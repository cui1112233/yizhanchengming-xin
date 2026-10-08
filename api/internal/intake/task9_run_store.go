package intake

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// CreateRunIdempotent stores the Task 9 logical execution key in the existing
// runs table. The unique (batch_project_id,idempotency_key) constraint is the
// concurrency authority; duplicate callers receive the already-created Run.
func (s *MySQLStore) CreateRunIdempotent(ctx context.Context, run Run, idempotencyKey string) (Run, error) {
	key := strings.TrimSpace(idempotencyKey)
	if key == "" {
		return Run{}, errors.New("idempotency key required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE`, run.BatchProjectID).Scan(&archivedAt); err != nil {
		return Run{}, err
	}
	if archivedAt.Valid {
		return Run{}, ErrBatchProjectArchived
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO runs (batch_project_id,idempotency_key,run_at,status) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`,
		run.BatchProjectID, key, run.RunAt, run.Status,
	)
	if err != nil {
		return Run{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Run{}, err
	}
	if err := tx.QueryRowContext(ctx,
		`SELECT id,batch_project_id,run_at,status FROM runs WHERE id=?`, id,
	).Scan(&run.ID, &run.BatchProjectID, &run.RunAt, &run.Status); err != nil {
		return Run{}, err
	}
	if err := tx.Commit(); err != nil {
		return Run{}, err
	}
	return run, nil
}
