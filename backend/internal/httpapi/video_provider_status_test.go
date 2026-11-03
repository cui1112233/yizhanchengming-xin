package httpapi

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestVideoProviderStatusDoesNotExposeCredentials(t *testing.T) {
	status := videoProviderStatus(map[string]any{
		"video": map[string]any{"ydApiKey": "secret-key"},
	}, "personal_api")
	if status["configured"] != true || status["model"] != "yd2.0-mini" {
		t.Fatalf("status=%#v", status)
	}
	if strings.Contains(strings.ToLower(strings.Join([]string{status["provider"].(string), status["model"].(string)}, " ")), "secret") {
		t.Fatal("video provider status leaked a credential")
	}
}

func TestV12VideoProviderStatusUsesGoUserConfig(t *testing.T) {
	now := time.Unix(1700000000, 0)
	config := NewUserConfigStore(filepath.Join(t.TempDir(), "users"))
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, UserConfigStore: config})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/video-provider/status?provider=personal_api", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"provider":"personal_api"`) || !strings.Contains(rec.Body.String(), `"configured":false`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12VideoProviderStatusUsesManagedCatalog(t *testing.T) {
	now := time.Unix(1700000000, 0)
	config := NewUserConfigStore(filepath.Join(t.TempDir(), "users"))
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{
		"modelCatalog":        []any{map[string]any{"id": "yd2-mini-video", "kind": "video", "modelId": "yd-model", "credential": "configured"}},
		"modelCatalogVersion": 1,
	}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, UserConfigStore: config, NovelFetchStore: catalog})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/video-provider/status?provider=personal_api", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"model":"yd-model"`) || !strings.Contains(rec.Body.String(), `"configured":true`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12VideoProviderStatusUsesRuntimeRegistry(t *testing.T) {
	now := time.Unix(1700000000, 0)
	config := NewUserConfigStore(filepath.Join(t.TempDir(), "users"))
	accountConfig := novelfetchworkshop.NewMemoryStore()
	registry := batchfactoryv11.NewPersistentVideoProviderRegistry(accountConfig)
	if err := registry.SaveConfig(context.Background(), "alice", batchfactoryv11.VideoProviderConfig{
		Provider: batchfactoryv11.VideoProviderPersonalAPI,
		APIKey:   "yd-secret",
	}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{
		BridgeSecret:    "secret",
		Now:             func() time.Time { return now },
		UserConfigStore: config,
		NovelFetchStore: accountConfig,
		Production:      &batchfactoryv11.ProductionService{ProviderRegistry: registry},
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/video-provider/status?provider=personal_api", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"model":"yd2.0-mini"`) || !strings.Contains(rec.Body.String(), `"configured":true`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "yd-secret") {
		t.Fatal("video provider status leaked a credential")
	}
}
