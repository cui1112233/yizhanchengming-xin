package giantmaterialexecutor

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) EnsureIdentityUniqueIndex(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM information_schema.statistics
WHERE table_schema = DATABASE() AND table_name = 'giant_executors'
AND index_name = 'uq_giant_executors_identity'`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `ALTER TABLE giant_executors
ADD UNIQUE KEY uq_giant_executors_identity (owner_username, os, device_name)`)
	return err
}

func (s *MySQLStore) ConsolidateDuplicateExecutors(ctx context.Context, now time.Time) (int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT owner_username, os, device_name
FROM giant_executors GROUP BY owner_username, os, device_name HAVING COUNT(*) > 1`)
	if err != nil {
		return 0, err
	}
	type groupRow struct{ owner, osName, deviceName string }
	groups := make([]groupRow, 0)
	for rows.Next() {
		var group groupRow
		if err := rows.Scan(&group.owner, &group.osName, &group.deviceName); err != nil {
			rows.Close()
			return 0, err
		}
		groups = append(groups, group)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	merged := 0
	for _, group := range groups {
		if err := s.consolidateOneGroup(ctx, group.owner, group.osName, group.deviceName, now); err != nil {
			return merged, err
		}
		merged++
	}
	return merged, nil
}

func (s *MySQLStore) consolidateOneGroup(ctx context.Context, owner, osName, deviceName string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id, app_version, last_seen_at, updated_at
FROM giant_executors WHERE owner_username = ? AND os = ? AND device_name = ? ORDER BY id ASC FOR UPDATE`,
		owner, osName, deviceName)
	if err != nil {
		return err
	}
	candidates := make([]executorCandidate, 0)
	for rows.Next() {
		var item executorCandidate
		var lastSeen sql.NullTime
		if err := rows.Scan(&item.id, &item.version, &lastSeen, &item.updatedAt); err != nil {
			rows.Close()
			return err
		}
		if lastSeen.Valid {
			seen := lastSeen.Time
			item.lastSeen = &seen
		}
		candidates = append(candidates, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil || len(candidates) < 2 {
		return err
	}
	keeping := pickKeeping(candidates)
	oldIDs := make([]string, 0, len(candidates)-1)
	for _, item := range candidates {
		if item.id != keeping {
			oldIDs = append(oldIDs, item.id)
		}
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(oldIDs)), ",")
	args := make([]any, 0, len(oldIDs)+1)
	args = append(args, keeping)
	for _, id := range oldIDs {
		args = append(args, id)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE giant_executor_jobs SET lease_executor_id = ? WHERE lease_executor_id IN (`+placeholders+`)`, args...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE giant_executor_job_events SET executor_id = ? WHERE executor_id IN (`+placeholders+`)`, args...); err != nil {
		return err
	}
	deleteArgs := make([]any, 0, len(oldIDs))
	for _, id := range oldIDs {
		deleteArgs = append(deleteArgs, id)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM giant_executors WHERE id IN (`+placeholders+`)`, deleteArgs...); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) RecentFailures(ctx context.Context, owner string, since time.Time) (map[string]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT executor_id, MAX(created_at)
FROM giant_executor_job_events
WHERE event_type = 'failed' AND created_at >= ?
AND executor_id IN (SELECT id FROM giant_executors WHERE owner_username = ?)
GROUP BY executor_id`, since, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]time.Time)
	for rows.Next() {
		var id string
		var value time.Time
		if err := rows.Scan(&id, &value); err != nil {
			return nil, err
		}
		out[id] = value
	}
	return out, rows.Err()
}

func (s *MySQLStore) CreatePairing(ctx context.Context, record PairingRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO giant_executor_pairings
(id, owner_username, platform, code_hash, expires_at, consumed_at, created_at)
VALUES (?, ?, ?, ?, ?, NULL, ?)`, record.ID, record.OwnerUsername, record.Platform, record.CodeHash[:], record.ExpiresAt, record.CreatedAt)
	return err
}

