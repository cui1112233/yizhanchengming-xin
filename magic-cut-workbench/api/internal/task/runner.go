package task

import (
	"context"
	"errors"
	"strings"
)

type Store interface {
	Save(context.Context, Task) error
	Recoverable(context.Context) ([]Task, error)
}

type Queue interface {
	Push(context.Context, string, string) error
}

type Runner struct {
	store  Store
	queue  Queue
	prefix string
}

func NewRunner(store Store, queue Queue, prefix string) *Runner {
	return &Runner{store: store, queue: queue, prefix: prefix}
}

func PendingQueueKey(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if !strings.HasPrefix(prefix, "magic_cut:dev:") {
		prefix = "magic_cut:dev:"
	}
	return strings.TrimSuffix(prefix, ":") + ":tasks:pending"
}

func (r *Runner) Enqueue(ctx context.Context, item Task) error {
	if r == nil || r.store == nil || r.queue == nil {
		return errors.New("task runner is not configured")
	}
	if item.Status == "" {
		item.Status = StatusPending
	}
	if err := r.store.Save(ctx, item); err != nil {
		return err
	}
	return r.queue.Push(ctx, PendingQueueKey(r.prefix), item.ID)
}

func (r *Runner) Recover(ctx context.Context) error {
	if r == nil || r.store == nil || r.queue == nil {
		return errors.New("task runner is not configured")
	}
	items, err := r.store.Recoverable(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := r.queue.Push(ctx, PendingQueueKey(r.prefix), item.ID); err != nil {
			return err
		}
	}
	return nil
}
