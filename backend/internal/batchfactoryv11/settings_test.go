package batchfactoryv11

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

func TestSparsePatchPreservesFalseEmptyZero(t *testing.T) {
	current := SettingsPatch{"enabled": raw(true), "prefix": raw("x"), "duration": raw(9)}
	got := ApplySparseUpdate(current, SettingsUpdate{Patch: SettingsPatch{"enabled": raw(false), "prefix": raw(""), "duration": raw(0)}})
	for key, want := range map[string]string{"enabled": "false", "prefix": "\"\"", "duration": "0"} {
		if string(got[key]) != want {
			t.Fatalf("%s=%s want %s", key, got[key], want)
		}
	}
}

func TestNormalizeProductionModeDefaultsToVideo(t *testing.T) {
	if got := normalizeProductionMode(SettingsPatch{}); got != "video" {
		t.Fatalf("got %q want video", got)
	}
	if got := normalizeProductionMode(SettingsPatch{"productionMode": raw("comic")}); got != "comic" {
		t.Fatalf("got %q want comic", got)
	}
	if got := normalizeProductionMode(SettingsPatch{"productionMode": raw("unknown")}); got != "video" {
		t.Fatalf("got %q want video", got)
	}
}

func TestRestoreInheritanceRemovesOnlySelectedCurrentScopeKeys(t *testing.T) {
	current := SettingsPatch{"quality": raw("book"), "enabled": raw(false), "duration": raw(0)}
	got := ApplySparseUpdate(current, SettingsUpdate{RestoreKeys: []string{"quality"}})
	if _, ok := got["quality"]; ok {
		t.Fatal("quality should inherit after restore")
	}
	if string(got["enabled"]) != "false" || string(got["duration"]) != "0" {
		t.Fatalf("got=%v", got)
	}
}

func TestResolveSettingsUsesPresenceNotTruthiness(t *testing.T) {
	effective := ResolveSettings(
		SettingsPatch{"enabled": raw(true), "prefix": raw("system"), "duration": raw(10)},
		SettingsPatch{"enabled": raw(false)},
		SettingsPatch{"prefix": raw("")},
		SettingsPatch{"duration": raw(0)},
	)
	if string(effective["enabled"]) != "false" || string(effective["prefix"]) != "\"\"" || string(effective["duration"]) != "0" {
		t.Fatalf("effective=%v", effective)
	}
}

func TestResolveSettingsGivesBookOverridesPriorityOverBatchUnifiedConfiguration(t *testing.T) {
	batch := SettingsPatch{
		"videoProvider":           raw("autodl_h3"),
		"videoModelId":            raw("minimax-h3"),
		"textModelId":             raw("ant-2/gemini-3.8-flash"),
		"imageModelId":            raw("gpt-image-2-c"),
		"videoResolution":         raw("720p"),
		"videoAspectRatio":        raw("9:16"),
		"storyboardDurationLimit": raw(5),
		"openingEnabled":          raw(false),
		"sensitiveAIModels":       raw([]string{"ant-2/gemini-3.8-flash"}),
		"publishSettings":         raw(map[string]any{"websiteProfileId": "unified-profile"}),
	}
	book := SettingsPatch{
		"videoProvider":    raw("yfai"),
		"videoModelId":     raw("seedance-2"),
		"storyboardDurationLimit": raw(5),
		"videoResolution":  raw("720p"),
	}

	effective := ResolveSettings(batch, book)
	if string(effective["videoProvider"]) != "\"yfai\"" || string(effective["videoModelId"]) != "\"seedance-2\"" {
		t.Fatalf("单书显式值应优先于批次统一配置，effective=%s / %s", effective["videoProvider"], effective["videoModelId"])
	}
	if string(effective["textModelId"]) != "\"ant-2/gemini-3.8-flash\"" || string(effective["imageModelId"]) != "\"gpt-image-2-c\"" {
		t.Fatalf("单书未设置的字段应由批次统一配置兜底，effective=%s / %s", effective["textModelId"], effective["imageModelId"])
	}
	if string(effective["openingEnabled"]) != "false" || string(effective["sensitiveAIModels"]) != "[\"ant-2/gemini-3.8-flash\"]" {
		t.Fatalf("批次统一配置中的换开头/修复模型应继续生效，effective=%s / %s", effective["openingEnabled"], effective["sensitiveAIModels"])
	}
}

