package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestNovelPanelHistoryAliasUsesGoHistoryStore(t *testing.T) {
	store := &authTestStore{account: auth.Account{Username: "alice", Active: true}, session: auth.Session{Token: "go-session", Username: "alice", ExpiresAt: time.Now().Add(time.Hour)}}
	mux := http.NewServeMux()
	RegisterScriptHistoryRoutes(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: store}, NewScriptHistoryStore(t.TempDir()))
	req := httptest.NewRequest(http.MethodGet, "/api/novel-panel/history", nil)
	req.Header.Set("Authorization", "Bearer go-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"history"`) {
		t.Fatalf("history alias response=%d %s", res.Code, res.Body.String())
	}
}

func TestNovelPanelSettingsAliasUsesGoConfigStore(t *testing.T) {
	store := &authTestStore{account: auth.Account{Username: "alice", Active: true}, session: auth.Session{Token: "go-session", Username: "alice", ExpiresAt: time.Now().Add(time.Hour)}}
	mux := http.NewServeMux()
	RegisterNovelPanelSettingsCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: store}, NewUserConfigStore(t.TempDir()), nil)
	req := httptest.NewRequest(http.MethodGet, "/api/novel-panel/settings", nil)
	req.Header.Set("Authorization", "Bearer go-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"ai_timeout_seconds"`) {
		t.Fatalf("settings alias response=%d %s", res.Code, res.Body.String())
	}
}
