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

type chatTestProvider struct{}

func (chatTestProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	return "生成结果", nil
}

func TestChatRouteRejectsMissingNovelText(t *testing.T) {
	mux := http.NewServeMux()
	registerChatRoute(mux, BridgeAuth{Secret: "secret"}, chatTestProvider{}, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"promptType":"script"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "需要完整原文") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
