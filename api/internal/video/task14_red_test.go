package video

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestProviderModelMapping(t *testing.T) {
	tests := []struct {
		model string
		want  string
	}{
		{ModelYD20Mini, ProviderPersonalAPI},
		{"yd2-mini-video", ProviderPersonalAPI},
		{"seedance-2-0-official", ProviderYFAISeedance},
		{"minimax-h3-video", ProviderAutoDLComfyUI},
		{"doubao-seedance", ProviderDoubaoLocalExecutor},
	}
	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			got, ok := ProviderForModel(tc.model)
			if !ok || got != tc.want {
				t.Fatalf("ProviderForModel(%q) = %q,%v want %q,true", tc.model, got, ok, tc.want)
			}
		})
	}
	if _, ok := ProviderForModel("unknown-model"); ok {
		t.Fatal("unknown model must not resolve to a provider")
	}
}

func TestProviderUnconfigured(t *testing.T) {
	_, err := NewPersonalAPIProvider(ProviderConfig{
		ProviderKey: ProviderPersonalAPI,
		Model:       ModelYD20Mini,
		CreateURL:   "https://example.invalid/create",
		TasksURL:    "https://example.invalid/tasks",
	}, "", http.DefaultClient)
	if err == nil {
		t.Fatal("expected unconfigured provider error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != ErrorProviderUnconfigured {
		t.Fatalf("err = %v, want code %s", err, ErrorProviderUnconfigured)
	}
}

func TestSecretEncryptedAndNotLeakedByView(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef")
	ciphertext, nonce, err := EncryptSecret(master, "top-secret-provider-key")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(ciphertext), "top-secret-provider-key") {
		t.Fatal("ciphertext contains plaintext secret")
	}
	plaintext, err := DecryptSecret(master, ciphertext, nonce)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "top-secret-provider-key" {
		t.Fatalf("plaintext = %q", plaintext)
	}

	cfg := ProviderConfig{
		ProviderKey:     ProviderPersonalAPI,
		Model:           ModelYD20Mini,
		EncryptedSecret: ciphertext,
		SecretNonce:     nonce,
		CreateURL:       "https://provider.example/create",
		TasksURL:        "https://provider.example/tasks",
		Enabled:         true,
	}
	payload, err := json.Marshal(cfg.View())
	if err != nil {
		t.Fatal(err)
	}
	body := string(payload)
	for _, forbidden := range []string{"top-secret-provider-key", hex.EncodeToString(ciphertext), hex.EncodeToString(nonce), "encryptedSecret", "secretNonce"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("safe provider view leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"configured":true`) {
		t.Fatalf("safe provider view = %s, want configured=true", body)
	}
}

func TestSecretDecryptFailureIsExplicit(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef")
	ciphertext, nonce, err := EncryptSecret(master, "secret")
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[0] ^= 0xff
	_, err = DecryptSecret(master, ciphertext, nonce)
	if err == nil {
		t.Fatal("expected decrypt failure")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != ErrorProviderConfigDecryptFailed {
		t.Fatalf("err = %v, want code %s", err, ErrorProviderConfigDecryptFailed)
	}
}

func TestPersonalAPISubmitSuccessAndFailure(t *testing.T) {
	const secret = "server-only-api-key"
	var fail bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/create" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+secret {
			t.Fatalf("Authorization = %q", got)
		}
		if fail {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, `{"error":"provider exploded; leaked=%s"}`, secret)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"task_id":"remote-123","status":"queued"}`))
	}))
	defer server.Close()

	provider, err := NewPersonalAPIProvider(ProviderConfig{
		ProviderKey: ProviderPersonalAPI,
		Model:       ModelYD20Mini,
		CreateURL:   server.URL + "/create",
		TasksURL:    server.URL + "/tasks",
		ResultURL:   server.URL + "/result/{id}",
	}, secret, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	result, err := provider.Submit(context.Background(), SubmitRequest{Model: ModelYD20Mini, Prompt: "video prompt", RequestID: "req-1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProviderJobID != "remote-123" || result.Status != TaskQueued {
		t.Fatalf("submit result = %+v", result)
	}

	fail = true
	_, err = provider.Submit(context.Background(), SubmitRequest{Model: ModelYD20Mini, Prompt: "video prompt", RequestID: "req-2"})
	if err == nil {
		t.Fatal("expected submit failure")
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "provider exploded") {
		t.Fatalf("provider failure leaked response body/secret: %v", err)
	}
}

func TestPersonalAPIPollQueuedRunningSucceededFailed(t *testing.T) {
	state := "queued"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/tasks/"):
			_, _ = fmt.Fprintf(w, `{"id":"remote-123","status":%q}`, state)
		case strings.HasPrefix(r.URL.Path, "/result/"):
			_, _ = w.Write([]byte(`{"status":"succeeded","url":"https://cdn.example/video.mp4"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := NewPersonalAPIProvider(ProviderConfig{
		ProviderKey: ProviderPersonalAPI,
		Model:       ModelYD20Mini,
		CreateURL:   server.URL + "/create",
		TasksURL:    server.URL + "/tasks",
		ResultURL:   server.URL + "/result/{id}",
	}, "secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		remote string
		want   TaskStatus
	}{
		{"queued", TaskQueued},
		{"processing", TaskRunning},
		{"completed", TaskSucceeded},
		{"failed", TaskFailed},
	} {
		state = tc.remote
		result, err := provider.Poll(context.Background(), "remote-123")
		if err != nil {
			t.Fatalf("state %s: %v", tc.remote, err)
		}
		if result.Status != tc.want {
			t.Fatalf("state %s => %s want %s", tc.remote, result.Status, tc.want)
		}
		if tc.want == TaskSucceeded && result.ArtifactURL != "https://cdn.example/video.mp4" {
			t.Fatalf("artifact URL = %q", result.ArtifactURL)
		}
	}
}

type redFinalPromptSource struct {
	prompt FinalPrompt
}

func (s redFinalPromptSource) ResolveFinalPrompt(context.Context, int64, int64) (FinalPrompt, error) {
	return s.prompt, nil
}

type redProvider struct {
	mu          sync.Mutex
	submitCalls int
	pollResult  PollResult
	submitErr   error
}

func (p *redProvider) Submit(context.Context, SubmitRequest) (SubmitResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.submitCalls++
	if p.submitErr != nil {
		return SubmitResult{}, p.submitErr
	}
	return SubmitResult{ProviderJobID: "remote-123", Status: TaskQueued}, nil
}

func (p *redProvider) Poll(context.Context, string) (PollResult, error) { return p.pollResult, nil }
func (p *redProvider) Cancel(context.Context, string) (CancelResult, error) {
	return CancelResult{Accepted: false}, nil
}

func (p *redProvider) calls() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.submitCalls
}

