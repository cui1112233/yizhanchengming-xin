package prompt

import (
	"crypto/sha256"
	"fmt"
)

const storyboard139OptionalImageDigest = "ac721cf64535af752dcfc841a80f497db9a31268dee0cc7dd74cdb21bc1d37ca"
const storyboard139OptionalImageParser = "cm_storyboard_139_optional_image_v2"

// The frozen 63-body catalog remains an evidence record. This independent CM
// adapter follows recovered generate/import/editor consumers for this exact
// body: omitted or empty image_prompt stays empty; it is never synthesized from
// video_desc. The native parser has not been recovered in full.
func applyRecoveredProtocolOverrides(m *CreativePromptMetadata) error {
	if m.SourceID != 139 {
		return nil
	}
	if m.ID != "luoshui_storyboard_139" || m.Digest != storyboard139OptionalImageDigest {
		return fmt.Errorf("139协议适配不匹配已核验正文摘要")
	}
	fields := make([]string, 0, len(m.RequiredFields))
	for _, field := range m.RequiredFields {
		if field != "image_prompt" {
			fields = append(fields, field)
		}
	}
	m.RequiredFields = fields
	m.Supported = true
	m.UnsupportedReason = ""
	m.ParserVersion = storyboard139OptionalImageParser
	m.Source += "; CM independent optional-image adapter from recovered generation/import/editor consumers; native parser not fully recovered"
	return nil
}

func StoryboardAllowsMissingImagePrompt(p CreativePromptSnapshot) bool {
	return p.SourceID == 139 && p.ID == "luoshui_storyboard_139" && p.Digest == storyboard139OptionalImageDigest && p.ParserVersion == storyboard139OptionalImageParser && p.Supported
}

// RefreshStoryboard139ForExplicitSelection is only for an explicit config PUT.
// It never replaces frozen text, and is never called by generation or GET.
func RefreshStoryboard139ForExplicitSelection(frozen CreativePromptSnapshot) (CreativePromptSnapshot, bool) {
	if frozen.ID != "luoshui_storyboard_139" || frozen.Type != "storyboard" || (frozen.SourceID != 0 && frozen.SourceID != 139) || frozen.Digest != storyboard139OptionalImageDigest || fmt.Sprintf("%x", sha256.Sum256([]byte(frozen.Body))) != storyboard139OptionalImageDigest {
		return frozen, false
	}
	current, err := ResolveCreativePrompt(frozen.ID, "storyboard")
	if err != nil || !StoryboardAllowsMissingImagePrompt(current) {
		return frozen, false
	}
	frozen.CreativePromptContract = current.CreativePromptContract
	return frozen, true
}
