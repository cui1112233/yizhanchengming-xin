package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
)

type agentTestProvider struct{ reply string }

func (p agentTestProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	return p.reply, nil
}

func agentRequest(t *testing.T, handler http.Handler, now time.Time, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	SignBridgeRequest(req, "alice", true, now, "secret")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestAgentWorkspaceImportsLegacyAndUsesStoreForCanvasAndChat(t *testing.T) {
	now := time.Date(2026, 10, 18, 3, 0, 0, 0, time.UTC)
	legacyRoot := t.TempDir()
	userDir := filepath.Join(legacyRoot, "alice")
	if err := os.MkdirAll(userDir, 0o700); err != nil {
		t.Fatal(err)
	}
	legacy := `{"version":1,"tasks":[{"id":"legacy-1","title":"旧项目","createdAt":"2026-10-17T01:00:00Z","updatedAt":"2026-10-17T01:00:00Z","messages":[],"canvas":{"version":1,"revision":0,"nodes":[],"edges":[]},"executions":[]}]}`
	if err := os.WriteFile(filepath.Join(userDir, "agent-tasks.json"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	store := agentworkspace.NewMemoryStore()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, legacyRoot, agentTestProvider{reply: "已收到节点引用"}, func() time.Time { return now })

	list := agentRequest(t, mux, now, http.MethodGet, "/api/agent/tasks", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "legacy-1") {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	created := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks", "")
	if created.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var createBody struct {
		Task agentworkspace.Project `json:"task"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createBody); err != nil || createBody.Task.ID == "" {
		t.Fatalf("create body=%s err=%v", created.Body.String(), err)
	}

	canvasBody := `{"canvas":{"version":1,"revision":0,"appearance":{"edgeColor":"cyan","edgeAnimated":false,"edgeWidth":"thick","backgroundDots":false,"alignmentGuides":true,"edgesVisible":false},"nodes":[{"id":"node-1","type":"input","position":{"x":12,"y":24},"title":"小说原文","version":1,"sourceNodeIds":[],"status":"confirmed","summary":"第一章"}],"edges":[]},"revision":0}`
	canvasPath := "/api/agent/tasks/" + createBody.Task.ID + "/canvas"
	saved := agentRequest(t, mux, now, http.MethodPatch, canvasPath, canvasBody)
	if saved.Code != http.StatusOK || !strings.Contains(saved.Body.String(), `"revision":1`) || !strings.Contains(saved.Body.String(), `"edgeColor":"cyan"`) || !strings.Contains(saved.Body.String(), `"edgesVisible":false`) {
		t.Fatalf("save status=%d body=%s", saved.Code, saved.Body.String())
	}
	loaded := agentRequest(t, mux, now, http.MethodGet, canvasPath, "")
	if loaded.Code != http.StatusOK || !strings.Contains(loaded.Body.String(), `"edgeColor":"cyan"`) || !strings.Contains(loaded.Body.String(), `"edgesVisible":false`) {
		t.Fatalf("load status=%d body=%s", loaded.Code, loaded.Body.String())
	}
	stale := agentRequest(t, mux, now, http.MethodPatch, canvasPath, canvasBody)
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale status=%d body=%s", stale.Code, stale.Body.String())
	}

	chatBody := `{"taskId":"` + createBody.Task.ID + `","prompt":"修改这个节点","context":{"selectedNodes":[{"id":"node-1","title":"小说原文","summary":"第一章"}]}}`
	chat := agentRequest(t, mux, now, http.MethodPost, "/api/agent/chat", chatBody)
	if chat.Code != http.StatusOK {
		t.Fatalf("chat status=%d body=%s", chat.Code, chat.Body.String())
	}
	project, err := store.Project(context.Background(), "alice", createBody.Task.ID)
	if err != nil || len(project.Messages) != 2 {
		t.Fatalf("project=%+v err=%v", project, err)
	}
	selected, ok := project.Messages[0].Context["selectedNodes"].([]any)
	if !ok || len(selected) != 1 {
		t.Fatalf("context=%+v", project.Messages[0].Context)
	}
}

func TestAgentWorkspaceCanvasReadAddsSafeAssetContentURL(t *testing.T) {
	now := time.Date(2026, 10, 20, 4, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{ID: "agent-1", Title: "视频项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.EmptyCanvas()}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{
		ID: "video-1", Type: "video", Title: "成片", Version: 2, Status: "confirmed",
		AssetRef: map[string]any{
			"id": "asset-1", "kind": "video", "name": "成片.mp4",
			"objectKey": "private/alice/agent-1/asset-1.mp4",
		},
	}}
	if _, err := store.SaveCanvas(context.Background(), "alice", project.ID, canvas, 0, now); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, "", nil, func() time.Time { return now })
	loaded := agentRequest(t, mux, now, http.MethodGet, "/api/agent/tasks/agent-1/canvas", "")
	if loaded.Code != http.StatusOK {
		t.Fatalf("load status=%d body=%s", loaded.Code, loaded.Body.String())
	}
	var body struct {
		Canvas agentworkspace.Canvas `json:"canvas"`
	}
	if err := json.Unmarshal(loaded.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	ref := body.Canvas.Nodes[0].AssetRef
	if ref["url"] != "/api/agent/tasks/agent-1/assets/asset-1/content" {
		t.Fatalf("asset url=%v", ref["url"])
	}
	if _, exists := ref["objectKey"]; exists {
		t.Fatalf("private object key leaked: %+v", ref)
	}
}
