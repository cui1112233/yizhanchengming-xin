package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestProgressivelyProtectedBusinessReadsReturn401WithoutSession(t *testing.T) {
	handler := NewHandler(Dependencies{Auth: &fakeAuthService{}})
	for _, path := range []string{
		"/api/v1/intakes",
		"/api/v1/batch-projects",
		"/api/v1/batch-projects/51/settings",
		"/api/v1/batch-projects/51/generation",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s, want 401", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHENTICATED"`) {
				t.Fatalf("body=%s, want AUTH_UNAUTHENTICATED", rec.Body.String())
			}
		})
	}
}

func TestProgressivelyProtectedSettingsAndGenerationReturn403WithoutCapabilities(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Role: "member"}}
	handler := NewHandler(Dependencies{Auth: auth})
	for _, path := range []string{
		"/api/v1/batch-projects/51/settings",
		"/api/v1/batch-projects/51/generation",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid-access"})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"AUTH_FORBIDDEN"`) {
				t.Fatalf("body=%s, want AUTH_FORBIDDEN", rec.Body.String())
			}
		})
	}
}
