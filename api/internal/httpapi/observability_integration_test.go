package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novelpanel"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
	_ "github.com/go-sql-driver/mysql"
)

const errorCanaries = "password=pass-canary Cookie: session=cookie-canary Bearer bearer-canary provider_api_key=provider-canary user:dsn-canary@tcp(localhost:3306)/db"

func assertErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status=%d want=%d body=%s", rec.Code, status, rec.Body.String())
	}
	id, ok := observability.ValidRequestID(rec.Header().Values(observability.RequestIDHeader))
	if !ok {
		t.Fatalf("invalid request ID: %v", rec.Header())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %s", rec.Body.String())
	}
	if body["request_id"] != id || body["code"] != code || body["message"] == nil || body["message"] == "" {
		t.Fatalf("invalid envelope: %v; header=%q", body, id)
	}
}

func TestErrorRequestIDContract(t *testing.T) {
	api := NewHandler(Dependencies{Auth: opsAuthService{user: authn.User{ID: 7, Role: "member"}}})
	for _, tc := range []struct {
		name, path, code string
		status           int
		cookie           bool
	}{
		{"unknown", "/api/v1/not-found", "NOT_FOUND", 404, false},
		{"anonymous", "/api/auth/current-user", "AUTH_UNAUTHENTICATED", 401, false},
		{"forbidden", "/api/v1/issues", "AUTH_FORBIDDEN", 403, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.cookie {
				req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
			}
			rec := httptest.NewRecorder()
			api.ServeHTTP(rec, req)
			assertErrorEnvelope(t, rec, tc.status, tc.code)
		})
	}
	t.Run("runtime", func(t *testing.T) {
		rec := httptest.NewRecorder()
		NewHandlerWithRuntime(Dependencies{Auth: opsAuthService{}}, &fakeRuntimeService{}).ServeHTTP(rec, runtimeRetryRequest())
		assertErrorEnvelope(t, rec, 401, "AUTH_UNAUTHENTICATED")
	})
}

type failingSettings struct{ UnifiedSettingsService }

func (failingSettings) GetCurrent(context.Context, int64) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, errors.New(errorCanaries)
}

func TestSafeErrorServiceAndPanicCanaries(t *testing.T) {
	for _, mode := range []string{"service", "panic", "settings"} {
		t.Run(mode, func(t *testing.T) {
			var logs bytes.Buffer
			h := handler{deps: Dependencies{Logger: observability.NewJSONLogger(&logs)}}
			wrapped := h.withObservability(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if mode == "panic" {
					panic(errors.New(errorCanaries))
				}
				h.writeServiceError(w, r, 500, "INTERNAL_ERROR", "服务暂时不可用", "test", "read", errors.New(errorCanaries))
			}))
			path, code := "/api/test", "INTERNAL_ERROR"
			if mode == "settings" {
				wrapped = NewHandler(Dependencies{Logger: h.deps.Logger, UnifiedSettings: failingSettings{}})
				path, code = "/api/v1/batch-projects/9/settings", "SETTINGS_READ_FAILED"
			}
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
			for _, canary := range []string{"pass-canary", "cookie-canary", "bearer-canary", "provider-canary", "dsn-canary"} {
				if strings.Contains(rec.Body.String()+logs.String(), canary) {
					t.Errorf("leaked %s", canary)
				}
			}
			assertErrorEnvelope(t, rec, 500, code)
			if mode == "settings" {
				var body map[string]any
				_ = json.Unmarshal(rec.Body.Bytes(), &body)
				if body["error"] != body["message"] {
					t.Fatal("legacy error field must retain safe public message")
				}
			}
			for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
				var entry map[string]any
				if json.Unmarshal([]byte(line), &entry) != nil {
					t.Fatalf("invalid structured log: %s", line)
				}
				if entry["request_id"] != rec.Header().Get(observability.RequestIDHeader) {
					t.Fatalf("uncorrelated log: %s", line)
				}
			}
		})
	}
}

func TestSafeErrorQuotedJSONAndCookieCanaries(t *testing.T) {
	for _, input := range []string{
		`upstream failed {"provider_api_key":"provider-canary","password":"pass-canary","reason":"diagnostic-visible"}`,
		"upstream failed\nCookie: locale=zh; session=cookie-canary; refresh=refresh-canary\ndiagnostic-visible",
		"upstream failed\nCookie: [REDACTED]; session=cookie-canary\ndiagnostic-visible",
		"upstream failed\nSet-Cookie: locale=zh; session=cookie-canary; Expires=Wed, 21 Oct 2026 07:28:00 GMT; refresh=refresh-canary\ndiagnostic-visible",
		`upstream failed {"Cookie":"locale=zh; session=cookie-canary; refresh=refresh-canary","reason":"diagnostic-visible"}`,
	} {
		t.Run(input, func(t *testing.T) {
			var logs bytes.Buffer
			h := handler{deps: Dependencies{Logger: observability.NewJSONLogger(&logs)}}
			rec := httptest.NewRecorder()
			h.withObservability(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.writeServiceError(w, r, 500, "INTERNAL_ERROR", "服务暂时不可用", "provider", "request", errors.New(input))
			})).ServeHTTP(rec, httptest.NewRequest("GET", "/api/test", nil))
			assertErrorEnvelope(t, rec, 500, "INTERNAL_ERROR")
			for _, secret := range []string{"provider-canary", "pass-canary", "cookie-canary", "refresh-canary"} {
				if strings.Contains(logs.String()+rec.Body.String(), secret) {
					t.Errorf("leaked %s", secret)
				}
			}
			var entry map[string]any
			if err := json.Unmarshal(bytes.Split(logs.Bytes(), []byte("\n"))[0], &entry); err != nil {
				t.Fatal(err)
			}
			safe, _ := entry["safe_error"].(string)
			if !strings.Contains(safe, "upstream failed") || !strings.Contains(safe, "diagnostic-visible") {
				t.Fatalf("diagnostic text lost: %s", safe)
			}
		})
	}
}

type failingRequestBody struct{}

func (failingRequestBody) Read([]byte) (int, error) { return 0, errors.New(errorCanaries) }
func (failingRequestBody) Close() error             { return nil }

func TestSafeErrorInputAndServiceWrappers(t *testing.T) {
	for _, mode := range []string{"body-read", "unknown-field", "novel-invalid", "media-service"} {
		t.Run(mode, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/test", strings.NewReader(`{"provider-canary":true}`))
			if mode == "body-read" {
				req.Body = failingRequestBody{}
			}
			h := handler{}
			rec := httptest.NewRecorder()
			h.withObservability(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch mode {
				case "novel-invalid":
					h.novelPanelError(w, r, fmt.Errorf("%w: %s", novelpanel.ErrInvalid, errorCanaries))
				case "media-service":
					h.swerr(w, r, errors.New(errorCanaries))
				default:
					var body struct{ Name string }
					if err := decodeJSON(w, r, &body); err != nil {
						writeError(w, 400, err.Error())
					}
				}
			})).ServeHTTP(rec, req)
			for _, canary := range []string{"pass-canary", "cookie-canary", "bearer-canary", "provider-canary", "dsn-canary"} {
				if strings.Contains(rec.Body.String(), canary) {
					t.Errorf("response leaked %s: %s", canary, rec.Body.String())
				}
			}
		})
	}
}

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
