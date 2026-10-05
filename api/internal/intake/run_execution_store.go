package intake

import (
	"context"
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
