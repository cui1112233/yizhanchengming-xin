package httpapi

import (
	"context"

	"qiantie/backend/internal/batchfactoryv11"
)

// comicPageLiteStore is optional so older/in-memory stores keep working. The
// MySQL runtime implements it and avoids reading prompt-sized columns for list
// and status requests.
type comicPageLiteStore interface {
	ListComicPagesLite(ctx context.Context, owner, batchID, bookID string) ([]batchfactoryv11.ComicPage, error)
}

func listComicPagesForView(ctx context.Context, store batchfactoryv11.ComicPageStore, owner, batchID, bookID string, lite bool) ([]batchfactoryv11.ComicPage, error) {
	if lite {
		if liteStore, ok := store.(comicPageLiteStore); ok {
			return liteStore.ListComicPagesLite(ctx, owner, batchID, bookID)
		}
	}
	return store.ListComicPages(ctx, owner, batchID, bookID)
}
