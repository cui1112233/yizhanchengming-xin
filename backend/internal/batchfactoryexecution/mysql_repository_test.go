package batchfactoryexecution

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newExecutionRepositoryMock(t *testing.T) (*MySQLRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewMySQLRepository(db)
	repo.now = func() time.Time { return time.Date(2026, 11, 22, 8, 0, 0, 0, time.UTC) }
	ids := []string{"execution-1", "execution-book-1", "outbox-1"}
	repo.newID = func(prefix string) string { id := ids[0]; ids = ids[1:]; return id }
	return repo, mock
}

func TestMySQLRepositoryEnsureCreatesExecutionAtomically(t *testing.T) {
	repo, mock := newExecutionRepositoryMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(selectExecutionByRequestForUpdateSQL)).
		WithArgs("alice", "batch-1", "request-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT eb.book_id").WithArgs("alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"book_id"}))
	mock.ExpectExec("INSERT INTO batch_executions").
		WithArgs("execution-1", "alice", "batch-1", "video", "published_confirmed", "request-1", "planned", nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO batch_execution_config_snapshots").
		WithArgs("execution-1", `{"mode":"video"}`, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO batch_execution_books").
		WithArgs("execution-book-1", "execution-1", "book-1", "Book One", "pending", nil, nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO batch_execution_outbox").
		WithArgs("outbox-1", "execution", "execution-1", "execution.ensure", sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, 0, nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := repo.Ensure(context.Background(), EnsureRecord{
		Owner: "alice", BatchID: "batch-1", Mode: "video", Target: "published_confirmed", ClientRequestID: "request-1",
		ConfigJSON: []byte(`{"mode":"video"}`), ConfigDigest: "digest", Books: []EnsureBook{{BookID: "book-1", Title: "Book One"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Created || result.Execution.ID != "execution-1" || len(result.QueuedBooks) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLRepositoryEnsureReusesSameClientRequest(t *testing.T) {
	repo, mock := newExecutionRepositoryMock(t)
	now := repo.now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(selectExecutionByRequestForUpdateSQL)).WithArgs("alice", "batch-1", "request-1").
		WillReturnRows(sqlmock.NewRows(executionColumns).AddRow("execution-existing", "alice", "batch-1", "video", "published_confirmed", "request-1", "running", nil, now, now))
	mock.ExpectQuery("SELECT book_id FROM batch_execution_books").WithArgs("execution-existing").
		WillReturnRows(sqlmock.NewRows([]string{"book_id"}).AddRow("book-1"))
	mock.ExpectCommit()

	result, err := repo.Ensure(context.Background(), EnsureRecord{Owner: "alice", BatchID: "batch-1", Mode: "video", Target: "published_confirmed", ClientRequestID: "request-1", Books: []EnsureBook{{BookID: "book-1", Title: "Book One"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Created || result.Execution.ID != "execution-existing" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLRepositoryEnsureRejectsInvalidBookBeforeWrite(t *testing.T) {
	repo, mock := newExecutionRepositoryMock(t)
	_, err := repo.Ensure(context.Background(), EnsureRecord{Owner: "alice", BatchID: "batch-1", ClientRequestID: "request-1", Books: []EnsureBook{{BookID: ""}}})
	if err == nil {
		t.Fatal("expected invalid book error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
