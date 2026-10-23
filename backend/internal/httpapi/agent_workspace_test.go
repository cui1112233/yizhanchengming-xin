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

type recordingAgentTestProvider struct {
	reply   string
	request batchfactoryv11.TextCompletionRequest
}

func (p *recordingAgentTestProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.request = request
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

func TestAgentWorkspaceCanvasReadAddsSafeAudioContentURL(t *testing.T) {
	now := time.Date(2026, 10, 21, 4, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{ID: "agent-audio-1", Title: "音频项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.EmptyCanvas()}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{
		ID: "audio-1", Type: "audio", Title: "配音", Version: 2, Status: "confirmed",
		AssetRef: map[string]any{
			"id": "asset-audio-1", "kind": "audio", "name": "speech.mp3",
			"objectKey": "private/alice/agent-audio-1/asset-audio-1.mp3",
		},
	}}
	if _, err := store.SaveCanvas(context.Background(), "alice", project.ID, canvas, 0, now); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, "", nil, func() time.Time { return now })
	loaded := agentRequest(t, mux, now, http.MethodGet, "/api/agent/tasks/agent-audio-1/canvas", "")
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
	if ref["url"] != "/api/agent/tasks/agent-audio-1/assets/asset-audio-1/content" {
		t.Fatalf("audio asset url=%v", ref["url"])
	}
	if _, exists := ref["objectKey"]; exists {
		t.Fatalf("private object key leaked: %+v", ref)
	}
}

func TestAgentWorkspaceApprovesOnlyTheSavedDirectorPlan(t *testing.T) {
	now := time.Date(2026, 10, 21, 5, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{
		ID: "director-project", Title: "导演项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{
			ID: "text-1", Type: "text", Position: agentworkspace.Position{X: 24, Y: 36}, Title: "原始文本", Version: 1, Status: "draft",
		}}},
		Messages: []agentworkspace.Message{{
			Role: "assistant", Content: "我整理了一个画布修改方案。", CreatedAt: now,
			Context: map[string]any{"directorPlan": map[string]any{
				"id": "plan-1", "summary": "更新开场文本", "commands": []any{map[string]any{
					"type": "update_node", "nodeId": "text-1", "patch": map[string]any{"title": "雨夜开场"},
				}},
			}},
		}},
	}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, "", nil, func() time.Time { return now })

	approved := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/director-project/director/plans/plan-1/approve", `{"revision":0}`)
	if approved.Code != http.StatusOK || !strings.Contains(approved.Body.String(), `"雨夜开场"`) || !strings.Contains(approved.Body.String(), `"revision":1`) {
		t.Fatalf("approve status=%d body=%s", approved.Code, approved.Body.String())
	}
	missing := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/director-project/director/plans/other/approve", `{"revision":1}`)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing plan status=%d body=%s", missing.Code, missing.Body.String())
	}
	stale := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/director-project/director/plans/plan-1/approve", `{"revision":0}`)
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale status=%d body=%s", stale.Code, stale.Body.String())
	}
}

func TestAgentWorkspaceDirectorApprovalStopsOperationalAndForbiddenPlans(t *testing.T) {
	now := time.Date(2026, 10, 21, 5, 10, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{
		ID: "director-safety", Title: "导演安全", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{
			ID: "image-1", Type: "image", Position: agentworkspace.Position{X: 24, Y: 36}, Title: "开场图", Version: 1, Status: "draft", Config: map[string]any{"modelId": "configured-image"},
		}}},
		Messages: []agentworkspace.Message{
			{Role: "assistant", Content: "生成方案", CreatedAt: now, Context: map[string]any{"directorPlan": map[string]any{
				"id": "operational-plan", "commands": []any{map[string]any{"type": "start_generation", "nodeId": "image-1"}},
			}}},
			{Role: "assistant", Content: "删除方案", CreatedAt: now, Context: map[string]any{"directorPlan": map[string]any{
				"id": "forbidden-plan", "commands": []any{map[string]any{"type": "delete_asset", "nodeId": "asset-1"}},
			}}},
		},
	}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, "", nil, func() time.Time { return now })

	operational := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/director-safety/director/plans/operational-plan/approve", `{"revision":0}`)
	if operational.Code != http.StatusUnprocessableEntity {
		t.Fatalf("operational status=%d body=%s", operational.Code, operational.Body.String())
	}
	forbidden := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/director-safety/director/plans/forbidden-plan/approve", `{"revision":0}`)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("forbidden status=%d body=%s", forbidden.Code, forbidden.Body.String())
	}
}

func TestAgentWorkspaceDirectorChatStoresOnlyValidatedPlan(t *testing.T) {
	now := time.Date(2026, 10, 21, 5, 20, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{
		ID: "director-chat", Title: "导演对话", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{
			ID: "text-1", Type: "text", Position: agentworkspace.Position{X: 20, Y: 20}, Title: "原始开场", Version: 1, Status: "draft",
		}}},
	}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	provider := &recordingAgentTestProvider{reply: `{"reply":"我整理了开场修改方案，等待你确认。","plan":{"summary":"调整开场标题","commands":[{"type":"update_node","nodeId":"text-1","patch":{"title":"雨夜开场"}}]}}`}
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, "", provider, func() time.Time { return now })

	chat := agentRequest(t, mux, now, http.MethodPost, "/api/agent/chat", `{"taskId":"director-chat","prompt":"把开场改成雨夜","context":{"director":{"enabled":true,"mode":"supervised"}}}`)
	if chat.Code != http.StatusOK || !strings.Contains(chat.Body.String(), "等待你确认") || !strings.Contains(chat.Body.String(), "directorPlan") {
		t.Fatalf("chat status=%d body=%s", chat.Code, chat.Body.String())
	}
	if provider.request.DisableJSONResponse || !strings.Contains(provider.request.SystemPrompt, "待用户确认") {
		t.Fatalf("director provider request=%+v", provider.request)
	}
	stored, err := store.Project(context.Background(), "alice", project.ID)
	if err != nil || len(stored.Messages) != 2 {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	if _, ok := stored.Messages[1].Context["directorPlan"]; !ok {
		t.Fatalf("assistant context=%+v", stored.Messages[1].Context)
	}
}

func TestAgentWorkspaceAutomaticDirectorAppliesOnlyValidatedStructuralPlan(t *testing.T) {
	now := time.Date(2026, 10, 21, 5, 25, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{
		ID: "director-auto", Title: "自动导演", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{
			ID: "text-1", Type: "text", Position: agentworkspace.Position{X: 20, Y: 20}, Title: "原始开场", Version: 1, Status: "draft",
		}}},
	}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	provider := &recordingAgentTestProvider{reply: `{"reply":"我将调整开场。","plan":{"summary":"调整开场标题","commands":[{"type":"update_node","nodeId":"text-1","patch":{"title":"雨夜开场"}}]}}`}
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, store, "", provider, func() time.Time { return now })

	chat := agentRequest(t, mux, now, http.MethodPost, "/api/agent/chat", `{"taskId":"director-auto","prompt":"全自动完成开场调整","context":{"director":{"enabled":true,"mode":"automatic"}}}`)
	if chat.Code != http.StatusOK || !strings.Contains(chat.Body.String(), `"status":"applied"`) || !strings.Contains(chat.Body.String(), `"title":"雨夜开场"`) {
		t.Fatalf("chat status=%d body=%s", chat.Code, chat.Body.String())
	}
	canvas, err := store.Canvas(context.Background(), "alice", project.ID)
	if err != nil || canvas.Revision != 1 || canvas.Nodes[0].Title != "雨夜开场" {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
}
