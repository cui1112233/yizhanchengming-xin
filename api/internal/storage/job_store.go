package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

var ErrJobNotFound = errors.New("pipeline job not found")

type jobTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type jobRows interface {
	Next() bool
	Scan(...any) error
	Close() error
	Err() error
}

type beginJobTx func(context.Context) (jobTx, error)
type queryJobs func(context.Context, string, ...any) (jobRows, error)

type SQLJobStore struct {
	begin beginJobTx
	query queryJobs
}

func NewSQLJobStore(db *sql.DB) *SQLJobStore {
	if db == nil {
		return &SQLJobStore{}
	}
	return &SQLJobStore{
		begin: func(ctx context.Context) (jobTx, error) { return db.BeginTx(ctx, nil) },
		query: func(ctx context.Context, query string, args ...any) (jobRows, error) { return db.QueryContext(ctx, query, args...) },
	}
}

func newSQLJobStoreWithBegin(begin beginJobTx) *SQLJobStore { return &SQLJobStore{begin: begin} }
func newSQLJobStoreWithQuery(query queryJobs) *SQLJobStore { return &SQLJobStore{query: query} }

func (s *SQLJobStore) CreateJob(ctx context.Context, job pipeline.Job) error {
	if s == nil || s.begin == nil { return errors.New("sql job store database is required") }
	tx, err := s.begin(ctx)
	if err != nil { return err }
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO pipeline_jobs (id, owner, intake_id, status, run_at, idempotency_key) SELECT ?, owner, id, ?, ?, ? FROM intakes WHERE id = ?`,
		job.ID, job.Status, nullableJobTime(job.RunAt), job.ID, job.IntakeID,
	)
	if err != nil { return err }
	rows, err := result.RowsAffected()
	if err != nil { return err }
	if rows != 1 { return errors.New("intake not found") }

	for index, stage := range job.Stages {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO pipeline_stages (job_id, stage, ordinal_no, status) VALUES (?, ?, ?, 'pending')`,
			job.ID, string(stage), index+1,
		); err != nil { return err }
	}
	return tx.Commit()
}

func (s *SQLJobStore) LoadJob(ctx context.Context, jobID string) (pipeline.Job, error) {
	if s == nil || s.query == nil { return pipeline.Job{}, errors.New("sql job store database is required") }
	rows, err := s.query(ctx, `SELECT id, intake_id, batch_id, status, run_at FROM pipeline_jobs WHERE id = ?`, jobID)
	if err != nil { return pipeline.Job{}, err }
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil { return pipeline.Job{}, err }
		return pipeline.Job{}, ErrJobNotFound
	}
	var job pipeline.Job
	var intakeID, batchID sql.NullString
	var runAt sql.NullTime
	if err := rows.Scan(&job.ID, &intakeID, &batchID, &job.Status, &runAt); err != nil { return pipeline.Job{}, err }
	if intakeID.Valid { job.IntakeID = intakeID.String }
	if batchID.Valid { job.BatchID = batchID.String }
	if runAt.Valid { job.RunAt = runAt.Time }
	if err := rows.Err(); err != nil { return pipeline.Job{}, err }

	stageRows, err := s.query(ctx, `SELECT stage, ordinal_no FROM pipeline_stages WHERE job_id = ? ORDER BY ordinal_no ASC`, jobID)
	if err != nil { return pipeline.Job{}, err }
	defer stageRows.Close()
	for stageRows.Next() {
		var stage string
		var ordinal int
		if err := stageRows.Scan(&stage, &ordinal); err != nil { return pipeline.Job{}, err }
		job.Stages = append(job.Stages, pipeline.Stage(stage))
	}
	if err := stageRows.Err(); err != nil { return pipeline.Job{}, err }
	return job, nil
}

func nullableJobTime(value time.Time) any {
	if value.IsZero() { return nil }
	return value
}
