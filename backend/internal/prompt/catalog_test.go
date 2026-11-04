package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogUsesAdministratorPreviewAssetForComicMaterial(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[{
		"id":"custom-comic-scene","module":"batch-factory","name":"自定义场景","kind":"base","version":1,"revision":1,"status":"published","body":"场景隐藏词","previewAssetId":"asset_custom_scene","protocolLock":{"slot":"batch.comic-reference-scene"}
	}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	items, err := NewCatalog(path).Published("batch-factory")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].PreviewURL != "/api/presets/comic-material/custom-comic-scene/preview" || items[0].ReferenceURL != "/api/local-executor/v1/batch-comic-materials/custom-comic-scene" {
		t.Fatalf("custom material URLs = %#v", items)
	}
}

func TestCatalogListsPublishedPresetsAndConstraintText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	data := `[{"id":"base","module":"script","name":"Base","kind":"base","description":"","compatibleBaseIds":[],"version":1,"revision":1,"status":"published"},{"id":"script-constraint-prefix-a","module":"script","name":"A","kind":"addon","description":"","compatibleBaseIds":["base"],"version":1,"status":"published","body":"PREFIX","protocolLock":{"format":"constraint","category":"prefix"}},{"id":"draft","module":"script","name":"Draft","kind":"base","description":"","compatibleBaseIds":[],"version":1,"status":"draft"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalog(path)
	items, err := catalog.Published("script")
	if err != nil || len(items) != 2 {
		t.Fatalf("Published() = %+v, %v", items, err)
	}
	all, err := catalog.All("script")
	hasDraft := false
	for _, item := range all {
		if item.ID == "draft" && item.Status == "draft" {
			hasDraft = true
		}
	}
	if err != nil || len(all) != 3 || !hasDraft {
		t.Fatalf("All() = %+v, %v", all, err)
	}
	texts, err := catalog.ConstraintText([]string{"script-constraint-prefix-a"})
	if err != nil || texts["script-constraint-prefix-a"] != "PREFIX" {
		t.Fatalf("ConstraintText() = %+v, %v", texts, err)
	}
}

func TestCatalogReloadsWhenPublishedFileChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	first := `[{"id":"base","module":"script","name":"第一版","kind":"base","compatibleBaseIds":[],"version":1,"status":"published"}]`
	if err := os.WriteFile(path, []byte(first), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalog(path)
	items, err := catalog.Published("script")
	if err != nil || len(items) != 1 || items[0].Name != "第一版" {
		t.Fatalf("first Published() = %+v, %v", items, err)
	}
	second := `[{"id":"base","module":"script","name":"第二版","kind":"base","compatibleBaseIds":[],"version":1,"status":"published"}]`
	if err := os.WriteFile(path, []byte(second), 0o600); err != nil {
		t.Fatal(err)
	}
	items, err = catalog.Published("script")
	if err != nil || len(items) != 1 || items[0].Name != "第二版" {
		t.Fatalf("cached Published() did not reload: %+v, %v", items, err)
	}
}

func TestCatalogCacheIsNotExposedToCallers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	data := `[{"id":"base","module":"script","name":"Base","kind":"base","compatibleBaseIds":["origin"],"categories":["cat"],"version":1,"status":"published"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog := NewCatalog(path)
	records, err := catalog.load()
	if err != nil || len(records) != 1 {
		t.Fatalf("load() = %+v, %v", records, err)
	}
	records[0].CompatibleBaseIDs[0] = "mutated"
	records[0].Categories[0] = "mutated"
	reloaded, err := catalog.load()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded[0].CompatibleBaseIDs[0] != "origin" || reloaded[0].Categories[0] != "cat" {
		t.Fatalf("cache was mutated through a caller copy: %+v", reloaded[0])
	}
}

func TestCatalogResolveRequiresCompatibleBase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "presets.json")
	data := `[{"id":"base","module":"script","name":"Base","kind":"base","compatibleBaseIds":[],"version":1,"status":"published"},{"id":"addon","module":"script","name":"Addon","kind":"addon","compatibleBaseIds":["other"],"version":1,"status":"published"}]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCatalog(path).Resolve("script", []string{"base", "addon"}); err == nil {
		t.Fatal("Resolve accepted an incompatible addon")
	}
}
