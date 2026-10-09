package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type fakeGenerationRuntime struct {
	ready      bool
	result     task9runtime.AdmissionResult
	run        task9runtime.GenerationRunStatus
	err        error
	req        task9runtime.GenerationRequest
	admitCalls int
}

func (f *fakeGenerationRuntime) Ready() bool { return f.ready }
func (f *fakeGenerationRuntime) AdmitGeneration(_ context.Context, req task9runtime.GenerationRequest) (task9runtime.AdmissionResult, error) {
	f.admitCalls++
	f.req = req
	return f.result, f.err
}
func (f *fakeGenerationRuntime) GenerationRun(context.Context, int64, int64) (task9runtime.GenerationRunStatus, error) {
	return f.run, f.err
}

type fakeGenerationService struct {
	project        generation.ProjectSummary
	book           generation.BookGenerationResult
	stage          generation.StageRun
	prompts        []generation.Prompt
	measurement    generation.AudioMeasurement
	measurementErr error
	measureReq     generation.AudioMeasurementRequest
	runBookReq     generation.RunBookRequest
	runBatchReq    generation.RunBatchRequest
	retryReq       generation.RetryStageRequest
}

func (f *fakeGenerationService) ProjectSummary(context.Context, int64) (generation.ProjectSummary, error) {
	return f.project, nil
}
func (f *fakeGenerationService) BookSummary(context.Context, int64, int64) (generation.BookGenerationResult, error) {
	return f.book, nil
}
func (f *fakeGenerationService) RunBook(_ context.Context, r generation.RunBookRequest) (generation.BookGenerationResult, error) {
	f.runBookReq = r
	return f.book, nil
}
func (f *fakeGenerationService) RunBatch(_ context.Context, r generation.RunBatchRequest) (generation.BatchGenerationResult, error) {
	f.runBatchReq = r
	return generation.BatchGenerationResult{BatchProjectID: r.BatchProjectID}, nil
}
func (f *fakeGenerationService) RetryStage(_ context.Context, r generation.RetryStageRequest) (generation.BookGenerationResult, error) {
	f.retryReq = r
	return f.book, nil
}
func (f *fakeGenerationService) StageResult(context.Context, int64, int64, generation.Stage) (generation.StageRun, error) {
	return f.stage, nil
}
func (f *fakeGenerationService) ListPrompts(context.Context) ([]generation.Prompt, error) {
	return f.prompts, nil
}
func (f *fakeGenerationService) AudioMeasurement(context.Context, int64, int64) (generation.AudioMeasurement, error) {
	return f.measurement, f.measurementErr
}
func (f *fakeGenerationService) MeasureAudio(_ context.Context, r generation.AudioMeasurementRequest) (generation.AudioMeasurement, error) {
	f.measureReq = r
	return f.measurement, f.measurementErr
}

func TestGenerationBookRouteBindsPathIDs(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: true, result: task9runtime.AdmissionResult{Run: task9runtime.RunRecord{ID: 9, BatchProjectID: 3}, TaskIDs: []int64{71}, Created: true, Dispatch: "queued"}}
	h := generationMutationHarness(runtime)
	body := bytes.NewBufferString(`{"hookEnabled":true,"directorMode":"normal","requestId":"r1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation", body)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if runtime.req.BatchProjectID != 3 || len(runtime.req.BookIDs) != 1 || runtime.req.BookIDs[0] != 11 || runtime.req.Action != "" {
		t.Fatalf("request=%#v", runtime.req)
	}
}

func TestGenerationRetryBindsTargetStage(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: true, result: task9runtime.AdmissionResult{Run: task9runtime.RunRecord{ID: 10, BatchProjectID: 3}, TaskIDs: []int64{72}, Created: true, Dispatch: "queued"}}
	h := generationMutationHarness(runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation/stages/DIRECTOR/retry", bytes.NewBufferString(`{"requestId":"retry-1","sourceBookRunId":70}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if runtime.req.RetryStage != string(generation.StageDirector) || len(runtime.req.BookIDs) != 1 || runtime.req.BookIDs[0] != 11 || runtime.req.SourceBookRunID != 70 {
		t.Fatalf("retry=%#v", runtime.req)
	}
}

