package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type agentVideoFaultStore struct {
	agentworkspace.Store
	failCreateGeneration     int
	failGenerationState      agentworkspace.GenerationState
	failGenerationTransition int
	raceVideoFailure         bool
	failProviderTaskPersist  int
}

func (s *agentVideoFaultStore) CreateGenerationTask(ctx context.Context, owner string, input agentworkspace.CreateGenerationTaskInput) (agentworkspace.GenerationTask, error) {
	if s.failCreateGeneration > 0 {
		s.failCreateGeneration--
		return agentworkspace.GenerationTask{}, errors.New("temporary generation create failure")
	}
	return s.Store.CreateGenerationTask(ctx, owner, input)
}

func (s *agentVideoFaultStore) TransitionGenerationTask(ctx context.Context, owner, projectID, id string, transition agentworkspace.GenerationTaskTransition) (agentworkspace.GenerationTask, error) {
	if s.failGenerationTransition > 0 && transition.State == s.failGenerationState {
		s.failGenerationTransition--
		return agentworkspace.GenerationTask{}, errors.New("temporary generation transition failure")
	}
	return s.Store.TransitionGenerationTask(ctx, owner, projectID, id, transition)
}

func (s *agentVideoFaultStore) UpdateVideoJob(ctx context.Context, owner string, input agentworkspace.VideoJob, now time.Time) (agentworkspace.VideoJob, error) {
	if s.failProviderTaskPersist > 0 && input.State == agentworkspace.VideoJobQueued {
		s.failProviderTaskPersist--
		return agentworkspace.VideoJob{}, errors.New("temporary provider task persistence failure")
	}
	if s.raceVideoFailure && input.State == agentworkspace.VideoJobFailed {
		s.raceVideoFailure = false
		current, err := s.Store.VideoJob(ctx, owner, input.ProjectID, input.ID)
		if err != nil {
			return agentworkspace.VideoJob{}, err
		}
		current.State = agentworkspace.VideoJobCanceling
		if _, err := s.Store.UpdateVideoJob(ctx, owner, current, now); err != nil {
			return agentworkspace.VideoJob{}, err
		}
		return agentworkspace.VideoJob{}, agentworkspace.ErrVideoJobState
	}
	return s.Store.UpdateVideoJob(ctx, owner, input, now)
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

func TestAgentVideoJobStoresResolvedCanvasInputSnapshot(t *testing.T) {
	now := time.Date(2026, 10, 23, 17, 5, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{
		Version: 1, Nodes: []agentworkspace.Node{
			{ID: "text-1", Type: "text", Title: "文案", Summary: "上游镜头文案", Version: 1, Status: "ready"},
			{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"},
		}, Edges: []agentworkspace.Edge{{ID: "text-to-video", Source: "text-1", Target: "video-1"}},
	}}); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{
		projects: projects, assets: agentworkspace.NewMemoryAssetStore(),
		videos: &scriptVideoService{localExecutor: localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })},
		now:    func() time.Time { return now },
	}
	job, _ := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
		NodeID: "video-1", NodeVersion: 1, IdempotencyKey: "video-input-snapshot", Prompt: "本次视频指令", Provider: "doubao_local_executor", ModelID: "Seedance 2.5",
		Parameters: map[string]any{"ratio": "16:9", "resolution": "1080P", "durationSeconds": 10, "count": 1},
	})
	if job.ID == "" {
		t.Fatal("video job was not persisted")
	}
	snapshot, ok := job.Settings["resolvedInputs"].(map[string]any)
	if !ok || snapshot["prompt"] != "本次视频指令" || snapshot["resolvedPrompt"] != "上游镜头文案" {
		t.Fatalf("snapshot=%+v", job.Settings["resolvedInputs"])
	}
}

