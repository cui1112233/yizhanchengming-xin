package prompt

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureComicPresetsSeedsOnlyMissingRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[{"id":"existing","version":1}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalog(path)
	preset, found, err := catalog.PublishedBody("batch-factory", "batch-comic-page")
	if err != nil || !found || !strings.Contains(preset.Body, "{{image_aspect_ratio}}") {
		t.Fatalf("comic page preset was not seeded: %#v, found=%v, err=%v", preset, found, err)
	}
	if err := EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	if _, found, err := catalog.PublishedBody("batch-factory", "existing"); err != nil || found {
		t.Fatalf("unrelated record should stay untouched: found=%v err=%v", found, err)
	}
}

func TestSystemComicPresetsArePublishedAndUsable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalog(path)
	tests := []struct {
		id, name, module, slot, category, phrase string
	}{
		{"batch-comic-assets", "CM·漫画资产统一推导", "batch-factory", "batch.comic-assets", "", "年龄段变体拆分"},
		{"vip-smart-character-derive", "CM·VIP 智能角色推导", "batch-factory", "batch.comic.assets.character", "", "首句融合规则"},
		{"batch-comic-assets-cm-scene", "CM·主要场景推导", "batch-factory", "batch.comic.assets.scene", "", "昼夜/时段强制锚定"},
		{"batch-comic-assets-cm-prop", "CM·关键道具推导", "batch-factory", "batch.comic.assets.prop", "", "无则空数组"},
		{"batch-comic-character-three-view-cm", "CM·角色三视图", "batch-factory", "batch.comic.character-image-template", "", "正面近照、正面全身、侧面全身、背面全身"},
		{"script-constraint-comic-quality", "漫画画质约束", "script", "", "comic_quality", "{{image_aspect_ratio}}"},
		{"script-constraint-comic-restriction", "漫画画面限制", "script", "", "comic_restriction", "三格版式"},
		{"script-constraint-comic-negative", "漫画负面提示词", "script", "", "comic_negative", "成人写实比例"},
		{"batch-comic-page", "漫画提示词", "batch-factory", "batch.comic-page", "", "本页版式"},
		{"batch-comic-page-fusion-bw", "剧情漫画融合白框版", "batch-factory", "batch.comic-page", "", "白色外边框"},
		{"batch-comic-page-light", "轻量分镜版", "batch-factory", "batch.comic-page", "", "轻量"},
		{"batch-comic-page-premium-strip", "精品商业条漫版", "batch-factory", "batch.comic-page", "", "商业条漫"},
		{"comic-image-derive-default", "图片描述词推导", "batch-factory", "batch.comic-image-derivation", "", "图片模型"},
	}
	for _, test := range tests {
		preset, found, err := catalog.PublishedBody(test.module, test.id)
		if err != nil || !found {
			t.Fatalf("%s is not published: found=%v err=%v", test.id, found, err)
		}
		if preset.Name != test.name || preset.Slot != test.slot || preset.ConstraintCategory != test.category || !strings.Contains(preset.Body, test.phrase) || strings.Contains(preset.Name, "雪漫画·") {
			t.Fatalf("%s has the wrong system-preset contract: %#v", test.id, preset)
		}
	}
}

func TestEnsureComicPresetsUpgradesRecognizedLegacySystemSeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	legacy := `[
		{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产提示词","kind":"base","description":"旧版","compatibleBaseIds":[],"version":2,"revision":2,"status":"published","body":"旧漫画资产正文","protocolLock":{"format":"json","operation":"comic-assets","slot":"batch.comic-assets"},"createdBy":"system"},
		{"id":"vip-smart-character-derive","module":"batch-factory","name":"VIP 智能角色推导","kind":"base","description":"旧版","compatibleBaseIds":[],"version":1,"revision":1,"status":"published","body":"旧角色正文","protocolLock":{"format":"json","operation":"comic-assets","slot":"batch.comic-assets"},"createdBy":"system"}
	]`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	records := readComicPresetRecords(t, path)
	assertComicPresetVersion(t, records, "batch-comic-assets", 2, "archived", "漫画资产提示词", "旧漫画资产正文")
	assertComicPresetVersion(t, records, "batch-comic-assets", 3, "published", "CM·漫画资产统一推导", "")
	assertComicPresetVersion(t, records, "vip-smart-character-derive", 1, "archived", "VIP 智能角色推导", "旧角色正文")
	assertComicPresetVersion(t, records, "vip-smart-character-derive", 2, "published", "CM·VIP 智能角色推导", "")
}

