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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localexecutor"
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

func TestAgentVideoJobSubmitRejectsVideoReferenceBeforeProviderEvenWhenIgnoreReferencesIsSet(t *testing.T) {
	now := time.Date(2026, 10, 20, 1, 15, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	video := agentworkspace.Asset{ID: "upstream-video", ProjectID: "agent-1", Kind: "video", Name: "upstream.mp4", ContentType: "video/mp4", ObjectKey: "agent-assets/alice/agent-1/upstream.mp4", ByteSize: 12, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", video); err != nil {
		t.Fatal(err)
	}
	providerCalls := 0
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		providerCalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task_id":"provider-job","status":"queued"}`))
	}))
	defer provider.Close()
	client := provider.Client()
	transport := client.Transport
	client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		copyReq := req.Clone(req.Context())
		parsed, parseErr := url.Parse(provider.URL + req.URL.Path)
		if parseErr != nil {
			return nil, parseErr
		}
		copyReq.URL = parsed
		return transport.RoundTrip(copyReq)
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderAutoDLH3, APIKey: "secret", Model: batchfactoryv11.AutoDLH3Model, CreateURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", TasksURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}", WorkflowID: batchfactoryv11.AutoDLH3NoImageWorkflow}, nil
	}
	coordinator := &agentVideoCoordinator{projects: projects, assets: assets, references: &recordingAgentReferenceURLProvider{url: "https://assets.example/reference"}, videos: &scriptVideoService{resolver: resolver, client: client}, now: func() time.Time { return now }}

	_, err := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
		NodeID: "video-1", IdempotencyKey: "request-video-reference", Prompt: "雨夜追逐", ModelKey: "minimax-h3-video",
		ReferenceAssetIDs: []string{"upstream-video"}, IgnoreReferences: true,
	})
	if err == nil || !strings.Contains(err.Error(), "暂不支持视频参考素材") {
		t.Fatalf("err=%v, want explicit unsupported video reference error", err)
	}
	if providerCalls != 0 {
		t.Fatalf("provider calls=%d, want 0", providerCalls)
	}
	jobs, listErr := projects.ListVideoJobs(context.Background(), "alice", "agent-1")
	if listErr != nil || len(jobs) != 0 {
		t.Fatalf("jobs=%+v err=%v, want no created job", jobs, listErr)
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

func TestAgentVideoRetryRouteRejectsNonRetryableStatesBeforeProviderSubmit(t *testing.T) {
	states := []agentworkspace.VideoJobState{
		agentworkspace.VideoJobQueued,
		agentworkspace.VideoJobRunning,
		agentworkspace.VideoJobResultReady,
		agentworkspace.VideoJobAwaitingConfirmation,
		agentworkspace.VideoJobApplied,
	}
	for _, targetState := range states {
		t.Run(string(targetState), func(t *testing.T) {
			now := time.Date(2026, 10, 20, 5, 0, 0, 0, time.UTC)
			projects := agentworkspace.NewMemoryStore()
			if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
				ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
				Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
			}); err != nil {
				t.Fatal(err)
			}
			job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{
				ID: "job-1", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1,
				ModelKey: "minimax-h3-video", Settings: map[string]any{"prompt": "雨夜追逐", "duration": 10}, IdempotencyKey: "request-1", Now: now,
			})
			if err != nil {
				t.Fatal(err)
			}
			job = advanceAgentVideoJobForTest(t, projects, now, job, targetState)

			var providerCalls int
			provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				providerCalls++
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"task_id":"provider-retry","status":"queued"}`))
			}))
			defer provider.Close()
			client := provider.Client()
			transport := client.Transport
			client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				copyReq := req.Clone(req.Context())
				parsed, parseErr := url.Parse(provider.URL + req.URL.Path)
				if parseErr != nil {
					return nil, parseErr
				}
				copyReq.URL = parsed
				return transport.RoundTrip(copyReq)
			})
			resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
				return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderAutoDLH3, APIKey: "secret", Model: batchfactoryv11.AutoDLH3Model, CreateURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", TasksURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}", WorkflowID: batchfactoryv11.AutoDLH3NoImageWorkflow}, nil
			}
			coordinator := &agentVideoCoordinator{projects: projects, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{resolver: resolver, client: client}, ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return now.Add(10 * time.Second) }}
			mux := http.NewServeMux()
			RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)

			request := agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs/"+job.ID+"/retry", `{"idempotencyKey":"request-retry"}`)
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusConflict {
				t.Fatalf("retry=%d %s, want conflict", recorder.Code, recorder.Body.String())
			}
			if providerCalls != 0 {
				t.Fatalf("provider calls=%d, want 0", providerCalls)
			}
		})
	}
}

func TestAgentVideoRetryCreatesDistinctLocalExecutorJob(t *testing.T) {
	now := time.Date(2026, 10, 21, 2, 0, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	executor := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	coordinator := &agentVideoCoordinator{
		projects: projects,
		assets:   agentworkspace.NewMemoryAssetStore(),
		videos:   &scriptVideoService{localExecutor: executor},
		now:      func() time.Time { return now },
	}

	first, err := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
		NodeID: "video-1", IdempotencyKey: "request-1", Prompt: "雨夜追逐", ModelKey: "local-doubao-executor-video", Duration: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	first.State, first.ErrorMessage = agentworkspace.VideoJobFailed, "模拟首次执行失败"
	first, err = projects.UpdateVideoJob(context.Background(), "alice", first, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	retried, err := coordinator.retry(context.Background(), "alice", "agent-1", first.ID, "request-2")
	if err != nil {
		t.Fatalf("retry err=%v", err)
	}
	if retried.ProviderTaskID == "" || retried.ProviderTaskID == first.ProviderTaskID {
		t.Fatalf("first provider task=%q retry provider task=%q, want distinct tasks", first.ProviderTaskID, retried.ProviderTaskID)
	}
}

func TestConcurrentSuccessfulAgentVideoPollsPersistOneResultAsset(t *testing.T) {
	now := time.Date(2026, 10, 20, 5, 30, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "job-1", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", Settings: map[string]any{"prompt": "雨夜追逐"}, IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State, job.ProviderTaskID = agentworkspace.VideoJobQueued, "h3:provider-job"
	job, err = projects.UpdateVideoJob(context.Background(), "alice", job, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"SUCCESS","url":"https://cdn.example/video.mp4"}`))
	}))
	defer provider.Close()
	client := provider.Client()
	transport := client.Transport
	client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		copyReq := req.Clone(req.Context())
		parsed, parseErr := url.Parse(provider.URL + req.URL.Path)
		if parseErr != nil {
			return nil, parseErr
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
		ingestor: &agentVideoAssetIngestor{assets: assetStore, objects: objects, now: func() time.Time { return now.Add(3 * time.Second) }, fetch: func(context.Context, string) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("mp4"))), "video/mp4", nil
		}},
		now: func() time.Time { return now.Add(2 * time.Second) },
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func(snapshot agentworkspace.VideoJob) {
			defer group.Done()
			<-start
			_, pollErr := coordinator.poll(context.Background(), "alice", snapshot)
			results <- pollErr
		}(job)
	}
	close(start)
	group.Wait()
	close(results)
	for pollErr := range results {
		if pollErr != nil {
			t.Fatalf("poll err=%v", pollErr)
		}
	}
	stored, err := projects.VideoJob(context.Background(), "alice", "agent-1", "job-1")
	if err != nil || stored.State != agentworkspace.VideoJobAwaitingConfirmation {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	assets, err := assetStore.ListAssets(context.Background(), "alice", "agent-1", "video")
	if err != nil || len(assets) != 1 || len(objects.files) != 1 {
		t.Fatalf("assets=%d objects=%d err=%v", len(assets), len(objects.files), err)
	}
}

