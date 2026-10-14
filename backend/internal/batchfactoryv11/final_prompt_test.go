package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
)

func seedCompiledVideo(t *testing.T) (*MemoryStore, Batch, Book, Video) {
	t.Helper()
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"qualityEnabled":  rawSetting(t, true),
			"quality":         rawSetting(t, "高质量商业成片"),
			"negativeEnabled": rawSetting(t, true),
			"negative":        rawSetting(t, "不要水印和畸形手指"),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aspectRatio": rawSetting(t, "16:9")}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	video := book.Videos[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"duration": rawSetting(t, 8), "prefixEnabled": rawSetting(t, true), "prefix": rawSetting(t, "电影感运镜")}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	return store, batch, batch.Books[0], batch.Books[0].Videos[0]
}

func TestEffectiveSettingsResolveSystemBatchBookVideoWithSources(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	service := &PromptCompilerService{Store: store}
	value, _, _, _, err := service.ResolveEffective(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := rawString(value.Values, "aspectRatio", ""); got != "16:9" || value.SourceByField["aspectRatio"] != "book" {
		t.Fatalf("aspect=%q sources=%+v", got, value.SourceByField)
	}
	if got := rawInt(value.Values, "duration", 0); got != 8 || value.SourceByField["duration"] != "video" {
		t.Fatalf("duration=%d sources=%+v", got, value.SourceByField)
	}
	if value.SourceByField["productionMode"] != "batch" || value.SourceByField["subtitlePolicy"] != "system" {
		t.Fatalf("sources=%+v", value.SourceByField)
	}
	if value.DirectorRevisionID == "" || len(value.SnapshotHash) != 64 {
		t.Fatalf("effective=%+v", value)
	}
}

func TestFinalPromptUsesDirectorAssetsAndEffectiveConstraints(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	value, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"视频提示词：镜头画面：", "00:00-00:03 | 中景｜缓慢推轨 | 林晚进入客厅", "00:03-00:09 | 特写｜固定 | 她握紧玻璃杯", "林晚：18岁中国女性", "林家客厅：现代中式客厅", "玻璃杯：透明厚底玻璃杯", "画面前缀：电影感运镜", "画面约束提示词：高质量商业成片", "负面提示词：不要水印和畸形手指"} {
		if !strings.Contains(value.CompiledPrompt, expected) {
			t.Fatalf("missing %q in\n%s", expected, value.CompiledPrompt)
		}
	}
	if strings.Contains(value.CompiledPrompt, "画面主体：") {
		t.Fatalf("visual prompt leaked into final video prompt:\n%s", value.CompiledPrompt)
	}
	if value.DirectorRevisionID != book.DirectorRevision.ID || value.SnapshotHash != value.EffectiveSettings.SnapshotHash {
		t.Fatalf("prompt identity=%+v", value)
	}
}

func TestFinalPromptUsesH3VideoRendererWhenTheH3PresetIsSelected(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"video": map[string]any{
				"enabled":   true,
				"scope":     "all",
				"presetId":  "batch-video-h3-director",
				"presetKey": "h3-video-normal",
				"body":      "H3 director output renderer",
			},
		})},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}

	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"【人物定义】",
		"<Subject 1> 林晚：18岁中国女性",
		"【H3视听时间轴】",
		"00:00-00:09",
		"镜头1：中景；缓慢推轨；林晚<Subject 1>进入客厅",
		"No on-screen text unless explicitly required by the story.",
	} {
		if !strings.Contains(prompt.CompiledPrompt, expected) {
			t.Fatalf("H3 renderer missing %q in:\n%s", expected, prompt.CompiledPrompt)
		}
	}
	if strings.Contains(prompt.CompiledPrompt, "视频提示词：镜头画面：") {
		t.Fatalf("generic renderer leaked into H3 output:\n%s", prompt.CompiledPrompt)
	}
	// 基础设定开关开启时，H3 的人物定义统一走外层“基础设定”组件注入（保留 <Subject> 格式）。
	if !strings.Contains(prompt.CompiledPrompt, "基础设定：【人物定义】") {
		t.Fatalf("H3 base-setup must inject subject definitions under the 基础设定 component:\n%s", prompt.CompiledPrompt)
	}
	if !strings.Contains(prompt.DisplayPrompt, "镜头画面：") || strings.Contains(prompt.DisplayPrompt, "【H3视听时间轴】") {
		t.Fatalf("the editable card must keep the director prompt, display=%q", prompt.DisplayPrompt)
	}
}

