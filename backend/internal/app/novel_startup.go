package app

import (
	"context"
	"qiantie/backend/internal/novelcreation"
	"sync"
	"time"
)

// Constructing services must not consume persisted tasks. Build calls Start
// only after configuration, all fallible setup, and the router are ready.
type novelCreationStartup struct {
	service *novelcreation.Service
	once    sync.Once
}

func (s *novelCreationStartup) Start(ctx context.Context) {
	s.once.Do(func() {
		novelcreation.StartWorker(ctx, s.service, time.Second)
		novelcreation.StartDocumentWorker(ctx, s.service)
		novelcreation.StartScheduler(ctx, s.service, s.service.Notify, time.Minute)
		novelcreation.StartBatchWorker(ctx, s.service, s.service.Notify, 5*time.Second)
	})
}
