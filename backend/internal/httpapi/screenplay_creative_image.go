package httpapi

import (
	"context"
	"fmt"
	"strings"

	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
)

// Resolve the selected asset and template before creating a billable attempt.
// Style preview images are deliberately not used as implicit image references.
func (s *ScreenplayMediaWorkbench) creativeImagePrompt(ctx context.Context, p screenplay.Project, target, extra string) (string, error) {
	parts := []string{}
	config := screenplay.EffectiveCreativeConfig(p)
	var asset *screenplay.Asset
	if target != "" && target != "narrator" {
		assets, err := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
		if err != nil {
			return "", err
		}
		for i := range assets {
			if assets[i].ID == target {
				asset = &assets[i]
				break
			}
		}
	}
	style := strings.TrimSpace(p.StyleDesc)
	if p.CreativeConfig != nil && p.CreativeConfig.Style != nil {
		style = strings.TrimSpace(p.CreativeConfig.Style.Prompt)
	}
	if asset != nil {
		description := strings.TrimSpace(asset.Description)
		if asset.Prompt != "" {
			description = strings.TrimSpace(description + "\n" + asset.Prompt)
		}
		if description == "" {
			return "", fmt.Errorf("%w: 请先补充目标资产的描述或提示词", screenplay.ErrInvalidArgument)
		}
		id, purpose := "", ""
		mode := ""
		switch asset.Type {
		case screenplay.AssetCharacter:
			mode = config.CharacterImageMode
			id, purpose = config.CharacterDesignPromptID, "character-design"
		case screenplay.AssetScene:
			mode = config.SceneImageMode
			id, purpose = config.SceneDesignPromptID, "scene-design"
		}
		if mode != "" {
			if selected := screenplay.CreativeModePromptID(mode); selected != "" {
				id = selected
			} else {
				id = ""
			}
		}
		if id != "" {
			preset, err := screenplay.ResolveProjectPrompt(p, id, purpose)
			if err != nil {
				return "", fmt.Errorf("%w: %v", screenplay.ErrInvalidArgument, err)
			}
			body := strings.NewReplacer("{name}", asset.Name, "{description}", description, "{sceneName}", asset.Name, "{sceneDescription}", description, "{stylePrompt}", style).Replace(preset.Body)
			parts = append(parts, body)
		} else {
			parts = append(parts, "资产："+asset.Name+"\n"+description)
			if asset.Type == screenplay.AssetCharacter {
				if body := prompt.ScreenplayLegacyCharacterImage(mode); body != "" {
					parts = append(parts, body)
				}
			}
		}
	} else if target != "" && target != "narrator" {
		boards, err := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, "")
		if err != nil {
			return "", err
		}
		for _, board := range boards {
			for _, shot := range board.Shots {
				if shot.ID == target {
					imagePrompt := strings.TrimSpace(shot.ImagePrompt)
					if imagePrompt == "" {
						if board.Engine == screenplay.StoryboardEngine("prompt_139") {
							if strings.TrimSpace(extra) == "" {
								return "", fmt.Errorf("%w: 当前分镜没有图片提示词，请先补充图片提示词", screenplay.ErrInvalidArgument)
							}
						} else {
							imagePrompt = shot.VideoDesc
						}
					}
					parts = append(parts, imagePrompt)
				}
			}
		}
	}
	if style != "" {
		parts = append(parts, "画风要求："+style)
	}
	if strings.TrimSpace(extra) != "" {
		parts = append(parts, "本次补充要求："+strings.TrimSpace(extra))
	}
	result := strings.TrimSpace(strings.Join(parts, "\n\n"))
	if result == "" {
		return "", fmt.Errorf("%w: 请填写图片提示词或选择有描述的目标资产", screenplay.ErrInvalidArgument)
	}
	return screenplayCorePrefix(p, result), nil
}

func (s *ScreenplayMediaWorkbench) creativeAssetProcessPrompt(ctx context.Context, p screenplay.Project, target, extra string) (string, error) {
	config := screenplay.EffectiveCreativeConfig(p)
	if strings.TrimSpace(config.AssetProcessPrompt) == "" {
		return "", fmt.Errorf("%w: 请在模型配置填写资产处理提示词", screenplay.ErrInvalidArgument)
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return "", e
	}
	found := false
	for _, a := range assets {
		if a.ID == target && a.Type == screenplay.AssetCharacter {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("%w: 资产处理只支持当前项目的角色资产", screenplay.ErrInvalidArgument)
	}
	text := config.AssetProcessPrompt
	if strings.TrimSpace(extra) != "" {
		text += "\n\n本次补充要求：" + extra
	}
	// Original qx explicitly uses the dedicated processing prompt, without the normal image core.
	return text, nil
}
