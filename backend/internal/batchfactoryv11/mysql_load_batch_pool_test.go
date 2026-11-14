package batchfactoryv11

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreGetBatchDoesNotHoldBookRowsWhileLoadingDrafts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT b.id,r.title,COALESCE\\(r.source_intake_id,''\\),b.revision,b.created_at,b.updated_at").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "source_intake_id", "revision", "created_at", "updated_at"}).
			AddRow("batch-1", "批量一", "intake-1", int64(1), now, now))
	mock.ExpectQuery("SELECT id FROM batch_factory_v12_books WHERE batch_id=\\? AND owner_username=\\? AND hidden_at IS NULL ORDER BY ordinal,id").
		WithArgs("batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("book-1"))
	mock.ExpectQuery("SELECT b.id,r.title,COALESCE\\(r.source_text,''\\)").
		WithArgs("batch-1", "alice", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "source_text", "source_book_id", "source_task_id", "platform",
			"txt_text", "txt_file_name", "source_metadata_json", "revision",
		}).AddRow("book-1", "小说一", "正文", "source-1", "", "qimao", "", "", []byte(`{}`), int64(1)))
	mock.ExpectQuery("SELECT content FROM batch_factory_v12_drafts").
		WithArgs("alice", "working-front:book-1", "batch-1").
		WillReturnRows(sqlmock.NewRows([]string{"content"}).AddRow("编辑后的正文"))
	mock.ExpectQuery("SELECT v.id,r.label,COALESCE\\(r.video_prompt,''\\)").
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "video_prompt", "visual_prompt", "duration_seconds", "ordinal", "compatibility_state", "revision"}))
	mock.ExpectQuery("SELECT v.id,COALESCE\\(r.label,''\\)").
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "ordinal", "compatibility_state", "revision"}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	batch, err := NewMySQLStore(db).GetBatch(ctx, "alice", "batch-1")
	if err != nil {
		t.Fatalf("GetBatch must complete with a one-connection pool: %v", err)
	}
	if len(batch.Books) != 1 || batch.Books[0].WorkingFrontContent != "编辑后的正文" {
		t.Fatalf("book draft = %#v", batch.Books)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
