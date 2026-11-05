package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestParseComicBootstrapOutputNormalizesVideoOnlyTimingFields(t *testing.T) {
	source := "林晚走进客厅。"
	boundaries := EnumerateComicSourceBoundaries(source)
	rawOutput := json.RawMessage(`{
		"characters":[{"name":"林晚","prompt":"现代女性"}],
		"scenes":[{"name":"客厅","prompt":"现代客厅"}],
		"props":[],
		"storyboard":[{
			"duration_sec":"1秒",
			"characters":["林晚"],
			"props":[],
			"scene":"客厅",
			"prefix_key":"general_anime",
			"shots":[{"start_sec":"0秒","end_sec":"1秒","description":"林晚走进客厅"}],
			"screen_prompt":"单页漫画导演说明",
			"image_prompt":"林晚走进客厅，完整中文漫画页",
			"end_boundary_id":"b1",
			"roles":["林晚"],
			"scene_presets":["客厅"]
		}],
		"source_coverage":{"source_complete":true,"has_remaining_source":false}
	}`)

	director, pages, segments, err := parseComicBootstrapOutput(rawOutput, DirectorSnapshot{MaxVideoDuration: 10}, source, boundaries)
	if err != nil {
		t.Fatal(err)
	}
	if len(director.Storyboard) != 1 || director.Storyboard[0].DurationSec != 1 {
		t.Fatalf("comic timing was not normalized: %+v", director.Storyboard)
	}
	if len(director.Storyboard[0].Shots) != 1 || director.Storyboard[0].Shots[0].StartSec != 0 || director.Storyboard[0].Shots[0].EndSec != 1 {
		t.Fatalf("comic page shot was not normalized: %+v", director.Storyboard[0].Shots)
	}
	if len(pages) != 1 || pages[0].ImagePrompt == "" {
		t.Fatalf("comic prompts were not preserved: %+v", pages)
	}
	if len(segments) != 1 || segments[0].SourceText != source {
		t.Fatalf("comic source segment was not preserved: %+v", segments)
	}
}

func TestBuildComicBootstrapContractUsesContentBoundariesInsteadOfFixedPageCount(t *testing.T) {
	book := Book{ID: "book-1", Title: "小妾造反后", SourceText: strings.Repeat("完整正文。", 700)}
	snapshot := DirectorSnapshot{Effective: SettingsPatch{
		"productionMode":   rawSetting(t, "comic"),
		"imageAspectRatio": rawSetting(t, "9:16"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
			"enabled": true,
			"body":    "生成完整中文漫画页",
			"assets":  map[string]any{"body": "提取人物场景道具"},
		}}),
	}}

	contract, err := BuildComicBootstrapContract(book, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(contract.SystemPrompt, "严格输出 30 个漫画页") || strings.Contains(contract.UserPrompt, "target_page_count") {
		t.Fatalf("bootstrap prompt still locks a fixed page count: system=%s user=%s", contract.SystemPrompt, contract.UserPrompt)
	}
	var payload struct {
		PagePlan struct {
			MinimumPageCount     int `json:"minimum_page_count"`
			RecommendedPageCount int `json:"recommended_page_count"`
		} `json:"page_plan"`
	}
	if err := json.Unmarshal([]byte(contract.UserPrompt), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.PagePlan.MinimumPageCount <= 0 || payload.PagePlan.RecommendedPageCount < payload.PagePlan.MinimumPageCount {
		t.Fatalf("bootstrap prompt does not carry a dynamic page plan: %+v", payload.PagePlan)
	}
	if !strings.Contains(contract.SystemPrompt, "minimum_page_count") || !strings.Contains(contract.SystemPrompt, "recommended_page_count") {
		t.Fatalf("bootstrap protocol does not enforce the dynamic page plan: %s", contract.SystemPrompt)
	}
	if !strings.Contains(contract.SystemPrompt, "end_boundary_id") || !strings.Contains(contract.UserPrompt, "b1") {
		t.Fatalf("bootstrap prompt does not carry safe source boundaries: system=%s user=%s", contract.SystemPrompt, contract.UserPrompt)
	}
}

func TestRunComicBootstrapRequiresComicAssetsFromTheAssetStage(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "林晚走进办公室。"}}})
	if err != nil {
		t.Fatal(err)
	}
	config := map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "漫画分页与提示词规则",
		"assets":  map[string]any{"body": "提取漫画资产"},
	}}
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{`{"characters":[],"scenes":[],"props":[],"storyboard":[]}`}}
	_, err = (&DirectorService{Store: store, Provider: provider}).RunComicBootstrap(ctx, "alice", batch.ID, batch.Books[0].ID)
	if err == nil || !strings.Contains(err.Error(), "请先提取漫画资产") {
		t.Fatalf("bootstrap error=%v, want comic asset stage prerequisite", err)
	}
	if len(provider.calls) != 0 {
		t.Fatalf("comic bootstrap called the text provider before comic assets existed: %d calls", len(provider.calls))
	}
}

