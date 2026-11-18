package storage

import "testing"

func TestNovelChapterImportReceiptMigrationIsAppendOnly(t *testing.T) {
	ms := NovelCreationMigrations()
	if ms[len(ms)-2].Version != 8800315 || ms[len(ms)-1].Version != 8800316 {
		t.Fatal("chapter import receipt migration missing")
	}
	seen := map[int]bool{}
	for _, m := range ms {
		if seen[m.Version] {
			t.Fatalf("duplicate migration %d", m.Version)
		}
		seen[m.Version] = true
	}
}
