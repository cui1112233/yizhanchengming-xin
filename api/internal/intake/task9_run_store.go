package intake

import (
	"context"
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
	result, err := s.db.ExecContext(ctx,
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
	if err := s.db.QueryRowContext(ctx,
		`SELECT id,batch_project_id,run_at,status FROM runs WHERE id=?`, id,
	).Scan(&run.ID, &run.BatchProjectID, &run.RunAt, &run.Status); err != nil {
		return Run{}, err
	}
	return run, nil
}
