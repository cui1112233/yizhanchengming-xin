package prompt

import (
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
	if err != nil || !found || !strings.Contains(preset.Body, "3:4") {
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
		{"script-constraint-comic-quality", "script", "", "comic_quality", "1440×1920"},
		{"script-constraint-comic-restriction", "script", "", "comic_restriction", "三格版式"},
		{"script-constraint-comic-negative", "script", "", "comic_negative", "成人写实比例"},
		{"batch-comic-page", "batch-factory", "batch.comic-page", "", "本页版式"},
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
