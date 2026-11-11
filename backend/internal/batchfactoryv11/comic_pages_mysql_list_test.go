package batchfactoryv11

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLListComicPagesLoadsImagesInOneBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=?")).
		WithArgs("book-1", "batch-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id,p.video_id,p.comic_prompt,p.image_prompt,COALESCE(p.model_id,''),p.aspect_ratio,p.status,COALESCE(p.last_error,''),COALESCE(p.snapshot_hash,''),COALESCE(p.reference_image_urls,JSON_ARRAY()),p.revision,p.created_at,p.updated_at")).
		WithArgs("alice", "batch-1", "book-1").WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "comic_prompt", "image_prompt", "model_id", "aspect_ratio", "status", "last_error", "snapshot_hash", "reference_image_urls", "revision", "created_at", "updated_at"}).
		AddRow("page-1", "video-1", "漫画页一", "图片提示词一", "model-a", "3:4", "generation_failed", "图片通道拒绝了本次请求\n上游原文：HTTP 400", "", "[]", 1, now, now).
		AddRow("page-2", "video-2", "漫画页二", "图片提示词二", "model-a", "3:4", "image_ready", "", "", "[]", 1, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.id,COALESCE(i.url,''),COALESCE(i.storage_ref,''),i.media_type,i.source,i.is_primary,i.revision,i.created_at")).
		WithArgs("alice", "batch-1", "book-1").WillReturnRows(sqlmock.NewRows([]string{"id", "url", "storage_ref", "media_type", "source", "is_primary", "revision", "created_at", "comic_page_id"}).
		AddRow("image-1", "", "images/one.png", "image/png", "provider", true, 1, now, "page-1").
		AddRow("image-2", "", "images/two.png", "image/png", "provider", true, 1, now, "page-2"))

	store := &MySQLStore{db: db}
	pages, err := store.ListComicPages(context.Background(), "alice", "batch-1", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || len(pages[0].Images) != 1 || len(pages[1].Images) != 1 {
		t.Fatalf("pages/images not assembled from batch result: %+v", pages)
	}
	if pages[0].Images[0].ID != "image-1" || pages[1].Images[0].ID != "image-2" {
		t.Fatalf("unexpected image mapping: first=%+v second=%+v", pages[0].Images, pages[1].Images)
	}
	if pages[0].GenerationError != "图片通道拒绝了本次请求\n上游原文：HTTP 400" || pages[1].GenerationError != "" {
		t.Fatalf("generationError not mapped from last_error: first=%q second=%q", pages[0].GenerationError, pages[1].GenerationError)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLListComicPagesLiteSkipsPromptPayload(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=?")).
		WithArgs("book-1", "batch-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	// 轻量列表只读取列表渲染需要的字段，不能把大段提示词和参考图 JSON 从 MySQL 搬出来。
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id,p.video_id,COALESCE(p.model_id,''),p.aspect_ratio,p.status,COALESCE(p.last_error,''),COALESCE(p.snapshot_hash,''),p.revision,p.created_at,p.updated_at")).
		WithArgs("alice", "batch-1", "book-1").WillReturnRows(sqlmock.NewRows([]string{"id", "video_id", "model_id", "aspect_ratio", "status", "last_error", "snapshot_hash", "revision", "created_at", "updated_at"}).
		AddRow("page-1", "video-1", "model-a", "3:4", "generation_failed", "提示词包含敏感内容", "", 1, now, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT i.id,COALESCE(i.url,''),COALESCE(i.storage_ref,''),i.media_type,i.source,i.is_primary,i.revision,i.created_at")).
		WithArgs("alice", "batch-1", "book-1").WillReturnRows(sqlmock.NewRows([]string{"id", "url", "storage_ref", "media_type", "source", "is_primary", "revision", "created_at", "comic_page_id"}).
		AddRow("image-1", "", "images/one.png", "image/png", "provider", true, 1, now, "page-1"))

	store := &MySQLStore{db: db}
	pages, err := store.ListComicPagesLite(context.Background(), "alice", "batch-1", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || len(pages[0].Images) != 1 {
		t.Fatalf("lite pages/images not assembled: %+v", pages)
	}
	if pages[0].ComicPrompt != "" || pages[0].ImagePrompt != "" || len(pages[0].ReferenceImageURLs) != 0 {
		t.Fatalf("lite list unexpectedly loaded prompt payload: %+v", pages[0])
	}
	if pages[0].GenerationError != "提示词包含敏感内容" {
		t.Fatalf("lite list must keep generationError: %+v", pages[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