type redFactory struct{ provider Provider }

func (f redFactory) Build(ProviderConfig, string) (Provider, error) { return f.provider, nil }

type redArtifactStore struct {
	artifact Artifact
	calls    int
}

func (s *redArtifactStore) Persist(context.Context, string, string) (Artifact, error) {
	s.calls++
	return s.artifact, nil
}

type redStore struct {
	mu          sync.Mutex
	config      ProviderConfig
	jobsByKey   map[string]ProductionJob
	jobsByID    map[int64]ProductionJob
	tasksByID   map[int64]ProductionTask
	nextJobID   int64
	nextTaskID  int64
	recoverable []ProductionTask
}

func newRedStore(config ProviderConfig) *redStore {
	return &redStore{config: config, jobsByKey: map[string]ProductionJob{}, jobsByID: map[int64]ProductionJob{}, tasksByID: map[int64]ProductionTask{}, nextJobID: 1, nextTaskID: 1}
}

func (s *redStore) GetProviderConfig(context.Context, string, string) (ProviderConfig, error) { return s.config, nil }
func (s *redStore) CreateOrGetProductionJob(_ context.Context, job ProductionJob) (ProductionJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.jobsByKey[job.IdempotencyKey]; ok {
		return existing, false, nil
	}
	job.ID = s.nextJobID
	s.nextJobID++
	s.jobsByKey[job.IdempotencyKey] = job
	s.jobsByID[job.ID] = job
	return job, true, nil
}
func (s *redStore) CreateProductionTask(_ context.Context, task ProductionTask) (ProductionTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task.ID = s.nextTaskID
	s.nextTaskID++
	s.tasksByID[task.ID] = task
	return task, nil
}
func (s *redStore) UpdateProductionTask(_ context.Context, task ProductionTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasksByID[task.ID] = task
	return nil
}
func (s *redStore) UpdateProductionJob(_ context.Context, job ProductionJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobsByID[job.ID] = job
	s.jobsByKey[job.IdempotencyKey] = job
	return nil
}
func (s *redStore) GetProductionTask(_ context.Context, id int64) (ProductionTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasksByID[id]
	if !ok {
		return ProductionTask{}, sql.ErrNoRows
	}
	return task, nil
}
func (s *redStore) GetProductionJob(_ context.Context, id int64) (ProductionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobsByID[id]
	if !ok {
		return ProductionJob{}, sql.ErrNoRows
	}
	return job, nil
}
func (s *redStore) LatestTaskForJob(_ context.Context, jobID int64) (ProductionTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var latest ProductionTask
	for _, task := range s.tasksByID {
		if task.ProductionJobID == jobID && task.ID > latest.ID {
			latest = task
		}
	}
	if latest.ID == 0 {
		return ProductionTask{}, sql.ErrNoRows
	}
	return latest, nil
}
func (s *redStore) ListRecoverableTasks(context.Context, int) ([]ProductionTask, error) {
	return append([]ProductionTask(nil), s.recoverable...), nil
}