func TestEnsureComicPresetsPreservesAdministratorPublishedVersions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	custom := `[{
		"id":"batch-comic-assets","module":"batch-factory","name":"管理员漫画资产提示词","kind":"base","description":"自定义","compatibleBaseIds":[],"version":3,"revision":1,"status":"published","body":"管理员自定义正文","protocolLock":{"format":"json","operation":"comic-assets","slot":"batch.comic-assets"},"createdBy":"admin","publishedBy":"admin"
	}]`
	if err := os.WriteFile(path, []byte(custom), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	records := readComicPresetRecords(t, path)
	assertComicPresetVersion(t, records, "batch-comic-assets", 3, "published", "管理员漫画资产提示词", "管理员自定义正文")
	published := 0
	for _, record := range records {
		if stringValue(record["id"]) == "batch-comic-assets" && stringValue(record["status"]) == "published" {
			published++
		}
	}
	if published != 1 {
		t.Fatalf("administrator preset published count = %d, want 1", published)
	}
}

func readComicPresetRecords(t *testing.T, path string) []map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	if err := json.Unmarshal(raw, &records); err != nil {
		t.Fatal(err)
	}
	return records
}

func assertComicPresetVersion(t *testing.T, records []map[string]any, id string, version int, status, name, body string) {
	t.Helper()
	for _, record := range records {
		if stringValue(record["id"]) != id || intValue(record["version"]) != version {
			continue
		}
		if stringValue(record["status"]) != status || stringValue(record["name"]) != name {
			t.Fatalf("%s V%d metadata = %#v", id, version, record)
		}
		if body != "" && stringValue(record["body"]) != body {
			t.Fatalf("%s V%d body = %q, want %q", id, version, stringValue(record["body"]), body)
		}
		return
	}
	t.Fatalf("%s V%d not found", id, version)
}

func TestSystemComicStylePresetsExposeTwentyFivePreviewCards(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalog(path)
	items, err := catalog.Published("batch-factory")
	if err != nil {
		t.Fatal(err)
	}
	styles := make([]PublicPreset, 0, 25)
	for _, item := range items {
		if item.Slot != nil && *item.Slot == "batch.comic-style" {
			styles = append(styles, item)
		}
	}
	if len(styles) != 25 {
		t.Fatalf("comic style count = %d, want 25", len(styles))
	}
	seen := map[string]bool{}
	for _, item := range styles {
		if seen[item.ID] {
			t.Fatalf("duplicate comic style id %q", item.ID)
		}
		seen[item.ID] = true
		if item.Name == "" || item.PreviewURL == "" {
			t.Fatalf("comic style metadata incomplete: %#v", item)
		}
		preset, found, err := catalog.PublishedBody("batch-factory", item.ID)
		if err != nil || !found || strings.TrimSpace(preset.Body) == "" {
			t.Fatalf("comic style body unavailable for %q: found=%v err=%v", item.ID, found, err)
		}
		preview, contentType, found := OpenComicStylePreview(item.ID)
		if !found || contentType != "image/jpeg" || len(preview) < 4 || !bytes.Equal(preview[:2], []byte{0xff, 0xd8}) {
			t.Fatalf("comic style preview invalid for %q: found=%v type=%q bytes=%d", item.ID, found, contentType, len(preview))
		}
	}
}
