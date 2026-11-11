package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestSafeUsernameRejectsPathTraversal(t *testing.T) {
	for _, raw := range []string{"", " ", ".", "..", "../alice", `..\alice`, "a/b", `a\b`, "C:evil", "alice..bob", "bad\x00name"} {
		if ValidUsername(raw) {
			t.Fatalf("ValidUsername(%q) = true, want false", raw)
		}
	}
	for _, raw := range []string{"alice", "alice.bob", "user-1", "测试用户"} {
		if username, err := SafeUsername(raw); err != nil || username != strings.TrimSpace(raw) {
			t.Fatalf("SafeUsername(%q) = %q, %v", raw, username, err)
		}
	}
}

func TestCanonicalWebSubmitAliasRejectsAnonymousIdentityHeaders(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "owner", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true}}}); err != nil {
		t.Fatal(err)
	}
	accounts := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", Now: func() time.Time { return now }, NovelFetchStore: store, AuthStore: accounts})

	anonymous := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/web-submit/config", nil)
	anonymous.Header.Set(HeaderUsername, "owner")
	anonymous.Header.Set(HeaderIsOwner, "true")
	anonymous.Header.Set(HeaderIssuedAt, strconv.FormatInt(now.Unix(), 10))
	anonymousResponse := httptest.NewRecorder()
	handler.ServeHTTP(anonymousResponse, anonymous)
	if anonymousResponse.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous alias status=%d body=%s", anonymousResponse.Code, anonymousResponse.Body.String())
	}

	// A valid HMAC signature must not help either: this alias only trusts a real
	// login session and never header identity.
	signed := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/web-submit/config", nil)
	SignBridgeRequest(signed, "owner", true, now, "test-secret")
	signedResponse := httptest.NewRecorder()
	handler.ServeHTTP(signedResponse, signed)
	if signedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("signed alias status=%d body=%s", signedResponse.Code, signedResponse.Body.String())
	}
}

func TestAgentWorkspaceRootPathIsConfinedToUserDirectory(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	usersDir := t.TempDir()
	accounts := &authTestStore{
		account: auth.Account{Username: "tester", Active: true},
		session: auth.Session{Username: "tester", Token: "agent-session", ExpiresAt: now.Add(time.Hour)},
	}
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "test-secret",
		Now:             func() time.Time { return now },
		UserConfigStore: NewUserConfigStore(usersDir),
		AuthStore:       accounts,
	})
	create := func(name, rootPath string) *httptest.ResponseRecorder {
		payload, _ := json.Marshal(map[string]any{"name": name, "rootPath": rootPath})
		req := httptest.NewRequest(http.MethodPost, "/api/agent/workspaces", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer agent-session")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}

	if response := create("data-dir", usersDir); response.Code != http.StatusBadRequest {
		t.Fatalf("data dir root status=%d body=%s", response.Code, response.Body.String())
	}
	volumeRoot := filepath.VolumeName(usersDir) + string(filepath.Separator)
	if response := create("volume-root", volumeRoot); response.Code != http.StatusBadRequest {
		t.Fatalf("volume root status=%d body=%s", response.Code, response.Body.String())
	}

	allowedRoot := filepath.Join(usersDir, "tester", "agent", "custom-workspace")
	if err := os.MkdirAll(allowedRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if response := create("custom", allowedRoot); response.Code != http.StatusOK {
		t.Fatalf("allowed root status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAgentWorkspaceFileReadRejectsSymlinkEscape(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	usersDir := t.TempDir()
	accounts := &authTestStore{
		account: auth.Account{Username: "tester", Active: true},
		session: auth.Session{Username: "tester", Token: "agent-session", ExpiresAt: now.Add(time.Hour)},
	}
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "test-secret",
		Now:             func() time.Time { return now },
		UserConfigStore: NewUserConfigStore(usersDir),
		AuthStore:       accounts,
	})
	workspaceRoot := filepath.Join(usersDir, "tester", "agent", "linked")
	if err := os.MkdirAll(workspaceRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"name": "linked", "rootPath": workspaceRoot})
	create := httptest.NewRequest(http.MethodPost, "/api/agent/workspaces", bytes.NewReader(payload))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Authorization", "Bearer agent-session")
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, create)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", createResponse.Code, createResponse.Body.String())
	}
	var created struct {
		Workspace struct {
			ID string `json:"id"`
		} `json:"workspace"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil || created.Workspace.ID == "" {
		t.Fatalf("create payload=%s err=%v", createResponse.Body.String(), err)
	}

	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(workspaceRoot, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	read := httptest.NewRequest(http.MethodGet, "/api/agent/workspaces/"+created.Workspace.ID+"/file?path=escape/secret.txt", nil)
	read.Header.Set("Authorization", "Bearer agent-session")
	readResponse := httptest.NewRecorder()
	handler.ServeHTTP(readResponse, read)
	if readResponse.Code != http.StatusBadRequest {
		t.Fatalf("symlink escape status=%d body=%s", readResponse.Code, readResponse.Body.String())
	}
}

func TestUserScopedPathConfinesUsernameToBaseDirectory(t *testing.T) {
	base := t.TempDir()
	for _, raw := range []string{"../escape", `..\escape`, "alice/bob", `alice\bob`, "..", ""} {
		if _, err := UserScopedPath(base, raw, "api-config.json"); err == nil {
			t.Fatalf("UserScopedPath(%q) succeeded, want error", raw)
		}
	}
	path, err := UserScopedPath(base, "测试 用户", "api-config.json")
	if err != nil {
		t.Fatalf("valid username rejected: %v", err)
	}
	want := filepath.Join(base, "测试 用户", "api-config.json")
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}

	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(base, "linked")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := UserScopedPath(base, "linked", "api-config.json"); err == nil {
		t.Fatal("symlinked username escaped the base directory")
	}
}

func TestLegacyWebSubmitAliasRejectsSignedIdentityHeaders(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := novelfetchworkshop.NewMemoryStore()
	accounts := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "test-secret",
		Now:             func() time.Time { return now },
		NovelFetchStore: store,
		UserDataDir:     t.TempDir(),
		AuthStore:       accounts,
	})
	for _, path := range []string{"/api/web-submit/config", "/api/web-submit/operations", "/api/web-submit/submit", "/api/batch-rewrite/web-submit/preview", "/api/batch-rewrite/web-submit/config", "/api/batch-rewrite/web-submit/environment"} {
		method := http.MethodGet
		if path == "/api/web-submit/operations" || path == "/api/web-submit/submit" || path == "/api/batch-rewrite/web-submit/preview" {
			method = http.MethodPost
		}
		req := httptest.NewRequest(method, path, strings.NewReader(`{}`))
		SignBridgeRequest(req, "owner", true, now, "test-secret")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s signed-only status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
}

func TestLoginRateLimitLocksIPAndUsername(t *testing.T) {
	store := &authTestStore{account: auth.Account{Username: "alice", Active: true, PasswordHash: "hash", PasswordAlgorithm: "sha256"}}
	now := time.Unix(1_800_000_000, 0)
	limiter := newAuthRateLimiter(AuthRateLimitConfig{LoginWindow: time.Minute, LoginMaxFailures: 3, LoginLockout: 5 * time.Minute}, func() time.Time { return now })
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, AuthOptions{Store: store, Now: func() time.Time { return now }, RateLimiter: limiter})
	attempt := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"alice","password":"wrong"}`))
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		return response
	}
	for i := 0; i < 3; i++ {
		if response := attempt(); response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status=%d body=%s", i+1, response.Code, response.Body.String())
		}
	}
	locked := attempt()
	if locked.Code != http.StatusTooManyRequests {
		t.Fatalf("locked status=%d body=%s", locked.Code, locked.Body.String())
	}
	if !strings.Contains(locked.Body.String(), "操作过于频繁") || locked.Header().Get("Retry-After") == "" {
		t.Fatalf("locked response missing Chinese message or Retry-After: %s", locked.Body.String())
	}
}

