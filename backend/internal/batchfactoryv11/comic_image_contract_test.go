package batchfactoryv11

import (
	"strings"
	"testing"
)

func TestBuildComicImageContractCompilesStyleAndExplicitReferences(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"enabled":              true,
			"style":                map[string]any{"presetId": "batch-comic-style-14", "body": "水墨画风正文"},
			"imagePromptPrefix":    "用户全局前缀",
			"imagePromptSuffix":    "用户全局后缀",
			"globalReferenceImage": map[string]any{"enabled": true, "url": "https://example.com/style.jpg", "applyToAssets": true, "applyToPages": true},
			"constraints":          []any{map[string]any{"constraintCategory": "comic_negative", "body": "不要水印"}},
		}}),
	}
	contract, err := BuildComicImageContract(settings, ComicImagePurposePage, "可爱大头短身、圆脸大眼、低饱和柔和赛璐璐，当前漫画页描述", []string{"https://example.com/hero.jpg", "https://example.com/style.jpg"}, []string{"https://example.com/page.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	parts := []string{"水墨画风正文", "用户全局前缀", "阅读顺序", "当前漫画页描述", "用户全局后缀", "不要水印"}
	last := -1
	for _, part := range parts {
		index := strings.Index(contract.Prompt, part)
		if index <= last {
			t.Fatalf("prompt order is wrong for %q: %s", part, contract.Prompt)
		}
		last = index
	}
	wantReferences := []string{"https://example.com/page.jpg", "https://example.com/hero.jpg", "https://example.com/style.jpg"}
	if len(contract.ReferenceImageURLs) != len(wantReferences) {
		t.Fatalf("reference images = %#v", contract.ReferenceImageURLs)
	}
	for index := range wantReferences {
		if contract.ReferenceImageURLs[index] != wantReferences[index] {
			t.Fatalf("reference images = %#v", contract.ReferenceImageURLs)
		}
	}
	if contract.StylePresetID != "batch-comic-style-14" || contract.SnapshotHash == "" {
		t.Fatalf("style snapshot is incomplete: %#v", contract)
	}
	if strings.Contains(contract.Prompt, "可爱大头短身") || strings.Contains(contract.Prompt, "圆脸大眼") {
		t.Fatalf("legacy Q style survived selected-style compilation: %s", contract.Prompt)
	}
}

func TestBuildComicImageContractScopesGlobalReferenceAndRejectsVideoMode(t *testing.T) {
	comicSettings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"enabled":              true,
			"style":                map[string]any{"presetId": "batch-comic-style-01", "body": "韩漫画风"},
			"globalReferenceImage": map[string]any{"enabled": true, "url": "https://example.com/global.jpg", "applyToAssets": false, "applyToPages": true},
		}}),
	}
	asset, err := BuildComicImageContract(comicSettings, ComicImagePurposeAsset, "角色资产", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(asset.ReferenceImageURLs) != 0 {
		t.Fatalf("asset references = %#v", asset.ReferenceImageURLs)
	}
	page, err := BuildComicImageContract(comicSettings, ComicImagePurposePage, "漫画页", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.ReferenceImageURLs) != 1 || page.ReferenceImageURLs[0] != "https://example.com/global.jpg" {
		t.Fatalf("page references = %#v", page.ReferenceImageURLs)
	}
	videoSettings := SettingsPatch{"productionMode": rawSetting(t, "video"), "aiPromptConfig": comicSettings["aiPromptConfig"]}
	if _, err := BuildComicImageContract(videoSettings, ComicImagePurposePage, "漫画页", nil, nil); err == nil {
		t.Fatal("video mode accepted comic image compilation")
	}
}
