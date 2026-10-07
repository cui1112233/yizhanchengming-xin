package intake

import (
	"context"
	"regexp"
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
	const query = "SELECT bp.id, bp.intake_id, bp.name, COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|'), ''), COUNT(b.id), COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.gender), '') ORDER BY b.gender SEPARATOR '|'), ''), COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.style), '') ORDER BY b.style SEPARATOR '|'), ''), COALESCE((SELECT r.status FROM runs r WHERE r.batch_project_id = bp.id ORDER BY r.run_at DESC, r.id DESC LIMIT 1), ''), bp.created_at, bp.updated_at FROM batch_projects bp LEFT JOIN books b ON b.intake_id = bp.intake_id GROUP BY bp.id, bp.intake_id, bp.name, bp.created_at, bp.updated_at ORDER BY bp.updated_at DESC, bp.id DESC LIMIT 100"
	columns := []string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "created_at", "updated_at"}

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows(columns))

	first, err := store.ListBatchProjects(context.Background())
	if err != nil {
		t.Fatalf("first ListBatchProjects: %v", err)
	}
	if len(first) != 0 {
		t.Fatalf("first projects = %+v, want empty list", first)
	}

	now := time.Date(2026, 10, 5, 8, 30, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(88, 18, "刚创建的批次", "知乎", 1, "女频", "情感", RunStatusPending, now, now))

	second, err := store.ListBatchProjects(context.Background())
	if err != nil {
		t.Fatalf("second ListBatchProjects: %v", err)
	}
	if len(second) != 1 || second[0].ID != 88 || second[0].Name != "刚创建的批次" {
		t.Fatalf("second projects = %+v, want newly created project", second)
	}
	if second[0].BookCount != 1 || second[0].RunStatus != RunStatusPending {
		t.Fatalf("new project summary = %+v", second[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected a fresh MySQL query for each list call: %v", err)
	}
}
