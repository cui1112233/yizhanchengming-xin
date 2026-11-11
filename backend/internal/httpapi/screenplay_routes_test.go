package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

type screenplayProviderFunc func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error)

func (fn screenplayProviderFunc) Complete(ctx context.Context, req batchfactoryv11.TextCompletionRequest) (string, error) {
	return fn(ctx, req)
}

func TestScreenplayRoutesProjectGenerateAndWorkspace(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	mux := http.NewServeMux()
	provider := screenplayProviderFunc(func(_ context.Context, _ batchfactoryv11.TextCompletionRequest) (string, error) {
		return "# 第1集：开始创作\n\n## 1-1：日 内 客厅\n人物：小Y\n\n△ 小Y打开电脑。\n小Y：开始创作漫剧。", nil
	})
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })
	prompts := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/comic-screenplays/prompts", nil)
	if prompts.Code != http.StatusOK || !containsText(prompts.Body.String(), "专业剧本") || containsText(prompts.Body.String(), "SystemPrompt") {
		t.Fatalf("prompts=%d %s", prompts.Code, prompts.Body.String())
	}

	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"id": "test-project", "title": "测试漫剧", "aspectRatio": "9:16", "modelId": "text-own", "promptId": "a1_novel_to_screenplay", "assetDegree": "master", "permissionMode": "bypass"})
	if created.Code != http.StatusCreated || !containsText(created.Body.String(), "测试漫剧") || !containsText(created.Body.String(), `"model_id":"text-own"`) || !containsText(created.Body.String(), `"prompt_id":"a1_novel_to_screenplay"`) {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	if list := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/comic-screenplays/projects", nil); list.Code != http.StatusOK || !containsText(list.Body.String(), "测试漫剧") {
		t.Fatalf("list=%d %s", list.Code, list.Body.String())
	}
	generated := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/test-project/generate", map[string]any{"mode": "a1", "content": "一段小说", "model": "demo", "confirmed": true})
	if generated.Code != http.StatusOK || !containsText(generated.Body.String(), `"episode_count":1`) {
		t.Fatalf("generate=%d %s", generated.Code, generated.Body.String())
	}
	workspace := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/comic-screenplays/projects/test-project/workspace", nil)
	if workspace.Code != http.StatusOK || !containsText(workspace.Body.String(), "开始创作漫剧") {
		t.Fatalf("workspace=%d %s", workspace.Code, workspace.Body.String())
	}
	started := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/test-project/batch", map[string]any{})
	if started.Code != http.StatusCreated || !containsText(started.Body.String(), `"status":"running"`) {
		t.Fatalf("batch start=%d %s", started.Code, started.Body.String())
	}
	paused := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/test-project/batch/cancel", map[string]any{})
	if paused.Code != http.StatusOK || !containsText(paused.Body.String(), `"status":"paused"`) {
		t.Fatalf("batch pause=%d %s", paused.Code, paused.Body.String())
	}
	resumed := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/test-project/batch", map[string]any{})
	if resumed.Code != http.StatusOK || !containsText(resumed.Body.String(), `"status":"running"`) {
		t.Fatalf("batch resume=%d %s", resumed.Code, resumed.Body.String())
	}
	exported := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/comic-screenplays/projects/test-project/export", nil)
	if exported.Code != http.StatusOK || !containsText(exported.Body.String(), `"format":"yizhanchengming-comic-screenplay-v1"`) {
		t.Fatalf("export=%d %s", exported.Code, exported.Body.String())
	}
}

func TestScreenplayRoutesRejectEmptyDuplicateAndForeignOwner(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"title": " "})
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("empty=%d %s", empty.Code, empty.Body.String())
	}
	first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"id": "same", "title": "第一版"})
	if first.Code != http.StatusCreated {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}
	duplicate := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"id": "same", "title": "重复"})
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate=%d %s", duplicate.Code, duplicate.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/comic-screenplays/projects/same/workspace", nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
}

