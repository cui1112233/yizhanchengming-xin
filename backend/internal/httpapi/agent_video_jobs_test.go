package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
)

func TestAgentVideoJobRoutesSaveResultBeforeConfirmingCanvasNode(t *testing.T) {
	now := time.Date(2026, 10, 20, 0, 10, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft", Position: agentworkspace.Position{}}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}

	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"task_id":"provider-job","status":"queued"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"SUCCESS","url":"https://cdn.example/video.mp4"}`))
	}))
	defer provider.Close()
	client := provider.Client()
	transport := client.Transport
	client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		copyReq := req.Clone(req.Context())
		parsed, err := url.Parse(provider.URL + req.URL.Path)
		if err != nil {
			return nil, err
		}
		copyReq.URL = parsed
		return transport.RoundTrip(copyReq)
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderAutoDLH3, APIKey: "secret", Model: batchfactoryv11.AutoDLH3Model, CreateURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", TasksURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}", WorkflowID: batchfactoryv11.AutoDLH3NoImageWorkflow}, nil
	}
	assetStore := agentworkspace.NewMemoryAssetStore()
	objects := &agentAssetObjectTestStore{files: map[string][]byte{}}
	coordinator := &agentVideoCoordinator{
		projects: projects, assets: assetStore,
		videos: &scriptVideoService{resolver: resolver, client: client},
		ingestor: &agentVideoAssetIngestor{assets: assetStore, objects: objects, now: func() time.Time { return now }, fetch: func(context.Context, string) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("mp4"))), "video/mp4", nil
		}},
		now: func() time.Time { return now },
	}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)

	post := agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs", `{"nodeId":"video-1","idempotencyKey":"request-1","prompt":"雨夜追逐","modelKey":"minimax-h3-video","duration":10}`)
	created := httptest.NewRecorder()
	mux.ServeHTTP(created, post)
	if created.Code != http.StatusAccepted {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	jobID := agentVideoJobID(t, created)

	poll := agentVideoJSONRequest(t, now, http.MethodGet, "/api/agent/tasks/agent-1/video-jobs/"+jobID, "")
	polled := httptest.NewRecorder()
	mux.ServeHTTP(polled, poll)
	if polled.Code != http.StatusOK || !strings.Contains(polled.Body.String(), `"state":"awaiting_confirmation"`) || !strings.Contains(polled.Body.String(), `"resultAsset"`) {
		t.Fatalf("poll=%d %s", polled.Code, polled.Body.String())
	}
	if len(objects.files) != 1 {
		t.Fatalf("private objects=%d", len(objects.files))
	}

	confirm := agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs/"+jobID+"/confirm", "")
	confirmed := httptest.NewRecorder()
	mux.ServeHTTP(confirmed, confirm)
	if confirmed.Code != http.StatusOK || !strings.Contains(confirmed.Body.String(), `"state":"applied"`) {
		t.Fatalf("confirm=%d %s", confirmed.Code, confirmed.Body.String())
	}
	canvas, err := projects.Canvas(context.Background(), "alice", "agent-1")
	if err != nil || canvas.Nodes[0].AssetRef["kind"] != "video" || canvas.Nodes[0].Status != "confirmed" {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
}

func agentVideoJSONRequest(t *testing.T, now time.Time, method, target, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	SignBridgeRequest(req, "alice", false, now, "secret")
	return req
}

func agentVideoJobID(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var payload struct {
		Job struct {
			ID string `json:"id"`
		} `json:"job"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil || payload.Job.ID == "" {
		t.Fatalf("response=%s err=%v", recorder.Body.String(), err)
	}
	return payload.Job.ID
}
