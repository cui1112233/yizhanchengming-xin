package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type jobTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type beginJobTx func(context.Context) (jobTx, error)

type SQLJobStore struct {
	begin beginJobTx
}

func NewSQLJobStore(db *sql.DB) *SQLJobStore {
	if db == nil {
		return &SQLJobStore{}
	}
	return &SQLJobStore{begin: func(ctx context.Context) (jobTx, error) {
		return db.BeginTx(ctx, nil)
	}}
}

func newSQLJobStoreWithBegin(begin beginJobTx) *SQLJobStore {
	return &SQLJobStore{begin: begin}
}

func (s *SQLJobStore) CreateJob(ctx context.Context, job pipeline.Job) error {
	if s == nil || s.begin == nil {
		return errors.New("sql job store database is required")
	}
	tx, err := s.begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO pipeline_jobs (id, owner, batch_id, status, run_at, idempotency_key) SELECT ?, owner, id, ?, ?, ? FROM batches WHERE id = ?`,
		job.ID, job.Status, nullableJobTime(job.RunAt), job.ID, job.BatchID,
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

func nullableJobTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
