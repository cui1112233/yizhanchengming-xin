package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
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

func TestAgentImageGenerationPersistsAssetBeforeApplyingNode(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &recordingAgentNodeObjectStore{}
	imageGenerator := func(_ context.Context, owner, modelID, prompt, size string) (string, int, error) {
		if owner != "alice" || modelID != "image-primary" || prompt != "雨夜街道" || size != "1024x1024" {
			t.Fatalf("owner=%q model=%q prompt=%q size=%q", owner, modelID, prompt, size)
		}
		return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB", http.StatusOK, nil
	}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: assets, Objects: objects, ImageGenerate: imageGenerator, Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","size":"1024x1024","idempotencyKey":"image-request-1"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"kind":"image"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	items, err := assets.ListAssets(context.Background(), "alice", "project-1", "image")
	if err != nil || len(items) != 1 || len(objects.putContents) == 0 {
		t.Fatalf("assets=%+v bytes=%d err=%v", items, len(objects.putContents), err)
	}
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil || canvas.Nodes[0].Version != 2 || canvas.Nodes[0].AssetRef["id"] != items[0].ID {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
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
		ImageGenerate: func(context.Context, string, string, string, string) (string, int, error) {
			called = true
			return "", http.StatusOK, nil
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
		ImageGenerate: func(context.Context, string, string, string, string) (string, int, error) {
			return "", http.StatusServiceUnavailable, errors.New("图片模型凭据未配置")
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

func TestAgentImageGenerationRejectsStaleNodeAfterProviderCompletes(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 15, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Version: 1, Status: "draft"})
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{
		Projects: store, Assets: agentworkspace.NewMemoryAssetStore(), Objects: &recordingAgentNodeObjectStore{},
		ImageGenerate: func(ctx context.Context, owner, _, _, _ string) (string, int, error) {
			canvas, err := store.Canvas(ctx, owner, "project-1")
			if err != nil {
				t.Fatal(err)
			}
			canvas.Nodes[0].Version++
			if _, err := store.SaveCanvas(ctx, owner, "project-1", canvas, canvas.Revision, now.Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB", http.StatusOK, nil
		},
		Now: func() time.Time { return now },
	})
	body := `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","idempotencyKey":"image-request-stale"}`
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", body)
	if recorder.Code != http.StatusConflict || !strings.Contains(recorder.Body.String(), "节点已变化") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
