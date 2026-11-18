package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type stubPipelineRunner struct {
	mu    sync.Mutex
	calls []RunContext
	run   func(context.Context, RunContext, EventSink) (ChapterBundle, error)
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

func TestServiceSendMessageRedisFailureStillPersistsDispatch(t *testing.T) {
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
	if err != nil {
		t.Fatal(err)
	}
	snapshot, getErr := service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if getErr != nil {
		t.Fatal(getErr)
	}
	if snapshot.Run.Status != RunQueued {
		t.Fatalf("redis wake failure changed durable run: %s", snapshot.Run.Status)
	}
	pending, err := store.ListDispatch(context.Background(), 10)
	if err != nil || len(pending) != 1 || pending[0].RunID != run.ID {
		t.Fatalf("dispatch lost: %+v %v", pending, err)
	}
	if len(snapshot.Events) != 1 || snapshot.Events[0].Type != "run.queued" {
		t.Fatalf("events=%+v", snapshot.Events)
	}

}

func TestServiceRunUsesServerOwnerAcquiresLockPersistsEventsAndBundle(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON,
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
	if completed.Status != RunSucceeded || len(provider.requests) != 7 {
		t.Fatalf("completed=%+v provider requests=%d", completed, len(provider.requests))
	}
	var checkSnapshot struct {
		LocalChecks map[string]json.RawMessage `json:"localChecks"`
	}
	if json.Unmarshal(completed.RuntimeSnapshot, &checkSnapshot) != nil || len(checkSnapshot.LocalChecks["post_write"]) == 0 {
		t.Fatal("saved run lost local validation report")
	}
	snapshot, err := service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Chapters) != 1 || snapshot.Chapters[0].Title == "" || snapshot.Chapters[0].Content == "" {
		t.Fatalf("snapshot chapters=%+v, want saved chapter body", snapshot.Chapters)
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

func newLongFormSession(t *testing.T, fixture serviceFixture) Session {
	t.Helper()
	session, err := fixture.service.CreateSession(context.Background(), "alice", CreateSessionInput{
		ProjectID: fixture.project.ID, Mode: "long_original", Title: "长篇创作",
	})
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func checkpointRunner() *stubPipelineRunner {
	runner := &stubPipelineRunner{}
	runner.run = func(_ context.Context, input RunContext, _ EventSink) (ChapterBundle, error) {
		if input.Checkpoint && len(input.ConfirmedFoundation) == 0 {
			return checkpointBundle(input, json.RawMessage(foundationJSON))
		}
		return serviceTestBundle(input, RunSucceeded), nil
	}
	return runner
}

func TestServiceCheckpointPausesThenConfirmsIntoChapterRun(t *testing.T) {
	fixture := newServiceFixture(t, checkpointRunner())
	session := newLongFormSession(t, fixture)
	run, err := fixture.service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: session.ID, ProjectID: fixture.project.ID, Content: "写长篇第一章",
		ModelID: "text-1", IdempotencyKey: "checkpoint-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	paused, err := fixture.service.Run(context.Background(), "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if paused.Status != RunAwaitingConfirmation || paused.FinishedAt != nil {
		t.Fatalf("paused=%+v, want awaiting_confirmation without finish time", paused)
	}
	if len(fixture.store.chapterBundles) != 0 {
		t.Fatal("checkpoint must not save a chapter body")
	}
	snapshot, err := fixture.service.GetSnapshot(context.Background(), "alice", run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	checkpointEvent := false
	for _, event := range snapshot.Events {
		if event.Type == "run.checkpoint" {
			checkpointEvent = true
		}
	}
	if !checkpointEvent {
		t.Fatalf("events=%+v, want run.checkpoint", snapshot.Events)
	}

	confirmed, err := fixture.service.ConfirmRun(context.Background(), "alice", paused.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if confirmed.Status != RunQueued || confirmed.ID == paused.ID {
		t.Fatalf("confirmed=%+v", confirmed)
	}
	completed, err := fixture.service.Run(context.Background(), "alice", confirmed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != RunSucceeded {
		t.Fatalf("completed=%+v", completed)
	}
	if len(fixture.runner.lastCall().ConfirmedFoundation) == 0 {
		t.Fatal("continuation run must carry the confirmed blueprint")
	}
}

func TestServiceConfirmRunValidatesAndPersistsEditedBlueprint(t *testing.T) {
	fixture := newServiceFixture(t, checkpointRunner())
	session := newLongFormSession(t, fixture)
	run, err := fixture.service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: session.ID, ProjectID: fixture.project.ID, Content: "写长篇第一章",
		ModelID: "text-1", IdempotencyKey: "checkpoint-edited",
	})
	if err != nil {
		t.Fatal(err)
	}
	paused, err := fixture.service.Run(context.Background(), "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.ConfirmRun(context.Background(), "alice", paused.ID, json.RawMessage(`{"title":""}`)); !IsCode(err, ErrInputRequired) {
		t.Fatalf("invalid blueprint err=%v, want %s", err, ErrInputRequired)
	}
	edited := strings.Replace(foundationJSON, "雨幕记者", "雨幕新记", 1)
	if _, err := fixture.service.ConfirmRun(context.Background(), "alice", paused.ID, json.RawMessage(edited)); err != nil {
		t.Fatal(err)
	}
	foundation, found, err := fixture.store.GetFoundation(context.Background(), "alice", fixture.project.ID)
	if err != nil || !found {
		t.Fatalf("foundation found=%v err=%v", found, err)
	}
	if !strings.Contains(string(foundation.Snapshot), "雨幕新记") {
		t.Fatalf("edited blueprint not persisted: %s", foundation.Snapshot)
	}
}

func TestServiceRejectsMissingModeInputBeforePersistingMessage(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	session, err := fixture.service.CreateSession(context.Background(), "alice", CreateSessionInput{
		ProjectID: fixture.project.ID, Mode: "continuation",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = fixture.service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: session.ID, ProjectID: fixture.project.ID, Content: "接着写",
		ModelID: "text-1", IdempotencyKey: "missing-mode-input",
	})
	if !IsCode(err, ErrInputRequired) {
		t.Fatalf("err=%v, want %s", err, ErrInputRequired)
	}
	messages, err := fixture.store.ListMessages(context.Background(), "alice", session.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 0 {
		t.Fatalf("messages=%d, invalid mode input must not persist", len(messages))
	}
	run, err := fixture.service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: session.ID, ProjectID: fixture.project.ID, Content: "接着写",
		ModelID: "text-1", IdempotencyKey: "with-mode-input",
		ModeInputs: map[string]string{"previousText": "  前文最后一段  "},
	})
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	if err := json.Unmarshal(run.PromptSnapshot, &snapshot); err != nil {
		t.Fatal(err)
	}
	modeInputs, _ := snapshot["mode_inputs"].(map[string]any)
	if modeInputs["previousText"] != "前文最后一段" || snapshot["mode"] != "continuation" {
		t.Fatalf("prompt snapshot=%v", snapshot)
	}
}

func TestServiceInjectsProjectSettingsIntoPipelineRun(t *testing.T) {
	runner := &stubPipelineRunner{}
	fixture := newServiceFixture(t, runner)
	if _, err := fixture.service.UpdateProjectSettings(context.Background(), "alice", fixture.project.ID, ProjectSettings{
		StyleProfile: " 冷峻短句 ", Rules: []string{" 不写血腥 ", ""},
	}); err != nil {
		t.Fatal(err)
	}
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("settings-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	call := runner.lastCall()
	if call.StyleProfile != "冷峻短句" || !reflect.DeepEqual(call.Rules, []string{"不写血腥"}) {
		t.Fatalf("pipeline input style=%q rules=%v", call.StyleProfile, call.Rules)
	}
}

func TestServiceExportProjectRendersMarkdownAndText(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("export-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	markdown, err := fixture.service.ExportProject(context.Background(), "alice", fixture.project.ID, "markdown")
	if err != nil {
		t.Fatal(err)
	}
	if markdown.Filename != "雨幕记者.md" || !strings.Contains(markdown.Content, "# 雨幕记者") || !strings.Contains(markdown.Content, "雨水敲着窗。") {
		t.Fatalf("markdown=%+v", markdown)
	}
	text, err := fixture.service.ExportProject(context.Background(), "alice", fixture.project.ID, "txt")
	if err != nil {
		t.Fatal(err)
	}
	if text.Filename != "雨幕记者.txt" || strings.Contains(text.Content, "# ") || !strings.Contains(text.Content, "第1章") {
		t.Fatalf("text=%+v", text)
	}
	if _, err := fixture.service.ExportProject(context.Background(), "alice", fixture.project.ID, "pdf"); !IsCode(err, ErrInputRequired) {
		t.Fatalf("unsupported format err=%v", err)
	}
}

func TestServiceKnowledgeMaterialsReachPipelineAndSnapshot(t *testing.T) {
	runner := &stubPipelineRunner{}
	fixture := newServiceFixture(t, runner)
	material, err := fixture.service.CreateMaterial(context.Background(), "alice", fixture.project.ID, CreateMaterialInput{
		Title: "雨夜参考", Content: "旧城排水系统在暴雨夜失效。",
	})
	if err != nil {
		t.Fatal(err)
	}
	input := fixture.sendInput("knowledge-1")
	input.KnowledgeIDs = []string{material.ID, material.ID}
	run, err := fixture.service.SendMessage(context.Background(), "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	if err := json.Unmarshal(run.PromptSnapshot, &snapshot); err != nil {
		t.Fatal(err)
	}
	if ids, _ := snapshot["knowledge_ids"].([]any); len(ids) != 1 || ids[0] != material.ID {
		t.Fatalf("knowledge_ids=%v", snapshot["knowledge_ids"])
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	call := runner.lastCall()
	if len(call.KnowledgeMaterials) != 1 || call.KnowledgeMaterials[0].Title != "雨夜参考" || !strings.Contains(call.KnowledgeMaterials[0].Content, "排水系统") {
		t.Fatalf("pipeline knowledge=%+v", call.KnowledgeMaterials)
	}
}

func TestServiceRejectsUnknownKnowledgeIDBeforePersistingMessage(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	input := fixture.sendInput("knowledge-missing")
	input.KnowledgeIDs = []string{"material-missing"}
	if _, err := fixture.service.SendMessage(context.Background(), "alice", input); !IsCode(err, ErrInputRequired) {
		t.Fatalf("err=%v, want %s", err, ErrInputRequired)
	}
	messages, err := fixture.store.ListMessages(context.Background(), "alice", fixture.session.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 0 {
		t.Fatalf("messages=%d, invalid knowledge must not persist", len(messages))
	}
}

func TestServiceAgentSettingsClampAndReachPipeline(t *testing.T) {
	runner := &stubPipelineRunner{}
	fixture := newServiceFixture(t, runner)
	settings, err := fixture.service.UpdateProjectSettings(context.Background(), "alice", fixture.project.ID, ProjectSettings{
		Agent: AgentSettings{TargetChapterWords: 3200, MaxRevisionRounds: 9, ThinkingLevel: "deep"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.Agent.TargetChapterWords != 3200 || settings.Agent.MaxRevisionRounds != maxMaxRevisionRounds || settings.Agent.ThinkingLevel != "deep" {
		t.Fatalf("agent settings=%+v", settings.Agent)
	}
	run, err := fixture.service.SendMessage(context.Background(), "alice", SendMessageInput{
		SessionID: fixture.session.ID, ProjectID: fixture.project.ID, Content: "写第一章",
		ModelID: "text-1", IdempotencyKey: "agent-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	call := runner.lastCall()
	if call.MaxRevisionRounds != maxMaxRevisionRounds || call.Project.TargetChapterWords != 3200 || call.Run.ThinkingLevel != "deep" {
		t.Fatalf("pipeline agent input=%+v", call)
	}
}

func TestServiceNotesCRUDAndDashboard(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	note, err := fixture.service.CreateNote(context.Background(), "alice", fixture.project.ID, ProjectNoteInput{Title: "伏笔", Content: "灯塔"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := fixture.service.UpdateNote(context.Background(), "alice", fixture.project.ID, note.ID, ProjectNoteInput{Title: "伏笔", Content: "灯塔与旧船票"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Content != "灯塔与旧船票" {
		t.Fatalf("note=%+v", updated)
	}
	notes, err := fixture.service.ListNotes(context.Background(), "alice", fixture.project.ID)
	if err != nil || len(notes) != 1 {
		t.Fatalf("notes=%v err=%v", notes, err)
	}
	run, err := fixture.service.SendMessage(context.Background(), "alice", fixture.sendInput("dashboard-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.Run(context.Background(), "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	dashboard, err := fixture.service.GetDashboard(context.Background(), "alice", fixture.project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.ChapterCount != 1 || dashboard.TotalWords != 7 || dashboard.RunTotal != 1 || dashboard.SuccessRate != 100 {
		t.Fatalf("dashboard=%+v", dashboard)
	}
	if err := fixture.service.DeleteNote(context.Background(), "alice", fixture.project.ID, note.ID); err != nil {
		t.Fatal(err)
	}
	notes, _ = fixture.service.ListNotes(context.Background(), "alice", fixture.project.ID)
	if len(notes) != 0 {
		t.Fatalf("notes after delete=%v", notes)
	}
}

func TestServiceScheduleDispatchCreatesRunAndAdvancesNextRun(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	schedule, err := fixture.service.CreateSchedule(context.Background(), "alice", fixture.project.ID, ScheduleInput{
		SessionID: fixture.session.ID, IntervalMinutes: 30, Enabled: true, Prompt: "继续写下一章",
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	schedule.NextRunAt = now.Add(-time.Minute)
	if _, err := fixture.store.UpdateSchedule(context.Background(), "alice", schedule); err != nil {
		t.Fatal(err)
	}
	dispatched, err := fixture.service.DispatchDueSchedules(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(dispatched) != 1 || dispatched[0].Run.ID == "" {
		t.Fatalf("dispatched=%+v", dispatched)
	}
	reloaded, err := fixture.store.GetSchedule(context.Background(), "alice", schedule.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.LastRunID != dispatched[0].Run.ID || !reloaded.NextRunAt.Equal(schedule.NextRunAt) {
		t.Fatalf("schedule after dispatch=%+v", reloaded)
	}
	second, err := fixture.service.DispatchDueSchedules(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 0 {
		t.Fatalf("second dispatch=%+v, want none", second)
	}
	finishQueuedRun(t, fixture.store, "alice", dispatched[0].Run, RunSucceeded, "")
	run, err := fixture.service.RunSchedule(context.Background(), "alice", fixture.project.ID, schedule.ID)
	if err != nil {
		t.Fatal(err)
	}
	if run.ID == "" || run.ProjectID != fixture.project.ID {
		t.Fatalf("run=%+v", run)
	}
}

func finishQueuedRun(t *testing.T, store *MemoryStore, owner string, run Run, status RunStatus, errorMessage string) {
	t.Helper()
	now := time.Now().UTC()
	if _, err := store.UpdateRun(context.Background(), owner, RunUpdate{
		ID: run.ID, Status: status, ErrorMessage: errorMessage, FinishedAt: &now,
	}); err != nil {
		t.Fatal(err)
	}
	if status == RunSucceeded || status == RunSucceededWithWarnings {
		n := len(store.chapterVersions) + 1
		if e := store.SaveChapterBundle(context.Background(), owner, ChapterBundle{ProjectID: run.ProjectID, RunID: run.ID, Chapter: Chapter{ID: fmt.Sprintf("chapter-%d", n), ProjectID: run.ProjectID, Number: n, Content: "正文", Version: 1}, ChapterVersion: ChapterVersion{ID: fmt.Sprintf("version-%d", n), ChapterID: fmt.Sprintf("chapter-%d", n), RunID: run.ID, Version: 1, Content: "正文"}, Memories: []Memory{{ID: fmt.Sprintf("memory-%d", n), ChapterID: fmt.Sprintf("chapter-%d", n), Snapshot: json.RawMessage(`{"summary":"已保存的前文"}`)}}, RuntimeSnapshot: json.RawMessage(`{"truthSettlement":{"facts":["已发生"]},"memoryIndex":{"summary":"已保存的前文"}}`), FinalRunStatus: status}); e != nil {
			t.Fatal(e)
		}
	}
	e, ok, err := store.GetExecution(context.Background(), owner, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		e.State = "done"
		if err = store.PutExecution(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}

}

func TestServiceBatchAdvancesChaptersAndCompletes(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	batch, err := fixture.service.CreateBatch(context.Background(), "alice", fixture.project.ID, BatchInput{
		SessionID: fixture.session.ID, TotalChapters: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if batch.Status != BatchRunning || batch.CompletedChapters != 0 || batch.Prompt != defaultBatchPrompt {
		t.Fatalf("batch=%+v", batch)
	}
	first, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(first) != 1 || first[0].Run.ID == "" {
		t.Fatalf("first dispatch=%+v err=%v", first, err)
	}
	finishQueuedRun(t, fixture.store, "alice", first[0].Run, RunSucceeded, "")
	second, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(second) != 1 || second[0].Run.ID == first[0].Run.ID {
		t.Fatalf("second dispatch=%+v err=%v", second, err)
	}
	progress, err := fixture.store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || progress.CompletedChapters != 1 {
		t.Fatalf("progress=%+v err=%v", progress, err)
	}
	finishQueuedRun(t, fixture.store, "alice", second[0].Run, RunSucceededWithWarnings, "")
	third, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(third) != 0 {
		t.Fatalf("third dispatch=%+v err=%v", third, err)
	}
	completed, err := fixture.store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || completed.Status != BatchCompleted || completed.CompletedChapters != 2 {
		t.Fatalf("completed=%+v err=%v", completed, err)
	}
}

func TestServiceBatchPauseResumeAndStop(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	batch, err := fixture.service.CreateBatch(context.Background(), "alice", fixture.project.ID, BatchInput{
		SessionID: fixture.session.ID, TotalChapters: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(first) != 1 {
		t.Fatalf("first dispatch=%+v err=%v", first, err)
	}
	paused, err := fixture.service.PauseBatch(context.Background(), "alice", fixture.project.ID, batch.ID)
	if err != nil || paused.Status != BatchPaused {
		t.Fatalf("paused=%+v err=%v", paused, err)
	}
	finishQueuedRun(t, fixture.store, "alice", first[0].Run, RunSucceeded, "")
	if pending, err := fixture.service.AdvanceBatches(context.Background()); err != nil || len(pending) != 0 {
		t.Fatalf("paused advance=%+v err=%v", pending, err)
	}
	if _, err := fixture.service.ResumeBatch(context.Background(), "alice", fixture.project.ID, batch.ID); err != nil {
		t.Fatal(err)
	}
	second, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(second) != 1 {
		t.Fatalf("resumed dispatch=%+v err=%v", second, err)
	}
	stopped, err := fixture.service.StopBatch(context.Background(), "alice", fixture.project.ID, batch.ID)
	if err != nil || stopped.Status != BatchStopped {
		t.Fatalf("stopped=%+v err=%v", stopped, err)
	}
	cancelled, err := fixture.store.GetRun(context.Background(), "alice", second[0].Run.ID)
	if err != nil || cancelled.Status != RunCancelled {
		t.Fatalf("cancelled run=%+v err=%v", cancelled, err)
	}
	if after, err := fixture.service.AdvanceBatches(context.Background()); err != nil || len(after) != 0 {
		t.Fatalf("stopped advance=%+v err=%v", after, err)
	}
}

func TestServiceBatchFailureThenResumeUsesNewRun(t *testing.T) {
	fixture := newServiceFixture(t, nil)
	batch, err := fixture.service.CreateBatch(context.Background(), "alice", fixture.project.ID, BatchInput{
		SessionID: fixture.session.ID, TotalChapters: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(first) != 1 {
		t.Fatalf("first dispatch=%+v err=%v", first, err)
	}
	finishQueuedRun(t, fixture.store, "alice", first[0].Run, RunFailed, "模型服务限流，请稍后重试")
	if _, err := fixture.service.AdvanceBatches(context.Background()); err != nil {
		t.Fatal(err)
	}
	failed, err := fixture.store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || failed.Status != BatchFailed || failed.LastError == "" {
		t.Fatalf("failed batch=%+v err=%v", failed, err)
	}
	if _, err := fixture.service.ResumeBatch(context.Background(), "alice", fixture.project.ID, batch.ID); err != nil {
		t.Fatal(err)
	}
	retried, err := fixture.service.AdvanceBatches(context.Background())
	if err != nil || len(retried) != 1 || retried[0].Run.ID == first[0].Run.ID {
		t.Fatalf("retried dispatch=%+v err=%v", retried, err)
	}
	reloaded, err := fixture.store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || reloaded.Status != BatchRunning || reloaded.CompletedChapters != 0 {
		t.Fatalf("reloaded batch=%+v err=%v", reloaded, err)
	}
}
