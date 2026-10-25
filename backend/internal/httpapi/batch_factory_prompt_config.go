package httpapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"qiantie/backend/internal/prompt"
)

const (
	comicAssetsSlot          = "batch.comic-assets"
	comicPageSlot            = "batch.comic-page"
	comicImageDerivationSlot = "batch.comic-image-derivation"
	comicStyleSlot           = "batch.comic-style"
)

func normalizeBatchFactoryProductionMode(value string) (string, error) {
	mode := strings.TrimSpace(value)
	switch mode {
	case "", "video", "comic", "original", "viral", "original_direct", "viral_hook":
		return mode, nil
	default:
		return "", fmt.Errorf("生产模式无效")
	}
}

// hydrateBatchFactoryPromptConfig turns the compact preset IDs saved by the
// browser into an immutable backend snapshot. This keeps prompt text owned by
// the system preset library while retaining the exact version used by a book.
func hydrateBatchFactoryPromptConfig(catalog *prompt.Catalog, raw json.RawMessage) (json.RawMessage, error) {
	if catalog == nil || len(raw) == 0 {
		return raw, nil
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil || config == nil {
		return nil, fmt.Errorf("漫画提示词配置格式不正确")
	}
	comic, ok := config["comic"].(map[string]any)
	if !ok || comic == nil {
		return raw, nil
	}
	if err := freezePromptSelection(catalog, comic, "batch-factory", comicPageSlot, ""); err != nil {
		return nil, fmt.Errorf("漫画提示词：%w", err)
	}
	if derivation, ok := comic["imageDerivation"].(map[string]any); ok && derivation != nil {
		if err := freezePromptSelection(catalog, derivation, "batch-factory", comicImageDerivationSlot, ""); err != nil {
			return nil, fmt.Errorf("图片描述词推导：%w", err)
		}
	}
	if assets, ok := comic["assets"].(map[string]any); ok && assets != nil {
		if err := freezePromptSelection(catalog, assets, "batch-factory", comicAssetsSlot, ""); err != nil {
			return nil, fmt.Errorf("漫画资产提示词：%w", err)
		}
	}
	if style, ok := comic["style"].(map[string]any); ok && style != nil {
		if err := freezePromptSelection(catalog, style, "batch-factory", comicStyleSlot, ""); err != nil {
			return nil, fmt.Errorf("漫画画风：%w", err)
		}
	}
	if values, ok := comic["constraints"].([]any); ok {
		for index, value := range values {
			selection, ok := value.(map[string]any)
			if !ok || selection == nil {
				return nil, fmt.Errorf("漫画约束提示词第 %d 项格式不正确", index+1)
			}
			selectedCategory := strings.TrimSpace(stringValue(selection["constraintCategory"]))
			if err := freezePromptSelection(catalog, selection, "script", "", selectedCategory); err != nil {
				return nil, fmt.Errorf("漫画约束提示词第 %d 项：%w", index+1, err)
			}
			switch strings.TrimSpace(stringValue(selection["constraintCategory"])) {
			case "comic_quality", "comic_restriction", "comic_negative":
			default:
				return nil, fmt.Errorf("漫画约束提示词第 %d 项分类无效", index+1)
			}
		}
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("保存漫画提示词配置失败")
	}
	return encoded, nil
}

func freezePromptSelection(catalog *prompt.Catalog, selection map[string]any, module, slot, category string) error {
	id := strings.TrimSpace(stringValue(selection["presetId"]))
	if id == "" {
		return nil
	}
	preset, found, err := catalog.PublishedBody(module, id)
	if err != nil {
		return fmt.Errorf("读取系统预设失败")
	}
	if !found || strings.TrimSpace(preset.Body) == "" {
		return fmt.Errorf("所选系统预设不存在或未发布")
	}
	if slot != "" && preset.Slot != slot {
		return fmt.Errorf("所选系统预设不属于此位置")
	}
	if category != "" && preset.ConstraintCategory != category {
		return fmt.Errorf("所选系统预设分类不匹配")
	}
	selection["presetId"] = preset.ID
	selection["presetName"] = preset.Name
	selection["presetSlot"] = preset.Slot
	selection["presetVersion"] = preset.Version
	selection["constraintCategory"] = preset.ConstraintCategory
	selection["body"] = preset.Body
	return nil
}
