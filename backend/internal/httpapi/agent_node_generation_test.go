package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

type recordingAgentNodeObjectStore struct {
	putAsset    agentworkspace.Asset
	putContents []byte
	deleted     []string
}

type failingAgentNodeAssetStore struct{ err error }

func (s failingAgentNodeAssetStore) ListAssets(context.Context, string, string, string) ([]agentworkspace.Asset, error) {
	return nil, s.err
}

func (s failingAgentNodeAssetStore) CreateAsset(context.Context, string, agentworkspace.Asset) error {
	return s.err
}

func (s failingAgentNodeAssetStore) Asset(context.Context, string, string, string) (agentworkspace.Asset, error) {
	return agentworkspace.Asset{}, s.err
}

func (s *recordingAgentNodeObjectStore) Put(_ context.Context, asset agentworkspace.Asset, body io.Reader) error {
	s.putAsset = asset
	s.putContents, _ = io.ReadAll(body)
	return nil
}

func (s *recordingAgentNodeObjectStore) Open(context.Context, agentworkspace.Asset) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.putContents)), nil
}

func (s *recordingAgentNodeObjectStore) Delete(_ context.Context, asset agentworkspace.Asset) error {
	s.deleted = append(s.deleted, asset.ID)
	return nil
}

type fixedAgentNodeTextProvider struct {
	result string
	err    error
	input  *batchfactoryv11.TextCompletionRequest
}

func (p *fixedAgentNodeTextProvider) Complete(_ context.Context, input batchfactoryv11.TextCompletionRequest) (string, error) {
	if p.input != nil {
		*p.input = input
	}
	return p.result, p.err
}