func TestResolveSettingsDeepMergesAIPromptConfigModules(t *testing.T) {
	batch := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"assets": map[string]any{"enabled": true, "presetId": "batch-assets"},
		"visual": map[string]any{"enabled": false},
	})}
	book := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"constraints": map[string]any{"enabled": true, "presetId": "book-constraints"},
	})}

	effective := ResolveSettings(batch, book)
	var prompt map[string]json.RawMessage
	if err := json.Unmarshal(effective["aiPromptConfig"], &prompt); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"assets", "constraints", "visual"} {
		if _, ok := prompt[key]; !ok {
			t.Fatalf("missing %s in %#v", key, prompt)
		}
	}
}

func TestEffectiveProductionModeKeepsVideoBatchOutOfComicStages(t *testing.T) {
	// 批次在页面上按视频渲染时，单书里遗留的 comic 模式不得把编排拉进
	// comic_assets 阶段。
	batchVideo := SettingsPatch{"productionMode": raw("video")}
	staleBookComic := SettingsPatch{"productionMode": raw("comic")}
	if mode := EffectiveProductionMode(batchVideo, staleBookComic); mode != "video" {
		t.Fatalf("mode=%q, want video", mode)
	}
	// 批次未写模式时，仍兼容旧的单书 comic 模式。
	if mode := EffectiveProductionMode(SettingsPatch{}, SettingsPatch{"productionMode": raw("comic")}); mode != "comic" {
		t.Fatalf("legacy book mode=%q, want comic", mode)
	}
	// 两侧都没写时，保留漫画配置推断。
	inferred := SettingsPatch{"aiPromptConfig": raw(map[string]any{"comic": map[string]any{"enabled": true, "body": "分镜规则"}})}
	if mode := EffectiveProductionMode(inferred, SettingsPatch{}); mode != "comic" {
		t.Fatalf("inferred mode=%q, want comic", mode)
	}
}

func TestResolveSettingsKeepsSiblingComicSelectionsWhenBookWritesPagesOnly(t *testing.T) {
	batch := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"comic": map[string]any{
			"enabled": true,
			"assets":  map[string]any{"presetId": "batch-comic-assets", "body": "漫画资产提示词正文"},
		},
	})}
	// 单书只改了漫画分页提示词，没有重复写 assets；旧的一层合并会把批次里
	// 已选的漫画资产提示词整个丢掉，阶段随后报“请选择漫画资产提示词”。
	book := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"comic": map[string]any{"body": "漫画分页提示词正文"},
	})}

	effective := ResolveSettings(batch, book)
	comic := aiReasoningPromptConfig(effective).Comic
	if strings.TrimSpace(comic.Assets.Body) == "" || comic.Assets.ID != "batch-comic-assets" {
		t.Fatalf("book page-only write dropped the batch comic asset selection: %#v", comic.Assets)
	}
	if comic.Body != "漫画分页提示词正文" {
		t.Fatalf("book comic page preset=%q, want the book override", comic.Body)
	}
}

func TestResolveSettingsKeepsLegacyFullBookPromptPayload(t *testing.T) {
	batch := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"assets": map[string]any{"presetId": "batch-assets"},
		"video":  map[string]any{"presetId": "batch-video"},
	})}
	book := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"assets": map[string]any{"presetId": "book-assets"},
		"video":  map[string]any{"presetId": "book-video"},
	})}

	effective := ResolveSettings(batch, book)
	var prompt map[string]map[string]string
	if err := json.Unmarshal(effective["aiPromptConfig"], &prompt); err != nil {
		t.Fatal(err)
	}
	if prompt["assets"]["presetId"] != "book-assets" || prompt["video"]["presetId"] != "book-video" {
		t.Fatalf("legacy book prompt payload changed: %#v", prompt)
	}
}

