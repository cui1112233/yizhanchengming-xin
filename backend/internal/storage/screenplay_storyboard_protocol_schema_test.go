package storage

import (
	"strings"
	"testing"
)

func TestStoryboardProtocolMigrationIsRegisteredAndAdditive(t *testing.T) {
	m := screenplayStoryboardProtocolMigration()
	if m.Version != 8800114 || len(m.SQL) != 1 || !strings.Contains(m.SQL[0], "ADD COLUMN protocol_json JSON NULL") {
		t.Fatal(m)
	}
	found := false
	for _, item := range ScreenplayMigrations() {
		if item.Version == m.Version {
			found = true
		}
	}
	if !found {
		t.Fatal("migration missing")
	}
}
