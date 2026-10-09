package task9runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type BookState string
type RunState string

const (
	BookPending         BookState = "pending"
	BookQueued          BookState = "queued"
	BookRunning         BookState = "running"
	BookSucceeded       BookState = "succeeded"
	BookFailed          BookState = "failed"
	BookRetryableFailed BookState = "retryable_failed"

	RunPending       RunState = "pending"
	RunRunning       RunState = "running"
	RunSucceeded     RunState = "succeeded"
	RunPartialFailed RunState = "partial_failed"
	RunFailed        RunState = "failed"
)

var ErrStaleExecution = errors.New("task9 stale execution")
var ErrProjectArchived = errors.New("batch project is archived")

type WorkItem struct {
	BookRunID int64
	BookID    int64
	Attempt   int
}
type Execution struct {
	BookRunID    int64
	Attempt      int
	FencingToken uint64
	Owner        string
}
type Failure struct {
	Code      string
	Message   string
	Retryable bool
}

// SafeExecutionError is the cross-domain error contract accepted by the
// runtime. It lets executors expose an allowlisted outcome without importing
// their package or persisting the wrapped dependency error.
type SafeExecutionError interface {
	error
	SafeCode() string
	SafeMessage() string
	Retryable() bool
	Unwrap() error
}
type BookAttempt struct {
	BookRunID int64
	BookID    int64
	Attempt   int
	State     BookState
	Retryable bool
}
type RecoveryCandidate struct {
	BookRunID     int64
	BookID        int64
	Attempt       int
	FencingToken  uint64
	RunningSince  time.Time
	LeaseDeadline time.Time
	MaxAttempts   int
	Retryable     bool
}
type RunClaim struct{ RunID int64 }

type Store interface {
	Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error)
	Renew(context.Context, Execution, time.Time) (bool, error)
	Complete(context.Context, Execution) (bool, error)
	Fail(context.Context, Execution, Failure) (bool, error)
}
type Executor interface {
	Execute(context.Context, Execution) error
}
type RuntimeCoordinator interface {
	Enqueue(context.Context, WorkItem) error
}

type QueueCoordinator struct{ queue taskruntime.Queue }

func NewQueueCoordinator(q taskruntime.Queue) *QueueCoordinator { return &QueueCoordinator{queue: q} }
func taskKey(item WorkItem) string {
	return fmt.Sprintf("bookrun:%d:book:%d", item.BookRunID, item.BookID)
}
func decodeWorkItem(msg taskruntime.Message) (WorkItem, error) {
	var br, b int64
	if _, err := fmt.Sscanf(msg.TaskKey, "bookrun:%d:book:%d", &br, &b); err != nil {
		return WorkItem{}, taskruntime.ErrMalformedEnvelope
	}
	item := WorkItem{BookRunID: br, BookID: b, Attempt: msg.Attempt}
	if br <= 0 || b <= 0 || msg.Attempt <= 0 || taskKey(item) != msg.TaskKey {
		return WorkItem{}, taskruntime.ErrMalformedEnvelope
	}
	return item, nil
}
func (c *QueueCoordinator) Enqueue(ctx context.Context, item WorkItem) error {
	return c.queue.Enqueue(ctx, taskruntime.Message{TaskKey: taskKey(item), Attempt: item.Attempt})
}

type Worker struct {
	store    Store
	leases   taskruntime.LeaseStore
	executor Executor
	owner    string
	leaseTTL time.Duration
	now      func() time.Time
}

func NewWorker(store Store, leases taskruntime.LeaseStore, executor Executor, owner string, leaseTTL time.Duration, now func() time.Time) *Worker {
	if now == nil {
		now = time.Now
	}
	return &Worker{store: store, leases: leases, executor: executor, owner: owner, leaseTTL: leaseTTL, now: now}
}

func (w *Worker) Process(ctx context.Context, item WorkItem) error {
	execution, ok, err := w.store.Claim(ctx, item, w.owner, w.now().Add(w.leaseTTL))
	if err != nil || !ok {
		return err
	}
	if err := w.executor.Execute(ctx, execution); err != nil {
		if nonBusinessTermination(ctx, err) {
			return err
		}
		code, message := SafeError(err)
		durable, failErr := w.store.Fail(ctx, execution, Failure{Code: code, Message: message, Retryable: retryableError(err)})
		if failErr == nil && !durable {
			return ErrStaleExecution
		}
		return failErr
	}
	ok, err = w.store.Complete(ctx, execution)
	if err == nil && !ok {
		return ErrStaleExecution
	}
	return err
}

func nonBusinessTermination(ctx context.Context, err error) bool {
	return errors.Is(err, ErrStaleExecution) || errors.Is(err, taskruntime.ErrLeaseNotOwner) || errors.Is(err, context.Canceled) || (ctx != nil && ctx.Err() != nil && errors.Is(err, ctx.Err()))
}

type retryability interface{ Retryable() bool }

func retryableError(err error) bool {
	var r retryability
	if errors.As(err, &r) {
		return r.Retryable()
	}
	return true
}

