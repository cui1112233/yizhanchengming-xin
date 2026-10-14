package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestNovelFetchUploadSessionDoesNotFallBackToDuplicateDirectHTTP(t *testing.T) {
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "alice"), 0o700); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"cookie": "sid=fixture", "targetUsername": "operator"})
	if err := os.WriteFile(filepath.Join(usersDir, "alice", "upload-target.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterNovelFetchUploadSessionCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{}}, usersDir)
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-upload/upload-session", nil)
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "Go 连接器未配置") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
