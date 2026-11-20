package task

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLStoreListsTasksOnlyForAuthenticatedOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(listTasksSQL)).WithArgs("alice", 100).WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "task_type", "status", "progress", "input_json", "output_json", "error_code", "error_message", "created_at", "updated_at"}).AddRow("mct_1", "mcp_1", "alice", "material_probe", "succeeded", 1, `{}`, `{}`, nil, nil, now, now))
	items, err := NewSQLStore(db, time.Now).List(context.Background(), "alice", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Owner != "alice" || items[0].Status != StatusSucceeded {
		t.Fatalf("items=%+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStoreCancelNeverTouchesAnotherOwnerTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta(cancelTaskSQL)).WithArgs(StatusCanceled, now, "mct_1", "alice", StatusPending, StatusRunning).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := NewSQLStore(db, func() time.Time { return now }).Cancel(context.Background(), "alice", "mct_1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStoreCancelReturnsTypedConflictForFinishedTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(cancelTaskSQL)).WillReturnResult(sqlmock.NewResult(0, 0))
	err = NewSQLStore(db, time.Now).Cancel(context.Background(), "alice", "mct_done")
	if ErrorCode(err) != "TASK_NOT_CANCELABLE" {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
}