func TestGenerationPromptListReturnsVersions(t *testing.T) {
	fake := &fakeGenerationService{prompts: []generation.Prompt{{Key: generation.PromptScript, Version: 2, Enabled: true, Content: "test-secret-prompt"}}}
	h := NewHandler(Dependencies{Generation: fake})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generation/prompts", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d", res.Code)
	}
	var payload struct {
		Prompts []generation.Prompt `json:"prompts"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Prompts) != 1 || payload.Prompts[0].Version != 2 || payload.Prompts[0].Content != "" {
		t.Fatalf("payload=%#v", payload)
	}
	if bytes.Contains(res.Body.Bytes(), []byte("test-secret-prompt")) {
		t.Fatalf("body leaked prompt: %s", res.Body.String())
	}
	if fake.prompts[0].Content != "test-secret-prompt" {
		t.Fatal("HTTP projection mutated service-owned prompt")
	}
}

func TestAudioMeasurementGetReturnsAuthoritativeDuration(t *testing.T) {
	fake := &fakeGenerationService{measurement: generation.AudioMeasurement{ID: 1, BatchProjectID: 3, BookID: 11, AudioAsset: "/audio/11.mp3", DurationMS: 28000, MeasuredAt: time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)}}
	h := NewHandler(Dependencies{Generation: fake})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/3/books/11/audio-measurement", nil)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if !bytes.Contains(res.Body.Bytes(), []byte(`"durationMs":28000`)) {
		t.Fatalf("body=%s", res.Body.String())
	}
}

func TestAudioMeasurementPostDoesNotAcceptClientDuration(t *testing.T) {
	fake := &fakeGenerationService{measurement: generation.AudioMeasurement{ID: 1, BatchProjectID: 3, BookID: 11, DurationMS: 28000}}
	h := NewHandler(Dependencies{Generation: fake})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/audio-measurement", bytes.NewBufferString(`{"audioAsset":"/audio/11.mp3","audioDurationSec":999}`))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if fake.measureReq.BatchProjectID != 3 || fake.measureReq.BookID != 11 || fake.measureReq.AudioAsset != "/audio/11.mp3" {
		t.Fatalf("measure req=%#v", fake.measureReq)
	}
}

func TestAudioProbeUnavailableReturnsServiceUnavailableStableCode(t *testing.T) {
	fake := &fakeGenerationService{measurementErr: generation.ErrAudioProbeUnavailable}
	h := NewHandler(Dependencies{Generation: fake})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/audio-measurement", bytes.NewBufferString(`{"audioAsset":"/audio/11.mp3"}`))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if !bytes.Contains(res.Body.Bytes(), []byte("audio_probe_unavailable")) {
		t.Fatalf("body=%s", res.Body.String())
	}
}

func TestMissingMeasurementReturnsNotFound(t *testing.T) {
	fake := &fakeGenerationService{measurementErr: errors.New("wrapped: " + generation.ErrNotFound.Error())}
	_ = fake
}

func generationMutationHarness(runtime *fakeGenerationRuntime) http.Handler {
	return NewHandler(Dependencies{
		Auth:                  &fakeAuthService{user: authn.User{ID: 5, TeamID: 2, Capabilities: []string{CapabilityBatchExecute, CapabilityBatchView}}},
		BatchProjectAccess:    &fakeRuntimeAccess{allowed: true},
		BatchProjectLifecycle: &fakeBatchProjectLifecycle{},
		GenerationRuntime:     runtime,
	})
}

func TestGenerationPOSTUsesHeaderIdempotencyAndReturnsAcceptedRuntimeContract(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: true, result: task9runtime.AdmissionResult{
		Run:     task9runtime.RunRecord{ID: 44, BatchProjectID: 3, Status: task9runtime.RunRunning},
		TaskIDs: []int64{91, 92}, Created: true, Dispatch: "queued",
	}}
	h := generationMutationHarness(runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/generation", bytes.NewBufferString(`{"requestId":"body-key","bookIds":[11,12]}`))
	req.Header.Set("Idempotency-Key", "header-key")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Location") != "/api/v1/batch-projects/3/generation/runs/44" || rec.Header().Get("Retry-After") != "1" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("headers=%v", rec.Header())
	}
	if runtime.req.RequestID != "header-key" || runtime.req.RequestedByUserID != 5 {
		t.Fatalf("request=%+v", runtime.req)
	}
	for _, want := range []string{`"runId":44`, `"taskIds":[91,92]`, `"pollUrl":"/api/v1/batch-projects/3/generation/runs/44"`, `"dispatch":"queued"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("missing %s in %s", want, rec.Body.String())
		}
	}
}

