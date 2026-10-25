package prompt

import (
	"bytes"
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
		id, module, slot, category, phrase string
	}{
		{"batch-comic-assets", "batch-factory", "batch.comic-assets", "", "二点五至三点五头身"},
		{"script-constraint-comic-quality", "script", "", "comic_quality", "{{image_aspect_ratio}}"},
		{"script-constraint-comic-restriction", "script", "", "comic_restriction", "三格版式"},
		{"script-constraint-comic-negative", "script", "", "comic_negative", "成人写实比例"},
		{"batch-comic-page", "batch-factory", "batch.comic-page", "", "本页版式"},
		{"batch-comic-page-fusion-bw", "batch-factory", "batch.comic-page", "", "白色外边框"},
		{"batch-comic-page-light", "batch-factory", "batch.comic-page", "", "轻量"},
		{"batch-comic-page-premium-strip", "batch-factory", "batch.comic-page", "", "商业条漫"},
		{"comic-image-derive-default", "batch-factory", "batch.comic-image-derivation", "", "图片模型"},
		{"vip-smart-character-derive", "batch-factory", "batch.comic-assets", "", "角色设定"},
	}
	for _, test := range tests {
		preset, found, err := catalog.PublishedBody(test.module, test.id)
		if err != nil || !found {
			t.Fatalf("%s is not published: found=%v err=%v", test.id, found, err)
		}
		if preset.Slot != test.slot || preset.ConstraintCategory != test.category || !strings.Contains(preset.Body, test.phrase) {
			t.Fatalf("%s has the wrong system-preset contract: %#v", test.id, preset)
		}
	}
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
