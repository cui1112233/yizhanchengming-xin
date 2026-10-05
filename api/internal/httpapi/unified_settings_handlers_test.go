package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

type fakeUnifiedSettings struct {
	current unifiedsettings.Current
	sync121Calls int
	syncStyleCalls int
}
func (f *fakeUnifiedSettings) GetCurrent(context.Context, int64) (unifiedsettings.Current, error) { return f.current, nil }
func (f *fakeUnifiedSettings) SaveProduction(_ context.Context, id int64, value map[string]any) (unifiedsettings.Current, error) { f.current.ProjectID = id; f.current.Project.Production = value; return f.current, nil }
func (f *fakeUnifiedSettings) SavePublishing(_ context.Context, id int64, value map[string]any) (unifiedsettings.Current, error) { f.current.ProjectID = id; f.current.Project.Publishing = value; return f.current, nil }
func (f *fakeUnifiedSettings) SaveProfile(_ context.Context, id int64, profile unifiedsettings.VersionProfile) (unifiedsettings.Current, error) { f.current.ProjectID = id; f.current.Profile = profile; return f.current, nil }
func (f *fakeUnifiedSettings) Sync121(_ context.Context, id int64) (unifiedsettings.Current, error) { f.sync121Calls++; f.current.ProjectID = id; return f.current, nil }
func (f *fakeUnifiedSettings) SyncStyleTypes(_ context.Context, id int64) (unifiedsettings.Current, error) { f.syncStyleCalls++; f.current.ProjectID = id; return f.current, nil }

func TestUnifiedSettingsRoutesStayUnderBatchProject(t *testing.T) {
	service := &fakeUnifiedSettings{current: unifiedsettings.Current{Priority: []string{"project", "version_profile", "system_default"}}}
	h := NewHandler(Dependencies{UnifiedSettings: service})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/batch-projects/9/settings/production", bytes.NewBufferString(`{"aiCopyEnabled":true,"aiCopyCount":3}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("production status = %d, body=%s", rec.Code, rec.Body.String()) }
	if service.current.Project.Production["aiCopyCount"] != float64(3) { t.Fatalf("production = %#v", service.current.Project.Production) }

	req = httptest.NewRequest(http.MethodPut, "/api/v1/batch-projects/9/settings/publishing", bytes.NewBufferString(`{"uploadVideoType":"individual"}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("publishing status = %d, body=%s", rec.Code, rec.Body.String()) }

	req = httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/9/settings", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("get status = %d, body=%s", rec.Code, rec.Body.String()) }
}

func TestVersionProfileAndSyncRoutes(t *testing.T) {
	service := &fakeUnifiedSettings{}
	h := NewHandler(Dependencies{UnifiedSettings: service})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/batch-projects/9/version-profile", bytes.NewBufferString(`{"name":"女频短剧版","version":"v3","settings":{"processingRulePromptRef":"processing-v3","knowledgePromptRef":"knowledge-v7"}}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("profile status = %d, body=%s", rec.Code, rec.Body.String()) }
	if service.current.Profile.Name != "女频短剧版" || service.current.Profile.Settings.ProcessingRulePromptRef != "processing-v3" || service.current.Profile.Settings.KnowledgePromptRef != "knowledge-v7" {
		t.Fatalf("profile = %#v", service.current.Profile)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/9/version-profile/sync-121", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || service.sync121Calls != 1 { t.Fatalf("sync121 status=%d calls=%d body=%s", rec.Code, service.sync121Calls, rec.Body.String()) }

	req = httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/9/version-profile/sync-style-types", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || service.syncStyleCalls != 1 { t.Fatalf("sync style status=%d calls=%d body=%s", rec.Code, service.syncStyleCalls, rec.Body.String()) }

	req = httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/9/version-profile", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("get profile status=%d body=%s", rec.Code, rec.Body.String()) }
}
