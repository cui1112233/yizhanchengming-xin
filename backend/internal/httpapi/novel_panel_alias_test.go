package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
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
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"ai_timeout_seconds":400`) {
		t.Fatalf("settings alias response=%d %s", res.Code, res.Body.String())
	}
}

func TestNovelPanelSettingsAliasUsesManagedTextModel(t *testing.T) {
	store := &authTestStore{account: auth.Account{Username: "alice", Active: true}, session: auth.Session{Token: "go-session", Username: "alice", ExpiresAt: time.Now().Add(time.Hour)}}
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{
		"modelCatalog": []any{map[string]any{"id": "managed-text", "kind": "text", "modelId": "gemini-configured", "baseUrl": "https://text.example/v1", "credential": "secret", "enabled": true}},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterNovelPanelSettingsCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: store}, NewUserConfigStore(t.TempDir()), nil, catalog)
	req := httptest.NewRequest(http.MethodGet, "/api/novel-panel/settings", nil)
	req.Header.Set("Authorization", "Bearer go-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"text_model_id":"managed-text"`) || !strings.Contains(res.Body.String(), `"model":"gemini-configured"`) || !strings.Contains(res.Body.String(), `"api_key_configured":true`) {
		t.Fatalf("settings catalog response=%d %s", res.Code, res.Body.String())
	}
}

func TestNovelPanelImageSettingsUsesManagedImageModel(t *testing.T) {
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{
		"modelCatalog": []any{map[string]any{
			"id": "managed-image", "kind": "image", "modelId": "image-configured", "baseUrl": "https://image.example/v1", "credential": "secret", "enabled": true,
			"capabilities": map[string]any{"supportsReferenceImages": true},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	settings := readPanelImageSettingsWithCatalog(context.Background(), NewUserConfigStore(t.TempDir()), "alice", catalog)
	if settings["base_url"] != "https://image.example/v1" || settings["model"] != "image-configured" || settings["api_key"] != "secret" || settings["supports_reference"] != true {
		t.Fatalf("unexpected image settings: %#v", settings)
	}
}
