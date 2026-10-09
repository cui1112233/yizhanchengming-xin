package task9runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

func TestRunOnceSupervisorRetriesTransientQueueErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exec := &cancellingExecutor{cancel: cancel}
	q := &supervisorQueue{errors: []error{taskruntime.ErrQueueUnavailable, context.DeadlineExceeded}}
	w := NewWorker(&fakeStore{}, nil, exec, "worker", time.Second, time.Now)
	if err := w.Run(ctx, q, 10*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run stopped before context cancellation: %v", err)
	}
	if exec.calls != 1 {
		t.Fatalf("transient errors prevented execution: %d", exec.calls)
	}
}

func TestRunOnceSupervisorRetriesLeaseErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exec := &cancellingExecutor{cancel: cancel}
	q := &supervisorQueue{}
	w := NewWorker(&fakeStore{}, &transientLease{noopLease: noopLease{}}, exec, "worker", time.Second, time.Now)
	if err := w.Run(ctx, q, 10*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("lease error stopped Run: %v", err)
	}
	if exec.calls != 1 {
		t.Fatalf("lease recovery execution calls=%d", exec.calls)
	}
}

func TestRunOnceSupervisorMalformedDeliveryAckedOnce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exec := &cancellingExecutor{cancel: cancel}
	q := &supervisorQueue{malformed: true}
	w := NewWorker(&fakeStore{}, nil, exec, "worker", time.Second, time.Now)
	if err := w.Run(ctx, q, 10*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("malformed delivery stopped Run: %v", err)
	}
	if q.malformedAcks != 1 || exec.calls != 1 {
		t.Fatalf("malformed ACKs=%d execution=%d", q.malformedAcks, exec.calls)
	}
}

func TestRunOnceSupervisorReclaimsAndRetriesWithoutBusyLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Millisecond)
	defer cancel()
	q := &reclaimingQueue{supervisorQueue: supervisorQueue{alwaysEmpty: true}}
	w := NewWorker(&fakeStore{}, nil, &fakeExecutor{}, "worker", time.Second, time.Now)
	started := time.Now()
	if err := w.Run(ctx, q, 0); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run cancellation=%v", err)
	}
	if q.reclaims == 0 {
		t.Fatal("processing recovery was not attempted")
	}
	if q.claims > 5 || q.reclaims > 6 {
		t.Fatalf("busy loop claims=%d reclaims=%d", q.claims, q.reclaims)
	}
	if time.Since(started) > 500*time.Millisecond {
		t.Fatal("backoff delayed context cancellation")
	}
}

func TestRunOnceRejectsMalformedTaskKeysBeforeExecution(t *testing.T) {
	for _, msg := range []taskruntime.Message{
		{TaskKey: "bookrun:91:book:9:extra", Attempt: 1},
		{TaskKey: "bookrun:91:book:0", Attempt: 1},
		{TaskKey: "bookrun:91:book:9", Attempt: 0},
	} {
		t.Run(msg.TaskKey+"-"+fmt.Sprint(msg.Attempt), func(t *testing.T) {
			exec := &fakeExecutor{}
			q := &fixedDeliveryQueue{message: msg}
			w := NewWorker(&fakeStore{}, nil, exec, "worker", time.Second, time.Now)
			if err := w.RunOnce(context.Background(), q, time.Millisecond); !errors.Is(err, taskruntime.ErrMalformedEnvelope) {
				t.Fatalf("malformed error=%v", err)
			}
			if exec.calls != 0 || q.acks != 1 {
				t.Fatalf("malformed executed=%d ACKs=%d", exec.calls, q.acks)
			}
		})
	}
}

func TestRunOnceSupervisorSafeObserverAndBoundedBackoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	q := &supervisorQueue{errors: make([]error, 9)}
	for i := range q.errors {
		q.errors[i] = fmt.Errorf("%w: Authorization: Bearer secret-data", taskruntime.ErrQueueUnavailable)
	}
	w := NewWorker(&fakeStore{}, nil, &fakeExecutor{}, "worker", time.Second, time.Now)
	var events []RetryEvent
	err := w.RunSupervised(ctx, q, time.Millisecond, func(e RetryEvent) {
		events = append(events, e)
		if len(events) == 9 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) || len(events) != 9 {
		t.Fatalf("observer err=%v events=%d", err, len(events))
	}
	for i, e := range events {
		if e.Code != "queue_unavailable" || e.Backoff <= 0 || e.Backoff > time.Second {
			t.Fatalf("unsafe/unbounded event=%+v", e)
		}
		if i > 0 && e.Backoff < events[i-1].Backoff {
			t.Fatal("backoff reset across failures")
		}
	}
	if events[8].Backoff != time.Second {
		t.Fatalf("backoff did not reach its cap: %+v", events[8])
	}
}

