package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

type task14ProjectAccessStub struct {
	allowed bool
}

func (s task14ProjectAccessStub) CanAccessBatchProject(context.Context, int64, int64, int64, bool) (bool, error) {
	return s.allowed, nil
}

type task14VideoResourceProjectStub struct{}

func (task14VideoResourceProjectStub) ProjectIDForProductionTask(context.Context, int64) (int64, error) { return 77, nil }
func (task14VideoResourceProjectStub) ProjectIDForMergeJob(context.Context, int64) (int64, error)       { return 77, nil }
func (task14VideoResourceProjectStub) ProjectIDForMergeAttempt(context.Context, int64) (int64, error)   { return 77, nil }

func TestTask14ProjectRoutesEnforceTask15OwnershipPolicy(t *testing.T) {
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView, CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{
		Auth:                  auth,
		BatchProjectAccess:    task14ProjectAccessStub{allowed: false},
		VideoResourceProjects: task14VideoResourceProjectStub{},
	})

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/batch-projects/77/video"},
		{http.MethodPost, "/api/v1/batch-projects/77/books/9/video"},
		{http.MethodPost, "/api/v1/batch-projects/77/books/9/merge"},
		{http.MethodPost, "/api/v1/video-tasks/31/poll"},
		{http.MethodPost, "/api/v1/video-tasks/31/cancel"},
		{http.MethodPost, "/api/v1/video-tasks/31/retry"},
		{http.MethodGet, "/api/v1/video-merge-jobs/41"},
		{http.MethodPost, "/api/v1/video-merge-attempts/51/retry"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://example.com"+tc.path, strings.NewReader(`{}`))
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
			if tc.method != http.MethodGet {
				req.Header.Set("Origin", "http://example.com")
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "AUTH_FORBIDDEN") {
				t.Fatalf("%s %s: status=%d body=%s", tc.method, tc.path, rec.Code, rec.Body.String())
			}
		})
	}
}
