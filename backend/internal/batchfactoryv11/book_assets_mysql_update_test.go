package batchfactoryv11

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLUpdateBookAssetPreservesComicSource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT 1 FROM batch_factory_v12_books WHERE id=\? AND batch_id=\? AND owner_username=\?`).
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(`UPDATE batch_factory_v12_book_assets SET name=\?,formal_name=\?,facts=\?,appearance=\?,relationships=\?,prompt=\?,image_model_id=\?,image_aspect_ratio=\?,revision=revision\+1,updated_at=\?`).
		WithArgs("阿月", "", "", "", "", "青衣少女，银簪", nil, nil, sqlmock.AnyArg(), "asset-1", "alice", "batch-1", "book-1", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT kind,name,COALESCE\(formal_name,''\),COALESCE\(facts,''\),COALESCE\(appearance,''\),COALESCE\(relationships,''\),prompt,COALESCE\(image_model_id,''\),COALESCE\(image_aspect_ratio,''\),source`).
		WithArgs("asset-1", "alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"kind", "name", "formal_name", "facts", "appearance", "relationships", "prompt", "image_model_id", "image_aspect_ratio", "source", "extraction_preset_id", "extraction_preset_version", "revision", "created_at", "updated_at"}).
			AddRow("character", "阿月", "", "", "", "", "青衣少女，银簪", "", "", "comic", "", int64(0), int64(2), now, now))
	mock.ExpectCommit()

	updated, err := NewMySQLStore(db).UpdateBookAsset(context.Background(), "alice", "batch-1", "book-1", "asset-1", UpdateBookAssetInput{
		Name: "阿月", Prompt: "青衣少女，银簪", ExpectedRevision: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Source != "comic" {
		t.Fatalf("updated source=%q, want comic", updated.Source)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