func (w *Worker) RunOnce(ctx context.Context, q taskruntime.Queue, poll time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if reclaimer, ok := q.(taskruntime.ProcessingReclaimer); ok {
		if _, err := reclaimer.ReclaimExpired(ctx, w.now(), 100); err != nil {
			return err
		}
	}
	delivery, err := q.Claim(ctx, w.owner, poll)
	if err != nil {
		return err
	}
	item, err := decodeWorkItem(delivery.Message)
	if err != nil {
		if ackErr := q.Ack(ctx, delivery); ackErr != nil {
			return ackErr
		}
		return taskruntime.ErrMalformedEnvelope
	}
	execution, ok, err := w.store.Claim(ctx, item, w.owner, w.now().Add(w.leaseTTL))
	if err != nil {
		_ = q.Nack(ctx, delivery, 0)
		return err
	}
	if !ok {
		return q.Ack(ctx, delivery)
	}
	var lease taskruntime.Lease
	leased := false
	if w.leases != nil {
		lease, ok, err = w.leases.Claim(ctx, delivery.Message.TaskKey, w.owner, execution.FencingToken, w.leaseTTL)
		if err != nil {
			_ = q.Ack(ctx, delivery)
			return err
		}
		if !ok {
			_ = q.Ack(ctx, delivery)
			return nil
		}
		leased = true
	}
	execErr := w.executeWithHeartbeat(ctx, execution, lease, leased)
	var durable bool
	if execErr != nil && nonBusinessTermination(ctx, execErr) {
		settleCtx, settleCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer settleCancel()
		if leased {
			_, _ = w.leases.Release(settleCtx, lease)
		}
		if errors.Is(execErr, context.Canceled) || (ctx.Err() != nil && errors.Is(execErr, ctx.Err())) {
			_ = q.Nack(settleCtx, delivery, 0)
		} else {
			_ = q.Ack(settleCtx, delivery)
		}
		return execErr
	} else if execErr != nil {
		code, message := SafeError(execErr)
		durable, err = w.store.Fail(ctx, execution, Failure{Code: code, Message: message, Retryable: retryableError(execErr)})
	} else {
		durable, err = w.store.Complete(ctx, execution)
	}
	if leased {
		if _, relErr := w.leases.Release(ctx, lease); relErr != nil && !errors.Is(relErr, taskruntime.ErrLeaseNotOwner) && err == nil {
			err = relErr
		}
	}
	if err != nil {
		_ = q.Nack(ctx, delivery, 0)
		return err
	}
	if !durable {
		_ = q.Ack(ctx, delivery)
		return ErrStaleExecution
	}
	return q.Ack(ctx, delivery)
}

func (w *Worker) executeWithHeartbeat(ctx context.Context, e Execution, lease taskruntime.Lease, leased bool) error {
	if !leased || w.leases == nil || w.leaseTTL <= 0 {
		return w.executor.Execute(ctx, e)
	}
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	interval := w.leaseTTL / 3
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}
	done := make(chan struct{})
	hbErr := make(chan error, 1)
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-execCtx.Done():
				return
			case <-ticker.C:
				ok, err := w.leases.Renew(execCtx, lease, w.leaseTTL)
				if err != nil || !ok {
					if err == nil {
						err = taskruntime.ErrLeaseNotOwner
					}
					select {
					case hbErr <- err:
					default:
						{
						}
					}
					cancel()
					return
				}
				ok, err = w.store.Renew(execCtx, e, w.now().Add(w.leaseTTL))
				if err != nil || !ok {
					if err == nil {
						err = ErrStaleExecution
					}
					select {
					case hbErr <- err:
					default:
						{
						}
					}
					cancel()
					return
				}
			}
		}
	}()
	err := w.executor.Execute(execCtx, e)
	cancel()
	<-done
	select {
	case h := <-hbErr:
		if err == nil {
			return h
		}
	default:
	}
	return err
}

func (w *Worker) Run(ctx context.Context, q taskruntime.Queue, poll time.Duration) error {
	return w.RunSupervised(ctx, q, poll, nil)
}

// RetryEvent exposes only fixed error categories, never queue envelopes or raw
// dependency errors. The app can attach its logger through the observer.
type RetryEvent struct {
	Code    string
	Backoff time.Duration
}

// RunSupervised keeps the existing worker alive across dependency failures.
// Only cancellation of the supervising context shuts it down; an individual
// Redis/lease operation's timeout is retried. Backoff is capped and cancellable.
func (w *Worker) RunSupervised(ctx context.Context, q taskruntime.Queue, poll time.Duration, onRetry func(RetryEvent)) error {
	if poll <= 0 {
		poll = time.Second
	}
	backoff := 25 * time.Millisecond
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := w.RunOnce(ctx, q, poll)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err == nil {
			backoff = 25 * time.Millisecond
			continue
		}
		if onRetry != nil && !errors.Is(err, taskruntime.ErrQueueEmpty) {
			code := "runtime_operation_failed"
			switch {
			case errors.Is(err, taskruntime.ErrQueueUnavailable):
				code = "queue_unavailable"
			case errors.Is(err, taskruntime.ErrLeaseUnavailable):
				code = "lease_unavailable"
			case errors.Is(err, taskruntime.ErrMalformedEnvelope):
				code = "malformed_envelope"
			case errors.Is(err, ErrStaleExecution), errors.Is(err, taskruntime.ErrLeaseNotOwner):
				code = "stale_execution"
			case errors.Is(err, context.DeadlineExceeded):
				code = "operation_timeout"
			}
			onRetry(RetryEvent{Code: code, Backoff: backoff})
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if backoff < time.Second {
			backoff *= 2
			if backoff > time.Second {
				backoff = time.Second
			}
		}
	}
}

