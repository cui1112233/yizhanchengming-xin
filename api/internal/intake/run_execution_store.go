package intake

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *MySQLStore) StartRun(ctx context.Context, runID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx,
		"UPDATE runs SET status = ? WHERE id = ? AND status = ?",
		RunStatusRunning, runID, RunStatusPending,
	)
	if err != nil {
		return false, fmt.Errorf("start run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("start run rows affected: %w", err)
	}
	return affected == 1, nil
}

func (s *MySQLStore) EnsureBookRuns(ctx context.Context, runID int64) error {
	const query = "INSERT INTO book_runs (run_id, book_id, status, idempotency_key) SELECT r.id, b.id, 'pending', CONCAT('run:', r.id, ':book:', b.id) FROM runs r JOIN batch_projects bp ON bp.id = r.batch_project_id JOIN books b ON b.intake_id = bp.intake_id WHERE r.id = ? ON DUPLICATE KEY UPDATE id = id"
	if _, err := s.db.ExecContext(ctx, query, runID); err != nil {
		return fmt.Errorf("ensure book runs: %w", err)
	}
	return nil
}

func (s *MySQLStore) RecoverExpiredBookRuns(ctx context.Context, runID int64, now time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		"UPDATE book_runs SET status = 'pending', lease_until = NULL WHERE run_id = ? AND status = 'running' AND lease_until IS NOT NULL AND lease_until < ?",
		runID, now,
	)
	if err != nil {
		return 0, fmt.Errorf("recover expired book runs: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("recover expired book runs rows affected: %w", err)
	}
	return affected, nil
}

func (s *MySQLStore) RetryBookRun(ctx context.Context, bookRunID int64) (bool, error) {
	result, err := s.db.ExecContext(ctx,
		"UPDATE book_runs SET status = 'pending', error_message = '', lease_until = NULL WHERE id = ? AND status = 'failed'",
		bookRunID,
	)
	if err != nil {
		return false, fmt.Errorf("retry book run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("retry book run rows affected: %w", err)
	}
	return affected == 1, nil
}

func (s *MySQLStore) ListPendingBookRuns(ctx context.Context, runID int64, limit int) ([]BookRun, error) {
	if limit <= 0 {
		limit = 100
	}
	const query = "SELECT id, run_id, book_id, status, attempt, error_message, idempotency_key, lease_until, created_at, updated_at FROM book_runs WHERE run_id = ? AND status = 'pending' ORDER BY id ASC LIMIT ?"
	rows, err := s.db.QueryContext(ctx, query, runID, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending book runs: %w", err)
	}
	defer rows.Close()

	items := make([]BookRun, 0)
	for rows.Next() {
		var item BookRun
		var lease sql.NullTime
		if err := rows.Scan(&item.ID, &item.RunID, &item.BookID, &item.Status, &item.Attempt, &item.ErrorMessage, &item.IdempotencyKey, &lease, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pending book run: %w", err)
		}
		if lease.Valid {
			value := lease.Time
			item.LeaseUntil = &value
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending book runs: %w", err)
	}
	return items, nil
}

func (s *MySQLStore) ClaimBookRun(ctx context.Context, bookRunID int64, leaseUntil time.Time) (bool, error) {
	result, err := s.db.ExecContext(ctx,
		"UPDATE book_runs SET status = 'running', attempt = attempt + 1, error_message = '', lease_until = ? WHERE id = ? AND status = 'pending'",
		leaseUntil, bookRunID,
	)
	if err != nil {
		return false, fmt.Errorf("claim book run: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("claim book run rows affected: %w", err)
	}
	return affected == 1, nil
}

func (s *MySQLStore) CompleteBookRun(ctx context.Context, bookRunID int64) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE book_runs SET status = 'completed', error_message = '', lease_until = NULL WHERE id = ? AND status = 'running'",
		bookRunID,
	)
	if err != nil {
		return fmt.Errorf("complete book run: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("complete book run rows affected: %w", err)
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *MySQLStore) FailBookRun(ctx context.Context, bookRunID int64, message string) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE book_runs SET status = 'failed', error_message = ?, lease_until = NULL WHERE id = ? AND status = 'running'",
		message, bookRunID,
	)
	if err != nil {
		return fmt.Errorf("fail book run: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("fail book run rows affected: %w", err)
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *MySQLStore) FinalizeRun(ctx context.Context, runID int64) error {
	const query = "UPDATE runs r SET status = CASE WHEN EXISTS (SELECT 1 FROM book_runs br WHERE br.run_id = r.id AND br.status IN ('pending', 'running')) THEN 'running' WHEN EXISTS (SELECT 1 FROM book_runs br WHERE br.run_id = r.id AND br.status = 'failed') THEN 'failed' ELSE 'completed' END WHERE r.id = ?"
	result, err := s.db.ExecContext(ctx, query, runID)
	if err != nil {
		return fmt.Errorf("finalize run: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("finalize run rows affected: %w", err)
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
