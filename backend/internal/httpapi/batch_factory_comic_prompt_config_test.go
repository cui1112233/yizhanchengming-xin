package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/prompt"
)

func TestHydrateBatchFactoryComicPromptConfigFreezesPublishedBodies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[
		{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产提示词","kind":"base","version":1,"revision":1,"status":"published","body":"漫画资产正文","protocolLock":{"format":"json","slot":"batch.comic-assets"}},
		{"id":"script-constraint-comic-quality","module":"script","name":"漫画画质约束","kind":"addon","version":1,"revision":1,"status":"published","body":"漫画画质正文","protocolLock":{"format":"constraint","category":"comic_quality"}},
		{"id":"script-constraint-comic-restriction","module":"script","name":"漫画画面限制","kind":"addon","version":1,"revision":1,"status":"published","body":"漫画限制正文","protocolLock":{"format":"constraint","category":"comic_restriction"}},
		{"id":"script-constraint-comic-negative","module":"script","name":"漫画负面提示词","kind":"addon","version":1,"revision":1,"status":"published","body":"漫画负面正文","protocolLock":{"format":"constraint","category":"comic_negative"}},
		{"id":"batch-comic-page","module":"batch-factory","name":"漫画提示词","kind":"base","version":1,"revision":1,"status":"published","body":"漫画页面正文","protocolLock":{"format":"json","slot":"batch.comic-page"}}
	]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]any{"comic": map[string]any{
		"enabled": true,
		"assets":  map[string]any{"presetId": "batch-comic-assets"},
		"constraints": []any{
			map[string]any{"presetId": "script-constraint-comic-quality"},
			map[string]any{"presetId": "script-constraint-comic-restriction"},
			map[string]any{"presetId": "script-constraint-comic-negative"},
		},
		"presetId": "batch-comic-page",
	}})
	if err != nil {
		t.Fatal(err)
	}

	hydrated, err := hydrateBatchFactoryPromptConfig(prompt.NewCatalog(path), raw)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(hydrated, &got); err != nil {
		t.Fatal(err)
	}
	comic := got["comic"].(map[string]any)
	if comic["body"] != "漫画页面正文" || comic["presetSlot"] != "batch.comic-page" {
		t.Fatalf("comic page was not frozen: %#v", comic)
	}
	assets := comic["assets"].(map[string]any)
	if assets["body"] != "漫画资产正文" || assets["presetSlot"] != "batch.comic-assets" {
		t.Fatalf("comic assets were not frozen: %#v", assets)
	}
	constraints := comic["constraints"].([]any)
	for _, item := range constraints {
		selection := item.(map[string]any)
		if selection["body"] == "" || selection["constraintCategory"] == "" {
			t.Fatalf("comic constraint was not frozen: %#v", selection)
		}
	}
}

func TestHydrateBatchFactoryComicPromptConfigRejectsWrongPresetSlot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[{"id":"wrong","module":"batch-factory","name":"错误","kind":"base","version":1,"revision":1,"status":"published","body":"错误正文","protocolLock":{"format":"json","slot":"batch.video-meta"}}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"comic":{"enabled":true,"presetId":"wrong"}}`)
	if _, err := hydrateBatchFactoryPromptConfig(prompt.NewCatalog(path), raw); err == nil {
		t.Fatal("accepted a non-comic page preset")
	}
}
