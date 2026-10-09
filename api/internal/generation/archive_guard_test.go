package generation

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateBookRunLocksProjectAndRejectsArchivedState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectRollback()

	_, err = store.CreateBookRun(context.Background(), BookRun{BatchProjectID: 51, BookID: 21, Status: StatusRunning})
	if !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("err=%v, want ErrProjectArchived", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRetryTransitionLocksProjectAndRejectsArchivedState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT run_id FROM book_runs WHERE id=?")).WithArgs(int64(71)).WillReturnRows(sqlmock.NewRows([]string{"run_id"}).AddRow(nil))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectRollback()
	_, err = store.UpdateBookRun(context.Background(), BookRun{ID: 71, BatchProjectID: 51, Status: StatusRunning})
	if !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("err=%v, want ErrProjectArchived", err)
	}
}
