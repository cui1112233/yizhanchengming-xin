package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchProjectDetailRouteIsRegistered(t *testing.T) {
	handler := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/51", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s, want 503 when detail reader is not wired", rec.Code, rec.Body.String())
	}
}
