package novelfetchworkshop

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

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

func TestReadLegacyConfigFileCachesUntilFileChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alice", "api-config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"modelCatalog":[{"id":"one"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store := &MySQLStore{LegacyConfigDir: dir}

	first := store.readLegacyConfigFile("alice")
	if catalog, _ := first["modelCatalog"].([]any); len(catalog) != 1 {
		t.Fatalf("first read catalog = %#v", first["modelCatalog"])
	}
	// A second read with an unchanged file must reuse the cached document.
	if second := store.readLegacyConfigFile("alice"); reflect.ValueOf(second).Pointer() != reflect.ValueOf(first).Pointer() {
		t.Fatal("unchanged legacy config was not served from cache")
	}

	if err := os.WriteFile(path, []byte(`{"modelCatalog":[{"id":"one"},{"id":"two"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	modTime := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
	updated := store.readLegacyConfigFile("alice")
	if catalog, _ := updated["modelCatalog"].([]any); len(catalog) != 2 {
		t.Fatalf("updated read catalog = %#v", updated["modelCatalog"])
	}
	if reflect.ValueOf(updated).Pointer() == reflect.ValueOf(first).Pointer() {
		t.Fatal("legacy config cache was not invalidated after the file changed")
	}
}
