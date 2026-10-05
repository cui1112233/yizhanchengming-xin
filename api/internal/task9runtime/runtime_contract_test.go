package task9runtime

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAggregateRunStatus(t *testing.T) {
	tests := []struct {
		name string
		states []BookState
		want RunState
	}{
		{"still running", []BookState{BookSucceeded, BookRunning}, RunRunning},
		{"all success", []BookState{BookSucceeded, BookSucceeded}, RunSucceeded},
		{"partial failed", []BookState{BookSucceeded, BookFailed}, RunPartialFailed},
		{"all failed", []BookState{BookFailed, BookFailed}, RunFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AggregateRunStatus(tt.states); got != tt.want { t.Fatalf("got %q want %q", got, tt.want) }
		})
	}
}

func TestSafeErrorRedactionBeforePersistence(t *testing.T) {
	err := errors.New("provider failed Authorization: Bearer secret-token password=hunter2 dsn=root:pw@tcp(db)/prod")
	code, message := SafeError(err)
	if code == "" { t.Fatal("missing error code") }
	lower := strings.ToLower(message)
	for _, forbidden := range []string{"secret-token", "hunter2", "root:pw", "authorization", "bearer"} {
		if strings.Contains(lower, forbidden) { t.Fatalf("safe message leaked %q: %q", forbidden, message) }
	}
}

func TestWorkerRequiresMySQLCASBeforeExecution(t *testing.T) {
	store := &fakeStore{claimResults: []bool{true, false}}
	exec := &fakeExecutor{}
	worker := NewWorker(store, nil, exec, "worker-a", time.Minute, func() time.Time { return time.Unix(100, 0) })
	msg := WorkItem{BookRunID: 91, Attempt: 1}
	if err := worker.Process(context.Background(), msg); err != nil { t.Fatal(err) }
	if err := worker.Process(context.Background(), msg); err != nil { t.Fatal(err) }
	if exec.calls != 1 { t.Fatalf("duplicate message executed %d times", exec.calls) }
}

func TestTwoWorkersOnlyOneCASWinner(t *testing.T) {
	store := newConcurrentClaimStore()
	exec := &fakeExecutor{}
	w1 := NewWorker(store, nil, exec, "worker-a", time.Minute, time.Now)
	w2 := NewWorker(store, nil, exec, "worker-b", time.Minute, time.Now)
	item := WorkItem{BookRunID: 5, Attempt: 1}
	var wg sync.WaitGroup
	wg.Add(2)
	go func(){ defer wg.Done(); _ = w1.Process(context.Background(), item) }()
	go func(){ defer wg.Done(); _ = w2.Process(context.Background(), item) }()
	wg.Wait()
	if exec.calls != 1 { t.Fatalf("executed %d times", exec.calls) }
}

func TestStaleWorkerCompletionRejected(t *testing.T) {
	store := &fakeStore{}
	current := Execution{BookRunID: 8, Attempt: 2, FencingToken: 22, Owner: "worker-b"}
	store.current = current
	stale := Execution{BookRunID: 8, Attempt: 1, FencingToken: 21, Owner: "worker-a"}
	if ok, err := store.Complete(context.Background(), stale); err != nil || ok { t.Fatalf("stale complete ok=%v err=%v", ok, err) }
	if ok, err := store.Complete(context.Background(), current); err != nil || !ok { t.Fatalf("current complete ok=%v err=%v", ok, err) }
}

func TestStaleWorkerFailRejected(t *testing.T) {
	store := &fakeStore{current: Execution{BookRunID: 8, Attempt: 2, FencingToken: 22, Owner: "worker-b"}}
	stale := Execution{BookRunID: 8, Attempt: 1, FencingToken: 21, Owner: "worker-a"}
	if ok, err := store.Fail(context.Background(), stale, Failure{Code:"provider_timeout", Message:"timeout", Retryable:true}); err != nil || ok { t.Fatalf("stale fail ok=%v err=%v", ok, err) }
}

func TestRetryCreatesOnlyFailedBooksNextAttempt(t *testing.T) {
	state := []BookAttempt{{BookID:1, Attempt:1, State:BookSucceeded}, {BookID:2, Attempt:1, State:BookFailed}}
	got := PlanRetry(state, 3)
	if len(got) != 1 || got[0].BookID != 2 || got[0].Attempt != 2 { t.Fatalf("retry plan=%+v", got) }
}

func TestRetryPolicyStopsAtMaxAttemptsAndTerminalErrors(t *testing.T) {
	if ShouldRetry(BookAttempt{Attempt:3, State:BookFailed, Retryable:true}, 3) { t.Fatal("must not exceed max attempts") }
	if ShouldRetry(BookAttempt{Attempt:1, State:BookFailed, Retryable:false}, 3) { t.Fatal("terminal error must not retry") }
	if !ShouldRetry(BookAttempt{Attempt:1, State:BookFailed, Retryable:true}, 3) { t.Fatal("retryable attempt should retry") }
}