// ListPendingPairings 复用现有 giant_executor_pairings 表读取待配对记录，
// 不新增 schema；读取时顺手清理超过保留期的过期记录。最近一条未消费记录
// 即使已过期也会返回，由调用方通过 ExpiresAt 判断是否仍然有效。
func (s *MySQLStore) ListPendingPairings(ctx context.Context, owner, platform string, now time.Time) ([]PairingRecord, error) {
	// 清理是尽力而为：失败不能影响状态读取和既有配对流程。
	_, _ = s.db.ExecContext(ctx, `DELETE FROM giant_executor_pairings
WHERE owner_username = ? AND expires_at < ?`, owner, now.Add(-pairingCleanupRetention))
	var record PairingRecord
	var codeHash, consumed []byte
	err := s.db.QueryRowContext(ctx, `SELECT id, owner_username, platform, code_hash, expires_at, consumed_at, created_at
FROM giant_executor_pairings
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
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		record, retryable, err := s.pairExecutorTx(ctx, codeHash, platform, executor, now)
		if err == nil {
			return record, nil
		}
		if !retryable {
			return ExecutorRecord{}, err
		}
		lastErr = err
	}
	return ExecutorRecord{}, lastErr
}

func (s *MySQLStore) pairExecutorTx(ctx context.Context, codeHash SecretHash, platform string, executor ExecutorRecord, now time.Time) (ExecutorRecord, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExecutorRecord{}, isRetryableMySQLError(err), err
	}
	defer tx.Rollback()
	var pairing PairingRecord
	var consumed sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT id, owner_username, platform, expires_at, consumed_at, created_at
FROM giant_executor_pairings WHERE code_hash = ? FOR UPDATE`, codeHash[:]).Scan(&pairing.ID, &pairing.OwnerUsername, &pairing.Platform, &pairing.ExpiresAt, &consumed, &pairing.CreatedAt); errors.Is(err, sql.ErrNoRows) {
		return ExecutorRecord{}, false, ErrPairingInvalid
	} else if err != nil {
		return ExecutorRecord{}, isRetryableMySQLError(err), err
	}
	if consumed.Valid || pairing.Platform != platform || !pairing.ExpiresAt.After(now) {
		return ExecutorRecord{}, false, ErrPairingInvalid
	}
	executor.OwnerUsername = pairing.OwnerUsername
	var existingID string
	lookupErr := tx.QueryRowContext(ctx, `SELECT id FROM giant_executors
WHERE owner_username = ? AND os = ? AND device_name = ?
ORDER BY updated_at DESC, id ASC LIMIT 1 FOR UPDATE`,
		executor.OwnerUsername, executor.OS, executor.DeviceName).Scan(&existingID)
	if errors.Is(lookupErr, sql.ErrNoRows) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO giant_executors
(id, owner_username, platform, token_hash, device_name, os, app_version, last_seen_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)`, executor.ID, executor.OwnerUsername, executor.Platform,
			executor.TokenHash[:], executor.DeviceName, executor.OS, executor.Version,
			executor.CreatedAt, executor.UpdatedAt); err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				// 并发配对竞态：身份行已被另一事务插入，同一事务内重新锁定该行后走 UPDATE。
				reErr := tx.QueryRowContext(ctx, `SELECT id FROM giant_executors
WHERE owner_username = ? AND os = ? AND device_name = ?
ORDER BY updated_at DESC, id ASC LIMIT 1 FOR UPDATE`,
					executor.OwnerUsername, executor.OS, executor.DeviceName).Scan(&existingID)
				switch {
				case reErr == nil:
					// 查到了竞争事务插入的行，下面统一走 UPDATE 分支。
				case errors.Is(reErr, sql.ErrNoRows):
					return ExecutorRecord{}, false, err
				default:
					return ExecutorRecord{}, isRetryableMySQLError(reErr), reErr
				}
			} else {
				return ExecutorRecord{}, isRetryableMySQLError(err), err
			}
		}
	} else if lookupErr != nil {
		return ExecutorRecord{}, isRetryableMySQLError(lookupErr), lookupErr
	}
	if existingID != "" {
		executor.ID = existingID
		executor.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executors
SET token_hash = ?, app_version = ?, updated_at = ? WHERE id = ?`,
			executor.TokenHash[:], executor.Version, now, existingID); err != nil {
			return ExecutorRecord{}, isRetryableMySQLError(err), err
		}
	}
	result, err := tx.ExecContext(ctx, `UPDATE giant_executor_pairings SET consumed_at = ? WHERE id = ? AND consumed_at IS NULL`, now, pairing.ID)
	if err != nil {
		return ExecutorRecord{}, isRetryableMySQLError(err), err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return ExecutorRecord{}, false, ErrPairingInvalid
	}
	if err := tx.Commit(); err != nil {
		return ExecutorRecord{}, isRetryableMySQLError(err), err
	}
	return executor, false, nil
}

