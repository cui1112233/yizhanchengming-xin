package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

type completedComicQueueStore struct {
	*batchfactoryv11.MemoryStore
	pages   []batchfactoryv11.ComicPage
	readErr error
}

func (s completedComicQueueStore) ListComicPages(context.Context, string, string, string) ([]batchfactoryv11.ComicPage, error) {
	return s.pages, s.readErr
}

func TestCompletedComicBookSkipsGenerationQueue(t *testing.T) {
	ctx := context.Background()
	store := completedComicQueueStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
	book := batchfactoryv11.Book{ID: "book", Videos: []batchfactoryv11.Video{{ID: "video-1"}, {ID: "video-2"}}}
	store.pages = []batchfactoryv11.ComicPage{
		{VideoID: "video-1", Images: []batchfactoryv11.ComicPageImage{{ID: "image-1", IsPrimary: true}}},
		{VideoID: "video-2", Images: []batchfactoryv11.ComicPageImage{{ID: "image-2", IsPrimary: true}}},
	}
	if !comicBookGenerationIdle(ctx, store, "alice", "batch", nil, book) {
		t.Fatal("completed pages must not wait for an image-generation slot")
	}
	store.pages[1].Images[0].IsPrimary = false
	if comicBookGenerationIdle(ctx, store, "alice", "batch", nil, book) {
		t.Fatal("a candidate image must not bypass the generation queue")
	}
	store.pages[1].Images[0].IsPrimary = true
	store.readErr = errors.New("database unavailable")
	if comicBookGenerationIdle(ctx, store, "alice", "batch", nil, book) {
		t.Fatal("a failed readiness read must not bypass the queue")
	}
}

func TestComicQueueReadinessChecksEnabledAssets(t *testing.T) {
	ctx := context.Background()
	store := completedComicQueueStore{MemoryStore: batchfactoryv11.NewMemoryStore(), pages: []batchfactoryv11.ComicPage{{VideoID: "video", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}}}}
	// The book does not exist in the asset store. With assets enabled the read
	// must fail closed, while disabled assets require no asset reads at all.
	book := batchfactoryv11.Book{ID: "missing", Videos: []batchfactoryv11.Video{{ID: "video"}}}
	if comicBookGenerationIdle(ctx, store, "alice", "batch", batchfactoryv11.SettingsPatch{"assetImagesEnabled": json.RawMessage(`true`)}, book) {
		t.Fatal("enabled assets were ignored")
	}
	if !comicBookGenerationIdle(ctx, store, "alice", "batch", batchfactoryv11.SettingsPatch{"assetImagesEnabled": json.RawMessage(`false`)}, book) {
		t.Fatal("disabled assets blocked completed pages")
	}
}
