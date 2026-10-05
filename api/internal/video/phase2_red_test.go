package video

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type phase2ConfigStore struct {
	cfg ProviderConfig
	err error
}

func (s *phase2ConfigStore) GetProviderConfig(context.Context, string, string) (ProviderConfig, error) {
	if s.err != nil {
		return ProviderConfig{}, s.err
	}
	return s.cfg, nil
}
func (s *phase2ConfigStore) UpsertProviderConfig(context.Context, ProviderConfig) error { return nil }

type phase2Provider struct {
	mu           sync.Mutex
	submitCalls  int
	submitInputs []SubmitRequest
	pollResult   PollResult
	cancelResult CancelResult
	cancelErr    error
	probeErr     error
}

func (p *phase2Provider) Submit(_ context.Context, req SubmitRequest) (SubmitResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.submitCalls++
	p.submitInputs = append(p.submitInputs, req)
	return SubmitResult{ProviderJobID: fmt.Sprintf("remote-%d", p.submitCalls), Status: TaskQueued}, nil
}
func (p *phase2Provider) Poll(context.Context, string) (PollResult, error) { return p.pollResult, nil }
func (p *phase2Provider) Cancel(context.Context, string) (CancelResult, error) {
	return p.cancelResult, p.cancelErr
}
func (p *phase2Provider) Probe(context.Context) error { return p.probeErr }

func (p *phase2Provider) inputs() []SubmitRequest {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]SubmitRequest(nil), p.submitInputs...)
}

type phase2Factory struct{ provider Provider }
func (f phase2Factory) Build(ProviderConfig, string) (Provider, error) { return f.provider, nil }

type countingFinalPromptSource struct {
	mu     sync.Mutex
	prompt FinalPrompt
	calls  int
}
func (s *countingFinalPromptSource) ResolveFinalPrompt(context.Context, int64, int64) (FinalPrompt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.prompt, nil
}
func (s *countingFinalPromptSource) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func TestProviderStatusStatesAndSecretNeverReturned(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef")
	ciphertext, nonce, err := EncryptSecret(master, "status-secret")
	if err != nil { t.Fatal(err) }
	cfg := ProviderConfig{
		ProviderKey: ProviderPersonalAPI,
		Model: ModelYD20Mini,
		EncryptedSecret: ciphertext,
		SecretNonce: nonce,
		Enabled: true,
		CreateURL: "https://provider.example/create",
		TasksURL: "https://provider.example/tasks",
	}

	t.Run("unconfigured", func(t *testing.T) {
		store := &phase2ConfigStore{err: providerError(ErrorProviderUnconfigured, "missing", nil)}
		service := NewConfigServiceWithProviders(store, master, phase2Factory{provider: &phase2Provider{}})
		status, err := service.Status(context.Background(), ProviderPersonalAPI, ModelYD20Mini)
		if err != nil { t.Fatal(err) }
		if status.Status != ProviderStatusUnconfigured || status.Configured {
			t.Fatalf("status = %+v", status)
		}
	})

	for _, tc := range []struct {
		name string
		probeErr error
		want ProviderAvailability
	}{
		{name: "available", want: ProviderStatusAvailable},
		{name: "auth_failed", probeErr: providerError(ErrorProviderAuthFailed, "bad credential", nil), want: ProviderStatusAuthFailed},
		{name: "unavailable", probeErr: providerError(ErrorProviderUnavailable, "network unavailable", nil), want: ProviderStatusUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := &phase2Provider{probeErr: tc.probeErr}
			service := NewConfigServiceWithProviders(&phase2ConfigStore{cfg: cfg}, master, phase2Factory{provider: provider})
			status, err := service.Status(context.Background(), ProviderPersonalAPI, ModelYD20Mini)
			if err != nil { t.Fatal(err) }
			if status.Status != tc.want || !status.Configured {
				t.Fatalf("status = %+v want %s", status, tc.want)
			}
			payload, err := json.Marshal(status)
			if err != nil { t.Fatal(err) }
			for _, forbidden := range []string{"status-secret", "encryptedSecret", "secretNonce"} {
				if strings.Contains(string(payload), forbidden) {
					t.Fatalf("status leaked %q: %s", forbidden, payload)
				}
			}
		})
	}
}

