package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
)

func newAgentScriptShotTestAPI(t *testing.T, now time.Time, provider batchfactoryv11.DirectorProvider) (*agentworkspace.MemoryStore, *http.ServeMux, *agentScriptSplitCoordinator) {
	t.Helper()
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{
		ID: "script-1", Type: "script", Title: "第一集", Version: 1, Status: "draft",
		Config: map[string]any{"sourceText": "雨夜里，林夏推开旧仓库的门。"},
	}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "短剧", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	services := AgentScriptShotServices{
		Projects: store,
		TextModels: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			if provider == nil {
				return nil, errors.New("model unavailable")
			}
			return provider, nil
		},
		Dispatch: func(context.Context, string, string, agentworkspace.Node, agentworkspace.GenerationTask) error {
			return nil
		},
		Models: func(_ context.Context, owner, modelID, kind string) (map[string]any, error) {
			if owner != "alice" || modelID == "" || kind == "" {
				return nil, errors.New("model unavailable")
			}
			return map[string]any{"id": modelID, "modelId": modelID, "kind": kind, "provider": "personal_api", "enabled": true}, nil
		},
		Now: func() time.Time { return now },
	}
	RegisterAgentScriptShotRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, services)
	return store, mux, newAgentScriptSplitCoordinator(services)
}

func TestScriptSplitPersistsStructuredShotsAndStaysProjectScoped(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	provider := &fixedAgentNodeTextProvider{result: `{"characters":[{"name":"林夏","prompt":"年轻女子"}],"scenes":[{"name":"旧仓库","prompt":"雨夜仓库"}],"props":[],"storyboard":[{"duration_sec":4,"characters":["林夏"],"props":[],"scene":"旧仓库","prefix_key":"modern_conflict","shots":[{"start_sec":0,"end_sec":4,"shot_type":"中景","camera":"缓慢推进","description":"林夏推开仓库门","audio":"林夏：有人吗？"}],"video_desc":"林夏在雨夜进入旧仓库。","visual_prompt":"雨夜旧仓库"}]}`}
	store, mux, worker := newAgentScriptShotTestAPI(t, now, provider)

	split := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/script-splits", map[string]any{"nodeId": "script-1", "nodeVersion": 1, "modelId": "text-model"})
	if split.Code != http.StatusAccepted || !strings.Contains(split.Body.String(), `"state":"queued"`) {
		t.Fatalf("split=%d %s", split.Code, split.Body.String())
	}
	if err := worker.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	shots := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/tasks/project-1/script-nodes/script-1/shots", nil)
	if shots.Code != http.StatusOK || !strings.Contains(shots.Body.String(), `"id":"script-1-shot-1"`) || !strings.Contains(shots.Body.String(), `"imagePrompt":"雨夜旧仓库"`) {
		t.Fatalf("shots=%d %s", shots.Code, shots.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/agent/tasks/project-1/script-nodes/script-1/shots", nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
	durable, err := store.ListShots(context.Background(), "alice", "project-1", "script-1")
	if err != nil || len(durable) != 1 || durable[0].TTSText != "林夏：有人吗？" {
		t.Fatalf("durable=%+v err=%v", durable, err)
	}
}

func TestScriptSplitUsesBatchFactoryDirectorContractWithoutCreatingBatchData(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	provider := &fixedAgentNodeTextProvider{result: `{
  "characters": [{"name":"林夏","prompt":"黑色雨衣的年轻女子"}],
  "scenes": [{"name":"旧仓库","prompt":"雨夜废弃仓库"}],
  "props": [{"name":"手电筒","prompt":"银色手电筒"}],
  "storyboard": [{
    "duration_sec": 6, "characters": ["林夏"], "props": ["手电筒"], "scene": "旧仓库", "prefix_key": "modern_conflict",
    "shots": [
      {"start_sec":0,"end_sec":3,"shot_type":"中景","camera":"缓慢推进","description":"林夏推开铁门","audio":"林夏：有人吗？"},
      {"start_sec":3,"end_sec":6,"shot_type":"近景","camera":"固定","description":"她举起手电筒","audio":"雨声"}
    ], "video_desc": "林夏在雨夜进入旧仓库。", "visual_prompt": "雨夜仓库，冷色电影光"
  }]
}`}
	store, mux, worker := newAgentScriptShotTestAPI(t, now, provider)

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/script-splits", map[string]any{"nodeId": "script-1", "nodeVersion": 1, "modelId": "text-model"})
	if response.Code != http.StatusAccepted {
		t.Fatalf("split=%d %s", response.Code, response.Body.String())
	}
	if err := worker.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	shots, err := store.ListShots(context.Background(), "alice", "project-1", "script-1")
	if err != nil || len(shots) != 1 {
		t.Fatalf("shots=%+v err=%v", shots, err)
	}
	if shots[0].ImagePrompt != "雨夜仓库，冷色电影光" || !strings.Contains(shots[0].VideoPrompt, "00:00-00:03") || !strings.Contains(shots[0].TTSText, "林夏：有人吗？") {
		t.Fatalf("director shot was not preserved: %+v", shots[0])
	}
}

func TestScriptSplitFailureKeepsPreviouslySavedShots(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 30, 0, 0, time.UTC)
	store, mux, worker := newAgentScriptShotTestAPI(t, now, &fixedAgentNodeTextProvider{result: "not-json"})
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "old-shot", Sequence: 1, Scene: "旧镜头"}}); err != nil {
		t.Fatal(err)
	}

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/script-splits", map[string]any{"nodeId": "script-1", "nodeVersion": 1, "modelId": "text-model"})
	if response.Code != http.StatusAccepted {
		t.Fatalf("split=%d %s", response.Code, response.Body.String())
	}
	if err := worker.pollPendingOnce(context.Background()); !errors.Is(err, errScriptSplitUpstream) {
		t.Fatalf("worker err=%v", err)
	}
	shots, err := store.ListShots(context.Background(), "alice", "project-1", "script-1")
	if err != nil || len(shots) != 1 || shots[0].ID != "old-shot" {
		t.Fatalf("shots=%+v err=%v", shots, err)
	}
}