func TestAgentVideoPollDuringResultIngestDoesNotSaveSecondAsset(t *testing.T) {
	now := time.Date(2026, 10, 20, 5, 45, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "job-1", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", Settings: map[string]any{"prompt": "雨夜追逐"}, IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State, job.ProviderTaskID = agentworkspace.VideoJobQueued, "h3:provider-job"
	job, err = projects.UpdateVideoJob(context.Background(), "alice", job, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"SUCCESS","url":"https://cdn.example/video.mp4"}`))
	}))
	defer provider.Close()
	client := provider.Client()
	transport := client.Transport
	client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		copyReq := req.Clone(req.Context())
		parsed, parseErr := url.Parse(provider.URL + req.URL.Path)
		if parseErr != nil {
			return nil, parseErr
		}
		copyReq.URL = parsed
		return transport.RoundTrip(copyReq)
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderAutoDLH3, APIKey: "secret", Model: batchfactoryv11.AutoDLH3Model, CreateURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", TasksURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}", WorkflowID: batchfactoryv11.AutoDLH3NoImageWorkflow}, nil
	}
	assetStore := agentworkspace.NewMemoryAssetStore()
	objects := &agentAssetObjectTestStore{files: map[string][]byte{}}
	firstFetchStarted := make(chan struct{})
	releaseFirstFetch := make(chan struct{})
	var fetchCalls atomic.Int32
	var clockNanos atomic.Int64
	clockNanos.Store(now.Add(2 * time.Second).UnixNano())
	coordinator := &agentVideoCoordinator{
		projects: projects, assets: assetStore,
		videos: &scriptVideoService{resolver: resolver, client: client},
		ingestor: &agentVideoAssetIngestor{assets: assetStore, objects: objects, now: func() time.Time { return now.Add(3 * time.Second) }, fetch: func(context.Context, string) (io.ReadCloser, string, error) {
			if fetchCalls.Add(1) == 1 {
				close(firstFetchStarted)
				<-releaseFirstFetch
			}
			return io.NopCloser(bytes.NewReader([]byte("mp4"))), "video/mp4", nil
		}},
		now: func() time.Time { return time.Unix(0, clockNanos.Load()).UTC() },
	}

	firstResult := make(chan error, 1)
	go func() {
		_, pollErr := coordinator.poll(context.Background(), "alice", job)
		firstResult <- pollErr
	}()
	<-firstFetchStarted
	duringIngest, err := projects.VideoJob(context.Background(), "alice", "agent-1", job.ID)
	if err != nil {
		t.Fatal(err)
	}
	clockNanos.Store(now.Add(agentVideoIngestLease + time.Minute).UnixNano())
	if _, err := coordinator.poll(context.Background(), "alice", duringIngest); err != nil {
		t.Fatalf("second poll err=%v", err)
	}
	close(releaseFirstFetch)
	if err := <-firstResult; err != nil {
		t.Fatalf("first poll err=%v", err)
	}

	stored, err := projects.VideoJob(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || stored.State != agentworkspace.VideoJobAwaitingConfirmation {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	assets, err := assetStore.ListAssets(context.Background(), "alice", "agent-1", "video")
	if err != nil || len(assets) != 1 || len(objects.files) != 1 {
		t.Fatalf("assets=%d objects=%d fetches=%d err=%v", len(assets), len(objects.files), fetchCalls.Load(), err)
	}
}

func TestAgentVideoCancelClaimsStateBeforeCallingProvider(t *testing.T) {
	now := time.Date(2026, 10, 20, 5, 55, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "job-1", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job = advanceAgentVideoJobForTest(t, projects, now, job, agentworkspace.VideoJobAwaitingConfirmation)
	coordinator := &agentVideoCoordinator{projects: projects, videos: &scriptVideoService{}, now: func() time.Time { return now.Add(10 * time.Second) }}

	if _, err := coordinator.cancel(context.Background(), "alice", job); !errors.Is(err, agentworkspace.ErrVideoJobState) {
		t.Fatalf("cancel err=%v, want ErrVideoJobState before provider call", err)
	}
	stored, err := projects.VideoJob(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || stored.State != agentworkspace.VideoJobAwaitingConfirmation {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
}

func advanceAgentVideoJobForTest(t *testing.T, store *agentworkspace.MemoryStore, now time.Time, job agentworkspace.VideoJob, target agentworkspace.VideoJobState) agentworkspace.VideoJob {
	t.Helper()
	steps := []agentworkspace.VideoJobState{agentworkspace.VideoJobQueued, agentworkspace.VideoJobRunning, agentworkspace.VideoJobResultReady, agentworkspace.VideoJobIngesting, agentworkspace.VideoJobAwaitingConfirmation}
	for index, state := range steps {
		job.State = state
		if state == agentworkspace.VideoJobAwaitingConfirmation {
			job.ResultAssetID = "result-1"
		}
		updated, err := store.UpdateVideoJob(context.Background(), "alice", job, now.Add(time.Duration(index+1)*time.Second))
		if err != nil {
			t.Fatal(err)
		}
		job = updated
		if state == target {
			return job
		}
	}
	if target == agentworkspace.VideoJobApplied {
		updated, err := store.ConfirmVideoJob(context.Background(), "alice", "agent-1", job.ID, now.Add(6*time.Second))
		if err != nil {
			t.Fatal(err)
		}
		return updated
	}
	t.Fatalf("unsupported target state %s", target)
	return agentworkspace.VideoJob{}
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
