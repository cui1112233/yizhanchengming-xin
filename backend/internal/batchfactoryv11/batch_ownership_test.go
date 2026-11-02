package batchfactoryv11

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLBatchOwnershipCheckIsScopedAndLightweight(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	query := `SELECT 1 FROM batch_factory_v12_batches WHERE id=\? AND owner_username=\?`

	mock.ExpectQuery(query).WithArgs("batch-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	if err := store.EnsureBatchOwnership(context.Background(), "alice", "batch-1"); err != nil {
		t.Fatalf("owned batch rejected: %v", err)
	}

	mock.ExpectQuery(query).WithArgs("batch-1", "mallory").WillReturnRows(sqlmock.NewRows([]string{"1"}))
	err = store.EnsureBatchOwnership(context.Background(), "mallory", "batch-1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner batch error = %v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
