package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type archivedGuardMergeSpy struct{ calls int }

func (s *archivedGuardMergeSpy) StartFromProductionTasks(context.Context, video.MergeProductionStartRequest) (video.MergeResult, error) {
	s.calls++
	return video.MergeResult{}, nil
}
func (s *archivedGuardMergeSpy) Get(context.Context, int64) (video.MergeJob, []video.MergeAttempt, error) {
	s.calls++
	return video.MergeJob{}, nil, nil
}
func (s *archivedGuardMergeSpy) RetryAttempt(context.Context, int64) (video.MergeResult, error) {
	s.calls++
	return video.MergeResult{}, nil
}

func TestArchivedBatchProjectRejectsEveryDirectBusinessMutationBeforeService(t *testing.T) {
	for _, route := range allDirectBatchProjectObjectRoutes() {
		if route.method == http.MethodGet {
			continue
		}
		t.Run(route.name, func(t *testing.T) {
			business := &batchProjectBusinessSpy{}
			lifecycle := &fakeBatchProjectLifecycle{archived: true}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{route.capability}}}
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{
				Auth: auth, BatchProjectAccess: &batchProjectAccessSpy{allowed: true}, BatchProjectLifecycle: lifecycle,
				BatchProjectDetails: business, ScriptBooks: business, ScriptStoryboards: business,
				UnifiedSettings: business, Generation: business,
			}).ServeHTTP(rec, batchProjectRouteRequest(route))

			if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_ARCHIVED"`) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if business.calls != 0 || lifecycle.stateCalls != 1 || lifecycle.projectID != 51 {
				t.Fatalf("business=%d lifecycle=%+v", business.calls, lifecycle)
			}
		})
	}
}

func TestArchivedBatchProjectRejectsIntakeMutationAliasesBeforeService(t *testing.T) {
	routes := []struct {
		name, method, path, body, capability string
	}{
		{name: "execute intake", method: http.MethodPost, path: "/api/v1/intakes/11/execute", body: `{"maxText":4000}`, capability: CapabilityBatchExecute},
		{name: "save workshop", method: http.MethodPut, path: "/api/v1/intakes/11/workshop", body: `{"settings":{}}`, capability: CapabilityBatchConfigure},
		{name: "restore book", method: http.MethodPost, path: "/api/v1/intakes/11/books/21/restore", body: `{"maxText":4000}`, capability: CapabilityBatchExecute},
		{name: "create another run", method: http.MethodPost, path: "/api/v1/intakes/11/batch-projects", body: `{}`, capability: CapabilityBatchExecute},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			business := &intakeBusinessSpy{}
			lifecycle := &fakeBatchProjectLifecycle{archived: true}
			auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{route.capability}}}
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{
				Auth: auth, IntakeAccess: &intakeAccessStub{allowed: true}, BatchProjectLifecycle: lifecycle,
				Intakes: business, Reader: business, Workshop: business, Pipeline: business,
			}).ServeHTTP(rec, authenticatedIntakeRequest(route.method, route.path, route.body))

			if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_ARCHIVED"`) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if business.calls != 0 || lifecycle.intakeStateCalls != 1 || lifecycle.projectID != 11 {
				t.Fatalf("business=%d lifecycle=%+v", business.calls, lifecycle)
			}
		})
	}
}

func TestArchivedBatchProjectStillAllowsReadsAndRestore(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Role: "member", Capabilities: []string{CapabilityBatchView, CapabilityBatchConfigure}}}
	access := &batchProjectAccessSpy{allowed: true}
	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodGet, path: "/api/v1/batch-projects/51"},
		{method: http.MethodGet, path: "/api/v1/batch-projects/51/generation"},
		{method: http.MethodPost, path: "/api/v1/batch-projects/51/restore"},
	} {
		lifecycle := &fakeBatchProjectLifecycle{archived: true}
		req := httptest.NewRequest(tc.method, "http://app.example"+tc.path, strings.NewReader(tc.body))
		req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
		if tc.method != http.MethodGet {
			sameOrigin(req)
		}
		rec := httptest.NewRecorder()
		NewHandler(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: lifecycle}).ServeHTTP(rec, req)
		if rec.Code == http.StatusConflict && strings.Contains(rec.Body.String(), "BATCH_PROJECT_ARCHIVED") {
			t.Fatalf("%s %s was incorrectly blocked: %s", tc.method, tc.path, rec.Body.String())
		}
		if tc.method == http.MethodGet && lifecycle.stateCalls != 0 {
			t.Fatalf("read route consulted mutation guard: %+v", lifecycle)
		}
		if strings.HasSuffix(tc.path, "/restore") && lifecycle.restoreCalls != 1 {
			t.Fatalf("restore calls=%d body=%s", lifecycle.restoreCalls, rec.Body.String())
		}
	}
}

func TestArchivedProjectBlocksIndirectVideoMutationsButAllowsCancel(t *testing.T) {
	auth := task14AuthStub{user: authn.User{ID: 7, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	for _, tc := range []struct {
		name        string
		path        string
		wantBlocked bool
		merge       bool
	}{
		{name: "poll", path: "/api/v1/video-tasks/31/poll", wantBlocked: true},
		{name: "retry", path: "/api/v1/video-tasks/31/retry", wantBlocked: true},
		{name: "merge retry", path: "/api/v1/video-merge-attempts/51/retry", wantBlocked: true, merge: true},
		{name: "cancel", path: "/api/v1/video-tasks/31/cancel"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lifecycle := &fakeBatchProjectLifecycle{archived: true}
			videoSpy := &videoBusinessSpy{}
			mergeSpy := &archivedGuardMergeSpy{}
			req := httptest.NewRequest(http.MethodPost, "http://app.example"+tc.path, strings.NewReader(`{}`))
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
			sameOrigin(req)
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{
				Auth: auth, BatchProjectAccess: task14ProjectAccessStub{allowed: true}, BatchProjectLifecycle: lifecycle,
				VideoResourceProjects: task14VideoResourceProjectStub{}, Video: videoSpy, VideoMerge: mergeSpy,
			}).ServeHTTP(rec, req)
			if tc.wantBlocked {
				if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_ARCHIVED"`) || lifecycle.stateCalls != 1 || videoSpy.calls != 0 || mergeSpy.calls != 0 {
					t.Fatalf("status=%d lifecycle=%+v video=%d merge=%d body=%s", rec.Code, lifecycle, videoSpy.calls, mergeSpy.calls, rec.Body.String())
				}
				return
			}
			if rec.Code != http.StatusOK || lifecycle.stateCalls != 0 || videoSpy.calls != 1 {
				t.Fatalf("cancel status=%d lifecycle=%+v video=%d body=%s", rec.Code, lifecycle, videoSpy.calls, rec.Body.String())
			}
		})
	}
}
