package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type localExecutorHTTPStub struct {
	registered video.LocalExecutorRegistrationResult
	identity   video.LocalExecutorIdentity
	completed  string
	failed     string
}

func (s *localExecutorHTTPStub) Register(context.Context, video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error) {
	return s.registered, nil
}
func (s *localExecutorHTTPStub) Identity(context.Context, string) (video.LocalExecutorIdentity, error) { return s.identity, nil }
func (s *localExecutorHTTPStub) Heartbeat(context.Context, string, video.LocalExecutorHeartbeatInput) error { return nil }
func (s *localExecutorHTTPStub) List(context.Context) ([]video.LocalExecutorIdentity, error) { return []video.LocalExecutorIdentity{s.identity}, nil }
func (s *localExecutorHTTPStub) CompleteTask(_ context.Context, _ string, taskID string, _ video.LocalExecutorCompleteInput) error {
	s.completed = taskID
	return nil
}
func (s *localExecutorHTTPStub) FailTask(_ context.Context, _ string, taskID string, _ video.LocalExecutorFailInput) error {
	s.failed = taskID
	return nil
}

func TestLocalExecutorHTTPRegistrationDoesNotExposeTokenHash(t *testing.T) {
	stub := &localExecutorHTTPStub{registered: video.LocalExecutorRegistrationResult{
		Token: "one-time-secret",
		Executor: video.LocalExecutorIdentity{ID: "lex_1", Name: "mac", ProviderKey: video.ProviderDoubaoLocalExecutor, Model: video.ModelDoubaoSeedance, Online: true, LastSeenAt: time.Now(), TokenConfigured: true},
	}}
	h := NewHandler(Dependencies{VideoLocalExecutor: stub})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/video/local-executors/register", strings.NewReader(`{"name":"mac","providerKey":"doubao_local_executor","model":"doubao-seedance"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "one-time-secret") || strings.Contains(rec.Body.String(), "tokenHash") {
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
