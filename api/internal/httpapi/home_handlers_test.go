package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workspace"
)

type fakeRecentReader struct {
	items         []workspace.RecentItem
	err           error
	userID        int64
	teamID        int64
	limit         int
	scopedCalls   int
	elevatedCalls int
}

func (f *fakeRecentReader) ListRecent(_ context.Context, userID, teamID int64, limit int) ([]workspace.RecentItem, error) {
	f.userID, f.teamID, f.limit = userID, teamID, limit
	f.scopedCalls++
	return f.items, f.err
}

func (f *fakeRecentReader) ListRecentElevated(_ context.Context, limit int) ([]workspace.RecentItem, error) {
	f.limit = limit
	f.elevatedCalls++
	return f.items, f.err
}

func recentRequest(handler http.Handler, path string, withCookie bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if withCookie {
		req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestWorkspaceRecentRequiresSessionAndBatchView(t *testing.T) {
	reader := &fakeRecentReader{}
	unauthenticated := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}, WorkspaceRecent: reader})
	if rec := recentRequest(unauthenticated, "/api/v1/workspace/recent", false); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status=%d body=%s", rec.Code, rec.Body.String())
	}

	forbidden := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7}}, WorkspaceRecent: reader})
	if rec := recentRequest(forbidden, "/api/v1/workspace/recent", true); rec.Code != http.StatusForbidden {
		t.Fatalf("forbidden status=%d body=%s", rec.Code, rec.Body.String())
	}
	if reader.scopedCalls != 0 || reader.elevatedCalls != 0 {
		t.Fatalf("reader called across auth boundary: %#v", reader)
	}
}

func TestWorkspaceRecentUsesScopedQueryForMemberAndDev(t *testing.T) {
	for _, tc := range []struct {
		name string
		role string
	}{
		{name: "member", role: "member"},
		{name: "dev is not elevated", role: "dev"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &fakeRecentReader{items: []workspace.RecentItem{}}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: tc.role, Capabilities: []string{CapabilityBatchView}}}
			rec := recentRequest(NewHandler(Dependencies{Auth: auth, WorkspaceRecent: reader}), "/api/v1/workspace/recent", true)
			if rec.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if reader.scopedCalls != 1 || reader.elevatedCalls != 0 || reader.userID != 7 || reader.teamID != 3 || reader.limit != 6 {
				t.Fatalf("reader=%#v", reader)
			}
			if strings.TrimSpace(rec.Body.String()) != `{"items":[]}` {
				t.Fatalf("body=%s", rec.Body.String())
			}
		})
	}
}

func TestWorkspaceRecentUsesDedicatedElevatedQueryForAdminAndOwner(t *testing.T) {
	for _, role := range []string{"admin", "owner"} {
		t.Run(role, func(t *testing.T) {
			reader := &fakeRecentReader{items: []workspace.RecentItem{}}
			auth := &fakeAuthService{user: authn.User{ID: 1, Role: role, Capabilities: []string{CapabilityBatchView}}}
			rec := recentRequest(NewHandler(Dependencies{Auth: auth, WorkspaceRecent: reader}), "/api/v1/workspace/recent?limit=20", true)
			if rec.Code != http.StatusOK || reader.elevatedCalls != 1 || reader.scopedCalls != 0 || reader.limit != 20 {
				t.Fatalf("status=%d body=%s reader=%#v", rec.Code, rec.Body.String(), reader)
			}
		})
	}
}

func TestWorkspaceRecentValidatesLimitWithoutCallingStore(t *testing.T) {
	for _, value := range []string{"0", "21", "abc", "1.5"} {
		t.Run(value, func(t *testing.T) {
			reader := &fakeRecentReader{}
			auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}
			rec := recentRequest(NewHandler(Dependencies{Auth: auth, WorkspaceRecent: reader}), "/api/v1/workspace/recent?limit="+value, true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if reader.scopedCalls != 0 || reader.elevatedCalls != 0 {
				t.Fatal("invalid input reached store")
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body["code"] != "INVALID_LIMIT" || body["request_id"] == "" {
				t.Fatalf("body=%s", rec.Body.String())
			}
		})
	}
}

func TestWorkspaceRecentResponseContainsOnlyPublicProjectionFields(t *testing.T) {
	updated := time.Date(2026, 10, 9, 4, 5, 6, 0, time.UTC)
	reader := &fakeRecentReader{items: []workspace.RecentItem{{Kind: "video", ID: "video:9", Title: "项目", Status: "running", UpdatedAt: updated, Href: "/batch-factory?projectId=12"}}}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}
	rec := recentRequest(NewHandler(Dependencies{Auth: auth, WorkspaceRecent: reader}), "/api/v1/workspace/recent?limit=1", true)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string][]map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	item := body["items"][0]
	if len(item) != 6 {
		t.Fatalf("public item fields=%v", item)
	}
	for _, key := range []string{"kind", "id", "title", "status", "updatedAt", "href"} {
		if _, ok := item[key]; !ok {
			t.Fatalf("missing %s in %v", key, item)
		}
	}
	for _, forbidden := range []string{"error", "provider", "bucket", "objectKey", "input", "output"} {
		if _, ok := item[forbidden]; ok {
			t.Fatalf("leaked %s in %v", forbidden, item)
		}
	}
}

func TestWorkspaceRecentReadFailureIsSafeAndCorrelated(t *testing.T) {
	reader := &fakeRecentReader{err: errors.New("mysql://root:secret@prod/private")}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}
	rec := recentRequest(NewHandler(Dependencies{Auth: auth, WorkspaceRecent: reader}), "/api/v1/workspace/recent", true)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "root") || strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "prod") {
		t.Fatalf("internal error leaked: %s", rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "RECENT_READ_FAILED" || body["message"] != "读取最近创作失败" || body["request_id"] == "" {
		t.Fatalf("body=%s", rec.Body.String())
	}
}
