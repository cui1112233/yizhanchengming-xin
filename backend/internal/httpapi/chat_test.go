package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type chatTestProvider struct{}

func (chatTestProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	return "生成结果", nil
}

func TestChatRouteUsesRequestedChannelModel(t *testing.T) {
	mux := http.NewServeMux()
	resolver := func(_ context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || modelID != "channel:deepseek" {
			t.Errorf("owner=%q modelID=%q", owner, modelID)
		}
		return nil, errors.New("渠道未启用")
	}
	registerChatRoute(mux, BridgeAuth{Secret: "secret"}, chatTestProvider{}, resolver, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"promptType":"script","textModelId":"channel:deepseek"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "所选文本模型不可用") {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestChatRouteRejectsMissingNovelText(t *testing.T) {
	mux := http.NewServeMux()
	registerChatRoute(mux, BridgeAuth{Secret: "secret"}, chatTestProvider{}, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"promptType":"script"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "需要完整原文") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
