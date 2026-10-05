package execution

import (
	"context"
	"fmt"
	"time"
)

type DueRunStore interface {
	ListRunnableRunIDs(context.Context, time.Time, time.Time, int) ([]int64, error)
}

type Queue interface {
	Enqueue(context.Context, int64) error
	Dequeue(context.Context) (int64, error)
}

type Locker interface {
	Acquire(context.Context, string, time.Duration) (release func(context.Context) error, acquired bool, err error)
}

type Runner interface {
	ExecuteRun(context.Context, int64) error
}

type Scheduler struct {
	store       DueRunStore
	queue       Queue
	now         Clock
	staleAfter  time.Duration
	batchSize   int
}

func NewScheduler(store DueRunStore, queue Queue, now Clock) *Scheduler {
	if now == nil {
		now = time.Now
	}
	return &Scheduler{store: store, queue: queue, now: now, staleAfter: 10 * time.Minute, batchSize: 100}
}

func (s *Scheduler) Tick(ctx context.Context) error {
	if s.store == nil || s.queue == nil {
		return fmt.Errorf("scheduler dependencies unavailable")
	}
	now := s.now()
	ids, err := s.store.ListRunnableRunIDs(ctx, now, now.Add(-s.staleAfter), s.batchSize)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.queue.Enqueue(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

type Worker struct {
	queue   Queue
	locker  Locker
	runner  Runner
	lockTTL time.Duration
}

func NewWorker(queue Queue, locker Locker, runner Runner) *Worker {
	return &Worker{queue: queue, locker: locker, runner: runner, lockTTL: 10 * time.Minute}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	if w.queue == nil || w.locker == nil || w.runner == nil {
		return fmt.Errorf("worker dependencies unavailable")
	}
	runID, err := w.queue.Dequeue(ctx)
	if err != nil {
		return err
	}
	release, acquired, err := w.locker.Acquire(ctx, fmt.Sprintf("batch-factory:run:%d", runID), w.lockTTL)
	if err != nil {
		return err
	}
	if !acquired {
		return nil
	}
	defer func() { _ = release(context.Background()) }()
	return w.runner.ExecuteRun(ctx, runID)
}
