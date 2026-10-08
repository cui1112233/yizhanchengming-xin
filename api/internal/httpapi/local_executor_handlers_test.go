package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type localExecutorHTTPStub struct {
	registered video.LocalExecutorRegistrationResult
	identity   video.LocalExecutorIdentity
	completed  string
	failed     string
	ownerID    int64
}

func (s *localExecutorHTTPStub) Register(context.Context, video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error) {
	return s.registered, nil
}
func (s *localExecutorHTTPStub) CreatePairingIntent(_ context.Context, ownerID int64) (video.LocalExecutorPairingResult, error) {
	s.ownerID = ownerID
	return video.LocalExecutorPairingResult{Payload: "pair_payload", ExpiresAt: time.Now().Add(time.Minute)}, nil
}
func (s *localExecutorHTTPStub) RedeemPairingIntent(context.Context, string, video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error) {
	return s.registered, nil
}
func (s *localExecutorHTTPStub) ListForOwner(_ context.Context, ownerID int64) ([]video.LocalExecutorIdentity, error) {
	s.ownerID = ownerID
	return []video.LocalExecutorIdentity{s.identity}, nil
}
func (s *localExecutorHTTPStub) UnbindForOwner(_ context.Context, ownerID int64, _ string) error {
	s.ownerID = ownerID
	return nil
}
func (s *localExecutorHTTPStub) Identity(context.Context, string) (video.LocalExecutorIdentity, error) {
	return s.identity, nil
}
func (s *localExecutorHTTPStub) Heartbeat(context.Context, string, video.LocalExecutorHeartbeatInput) error {
	return nil
}
func (s *localExecutorHTTPStub) List(context.Context) ([]video.LocalExecutorIdentity, error) {
	return []video.LocalExecutorIdentity{s.identity}, nil
}
func (s *localExecutorHTTPStub) CompleteTask(_ context.Context, _ string, taskID string, _ video.LocalExecutorCompleteInput) error {
	s.completed = taskID
	return nil
}
func (s *localExecutorHTTPStub) FailTask(_ context.Context, _ string, taskID string, _ video.LocalExecutorFailInput) error {
	s.failed = taskID
	return nil
}

func TestLocalExecutorHTTPBootstrapRegistrationRequiresSecurePairing(t *testing.T) {
	stub := &localExecutorHTTPStub{registered: video.LocalExecutorRegistrationResult{
		Token:    "one-time-secret",
		Executor: video.LocalExecutorIdentity{ID: "lex_1", Name: "mac", ProviderKey: video.ProviderDoubaoLocalExecutor, Model: video.ModelDoubaoSeedance, Online: true, LastSeenAt: time.Now(), TokenConfigured: true},
	}}
	h := NewHandler(Dependencies{VideoLocalExecutor: stub, VideoExecutorBootstrapToken: "bootstrap-secret"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/video/local-executors/register", strings.NewReader(`{"name":"mac","providerKey":"doubao_local_executor","model":"doubao-seedance"}`))
	req.Header.Set("Authorization", "Bearer bootstrap-secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusGone {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "one-time-secret") || strings.Contains(rec.Body.String(), "tokenHash") || !strings.Contains(rec.Body.String(), "local_executor_pairing_required") {
		t.Fatalf("registration response = %s", rec.Body.String())
	}
}

func TestLocalExecutorHTTPUsesBearerTokenForIdentityAndTaskCompletion(t *testing.T) {
	stub := &localExecutorHTTPStub{identity: video.LocalExecutorIdentity{ID: "lex_1", ProviderKey: video.ProviderDoubaoLocalExecutor, Model: video.ModelDoubaoSeedance, Online: true, TokenConfigured: true}}
	h := NewHandler(Dependencies{VideoLocalExecutor: stub})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/video/local-executors/me", nil)
	req.Header.Set("Authorization", "Bearer executor-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("identity status=%d body=%s", rec.Code, rec.Body.String())
	}
	var identity video.LocalExecutorIdentity
	if err := json.Unmarshal(rec.Body.Bytes(), &identity); err != nil || identity.ID != "lex_1" {
		t.Fatalf("identity response=%s err=%v", rec.Body.String(), err)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/video/local-executor-tasks/let_1/complete", strings.NewReader(`{"artifactUrl":"https://cdn.example/video.mp4"}`))
	req.Header.Set("Authorization", "Bearer executor-token")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || stub.completed != "let_1" {
		t.Fatalf("complete status=%d task=%q body=%s", rec.Code, stub.completed, rec.Body.String())
	}
}

func TestLocalExecutorPairingBrowserRoutesRequireSessionCSRFAndBindCurrentUser(t *testing.T) {
	stub := &localExecutorHTTPStub{registered: video.LocalExecutorRegistrationResult{Token: "executor-only-token"}}
	h := NewHandler(Dependencies{Auth: task14AuthStub{user: authn.User{ID: 17, Role: "member", Capabilities: []string{CapabilityBatchConfigure}}}, VideoLocalExecutor: stub})

	unauthenticated := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/video/local-executors/pairings", nil)
	unauthenticated.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, unauthenticated)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d body=%s", rec.Code, rec.Body.String())
	}

	crossOrigin := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/video/local-executors/pairings", nil)
	crossOrigin.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	crossOrigin.Header.Set("Origin", "https://attacker.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, crossOrigin)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin status=%d body=%s", rec.Code, rec.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/video/local-executors/pairings", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	req.Header.Set("Origin", "http://example.com")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated || stub.ownerID != 17 {
		t.Fatalf("pairing status=%d owner=%d body=%s", rec.Code, stub.ownerID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "executor-only-token") || !strings.Contains(rec.Body.String(), "ycm-executor://pair?intent=") {
		t.Fatalf("pairing response=%s", rec.Body.String())
	}

	list := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/video/local-executors", nil)
	list.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, list)
	if rec.Code != http.StatusOK || stub.ownerID != 17 {
		t.Fatalf("list status=%d owner=%d body=%s", rec.Code, stub.ownerID, rec.Body.String())
	}
}