func TestVideoRetryCreatesNewAttemptWithoutRerunningFinalPrompt(t *testing.T) {
	store, master := configuredRedStore(t)
	provider := &phase2Provider{}
	promptSource := &countingFinalPromptSource{prompt: FinalPrompt{StageRunID: 44, PromptVersion: 3, InputRevision: "rev-44", Text: "frozen final prompt"}}
	service := NewService(store, promptSource, phase2Factory{provider: provider}, &redArtifactStore{}, master)

	first, err := service.Start(context.Background(), StartRequest{BatchProjectID: 9, BookID: 7, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, RequestID: "req-1"})
	if err != nil { t.Fatal(err) }
	if first.Job.FinalPromptText != "frozen final prompt" {
		t.Fatalf("job final prompt snapshot = %q", first.Job.FinalPromptText)
	}
	first.Task.Status = TaskFailed
	if err := store.UpdateProductionTask(context.Background(), first.Task); err != nil { t.Fatal(err) }
	first.Job.Status = JobFailed
	if err := store.UpdateProductionJob(context.Background(), first.Job); err != nil { t.Fatal(err) }

	retried, err := service.RetryTask(context.Background(), first.Task.ID, "req-2")
	if err != nil { t.Fatal(err) }
	if retried.Job.ID != first.Job.ID || retried.Task.ID == first.Task.ID || retried.Task.Attempt != 2 {
		t.Fatalf("retry = %+v first = %+v", retried, first)
	}
	oldTask, err := store.GetProductionTask(context.Background(), first.Task.ID)
	if err != nil { t.Fatal(err) }
	if oldTask.Attempt != 1 || oldTask.Status != TaskFailed {
		t.Fatalf("old attempt mutated: %+v", oldTask)
	}
	if promptSource.count() != 1 {
		t.Fatalf("FINAL_PROMPT source calls = %d want 1; retry must not rerun generation", promptSource.count())
	}
	inputs := provider.inputs()
	if len(inputs) != 2 || inputs[1].Prompt != "frozen final prompt" {
		t.Fatalf("provider inputs = %+v", inputs)
	}
}

func TestVideoCancelOnlyMarksCancelledWhenProviderAccepts(t *testing.T) {
	makeService := func(t *testing.T, cancel CancelResult, cancelErr error) (*Service, *redStore, StartResult) {
		t.Helper()
		store, master := configuredRedStore(t)
		provider := &phase2Provider{cancelResult: cancel, cancelErr: cancelErr}
		service := NewService(store, redFinalPromptSource{prompt: FinalPrompt{StageRunID: 55, PromptVersion: 1, InputRevision: "rev-55", Text: "prompt"}}, phase2Factory{provider: provider}, &redArtifactStore{}, master)
		started, err := service.Start(context.Background(), StartRequest{BatchProjectID: 3, BookID: 4, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, RequestID: "cancel-start"})
		if err != nil { t.Fatal(err) }
		return service, store, started
	}

	t.Run("accepted", func(t *testing.T) {
		service, store, started := makeService(t, CancelResult{Accepted: true, Status: TaskCancelled}, nil)
		cancelled, err := service.CancelTask(context.Background(), started.Task.ID)
		if err != nil { t.Fatal(err) }
		if cancelled.Status != TaskCancelled { t.Fatalf("task = %+v", cancelled) }
		job, err := store.GetProductionJob(context.Background(), started.Job.ID)
		if err != nil { t.Fatal(err) }
		if job.Status != JobCancelled { t.Fatalf("job = %+v", job) }
	})

	t.Run("unsupported", func(t *testing.T) {
		service, store, started := makeService(t, CancelResult{Accepted: false}, nil)
		_, err := service.CancelTask(context.Background(), started.Task.ID)
		if err == nil { t.Fatal("expected cancel unsupported error") }
		var providerErr *ProviderError
		if !errors.As(err, &providerErr) || providerErr.Code != ErrorProviderCancelUnsupported {
			t.Fatalf("err = %v", err)
		}
		task, err := store.GetProductionTask(context.Background(), started.Task.ID)
		if err != nil { t.Fatal(err) }
		if task.Status == TaskCancelled { t.Fatalf("unsupported cancellation mutated task: %+v", task) }
	})
}