func configuredRedStore(t *testing.T) (*redStore, []byte) {
	t.Helper()
	master := []byte("0123456789abcdef0123456789abcdef")
	ciphertext, nonce, err := EncryptSecret(master, "provider-secret")
	if err != nil {
		t.Fatal(err)
	}
	return newRedStore(ProviderConfig{
		ProviderKey:     ProviderPersonalAPI,
		Model:           ModelYD20Mini,
		EncryptedSecret: ciphertext,
		SecretNonce:     nonce,
		CreateURL:       "https://provider.example/create",
		TasksURL:        "https://provider.example/tasks",
		Enabled:         true,
	}), master
}

func TestServiceDuplicateSubmitAndFinalPromptRevisionTracking(t *testing.T) {
	store, master := configuredRedStore(t)
	provider := &redProvider{}
	artifacts := &redArtifactStore{}
	inputSnapshot := `{"director":"revision-7"}`
	hash := sha256.Sum256([]byte(inputSnapshot))
	prompt := FinalPrompt{StageRunID: 901, PromptVersion: 7, InputRevision: hex.EncodeToString(hash[:]), Text: "compiled final prompt"}
	service := NewService(store, redFinalPromptSource{prompt: prompt}, redFactory{provider: provider}, artifacts, master)

	first, err := service.Start(context.Background(), StartRequest{BatchProjectID: 51, BookID: 31, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, RequestID: "req-1"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Start(context.Background(), StartRequest{BatchProjectID: 51, BookID: 31, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, RequestID: "req-2"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Job.ID != second.Job.ID || first.Task.ID != second.Task.ID {
		t.Fatalf("duplicate Start created new work: first=%+v second=%+v", first, second)
	}
	if provider.calls() != 1 {
		t.Fatalf("provider Submit calls = %d want 1", provider.calls())
	}
	if first.Job.FinalPromptStageRunID != 901 || first.Job.FinalPromptVersion != 7 || first.Job.InputRevision != prompt.InputRevision {
		t.Fatalf("job lost Final Prompt identity: %+v", first.Job)
	}
}

func TestServicePollQueuedRunningSucceededFailed(t *testing.T) {
	store, master := configuredRedStore(t)
	provider := &redProvider{}
	artifacts := &redArtifactStore{artifact: Artifact{Bucket: "video", ObjectKey: "task14/out.mp4", URL: "https://tos.example/task14/out.mp4"}}
	service := NewService(store, redFinalPromptSource{prompt: FinalPrompt{StageRunID: 1, PromptVersion: 1, InputRevision: "rev", Text: "prompt"}}, redFactory{provider: provider}, artifacts, master)
	started, err := service.Start(context.Background(), StartRequest{BatchProjectID: 51, BookID: 31, Provider: ProviderPersonalAPI, Model: ModelYD20Mini})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		result PollResult
		want   TaskStatus
	}{
		{PollResult{Status: TaskQueued}, TaskQueued},
		{PollResult{Status: TaskRunning}, TaskRunning},
		{PollResult{Status: TaskFailed, ErrorCode: ErrorProviderRequestFailed, ErrorMessage: "failed"}, TaskFailed},
		{PollResult{Status: TaskSucceeded, ArtifactURL: "https://provider.example/temp.mp4"}, TaskSucceeded},
	} {
		provider.pollResult = tc.result
		got, err := service.PollTask(context.Background(), started.Task.ID)
		if err != nil {
			t.Fatalf("poll %+v: %v", tc.result, err)
		}
		if got.Status != tc.want {
			t.Fatalf("poll %+v => %s want %s", tc.result, got.Status, tc.want)
		}
		if tc.want == TaskSucceeded {
			if got.OutputObjectKey != "task14/out.mp4" || artifacts.calls != 1 {
				t.Fatalf("succeeded task did not persist artifact: task=%+v calls=%d", got, artifacts.calls)
			}
		}
	}
}

func TestServiceSubmitFailureIsDurable(t *testing.T) {
	store, master := configuredRedStore(t)
	provider := &redProvider{submitErr: &ProviderError{Code: ErrorProviderRequestFailed, Message: "submit failed"}}
	service := NewService(store, redFinalPromptSource{prompt: FinalPrompt{StageRunID: 1, PromptVersion: 1, InputRevision: "rev", Text: "prompt"}}, redFactory{provider: provider}, &redArtifactStore{}, master)
	result, err := service.Start(context.Background(), StartRequest{BatchProjectID: 51, BookID: 31, Provider: ProviderPersonalAPI, Model: ModelYD20Mini})
	if err == nil {
		t.Fatal("expected submit failure")
	}
	if result.Task.Status != TaskFailed || result.Task.ErrorCode != ErrorProviderRequestFailed {
		t.Fatalf("failed submit not persisted on task: %+v", result.Task)
	}
}

func TestMySQLStoreRecoversPendingPollTasksAfterRestart(t *testing.T) {
	dsn := os.Getenv("TASK14_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TASK14_MYSQL_DSN is set by the Goose/MySQL CI job")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	store1 := NewMySQLStore(db)
	config := ProviderConfig{ProviderKey: ProviderPersonalAPI, Model: ModelYD20Mini, EncryptedSecret: []byte("cipher"), SecretNonce: []byte("nonce"), Enabled: true}
	if err := store1.UpsertProviderConfig(ctx, config); err != nil {
		t.Fatal(err)
	}
	job, _, err := store1.CreateOrGetProductionJob(ctx, ProductionJob{BatchProjectID: 51, BookID: 31, Status: JobRunning, InputRevision: "rev-restart", FinalPromptStageRunID: 901, FinalPromptVersion: 7, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, IdempotencyKey: "restart-contract"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := store1.CreateProductionTask(ctx, ProductionTask{ProductionJobID: job.ID, Attempt: 1, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, ProviderJobID: "remote-restart", Status: TaskRunning})
	if err != nil {
		t.Fatal(err)
	}

	store2 := NewMySQLStore(db)
	recoverable, err := store2.ListRecoverableTasks(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range recoverable {
		if got.ID == task.ID && got.ProviderJobID == "remote-restart" && got.Status == TaskRunning {
			return
		}
	}
	t.Fatalf("restarted store did not recover task %d: %+v", task.ID, recoverable)
}
