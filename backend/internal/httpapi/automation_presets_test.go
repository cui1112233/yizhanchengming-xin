package httpapi

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestFileAutomationPresetStoreKeepsNodeCompatibleData(t *testing.T) {
	now := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "batch-factory-v11-automation-presets.json")
	store, err := NewFileAutomationPresetStore(path, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	created, err := store.Create("alice", "默认预设", map[string]any{"openingEnabled": true, "nested": map[string]any{"count": 2}})
	if err != nil || created.Version != 1 || !strings.HasPrefix(created.ID, "auto-preset-") {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	items, err := store.List("alice")
	if err != nil || len(items) != 1 || items[0].Config["openingEnabled"] != true {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if _, err := store.Update("alice", created.ID, 99, "bad", map[string]any{}); err != ErrAutomationPresetConflict {
		t.Fatalf("conflict err=%v", err)
	}
	updated, err := store.Update("alice", created.ID, 1, "新版", map[string]any{"openingEnabled": false})
	if err != nil || updated.Version != 2 || updated.Name != "新版" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if err := store.Remove("alice", created.ID); err != nil {
		t.Fatal(err)
	}
	reloaded, err := NewFileAutomationPresetStore(path, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	items, err = reloaded.List("alice")
	if err != nil || len(items) != 0 {
		t.Fatalf("reloaded=%+v err=%v", items, err)
	}
}

func TestV12AutomationPresetRoutesUseGoStore(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	presets, err := NewFileAutomationPresetStore(filepath.Join(t.TempDir(), "presets.json"), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, AutomationPresetStore: presets})
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/automation-presets", map[string]any{"name": "预设 A", "config": map[string]any{"openingEnabled": true}})
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), "预设 A") {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/automation-presets", nil)
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "预设 A") {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	if !strings.Contains(created.Body.String(), "auto-preset-") {
		t.Fatal("created preset id missing")
	}
}
