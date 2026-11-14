package prompt

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestScreenplayModelConfigDefaultsMatchOriginalAgaAndZga(t *testing.T) {
	for name, pair := range map[string][2]string{"Aga": {ScreenplayImageCoreDefault, "89f6ed5e36c084b4bd96f744c667d03c43cddb0fdc3e5a967f81e83c156f86a3"}, "zga": {ScreenplayAssetProcessDefault, "8290758dbf6d1df0bf75aac79b45fd80aaf9d8e847076bf2e60296de21e6d3e3"}} {
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(pair[0]))); got != pair[1] {
			t.Fatal(name, got)
		}
	}
}

func TestScreenplayTypedCatalogHasAllOriginal12And6AndStrictPurposes(t *testing.T) {
	counts := map[string]int{}
	seen := map[string]bool{}
	for _, p := range CreativePromptOptions() {
		if seen[p.ID] {
			t.Fatal("duplicate option", p.ID)
		}
		seen[p.ID] = true
		counts[p.Type]++
	}
	if counts["character-design"] != 12 || counts["scene-design"] != 6 {
		t.Fatal(counts)
	}
	for _, id := range []string{"212", "213", "214", "219", "249", "250", "285", "290", "294", "298", "299", "300"} {
		p, e := ResolveCreativePrompt(id, "character-design")
		if e != nil || p.Body == "" || len(p.Digest) != 64 {
			t.Fatal(id, p, e)
		}
		if _, e = ResolveCreativePrompt(id, "scene-design"); e == nil {
			t.Fatal("character accepted for scene", id)
		}
	}
	for _, id := range []string{"251", "286", "287", "288", "289", "295"} {
		p, e := ResolveCreativePrompt(id, "scene-design")
		if e != nil || p.Body == "" || len(p.Digest) != 64 {
			t.Fatal(id, p, e)
		}
		if _, e = ResolveCreativePrompt(id, "character-design"); e == nil {
			t.Fatal("scene accepted for character", id)
		}
	}
}
