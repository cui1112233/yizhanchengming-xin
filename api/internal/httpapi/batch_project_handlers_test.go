package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type fakeBatchProjectReader struct {
	projects []intake.BatchProject
	err      error
	calls    int
}

func (f *fakeBatchProjectReader) ListBatchProjects(context.Context) ([]intake.BatchProject, error) {
	f.calls++
	return f.projects, f.err
}

func TestBatchProjectListRouteIsRegistered(t *testing.T) {
	handler := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s, want 503 when reader is not wired", rec.Code, rec.Body.String())
	}
}

func TestBatchProjectListReturnsReaderProjects(t *testing.T) {
	reader := &fakeBatchProjectReader{projects: []intake.BatchProject{
		{ID: 51, IntakeID: 11, Name: "知乎批次"},
		{ID: 52, IntakeID: 12, Name: "点众批次"},
	}}
	handler := NewHandler(Dependencies{BatchProjects: reader})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	if reader.calls != 1 {
		t.Fatalf("reader calls = %d, want 1", reader.calls)
	}
	body := rec.Body.String()
	for _, want := range []string{`"projects"`, `"id":51`, `"intakeId":11`, `"name":"知乎批次"`, `"id":52`, `"name":"点众批次"`} {
		if !contains(body, want) {
			t.Fatalf("body = %s, missing %s", body, want)
		}
	}
}

func contains(value, part string) bool {
	return len(part) == 0 || len(value) >= len(part) && index(value, part) >= 0
}

func index(value, part string) int {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
