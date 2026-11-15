package prompt

import (
	"strings"
	"testing"
)

func TestStoryboard139ExplicitRefreshPinsActualBodyAndLeavesOtherFrozenContracts(t *testing.T) {
	current, err := ResolveCreativePrompt("139", "storyboard")
	if err != nil {
		t.Fatal(err)
	}
	old := current
	old.Supported = false
	old.ParserVersion = "storyboard_image_voice_v1"
	refreshed, ok := RefreshStoryboard139ForExplicitSelection(old)
	if !ok || !refreshed.Supported || refreshed.Body != old.Body || refreshed.Digest != old.Digest || refreshed.ParserVersion != storyboard139OptionalImageParser {
		t.Fatal(refreshed, ok)
	}
	for _, change := range []func(*CreativePromptSnapshot){func(p *CreativePromptSnapshot) { p.Body += "\n" }, func(p *CreativePromptSnapshot) { p.Digest = "changed" }, func(p *CreativePromptSnapshot) { p.SourceID = 138 }, func(p *CreativePromptSnapshot) { p.ID = "other" }} {
		bad := old
		change(&bad)
		got, ok := RefreshStoryboard139ForExplicitSelection(bad)
		if ok || got.Supported {
			t.Fatal("unreviewed frozen body upgraded", ok, got.ID)
		}
	}
	required := strings.Join(current.RequiredFields, ",")
	if strings.Contains(required, "image_prompt") || !strings.Contains(required, "source_elements") || !strings.Contains(required, "video_desc") || !strings.Contains(required, "voice_actor") {
		t.Fatal(required)
	}
	other, _ := ResolveCreativePrompt("138", "storyboard")
	if other.Supported {
		t.Fatal("138 enabled without continuation implementation")
	}
	if _, ok := RefreshStoryboard139ForExplicitSelection(other); ok {
		t.Fatal("138 refreshed")
	}
}
