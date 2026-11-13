package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"strings"
	"testing"
	"time"
)

func TestCanvasPetChatAndApprovalPersistActualEditAndRejectReplay(t *testing.T) {
	now := time.Now().UTC()
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{ID: "canvas-pet", Title: "画布验收", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "n1", Type: "text", Title: "原故事", Version: 1, Status: "draft"}}}}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	provider := &recordingAgentTestProvider{reply: `{"reply":"请确认","plan":{"id":"plan-pet","summary":"修改标题","commands":[{"type":"update_node","nodeId":"n1","patch":{"title":"新故事"}}]}}`}
	mux := http.NewServeMux()
	RegisterAgentWorkspaceCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, "", provider, func() time.Time { return now })
	chat := agentRequest(t, mux, now, http.MethodPost, "/api/agent/chat", `{"taskId":"canvas-pet","prompt":"改标题","context":{"canvasPet":{"revision":0,"selectedNodeIds":["n1"]}}}`)
	if chat.Code != 200 || !strings.Contains(provider.request.UserPrompt, "原故事") {
		t.Fatalf("chat=%d %s request=%s", chat.Code, chat.Body.String(), provider.request.UserPrompt)
	}
	result := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/canvas-pet/director/plans/plan-pet/approve", `{"revision":0}`)
	if result.Code != 200 {
		t.Fatalf("apply=%d %s", result.Code, result.Body.String())
	}
	stored, _ := store.Project(context.Background(), "alice", project.ID)
	if stored.Canvas.Nodes[0].Title != "新故事" {
		t.Fatalf("no real edit: %#v", stored.Canvas)
	}
	replay := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/canvas-pet/director/plans/plan-pet/approve", fmt.Sprintf(`{"revision":%d}`, stored.Canvas.Revision))
	if replay.Code != http.StatusConflict {
		t.Fatalf("replay=%d %s", replay.Code, replay.Body.String())
	}
	invalid := agentRequest(t, mux, now, http.MethodPost, "/api/agent/chat", `{"taskId":"canvas-pet","prompt":"x","context":{"canvasPet":{"revision":0,"selectedNodeIds":["n1"]}}}`)
	if invalid.Code != http.StatusConflict {
		t.Fatalf("stale request=%d", invalid.Code)
	}
	legacyPlan := agentworkspace.DirectorPlan{ID: "already-applied", Status: "applied", Summary: "旧自动方案", Commands: []agentworkspace.DirectorCommand{{Type: "update_node", NodeID: "n1", Patch: map[string]any{"title": "旧值"}}}}
	_, err := store.AppendExchange(context.Background(), "alice", project.ID, agentworkspace.Message{Role: "user", Content: "历史请求", CreatedAt: now}, agentworkspace.Message{Role: "assistant", Content: "已经执行", CreatedAt: now, Context: map[string]any{"directorPlan": legacyPlan}})
	if err != nil {
		t.Fatal(err)
	}
	legacy := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/canvas-pet/director/plans/already-applied/approve", fmt.Sprintf(`{"revision":%d}`, stored.Canvas.Revision))
	if legacy.Code != http.StatusConflict {
		t.Fatalf("applied legacy plan accepted: %d %s", legacy.Code, legacy.Body.String())
	}
}

func TestCanvasPetContextReadsOwnedSavedNodes(t *testing.T) {
	project := agentworkspace.Project{Canvas: agentworkspace.Canvas{Revision: 4, Nodes: []agentworkspace.Node{{ID: "n1", Type: "text", Title: "真实标题", Summary: "真实正文"}}}}
	input := map[string]any{"canvasPet": map[string]any{"revision": float64(4), "selectedNodeIds": []any{"n1"}}, "selectedNodes": []any{map[string]any{"summary": "伪造正文"}}}
	got, err := prepareCanvasPetContext(project, input)
	if err != nil {
		t.Fatal(err)
	}
	nodes := got["selectedNodes"].([]any)
	if !strings.Contains(nodes[0].(map[string]any)["summary"].(string), "真实正文") {
		t.Fatalf("not server nodes: %#v", got)
	}
	if got["director"].(map[string]any)["mode"] != "supervised" {
		t.Fatalf("must supervise: %#v", got)
	}
}

func TestCanvasPetContextRejectsStaleOrForeignSelection(t *testing.T) {
	project := agentworkspace.Project{Canvas: agentworkspace.Canvas{Revision: 4, Nodes: []agentworkspace.Node{{ID: "n1", Type: "text"}}}}
	for _, pet := range []map[string]any{{"revision": float64(3), "selectedNodeIds": []any{"n1"}}, {"revision": float64(4), "selectedNodeIds": []any{"foreign"}}, {"revision": float64(4.5), "selectedNodeIds": []any{}}} {
		if _, err := prepareCanvasPetContext(project, map[string]any{"canvasPet": pet}); err == nil {
			t.Fatalf("accepted %#v", pet)
		}
	}
}
