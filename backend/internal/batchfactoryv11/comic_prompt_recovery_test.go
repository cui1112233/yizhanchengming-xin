package batchfactoryv11

import (
	"context"
	"fmt"
	"testing"
)

// 回补只调用缺失页：已成功的 7 条提示词绝不能被重新请求。
func TestPromptRecoveryOnlyCallsMissing(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	videos := make([]CreateVideoInput, 0, 10)
	for index := 0; index < 10; index++ {
		videos = append(videos, CreateVideoInput{
			Label:       fmt.Sprintf("漫画页 %02d", index+1),
			VideoPrompt: fmt.Sprintf("第 %d 页画面。", index+1),
		})
	}
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{
		Title: "K", SourceText: "女主依次走过十个房间。", Videos: videos,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "整书漫画页提示词"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": config}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	// 前 7 页视为已完成。
	for index := 0; index < 7; index++ {
		latest, err := store.GetBatch(ctx, "alice", batch.ID)
		if err != nil {
			t.Fatal(err)
		}
		video := latest.Books[0].Videos[index]
		if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
			Patch: SettingsPatch{
				"comicPrompt":      rawSetting(t, fmt.Sprintf("已完成导演说明 %d", index+1)),
				"comicImagePrompt": rawSetting(t, fmt.Sprintf("已完成图片词 %d", index+1)),
			},
			ExpectedRevision: video.Revision,
		}); err != nil {
			t.Fatal(err)
		}
	}
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	values := make([]string, 0, 3)
	for index := 7; index < 10; index++ {
		video := latest.Books[0].Videos[index]
		values = append(values, fmt.Sprintf(`{"pages":[{"video_id":%q,"role_stage":{"女主":"女主"},"scene_presets":[],"screen_prompt":"补跑第 %d 页","image_prompt":"补跑第 %d 页图片词","roles":["女主"]}]}`, video.ID, index+1, index+1))
	}
	provider := &queuedDirectorProvider{values: values}

	result, err := (&DirectorService{Store: store, Provider: provider}).RunMissingComicPromptExtraction(ctx, "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatalf("回补不应因已完成页而失败: %v", err)
	}
	if len(provider.calls) != 3 {
		t.Fatalf("provider calls=%d, want 3（已成功的 7 条不得重新请求）", len(provider.calls))
	}
	if len(result.Updated) != 3 {
		t.Fatalf("updated=%d, want 3", len(result.Updated))
	}
	if len(result.Failed) != 0 {
		t.Fatalf("failed=%+v, want none", result.Failed)
	}
	persisted, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 7; index++ {
		if got := persisted.Books[0].Videos[index].ComicPrompt; got != fmt.Sprintf("已完成导演说明 %d", index+1) {
			t.Fatalf("page %d 已完成提示词被覆盖: %q", index+1, got)
		}
	}
	for index := 7; index < 10; index++ {
		if got := persisted.Books[0].Videos[index].ComicPrompt; got != fmt.Sprintf("补跑第 %d 页", index+1) {
			t.Fatalf("page %d 未补齐: %q", index+1, got)
		}
	}
}

func TestPromptExtractionOnlyTargetsRequestedPages(t *testing.T) {
	ctx := WithComicPromptConcurrency(context.Background(), 1)
	store := NewMemoryStore()
	videos := make([]CreateVideoInput, 0, 3)
	for index := 0; index < 3; index++ {
		videos = append(videos, CreateVideoInput{
			Label:       fmt.Sprintf("漫画页 %02d", index+1),
			VideoPrompt: fmt.Sprintf("第 %d 页画面。", index+1),
		})
	}
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{
		Title: "K", SourceText: "女主依次走过三个房间。", Videos: videos,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "整书漫画页提示词"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": config}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.bookAssets["comic-asset"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-asset", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "女主", Prompt: "漫画少女", Source: "comic"}}
	store.mu.Unlock()
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 中间页已有正常提示词，重抽时必须原样保留。
	keep := latest.Books[0].Videos[1]
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: keep.ID}, SettingsUpdate{
		Patch:            SettingsPatch{"comicPrompt": rawSetting(t, "已确认导演说明"), "comicImagePrompt": rawSetting(t, "已确认图片词")},
		ExpectedRevision: keep.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	values := make([]string, 0, 2)
	for _, index := range []int{0, 2} {
		video := latest.Books[0].Videos[index]
		values = append(values, fmt.Sprintf(`{"pages":[{"video_id":%q,"screen_prompt":"重抽第 %d 页","image_prompt":"重抽出图 %d"}]}`, video.ID, index+1, index+1))
	}
	provider := &queuedDirectorProvider{values: values}
	result, err := (&DirectorService{Store: store, Provider: provider}).RunComicPromptExtractionForPages(ctx, "alice", batch.ID, book.ID, []string{latest.Books[0].Videos[0].ID, latest.Books[0].Videos[2].ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("provider calls=%d, want 2（未指定的页不得重抽）", len(provider.calls))
	}
	if len(result.Updated) != 2 || len(result.Failed) != 0 {
		t.Fatalf("updated=%d failed=%d, want 2/0", len(result.Updated), len(result.Failed))
	}
	persisted, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := persisted.Books[0].Videos[1].ComicPrompt; got != "已确认导演说明" {
		t.Fatalf("未指定的页被覆盖: %q", got)
	}
	if got := persisted.Books[0].Videos[0].ComicPrompt; got != "重抽第 1 页" {
		t.Fatalf("指定页未重抽: %q", got)
	}
}
