package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type progressExec func(context.Context, string, ...any) (sql.Result, error)

type SQLProgressStore struct {
	exec progressExec
}

func NewSQLProgressStore(db *sql.DB) *SQLProgressStore {
	if db == nil {
		return &SQLProgressStore{}
	}
	return &SQLProgressStore{exec: db.ExecContext}
}

func newSQLProgressStoreWithExec(exec progressExec) *SQLProgressStore {
	return &SQLProgressStore{exec: exec}
}

func (s *SQLProgressStore) CompleteStage(ctx context.Context, jobID string, stage pipeline.Stage) error {
	if s == nil || s.exec == nil {
		return errors.New("sql progress store database is required")
	}
	result, err := s.exec(ctx,
		`UPDATE pipeline_stages SET status = 'succeeded', completed_at = CURRENT_TIMESTAMP(6), error_message = '' WHERE job_id = ? AND stage = ? AND status <> 'succeeded'`,
		jobID, string(stage),
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("pipeline stage not found or already complete")
	}
	_, err = s.exec(ctx,
		`UPDATE pipeline_jobs SET current_stage = ?, status = 'running', error_message = '' WHERE id = ?`,
		string(stage), jobID,
	)
	return err
}

func (s *SQLProgressStore) CompleteJob(ctx context.Context, jobID string) error {
	if s == nil || s.exec == nil {
		return errors.New("sql progress store database is required")
	}
	_, err := s.exec(ctx,
		`UPDATE pipeline_jobs SET status = 'succeeded', current_stage = '', error_message = '' WHERE id = ?`,
		jobID,
	)
	return err
}