func TestScriptSplitTimeoutMarksGenerationFailedWithUsefulMessage(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: "script-1", Type: "script", Title: "第一集", Version: 1, Status: "draft", Config: map[string]any{"sourceText": "雨夜"}}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "短剧", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	services := AgentScriptShotServices{
		Projects: store,
		Models: func(_ context.Context, _ string, modelID, kind string) (map[string]any, error) {
			return map[string]any{"id": modelID, "modelId": modelID, "kind": kind, "enabled": true}, nil
		},
		TextComplete: func(ctx context.Context, _, _ string, _ batchfactoryv11.TextCompletionRequest) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
		ScriptSplitTimeout: time.Millisecond,
		Now:                func() time.Time { return now },
	}
	RegisterAgentScriptShotRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, services)
	worker := newAgentScriptSplitCoordinator(services)

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/script-splits", map[string]any{"nodeId": "script-1", "nodeVersion": 1, "modelId": "text-model"})
	if response.Code != http.StatusAccepted {
		t.Fatalf("split=%d %s", response.Code, response.Body.String())
	}
	if err := worker.pollPendingOnce(context.Background()); !errors.Is(err, errScriptSplitUpstream) {
		t.Fatalf("worker err=%v", err)
	}
	tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}
	if tasks[0].State != agentworkspace.GenerationFailed || !strings.Contains(tasks[0].ErrorMessage, "文字模型请求超时") {
		t.Fatalf("task=%+v", tasks[0])
	}
}

func TestShotBatchCreatesValidNodesAndSkipsMissingShots(t *testing.T) {
	now := time.Date(2026, 10, 22, 11, 0, 0, 0, time.UTC)
	store, mux, _ := newAgentScriptShotTestAPI(t, now, &fixedAgentNodeTextProvider{})
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "shot-1", Sequence: 1, Scene: "旧仓库", ImagePrompt: "电影感雨夜"}}); err != nil {
		t.Fatal(err)
	}

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", map[string]any{
		"scriptNodeId": "script-1", "shotIds": []string{"shot-1", "missing-shot"}, "kind": "image", "modelId": "image-model",
	})
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"shotId":"missing-shot"`) || !strings.Contains(response.Body.String(), `"operation":"text_to_image"`) {
		t.Fatalf("batch=%d %s", response.Code, response.Body.String())
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || len(canvas.Nodes) != 2 || canvas.Nodes[1].Type != "image" {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
	links, err := store.ListGenerationLinks(context.Background(), "alice", "project-1", "shot-1")
	if err != nil || len(links) != 1 || links[0].NodeID != canvas.Nodes[1].ID || links[0].GenerationID == "" {
		t.Fatalf("links=%+v err=%v", links, err)
	}
	tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(tasks) != 1 || tasks[0].Kind != "image" || tasks[0].NodeID != canvas.Nodes[1].ID {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}

	duplicate := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", map[string]any{
		"scriptNodeId": "script-1", "shotIds": []string{"shot-1", "shot-1"}, "kind": "image", "modelId": "image-model",
	})
	if duplicate.Code != http.StatusBadRequest {
		t.Fatalf("duplicate=%d %s", duplicate.Code, duplicate.Body.String())
	}
}

func TestShotPatchUsesCurrentRevisionAndProjectScope(t *testing.T) {
	now := time.Date(2026, 10, 22, 11, 30, 0, 0, time.UTC)
	store, mux, _ := newAgentScriptShotTestAPI(t, now, &fixedAgentNodeTextProvider{})
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "shot-1", Sequence: 1, Scene: "旧仓库"}}); err != nil {
		t.Fatal(err)
	}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/agent/tasks/project-1/shots/shot-1", map[string]any{
		"scriptNodeId": "script-1", "sequence": 1, "scene": "仓库内景", "revision": 1,
	})
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"revision":2`) {
		t.Fatalf("patch=%d %s", response.Code, response.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPatch, "/api/agent/tasks/project-1/shots/shot-1", map[string]any{"scriptNodeId": "script-1", "sequence": 1, "revision": 2})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
	missingRevision := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/agent/tasks/project-1/shots/shot-1", map[string]any{"scriptNodeId": "script-1", "scene": "不应保存"})
	if missingRevision.Code != http.StatusBadRequest {
		t.Fatalf("missing revision=%d %s", missingRevision.Code, missingRevision.Body.String())
	}
}

