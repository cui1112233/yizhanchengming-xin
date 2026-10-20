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
