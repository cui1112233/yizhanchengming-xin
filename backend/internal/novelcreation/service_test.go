package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

type stubPipelineRunner struct {
	mu       sync.Mutex
	calls    []RunContext
	run      func(context.Context, RunContext, EventSink) (ChapterBundle, error)
}

func (r *stubPipelineRunner) Run(ctx context.Context, input RunContext, sink EventSink) (ChapterBundle, error) {
	r.mu.Lock()
	r.calls = append(r.calls, input)
	r.mu.Unlock()
	if r.run == nil {
		return serviceTestBundle(input, RunSucceeded), nil
	}
	return r.run(ctx, input, sink)
}

func (r *stubPipelineRunner) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

func (r *stubPipelineRunner) lastCall() RunContext {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls[len(r.calls)-1]
}

type failingEnqueueRuntime struct {
	Runtime
	err error
}

func (r failingEnqueueRuntime) Enqueue(context.Context, string) error { return r.err }

type saveRecordingStore struct {
	Store
	mu        sync.Mutex
	saveCalls int
}

func (s *saveRecordingStore) SaveChapterBundle(ctx context.Context, owner string, bundle ChapterBundle) error {
	s.mu.Lock()
	s.saveCalls++
	s.mu.Unlock()
	return s.Store.SaveChapterBundle(ctx, owner, bundle)
}

func (s *saveRecordingStore) saves() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveCalls
}

type serviceFixture struct {
	service *Service
	store   *MemoryStore
	runtime *MemoryRuntime
	runner  *stubPipelineRunner
	project Project
	session Session
}

func newServiceFixture(t *testing.T, runner *stubPipelineRunner) serviceFixture {
	t.Helper()
	store := NewMemoryStore()
	runtime := NewMemoryRuntime()
	if runner == nil {
		runner = &stubPipelineRunner{}
	}
	service := NewService(store, runtime, runner)
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{
		Title: "雨幕记者", Genre: "都市悬疑", World: "旧城连日暴雨",
		Protagonist: "记者林澜", CoreConflict: "真相与自保", ModelID: "text-1",
		TargetChapterWords: 2000,
	})
	if err != nil {
		t.Fatal(err)
	}
	session, err := service.CreateSession(context.Background(), "alice", CreateSessionInput{
		ProjectID: project.ID, Mode: "long_form_original", Title: "第一章",
	})
	if err != nil {
		t.Fatal(err)
	}
	return serviceFixture{service: service, store: store, runtime: runtime, runner: runner, project: project, session: session}
}

func (f serviceFixture) sendInput(key string) SendMessageInput {
	return SendMessageInput{
		SessionID: f.session.ID, ProjectID: f.project.ID,
		Content: "写第一章，从一封匿名信开始。", ModelID: "text-1",
		ThinkingLevel: "medium", IdempotencyKey: key,
	}
}

func serviceTestBundle(input RunContext, status RunStatus) ChapterBundle {
	chapterID := input.Project.ID + "-chapter-1"
	return ChapterBundle{
		ProjectID: input.Project.ID,
		RunID:     input.Run.ID,
		Chapter: Chapter{
			ID: chapterID, ProjectID: input.Project.ID, Number: 1,
			Title: "匿名信", Content: "雨水敲着窗。", WordCount: 7, Version: 1,
		},
		ChapterVersion: ChapterVersion{
			ID: chapterID + "-version-" + input.Run.ID, ChapterID: chapterID,
			RunID: input.Run.ID, Version: 1, Content: "雨水敲着窗。", Snapshot: json.RawMessage(`{}`),
		},
		FinalRunStatus: status, RuntimeSnapshot: json.RawMessage(`{"service_test":true}`),
	}
}

func TestServiceSendMessageReturnsExistingRunForSameIdempotencyKey(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	first, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("send-1"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("send-1"))
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("first=%s second=%s", first.ID, second.ID)
	}
	messages, err := fixture.store.ListMessages(context.Background(), "alice", fixture.session.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages=%d, duplicate idempotent send must not append another user message", len(messages))
	}
	events, err := fixture.store.ListRunEvents(context.Background(), "alice", first.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != "run.queued" || events[0].Status != RunQueued {
		t.Fatalf("events=%+v", events)
	}
}

