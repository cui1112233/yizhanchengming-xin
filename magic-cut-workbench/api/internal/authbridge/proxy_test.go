package authbridge

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxyForwardsOnlyLoginSessionWithBearerToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/login/session" {
			t.Fatalf("upstream path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer account-token" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Set-Cookie", "account_session=opaque; HttpOnly; SameSite=Lax")
		_, _ = w.Write([]byte(`{"username":"owner","active":true}`))
	}))
	defer upstream.Close()

	handler := New(upstream.URL, upstream.Client())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/login/session", nil)
	request.Header.Set("Authorization", "Bearer account-token")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Values("Set-Cookie"); len(got) != 1 || !strings.Contains(got[0], "HttpOnly") {
		t.Fatalf("Set-Cookie = %#v", got)
	}
}

func TestProxyUsesTheHttpOnlyBridgeCookieWhenBearerIsMissing(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer cookie-token" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"owner"}`))
	}))
	defer upstream.Close()

	handler := New(upstream.URL, upstream.Client())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/login/session", nil)
	request.AddCookie(&http.Cookie{Name: accountCookieName, Value: "cookie-token"})
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestBootstrapValidates4001TokenAndSetsHttpOnlyCookie(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/login/session" || r.Header.Get("Authorization") != "Bearer account-token" {
			t.Fatalf("path=%q auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"owner","active":true}`))
	}))
	defer upstream.Close()

	handler := NewBootstrap(upstream.URL, upstream.Client())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4040/api/auth/bootstrap", nil)
	request.Host = "127.0.0.1:4040"
	request.Header.Set("Origin", "http://127.0.0.1:4001")
	request.Header.Set("Authorization", "Bearer account-token")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:4001" {
		t.Fatalf("Access-Control-Allow-Origin = %q", recorder.Header().Get("Access-Control-Allow-Origin"))
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != accountCookieName || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookies = %#v", cookies)
	}
}

func TestProxyRejectsNonLoginBusinessPath(t *testing.T) {
	handler := New("http://127.0.0.1:4001", http.DefaultClient)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/projects", nil))

	if recorder.Code != http.StatusNotFound || !strings.Contains(recorder.Body.String(), "AUTH_BRIDGE_PATH_NOT_ALLOWED") {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestProxyReturnsStableUnavailableError(t *testing.T) {
	handler := New("http://127.0.0.1:1", &http.Client{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/login/session", nil))

	if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), "ACCOUNT_SERVICE_UNAVAILABLE") {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}
