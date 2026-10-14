package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestSyncScriptVideoModelConfigUsesSelectedOwnerModel(t *testing.T) {
	registry := batchfactoryv11.NewMemoryVideoProviderRegistry()
	config := map[string]any{"modelCatalog": []any{
		map[string]any{"kind": "video", "enabled": true, "modelId": "minimax-h3-video", "endpoint": "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}", "tasksEndpoint": "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}", "credential": "h3-secret"},
		map[string]any{"kind": "video", "enabled": true, "modelId": "seedance-2-0-official", "baseUrl": "https://yf.token6688.com", "credential": "seedance-secret"},
		map[string]any{"kind": "video", "enabled": true, "modelId": "yd2-mini-video", "credential": "yd-secret"},
	}}
	for _, tc := range []struct {
		modelKey string
		provider string
		model    string
	}{
		{"minimax-h3-video", batchfactoryv11.VideoProviderAutoDLH3, "minimax-h3-video"},
		{"seedance-2-0-official", batchfactoryv11.VideoProviderYFAISeedance, "seedance-2-0-official"},
		{"yd2-mini-video", batchfactoryv11.VideoProviderPersonalAPI, "yd2-mini-video"},
	} {
		if err := syncScriptVideoModelConfig(context.Background(), registry, "alice", config, tc.modelKey); err != nil {
			t.Fatalf("sync %s: %v", tc.modelKey, err)
		}
		got, err := registry.Resolve(context.Background(), "alice", tc.provider)
		if err != nil || got.Model != tc.model || got.APIKey == "" {
			t.Fatalf("config for %s = %#v, %v", tc.modelKey, got, err)
		}
	}
}

func TestSyncScriptVideoModelConfigRejectsDisabledOrMissingModel(t *testing.T) {
	registry := batchfactoryv11.NewMemoryVideoProviderRegistry()
	config := map[string]any{"modelCatalog": []any{
		map[string]any{"kind": "video", "enabled": false, "modelId": "minimax-h3-video", "credential": "secret"},
	}}
	if err := syncScriptVideoModelConfig(context.Background(), registry, "alice", config, "minimax-h3-video"); err == nil {
		t.Fatal("disabled model must not be configured")
	}
}