func TestScreenplayRoutesShotCreateUpdateMoveDelete(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	mux := http.NewServeMux()
	provider := screenplayProviderFunc(func(_ context.Context, _ batchfactoryv11.TextCompletionRequest) (string, error) {
		return "# 第1集：测试\n\n## 1-1：日 内 客厅\n人物：小Y\n\n△ 第一场。", nil
	})
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }, func() time.Time { return now })

	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"id": "shot-project", "title": "镜头操作"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create project=%d %s", created.Code, created.Body.String())
	}
	generated := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/shot-project/generate", map[string]any{"mode": "a1", "content": "测试", "model": "demo", "confirmed": true})
	if generated.Code != http.StatusOK {
		t.Fatalf("generate=%d %s", generated.Code, generated.Body.String())
	}

	added := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/shot-project/shots", map[string]any{
		"episodeNo": 1, "heading": "1-2：夜 外 天台", "scriptText": "新增的一场", "characters": []string{"小Y"},
	})
	if added.Code != http.StatusCreated || !containsText(added.Body.String(), "新增的一场") {
		t.Fatalf("add shot=%d %s", added.Code, added.Body.String())
	}
	var addedPayload struct {
		Shot screenplay.Shot `json:"shot"`
	}
	if err := json.Unmarshal(added.Body.Bytes(), &addedPayload); err != nil {
		t.Fatalf("decode added shot: %v", err)
	}
	if addedPayload.Shot.ID == "" || addedPayload.Shot.SortOrder != 2 {
		t.Fatalf("unexpected added shot: %+v", addedPayload.Shot)
	}

	updated := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/comic-screenplays/projects/shot-project/shots/"+addedPayload.Shot.ID, map[string]any{
		"episodeNo": 1, "heading": "1-2：夜 外 天台", "scriptText": "编辑后的内容", "characters": []string{"小Y", "CM"},
	})
	if updated.Code != http.StatusOK || !containsText(updated.Body.String(), "编辑后的内容") {
		t.Fatalf("update shot=%d %s", updated.Code, updated.Body.String())
	}

	moved := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/shot-project/shots/"+addedPayload.Shot.ID+"/move", map[string]any{"direction": "up"})
	if moved.Code != http.StatusOK || !containsText(moved.Body.String(), `"sort_order":1`) {
		t.Fatalf("move shot=%d %s", moved.Code, moved.Body.String())
	}
	workspace := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/comic-screenplays/projects/shot-project/workspace", nil)
	if workspace.Code != http.StatusOK || !containsText(workspace.Body.String(), "编辑后的内容") {
		t.Fatalf("workspace after edit=%d %s", workspace.Code, workspace.Body.String())
	}

	deleted := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, "/api/comic-screenplays/projects/shot-project/shots/"+addedPayload.Shot.ID, nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete shot=%d %s", deleted.Code, deleted.Body.String())
	}
	repeated := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, "/api/comic-screenplays/projects/shot-project/shots/"+addedPayload.Shot.ID, nil)
	if repeated.Code != http.StatusNotFound {
		t.Fatalf("repeat delete=%d %s", repeated.Code, repeated.Body.String())
	}
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/shot-project/shots", map[string]any{"episodeNo": 1, "heading": " ", "scriptText": " "})
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("empty shot=%d %s", empty.Code, empty.Body.String())
	}
}

func TestScreenplayRoutesAssetCreateUpdateDelete(t *testing.T) {
	now := time.Date(2026, 11, 11, 11, 0, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	createdProject := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"id": "asset-project", "title": "资产操作"})
	if createdProject.Code != http.StatusCreated {
		t.Fatalf("create project=%d %s", createdProject.Code, createdProject.Body.String())
	}

	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/asset-project/assets", map[string]any{
		"type": "character", "name": "小Y", "identity": "女主角", "description": "短发，行动果断", "prompt": "成年女性角色，短发",
	})
	if created.Code != http.StatusCreated || !containsText(created.Body.String(), "小Y") {
		t.Fatalf("create asset=%d %s", created.Code, created.Body.String())
	}
	var payload struct {
		Asset screenplay.Asset `json:"asset"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Asset.ID == "" || payload.Asset.Revision != 1 {
		t.Fatalf("asset=%+v", payload.Asset)
	}

	duplicate := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/asset-project/assets", map[string]any{"type": "character", "name": "小Y"})
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("duplicate=%d %s", duplicate.Code, duplicate.Body.String())
	}
	updated := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/comic-screenplays/projects/asset-project/assets/"+payload.Asset.ID, map[string]any{
		"type": "character", "name": "小Y", "identity": "核心女主角", "description": "短发，行动果断", "prompt": "写实成年女性角色，短发", "expectedRevision": 1,
	})
	if updated.Code != http.StatusOK || !containsText(updated.Body.String(), "核心女主角") || !containsText(updated.Body.String(), `"revision":2`) {
		t.Fatalf("update=%d %s", updated.Code, updated.Body.String())
	}
	conflict := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/comic-screenplays/projects/asset-project/assets/"+payload.Asset.ID, map[string]any{"type": "character", "name": "旧数据覆盖", "expectedRevision": 1})
	if conflict.Code != http.StatusConflict {
		t.Fatalf("revision conflict=%d %s", conflict.Code, conflict.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodDelete, "/api/comic-screenplays/projects/asset-project/assets/"+payload.Asset.ID, nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign delete=%d %s", foreign.Code, foreign.Body.String())
	}
	deleted := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, "/api/comic-screenplays/projects/asset-project/assets/"+payload.Asset.ID, nil)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete=%d %s", deleted.Code, deleted.Body.String())
	}
	repeated := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, "/api/comic-screenplays/projects/asset-project/assets/"+payload.Asset.ID, nil)
	if repeated.Code != http.StatusNotFound {
		t.Fatalf("repeat delete=%d %s", repeated.Code, repeated.Body.String())
	}
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects/asset-project/assets", map[string]any{"type": "scene", "name": " "})
	if empty.Code != http.StatusBadRequest {
		t.Fatalf("empty=%d %s", empty.Code, empty.Body.String())
	}
}
