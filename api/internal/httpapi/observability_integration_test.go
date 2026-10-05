package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	_ "github.com/go-sql-driver/mysql"
)

func TestReadyzFailsWhenDatabasePingFails(t *testing.T) {
	db, err := sql.Open("mysql", "user:password@tcp(127.0.0.1:1)/missing?timeout=10ms")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Database: db, AppInitialized: true}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["code"] != "DATABASE_UNAVAILABLE" {
		t.Fatalf("code = %#v, want DATABASE_UNAVAILABLE", body["code"])
	}
	if body["request_id"] == "" || rec.Header().Get(observability.RequestIDHeader) == "" {
		t.Fatal("readiness failure is missing request correlation")
	}
}

func TestHealthzIgnoresOptionalSubsystemAvailability(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{
		AppInitialized:     false,
		Database:           nil,
		VideoConfig:        nil,
		VideoLocalExecutor: nil,
		VideoMerge:         nil,
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200 despite DB/provider/executor/ffmpeg/runtime state", rec.Code)
	}
}

func TestDiagnosticsRequireAdminNotBatchView(t *testing.T) {
	auth := opsAuthService{user: authn.User{ID: 42, Role: "user", Capabilities: []string{CapabilityBatchView}}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Auth: auth}).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("batch.view diagnostics status = %d, want 403", rec.Code)
	}
}

func TestDiagnosticsAdminDTOHasNoSecretLikeFields(t *testing.T) {
	auth := opsAuthService{user: authn.User{ID: 1, Role: "admin"}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Auth: auth}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("admin diagnostics status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	lower := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"authorization", "set-cookie", "ciphertext", "nonce", "client_secret", "access_token", "refresh_token", "token_hash", "dsn"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("diagnostics DTO contains forbidden field %q: %s", forbidden, rec.Body.String())
		}
	}
}

func TestAccessLogDoesNotRecordQueryHeadersOrBodies(t *testing.T) {
	var logs bytes.Buffer
	logger := observability.NewJSONLogger(&logs)
	req := httptest.NewRequest(http.MethodPost, "/missing?token=query-secret", strings.NewReader(`{"password":"body-secret"}`))
	req.Header.Set("Authorization", "Bearer header-secret")
	req.Header.Set("Cookie", "session=cookie-secret")
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Logger: logger}).ServeHTTP(rec, req)

	output := logs.String()
	for _, secret := range []string{"query-secret", "body-secret", "header-secret", "cookie-secret"} {
		if strings.Contains(output, secret) {
			t.Fatalf("access log leaked %q: %s", secret, output)
		}
	}
	if !strings.Contains(output, `"path":"/missing"`) {
		t.Fatalf("access log missing safe path: %s", output)
	}
}

func TestPanicRecoveryReturnsRequestIDWithoutStackOrCause(t *testing.T) {
	var logs bytes.Buffer
	h := handler{deps: Dependencies{Logger: observability.NewJSONLogger(&logs)}}
	wrapped := h.withObservability(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(errors.New("password=panic-secret\nSTACK-LINE"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic status = %d, want 500", rec.Code)
	}
	if rec.Header().Get(observability.RequestIDHeader) == "" {
		t.Fatal("panic response missing X-Request-ID")
	}
	if strings.Contains(rec.Body.String(), "panic-secret") || strings.Contains(rec.Body.String(), "STACK-LINE") {
		t.Fatalf("panic response leaked internal cause: %s", rec.Body.String())
	}
	if strings.Contains(logs.String(), "panic-secret") || strings.Contains(logs.String(), "\nSTACK-LINE") {
		t.Fatalf("panic log leaked secret/control sequence: %s", logs.String())
	}
}

type opsAuthService struct {
	user authn.User
}

func (s opsAuthService) Login(context.Context, string, string) (authn.Credentials, authn.User, error) {
	return authn.Credentials{}, authn.User{}, authn.ErrUnauthenticated
}
func (s opsAuthService) AuthenticateAccess(context.Context, string) (authn.User, error) {
	if s.user.ID <= 0 {
		return authn.User{}, authn.ErrUnauthenticated
	}
	return s.user, nil
}
func (s opsAuthService) Refresh(context.Context, string) (authn.Credentials, authn.User, error) {
	return authn.Credentials{}, authn.User{}, authn.ErrUnauthenticated
}
func (s opsAuthService) Logout(context.Context, string, string) error { return nil }

func TestProcessDiagnosticsRemainAggregateOnly(t *testing.T) {
	stats := observability.ReadProcessStats(time.Now().Add(-time.Minute), time.Now())
	if stats.Goroutines <= 0 || stats.Sys == 0 || stats.UptimeSec < 50 {
		t.Fatalf("unexpected process stats: %#v", stats)
	}
}