func newAgentNodeGenerationTestStore(t *testing.T, now time.Time, node agentworkspace.Node) *agentworkspace.MemoryStore {
	t.Helper()
	store := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{
		Version: 1, Appearance: agentworkspace.DefaultAppearance(), Nodes: []agentworkspace.Node{node}, Edges: []agentworkspace.Edge{},
	}}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestAgentImageGenerationRejectsDisabledSelectedCatalogModel(t *testing.T) {
	now := time.Date(2026, 10, 23, 18, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Version: 1, Status: "draft"})
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "image-disabled", "kind": "image", "enabled": false},
	}}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Catalog: catalog, Assets: agentworkspace.NewMemoryAssetStore(), QueueImages: true, Now: func() time.Time { return now },
	})
	response := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-disabled","prompt":"雨夜街道","idempotencyKey":"disabled-model"}`)
	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "当前模型不可用") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAgentTextGenerationUsesRequestedOwnerModelAndAppliesResult(t *testing.T) {
	now := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "text-1", Type: "text", Title: "故事", Summary: "旧内容", Version: 1, Status: "draft"})
	var request batchfactoryv11.TextCompletionRequest
	provider := &fixedAgentNodeTextProvider{result: "生成后的正文", input: &request}
	resolver := func(_ context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || modelID != "text-secondary" {
			t.Fatalf("owner=%q modelID=%q", owner, modelID)
		}
		return provider, nil
	}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, TextModels: resolver, Now: func() time.Time { return now },
	})
	body := `{"nodeId":"text-1","nodeVersion":1,"modelId":"text-secondary","prompt":"写一个雨夜故事","idempotencyKey":"text-request-1"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/text-generations", body)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"summary":"生成后的正文"`) || !strings.Contains(recorder.Body.String(), `"revision":1`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if request.UserPrompt != "写一个雨夜故事" || !request.DisableJSONResponse || request.SystemPrompt == "" {
		t.Fatalf("provider request=%+v", request)
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || canvas.Nodes[0].Summary != "生成后的正文" || canvas.Nodes[0].Version != 2 {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
	generations, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	definition, _ := generations[0].Settings["promptDefinition"].(map[string]any)
	if err != nil || len(generations) != 1 || generations[0].State != agentworkspace.GenerationCompleted || generations[0].Kind != "text" || definition["id"] != "agent.text-node" || fmt.Sprint(definition["version"]) != "1" {
		t.Fatalf("generations=%+v err=%v", generations, err)
	}
}

func TestNormalizeAgentImageGenerationInputUsesNestedNodeSettings(t *testing.T) {
	input, err := normalizeAgentImageGenerationInput(agentImageGenerationRequest{
		Provider: "doubao_local_executor", ModelID: "Seedream 5.0 Pro", Prompt: "电影感雨夜",
		Mode:       "image_to_image",
		Parameters: map[string]any{"ratio": "9:16", "quality": "high", "resolution": "4K", "count": float64(2)},
		Advanced:   map[string]any{"seed": float64(42)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if input.Provider != "doubao_local_executor" || input.ModelID != "Seedream 5.0 Pro" || input.AspectRatio != "9:16" || input.Quality != "high" || input.Resolution != "4K" || input.Count != 2 || input.Size != "2304x4096" {
		t.Fatalf("normalized input=%+v", input)
	}
}

func TestNormalizeAgentImageGenerationInputSupportsPortraitAndLandscapeRatios(t *testing.T) {
	cases := []struct{ ratio, size string }{
		{"2:3", "768x1152"},
		{"3:4", "768x1024"},
		{"4:3", "1024x768"},
		{"3:2", "1152x768"},
		{"21:9", "1344x576"},
	}
	for _, tc := range cases {
		input, err := normalizeAgentImageGenerationInput(agentImageGenerationRequest{
			Provider: "doubao_local_executor", ModelID: "Seedream 5.0 Pro", Prompt: "海报",
			Mode:       "text_to_image",
			Parameters: map[string]any{"ratio": tc.ratio, "quality": "standard", "resolution": "1K", "count": float64(1)},
		})
		if err != nil {
			t.Fatalf("%s: %v", tc.ratio, err)
		}
		if input.AspectRatio != tc.ratio || input.Size != tc.size {
			t.Fatalf("%s normalized input=%+v", tc.ratio, input)
		}
	}
}

func TestAgentImageGenerationRejectsUnsupportedNestedCapabilityValue(t *testing.T) {
	now := time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "图片", Version: 1, Status: "draft"})
	calls := 0
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: agentworkspace.NewMemoryAssetStore(),
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			calls++
			return nil, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"provider":"doubao_local_executor","modelId":"Seedream 5.0 Pro","prompt":"电影感雨夜","mode":"text_to_image","parameters":{"ratio":"16:9","quality":"standard","resolution":"8K","count":1},"idempotencyKey":"image-invalid-resolution"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "resolution") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if calls != 0 {
		t.Fatalf("image generator calls=%d, want 0", calls)
	}
}

func TestAgentImageGenerationStoresResolvedCanvasInputSnapshot(t *testing.T) {
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "text-1", Type: "text", Title: "设定", Summary: "雨夜街道", Version: 1, Status: "confirmed"})
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes = append(canvas.Nodes, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft", Config: map[string]any{"prompt": "霓虹倒影"}})
	canvas.Edges = []agentworkspace.Edge{{ID: "edge-text", Source: "text-1", Target: "image-1"}}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, canvas.Revision, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: agentworkspace.NewMemoryAssetStore(), QueueImages: true, Now: func() time.Time { return now },
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			return nil, http.StatusOK, nil
		},
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道\n霓虹倒影","parameters":{"ratio":"16:9","quality":"standard","resolution":"2K","count":1},"idempotencyKey":"image-resolved-inputs"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}
	snapshot, ok := tasks[0].Settings["resolvedInputs"].(map[string]any)
	if !ok {
		t.Fatalf("resolvedInputs=%#v", tasks[0].Settings["resolvedInputs"])
	}
	if got, want := snapshot["prompt"], "雨夜街道\n霓虹倒影"; got != want {
		t.Fatalf("snapshot prompt=%#v want=%q", got, want)
	}
	if got, want := snapshot["sourceNodeIds"], []any{"text-1", "image-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("snapshot sources=%#v want=%#v", got, want)
	}
}