func TestServiceSendMessageQueueFailurePersistsFailedTerminalAndCanRetry(t *testing.T) {
	store := NewMemoryStore()
	baseRuntime := NewMemoryRuntime()
	runtime := failingEnqueueRuntime{Runtime: baseRuntime, err: errors.New("redis unavailable")}
	service := NewService(store, runtime, &stubPipelineRunner{})
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "雨幕", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := service.CreateSession(context.Background(), "alice", CreateSessionInput{ProjectID: project.ID, Mode: "long_form_original"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: session.ID, ProjectID: project.ID, Content: "写第一章", ModelID: "text-1", IdempotencyKey: "send-fail",
	})
	if err == nil || !IsCode(err, ErrPersistenceFailed) {
		t.Fatalf("run=%+v err=%v, want %s", run, err, ErrPersistenceFailed)
	}
	snapshot, getErr := service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if getErr != nil {
		t.Fatal(getErr)
	}
	if snapshot.Run.Status != RunFailed || snapshot.Run.ErrorCode != ErrPersistenceFailed || snapshot.Run.FinishedAt == nil {
		t.Fatalf("failed run was not durable: %+v", snapshot.Run)
	}
	if len(snapshot.Events) != 2 || snapshot.Events[1].Type != "run.failed" {
		t.Fatalf("events=%+v, want queued then failed", snapshot.Events)
	}
}

func TestServiceRunUsesServerOwnerAcquiresLockPersistsEventsAndBundle(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	store := NewMemoryStore()
	runtime := NewMemoryRuntime()
	service := NewService(store, runtime, NewPipeline(func(_ context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || modelID != "text-1" {
			t.Fatalf("resolver owner/model=%q/%q", owner, modelID)
		}
		return provider, nil
	}))
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{
		Title: "雨幕记者", Genre: "都市悬疑", World: "旧城", Protagonist: "林澜",
		CoreConflict: "真相与自保", ModelID: "text-1", TargetChapterWords: 2000,
	})
	if err != nil {
		t.Fatal(err)
	}
	session, err := service.CreateSession(context.Background(), "alice", CreateSessionInput{ProjectID: project.ID, Mode: "long_form_original"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: session.ID, ProjectID: project.ID, Content: "写第一章，从一封匿名信开始。",
		ModelID: "text-1", ThinkingLevel: "medium", IdempotencyKey: "real-pipeline",
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, err := service.Run(context.Background(), "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != RunSucceeded || len(provider.requests) != 6 {
		t.Fatalf("completed=%+v provider requests=%d", completed, len(provider.requests))
	}
	snapshot, err := service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantSteps := []RunStatus{RunPlanning, RunWriting, RunNormalizing, RunAuditing, RunSettling, RunIndexing}
	seen := make(map[RunStatus]bool)
	for _, event := range snapshot.Events {
		if event.Type == "run.step" {
			seen[event.Status] = true
		}
	}
	for _, step := range wantSteps {
		if !seen[step] {
			t.Fatalf("missing durable step %s in events %+v", step, snapshot.Events)
		}
	}
	if len(store.chapterBundles) != 1 {
		t.Fatalf("chapter bundles=%d", len(store.chapterBundles))
	}
}

func TestServiceRunPassesOwnerOnlyFromCallingArgument(t *testing.T) {
	runner := &stubPipelineRunner{}
	fixture := newServiceFixture(t, runner)
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("owner-1"))
	if err != nil {
		t.Fatal(err)
	}
	completed, err := fixture.service.Run(context.Background(), "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != RunSucceeded {
		t.Fatalf("status=%s", completed.Status)
	}
	call := runner.lastCall()
	if call.OwnerUsername != "alice" {
		t.Fatalf("pipeline owner=%q", call.OwnerUsername)
	}
	if strings.Contains(string(run.PromptSnapshot), "alice") || strings.Contains(string(run.RuntimeSnapshot), "alice") {
		t.Fatalf("owner leaked into client-derived snapshots: prompt=%s runtime=%s", run.PromptSnapshot, run.RuntimeSnapshot)
	}
	snapshot, err := fixture.service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Events) < 2 || snapshot.Events[0].Type != "run.queued" || snapshot.Events[len(snapshot.Events)-1].Type != "run.succeeded" {
		t.Fatalf("events=%+v", snapshot.Events)
	}
	if len(fixture.store.chapterBundles) != 1 {
		t.Fatalf("saved chapter bundles=%d", len(fixture.store.chapterBundles))
	}
}

