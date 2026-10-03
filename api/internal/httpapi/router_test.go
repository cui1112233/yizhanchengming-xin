package httpapi

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestRouterExposesHealthAndBatchFactory(t *testing.T) {
    starter := &fakeStarter{}
    handler := NewRouter(starter)

    health := httptest.NewRecorder()
    handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
    if health.Code != http.StatusOK { t.Fatalf("health status=%d", health.Code) }

    batch := httptest.NewRecorder()
    handler.ServeHTTP(batch, httptest.NewRequest(http.MethodGet, "/api/batch-factory/jobs", nil))
    if batch.Code != http.StatusMethodNotAllowed { t.Fatalf("batch route status=%d", batch.Code) }
}
