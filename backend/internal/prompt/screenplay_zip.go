package prompt

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"strings"
)

// Supplement from the supplied ZIP; the older 63-row recovered corpus remains frozen.
//
//go:embed zip/prompt-164.txt
var zipStoryboard164Body string

const CreativeStoryboard164 = "luoshui_storyboard_164"

func resolveZIPCreativePrompt(id, purpose string) (CreativePromptSnapshot, bool, error) {
	if id != CreativeStoryboard164 {
		return CreativePromptSnapshot{}, false, nil
	}
	if purpose != "storyboard" {
		return CreativePromptSnapshot{}, true, fmt.Errorf("提示词 %s 不适用于 %s", id, purpose)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(zipStoryboard164Body)))
	if digest != "b3ea2ac397a09cee548464e5834d0c1dafcb2439649ae642fbe8547627dcf328" {
		return CreativePromptSnapshot{}, true, fmt.Errorf("ZIP 164号提示词正文校验失败")
	}
	return CreativePromptSnapshot{
		ID: id, Name: "Vidu-Q3多4宫格q2、q3通用-3月12日", Type: "storyboard", Body: zipStoryboard164Body, Digest: digest, Version: digest[:16],
		CreatedAt: "2026-03-14 17:59:47", UpdatedAt: "2026-03-14 17:59:47", CatalogGroup: "zip-original",
		CreativePromptContract: CreativePromptContract{
			SourceID: 164, Purpose: purpose, NormalizedDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(zipStoryboard164Body, "\r\n", "\n")))),
			Source: "supplied 洛水2026.zip / data/app.db / prompts.id=164", InputSource: "project_content_and_asset_libraries",
			Variables:      []string{"ROLES_PLACEHOLDER", "SCENES_PLACEHOLDER", "PROPS_PLACEHOLDER", "剧本内容", "前场景末尾/后场景开头分镜"},
			OutputProtocol: "storyboard_four_grid_voice_v1", ResponseMode: "json_object", ParserVersion: "cm_zip164_v1", ModelCapabilities: []string{"text_generation"}, Supported: true,
			RequiredFields: []string{"id", "source_elements", "characters", "scene", "image_prompt", "voice_actor", "emotion", "line", "video_desc"},
		},
	}, true, nil
}