func TestYFAISeedanceSubmitAndPollProtocol(t *testing.T) {
	const secret = "yf-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+secret {
			t.Fatalf("Authorization = %q", got)
		}
		switch r.URL.Path {
		case "/v1/media/generate":
			if r.Method != http.MethodPost { t.Fatalf("method = %s", r.Method) }
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { t.Fatal(err) }
			if payload["model"] != "seedance-2-0-official" || payload["prompt"] != "seedance prompt" {
				t.Fatalf("payload = %+v", payload)
			}
			params, _ := payload["params"].(map[string]any)
			if params["mode"] != "reference" || params["duration"] != "8" || params["resolution"] != "720p" || params["aspect_ratio"] != "9:16" {
				t.Fatalf("params = %+v", params)
			}
			_, _ = w.Write([]byte(`{"task_id":"yf-123"}`))
		case "/v1/tasks/yf-123":
			_, _ = w.Write([]byte(`{"status":"completed","output_url":"https://cdn.example/yf.mp4"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := NewYFAISeedanceProvider(ProviderConfig{ProviderKey: ProviderYFAISeedance, Model: "seedance-2-0-official", CreateURL: server.URL + "/v1/media/generate", TasksURL: server.URL + "/v1/tasks", Enabled: true}, secret, server.Client())
	if err != nil { t.Fatal(err) }
	submit, err := provider.Submit(context.Background(), SubmitRequest{Model: "seedance-2-0-official", Prompt: "seedance prompt", DurationSeconds: 8, Resolution: "720p", AspectRatio: "9:16", ReferenceImageURLs: []string{"https://img.example/ref.png"}})
	if err != nil { t.Fatal(err) }
	if submit.ProviderJobID != "yf-123" || submit.Status != TaskQueued { t.Fatalf("submit = %+v", submit) }
	poll, err := provider.Poll(context.Background(), "yf-123")
	if err != nil { t.Fatal(err) }
	if poll.Status != TaskSucceeded || poll.ArtifactURL != "https://cdn.example/yf.mp4" { t.Fatalf("poll = %+v", poll) }
}

func TestAutoDLComfyUISubmitAndPollProtocol(t *testing.T) {
	const secret = "autodl-raw-key"
	var submitPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != secret {
			t.Fatalf("Authorization = %q want raw key", got)
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/workflow/"):
			submitPath = r.URL.Path
			_, _ = w.Write([]byte(`{"id":"ad-1","status":"queued"}`))
		case r.URL.Path == "/result/ad-1":
			_, _ = w.Write([]byte(`{"status":"completed","results":["https://cdn.example/ad.mp4"]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := NewAutoDLComfyUIProvider(ProviderConfig{ProviderKey: ProviderAutoDLComfyUI, Model: "minimax-h3-video", CreateURL: server.URL + "/workflow/{workflow}", TasksURL: server.URL + "/result/{id}", Enabled: true}, secret, server.Client())
	if err != nil { t.Fatal(err) }
	submit, err := provider.Submit(context.Background(), SubmitRequest{Model: "minimax-h3-video", Prompt: "h3 prompt", DurationSeconds: 10, Resolution: "480p竖"})
	if err != nil { t.Fatal(err) }
	if submit.ProviderJobID != "ad-1" || !strings.HasSuffix(submitPath, "/minimax_h3_lightx2v_no_pic") {
		t.Fatalf("submit = %+v path=%s", submit, submitPath)
	}
	poll, err := provider.Poll(context.Background(), "ad-1")
	if err != nil { t.Fatal(err) }
	if poll.Status != TaskSucceeded || poll.ArtifactURL != "https://cdn.example/ad.mp4" { t.Fatalf("poll = %+v", poll) }
}
