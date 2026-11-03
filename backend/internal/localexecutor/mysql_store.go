package localexecutor

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) CreatePairing(ctx context.Context, record PairingRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO local_executor_pairings
(id, owner_username, platform, code_hash, expires_at, consumed_at, created_at)
VALUES (?, ?, ?, ?, ?, NULL, ?)`, record.ID, record.OwnerUsername, record.Platform, record.CodeHash[:], record.ExpiresAt, record.CreatedAt)
	return err
}

// ListPendingPairings 复用现有 local_executor_pairings 表读取待配对记录，
// 不新增 schema；读取时顺手清理超过保留期的过期记录。最近一条未消费记录
// 即使已过期也会返回，由调用方通过 ExpiresAt 判断是否仍然有效。
func (s *MySQLStore) ListPendingPairings(ctx context.Context, owner, platform string, now time.Time) ([]PairingRecord, error) {
	// 清理是尽力而为：失败不能影响状态读取和既有配对流程。
	_, _ = s.db.ExecContext(ctx, `DELETE FROM local_executor_pairings
WHERE owner_username = ? AND expires_at < ?`, owner, now.Add(-pairingCleanupRetention))
	var record PairingRecord
	var codeHash, consumed []byte
	err := s.db.QueryRowContext(ctx, `SELECT id, owner_username, platform, code_hash, expires_at, consumed_at, created_at
FROM local_executor_pairings
WHERE owner_username = ? AND platform = ? AND consumed_at IS NULL AND created_at >= ?
ORDER BY created_at DESC, id ASC LIMIT 1`, owner, platform, now.Add(-pairingCleanupRetention)).Scan(
		&record.ID, &record.OwnerUsername, &record.Platform, &codeHash, &record.ExpiresAt, &consumed, &record.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	copy(record.CodeHash[:], codeHash)
	return []PairingRecord{record}, nil
}

func (s *MySQLStore) PairExecutor(ctx context.Context, codeHash SecretHash, platform string, executor ExecutorRecord, now time.Time) (ExecutorRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExecutorRecord{}, err
	}
	defer tx.Rollback()

	var pairing PairingRecord
	var consumed sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT id, owner_username, platform, expires_at, consumed_at, created_at
FROM local_executor_pairings WHERE code_hash = ? FOR UPDATE`, codeHash[:]).Scan(
		&pairing.ID, &pairing.OwnerUsername, &pairing.Platform, &pairing.ExpiresAt, &consumed, &pairing.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ExecutorRecord{}, ErrPairingInvalid
	}
	if err != nil {
		return ExecutorRecord{}, err
	}
	if consumed.Valid || pairing.Platform != platform || !pairing.ExpiresAt.After(now) {
		return ExecutorRecord{}, ErrPairingInvalid
	}

	executor.OwnerUsername = pairing.OwnerUsername
	var existingID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM local_executors
WHERE owner_username = ? AND platform = ? AND device_name = ? AND os = ?
ORDER BY updated_at DESC, id ASC LIMIT 1 FOR UPDATE`, executor.OwnerUsername, platform, executor.DeviceName, executor.OS).Scan(&existingID)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return ExecutorRecord{}, err
	}
	if existingID != "" {
		executor.ID = existingID
		_, err = tx.ExecContext(ctx, `UPDATE local_executors SET token_hash = ?, app_version = ?,
accounts_total = 0, accounts_available = 0, accounts_busy = 0, accounts_quota_exhausted = 0,
accounts_login_error = 0, accounts_human_verification = 0, last_seen_at = NULL, updated_at = ?
WHERE id = ?`, executor.TokenHash[:], executor.Version, executor.UpdatedAt, executor.ID)
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO local_executors
(id, owner_username, platform, token_hash, device_name, os, app_version,
 accounts_total, accounts_available, accounts_busy, accounts_quota_exhausted,
 accounts_login_error, accounts_human_verification, last_seen_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, 0, 0, NULL, ?, ?)`,
			executor.ID, executor.OwnerUsername, executor.Platform, executor.TokenHash[:], executor.DeviceName,
			executor.OS, executor.Version, executor.CreatedAt, executor.UpdatedAt)
	}
	if err != nil {
		return ExecutorRecord{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE local_executor_pairings SET consumed_at = ?
WHERE id = ? AND consumed_at IS NULL`, now, pairing.ID)
	if err != nil {
		return ExecutorRecord{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return ExecutorRecord{}, err
	}
	if rows != 1 {
		return ExecutorRecord{}, ErrPairingInvalid
	}
	if err := tx.Commit(); err != nil {
		return ExecutorRecord{}, err
	}
	return executor, nil
}

func (s *MySQLStore) ExecutorByTokenHash(ctx context.Context, tokenHash SecretHash) (ExecutorRecord, error) {
	row := s.db.QueryRowContext(ctx, executorSelect+` WHERE token_hash = ?`, tokenHash[:])
	record, err := scanExecutor(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	return record, err
}

func (s *MySQLStore) UpdateHeartbeat(ctx context.Context, id string, input HeartbeatInput, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE local_executors SET
 device_name = CASE WHEN ? = '' THEN device_name ELSE ? END,
 os = CASE WHEN ? = '' THEN os ELSE ? END,
 app_version = CASE WHEN ? = '' THEN app_version ELSE ? END,
 accounts_total = ?, accounts_available = ?, accounts_busy = ?, accounts_quota_exhausted = ?,
 accounts_login_error = ?, accounts_human_verification = ?, last_seen_at = ?, updated_at = ?
WHERE id = ?`,
		input.DeviceName, input.DeviceName, input.OS, input.OS, input.Version, input.Version,
		input.Accounts.Total, input.Accounts.Available, input.Accounts.Busy, input.Accounts.QuotaExhausted,
		input.Accounts.LoginError, input.Accounts.HumanVerification, now, now, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrExecutorUnauthorized
	}
	return nil
}

func (s *MySQLStore) ListExecutors(ctx context.Context, owner string) ([]ExecutorRecord, error) {
	rows, err := s.db.QueryContext(ctx, executorSelect+` WHERE owner_username = ? ORDER BY updated_at DESC, id ASC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ExecutorRecord, 0)
	for rows.Next() {
		record, err := scanExecutor(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

const executorSelect = `SELECT id, owner_username, platform, device_name, os, app_version,
 accounts_total, accounts_available, accounts_busy, accounts_quota_exhausted,
 accounts_login_error, accounts_human_verification, last_seen_at, created_at, updated_at
FROM local_executors`

type scanFunc func(dest ...any) error

func scanExecutor(scan scanFunc) (ExecutorRecord, error) {
	var record ExecutorRecord
	var lastSeen sql.NullTime
	err := scan(
		&record.ID, &record.OwnerUsername, &record.Platform, &record.DeviceName, &record.OS, &record.Version,
		&record.Accounts.Total, &record.Accounts.Available, &record.Accounts.Busy, &record.Accounts.QuotaExhausted,
		&record.Accounts.LoginError, &record.Accounts.HumanVerification, &lastSeen, &record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return ExecutorRecord{}, err
	}
	if lastSeen.Valid {
		seen := lastSeen.Time
		record.LastSeenAt = &seen
	}
	return record, nil
}
