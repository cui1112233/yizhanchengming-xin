package releasenotes

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestDefaultCatalogContainsCurrentReleaseAndSkipsDraft(t *testing.T) {
	// 从发布指针读取当前 releaseId，而不是写死某个版本：改了发布指针却没写
	// 对应说明时，这条门禁必须直接失败。
	releaseID := ReleaseIDFromManifest(filepath.Join("..", "..", "current-release.json"))
	if releaseID == "" {
		t.Fatal("backend/current-release.json has no releaseId")
	}

	catalog := Default()

	note, ok := catalog.ForReleaseID(releaseID)
	if !ok {
		t.Fatalf("release notes for %q are missing from the embedded catalog", releaseID)
	}
	if note.Title == "" || len(note.Sections) == 0 {
		t.Fatalf("note = %#v, want a usable release note", note)
	}

	if _, draftOK := catalog.ForReleaseID(""); draftOK {
		t.Fatal("a draft with an empty releaseId must never be served")
	}
}

func TestLoadRejectsUnusableNotesAndSanitizesSections(t *testing.T) {
	fsys := fstest.MapFS{
		"notes/empty-release.json": {Data: []byte(`{"releaseId":"","title":"产品更新","sections":[{"type":"fix","label":"修复","items":[{"title":"x"}]}]}`)},
		"notes/no-title.json":      {Data: []byte(`{"releaseId":"r1","title":" ","sections":[{"type":"fix","label":"修复","items":[{"title":"x"}]}]}`)},
		"notes/bad-section.json":   {Data: []byte(`{"releaseId":"r2","title":"产品更新","sections":[{"type":"unknown","label":"其它","items":[{"title":"x"}]}]}`)},
		"notes/ok.json": {Data: []byte(`{
			"releaseId":" r3 ",
			"version":" 3 ",
			"title":" 产品更新 ",
			"sections":[
				{"type":"fix","label":" 修复 ","items":[{"title":" 标题 ","desc":" 说明 ","tags":[" 标签 ",""],"link":" /x "},{"title":" "}]},
				{"type":"feature","label":"","items":[{"title":"空标签"}]}
			]
		}`)},
	}

	catalog, err := Load(fsys, "notes")
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.notes) != 1 {
		t.Fatalf("notes = %#v, want only the usable note", catalog.notes)
	}
	note, ok := catalog.ForReleaseID("r3")
	if !ok {
		t.Fatal("sanitized releaseId r3 is missing")
	}
	if note.Version != "3" || note.Title != "产品更新" {
		t.Fatalf("note = %#v, want trimmed metadata", note)
	}
	if len(note.Sections) != 1 || note.Sections[0].Type != "fix" || note.Sections[0].Label != "修复" {
		t.Fatalf("sections = %#v, want only the usable fix section", note.Sections)
	}
	item := note.Sections[0].Items[0]
	if item.Title != "标题" || item.Desc != "说明" || item.Link != "/x" {
		t.Fatalf("item = %#v, want trimmed values", item)
	}
	if len(item.Tags) != 1 || item.Tags[0] != "标签" {
		t.Fatalf("tags = %#v, want one trimmed tag", item.Tags)
	}
	if len(note.Sections[0].Items) != 1 {
		t.Fatalf("items = %#v, want the empty title to be dropped", note.Sections[0].Items)
	}
}

func TestReleaseIDFromManifestReadsBOMAndMissingFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime-release.json")
	contents := append([]byte{0xef, 0xbb, 0xbf}, []byte(`{"releaseId":"qiantie-test","gitSha":"abc"}`)...)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	if id := ReleaseIDFromManifest(path); id != "qiantie-test" {
		t.Fatalf("releaseId = %q, want qiantie-test", id)
	}
	if id := ReleaseIDFromManifest(filepath.Join(dir, "missing.json")); id != "" {
		t.Fatalf("releaseId = %q, want empty for a missing manifest", id)
	}
	if id := ReleaseIDFromManifest(""); id != "" {
		t.Fatalf("releaseId = %q, want empty for an empty path", id)
	}
}

func TestCurrentReleaseIDPrefersRuntimeManifest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime-release.json")
	if err := os.WriteFile(path, []byte(`{"releaseId":"runtime-release"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if id := CurrentReleaseID(path); id != "runtime-release" {
		t.Fatalf("releaseId = %q, want runtime-release", id)
	}
}
