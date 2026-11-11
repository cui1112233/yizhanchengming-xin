package storage

import (
	"strings"
	"testing"
)

func TestV11DirectorMigrationIsAdditiveAndOrdered(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 4 || migrations[3].Version != 1100004 {
		t.Fatalf("migrations=%+v", migrations)
	}
	joined := strings.ToLower(strings.Join(V11DirectorStatements(), "\n"))
	for _, table := range []string{
		"batch_factory_v11_hook_revisions",
		"batch_factory_v11_director_revisions",
		"batch_factory_v11_director_video_links",
		"batch_factory_v11_orphaned_overrides",
	} {
		if !strings.Contains(joined, table) {
			t.Fatalf("missing %s", table)
		}
	}
	if strings.Contains(joined, "drop table") || strings.Contains(joined, "delete from") {
		t.Fatal("Director migration must be additive")
	}
}
