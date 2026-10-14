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