func TestAgentVideoJobSubmitCarriesDisplayedSettingsIntoLocalExecutor(t *testing.T) {
	now := time.Date(2026, 10, 21, 3, 0, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 2, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	executorStore := localexecutor.NewMemoryStore()
	executor := localexecutor.NewService(executorStore, func() time.Time { return now })
	assets := agentworkspace.NewMemoryAssetStore()
	reference := agentworkspace.Asset{ID: "reference-image", ProjectID: "agent-1", Kind: "image", Name: "reference.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/agent-1/reference-image.png", ByteSize: 12, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", reference); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{
		projects: projects,
		assets:   assets,
		videos:   &scriptVideoService{localExecutor: executor},
		now:      func() time.Time { return now },
	}

	job, err := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
		NodeID: "video-1", NodeVersion: 2, IdempotencyKey: "request-settings", Prompt: "缓慢推进",
		Provider: "doubao_local_executor", ModelID: "Seedance 2.5", Mode: "first_frame",
		Parameters: map[string]any{
			"ratio": "16:9", "resolution": "1080P", "durationSeconds": 10,
			"count": 1, "generateAudio": true,
		},
		Advanced:          map[string]any{"camera_motion": "push_in"},
		ReferenceAssetIDs: []string{"reference-image"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.ModelKey != "local-doubao-executor-video" {
		t.Fatalf("model key=%q", job.ModelKey)
	}
	for key, want := range map[string]any{
		"provider": "doubao_local_executor", "modelId": "Seedance 2.5", "mode": "first_frame",
		"durationSeconds": 10, "count": 1, "generateAudio": true,
	} {
		if got := job.Settings[key]; fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("setting %s=%v, want %v; all=%+v", key, got, want, job.Settings)
		}
	}
	localJob, err := executor.GetJob(context.Background(), "alice", strings.TrimPrefix(job.ProviderTaskID, localScriptVideoTaskPrefix))
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(localJob.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["model"] != "Seedance 2.5" || payload["generateAudio"] != true || payload["duration"] != float64(10) {
		t.Fatalf("executor payload=%+v", payload)
	}
	urls, ok := payload["imageUrls"].([]any)
	if !ok || len(urls) != 1 || urls[0] != "/api/local-executor/v1/agent-reference-assets/agent-1/reference-image" {
		t.Fatalf("executor reference urls=%+v", payload["imageUrls"])
	}
}

func TestAgentVideoJobSubmitRejectsUnsupportedDisplayedSettingsBeforeCreatingJob(t *testing.T) {
	now := time.Date(2026, 10, 21, 3, 10, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		parameters map[string]any
	}{
		{name: "resolution", parameters: map[string]any{"ratio": "16:9", "resolution": "4K", "durationSeconds": 10, "count": 1}},
		{name: "duration", parameters: map[string]any{"ratio": "16:9", "resolution": "1080P", "durationSeconds": 20, "count": 1}},
		{name: "count", parameters: map[string]any{"ratio": "16:9", "resolution": "1080P", "durationSeconds": 10, "count": 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			projects := agentworkspace.NewMemoryStore()
			if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
				ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
				Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
			}); err != nil {
				t.Fatal(err)
			}
			coordinator := &agentVideoCoordinator{
				projects: projects, assets: agentworkspace.NewMemoryAssetStore(),
				videos: &scriptVideoService{localExecutor: localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })},
				now:    func() time.Time { return now },
			}
			_, err := coordinator.submit(context.Background(), "alice", "agent-1", agentVideoJobRequest{
				NodeID: "video-1", IdempotencyKey: "invalid-" + tc.name, Prompt: "缓慢推进",
				Provider: "doubao_local_executor", ModelID: "Seedance 2.5", Mode: "text_to_video", Parameters: tc.parameters,
			})
			if err == nil || !strings.Contains(err.Error(), tc.name) {
				t.Fatalf("err=%v, want field %s", err, tc.name)
			}
			jobs, listErr := projects.ListVideoJobs(context.Background(), "alice", "agent-1")
			if listErr != nil || len(jobs) != 0 {
				t.Fatalf("jobs=%+v err=%v, want no created job", jobs, listErr)
			}
		})
	}
}

