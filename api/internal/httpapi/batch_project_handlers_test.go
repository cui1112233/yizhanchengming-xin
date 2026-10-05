package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchProjectListRouteIsRegistered(t *testing.T) {
	handler := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s, want 503 when reader is not wired", rec.Code, rec.Body.String())
	}
}
