package batchfactoryv11

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreListBatchesReturnsLibrarySummariesWithoutFullSourceHydration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT b.id,r.title,COALESCE\\(r.source_intake_id,''\\),b.revision,b.created_at,b.updated_at").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "source_intake_id", "revision", "created_at", "updated_at"}).
			AddRow("batch-1", "批量一", "intake-1", int64(7), now, now))
	mock.ExpectQuery("SELECT b.batch_id,b.id,r.title,COALESCE\\(r.source_book_id,''\\),COALESCE\\(r.platform,''\\),b.revision").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"batch_id", "id", "title", "source_book_id", "platform", "revision"}).
			AddRow("batch-1", "book-1", "小说一", "source-1", "hei-yan", int64(3)))

	batches, err := NewMySQLStore(db).ListBatches(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 1 || batches[0].ID != "batch-1" || batches[0].Title != "批量一" || batches[0].Revision != 7 {
		t.Fatalf("batch summary = %#v", batches)
	}
	if len(batches[0].Books) != 1 {
		t.Fatalf("books = %#v", batches[0].Books)
	}
	book := batches[0].Books[0]
	if book.ID != "book-1" || book.BookID != "source-1" || book.Title != "小说一" || book.Platform != "hei-yan" || book.SourceText != "" || len(book.Videos) != 0 {
		t.Fatalf("book summary = %#v", book)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReadbackMySQLStoreListBatchesKeepsLibraryReadsSummaryOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT b.id,r.title,COALESCE\\(r.source_intake_id,''\\),b.revision,b.created_at,b.updated_at").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "source_intake_id", "revision", "created_at", "updated_at"}).
			AddRow("batch-1", "批量一", "intake-1", int64(7), now, now))
	mock.ExpectQuery("SELECT b.batch_id,b.id,r.title,COALESCE\\(r.source_book_id,''\\),COALESCE\\(r.platform,''\\),b.revision").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"batch_id", "id", "title", "source_book_id", "platform", "revision"}).
			AddRow("batch-1", "book-1", "小说一", "source-1", "hei-yan", int64(3)))

	batches, err := NewReadbackMySQLStore(db).ListBatches(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 1 || batches[0].SettingsState.Patch != nil || len(batches[0].Books) != 1 || batches[0].Books[0].SettingsState.Patch != nil {
		t.Fatalf("library summaries must not hydrate settings: %#v", batches)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReadbackMySQLStoreHydratesComicSourceSegments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 4, 16, 0, 0, 0, time.UTC)

	for _, scope := range []struct {
		kind ScopeKind
		id   string
	}{
		{kind: ScopeBatch, id: "batch-1"},
		{kind: ScopeBook, id: "book-1"},
	} {
		mock.ExpectQuery("SELECT patch_json FROM batch_factory_v12_settings_patches").
			WithArgs(string(scope.kind), scope.id, "alice").
			WillReturnRows(sqlmock.NewRows([]string{"patch_json"}))
	}
	mock.ExpectQuery("SELECT id,revision,status,hook_text,source_digest,created_at,approved_at FROM batch_factory_v12_hook_revisions").
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "status", "hook_text", "source_digest", "created_at", "approved_at"}))
	mock.ExpectQuery("SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions").
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "snapshot_id", "mode", "source_digest", "hook_revision_id", "output_json", "created_at"}))
	mock.ExpectQuery("SELECT 1 FROM batch_factory_v12_books").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT id,kind,name").
		WithArgs("alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kind", "name", "formal_name", "facts", "appearance", "relationships", "prompt", "image_model_id", "image_aspect_ratio", "source", "extraction_preset_id", "extraction_preset_version", "revision", "created_at", "updated_at"}))
	mock.ExpectQuery("SELECT patch_json FROM batch_factory_v12_settings_patches").
		WithArgs(string(ScopeVideo), "video-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"patch_json"}).AddRow(`{"comicPrompt":"第一页","comicImagePrompt":"第一页图片"}`))
	// 非 active 的历史槽位必须一并下发，并带上最近成功任务的可播地址，
	// 供媒体库展示历史成片、供自动化判断跳过。
	mock.ExpectQuery("SELECT v.id,v.ordinal,v.compatibility_state,COALESCE\\(t.id,''\\)").
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordinal", "compatibility_state", "task_id", "media_url", "provider", "status", "created_at"}).
			AddRow("video-old", 0, "orphaned", "task-old", "https://cdn.example/video-old.mp4", "doubao", "succeeded", now))

	digest := sourceDigest("第一段。")
	mock.ExpectQuery("SELECT COALESCE\\(NULLIF\\(d.content,''\\),r.source_text,''\\)").
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"source"}).AddRow("第一段。"))
	mock.ExpectQuery("SELECT s.video_id,v.ordinal,s.start_offset").
		WithArgs("alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"video_id", "ordinal", "start_offset", "end_offset", "source_text", "character_count", "boundary_type", "reason", "source_digest", "segmentation_version", "prompt_stale", "media_stale", "revision"}).
			AddRow("video-1", 0, 0, 12, "第一段。", 4, "sentence", "model", digest, int64(1), false, false, int64(1)))

	store := NewReadbackMySQLStore(db)
	batch := Batch{ID: "batch-1", Revision: 1, Books: []Book{{
		ID: "book-1", BatchID: "batch-1", Revision: 1,
		Videos: []Video{{ID: "video-1", BatchID: "batch-1", BookID: "book-1", CompatibilityState: "active", Revision: 1}},
	}}}
	got, err := store.hydrateSettingsState(context.Background(), "alice", batch)
	if err != nil {
		t.Fatal(err)
	}
	segment := got.Books[0].Videos[0].ComicSourceSegment
	if segment == nil || segment.SourceText != "第一段。" || segment.VideoID != "video-1" {
		t.Fatalf("comic source segment = %#v", segment)
	}
	if len(got.Books[0].VideoHistory) != 1 || got.Books[0].VideoHistory[0].ID != "video-old" || got.Books[0].VideoHistory[0].Ordinal != 0 {
		t.Fatalf("video history = %#v", got.Books[0].VideoHistory)
	}
	history := got.Books[0].VideoHistory[0]
	if history.MediaURL != "https://cdn.example/video-old.mp4" || history.ProductionTaskID != "task-old" || history.Provider != "doubao" || !history.Succeeded {
		t.Fatalf("history media fields = %#v, want playable orphaned slot", history)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
