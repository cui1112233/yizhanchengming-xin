package storage

import (
	"strings"
	"testing"
)

func TestOrdinaryAssetLibraryTombstoneMigrationIsAdditiveAndRegistered(t *testing.T) {
	m := screenplayAssetLibraryTombstoneMigration()
	if m.Version != 8800117 || len(m.SQL) != 1 || !strings.Contains(m.SQL[0], "ADD COLUMN deleted_at DATETIME(6) NULL") || !strings.Contains(m.SQL[0], "idx_screenplay_library_active") {
		t.Fatal(m)
	}
	for _, bad := range []string{"DROP TABLE", "DELETE FROM", "TRUNCATE"} {
		if strings.Contains(m.SQL[0], bad) {
			t.Fatal(m.SQL)
		}
	}
	found := false
	for _, x := range ScreenplayMigrations() {
		if x.Version == m.Version {
			found = true
		}
	}
	if !found {
		t.Fatal("tombstone migration missing")
	}
}
