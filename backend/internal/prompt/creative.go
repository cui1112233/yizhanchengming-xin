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
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}
type CreativePromptSnapshot struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Body    string `json:"-"`
	Digest  string `json:"digest"`
}

var creativePromptTypes = map[string]string{CreativeScript144: "script", CreativeStoryboard307: "storyboard", CreativeCharacter213: "character-design", CreativeScene289: "scene-design", "a1_novel_to_screenplay": "script", "a2_novel_to_screenplay": "script", "b_director_screenplay": "script", "storyboard_xiaoluo_v7_15s": "storyboard", "storyboard_xiaoyunque_v3": "storyboard"}

func CreativePromptOptions() []CreativePromptMetadata {
	out := []CreativePromptMetadata{}
	for _, id := range []string{CreativeScript144, CreativeStoryboard307, CreativeCharacter213, CreativeScene289, "a1_novel_to_screenplay", "b_director_screenplay", "storyboard_xiaoluo_v7_15s", "storyboard_xiaoyunque_v3"} {
		p, e := ResolveCreativePrompt(id, creativePromptTypes[id])
		if e == nil {
			out = append(out, CreativePromptMetadata{ID: p.ID, Name: p.Name, Type: p.Type, Version: p.Version, Digest: p.Digest})
		}
	}
	return out
}
func ResolveCreativePrompt(id, purpose string) (CreativePromptSnapshot, error) {
	id = strings.TrimSpace(id)
	kind, ok := creativePromptTypes[id]
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
	return CreativePromptSnapshot{ID: id, Name: entry.Name, Type: kind, Version: digest[:16], Body: body, Digest: digest}, nil
}
