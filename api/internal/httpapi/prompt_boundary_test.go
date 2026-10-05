package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSystemPresetPromptEndpointIsPublicReadOnlyNotBroadWhitelist(t *testing.T) {
	handler := NewHandler(Dependencies{Auth: &fakeAuthService{}})

	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/v1/generation/prompts", nil))
	if get.Code == http.StatusUnauthorized || get.Code == http.StatusForbidden {
		t.Fatalf("public system preset read unexpectedly requires user auth: %d body=%s", get.Code, get.Body.String())
	}

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/generation/prompts", nil)
	postReq.Host = "app.example"
	postReq.Header.Set("Origin", "http://app.example")
	post := httptest.NewRecorder()
	handler.ServeHTTP(post, postReq)
	if post.Code == http.StatusOK || post.Code == http.StatusCreated || post.Code == http.StatusNoContent {
		t.Fatalf("system preset public route must be GET-only, got %d", post.Code)
	}
}

func TestHistoricalScriptConstraintPromptPathIsNotWhitelisted(t *testing.T) {
	handler := NewHandler(Dependencies{Auth: &fakeAuthService{}})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/script-constraint-prompts?category=prefix", nil))
	if rec.Code >= 200 && rec.Code < 300 {
		t.Fatalf("historical prompt path must not become an unauthenticated whitelist shortcut: %d body=%s", rec.Code, rec.Body.String())
	}
}
