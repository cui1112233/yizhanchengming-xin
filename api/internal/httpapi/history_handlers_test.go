package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workspace"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type historyReaderFake struct {
	query  workspace.HistoryQuery
	result workspace.HistoryPage
	err    error
	calls  int
}

func (f *historyReaderFake) ListHistory(_ context.Context, q workspace.HistoryQuery) (workspace.HistoryPage, error) {
	f.query = q
	f.calls++
	return f.result, f.err
}
func historyRequest(h http.Handler, path string, cookie bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie {
		r.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	}
	r.Header.Set("X-Request-ID", "history-test-request")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestHistoryOwnershipRoleAndTeamBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, role string
		team       int64
		elevated   bool
	}{{"zero team member", "member", 0, false}, {"nonzero team member", "member", 3, false}, {"dev remains scoped", "dev", 3, false}, {"admin", "admin", 0, true}, {"owner", "owner", 0, true}} {
		t.Run(tc.name, func(t *testing.T) {
			f := &historyReaderFake{result: workspace.HistoryPage{Entries: []workspace.HistoryItem{}}}
			h := NewHandler(Dependencies{WorkspaceHistory: f, Auth: &fakeAuthService{user: authn.User{ID: 7, TeamID: tc.team, Role: tc.role, Capabilities: []string{CapabilityBatchView}}}})
			w := historyRequest(h, "/api/v1/history", true)
			if w.Code != 200 {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if f.query.UserID != 7 || f.query.TeamID != tc.team || f.query.Elevated != tc.elevated || f.query.Page != 1 || f.query.Limit != 20 || f.query.Archived != "all" {
				t.Fatalf("scope/defaults=%+v", f.query)
			}
		})
	}
}
func TestHistoryValidatesFiltersAndPaginates(t *testing.T) {
	f := &historyReaderFake{result: workspace.HistoryPage{Entries: []workspace.HistoryItem{{ID: "intake:intake:42", Kind: "intake", Origin: "intake", SourceID: "42", IntakeID: 42, Title: "独立获取", Status: "unknown", SourceStatus: "unknown", UpdatedAt: time.Now()}}, Total: 31}}
	h := NewHandler(Dependencies{WorkspaceHistory: f, Auth: &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}})
	w := historyRequest(h, "/api/v1/history?page=2&limit=10&q=故事&kind=script&status=completed&archived=archived", true)
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["page"] != float64(2) || got["limit"] != float64(10) || got["total"] != float64(31) || f.query.Kind != "script" || f.query.Status != "completed" || f.query.Q != "故事" || f.query.Archived != "archived" {
		t.Fatalf("body=%s query=%+v", w.Body.String(), f.query)
	}
	for _, suffix := range []string{"page=0", "page=oops", "page=2147483648", "limit=101", "limit=0", "kind=evil", "status=succeeded", "archived=restore", "q=" + strings.Repeat("a", 201)} {
		f.calls = 0
		w := historyRequest(h, "/api/v1/history?"+suffix, true)
		if w.Code != 400 || f.calls != 0 {
			t.Fatalf("%s status=%d calls=%d", suffix, w.Code, f.calls)
		}
	}
}
func TestHistoryAuthCapabilityAndSafeErrors(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}
	f := &historyReaderFake{err: errors.New("password=private-provider-payload")}
	for _, tc := range []struct {
		name   string
		deps   Dependencies
		cookie bool
		want   int
	}{{"unauthenticated", Dependencies{Auth: auth, WorkspaceHistory: f}, false, 401}, {"forbidden", Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7}}, WorkspaceHistory: f}, true, 403}, {"unavailable", Dependencies{Auth: auth}, true, 503}, {"read failure", Dependencies{Auth: auth, WorkspaceHistory: f}, true, 500}} {
		t.Run(tc.name, func(t *testing.T) {
			w := historyRequest(NewHandler(tc.deps), "/api/v1/history", tc.cookie)
			if w.Code != tc.want || strings.Contains(w.Body.String(), "private-provider") {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if w.Header().Get("X-Request-ID") == "" {
				t.Fatal("missing request correlation")
			}
			if tc.want >= 500 && !strings.Contains(w.Body.String(), "history-test-request") {
				t.Fatal(w.Body.String())
			}
		})
	}
}
func TestHistoryDoesNotRegisterDeletionOrRestore(t *testing.T) {
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityBatchView}}}})
	for _, tc := range []struct{ method, path string }{{"DELETE", "/api/v1/history"}, {"POST", "/api/v1/history/clear"}, {"POST", "/api/v1/history/42/restore"}} {
		r := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 404 && w.Code != 405 {
			t.Fatalf("%s %s: %d", tc.method, tc.path, w.Code)
		}
	}
}
