package novelcreation

import (
	"context"
	"errors"
	"fmt"
	"qiantie/backend/internal/batchfactoryv11"
	"sync"
	"testing"
	"time"
)

type onceWriteFault struct {
	mu    sync.Mutex
	kind  string
	fired bool
}

func (f *onceWriteFault) fail(kind string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.fired && f.kind == kind {
		f.fired = true
		return true
	}
	return false
}

type faultJournalStore struct {
	Store
	fault *onceWriteFault
}

func (s faultJournalStore) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(faultJournalStore{tx, s.fault}) })
}
func (s faultJournalStore) PutExecution(ctx context.Context, e Execution) error {
	for _, call := range e.Calls {
		if (call.State == "completed" || call.State == "rejected") && s.fault.fail(call.State) {
			return errors.New("temporary journal write failure")
		}
	}
	return s.Store.PutExecution(ctx, e)
}
func (s faultJournalStore) AppendRunEvent(ctx context.Context, o string, e RunEvent) (RunEvent, error) {
	if (e.Type == "assistant.delta" && s.fault.fail("event")) || (e.Status == RunWriting && e.Type != "assistant.delta" && s.fault.fail("next_stage")) {
		return RunEvent{}, errors.New("temporary event write failure")
	}
	return s.Store.AppendRunEvent(ctx, o, e)
}

func TestWorkerFixR2JournalWriteFailurePausesUnknown(t *testing.T) {
	for _, kind := range []string{"completed", "rejected"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			provider := &queuedPipelineProvider{responses: []string{foundationJSON}}
			if kind == "rejected" {
				provider.errors = []error{&batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}}
			}
			f.service.pipeline = pipelineWithProvider(provider)
			f.service.store = faultJournalStore{f.store, &onceWriteFault{kind: kind}}
			r := queuedWorkerRun(t, f)
			_, _ = f.service.Run(context.Background(), "alice", r.ID)
			saved, _ := f.store.GetRun(context.Background(), "alice", r.ID)
			e, _, _ := f.store.GetExecution(context.Background(), "alice", r.ID)
			if saved.Status != RunPaused || e.State != "paused" || saved.ErrorCode != ErrResultUnknown {
				t.Fatalf("uncertain journal sealed: run=%s execution=%s code=%s", saved.Status, e.State, saved.ErrorCode)
			}
			if _, err := f.service.RetryRun(context.Background(), "alice", r.ID); !IsCode(err, ErrResultUnknown) {
				t.Fatalf("unknown retry=%v", err)
			}
			if len(provider.requests) != 1 {
				t.Fatal("uncertain request resent")
			}
		})
	}
}

func TestWorkerFixR3EventFailureReplaysSameRun(t *testing.T) {
	for _, kind := range []string{"event", "next_stage"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, draftJSON, draftJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
			f.service.pipeline = pipelineWithProvider(provider)
			f.service.store = faultJournalStore{f.store, &onceWriteFault{kind: kind}}
			r := queuedWorkerRun(t, f)
			if _, err := f.service.Run(context.Background(), "alice", r.ID); err == nil {
				t.Fatal("expected local event failure")
			}
			e, _, _ := f.store.GetExecution(context.Background(), "alice", r.ID)
			if e.State != "ready" {
				t.Fatalf("successful journal made unrecoverable: %s", e.State)
			}
			retry, err := f.service.RetryRun(context.Background(), "alice", r.ID)
			if err != nil || retry.ID != r.ID {
				t.Fatalf("local failure created another paid run: %s %v", retry.ID, err)
			}
			resumed := NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(provider))
			got, err := resumed.Run(context.Background(), "alice", r.ID)
			if err != nil || got.Status != RunSucceeded {
				t.Fatalf("recovery=%s %v", got.Status, err)
			}
			if len(provider.requests) != 9 {
				t.Fatalf("completed response resent: calls=%d", len(provider.requests))
			}
		})
	}
}

