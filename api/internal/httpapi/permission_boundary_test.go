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

type permissionBatchReader struct {
	calls int
}

func (f *permissionBatchReader) ListBatchProjects(context.Context) ([]intake.BatchProject, error) {
	f.calls++
	return []intake.BatchProject{{ID: 51, IntakeID: 11, Name: "项目"}}, nil
}

func TestBatchListReturns403ForAuthenticatedUserWithoutCapability(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Username: "alice", Role: "member"}}
	reader := &permissionBatchReader{}
	handler := NewHandler(Dependencies{Auth: auth, BatchProjects: reader})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid-access"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want 403", rec.Code, rec.Body.String())
	}
	if reader.calls != 0 {
		t.Fatalf("reader calls = %d, want 0 when permission is denied", reader.calls)
	}
	if !strings.Contains(rec.Body.String(), `"code":"AUTH_FORBIDDEN"`) {
		t.Fatalf("body = %s, want AUTH_FORBIDDEN", rec.Body.String())
	}
}

func TestSystemGenerationPromptReadDoesNotBecomeAuthWhitelistHack(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Username: "alice", Role: "member"}}
	handler := NewHandler(Dependencies{Auth: auth})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/generation/prompts", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s, want generation service 503 instead of auth 401", rec.Code, rec.Body.String())
	}
}
