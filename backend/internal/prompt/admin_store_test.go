package prompt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAdminStoreDraftPublishRollback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	seed := []map[string]any{{
		"id": "demo", "module": "script", "name": "旧版", "kind": "base", "description": "", "compatibleBaseIds": []any{}, "version": 1, "revision": 1, "status": "published", "body": "old", "protocolLock": map[string]any{"slot": "script.demo"}, "createdAt": "2026-01-01T00:00:00Z", "createdBy": "tester", "publishedAt": "2026-01-01T00:00:00Z", "publishedBy": "tester",
	}}
	raw, _ := json.Marshal(seed)
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatal(err)
	}
	store := NewAdminStore(path)
	draft, err := store.CreateDraft("tester", map[string]any{"id": "demo", "module": "script", "name": "新版", "kind": "base", "description": "desc", "compatibleBaseIds": []any{}, "body": "new", "protocolLock": map[string]any{"slot": "script.demo"}})
	if err != nil || intValue(draft["version"]) != 2 || stringValue(draft["status"]) != "draft" {
		t.Fatalf("draft=%v err=%v", draft, err)
	}
	updated, err := store.UpdateDraft("tester", "demo", 2, map[string]any{"id": "demo", "module": "script", "name": "新版改过", "kind": "base", "description": "desc", "compatibleBaseIds": []any{}, "body": "newer", "protocolLock": map[string]any{"slot": "script.demo"}, "expectedRevision": 1})
	if err != nil || intValue(updated["revision"]) != 2 {
		t.Fatalf("updated=%v err=%v", updated, err)
	}
	if _, err := store.Publish("tester", "demo", 2); err != nil {
		t.Fatal(err)
	}
	old, err := store.Get("demo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if stringValue(old["status"]) != "archived" {
		t.Fatalf("old status=%v", old["status"])
	}
	if _, err := store.Rollback("tester", "demo", 1); err != nil {
		t.Fatal(err)
	}
	current, err := store.Get("demo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if stringValue(current["status"]) != "published" {
		t.Fatalf("rollback status=%v", current["status"])
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestAdminStoreDeleteRemovesAnyVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	seed := []map[string]any{{
		"id": "demo", "module": "script", "name": "已发布", "kind": "base", "version": 1, "revision": 1, "status": "published", "body": "old", "protocolLock": map[string]any{"slot": "script.demo"},
	}}
	raw, _ := json.Marshal(seed)
	if err := os.WriteFile(path, raw, 0644); err != nil {
		t.Fatal(err)
	}
	store := NewAdminStore(path)
	draft, err := store.CreateDraft("tester", map[string]any{"id": "demo", "module": "script", "name": "草稿", "kind": "base", "description": "", "compatibleBaseIds": []any{}, "body": "draft", "protocolLock": map[string]any{"slot": "script.demo"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Delete("tester", "demo", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("demo", 1); err == nil {
		t.Fatal("deleted published version is still readable")
	}
	version := intValue(draft["version"])
	if _, err := store.Delete("tester", "demo", version); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("demo", version); err == nil {
		t.Fatal("deleted draft is still readable")
	}
}
