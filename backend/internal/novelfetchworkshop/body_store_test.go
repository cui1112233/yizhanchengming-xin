package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestMemoryStoreDocumentRoundTripPreservesProcessedFields(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	want := Document{
		BookID:           "book-123",
		OriginalRaw:      "完整原文",
		ProcessedText:    "处理后",
		OriginalRawChars: 4,
		ProcessedChars:   3,
		MaxTxt:           3,
	}
	if err := store.PutDocument(ctx, "alice", want); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetDocument(ctx, "alice", want.BookID)
	if err != nil {
		t.Fatal(err)
	}
	if got.OriginalRaw != want.OriginalRaw || got.ProcessedText != want.ProcessedText || got.OriginalRawChars != want.OriginalRawChars || got.ProcessedChars != want.ProcessedChars || got.MaxTxt != want.MaxTxt {
		t.Fatalf("got = %+v, want = %+v", got, want)
	}
}

func TestMemoryStoreBodyLifecycle(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	ref, err := store.PutBody(ctx, "alice", BodyRecord{
		BookID: "123",
		BodyRef: BodyRef{
			VersionID: "ai3",
			State:     "ready",
		},
		Content: "第一版正文",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ref.VersionID != "ai3" || ref.Revision != 1 || ref.ContentHash == "" || ref.CharCount == 0 {
		t.Fatalf("unexpected ref: %+v", ref)
	}

	body, err := store.GetBody(ctx, "alice", "123", "ai3")
	if err != nil {
		t.Fatal(err)
	}
	if body.Content != "第一版正文" {
		t.Fatalf("content=%q", body.Content)
	}

	refs, err := store.ListBodyRefs(ctx, "alice", "123")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].VersionID != "ai3" {
		t.Fatalf("refs=%+v", refs)
	}

	ref, err = store.PutBody(ctx, "alice", BodyRecord{
		BookID: "123",
		BodyRef: BodyRef{
			VersionID: "ai3",
			State:     "ready",
		},
		Content: "第二版正文",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ref.Revision != 2 {
		t.Fatalf("revision=%d", ref.Revision)
	}

	deleted, err := store.DeleteBody(ctx, "alice", "123", "ai3")
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Fatal("expected body to be deleted")
	}
	if _, err := store.GetBody(ctx, "alice", "123", "ai3"); err != ErrNotFound {
		t.Fatalf("err=%v", err)
	}
}
