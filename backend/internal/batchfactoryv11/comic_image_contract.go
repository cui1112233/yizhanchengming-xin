package batchfactoryv11

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type ComicImagePurpose string

const (
	ComicImagePurposeAsset ComicImagePurpose = "asset"
	ComicImagePurposePage  ComicImagePurpose = "page"
)

type ComicImageContract struct {
	Prompt             string   `json:"prompt"`
	ReferenceImageURLs []string `json:"referenceImageUrls,omitempty"`
	StylePresetID      string   `json:"stylePresetId,omitempty"`
	SnapshotHash       string   `json:"snapshotHash"`
}

func ComicImageStyleConfigured(settings SettingsPatch) bool {
	return normalizeProductionMode(settings) == "comic" && strings.TrimSpace(aiReasoningPromptConfig(settings).Comic.Style.Body) != ""
}

func BuildComicImageContract(settings SettingsPatch, purpose ComicImagePurpose, assetKind, content string, assetImageURLs, pageReferenceURLs []string) (ComicImageContract, error) {
	if normalizeProductionMode(settings) != "comic" {
		return ComicImageContract{}, fmt.Errorf("%w: 当前书不是漫画生成模式", ErrConflict)
	}
	if purpose != ComicImagePurposeAsset && purpose != ComicImagePurposePage {
		return ComicImageContract{}, fmt.Errorf("%w: 漫画图片用途无效", ErrInvalid)
	}
	config := aiReasoningPromptConfig(settings).Comic
	style := strings.TrimSpace(config.Style.Body)
	if style == "" {
		return ComicImageContract{}, fmt.Errorf("%w: 请先选择画风", ErrConflict)
	}
	content = stripLegacyComicStyle(content)
	parts := make([]string, 0, 8)
	global := config.GlobalReferenceImage
	legacySceneReference := strings.HasPrefix(strings.TrimSpace(global.AssetID), "scene:") || global.Source == "scene-library" || global.Source == "builtin" || strings.TrimSpace(global.Scene.ID) != ""
	globalReferenceURL := strings.TrimSpace(global.URL)
	if legacySceneReference && strings.TrimSpace(config.Style.ID) != "" {
		globalReferenceURL = "/api/local-executor/v1/batch-comic-style-previews/" + url.PathEscape(strings.TrimSpace(config.Style.ID))
	}
	useGlobalReference := global.Enabled && globalReferenceURL != "" && ((purpose == ComicImagePurposeAsset && strings.TrimSpace(assetKind) == "character" && global.ApplyToAssets) || (purpose == ComicImagePurposePage && global.ApplyToPages))
	if useGlobalReference && !legacySceneReference {
		if scenePrompt := strings.TrimSpace(global.Scene.Body); scenePrompt != "" {
			parts = append(parts, scenePrompt)
		}
	}
	if styleName := strings.TrimSpace(config.Style.Name); styleName != "" {
		parts = append(parts, "画风名称："+styleName)
	}
	parts = append(parts, style)
	if useGlobalReference {
		parts = append(parts, "参考图使用规则：第1张参考图是画风库参考图，只用于锁定线条、上色、人物比例和整体画风；其余人物资产参考图只用于锁定角色身份、脸型、发型和服装，不得覆盖第1张画风。")
	}
	if prefix := strings.TrimSpace(config.ImagePromptPrefix); prefix != "" {
		parts = append(parts, prefix)
	}
	if purpose == ComicImagePurposePage {
		parts = append(parts, comicPromptTextForAspect(comicPageStructureLock, EffectiveImageAspectRatio(settings)))
	} else {
		parts = append(parts, "保留资产参考图中的身份、脸型、发型、服装、配饰或物体结构；忽略原始参考图的摄影、渲染和画风，最终画风严格遵循本书画风预设。")
	}
	if content = strings.TrimSpace(content); content != "" {
		parts = append(parts, content)
	}
	if suffix := strings.TrimSpace(config.ImagePromptSuffix); suffix != "" {
		parts = append(parts, suffix)
	}
	if purpose == ComicImagePurposePage {
		for _, selection := range config.Constraints {
			if constraintCategory(selection) == comicNegativeConstraint {
				if body := strings.TrimSpace(stripLegacyComicStyle(selection.Body)); body != "" {
					parts = append(parts, body)
				}
			}
		}
	}
	styleLock := []string{"【最终画风锁定】"}
	if styleName := strings.TrimSpace(config.Style.Name); styleName != "" {
		styleLock = append(styleLock, "画风名称："+styleName)
	}
	styleLock = append(styleLock, style)
	priorityRule := "最高优先级：最终图的线条、上色、材质、人物比例和渲染方式必须严格遵循上述所选画风；剧情描述和人物资产图只提供内容与角色身份，不得把画风改成写实、摄影、3D或其他风格。"
	if useGlobalReference {
		priorityRule += " 第1张参考图是画风库参考图，其画风优先于后续所有参考图。"
	}
	styleLock = append(styleLock, priorityRule)
	parts = append(parts, strings.Join(styleLock, "\n"))
	references := make([]string, 0, len(pageReferenceURLs)+len(assetImageURLs)+1)
	if useGlobalReference {
		references = appendUniqueComicReferences(references, globalReferenceURL)
	}
	if purpose == ComicImagePurposePage {
		for _, referenceURL := range pageReferenceURLs {
			if isLegacyComicSceneReferenceURL(referenceURL) {
				continue
			}
			references = appendUniqueComicReferences(references, referenceURL)
		}
	}
	references = appendUniqueComicReferences(references, assetImageURLs...)
	if len(references) > 6 {
		references = references[:6]
	}
	contract := ComicImageContract{Prompt: strings.Join(parts, "\n\n"), ReferenceImageURLs: references, StylePresetID: strings.TrimSpace(config.Style.ID)}
	hashInput, _ := json.Marshal(map[string]any{"purpose": purpose, "prompt": contract.Prompt, "references": references, "stylePresetId": contract.StylePresetID})
	sum := sha256.Sum256(hashInput)
	contract.SnapshotHash = hex.EncodeToString(sum[:])
	return contract, nil
}

func isLegacyComicSceneReferenceURL(value string) bool {
	value = strings.TrimSpace(value)
	return strings.Contains(value, "/batch-comic-reference-scenes/") || strings.Contains(value, "/comic-reference-scene/")
}

func appendUniqueComicReferences(target []string, values ...string) []string {
	seen := make(map[string]bool, len(target)+len(values))
	for _, value := range target {
		seen[value] = true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		target = append(target, value)
	}
	return target
}

func stripLegacyComicStyle(value string) string {
	replacer := strings.NewReplacer(
		"可爱大头短身比例（二点五至三点五头身）、圆脸大眼、淡腮红、低饱和柔和赛璐璐平涂。", "",
		"可爱大头短身、圆脸大眼、低饱和柔和赛璐璐，", "",
		"可爱大头短身、圆脸大眼、低饱和柔和赛璐璐上色；", "",
		"人物按二点五至三点五头身的可爱短身比例绘制，", "",
		"人物按二点五至三点五头身的可爱短身比例绘制", "",
		"二点五至三点五头身", "",
		"可爱大头短身", "",
		"圆脸大眼", "",
	)
	return strings.TrimSpace(replacer.Replace(value))
}
