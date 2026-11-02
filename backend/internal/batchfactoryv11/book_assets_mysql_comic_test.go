package batchfactoryv11

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPersistComicExtractedBookAssetsDeletesOldImagesBeforeReplacingAssets(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM batch_factory_v12_books").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("DELETE i FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id").
		WithArgs("alice", "batch-1", "book-1", "batch-comic-assets", "vip-smart-character-derive").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM batch_factory_v12_book_assets").
		WithArgs("alice", "batch-1", "book-1", "batch-comic-assets", "vip-smart-character-derive").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery("SELECT id,kind,name").
		WithArgs("alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kind", "name", "formal_name", "facts", "appearance", "relationships", "prompt", "image_model_id", "image_aspect_ratio", "source", "extraction_preset_id", "extraction_preset_version", "revision", "created_at", "updated_at"}))
	mock.ExpectCommit()

	_, err = NewMySQLStore(db).PersistComicExtractedBookAssets(context.Background(), "alice", Book{
		ID: "book-1", BatchID: "batch-1",
	}, DirectorSnapshot{}, DirectorAssets{})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