func TestGenerationPOSTNeverFallsBackToLegacySynchronousService(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: true, result: task9runtime.AdmissionResult{
		Run: task9runtime.RunRecord{ID: 44, BatchProjectID: 3}, TaskIDs: []int64{91}, Created: true, Dispatch: "queued",
	}}
	legacy := &fakeGenerationService{}
	h := NewHandler(Dependencies{
		Auth:                  &fakeAuthService{user: authn.User{ID: 5, TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}},
		BatchProjectAccess:    &fakeRuntimeAccess{allowed: true},
		BatchProjectLifecycle: &fakeBatchProjectLifecycle{},
		Generation:            legacy,
		GenerationRuntime:     runtime,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/generation", bytes.NewBufferString(`{"requestId":"request-key","bookIds":[11]}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted || runtime.admitCalls != 1 {
		t.Fatalf("status=%d runtime calls=%d body=%s", rec.Code, runtime.admitCalls, rec.Body.String())
	}
	if legacy.runBatchReq.BatchProjectID != 0 || legacy.runBookReq.BatchProjectID != 0 || legacy.retryReq.BatchProjectID != 0 {
		t.Fatalf("legacy synchronous generation was called: batch=%+v book=%+v retry=%+v", legacy.runBatchReq, legacy.runBookReq, legacy.retryReq)
	}
}

func TestGenerationPOSTFailsClosedWhenRuntimeNotReady(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: false}
	h := generationMutationHarness(runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation", bytes.NewBufferString(`{"requestId":"key"}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable || runtime.admitCalls != 0 || !strings.Contains(rec.Body.String(), "GENERATION_RUNTIME_UNAVAILABLE") {
		t.Fatalf("status=%d calls=%d body=%s", rec.Code, runtime.admitCalls, rec.Body.String())
	}
}

func TestGenerationRetryRequiresFrozenSourceBookRun(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: true, result: task9runtime.AdmissionResult{Run: task9runtime.RunRecord{ID: 45, BatchProjectID: 3}, TaskIDs: []int64{93}, Created: true, Dispatch: "queued"}}
	h := generationMutationHarness(runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation/stages/DIRECTOR/retry", bytes.NewBufferString(`{"requestId":"retry-key","sourceBookRunId":81}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted || runtime.req.Action != task9runtime.GenerationActionStageRetry || runtime.req.SourceBookRunID != 81 || runtime.req.RetryStage != "DIRECTOR" {
		t.Fatalf("status=%d request=%+v body=%s", rec.Code, runtime.req, rec.Body.String())
	}
}

func TestGenerationRunGETReturnsOnlyLifecycleProjection(t *testing.T) {
	runtime := &fakeGenerationRuntime{ready: false, run: task9runtime.GenerationRunStatus{RunID: 44, BatchProjectID: 3, Status: "completed", Terminal: true, Counts: task9runtime.GenerationRunCounts{Total: 1, Completed: 1}, Tasks: []task9runtime.GenerationTaskStatus{{TaskID: 91, BookID: 11, Attempt: 1, Status: "completed"}}}}
	h := generationMutationHarness(runtime)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/3/generation/runs/44", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"terminal":true`) || !strings.Contains(rec.Body.String(), `"counts":{"total":1`) || strings.Contains(rec.Body.String(), "request_snapshot") || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGenerationAdmissionErrorsUseStableCodesAndNeverLeakCause(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
		code string
	}{
		{task9runtime.ErrInvalidGenerationRequest, http.StatusBadRequest, "GENERATION_INVALID"},
		{task9runtime.ErrIdempotencyConflict, http.StatusConflict, "GENERATION_IDEMPOTENCY_CONFLICT"},
		{task9runtime.ErrProjectArchived, http.StatusConflict, "BATCH_PROJECT_ARCHIVED"},
		{task9runtime.ErrExecutorUnavailable, http.StatusServiceUnavailable, "GENERATION_RUNTIME_UNAVAILABLE"},
		{errors.New(generationDiagnostic), http.StatusInternalServerError, "GENERATION_ADMISSION_FAILED"},
	} {
		runtime := &fakeGenerationRuntime{ready: true, err: tc.err}
		h := generationMutationHarness(runtime)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation", bytes.NewBufferString(`{"requestId":"key"}`))
		req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
		sameOrigin(req)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.want || !strings.Contains(rec.Body.String(), tc.code) {
			t.Fatalf("err=%v status=%d body=%s", tc.err, rec.Code, rec.Body.String())
		}
		assertSafeOutcomeBody(t, rec.Body.String())
	}
}
