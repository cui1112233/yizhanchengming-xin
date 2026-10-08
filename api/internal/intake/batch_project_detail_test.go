package intake

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type batchProjectDetailReader interface {
	GetBatchProject(context.Context, int64) (BatchProject, error)
}

func TestMySQLStoreProvidesBatchProjectDetailReader(t *testing.T) {
	var db *sql.DB
	store := NewMySQLStore(db)
	if _, ok := any(store).(batchProjectDetailReader); !ok {
		t.Fatalf("MySQLStore must implement GetBatchProject for the Batch Factory detail workbench")
	}
}

func TestMySQLStoreGetsBatchProjectByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 8, 30, 0, 0, time.UTC)
	archivedAt := now.Add(time.Hour)
	query := "SELECT id, intake_id, name, created_at, updated_at, archived_at FROM batch_projects WHERE id = ?"
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(int64(51)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "created_at", "updated_at", "archived_at"}).
			AddRow(51, 11, "真实批次", now, now, archivedAt))

	project, err := store.GetBatchProject(context.Background(), 51)
	if err != nil {
		t.Fatalf("GetBatchProject: %v", err)
	}
	if project.ID != 51 || project.IntakeID != 11 || project.Name != "真实批次" {
		t.Fatalf("project = %+v", project)
	}
	if project.ArchivedAt == nil || !project.ArchivedAt.Equal(archivedAt) {
		t.Fatalf("archivedAt = %v, want %v", project.ArchivedAt, archivedAt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
