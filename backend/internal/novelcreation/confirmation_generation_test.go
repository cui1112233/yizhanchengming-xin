package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"testing"
)

func confirmationGenerationFixture(t *testing.T) (serviceFixture, *metadataQueuedProvider, Run) {
	t.Helper()
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"requireSecondaryConfirm":true,"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	if _, err := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a}); err != nil {
		t.Fatal(err)
	}
	provider := &metadataQueuedProvider{queuedPipelineProvider: &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}, metadata: batchfactoryv11.TextModelMetadata{ActualModel: "actual-model", ChannelID: "safe-channel", ConfigRevision: "A"}}
	f.service.pipeline = pipelineWithProvider(provider)
	run := queuedWorkerRun(t, f)
	if _, err := f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	return f, provider, run
}
func pendingGenerationID(t *testing.T, f serviceFixture, id string) string {
	t.Helper()
	run, err := f.store.GetRun(context.Background(), "alice", id)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		ID string `json:"confirmationId"`
	}
	if err = json.Unmarshal(run.RuntimeSnapshot, &snapshot); err != nil {
		t.Fatal(err)
	}
	if run.Status != RunAwaitingConfirmation || snapshot.ID == "" {
		t.Fatalf("missing pending confirmation: status=%s runtime=%s", run.Status, run.RuntimeSnapshot)
	}
	return snapshot.ID
}

func TestAgentConfirmationGenerationConfigAToBToA(t *testing.T) {
	ctx := context.Background()
	f, provider, run := confirmationGenerationFixture(t)
	first := pendingGenerationID(t, f, run.ID)
	if _, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, first); err != nil {
		t.Fatal(err)
	}
	provider.metadata.ConfigRevision = "B"
	if _, err := f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	second := pendingGenerationID(t, f, run.ID)
	if _, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, second); err != nil {
		t.Fatal(err)
	}
	provider.metadata.ConfigRevision = "A"
	if _, err := f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	third := pendingGenerationID(t, f, run.ID)
	if first == second || first == third || second == third {
		t.Fatalf("A1/B1/A2 must be unique: first==third=%v", first == third)
	}
	f.service = NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(provider))
	for _, old := range []string{first, second, first} {
		got, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, old)
		if err != nil || got.Status != RunAwaitingConfirmation {
			t.Fatalf("old approval authorized new pending: %+v %v", got, err)
		}
		if pendingGenerationID(t, f, run.ID) != third {
			t.Fatal("observing old approval changed the current pending ID")
		}
	}
	if _, err := f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	if pendingGenerationID(t, f, run.ID) != third || len(provider.requests) != 1 {
		t.Fatal("restart/observation changed pending authorization or resent history")
	}
	for i := 0; i < 2; i++ {
		got, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, third)
		if err != nil || got.Status != RunQueued {
			t.Fatalf("current A2 did not resume idempotently: status=%s err=%v", got.Status, err)
		}
	}
	f.service = NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(provider))
	if _, err := f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	fourth := pendingGenerationID(t, f, run.ID)
	if fourth == third || len(provider.requests) != 2 {
		t.Fatalf("A2 not sent exactly once: calls=%d", len(provider.requests))
	}
	for _, old := range []string{first, second, third} {
		if _, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, old); err != nil {
			t.Fatal(err)
		}
		if pendingGenerationID(t, f, run.ID) != fourth {
			t.Fatal("late approval authorized following call")
		}
	}
}

