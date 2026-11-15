package storage

import (
	"strings"
	"testing"
)

func TestScreenplayAssetLibraryCategoriesMigration(t *testing.T) {
	found := 0
	for _, m := range ScreenplayMigrations() {
		if m.Version == 8800118 {
			found++
			if len(m.SQL) != 1 || !strings.Contains(m.SQL[0], "CREATE TABLE IF NOT EXISTS screenplay_asset_library_categories") || !strings.Contains(m.SQL[0], "PRIMARY KEY(owner_username)") || !strings.Contains(m.SQL[0], "state_json JSON NOT NULL") {
				t.Fatal(m)
			}
		}
	}
	if found != 1 {
		t.Fatal("category migration must be registered once", found)
	}
}