func TestFinalPromptH3BaseSetupSwitchOffOmitsSubjectDefinitions(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"aiPromptConfig": rawSetting(t, map[string]any{
				"video": map[string]any{
					"enabled":   true,
					"scope":     "all",
					"presetId":  "batch-video-h3-director",
					"presetKey": "h3-video-normal",
					"body":      "H3 director output renderer",
				},
			}),
			"injectBaseSettings": rawSetting(t, false),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt.CompiledPrompt, "【人物定义】") || strings.Contains(prompt.CompiledPrompt, "<Subject 1>") {
		t.Fatalf("base-setup off must omit H3 subject definitions:\n%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptStripsSharedSectionsFromOverride(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch:            SettingsPatch{"videoPrompt": rawSetting(t, "【基础设定】\n人物：张三\n\n镜头画面：\n00:00 | 中景 | 动作\n\n负面提示词：模糊")},
		ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt.CompiledPrompt, "【基础设定】") || strings.Contains(prompt.CompiledPrompt, "张三") || strings.Contains(prompt.CompiledPrompt, "模糊") {
		t.Fatalf("override shared sections must be stripped:\n%s", prompt.CompiledPrompt)
	}
	if !strings.Contains(prompt.CompiledPrompt, "镜头画面：") {
		t.Fatalf("override body must remain:\n%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptPreviewUsesLatestFrozenH3Compilation(t *testing.T) {
	ctx := context.Background()
	store, batch, book := seedH3DirectorRevision(t)
	document := *book.DirectorRevision.Output.H3Director
	seedH3AudioMeasurement(t, store, batch, book, document)
	compiled, err := (&H3KernelService{Store: store}).Compile(ctx, "alice", batch.ID, book.ID, H3KernelCompileRequest{
		DirectorRevisionID: book.DirectorRevision.ID,
		AudioAssetID:       "audio-1",
		Preset:             completeH3CompileInput(document, H3CanonicalTimeline{}).Preset,
		Switches:           H3PromptSwitches{SmartUnified: true, BaseSetup: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	segment := compiled.Compilation.Compilation.Segments[0]
	prompt, err := (&PromptCompilerService{Store: store}).Compile(ctx, "alice", batch.ID, book.ID, compiled.Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.CompiledPrompt != segment.CompiledPrompt || prompt.SnapshotHash != segment.CompiledPromptHash {
		t.Fatalf("preview differs from latest frozen H3 submission prompt:\npreview=%#v\nsegment=%#v", prompt, segment)
	}
	if prompt.DisplayPrompt != segment.EditableCopy || prompt.CompilationID != compiled.Compilation.ID {
		t.Fatalf("preview lost separate editable copy or compilation identity: %#v", prompt)
	}
}

func TestH3RendererOmitsAssetDefinitionsWhenBaseSetupIsOff(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"video":       map[string]any{"enabled": true, "scope": "all", "presetKey": h3VideoRendererKey},
			"constraints": map[string]any{"enabled": true, "scope": "all", "baseSetup": map[string]any{"enabled": false}},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}

	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"【人物定义】", "18岁中国女性", "现代中式客厅"} {
		if strings.Contains(prompt.CompiledPrompt, forbidden) {
			t.Fatalf("base setup is off but asset prompt %q entered H3 output:\n%s", forbidden, prompt.CompiledPrompt)
		}
	}
	if !strings.Contains(prompt.CompiledPrompt, "林晚进入客厅") {
		t.Fatalf("base setup must not remove the director's character name and action:\n%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptExposesReadOnlyBaseSetupForTheStoryboardCard(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"林晚：18岁中国女性", "林家客厅：现代中式客厅"} {
		if !strings.Contains(prompt.BaseSetupPrompt, expected) {
			t.Fatalf("the enabled base setup must be exposed separately for the storyboard card; missing %q in %s", expected, prompt.BaseSetupPrompt)
		}
	}
	if strings.Contains(prompt.BaseSetupPrompt, "玻璃杯：透明厚底玻璃杯") {
		t.Fatalf("the storyboard-card base setup only shows people and scenes, got %s", prompt.BaseSetupPrompt)
	}
}

func TestH3VideoRendererRequiresTheExplicitSmartUnifiedConstraintForItsVisualBaseline(t *testing.T) {
	book := Book{DirectorRevision: &DirectorRevision{Output: DirectorResult{SmartUnifiedStyle: "影像媒介：数字电影；光线与明暗层次：冷暖对照"}}}
	config := AIReasoningPromptConfig{Video: AIReasoningPromptModule{PresetSnapshot: PresetSnapshot{Key: h3VideoRendererKey}}}
	if got := smartUnifiedStyleForRevision(book, config); got != "" {
		t.Fatalf("H3 must not use a visual baseline without the explicit setting, got %q", got)
	}
}

func TestH3RendererCompilesSubjectRhythmAndAudioFromDirectorOutput(t *testing.T) {
	raw := strings.Replace(validDirectorJSON(),
		`"camera":"缓慢推轨","description":"林晚进入客厅"`,
		`"camera":"缓慢推轨","rhythm":"克制推进","audio":"林晚（低声）：“我回来了。”","description":"林晚进入客厅"`,
		1,
	)
	result, err := NormalizeDirectorOutput(json.RawMessage(raw), DirectorSettings{
		MaxVideoDuration: 10,
		AspectRatio:      "9:16",
		AllowedPrefixKeys: []string{
			"modern_conflict",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	prompt := compileH3VideoPrompt(result.Storyboard[0], []compiledAsset{{Name: "林晚", Prompt: "18岁中国女性，黑色长发。"}}, []compiledAsset{{Name: "林家客厅", Prompt: "现代中式客厅。"}})
	for _, expected := range []string{
		"画面：镜头1（节奏：克制推进）：中景；缓慢推轨；林晚<Subject 1>进入客厅",
		"Audio:",
		"镜头1：林晚<Subject 1>（低声）：“我回来了。”",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("H3 canonical output missing %q in:\n%s", expected, prompt)
		}
	}
}

func TestH3RendererKeepsPerShotVisualContextAndLighting(t *testing.T) {
	raw := strings.Replace(validDirectorJSON(),
		`"camera":"缓慢推轨","description":"林晚进入客厅"`,
		`"camera":"缓慢推轨","visual_context":"现代中式客厅，落地窗透入阴天冷光，玻璃杯放在茶几边缘","lighting":"侧向冷光勾勒人物轮廓，暗部保留木材纹理","description":"林晚进入客厅"`,
		1,
	)
	result, err := NormalizeDirectorOutput(json.RawMessage(raw), DirectorSettings{
		MaxVideoDuration: 10, AspectRatio: "9:16", AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err != nil {
		t.Fatal(err)
	}
	prompt := compileH3VideoPrompt(result.Storyboard[0], []compiledAsset{{Name: "林晚", Prompt: "18岁中国女性，黑色长发。"}}, nil)
	for _, expected := range []string{
		"场景：现代中式客厅，落地窗透入阴天冷光，玻璃杯放在茶几边缘",
		"光影：侧向冷光勾勒人物轮廓，暗部保留木材纹理",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("H3 rich visual timeline missing %q in:\n%s", expected, prompt)
		}
	}
}

func TestFinalPromptPreviewKeepsShowingAnExistingOverLimitStoryboard(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"duration": rawSetting(t, 12), "storyboardDurationLimit": rawSetting(t, 10)}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book, video = batch.Books[0], batch.Books[0].Videos[0]
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatalf("preview should remain readable for an old over-limit storyboard: %v", err)
	}
	if prompt.DurationSeconds != 12 {
		t.Fatalf("duration=%d", prompt.DurationSeconds)
	}
	if len(prompt.EffectiveSettings.Compatibility) == 0 || prompt.EffectiveSettings.Compatibility[0].State != "incompatible" {
		t.Fatalf("preview must retain the production warning: %+v", prompt.EffectiveSettings.Compatibility)
	}
	if _, err := (&PromptCompilerService{Store: store}).CompileForProduction(context.Background(), "alice", batch.ID, book.ID, video.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("production must still reject the over-limit storyboard: %v", err)
	}
}

func TestFinalPromptIgnoresRetiredVideoStylePrefixSnapshot(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"prefix": map[string]any{"enabled": true, "scope": "all", "presetId": "batch-prefix-modern-conflict", "presetKey": "modern_conflict", "body": "FROZEN PREFIX BODY"},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt.CompiledPrompt, "FROZEN PREFIX BODY") {
		t.Fatalf("retired video style prefix leaked into final VIDEO prompt: %s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptMarksSmartUnifiedAsPendingUntilTheDirectorAnalysisExists(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	constraints := map[string]any{"enabled": true, "scope": "all", "selections": []any{
		map[string]any{"presetId": smartUnifiedPrefixPresetID, "constraintCategory": "prefix", "body": "智能统一元提示词不能进入视频提示词"},
	}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"constraints": constraints})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !prompt.SmartUnifiedPending {
		t.Fatal("smart unified without a fresh director analysis must be reported as pending")
	}
	if strings.Contains(prompt.CompiledPrompt, "智能统一元提示词不能进入视频提示词") {
		t.Fatalf("smart unified meta prompt leaked into VIDEO prompt: %s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptIncludesSelectedAIReasoningConstraintLayers(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	constraints := map[string]any{
		"enabled": true,
		"scope":   "all",
		"selections": []any{
			map[string]any{"constraintCategory": "prefix", "body": "统一冷青电影光"},
			map[string]any{"constraintCategory": "quality", "body": "4K 细节清晰"},
			map[string]any{"constraintCategory": "restriction", "body": "禁止镜头跳轴"},
			map[string]any{"constraintCategory": "negative", "body": "不要字幕与水印"},
		},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"constraints": constraints})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"画面前缀：统一冷青电影光", "画面约束提示词：4K 细节清晰", "画面限制：禁止镜头跳轴", "负面提示词：不要字幕与水印"} {
		if !strings.Contains(prompt.CompiledPrompt, expected) {
			t.Fatalf("missing selected constraint %q in\n%s", expected, prompt.CompiledPrompt)
		}
	}
}

func TestFinalPromptUsesTheConfirmedVideoInputOrder(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	constraints := map[string]any{
		"enabled":   true,
		"scope":     "all",
		"baseSetup": map[string]any{"enabled": true},
		"selections": []any{
			map[string]any{"constraintCategory": "prefix", "body": "PREFIX"},
			map[string]any{"constraintCategory": "quality", "body": "QUALITY"},
			map[string]any{"constraintCategory": "restriction", "body": "RESTRICTION"},
			map[string]any{"constraintCategory": "negative", "body": "NEGATIVE"},
		},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"constraints": constraints})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"画面前缀：PREFIX", "基础设定：", "画面约束提示词：QUALITY", "视频提示词：镜头画面：", "画面限制：RESTRICTION", "负面提示词：NEGATIVE"}
	last := -1
	for _, part := range want {
		at := strings.Index(prompt.CompiledPrompt, part)
		if at < 0 || at <= last {
			t.Fatalf("compiled order must be prefix/base/quality/video/restriction/negative; missing or misplaced %q in:\n%s", part, prompt.CompiledPrompt)
		}
		last = at
	}
	if strings.Contains(prompt.CompiledPrompt, "字幕策略") || strings.Contains(prompt.CompiledPrompt, "视频规格") {
		t.Fatalf("non-prompt transport fields must not be displayed in the final input: %s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptWithConstraintsDisabledContainsOnlyStoryboardPrompt(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"constraints": map[string]any{"enabled": false}})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.CompiledPrompt != "视频提示词：镜头画面：\n00:00-00:03 | 中景｜缓慢推轨 | 林晚进入客厅\n00:03-00:09 | 特写｜固定 | 她握紧玻璃杯\n资产：@林晚 @林家客厅 @玻璃杯" {
		t.Fatalf("constraints disabled must leave only storyboard prompt, got:\n%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptUsesPersistedManualBookAsset(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	var character BookAsset
	for _, asset := range book.AssetRecords {
		if asset.Kind == "character" && asset.Name == "林晚" {
			character = asset
			break
		}
	}
	if character.ID == "" {
		t.Fatalf("assets=%+v", book.AssetRecords)
	}
	if _, err := store.UpdateBookAsset(context.Background(), "alice", batch.ID, book.ID, character.ID, UpdateBookAssetInput{Name: character.Name, Prompt: "用户手动确认的角色一致性 Prompt", ExpectedRevision: character.Revision}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt.CompiledPrompt, "林晚：用户手动确认的角色一致性 Prompt") || strings.Contains(prompt.CompiledPrompt, "林晚：18岁中国女性") {
		t.Fatalf("compiled prompt=%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptExcludesDeselectedStoryboardAssetFromTextAndReferences(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	var character, scene BookAsset
	for _, asset := range book.AssetRecords {
		switch {
		case asset.Kind == "character" && asset.Name == "林晚":
			character = asset
		case asset.Kind == "scene" && asset.Name == "林家客厅":
			scene = asset
		}
	}
	if character.ID == "" || scene.ID == "" {
		t.Fatalf("book asset records=%+v", book.AssetRecords)
	}
	for _, asset := range []BookAsset{character, scene} {
		if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, asset.ID, CreateBookAssetImageInput{URL: "https://images.example/" + asset.ID + ".png", MediaType: "image/png", Source: "provider"}); err != nil {
			t.Fatal(err)
		}
	}
	selection := map[string]any{
		"autoAssetIds":     []string{character.ID, scene.ID},
		"addedAssetIds":    []string{},
		"excludedAssetIds": []string{character.ID},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"assetSelection": rawSetting(t, selection)}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, latest.Books[0].Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(compiled.CompiledPrompt, "林晚：18岁中国女性") || slices.Contains(compiled.ReferenceImageURLs, "https://images.example/"+character.ID+".png") {
		t.Fatalf("deselected character leaked into production request: %+v\n%s", compiled.ReferenceImageURLs, compiled.CompiledPrompt)
	}
	if strings.Contains(compiled.CompiledPrompt, "@林晚") {
		t.Fatalf("deselected character asset mention leaked into production request: %s", compiled.CompiledPrompt)
	}
	if !slices.Contains(compiled.ReferenceImageURLs, "https://images.example/"+scene.ID+".png") {
		t.Fatalf("selected scene image missing: %+v", compiled.ReferenceImageURLs)
	}
	if !strings.Contains(compiled.CompiledPrompt, "@林家客厅") {
		t.Fatalf("selected scene asset mention missing from production request: %s", compiled.CompiledPrompt)
	}
}

func TestExplicitAssetMentionsLimitVideoReferenceImages(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	var character, scene BookAsset
	for _, asset := range book.AssetRecords {
		if asset.Kind == "character" && asset.Name == "林晚" {
			character = asset
		}
		if asset.Kind == "scene" && asset.Name == "林家客厅" {
			scene = asset
		}
	}
	for _, asset := range []BookAsset{character, scene} {
		if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, asset.ID, CreateBookAssetImageInput{URL: "https://images.example/" + asset.ID + ".png", MediaType: "image/png", Source: "provider"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"videoPrompt": rawSetting(t, "镜头中出现 @林家客厅。")}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	compiled, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(compiled.ReferenceImageURLs, "https://images.example/"+character.ID+".png") {
		t.Fatalf("unmentioned @ asset was uploaded: %+v", compiled.ReferenceImageURLs)
	}
	if !slices.Contains(compiled.ReferenceImageURLs, "https://images.example/"+scene.ID+".png") {
		t.Fatalf("mentioned asset image missing: %+v", compiled.ReferenceImageURLs)
	}
}

func TestFinalPromptOmitsTextForImageBackedAssets(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	var scene BookAsset
	for _, asset := range book.AssetRecords {
		if asset.Kind == "scene" && asset.Name == "林家客厅" {
			scene = asset
		}
	}
	if scene.ID == "" {
		t.Fatalf("scene asset not found: %+v", book.AssetRecords)
	}
	if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, scene.ID, CreateBookAssetImageInput{URL: "https://images.example/scene-primary.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}
	compiled, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 有主图的场景：走图片生视频，文字设定不发。
	if strings.Contains(compiled.CompiledPrompt, "林家客厅：现代中式客厅") {
		t.Fatalf("image-backed scene text must be omitted:\n%s", compiled.CompiledPrompt)
	}
	if !slices.Contains(compiled.ReferenceImageURLs, "https://images.example/scene-primary.png") {
		t.Fatalf("scene primary image missing: %+v", compiled.ReferenceImageURLs)
	}
	// 没图片的人物：文字照常发，走文生视频。
	if !strings.Contains(compiled.CompiledPrompt, "林晚：18岁中国女性") {
		t.Fatalf("text-only character prompt must remain:\n%s", compiled.CompiledPrompt)
	}
}

func TestFinalPromptIgnoresLegacyAssetPromptDraftWhenAssetRecordExists(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveDraft(context.Background(), "alice", Draft{Key: "asset:character:林晚", Kind: "asset-prompt", Scope: batch.ID, Content: "过期草稿"}); err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt.CompiledPrompt, "过期草稿") || !strings.Contains(prompt.CompiledPrompt, "林晚：18岁中国女性") {
		t.Fatalf("compiled prompt=%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptUsesSavedVideoPromptOverrideWithoutVisualPrompt(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	override := "镜头提示词已由用户确认，保持人物连续性。"
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"videoPrompt": rawSetting(t, override)}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, latest.Books[0].Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt.CompiledPrompt, "视频提示词："+override) {
		t.Fatalf("video prompt override missing=%s", prompt.CompiledPrompt)
	}
	if strings.Contains(prompt.CompiledPrompt, "画面主体：") {
		t.Fatalf("visual prompt leaked into video prompt=%s", prompt.CompiledPrompt)
	}
}

func TestFinalPromptRejectsOrphanedVideoIdentity(t *testing.T) {
	store, batch, book, oldVideo := seedCompiledVideo(t)
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, oldVideo.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for orphaned VIDEO, got %v", err)
	}
}

func TestSDDirectorReturnsPlainTextCards(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"video": map[string]any{
				"enabled":   true,
				"scope":     "all",
				"presetId":  "batch-video-sd",
				"presetKey": "sd-video-normal",
				"body":      "SD视频提示词预设正文",
			},
		})},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]

	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：全程无台词，仅靠眼神与动作。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，缓慢推轨，新娘低头抚平嫁衣袖口。`

	// 真实流程：先点“智能预设”提取资产，再选 SD 重新生成导演分镜。
	assetJSON := `{"characters":[{"name":"林溪","prompt":"短发女主，右耳一颗小痣"}],"scenes":[{"name":"客厅","prompt":"现代中式客厅"}],"props":[]}`
	provider := &queuedDirectorProvider{values: []string{assetJSON, sdResponse}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunAssetExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	revision, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(revision.Output.Storyboard) != 1 {
		t.Fatalf("storyboard count=%d, want 1", len(revision.Output.Storyboard))
	}
	card := revision.Output.Storyboard[0]
	if card.FinalPrompt == "" {
		t.Fatal("FinalPrompt must be populated for SD plain-text response")
	}
	if !strings.Contains(card.FinalPrompt, "段内执行约束") {
		t.Fatalf("FinalPrompt missing SD content:\n%s", card.FinalPrompt)
	}
	if strings.Contains(card.FinalPrompt, "统一风格") {
		t.Fatalf("FinalPrompt must not contain 统一风格 (injected by switch):\n%s", card.FinalPrompt)
	}
	if strings.Contains(card.FinalPrompt, "最终导出画质约束") {
		t.Fatalf("FinalPrompt must not contain quality section (injected by switch):\n%s", card.FinalPrompt)
	}
	if card.DurationSec != 10 {
		t.Fatalf("duration=%d, want 10", card.DurationSec)
	}

	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	video := book.Videos[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"constraints": map[string]any{
			"enabled":   true,
			"scope":     "all",
			"baseSetup": map[string]any{"enabled": true},
			"selections": []any{
				map[string]any{"constraintCategory": "prefix", "body": "统一的都市电影质感"},
				map[string]any{"constraintCategory": "quality", "body": "4K级细节"},
				map[string]any{"constraintCategory": "restriction", "body": "镜头不得跳轴"},
				map[string]any{"constraintCategory": "negative", "body": "不要字幕和水印"},
			},
		}})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	video = book.Videos[0]
	compiled, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(compiled.CompiledPrompt, "段内执行约束") {
		t.Fatalf("CompiledPrompt must contain SD body:\n%s", compiled.CompiledPrompt)
	}
	if !strings.Contains(compiled.DisplayPrompt, "段内执行约束") {
		t.Fatalf("DisplayPrompt must show SD card body:\n%s", compiled.DisplayPrompt)
	}
	// 基础设定默认开启：必须注入已提取的人物与场景。
	if !strings.Contains(compiled.CompiledPrompt, "基础设定") || !strings.Contains(compiled.CompiledPrompt, "林溪：短发女主，右耳一颗小痣") {
		t.Fatalf("基础设定必须注入已提取人物:\n%s", compiled.CompiledPrompt)
	}
	if !strings.Contains(compiled.CompiledPrompt, "客厅：现代中式客厅") {
		t.Fatalf("基础设定必须注入已提取场景:\n%s", compiled.CompiledPrompt)
	}
	want := []string{"基础设定：", "画面质量提示词：统一的都市电影质感\n4K级细节", "视频提示词：", "画面限制：镜头不得跳轴", "负面提示词：不要字幕和水印"}
	last := -1
	for _, part := range want {
		at := strings.Index(compiled.CompiledPrompt, part)
		if at < 0 || at <= last {
			t.Fatalf("SD card must keep the legacy base/quality/video/restriction/negative order; missing or misplaced %q in:\n%s", part, compiled.CompiledPrompt)
		}
		last = at
	}
	if strings.Contains(compiled.CompiledPrompt, "画面前缀：") {
		t.Fatalf("SD card must not expose a separate picture-prefix section:\n%s", compiled.CompiledPrompt)
	}
}

func TestSelectedAssetsSuppressesRepeatedNames(t *testing.T) {
	assets := selectedAssets("prop", []string{"茶杯", "茶杯"}, map[string]compiledAsset{
		assetKey("prop", "茶杯"): {ID: "prop-1", Kind: "prop", Name: "茶杯", Prompt: "白瓷茶杯"},
	}, map[string]string{}, nil, false)
	if got := selectResolvedPrompts(assets, map[string]bool{}); got != "茶杯：白瓷茶杯" {
		t.Fatalf("duplicate asset names must only appear once, got %q", got)
	}
}

func TestSDDirectorSplitsMultipleVideosByDurationLimit(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"video": map[string]any{"enabled": true, "scope": "all", "presetId": "batch-video-sd", "presetKey": "sd-video-normal", "body": "SD预设"},
		})},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]

	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：按场景编号依次完成。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，林晚推开门。

===VIDEO 02===
时长：4秒
段内执行约束：按场景编号依次完成。
[场景 6] 总时长：4.000秒
[镜头 1] 近景，林晚转身。`

	provider := &queuedDirectorProvider{values: []string{sdResponse}}
	revision, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(revision.Output.Storyboard) != 2 {
		t.Fatalf("storyboard count=%d, want 2", len(revision.Output.Storyboard))
	}
	if revision.Output.Storyboard[0].DurationSec != 10 || revision.Output.Storyboard[1].DurationSec != 4 {
		t.Fatalf("durations=%d,%d", revision.Output.Storyboard[0].DurationSec, revision.Output.Storyboard[1].DurationSec)
	}
	if !strings.Contains(revision.Output.Storyboard[1].FinalPrompt, "林晚转身") {
		t.Fatalf("second card body wrong:\n%s", revision.Output.Storyboard[1].FinalPrompt)
	}
	if !strings.Contains(revision.Videos[1].VideoPrompt, "[场景 6]") {
		t.Fatalf("second video prompt wrong:\n%s", revision.Videos[1].VideoPrompt)
	}
}

func TestNormalDirectorStillUsesJSONContract(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON()}}
	_, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	if len(book.Videos) == 0 {
		t.Fatal("normal JSON contract must still produce videos")
	}
	if book.Videos[0].VideoPrompt == "" {
		t.Fatal("normal JSON contract must still produce video prompts")
	}
}