func isRetryableMySQLError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1213 || mysqlErr.Number == 1205
	}
	return false
}

func (s *MySQLStore) ExecutorByTokenHash(ctx context.Context, hash SecretHash) (ExecutorRecord, error) {
	var record ExecutorRecord
	var lastSeen sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id, owner_username, platform, device_name, os, app_version, last_seen_at, created_at, updated_at
FROM giant_executors WHERE token_hash = ?`, hash[:]).Scan(&record.ID, &record.OwnerUsername, &record.Platform, &record.DeviceName, &record.OS, &record.Version, &lastSeen, &record.CreatedAt, &record.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	if err != nil {
		return ExecutorRecord{}, err
	}
	if lastSeen.Valid {
		seen := lastSeen.Time
		record.LastSeenAt = &seen
	}
	return record, nil
}

func (s *MySQLStore) UpdateHeartbeat(ctx context.Context, id string, input HeartbeatInput, now time.Time) error {
	deviceName := strings.TrimSpace(input.DeviceName)
	osName := strings.TrimSpace(input.OS)
	version := strings.TrimSpace(input.Version)
	result, err := s.db.ExecContext(ctx, `UPDATE giant_executors SET
 device_name = CASE WHEN ? = '' THEN device_name ELSE ? END,
 os = CASE WHEN ? = '' THEN os ELSE ? END,
 app_version = CASE WHEN ? = '' THEN app_version ELSE ? END,
 last_seen_at = ?, updated_at = ? WHERE id = ?`, deviceName, deviceName, osName, osName, version, version, now, now, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return ErrExecutorUnauthorized
	}
	return nil
}

func (s *MySQLStore) ListExecutors(ctx context.Context, owner string) ([]ExecutorRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, owner_username, platform, device_name, os, app_version, last_seen_at, created_at, updated_at
FROM giant_executors WHERE owner_username = ? ORDER BY updated_at DESC, id ASC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ExecutorRecord, 0)
	for rows.Next() {
		var record ExecutorRecord
		var lastSeen sql.NullTime
		if err := rows.Scan(&record.ID, &record.OwnerUsername, &record.Platform, &record.DeviceName, &record.OS, &record.Version, &lastSeen, &record.CreatedAt, &record.UpdatedAt); err != nil {
			return nil, err
		}
		if lastSeen.Valid {
			seen := lastSeen.Time
			record.LastSeenAt = &seen
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

func (s *MySQLStore) LatestPlatformFailure(ctx context.Context, owner, osName string, since time.Time) (*time.Time, error) {
	var latest sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT MAX(e.created_at)
FROM giant_executor_job_events e
WHERE e.event_type = 'failed' AND e.created_at >= ?
AND e.executor_id IN (SELECT id FROM giant_executors WHERE owner_username = ? AND os = ?)`,
		since, owner, osName).Scan(&latest)
	if err != nil {
		return nil, err
	}
	if !latest.Valid {
		return nil, nil
	}
	value := latest.Time
	return &value, nil
}

const jobSelect = `SELECT j.id, j.owner_username, j.platform, j.material_id, j.platform_book_id, j.title,
 j.video_url, j.video_expires_at, j.duration_seconds, j.model_version, j.content_range_lines, j.target_executor_id, j.state, j.cancel_requested,
 j.progress_completed, j.progress_total, j.progress_percent, j.lease_executor_id, j.lease_token_hash,
 j.lease_generation, j.lease_expires_at, j.error_code, j.error_message, j.created_at, j.updated_at,
 r.text_body, r.word_count, r.created_at
FROM giant_executor_jobs j LEFT JOIN giant_executor_results r ON r.job_id = j.id`

