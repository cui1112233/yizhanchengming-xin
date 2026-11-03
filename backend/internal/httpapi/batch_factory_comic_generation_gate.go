package httpapi

import (
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// The page status is the durable lease for a comic-image generation request.
// A provider call is allowed to run for up to 15 minutes; the lease is
// deliberately longer than that generation context so a still-running worker
// is never mistaken for an orphan and taken over by a second request.
const comicPageGenerationStaleAfter = 25 * time.Minute

// comicPageGenerationGate prevents duplicate requests for the same page from
// queueing multiple expensive provider jobs at the same time.
type comicPageGenerationGate struct {
	mu     sync.Mutex
	active map[string]struct{}
}

func newComicPageGenerationGate() *comicPageGenerationGate {
	return &comicPageGenerationGate{active: make(map[string]struct{})}
}

// comicPageGenerationTaskExists is the durable half of the duplicate guard.
// A completed image is deliberately not considered an active task: an explicit
// regeneration is allowed after the previous generation has finished.
func comicPageGenerationTaskExists(page batchfactoryv11.ComicPage) bool {
	return strings.EqualFold(strings.TrimSpace(page.Status), "generating")
}

func comicPageGenerationTaskIsStale(page batchfactoryv11.ComicPage, now time.Time) bool {
	if !comicPageGenerationTaskExists(page) || page.UpdatedAt.IsZero() {
		return false
	}
	return !now.Before(page.UpdatedAt.Add(comicPageGenerationStaleAfter))
}

func (g *comicPageGenerationGate) tryAcquire(key string) (func(), bool) {
	if g == nil {
		return nil, false
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.active[key]; exists {
		return nil, false
	}
	g.active[key] = struct{}{}
	var once sync.Once
	return func() {
		once.Do(func() {
			g.mu.Lock()
			delete(g.active, key)
			g.mu.Unlock()
		})
	}, true
}
