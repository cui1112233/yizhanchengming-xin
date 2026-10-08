package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

type fakeAdminPromptService struct {
	prompts       []generation.AdminPrompt
	createCalls   int
	updateCalls   int
	publishCalls  int
	restoreCalls  int
	lastRequestID string
}

func (f *fakeAdminPromptService) ListAdminPrompts(context.Context, string) ([]generation.AdminPrompt, error) {
	return append([]generation.AdminPrompt(nil), f.prompts...), nil
}

func (f *fakeAdminPromptService) GetAdminPrompt(_ context.Context, key string, version int) (generation.AdminPrompt, error) {
	for _, prompt := range f.prompts {
		if prompt.Key == key && prompt.Version == version {
			return prompt, nil
		}
	}
	return generation.AdminPrompt{}, generation.ErrNotFound
}

func (f *fakeAdminPromptService) CreateAdminPromptDraft(ctx context.Context, _ int64, key, content string) error {
	f.createCalls++
	f.lastRequestID = observability.RequestID(ctx)
	f.prompts = append(f.prompts, generation.AdminPrompt{ID: 22, Key: key, Version: 4, Content: content, Lifecycle: "draft", Enabled: false})
	return nil
}

func (f *fakeAdminPromptService) UpdateAdminPromptDraft(ctx context.Context, _ int64, key string, version int, content string) error {
	f.updateCalls++
	f.lastRequestID = observability.RequestID(ctx)
	for index := range f.prompts {
		if f.prompts[index].Key == key && f.prompts[index].Version == version {
			f.prompts[index].Content = content
		}
	}
	return nil
}

func (f *fakeAdminPromptService) PublishAdminPrompt(ctx context.Context, _ int64, key string, version int) error {
	f.publishCalls++
	f.lastRequestID = observability.RequestID(ctx)
	for index := range f.prompts {
		if f.prompts[index].Key == key {
			f.prompts[index].Lifecycle = "archived"
			f.prompts[index].Enabled = false
		}
		if f.prompts[index].Key == key && f.prompts[index].Version == version {
			f.prompts[index].Lifecycle = "published"
			f.prompts[index].Enabled = true
		}
	}
	return nil
}

func (f *fakeAdminPromptService) RestoreAdminPrompt(ctx context.Context, _ int64, key string, version int) error {
	f.restoreCalls++
	f.lastRequestID = observability.RequestID(ctx)
	for _, prompt := range f.prompts {
		if prompt.Key == key && prompt.Version == version {
			f.prompts = append(f.prompts, generation.AdminPrompt{ID: 30, Key: key, Version: 5, Content: prompt.Content, Lifecycle: "published", Enabled: true})
			return nil
		}
	}
	return generation.ErrNotFound
}

func adminPromptFixture() *fakeAdminPromptService {
	return &fakeAdminPromptService{prompts: []generation.AdminPrompt{
		{ID: 11, Key: generation.PromptScript, Version: 3, Content: "server-only-prompt-body", Lifecycle: "published", Enabled: true, ContentSHA256: "sha"},
		{ID: 12, Key: generation.PromptScript, Version: 4, Content: "draft-body", Lifecycle: "draft", Enabled: false, ContentSHA256: "sha2"},
	}}
}

func adminPromptHandlerRequest(method, path, body string, user authn.User) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	return req
}

