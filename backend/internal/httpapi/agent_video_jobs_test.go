package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type recordingAgentReferenceURLProvider struct {
	assets []agentworkspace.Asset
	url    string
}

func (p *recordingAgentReferenceURLProvider) ReferenceURL(_ context.Context, asset agentworkspace.Asset, _ time.Duration) (string, error) {
	p.assets = append(p.assets, asset)
	return p.url, nil
}

func TestAgentVideoJobSubmitUsesRequestedReferenceAndEffectiveH3Settings(t *testing.T) {
	now := time.Date(2026, 10, 20, 1, 0, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	image := agentworkspace.Asset{ID: "manual-image", ProjectID: "agent-1", Kind: "image", Name: "reference.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/agent-1/manual-image.png", ByteSize: 12, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", image); err != nil {
		t.Fatal(err)
	}
	references := &recordingAgentReferenceURLProvider{url: "https://assets.example/reference.png"}
	var providerPayload map[string]any
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := json.NewDecoder(req.Body).Decode(&providerPayload); err != nil {
			t.Errorf("decode provider payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task_id":"provider-job","status":"queued"}`))
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
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderAutoDLH3, APIKey: "secret", Model: batchfactoryv11.AutoDLH3Model, CreateURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", TasksURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}"}, nil
	}
	coordinator := &agentVideoCoordinator{
		projects: projects, assets: assets, references: references,
		videos: &scriptVideoService{resolver: resolver, client: client},
		now:    func() time.Time { return now },
	}
	job, err := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
		NodeID: "video-1", IdempotencyKey: "request-reference", Prompt: "雨夜追逐", ModelKey: "minimax-h3-video",
		Duration: 10, AspectRatio: "16:9", Resolution: "720p", ReferenceAssetIDs: []string{"manual-image", "manual-image"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(references.assets) != 1 || references.assets[0].ID != image.ID {
		t.Fatalf("reference assets=%+v", references.assets)
	}
	if len(job.ReferenceAssetIDs) != 1 || job.ReferenceAssetIDs[0] != image.ID {
		t.Fatalf("job reference ids=%v", job.ReferenceAssetIDs)
	}
	if providerPayload["ref_image_0"] != references.url || providerPayload["resolution"] != "768p横" {
		t.Fatalf("provider payload=%+v", providerPayload)
	}
}

func TestAgentVideoJobSubmitRejectsReferenceOutsideOwnerProject(t *testing.T) {
	now := time.Date(2026, 10, 20, 1, 10, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	foreign := agentworkspace.Asset{ID: "foreign-image", ProjectID: "agent-1", Kind: "image", Name: "foreign.png", ContentType: "image/png", ObjectKey: "agent-assets/bob/agent-1/foreign-image.png", ByteSize: 12, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "bob", foreign); err != nil {
		t.Fatal(err)
	}
	references := &recordingAgentReferenceURLProvider{url: "https://assets.example/foreign.png"}
	coordinator := &agentVideoCoordinator{projects: projects, assets: assets, references: references, videos: &scriptVideoService{}, now: func() time.Time { return now }}
	_, err := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
		NodeID: "video-1", IdempotencyKey: "request-foreign", Prompt: "雨夜追逐", ModelKey: "minimax-h3-video", ReferenceAssetIDs: []string{"foreign-image"},
	})
	if !errors.Is(err, agentworkspace.ErrAssetNotFound) {
		t.Fatalf("err=%v, want ErrAssetNotFound", err)
	}
	if len(references.assets) != 0 {
		t.Fatalf("reference provider calls=%d, want 0", len(references.assets))
	}
}

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

func TestAgentVideoJobListRouteIsOwnerProjectScoped(t *testing.T) {
	now := time.Date(2026, 10, 20, 2, 0, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	for _, item := range []struct{ owner, id string }{{"alice", "agent-1"}, {"alice", "agent-2"}, {"bob", "agent-1"}} {
		if err := projects.CreateProject(context.Background(), item.owner, agentworkspace.Project{ID: item.id, Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range []struct{ owner, id, projectID, nodeID, key string }{
		{"alice", "alice-current", "agent-1", "video-1", "alice-current-key"},
		{"alice", "alice-other-project", "agent-2", "video-2", "alice-other-key"},
		{"bob", "bob-current", "agent-1", "video-3", "bob-current-key"},
	} {
		if _, err := projects.CreateVideoJob(context.Background(), input.owner, agentworkspace.CreateVideoJobInput{
			ID: input.id, ProjectID: input.projectID, NodeID: input.nodeID, NodeVersion: 1,
			ModelKey: "minimax-h3-video", IdempotencyKey: input.key, Now: now,
		}); err != nil {
			t.Fatal(err)
		}
	}
	assets := agentworkspace.NewMemoryAssetStore()
	coordinator := &agentVideoCoordinator{projects: projects, assets: assets, videos: &scriptVideoService{}, ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return now }}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)

	request := agentVideoJSONRequest(t, now, http.MethodGet, "/api/agent/tasks/agent-1/video-jobs", "")
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("list=%d %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Jobs []agentworkspace.VideoJob `json:"jobs"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Jobs) != 1 || payload.Jobs[0].ID != "alice-current" {
		t.Fatalf("jobs=%+v", payload.Jobs)
	}
}

func TestAgentVideoJobImmediateProviderFailureReturnsPersistedFailedJob(t *testing.T) {
	now := time.Date(2026, 10, 20, 2, 30, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
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
	assets := agentworkspace.NewMemoryAssetStore()
	coordinator := &agentVideoCoordinator{projects: projects, assets: assets, videos: &scriptVideoService{resolver: resolver, client: client}, ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return now }}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)

	request := agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs", `{"nodeId":"video-1","idempotencyKey":"immediate-failure","prompt":"雨夜追逐","modelKey":"minimax-h3-video","duration":10}`)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("create=%d %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Job   agentworkspace.VideoJob `json:"job"`
		Error string                  `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Job.ID == "" || payload.Job.State != agentworkspace.VideoJobFailed || payload.Job.ErrorMessage == "" || payload.Error == "" {
		t.Fatalf("payload=%+v", payload)
	}
	stored, err := projects.VideoJob(context.Background(), "alice", "agent-1", payload.Job.ID)
	if err != nil || stored.State != agentworkspace.VideoJobFailed {
		t.Fatalf("stored=%+v err=%v", stored, err)
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
