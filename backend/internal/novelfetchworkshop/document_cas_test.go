package novelfetchworkshop

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMemoryDocumentCASPreservesConcurrentStop(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	doc := Document{BookID: "book", Meta: map[string]any{"status": "queued"}}
	version, err := store.CompareAndSwapDocument(ctx, "alice", doc)
	if err != nil || version == "" {
		t.Fatal(version, err)
	}
	if _, err = store.CompareAndSwapDocument(ctx, "alice", doc); !errors.Is(err, ErrDocumentChanged) {
		t.Fatalf("duplicate insert: %v", err)
	}
	current, err := store.GetDocument(ctx, "alice", doc.BookID)
	if err != nil || current.UpdatedAt != version {
		t.Fatal(current.UpdatedAt, err)
	}
	current.Meta["status"] = "stopped"
	if err = store.PutDocument(ctx, "alice", current); err != nil {
		t.Fatal(err)
	}
	doc.UpdatedAt = version
	if _, err = store.CompareAndSwapDocument(ctx, "alice", doc); !errors.Is(err, ErrDocumentChanged) {
		t.Fatalf("stale compensation: %v", err)
	}
	current, err = store.GetDocument(ctx, "alice", doc.BookID)
	if err != nil || current.Meta["status"] != "stopped" {
		t.Fatal(current.Meta, err)
	}
}
func TestMySQLDocumentCASChecksOwnerBookAndVersion(t *testing.T) {
	for _, affected := range []int64{0, 1} {
		_, mock, store := newBodySQLMock(t)
		previous := time.Date(2026, 11, 23, 1, 0, 0, 123456000, time.UTC)
		doc := Document{BookID: "book", Meta: map[string]any{"status": "queued"}, UpdatedAt: previous.Format(time.RFC3339Nano)}
		mock.ExpectExec(regexp.QuoteMeta("UPDATE novel_fetch_workshop_documents SET document_json=CAST(? AS CHAR CHARACTER SET utf8mb4), updated_at=? WHERE owner_username=? AND book_id=? AND updated_at=?")).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", "book", previous).WillReturnResult(sqlmock.NewResult(0, affected))
		version, err := store.CompareAndSwapDocument(context.Background(), "alice", doc)
		if affected == 0 && !errors.Is(err, ErrDocumentChanged) || affected == 1 && (err != nil || version == "" || version == doc.UpdatedAt) {
			t.Fatalf("affected=%d version=%s err=%v", affected, version, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}
