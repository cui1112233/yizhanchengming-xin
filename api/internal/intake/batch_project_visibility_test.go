package intake

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreBatchProjectListSeesProjectCreatedBetweenCalls(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	columns := []string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}

	mock.ExpectQuery("^SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("^SELECT bp.id").WithArgs(20, 0).WillReturnRows(sqlmock.NewRows(columns))

	first, err := store.ListBatchProjects(context.Background(), BatchProjectListQuery{Elevated: true, Archived: BatchProjectArchivedAll, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("first ListBatchProjects: %v", err)
	}
	if len(first.Projects) != 0 {
		t.Fatalf("first projects = %+v, want empty list", first)
	}

	now := time.Date(2026, 10, 5, 8, 30, 0, 0, time.UTC)
	mock.ExpectQuery("^SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("^SELECT bp.id").WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(88, 18, "刚创建的批次", "知乎", 1, "女频", "情感", RunStatusPending, 0, now, now, nil))

	second, err := store.ListBatchProjects(context.Background(), BatchProjectListQuery{Elevated: true, Archived: BatchProjectArchivedAll, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("second ListBatchProjects: %v", err)
	}
	if len(second.Projects) != 1 || second.Projects[0].ID != 88 || second.Projects[0].Name != "刚创建的批次" {
		t.Fatalf("second projects = %+v, want newly created project", second)
	}
	if second.Projects[0].BookCount != 1 || second.Projects[0].RunStatus != RunStatusPending {
		t.Fatalf("new project summary = %+v", second.Projects[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected a fresh MySQL query for each list call: %v", err)
	}
}
