package app

import (
	"os"
	"testing"

	"qiantie/backend/internal/config"
	"qiantie/backend/internal/giantmaterialresolver"
)

func TestVideoManagementTargetBaseURLDefaultsToHTTPS(t *testing.T) {
	previous := os.Getenv("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL")
	t.Cleanup(func() { _ = os.Setenv("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL", previous) })
	_ = os.Unsetenv("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL")
	if got := videoManagementTargetBaseURL(); got != "https://two.121w.com" {
		t.Fatalf("videoManagementTargetBaseURL() = %q, want HTTPS target", got)
	}
	_ = os.Setenv("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL", "http://two.121w.com/")
	if got := videoManagementTargetBaseURL(); got != "https://two.121w.com" {
		t.Fatalf("legacy target URL = %q, want upgraded HTTPS target", got)
	}
}

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

func TestSelectedTextModelIDAcceptsPersistedSnakeCaseSetting(t *testing.T) {
	settings := map[string]any{"text_model_id": "managed-text"}
	if got := selectedTextModelID(settings); got != "managed-text" {
		t.Fatalf("selectedTextModelID() = %q, want managed-text", got)
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

func TestLocalArtifactCleanupStartsWithoutProductionOrMergeEnabled(t *testing.T) {
	cfg := config.Config{
		Slice:                    3,
		LocalMergeEnabled:        false,
		LocalExecutorArtifactDir: `F:\local-artifacts`,
	}
	if !shouldStartLocalArtifactPurger(cfg) {
		t.Fatal("local artifact cleanup must remain available before production and merge are enabled")
	}
	cfg.LocalExecutorArtifactDir = ""
	if shouldStartLocalArtifactPurger(cfg) {
		t.Fatal("local artifact cleanup must not start without a configured artifact directory")
	}
}
