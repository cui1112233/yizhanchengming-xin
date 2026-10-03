package storage

import (
    "context"
    "database/sql"
    "testing"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeExec struct{ queries []string; args [][]any }
func (f *fakeExec) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
    f.queries = append(f.queries, query)
    f.args = append(f.args, append([]any(nil), args...))
    return fakeResult{rows:1}, nil
}

func TestSQLProgressStoreCompletesStageAndJob(t *testing.T) {
    exec := &fakeExec{}
    store := newSQLProgressStoreWithExec(exec.ExecContext)
    if err := store.CompleteStage(context.Background(), "job-1", pipeline.StageFetchBook); err != nil { t.Fatal(err) }
    if err := store.CompleteJob(context.Background(), "job-1"); err != nil { t.Fatal(err) }
    if len(exec.queries) != 2 { t.Fatalf("queries=%v", exec.queries) }
    if exec.args[0][0] != "job-1" || exec.args[0][1] != string(pipeline.StageFetchBook) { t.Fatalf("stage args=%v", exec.args[0]) }
    if exec.args[1][0] != "job-1" { t.Fatalf("job args=%v", exec.args[1]) }
}
