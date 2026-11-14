package screenplay

import (
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
)

var CreativeAspectRatios = []string{"9:16", "16:9", "1:1", "3:4", "4:3", "2:3", "3:2", "21:9", "4:5", "5:4"}

func ValidCreativeAspectRatio(value string) bool {
	for _, v := range CreativeAspectRatios {
		if value == v {
			return true
		}
	}
	return false
}

// EffectiveCreativeConfig supplies absent legacy metadata without refilling a saved empty prompt.
func EffectiveCreativeConfig(p Project) CreativeConfig {
	c := CreativeConfig{ConfigVersion: 1, ScriptPromptID: p.PromptID, ImageResolution: "2K", AssetPromptType: "general", VoiceDesignModel: "minimax", ShowOriginalText: true, CharacterImageMode: "typedprompt:" + prompt.CreativePromptAlias("212"), SceneImageMode: "typedprompt:" + prompt.CreativeScene289, CharacterDesignPromptID: prompt.CreativePromptAlias("212"), SceneDesignPromptID: prompt.CreativeScene289, ImageCoreInstruction: prompt.ScreenplayImageCoreDefault, AssetProcessPrompt: prompt.ScreenplayAssetProcessDefault}
	if p.CreativeConfig != nil {
		c = *cloneCreativeConfig(p.CreativeConfig)
		if c.ConfigVersion == 0 {
			c.ShowOriginalText = true
			// This field did not exist before version 1. The legacy decoder has no
			// field-presence bit, so version 0 + empty + not-cleared means absent.
			// Never apply this inference to the older image_core_instruction field.
			if c.AssetProcessPrompt == "" && !c.AssetProcessPromptCleared {
				c.AssetProcessPrompt = prompt.ScreenplayAssetProcessDefault
			}
		}
		if c.ImageResolution == "" {
			c.ImageResolution = "2K"
		}
		if c.AssetPromptType == "" {
			c.AssetPromptType = "general"
		}
		if c.VoiceDesignModel == "" {
			c.VoiceDesignModel = "minimax"
		}
		if c.CharacterImageMode == "" {
			if c.CharacterDesignPromptID != "" {
				c.CharacterImageMode = "typedprompt:" + c.CharacterDesignPromptID
			} else {
				c.CharacterImageMode = "single"
			}
		}
		if c.SceneImageMode == "" {
			if c.SceneDesignPromptID != "" {
				c.SceneImageMode = "typedprompt:" + c.SceneDesignPromptID
			} else {
				c.SceneImageMode = "none"
			}
		}
	}
	if p.CreativeConfig == nil && (p.SelectedStyle != "" || p.StyleDesc != "") {
		c.Style = &CreativeStyleSnapshot{ID: p.SelectedStyle, Name: p.ArtStyle, Prompt: p.StyleDesc, Images: []string{}}
	}
	for _, v := range []struct {
		mode    *string
		id      *string
		purpose string
	}{{&c.CharacterImageMode, &c.CharacterDesignPromptID, "character-design"}, {&c.SceneImageMode, &c.SceneDesignPromptID, "scene-design"}} {
		if id := CreativeModePromptID(*v.mode); id != "" {
			if _, e := prompt.ResolveCreativePrompt(id, v.purpose); e == nil {
				*v.mode = "typedprompt:" + id
				*v.id = id
			}
		}
	}
	c.AspectRatio = p.AspectRatio
	if c.AspectRatio == "" {
		c.AspectRatio = "9:16"
	}
	return c
}

func CreativeModePromptID(mode string) string {
	for _, prefix := range []string{"typedprompt:", "prompt:"} {
		if strings.HasPrefix(mode, prefix) {
			return CreativePromptAlias(strings.TrimPrefix(mode, prefix))
		}
	}
	return ""
}

func CreativePromptAlias(id string) string {
	switch strings.TrimSpace(id) {
	case "144":
		return prompt.CreativeScript144
	case "307":
		return prompt.CreativeStoryboard307
	case "213":
		return prompt.CreativeCharacter213
	case "289":
		return prompt.CreativeScene289
	}
	return prompt.CreativePromptAlias(strings.TrimSpace(id))
}

func NormalizeCreativeImageMode(mode, purpose string) (string, string, error) {
	mode = strings.TrimSpace(mode)
	if id := CreativeModePromptID(mode); id != "" {
		if _, e := prompt.ResolveCreativePrompt(id, purpose); e != nil {
			return "", "", e
		}
		return "typedprompt:" + id, id, nil
	}
	if purpose == "character-design" {
		switch mode {
		case "single", "three_view", "three_view_realistic_sketch", "all_around":
			return mode, "", nil
		}
	} else if purpose == "scene-design" && mode == "none" {
		return mode, "", nil
	}
	return "", "", fmt.Errorf("出图模式%s不适用于%s", mode, purpose)
}
