package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

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
	if err != nil || len(all) != 3 || all[2].Status != "draft" {
		t.Fatalf("All() = %+v, %v", all, err)
	}
	texts, err := catalog.ConstraintText([]string{"script-constraint-prefix-a"})
	if err != nil || texts["script-constraint-prefix-a"] != "PREFIX" {
		t.Fatalf("ConstraintText() = %+v, %v", texts, err)
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
