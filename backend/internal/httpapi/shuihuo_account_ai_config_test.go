package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestRegisterShuihuoAccountAIConfigCompatPersistsLegacyModelsWithoutReturningSecrets(t *testing.T) {
	now := time.Unix(100, 0)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "alice", map[string]any{
		"modelCatalog": []any{map[string]any{"id": "keep-me", "kind": "text", "modelId": "existing", "enabled": true}},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterShuihuoAccountAIConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store)
	recorder := signedJSONRequest(t, mux, now, "alice", http.MethodPut, "/api/shuihuo-production/account-ai-config", map[string]any{
		"provider": "custom", "baseUrl": "https://text.example/v1", "model": "text-model", "apiKey": "text-secret",
		"image": map[string]any{"provider": "openai_compatible", "baseUrl": "https://image.example/v1", "model": "image-model", "apiKey": "image-secret"},
		"video": map[string]any{"provider": "yd_video", "apiKey": "video-secret"},
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if containsAny(recorder.Body.String(), "text-secret", "image-secret", "video-secret") {
		t.Fatalf("response leaked credentials: %s", recorder.Body.String())
	}
	settings, err := store.GetConfig(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	catalog, ok := settings["modelCatalog"].([]any)
	if !ok || len(catalog) != 3 {
		t.Fatalf("unexpected catalog: %#v", settings["modelCatalog"])
	}
	for _, raw := range catalog {
		model := raw.(map[string]any)
		if model["kind"] == "text" && model["modelId"] == "text-model" && model["credential"] != "text-secret" {
			t.Fatalf("text credential was not stored for server-side use: %#v", model)
		}
	}
}

func TestRegisterShuihuoAccountAIConfigCompatRejectsMissingTextFields(t *testing.T) {
	mux := http.NewServeMux()
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Unix(100, 0)
	RegisterShuihuoAccountAIConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store)
	recorder := signedJSONRequest(t, mux, now, "alice", http.MethodPut, "/api/shuihuo-production/account-ai-config", map[string]any{"provider": "custom"})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
