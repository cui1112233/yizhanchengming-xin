package storage

import (
	"strings"
	"testing"
)

func TestScreenplayProjectConfigMigrationUsesCompatibleAtomicAlter(t *testing.T) {
	var migration Migration
	for _, candidate := range ScreenplayMigrations() {
		if candidate.Version == 8800103 {
			migration = candidate
			break
		}
	}
	if migration.Version == 0 {
		t.Fatal("missing screenplay project config migration")
	}
	if len(migration.SQL) != 1 {
		t.Fatalf("project config migration must use one atomic ALTER, got %d statements", len(migration.SQL))
	}
	statement := migration.SQL[0]
	if strings.Contains(strings.ToUpper(statement), "ADD COLUMN IF NOT EXISTS") {
		t.Fatalf("current MySQL rejects ADD COLUMN IF NOT EXISTS: %s", statement)
	}
	for _, column := range []string{"model_id", "prompt_id", "asset_degree"} {
		if !strings.Contains(statement, column) {
			t.Fatalf("atomic ALTER missing %s: %s", column, statement)
		}
	}
}
