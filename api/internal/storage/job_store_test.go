package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeResult struct{ rows int64 }
func (f fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (f fakeResult) RowsAffected() (int64, error) { return f.rows, nil }

type fakeTx struct {
	queries []string
	args    [][]any
	commit  bool
}
func (f *fakeTx) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.queries = append(f.queries, query)
	f.args = append(f.args, append([]any(nil), args...))
	return fakeResult{rows: 1}, nil
}
func (f *fakeTx) Commit() error { f.commit = true; return nil }
func (f *fakeTx) Rollback() error { return nil }

func TestSQLJobStoreCreatesJobAndStagesInOneTransaction(t *testing.T) {
	tx := &fakeTx{}
	store := newSQLJobStoreWithBegin(func(context.Context) (jobTx, error) { return tx, nil })
	runAt := time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC)
	job := pipeline.Job{
		ID: "job-1", BatchID: "batch-1", RunAt: runAt, Status: "queued",
		Stages: []pipeline.Stage{pipeline.StageFetchBook, pipeline.StageResolveMetadata, pipeline.StageCreateBatch},
	}
	if err := store.CreateJob(context.Background(), job); err != nil { t.Fatal(err) }
	if !tx.commit { t.Fatal("transaction was not committed") }
	if len(tx.queries) != 4 { t.Fatalf("query count=%d", len(tx.queries)) }
	if tx.args[0][0] != "job-1" || tx.args[0][4] != "batch-1" { t.Fatalf("job args=%v", tx.args[0]) }
	for i, stage := range job.Stages {
		if tx.args[i+1][0] != "job-1" || tx.args[i+1][1] != string(stage) || tx.args[i+1][2] != i+1 {
			t.Fatalf("stage %d args=%v", i, tx.args[i+1])
		}
	}
}