func TestAgentVideoJobRouteReturnsUnprocessableEntityForUnsupportedDisplayedSettings(t *testing.T) {
	now := time.Date(2026, 10, 21, 3, 20, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{
		ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Title: "镜头", Version: 1, Status: "draft"}}, Edges: []agentworkspace.Edge{}},
	}); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{
		projects: projects, assets: agentworkspace.NewMemoryAssetStore(),
		videos:   &scriptVideoService{localExecutor: localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })},
		ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return now },
	}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)
	request := agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs", `{"nodeId":"video-1","nodeVersion":1,"idempotencyKey":"invalid-resolution","prompt":"缓慢推进","provider":"doubao_local_executor","modelId":"Seedance 2.5","mode":"text_to_video","parameters":{"ratio":"16:9","resolution":"4K","durationSeconds":10,"count":1,"generateAudio":true},"advanced":{}}`)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "resolution") {
		t.Fatalf("response=%d %s", recorder.Code, recorder.Body.String())
	}
	jobs, err := projects.ListVideoJobs(context.Background(), "alice", "agent-1")
	if err != nil || len(jobs) != 0 {
		t.Fatalf("jobs=%+v err=%v, want no created job", jobs, err)
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
	generations, err := projects.ListGenerationTasks(context.Background(), "alice", "agent-1")
	if err != nil || len(generations) != 1 || generations[0].ID != jobID || generations[0].State != agentworkspace.GenerationAwaitingConfirmation || len(generations[0].Results) != 1 {
		t.Fatalf("generations after poll=%+v err=%v", generations, err)
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
	generations, err = projects.ListGenerationTasks(context.Background(), "alice", "agent-1")
	if err != nil || generations[0].State != agentworkspace.GenerationCompleted || generations[0].SelectedResultID == "" {
		t.Fatalf("generations after confirm=%+v err=%v", generations, err)
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

func TestAgentVideoCoordinatorPollPendingOnceCompletesVideoWithoutBrowserRequest(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 30, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "job-background", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", Settings: map[string]any{"prompt": "雨夜追逐"}, IdempotencyKey: "request-background", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State, job.ProviderTaskID = agentworkspace.VideoJobQueued, "h3:provider-background"
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

	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatalf("background poll err=%v", err)
	}
	stored, err := projects.VideoJob(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || stored.State != agentworkspace.VideoJobAwaitingConfirmation {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	assets, err := assetStore.ListAssets(context.Background(), "alice", "agent-1", "video")
	if err != nil || len(assets) != 1 || len(objects.files) != 1 {
		t.Fatalf("assets=%d objects=%d err=%v", len(assets), len(objects.files), err)
	}
}

func TestAgentVideoCoordinatorPollPendingOnceAbandonsExpiredSubmittingJobWithoutResubmit(t *testing.T) {
	now := time.Date(2026, 10, 23, 13, 0, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1}}}}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "job-submitting", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", Settings: map[string]any{"prompt": "雨夜追逐"}, IdempotencyKey: "request-submitting", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State = agentworkspace.VideoJobSubmitting
	job, err = projects.UpdateVideoJob(context.Background(), "alice", job, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	clock := now.Add(agentVideoSubmissionLease + 2*time.Second)
	coordinator := &agentVideoCoordinator{
		projects: projects, assets: agentworkspace.NewMemoryAssetStore(), ingestor: &agentVideoAssetIngestor{},
		videos: &scriptVideoService{},
		now:    func() time.Time { return clock },
	}
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	stored, err := projects.VideoJob(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || stored.State != agentworkspace.VideoJobAbandoned {
		t.Fatalf("stored=%+v err=%v", stored, err)
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

func TestAgentVideoPollHealsQueuedGenerationBeforeAwaitingConfirmation(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 10, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "video-heal", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "model", IdempotencyKey: "video-heal", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job = advanceAgentVideoJobForTest(t, projects, now, job, agentworkspace.VideoJobAwaitingConfirmation)
	assets := agentworkspace.NewMemoryAssetStore()
	if err := assets.CreateAsset(context.Background(), "alice", agentworkspace.Asset{ID: "result-1", ProjectID: "agent-1", Kind: "video", Name: "result.mp4", ContentType: "video/mp4", ObjectKey: "agent-assets/alice/agent-1/result-1.mp4", ByteSize: 1, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: job.ID, ProjectID: job.ProjectID, NodeID: job.NodeID, NodeVersion: job.NodeVersion, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "generation-heal", Now: now}); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{projects: projects, assets: assets, videos: &scriptVideoService{}, now: func() time.Time { return now.Add(10 * time.Second) }}
	if _, err := coordinator.poll(context.Background(), "alice", job); err != nil {
		t.Fatalf("poll err=%v", err)
	}
	task, err := projects.GenerationTask(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || task.State != agentworkspace.GenerationAwaitingConfirmation || len(task.Results) != 1 {
		t.Fatalf("generation=%+v err=%v", task, err)
	}
}

func TestAgentVideoSubmitReplayRecoversMissingGeneration(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 20, 0, 0, time.UTC)
	base := agentworkspace.NewMemoryStore()
	if err := base.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	projects := &agentVideoFaultStore{Store: base, failCreateGeneration: 1}
	executor := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	coordinator := &agentVideoCoordinator{projects: projects, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{localExecutor: executor}, now: func() time.Time { return now }}
	request := agentVideoJobRequest{NodeID: "video-1", IdempotencyKey: "video-replay", Prompt: "雨夜追逐", Provider: "doubao_local_executor", ModelID: "Seedance 2.5", Parameters: map[string]any{"ratio": "16:9", "resolution": "1080P", "durationSeconds": 10, "count": 1}}
	first, err := coordinator.submit(context.Background(), "alice", "agent-1", request)
	if err == nil || first.ID == "" {
		t.Fatalf("first=%+v err=%v, want interrupted generic creation", first, err)
	}
	second, err := coordinator.submit(context.Background(), "alice", "agent-1", request)
	if err != nil || second.ID != first.ID || second.ProviderTaskID == "" {
		t.Fatalf("second=%+v err=%v", second, err)
	}
	task, err := base.GenerationTask(context.Background(), "alice", "agent-1", second.ID)
	if err != nil || task.State != agentworkspace.GenerationRunning {
		t.Fatalf("generation=%+v err=%v", task, err)
	}
}

func TestAgentVideoSubmitReplayAbandonsUncertainProviderSubmission(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 25, 0, 0, time.UTC)
	base := agentworkspace.NewMemoryStore()
	if err := base.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	projects := &agentVideoFaultStore{Store: base, failProviderTaskPersist: 1}
	executor := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	clock := now
	coordinator := &agentVideoCoordinator{projects: projects, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{localExecutor: executor}, ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return clock }}
	request := agentVideoJobRequest{NodeID: "video-1", IdempotencyKey: "video-uncertain", Prompt: "雨夜追逐", Provider: "doubao_local_executor", ModelID: "Seedance 2.5", Parameters: map[string]any{"ratio": "16:9", "resolution": "1080P", "durationSeconds": 10, "count": 1}}
	if _, err := coordinator.submit(context.Background(), "alice", "agent-1", request); err == nil {
		t.Fatal("first submit succeeded, want provider-task persistence failure")
	}
	jobs, err := base.ListVideoJobs(context.Background(), "alice", "agent-1")
	if err != nil || len(jobs) != 1 || jobs[0].State != agentworkspace.VideoJobSubmitting {
		t.Fatalf("jobs=%+v err=%v", jobs, err)
	}
	job, err := coordinator.submit(context.Background(), "alice", "agent-1", request)
	if err != nil || job.State != agentworkspace.VideoJobSubmitting || job.ProviderTaskID != "" {
		t.Fatalf("replay job=%+v err=%v", job, err)
	}
	clock = clock.Add(agentVideoSubmissionLease + time.Second)
	job, err = coordinator.poll(context.Background(), "alice", job)
	if err != nil || job.State != agentworkspace.VideoJobAbandoned || job.ProviderTaskID != "" {
		t.Fatalf("expired submission job=%+v err=%v", job, err)
	}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return clock }}, coordinator)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, agentVideoJSONRequest(t, clock, http.MethodGet, "/api/agent/tasks/agent-1/video-jobs/"+job.ID, ""))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"abandoned"`) {
		t.Fatalf("detail=%d %s", response.Code, response.Body.String())
	}
	task, err := base.GenerationTask(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || task.State != agentworkspace.GenerationFailed {
		t.Fatalf("generation=%+v err=%v", task, err)
	}
}

func TestAgentVideoSubmitReplayDoesNotAbandonLiveProviderCall(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 25, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}

	providerEntered := make(chan struct{})
	providerRelease := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(providerRelease) })
	var enteredOnce sync.Once
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		enteredOnce.Do(func() { close(providerEntered) })
		<-providerRelease
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"task_id":"provider-job","status":"queued"}`)),
			Request:    req,
		}, nil
	})}
	resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{Provider: batchfactoryv11.VideoProviderAutoDLH3, APIKey: "secret", Model: batchfactoryv11.AutoDLH3Model, CreateURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", TasksURL: "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}", WorkflowID: batchfactoryv11.AutoDLH3NoImageWorkflow}, nil
	}
	var clockNanos atomic.Int64
	clockNanos.Store(now.UnixNano())
	coordinator := &agentVideoCoordinator{
		projects: projects,
		assets:   agentworkspace.NewMemoryAssetStore(),
		videos:   &scriptVideoService{resolver: resolver, client: client},
		now:      func() time.Time { return time.Unix(0, clockNanos.Load()).UTC() },
	}
	request := agentVideoJobRequest{NodeID: "video-1", IdempotencyKey: "video-live-submit", Prompt: "雨夜追逐", ModelKey: "minimax-h3-video", Duration: 10, AspectRatio: "16:9", Resolution: "720P"}

	type submitResult struct {
		job agentworkspace.VideoJob
		err error
	}
	firstResult := make(chan submitResult, 1)
	go func() {
		job, err := coordinator.submit(context.Background(), "alice", "agent-1", request)
		firstResult <- submitResult{job: job, err: err}
	}()
	select {
	case <-providerEntered:
	case result := <-firstResult:
		t.Fatalf("provider submit returned before network call: job=%+v err=%v", result.job, result.err)
	case <-time.After(time.Second):
		t.Fatal("provider submit was not entered")
	}

	clockNanos.Store(now.Add(2*time.Minute + time.Second).UnixNano())
	replayed, err := coordinator.submit(context.Background(), "alice", "agent-1", request)
	if err != nil || replayed.State != agentworkspace.VideoJobSubmitting {
		t.Fatalf("replayed=%+v err=%v, want live submission to remain submitting", replayed, err)
	}

	releaseOnce.Do(func() { close(providerRelease) })
	result := <-firstResult
	if result.err != nil || result.job.State != agentworkspace.VideoJobQueued || result.job.ProviderTaskID == "" {
		t.Fatalf("first submit=%+v err=%v, want accepted provider task to persist", result.job, result.err)
	}
}

