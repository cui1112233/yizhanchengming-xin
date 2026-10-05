package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestLoginRateLimiterBlocksRepeatedFailuresWithoutCallingAuthAgain(t *testing.T) {
	now := time.Date(2026, 10, 5, 10, 45, 0, 0, time.UTC)
	limiter := NewLoginRateLimiter(LoginRateLimitOptions{
		MaxFailures: 2,
		Window:      time.Minute,
		Now:         func() time.Time { return now },
	})
	auth := &fakeAuthService{loginErr: authn.ErrUnauthenticated}
	handler := NewHandler(Dependencies{Auth: auth, LoginLimiter: limiter})

	for attempt := 1; attempt <= 2; attempt++ {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"alice","password":"wrong-password"}`))
		sameOrigin(req)
		req.RemoteAddr = "203.0.113.10:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status=%d body=%s, want 401", attempt, rec.Code, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"alice","password":"wrong-password"}`))
	sameOrigin(req)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status=%d body=%s, want 429", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"AUTH_RATE_LIMITED"`) {
		t.Fatalf("body=%s, want AUTH_RATE_LIMITED", rec.Body.String())
	}
	if auth.loginCalls != 2 {
		t.Fatalf("login calls=%d want 2; blocked attempt must not hit password verification", auth.loginCalls)
	}
}

func TestLoginRateLimiterResetsAfterSuccessfulLogin(t *testing.T) {
	now := time.Date(2026, 10, 5, 10, 45, 0, 0, time.UTC)
	limiter := NewLoginRateLimiter(LoginRateLimitOptions{MaxFailures: 2, Window: time.Minute, Now: func() time.Time { return now }})
	auth := &fakeAuthService{loginErr: authn.ErrUnauthenticated}
	handler := NewHandler(Dependencies{Auth: auth, LoginLimiter: limiter})

	request := func(password string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"alice","password":"`+password+`"}`))
		sameOrigin(req)
		req.RemoteAddr = "203.0.113.11:4321"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	if rec := request("wrong"); rec.Code != http.StatusUnauthorized { t.Fatalf("first failure status=%d", rec.Code) }
	auth.loginErr = nil
	auth.loginCreds = authn.Credentials{AccessToken: "access", RefreshToken: "refresh"}
	if rec := request("correct-password"); rec.Code != http.StatusOK { t.Fatalf("success status=%d body=%s", rec.Code, rec.Body.String()) }
	auth.loginErr = errors.New("backend unavailable")
	if rec := request("anything"); rec.Code != http.StatusInternalServerError {
		t.Fatalf("post-success request status=%d body=%s; limiter should have reset", rec.Code, rec.Body.String())
	}
}

func TestLoginRateLimiterWindowExpires(t *testing.T) {
	now := time.Date(2026, 10, 5, 10, 45, 0, 0, time.UTC)
	limiter := NewLoginRateLimiter(LoginRateLimitOptions{MaxFailures: 1, Window: time.Minute, Now: func() time.Time { return now }})
	key := limiterKey("203.0.113.12:9999", "Alice")
	limiter.RecordFailure(key)
	if limiter.Allow(key) { t.Fatal("key should be blocked inside window") }
	now = now.Add(time.Minute + time.Second)
	if !limiter.Allow(key) { t.Fatal("key should be allowed after window expires") }
}
