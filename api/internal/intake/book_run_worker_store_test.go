package intake

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type bookRunWorkerStore interface {
	ListPendingBookRuns(context.Context, int64, int) ([]BookRun, error)
	ClaimBookRun(context.Context, int64, time.Time) (bool, error)
	CompleteBookRun(context.Context, int64) error
	FailBookRun(context.Context, int64, string) error
	FinalizeRun(context.Context, int64) error
}

func TestMySQLStoreProvidesBookRunWorkerOperations(t *testing.T) {
	store := &MySQLStore{}
	if _, ok := any(store).(bookRunWorkerStore); !ok {
		t.Fatal("MySQLStore must provide per-book worker operations")
	}
}

func TestMySQLStoreListsClaimsCompletesAndFailsBookRuns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	lease := now.Add(5 * time.Minute)

	listSQL := "SELECT id, run_id, book_id, status, attempt, error_message, idempotency_key, lease_until, created_at, updated_at FROM run_book_executions WHERE run_id = ? AND status = 'pending' ORDER BY id ASC LIMIT ?"
	mock.ExpectQuery(regexp.QuoteMeta(listSQL)).WithArgs(int64(71), 1000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "run_id", "book_id", "status", "attempt", "error_message", "idempotency_key", "lease_until", "created_at", "updated_at"}).
			AddRow(901, 71, 101, BookRunStatusPending, 0, "", "run:71:book:101", nil, now, now).
			AddRow(902, 71, 102, BookRunStatusPending, 0, "", "run:71:book:102", nil, now, now))
	items, err := store.ListPendingBookRuns(context.Background(), 71, 1000)
	if err != nil || len(items) != 2 || items[0].BookID != 101 { t.Fatalf("items=%+v err=%v", items, err) }

	claimSQL := "UPDATE run_book_executions SET status = 'running', attempt = attempt + 1, error_message = '', lease_until = ? WHERE id = ? AND status = 'pending'"
	mock.ExpectExec(regexp.QuoteMeta(claimSQL)).WithArgs(lease, int64(901)).WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err := store.ClaimBookRun(context.Background(), 901, lease)
	if err != nil || !claimed { t.Fatalf("claimed=%v err=%v", claimed, err) }

	completeSQL := "UPDATE run_book_executions SET status = 'completed', error_message = '', lease_until = NULL WHERE id = ? AND status = 'running'"
	mock.ExpectExec(regexp.QuoteMeta(completeSQL)).WithArgs(int64(901)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.CompleteBookRun(context.Background(), 901); err != nil { t.Fatal(err) }

	mock.ExpectExec(regexp.QuoteMeta(claimSQL)).WithArgs(lease, int64(902)).WillReturnResult(sqlmock.NewResult(0, 1))
	claimed, err = store.ClaimBookRun(context.Background(), 902, lease)
	if err != nil || !claimed { t.Fatalf("claimed=%v err=%v", claimed, err) }
	failSQL := "UPDATE run_book_executions SET status = 'failed', error_message = ?, lease_until = NULL WHERE id = ? AND status = 'running'"
	mock.ExpectExec(regexp.QuoteMeta(failSQL)).WithArgs("provider timeout", int64(902)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.FailBookRun(context.Background(), 902, "provider timeout"); err != nil { t.Fatal(err) }

	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreFinalizesRunFromPerBookStates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	query := "UPDATE runs r SET status = CASE WHEN EXISTS (SELECT 1 FROM run_book_executions rbe WHERE rbe.run_id = r.id AND rbe.status IN ('pending', 'running')) THEN 'running' WHEN EXISTS (SELECT 1 FROM run_book_executions rbe WHERE rbe.run_id = r.id AND rbe.status = 'failed') THEN 'failed' ELSE 'completed' END WHERE r.id = ?"
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(int64(71)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.FinalizeRun(context.Background(), 71); err != nil { t.Fatal(err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