func TestAgentVideoSubmitFailureResponseUsesDurableSubmittingJob(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 26, 0, 0, time.UTC)
	base := agentworkspace.NewMemoryStore()
	if err := base.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{projects: &agentVideoFaultStore{Store: base, failProviderTaskPersist: 1}, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{localExecutor: localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })}, ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return now }}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs", `{"nodeId":"video-1","idempotencyKey":"video-durable-response","prompt":"雨夜追逐","provider":"doubao_local_executor","modelId":"Seedance 2.5","parameters":{"ratio":"16:9","resolution":"1080P","durationSeconds":10,"count":1}}`))
	if recorder.Code != http.StatusBadGateway || !strings.Contains(recorder.Body.String(), `"state":"submitting"`) || strings.Contains(recorder.Body.String(), `"state":"queued"`) {
		t.Fatalf("response=%d %s", recorder.Code, recorder.Body.String())
	}
}

func TestAgentVideoPollReconcilesTerminalVideoState(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 30, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		videoState agentworkspace.VideoJobState
		want       agentworkspace.GenerationState
	}{{"failed", agentworkspace.VideoJobFailed, agentworkspace.GenerationFailed}, {"cancelled", agentworkspace.VideoJobCanceled, agentworkspace.GenerationCancelled}} {
		t.Run(tc.name, func(t *testing.T) {
			projects := agentworkspace.NewMemoryStore()
			if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
				t.Fatal(err)
			}
			job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "video-" + tc.name, ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "model", IdempotencyKey: "video-" + tc.name, Now: now})
			if err != nil {
				t.Fatal(err)
			}
			job = advanceAgentVideoJobForTest(t, projects, now, job, agentworkspace.VideoJobRunning)
			job.State = tc.videoState
			job, err = projects.UpdateVideoJob(context.Background(), "alice", job, now.Add(10*time.Second))
			if err != nil {
				t.Fatal(err)
			}
			task, err := projects.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: job.ID, ProjectID: job.ProjectID, NodeID: job.NodeID, NodeVersion: job.NodeVersion, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "generation-" + tc.name, Now: now})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = projects.TransitionGenerationTask(context.Background(), "alice", job.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)}); err != nil {
				t.Fatal(err)
			}
			coordinator := &agentVideoCoordinator{projects: projects, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{}, now: func() time.Time { return now.Add(20 * time.Second) }}
			if _, err := coordinator.poll(context.Background(), "alice", job); err != nil {
				t.Fatalf("poll err=%v", err)
			}
			stored, err := projects.GenerationTask(context.Background(), "alice", "agent-1", job.ID)
			if err != nil || stored.State != tc.want {
				t.Fatalf("generation=%+v err=%v want=%s", stored, err, tc.want)
			}
		})
	}
}

