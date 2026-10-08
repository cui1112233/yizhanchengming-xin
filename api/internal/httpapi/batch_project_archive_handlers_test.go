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
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type fakeBatchProjectLifecycle struct {
	archiveErr       error
	restoreErr       error
	archived         bool
	archivedErr      error
	archiveCalls     int
	restoreCalls     int
	stateCalls       int
	intakeStateCalls int
	projectID        int64
	actorID          int64
}

func (f *fakeBatchProjectLifecycle) ArchiveBatchProject(_ context.Context, projectID, actorID int64) error {
	f.archiveCalls++
	f.projectID, f.actorID = projectID, actorID
	return f.archiveErr
}
func (f *fakeBatchProjectLifecycle) RestoreBatchProject(_ context.Context, projectID int64) error {
	f.restoreCalls++
	f.projectID = projectID
	return f.restoreErr
}
func (f *fakeBatchProjectLifecycle) IsBatchProjectArchived(_ context.Context, projectID int64) (bool, error) {
	f.stateCalls++
	f.projectID = projectID
	return f.archived, f.archivedErr
}
func (f *fakeBatchProjectLifecycle) IsIntakeBatchProjectArchived(_ context.Context, intakeID int64) (bool, error) {
	f.intakeStateCalls++
	f.projectID = intakeID
	return f.archived, f.archivedErr
}

func TestArchiveAndRestoreBatchProjectRequireCSRFConfigureAndOwnership(t *testing.T) {
	for _, tc := range []struct {
		name      string
		path      string
		configure bool
		origin    bool
		allowed   bool
		want      int
	}{
		{name: "missing csrf", path: "/api/v1/batch-projects/51/archive", configure: true, allowed: true, want: http.StatusForbidden},
		{name: "missing capability", path: "/api/v1/batch-projects/51/archive", origin: true, allowed: true, want: http.StatusForbidden},
		{name: "foreign", path: "/api/v1/batch-projects/51/archive", configure: true, origin: true, want: http.StatusForbidden},
		{name: "archive allowed", path: "/api/v1/batch-projects/51/archive", configure: true, origin: true, allowed: true, want: http.StatusOK},
		{name: "restore allowed", path: "/api/v1/batch-projects/51/restore", configure: true, origin: true, allowed: true, want: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caps := []string{}
			if tc.configure {
				caps = []string{CapabilityBatchConfigure}
			}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: caps}}
			access := &batchProjectAccessSpy{allowed: tc.allowed}
			lifecycle := &fakeBatchProjectLifecycle{}
			req := httptest.NewRequest(http.MethodPost, "http://app.example"+tc.path, nil)
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
			if tc.origin {
				sameOrigin(req)
			}
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{Auth: auth, BatchProjectAccess: access, BatchProjectLifecycle: lifecycle}).ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d body=%s want=%d", rec.Code, rec.Body.String(), tc.want)
			}
			wantCalls := 0
			if tc.want == http.StatusOK {
				wantCalls = 1
			}
			if lifecycle.archiveCalls+lifecycle.restoreCalls != wantCalls {
				t.Fatalf("archive=%d restore=%d", lifecycle.archiveCalls, lifecycle.restoreCalls)
			}
			if wantCalls == 1 && (lifecycle.projectID != 51 || (strings.HasSuffix(tc.path, "/archive") && lifecycle.actorID != 7)) {
				t.Fatalf("lifecycle=%+v", lifecycle)
			}
		})
	}
}

func TestArchiveBatchProjectReturnsConflictForActiveWorkWithoutDeletingAnything(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchConfigure}}}
	lifecycle := &fakeBatchProjectLifecycle{archiveErr: intake.ErrBatchProjectActive}
	req := httptest.NewRequest(http.MethodPost, "http://app.example/api/v1/batch-projects/51/archive", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	sameOrigin(req)
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Auth: auth, BatchProjectAccess: &batchProjectAccessSpy{allowed: true}, BatchProjectLifecycle: lifecycle}).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_ACTIVE"`) || lifecycle.archiveCalls != 1 {
		t.Fatalf("status=%d calls=%d body=%s", rec.Code, lifecycle.archiveCalls, rec.Body.String())
	}
}

func TestArchiveBatchProjectErrorsAreSafeAndMissingLifecycleFailsClosed(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchConfigure}}}
	for _, tc := range []struct {
		name      string
		lifecycle BatchProjectLifecycle
		want      int
		code      string
	}{
		{name: "missing", want: http.StatusServiceUnavailable, code: "BATCH_PROJECT_POLICY_UNAVAILABLE"},
		{name: "not found", lifecycle: &fakeBatchProjectLifecycle{archiveErr: sql.ErrNoRows}, want: http.StatusNotFound, code: "BATCH_PROJECT_NOT_FOUND"},
		{name: "internal", lifecycle: &fakeBatchProjectLifecycle{archiveErr: errors.New("mysql password=secret")}, want: http.StatusInternalServerError, code: "BATCH_PROJECT_ARCHIVE_FAILED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://app.example/api/v1/batch-projects/51/archive", nil)
			req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
			sameOrigin(req)
			rec := httptest.NewRecorder()
			NewHandler(Dependencies{Auth: auth, BatchProjectAccess: &batchProjectAccessSpy{allowed: true}, BatchProjectLifecycle: tc.lifecycle}).ServeHTTP(rec, req)
			if rec.Code != tc.want || !strings.Contains(rec.Body.String(), `"code":"`+tc.code+`"`) || strings.Contains(rec.Body.String(), "password=secret") {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