func TestAgentConfirmationGenerationRepairsLegacyApprovedPending(t *testing.T) {
	ctx := context.Background()
	f, provider, run := confirmationGenerationFixture(t)
	old := pendingGenerationID(t, f, run.ID)
	state, _, _ := f.store.GetExecution(ctx, "alice", run.ID)
	key := state.PendingConfirmation.Key
	call := state.Calls[key]
	old = fingerprint(struct {
		Run, Key, Request string
		Metadata          *ModelCallMetadata
	}{run.ID, key, call.Fingerprint, state.PendingConfirmation.Metadata})
	state.PendingConfirmation.ID = old
	call.ConfirmationID, call.Confirmed = old, false
	state.Calls[key] = call
	state.ConfirmedCalls = map[string]string{old: state.PendingConfirmation.Key}
	// Seed the exact persisted legacy deadlock shape, with no new generation fields.
	raw, _ := json.Marshal(state)
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields)
	for _, call := range fields["Calls"].(map[string]any) {
		delete(call.(map[string]any), "ConfirmationGeneration")
		delete(call.(map[string]any), "ConfirmationFingerprint")
	}
	raw, _ = json.Marshal(fields)
	state = Execution{}
	_ = json.Unmarshal(raw, &state)
	if err := f.store.PutExecution(ctx, state); err != nil {
		t.Fatal(err)
	}
	current, _ := f.store.GetRun(ctx, "alice", run.ID)
	var snapshot map[string]any
	_ = json.Unmarshal(current.RuntimeSnapshot, &snapshot)
	snapshot["confirmationId"] = old
	raw, _ = json.Marshal(snapshot)
	if _, err := f.store.UpdateRun(ctx, "alice", RunUpdate{ID: run.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: raw}); err != nil {
		t.Fatal(err)
	}
	f.service = NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(provider))
	current, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, old)
	if err != nil || current.Status != RunAwaitingConfirmation {
		t.Fatalf("legacy repair should still require confirmation: %+v %v", current, err)
	}
	fresh := pendingGenerationID(t, f, run.ID)
	if fresh == old || len(provider.requests) != 1 {
		t.Fatal("legacy approved pending was neither refreshed nor kept unsent")
	}
	if _, err = f.service.ConfirmModelCall(ctx, "alice", run.ID, old); err != nil {
		t.Fatal(err)
	}
	if pendingGenerationID(t, f, run.ID) != fresh {
		t.Fatal("old token regenerated the repaired pending on retry")
	}
	current, err = f.service.ConfirmModelCall(ctx, "alice", run.ID, fresh)
	if err != nil || current.Status != RunQueued {
		t.Fatalf("fresh legacy repair token not usable: %+v %v", current, err)
	}
	if _, err = f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	if len(provider.requests) != 2 {
		t.Fatal("legacy repair did not resume exactly one external call")
	}
}

type confirmationCheckpointFault struct {
	Store
	fail *bool
}

func (s confirmationCheckpointFault) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, owner, project, func(tx Store) error { return fn(confirmationCheckpointFault{Store: tx, fail: s.fail}) })
}

func (s confirmationCheckpointFault) UpdateRun(ctx context.Context, owner string, update RunUpdate) (Run, error) {
	if update.Status == RunAwaitingConfirmation && *s.fail {
		*s.fail = false
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "injected checkpoint save failure"}
	}
	return s.Store.UpdateRun(ctx, owner, update)
}

func TestAgentConfirmationGenerationSurvivesCheckpointSaveFailure(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	var settings AgentSettings
	_ = json.Unmarshal([]byte(`{"requireSecondaryConfirm":true,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &settings)
	if _, err := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: settings}); err != nil {
		t.Fatal(err)
	}
	provider := &metadataQueuedProvider{queuedPipelineProvider: &queuedPipelineProvider{responses: []string{foundationJSON}}, metadata: batchfactoryv11.TextModelMetadata{ConfigRevision: "A"}}
	f.service.pipeline = pipelineWithProvider(provider)
	fail := true
	f.service.store = confirmationCheckpointFault{Store: f.store, fail: &fail}
	run := queuedWorkerRun(t, f)
	if _, err := f.service.Run(ctx, "alice", run.ID); err == nil {
		t.Fatal("expected checkpoint persistence failure")
	}
	before, _, _ := f.store.GetExecution(ctx, "alice", run.ID)
	if before.PendingConfirmation == nil || before.State != "ready" {
		t.Fatalf("pending was not durable before checkpoint failure: %+v", before)
	}
	f.service = NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(provider))
	if _, err := f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	if pendingGenerationID(t, f, run.ID) != before.PendingConfirmation.ID || len(provider.requests) != 1 {
		t.Fatal("re-observing durable pending changed its ID or resent history")
	}
}
