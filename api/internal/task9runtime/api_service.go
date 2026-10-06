package task9runtime

import (
	"context"
	"errors"
)

var ErrRuntimeQueueUnavailable = errors.New("task9 runtime queue unavailable")

type RetryStore interface {
	RetryBookRun(context.Context, int64) (WorkItem, bool, error)
	ProjectIDForBookRun(context.Context, int64) (int64, error)
}

type RetryService struct {
	store RetryStore
	coordinator RuntimeCoordinator
}

func NewRetryService(store RetryStore, coordinator RuntimeCoordinator) *RetryService {
	return &RetryService{store: store, coordinator: coordinator}
}

func (s *RetryService) ProjectIDForBookRun(ctx context.Context, bookRunID int64) (int64, error) {
	if s == nil || s.store == nil { return 0, ErrRuntimeQueueUnavailable }
	return s.store.ProjectIDForBookRun(ctx, bookRunID)
}

func (s *RetryService) RetryBookRun(ctx context.Context, failedBookRunID int64) (WorkItem, bool, error) {
	if s == nil || s.store == nil || s.coordinator == nil { return WorkItem{}, false, ErrRuntimeQueueUnavailable }
	item, created, err := s.store.RetryBookRun(ctx, failedBookRunID)
	if err != nil { return WorkItem{}, false, err }
	if created {
		if err := s.coordinator.Enqueue(ctx, item); err != nil { return item, true, err }
	}
	return item, created, nil
}
