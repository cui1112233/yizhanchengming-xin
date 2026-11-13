package batchfactoryv11

import (
	"encoding/json"
	"reflect"
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
	contract, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "可爱大头短身、圆脸大眼、低饱和柔和赛璐璐，当前漫画页描述", []string{"https://example.com/hero.jpg", "https://example.com/style.jpg"}, []string{"https://example.com/page.jpg"})
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
	wantReferences := []string{"https://example.com/style.jpg", "https://example.com/page.jpg", "https://example.com/hero.jpg"}
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

func TestBuildComicImageContractIsIdempotentAcrossRepeatedCompiles(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"enabled":              true,
			"style":                map[string]any{"presetId": "batch-comic-style-16", "presetName": "精致火柴2DQ版漫风", "body": "精致头部，极简火柴四肢，细长黑线腿部"},
			"imagePromptPrefix":    "用户全局前缀",
			"imagePromptSuffix":    "用户全局后缀",
			"globalReferenceImage": map[string]any{"enabled": true, "url": "https://example.com/style.jpg", "applyToAssets": true, "applyToPages": true},
			"constraints":          []any{map[string]any{"constraintCategory": "comic_negative", "body": "不要水印"}},
		}}),
	}
	assetImages := []string{"https://example.com/hero.jpg"}
	pageReferences := []string{"https://example.com/page.jpg"}
	first, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "原始漫画页内容", assetImages, pageReferences)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildComicImageContract(settings, ComicImagePurposePage, "", first.Prompt, assetImages, pageReferences)
	if err != nil {
		t.Fatal(err)
	}
	third, err := BuildComicImageContract(settings, ComicImagePurposePage, "", second.Prompt, assetImages, pageReferences)
	if err != nil {
		t.Fatal(err)
	}
	for index, contract := range []ComicImageContract{second, third} {
		if contract.Prompt != first.Prompt {
			t.Fatalf("compile %d changed the prompt:\nfirst:  %s\nactual: %s", index+2, first.Prompt, contract.Prompt)
		}
		if contract.SnapshotHash != first.SnapshotHash {
			t.Fatalf("compile %d changed the snapshot hash: %s != %s", index+2, contract.SnapshotHash, first.SnapshotHash)
		}
	}
	if count := strings.Count(first.Prompt, comicStyleLockMarker); count != 1 {
		t.Fatalf("style lock block count = %d, want 1: %s", count, first.Prompt)
	}
	if !strings.Contains(first.Prompt, "原始漫画页内容") {
		t.Fatalf("original content was dropped: %s", first.Prompt)
	}
}

func TestStripCompiledComicImagePromptRemovesStackedStyleLocks(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"style": map[string]any{"presetId": "batch-comic-style-16", "presetName": "精致火柴2DQ版漫风", "body": "精致头部，极简火柴四肢，细长黑线腿部"},
		}}),
	}
	dirty := "原始漫画页内容\n\n画风名称：精致火柴2DQ版漫风\n\n精致头部，极简火柴四肢，细长黑线腿部\n\n【最终画风锁定】\n画风名称：精致火柴2DQ版漫风\n精致头部，极简火柴四肢，细长黑线腿部\n最高优先级：必须遵循上述画风。\n\n【最终画风锁定】\n画风名称：精致火柴2DQ版漫风\n精致头部，极简火柴四肢，细长黑线腿部\n最高优先级：必须遵循上述画风。"
	if got := StripCompiledComicImagePrompt(settings, ComicImagePurposePage, dirty); got != "原始漫画页内容" {
		t.Fatalf("stripped prompt = %q", got)
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
	asset, err := BuildComicImageContract(comicSettings, ComicImagePurposeAsset, "character", "角色资产", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(asset.ReferenceImageURLs) != 0 {
		t.Fatalf("asset references = %#v", asset.ReferenceImageURLs)
	}
	page, err := BuildComicImageContract(comicSettings, ComicImagePurposePage, "", "漫画页", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.ReferenceImageURLs) != 1 || page.ReferenceImageURLs[0] != "https://example.com/global.jpg" {
		t.Fatalf("page references = %#v", page.ReferenceImageURLs)
	}
	videoSettings := SettingsPatch{"productionMode": rawSetting(t, "video"), "aiPromptConfig": comicSettings["aiPromptConfig"]}
	if _, err := BuildComicImageContract(videoSettings, ComicImagePurposePage, "", "漫画页", nil, nil); err == nil {
		t.Fatal("video mode accepted comic image compilation")
	}
}

func TestBuildComicImageContractSendsStyleLibraryReferenceToPages(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"style": map[string]any{"presetId": "batch-comic-style-01", "body": "韩漫画风"},
			"globalReferenceImage": map[string]any{
				"assetId": "style:batch-comic-style-01", "url": "https://example.com/style-preview.jpg", "source": "style-library",
				"enabled": true, "applyToAssets": true, "applyToPages": true,
			},
		}}),
	}
	contract, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "漫画页", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(contract.ReferenceImageURLs) != 1 || contract.ReferenceImageURLs[0] != "https://example.com/style-preview.jpg" {
		t.Fatalf("style reference was not sent to page generation: %#v", contract.ReferenceImageURLs)
	}
}

