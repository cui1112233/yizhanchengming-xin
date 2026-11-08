package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

type trackingComicPageLiteStore struct {
	batchfactoryv11.ComicPageStore
	fullCalls int
	liteCalls int
}

func (s *trackingComicPageLiteStore) ListComicPages(context.Context, string, string, string) ([]batchfactoryv11.ComicPage, error) {
	s.fullCalls++
	return []batchfactoryv11.ComicPage{{ID: "full"}}, nil
}

func (s *trackingComicPageLiteStore) ListComicPagesLite(context.Context, string, string, string) ([]batchfactoryv11.ComicPage, error) {
	s.liteCalls++
	return []batchfactoryv11.ComicPage{{ID: "lite"}}, nil
}

func TestListComicPagesForViewUsesLiteReader(t *testing.T) {
	store := &trackingComicPageLiteStore{}
	pages, err := listComicPagesForView(context.Background(), store, "alice", "batch-1", "book-1", true)
	if err != nil {
		t.Fatal(err)
	}
	if store.liteCalls != 1 || store.fullCalls != 0 {
		t.Fatalf("lite view used wrong reader: lite=%d full=%d", store.liteCalls, store.fullCalls)
	}
	if len(pages) != 1 || pages[0].ID != "lite" {
		t.Fatalf("unexpected pages: %+v", pages)
	}
}

func TestListComicPagesForViewKeepsFullReaderForEditor(t *testing.T) {
	store := &trackingComicPageLiteStore{}
	pages, err := listComicPagesForView(context.Background(), store, "alice", "batch-1", "book-1", false)
	if err != nil {
		t.Fatal(err)
	}
	if store.fullCalls != 1 || store.liteCalls != 0 {
		t.Fatalf("full view used wrong reader: lite=%d full=%d", store.liteCalls, store.fullCalls)
	}
	if len(pages) != 1 || pages[0].ID != "full" {
		t.Fatalf("unexpected pages: %+v", pages)
	}
}
