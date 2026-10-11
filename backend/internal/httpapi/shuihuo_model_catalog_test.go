package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestLoadShuihuoOwnerConfigUsesManagedModelCatalog(t *testing.T) {
	shuihuoStore := batchfactoryv11.NewMemoryStore()
	if err := shuihuoStore.SaveShuihuoConfig(context.Background(), "tester", map[string]any{"textModelId": "legacy"}); err != nil {
		t.Fatal(err)
	}
	catalogStore := novelfetchworkshop.NewMemoryStore()
	if err := catalogStore.PutConfig(context.Background(), "tester", map[string]any{
		"modelCatalog":        []any{map[string]any{"id": "managed-text", "kind": "text", "enabled": true}},
		"modelCatalogVersion": 1,
	}); err != nil {
		t.Fatal(err)
	}
	config, err := loadShuihuoOwnerConfig(context.Background(), shuihuoStore, "tester", catalogStore)
	if err != nil {
		t.Fatal(err)
	}
	catalog, ok := config["modelCatalog"].([]any)
	if !ok || len(catalog) != 1 || catalog[0].(map[string]any)["id"] != "managed-text" {
		t.Fatalf("unexpected managed catalog: %#v", config["modelCatalog"])
	}
}

func TestSafeShuihuoConfigForResponseHidesModelCredentials(t *testing.T) {
	config := map[string]any{
		"textModelId": "managed-text",
		"modelCatalog": []any{map[string]any{
			"id": "managed-text", "kind": "text", "modelId": "gemini",
			"credential": "secret-value", "apiKey": "api-key-value", "enabled": true,
		}},
	}

	got := safeShuihuoConfigForResponse(config)
	catalog, ok := got["modelCatalog"].([]map[string]any)
	if !ok || len(catalog) != 1 {
		t.Fatalf("catalog = %#v", got["modelCatalog"])
	}
	if _, exists := catalog[0]["credential"]; exists {
		t.Fatalf("credential leaked: %#v", catalog[0])
	}
	if _, exists := catalog[0]["apiKey"]; exists {
		t.Fatalf("api key leaked: %#v", catalog[0])
	}
	if catalog[0]["hasCredential"] != true || got["textModelId"] != "managed-text" {
		t.Fatalf("safe config lost compatibility fields: %#v", got)
	}
}

func TestShuihuoTextModelConfiguredUsesManagedCatalog(t *testing.T) {
	configured := map[string]any{
		"modelCatalog": []any{map[string]any{
			"kind": "text", "enabled": true, "modelId": "[anti-2]gemini-3.8-flash",
			"baseUrl": "https://example.test/v1", "credential": "secret",
		}},
	}
	if !shuihuoTextModelConfigured(configured) {
		t.Fatal("enabled managed text model should make Shuihuo text readiness true")
	}
	if shuihuoTextModelConfigured(map[string]any{
		"modelCatalog": []any{map[string]any{
			"kind": "text", "enabled": false, "modelId": "model",
			"baseUrl": "https://example.test/v1", "credential": "secret",
		}},
	}) {
		t.Fatal("disabled managed text model must not make Shuihuo text readiness true")
	}
}

func TestShuihuoModelKindConfiguredUsesCredentialReferenceAndAudioFields(t *testing.T) {
	t.Setenv("V88_TEXT_MODEL_KEY", "text-secret")
	config := map[string]any{"modelCatalog": []any{
		map[string]any{"kind": "text", "enabled": true, "modelId": "text-model", "baseUrl": "https://text.example/v1", "credentialRef": "V88_TEXT_MODEL_KEY"},
		map[string]any{"kind": "audio", "enabled": true, "modelId": "tts-model", "endpoint": "https://audio.example/speech", "credential": "audio-secret", "requestTemplate": `{\"body\":{}}`, "responseMapping": `{\"resultUrl\":\"data.url\"}`},
	}}
	if !shuihuoModelKindConfigured(config, "text") {
		t.Fatal("text model with a credential reference should be ready")
	}
	if !shuihuoModelKindConfigured(config, "audio") {
		t.Fatal("audio model with endpoint, credential, template and mapping should be ready")
	}
	config["modelCatalog"] = []any{map[string]any{"kind": "audio", "enabled": true, "modelId": "tts-model", "endpoint": "https://audio.example/speech", "credential": "audio-secret"}}
	if shuihuoModelKindConfigured(config, "audio") {
		t.Fatal("audio model without request template and response mapping must not be ready")
	}
}

func TestShuihuoImageModelReadyUsesBaseURLAndCredentialRef(t *testing.T) {
	t.Setenv("V88_IMAGE_MODEL_KEY", "image-secret")
	model := map[string]any{
		"kind":          "image",
		"enabled":       true,
		"modelId":       "image-model",
		"baseUrl":       "https://images.example/v1",
		"credentialRef": "V88_IMAGE_MODEL_KEY",
	}
	if !shuihuoImageModelReady(model) {
		t.Fatal("image model with a valid base URL and credential reference should be ready")
	}
	model["enabled"] = false
	if shuihuoImageModelReady(model) {
		t.Fatal("disabled image model must not be ready")
	}
}

func TestPatchShuihuoModelAllowsStatusOnlyUpdates(t *testing.T) {
	current := map[string]any{"id": "text-model", "name": "文本模型", "kind": "text", "enabled": true}
	updated, err := patchShuihuoModel(current, map[string]any{"enabled": false})
	if err != nil {
		t.Fatalf("status-only patch should be accepted: %v", err)
	}
	if updated["enabled"] != false || updated["name"] != "文本模型" || updated["kind"] != "text" {
		t.Fatalf("unexpected patched model: %#v", updated)
	}
}

func TestPatchShuihuoModelAcceptsSharedCatalogDisplayName(t *testing.T) {
	current := map[string]any{"id": "text-model", "displayName": "共享文本模型", "kind": "text", "enabled": true}
	updated, err := patchShuihuoModel(current, map[string]any{"enabled": false})
	if err != nil || updated["enabled"] != false {
		t.Fatalf("shared catalog status-only patch should be accepted: %#v, %v", updated, err)
	}
}
