package novelfetchworkshop

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestMemoryStoreDeleteDocumentsMarksBodiesReleasable(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if err := store.PutDocument(ctx, "alice", Document{BookID: "123", Original: "原文"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(ctx, "alice", BodyRecord{
		BookID:  "123",
		BodyRef: BodyRef{VersionID: "ai1", State: "ready"},
		Content: "改写正文",
	}); err != nil {
		t.Fatal(err)
	}
	result, err := store.DeleteDocuments(ctx, "alice", []string{"123"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Deleted != 1 {
		t.Fatalf("deleted=%d result=%+v", result.Deleted, result)
	}
	body, err := store.GetBody(ctx, "alice", "123", "ai1")
	if err != nil {
		t.Fatalf("body should be retained for history, got %v", err)
	}
	if body.State != "releasable" || body.ExpiresAt == "" {
		t.Fatalf("body not marked releasable: %+v", body.BodyRef)
	}

	cleaned, err := store.CleanupExpiredBodies(ctx, time.Now().UTC().Add(DefaultBodyRetentionDays*24*time.Hour+time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if cleaned.Deleted != 1 {
		t.Fatalf("expired bodies not reclaimed: %+v", cleaned)
	}
	if _, err := store.GetBody(ctx, "alice", "123", "ai1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("body kept after expiry cleanup: %v", err)
	}
}

func TestMySQLStoreDeleteDocumentsMarksBodyRowsReleasable(t *testing.T) {
	_, mock, store := newBodySQLMock(t)
	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_fetch_workshop_bodies SET state='releasable', releasable_at=?, expires_at=?, updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=? AND state<>'releasable'`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", "123").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`)).
		WithArgs("alice", "123").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := store.DeleteDocuments(ctx, "alice", []string{"123"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Deleted != 1 {
		t.Fatalf("deleted=%d result=%+v", result.Deleted, result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
