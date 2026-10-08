package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type fakeRuntimeService struct {
	projectID    int64
	item         task9runtime.WorkItem
	created      bool
	err          error
	resolveErr   error
	retryErr     error
	resolveCalls int
	retryCalls   int
}

func (f *fakeRuntimeService) ProjectIDForBookRun(context.Context, int64) (int64, error) {
	f.resolveCalls++
	if f.resolveErr != nil {
		return 0, f.resolveErr
	}
	if f.err != nil {
		return 0, f.err
	}
	return f.projectID, nil
}

func TestRuntimeRetryRejectsForeignURLProjectBeforeResolvingBookRun(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: false}
	runtime := &fakeRuntimeService{projectID: 7}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: &fakeBatchProjectLifecycle{}}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "AUTH_FORBIDDEN") || !strings.Contains(rec.Body.String(), "request_id") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if access.calls != 1 || runtime.resolveCalls != 0 || runtime.retryCalls != 0 {
		t.Fatalf("access=%d resolve=%d retry=%d", access.calls, runtime.resolveCalls, runtime.retryCalls)
	}
}

func TestRuntimeRetryMissingProjectPolicyFailsBeforeResolvingBookRun(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
	runtime := &fakeRuntimeService{projectID: 7}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "AUTH_POLICY_UNAVAILABLE") || !strings.Contains(rec.Body.String(), "request_id") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if runtime.resolveCalls != 0 || runtime.retryCalls != 0 {
		t.Fatalf("resolve=%d retry=%d", runtime.resolveCalls, runtime.retryCalls)
	}
}

func TestRuntimeRetryRejectsBookRunWhoseResolvedProjectDoesNotMatchURL(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: true}
	runtime := &fakeRuntimeService{projectID: 99}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: &fakeBatchProjectLifecycle{}}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "RUNTIME_NOT_FOUND") || !strings.Contains(rec.Body.String(), "request_id") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if access.calls != 1 || runtime.resolveCalls != 1 || runtime.retryCalls != 0 {
		t.Fatalf("access=%d resolve=%d retry=%d", access.calls, runtime.resolveCalls, runtime.retryCalls)
	}
}

func TestRuntimeRetryDoesNotRevealWhetherBookRunIsMissingOrBelongsToAnotherProject(t *testing.T) {
	serve := func(runtime *fakeRuntimeService) *httptest.ResponseRecorder {
		auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
		access := &fakeRuntimeAccess{allowed: true}
		h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access}, runtime)
		req := runtimeRetryRequest()
		req.Header.Set(testRequestIDHeader, "runtime-oracle-test")
		req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if access.calls != 1 || runtime.resolveCalls != 1 || runtime.retryCalls != 0 {
			t.Fatalf("access=%d resolve=%d retry=%d", access.calls, runtime.resolveCalls, runtime.retryCalls)
		}
		return rec
	}

	missing := serve(&fakeRuntimeService{resolveErr: sql.ErrNoRows})
	mismatch := serve(&fakeRuntimeService{projectID: 99})
	if missing.Code != http.StatusNotFound || mismatch.Code != http.StatusNotFound {
		t.Fatalf("missing=%d mismatch=%d", missing.Code, mismatch.Code)
	}
	if missing.Body.String() != mismatch.Body.String() {
		t.Fatalf("resource oracle differs:\nmissing=%s\nmismatch=%s", missing.Body.String(), mismatch.Body.String())
	}
	if !strings.Contains(missing.Body.String(), `"code":"RUNTIME_NOT_FOUND"`) || !strings.Contains(missing.Body.String(), `"message":"BookRun 不存在"`) {
		t.Fatalf("unexpected envelope: %s", missing.Body.String())
	}
}

func TestRuntimeRetryResolverFailureIsSafeAndDoesNotMutate(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: true}
	runtime := &fakeRuntimeService{resolveErr: errors.New("mysql password=secret")}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "RUNTIME_UNAVAILABLE") || !strings.Contains(rec.Body.String(), "request_id") || strings.Contains(rec.Body.String(), "password=secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if access.calls != 1 || runtime.resolveCalls != 1 || runtime.retryCalls != 0 {
		t.Fatalf("access=%d resolve=%d retry=%d", access.calls, runtime.resolveCalls, runtime.retryCalls)
	}
}

