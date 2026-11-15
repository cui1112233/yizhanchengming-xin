package app

import (
	"context"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

type fixedTextProvider func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error)

func (f fixedTextProvider) Complete(ctx context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	return f(ctx, request)
}

func TestOwnerTextProviderUsesFallbackProvider(t *testing.T) {
	provider := ownerTextProvider{fallback: fixedTextProvider(func(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
		if request.UserPrompt != "ping" {
			t.Fatalf("unexpected prompt: %q", request.UserPrompt)
		}
		return "pong", nil
	})}
	result, err := provider.Complete(context.Background(), batchfactoryv11.TextCompletionRequest{UserPrompt: "ping"})
	if err != nil || result != "pong" {
		t.Fatalf("fallback provider result=%q err=%v", result, err)
	}
}

func TestOwnerTextProviderRejectsTypedNilFallback(t *testing.T) {
	var typedNil *batchfactoryv11.OpenAICompatibleProvider
	provider := ownerTextProvider{fallback: typedNil}
	if _, err := provider.Complete(context.Background(), batchfactoryv11.TextCompletionRequest{UserPrompt: "ping"}); err == nil {
		t.Fatal("typed nil fallback must return an unavailable-provider error")
	}
}

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

func TestResolveNovelFetchTextProviderFallsBackWhenSavedDefaultIsUnavailable(t *testing.T) {
	settings := map[string]any{
		// 历史配置里保存的默认模型已经下架。
		"text_model_id": "[anti-2]gemini-3.8-flash",
		"modelCatalog": []any{
			map[string]any{"id": "text-primary", "kind": "text", "enabled": true, "baseUrl": "https://text.example/v1", "modelId": "text-model", "credential": "secret"},
		},
	}
	provider, err := resolveNovelFetchTextProvider(settings, "")
	if err != nil {
		t.Fatalf("resolveNovelFetchTextProvider returned error: %v", err)
	}
	if provider == nil || provider.Model != "text-model" {
		t.Fatalf("expected fallback to first enabled model, got %#v", provider)
	}
}

func TestResolveNovelFetchTextProviderKeepsExplicitSelectionStrict(t *testing.T) {
	settings := map[string]any{
		"modelCatalog": []any{
			map[string]any{"id": "text-primary", "kind": "text", "enabled": true, "baseUrl": "https://text.example/v1", "modelId": "text-model", "credential": "secret"},
		},
	}
	provider, err := resolveNovelFetchTextProvider(settings, "removed-model")
	if err != nil {
		t.Fatalf("resolveNovelFetchTextProvider returned error: %v", err)
	}
	if provider != nil {
		t.Fatalf("explicit unavailable selection must not fall back, got %#v", provider)
	}
}

func TestResolveNovelFetchTextProviderPrefersSavedDefault(t *testing.T) {
	settings := map[string]any{
		"text_model_id": "text-preferred",
		"modelCatalog": []any{
			map[string]any{"id": "text-first", "kind": "text", "enabled": true, "baseUrl": "https://first.example/v1", "modelId": "first-model", "credential": "x"},
			map[string]any{"id": "text-preferred", "kind": "text", "enabled": true, "baseUrl": "https://preferred.example/v1", "modelId": "preferred-model", "credential": "y"},
		},
	}
	provider, err := resolveNovelFetchTextProvider(settings, "")
	if err != nil {
		t.Fatalf("resolveNovelFetchTextProvider returned error: %v", err)
	}
	if provider == nil || provider.Model != "preferred-model" {
		t.Fatalf("expected saved default model, got %#v", provider)
	}
}