func TestPasswordResetRateLimitLocksAfterConfiguredAttempts(t *testing.T) {
	t.Setenv("QIANTIE_MAIL_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("QIANTIE_MAIL_FROM", "noreply@example.com")
	now := time.Unix(1_800_000_000, 0)
	limiter := newAuthRateLimiter(AuthRateLimitConfig{PasswordResetWindow: time.Minute, PasswordResetMax: 2, PasswordResetLockout: 5 * time.Minute}, func() time.Time { return now })
	mux := http.NewServeMux()
	RegisterAccountRecoveryStatusCompat(mux, BridgeAuth{Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "alice", Active: true}}}, t.TempDir(), limiter)
	attempt := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/account-recovery/password/request", strings.NewReader(`{"email":"nobody@example.com"}`))
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		return response
	}
	for i := 0; i < 2; i++ {
		if response := attempt(); response.Code != http.StatusAccepted {
			t.Fatalf("attempt %d status=%d body=%s", i+1, response.Code, response.Body.String())
		}
	}
	locked := attempt()
	if locked.Code != http.StatusTooManyRequests {
		t.Fatalf("locked status=%d body=%s", locked.Code, locked.Body.String())
	}
	if !strings.Contains(locked.Body.String(), "操作过于频繁") {
		t.Fatalf("locked response missing Chinese message: %s", locked.Body.String())
	}
}

func TestRecoveryResetLinkUsesConfiguredSiteNotRequestHost(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("QIANTIE_PUBLIC_BASE_URL", "https://configured.example")
	var mailBody string
	mailServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		raw, _ := io.ReadAll(req.Body)
		mailBody = string(raw)
		w.WriteHeader(http.StatusOK)
	}))
	defer mailServer.Close()
	t.Setenv("QIANTIE_MAIL_ENDPOINT", mailServer.URL)
	t.Setenv("QIANTIE_MAIL_FROM", "noreply@example.com")

	state, _ := json.Marshal(map[string]any{
		"verifiedEmails": []any{map[string]any{"username": "alice", "email": "alice@example.com", "verifiedAt": time.Now().UTC().Format(time.RFC3339)}},
	})
	if err := os.WriteFile(filepath.Join(dir, "account-recovery.json"), state, 0o600); err != nil {
		t.Fatal(err)
	}
	profiles, _ := json.Marshal([]map[string]any{{"username": "alice", "email": "alice@example.com"}})
	if err := os.WriteFile(filepath.Join(dir, "member-profile-details.json"), profiles, 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterAccountRecoveryStatusCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{account: auth.Account{Username: "alice", Active: true}}}, dir)
	req := httptest.NewRequest(http.MethodPost, "/api/account-recovery/password/request", strings.NewReader(`{"email":"alice@example.com"}`))
	req.Host = "evil.example"
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusAccepted {
		t.Fatalf("recovery request status=%d body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(mailBody, "https://configured.example/recover?reset=") {
		t.Fatalf("recovery mail body missing configured origin: %s", mailBody)
	}
	if strings.Contains(mailBody, "evil.example") {
		t.Fatalf("recovery mail body trusted Host header: %s", mailBody)
	}
}