func TestBuildComicImageContractReplacesLegacySceneReferenceWithSelectedStyle(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"style": map[string]any{"presetId": "batch-comic-style-01", "body": "漫画画风"},
			"globalReferenceImage": map[string]any{
				"assetId": "scene:batch-comic-reference-scene-campus-b", "url": "/api/local-executor/v1/batch-comic-reference-scenes/batch-comic-reference-scene-campus-b",
				"enabled": true, "applyToAssets": false, "applyToPages": true,
				"scene": map[string]any{"presetId": "batch-comic-reference-scene-campus-b", "body": "场景隐藏提示词"},
			},
		}}),
	}
	page, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "漫画页", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(page.Prompt, "场景隐藏提示词") || !reflect.DeepEqual(page.ReferenceImageURLs, []string{"/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-01"}) {
		t.Fatalf("legacy scene reference leaked into page contract: %#v", page)
	}
}

func TestBuildComicImageContractReplacesPersistedLegacySceneWithSelectedStyleReference(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"style": map[string]any{"presetId": "batch-comic-style-16", "presetName": "精致火柴2DQ版漫风", "body": "精致头部，极简四肢"},
			"globalReferenceImage": map[string]any{
				"assetId": "scene:batch-comic-reference-scene-urban", "url": "/api/local-executor/v1/batch-comic-reference-scenes/batch-comic-reference-scene-urban",
				"source": "builtin", "enabled": true, "applyToAssets": true, "applyToPages": true,
			},
		}}),
	}
	contract, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "漫画页", []string{"/api/assets/character.png"}, []string{
		"/api/local-executor/v1/batch-comic-reference-scenes/batch-comic-reference-scene-urban",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-16",
		"/api/assets/character.png",
	}
	if !reflect.DeepEqual(contract.ReferenceImageURLs, want) {
		t.Fatalf("references = %#v, want selected style first and no legacy scene: %#v", contract.ReferenceImageURLs, want)
	}
	if !strings.Contains(contract.Prompt, "第1张参考图是画风库参考图") || !strings.Contains(contract.Prompt, "人物资产参考图只用于锁定角色") {
		t.Fatalf("reference roles are not explicit in prompt: %s", contract.Prompt)
	}
}

func TestBuildComicImageContractIncludesStyleNameAndCapsReferencesAtSix(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"enabled":              true,
			"style":                map[string]any{"presetId": "batch-comic-style-14", "presetName": "水墨古风", "body": "水墨晕染，毛笔笔触"},
			"globalReferenceImage": map[string]any{"enabled": true, "applyToPages": true, "url": "https://example.com/global.jpg"},
		}}),
	}
	contract, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "当前漫画页", []string{
		"https://example.com/asset-1.jpg", "https://example.com/asset-2.jpg", "https://example.com/asset-3.jpg", "https://example.com/asset-4.jpg",
	}, []string{
		"https://example.com/page-1.jpg", "https://example.com/page-2.jpg", "https://example.com/page-3.jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.Prompt, "水墨古风") || !strings.Contains(contract.Prompt, "水墨晕染，毛笔笔触") {
		t.Fatalf("style name and body must both reach the image prompt: %s", contract.Prompt)
	}
	want := []string{
		"https://example.com/global.jpg", "https://example.com/page-1.jpg", "https://example.com/page-2.jpg",
		"https://example.com/page-3.jpg", "https://example.com/asset-1.jpg", "https://example.com/asset-2.jpg",
	}
	if !reflect.DeepEqual(contract.ReferenceImageURLs, want) {
		t.Fatalf("references = %#v, want %#v", contract.ReferenceImageURLs, want)
	}
}