func (f *fakeRuntimeService) RetryBookRun(context.Context, int64) (task9runtime.WorkItem, bool, error) {
	f.retryCalls++
	if f.retryErr != nil {
		return f.item, f.created, f.retryErr
	}
	return f.item, f.created, f.err
}

type fakeRuntimeAccess struct {
	allowed bool
	calls   int
}

func (f *fakeRuntimeAccess) CanAccessBatchProject(context.Context, int64, int64, int64, bool) (bool, error) {
	f.calls++
	return f.allowed, nil
}

func runtimeRetryRequest() *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/7/book-runs/41/retry", nil)
	sameOrigin(req)
	return req
}

func TestRuntimeRetryRequiresAuthentication(t *testing.T) {
	auth := &fakeAuthService{authErr: authn.ErrUnauthenticated}
	runtime := &fakeRuntimeService{projectID: 7, item: task9runtime.WorkItem{BookRunID: 42, BookID: 9, Attempt: 2}, created: true}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth}, runtime)
	req := runtimeRetryRequest()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if runtime.retryCalls != 0 {
		t.Fatalf("retry calls=%d want 0", runtime.retryCalls)
	}
}

func TestRuntimeRetryRequiresBatchExecuteCapability(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	runtime := &fakeRuntimeService{projectID: 7, item: task9runtime.WorkItem{BookRunID: 42, BookID: 9, Attempt: 2}, created: true}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "AUTH_FORBIDDEN") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if runtime.retryCalls != 0 {
		t.Fatalf("retry calls=%d want 0", runtime.retryCalls)
	}
}

func TestRuntimeRetryRequiresProjectOwnership(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: false}
	runtime := &fakeRuntimeService{projectID: 7, item: task9runtime.WorkItem{BookRunID: 42, BookID: 9, Attempt: 2}, created: true}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "AUTH_FORBIDDEN") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if access.calls != 1 || runtime.retryCalls != 0 {
		t.Fatalf("access=%d retry=%d", access.calls, runtime.retryCalls)
	}
}

func TestRuntimeRetryUsesExistingTask15BoundariesAndReturnsAttempt(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", TeamID: 2, Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: true}
	runtime := &fakeRuntimeService{projectID: 7, item: task9runtime.WorkItem{BookRunID: 42, BookID: 9, Attempt: 2}, created: true}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: &fakeBatchProjectLifecycle{}}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if runtime.retryCalls != 1 || !strings.Contains(rec.Body.String(), `"attempt":2`) {
		t.Fatalf("calls=%d body=%s", runtime.retryCalls, rec.Body.String())
	}
}

func TestRuntimeRetryRejectsBadOriginBeforeMutation(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: true}
	runtime := &fakeRuntimeService{projectID: 7, item: task9runtime.WorkItem{BookRunID: 42, BookID: 9, Attempt: 2}, created: true}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access}, runtime)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/7/book-runs/41/retry", nil)
	req.Host = "app.example"
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "CSRF_REJECTED") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if runtime.retryCalls != 0 {
		t.Fatalf("retry calls=%d", runtime.retryCalls)
	}
}

func TestRuntimeRetryMapsNonRetryableWithoutLeakingInternalError(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 5, Role: "admin", Capabilities: authn.EffectiveCapabilities("admin", []string{CapabilityBatchExecute})}}
	access := &fakeRuntimeAccess{allowed: true}
	runtime := &fakeRuntimeService{projectID: 7, retryErr: task9runtime.ErrBookRunNotRetryable}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: &fakeBatchProjectLifecycle{}}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "secret") {
		t.Fatal(errors.New("sensitive error leaked"))
	}
}

func TestRuntimeRetryRejectsArchivedProjectBeforeCreatingAttempt(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	access := &fakeRuntimeAccess{allowed: true}
	lifecycle := &fakeBatchProjectLifecycle{archived: true}
	runtime := &fakeRuntimeService{projectID: 7}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: lifecycle}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_ARCHIVED"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if runtime.retryCalls != 0 || lifecycle.stateCalls != 1 {
		t.Fatalf("retry=%d lifecycle=%+v", runtime.retryCalls, lifecycle)
	}
}