func TestRunComicBootstrapSegmentsThenExtractsPromptsUsingExistingComicAssets(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "林晚走进办公室。"}}})
	if err != nil {
		t.Fatal(err)
	}
	config := map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "漫画分页与提示词规则",
		"assets":  map[string]any{"body": "提取漫画资产"},
	}}
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"productionMode":       rawSetting(t, "comic"),
			"aiPromptConfig":       rawSetting(t, config),
			"audioPlanningEnabled": rawSetting(t, true),
		}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBookAsset(ctx, "alice", batch.ID, batch.Books[0].ID, CreateBookAssetInput{Kind: "character", Source: "comic", Name: "林晚", Prompt: "黑发女性"}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		`{"boundaryIds":["b1"]}`,
		`{"screen_prompt":"单页漫画导演说明","image_prompt":"单页漫画图片提示词","roles":["林晚"],"scene_presets":[]}`,
	}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunComicBootstrap(ctx, "alice", batch.ID, batch.Books[0].ID); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("comic bootstrap calls=%d, want separate segmentation and prompt calls", len(provider.calls))
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(latest.Books[0].AssetRecords) != 1 || latest.Books[0].AssetRecords[0].Name != "林晚" {
		t.Fatalf("comic prompt stage changed existing assets: %+v", latest.Books[0].AssetRecords)
	}
	if len(latest.Books[0].Videos) != 1 || latest.Books[0].Videos[0].ComicPrompt != "单页漫画导演说明" {
		t.Fatalf("comic page prompt was not saved after segmentation: %+v", latest.Books[0].Videos)
	}
}

func TestParseComicBootstrapOutputRejectsIncompleteSourceCoverage(t *testing.T) {
	source := "第一页原文。第二页原文。"
	boundaries := EnumerateComicSourceBoundaries(source)
	raw := json.RawMessage(`{"characters":[],"scenes":[],"props":[],"storyboard":[{"end_boundary_id":"b1","screen_prompt":"第一页","image_prompt":"第一页图片"}],"source_coverage":{"source_complete":false,"has_remaining_source":true}}`)

	_, _, _, err := parseComicBootstrapOutput(raw, DirectorSnapshot{}, source, boundaries)
	if err == nil || !strings.Contains(err.Error(), "没有完整覆盖") {
		t.Fatalf("parse error=%v, want incomplete source rejection", err)
	}
}

func TestParseComicBootstrapOutputRepairsUnsafeBoundariesWhenPageCountMatches(t *testing.T) {
	sentence := strings.Repeat("甲", 49) + "。"
	source := strings.Repeat(sentence, 5)
	boundaries := EnumerateComicSourceBoundaries(source)
	raw := json.RawMessage(`{
		"characters":[],"scenes":[],"props":[],
		"storyboard":[
			{"end_boundary_id":"b4","screen_prompt":"第一页","image_prompt":"第一页图片"},
			{"end_boundary_id":"b5","screen_prompt":"第二页","image_prompt":"第二页图片"}
		],
		"source_coverage":{"source_complete":true,"has_remaining_source":false}
	}`)

	_, pages, segments, err := parseComicBootstrapOutput(raw, DirectorSnapshot{}, source, boundaries)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || len(segments) != 2 {
		t.Fatalf("pages=%d segments=%d", len(pages), len(segments))
	}
	if segments[0].CharacterCount != 100 || segments[1].CharacterCount != 150 {
		t.Fatalf("unsafe model boundaries were not repaired: %+v", segments)
	}
	if segments[0].SourceText+segments[1].SourceText != source {
		t.Fatal("repaired segments do not reconstruct the source")
	}
}