func TestAgentTextGenerationFailureDoesNotChangeNode(t *testing.T) {
	now := time.Date(2026, 10, 20, 9, 10, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "text-1", Type: "text", Title: "故事", Summary: "保留内容", Version: 1, Status: "draft"})
	provider := &fixedAgentNodeTextProvider{err: errors.New("provider unavailable")}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects:   store,
		TextModels: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil },
		Now:        func() time.Time { return now },
	})
	body := `{"nodeId":"text-1","nodeVersion":1,"modelId":"text-primary","prompt":"继续写","idempotencyKey":"text-request-failed"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/text-generations", body)
	if recorder.Code != http.StatusBadGateway || !strings.Contains(recorder.Body.String(), "文字模型生成失败") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	project, err := store.Project(context.Background(), "alice", "project-1")
	if err != nil || project.Canvas.Nodes[0].Summary != "保留内容" || project.Canvas.Nodes[0].Version != 1 {
		t.Fatalf("project=%+v err=%v", project, err)
	}
	if len(project.Executions) != 1 || project.Executions[0].Status != "failed" {
		t.Fatalf("executions=%+v", project.Executions)
	}
}

func TestAgentImageGenerationPersistsAssetBeforeAwaitingConfirmation(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &recordingAgentNodeObjectStore{}
	imageGenerator := func(_ context.Context, owner string, input AgentImageGenerationInput) ([]string, int, error) {
		if owner != "alice" || input.ModelID != "image-primary" || input.Prompt != "雨夜街道" || input.Size != "2048x1152" || input.AspectRatio != "16:9" || input.Quality != "standard" || input.Resolution != "2K" || input.Count != 1 {
			t.Fatalf("owner=%q input=%+v", owner, input)
		}
		return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
	}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: objects, ImageGenerate: imageGenerator, Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","aspectRatio":"16:9","quality":"standard","resolution":"2K","count":1,"size":"2048x1152","idempotencyKey":"image-request-1"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"kind":"image"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	items, err := assets.ListAssets(context.Background(), "alice", "project-1", "image")
	if err != nil || len(items) != 1 || len(objects.putContents) == 0 {
		t.Fatalf("assets=%+v bytes=%d err=%v", items, len(objects.putContents), err)
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || canvas.Nodes[0].Version != 1 || len(canvas.Nodes[0].AssetRef) != 0 {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
	executions, err := store.Executions(context.Background(), "alice", "project-1")
	if err != nil || len(executions) != 1 || executions[0].Status != "awaiting_confirmation" {
		t.Fatalf("executions=%+v err=%v", executions, err)
	}
	generations, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(generations) != 1 || generations[0].State != agentworkspace.GenerationAwaitingConfirmation || len(generations[0].Results) != 1 {
		t.Fatalf("generations=%+v err=%v", generations, err)
	}
}

func TestAgentImageGenerationPersistsEveryRequestedResult(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 2, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &recordingAgentNodeObjectStore{}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: objects,
		ImageGenerate: func(_ context.Context, _ string, input AgentImageGenerationInput) ([]string, int, error) {
			if input.Count != 2 || input.AspectRatio != "9:16" || input.Quality != "high" || input.Resolution != "4K" || input.Size != "2304x4096" {
				t.Fatalf("input=%+v", input)
			}
			return []string{
				"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB",
				"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAC",
			}, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"竖版海报","aspectRatio":"9:16","quality":"high","resolution":"4K","count":2,"idempotencyKey":"image-request-2"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"resultAssets":[`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	items, err := assets.ListAssets(context.Background(), "alice", "project-1", "image")
	if err != nil || len(items) != 2 {
		t.Fatalf("assets=%+v err=%v", items, err)
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || canvas.Nodes[0].Version != 1 || len(canvas.Nodes[0].AssetRef) != 0 {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
	executions, err := store.Executions(context.Background(), "alice", "project-1")
	if err != nil || len(executions) != 1 || executions[0].Status != "awaiting_confirmation" {
		t.Fatalf("executions=%+v err=%v", executions, err)
	}
	ids, ok := executions[0].Detail["resultAssetIds"].([]any)
	assetIDSet := map[string]bool{}
	for _, item := range items {
		assetIDSet[item.ID] = true
	}
	if !ok || len(ids) != 2 || !assetIDSet[ids[0].(string)] || !assetIDSet[ids[1].(string)] {
		t.Fatalf("resultAssetIds=%#v assets=%+v", executions[0].Detail["resultAssetIds"], items)
	}
}

func TestConfirmAgentImageGenerationAppliesTheSelectedCandidate(t *testing.T) {
	now := time.Date(2026, 10, 21, 10, 20, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: &recordingAgentNodeObjectStore{},
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			return []string{
				"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB",
				"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAC",
			}, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	generated := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"两个候选","parameters":{"count":2},"idempotencyKey":"image-confirm-1"}`)
	if generated.Code != http.StatusOK {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var response struct {
		Execution struct {
			ID string `json:"id"`
		} `json:"execution"`
		ResultAssets []agentworkspace.Asset `json:"resultAssets"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &response); err != nil || response.Execution.ID == "" || len(response.ResultAssets) != 2 {
		t.Fatalf("response=%+v err=%v body=%s", response, err, generated.Body.String())
	}
	selected := response.ResultAssets[1]
	confirmed := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations/"+response.Execution.ID+"/confirm", `{"assetId":"`+selected.ID+`"}`)
	if confirmed.Code != http.StatusOK || !strings.Contains(confirmed.Body.String(), `"status":"completed"`) {
		t.Fatalf("confirm status=%d body=%s", confirmed.Code, confirmed.Body.String())
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || canvas.Nodes[0].Version != 2 || canvas.Nodes[0].AssetRef["id"] != selected.ID {
		t.Fatalf("canvas=%+v selected=%+v err=%v", canvas, selected, err)
	}
	generations, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(generations) != 1 || generations[0].State != agentworkspace.GenerationCompleted || generations[0].SelectedResultID == "" {
		t.Fatalf("generations=%+v err=%v", generations, err)
	}
}

func TestConfirmAgentImageGenerationRejectsAssetOutsideExecution(t *testing.T) {
	now := time.Date(2026, 10, 21, 10, 30, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	if err := assets.CreateAsset(context.Background(), "alice", agentworkspace.Asset{ID: "other-asset", ProjectID: "project-1", Kind: "image", Name: "other.png", ContentType: "image/png", ObjectKey: "agent-assets/other.png", ByteSize: 1, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: &recordingAgentNodeObjectStore{},
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	generated := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"一个候选","idempotencyKey":"image-confirm-reject"}`)
	var response struct {
		Execution struct {
			ID string `json:"id"`
		} `json:"execution"`
	}
	_ = json.Unmarshal(generated.Body.Bytes(), &response)
	rejected := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations/"+response.Execution.ID+"/confirm", `{"assetId":"other-asset"}`)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rejected.Code, rejected.Body.String())
	}
	canvas, _ := store.Canvas(context.Background(), "alice", "project-1")
	if canvas.Nodes[0].Version != 1 || len(canvas.Nodes[0].AssetRef) != 0 {
		t.Fatalf("canvas=%+v", canvas)
	}
}

func TestAgentImageGenerationRejectsCrossProjectReference(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 5, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Version: 1})
	assets := agentworkspace.NewMemoryAssetStore()
	if err := assets.CreateAsset(context.Background(), "alice", agentworkspace.Asset{ID: "foreign-image", ProjectID: "project-2", Kind: "image", Name: "foreign.png", ContentType: "image/png", ObjectKey: "agent-assets/foreign.png", ByteSize: 10, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	called := false
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: &recordingAgentNodeObjectStore{},
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			called = true
			return nil, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","referenceAssetIds":["foreign-image"],"idempotencyKey":"image-request-foreign"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusBadRequest || called {
		t.Fatalf("status=%d called=%v body=%s", recorder.Code, called, recorder.Body.String())
	}
}

func TestAgentImageGenerationProviderFailureLeavesNodeUnchanged(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 10, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Version: 1, Status: "draft"})
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: agentworkspace.NewMemoryAssetStore(), Objects: &recordingAgentNodeObjectStore{},
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			return nil, http.StatusServiceUnavailable, errors.New("图片模型凭据未配置")
		},
		Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"missing-model","prompt":"雨夜街道","idempotencyKey":"image-request-failed"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), "图片模型凭据未配置") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	project, err := store.Project(context.Background(), "alice", "project-1")
	if err != nil || project.Canvas.Nodes[0].Version != 1 || len(project.Canvas.Nodes[0].AssetRef) != 0 || len(project.Executions) != 1 || project.Executions[0].Status != "failed" {
		t.Fatalf("project=%+v err=%v", project, err)
	}
}

func TestAgentImageGenerationRejectsStaleNodeWhenCandidateIsConfirmed(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 15, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Version: 1, Status: "draft"})
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: agentworkspace.NewMemoryAssetStore(), Objects: &recordingAgentNodeObjectStore{},
		ImageGenerate: func(ctx context.Context, owner string, _ AgentImageGenerationInput) ([]string, int, error) {
			canvas, err := store.Canvas(ctx, owner, "project-1")
			if err != nil {
				t.Fatal(err)
			}
			canvas.Nodes[0].Version++
			if _, err := store.SaveCanvas(ctx, owner, "project-1", canvas, canvas.Revision, now.Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","idempotencyKey":"image-request-stale"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusOK {
		t.Fatalf("generate status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Execution struct {
			ID string `json:"id"`
		} `json:"execution"`
		ResultAssets []agentworkspace.Asset `json:"resultAssets"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Execution.ID == "" || len(response.ResultAssets) != 1 {
		t.Fatalf("response=%+v err=%v", response, err)
	}
	confirmed := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations/"+response.Execution.ID+"/confirm", `{"assetId":"`+response.ResultAssets[0].ID+`"}`)
	if confirmed.Code != http.StatusConflict || !strings.Contains(confirmed.Body.String(), "节点已变化") {
		t.Fatalf("confirm status=%d body=%s", confirmed.Code, confirmed.Body.String())
	}
}
