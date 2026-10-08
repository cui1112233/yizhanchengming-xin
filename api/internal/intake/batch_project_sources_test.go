package intake

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreListsDistinctBookSourcesForBatchProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 6, 0, 0, 0, time.UTC)
	mock.ExpectQuery("^SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("^SELECT bp.id").WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}).
			AddRow(52, 12, "混合书城批次", "点众|知乎", 3, "女频|男频", "情感|悬疑", RunStatusRunning, 0, now, now, nil))

	page, err := store.ListBatchProjects(context.Background(), BatchProjectListQuery{Elevated: true, Archived: BatchProjectArchivedAll, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("ListBatchProjects: %v", err)
	}
	if len(page.Projects) != 1 {
		t.Fatalf("projects = %+v", page.Projects)
	}

	field := reflect.ValueOf(page.Projects[0]).FieldByName("Sources")
	if !field.IsValid() {
		t.Fatalf("BatchProject must expose Sources for multi-bookstore projects")
	}
	if got, want := field.Interface(), []string{"点众", "知乎"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Sources = %#v, want %#v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
