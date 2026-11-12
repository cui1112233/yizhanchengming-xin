package storage

import (
	"strings"
	"testing"
)

func TestAICutAssetFolderMigrationScopesMembershipAndPreservesMedia(t *testing.T) {
	migrations := AICutAssetFolderMigrations()
	if len(migrations) != 1 || migrations[0].Version != 8800033 {
		t.Fatalf("migration=%+v", migrations)
	}
	sql := strings.Join(migrations[0].SQL, "\n")
	for _, want := range []string{"PRIMARY KEY (asset_id, project_id, owner_username)", "REFERENCES aicut_assets(id, project_id, owner_username) ON DELETE CASCADE", "REFERENCES aicut_asset_folders(id, project_id, owner_username) ON DELETE CASCADE"} {
		if !strings.Contains(sql, want) {
			t.Fatalf("missing %s", want)
		}
	}
	for _, forbidden := range []string{"DELETE FROM aicut_assets", "DROP TABLE", "ALTER TABLE aicut_assets"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("unsafe migration %s", forbidden)
		}
	}
}
