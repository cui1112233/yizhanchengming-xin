package batchfactoryv11

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

func BuildComicImageContract(settings SettingsPatch, purpose ComicImagePurpose, content string, assetImageURLs, pageReferenceURLs []string) (ComicImageContract, error) {
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
	parts := []string{style}
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
	references := make([]string, 0, len(pageReferenceURLs)+len(assetImageURLs)+1)
	if purpose == ComicImagePurposePage {
		references = appendUniqueComicReferences(references, pageReferenceURLs...)
	}
	references = appendUniqueComicReferences(references, assetImageURLs...)
	global := config.GlobalReferenceImage
	if global.Enabled && ((purpose == ComicImagePurposeAsset && global.ApplyToAssets) || (purpose == ComicImagePurposePage && global.ApplyToPages)) {
		references = appendUniqueComicReferences(references, global.URL)
	}
	contract := ComicImageContract{Prompt: strings.Join(parts, "\n\n"), ReferenceImageURLs: references, StylePresetID: strings.TrimSpace(config.Style.ID)}
	hashInput, _ := json.Marshal(map[string]any{"purpose": purpose, "prompt": contract.Prompt, "references": references, "stylePresetId": contract.StylePresetID})
	sum := sha256.Sum256(hashInput)
	contract.SnapshotHash = hex.EncodeToString(sum[:])
	return contract, nil
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
