package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/aicut"
)

func TestAICutAgentRoutesCreateReplayListEventsAndCancel(t *testing.T) {
	now := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	timelines := aicut.NewMemoryStore()
	if err := timelines.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	agents := aicut.NewMemoryAgentStore()
	started := make([]string, 0)
	mux := http.NewServeMux()
	RegisterAICutAgentRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, timelines, agents, func(_ context.Context, runID string) error { started = append(started, runID); return nil }, func() time.Time { return now })

	input := map[string]any{"instruction": "删除第二段", "mode": "agent", "modelId": "text-main", "expectedRevision": 0, "clientRequestId": "client-1"}
	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/agent-runs", input)
	if created.Code != http.StatusAccepted || !containsText(created.Body.String(), `"state":"queued"`) {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	var payload struct {
		Run aicut.AgentRun `json:"run"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(started) != 1 || started[0] != payload.Run.ID {
		t.Fatalf("started=%v run=%s", started, payload.Run.ID)
	}

	replayed := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/agent-runs", input)
	if replayed.Code != http.StatusAccepted || !containsText(replayed.Body.String(), payload.Run.ID) || len(started) != 1 {
		t.Fatalf("replay=%d %s started=%v", replayed.Code, replayed.Body.String(), started)
	}

	if _, err := agents.AppendAgentEvent(t.Context(), aicut.AgentEvent{RunID: payload.Run.ID, Owner: "alice", ProjectID: "cut-1", Type: aicut.AgentEventThinking, Message: "分析中", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	events := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/ai-cut/projects/cut-1/agent-runs/"+payload.Run.ID+"/events?after=0", nil)
	if events.Code != http.StatusOK || !containsText(events.Body.String(), `"sequence":1`) {
		t.Fatalf("events=%d %s", events.Code, events.Body.String())
	}

	canceled := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/agent-runs/"+payload.Run.ID+"/cancel", map[string]any{})
	if canceled.Code != http.StatusOK || !containsText(canceled.Body.String(), `"state":"cancel_requested"`) {
		t.Fatalf("cancel=%d %s", canceled.Code, canceled.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/ai-cut/projects/cut-1/agent-runs/"+payload.Run.ID, nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
}

func TestAICutAgentRoutesRejectEmptyInstructionAndMissingProject(t *testing.T) {
	now := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	timelines := aicut.NewMemoryStore()
	agents := aicut.NewMemoryAgentStore()
	mux := http.NewServeMux()
	RegisterAICutAgentRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, timelines, agents, nil, func() time.Time { return now })
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/missing/agent-runs", map[string]any{"instruction": " ", "mode": "agent", "modelId": "text-main", "expectedRevision": 0, "clientRequestId": "client-1"})
	if empty.Code != http.StatusBadRequest || !containsText(empty.Body.String(), "请输入剪辑要求") {
		t.Fatalf("empty=%d %s", empty.Code, empty.Body.String())
	}
	missing := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/missing/agent-runs", map[string]any{"instruction": "粗剪", "mode": "agent", "modelId": "text-main", "expectedRevision": 0, "clientRequestId": "client-2"})
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing=%d %s", missing.Code, missing.Body.String())
	}
}

func TestAICutAgentRoutesAreMountedByApplicationRouter(t *testing.T) {
	now := time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)
	timelines := aicut.NewMemoryStore()
	if err := timelines.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	agents := aicut.NewMemoryAgentStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: timelines, AICutAgentStore: agents})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/ai-cut/projects/cut-1/agent-messages", nil)
	if response.Code != http.StatusOK || !containsText(response.Body.String(), `"messages":[]`) {
		t.Fatalf("messages=%d %s", response.Code, response.Body.String())
	}
}
