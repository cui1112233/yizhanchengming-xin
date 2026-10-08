package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type fakeBatchProjectReader struct {
	projects []intake.BatchProject
	err      error
	calls    int
	query    intake.BatchProjectListQuery
}

func (f *fakeBatchProjectReader) ListBatchProjects(_ context.Context, query intake.BatchProjectListQuery) (intake.BatchProjectPage, error) {
	f.calls++
	f.query = query
	return intake.BatchProjectPage{Projects: f.projects, Page: query.Page, Limit: query.Limit, Total: len(f.projects)}, f.err
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
	now := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC)
	reader := &fakeBatchProjectReader{projects: []intake.BatchProject{
		{
			ID: 51, IntakeID: 11, Name: "跨书城批次",
			Sources: []string{"点众", "知乎"}, BookCount: 3,
			Genders: []string{"女频", "男频"}, Styles: []string{"情感", "悬疑"},
			RunStatus: intake.RunStatusRunning, FailureCount: 2, CreatedAt: now, UpdatedAt: now,
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
		`"runStatus":"running"`, `"failureCount":2`, `"id":52`, `"name":"点众批次"`,
		`"createdAt":"2026-10-09T08:00:00Z"`, `"updatedAt":"2026-10-09T08:00:00Z"`, `"archivedAt":null`,
		`"page":1`, `"limit":20`, `"total":2`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, missing %s", body, want)
		}
	}
}

func TestBatchProjectListParsesScopedFiltersPaginationAndAdminElevation(t *testing.T) {
	reader := &fakeBatchProjectReader{}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "admin", Capabilities: []string{CapabilityBatchView}}}
	handler := NewHandler(Dependencies{Auth: auth, BatchProjects: reader, BatchProjectAccess: &batchProjectAccessSpy{allowed: true}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects?q=Hero&source=%E7%9F%A5%E4%B9%8E&status=failed&archived=all&page=2&pageSize=12&sort=name_asc", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	query := reader.query
	if query.UserID != 7 || query.TeamID != 3 || !query.Elevated || query.Query != "Hero" || query.Source != "知乎" || query.Status != intake.RunStatusFailed || query.Archived != intake.BatchProjectArchivedAll || query.Page != 2 || query.Limit != 12 || query.Sort != intake.BatchProjectSortNameAsc {
		t.Fatalf("query=%+v", query)
	}
}

func TestBatchProjectListRejectsInvalidQueryBeforeReader(t *testing.T) {
	for _, query := range []string{"page=0", "limit=101", "status=secret", "archived=false", "sort=random"} {
		t.Run(query, func(t *testing.T) {
			reader := &fakeBatchProjectReader{}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects?"+query, nil)
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{BatchProjects: reader}).ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_INVALID_REQUEST"`) || reader.calls != 0 {
				t.Fatalf("status=%d calls=%d body=%s", rec.Code, reader.calls, rec.Body.String())
			}
		})
	}
}

func TestBatchProjectListElevatesOnlyAdminAndOwner(t *testing.T) {
	for _, tc := range []struct {
		role string
		want bool
	}{
		{role: "admin", want: true}, {role: "owner", want: true},
		{role: "dev"}, {role: "manager"}, {role: "member"},
	} {
		t.Run(tc.role, func(t *testing.T) {
			reader := &fakeBatchProjectReader{}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: tc.role, Capabilities: []string{CapabilityBatchView}}}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{Auth: auth, BatchProjects: reader, BatchProjectAccess: &batchProjectAccessSpy{allowed: true}}).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK || reader.query.Elevated != tc.want {
				t.Fatalf("status=%d elevated=%v body=%s", rec.Code, reader.query.Elevated, rec.Body.String())
			}
		})
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
