package storage

import (
	"strings"
	"testing"
)

func TestAssetCheckMigrationPersistsIndexedDraftAndExactNames(t *testing.T) {
	m := screenplayAssetCheckMigration()
	if m.Version != 8800115 || len(m.SQL) != 2 {
		t.Fatal(m)
	}
	for _, key := range []string{"PRIMARY KEY(id)", "UNIQUE KEY uk_asset_check_request(owner_username,project_id,request_key)", "state_json JSON NOT NULL", "idx_asset_check_runnable(status,updated_at)", "candidate_count", "applied"} {
		if !strings.Contains(m.SQL[0], key) {
			t.Fatal(key)
		}
	}
	if !strings.Contains(m.SQL[1], "COLLATE utf8mb4_bin") || strings.Contains(m.SQL[1], "UPDATE") || strings.Contains(m.SQL[1], "DELETE") {
		t.Fatal(m.SQL[1])
	}
	found := false
	for _, x := range ScreenplayMigrations() {
		if x.Version == 8800115 {
			found = true
		}
	}
	if !found {
		t.Fatal("not registered")
	}
}