func (s *MySQLStore) FindJobByKey(ctx context.Context, owner, key string) (JobRecord, error) {
	parts := splitJobKey(key)
	if len(parts) != 3 {
		return JobRecord{}, ErrJobNotFound
	}
	return s.queryJob(ctx, jobSelect+` WHERE j.owner_username = ? AND j.material_id = ? AND j.platform_book_id = ? AND j.model_version = ?`, owner, parts[0], parts[1], parts[2])
}

func (s *MySQLStore) CreateJob(ctx context.Context, record JobRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO giant_executor_jobs
(id, owner_username, platform, material_id, platform_book_id, title, video_url, video_expires_at, duration_seconds, model_version, content_range_lines, target_executor_id, state, cancel_requested,
 progress_completed, progress_total, progress_percent, lease_executor_id, lease_token_hash, lease_generation, lease_expires_at, error_code, error_message, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, FALSE, 0, 0, 0, NULL, NULL, 0, NULL, NULL, NULL, ?, ?)`, record.ID, record.OwnerUsername, record.Platform, record.MaterialID, record.PlatformBookID, record.Title, record.VideoURL, record.VideoExpiresAt, record.DurationSeconds, record.ModelVersion, record.ContentRangeLines, record.TargetExecutorID, record.State, record.CreatedAt, record.UpdatedAt)
	return err
}

func (s *MySQLStore) JobForOwner(ctx context.Context, owner, id string) (JobRecord, error) {
	return s.queryJob(ctx, jobSelect+` WHERE j.owner_username = ? AND j.id = ?`, owner, id)
}

func (s *MySQLStore) FailQueuedJob(ctx context.Context, owner, id string, failure FailureInput, now time.Time) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	record, err := queryJobTx(ctx, tx, jobSelect+` WHERE j.owner_username = ? AND j.id = ? FOR UPDATE`, owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	if err != nil {
		return JobRecord{}, err
	}
	if record.State == JobQueued {
		record.State = JobFailed
		record.ErrorCode = bounded(failure.Code, 96)
		record.ErrorMessage = bounded(failure.Message, 512)
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs SET state = ?, error_code = ?, error_message = ?, updated_at = ? WHERE id = ? AND state = ?`, JobFailed, record.ErrorCode, record.ErrorMessage, now, id, JobQueued); err != nil {
			return JobRecord{}, err
		}
		if err := appendEvent(ctx, tx, id, "", "failed", JobFailed, record.ErrorMessage, now); err != nil {
			return JobRecord{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) CancelJob(ctx context.Context, owner, id string, now time.Time) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	record, err := queryJobTx(ctx, tx, jobSelect+` WHERE j.owner_username = ? AND j.id = ? FOR UPDATE`, owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	if err != nil {
		return JobRecord{}, err
	}
	if record.State != JobSucceeded && record.State != JobFailed && record.State != JobCancelled {
		record.CancelRequested = true
		record.State = JobCancelled
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs SET cancel_requested = TRUE, state = ?, updated_at = ? WHERE id = ?`, JobCancelled, now, id); err != nil {
			return JobRecord{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) ClaimJob(ctx context.Context, executor ExecutorRecord, leaseHash SecretHash, expires, now time.Time, targetedOnly bool) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	targetFilter := "AND j.target_executor_id IS NULL"
	args := []any{executor.OwnerUsername, PlatformGiantMaterial}
	if targetedOnly {
		// 只领明确点名给本执行器的任务（定向重试），平台偏好门控不拦截。
		targetFilter = "AND j.target_executor_id = ?"
		args = append(args, executor.ID)
	}
	args = append(args, JobQueued,
		now, JobSucceeded, JobFailed, JobCancelled,
		JobLeased, JobRunning, JobCleaning, JobUploading,
		now, now.Add(-StuckProgressLimit))
	record, err := queryJobTx(ctx, tx, jobSelect+` WHERE j.owner_username = ? AND j.platform = ? AND j.cancel_requested = FALSE
	`+targetFilter+`
	AND (j.state = ?
 OR (j.lease_expires_at IS NOT NULL AND j.lease_expires_at <= ? AND j.state NOT IN (?, ?, ?))
 OR (j.state IN (?, ?, ?, ?) AND j.lease_expires_at > ? AND j.progress_changed_at IS NOT NULL AND j.progress_changed_at <= ?))
 ORDER BY j.created_at ASC, j.id ASC LIMIT 1 FOR UPDATE SKIP LOCKED`, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrNoClaimableJob
	}
	if err != nil {
		return JobRecord{}, err
	}
	wasStuck := record.State != JobQueued
	record.State = JobLeased
	record.LeaseExecutorID = executor.ID
	record.LeaseTokenHash = leaseHash
	record.LeaseGeneration++
	record.LeaseExpiresAt = &expires
	record.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs
SET state = ?, lease_executor_id = ?, lease_token_hash = ?, lease_generation = ?, lease_expires_at = ?,
 progress_changed_at = ?, error_code = NULL, error_message = NULL, updated_at = ? WHERE id = ?`,
		JobLeased, executor.ID, leaseHash[:], record.LeaseGeneration, expires, now, now, record.ID); err != nil {
		return JobRecord{}, err
	}
	eventType := "claimed"
	if wasStuck {
		eventType = "stuck_reclaimed"
	}
	if err := appendEvent(ctx, tx, record.ID, executor.ID, eventType, JobLeased, "", now); err != nil {
		return JobRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) RenewJob(ctx context.Context, executorID, id string, leaseHash SecretHash, generation int64, expires, now time.Time) (JobRecord, error) {
	return s.mutateLease(ctx, executorID, id, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		record.LeaseExpiresAt = &expires
		record.UpdatedAt = now
		_, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs SET lease_expires_at = ?, updated_at = ? WHERE id = ?`, expires, now, id)
		return err
	})
}

func (s *MySQLStore) SetProgress(ctx context.Context, executorID, id string, leaseHash SecretHash, generation int64, next JobState, progress ProgressInput, now time.Time) (JobRecord, error) {
	return s.mutateLease(ctx, executorID, id, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		if err := validJobTransition(record.State, next); err != nil {
			return err
		}
		record.State = next
		record.Progress = progress
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs
SET state = ?, progress_completed = ?, progress_total = ?, progress_percent = ?,
 progress_changed_at = CASE WHEN progress_completed <> ? OR progress_total <> ? OR progress_percent <> ? THEN ? ELSE progress_changed_at END,
 updated_at = ? WHERE id = ?`,
			next, progress.Completed, progress.Total, progress.Percent,
			progress.Completed, progress.Total, progress.Percent, now, now, id); err != nil {
			return err
		}
		return appendEvent(ctx, tx, id, executorID, "progress", next, "", now)
	})
}

