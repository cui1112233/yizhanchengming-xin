package batchfactorytext

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBatchPageCountsReadCurrentBodyPagesInOneQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT p.book_id, COUNT").WithArgs("owner", "batch").WillReturnRows(sqlmock.NewRows([]string{"book_id", "total", "ready"}).AddRow("book-a", 28, 27).AddRow("book-b", 8, 8))
	got, err := NewMySQLRepository(db).BatchPageCounts(context.Background(), "owner", "batch")
	if err != nil {
		t.Fatal(err)
	}
	if got["book-a"].Total != 28 || got["book-a"].Ready != 27 || got["book-b"].Ready != 8 {
		t.Fatalf("counts=%+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
