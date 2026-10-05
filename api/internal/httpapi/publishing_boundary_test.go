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
}

func (f *fakePublishingService) CreateAccount(_ context.Context, actor authn.User, input publishing.CreateAccountInput) (publishing.Account, error) {
	f.createAccountCalls++
	return publishing.Account{ID: 41, OwnerUserID: actor.ID, TeamID: actor.TeamID, Platform: input.Platform, DisplayName: input.DisplayName, CredentialRef: input.CredentialRef, Active: true}, nil
}
func (f *fakePublishingService) ListAccounts(context.Context, authn.User) ([]publishing.Account, error) { return nil, nil }
func (f *fakePublishingService) CreateIntent(_ context.Context, actor authn.User, input publishing.CreateIntentInput) (publishing.Intent, error) {
	f.createIntentCalls++
	return publishing.Intent{ID: 73, BatchProjectID: input.BatchProjectID, PublishingAccountID: input.PublishingAccountID, RequestedByUserID: actor.ID, Platform: "douyin", Status: publishing.IntentStatusPending}, nil
}
func (f *fakePublishingService) GetIntent(context.Context, authn.User, int64) (publishing.Intent, error) { return publishing.Intent{}, nil }
func (f *fakePublishingService) ListAudits(context.Context, authn.User, int64) ([]publishing.Audit, error) { f.listAuditCalls++; return nil, nil }

func publishingRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	return req
}

func TestPublishingAccountCreateRequiresDedicatedCapability(t *testing.T) {
	publishingService := &fakePublishingService{}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}
	handler := NewHandler(Dependencies{Auth: auth, Publishing: publishingService})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/accounts", `{"platform":"douyin","displayName":"Alice 抖音","credentialRef":"vault://publisher/alice"}`))
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if publishingService.createAccountCalls != 0 { t.Fatalf("service calls=%d; missing capability must stop before business service", publishingService.createAccountCalls) }
}

func TestPublishingAccountCreateHidesCredentialReference(t *testing.T) {
	publishingService := &fakePublishingService{}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Capabilities: []string{CapabilityPublishAccountConfigure}}}
	handler := NewHandler(Dependencies{Auth: auth, Publishing: publishingService})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/accounts", `{"platform":"douyin","displayName":"Alice 抖音","credentialRef":"vault://publisher/alice"}`))
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s, want 201", rec.Code, rec.Body.String()) }
	if strings.Contains(rec.Body.String(), "vault://") || strings.Contains(strings.ToLower(rec.Body.String()), "credentialref") {
		t.Fatalf("credential reference leaked to browser: %s", rec.Body.String())
	}
}

func TestPublishIntentRequiresExecuteCapability(t *testing.T) {
	publishingService := &fakePublishingService{}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishAccountConfigure}}}
	handler := NewHandler(Dependencies{Auth: auth, Publishing: publishingService})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, publishingRequest(http.MethodPost, "/api/v1/publishing/intents", `{"batchProjectId":21,"publishingAccountId":11,"outputRef":"tos://video/final.mp4"}`))
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if publishingService.createIntentCalls != 0 { t.Fatalf("service calls=%d; missing publish.execute must stop before service", publishingService.createIntentCalls) }
}

func TestPublishAuditRequiresAuditCapability(t *testing.T) {
	publishingService := &fakePublishingService{}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityPublishExecute}}}
	handler := NewHandler(Dependencies{Auth: auth, Publishing: publishingService})
	req := publishingRequest(http.MethodGet, "/api/v1/publishing/audits?batchProjectId=21", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden { t.Fatalf("status=%d body=%s, want 403", rec.Code, rec.Body.String()) }
	if publishingService.listAuditCalls != 0 { t.Fatalf("service calls=%d; missing audit capability must stop before service", publishingService.listAuditCalls) }
}