// 统一配置保存只允许写批次作用域。书级 patch 里既有单书覆盖值，也有已生成的
// 内容/测量产物（分镜文本、配音逐行时长、配音指纹、H3 风格分析）。这些键必须
// 逐字节保持原样，否则“改配置”会顺带把已完成阶段变成待重跑。
func TestBatchUnifiedSettingsSaveKeepsBookContentAndOverridesUntouched(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批次", Books: []CreateBookInput{{Title: "书", SourceText: "她被当众羞辱后沉默离开。"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	bookPatch := SettingsPatch{
		"videoProvider":            raw("yfai"),
		"videoModelId":             raw("seedance-2-0-official"),
		"audioLineDurations":       raw([]float64{1.2, 2.4}),
		"audioDurationFingerprint": raw("fp-voice-1"),
		"ttsAudioArtifactId":       raw("audio-artifact-1"),
		"h3StyleAnalysis":          raw(map[string]any{"summary": "已生成的风格分析"}),
	}
	saved, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: bookPatch, ExpectedRevision: book.Revision})
	if err != nil {
		t.Fatal(err)
	}

	// 模拟“统一配置”面板保存：只写批次作用域，等价于后端收到的唯一写请求。
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch:            SettingsPatch{"videoModelId": raw("minimax-h3-video")},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}

	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	latestBook := latest.Books[0]
	if latestBook.SettingsState.Revision != saved.Revision {
		t.Fatalf("统一配置保存改动书级 revision：got %d want %d", latestBook.SettingsState.Revision, saved.Revision)
	}
	for key, want := range bookPatch {
		got, ok := latestBook.SettingsState.Patch[key]
		if !ok || string(got) != string(want) {
			t.Fatalf("书级 %s 被统一配置保存改动：got %s want %s", key, got, want)
		}
	}
	// 运行时仍然是单书优先：批次新模型不能盖掉这本单书的选择。
	effective := ResolveSettings(latest.SettingsState.Patch, latestBook.SettingsState.Patch)
	if string(effective["videoModelId"]) != "\"seedance-2-0-official\"" {
		t.Fatalf("运行时单书模型未生效：%s", effective["videoModelId"])
	}
}

func TestNormalizeLegacyNestedPatchRestoresTheSavedUnifiedConfiguration(t *testing.T) {
	legacy := SettingsPatch{
		"patch": raw(map[string]any{
			"patch": map[string]any{
				"textModelId": "gpt-5.4",
				"publishSettings": map[string]any{
					"websiteProfileId": "121-330",
					"horizontalFlip":   true,
				},
			},
		}),
		"aiPromptConfig":    raw(map[string]any{"video": map[string]any{"presetId": "h3-director"}}),
		"expectedRevision": raw(3),
	}

	got := normalizeLegacyNestedPatch(legacy)
	if _, exists := got["patch"]; exists {
		t.Fatalf("legacy wrapper leaked into the effective settings: %#v", got)
	}
	if _, exists := got["expectedRevision"]; exists {
		t.Fatalf("request metadata leaked into the effective settings: %#v", got)
	}
	if string(got["textModelId"]) != "\"gpt-5.4\"" {
		t.Fatalf("text model not restored: %s", got["textModelId"])
	}
	var publish map[string]any
	if err := json.Unmarshal(got["publishSettings"], &publish); err != nil {
		t.Fatal(err)
	}
	if publish["websiteProfileId"] != "121-330" || publish["horizontalFlip"] != true {
		t.Fatalf("publish settings not restored: %#v", publish)
	}
	if string(got["aiPromptConfig"]) == "" {
		t.Fatalf("newer outer settings were not preserved: %#v", got)
	}
}
