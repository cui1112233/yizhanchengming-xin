package storage

import (
    "context"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

func TestSQLJobStoreCreatesJobAndStagesInOneTransaction(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil { t.Fatal(err) }
    defer db.Close()

    runAt := time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC)
    job := pipeline.Job{
        ID: "job-1", BatchID: "batch-1", RunAt: runAt, Status: "queued",
        Stages: []pipeline.Stage{pipeline.StageFetchBook, pipeline.StageResolveMetadata, pipeline.StageCreateBatch},
    }

    mock.ExpectBegin()
    mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO pipeline_jobs (id, owner, batch_id, status, run_at, idempotency_key) SELECT ?, owner, id, ?, ?, ? FROM batches WHERE id = ?`)).
        WithArgs("job-1", "queued", runAt, "job-1", "batch-1").
        WillReturnResult(sqlmock.NewResult(1, 1))
    for i, stage := range job.Stages {
        mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO pipeline_stages (job_id, stage, ordinal_no, status) VALUES (?, ?, ?, 'pending')`)).
            WithArgs("job-1", string(stage), i+1).
            WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
    }
    mock.ExpectCommit()

    store := NewSQLJobStore(db)
    if err := store.CreateJob(context.Background(), job); err != nil { t.Fatal(err) }
    if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
