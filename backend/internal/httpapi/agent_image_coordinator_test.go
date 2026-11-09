package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func TestAgentImageCoordinatorRunsQueuedWorkAfterHTTPReturns(t *testing.T) {
	now := time.Date(2026, 10, 23, 15, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	services := AgentNodeGenerationServices{Projects: store, Assets: assets, Objects: &recordingAgentNodeObjectStore{}, QueueImages: true, Now: func() time.Time { return now }, ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		started <- struct{}{}
		<-release
		return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
	}}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, services)
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","mode":"text_to_image","idempotencyKey":"queued-image"}`)
	if recorder.Code != http.StatusAccepted || !strings.Contains(recorder.Body.String(), `"state":"queued"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	coordinator := newAgentImageCoordinator(services)
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("image generator did not start")
	}
	close(release)
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
		if err == nil && len(tasks) == 1 && tasks[0].State == agentworkspace.GenerationAwaitingConfirmation && len(tasks[0].Results) == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	tasks, _ := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	t.Fatalf("task did not finish: %+v", tasks)
}

func TestAgentImageCoordinatorRunsDefaultImageGenerationMode(t *testing.T) {
	now := time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	called := make(chan AgentImageGenerationInput, 1)
	services := AgentNodeGenerationServices{Projects: store, Assets: assets, Objects: &recordingAgentNodeObjectStore{}, QueueImages: true, Now: func() time.Time { return now }, ImageGenerate: func(_ context.Context, _ string, input AgentImageGenerationInput) ([]string, int, error) {
		called <- input
		return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
	}}
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, services)
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-generations", `{"nodeId":"image-1","nodeVersion":1,"modelId":"image-primary","prompt":"雨夜街道","idempotencyKey":"default-image-mode"}`)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	coordinator := newAgentImageCoordinator(services)
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case input := <-called:
		if input.Mode != "image_generation" {
			t.Fatalf("mode=%q", input.Mode)
		}
	case <-time.After(time.Second):
		t.Fatal("default image mode did not reach image generator")
	}
}

func TestAgentImageCoordinatorExplainsMissingResultStore(t *testing.T) {
	now := time.Date(2026, 10, 25, 1, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	task, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "generation-missing-store", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "configured_model", ModelID: "image-primary", Mode: "image_generation",
		Settings: map[string]any{"prompt": "雨夜街道"}, IdempotencyKey: "missing-result-store", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	services := AgentNodeGenerationServices{Projects: store, Assets: assets, Now: func() time.Time { return now }, ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		return nil, http.StatusOK, nil
	}}
	coordinator := newAgentImageCoordinator(services)
	if err := coordinator.execute(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GenerationTask(context.Background(), "alice", "project-1", task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ErrorMessage != "图片生成服务未配置：缺少结果素材存储" {
		t.Fatalf("error=%q", updated.ErrorMessage)
	}
}

func TestAgentImageCoordinatorRecoversStaleRunningTaskAfterRestart(t *testing.T) {
	createdAt := time.Date(2026, 10, 25, 2, 0, 0, 0, time.UTC)
	recoveredAt := createdAt.Add(agentImageLease + time.Second)
	store := newAgentNodeGenerationTestStore(t, createdAt, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "draft"})
	task, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "generation-stale-running", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "configured_model", ModelID: "image-primary", Mode: "image_generation",
		Settings: map[string]any{"prompt": "雨夜街道", "parameters": map[string]any{"ratio": "16:9", "quality": "standard", "resolution": "2K", "count": 1}}, IdempotencyKey: "stale-running-image", Now: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.TransitionGenerationTask(context.Background(), "alice", "project-1", task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: createdAt.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{}, 1)
	services := AgentNodeGenerationServices{
		Projects: store, Assets: agentworkspace.NewMemoryAssetStore(), Objects: &recordingAgentNodeObjectStore{}, Now: func() time.Time { return recoveredAt },
		ImageGenerate: func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
			started <- struct{}{}
			return []string{"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB"}, http.StatusOK, nil
		},
	}
	coordinator := newAgentImageCoordinator(services)
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		stored, _ := store.GenerationTask(context.Background(), "alice", "project-1", task.ID)
		t.Fatalf("stale running image task was not recovered: %+v", stored)
	}
}
