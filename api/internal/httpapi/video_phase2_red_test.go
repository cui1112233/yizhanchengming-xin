package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type phase2HTTPVideoService struct {
	cancelCalls int
	retryCalls  int
}
func (s *phase2HTTPVideoService) Start(context.Context, video.StartRequest) (video.StartResult, error) { return video.StartResult{}, nil }
func (s *phase2HTTPVideoService) PollTask(context.Context, int64) (video.ProductionTask, error) { return video.ProductionTask{}, nil }
func (s *phase2HTTPVideoService) CancelTask(context.Context, int64) (video.ProductionTask, error) {
	s.cancelCalls++
	return video.ProductionTask{ID: 7, Status: video.TaskCancelled}, nil
}
func (s *phase2HTTPVideoService) RetryTask(context.Context, int64, string) (video.StartResult, error) {
	s.retryCalls++
	return video.StartResult{Task: video.ProductionTask{ID: 8, Attempt: 2, Status: video.TaskQueued}}, nil
}

type phase2HTTPConfigService struct{}
func (phase2HTTPConfigService) Get(context.Context, string, string) (video.ProviderConfigView, error) {
	return video.ProviderConfigView{ProviderKey: video.ProviderPersonalAPI, Model: video.ModelYD20Mini, Configured: true, Enabled: true}, nil
}
func (phase2HTTPConfigService) Save(context.Context, video.ProviderConfigInput) (video.ProviderConfigView, error) {
	return video.ProviderConfigView{ProviderKey: video.ProviderPersonalAPI, Model: video.ModelYD20Mini, Configured: true, Enabled: true}, nil
}
func (phase2HTTPConfigService) Status(context.Context, string, string) (video.ProviderStatusView, error) {
	return video.ProviderStatusView{ProviderKey: video.ProviderPersonalAPI, Model: video.ModelYD20Mini, Configured: true, Status: video.ProviderStatusAvailable}, nil
}

func TestVideoProviderStatusRouteAndSecretSafety(t *testing.T) {
	h := NewHandler(Dependencies{VideoConfig: phase2HTTPConfigService{}})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/video-providers/personal_api/models/yd2.0-mini/status", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK { t.Fatalf("status = %d body=%s", w.Code, w.Body.String()) }
	if strings.Contains(strings.ToLower(w.Body.String()), "secret") || strings.Contains(w.Body.String(), "encryptedSecret") {
		t.Fatalf("status response leaked secret fields: %s", w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if body["status"] != string(video.ProviderStatusAvailable) { t.Fatalf("body = %+v", body) }
}

func TestVideoCancelAndRetryRoutes(t *testing.T) {
	service := &phase2HTTPVideoService{}
	h := NewHandler(Dependencies{Video: service})

	cancel := httptest.NewRequest(http.MethodPost, "/api/v1/video-tasks/7/cancel", nil)
	cancelW := httptest.NewRecorder()
	h.ServeHTTP(cancelW, cancel)
	if cancelW.Code != http.StatusOK || service.cancelCalls != 1 {
		t.Fatalf("cancel status=%d calls=%d body=%s", cancelW.Code, service.cancelCalls, cancelW.Body.String())
	}

	retry := httptest.NewRequest(http.MethodPost, "/api/v1/video-tasks/7/retry", strings.NewReader(`{"requestId":"retry-2"}`))
	retry.Header.Set("Content-Type", "application/json")
	retryW := httptest.NewRecorder()
	h.ServeHTTP(retryW, retry)
	if retryW.Code != http.StatusAccepted || service.retryCalls != 1 {
		t.Fatalf("retry status=%d calls=%d body=%s", retryW.Code, service.retryCalls, retryW.Body.String())
	}
}
