package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func newAgentGenerationLifecycle(t *testing.T, now time.Time) (*agentworkspace.MemoryStore, *http.ServeMux) {
	t.Helper()
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: "image-1", Type: "image", Title: "图片", Version: 1, Status: "draft"}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentGenerationTaskRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })
	return store, mux
}

func TestAgentGenerationLifecycleIsOwnerScopedAndIdempotent(t *testing.T) {
	now := time.Date(2026, 10, 21, 15, 0, 0, 0, time.UTC)
	_, mux := newAgentGenerationLifecycle(t, now)
	body := map[string]any{
		"nodeId": "image-1", "nodeVersion": 1, "kind": "image", "provider": "personal_api", "modelId": "image-model",
		"mode": "text_to_image", "parameters": map[string]any{"ratio": "16:9", "count": 1}, "idempotencyKey": "image-1-v1",
	}
	first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations", body)
	second := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations", body)
	if first.Code != http.StatusAccepted || second.Code != http.StatusAccepted {
		t.Fatalf("first=%d %s second=%d %s", first.Code, first.Body.String(), second.Code, second.Body.String())
	}
	var firstPayload, secondPayload struct {
		Task agentworkspace.GenerationTask `json:"task"`
	}
	if json.Unmarshal(first.Body.Bytes(), &firstPayload) != nil || json.Unmarshal(second.Body.Bytes(), &secondPayload) != nil || firstPayload.Task.ID == "" || firstPayload.Task.ID != secondPayload.Task.ID {
		t.Fatalf("first=%s second=%s", first.Body.String(), second.Body.String())
	}

	list := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/tasks/project-1/generations", nil)
	if list.Code != http.StatusOK || !containsText(list.Body.String(), firstPayload.Task.ID) {
		t.Fatalf("list=%d %s", list.Code, list.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/agent/tasks/project-1/generations/"+firstPayload.Task.ID, nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
}

func TestAgentGenerationSummaryCountsPersistedStates(t *testing.T) {
	now := time.Date(2026, 10, 23, 14, 0, 0, 0, time.UTC)
	store, mux := newAgentGenerationLifecycle(t, now)
	states := []agentworkspace.GenerationState{agentworkspace.GenerationQueued, agentworkspace.GenerationRunning, agentworkspace.GenerationAwaitingConfirmation, agentworkspace.GenerationCompleted, agentworkspace.GenerationFailed, agentworkspace.GenerationCancelled}
	for index, state := range states {
		input := agentworkspace.CreateGenerationTaskInput{ID: "summary-" + string(rune('a'+index)), ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "provider", ModelID: "model", IdempotencyKey: "summary-" + string(rune('a'+index)), Now: now.Add(time.Duration(index) * time.Second)}
		task, err := store.CreateGenerationTask(context.Background(), "alice", input)
		if err != nil {
			t.Fatal(err)
		}
		if state == agentworkspace.GenerationQueued {
			continue
		}
		task, err = store.TransitionGenerationTask(context.Background(), "alice", "project-1", task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Duration(index+10) * time.Second)})
		if err != nil {
			t.Fatal(err)
		}
		if state == agentworkspace.GenerationRunning {
			continue
		}
		transition := agentworkspace.GenerationTaskTransition{State: state, Now: now.Add(time.Duration(index+20) * time.Second)}
		if state == agentworkspace.GenerationAwaitingConfirmation || state == agentworkspace.GenerationCompleted {
			transition.State = agentworkspace.GenerationAwaitingConfirmation
			transition.Results = []agentworkspace.GenerationResult{{ID: task.ID + "-result", AssetID: task.ID + "-asset", Kind: "image"}}
			transition.SelectedResultID = task.ID + "-result"
		}
		task, err = store.TransitionGenerationTask(context.Background(), "alice", "project-1", task.ID, transition)
		if err != nil {
			t.Fatal(err)
		}
		if state == agentworkspace.GenerationCompleted {
			if _, err = store.TransitionGenerationTask(context.Background(), "alice", "project-1", task.ID, agentworkspace.GenerationTaskTransition{State: state, Now: now.Add(time.Duration(index+30) * time.Second)}); err != nil {
				t.Fatal(err)
			}
		}
	}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/tasks/project-1/generation-summary", nil)
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"total":6`) || !containsText(response.Body.String(), `"awaitingConfirmation":1`) || !containsText(response.Body.String(), `"failed":1`) {
		t.Fatalf("summary=%d %s", response.Code, response.Body.String())
	}
}

func TestAgentGenerationConfirmAppliesSelectedResult(t *testing.T) {
	now := time.Date(2026, 10, 21, 16, 0, 0, 0, time.UTC)
	store, mux := newAgentGenerationLifecycle(t, now)
	created, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "generation-1", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "personal_api", ModelID: "image-model", IdempotencyKey: "image-1-v1", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	running, err := store.TransitionGenerationTask(context.Background(), "alice", created.ProjectID, created.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.TransitionGenerationTask(context.Background(), "alice", created.ProjectID, created.ID, agentworkspace.GenerationTaskTransition{
		State: agentworkspace.GenerationAwaitingConfirmation, ExpectedUpdatedAt: running.UpdatedAt, SelectedResultID: "result-2", Now: now.Add(2 * time.Second),
		Results: []agentworkspace.GenerationResult{
			{ID: "result-1", AssetID: "asset-1", Kind: "image", Asset: map[string]any{"id": "asset-1", "kind": "image", "name": "一.png"}},
			{ID: "result-2", AssetID: "asset-2", Kind: "image", Asset: map[string]any{"id": "asset-2", "kind": "image", "name": "二.png"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	confirm := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/generation-1/confirm", map[string]any{"resultId": "result-2"})
	if confirm.Code != http.StatusOK || !containsText(confirm.Body.String(), "asset-2") || !containsText(confirm.Body.String(), "completed") {
		t.Fatalf("confirm=%d %s", confirm.Code, confirm.Body.String())
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || canvas.Nodes[0].AssetRef["id"] != "asset-2" {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
}

func TestAgentGenerationCanActivateHistoricalResultWithVersionGuard(t *testing.T) {
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	store, mux := newAgentGenerationLifecycle(t, now)
	created, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "generation-1", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "personal_api", ModelID: "image-model", IdempotencyKey: "image-1-v1", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	running, err := store.TransitionGenerationTask(context.Background(), "alice", created.ProjectID, created.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.TransitionGenerationTask(context.Background(), "alice", created.ProjectID, created.ID, agentworkspace.GenerationTaskTransition{
		State: agentworkspace.GenerationAwaitingConfirmation,
		Results: []agentworkspace.GenerationResult{
			{ID: "result-1", AssetID: "asset-1", Kind: "image", Asset: map[string]any{"id": "asset-1", "kind": "image"}},
			{ID: "result-2", AssetID: "asset-2", Kind: "image", Asset: map[string]any{"id": "asset-2", "kind": "image"}},
		},
		SelectedResultID: "result-2", ExpectedUpdatedAt: running.UpdatedAt, Now: now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ConfirmGenerationTask(context.Background(), "alice", created.ProjectID, created.ID, "result-2", now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}

	activate := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/generation-1/activate", map[string]any{"resultId": "result-1", "expectedNodeVersion": 2})
	if activate.Code != http.StatusOK || !containsText(activate.Body.String(), `"id":"asset-1"`) || !containsText(activate.Body.String(), `"version":3`) {
		t.Fatalf("activate=%d %s", activate.Code, activate.Body.String())
	}
	stale := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/generation-1/activate", map[string]any{"resultId": "result-2", "expectedNodeVersion": 2})
	if stale.Code != http.StatusConflict || !containsText(stale.Body.String(), "节点版本") {
		t.Fatalf("stale activation=%d %s", stale.Code, stale.Body.String())
	}
}

func TestAgentGenerationCancelAndRetryCreatesLinkedTask(t *testing.T) {
	now := time.Date(2026, 10, 21, 17, 0, 0, 0, time.UTC)
	store, mux := newAgentGenerationLifecycle(t, now)
	created, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "generation-1", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "personal_api", ModelID: "image-model", IdempotencyKey: "image-1-v1", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/"+created.ID+"/cancel", nil)
	if cancel.Code != http.StatusOK || !containsText(cancel.Body.String(), "cancelled") {
		t.Fatalf("cancel=%d %s", cancel.Code, cancel.Body.String())
	}
	retry := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/"+created.ID+"/retry", map[string]any{"idempotencyKey": "image-1-v1-retry"})
	if retry.Code != http.StatusAccepted || !containsText(retry.Body.String(), "generation-1") || !containsText(retry.Body.String(), "queued") {
		t.Fatalf("retry=%d %s", retry.Code, retry.Body.String())
	}
}

func TestAgentGenerationVideoActionsRequireVideoJobRoutes(t *testing.T) {
	now := time.Date(2026, 10, 21, 18, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentGenerationTaskRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func() time.Time { return now })

	queued, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: "video-queued", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "video-queued", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	running, err := store.TransitionGenerationTask(context.Background(), "alice", "project-1", queued.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	awaiting, err := store.TransitionGenerationTask(context.Background(), "alice", "project-1", running.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: []agentworkspace.GenerationResult{{ID: "video-result", AssetID: "video-asset", Kind: "video"}}, SelectedResultID: "video-result", Now: now.Add(2 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}

	confirm := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/"+awaiting.ID+"/confirm", map[string]any{"resultId": "video-result"})
	if confirm.Code != http.StatusConflict || !containsText(confirm.Body.String(), "视频任务") {
		t.Fatalf("confirm=%d %s", confirm.Code, confirm.Body.String())
	}
	cancel := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/"+queued.ID+"/cancel", nil)
	if cancel.Code != http.StatusConflict || !containsText(cancel.Body.String(), "视频任务") {
		t.Fatalf("cancel=%d %s", cancel.Code, cancel.Body.String())
	}

	failed, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: "video-failed", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "video-failed", Now: now.Add(3 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	failed, err = store.TransitionGenerationTask(context.Background(), "alice", "project-1", failed.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: "provider failed", Now: now.Add(4 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	retry := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/generations/"+failed.ID+"/retry", map[string]any{"idempotencyKey": "video-retry"})
	if retry.Code != http.StatusConflict || !containsText(retry.Body.String(), "视频任务") {
		t.Fatalf("retry=%d %s", retry.Code, retry.Body.String())
	}

	for _, id := range []string{queued.ID, awaiting.ID, failed.ID} {
		task, getErr := store.GenerationTask(context.Background(), "alice", "project-1", id)
		if getErr != nil || task.State == agentworkspace.GenerationCompleted || task.State == agentworkspace.GenerationCancelled {
			t.Fatalf("task %s=%+v err=%v", id, task, getErr)
		}
	}
}

func containsText(value, needle string) bool {
	return needle == "" || (json.Valid([]byte(value)) && strings.Contains(value, needle))
}
