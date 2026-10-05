package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

// CanAccessBatchProject is intentionally declared on the existing publishing
// test double before the production HTTP dependency requires it. The security
// audit RED tests below prove that BatchProject browser routes currently ignore
// the Task 15 ownership boundary even though the publishing layer already owns
// the project-ownership fact source.
func (f *fakePublishingService) CanAccessBatchProject(context.Context, authn.User, int64) (bool, error) {
	return false, nil
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
