package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"sync"
	"testing"
	"time"
)

func TestWorkerAtomicDuplicateInputConflict(t *testing.T) {
	f := newServiceFixture(t, nil)
	input := f.sendInput("atomic-key")
	if _, err := f.service.SendMessage(context.Background(), "alice", input); err != nil {
		t.Fatal(err)
	}
	input.Content = "different request"
	if _, err := f.service.SendMessage(context.Background(), "alice", input); err == nil {
		t.Fatal("same idempotency key accepted different content")
	}
}

func queuedWorkerRun(t *testing.T, f serviceFixture) Run {
	t.Helper()
	r, e := f.service.SendMessage(context.Background(), "alice", f.sendInput("worker"))
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func expireWorkerLease(t *testing.T, f serviceFixture, l RunLease) {
	t.Helper()
	err := f.store.Atomic(context.Background(), l.Owner, l.ProjectID, func(tx Store) error {
		e, _, err := tx.GetExecution(context.Background(), l.Owner, l.RunID)
		if err != nil {
			return err
		}
		e.LeaseUntil = time.Now().Add(-time.Second)
		if err = tx.PutExecution(context.Background(), e); err != nil {
			return err
		}
		slot, err := tx.GetSlot(context.Background(), l.Owner, l.ProjectID)
		if err != nil {
			return err
		}
		slot.LeaseUntil = e.LeaseUntil
		return tx.PutSlot(context.Background(), l.Owner, l.ProjectID, slot)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorkerConcurrentClaimCallsOnce(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	runner := &stubPipelineRunner{run: func(ctx context.Context, input RunContext, sink EventSink) (ChapterBundle, error) {
		close(entered)
		<-release
		return serviceTestBundle(input, RunSucceeded), nil
	}}
	f := newServiceFixture(t, runner)
	r := queuedWorkerRun(t, f)
	done := make(chan error, 1)
	go func() { _, e := f.service.Run(context.Background(), "alice", r.ID); done <- e }()
	<-entered
	other := NewService(f.store, NewMemoryRuntime(), runner)
	if _, e := other.Run(context.Background(), "alice", r.ID); !IsCode(e, ErrProjectLocked) {
		t.Fatalf("second claim=%v", e)
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	if runner.callCount() != 1 {
		t.Fatal("duplicate upstream dispatch")
	}
}

func TestWorkerExpiredFenceCannotWrite(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	l, _, e := f.service.claim(context.Background(), "alice", r.ID)
	if e != nil {
		t.Fatal(e)
	}
	expireWorkerLease(t, f, l)
	next, _, e := f.service.claim(context.Background(), "alice", r.ID)
	if e != nil {
		t.Fatal(e)
	}
	if next.Fence <= l.Fence {
		t.Fatal("fence not advanced")
	}
	old := &guardedStore{Store: f.store, lease: l}
	if _, e := old.UpdateRun(context.Background(), "alice", RunUpdate{ID: r.ID, Status: RunSucceeded}); !IsCode(e, ErrLeaseLost) {
		t.Fatalf("expired update=%v", e)
	}
	if _, e := old.AppendRunEvent(context.Background(), "alice", RunEvent{RunID: r.ID}); !IsCode(e, ErrLeaseLost) {
		t.Fatalf("expired event=%v", e)
	}
	if e := old.SaveFoundation(context.Background(), "alice", Foundation{ProjectID: r.ProjectID, ID: "f", Snapshot: json.RawMessage(`{}`)}); !IsCode(e, ErrLeaseLost) {
		t.Fatalf("expired foundation=%v", e)
	}
	if e := old.SaveChapterBundle(context.Background(), "alice", serviceTestBundle(RunContext{Run: r, Project: f.project}, RunSucceeded)); !IsCode(e, ErrLeaseLost) {
		t.Fatalf("expired bundle=%v", e)
	}
}

func TestWorkerRecoveryUnknownPausesPreparedResumesAndOldActivePauses(t *testing.T) {
	for _, state := range []string{"may_have_sent", "prepared", "completed", "legacy-active", "legacy-queued"} {
		t.Run(state, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			r := queuedWorkerRun(t, f)
			e, _, _ := f.store.GetExecution(context.Background(), "alice", r.ID)
			if state == "legacy-active" || state == "legacy-queued" {
				delete(f.store.executions, memoryKey("alice", r.ID))
				if state == "legacy-active" {
					_, _ = f.store.UpdateRun(context.Background(), "alice", RunUpdate{ID: r.ID, Status: RunWriting})
				}
			} else {
				e.Calls["draft"] = CallRecord{State: state}
				_ = f.store.PutExecution(context.Background(), e)
			}
			_, _, err := f.service.claim(context.Background(), "alice", r.ID)
			if state == "may_have_sent" || state == "legacy-active" {
				if !IsCode(err, ErrResultUnknown) {
					t.Fatalf("unknown sent recovery=%v", err)
				}
				got, _ := f.store.GetRun(context.Background(), "alice", r.ID)
				if got.Status != RunPaused {
					t.Fatalf("status=%s", got.Status)
				}
				if _, err = f.service.RetryRun(context.Background(), "alice", r.ID); !IsCode(err, ErrResultUnknown) {
					t.Fatalf("retry may charge again: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

type workerProvider struct {
	calls    int
	response string
	err      error
}

func (p *workerProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	return p.response, p.err
}
func TestWorkerJournalReplaysRawAndPausesUnknown(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	l, _, _ := f.service.claim(context.Background(), "alice", r.ID)
	ctx := context.WithValue(context.Background(), journalKey{}, &callJournal{f.store, l})
	p := &workerProvider{response: `{"content":"saved before parse"}`}
	request := batchfactoryv11.TextCompletionRequest{UserPrompt: "draft"}
	got, err := journalComplete(ctx, "draft", p, request)
	if err != nil {
		t.Fatal(err)
	}
	e, _, _ := f.store.GetExecution(ctx, "alice", r.ID)
	if e.Calls["draft"].Response != got || e.Calls["draft"].State != "completed" {
		t.Fatal("raw result not persisted")
	}
	expireWorkerLease(t, f, l)
	l, _, err = f.service.claim(context.Background(), "alice", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	ctx = context.WithValue(context.Background(), journalKey{}, &callJournal{f.store, l})
	if replay, err := journalComplete(ctx, "draft", p, request); err != nil || replay != got || p.calls != 1 {
		t.Fatalf("replay=%s count=%d err=%v", replay, p.calls, err)
	}
	p.err = errors.New("connection lost")
	if _, err = journalComplete(ctx, "audit:0", p, request); !IsCode(err, ErrResultUnknown) {
		t.Fatal(err)
	}
	if _, err = journalComplete(ctx, "audit:0", p, request); !IsCode(err, ErrResultUnknown) || p.calls != 2 {
		t.Fatalf("unknown resent count=%d err=%v", p.calls, err)
	}
}

func TestWorkerCancelImmediatelyStopsProviderAndPreventsCommit(t *testing.T) {
	entered := make(chan struct{})
	stopped := make(chan struct{})
	runner := &stubPipelineRunner{run: func(ctx context.Context, in RunContext, sink EventSink) (ChapterBundle, error) {
		close(entered)
		<-ctx.Done()
		close(stopped)
		return serviceTestBundle(in, RunSucceeded), nil
	}}
	f := newServiceFixture(t, runner)
	r := queuedWorkerRun(t, f)
	done := make(chan struct{})
	go func() { _, _ = f.service.Run(context.Background(), "alice", r.ID); close(done) }()
	<-entered
	if _, err := f.service.CancelRun(context.Background(), "alice", r.ID); err != nil {
		t.Fatal(err)
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("cancel did not interrupt context")
	}
	<-done
	chapters, _ := f.store.ListChapters(context.Background(), "alice", f.project.ID)
	if len(chapters) != 0 {
		t.Fatal("cancelled result committed")
	}
	got, _ := f.store.GetRun(context.Background(), "alice", r.ID)
	if got.Status != RunCancelled {
		t.Fatalf("cancel overwritten: %s", got.Status)
	}
}

func TestWorkerConfirmIsIdempotentAfterSourceSucceeded(t *testing.T) {
	f := newServiceFixture(t, checkpointRunner())
	session := newLongFormSession(t, f)
	input := f.sendInput("confirm-repeat")
	input.SessionID = session.ID
	r, err := f.service.SendMessage(context.Background(), "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Run(context.Background(), "alice", r.ID); err != nil {
		t.Fatal(err)
	}
	a, err := f.service.ConfirmRun(context.Background(), "alice", r.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := f.service.ConfirmRun(context.Background(), "alice", r.ID, nil)
	if err != nil || a.ID != b.ID {
		t.Fatalf("confirm=%s/%s %v", a.ID, b.ID, err)
	}
	if _, err = f.service.ConfirmRun(context.Background(), "alice", r.ID, json.RawMessage(foundationJSON)); !IsCode(err, ErrIdempotencyConflict) {
		t.Fatalf("changed confirm=%v", err)
	}
}

func TestWorkerCancelReleasesProjectSlot(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	if _, _, err := f.service.claim(context.Background(), "alice", r.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.CancelRun(context.Background(), "alice", r.ID); err != nil {
		t.Fatal(err)
	}
	next, err := f.service.SendMessage(context.Background(), "alice", f.sendInput("after-cancel"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = f.service.claim(context.Background(), "alice", next.ID); err != nil {
		t.Fatalf("cancelled run still blocks project: %v", err)
	}
}

func TestWorkerMemoryBundlePersistsTerminalLikeSQL(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	if err := f.store.SaveChapterBundle(context.Background(), "alice", serviceTestBundle(RunContext{Run: r, Project: f.project}, RunSucceeded)); err != nil {
		t.Fatal(err)
	}
	got, _ := f.store.GetRun(context.Background(), "alice", r.ID)
	if got.Status != RunSucceeded || got.FinishedAt == nil {
		t.Fatal("memory bundle did not persist terminal run like SQL")
	}
}

type saveFailure struct {
	mu             sync.Mutex
	left, attempts int
}
type flakySaveStore struct {
	Store
	failure *saveFailure
}

func (s flakySaveStore) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, owner, project, func(tx Store) error { return fn(flakySaveStore{tx, s.failure}) })
}
func (s flakySaveStore) SaveChapterBundle(ctx context.Context, owner string, b ChapterBundle) error {
	s.failure.mu.Lock()
	s.failure.attempts++
	fail := s.failure.left > 0
	if fail {
		s.failure.left--
	}
	s.failure.mu.Unlock()
	if fail {
		return errors.New("temporary disk failure")
	}
	return s.Store.SaveChapterBundle(ctx, owner, b)
}
func TestWorkerSavePendingRetriesOnlySave(t *testing.T) {
	f := newServiceFixture(t, nil)
	failure := &saveFailure{left: 1}
	f.service.store = flakySaveStore{f.store, failure}
	r := queuedWorkerRun(t, f)
	if _, err := f.service.Run(context.Background(), "alice", r.ID); err == nil {
		t.Fatal("expected save failure")
	}
	e, _, _ := f.store.GetExecution(context.Background(), "alice", r.ID)
	if e.State != "save_pending" || e.Bundle == nil {
		t.Fatal("full completed bundle not retained")
	}
	snapshot, snapErr := f.service.GetSessionSnapshot(context.Background(), "alice", f.session.ID)
	if snapErr != nil || snapshot.ExecutionState != "save_pending" {
		t.Fatalf("save recovery invisible to browser: %s %v", snapshot.ExecutionState, snapErr)
	}
	retry, err := f.service.RetryRun(context.Background(), "alice", r.ID)
	if err != nil || retry.ID != r.ID {
		t.Fatalf("save retry created a paid run: %s %v", retry.ID, err)
	}
	restarted := NewService(f.store, NewMemoryRuntime(), f.runner)
	got, err := restarted.Run(context.Background(), "alice", r.ID)
	if err != nil || got.Status != RunSucceeded {
		t.Fatalf("save retry=%s %v", got.Status, err)
	}
	if f.runner.callCount() != 1 {
		t.Fatalf("save retry regenerated: %d", f.runner.callCount())
	}
	if _, err = restarted.Run(context.Background(), "alice", r.ID); err != nil {
		t.Fatal(err)
	}
	if len(f.store.chapterVersions) != 1 {
		t.Fatal("duplicate final commit created history")
	}
}

func TestWorkerQueuedSettingsAreImmutable(t *testing.T) {
	f := newServiceFixture(t, nil)
	_, err := f.service.UpdateProjectSettings(context.Background(), "alice", f.project.ID, ProjectSettings{Rules: []string{"original rule"}, Agent: AgentSettings{TargetChapterWords: 1800, MaxRevisionRounds: 1}})
	if err != nil {
		t.Fatal(err)
	}
	r := queuedWorkerRun(t, f)
	_, err = f.service.UpdateProjectSettings(context.Background(), "alice", f.project.ID, ProjectSettings{Rules: []string{"changed rule"}, Agent: AgentSettings{TargetChapterWords: 4000, MaxRevisionRounds: 4}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Run(context.Background(), "alice", r.ID); err != nil {
		t.Fatal(err)
	}
	got := f.runner.lastCall()
	if got.Project.TargetChapterWords != 1800 || got.MaxRevisionRounds != 1 || len(got.Rules) != 1 || got.Rules[0] != "original rule" {
		t.Fatalf("queued request settings changed: %+v", got)
	}
}

type offlineRuntime struct{ Runtime }

func (offlineRuntime) IsCancelled(context.Context, string) (bool, error) {
	return false, errors.New("redis offline")
}
func (offlineRuntime) Publish(context.Context, RunEvent) error { return errors.New("redis offline") }
func TestWorkerRedisOfflineDoesNotLoseSavedRun(t *testing.T) {
	f := newServiceFixture(t, nil)
	f.service.runtime = offlineRuntime{f.runtime}
	r := queuedWorkerRun(t, f)
	got, err := f.service.Run(context.Background(), "alice", r.ID)
	if err != nil || got.Status != RunSucceeded {
		t.Fatalf("Redis notification controls durable execution: %s %v", got.Status, err)
	}
}

func TestWorkerPipelineRecoveryReplaysRepairAndAuditRounds(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	r = legacyAgentRun(t, f, r)
	l, _, err := f.service.claim(context.Background(), "alice", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	provider := &queuedPipelineProvider{responses: []string{`{"title":`, foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	pipeline := pipelineWithProvider(provider)
	ctx := context.WithValue(context.Background(), journalKey{}, &callJournal{f.store, l})
	input := RunContext{OwnerUsername: "alice", Run: r, Project: f.project, Session: f.session, UserMessage: f.sendInput("worker").Content, ChapterNumber: 1, PreviousTruth: json.RawMessage(`{}`), PreviousMemory: json.RawMessage(`{}`), PreviousChapters: []PreviousChapter{}, Rules: []string{}, Runtime: f.runtime, MaxRevisionRounds: 2}
	if _, err = pipeline.Run(ctx, input, &recordingEventSink{}); err != nil {
		t.Fatal(err)
	}
	e, _, _ := f.store.GetExecution(ctx, "alice", r.ID)
	if len(e.Calls) != 9 {
		t.Fatalf("repair/audit rounds collided: %d calls", len(e.Calls))
	}
	expireWorkerLease(t, f, l)
	restarted := NewService(f.store, NewMemoryRuntime(), pipeline)
	got, err := restarted.Run(context.Background(), "alice", r.ID)
	if err != nil || got.Status != RunSucceeded {
		t.Fatalf("pipeline recovery=%s %v", got.Status, err)
	}
	if len(provider.requests) != 9 {
		t.Fatalf("successful steps resent: %d", len(provider.requests))
	}
}

func TestWorkerFinalCommitIsImmutableAndIdempotent(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	l, _, err := f.service.claim(context.Background(), "alice", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	g := &guardedStore{Store: f.store, lease: l}
	bundle := serviceTestBundle(RunContext{Run: r, Project: f.project}, RunSucceeded)
	if err = g.SaveChapterBundle(context.Background(), "alice", bundle); err != nil {
		t.Fatal(err)
	}
	if err = g.SaveChapterBundle(context.Background(), "alice", bundle); err != nil {
		t.Fatalf("idempotent save failed: %v", err)
	}
	bundle.Chapter.Content = "must not overwrite"
	if err = g.SaveChapterBundle(context.Background(), "alice", bundle); !IsCode(err, ErrIdempotencyConflict) {
		t.Fatalf("changed final artifact accepted: %v", err)
	}
	chapters, _ := f.store.ListChapters(context.Background(), "alice", f.project.ID)
	if len(chapters) != 1 || chapters[0].Content == bundle.Chapter.Content || len(f.store.chapterVersions) != 1 {
		t.Fatal("immutable history overwritten")
	}
}

func TestWorkerRenewExtendsLeaseAndExpiredCannotRenew(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	l, e, err := f.service.claim(context.Background(), "alice", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.service.renew(context.Background(), l); err != nil {
		t.Fatal(err)
	}
	next, _, _ := f.store.GetExecution(context.Background(), "alice", r.ID)
	if next.LeaseUntil.Before(e.LeaseUntil) || next.Fence != e.Fence {
		t.Fatal("renew did not preserve fence and extend lease")
	}
	expireWorkerLease(t, f, l)
	if err = f.service.renew(context.Background(), l); !IsCode(err, ErrLeaseLost) {
		t.Fatalf("expired lease revived: %v", err)
	}
}

func TestWorkerStartupScansWithoutBrowserOrRedisQueue(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	restarted := NewService(f.store, NewMemoryRuntime(), f.runner)
	StartWorker(ctx, restarted, 5*time.Millisecond)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		got, _ := f.store.GetRun(ctx, "alice", r.ID)
		if got.Status == RunSucceeded {
			if f.runner.callCount() != 1 {
				t.Fatal("duplicate startup dispatch")
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("startup did not scan durable dispatch")
}

type checkpointFailureStore struct{ Store }

func (s checkpointFailureStore) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, owner, project, func(tx Store) error { return fn(checkpointFailureStore{tx}) })
}
func (s checkpointFailureStore) UpdateRun(ctx context.Context, owner string, u RunUpdate) (Run, error) {
	if u.Status == RunAwaitingConfirmation {
		return Run{}, errors.New("checkpoint status failed")
	}
	return s.Store.UpdateRun(ctx, owner, u)
}
func TestWorkerCheckpointFailureRollsBackFoundation(t *testing.T) {
	f := newServiceFixture(t, checkpointRunner())
	session := newLongFormSession(t, f)
	input := f.sendInput("checkpoint-atomic")
	input.SessionID = session.ID
	r, err := f.service.SendMessage(context.Background(), "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	f.service.store = checkpointFailureStore{f.store}
	if _, err = f.service.Run(context.Background(), "alice", r.ID); err == nil {
		t.Fatal("expected checkpoint save failure")
	}
	if _, found, err := f.store.GetFoundation(context.Background(), "alice", f.project.ID); err != nil || found {
		t.Fatalf("partial blueprint escaped checkpoint transaction: found=%v err=%v", found, err)
	}
}

func TestWorkerKnownRateRejectionIsPersistedAndNeverReplayed(t *testing.T) {
	f := newServiceFixture(t, nil)
	r := queuedWorkerRun(t, f)
	l, _, _ := f.service.claim(context.Background(), "alice", r.ID)
	ctx := context.WithValue(context.Background(), journalKey{}, &callJournal{f.store, l})
	p := &workerProvider{err: &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}}
	for i := 0; i < 2; i++ {
		if _, err := journalComplete(ctx, "draft", p, batchfactoryv11.TextCompletionRequest{}); !IsCode(err, ErrModelRateLimited) {
			t.Fatalf("known rejection misclassified: %v", err)
		}
	}
	if p.calls != 1 {
		t.Fatal("rejected step replay sent another request")
	}
}

func TestWorkerAtomicConcurrentMessages(t *testing.T) {
	f := newServiceFixture(t, nil)
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.service.SendMessage(context.Background(), "alice", f.sendInput("concurrent-key"))
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	msgs, err := f.store.ListMessages(context.Background(), "alice", f.session.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("atomic request left %d messages", len(msgs))
	}
}