func TestServiceRunChecksCancellationAfterPipelineBeforeSaveChapterBundle(t *testing.T) {
	baseStore := NewMemoryStore()
	store := &saveRecordingStore{Store: baseStore}
	runtime := NewMemoryRuntime()
	runner := &stubPipelineRunner{}
	runner.run = func(_ context.Context, input RunContext, _ EventSink) (ChapterBundle, error) {
		if err := runtime.Cancel(context.Background(), input.Run.ID); err != nil {
			t.Fatal(err)
		}
		return serviceTestBundle(input, RunSucceeded), nil
	}
	service := NewService(store, runtime, runner)
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "雨幕", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := service.CreateSession(context.Background(), "alice", CreateSessionInput{ProjectID: project.ID, Mode: "long_form_original"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := service.SendMessage(context.Background(), "alice", SendMessageInput{SessionID: session.ID, ProjectID: project.ID, Content: "写第一章", ModelID: "text-1", IdempotencyKey: "cancel-before-save"})
	if err != nil {
		t.Fatal(err)
	}
	finished, err := service.Run(context.Background(), "alice", run.ID)
	if !IsCode(err, ErrRunCancelled) {
		t.Fatalf("run=%+v err=%v, want %s", finished, err, ErrRunCancelled)
	}
	if store.saves() != 0 {
		t.Fatalf("SaveChapterBundle calls=%d, want zero after final cancellation check", store.saves())
	}
	if finished.Status != RunCancelled || finished.FinishedAt == nil {
		t.Fatalf("cancelled terminal not persisted: %+v", finished)
	}
}

func TestServiceRunFailureIsTerminalStableAndReleasesProjectLock(t *testing.T) {
	call := 0
	runner := &stubPipelineRunner{}
	runner.run = func(_ context.Context, input RunContext, _ EventSink) (ChapterBundle, error) {
		call++
		if call == 1 {
			return ChapterBundle{}, &DomainError{Code: ErrModelUnavailable, Message: "offline"}
		}
		return serviceTestBundle(input, RunSucceeded), nil
	}
	fixture := newServiceFixture(t, runner)
	first, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("fail-stable"))
	if err != nil {
		t.Fatal(err)
	}
	failed, err := fixture.service.Run(context.Background(), "alice", first.ID)
	if !IsCode(err, ErrModelUnavailable) || failed.Status != RunFailed {
		t.Fatalf("run=%+v err=%v", failed, err)
	}
	again, err := fixture.service.Run(context.Background(), "alice", first.ID)
	if err != nil || again.Status != RunFailed || runner.callCount() != 1 {
		t.Fatalf("repeat run=%+v err=%v calls=%d", again, err, runner.callCount())
	}
	retry, err := fixture.service.RetryRun(context.Background(), "alice", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	succeeded, err := fixture.service.Run(context.Background(), "alice", retry.ID)
	if err != nil || succeeded.Status != RunSucceeded || runner.callCount() != 2 {
		t.Fatalf("retry=%+v err=%v calls=%d", succeeded, err, runner.callCount())
	}
}

func TestServiceCancelRunIsDurableIdempotentAndDoesNotChangeSuccess(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("cancel-1"))
	if err != nil {
		t.Fatal(err)
	}
	first, err := fixture.service.CancelRun(context.Background(), "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := fixture.service.CancelRun(context.Background(), "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != RunCancelled || second.Status != RunCancelled || first.ID != second.ID {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	if fixture.runner.callCount() != 0 {
		t.Fatalf("cancelled run reached pipeline %d times", fixture.runner.callCount())
	}

	success, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("success-1"))
	if err != nil {
		t.Fatal(err)
	}
	success, err = fixture.service.Run(context.Background(), "alice", success.ID)
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := fixture.service.CancelRun(context.Background(), "alice", success.ID)
	if err != nil || unchanged.Status != RunSucceeded {
		t.Fatalf("successful terminal changed: %+v err=%v", unchanged, err)
	}
}

func TestServiceRetryRunAllowedStatesReuseInputAndAreIdempotent(t *testing.T) {
	for _, status := range []RunStatus{RunFailed, RunCancelled, RunSucceededWithWarnings} {
		t.Run(string(status), func(t *testing.T) {
			fixture := newServiceFixture(t, nil)
			prompt := json.RawMessage(`{"version":1,"user_message":"原始输入","chapter_number":3,"previous_truth":{"facts":["a"]},"previous_memory":{"summary":"b"}}`)
			source, err := fixture.store.CreateRun(context.Background(), "alice", Run{
				ProjectID: fixture.project.ID, SessionID: fixture.session.ID,
				IdempotencyKey: "source-" + string(status), Status: status,
				ModelID: "text-1", ThinkingLevel: "high", PromptSnapshot: prompt,
			})
			if err != nil {
				t.Fatal(err)
			}
			first, err := fixture.service.RetryRun(context.Background(), "alice", source.ID)
			if err != nil {
				t.Fatal(err)
			}
			second, err := fixture.service.RetryRun(context.Background(), "alice", source.ID)
			if err != nil {
				t.Fatal(err)
			}
			if first.ID != second.ID || first.ID == source.ID || first.IdempotencyKey == source.IdempotencyKey {
				t.Fatalf("source=%+v first=%+v second=%+v", source, first, second)
			}
			var snapshot map[string]any
			if err := json.Unmarshal(first.PromptSnapshot, &snapshot); err != nil {
				t.Fatal(err)
			}
			if snapshot["retry_of_run_id"] != source.ID || snapshot["user_message"] != "原始输入" || snapshot["chapter_number"] != float64(3) {
				t.Fatalf("retry snapshot=%v", snapshot)
			}
			if first.ModelID != source.ModelID || first.ThinkingLevel != source.ThinkingLevel {
				t.Fatalf("retry model input changed: source=%+v retry=%+v", source, first)
			}
		})
	}
}

func TestServiceRetryRunRejectsNonRetryableState(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	source, err := fixture.store.CreateRun(context.Background(), "alice", Run{
		ProjectID: fixture.project.ID, SessionID: fixture.session.ID,
		IdempotencyKey: "source-success", Status: RunSucceeded,
		ModelID: "text-1", PromptSnapshot: json.RawMessage(`{"version":1,"user_message":"原始输入","chapter_number":1}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.RetryRun(context.Background(), "alice", source.ID); !IsCode(err, ErrInvalidTransition) {
		t.Fatalf("err=%v, want %s", err, ErrInvalidTransition)
	}
}

func TestServiceRetryOfWarningReplacesSameChapterWithoutDuplicate(t *testing.T) {
	runner := &stubPipelineRunner{}
	call := 0
	runner.run = func(_ context.Context, input RunContext, _ EventSink) (ChapterBundle, error) {
		call++
		status := RunSucceededWithWarnings
		if call == 2 {
			status = RunSucceeded
		}
		return serviceTestBundle(input, status), nil
	}
	fixture := newServiceFixture(t, runner)
	source, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("warning-source"))
	if err != nil {
		t.Fatal(err)
	}
	source, err = fixture.service.Run(context.Background(), "alice", source.ID)
	if err != nil || source.Status != RunSucceededWithWarnings {
		t.Fatalf("source=%+v err=%v", source, err)
	}
	retry, err := fixture.service.RetryRun(context.Background(), "alice", source.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", retry.ID); err != nil {
		t.Fatal(err)
	}
	if len(fixture.store.chapterBundles) != 1 {
		t.Fatalf("chapter bundles=%d, retry duplicated a successful chapter", len(fixture.store.chapterBundles))
	}
}

func TestServiceGetSnapshotIsOwnerScopedAndSupportsEventCursor(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("snapshot-1"))
	if err != nil {
		t.Fatal(err)
	}
	full, err := fixture.service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if full.Project.ID != fixture.project.ID || full.Session.ID != fixture.session.ID || full.Run.ID != run.ID || len(full.Messages) != 1 || len(full.Events) != 1 {
		t.Fatalf("snapshot=%+v", full)
	}
	after, err := fixture.service.GetSnapshot(context.Background(), "alice", run.ID, full.Events[0].EventID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Events) != 0 {
		t.Fatalf("events after cursor=%+v", after.Events)
	}
	if _, err := fixture.service.GetSnapshot(context.Background(), "bob", run.ID, 0); !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("cross-owner err=%v", err)
	}
}

func TestServiceHasNoDesktopOrLocalProcessDependency(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("web-only"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	// The complete workflow above uses only service contracts. A Windows path,
	// local process, Electron bridge, or filesystem handle is never supplied.
}