func TestCompiledVideoInputUsesAssetImagesBeforeAssetText(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	var character BookAsset
	for _, asset := range book.AssetRecords {
		if asset.Kind == "character" && asset.Name == "林晚" {
			character = asset
			break
		}
	}
	if character.ID == "" {
		t.Fatalf("book asset records=%+v", book.AssetRecords)
	}
	if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, character.ID, CreateBookAssetImageInput{URL: "https://images.example/lin.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}

	input, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(input.CompiledPrompt, "林晚：18岁中国女性") {
		t.Fatalf("primary asset image still injected text: %s", input.CompiledPrompt)
	}
	if !slices.Contains(input.ReferenceImageURLs, "https://images.example/lin.png") {
		t.Fatalf("reference images=%+v", input.ReferenceImageURLs)
	}
}

func TestCompiledVideoInputKeepsAnAssetReferenceImageWhenItsPromptIsBlank(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	var character BookAsset
	for _, asset := range book.AssetRecords {
		if asset.Kind == "character" && asset.Name == "林晚" {
			character = asset
			break
		}
	}
	if character.ID == "" {
		t.Fatalf("book asset records=%+v", book.AssetRecords)
	}
	if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, character.ID, CreateBookAssetImageInput{URL: "https://images.example/lin-image-only.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}
	// An image remains a valid VIDEO reference even if its source prompt was
	// deliberately cleared. The "has image, do not send text" rule must not
	// accidentally turn that asset into no input at all.
	owned := store.bookAssets[character.ID]
	owned.Value.Prompt = ""
	store.bookAssets[character.ID] = owned

	input, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(input.ReferenceImageURLs, "https://images.example/lin-image-only.png") {
		t.Fatalf("image-only asset was dropped from references=%+v", input.ReferenceImageURLs)
	}
	if strings.Contains(input.CompiledPrompt, "林晚：") {
		t.Fatalf("blank image-only asset leaked text=%s", input.CompiledPrompt)
	}
}

func TestCompiledVideoInputUsesDroppedAssetTextWhenReferenceLimitExceeded(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	for _, asset := range book.AssetRecords {
		if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, asset.ID, CreateBookAssetImageInput{URL: "https://images.example/" + asset.Kind + ".png", MediaType: "image/png", Source: "provider"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"referenceImageLimit": rawSetting(t, 2)}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	input, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, latest.Books[0].Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	var propID string
	for _, asset := range book.AssetRecords {
		if asset.Kind == "prop" && asset.Name == "玻璃杯" {
			propID = asset.ID
			break
		}
	}
	if len(input.ReferenceImageURLs) != 2 || !slices.Contains(input.DowngradedAssetIDs, propID) {
		t.Fatalf("compiled input=%+v", input)
	}
	if !strings.Contains(input.CompiledPrompt, "玻璃杯：透明厚底玻璃杯") {
		t.Fatalf("dropped asset text missing: %s", input.CompiledPrompt)
	}
	if strings.Contains(input.CompiledPrompt, "画面主体：") {
		t.Fatalf("visual prompt leaked into video prompt: %s", input.CompiledPrompt)
	}
}

func TestFinalPromptAllowsPureTextVideoWhenBaseSetupIsDisabled(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"constraints": map[string]any{"baseSetup": map[string]any{"enabled": false}}})}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, latest.Books[0].Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt.CompiledPrompt, "视频提示词：镜头画面：") || !strings.Contains(prompt.CompiledPrompt, "00:00-00:03 | 中景｜缓慢推轨 | 林晚进入客厅") {
		t.Fatalf("video prompt missing: %s", prompt.CompiledPrompt)
	}
	for _, forbidden := range []string{"林晚：18岁中国女性", "林家客厅：现代中式客厅", "玻璃杯：透明厚底玻璃杯"} {
		if strings.Contains(prompt.CompiledPrompt, forbidden) {
			t.Fatalf("base setup disabled but asset prompt leaked %q in %s", forbidden, prompt.CompiledPrompt)
		}
	}
	if len(prompt.ReferenceImageURLs) != 0 {
		t.Fatalf("pure text video should not require references: %+v", prompt.ReferenceImageURLs)
	}
}

