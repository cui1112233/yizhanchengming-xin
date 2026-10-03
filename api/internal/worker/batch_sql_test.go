package worker

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type batchFakeResult struct{ rows int64 }
func (r batchFakeResult) LastInsertId() (int64, error) { return 0, nil }
func (r batchFakeResult) RowsAffected() (int64, error) { return r.rows, nil }

type batchFakeTx struct {
	queries []string
	args [][]any
	committed bool
}
func (t *batchFakeTx) ExecContext(_ context.Context, q string, args ...any) (sql.Result, error) {
	t.queries = append(t.queries, q)
	t.args = append(t.args, append([]any(nil), args...))
	return batchFakeResult{rows: 1}, nil
}
func (t *batchFakeTx) Commit() error { t.committed = true; return nil }
func (t *batchFakeTx) Rollback() error { return nil }

func TestSQLBatchCreatorCreatesBatchBooksAndLinksJobAtomically(t *testing.T) {
	tx := &batchFakeTx{}
	creator := newSQLBatchCreatorWithBegin(func(context.Context) (batchTx, error) { return tx, nil })
	batchID, err := creator.CreateFromIntake(context.Background(), "intake-1", "job-1")
	if err != nil { t.Fatal(err) }
	if batchID == "" { t.Fatal("batch id is empty") }
	if !tx.committed { t.Fatal("transaction not committed") }
	if len(tx.queries) != 3 { t.Fatalf("queries=%d", len(tx.queries)) }
	if !strings.Contains(tx.queries[0], "INSERT INTO batches") { t.Fatalf("q0=%s", tx.queries[0]) }
	if !strings.Contains(tx.queries[1], "INSERT INTO batch_books") || !strings.Contains(tx.queries[1], "FROM intake_books") { t.Fatalf("q1=%s", tx.queries[1]) }
	if !strings.Contains(tx.queries[2], "UPDATE pipeline_jobs SET batch_id") { t.Fatalf("q2=%s", tx.queries[2]) }
	if tx.args[0][0] != batchID || tx.args[0][1] != "intake-1" { t.Fatalf("batch args=%v", tx.args[0]) }
	if tx.args[1][0] != batchID || tx.args[1][1] != "intake-1" { t.Fatalf("books args=%v", tx.args[1]) }
	if tx.args[2][0] != batchID || tx.args[2][1] != "job-1" || tx.args[2][2] != "intake-1" { t.Fatalf("job args=%v", tx.args[2]) }
}
