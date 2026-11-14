package storage

import (
	"strings"
	"testing"
)

func TestOrdinaryAssetLibraryMigrationIsScopedAndAllowsSameName(t *testing.T) {
	m := screenplayAssetLibraryMigration()
	if m.Version != 8800116 || len(m.SQL) != 1 {
		t.Fatal(m)
	}
	ddl := m.SQL[0]
	for _, part := range []string{"PRIMARY KEY(id)", "owner_username", "asset_type", "image_json JSON NOT NULL", "source_media_id VARCHAR(64)", "image_object_json JSON NOT NULL", "snapshot_json JSON NOT NULL", "revision BIGINT", "idx_screenplay_library_owner_type", "idx_screenplay_library_category"} {
		if !strings.Contains(ddl, part) {
			t.Fatal(part)
		}
	}
	if strings.Contains(ddl, "UNIQUE KEY") {
		t.Fatal("names must not gain an invented overwrite/unique policy")
	}
	found := false
	for _, x := range ScreenplayMigrations() {
		if x.Version == m.Version {
			found = true
		}
	}
	if !found {
		t.Fatal("migration not registered")
	}
}
