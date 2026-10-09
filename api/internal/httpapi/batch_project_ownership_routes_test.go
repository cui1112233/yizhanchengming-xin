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
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

type batchProjectAccessSpy struct {
	allowed   bool
	err       error
	calls     int
	projectID int64
	userID    int64
	teamID    int64
	elevated  bool
}

func (s *batchProjectAccessSpy) CanAccessBatchProject(_ context.Context, projectID, userID, teamID int64, elevated bool) (bool, error) {
	s.calls++
	s.projectID, s.userID, s.teamID, s.elevated = projectID, userID, teamID, elevated
	return s.allowed, s.err
}

type batchProjectBusinessSpy struct{ calls int }

func (s *batchProjectBusinessSpy) called() { s.calls++ }

func (s *batchProjectBusinessSpy) GetBatchProject(context.Context, int64) (intake.BatchProject, error) {
	s.called()
	return intake.BatchProject{ID: 51, IntakeID: 11}, nil
}
func (s *batchProjectBusinessSpy) ListBooks(context.Context, int64) ([]intake.Book, error) {
	s.called()
	return nil, nil
}
func (s *batchProjectBusinessSpy) UpdateBookOriginalText(context.Context, int64, int64, string) (intake.Book, error) {
	s.called()
	return intake.Book{}, nil
}
func (s *batchProjectBusinessSpy) Storyboard(context.Context, int64, int64) (generation.StoryboardDocument, error) {
	s.called()
	return generation.StoryboardDocument{}, nil
}
func (s *batchProjectBusinessSpy) SaveStoryboardCard(context.Context, int64, int64, generation.SaveStoryboardCardRequest) (generation.StoryboardDocument, error) {
	s.called()
	return generation.StoryboardDocument{}, nil
}
func (s *batchProjectBusinessSpy) DeleteStoryboardCard(context.Context, int64, int64, int64, int) (generation.StoryboardDocument, error) {
	s.called()
	return generation.StoryboardDocument{}, nil
}
func (s *batchProjectBusinessSpy) ReorderStoryboard(context.Context, int64, int64, []int64, int) (generation.StoryboardDocument, error) {
	s.called()
	return generation.StoryboardDocument{}, nil
}
func (s *batchProjectBusinessSpy) RecompileStoryboard(context.Context, int64, int64, string) (generation.BookGenerationResult, error) {
	s.called()
	return generation.BookGenerationResult{}, nil
}
func (s *batchProjectBusinessSpy) GetCurrent(context.Context, int64) (unifiedsettings.Current, error) {
	s.called()
	return unifiedsettings.Current{}, nil
}
func (s *batchProjectBusinessSpy) SaveProduction(context.Context, int64, map[string]any) (unifiedsettings.Current, error) {
	s.called()
	return unifiedsettings.Current{}, nil
}
func (s *batchProjectBusinessSpy) SavePublishing(context.Context, int64, map[string]any) (unifiedsettings.Current, error) {
	s.called()
	return unifiedsettings.Current{}, nil
}
func (s *batchProjectBusinessSpy) SaveProfile(context.Context, int64, unifiedsettings.VersionProfile) (unifiedsettings.Current, error) {
	s.called()
	return unifiedsettings.Current{}, nil
}
func (s *batchProjectBusinessSpy) Sync121(context.Context, int64) (unifiedsettings.Current, error) {
	s.called()
	return unifiedsettings.Current{}, nil
}
func (s *batchProjectBusinessSpy) SyncStyleTypes(context.Context, int64) (unifiedsettings.Current, error) {
	s.called()
	return unifiedsettings.Current{}, nil
}
func (s *batchProjectBusinessSpy) ProjectSummary(context.Context, int64) (generation.ProjectSummary, error) {
	s.called()
	return generation.ProjectSummary{}, nil
}
func (s *batchProjectBusinessSpy) BookSummary(context.Context, int64, int64) (generation.BookGenerationResult, error) {
	s.called()
	return generation.BookGenerationResult{}, nil
}
func (s *batchProjectBusinessSpy) RunBook(context.Context, generation.RunBookRequest) (generation.BookGenerationResult, error) {
	s.called()
	return generation.BookGenerationResult{}, nil
}
func (s *batchProjectBusinessSpy) RunBatch(context.Context, generation.RunBatchRequest) (generation.BatchGenerationResult, error) {
	s.called()
	return generation.BatchGenerationResult{}, nil
}
func (s *batchProjectBusinessSpy) RetryStage(context.Context, generation.RetryStageRequest) (generation.BookGenerationResult, error) {
	s.called()
	return generation.BookGenerationResult{}, nil
}
func (s *batchProjectBusinessSpy) StageResult(context.Context, int64, int64, generation.Stage) (generation.StageRun, error) {
	s.called()
	return generation.StageRun{}, nil
}
func (s *batchProjectBusinessSpy) ListPrompts(context.Context) ([]generation.Prompt, error) {
	s.called()
	return nil, nil
}
func (s *batchProjectBusinessSpy) AudioMeasurement(context.Context, int64, int64) (generation.AudioMeasurement, error) {
	s.called()
	return generation.AudioMeasurement{}, nil
}
func (s *batchProjectBusinessSpy) MeasureAudio(context.Context, generation.AudioMeasurementRequest) (generation.AudioMeasurement, error) {
	s.called()
	return generation.AudioMeasurement{}, nil
}

