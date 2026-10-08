package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type task14ProjectAccessStub struct {
	allowed bool
	err     error
}

func (s task14ProjectAccessStub) CanAccessBatchProject(context.Context, int64, int64, int64, bool) (bool, error) {
	return s.allowed, s.err
}

type task14VideoResourceProjectStub struct{}

func (task14VideoResourceProjectStub) ProjectIDForProductionTask(context.Context, int64) (int64, error) {
	return 77, nil
}
func (task14VideoResourceProjectStub) ProjectIDForMergeJob(context.Context, int64) (int64, error) {
	return 77, nil
}
func (task14VideoResourceProjectStub) ProjectIDForMergeAttempt(context.Context, int64) (int64, error) {
	return 77, nil
}

type videoResourceResolverSpy struct {
	projectID int64
	err       error
	calls     int
}

func (s *videoResourceResolverSpy) resolve() (int64, error) {
	s.calls++
	return s.projectID, s.err
}
func (s *videoResourceResolverSpy) ProjectIDForProductionTask(context.Context, int64) (int64, error) {
	return s.resolve()
}
func (s *videoResourceResolverSpy) ProjectIDForMergeJob(context.Context, int64) (int64, error) {
	return s.resolve()
}
func (s *videoResourceResolverSpy) ProjectIDForMergeAttempt(context.Context, int64) (int64, error) {
	return s.resolve()
}

type videoBusinessSpy struct{ calls int }

func (s *videoBusinessSpy) Start(context.Context, video.StartRequest) (video.StartResult, error) {
	s.calls++
	return video.StartResult{}, nil
}
func (s *videoBusinessSpy) PollTask(context.Context, int64) (video.ProductionTask, error) {
	s.calls++
	return video.ProductionTask{}, nil
}
func (s *videoBusinessSpy) CancelTask(context.Context, int64) (video.ProductionTask, error) {
	s.calls++
	return video.ProductionTask{}, nil
}
func (s *videoBusinessSpy) RetryTask(context.Context, int64, string) (video.StartResult, error) {
	s.calls++
	return video.StartResult{}, nil
}

func videoResourceRequest() *http.Request {
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/video-tasks/31/poll", strings.NewReader(`{}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	req.Header.Set("Origin", "http://example.com")
	return req
}

func TestVideoResourceAccessFailsBeforeResolverWhenProjectPolicyIsMissing(t *testing.T) {
	resolver := &videoResourceResolverSpy{projectID: 77}
	business := &videoBusinessSpy{}
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{Auth: auth, VideoResourceProjects: resolver, Video: business})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, videoResourceRequest())

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "AUTH_POLICY_UNAVAILABLE") || !strings.Contains(rec.Body.String(), "request_id") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if resolver.calls != 0 || business.calls != 0 {
		t.Fatalf("resolver=%d business=%d", resolver.calls, business.calls)
	}
}

func TestVideoResourceAccessSafelyRejectsResolverFailureBeforeBusiness(t *testing.T) {
	resolver := &videoResourceResolverSpy{err: errors.New("mysql dsn=secret")}
	business := &videoBusinessSpy{}
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{Auth: auth, BatchProjectAccess: task14ProjectAccessStub{allowed: true}, VideoResourceProjects: resolver, Video: business})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, videoResourceRequest())

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "AUTH_POLICY_UNAVAILABLE") || !strings.Contains(rec.Body.String(), "request_id") || strings.Contains(rec.Body.String(), "dsn=secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if resolver.calls != 1 || business.calls != 0 {
		t.Fatalf("resolver=%d business=%d", resolver.calls, business.calls)
	}
}

func TestVideoResourceAccessRejectsForeignResolvedProjectBeforeBusiness(t *testing.T) {
	resolver := &videoResourceResolverSpy{projectID: 77}
	business := &videoBusinessSpy{}
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{Auth: auth, BatchProjectAccess: task14ProjectAccessStub{allowed: false}, VideoResourceProjects: resolver, Video: business})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, videoResourceRequest())

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "VIDEO_NOT_FOUND") || !strings.Contains(rec.Body.String(), "request_id") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if resolver.calls != 1 || business.calls != 0 {
		t.Fatalf("resolver=%d business=%d", resolver.calls, business.calls)
	}
}

func TestVideoResourceAccessDoesNotRevealMissingVersusForeignProject(t *testing.T) {
	serve := func(resolver *videoResourceResolverSpy, access task14ProjectAccessStub) (*httptest.ResponseRecorder, *videoBusinessSpy) {
		business := &videoBusinessSpy{}
		auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
		h := NewHandler(Dependencies{Auth: auth, BatchProjectAccess: access, VideoResourceProjects: resolver, Video: business})
		req := videoResourceRequest()
		req.Header.Set(testRequestIDHeader, "video-oracle-test")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec, business
	}

	missingResolver := &videoResourceResolverSpy{err: video.ErrNotFound}
	missing, missingBusiness := serve(missingResolver, task14ProjectAccessStub{allowed: true})
	foreignResolver := &videoResourceResolverSpy{projectID: 77}
	foreign, foreignBusiness := serve(foreignResolver, task14ProjectAccessStub{allowed: false})
	if missing.Code != http.StatusNotFound || foreign.Code != http.StatusNotFound {
		t.Fatalf("missing=%d foreign=%d", missing.Code, foreign.Code)
	}
	if missing.Body.String() != foreign.Body.String() {
		t.Fatalf("resource oracle differs:\nmissing=%s\nforeign=%s", missing.Body.String(), foreign.Body.String())
	}
	if !strings.Contains(missing.Body.String(), `"code":"VIDEO_NOT_FOUND"`) || !strings.Contains(missing.Body.String(), `"message":"VIDEO resource 不存在"`) {
		t.Fatalf("unexpected envelope: %s", missing.Body.String())
	}
	if missingResolver.calls != 1 || foreignResolver.calls != 1 || missingBusiness.calls != 0 || foreignBusiness.calls != 0 {
		t.Fatalf("missingResolver=%d foreignResolver=%d missingBusiness=%d foreignBusiness=%d", missingResolver.calls, foreignResolver.calls, missingBusiness.calls, foreignBusiness.calls)
	}
}

