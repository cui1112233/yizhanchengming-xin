package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProductionHandlerRequiresAuthenticationForBusinessAPI(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	handler := NewHandler(db, fakeFetcher{}, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s, want 401", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHENTICATED"`) {
		t.Fatalf("body = %s, want stable auth error code", rec.Body.String())
	}
}

func TestSecureCookiesDefaultToTrueOutsideDevelopment(t *testing.T) {
	t.Setenv("QIANTIE_ENV", "")
	t.Setenv("QIANTIE_COOKIE_SECURE", "false")
	if !secureCookiesEnabled() {
		t.Fatal("production-like environment must keep Secure cookies even when a stale false override exists")
	}
}

func TestDevelopmentCanExplicitlyDisableSecureCookiesForHTTP(t *testing.T) {
	t.Setenv("QIANTIE_ENV", "development")
	t.Setenv("QIANTIE_COOKIE_SECURE", "false")
	if secureCookiesEnabled() {
		t.Fatal("development environment should allow HTTP-compatible Secure=false")
	}

	t.Setenv("QIANTIE_COOKIE_SECURE", "true")
	if !secureCookiesEnabled() {
		t.Fatal("development environment should still allow opting into Secure cookies")
	}
}
