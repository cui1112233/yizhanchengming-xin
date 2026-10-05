package task9runtime

import (
	"context"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type BookState string
type RunState string

const (
	BookPending BookState = "pending"
	BookQueued BookState = "queued"
	BookRunning BookState = "running"
	BookSucceeded BookState = "succeeded"
	BookFailed BookState = "failed"
	BookRetryableFailed BookState = "retryable_failed"

	RunPending RunState = "pending"
	RunRunning RunState = "running"
	RunSucceeded RunState = "succeeded"
	RunPartialFailed RunState = "partial_failed"
	RunFailed RunState = "failed"
)

type WorkItem struct {
	BookRunID int64
	BookID int64
	Attempt int
}

type Execution struct {
	BookRunID int64
	Attempt int
	FencingToken uint64
	Owner string
}

type Failure struct {
	Code string
	Message string
	Retryable bool
}

type BookAttempt struct {
	BookRunID int64
	BookID int64
	Attempt int
	State BookState
	Retryable bool
}

type RecoveryCandidate struct {
	BookRunID int64
	BookID int64
	Attempt int
	FencingToken uint64
	RunningSince time.Time
	LeaseDeadline time.Time
	MaxAttempts int
	Retryable bool
}

type RunClaim struct { RunID int64 }

type Store interface {
	Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error)
	Renew(context.Context, Execution, time.Time) (bool, error)
	Complete(context.Context, Execution) (bool, error)
	Fail(context.Context, Execution, Failure) (bool, error)
}

type Executor interface { Execute(context.Context, Execution) error }

type RuntimeCoordinator interface { Enqueue(context.Context, WorkItem) error }

type Worker struct {
	store Store
	leases taskruntime.LeaseStore
	executor Executor
	owner string
	leaseTTL time.Duration
	now func() time.Time
}

func NewWorker(store Store, leases taskruntime.LeaseStore, executor Executor, owner string, leaseTTL time.Duration, now func() time.Time) *Worker {
	return &Worker{store: store, leases: leases, executor: executor, owner: owner, leaseTTL: leaseTTL, now: now}
}

func (w *Worker) Process(ctx context.Context, item WorkItem) error {
	execution, ok, err := w.store.Claim(ctx, item, w.owner, w.now().Add(w.leaseTTL))
	if err != nil || !ok { return err }
	if err := w.executor.Execute(ctx, execution); err != nil {
		code, message := SafeError(err)
		_, failErr := w.store.Fail(ctx, execution, Failure{Code:code, Message:message, Retryable:true})
		return failErr
	}
	_, err = w.store.Complete(ctx, execution)
	return err
}

func AggregateRunStatus(states []BookState) RunState {
	if len(states) == 0 { return RunPending }
	succeeded, failed := 0, 0
	for _, state := range states {
		switch state {
		case BookPending, BookQueued, BookRunning, BookRetryableFailed:
			return RunRunning
		case BookSucceeded:
			succeeded++
		case BookFailed:
			failed++
		}
	}
	if succeeded == len(states) { return RunSucceeded }
	if failed == len(states) { return RunFailed }
	if succeeded > 0 && failed > 0 { return RunPartialFailed }
	return RunRunning
}

func SafeError(err error) (string, string) {
	if err == nil { return "", "" }
	message := err.Error()
	lower := strings.ToLower(message)
	for _, marker := range []string{"authorization", "bearer ", "password", "cookie", "access_token", "refresh_token", "api key", "apikey", "secret", "dsn=", "@tcp("} {
		if strings.Contains(lower, marker) { return "internal_error", "internal provider error" }
	}
	if len(message) > 1024 { message = message[:1024] }
	return "execution_error", message
}

func PlanRetry(attempts []BookAttempt, maxAttempts int) []WorkItem {
	latest := map[int64]BookAttempt{}
	for _, a := range attempts {
		if prev, ok := latest[a.BookID]; !ok || a.Attempt > prev.Attempt { latest[a.BookID] = a }
	}
	out := make([]WorkItem, 0)
	for _, a := range latest {
		if a.State == BookFailed && a.Attempt < maxAttempts {
			out = append(out, WorkItem{BookRunID:a.BookRunID, BookID:a.BookID, Attempt:a.Attempt+1})
		}
	}
	return out
}

func ShouldRetry(a BookAttempt, maxAttempts int) bool {
	return a.State == BookFailed && a.Retryable && a.Attempt < maxAttempts
}

func PlanRecovery(candidates []RecoveryCandidate, now time.Time) []WorkItem {
	out := make([]WorkItem, 0)
	for _, c := range candidates {
		if c.LeaseDeadline.IsZero() || !c.LeaseDeadline.Before(now) || c.Attempt >= c.MaxAttempts { continue }
		out = append(out, WorkItem{BookRunID:c.BookRunID, BookID:c.BookID, Attempt:c.Attempt+1})
	}
	return out
}

type SchedulerStore interface {
	ClaimDueRun(context.Context, time.Time) (RunClaim, bool, error)
	EnsureQueuedBooks(context.Context, RunClaim) ([]WorkItem, error)
}

type Scheduler struct { store SchedulerStore; coordinator RuntimeCoordinator; now func() time.Time }
func NewScheduler(store SchedulerStore, coordinator RuntimeCoordinator, now func() time.Time) *Scheduler { return &Scheduler{store:store, coordinator:coordinator, now:now} }
func (s *Scheduler) Tick(ctx context.Context, limit int) (int, error) {
	if limit <= 0 { return 0, nil }
	claimed := 0
	for i := 0; i < limit; i++ {
		run, ok, err := s.store.ClaimDueRun(ctx, s.now())
		if err != nil { return claimed, err }
		if !ok { break }
		items, err := s.store.EnsureQueuedBooks(ctx, run)
		if err != nil { return claimed, err }
		if s.coordinator != nil {
			for _, item := range items { if err := s.coordinator.Enqueue(ctx, item); err != nil { return claimed, err } }
		}
		claimed++
	}
	return claimed, nil
}
