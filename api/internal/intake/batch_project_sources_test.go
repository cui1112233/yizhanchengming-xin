package intake

import (
	"context"
	"reflect"
	"regexp"
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
	query := "SELECT bp.id, bp.intake_id, bp.name, COALESCE(GROUP_CONCAT(DISTINCT NULLIF(b.source, '') ORDER BY b.source SEPARATOR ','), ''), bp.created_at, bp.updated_at FROM batch_projects bp LEFT JOIN books b ON b.intake_id = bp.intake_id GROUP BY bp.id, bp.intake_id, bp.name, bp.created_at, bp.updated_at ORDER BY bp.id DESC LIMIT 100"
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "created_at", "updated_at"}).
			AddRow(52, 12, "混合书城批次", "点众,知乎", now, now))

	projects, err := store.ListBatchProjects(context.Background())
	if err != nil {
		t.Fatalf("ListBatchProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("projects = %+v", projects)
	}

	field := reflect.ValueOf(projects[0]).FieldByName("Sources")
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
