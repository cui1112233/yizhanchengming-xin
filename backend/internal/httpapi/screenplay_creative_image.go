package httpapi

import (
	"context"
	"fmt"
	"strings"

	"qiantie/backend/internal/screenplay"
)

// Resolve the selected asset and template before creating a billable attempt.
// Style preview images are deliberately not used as implicit image references.
func (s *ScreenplayMediaWorkbench) creativeImagePrompt(ctx context.Context, p screenplay.Project, target, extra string) (string, error) {
	parts := []string{}
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
		if p.CreativeConfig != nil {
			switch asset.Type {
			case screenplay.AssetCharacter:
				id, purpose = p.CreativeConfig.CharacterDesignPromptID, "character-design"
			case screenplay.AssetScene:
				id, purpose = p.CreativeConfig.SceneDesignPromptID, "scene-design"
			}
		}
		if purpose != "" && id == "" {
			return "", fmt.Errorf("%w: 请先选择对应的资产设计提示词", screenplay.ErrInvalidArgument)
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
		}
	} else if target != "" && target != "narrator" {
		boards, err := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, "")
		if err != nil {
			return "", err
		}
		for _, board := range boards {
			for _, shot := range board.Shots {
				if shot.ID == target {
					parts = append(parts, shot.VideoDesc)
				}
			}
		}
	}
	if style != "" {
		parts = append(parts, "画风要求："+style)
	}
	if p.CreativeConfig != nil && strings.TrimSpace(p.CreativeConfig.ImageCoreInstruction) != "" {
		parts = append(parts, "图片核心要求："+p.CreativeConfig.ImageCoreInstruction)
	}
	if strings.TrimSpace(extra) != "" {
		parts = append(parts, "本次补充要求："+strings.TrimSpace(extra))
	}
	result := strings.TrimSpace(strings.Join(parts, "\n\n"))
	if result == "" {
		return "", fmt.Errorf("%w: 请填写图片提示词或选择有描述的目标资产", screenplay.ErrInvalidArgument)
	}
	return result, nil
}
