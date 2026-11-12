package aicut

import (
	"context"
	"strings"
	"testing"
)

func TestRepairAgentCanListASRJobStatus(t *testing.T) {
	store, _, run, exec := newAgentToolFixture(t, false)
	now := exec.now()
	if err := store.CreateASRJob(context.Background(), run.Owner, ASRJob{ID: "asr-job", Owner: run.Owner, ProjectID: run.ProjectID, AssetID: "a", ModelID: "m", State: ASRRunning, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	got, err := exec.Execute(context.Background(), run, ToolCall{ID: "jobs", Name: "get_asr_jobs", Arguments: toolArgs(map[string]any{})})
	if err != nil || !got.OK || !strings.Contains(string(got.Data), "asr-job") {
		t.Fatalf("status unavailable: %#v %v", got, err)
	}
}
func TestRepairAgentOralCleanDefaultsToRunModel(t *testing.T) {
	_, _, run, exec := newAgentToolFixture(t, false)
	run.ModelID = "current-text"
	called := ""
	exec.ConfigureCompositeTools(AgentCompositeTools{OralClean: func(_ context.Context, _, _, model string, _, _ int64) (Timeline, FillerTrimSummary, string, error) {
		called = model
		return Timeline{}, FillerTrimSummary{}, "done", nil
	}})
	got, err := exec.Execute(context.Background(), run, ToolCall{ID: "oral", Name: "clean_oral_pauses", Arguments: toolArgs(map[string]any{})})
	if err != nil || !got.OK || called != "current-text" {
		t.Fatalf("current model not used: %#v %v", got, err)
	}
}
func TestRepairAgentASRCatalogFiltersAndRedacts(t *testing.T) {
	settings := map[string]any{"modelCatalog": []any{
		map[string]any{"id": "ready", "kind": "asr", "enabled": true, "name": "可用语音", "modelId": "whisper", "baseUrl": "https://example.test/v1", "credential": "top-secret"},
		map[string]any{"id": "disabled", "kind": "asr", "enabled": false, "modelId": "whisper", "baseUrl": "https://example.test/v1", "credential": "top-secret"},
		map[string]any{"id": "missing-key", "kind": "asr", "enabled": true, "modelId": "whisper", "baseUrl": "https://example.test/v1"},
		map[string]any{"id": "missing-model", "kind": "asr", "enabled": true, "baseUrl": "https://example.test/v1", "credential": "top-secret"},
	}}
	got := AgentASRModelsFromCatalog(settings)
	if len(got) != 1 || got[0].ID != "ready" {
		t.Fatalf("unavailable models exposed: %#v", got)
	}
	if strings.Contains(string(toolResultData(got)), "top-secret") || strings.Contains(string(toolResultData(got)), "example.test") {
		t.Fatal("provider details leaked")
	}
}
func TestRepairAgentNamedASRStatusUsesProjectJobStoreAndFiltersScope(t *testing.T) {
	ctx := context.Background()
	store, now := createASRFixture(t)
	collection := NewMemoryTimelineCollectionStore(store)
	a, err := collection.Create(ctx, "alice", "cut-1", "A")
	if err != nil {
		t.Fatal(err)
	}
	agents := NewMemoryAgentStore()
	run := AgentRun{ID: "named-run", ClientRequestID: "named-request", Owner: "alice", ProjectID: "cut-1", TimelineID: a.ID, Instruction: "字幕任务状态", Mode: AgentModeAgent, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	if _, err = agents.CreateAgentRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{a.ID, "other", ""} {
		id := "status-" + scope
		if scope == "" {
			id = "legacy-status"
		}
		if err = store.CreateASRJob(ctx, "alice", ASRJob{ID: id, Owner: "alice", ProjectID: "cut-1", TimelineID: scope, AssetID: "asset-1", ModelID: "asr", State: ASRRunning, CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	exec := NewAgentToolExecutor(store, agents, nil, collection)
	got, err := exec.Execute(ctx, run, ToolCall{ID: "read-named-status", Name: "get_asr_jobs", Arguments: toolArgs(map[string]any{})})
	if err != nil || !got.OK || !strings.Contains(string(got.Data), "status-"+a.ID) {
		t.Fatalf("named status unavailable: %#v %v", got, err)
	}
	if strings.Contains(string(got.Data), "status-other") || strings.Contains(string(got.Data), "legacy-status") {
		t.Fatalf("other timeline job leaked: %s", got.Data)
	}
}