func TestParseComicBootstrapOutputRepairsUnsafeBoundariesToRequestedPageCount(t *testing.T) {
	sentence := strings.Repeat("乙", 49) + "。"
	source := strings.Repeat(sentence, 58)
	boundaries := EnumerateComicSourceBoundaries(source)
	storyboard := make([]map[string]any, 30)
	for index := range storyboard {
		boundaryIndex := 4 + index*2
		if boundaryIndex > len(boundaries) || index == len(storyboard)-1 {
			boundaryIndex = len(boundaries)
		}
		storyboard[index] = map[string]any{
			"end_boundary_id": fmt.Sprintf("b%d", boundaryIndex),
			"screen_prompt":   fmt.Sprintf("第%d页", index+1),
			"image_prompt":    fmt.Sprintf("第%d页图片", index+1),
		}
	}
	raw, err := json.Marshal(map[string]any{
		"characters": []any{}, "scenes": []any{}, "props": []any{},
		"storyboard":      storyboard,
		"source_coverage": map[string]any{"source_complete": true, "has_remaining_source": false},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, pages, segments, err := parseComicBootstrapOutput(raw, DirectorSnapshot{}, source, boundaries)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 30 || len(segments) != 30 {
		t.Fatalf("pages=%d segments=%d", len(pages), len(segments))
	}
	var rebuilt strings.Builder
	for _, segment := range segments {
		if segment.CharacterCount > 150 {
			t.Fatalf("segment exceeds 150 characters: %+v", segment)
		}
		rebuilt.WriteString(segment.SourceText)
	}
	if rebuilt.String() != source {
		t.Fatal("repaired segments do not reconstruct the source")
	}
}

func TestParseComicBootstrapOutputKeepsIndivisibleOverlongSentenceDuringRepair(t *testing.T) {
	longSentence := strings.Repeat("丙", 160) + "。"
	shortSentence := strings.Repeat("丁", 49) + "。"
	source := longSentence + strings.Repeat(shortSentence, 4)
	boundaries := EnumerateComicSourceBoundaries(source)
	storyboard := []map[string]any{
		{"end_boundary_id": "b2", "screen_prompt": "第一页", "image_prompt": "第一页图片"},
		{"end_boundary_id": "b3", "screen_prompt": "第二页", "image_prompt": "第二页图片"},
		{"end_boundary_id": "b4", "screen_prompt": "第三页", "image_prompt": "第三页图片"},
		{"end_boundary_id": "b5", "screen_prompt": "第四页", "image_prompt": "第四页图片"},
	}
	raw, err := json.Marshal(map[string]any{
		"characters": []any{}, "scenes": []any{}, "props": []any{}, "storyboard": storyboard,
		"source_coverage": map[string]any{"source_complete": true, "has_remaining_source": false},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _, segments, err := parseComicBootstrapOutput(raw, DirectorSnapshot{}, source, boundaries)
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 4 || segments[0].CharacterCount != 161 || segments[0].Reason != ComicSegmentReasonOverlongUnit {
		t.Fatalf("indivisible sentence was not preserved: %+v", segments)
	}
}

func TestParseComicBootstrapOutputReportsPageCountRepairFailure(t *testing.T) {
	sentence := strings.Repeat("戊", 49) + "。"
	source := strings.Repeat(sentence, 4)
	boundaries := EnumerateComicSourceBoundaries(source)
	storyboard := []map[string]any{
		{"end_boundary_id": "b4", "screen_prompt": "整页", "image_prompt": "整页图片"},
	}
	rawOutput, marshalErr := json.Marshal(map[string]any{
		"characters": []any{}, "scenes": []any{}, "props": []any{}, "storyboard": storyboard,
		"source_coverage": map[string]any{"source_complete": true, "has_remaining_source": false},
	})
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}

	_, _, _, err := parseComicBootstrapOutput(rawOutput, DirectorSnapshot{MaxVideoDuration: 10}, source, boundaries)
	if err == nil || !strings.Contains(err.Error(), "模型返回 1 页") || !strings.Contains(err.Error(), "无法按安全边界生成 1 页") {
		t.Fatalf("err=%v, want page-count repair detail", err)
	}
}

func TestBuildComicPromptContractUsesComicAssetsAndConstraints(t *testing.T) {
	book := Book{
		ID: "book-1", Title: "风入琼楼", SourceText: "女主抱着孩子，太子提着礼盒走进病房。",
		AssetRecords: []BookAsset{
			{Kind: "character", Name: "女主", Prompt: "Q版古装少女", Source: "comic"},
			{Kind: "scene", Name: "病房", Prompt: "明亮温暖的现代病房", Source: "comic"},
			{Kind: "character", Name: "太子", Prompt: "电影古装太子", Source: "director"},
		},
	}
	video := Video{ID: "video-1", Label: "VIDEO 01", VideoPrompt: "10秒镜头缓慢推进"}
	settings, _ := json.Marshal(map[string]any{
		"comic": map[string]any{
			"enabled": true, "scope": "all", "body": "直接生成一张完整3:4竖版条漫页。",
			"constraints": []any{
				map[string]any{"constraintCategory": "comic_quality", "body": "Q版日系条漫，清晰线稿"},
				map[string]any{"constraintCategory": "comic_restriction", "body": "必须有分格、旁白框和气泡"},
				map[string]any{"constraintCategory": "comic_negative", "body": "不要水印，不要错误文字"},
			},
		},
	})

	contract, err := BuildComicPromptContract(book, []Video{video}, 0, DirectorSnapshot{Effective: SettingsPatch{"productionMode": raw("comic"), "imageAspectRatio": raw("9:16"), "aiPromptConfig": settings}, AspectRatio: "9:16"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "完整9:16竖幅构图条漫页") || !strings.Contains(contract.SystemPrompt, "阅读顺序") || !strings.Contains(contract.SystemPrompt, "Q版日系条漫") || !strings.Contains(contract.SystemPrompt, "必须有分格") || !strings.Contains(contract.SystemPrompt, "不要水印") {
		t.Fatalf("system prompt misses comic rules: %s", contract.SystemPrompt)
	}
	if strings.Contains(contract.SystemPrompt, "二点五至三点五头身") || strings.Contains(contract.SystemPrompt, "圆脸大眼") {
		t.Fatalf("system prompt still forces the legacy Q style: %s", contract.SystemPrompt)
	}
	if strings.Contains(contract.SystemPrompt, "3:4") || strings.Contains(contract.UserPrompt, "3:4") {
		t.Fatalf("comic contract retained the obsolete 3:4 ratio: system=%s user=%s", contract.SystemPrompt, contract.UserPrompt)
	}
	if !strings.Contains(contract.UserPrompt, `"aspect_ratio":"9:16"`) {
		t.Fatalf("comic payload does not use image aspect ratio: %s", contract.UserPrompt)
	}
	if strings.Contains(contract.SystemPrompt, "10秒镜头缓慢推进") {
		t.Fatalf("comic prompt must not inherit video motion: %s", contract.SystemPrompt)
	}
	if !strings.Contains(contract.UserPrompt, "Q版古装少女") || strings.Contains(contract.UserPrompt, "电影古装太子") {
		t.Fatalf("comic payload must use only comic assets: %s", contract.UserPrompt)
	}
}

func TestBuildComicPromptContractCarriesStableComicAssetIdentity(t *testing.T) {
	book := Book{
		ID: "book-identity", Title: "身份连续性", SourceText: "沈晚看向陆夫人。",
		AssetRecords: []BookAsset{
			{ID: "asset-shen-wan", Kind: "character", Name: "我", FormalName: "沈晚", Facts: "侯府妾室", Appearance: "乌发红衣", Relationships: "与陆夫人为妾室与主母关系", Prompt: "乌发红衣的古装女子", Source: "comic"},
			{ID: "asset-lu-lady", Kind: "character", Name: "夫人", FormalName: "陆夫人", Facts: "当家主母", Appearance: "高髻华服", Relationships: "与沈晚为主母与妾室关系", Prompt: "高髻华服的古装妇人", Source: "comic"},
		},
	}
	settings := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "生成完整漫画页"}})
	contract, err := BuildComicPromptContract(book, []Video{{ID: "page-identity", Label: "漫画页 01", VideoPrompt: "沈晚看向陆夫人。"}}, 0, DirectorSnapshot{Effective: SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": settings,
	}})
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		ComicAssets []map[string]string `json:"comic_assets"`
	}
	if err := json.Unmarshal([]byte(contract.UserPrompt), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.ComicAssets) != 2 {
		t.Fatalf("comic asset identity table length=%d, want 2: %s", len(payload.ComicAssets), contract.UserPrompt)
	}
	for _, want := range []struct{ key, value string }{
		{"asset_id", "asset-shen-wan"}, {"formal_name", "沈晚"}, {"facts", "侯府妾室"}, {"relationships", "与陆夫人为妾室与主母关系"},
	} {
		if !strings.Contains(contract.UserPrompt, want.value) || !strings.Contains(contract.UserPrompt, want.key) {
			t.Fatalf("comic prompt lost stable asset identity %s=%s: %s", want.key, want.value, contract.UserPrompt)
		}
	}
	if !strings.Contains(contract.SystemPrompt, "不得只写“我”“他”“她”") || !strings.Contains(contract.SystemPrompt, "asset_id") {
		t.Fatalf("comic prompt does not enforce stable identity usage: %s", contract.SystemPrompt)
	}
}