func TestFinalPromptBaseSetupUsesOnlyAssetsBoundToTheCurrentShot(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, CreateBookAssetInput{Kind: "character", Name: "顾遥", Prompt: "短发女律师，深色西装，冷静坚定"}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, latest.Books[0].Videos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt.CompiledPrompt, "顾遥：短发女律师，深色西装，冷静坚定") || strings.Contains(prompt.BaseSetupPrompt, "顾遥：短发女律师，深色西装，冷静坚定") {
		t.Fatalf("an unbound book asset must not leak into this VIDEO prompt or its card setup:\nprovider=%s\ncard=%s", prompt.CompiledPrompt, prompt.BaseSetupPrompt)
	}
	_ = video
}

func TestEffectiveSettingsMergesBookPromptModulesWithTheBatchH3VideoPreset(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"video":       map[string]any{"enabled": true, "scope": "all", "presetId": "batch-video-h3-director", "presetKey": h3VideoRendererKey, "body": "H3 renderer"},
			"constraints": map[string]any{"enabled": true, "scope": "all", "baseSetup": map[string]any{"enabled": true}},
		})}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book = latest.Books[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
			"scriptComposition": map[string]any{"general": map[string]any{"presetId": "script-general", "body": "book-only storyboard composition"}},
		})}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	latest, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, latest.Books[0].ID, video.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt.CompiledPrompt, "【H3视听时间轴】") {
		t.Fatalf("a partial book prompt config must retain the batch H3 VIDEO renderer, got:\n%s", prompt.CompiledPrompt)
	}
}