type failingShotLinkListStore struct {
	agentworkspace.Store
	err error
}

func (s failingShotLinkListStore) ListGenerationLinks(context.Context, string, string, string) ([]agentworkspace.ShotGenerationLink, error) {
	return nil, s.err
}

type failingAtomicShotStore struct {
	agentworkspace.Store
	err error
}

func (s failingAtomicShotStore) CreateShotGeneration(context.Context, string, agentworkspace.CreateShotGenerationInput) (agentworkspace.Canvas, agentworkspace.GenerationTask, agentworkspace.ShotGenerationLink, error) {
	return agentworkspace.Canvas{}, agentworkspace.GenerationTask{}, agentworkspace.ShotGenerationLink{}, s.err
}

func TestShotBatchDoesNotSuppressLineageReadOrAtomicCreateFailures(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	store, _, _ := newAgentScriptShotTestAPI(t, now, &fixedAgentNodeTextProvider{})
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "shot-1", Sequence: 1, ImagePrompt: "雨夜"}}); err != nil {
		t.Fatal(err)
	}
	register := func(projects agentworkspace.Store) *http.ServeMux {
		mux := http.NewServeMux()
		RegisterAgentScriptShotRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentScriptShotServices{
			Projects: projects,
			Models: func(context.Context, string, string, string) (map[string]any, error) {
				return map[string]any{"id": "image-model", "kind": "image", "provider": "personal_api", "enabled": true}, nil
			},
			Dispatch: func(context.Context, string, string, agentworkspace.Node, agentworkspace.GenerationTask) error {
				return nil
			},
			Now: func() time.Time { return now },
		})
		return mux
	}
	body := map[string]any{"scriptNodeId": "script-1", "shotIds": []string{"shot-1"}, "kind": "image", "modelId": "image-model"}
	listFailure := signedJSONRequest(t, register(failingShotLinkListStore{Store: store, err: errors.New("lineage read failed")}), now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", body)
	if listFailure.Code != http.StatusInternalServerError {
		t.Fatalf("list failure=%d %s", listFailure.Code, listFailure.Body.String())
	}
	atomicFailure := signedJSONRequest(t, register(failingAtomicShotStore{Store: store, err: errors.New("atomic create failed")}), now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", body)
	if atomicFailure.Code != http.StatusInternalServerError {
		t.Fatalf("atomic failure=%d %s", atomicFailure.Code, atomicFailure.Body.String())
	}
	canvas, _ := store.Canvas(context.Background(), "alice", "project-1")
	tasks, _ := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	links, _ := store.ListGenerationLinks(context.Background(), "alice", "project-1", "shot-1")
	if len(canvas.Nodes) != 1 || len(tasks) != 0 || len(links) != 0 {
		t.Fatalf("detached state nodes=%d tasks=%d links=%d", len(canvas.Nodes), len(tasks), len(links))
	}
}

func TestShotBatchUsesExecutableModesAndAttemptsPerOperation(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 30, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: "script-1", Type: "script", Version: 1, Config: map[string]any{"sourceText": "正文"}}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "shot-1", Sequence: 1, ImagePrompt: "图", VideoPrompt: "视频", TTSText: "旁白"}}); err != nil {
		t.Fatal(err)
	}
	var dispatched []agentworkspace.GenerationTask
	mux := http.NewServeMux()
	RegisterAgentScriptShotRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentScriptShotServices{
		Projects: store,
		Models: func(_ context.Context, _ string, modelID, kind string) (map[string]any, error) {
			return map[string]any{"id": modelID, "kind": kind, "provider": "personal_api", "enabled": true}, nil
		},
		Dispatch: func(_ context.Context, _, _ string, _ agentworkspace.Node, task agentworkspace.GenerationTask) error {
			dispatched = append(dispatched, task)
			return nil
		},
		Now: func() time.Time { return now },
	})
	for _, item := range []struct{ kind, mode string }{{"image", "text_to_image"}, {"video", "text_to_video"}, {"audio", "speech"}, {"image", "text_to_image"}} {
		response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", map[string]any{"scriptNodeId": "script-1", "shotIds": []string{"shot-1"}, "kind": item.kind, "modelId": item.kind + "-model"})
		if response.Code != http.StatusOK {
			t.Fatalf("%s batch=%d %s", item.kind, response.Code, response.Body.String())
		}
		last := dispatched[len(dispatched)-1]
		if last.Mode != item.mode || last.State != agentworkspace.GenerationQueued {
			t.Fatalf("kind=%s task=%+v", item.kind, last)
		}
	}
	imageLinks, _ := store.ListGenerationLinks(context.Background(), "alice", "project-1", "shot-1")
	attempts := map[string][]int{}
	for _, link := range imageLinks {
		task, err := store.GenerationTask(context.Background(), "alice", "project-1", link.GenerationID)
		if err != nil {
			t.Fatal(err)
		}
		attempts[task.Kind] = append(attempts[task.Kind], link.Attempt)
	}
	if len(attempts["image"]) != 2 || attempts["image"][0] != 1 || attempts["image"][1] != 2 || attempts["video"][0] != 1 || attempts["audio"][0] != 1 {
		t.Fatalf("attempts=%v", attempts)
	}
}

