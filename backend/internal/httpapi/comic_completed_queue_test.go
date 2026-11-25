package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type completedComicQueueStore struct {
	*batchfactoryv11.MemoryStore
	pages   []batchfactoryv11.ComicPage
	readErr error
}

func TestSmallComicBackfillDoesNotRequeueWholeBook(t *testing.T) {
	for _, tc := range []struct {
		total, missing int
		bypass         bool
	}{
		{28, 3, true}, {32, 3, true}, {28, 4, false}, {10, 2, true}, {10, 3, false}, {3, 3, false},
	} {
		t.Run(fmt.Sprintf("%d-of-%d", tc.missing, tc.total), func(t *testing.T) {
			store := completedComicQueueStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
			book := batchfactoryv11.Book{ID: "book"}
			for i := 0; i < tc.total; i++ {
				id := fmt.Sprintf("video-%d", i)
				book.Videos = append(book.Videos, batchfactoryv11.Video{ID: id})
				if i >= tc.missing {
					store.pages = append(store.pages, batchfactoryv11.ComicPage{VideoID: id, Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}})
				}
			}
			if got := comicBookCanBypassGenerationSlot(context.Background(), store, "alice", "batch", nil, book); got != tc.bypass {
				t.Fatalf("backfill bypass=%v want %v", got, tc.bypass)
			}
		})
	}
}

func TestComicBackfillDoesNotBypassLivePageTask(t *testing.T) {
	store := completedComicQueueStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
	book := batchfactoryv11.Book{ID: "book"}
	for i := 0; i < 28; i++ {
		id := fmt.Sprintf("video-%d", i)
		book.Videos = append(book.Videos, batchfactoryv11.Video{ID: id})
		page := batchfactoryv11.ComicPage{VideoID: id, Status: "generating", UpdatedAt: time.Now().UTC()}
		if i >= 3 {
			page.Status = "ready"
			page.Images = []batchfactoryv11.ComicPageImage{{IsPrimary: true}}
		}
		store.pages = append(store.pages, page)
	}
	if comicBookCanBypassGenerationSlot(context.Background(), store, "alice", "batch", nil, book) {
		t.Fatal("live page tasks must not be submitted again as a small backfill")
	}
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
	if !comicBookCanBypassGenerationSlot(ctx, store, "alice", "batch", nil, book) {
		t.Fatal("completed pages must not wait for an image-generation slot")
	}
	store.pages[1].Images[0].IsPrimary = false
	if comicBookCanBypassGenerationSlot(ctx, store, "alice", "batch", nil, book) {
		t.Fatal("a candidate image must not bypass the generation queue")
	}
	store.pages[1].Images[0].IsPrimary = true
	store.readErr = errors.New("database unavailable")
	if comicBookCanBypassGenerationSlot(ctx, store, "alice", "batch", nil, book) {
		t.Fatal("a failed readiness read must not bypass the queue")
	}
}

func TestComicQueueReadinessChecksEnabledAssets(t *testing.T) {
	ctx := context.Background()
	store := completedComicQueueStore{MemoryStore: batchfactoryv11.NewMemoryStore(), pages: []batchfactoryv11.ComicPage{{VideoID: "video", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}}}}
	// The book does not exist in the asset store. With assets enabled the read
	// must fail closed, while disabled assets require no asset reads at all.
	book := batchfactoryv11.Book{ID: "missing", Videos: []batchfactoryv11.Video{{ID: "video"}}}
	if comicBookCanBypassGenerationSlot(ctx, store, "alice", "batch", batchfactoryv11.SettingsPatch{"assetImagesEnabled": json.RawMessage(`true`)}, book) {
		t.Fatal("enabled assets were ignored")
	}
	if !comicBookCanBypassGenerationSlot(ctx, store, "alice", "batch", batchfactoryv11.SettingsPatch{"assetImagesEnabled": json.RawMessage(`false`)}, book) {
		t.Fatal("disabled assets blocked completed pages")
	}
}
