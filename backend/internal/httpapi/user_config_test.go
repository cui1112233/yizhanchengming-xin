package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestUserConfigReportsModelManagementForOwner(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := &authTestStore{account: auth.Account{Username: "tester", Active: true, IsOwner: true}}
	catalogStore := novelfetchworkshop.NewMemoryStore()
	if err := catalogStore.PutConfig(context.Background(), "tester", map[string]any{
		"modelCatalog":        []any{map[string]any{"id": "text-primary", "kind": "text", "enabled": true, "modelId": "managed-text", "baseUrl": "https://text.example/v1", "credential": "secret"}},
		"modelCatalogVersion": 1,
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerUserConfigRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: store}, NewUserConfigStore(t.TempDir()), store, catalogStore)
	handler := mux
	request := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	SignBridgeRequest(request, "tester", true, now, "test-secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("config status=%d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["canManageApi"] != true {
		t.Fatalf("owner config must expose model management: %#v", payload)
	}
	if _, ok := payload["managedBy"]; !ok {
		t.Fatalf("owner config must expose managedBy for frontend parity: %#v", payload)
	}
	if payload["provider"] != "managed" || payload["baseUrl"] != "https://text.example/v1" || payload["model"] != "managed-text" || payload["hasApiKey"] != true {
		t.Fatalf("owner config must expose the enabled managed text model as the effective summary: %#v", payload)
	}
	catalog, ok := payload["modelCatalog"].([]any)
	if !ok || len(catalog) != 1 || catalog[0].(map[string]any)["id"] != "text-primary" {
		t.Fatalf("owner config must expose the managed model catalog: %#v", payload["modelCatalog"])
	}
	model := catalog[0].(map[string]any)
	if model["hasCredential"] != true {
		t.Fatalf("owner config must expose hasCredential for the current settings page: %#v", model)
	}
	defaults, ok := payload["scriptDefaults"].(map[string]any)
	if !ok || defaults["videoResolution"] != "720p" {
		t.Fatalf("owner config must expose script defaults: %#v", payload["scriptDefaults"])
	}
}

func TestUserConfigSeedsImportedWorkshopCatalog(t *testing.T) {
	store := NewUserConfigStore(t.TempDir())
	settings, err := store.GetConfig(context.Background(), "new-user")
	if err != nil {
		t.Fatal(err)
	}
	app, ok := settings["appConfig"].(map[string]any)
	if !ok {
		t.Fatalf("appConfig missing: %#v", settings)
	}
	knowledge, ok := app["knowledge"].(map[string]any)
	if !ok {
		t.Fatalf("appConfig.knowledge missing: %#v", app)
	}
	opening, _ := knowledge["opening_phrases"].(map[string]any)
	items, _ := opening["items"].([]any)
	if len(items) != 174 {
		t.Fatalf("opening phrase seed count = %d, want 174", len(items))
	}
	sensitive, _ := app["sensitive"].(map[string]any)
	groups, _ := sensitive["groups"].([]any)
	if len(groups) != 2 {
		t.Fatalf("sensitive groups = %d, want 2", len(groups))
	}
}
