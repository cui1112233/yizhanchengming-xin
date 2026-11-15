package batchfactoryv11

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDeleteBookKeepsSiblingBook(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "first"}, {Title: "second"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteBook(ctx, "alice", batch.ID, batch.Books[0].ID); err != nil {
		t.Fatal(err)
	}
	remaining, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Books) != 1 || remaining.Books[0].ID != batch.Books[1].ID {
		t.Fatalf("remaining books = %#v", remaining.Books)
	}
}

func TestDeleteBatchRejectsOtherOwner(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "batch"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteBatch(ctx, "bob", batch.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if _, err := store.GetBatch(ctx, "alice", batch.ID); err != nil {
		t.Fatalf("alice batch unexpectedly changed: %v", err)
	}
}

// TestDeleteBookStmtsPlaceholderCountMatchesArgs guards against the bug where
// the drafts DELETE had 2 placeholders but was handed 3 arguments, which made
// every book/batch deletion fail with a 500 internal error.
func TestDeleteBookClearsComicCollectionReferencesBeforeVideoParents(t *testing.T) {
	collections, references, videos := -1, -1, -1
	for i, s := range deleteBookStmts {
		if strings.Contains(s.query, "DELETE FROM batch_factory_v12_comic_collections ") {
			collections = i
		}
		if strings.Contains(s.query, "DELETE i FROM batch_factory_v12_comic_collection_items i") {
			references = i
		}
		if strings.Contains(s.query, "DELETE FROM batch_factory_v12_video_records ") {
			videos = i
		}
	}
	if collections < 0 || references < 0 || videos < 0 || collections >= videos || references >= videos {
		t.Fatalf("restricting comic collection references must precede video records: collections=%d references=%d videos=%d", collections, references, videos)
	}
}

func TestDeleteBookStmtsPlaceholderCountMatchesArgs(t *testing.T) {
	for _, stmt := range deleteBookStmts {
		placeholders := strings.Count(stmt.query, "?")
		args := stmt.args("owner", "batch", "book")
		if placeholders != len(args) {
			t.Errorf("delete statement has %d placeholders but %d args:\n%s", placeholders, len(args), stmt.query)
		}
	}
}

func TestDeleteBookStmtsRemoveV12DependentsBeforeTheirParents(t *testing.T) {
	position := map[string]int{}
	for index, stmt := range deleteBookStmts {
		position[stmt.query] = index
	}
	requirePosition := func(query string) int {
		value, ok := position[query]
		if !ok {
			t.Fatalf("missing cleanup statement: %s", query)
		}
		return value
	}
	compilation := requirePosition(`DELETE FROM batch_factory_v12_video_compilations WHERE owner_username=? AND batch_id=? AND book_id=?`)
	timeline := requirePosition(`DELETE FROM batch_factory_v12_canonical_timelines WHERE owner_username=? AND batch_id=? AND book_id=?`)
	audio := requirePosition(`DELETE FROM batch_factory_v12_audio_measurements WHERE owner_username=? AND batch_id=? AND book_id=?`)
	director := requirePosition(`DELETE FROM batch_factory_v12_director_revisions WHERE owner_username=? AND batch_id=? AND book_id=?`)
	book := requirePosition(`DELETE FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=?`)
	if !(compilation < timeline && timeline < audio && audio < director && director < book) {
		t.Fatalf("invalid V12 cleanup order: compilation=%d timeline=%d audio=%d director=%d book=%d", compilation, timeline, audio, director, book)
	}
}

func TestLocalMergedArtifactStorageRefAcceptsOnlyTheExactBatchURL(t *testing.T) {
	storageRef, ok := localMergedArtifactStorageRef("batch_123", "/api/batch-factory/v12/batches/batch_123/merge-media/merge_AbC-123")
	if !ok {
		t.Fatal("local merge URL was not recognized")
	}
	if want := "merge_AbC-123.mp4"; storageRef != want {
		t.Fatalf("storageRef = %q, want %q", storageRef, want)
	}
}

func TestLocalMergedArtifactStorageRefRejectsRemoteAndOtherBatchURLs(t *testing.T) {
	for _, output := range []string{
		"https://tos.example/batch-merged/merge_AbC-123.mp4",
		"/api/batch-factory/v12/batches/other/merge-media/merge_AbC-123",
		"/api/batch-factory/v12/batches/batch_123/merge-media/../../sessions",
		"/api/batch-factory/v12/batches/batch_123/merge-media/local-merge_AbC-123",
	} {
		if storageRef, ok := localMergedArtifactStorageRef("batch_123", output); ok || storageRef != "" {
			t.Fatalf("output %q was incorrectly accepted as %q", output, storageRef)
		}
	}
}

func TestQueueLocalMergedArtifactPurgesQueuesOnlyLocalMergeURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	purgeAfter := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT COALESCE\\(book_id,''\\),COALESCE\\(output_url,''\\) FROM batch_factory_v12_merge_jobs").
		WithArgs("alice", "batch-1").
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "output_url"}).
			AddRow("book-1", "/api/batch-factory/v12/batches/batch-1/merge-media/merge_keep123").
			AddRow("book-2", "tos://private-bucket/batch-merged/remote.mp4"))
	mock.ExpectExec("INSERT IGNORE INTO batch_factory_v12_local_artifact_purges").
		WithArgs("alice", "batch-1", "book-1", "merge_keep123.mp4", purgeAfter).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := queueLocalMergedArtifactPurges(ctx, tx, "alice", "batch-1", "", purgeAfter); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