func TestBuildComicImageContractLocksSelectedStyleAfterConflictingPageDescription(t *testing.T) {
	settings := SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"style": map[string]any{
				"presetId":   "batch-comic-style-16",
				"presetName": "精致火柴2DQ版漫风",
				"body":       "精致头部，极简火柴四肢，细长黑线腿部",
			},
			"globalReferenceImage": map[string]any{
				"enabled": true, "applyToPages": true,
				"url": "/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-16",
			},
		}}),
	}
	contract, err := BuildComicImageContract(settings, ComicImagePurposePage, "", "都市现代风格，写实背景，时尚人物。", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	conflictIndex := strings.Index(contract.Prompt, "写实背景，时尚人物")
	styleLockIndex := strings.LastIndex(contract.Prompt, "精致头部，极简火柴四肢，细长黑线腿部")
	if conflictIndex < 0 || styleLockIndex <= conflictIndex {
		t.Fatalf("所选画风没有在冲突页面描述后再次锁定: %s", contract.Prompt)
	}
	if !strings.Contains(contract.Prompt[styleLockIndex:], "最高优先级") {
		t.Fatalf("末尾画风锁定没有明确优先级: %s", contract.Prompt)
	}
}

func comicPageLiteralTestSettings(t *testing.T, presetID string) SettingsPatch {
	t.Helper()
	return SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"presetId": presetID,
			"style": map[string]any{
				"presetId": "batch-comic-style-08",
				"body":     "韩漫条漫画风",
			},
		}}),
	}
}

func illustratedTextPromptSettings(t *testing.T, presetID string) SettingsPatch {
	t.Helper()
	return SettingsPatch{
		"productionMode":     rawSetting(t, "comic"),
		"assetImagesEnabled": rawSetting(t, false),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"enabled":  true,
			"presetId": presetID,
			"body":     "单页图文卡：上半文字区按原文排版，下半插画，不分格，不画气泡。",
			"style": map[string]any{
				"presetId": "batch-comic-style-08",
				"body":     "韩漫条漫画风",
			},
		}}),
	}
}

func TestIllustratedTextPresetUsesCardStructureLock(t *testing.T) {
	book := Book{ID: "book_1", Title: "测试书", SourceText: "第一页正文。第二页正文。"}
	videos := []Video{{ID: "video_1", Label: "漫画页 01", VideoPrompt: "第一页正文。"}}
	contract, err := BuildComicPromptContract(book, videos, 0, DirectorSnapshot{Effective: illustratedTextPromptSettings(t, ComicPageIllustratedTextPresetID)})
	if err != nil {
		t.Fatalf("build contract: %v", err)
	}
	if !strings.Contains(contract.SystemPrompt, "【单页图文卡结构约束】") {
		t.Fatalf("文字插画版必须使用图文卡版式锁:\n%s", contract.SystemPrompt)
	}
	if strings.Contains(contract.SystemPrompt, "【漫画页面结构约束】") {
		t.Fatalf("文字插画版不能注入分格漫画版式锁:\n%s", contract.SystemPrompt)
	}
	if strings.Contains(contract.SystemPrompt, "漫画文字密度") {
		t.Fatalf("文字插画版不能注入 45-70 字密度规则:\n%s", contract.SystemPrompt)
	}
	if !strings.Contains(contract.SystemPrompt, "上半为文字区") {
		t.Fatalf("文字插画版必须声明上半文字区/下半插画区:\n%s", contract.SystemPrompt)
	}
}