func (s *MySQLStore) CompleteJob(ctx context.Context, executorID, id string, leaseHash SecretHash, generation int64, result ResultInput, now time.Time) (JobRecord, error) {
	return s.mutateLease(ctx, executorID, id, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		if record.State != JobUploading {
			return ErrInvalidJobState
		}
		record.State = JobSucceeded
		record.Result = ResultRecord{Text: result.Text, WordCount: result.WordCount, CreatedAt: now}
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs SET state = ?, updated_at = ? WHERE id = ?`, JobSucceeded, now, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO giant_executor_results(job_id, text_body, word_count, created_at) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE text_body = VALUES(text_body), word_count = VALUES(word_count)`, id, result.Text, result.WordCount, now); err != nil {
			return err
		}
		return appendEvent(ctx, tx, id, executorID, "succeeded", JobSucceeded, "", now)
	})
}

func (s *MySQLStore) FailJob(ctx context.Context, executorID, id string, leaseHash SecretHash, generation int64, failure FailureInput, now time.Time) (JobRecord, error) {
	return s.mutateLease(ctx, executorID, id, leaseHash, generation, now, func(tx *sql.Tx, record *JobRecord) error {
		record.State = JobFailed
		record.ErrorCode = bounded(failure.Code, 96)
		record.ErrorMessage = bounded(failure.Message, 512)
		record.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs SET state = ?, error_code = ?, error_message = ?, updated_at = ? WHERE id = ?`, JobFailed, record.ErrorCode, record.ErrorMessage, now, id); err != nil {
			return err
		}
		return appendEvent(ctx, tx, id, executorID, "failed", JobFailed, record.ErrorCode, now)
	})
}

func (s *MySQLStore) RequeueJob(ctx context.Context, id string, update JobRecord, now time.Time) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	record, err := queryJobTx(ctx, tx, jobSelect+` WHERE j.id = ? FOR UPDATE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	if err != nil {
		return JobRecord{}, err
	}
	if record.State != JobFailed && record.State != JobCancelled {
		return record, nil
	}
	record.State = JobQueued
	record.CancelRequested = false
	record.Title = update.Title
	record.VideoURL = update.VideoURL
	record.VideoExpiresAt = update.VideoExpiresAt
	record.DurationSeconds = update.DurationSeconds
	record.ContentRangeLines = update.ContentRangeLines
	record.TargetExecutorID = update.TargetExecutorID
	record.LeaseExecutorID = ""
	record.LeaseTokenHash = SecretHash{}
	record.LeaseGeneration++
	record.LeaseExpiresAt = nil
	record.ErrorCode = ""
	record.ErrorMessage = ""
	record.Progress = ProgressInput{}
	record.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs SET state = ?, cancel_requested = FALSE, title = ?, video_url = ?, video_expires_at = ?, duration_seconds = ?, content_range_lines = ?, target_executor_id = NULLIF(?, ''),
	 lease_executor_id = NULL, lease_token_hash = NULL, lease_generation = ?, lease_expires_at = NULL, progress_changed_at = NULL,
	 error_code = NULL, error_message = NULL, progress_completed = 0, progress_total = 0, progress_percent = 0, updated_at = ?
	 WHERE id = ?`, record.State, record.Title, record.VideoURL, record.VideoExpiresAt, record.DurationSeconds, record.ContentRangeLines, record.TargetExecutorID, record.LeaseGeneration, now, id); err != nil {
		return JobRecord{}, err
	}
	if err := appendEvent(ctx, tx, id, "", "requeued", JobQueued, "", now); err != nil {
		return JobRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) mutateLease(ctx context.Context, executorID, id string, leaseHash SecretHash, generation int64, now time.Time, mutate func(*sql.Tx, *JobRecord) error) (JobRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return JobRecord{}, err
	}
	defer tx.Rollback()
	record, err := queryJobTx(ctx, tx, jobSelect+` WHERE j.id = ? FOR UPDATE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	if err != nil {
		return JobRecord{}, err
	}
	if err := checkLease(record, executorID, leaseHash, generation, now); err != nil {
		return JobRecord{}, err
	}
	if err := mutate(tx, &record); err != nil {
		return JobRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return JobRecord{}, err
	}
	return record, nil
}

func appendEvent(ctx context.Context, tx *sql.Tx, jobID, executorID, eventType string, state JobState, detail string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO giant_executor_job_events(job_id, executor_id, event_type, state, detail, created_at) VALUES (?, NULLIF(?, ''), ?, ?, NULLIF(?, ''), ?)`, jobID, executorID, eventType, state, bounded(detail, 512), now)
	return err
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *MySQLStore) queryJob(ctx context.Context, query string, args ...any) (JobRecord, error) {
	record, err := scanJob(s.db.QueryRowContext(ctx, query, args...).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return JobRecord{}, ErrJobNotFound
	}
	return record, err
}

func queryJobTx(ctx context.Context, tx *sql.Tx, query string, args ...any) (JobRecord, error) {
	return scanJob(tx.QueryRowContext(ctx, query, args...).Scan)
}

type scanFunc func(...any) error

func scanJob(scan scanFunc) (JobRecord, error) {
	var record JobRecord
	var videoExpires, leaseExpires, resultCreated sql.NullTime
	var targetExecutor, leaseExecutor, errorCode, errorMessage, resultText sql.NullString
	var leaseHash []byte
	var wordCount sql.NullInt64
	err := scan(&record.ID, &record.OwnerUsername, &record.Platform, &record.MaterialID, &record.PlatformBookID, &record.Title, &record.VideoURL, &videoExpires, &record.DurationSeconds, &record.ModelVersion, &record.ContentRangeLines, &targetExecutor, &record.State, &record.CancelRequested, &record.Progress.Completed, &record.Progress.Total, &record.Progress.Percent, &leaseExecutor, &leaseHash, &record.LeaseGeneration, &leaseExpires, &errorCode, &errorMessage, &record.CreatedAt, &record.UpdatedAt, &resultText, &wordCount, &resultCreated)
	if err != nil {
		return JobRecord{}, err
	}
	if videoExpires.Valid {
		value := videoExpires.Time
		record.VideoExpiresAt = &value
	}
	if leaseExecutor.Valid {
		record.LeaseExecutorID = leaseExecutor.String
	}
	if targetExecutor.Valid {
		record.TargetExecutorID = targetExecutor.String
	}
	if len(leaseHash) == len(record.LeaseTokenHash) {
		copy(record.LeaseTokenHash[:], leaseHash)
	}
	if leaseExpires.Valid {
		value := leaseExpires.Time
		record.LeaseExpiresAt = &value
	}
	if errorCode.Valid {
		record.ErrorCode = errorCode.String
	}
	if errorMessage.Valid {
		record.ErrorMessage = errorMessage.String
	}
	if resultText.Valid {
		record.Result.Text = resultText.String
	}
	if wordCount.Valid {
		record.Result.WordCount = int(wordCount.Int64)
	}
	if resultCreated.Valid {
		record.Result.CreatedAt = resultCreated.Time
	}
	return record, nil
}

func splitJobKey(key string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for len(parts) < 2 {
		idx := indexNull(key[start:])
		if idx < 0 {
			break
		}
		parts = append(parts, key[start:start+idx])
		start += idx + 1
	}
	parts = append(parts, key[start:])
	return parts
}

func indexNull(value string) int {
	for i := range value {
		if value[i] == 0 {
			return i
		}
	}
	return -1
}

func checkLease(record JobRecord, executorID string, leaseHash SecretHash, generation int64, now time.Time) error {
	if record.CancelRequested || record.State == JobCancelled {
		return ErrJobCancelled
	}
	if record.LeaseExecutorID != executorID || record.LeaseGeneration != generation || record.LeaseTokenHash != leaseHash || record.LeaseExpiresAt == nil || !record.LeaseExpiresAt.After(now) {
		return ErrStaleLease
	}
	return nil
}

func (s *MySQLStore) GetPreference(ctx context.Context, owner string) (PreferenceRecord, error) {
	var record PreferenceRecord
	err := s.db.QueryRowContext(ctx, `SELECT owner_username, preferred_os, updated_at
FROM giant_executor_preferences WHERE owner_username = ?`, owner).Scan(&record.OwnerUsername, &record.PreferredOS, &record.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return PreferenceRecord{}, ErrPreferenceNotFound
	}
	if err != nil {
		return PreferenceRecord{}, err
	}
	return record, nil
}

func (s *MySQLStore) SavePreference(ctx context.Context, record PreferenceRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO giant_executor_preferences (owner_username, preferred_os, updated_at)
VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE preferred_os = VALUES(preferred_os), updated_at = VALUES(updated_at)`,
		record.OwnerUsername, record.PreferredOS, record.UpdatedAt)
	return err
}

func (s *MySQLStore) DeleteExecutor(ctx context.Context, owner, id string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var dbOwner string
	err = tx.QueryRowContext(ctx, `SELECT owner_username FROM giant_executors WHERE id = ? FOR UPDATE`, id).Scan(&dbOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrExecutorNotFound
	}
	if err != nil {
		return err
	}
	if dbOwner != owner {
		return ErrExecutorNotFound
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM giant_executor_jobs
WHERE lease_executor_id = ? AND state NOT IN (?, ?, ?) FOR UPDATE`,
		id, JobSucceeded, JobFailed, JobCancelled)
	if err != nil {
		return err
	}
	jobIDs := make([]string, 0)
	for rows.Next() {
		var jobID string
		if err := rows.Scan(&jobID); err != nil {
			rows.Close()
			return err
		}
		jobIDs = append(jobIDs, jobID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		if _, err := tx.ExecContext(ctx, `UPDATE giant_executor_jobs
SET state = 'queued', cancel_requested = FALSE, lease_executor_id = NULL, lease_token_hash = NULL,
 lease_expires_at = NULL, lease_generation = lease_generation + 1,
 progress_completed = 0, progress_total = 0, progress_percent = 0, progress_changed_at = NULL,
 error_code = NULL, error_message = NULL, updated_at = ? WHERE id = ?`, now, jobID); err != nil {
			return err
		}
		if err := appendEvent(ctx, tx, jobID, "", "requeued", JobQueued, "", now); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM giant_executors WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