func TestCompileForOpeningVariant(t *testing.T) {
	store, batch, book, video := seedCompiledSDOpening(t, []OpeningVariant{
		{Index: 1, Label: "分镜一 | 换开头1", Prompt: "变体开场", Status: "success"},
		{Index: 2, Label: "分镜一 | 换开头2", Prompt: "变体二", Status: "failed"},
	})

	t.Run("variant overrides VIDEO01 prompt", func(t *testing.T) {
		prompt, err := (&PromptCompilerService{Store: store}).CompileForOpeningVariant(context.Background(), "alice", batch.ID, book.ID, video.ID, 1, -1)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(prompt.CompiledPrompt, "变体开场") {
			t.Fatalf("expected compiled prompt to contain variant prompt, got:\n%s", prompt.CompiledPrompt)
		}
		if strings.Contains(prompt.CompiledPrompt, "原开场") {
			t.Fatalf("expected compiled prompt NOT to contain original prompt, got:\n%s", prompt.CompiledPrompt)
		}
		if !strings.Contains(prompt.DisplayPrompt, "变体开场") {
			t.Fatalf("expected display prompt to contain variant prompt, got:\n%s", prompt.DisplayPrompt)
		}
	})

	t.Run("variantIndex 0 uses original prompt", func(t *testing.T) {
		prompt, err := (&PromptCompilerService{Store: store}).CompileForOpeningVariant(context.Background(), "alice", batch.ID, book.ID, video.ID, 0, -1)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(prompt.CompiledPrompt, "原开场") {
			t.Fatalf("expected compiled prompt to contain original prompt, got:\n%s", prompt.CompiledPrompt)
		}
		if strings.Contains(prompt.CompiledPrompt, "变体开场") {
			t.Fatalf("expected compiled prompt NOT to contain variant prompt, got:\n%s", prompt.CompiledPrompt)
		}
	})

	t.Run("normal Compile uses original prompt", func(t *testing.T) {
		prompt, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(prompt.CompiledPrompt, "原开场") {
			t.Fatalf("expected compiled prompt to contain original prompt, got:\n%s", prompt.CompiledPrompt)
		}
	})

	t.Run("nonexistent variant returns ErrConflict", func(t *testing.T) {
		_, err := (&PromptCompilerService{Store: store}).CompileForOpeningVariant(context.Background(), "alice", batch.ID, book.ID, video.ID, 99, -1)
		if !errors.Is(err, ErrConflict) {
			t.Fatalf("expected ErrConflict for nonexistent variant, got %v", err)
		}
	})

	t.Run("failed variant returns ErrConflict", func(t *testing.T) {
		_, err := (&PromptCompilerService{Store: store}).CompileForOpeningVariant(context.Background(), "alice", batch.ID, book.ID, video.ID, 2, -1)
		if !errors.Is(err, ErrConflict) {
			t.Fatalf("expected ErrConflict for failed variant, got %v", err)
		}
	})

	t.Run("opening variant on non-VIDEO01 returns ErrInvalid", func(t *testing.T) {
		video2 := book.Videos[1]
		_, err := (&PromptCompilerService{Store: store}).CompileForOpeningVariant(context.Background(), "alice", batch.ID, book.ID, video2.ID, 1, -1)
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected ErrInvalid for non-VIDEO01 variant request, got %v", err)
		}
	})

	t.Run("component injection consistent with normal compile", func(t *testing.T) {
		if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
			Patch: SettingsPatch{
				"qualityEnabled":  rawSetting(t, true),
				"quality":         rawSetting(t, "高质量商业成片"),
				"negativeEnabled": rawSetting(t, true),
				"negative":        rawSetting(t, "不要水印和畸形手指"),
			},
			ExpectedRevision: batch.Revision,
		}); err != nil {
			t.Fatal(err)
		}
		batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
		book = batch.Books[0]
		video = book.Videos[0]

		normal, err := (&PromptCompilerService{Store: store}).Compile(context.Background(), "alice", batch.ID, book.ID, video.ID)
		if err != nil {
			t.Fatal(err)
		}
		variant, err := (&PromptCompilerService{Store: store}).CompileForOpeningVariant(context.Background(), "alice", batch.ID, book.ID, video.ID, 1, -1)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(variant.CompiledPrompt, "画面质量提示词：高质量商业成片") {
			t.Fatalf("variant missing quality component:\n%s", variant.CompiledPrompt)
		}
		if !strings.Contains(variant.CompiledPrompt, "负面提示词：不要水印和畸形手指") {
			t.Fatalf("variant missing negative component:\n%s", variant.CompiledPrompt)
		}

		normalKeys := make([]string, len(normal.Components))
		for i, c := range normal.Components {
			normalKeys[i] = c.Key
		}
		variantKeys := make([]string, len(variant.Components))
		for i, c := range variant.Components {
			variantKeys[i] = c.Key
		}
		if !slices.Equal(normalKeys, variantKeys) {
			t.Fatalf("component order mismatch: normal=%v variant=%v", normalKeys, variantKeys)
		}
	})
}

func seedCompiledSDOpening(t *testing.T, variants []OpeningVariant) (*MemoryStore, Batch, Book, Video) {
	t.Helper()
	store, batch, book := seedOpeningBook(t, true, 2, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原开场"},
		{DurationSec: 10, FinalPrompt: "分镜二"},
	})
	video := book.Videos[0]
	patch := SettingsPatch{}
	if variants != nil {
		patch["openingVariants"] = rawSetting(t, variants)
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: patch, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	video = book.Videos[0]
	return store, batch, book, video
}
