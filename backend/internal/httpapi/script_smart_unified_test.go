package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type smartUnifiedTestProvider struct{}

func (smartUnifiedTestProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	return `{"final_genre":"都市情感短剧","trailer_style":"数字电影摄影，柔和高光滚降，冷暖对照","story_era":"现代都市"}`, nil
}

type directorPipelineTestProvider struct {
	request batchfactoryv11.TextCompletionRequest
}

func (p *directorPipelineTestProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.request = request
	return "Go 导演结果", nil
}

func TestNormalizeSmartUnifiedResultBuildsCanonicalPrompt(t *testing.T) {
	result, err := normalizeSmartUnifiedResult("```json\n{\"final_genre\":\"悬疑短剧\",\"trailer_style\":\"电影摄影\",\"story_era\":\"当代\"}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if result["prompt"] != "悬疑短剧；电影摄影；当代。" {
		t.Fatalf("prompt = %v", result["prompt"])
	}
}

func TestScriptSmartUnifiedRouteRejectsEmptyNovelText(t *testing.T) {
	mux := http.NewServeMux()
	registerScriptSmartUnifiedRoute(mux, BridgeAuth{Secret: "secret"}, smartUnifiedTestProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/script/smart-unified-style", strings.NewReader(`{"novelText":""}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "需要完整原文") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestScriptDirectorPipelineCompatRunsOnGoProvider(t *testing.T) {
	provider := &directorPipelineTestProvider{}
	mux := http.NewServeMux()
	registerScriptDirectorPipelineCompat(mux, BridgeAuth{Secret: "secret"}, provider, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/script/director-pipeline", strings.NewReader(`{"novelText":"林夏推开病房门。","duration":"10s"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"output_text":"Go 导演结果"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(provider.request.UserPrompt, "林夏推开病房门") {
		t.Fatalf("Go provider did not receive the novel text: %q", provider.request.UserPrompt)
	}
}
