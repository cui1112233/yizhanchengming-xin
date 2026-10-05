package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
		{
			ID: 51, IntakeID: 11, Name: "跨书城批次",
			Sources: []string{"点众", "知乎"}, BookCount: 3,
			Genders: []string{"女频", "男频"}, Styles: []string{"情感", "悬疑"},
			RunStatus: intake.RunStatusRunning,
		},
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
	for _, want := range []string{
		`"projects"`, `"id":51`, `"intakeId":11`, `"name":"跨书城批次"`,
		`"sources":["点众","知乎"]`, `"bookCount":3`,
		`"genders":["女频","男频"]`, `"styles":["情感","悬疑"]`,
		`"runStatus":"running"`, `"id":52`, `"name":"点众批次"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, missing %s", body, want)
		}
	}
}

func TestBatchProjectListDoesNotLeakReaderError(t *testing.T) {
	reader := &fakeBatchProjectReader{err: errors.New("mysql password=secret")}
	handler := NewHandler(Dependencies{BatchProjects: reader})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s, want 500", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password=secret") {
		t.Fatalf("reader error leaked to response: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "读取批量项目列表失败") {
		t.Fatalf("body = %s, want generic list error", rec.Body.String())
	}
}