func AggregateRunStatus(states []BookState) RunState {
	if len(states) == 0 {
		return RunPending
	}
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
	if succeeded == len(states) {
		return RunSucceeded
	}
	if failed == len(states) {
		return RunFailed
	}
	if succeeded > 0 && failed > 0 {
		return RunPartialFailed
	}
	return RunRunning
}
func SafeError(err error) (string, string) {
	if err == nil {
		return "", ""
	}
	var safe SafeExecutionError
	if errors.As(err, &safe) {
		return safe.SafeCode(), safe.SafeMessage()
	}
	message := err.Error()
	lower := strings.ToLower(message)
	for _, marker := range []string{"authorization", "bearer ", "password", "cookie", "access_token", "refresh_token", "api key", "apikey", "secret", "dsn=", "@tcp("} {
		if strings.Contains(lower, marker) {
			return "internal_error", "internal provider error"
		}
	}
	if len(message) > 1024 {
		message = message[:1024]
	}
	return "execution_error", message
}
func PlanRetry(attempts []BookAttempt, maxAttempts int) []WorkItem {
	latest := map[int64]BookAttempt{}
	for _, a := range attempts {
		if prev, ok := latest[a.BookID]; !ok || a.Attempt > prev.Attempt {
			latest[a.BookID] = a
		}
	}
	out := make([]WorkItem, 0)
	for _, a := range latest {
		if a.State == BookFailed && a.Retryable && a.Attempt < maxAttempts {
			out = append(out, WorkItem{BookRunID: a.BookRunID, BookID: a.BookID, Attempt: a.Attempt + 1})
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
		if c.LeaseDeadline.IsZero() || !c.LeaseDeadline.Before(now) || c.Attempt >= c.MaxAttempts || !c.Retryable {
			continue
		}
		out = append(out, WorkItem{BookRunID: c.BookRunID, BookID: c.BookID, Attempt: c.Attempt + 1})
	}
	return out
}

type SchedulerStore interface {
	ClaimAndMaterializeDueRun(context.Context, time.Time) (RunClaim, []WorkItem, bool, error)
}
type Scheduler struct {
	store       SchedulerStore
	coordinator RuntimeCoordinator
	now         func() time.Time
}

func NewScheduler(store SchedulerStore, coordinator RuntimeCoordinator, now func() time.Time) *Scheduler {
	if now == nil {
		now = time.Now
	}
	return &Scheduler{store: store, coordinator: coordinator, now: now}
}
func (s *Scheduler) Tick(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	claimed := 0
	for i := 0; i < limit; i++ {
		_, items, ok, err := s.store.ClaimAndMaterializeDueRun(ctx, s.now())
		if err != nil {
			return claimed, err
		}
		if !ok {
			break
		}
		if s.coordinator != nil {
			for _, item := range items {
				if err := s.coordinator.Enqueue(ctx, item); err != nil {
					return claimed, err
				}
			}
		}
		claimed++
	}
	return claimed, nil
}
func (s *Scheduler) Run(ctx context.Context, interval time.Duration, limit int) error {
	if interval <= 0 {
		interval = time.Second
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := s.Tick(ctx, limit); err != nil {
			return err
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

type RecoveryStore interface {
	RecoverUnmaterializedRuns(context.Context, time.Time, int) (int, error)
	ListQueuedBookRuns(context.Context, int) ([]WorkItem, error)
	RecoverStaleBookRuns(context.Context, time.Time, int) ([]WorkItem, error)
}
type Recovery struct {
	store       RecoveryStore
	coordinator RuntimeCoordinator
}

func NewRecovery(store RecoveryStore, coordinator RuntimeCoordinator) *Recovery {
	return &Recovery{store: store, coordinator: coordinator}
}
func (r *Recovery) Rebuild(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	if _, err := r.store.RecoverUnmaterializedRuns(ctx, now, limit); err != nil {
		return 0, err
	}
	queued, err := r.store.ListQueuedBookRuns(ctx, limit)
	if err != nil {
		return 0, err
	}
	remaining := limit - len(queued)
	var recovered []WorkItem
	if remaining > 0 {
		recovered, err = r.store.RecoverStaleBookRuns(ctx, now, remaining)
		if err != nil {
			return 0, err
		}
	}
	seen := map[int64]struct{}{}
	count := 0
	for _, item := range append(queued, recovered...) {
		if _, ok := seen[item.BookRunID]; ok {
			continue
		}
		seen[item.BookRunID] = struct{}{}
		if r.coordinator != nil {
			if err := r.coordinator.Enqueue(ctx, item); err != nil {
				return count, err
			}
		}
		count++
	}
	return count, nil
}
