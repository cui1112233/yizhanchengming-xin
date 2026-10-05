package intake

import (
	"context"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreListsBatchProjectSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 7, 30, 0, 0, time.UTC)
	query := "SELECT bp.id, bp.intake_id, bp.name, COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|'), ''), COUNT(b.id), COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.gender), '') ORDER BY b.gender SEPARATOR '|'), ''), COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.style), '') ORDER BY b.style SEPARATOR '|'), ''), COALESCE((SELECT r.status FROM runs r WHERE r.batch_project_id = bp.id ORDER BY r.run_at DESC, r.id DESC LIMIT 1), ''), bp.created_at, bp.updated_at FROM batch_projects bp LEFT JOIN books b ON b.intake_id = bp.intake_id GROUP BY bp.id, bp.intake_id, bp.name, bp.created_at, bp.updated_at ORDER BY bp.id DESC LIMIT 100"
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "created_at", "updated_at"}).
			AddRow(52, 12, "混合书城批次", "点众|知乎", 3, "女频|男频", "情感|悬疑", RunStatusRunning, now, now))

	projects, err := store.ListBatchProjects(context.Background())
	if err != nil {
		t.Fatalf("ListBatchProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("projects = %+v", projects)
	}
	got := projects[0]
	if !reflect.DeepEqual(got.Sources, []string{"点众", "知乎"}) {
		t.Fatalf("Sources = %#v", got.Sources)
	}
	if got.BookCount != 3 {
		t.Fatalf("BookCount = %d, want 3", got.BookCount)
	}
	if !reflect.DeepEqual(got.Genders, []string{"女频", "男频"}) {
		t.Fatalf("Genders = %#v", got.Genders)
	}
	if !reflect.DeepEqual(got.Styles, []string{"情感", "悬疑"}) {
		t.Fatalf("Styles = %#v", got.Styles)
	}
	if got.RunStatus != RunStatusRunning {
		t.Fatalf("RunStatus = %q, want %q", got.RunStatus, RunStatusRunning)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
