package storage

import (
	"strings"
	"testing"
)

func TestScreenplayRawResultMigrationKeepsExistingRowsCompatible(t *testing.T) {
	m := screenplayRawResultMigration()
	if m.Version != 8800110 || len(m.SQL) != 1 || !strings.Contains(m.SQL[0], "raw_result LONGTEXT NULL") {
		t.Fatalf("migration=%+v", m)
	}
	found := false
	for _, item := range ScreenplayMigrations() {
		if item.Version == m.Version {
			found = true
		}
	}
	if !found {
		t.Fatal("raw result migration not registered")
	}
}
