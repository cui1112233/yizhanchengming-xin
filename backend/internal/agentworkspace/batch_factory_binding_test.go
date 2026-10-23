package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestValidateBatchFactoryBindingRejectsEmptyIDs(t *testing.T) {
	if err := ValidateBatchFactoryBinding(BatchFactoryBinding{Owner: "alice", ProjectID: "agent-1"}); err == nil {
		t.Fatal("missing batch ID was accepted")
	}
	if err := ValidateBatchFactoryBinding(BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: "agent-1"}); err != nil {
		t.Fatalf("valid binding rejected: %v", err)
	}
}

func TestNodeOriginNeverContainsMediaBytes(t *testing.T) {
	origin := BatchFactoryNodeOrigin{BatchID: "batch-1", BookID: "book-1", ResourceType: "book_asset", ResourceID: "asset-1"}
	if err := ValidateBatchFactoryNodeOrigin(origin); err != nil {
		t.Fatalf("safe node origin rejected: %v", err)
	}
}

func TestMemoryBatchFactoryBindingsAreOwnerScopedAndAppendOnly(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 23, 9, 0, 0, 123456789, time.UTC)
	binding := BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: "project-1", CreatedAt: now, UpdatedAt: now}

	created, err := store.BindBatchFactoryProject(ctx, binding)
	if err != nil {
		t.Fatal(err)
	}
	if created.CreatedAt.Nanosecond()%1000 != 0 || created.UpdatedAt.Nanosecond()%1000 != 0 {
		t.Fatalf("timestamps must use microsecond precision: %+v", created)
	}
	duplicate, err := store.BindBatchFactoryProject(ctx, binding)
	if err != nil || duplicate != created {
		t.Fatalf("duplicate=%+v err=%v, want original binding", duplicate, err)
	}
	if _, err := store.BindBatchFactoryProject(ctx, BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: "project-2"}); !errors.Is(err, ErrBatchFactoryBindingConflict) {
		t.Fatalf("changed batch binding err=%v, want conflict", err)
	}
	if _, err := store.BindBatchFactoryProject(ctx, BatchFactoryBinding{Owner: "alice", BatchID: "batch-2", ProjectID: "project-1"}); !errors.Is(err, ErrBatchFactoryBindingConflict) {
		t.Fatalf("changed project binding err=%v, want conflict", err)
	}
	if _, err := store.BatchFactoryBinding(ctx, "bob", "batch-1"); !errors.Is(err, ErrBatchFactoryBindingNotFound) {
		t.Fatalf("cross-owner binding err=%v", err)
	}

	origin := BatchFactoryNodeOrigin{ProjectID: "project-1", NodeID: "image-1", BatchID: "batch-1", BookID: "book-1", ResourceType: "book_asset", ResourceID: "asset-1", VersionID: "version-1"}
	if err := store.BindBatchFactoryNode(ctx, "alice", origin); err != nil {
		t.Fatal(err)
	}
	origins, err := store.BatchFactoryNodeOrigins(ctx, "alice", "project-1", "book-1")
	if err != nil || len(origins) != 1 || origins[0] != origin {
		t.Fatalf("origins=%+v err=%v", origins, err)
	}
	if err := store.RemoveBatchFactoryNodeBinding(ctx, "bob", "project-1", "image-1"); err != nil {
		t.Fatalf("cross-owner remove err=%v", err)
	}
	origins, err = store.BatchFactoryNodeOrigins(ctx, "alice", "project-1", "book-1")
	if err != nil || len(origins) != 1 {
		t.Fatalf("cross-owner remove changed origins=%+v err=%v", origins, err)
	}
	if err := store.RemoveBatchFactoryNodeBinding(ctx, "alice", "project-1", "image-1"); err != nil {
		t.Fatal(err)
	}
	origins, err = store.BatchFactoryNodeOrigins(ctx, "alice", "project-1", "book-1")
	if err != nil || len(origins) != 0 {
		t.Fatalf("origins after remove=%+v err=%v", origins, err)
	}
}
