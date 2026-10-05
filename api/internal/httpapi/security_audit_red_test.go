package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

// CanAccessBatchProject is intentionally declared on the existing publishing
// test double before the production HTTP dependency requires it. The security
// audit RED tests below prove that BatchProject browser routes currently ignore
// the Task 15 ownership boundary even though the publishing layer already owns
// the project-ownership fact source.
func (f *fakePublishingService) CanAccessBatchProject(context.Context, authn.User, int64) (bool, error) {
	return false, nil
}

type failingUnifiedSettings struct{ err error }

func (f *failingUnifiedSettings) GetCurrent(context.Context, int64) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, f.err
}
func (f *failingUnifiedSettings) SaveProduction(context.Context, int64, map[string]any) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, f.err
}
func (f *failingUnifiedSettings) SavePublishing(context.Context, int64, map[string]any) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, f.err
}
func (f *failingUnifiedSettings) SaveProfile(context.Context, int64, unifiedsettings.VersionProfile) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, f.err
}
func (f *failingUnifiedSettings) Sync121(context.Context, int64) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, f.err
}
func (f *failingUnifiedSettings) SyncStyleTypes(context.Context, int64) (unifiedsettings.Current, error) {
	return unifiedsettings.Current{}, f.err
}

func authenticatedBatchRequest(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	return req
}

func TestSecurityAuditBatchProjectDetailBlocksCrossUserCrossTeamIDOR(t *testing.T) {
	reader := &fakeBatchProjectDetailReader{
		project: intake.BatchProject{ID: 51, IntakeID: 11, Name: "user-a-private-project"},
		books: []intake.Book{{ID: 31, IntakeID: 11, ExternalBookID: "1001", Title: "private-book"}},
	}
	auth := &fakeAuthService{user: authn.User{
		ID:           202,
		TeamID:       22,
		Role:         "member",
		Capabilities: []string{CapabilityBatchView},
	}}
	handler := NewHandler(Dependencies{
		Auth:                auth,
		Publishing:          &fakePublishingService{},
		BatchProjectDetails: reader,
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, authenticatedBatchRequest(http.MethodGet, "/api/v1/batch-projects/51"))

	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s, want 403 or 404 for cross-user/cross-team project access", rec.Code, rec.Body.String())
	}
	body := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"user-a-private-project", "private-book", `"id":51`} {
		if strings.Contains(body, strings.ToLower(forbidden)) {
			t.Fatalf("cross-user project material leaked: %s", rec.Body.String())
		}
	}
}

func TestSecurityAuditBatchProjectListHidesProjectsOutsideOwnershipBoundary(t *testing.T) {
	reader := &fakeBatchProjectReader{projects: []intake.BatchProject{
		{ID: 51, IntakeID: 11, Name: "user-a-private-project"},
	}}
	auth := &fakeAuthService{user: authn.User{
		ID:           202,
		TeamID:       22,
		Role:         "member",
		Capabilities: []string{CapabilityBatchView},
	}}
	handler := NewHandler(Dependencies{
		Auth:          auth,
		Publishing:    &fakePublishingService{},
		BatchProjects: reader,
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, authenticatedBatchRequest(http.MethodGet, "/api/v1/batch-projects"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s, want 200 with only visible projects", rec.Code, rec.Body.String())
	}
	body := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"user-a-private-project", `"id":51`} {
		if strings.Contains(body, strings.ToLower(forbidden)) {
			t.Fatalf("project outside ownership boundary leaked in list: %s", rec.Body.String())
		}
	}
}

func TestSecurityAuditUnifiedSettingsBlocksCrossUserProjectRead(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 202, TeamID: 22, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	handler := NewHandler(Dependencies{
		Auth:            auth,
		Publishing:      &fakePublishingService{},
		UnifiedSettings: &fakeUnifiedSettings{},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, authenticatedBatchRequest(http.MethodGet, "/api/v1/batch-projects/51/settings"))
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s, want 403 or 404 for cross-user settings read", rec.Code, rec.Body.String())
	}
}

func TestSecurityAuditGenerationBlocksCrossUserProjectRead(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 202, TeamID: 22, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	handler := NewHandler(Dependencies{
		Auth:       auth,
		Publishing: &fakePublishingService{},
		Generation: &fakeGenerationService{},
	})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, authenticatedBatchRequest(http.MethodGet, "/api/v1/batch-projects/51/generation"))
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s, want 403 or 404 for cross-user generation read", rec.Code, rec.Body.String())
	}
}

func TestSecurityAuditIntakeServiceErrorIsRedacted(t *testing.T) {
	const sensitiveDetail = "mysql dsn contains audit-fixture-secret"
	service := &fakeIntakeAPI{createErr: errors.New(sensitiveDetail)}
	handler := NewHandler(Dependencies{Intakes: service})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", bytes.NewBufferString(`{"name":"test","groups":[{"source":"x","platformId":"1","books":[{"bookId":"1"}]}]}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code < 400 {
		t.Fatalf("status=%d body=%s, want error status", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), sensitiveDetail) || strings.Contains(strings.ToLower(rec.Body.String()), "dsn") {
		t.Fatalf("internal credential/database detail leaked to browser: %s", rec.Body.String())
	}
}

func TestSecurityAuditUnifiedSettingsServiceErrorIsRedacted(t *testing.T) {
	service := &failingUnifiedSettings{err: errors.New("redis password=audit-fixture-secret")}
	handler := NewHandler(Dependencies{UnifiedSettings: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/51/settings", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s, want 500", rec.Code, rec.Body.String())
	}
	body := strings.ToLower(rec.Body.String())
	if strings.Contains(body, "audit-fixture-secret") || strings.Contains(body, "redis password") {
		t.Fatalf("internal settings error leaked to browser: %s", rec.Body.String())
	}
}