func TestWorkerFixR4BusyProjectCannotFillDispatchWindow(t *testing.T) {
	f := newServiceFixture(t, nil)
	active := queuedWorkerRun(t, f)
	if _, _, err := f.service.claim(context.Background(), "alice", active.ID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 70; i++ {
		if _, err := f.service.SendMessage(context.Background(), "alice", f.sendInput(fmt.Sprintf("busy-%d", i))); err != nil {
			t.Fatal(err)
		}
	}
	project, err := f.service.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "free", Genre: "test", ModelID: "model"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := f.service.CreateSession(context.Background(), "alice", CreateSessionInput{ProjectID: project.ID, Mode: "long_form_original"})
	if err != nil {
		t.Fatal(err)
	}
	input := f.sendInput("free")
	input.ProjectID, input.SessionID = project.ID, session.ID
	free, err := f.service.SendMessage(context.Background(), "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := f.store.ListDispatch(context.Background(), 64)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].RunID != free.ID {
		t.Fatalf("busy project occupied dispatch window: candidates=%d", len(entries))
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	StartWorker(ctx, f.service, 5*time.Millisecond)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		got, _ := f.store.GetRun(ctx, "alice", free.ID)
		if got.Status == RunSucceeded {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("idle project starved")
}

type outageSwitch struct {
	mu      sync.Mutex
	offline bool
	failed  bool
}
type outageAfterResponseStore struct {
	Store
	switcher *outageSwitch
}

func (s outageAfterResponseStore) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	s.switcher.mu.Lock()
	offline := s.switcher.offline
	s.switcher.mu.Unlock()
	if offline {
		return errors.New("database still unavailable")
	}
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(outageAfterResponseStore{tx, s.switcher}) })
}
func (s outageAfterResponseStore) PutExecution(ctx context.Context, e Execution) error {
	s.switcher.mu.Lock()
	fail := false
	for _, call := range e.Calls {
		if call.State == "completed" && !s.switcher.failed {
			fail = true
			s.switcher.failed = true
			s.switcher.offline = true
		}
	}
	s.switcher.mu.Unlock()
	if fail {
		return errors.New("database disappeared after upstream response")
	}
	return s.Store.PutExecution(ctx, e)
}
func TestWorkerFixR2DatabaseOutageLeavesRunScannable(t *testing.T) {
	f := newServiceFixture(t, nil)
	p := &queuedPipelineProvider{responses: []string{foundationJSON}}
	switcher := &outageSwitch{}
	f.service.pipeline = pipelineWithProvider(p)
	f.service.store = outageAfterResponseStore{f.store, switcher}
	r := queuedWorkerRun(t, f)
	if _, err := f.service.Run(context.Background(), "alice", r.ID); err == nil {
		t.Fatal("expected database outage")
	}
	e, _, _ := f.store.GetExecution(context.Background(), "alice", r.ID)
	saved, _ := f.store.GetRun(context.Background(), "alice", r.ID)
	if e.State != "leased" || isTerminalRunStatus(saved.Status) {
		t.Fatalf("unreadable journal falsely finalized: %s/%s", saved.Status, e.State)
	}
	switcher.mu.Lock()
	switcher.offline = false
	switcher.mu.Unlock()
	expireWorkerLease(t, f, RunLease{Owner: "alice", ProjectID: r.ProjectID, RunID: r.ID, Fence: e.Fence})
	pending, err := f.store.ListDispatch(context.Background(), 64)
	if err != nil || len(pending) != 1 || pending[0].RunID != r.ID {
		t.Fatalf("outage run disappeared from scan: %+v %v", pending, err)
	}
	restarted := NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(p))
	if _, err = restarted.Run(context.Background(), "alice", r.ID); !IsCode(err, ErrResultUnknown) {
		t.Fatalf("recovery guessed unsent: %v", err)
	}
	if len(p.requests) != 1 {
		t.Fatal("outage recovery repeated paid call")
	}
}
