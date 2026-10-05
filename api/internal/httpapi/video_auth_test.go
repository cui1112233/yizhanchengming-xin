package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

type task14AuthStub struct{ user authn.User }

func (s task14AuthStub) Login(context.Context, string, string) (authn.Credentials, authn.User, error) {
	return authn.Credentials{}, authn.User{}, authn.ErrUnauthenticated
}
func (s task14AuthStub) AuthenticateAccess(context.Context, string) (authn.User, error) { return s.user, nil }
func (s task14AuthStub) Refresh(context.Context, string) (authn.Credentials, authn.User, error) {
	return authn.Credentials{}, authn.User{}, authn.ErrUnauthenticated
}
func (s task14AuthStub) Logout(context.Context, string, string) error { return nil }

func TestTask14BrowserRoutesRequireTask15Session(t *testing.T) {
	h := NewHandler(Dependencies{Auth: task14AuthStub{user: authn.User{ID: 1, Role: "member"}}})
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/video-providers/personal_api/models/yd2.0-mini/status"},
		{http.MethodGet, "/api/v1/video-providers/personal_api/models/yd2.0-mini"},
		{http.MethodGet, "/api/v1/batch-projects/1/video"},
		{http.MethodPost, "/api/v1/batch-projects/1/books/1/video"},
		{http.MethodPost, "/api/v1/video-tasks/1/cancel"},
		{http.MethodPost, "/api/v1/video-tasks/1/retry"},
		{http.MethodPost, "/api/v1/batch-projects/1/books/1/merge"},
	}
	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://example.com"+tc.path, strings.NewReader(`{}`))
			if tc.method != http.MethodGet {
				req.Header.Set("Origin", "http://example.com")
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "AUTH_UNAUTHENTICATED") {
				t.Fatalf("%s %s: status=%d body=%s", tc.method, tc.path, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestTask14ProviderConfigRequiresConfigureCapability(t *testing.T) {
	h := NewHandler(Dependencies{Auth: task14AuthStub{user: authn.User{ID: 1, Role: "member", Capabilities: []string{CapabilityBatchView}}}})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/video-providers/personal_api/models/yd2.0-mini", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "AUTH_FORBIDDEN") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTask14ExecutorRegistrationIsNeverPublic(t *testing.T) {
	h := NewHandler(Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/video/local-executors/register", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "VIDEO_EXECUTOR_BOOTSTRAP_UNAVAILABLE") {
		t.Fatalf("unconfigured status=%d body=%s", rec.Code, rec.Body.String())
	}

	h = NewHandler(Dependencies{VideoExecutorBootstrapToken: "bootstrap-secret"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/video/local-executors/register", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer wrong-secret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || !strings.Contains(rec.Body.String(), "VIDEO_EXECUTOR_UNAUTHENTICATED") {
		t.Fatalf("invalid credential status=%d body=%s", rec.Code, rec.Body.String())
	}
}