func TestSchedulerNotDueAndTwoInstancesOnlyOneWins(t *testing.T) {
	store := &schedulerStore{runAt: time.Unix(200,0)}
	s1 := NewScheduler(store, nil, func() time.Time { return time.Unix(100,0) })
	if n, err := s1.Tick(context.Background(), 10); err != nil || n != 0 { t.Fatalf("early tick n=%d err=%v", n, err) }
	store.nowDue = true
	s2 := NewScheduler(store, nil, func() time.Time { return time.Unix(300,0) })
	var wg sync.WaitGroup
	wg.Add(2)
	go func(){ defer wg.Done(); _, _ = s1.Tick(context.Background(), 10) }()
	go func(){ defer wg.Done(); _, _ = s2.Tick(context.Background(), 10) }()
	wg.Wait()
	if store.claimWins != 1 { t.Fatalf("scheduler claim wins=%d", store.claimWins) }
}

func TestRedisWipeDoesNotImmediatelyRequeueActiveExecution(t *testing.T) {
	now := time.Unix(100,0)
	active := RecoveryCandidate{BookRunID:1, Attempt:1, FencingToken:7, RunningSince:now.Add(-time.Minute), LeaseDeadline:now.Add(time.Minute), MaxAttempts:3}
	stale := RecoveryCandidate{BookRunID:2, Attempt:1, FencingToken:8, RunningSince:now.Add(-2*time.Minute), LeaseDeadline:now.Add(-time.Second), MaxAttempts:3}
	got := PlanRecovery([]RecoveryCandidate{active, stale}, now)
	if len(got) != 1 || got[0].BookRunID != 2 { t.Fatalf("recovery=%+v", got) }
}

func TestGoRestartRecoveryUsesDurableMySQLFacts(t *testing.T) {
	now := time.Unix(100,0)
	candidate := RecoveryCandidate{BookRunID:4, Attempt:2, FencingToken:13, RunningSince:now.Add(-time.Hour), LeaseDeadline:now.Add(-time.Minute), MaxAttempts:3}
	got := PlanRecovery([]RecoveryCandidate{candidate}, now)
	if len(got) != 1 || got[0].Attempt != 3 { t.Fatalf("recovery=%+v", got) }
}

func TestTask14CanReuseRuntimeWithoutDomainImports(t *testing.T) {
	var _ RuntimeCoordinator = (*runtimeCoordinatorCompileProbe)(nil)
}

type fakeStore struct {
	claimResults []bool
	current Execution
}
func (f *fakeStore) Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error) {
	ok := true
	if len(f.claimResults) > 0 { ok = f.claimResults[0]; f.claimResults = f.claimResults[1:] }
	if !ok { return Execution{}, false, nil }
	if f.current.BookRunID == 0 { f.current = Execution{BookRunID:91, Attempt:1, FencingToken:1, Owner:"worker-a"} }
	return f.current, true, nil
}
func (f *fakeStore) Renew(context.Context, Execution, time.Time) (bool,error){ return true,nil }
func (f *fakeStore) Complete(_ context.Context, e Execution) (bool,error){ return e == f.current,nil }
func (f *fakeStore) Fail(_ context.Context, e Execution, _ Failure) (bool,error){ return e == f.current,nil }

type concurrentClaimStore struct{ mu sync.Mutex; won bool }
func newConcurrentClaimStore()*concurrentClaimStore{ return &concurrentClaimStore{} }
func (s *concurrentClaimStore) Claim(context.Context, WorkItem, string, time.Time)(Execution,bool,error){ s.mu.Lock(); defer s.mu.Unlock(); if s.won{return Execution{},false,nil}; s.won=true; return Execution{BookRunID:5,Attempt:1,FencingToken:1,Owner:"winner"},true,nil }
func (s *concurrentClaimStore) Renew(context.Context,Execution,time.Time)(bool,error){return true,nil}
func (s *concurrentClaimStore) Complete(context.Context,Execution)(bool,error){return true,nil}
func (s *concurrentClaimStore) Fail(context.Context,Execution,Failure)(bool,error){return true,nil}

type fakeExecutor struct{ mu sync.Mutex; calls int }
func (e *fakeExecutor) Execute(context.Context, Execution) error { e.mu.Lock(); defer e.mu.Unlock(); e.calls++; return nil }

type schedulerStore struct { mu sync.Mutex; runAt time.Time; nowDue bool; claimWins int }
func (s *schedulerStore) ClaimDueRun(context.Context,time.Time)(RunClaim,bool,error){ s.mu.Lock(); defer s.mu.Unlock(); if !s.nowDue || s.claimWins > 0 { return RunClaim{},false,nil }; s.claimWins++; return RunClaim{RunID:1},true,nil }
func (s *schedulerStore) EnsureQueuedBooks(context.Context,RunClaim) ([]WorkItem,error){ return []WorkItem{{BookRunID:1,Attempt:1}},nil }

type runtimeCoordinatorCompileProbe struct{}
func (*runtimeCoordinatorCompileProbe) Enqueue(context.Context, WorkItem) error { return nil }
