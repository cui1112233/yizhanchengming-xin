package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
)

type fakePublishingService struct {
	createAccountCalls int
	createIntentCalls  int
	listAuditCalls     int
	audits             []publishing.Audit
	createIntentErr    error
}
func (f *fakePublishingService) CreateAccount(_ context.Context, actor authn.User, input publishing.CreateAccountInput) (publishing.Account, error) {
	f.createAccountCalls++
	return publishing.Account{ID: 41, OwnerUserID: actor.ID, TeamID: actor.TeamID, Platform: input.Platform, DisplayName: input.DisplayName, CredentialRefID: "cred-internal", Active: true}, nil
}
func (f *fakePublishingService) ListAccounts(context.Context, authn.User) ([]publishing.Account, error) { return nil, nil }
func (f *fakePublishingService) CreateIntent(_ context.Context, actor authn.User, input publishing.CreateIntentInput) (publishing.Intent, error) {
	f.createIntentCalls++
	return publishing.Intent{ID: 73, BatchProjectID: input.BatchProjectID, PublishingAccountID: input.PublishingAccountID, RequestedByUserID: actor.ID, Platform: "douyin", Status: publishing.IntentStatusPending}, f.createIntentErr
}
func (f *fakePublishingService) GetIntent(context.Context, authn.User, int64) (publishing.Intent, error) { return publishing.Intent{}, nil }
func (f *fakePublishingService) ListAudits(context.Context, authn.User, int64) ([]publishing.Audit, error) { f.listAuditCalls++; return f.audits, nil }

func publishingRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body)); req.Host = "app.example"; req.Header.Set("Origin", "http://app.example"); req.Header.Set("Content-Type", "application/json"); req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"}); return req
}

func TestPublishingEndpointReturns401WhenUnauthenticated(t *testing.T) {
	service := &fakePublishingService{}
	handler := NewHandler(Dependencies{Auth: &fakeAuthService{}, Publishing: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/publishing/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized { t.Fatalf("status=%d body=%s, want 401", rec.Code, rec.Body.String()) }
	if !strings.Contains(rec.Body.String(), `"code":"AUTH_UNAUTHENTICATED"`) { t.Fatalf("body=%s", rec.Body.String()) }
}

func TestPublishingAccountCreateRequiresDedicatedCapability(t *testing.T) {
	service := &fakePublishingService{}; auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/accounts", `{"platform":"douyin","displayName":"Alice 抖音","credential":{"name":"alice","secret":"platform-secret"}}`))
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if service.createAccountCalls != 0 { t.Fatalf("service calls=%d", service.createAccountCalls) }
}

func TestPublishExecuteDoesNotGrantAccountConfigure(t *testing.T) {
	service := &fakePublishingService{}; auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishExecute}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/accounts", `{"platform":"douyin","displayName":"Alice 抖音","credential":{"name":"alice","secret":"platform-secret"}}`))
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if service.createAccountCalls != 0 { t.Fatalf("service calls=%d", service.createAccountCalls) }
}

func TestPublishingAccountCreateHidesCredentialMaterial(t *testing.T) {
	service := &fakePublishingService{}; auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Capabilities: []string{CapabilityPublishAccountConfigure}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/accounts", `{"platform":"douyin","displayName":"Alice 抖音","credential":{"name":"alice","secret":"platform-secret"}}`))
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s, want 201", rec.Code, rec.Body.String()) }
	body := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"platform-secret", "credentialref", "cred-internal", "ciphertext", "nonce", "keyid", "master key", "refresh token", "session token", "cookie", "api secret"} { if strings.Contains(body, strings.ToLower(forbidden)) { t.Fatalf("credential material leaked to browser: %s", rec.Body.String()) } }
}

func TestPublishIntentRequiresExecuteCapability(t *testing.T) {
	service := &fakePublishingService{}; auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishAccountConfigure}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/intents", `{"batchProjectId":21,"publishingAccountId":11}`))
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if service.createIntentCalls != 0 { t.Fatalf("service calls=%d", service.createIntentCalls) }
}

func TestPublishIntentRejectsBrowserSuppliedOutputURL(t *testing.T) {
	service := &fakePublishingService{}; auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishExecute}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/intents", `{"batchProjectId":21,"publishingAccountId":11,"outputRef":"https://evil.example/video.mp4"}`))
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s, want 400", rec.Code, rec.Body.String()) }
	if service.createIntentCalls != 0 { t.Fatalf("service calls=%d; browser media URL must never reach publish service", service.createIntentCalls) }
}

func TestPublishIntentMapsArchivedProjectToSharedConflictCode(t *testing.T) {
	service := &fakePublishingService{createIntentErr: publishing.ErrProjectArchived}; auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishExecute}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/intents", `{"batchProjectId":21,"publishingAccountId":11}`))
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"BATCH_PROJECT_ARCHIVED"`) || service.createIntentCalls != 1 { t.Fatalf("status=%d calls=%d body=%s", rec.Code, service.createIntentCalls, rec.Body.String()) }
}

func TestPublishAuditRequiresAuditCapability(t *testing.T) {
	service := &fakePublishingService{}; auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishExecute}}}; handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodGet, "/api/v1/publishing/audits?batchProjectId=21", ""))
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if service.listAuditCalls != 0 { t.Fatalf("service calls=%d", service.listAuditCalls) }
}

func TestPublishAuditResponseRedactsSensitiveErrorMaterial(t *testing.T) {
	service := &fakePublishingService{audits: []publishing.Audit{{ID: 1, IntentID: 2, BatchProjectID: 21, AccountID: 11, ActorUserID: 7, Platform: "douyin", Action: "publish.failed", Result: "failed", ErrorSummary: "provider returned refresh_token=very-secret cookie=session-value"}}}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishAuditView}}}
	handler := NewHandler(Dependencies{Auth: auth, Publishing: service})
	rec := httptest.NewRecorder(); handler.ServeHTTP(rec, publishingRequest(http.MethodGet, "/api/v1/publishing/audits?batchProjectId=21", ""))
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s, want 200", rec.Code, rec.Body.String()) }
	body := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"very-secret", "session-value", "refresh_token="} { if strings.Contains(body, forbidden) { t.Fatalf("sensitive audit material leaked: %s", rec.Body.String()) } }
	if !strings.Contains(rec.Body.String(), "敏感错误详情已脱敏") { t.Fatalf("expected safe redaction marker, body=%s", rec.Body.String()) }
}
