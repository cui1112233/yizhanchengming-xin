package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type BookRun = intake.BookRun

type Store interface {
	StartRun(context.Context, int64) (bool, error)
	EnsureBookRuns(context.Context, int64) error
	RecoverExpiredBookRuns(context.Context, int64, time.Time) (int64, error)
	ListPendingBookRuns(context.Context, int64, int) ([]intake.BookRun, error)
	ClaimBookRun(context.Context, int64, time.Time) (bool, error)
	CompleteBookRun(context.Context, int64) error
	FailBookRun(context.Context, int64, string) error
	FinalizeRun(context.Context, int64) error
}

type BookExecutor interface {
	ExecuteBook(context.Context, int64) error
}

type Clock func() time.Time

type Engine struct {
	store     Store
	executor  BookExecutor
	now       Clock
	leaseTTL  time.Duration
	batchSize int
}

func NewEngine(store Store, executor BookExecutor, now Clock) *Engine {
	if now == nil {
		now = time.Now
	}
	return &Engine{store: store, executor: executor, now: now, leaseTTL: 5 * time.Minute, batchSize: 1000}
}

func (e *Engine) ExecuteRun(ctx context.Context, runID int64) error {
	if e.store == nil || e.executor == nil {
		return fmt.Errorf("execution engine dependencies unavailable")
	}
	if _, err := e.store.StartRun(ctx, runID); err != nil {
		return err
	}
	if err := e.store.EnsureBookRuns(ctx, runID); err != nil {
		return err
	}
	now := e.now()
	if _, err := e.store.RecoverExpiredBookRuns(ctx, runID, now); err != nil {
		return err
	}
	items, err := e.store.ListPendingBookRuns(ctx, runID, e.batchSize)
	if err != nil {
		return err
	}
	for _, item := range items {
		claimed, err := e.store.ClaimBookRun(ctx, item.ID, e.now().Add(e.leaseTTL))
		if err != nil {
			return err
		}
		if !claimed {
			continue
		}
		if err := e.executor.ExecuteBook(ctx, item.BookID); err != nil {
			if persistErr := e.store.FailBookRun(ctx, item.ID, err.Error()); persistErr != nil {
				return persistErr
			}
			continue
		}
		if err := e.store.CompleteBookRun(ctx, item.ID); err != nil {
			return err
		}
	}
	return e.store.FinalizeRun(ctx, runID)
}
