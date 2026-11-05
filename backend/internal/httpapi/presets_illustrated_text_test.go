package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/prompt"
)

func seededCatalog(t *testing.T) *prompt.Catalog {
	t.Helper()
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	return prompt.NewCatalog(path)
}

func bodyDigest(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

func TestIllustratedTextComicPagePresetIsPublishedInCatalog(t *testing.T) {
	catalog := seededCatalog(t)
	presets, err := catalog.Published("batch-factory")
	if err != nil {
		t.Fatal(err)
	}
	var found *prompt.PublicPreset
	for index := range presets {
		if presets[index].ID == "batch-comic-page-illustrated-text" {
			found = &presets[index]
			break
		}
	}
	if found == nil {
		t.Fatal("catalog 缺少 batch-comic-page-illustrated-text")
	}
	if found.Name != "CM·文字插画版" {
		t.Fatalf("name = %q", found.Name)
	}
	if found.Version != 1 || found.Revision != 1 {
		t.Fatalf("version/revision = %d/%d", found.Version, found.Revision)
	}
	if found.Status != "published" {
		t.Fatalf("status = %q", found.Status)
	}
	if found.Slot == nil || *found.Slot != "batch.comic-page" {
		t.Fatalf("slot = %v", found.Slot)
	}
}

func TestIllustratedTextComicPagePresetBodyConstraints(t *testing.T) {
	catalog := seededCatalog(t)
	preset, found, err := catalog.PublishedBody("batch-factory", "batch-comic-page-illustrated-text")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("published body not found")
	}
	for _, required := range []string{"（未完见下页）", "最多 3 处", "不写任何画风名", "不出现任何文字"} {
		if !strings.Contains(preset.Body, required) {
			t.Fatalf("body missing %q", required)
		}
	}
	for _, forbidden := range []string{"水彩", "彩铅", "赛璐璐", "韩漫", "webtoon", "3:4", "9:16", "16:9", "4:3", "1440×1920"} {
		if strings.Contains(strings.ToLower(preset.Body), strings.ToLower(forbidden)) {
			t.Fatalf("body must not hardcode %q", forbidden)
		}
	}
}

func TestIllustratedTextPresetKeepsExistingComicPagePresetsUnchanged(t *testing.T) {
	snapshots := map[string]struct {
		name   string
		digest string
	}{
		"batch-comic-page-premium-strip": {"CM·精品商业条漫版", "a593e5c8bd327622ea958e88e8fd9cbbf4e1931af80535a7f70937f876d3bef5"},
		"batch-comic-page-fusion-bw":     {"CM·剧情漫画融合白框版", "1998f506b5d3d6bf192894cec130089d5cd03825ffc3cbed29990e6fb5eb0420"},
		"batch-comic-page":               {"CM·剧情漫画融合彩色版（默认）", "2a296e90f2ddb761b98935b02aedf64f75a397a4c1d6ddfe20911d2751f30f32"},
		"batch-comic-page-light":         {"CM·轻量分镜版", "ce5a82b3b1b0b1b6407a7ebdeb76742cf9f5376cbbfc807d46421f41d21b9bc7"},
	}
	catalog := seededCatalog(t)
	for id, snapshot := range snapshots {
		preset, found, err := catalog.PublishedBody("batch-factory", id)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatalf("%s published body not found", id)
		}
		if preset.Version != 3 {
			t.Fatalf("%s version = %d", id, preset.Version)
		}
		if preset.Name != snapshot.name {
			t.Fatalf("%s name = %q", id, preset.Name)
		}
		if preset.Slot != "batch.comic-page" {
			t.Fatalf("%s slot = %q", id, preset.Slot)
		}
		if got := bodyDigest(preset.Body); got != snapshot.digest {
			t.Fatalf("%s body changed: %s", id, got)
		}
	}
}