type batchProjectRouteCase struct {
	name       string
	method     string
	path       string
	body       string
	capability string
}

func newlyProtectedBatchProjectRoutes() []batchProjectRouteCase {
	return []batchProjectRouteCase{
		{name: "original text save", method: http.MethodPut, path: "/api/v1/batch-projects/51/books/21/original-text", body: `{"originalText":"secret"}`, capability: CapabilityBatchConfigure},
		{name: "storyboard read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/storyboard", capability: CapabilityBatchView},
		{name: "storyboard card create", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/storyboard/cards", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "storyboard card update", method: http.MethodPut, path: "/api/v1/batch-projects/51/books/21/storyboard/cards/31", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "storyboard card delete", method: http.MethodDelete, path: "/api/v1/batch-projects/51/books/21/storyboard/cards/31", body: `{"expectedVersion":1}`, capability: CapabilityBatchConfigure},
		{name: "storyboard reorder", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/storyboard/reorder", body: `{"cardIds":[31],"expectedVersion":1}`, capability: CapabilityBatchConfigure},
		{name: "storyboard recompile", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/storyboard/recompile", body: `{"requestId":"test"}`, capability: CapabilityBatchExecute},
		{name: "settings read", method: http.MethodGet, path: "/api/v1/batch-projects/51/settings", capability: CapabilityBatchView},
		{name: "production settings save", method: http.MethodPut, path: "/api/v1/batch-projects/51/settings/production", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "publishing settings save", method: http.MethodPut, path: "/api/v1/batch-projects/51/settings/publishing", body: `{}`, capability: CapabilityPublishConfigure},
		{name: "version profile read", method: http.MethodGet, path: "/api/v1/batch-projects/51/version-profile", capability: CapabilityBatchView},
		{name: "version profile save", method: http.MethodPut, path: "/api/v1/batch-projects/51/version-profile", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "version profile sync 121", method: http.MethodPost, path: "/api/v1/batch-projects/51/version-profile/sync-121", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "version profile sync styles", method: http.MethodPost, path: "/api/v1/batch-projects/51/version-profile/sync-style-types", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "project generation read", method: http.MethodGet, path: "/api/v1/batch-projects/51/generation", capability: CapabilityBatchView},
		{name: "generation run read", method: http.MethodGet, path: "/api/v1/batch-projects/51/generation/runs/61", capability: CapabilityBatchView},
		{name: "project generation start", method: http.MethodPost, path: "/api/v1/batch-projects/51/generation", body: `{}`, capability: CapabilityBatchExecute},
		{name: "book generation read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/generation", capability: CapabilityBatchView},
		{name: "book generation start", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/generation", body: `{}`, capability: CapabilityBatchExecute},
		{name: "audio measurement read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/audio-measurement", capability: CapabilityBatchView},
		{name: "audio measurement write", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/audio-measurement", body: `{"audioAsset":"/secret.mp3"}`, capability: CapabilityBatchExecute},
		{name: "generation stage retry", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/generation/stages/DIRECTOR/retry", body: `{}`, capability: CapabilityBatchExecute},
		{name: "generation stage read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/generation/stages/DIRECTOR", capability: CapabilityBatchView},
	}
}

func allDirectBatchProjectObjectRoutes() []batchProjectRouteCase {
	routes := []batchProjectRouteCase{
		{name: "project detail", method: http.MethodGet, path: "/api/v1/batch-projects/51", capability: CapabilityBatchView},
		{name: "novel panel read", method: http.MethodGet, path: "/api/v1/batch-projects/51/novel-panel", capability: CapabilityBatchView},
		{name: "novel panel save", method: http.MethodPut, path: "/api/v1/batch-projects/51/novel-panel", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "novel panel history", method: http.MethodGet, path: "/api/v1/batch-projects/51/novel-panel/history", capability: CapabilityBatchView},
		{name: "novel panel restore", method: http.MethodPost, path: "/api/v1/batch-projects/51/novel-panel/history/rev-1/restore", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "shuihuo segments read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/shuihuo/segments", capability: CapabilityBatchView},
		{name: "shuihuo segment create", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/shuihuo/segments", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "shuihuo segment update", method: http.MethodPut, path: "/api/v1/batch-projects/51/books/21/shuihuo/segments/31", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "shuihuo segment reorder", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/shuihuo/segments/reorder", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "shuihuo assets read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/shuihuo/assets", capability: CapabilityBatchView},
		{name: "shuihuo asset upload", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/shuihuo/assets/upload", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "shuihuo asset content", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/shuihuo/assets/41/content", capability: CapabilityBatchView},
		{name: "shuihuo media tasks read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/shuihuo/media-tasks", capability: CapabilityBatchView},
		{name: "shuihuo media task create", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/shuihuo/media-tasks", body: `{}`, capability: CapabilityBatchExecute},
		{name: "shuihuo candidates read", method: http.MethodGet, path: "/api/v1/batch-projects/51/books/21/shuihuo/media-tasks/41/candidates", capability: CapabilityBatchView},
		{name: "shuihuo candidate select", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/shuihuo/media-tasks/41/candidates/61/select", body: `{}`, capability: CapabilityBatchConfigure},
		{name: "shuihuo task retry", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/shuihuo/media-tasks/41/retry", body: `{}`, capability: CapabilityBatchExecute},
		{name: "project video status", method: http.MethodGet, path: "/api/v1/batch-projects/51/video", capability: CapabilityBatchView},
		{name: "book video start", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/video", body: `{}`, capability: CapabilityBatchExecute},
		{name: "book video merge", method: http.MethodPost, path: "/api/v1/batch-projects/51/books/21/merge", body: `{}`, capability: CapabilityBatchExecute},
	}
	return append(routes, newlyProtectedBatchProjectRoutes()...)
}

func newBatchProjectOwnershipHandler(auth AuthService, access BatchProjectAccessChecker, business *batchProjectBusinessSpy) http.Handler {
	return NewHandler(Dependencies{
		Auth:                  auth,
		BatchProjectAccess:    access,
		BatchProjectLifecycle: &fakeBatchProjectLifecycle{},
		BatchProjectDetails:   business,
		ScriptBooks:           business,
		ScriptStoryboards:     business,
		UnifiedSettings:       business,
		Generation:            business,
	})
}

func batchProjectRouteRequest(route batchProjectRouteCase) *http.Request {
	req := httptest.NewRequest(route.method, "http://example.com"+route.path, bytes.NewBufferString(route.body))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	if route.method != http.MethodGet {
		sameOrigin(req)
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestBatchProjectObjectRoutesRejectForeignUsersBeforeBusinessServices(t *testing.T) {
	for _, route := range newlyProtectedBatchProjectRoutes() {
		t.Run(route.name, func(t *testing.T) {
			business := &batchProjectBusinessSpy{}
			access := &batchProjectAccessSpy{allowed: false}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{route.capability}}}
			rec := httptest.NewRecorder()
			newBatchProjectOwnershipHandler(auth, access, business).ServeHTTP(rec, batchProjectRouteRequest(route))

			if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), `"code":"AUTH_FORBIDDEN"`) || !strings.Contains(rec.Body.String(), `"request_id"`) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if business.calls != 0 {
				t.Fatalf("business calls=%d, want zero before ownership rejection", business.calls)
			}
		})
	}
}

func TestEveryDirectBatchProjectObjectRouteUsesTheSharedOwnershipBoundary(t *testing.T) {
	for _, route := range allDirectBatchProjectObjectRoutes() {
		t.Run(route.name, func(t *testing.T) {
			access := &batchProjectAccessSpy{allowed: false}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{route.capability}}}
			rec := httptest.NewRecorder()
			newBatchProjectOwnershipHandler(auth, access, &batchProjectBusinessSpy{}).ServeHTTP(rec, batchProjectRouteRequest(route))

			if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), `"code":"AUTH_FORBIDDEN"`) || !strings.Contains(rec.Body.String(), `"request_id"`) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if access.calls != 1 || access.projectID != 51 {
				t.Fatalf("access=%+v", access)
			}
		})
	}
}

func TestBatchProjectObjectRoutesFailClosedWhenPolicyIsUnavailable(t *testing.T) {
	for _, route := range newlyProtectedBatchProjectRoutes() {
		for _, policy := range []struct {
			name   string
			access BatchProjectAccessChecker
		}{
			{name: "missing checker"},
			{name: "checker error", access: &batchProjectAccessSpy{err: errors.New("ownership database unavailable")}},
		} {
			t.Run(route.name+"/"+policy.name, func(t *testing.T) {
				business := &batchProjectBusinessSpy{}
				auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{route.capability}}}
				rec := httptest.NewRecorder()
				newBatchProjectOwnershipHandler(auth, policy.access, business).ServeHTTP(rec, batchProjectRouteRequest(route))

				if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"code":"AUTH_POLICY_UNAVAILABLE"`) || !strings.Contains(rec.Body.String(), `"request_id"`) {
					t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
				}
				if business.calls != 0 {
					t.Fatalf("business calls=%d, want zero while policy unavailable", business.calls)
				}
			})
		}
	}
}

func TestBatchProjectListFailsClosedBeforeScopedReaderWhenPolicyIsMissing(t *testing.T) {
	reader := &fakeBatchProjectReader{projects: []intake.BatchProject{{ID: 51, Name: "foreign-secret-project"}}}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/batch-projects", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	rec := httptest.NewRecorder()
	NewHandler(Dependencies{Auth: auth, BatchProjects: reader}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"code":"AUTH_POLICY_UNAVAILABLE"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if reader.calls != 0 || strings.Contains(rec.Body.String(), "foreign-secret-project") {
		t.Fatalf("reader crossed missing-policy boundary: calls=%d body=%s", reader.calls, rec.Body.String())
	}
}

func TestBatchProjectAccessElevatesOnlyAdminAndOwner(t *testing.T) {
	for _, tc := range []struct {
		role         string
		allowed      bool
		wantStatus   int
		wantElevated bool
	}{
		{role: "member", allowed: true, wantStatus: http.StatusOK},
		{role: "owner", allowed: true, wantStatus: http.StatusOK, wantElevated: true},
		{role: "admin", allowed: true, wantStatus: http.StatusOK, wantElevated: true},
		{role: "dev", allowed: false, wantStatus: http.StatusForbidden},
		{role: "manager", allowed: false, wantStatus: http.StatusForbidden},
	} {
		t.Run(tc.role, func(t *testing.T) {
			business := &batchProjectBusinessSpy{}
			access := &batchProjectAccessSpy{allowed: tc.allowed}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: tc.role, Capabilities: []string{CapabilityBatchView}}}
			rec := httptest.NewRecorder()
			route := batchProjectRouteCase{method: http.MethodGet, path: "/api/v1/batch-projects/51/generation", capability: CapabilityBatchView}
			newBatchProjectOwnershipHandler(auth, access, business).ServeHTTP(rec, batchProjectRouteRequest(route))

			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
			if access.calls != 1 || access.projectID != 51 || access.userID != 7 || access.teamID != 3 || access.elevated != tc.wantElevated {
				t.Fatalf("access=%+v", access)
			}
			wantCalls := 0
			if tc.allowed {
				wantCalls = 1
			}
			if business.calls != wantCalls {
				t.Fatalf("business calls=%d, want %d", business.calls, wantCalls)
			}
		})
	}
}

func TestBatchProjectOwnershipCannotBeProbedWithoutRouteCapability(t *testing.T) {
	business := &batchProjectBusinessSpy{}
	access := &batchProjectAccessSpy{allowed: true}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member"}}
	rec := httptest.NewRecorder()
	route := batchProjectRouteCase{method: http.MethodGet, path: "/api/v1/batch-projects/51/generation", capability: CapabilityBatchView}
	newBatchProjectOwnershipHandler(auth, access, business).ServeHTTP(rec, batchProjectRouteRequest(route))

	if rec.Code != http.StatusForbidden || access.calls != 0 || business.calls != 0 {
		t.Fatalf("status=%d access=%d business=%d body=%s", rec.Code, access.calls, business.calls, rec.Body.String())
	}
}

func TestBatchProjectMutationStillRequiresSameOriginBeforeOwnershipLookup(t *testing.T) {
	business := &batchProjectBusinessSpy{}
	access := &batchProjectAccessSpy{allowed: true}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchExecute}}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/v1/batch-projects/51/generation", bytes.NewBufferString(`{}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access-token"})
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	newBatchProjectOwnershipHandler(auth, access, business).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden || access.calls != 0 || business.calls != 0 {
		t.Fatalf("status=%d access=%d business=%d body=%s", rec.Code, access.calls, business.calls, rec.Body.String())
	}
}
