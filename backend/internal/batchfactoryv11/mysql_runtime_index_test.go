package batchfactoryv11

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetBatchRuntimeIndexReadsOnlyBatchAndBookIdentities(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT b.id,b.revision,b.created_at,b.updated_at FROM batch_factory_v12_batches b WHERE b.id=? AND b.owner_username=?`)).
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "created_at", "updated_at"}).AddRow("batch-1", int64(7), now, now))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,revision FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=? ORDER BY ordinal,id`)).
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision"}).AddRow("book-1", int64(2)).AddRow("book-2", int64(3)))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,ordinal FROM batch_factory_v12_videos WHERE batch_id=? AND book_id=? AND owner_username=? AND compatibility_state='active' ORDER BY ordinal,id`)).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordinal"}).AddRow("video-1", 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT v.id,v.ordinal,v.compatibility_state,COALESCE(t.id,''),COALESCE(t.media_url,'')`)).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordinal", "compatibility_state", "task_id", "media_url", "provider", "status", "created_at"}).
			AddRow("video-old-1", 0, "orphaned", "task-old-1", "https://cdn.example/video-old-1.mp4", "doubao", "succeeded", now))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,ordinal FROM batch_factory_v12_videos WHERE batch_id=? AND book_id=? AND owner_username=? AND compatibility_state='active' ORDER BY ordinal,id`)).
		WithArgs("batch-1", "book-2", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordinal"}).AddRow("video-2", 0).AddRow("video-3", 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT v.id,v.ordinal,v.compatibility_state,COALESCE(t.id,''),COALESCE(t.media_url,'')`)).
		WithArgs("batch-1", "book-2", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordinal", "compatibility_state", "task_id", "media_url", "provider", "status", "created_at"}))

	got, err := NewMySQLStore(db).GetBatchRuntimeIndex(context.Background(), "alice", "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "batch-1" || got.Revision != 7 || len(got.Books) != 2 || got.Books[1].ID != "book-2" || len(got.Books[0].Videos) != 1 || got.Books[0].Videos[0].ID != "video-1" || len(got.Books[1].Videos) != 2 {
		t.Fatalf("unexpected runtime index: %#v", got)
	}
	if len(got.Books[0].VideoHistory) != 1 || got.Books[0].VideoHistory[0].ID != "video-old-1" || got.Books[0].VideoHistory[0].Ordinal != 0 {
		t.Fatalf("runtime index must expose historical ordinal for the previous version: %#v", got.Books[0].VideoHistory)
	}
	if history := got.Books[0].VideoHistory[0]; history.MediaURL != "https://cdn.example/video-old-1.mp4" || history.ProductionTaskID != "task-old-1" || !history.Succeeded {
		t.Fatalf("runtime index history media fields = %#v", history)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
