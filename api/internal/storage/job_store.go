package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type SQLJobStore struct {
	db *sql.DB
}

func NewSQLJobStore(db *sql.DB) *SQLJobStore {
	return &SQLJobStore{db: db}
}

func (s *SQLJobStore) CreateJob(ctx context.Context, job pipeline.Job) error {
	if s == nil || s.db == nil {
		return errors.New("sql job store database is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO pipeline_jobs (id, owner, batch_id, status, run_at, idempotency_key) SELECT ?, owner, id, ?, ?, ? FROM batches WHERE id = ?`,
		job.ID, job.Status, nullableTime(job.RunAt), job.ID, job.BatchID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("batch not found")
	}

	for index, stage := range job.Stages {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pipeline_stages (job_id, stage, ordinal_no, status) VALUES (?, ?, ?, 'pending')`,
			job.ID, string(stage), index+1,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nullableTime(value interface{ IsZero() bool }) any {
	if value.IsZero() {
		return nil
	}
	return value
}