func TestAdminPromptAPIPermissionsUseEffectiveCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name   string
		user   authn.User
		method string
		path   string
		want   int
	}{
		{name: "owner", user: authn.User{ID: 1, Role: "owner", Capabilities: authn.EffectiveCapabilities("owner", nil)}, method: http.MethodGet, path: "/api/v1/admin/prompts", want: http.StatusOK},
		{name: "dev", user: authn.User{ID: 2, Role: "dev", Capabilities: authn.EffectiveCapabilities("dev", nil)}, method: http.MethodGet, path: "/api/v1/admin/prompts", want: http.StatusOK},
		{name: "legacy admin", user: authn.User{ID: 3, Role: "admin", Capabilities: authn.EffectiveCapabilities("admin", nil)}, method: http.MethodGet, path: "/api/v1/admin/prompts", want: http.StatusOK},
		{name: "member denied", user: authn.User{ID: 4, Role: "member"}, method: http.MethodGet, path: "/api/v1/admin/prompts", want: http.StatusForbidden},
		{name: "manager view only", user: authn.User{ID: 5, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptView}}, method: http.MethodGet, path: "/api/v1/admin/prompts", want: http.StatusOK},
		{name: "manager cannot edit", user: authn.User{ID: 6, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptView}}, method: http.MethodPost, path: "/api/v1/admin/prompts/script.default/drafts", want: http.StatusForbidden},
		{name: "manager cannot publish", user: authn.User{ID: 7, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptView}}, method: http.MethodPost, path: "/api/v1/admin/prompts/script.default/versions/3/publish", want: http.StatusForbidden},
		{name: "manager cannot restore", user: authn.User{ID: 8, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptView}}, method: http.MethodPost, path: "/api/v1/admin/prompts/script.default/versions/3/restore", want: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := adminPromptFixture()
			h := NewHandler(Dependencies{Auth: &fakeAuthService{user: tc.user}, AdminPrompts: service})
			req := adminPromptHandlerRequest(tc.method, tc.path, "", tc.user)
			req.Body = http.NoBody
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestAdminPromptDetailOnlyViewCapabilityReturnsBody(t *testing.T) {
	service := adminPromptFixture()
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptView}}}, AdminPrompts: service})
	req := adminPromptHandlerRequest(http.MethodGet, "/api/v1/admin/prompts/script.default/versions/3", "", authn.User{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "server-only-prompt-body") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	denied := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 8, Role: "manager"}}, AdminPrompts: service})
	rec = httptest.NewRecorder()
	denied.ServeHTTP(rec, adminPromptHandlerRequest(http.MethodGet, "/api/v1/admin/prompts/script.default/versions/3", "", authn.User{}))
	if rec.Code != http.StatusForbidden || strings.Contains(rec.Body.String(), "server-only-prompt-body") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminCapabilitiesReturnsOnlyEffectiveAdminSet(t *testing.T) {
	manager := authn.User{ID: 7, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptView, "batch.view"}}
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: manager}, AdminPrompts: adminPromptFixture()})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, adminPromptHandlerRequest(http.MethodGet, "/api/v1/admin/capabilities", "", manager))
	if rec.Code != http.StatusOK || rec.Body.String() != `{"capabilities":["admin.prompt.view"]}
` {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	member := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 8, Role: "member"}}, AdminPrompts: adminPromptFixture()})
	rec = httptest.NewRecorder()
	member.ServeHTTP(rec, adminPromptHandlerRequest(http.MethodGet, "/api/v1/admin/capabilities", "", authn.User{}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("member status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestManagerSingleCapabilityCanUseOnlyGrantedPromptOperation(t *testing.T) {
	editService := adminPromptFixture()
	editUser := authn.User{ID: 9, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptEdit}}
	editHandler := NewHandler(Dependencies{Auth: &fakeAuthService{user: editUser}, AdminPrompts: editService})
	rec := httptest.NewRecorder()
	editHandler.ServeHTTP(rec, adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", `{"content":"manager-draft"}`, editUser))
	if rec.Code != http.StatusCreated || editService.createCalls != 1 {
		t.Fatalf("manager edit status=%d calls=%d body=%s", rec.Code, editService.createCalls, rec.Body.String())
	}

	publishService := adminPromptFixture()
	publishUser := authn.User{ID: 10, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptPublish}}
	publishHandler := NewHandler(Dependencies{Auth: &fakeAuthService{user: publishUser}, AdminPrompts: publishService})
	rec = httptest.NewRecorder()
	publishHandler.ServeHTTP(rec, adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/versions/4/publish", "", publishUser))
	if rec.Code != http.StatusOK || publishService.publishCalls != 1 {
		t.Fatalf("manager publish status=%d calls=%d body=%s", rec.Code, publishService.publishCalls, rec.Body.String())
	}
}

func TestAdminPromptListNeverReturnsBodyAndPublicPromptRemainsSafe(t *testing.T) {
	service := adminPromptFixture()
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Capabilities: authn.EffectiveCapabilities("owner", nil)}}, AdminPrompts: service})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, adminPromptHandlerRequest(http.MethodGet, "/api/v1/admin/prompts", "", authn.User{}))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "server-only-prompt-body") || strings.Contains(rec.Body.String(), "draft-body") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	public := NewHandler(Dependencies{Generation: &fakeGenerationService{prompts: []generation.Prompt{{Key: generation.PromptScript, Version: 3, Enabled: true, Content: "server-only-prompt-body"}}}})
	rec = httptest.NewRecorder()
	public.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/generation/prompts", nil))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "server-only-prompt-body") {
		t.Fatalf("public prompt leaked body: status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminPromptWritesRequireCookieSessionAndSameOriginCSRF(t *testing.T) {
	service := adminPromptFixture()
	user := authn.User{ID: 7, Role: "manager", Capabilities: []string{authn.CapabilityAdminPromptEdit}}
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: user}, AdminPrompts: service})

	missingSession := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", strings.NewReader(`{"content":"new"}`))
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
	h.ServeHTTP(missingSession, req)
	if missingSession.Code != http.StatusUnauthorized {
		t.Fatalf("missing session status=%d body=%s", missingSession.Code, missingSession.Body.String())
	}

	missingOrigin := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", strings.NewReader(`{"content":"new"}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	h.ServeHTTP(missingOrigin, req)
	if missingOrigin.Code != http.StatusForbidden || service.createCalls != 0 {
		t.Fatalf("missing origin status=%d calls=%d body=%s", missingOrigin.Code, service.createCalls, missingOrigin.Body.String())
	}

	crossOrigin := httptest.NewRecorder()
	req = adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", `{"content":"new"}`, user)
	req.Header.Set("Origin", "https://evil.example")
	h.ServeHTTP(crossOrigin, req)
	if crossOrigin.Code != http.StatusForbidden || service.createCalls != 0 {
		t.Fatalf("cross origin status=%d calls=%d body=%s", crossOrigin.Code, service.createCalls, crossOrigin.Body.String())
	}

	ok := httptest.NewRecorder()
	h.ServeHTTP(ok, adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", `{"content":"new"}`, user))
	if ok.Code != http.StatusCreated || service.createCalls != 1 {
		t.Fatalf("same origin status=%d calls=%d body=%s", ok.Code, service.createCalls, ok.Body.String())
	}
	if service.lastRequestID == "" || service.lastRequestID == "client-forged" {
		t.Fatalf("request id was not server-controlled: %q", service.lastRequestID)
	}
	forged := httptest.NewRecorder()
	h.ServeHTTP(forged, adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", `{"content":"new","requestId":"client-forged"}`, user))
	if forged.Code != http.StatusBadRequest || service.createCalls != 1 {
		t.Fatalf("forged request id was accepted: status=%d calls=%d body=%s", forged.Code, service.createCalls, forged.Body.String())
	}
}

func TestAdminPromptEndpointsRequireLogin(t *testing.T) {
	h := NewHandler(Dependencies{Auth: &fakeAuthService{authErr: authn.ErrUnauthenticated}, AdminPrompts: adminPromptFixture()})
	for _, path := range []string{"/api/v1/admin/capabilities", "/api/v1/admin/prompts", "/api/v1/admin/prompts/script.default/versions/3"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("path=%s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
	}
}

func TestAdminPromptWriteRejectsUnknownSensitiveRequestFields(t *testing.T) {
	service := adminPromptFixture()
	user := authn.User{ID: 7, Capabilities: []string{authn.CapabilityAdminPromptEdit}}
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: user}, AdminPrompts: service})
	req := adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/drafts", `{"content":"safe","password":"secret"}`, user)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") || service.createCalls != 0 {
		t.Fatalf("sensitive request leaked or was processed: status=%d body=%s", rec.Code, rec.Body.String())
	}
}
