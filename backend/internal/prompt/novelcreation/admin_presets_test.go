package novelcreation

import (
	"os"
	"path/filepath"
	catalog "qiantie/backend/internal/prompt"
	"testing"
)

func TestNovelAdminPresetsSeedAndPublishedSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	_ = os.WriteFile(path, []byte(`[]`), 0600)
	if e := EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	if e := EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	store := catalog.NewAdminStore(path)
	rows, e := store.List(AdminModule)
	if e != nil || len(rows) != 24 {
		t.Fatalf("expected24 active prompt entries, got%d %v", len(rows), e)
	}
	bodies, e := catalog.NewCatalog(path).PublishedBodies(AdminModule)
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range []string{NativeRevisionPresetID, PatchOnlyPresetID, RewriteOnlyPresetID, IncrementalSettlementPresetID, ChapterDraft.ID} {
		found := false
		for _, v := range bodies {
			if v.ID == id && v.Body != "" {
				found = true
			}
		}
		if !found {
			t.Fatal("missing actual runtime prompt", id)
		}
	}
	if _, e = store.Archive("admin", ChapterDraft.ID, 1); e != nil {
		t.Fatal(e)
	}
	if e = EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	if _, found, e := catalog.NewCatalog(path).PublishedBody(AdminModule, ChapterDraft.ID); e != nil || found {
		t.Fatal("restart re-enabled archived preset", e)
	}
}
