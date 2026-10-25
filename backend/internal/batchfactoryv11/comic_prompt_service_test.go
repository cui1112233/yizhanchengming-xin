package batchfactoryv11

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

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

	contract, err := BuildComicPromptContract(book, video, DirectorSnapshot{Effective: SettingsPatch{"productionMode": raw("comic"), "imageAspectRatio": raw("9:16"), "aiPromptConfig": settings}, AspectRatio: "9:16"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "完整9:16画幅条漫页") || !strings.Contains(contract.SystemPrompt, "阅读顺序") || !strings.Contains(contract.SystemPrompt, "Q版日系条漫") || !strings.Contains(contract.SystemPrompt, "必须有分格") || !strings.Contains(contract.SystemPrompt, "不要水印") {
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

func TestBuildComicPromptContractLetsSelectedStyleReplaceLegacyQWording(t *testing.T) {
	book := Book{ID: "book-1", Title: "风入琼楼", SourceText: "女主推门。", AssetRecords: []BookAsset{{Kind: "character", Name: "女主", Prompt: "年轻女子", Source: "comic"}}}
	settings := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true,
		"body":    "固定为可爱大头短身、圆脸大眼、低饱和柔和赛璐璐上色；人物按二点五至三点五头身的可爱短身比例绘制，当前页分三格。",
		"style":   map[string]any{"presetId": "batch-comic-style-14", "body": "水墨晕染，毛笔笔触"},
	}})
	contract, err := BuildComicPromptContract(book, Video{ID: "video-1", Label: "第一页"}, DirectorSnapshot{Effective: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": settings}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, "水墨晕染") || strings.Contains(contract.SystemPrompt, "可爱大头短身") || strings.Contains(contract.SystemPrompt, "二点五至三点五头身") {
		t.Fatalf("selected style did not replace legacy Q wording: %s", contract.SystemPrompt)
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
