package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestCSRFBrowserMutationRejectsCrossSiteOriginBeforeBusinessAction(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Role: "admin"}}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Host = "app.example"
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "old-refresh"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if !strings.Contains(rec.Body.String(), "CSRF_REJECTED") { t.Fatalf("body=%s, want CSRF_REJECTED", rec.Body.String()) }
	if auth.refreshCalls != 0 { t.Fatalf("refresh calls=%d, cross-site request must be rejected before auth mutation", auth.refreshCalls) }
}

func TestCSRFBrowserMutationAllowsSameOrigin(t *testing.T) {
	auth := &fakeAuthService{
		user: authn.User{ID: 7, Role: "admin"},
		refreshCreds: authn.Credentials{AccessToken: "next-access", RefreshToken: "next-refresh"},
	}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Host = "app.example"
	req.Header.Set("Origin", "https://app.example")
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "old-refresh"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s, want 200", rec.Code, rec.Body.String()) }
	if auth.refreshCalls != 1 { t.Fatalf("refresh calls=%d want 1", auth.refreshCalls) }
}

func TestCSRFBrowserMutationRejectsMissingOriginAndReferer(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Role: "admin"}}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Host = "app.example"
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if auth.logoutCalls != 0 { t.Fatalf("logout calls=%d, missing CSRF source must be rejected", auth.logoutCalls) }
}

func TestRefreshRotationRaceFailureDoesNotClearPotentiallyNewerBrowserCookies(t *testing.T) {
	auth := &fakeAuthService{refreshErr: authn.ErrUnauthenticated}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.Host = "app.example"
	req.Header.Set("Origin", "https://app.example")
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "already-rotated-in-another-tab"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized { t.Fatalf("status=%d body=%s, want 401", rec.Code, rec.Body.String()) }
	for _, cookie := range rec.Result().Cookies() {
		if (cookie.Name == AccessCookieName || cookie.Name == RefreshCookieName) && cookie.MaxAge < 0 {
			t.Fatalf("stale refresh response must not clear a newer cookie from another tab: %#v", cookie)
		}
	}
}
