package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

type fakeAuthService struct {
	user          authn.User
	loginCreds    authn.Credentials
	refreshCreds  authn.Credentials
	authErr       error
	loginErr      error
	refreshErr    error
	logoutErr     error
	loginCalls    int
	refreshCalls  int
	logoutCalls   int
}

func (f *fakeAuthService) Login(context.Context, string, string) (authn.Credentials, authn.User, error) {
	f.loginCalls++
	return f.loginCreds, f.user, f.loginErr
}
func (f *fakeAuthService) AuthenticateAccess(context.Context, string) (authn.User, error) {
	if f.authErr != nil { return authn.User{}, f.authErr }
	return f.user, nil
}
func (f *fakeAuthService) Refresh(context.Context, string) (authn.Credentials, authn.User, error) {
	f.refreshCalls++
	return f.refreshCreds, f.user, f.refreshErr
}
func (f *fakeAuthService) Logout(context.Context, string, string) error {
	f.logoutCalls++
	return f.logoutErr
}

func sameOrigin(req *http.Request) {
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
}

func TestCurrentUserRequiresValidSession(t *testing.T) {
	auth := &fakeAuthService{authErr: authn.ErrUnauthenticated}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodGet, "/api/auth/current-user", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "expired-access"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s, want 401", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if body["code"] != "AUTH_UNAUTHENTICATED" {
		t.Fatalf("body = %s, want AUTH_UNAUTHENTICATED", rec.Body.String())
	}
}

func TestCurrentUserReturnsOnlySafeProfile(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Username: "alice", DisplayName: "Alice", Role: "member", TeamID: 3, Capabilities: []string{"batch.read"}}}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodGet, "/api/auth/current-user", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	body := rec.Body.String()
	for _, want := range []string{`"id":7`, `"username":"alice"`, `"role":"member"`, `"batch.read"`} {
		if !strings.Contains(body, want) { t.Fatalf("body=%s missing %s", body, want) }
	}
	for _, forbidden := range []string{"password", "refreshToken", "accessToken", "apiKey", "secret"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) { t.Fatalf("sensitive field leaked: %s", body) }
	}
}

func TestRefreshRotatesHttpOnlyCookiesWithoutReturningRawCredentials(t *testing.T) {
	auth := &fakeAuthService{
		user: authn.User{ID: 7, Username: "alice"},
		refreshCreds: authn.Credentials{AccessToken: "new-access-secret", RefreshToken: "new-refresh-secret"},
	}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	sameOrigin(req)
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "old-refresh"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if auth.refreshCalls != 1 { t.Fatalf("refresh calls=%d want 1", auth.refreshCalls) }
	cookies := rec.Result().Cookies()
	if len(cookies) < 2 { t.Fatalf("cookies=%v, want access + refresh", cookies) }
	seen := map[string]bool{}
	for _, cookie := range cookies {
		if cookie.HttpOnly != true { t.Fatalf("cookie %s must be HttpOnly", cookie.Name) }
		if cookie.SameSite != http.SameSiteLaxMode { t.Fatalf("cookie %s SameSite=%v, want Lax", cookie.Name, cookie.SameSite) }
		seen[cookie.Name] = true
	}
	if !seen[AccessCookieName] || !seen[RefreshCookieName] { t.Fatalf("cookies missing auth names: %v", seen) }
	if strings.Contains(rec.Body.String(), "new-access-secret") || strings.Contains(rec.Body.String(), "new-refresh-secret") {
		t.Fatalf("raw credential leaked in response body: %s", rec.Body.String())
	}
}

func TestLogoutRevokesServerSessionAndClearsCookies(t *testing.T) {
	auth := &fakeAuthService{}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	sameOrigin(req)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "refresh"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if auth.logoutCalls != 1 { t.Fatalf("logout calls=%d want 1", auth.logoutCalls) }
	for _, cookie := range rec.Result().Cookies() {
		if (cookie.Name == AccessCookieName || cookie.Name == RefreshCookieName) && cookie.MaxAge >= 0 {
			t.Fatalf("cookie %s was not cleared: %#v", cookie.Name, cookie)
		}
	}
}

func TestStaleRefreshReturns401WithoutLeakingSecrets(t *testing.T) {
	auth := &fakeAuthService{refreshErr: authn.ErrUnauthenticated}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	sameOrigin(req)
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "stale"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized { t.Fatalf("status=%d body=%s, want 401", rec.Code, rec.Body.String()) }
}

func TestRefreshBackendFailureReturnsSafe5xxWithoutLeakingSecrets(t *testing.T) {
	auth := &fakeAuthService{refreshErr: errors.New("database says refresh=super-secret")}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	sameOrigin(req)
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "present"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code < 500 { t.Fatalf("status=%d body=%s, want safe 5xx", rec.Code, rec.Body.String()) }
	if strings.Contains(rec.Body.String(), "super-secret") { t.Fatalf("auth error leaked secret: %s", rec.Body.String()) }
}
