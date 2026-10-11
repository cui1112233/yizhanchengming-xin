package novelfetchworkshop

import "testing"

func TestMergeLegacyRuntimeModelFieldsHydratesOnlyMissingValues(t *testing.T) {
	current := []any{map[string]any{
		"id":         "managed-text",
		"modelId":    "text-model",
		"baseUrl":    "https://db.example/v1",
		"credential": "",
	}}
	legacy := []any{map[string]any{
		"id":         "managed-text",
		"modelId":    "text-model",
		"baseUrl":    "https://legacy.example/v1",
		"credential": "local-secret",
		"endpoint":   "https://legacy.example/v1/chat/completions",
	}}

	mergeLegacyRuntimeModelFields(current, legacy)
	model := current[0].(map[string]any)
	if model["baseUrl"] != "https://db.example/v1" {
		t.Fatalf("existing database address was overwritten: %#v", model)
	}
	if model["credential"] != "local-secret" || model["endpoint"] != "https://legacy.example/v1/chat/completions" {
		t.Fatalf("missing runtime fields were not hydrated: %#v", model)
	}
}

func TestMergeLegacyRuntimeModelFieldsDoesNotMatchDifferentModels(t *testing.T) {
	current := []any{map[string]any{"id": "current", "modelId": "current-model"}}
	legacy := []any{map[string]any{"id": "other", "modelId": "other-model", "credential": "secret"}}

	mergeLegacyRuntimeModelFields(current, legacy)
	if _, exists := current[0].(map[string]any)["credential"]; exists {
		t.Fatal("credential from a different model was copied")
	}
}