func TestBuildComicPromptContractCarriesPageContextAndImageDerivation(t *testing.T) {
	book := Book{
		ID: "book-1", Title: "风入琼楼", SourceText: "第一幕。\n第二幕。\n第三幕。",
		AssetRecords: []BookAsset{{Kind: "character", Name: "林晚", Prompt: "黑发少女", Source: "comic"}},
	}
	videos := []Video{
		{ID: "page-1", Label: "漫画页 01", VideoPrompt: "林晚推开病房门。"},
		{ID: "page-2", Label: "漫画页 02", VideoPrompt: "林晚看见桌上的诊断书。"},
		{ID: "page-3", Label: "漫画页 03", VideoPrompt: "林晚攥紧诊断书。"},
	}
	settings := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "生成完整漫画页",
		"imageDerivation": map[string]any{
			"body": "DERIVE IMAGE PROMPT",
		},
	}})

	contract, err := BuildComicPromptContract(book, videos, 1, DirectorSnapshot{Effective: SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": settings,
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"DERIVE IMAGE PROMPT", "林晚推开病房门。", "林晚看见桌上的诊断书。", "林晚攥紧诊断书。"} {
		if !strings.Contains(contract.SystemPrompt+contract.UserPrompt, want) {
			t.Fatalf("comic page contract missing %q: system=%s user=%s", want, contract.SystemPrompt, contract.UserPrompt)
		}
	}
	if strings.Contains(contract.SystemPrompt+contract.UserPrompt, "第一幕。") || strings.Contains(contract.SystemPrompt+contract.UserPrompt, "第三幕。") {
		t.Fatalf("comic page contract leaked the full novel source text: system=%s user=%s", contract.SystemPrompt, contract.UserPrompt)
	}
}