type fixedDeliveryQueue struct {
	supervisorQueue
	message taskruntime.Message
	acks    int
}

func (q *fixedDeliveryQueue) Claim(context.Context, string, time.Duration) (taskruntime.Delivery, error) {
	return taskruntime.Delivery{Receipt: "malformed", Message: q.message}, nil
}
func (q *fixedDeliveryQueue) Ack(context.Context, taskruntime.Delivery) error { q.acks++; return nil }

type supervisorQueue struct {
	errors        []error
	malformed     bool
	malformedAcks int
	alwaysEmpty   bool
	claims        int
}

func (*supervisorQueue) Enqueue(context.Context, taskruntime.Message) error { return nil }
func (q *supervisorQueue) Claim(ctx context.Context, _ string, _ time.Duration) (taskruntime.Delivery, error) {
	q.claims++
	if len(q.errors) > 0 {
		err := q.errors[0]
		q.errors = q.errors[1:]
		return taskruntime.Delivery{}, err
	}
	if q.alwaysEmpty {
		return taskruntime.Delivery{}, taskruntime.ErrQueueEmpty
	}
	if q.malformed {
		q.malformed = false
		return taskruntime.Delivery{Receipt: "poison", Message: taskruntime.Message{TaskKey: "invalid", Attempt: 1}}, nil
	}
	if err := ctx.Err(); err != nil {
		return taskruntime.Delivery{}, err
	}
	return taskruntime.Delivery{Receipt: "valid", Message: taskruntime.Message{TaskKey: "bookrun:91:book:9", Attempt: 1}}, nil
}
func (q *supervisorQueue) Ack(_ context.Context, d taskruntime.Delivery) error {
	if d.Receipt == "poison" {
		q.malformedAcks++
	}
	return nil
}
func (*supervisorQueue) Nack(context.Context, taskruntime.Delivery, time.Duration) error { return nil }

type reclaimingQueue struct {
	supervisorQueue
	reclaims int
}

func (q *reclaimingQueue) ReclaimExpired(context.Context, time.Time, int) (int, error) {
	q.reclaims++
	if q.reclaims == 1 {
		return 0, taskruntime.ErrQueueUnavailable
	}
	return 0, nil
}

type cancellingExecutor struct {
	cancel context.CancelFunc
	calls  int
}

func (e *cancellingExecutor) Execute(context.Context, Execution) error {
	e.calls++
	e.cancel()
	return nil
}

type transientLease struct {
	noopLease
	claims int
}

func (l *transientLease) Claim(context.Context, string, string, uint64, time.Duration) (taskruntime.Lease, bool, error) {
	l.claims++
	if l.claims == 1 {
		return taskruntime.Lease{}, false, taskruntime.ErrLeaseUnavailable
	}
	return taskruntime.Lease{}, true, nil
}