func TestShotBatchDispatchesIntoExistingImageGenerationRoute(t *testing.T) {
	now := time.Date(2026, 10, 22, 14, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: "script-1", Type: "script", Version: 1, Config: map[string]any{"sourceText": "正文"}}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "shot-1", Sequence: 1, ImagePrompt: "雨夜仓库"}}); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &recordingAgentNodeObjectStore{}
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, auth, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: objects,
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	RegisterAgentScriptShotRoutes(mux, auth, AgentScriptShotServices{
		Projects: store,
		Models: func(context.Context, string, string, string) (map[string]any, error) {
			return map[string]any{"id": "image-model", "kind": "image", "provider": "personal_api", "enabled": true}, nil
		},
		Dispatch: newAgentShotGenerationDispatcher(mux, auth, func() time.Time { return now }, nil),
		Now:      func() time.Time { return now },
	})
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", map[string]any{"scriptNodeId": "script-1", "shotIds": []string{"shot-1"}, "kind": "image", "modelId": "image-model"})
	if response.Code != http.StatusOK {
		t.Fatalf("batch=%d %s", response.Code, response.Body.String())
	}
	tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(tasks) != 1 || tasks[0].State != agentworkspace.GenerationAwaitingConfirmation || len(tasks[0].Results) != 1 {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}
	executions, _ := store.Executions(context.Background(), "alice", "project-1")
	if len(executions) != 1 || executions[0].ToolName != "image-generation" || executions[0].Status != "awaiting_confirmation" {
		t.Fatalf("executions=%+v", executions)
	}
}

func TestShotBatchAcceptsLocalDoubaoImageModelWithoutCatalogEntry(t *testing.T) {
	now := time.Date(2026, 10, 22, 15, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: "script-1", Type: "script", Version: 1, Config: map[string]any{"sourceText": "正文"}}}
	if err := store.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []agentworkspace.ScriptShot{{ID: "shot-1", Sequence: 1, ImagePrompt: "雨夜仓库"}}); err != nil {
		t.Fatal(err)
	}
	var dispatched agentworkspace.GenerationTask
	mux := http.NewServeMux()
	RegisterAgentScriptShotRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentScriptShotServices{
		Projects: store,
		Models: func(context.Context, string, string, string) (map[string]any, error) {
			return nil, errors.New("local Doubao models are not in the account catalog")
		},
		Dispatch: func(_ context.Context, _, _ string, _ agentworkspace.Node, task agentworkspace.GenerationTask) error {
			dispatched = task
			return nil
		},
		Now: func() time.Time { return now },
	})
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/shots/batch-nodes", map[string]any{
		"scriptNodeId": "script-1", "shotIds": []string{"shot-1"}, "kind": "image",
		"provider": "doubao_local_executor", "modelId": "Seedream 4.5",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("batch=%d %s", response.Code, response.Body.String())
	}
	if dispatched.Provider != "doubao_local_executor" || dispatched.ModelID != "Seedream 4.5" {
		t.Fatalf("dispatched provider/model=%q/%q", dispatched.Provider, dispatched.ModelID)
	}
}
