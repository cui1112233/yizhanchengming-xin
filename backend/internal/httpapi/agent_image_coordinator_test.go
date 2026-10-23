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