func TestAggregateRunStatus(t *testing.T) {
	tests := []struct {
		name   string
		states []BookState
		want   RunState
	}{
		{"still running", []BookState{BookSucceeded, BookRunning}, RunRunning},
		{"all success", []BookState{BookSucceeded, BookSucceeded}, RunSucceeded},
		{"partial failed", []BookState{BookSucceeded, BookFailed}, RunPartialFailed},
		{"all failed", []BookState{BookFailed, BookFailed}, RunFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AggregateRunStatus(tt.states); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestSafeErrorRedactionBeforePersistence(t *testing.T) {
	err := errors.New("provider failed Authorization: Bearer secret-token password=hunter2 dsn=root:pw@tcp(db)/prod")
	code, message := SafeError(err)
	if code == "" {
		t.Fatal("missing error code")
	}
	lower := strings.ToLower(message)
	for _, forbidden := range []string{"secret-token", "hunter2", "root:pw", "authorization", "bearer"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("safe message leaked %q: %q", forbidden, message)
		}
	}
}

func TestWorkerRequiresMySQLCASBeforeExecution(t *testing.T) {
	store := &fakeStore{claimResults: []bool{true, false}}
	exec := &fakeExecutor{}
	worker := NewWorker(store, nil, exec, "worker-a", time.Minute, func() time.Time { return time.Unix(100, 0) })
	msg := WorkItem{BookRunID: 91, Attempt: 1}
	if err := worker.Process(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if err := worker.Process(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	if exec.calls != 1 {
		t.Fatalf("duplicate message executed %d times", exec.calls)
	}
}

func TestTwoWorkersOnlyOneCASWinner(t *testing.T) {
	store := newConcurrentClaimStore()
	exec := &fakeExecutor{}
	w1 := NewWorker(store, nil, exec, "worker-a", time.Minute, time.Now)
	w2 := NewWorker(store, nil, exec, "worker-b", time.Minute, time.Now)
	item := WorkItem{BookRunID: 5, Attempt: 1}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _ = w1.Process(context.Background(), item) }()
	go func() { defer wg.Done(); _ = w2.Process(context.Background(), item) }()
	wg.Wait()
	if exec.calls != 1 {
		t.Fatalf("executed %d times", exec.calls)
	}
}

func TestStaleWorkerCompletionRejected(t *testing.T) {
	store := &fakeStore{}
	current := Execution{BookRunID: 8, Attempt: 2, FencingToken: 22, Owner: "worker-b"}
	store.current = current
	stale := Execution{BookRunID: 8, Attempt: 1, FencingToken: 21, Owner: "worker-a"}
	if ok, err := store.Complete(context.Background(), stale); err != nil || ok {
		t.Fatalf("stale complete ok=%v err=%v", ok, err)
	}
	if ok, err := store.Complete(context.Background(), current); err != nil || !ok {
		t.Fatalf("current complete ok=%v err=%v", ok, err)
	}
}

func TestStaleWorkerFailRejected(t *testing.T) {
	store := &fakeStore{current: Execution{BookRunID: 8, Attempt: 2, FencingToken: 22, Owner: "worker-b"}}
	stale := Execution{BookRunID: 8, Attempt: 1, FencingToken: 21, Owner: "worker-a"}
	if ok, err := store.Fail(context.Background(), stale, Failure{Code: "provider_timeout", Message: "timeout", Retryable: true}); err != nil || ok {
		t.Fatalf("stale fail ok=%v err=%v", ok, err)
	}
}

func TestRetryCreatesOnlyFailedBooksNextAttempt(t *testing.T) {
	state := []BookAttempt{{BookID: 1, Attempt: 1, State: BookSucceeded}, {BookID: 2, Attempt: 1, State: BookFailed, Retryable: true}}
	got := PlanRetry(state, 3)
	if len(got) != 1 || got[0].BookID != 2 || got[0].Attempt != 2 {
		t.Fatalf("retry plan=%+v", got)
	}
}

func TestRetryPolicyStopsAtMaxAttemptsAndTerminalErrors(t *testing.T) {
	if ShouldRetry(BookAttempt{Attempt: 3, State: BookFailed, Retryable: true}, 3) {
		t.Fatal("must not exceed max attempts")
	}
	if ShouldRetry(BookAttempt{Attempt: 1, State: BookFailed, Retryable: false}, 3) {
		t.Fatal("terminal error must not retry")
	}
	if !ShouldRetry(BookAttempt{Attempt: 1, State: BookFailed, Retryable: true}, 3) {
		t.Fatal("retryable attempt should retry")
	}
}

func TestSchedulerNotDueAndTwoInstancesOnlyOneWins(t *testing.T) {
	store := &schedulerStore{runAt: time.Unix(200, 0)}
	s1 := NewScheduler(store, nil, func() time.Time { return time.Unix(100, 0) })
	if n, err := s1.Tick(context.Background(), 10); err != nil || n != 0 {
		t.Fatalf("early tick n=%d err=%v", n, err)
	}
	store.nowDue = true
	s2 := NewScheduler(store, nil, func() time.Time { return time.Unix(300, 0) })
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = s1.Tick(context.Background(), 10) }()
	go func() { defer wg.Done(); _, _ = s2.Tick(context.Background(), 10) }()
	wg.Wait()
	if store.claimWins != 1 {
		t.Fatalf("scheduler claim wins=%d", store.claimWins)
	}
}

func TestRedisWipeDoesNotImmediatelyRequeueActiveExecution(t *testing.T) {
	now := time.Unix(100, 0)
	active := RecoveryCandidate{BookRunID: 1, Attempt: 1, FencingToken: 7, RunningSince: now.Add(-time.Minute), LeaseDeadline: now.Add(time.Minute), MaxAttempts: 3, Retryable: true}
	stale := RecoveryCandidate{BookRunID: 2, Attempt: 1, FencingToken: 8, RunningSince: now.Add(-2 * time.Minute), LeaseDeadline: now.Add(-time.Second), MaxAttempts: 3, Retryable: true}
	got := PlanRecovery([]RecoveryCandidate{active, stale}, now)
	if len(got) != 1 || got[0].BookRunID != 2 {
		t.Fatalf("recovery=%+v", got)
	}
}

func TestGoRestartRecoveryUsesDurableMySQLFacts(t *testing.T) {
	now := time.Unix(100, 0)
	candidate := RecoveryCandidate{BookRunID: 4, Attempt: 2, FencingToken: 13, RunningSince: now.Add(-time.Hour), LeaseDeadline: now.Add(-time.Minute), MaxAttempts: 3, Retryable: true}
	got := PlanRecovery([]RecoveryCandidate{candidate}, now)
	if len(got) != 1 || got[0].Attempt != 3 {
		t.Fatalf("recovery=%+v", got)
	}
}

func TestTask14CanReuseRuntimeWithoutDomainImports(t *testing.T) {
	var _ RuntimeCoordinator = (*runtimeCoordinatorCompileProbe)(nil)
}

type fakeStore struct {
	claimResults []bool
	current      Execution
}

func (f *fakeStore) Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error) {
	ok := true
	if len(f.claimResults) > 0 {
		ok = f.claimResults[0]
		f.claimResults = f.claimResults[1:]
	}
	if !ok {
		return Execution{}, false, nil
	}
	if f.current.BookRunID == 0 {
		f.current = Execution{BookRunID: 91, Attempt: 1, FencingToken: 1, Owner: "worker-a"}
	}
	return f.current, true, nil
}
func (f *fakeStore) Renew(context.Context, Execution, time.Time) (bool, error) { return true, nil }
func (f *fakeStore) Complete(_ context.Context, e Execution) (bool, error) {
	return e == f.current, nil
}
func (f *fakeStore) Fail(_ context.Context, e Execution, _ Failure) (bool, error) {
	return e == f.current, nil
}

type concurrentClaimStore struct {
	mu  sync.Mutex
	won bool
}

func newConcurrentClaimStore() *concurrentClaimStore { return &concurrentClaimStore{} }
func (s *concurrentClaimStore) Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.won {
		return Execution{}, false, nil
	}
	s.won = true
	return Execution{BookRunID: 5, Attempt: 1, FencingToken: 1, Owner: "winner"}, true, nil
}
func (s *concurrentClaimStore) Renew(context.Context, Execution, time.Time) (bool, error) {
	return true, nil
}
func (s *concurrentClaimStore) Complete(context.Context, Execution) (bool, error) { return true, nil }
func (s *concurrentClaimStore) Fail(context.Context, Execution, Failure) (bool, error) {
	return true, nil
}

