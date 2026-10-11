package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestHydrateNovelFetchDocumentRestoresRawAndProcessedBodies(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if _, err := store.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{
		BookID: "book-1", BodyRef: novelfetchworkshop.BodyRef{VersionID: "original_raw"}, Content: "完整原文，未截断",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{
		BookID: "book-1", BodyRef: novelfetchworkshop.BodyRef{VersionID: "original"}, Content: "处理后的正文",
	}); err != nil {
		t.Fatal(err)
	}
	document, err := hydrateNovelFetchDocument(ctx, store, "tester", novelfetchworkshop.Document{BookID: "book-1"})
	if err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "完整原文，未截断" || document.Original != "处理后的正文" || document.ProcessedText != "处理后的正文" {
		t.Fatalf("hydrated document lost body content: %#v", document)
	}
	if document.OriginalRawChars != 8 || document.ProcessedChars != 6 {
		t.Fatalf("hydrated character counts are wrong: raw=%d processed=%d", document.OriginalRawChars, document.ProcessedChars)
	}
}

func TestPersistNovelFetchBodiesKeepsFullRawSeparate(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	document := novelfetchworkshop.Document{BookID: "book-2"}
	if err := persistNovelFetchBodies(ctx, store, "tester", &document, "完整原文", "截断处理文"); err != nil {
		t.Fatal(err)
	}
	if document.BodyRefs["original_raw"].CharCount != 4 || document.BodyRefs["original"].CharCount != 5 {
		t.Fatalf("body refs=%+v", document.BodyRefs)
	}
	raw, err := store.GetBody(ctx, "tester", "book-2", "original_raw")
	if err != nil || raw.Content != "完整原文" {
		t.Fatalf("raw body=%+v err=%v", raw, err)
	}
}
