package app

import (
	"testing"

	"qiantie/backend/internal/config"
	"qiantie/backend/internal/giantmaterialresolver"
)

func TestConfiguredTextProviderAcceptsCatalogIDOrProviderModelID(t *testing.T) {
	settings := map[string]any{"modelCatalog": []any{map[string]any{
		"id": "managed-text", "kind": "text", "enabled": true,
		"baseUrl": "https://text.example/v1", "modelId": "gemini-configured", "credential": "secret",
	}}}
	for _, selected := range []string{"managed-text", "gemini-configured", ""} {
		provider, err := configuredTextProvider(settings, selected)
		if err != nil {
			t.Fatalf("selected=%q: %v", selected, err)
		}
		if provider == nil || provider.Model != "gemini-configured" {
			t.Fatalf("selected=%q provider=%#v", selected, provider)
		}
	}
	provider, err := configuredTextProvider(settings, "other-model")
	if err != nil {
		t.Fatal(err)
	}
	if provider != nil {
		t.Fatalf("unexpected provider for unknown selection: %#v", provider)
	}
}

func TestSelectGiantMaterialResolverUsesHTTPPluginWhenConfigured(t *testing.T) {
	resolver := selectGiantMaterialResolver(config.Config{
		GiantMaterialPluginURL:   "http://127.0.0.1:18180",
		GiantMaterialPluginToken: "plugin-token",
	})
	plugin, ok := resolver.(*giantmaterialresolver.HTTPPluginClient)
	if !ok {
		t.Fatalf("resolver type = %T, want *HTTPPluginClient", resolver)
	}
	if plugin.BaseURL != "http://127.0.0.1:18180" || plugin.Token != "plugin-token" {
		t.Fatalf("plugin = %#v", plugin)
	}
}
