package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
)

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
	fake := &fakeGenerationService{book: generation.BookGenerationResult{Run: generation.BookRun{ID: 9, Status: generation.StatusCompleted}}}
	h := NewHandler(Dependencies{Generation: fake})
	body := bytes.NewBufferString(`{"hookEnabled":true,"directorMode":"normal","requestId":"r1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation", body)
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if fake.runBookReq.BatchProjectID != 3 || fake.runBookReq.BookID != 11 {
		t.Fatalf("request=%#v", fake.runBookReq)
	}
}

func TestGenerationRetryBindsTargetStage(t *testing.T) {
	fake := &fakeGenerationService{}
	h := NewHandler(Dependencies{Generation: fake})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/generation/stages/DIRECTOR/retry", bytes.NewBufferString(`{"requestId":"retry-1"}`))
	res := httptest.NewRecorder()
	h.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if fake.retryReq.Stage != generation.StageDirector || fake.retryReq.BookID != 11 {
		t.Fatalf("retry=%#v", fake.retryReq)
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