func TestVideoResourceAccessPolicyErrorFailsClosedBeforeBusiness(t *testing.T) {
	resolver := &videoResourceResolverSpy{projectID: 77}
	business := &videoBusinessSpy{}
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{Auth: auth, BatchProjectAccess: task14ProjectAccessStub{err: errors.New("mysql secret")}, VideoResourceProjects: resolver, Video: business})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, videoResourceRequest())

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "AUTH_POLICY_UNAVAILABLE") || strings.Contains(rec.Body.String(), "mysql secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if resolver.calls != 1 || business.calls != 0 {
		t.Fatalf("resolver=%d business=%d", resolver.calls, business.calls)
	}
}

func TestVideoResourceAccessAuthorizedRequestReachesBusiness(t *testing.T) {
	resolver := &videoResourceResolverSpy{projectID: 77}
	business := &videoBusinessSpy{}
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{Auth: auth, BatchProjectAccess: task14ProjectAccessStub{allowed: true}, VideoResourceProjects: resolver, Video: business})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, videoResourceRequest())

	if rec.Code != http.StatusOK || resolver.calls != 1 || business.calls != 1 {
		t.Fatalf("status=%d resolver=%d business=%d body=%s", rec.Code, resolver.calls, business.calls, rec.Body.String())
	}
}

func TestTask14ProjectRoutesEnforceTask15OwnershipPolicy(t *testing.T) {
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView, CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{
		Auth:                  auth,
		BatchProjectAccess:    task14ProjectAccessStub{allowed: false},
		VideoResourceProjects: task14VideoResourceProjectStub{},
	})

	tests := []struct {
		method     string
		path       string
		wantStatus int
		wantCode   string
	}{
		{http.MethodGet, "/api/v1/batch-projects/77/video", http.StatusForbidden, "AUTH_FORBIDDEN"},
		{http.MethodPost, "/api/v1/batch-projects/77/books/9/video", http.StatusForbidden, "AUTH_FORBIDDEN"},
		{http.MethodPost, "/api/v1/batch-projects/77/books/9/merge", http.StatusForbidden, "AUTH_FORBIDDEN"},
		{http.MethodPost, "/api/v1/video-tasks/31/poll", http.StatusNotFound, "VIDEO_NOT_FOUND"},
		{http.MethodPost, "/api/v1/video-tasks/31/cancel", http.StatusNotFound, "VIDEO_NOT_FOUND"},
		{http.MethodPost, "/api/v1/video-tasks/31/retry", http.StatusNotFound, "VIDEO_NOT_FOUND"},
		{http.MethodGet, "/api/v1/video-merge-jobs/41", http.StatusNotFound, "VIDEO_NOT_FOUND"},
		{http.MethodPost, "/api/v1/video-merge-attempts/51/retry", http.StatusNotFound, "VIDEO_NOT_FOUND"},
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
			if rec.Code != tc.wantStatus || !strings.Contains(rec.Body.String(), tc.wantCode) {
				t.Fatalf("%s %s: status=%d body=%s", tc.method, tc.path, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestBatchProjectAndVideoResourceInvalidIDsKeepDomainSpecificSafeCodes(t *testing.T) {
	auth := task14AuthStub{user: authn.User{ID: 12, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView, CapabilityBatchExecute}}}
	h := NewHandler(Dependencies{
		Auth:                  auth,
		BatchProjectAccess:    task14ProjectAccessStub{allowed: true},
		VideoResourceProjects: task14VideoResourceProjectStub{},
	})

	projectReq := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/batch-projects/0", nil)
	projectReq.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	projectRec := httptest.NewRecorder()
	h.ServeHTTP(projectRec, projectReq)
	if projectRec.Code != http.StatusBadRequest || !strings.Contains(projectRec.Body.String(), "BATCH_PROJECT_INVALID_REQUEST") || !strings.Contains(projectRec.Body.String(), "批量项目 ID 无效") {
		t.Fatalf("project status=%d body=%s", projectRec.Code, projectRec.Body.String())
	}

	videoReq := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/video-tasks/0/poll", strings.NewReader(`{}`))
	videoReq.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	videoReq.Header.Set("Origin", "http://example.com")
	videoRec := httptest.NewRecorder()
	h.ServeHTTP(videoRec, videoReq)
	if videoRec.Code != http.StatusBadRequest || !strings.Contains(videoRec.Body.String(), "VIDEO_INVALID_REQUEST") {
		t.Fatalf("video status=%d body=%s", videoRec.Code, videoRec.Body.String())
	}
}
