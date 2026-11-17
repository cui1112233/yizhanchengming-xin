package prompt

import "testing"

func TestZIP164CatalogAndExactBody(t *testing.T) {
	p, err := ResolveCreativePrompt("164", "storyboard")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "luoshui_storyboard_164" || p.SourceID != 164 || p.CatalogGroup != "zip-original" || !p.Supported {
		t.Fatalf("unexpected metadata: %+v", p)
	}
	if p.Digest != "b3ea2ac397a09cee548464e5834d0c1dafcb2439649ae642fbe8547627dcf328" {
		t.Fatalf("body changed: %s", p.Digest)
	}
	if p.MinDurationSec != 0 || p.MaxDurationSec != 0 || p.FixedDurationSec != 0 {
		t.Fatal("original has no fixed duration")
	}
	if p.CreatedAt != "2026-03-14 17:59:47" || p.UpdatedAt != p.CreatedAt {
		t.Fatal("original dates lost")
	}
	n := 0
	for _, option := range CreativePromptOptions() {
		if option.SourceID == 164 {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("visible count = %d", n)
	}
	if _, err := ResolveCreativePrompt(p.ID, "script"); err == nil {
		t.Fatal("wrong purpose accepted")
	}
}