func TestBuildComicBookPromptContractUsesBoundSourceSegmentsAsPageContent(t *testing.T) {
	settings := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "漫画规则"}})
	book := Book{ID: "book-1", Title: "连续分页", SourceText: "第一段原文。第二段原文。", AssetRecords: []BookAsset{{Kind: "character", Name: "林晚", Prompt: "黑发女性", Source: "comic"}}}
	videos := []Video{
		{ID: "page-1", Label: "漫画页 01", VideoPrompt: "旧视频分镜内容", ComicSourceSegment: &ComicSourceSegment{SourceText: "第一段原文。"}},
		{ID: "page-2", Label: "漫画页 02", ComicSourceSegment: &ComicSourceSegment{SourceText: "第二段原文。"}},
	}
	contract, err := BuildComicBookPromptContract(book, videos, DirectorSnapshot{Effective: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": settings}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.UserPrompt, "第一段原文。") || !strings.Contains(contract.UserPrompt, "第二段原文。") || strings.Contains(contract.UserPrompt, "旧视频分镜内容") {
		t.Fatalf("comic page contract did not use bound source segments: %s", contract.UserPrompt)
	}
}

func TestBuildComicPromptContractLimitsTextDensityPerComicPage(t *testing.T) {
	book := Book{ID: "book-1", Title: "测试", SourceText: "当前页原文", AssetRecords: []BookAsset{{Kind: "character", Name: "主角", Prompt: "漫画角色", Source: "comic"}}}
	videos := []Video{{ID: "video-1", Label: "漫画页 01", VideoPrompt: "当前页原文"}}
	snapshot := DirectorSnapshot{Effective: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, map[string]any{
		"comic": map[string]any{
			"enabled": true,
			"style":   map[string]any{"id": "style-1", "name": "测试画风", "body": "测试画风"},
			"body":    "测试漫画规则",
		},
	})}}
	contract, err := BuildComicPromptContract(book, videos, 0, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"每页画面文字总量控制在45-70字", "旁白拆成多个8-24字的短框", "整页最多3个旁白框"} {
		if !strings.Contains(contract.SystemPrompt, expected) {
			t.Fatalf("comic prompt contract missing text-density rule %q: %s", expected, contract.SystemPrompt)
		}
	}
}

