package httpapi

import (
	"encoding/json"
	"sync"

	cp "qiantie/backend/internal/collectionpublish"
)

type collectionPublishLockEntry struct {
	mu    sync.Mutex
	users int
}

type collectionPublishScopedLock struct {
	publisher *BatchFactoryN8ComicPublisher
	key       string
}

// Different books must not share a network-duration lock. All requests for the
// same immutable owner/book/kind/collection still serialize, including status
// readback and repeated submissions. Waiting users retain the entry until their
// upload completes, then the entry is removed instead of growing indefinitely.
func (p *BatchFactoryN8ComicPublisher) collectionPublishLock(r cp.PublishRequest) sync.Locker {
	key, _ := json.Marshal([]string{r.Owner, r.BatchID, r.BookID, string(r.Kind), r.CollectionID})
	return &collectionPublishScopedLock{publisher: p, key: string(key)}
}

func (l *collectionPublishScopedLock) Lock() {
	p := l.publisher
	p.mu.Lock()
	if p.collectionLocks == nil {
		p.collectionLocks = make(map[string]*collectionPublishLockEntry)
	}
	entry := p.collectionLocks[l.key]
	if entry == nil {
		entry = &collectionPublishLockEntry{}
		p.collectionLocks[l.key] = entry
	}
	entry.users++
	p.mu.Unlock()
	entry.mu.Lock()
}

func (l *collectionPublishScopedLock) Unlock() {
	p := l.publisher
	p.mu.Lock()
	defer p.mu.Unlock()
	entry := p.collectionLocks[l.key]
	entry.mu.Unlock()
	entry.users--
	if entry.users == 0 {
		delete(p.collectionLocks, l.key)
	}
}
