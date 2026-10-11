package app

import "testing"

func TestConfiguredTextProviderUsesFirstEnabledCatalogModelWhenNoSelection(t *testing.T) {
	provider, err := configuredTextProvider(map[string]any{
		"modelCatalog": []any{
			map[string]any{"id": "disabled", "kind": "text", "enabled": false, "baseUrl": "https://disabled.example/v1", "modelId": "disabled", "credential": "x"},
			map[string]any{"id": "text-primary", "kind": "text", "enabled": true, "baseUrl": "https://text.example/v1", "modelId": "text-model", "credential": "secret"},
		},
	}, "")
	if err != nil {
		t.Fatalf("configuredTextProvider returned error: %v", err)
	}
	if provider == nil {
		t.Fatal("configuredTextProvider returned nil for an enabled catalog model")
	}
	if provider.Endpoint != "https://text.example/v1/chat/completions" || provider.Model != "text-model" || provider.APIKey != "secret" {
		t.Fatalf("unexpected provider: %#v", provider)
	}
}