func TestComicAssetExtractionUsesFullCapturedTextInsteadOfContentRangeLines(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{
		Title:          "K",
		SourceText:     "第一行\n第二行\n第三行\n第四行\n第五行\n第六行关键人物登场",
		SourceMetadata: map[string]any{"contentRangeLines": 5},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := map[string]any{"comic": map[string]any{
		"enabled": true,
		"assets":  map[string]any{"presetId": "batch-comic-assets", "body": "提取漫画资产"},
	}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{`{"characters":[{"name":"关键人物","prompt":"黑发青年"}],"scenes":[],"props":[]}`}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunComicAssetExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 || !strings.Contains(provider.calls[0].UserPrompt, "第六行关键人物登场") {
		t.Fatalf("comic asset extraction did not receive the complete captured text: calls=%+v", provider.calls)
	}
}

func TestBuildComicPromptContractLetsSelectedStyleReplaceLegacyQWording(t *testing.T) {
	book := Book{ID: "book-1", Title: "风入琼楼", SourceText: "女主推门。", AssetRecords: []BookAsset{{Kind: "character", Name: "女主", Prompt: "年轻女子", Source: "comic"}}}
	settings := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "固定为可爱大头短身、圆脸大眼、低饱和柔和赛璐璐上色；人物按二点五至三点五头身的可爱短身比例绘制，当前页分三格。",
		"style":   map[string]any{"presetId": "batch-comic-style-14", "body": "水墨晕染，毛笔笔触"},
	}})
	contract, err := BuildComicPromptContract(book, []Video{{ID: "video-1", Label: "第一页"}}, 0, DirectorSnapshot{Effective: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": settings}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "水墨晕染") || strings.Contains(contract.SystemPrompt, "可爱大头短身") || strings.Contains(contract.SystemPrompt, "二点五至三点五头身") {
		t.Fatalf("selected style did not replace legacy Q wording: %s", contract.SystemPrompt)
	}
}

func TestBuildComicPromptContractExpandsRecoveredCMPlaceholders(t *testing.T) {
	book := Book{ID: "book-1", Title: "风入琼楼", SourceText: "女主推门。", AssetRecords: []BookAsset{{Kind: "character", Name: "林晚", FormalName: "林晚（现代）", Prompt: "黑发少女", Source: "comic"}}}
	settings := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "梗概：{summary}\n预设：{roles_str}\n窗口：{context_window}\n示例：{default_img_example}\n{{\"pages\": []}}",
	}})
	contract, err := BuildComicBookPromptContract(book, []Video{{ID: "page-1", Label: "漫画页 01", VideoPrompt: "林晚推门。"}}, DirectorSnapshot{Effective: SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": settings,
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, unresolved := range []string{"{summary}", "{roles_str}", "{context_window}", "{default_img_example}", "{{"} {
		if strings.Contains(contract.SystemPrompt, unresolved) {
			t.Fatalf("recovered CM placeholder %q was not expanded: %s", unresolved, contract.SystemPrompt)
		}
	}
	for _, want := range []string{"林晚推门", "林晚（现代）", `{"pages": []}`} {
		if !strings.Contains(contract.SystemPrompt, want) {
			t.Fatalf("expanded CM prompt missing %q: %s", want, contract.SystemPrompt)
		}
	}
}

func TestBuildComicPromptContractExposesCurrentPageForIllustratedTextPreset(t *testing.T) {
	book := Book{ID: "book-1", Title: "图文卡", SourceText: "上一页原文。\n当前页正文。\n下一页原文。", AssetRecords: []BookAsset{{Kind: "character", Name: "林晚", Prompt: "黑发少女", Source: "comic"}}}
	videos := []Video{
		{ID: "page-1", Label: "漫画页 01", VideoPrompt: "上一页原文。"},
		{ID: "page-2", Label: "漫画页 02", VideoPrompt: "当前页正文。"},
		{ID: "page-3", Label: "漫画页 03", VideoPrompt: "下一页原文。"},
	}
	settings := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "单页图文卡：按 current_page 原样排版，句末补「（未完见下页）」。",
	}})
	contract, err := BuildComicPromptContract(book, videos, 1, DirectorSnapshot{Effective: SettingsPatch{
		"productionMode": rawSetting(t, "comic"),
		"aiPromptConfig": settings,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.UserPrompt, `"current_page":"当前页正文。"`) {
		t.Fatalf("user payload 未携带 current_page 当前页正文: %s", contract.UserPrompt)
	}
	if !strings.Contains(contract.SystemPrompt, "（未完见下页）") {
		t.Fatalf("system prompt 未保留图文卡约束: %s", contract.SystemPrompt)
	}
}

func TestRunComicPromptExtractionPersistsComicPromptWithoutChangingVideoPrompt(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{"comic": map[string]any{"enabled": true, "scope": "all", "body": "完整竖版条漫", "constraints": []any{
		map[string]any{"constraintCategory": "comic_quality", "body": "Q版"},
		map[string]any{"constraintCategory": "comic_restriction", "body": "分格"},
		map[string]any{"constraintCategory": "comic_negative", "body": "无水印"},
	}}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "Q版少女", Source: "comic"}}
	store.mu.Unlock()
	service := &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{"完整漫画页提示词"}}}
	if _, err := service.RunComicPromptExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	latest, _ := store.GetBatch(context.Background(), "alice", batch.ID)
	if latest.Books[0].Videos[0].ComicPrompt != "完整漫画页提示词" || latest.Books[0].Videos[0].VideoPrompt != book.Videos[0].VideoPrompt {
		t.Fatalf("comic=%q video=%q", latest.Books[0].Videos[0].ComicPrompt, latest.Books[0].Videos[0].VideoPrompt)
	}
}

