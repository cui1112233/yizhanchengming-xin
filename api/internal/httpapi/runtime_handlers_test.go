package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type fakeRuntimeService struct {
	projectID  int64
	item       task9runtime.WorkItem
	created    bool
	err        error
	retryCalls int
}

func (f *fakeRuntimeService) ProjectIDForBookRun(context.Context, int64) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.projectID, nil
}
func (f *fakeRuntimeService) RetryBookRun(context.Context, int64) (task9runtime.WorkItem, bool, error) {
	f.retryCalls++
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
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access}, runtime)
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
	runtime := &fakeRuntimeService{projectID: 7, err: task9runtime.ErrBookRunNotRetryable}
	h := NewHandlerWithRuntime(Dependencies{Auth: auth, BatchProjectAccess: access}, runtime)
	req := runtimeRetryRequest()
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable && rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(strings.ToLower(rec.Body.String()), "secret") {
		t.Fatal(errors.New("sensitive error leaked"))
	}
}
