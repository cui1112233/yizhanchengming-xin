package intake

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type runExecutionStore interface {
	StartRun(context.Context, int64) (bool, error)
	EnsureBookRuns(context.Context, int64) error
	RecoverExpiredBookRuns(context.Context, int64, time.Time) (int64, error)
	RetryBookRun(context.Context, int64) (bool, error)
}

func TestMySQLStoreProvidesRunExecutionMethods(t *testing.T) {
	store := &MySQLStore{}
	if _, ok := any(store).(runExecutionStore); !ok {
		t.Fatal("MySQLStore must provide run execution methods")
	}
}

func TestMySQLStoreStartsPendingRunOnlyOnce(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	query := "UPDATE runs SET status = ? WHERE id = ? AND status = ?"
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(RunStatusRunning, int64(71), RunStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	started, err := store.StartRun(context.Background(), 71)
	if err != nil || !started { t.Fatalf("started=%v err=%v", started, err) }

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(RunStatusRunning, int64(71), RunStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 0))
	started, err = store.StartRun(context.Background(), 71)
	if err != nil || started { t.Fatalf("duplicate started=%v err=%v", started, err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreEnsuresOneBookRunPerRunAndBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	query := "INSERT INTO book_runs (run_id, book_id, status, idempotency_key) SELECT r.id, b.id, 'pending', CONCAT('run:', r.id, ':book:', b.id) FROM runs r JOIN batch_projects bp ON bp.id = r.batch_project_id JOIN books b ON b.intake_id = bp.intake_id WHERE r.id = ? ON DUPLICATE KEY UPDATE id = id"
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(int64(71)).WillReturnResult(sqlmock.NewResult(0, 2))
	if err := store.EnsureBookRuns(context.Background(), 71); err != nil { t.Fatal(err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreRecoversExpiredBookRunsAndRetriesFailedBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

	recoverSQL := "UPDATE book_runs SET status = 'pending', lease_until = NULL WHERE run_id = ? AND status = 'running' AND lease_until IS NOT NULL AND lease_until < ?"
	mock.ExpectExec(regexp.QuoteMeta(recoverSQL)).WithArgs(int64(71), now).WillReturnResult(sqlmock.NewResult(0, 1))
	recovered, err := store.RecoverExpiredBookRuns(context.Background(), 71, now)
	if err != nil || recovered != 1 { t.Fatalf("recovered=%d err=%v", recovered, err) }

	retrySQL := "UPDATE book_runs SET status = 'pending', error_message = '', lease_until = NULL WHERE id = ? AND status = 'failed'"
	mock.ExpectExec(regexp.QuoteMeta(retrySQL)).WithArgs(int64(901)).WillReturnResult(sqlmock.NewResult(0, 1))
	retried, err := store.RetryBookRun(context.Background(), 901)
	if err != nil || !retried { t.Fatalf("retried=%v err=%v", retried, err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