func TestRunComicPromptExtractionPersistsSeparateScreenAndImagePrompts(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{"comic": map[string]any{
		"enabled":         true,
		"body":            "完整竖版条漫",
		"imageDerivation": map[string]any{"body": "整理图片提示词"},
	}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	provider := &queuedDirectorProvider{values: []string{`{"screen_prompt":"三格漫画导演说明","image_prompt":"可直接出图的三格漫画提示词","roles":["女主"],"scene_presets":[]}`}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunComicPromptExtraction(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	latest, _ := store.GetBatch(context.Background(), "alice", batch.ID)
	video := latest.Books[0].Videos[0]
	if video.ComicPrompt != "三格漫画导演说明" || video.ComicImagePrompt != "可直接出图的三格漫画提示词" || len(video.ComicRoles) != 1 || video.ComicRoles[0] != "女主" {
		t.Fatalf("comic prompts were not persisted separately: %+v", video)
	}
}

func TestRunComicPromptExtractionRetriesOneTruncatedPageWithCompactOutput(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{"comic": map[string]any{
		"enabled":         true,
		"body":            "完整竖版条漫",
		"imageDerivation": map[string]any{"body": "整理图片提示词"},
	}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	provider := &truncatingDirectorProvider{values: []string{
		"truncated",
		`{"screen_prompt":"紧凑的三格导演说明","image_prompt":"紧凑的三格图片提示词","roles":["女主"],"scene_presets":[]}`,
	}}

	updated, err := (&DirectorService{Store: store, Provider: provider}).RunComicPromptExtraction(context.Background(), "alice", batch.ID, book.ID)

	if err != nil {
		t.Fatalf("one truncated comic page should retry with compact output: %v", err)
	}
	if len(updated) != 1 || updated[0].ComicImagePrompt != "紧凑的三格图片提示词" {
		t.Fatalf("compact retry did not persist the completed page: %+v", updated)
	}
	if len(provider.calls) != 2 || !strings.Contains(provider.calls[1].SystemPrompt, "更紧凑") || !provider.calls[1].DisableJSONResponse {
		t.Fatalf("retry must ask for compact JSON output: %#v", provider.calls)
	}
}

func TestRunComicPromptExtractionCallsTextProviderOncePerPage(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "女主住院后独自离开。", Videos: []CreateVideoInput{{Label: "漫画页 01", VideoPrompt: "女主坐在病床边。"}, {Label: "漫画页 02", VideoPrompt: "女主走出病房。"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "整书漫画页提示词",
	}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	provider := &queuedDirectorProvider{values: []string{
		`{"screen_prompt":"第一页导演说明","image_prompt":"第一页图片词","roles":["女主"],"role_stage":{"女主":"女主"},"scene_presets":[]}`,
		`{"screen_prompt":"第二页导演说明","image_prompt":"第二页图片词","roles":["女主"],"role_stage":{"女主":"女主"},"scene_presets":[]}`,
	}}

	updated, err := (&DirectorService{Store: store, Provider: provider}).RunComicPromptExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != len(book.Videos) {
		t.Fatalf("comic prompt extraction calls = %d, want %d", len(provider.calls), len(book.Videos))
	}
	if len(updated) != 2 || updated[0].ComicPrompt != "第一页导演说明" || updated[1].ComicPrompt != "第二页导演说明" {
		t.Fatalf("whole-book comic prompts were not applied by video_id: %+v", updated)
	}
	if updated[0].ComicRoleStage["女主"] != "女主" {
		t.Fatalf("role_stage was not preserved: %+v", updated[0].ComicRoleStage)
	}
}

func TestRunMissingComicPromptExtractionPreservesCompletedPages(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "女主先坐在病床边，再走出病房。", Videos: []CreateVideoInput{{Label: "漫画页 01", VideoPrompt: "女主坐在病床边。"}, {Label: "漫画页 02", VideoPrompt: "女主走出病房。"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "整书漫画页提示词"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": config}, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	first := book.Videos[0]
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: first.ID}, SettingsUpdate{Patch: SettingsPatch{"comicPrompt": rawSetting(t, "用户已编辑第一页"), "comicImagePrompt": rawSetting(t, "用户已编辑第一页图片词")}, ExpectedRevision: first.Revision}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	secondID := book.Videos[1].ID
	provider := &queuedDirectorProvider{values: []string{`{"pages":[{"video_id":"` + secondID + `","role_stage":{"女主":"女主"},"scene_presets":[],"screen_prompt":"补跑第二页","image_prompt":"补跑第二页图片词","roles":["女主"]}]}`}}

	updated, err := (&DirectorService{Store: store, Provider: provider}).RunMissingComicPromptExtraction(ctx, "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 || len(updated) != 1 || updated[0].ID != secondID {
		t.Fatalf("missing-page retry result=%+v calls=%d", updated, len(provider.calls))
	}
	latest, _ := store.GetBatch(ctx, "alice", batch.ID)
	if latest.Books[0].Videos[0].ComicPrompt != "用户已编辑第一页" || latest.Books[0].Videos[1].ComicPrompt != "补跑第二页" {
		t.Fatalf("missing-page retry overwrote completed work: %+v", latest.Books[0].Videos)
	}
}

func TestRunComicPromptExtractionAllowsLightPresetEmptyImagePrompt(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{"comic": map[string]any{
		"enabled":  true,
		"presetId": "batch-comic-page-light",
		"body":     "轻量分镜版允许图片词为空",
	}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	provider := &queuedDirectorProvider{values: []string{`{"pages":[{"video_id":"` + book.Videos[0].ID + `","role_stage":{},"scene_presets":[],"screen_prompt":"轻量版导演说明","image_prompt":"","roles":["女主"]}]}`}}

	updated, err := (&DirectorService{Store: store, Provider: provider}).RunComicPromptExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated[0].ComicPrompt != "轻量版导演说明" || updated[0].ComicImagePrompt != "" {
		t.Fatalf("light preset result changed unexpectedly: %+v", updated[0])
	}
}

func TestComicPromptTextForAspectSwitchesOrientationWording(t *testing.T) {
	portrait := comicPromptTextForAspect("一张3:4竖版条漫页", "2:3")
	if !strings.Contains(portrait, "2:3竖幅构图") || strings.Contains(portrait, "竖版") {
		t.Fatalf("portrait ratio kept stale wording: %s", portrait)
	}
	landscape := comicPromptTextForAspect("一张3:4竖版条漫页", "4:3")
	if !strings.Contains(landscape, "4:3横幅构图") || strings.Contains(landscape, "竖版") || strings.Contains(landscape, "条漫") {
		t.Fatalf("landscape ratio kept portrait wording: %s", landscape)
	}
	square := comicPromptTextForAspect("一张3:4竖版条漫页", "1:1")
	if !strings.Contains(square, "1:1方形构图") || strings.Contains(square, "竖版") {
		t.Fatalf("square ratio kept portrait wording: %s", square)
	}
	if got := comicPromptTextForAspect("成图为{{image_aspect_ratio}}{{image_orientation}}的中文漫画页", "3:2"); !strings.Contains(got, "3:2横幅构图") {
		t.Fatalf("structure lock did not use landscape orientation: %s", got)
	}
}
