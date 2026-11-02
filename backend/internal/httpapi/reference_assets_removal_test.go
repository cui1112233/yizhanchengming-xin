package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestRouterRejectsRemovedNovelPanelAPI(t *testing.T) {
	now := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "secret",
		Now:             func() time.Time { return now },
		AuthStore:       &authTestStore{account: auth.Account{Username: "alice", Active: true}},
		UserDataDir:     t.TempDir(),
		UserConfigStore: NewUserConfigStore(t.TempDir()),
	})
	for _, testCase := range []struct {
		method string
		endpoint string
	}{
		{http.MethodGet, "/api/novel-panel/projects"},
		{http.MethodGet, "/api/novel-panel/diagnostics"},
		{http.MethodPost, "/api/novel-panel/reference-assets/upload"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(testCase.method, testCase.endpoint, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s %s status = %d, want %d", testCase.method, testCase.endpoint, response.Code, http.StatusNotFound)
		}
	}
}