func TestAgentVideoFailureMirrorsStoredCancelRaceWinner(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 40, 0, 0, time.UTC)
	base := agentworkspace.NewMemoryStore()
	if err := base.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	job, err := base.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "video-race", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "model", IdempotencyKey: "video-race", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job = advanceAgentVideoJobForTest(t, base, now, job, agentworkspace.VideoJobRunning)
	task, err := base.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: job.ID, ProjectID: job.ProjectID, NodeID: job.NodeID, NodeVersion: job.NodeVersion, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "generation-race", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = base.TransitionGenerationTask(context.Background(), "alice", job.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	faults := &agentVideoFaultStore{Store: base, raceVideoFailure: true}
	coordinator := &agentVideoCoordinator{projects: faults, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{}, now: func() time.Time { return now.Add(30 * time.Second) }}
	if _, err := coordinator.fail(context.Background(), "alice", job, errors.New("provider failed")); err != nil {
		t.Fatalf("fail err=%v", err)
	}
	stored, err := base.GenerationTask(context.Background(), "alice", "agent-1", job.ID)
	if err != nil || stored.State != agentworkspace.GenerationRunning {
		t.Fatalf("generation=%+v err=%v, want running while video cancel wins", stored, err)
	}
}

func TestAgentVideoConfirmReplayCompletesAppliedGeneration(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 50, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "video-confirm-replay", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "model", IdempotencyKey: "video-confirm-replay", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job = advanceAgentVideoJobForTest(t, projects, now, job, agentworkspace.VideoJobAwaitingConfirmation)
	task, err := projects.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: job.ID, ProjectID: job.ProjectID, NodeID: job.NodeID, NodeVersion: job.NodeVersion, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "generation-confirm-replay", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	task, err = projects.TransitionGenerationTask(context.Background(), "alice", job.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = projects.TransitionGenerationTask(context.Background(), "alice", job.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: []agentworkspace.GenerationResult{{ID: "result-1", AssetID: "result-1", Kind: "video"}}, SelectedResultID: "result-1", Now: now.Add(2 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err = projects.ConfirmVideoJob(context.Background(), "alice", job.ProjectID, job.ID, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{projects: projects, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{}, ingestor: &agentVideoAssetIngestor{}, now: func() time.Time { return now.Add(4 * time.Second) }}
	mux := http.NewServeMux()
	RegisterAgentVideoJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, coordinator)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, agentVideoJSONRequest(t, now, http.MethodPost, "/api/agent/tasks/agent-1/video-jobs/"+job.ID+"/confirm", ""))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"state":"applied"`) {
		t.Fatalf("confirm replay=%d %s", recorder.Code, recorder.Body.String())
	}
	completed, err := projects.GenerationTask(context.Background(), "alice", job.ProjectID, job.ID)
	if err != nil || completed.State != agentworkspace.GenerationCompleted {
		t.Fatalf("generation=%+v err=%v", completed, err)
	}
}

func TestAgentVideoAppliedReconciliationHealsRunningGeneration(t *testing.T) {
	now := time.Date(2026, 10, 21, 20, 10, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	job, err := projects.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "video-applied-heal", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "model", IdempotencyKey: "video-applied-heal", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job = advanceAgentVideoJobForTest(t, projects, now, job, agentworkspace.VideoJobApplied)
	assets := agentworkspace.NewMemoryAssetStore()
	if err := assets.CreateAsset(context.Background(), "alice", agentworkspace.Asset{ID: "result-1", ProjectID: "agent-1", Kind: "video", Name: "result.mp4", ContentType: "video/mp4", ObjectKey: "agent-assets/alice/agent-1/result-1.mp4", ByteSize: 1, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	task, err := projects.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{ID: job.ID, ProjectID: job.ProjectID, NodeID: job.NodeID, NodeVersion: job.NodeVersion, Kind: "video", Provider: "provider", ModelID: "model", IdempotencyKey: "generation-applied-heal", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = projects.TransitionGenerationTask(context.Background(), "alice", job.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	coordinator := &agentVideoCoordinator{projects: projects, assets: assets, videos: &scriptVideoService{}, now: func() time.Time { return now.Add(10 * time.Second) }}
	if err := coordinator.reconcileGeneration(context.Background(), "alice", job); err != nil {
		t.Fatalf("reconcile err=%v", err)
	}
	completed, err := projects.GenerationTask(context.Background(), "alice", job.ProjectID, job.ID)
	if err != nil || completed.State != agentworkspace.GenerationCompleted {
		t.Fatalf("generation=%+v err=%v", completed, err)
	}
}

func TestAgentVideoRetryReplayRecoversMissingGeneration(t *testing.T) {
	now := time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC)
	base := agentworkspace.NewMemoryStore()
	if err := base.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.Canvas{Version: 1, Nodes: []agentworkspace.Node{{ID: "video-1", Type: "video", Version: 1, Status: "draft"}}}}); err != nil {
		t.Fatal(err)
	}
	original, err := base.CreateVideoJob(context.Background(), "alice", agentworkspace.CreateVideoJobInput{ID: "video-original", ProjectID: "agent-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "local-doubao-executor-video", Settings: map[string]any{"prompt": "雨夜追逐", "provider": "doubao_local_executor", "modelId": "Seedance 2.5", "durationSeconds": 10}, IdempotencyKey: "video-original", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	original.State, original.ErrorMessage = agentworkspace.VideoJobFailed, "provider failed"
	if _, err = base.UpdateVideoJob(context.Background(), "alice", original, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	projects := &agentVideoFaultStore{Store: base, failCreateGeneration: 1}
	executor := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	coordinator := &agentVideoCoordinator{projects: projects, assets: agentworkspace.NewMemoryAssetStore(), videos: &scriptVideoService{localExecutor: executor}, now: func() time.Time { return now.Add(2 * time.Second) }}
	first, err := coordinator.retry(context.Background(), "alice", "agent-1", original.ID, "video-retry")
	if err == nil || first.ID == "" {
		t.Fatalf("first retry=%+v err=%v, want interrupted generic creation", first, err)
	}
	second, err := coordinator.retry(context.Background(), "alice", "agent-1", original.ID, "video-retry")
	if err != nil || second.ID != first.ID || second.ProviderTaskID == "" {
		t.Fatalf("second retry=%+v err=%v", second, err)
	}
	task, err := base.GenerationTask(context.Background(), "alice", "agent-1", second.ID)
	if err != nil || task.RetryOfTaskID != original.ID || task.State != agentworkspace.GenerationRunning {
		t.Fatalf("generation=%+v err=%v", task, err)
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
