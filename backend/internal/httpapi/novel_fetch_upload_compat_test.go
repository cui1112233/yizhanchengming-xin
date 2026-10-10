package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNovelFetchUploadSessionMissingCookie(t *testing.T) {
	mux := http.NewServeMux()
	RegisterNovelFetchUploadSessionCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{}}, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-upload/upload-session", nil)
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || res.Body.String() == "" {
		t.Fatalf("missing session response=%d %s", res.Code, res.Body.String())
	}
}
