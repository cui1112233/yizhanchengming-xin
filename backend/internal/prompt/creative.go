package prompt

import (
	"crypto/sha256"
	"fmt"
	"qiantie/backend/internal/prompt/luoshui"
	"strings"
)

const (
	CreativeScript144     = "luoshui_professional_screenplay_144"
	CreativeStoryboard307 = "luoshui_storyboard_v9_30s_307"
	CreativeCharacter213  = "luoshui_character_design_213"
	CreativeScene289      = "luoshui_scene_design_289"
)

type CreativePromptMetadata struct {
	CreativePromptContract
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Version      string `json:"version"`
	Digest       string `json:"digest"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	CatalogGroup string `json:"catalog_group,omitempty"`
}
type CreativePromptSnapshot struct {
	CreativePromptContract
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Version      string `json:"version"`
	Body         string `json:"-"`
	Digest       string `json:"digest"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	CatalogGroup string `json:"catalog_group,omitempty"`
}

var creativePromptTypes = map[string]string{CreativeScript144: "script", CreativeStoryboard307: "storyboard", CreativeCharacter213: "character-design", CreativeScene289: "scene-design", "a1_novel_to_screenplay": "script", "a2_novel_to_screenplay": "script", "b_director_screenplay": "script", "storyboard_xiaoluo_v7_15s": "storyboard", "storyboard_xiaoyunque_v3": "storyboard"}

func CreativePromptOptions() []CreativePromptMetadata {
	out := []CreativePromptMetadata{}
	if records, err := RecoveredCreativePromptCatalog(); err == nil {
		for _, r := range records {
			if r.Type != "voice-design" {
				out = append(out, r)
			}
		}
	}
	for _, id := range []string{"a1_novel_to_screenplay", "b_director_screenplay", "storyboard_xiaoluo_v7_15s", "storyboard_xiaoyunque_v3"} {
		p, e := ResolveCreativePrompt(id, creativePromptTypes[id])
		if e == nil {
			out = append(out, CreativePromptMetadata{ID: p.ID, Name: p.Name, Type: p.Type, Version: p.Version, Digest: p.Digest, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, CatalogGroup: p.CatalogGroup, CreativePromptContract: p.CreativePromptContract})
		}
	}
	return out
}
func ResolveCreativePrompt(id, purpose string) (CreativePromptSnapshot, error) {
	id = strings.TrimSpace(id)
	id = CreativePromptAlias(id)
	if p, found, err := resolveRecoveredCreativePrompt(id, purpose); err != nil || found {
		return p, err
	}
	kind, ok := creativePromptTypes[id]
	if !ok {
		if p, found, e := resolveTypedCreativePrompt(id, purpose); e != nil || found {
			return p, e
		}
	}
	if !ok || kind != purpose {
		return CreativePromptSnapshot{}, fmt.Errorf("提示词 %s 不适用于 %s", id, purpose)
	}
	r, e := luoshui.Default()
	if e != nil {
		return CreativePromptSnapshot{}, e
	}
	entry, ok := r.Get(id)
	if !ok {
		return CreativePromptSnapshot{}, fmt.Errorf("提示词 %s 不存在", id)
	}
	body, ok := r.RuntimeBody(id)
	if !ok || strings.TrimSpace(body) == "" {
		return CreativePromptSnapshot{}, fmt.Errorf("提示词 %s 正文为空", id)
	}
	// Catalog baseline normalization must not alter the verified embedded source.
	if strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n")) == strings.TrimSpace(strings.ReplaceAll(entry.Body(), "\r\n", "\n")) {
		body = entry.Body()
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(body)))
	contract := CreativePromptContract{Purpose: kind, NormalizedDigest: fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(body, "\r\n", "\n")))), Source: "existing CM backend corpus", InputSource: "project_content_and_asset_libraries", Variables: []string{}, ModelCapabilities: []string{"text_generation"}, Supported: true}
	switch id {
	case "storyboard_xiaoluo_v7_15s":
		contract.OutputProtocol = "storyboard_video_props_v1"
		contract.ResponseMode = "json_object"
		contract.MaxDurationSec = 15
	case "storyboard_xiaoyunque_v3":
		contract.OutputProtocol = "storyboard_video_props_v1"
		contract.ResponseMode = "json_object"
		contract.MinDurationSec = 10
		contract.MaxDurationSec = 15
	default:
		contract.OutputProtocol = "legacy_script_text_v1"
		contract.ResponseMode = "text"
	}
	contract.ParserVersion = contract.OutputProtocol
	return CreativePromptSnapshot{ID: id, Name: strings.NewReplacer("小洛", "小Y", "洛水", "CM").Replace(entry.Name), Type: kind, Version: digest[:16], Body: body, Digest: digest, CatalogGroup: "compatibility", CreativePromptContract: contract}, nil
}
