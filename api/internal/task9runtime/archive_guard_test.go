package task9runtime

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRetryBookRunLocksProjectAndRejectsArchivedState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 11, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status FROM book_runs WHERE id=? FOR UPDATE")).WithArgs(int64(41)).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id", "book_id", "attempt", "max_attempts", "retryable", "status"}).AddRow(71, 51, 21, 1, 3, true, "failed"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectRollback()

	_, created, err := store.RetryBookRun(context.Background(), 41)
	if created || !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("created=%v err=%v", created, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
