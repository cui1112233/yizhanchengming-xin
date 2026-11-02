package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/videomanagementconnector"
)

type failingVideoManagementLogin struct{ err error }

func (f failingVideoManagementLogin) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, f.err
}
func (f failingVideoManagementLogin) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}
func (f failingVideoManagementLogin) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}
func (f failingVideoManagementLogin) Submit(context.Context, videomanagementconnector.Session, videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{}, nil
}
func (f failingVideoManagementLogin) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{}, nil
}

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

func TestNovelFetchUploadCanonicalSessionAliasUsesGoHandler(t *testing.T) {
	mux := http.NewServeMux()
	RegisterNovelFetchUploadSessionCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{}}, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/upload/session", nil)
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"loggedIn":false`) {
		t.Fatalf("canonical alias response=%d %s", res.Code, res.Body.String())
	}
}

func TestNovelFetchUploadSessionReturnsTargetUsername(t *testing.T) {
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "alice"), 0o700); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"cookie": "sid=fixture", "targetUsername": " operator "})
	if err := os.WriteFile(filepath.Join(usersDir, "alice", "upload-target.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterNovelFetchUploadSessionCompat(
		mux,
		BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{}},
		usersDir,
		failingVideoManagementLogin{},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/upload/session", nil)
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Username != "operator" {
		t.Fatalf("username=%q, want operator", response.Username)
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

func TestNovelFetchUploadLoginExplainsUnavailableTargetWithoutLegacyName(t *testing.T) {
	mux := http.NewServeMux()
	RegisterNovelFetchUploadSessionCompat(
		mux,
		BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{}},
		t.TempDir(),
		failingVideoManagementLogin{err: errors.New("target site login page returned HTTP 404")},
	)
	req := httptest.NewRequest(http.MethodPost, "/api/novel-fetch-workshop/upload/login", strings.NewReader(`{"username":"operator","password":"secret"}`))
	SignBridgeRequest(req, "alice", false, time.Now(), "secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "121") || !strings.Contains(rec.Body.String(), "视频管理系统登录入口不可用") {
		t.Fatalf("body=%s, want actionable video-management error without legacy name", rec.Body.String())
	}
}
