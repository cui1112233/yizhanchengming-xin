package intake

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreListIntakes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	query := "SELECT id, name, status, created_at, updated_at FROM intakes ORDER BY id DESC LIMIT 100"
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at"}).
			AddRow(12, "番茄", StatusCompleted, now, now).
			AddRow(11, "知乎", StatusPending, now, now),
	)

	rows, err := store.ListIntakes(context.Background())
	if err != nil {
		t.Fatalf("ListIntakes: %v", err)
	}
	if len(rows) != 2 || rows[0].ID != 12 || rows[1].ID != 11 {
		t.Fatalf("rows = %+v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
