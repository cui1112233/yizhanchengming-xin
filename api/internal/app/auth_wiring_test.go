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
