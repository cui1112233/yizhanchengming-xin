package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/prompt"
)

func TestNormalizeBatchFactoryProductionModeAcceptsComicAndLegacyDirectorModes(t *testing.T) {
	for _, mode := range []string{"", "video", "comic", "original", "viral"} {
		if _, err := normalizeBatchFactoryProductionMode(mode); err != nil {
			t.Fatalf("mode %q rejected: %v", mode, err)
		}
	}
	if _, err := normalizeBatchFactoryProductionMode("invalid"); err == nil {
		t.Fatal("invalid mode was accepted")
	}
}

func TestHydrateBatchFactoryComicPromptConfigFreezesPublishedBodies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[
		{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产提示词","kind":"base","version":1,"revision":1,"status":"published","body":"漫画资产正文","protocolLock":{"format":"json","slot":"batch.comic-assets"}},
		{"id":"script-constraint-comic-quality","module":"script","name":"漫画画质约束","kind":"addon","version":1,"revision":1,"status":"published","body":"漫画画质正文","protocolLock":{"format":"constraint","category":"comic_quality"}},
		{"id":"script-constraint-comic-restriction","module":"script","name":"漫画画面限制","kind":"addon","version":1,"revision":1,"status":"published","body":"漫画限制正文","protocolLock":{"format":"constraint","category":"comic_restriction"}},
		{"id":"script-constraint-comic-negative","module":"script","name":"漫画负面提示词","kind":"addon","version":1,"revision":1,"status":"published","body":"漫画负面正文","protocolLock":{"format":"constraint","category":"comic_negative"}},
		{"id":"batch-comic-page","module":"batch-factory","name":"漫画提示词","kind":"base","version":1,"revision":1,"status":"published","body":"漫画页面正文","protocolLock":{"format":"json","slot":"batch.comic-page"}},
		{"id":"comic-image-derive-default","module":"batch-factory","name":"图片描述词推导","kind":"base","version":1,"revision":1,"status":"published","body":"图片推导正文","protocolLock":{"format":"json","slot":"batch.comic-image-derivation"}},
		{"id":"batch-comic-style-01","module":"batch-factory","name":"顶级韩漫风格","kind":"base","version":1,"revision":1,"status":"published","body":"韩漫画风正文","protocolLock":{"format":"style","slot":"batch.comic-style"}}
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
		"presetId":        "batch-comic-page",
		"imageDerivation": map[string]any{"presetId": "comic-image-derive-default"},
		"style":           map[string]any{"presetId": "batch-comic-style-01"},
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
	derivation := comic["imageDerivation"].(map[string]any)
	if derivation["body"] != "图片推导正文" || derivation["presetSlot"] != "batch.comic-image-derivation" {
		t.Fatalf("comic image derivation was not frozen: %#v", derivation)
	}
	style := comic["style"].(map[string]any)
	if style["body"] != "韩漫画风正文" || style["presetSlot"] != "batch.comic-style" {
		t.Fatalf("comic style was not frozen: %#v", style)
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
