package storage

import (
	"sort"
	"strings"
	"testing"
)

func TestNovelContentMigrationsHaveUniqueOrderedDependencies(t *testing.T) {
	all := AppMigrations()
	sort.Slice(all, func(i, j int) bool { return all[i].Version < all[j].Version })
	seen := map[int]bool{}
	for _, m := range all {
		if seen[m.Version] {
			t.Fatalf("duplicate migration %d", m.Version)
		}
		seen[m.Version] = true
		if m.Version == 8800314 {
			joined := strings.Join(m.SQL, "\n")
			if !strings.Contains(joined, "novel_content_records") || !strings.Contains(joined, "archived_at") || strings.Contains(joined, "ALTER TABLE project_retention") {
				t.Fatal("content migration dependencies invalid")
			}
		}
	}
	if !seen[8800400] || !seen[8800401] || !seen[8800314] {
		t.Fatal("required migrations absent")
	}
	retention := ProjectRetentionMigrations()
	last := retention[len(retention)-1]
	if last.Version != 8800401 || !strings.Contains(strings.Join(last.SQL, "\n"), "manual_archived") {
		t.Fatal("manual archive marker missing")
	}
}