func TestPanelPresetKeepsPanelStructureLock(t *testing.T) {
	book := Book{ID: "book_1", Title: "测试书", SourceText: "第一页正文。"}
	videos := []Video{{ID: "video_1", Label: "漫画页 01", VideoPrompt: "第一页正文。"}}
	contract, err := BuildComicPromptContract(book, videos, 0, DirectorSnapshot{Effective: illustratedTextPromptSettings(t, "batch-comic-page")})
	if err != nil {
		t.Fatalf("build contract: %v", err)
	}
	if !strings.Contains(contract.SystemPrompt, "【漫画页面结构约束】") || !strings.Contains(contract.SystemPrompt, "漫画文字密度") {
		t.Fatalf("分格漫画预设必须保留原版式锁和密度规则:\n%s", contract.SystemPrompt)
	}
}

func TestIllustratedTextPresetImageContractUsesCardLock(t *testing.T) {
	contract, err := BuildComicImageContract(illustratedTextPromptSettings(t, ComicPageIllustratedTextPresetID), ComicImagePurposePage, "", "上半文字区，下半插画。", nil, nil)
	if err != nil {
		t.Fatalf("build image contract: %v", err)
	}
	if !strings.Contains(contract.Prompt, "【单页图文卡结构约束】") {
		t.Fatalf("文字插画版最终出图 prompt 必须是图文卡结构:\n%s", contract.Prompt)
	}
	if strings.Contains(contract.Prompt, "【漫画页面结构约束】") {
		t.Fatalf("文字插画版最终出图 prompt 不能出现分格版式锁:\n%s", contract.Prompt)
	}
}

func TestComicPageIllustratedTextPresetGatesOnlyIllustratedText(t *testing.T) {
	if !ComicPageIllustratedTextPreset(comicPageLiteralTestSettings(t, ComicPageIllustratedTextPresetID)) {
		t.Fatal("CM·文字插画版必须开启原文注入")
	}
	for _, presetID := range []string{
		"batch-comic-page",
		"batch-comic-page-fusion-bw",
		"batch-comic-page-light",
		"batch-comic-page-premium-strip",
	} {
		if ComicPageIllustratedTextPreset(comicPageLiteralTestSettings(t, presetID)) {
			t.Fatalf("预设 %s 不应开启原文注入", presetID)
		}
	}
}

func TestAppendComicPageLiteralTextKeepsSourceVerbatim(t *testing.T) {
	source := "夜色压低了整条街。\n「你迟到了。」陆淮安把外套递过来，声音很轻。"
	got, appended := AppendComicPageLiteralText("画面上方文字区，下方插画区。", source)
	if !appended {
		t.Fatal("原文应被追加")
	}
	if !strings.Contains(got, source) {
		t.Fatalf("最终 prompt 必须逐字包含原文（含换行）:\n%s", got)
	}
	if !strings.Contains(got, comicPageLiteralTextHeader) {
		t.Fatalf("最终 prompt 必须包含硬约束标题:\n%s", got)
	}
	if strings.Count(got, source) != 1 {
		t.Fatalf("原文只应出现一次:\n%s", got)
	}
	payload, err := json.Marshal(map[string]string{"prompt": got})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["prompt"] != got {
		t.Fatal("JSON 往返后 prompt 必须保持不变（引号与换行安全转义）")
	}
}

func TestAppendComicPageLiteralTextDedupesExistingSource(t *testing.T) {
	source := "我推开门。\n「你来了。」"
	prompt := "版式说明\n\n" + source
	got, appended := AppendComicPageLiteralText(prompt, source)
	if appended {
		t.Fatal("已带原文时不得重复拼接")
	}
	if strings.Count(got, source) != 1 {
		t.Fatalf("原文只应保留一份:\n%s", got)
	}
}

func TestAppendComicPageLiteralTextSkipsEmptySource(t *testing.T) {
	prompt := "版式说明"
	got, appended := AppendComicPageLiteralText(prompt, " \n ")
	if appended || got != prompt {
		t.Fatalf("空原文时不得改动 prompt: %q", got)
	}
}