type fakeExecutor struct {
	mu    sync.Mutex
	calls int
}

func (e *fakeExecutor) Execute(context.Context, Execution) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls++
	return nil
}

type schedulerStore struct {
	mu        sync.Mutex
	runAt     time.Time
	nowDue    bool
	claimWins int
}

func (s *schedulerStore) ClaimAndMaterializeDueRun(context.Context, time.Time) (RunClaim, []WorkItem, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.nowDue || s.claimWins > 0 {
		return RunClaim{}, nil, false, nil
	}
	s.claimWins++
	return RunClaim{RunID: 1}, []WorkItem{{BookRunID: 1, BookID: 2, Attempt: 1}}, true, nil
}

func TestGenerationRecoveryRebuildRepairsUnmaterializedRunsFirst(t *testing.T) {
	store := &generationRecoveryProbe{}
	if _, err := NewRecovery(store, nil).Rebuild(context.Background(), time.Now(), 2); err != nil {
		t.Fatal(err)
	}
	if !store.repaired {
		t.Fatal("Rebuild did not repair valid zero-BookRun generation candidates")
	}
}

type generationRecoveryProbe struct{ repaired bool }

func (s *generationRecoveryProbe) RecoverUnmaterializedRuns(context.Context, time.Time, int) (int, error) {
	s.repaired = true
	return 1, nil
}
func (s *generationRecoveryProbe) ListQueuedBookRuns(context.Context, int) ([]WorkItem, error) {
	if !s.repaired {
		return nil, errors.New("queued lookup preceded repair")
	}
	return nil, nil
}
func (*generationRecoveryProbe) RecoverStaleBookRuns(context.Context, time.Time, int) ([]WorkItem, error) {
	return nil, nil
}

type runtimeCoordinatorCompileProbe struct{}

func (*runtimeCoordinatorCompileProbe) Enqueue(context.Context, WorkItem) error { return nil }
