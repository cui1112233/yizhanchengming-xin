package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLatestFailedBookStageRunStaysWithinBook(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "first"}, {Title: "second"}}})
	if err != nil {
		t.Fatal(err)
	}
	first, second := batch.Books[0], batch.Books[1]
	_, err = store.CreateBookStageRun(context.Background(), BookStageRun{Owner: "alice", BatchID: batch.ID, BookID: first.ID, Stage: BookStageImage, Status: ProductionFailed, Attempt: 1, ErrorMessage: "image failed", UpdatedAt: time.Unix(10, 0)})
	if err != nil {
		t.Fatal(err)
	}
	secondFailure, err := store.CreateBookStageRun(context.Background(), BookStageRun{Owner: "alice", BatchID: batch.ID, BookID: second.ID, Stage: BookStageVideo, Status: ProductionFailed, Attempt: 1, ErrorMessage: "video failed", UpdatedAt: time.Unix(20, 0)})
	if err != nil {
		t.Fatal(err)
	}
	all, err := store.ListBookStageRuns(context.Background(), "alice", batch.ID, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].BookID != first.ID || all[0].ID == secondFailure.ID {
		t.Fatalf("runs=%+v second=%+v", all, secondFailure)
	}
	latest := LatestFailedBookStageRun(all)
	if latest == nil || latest.Stage != BookStageImage {
		t.Fatalf("latest=%+v", latest)
	}
}

func TestBookStageServiceRetriesOnlyLatestFailedStageForBook(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: "source"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	_, err = store.CreateBookStageRun(context.Background(), BookStageRun{Owner: "alice", BatchID: batch.ID, BookID: book.ID, Stage: BookStageDirector, Status: ProductionFailed, Attempt: 1, ErrorMessage: "director failed"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreateBookStageRun(context.Background(), BookStageRun{Owner: "alice", BatchID: batch.ID, BookID: book.ID, Stage: BookStageImage, Status: ProductionFailed, Attempt: 1, ErrorMessage: "image failed"})
	if err != nil {
		t.Fatal(err)
	}
	service := &BookStageService{Store: store}
	summary, err := service.RetryLastFailed(context.Background(), "alice", batch.ID, book.ID, "retry-1", "")
	if err == nil {
		t.Fatal("expected unavailable image retry error")
	}
	if summary.LastFailed == nil || summary.LastFailed.Stage != BookStageImage {
		t.Fatalf("latest failed stage = %#v", summary.LastFailed)
	}
	if len(summary.Runs) != 3 || summary.Runs[2].Stage != BookStageImage || summary.Runs[2].Attempt != 2 {
		t.Fatalf("runs = %#v", summary.Runs)
	}
}

func TestBookStageImageUsesConfiguredGenerator(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	called := false
	service := &BookStageService{
		Store:      store,
		ImageReady: func(context.Context, string) (bool, string) { return true, "" },
		ImageGenerator: func(_ context.Context, owner, batchID, bookID string) error {
			called = owner == "alice" && batchID == batch.ID && bookID == book.ID
			return nil
		},
	}
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageImage, StageModeMissing, "image-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if !called || len(summary.Runs) != 1 || summary.Runs[0].Status != ProductionSucceeded {
		t.Fatalf("image run did not use configured generator: called=%v runs=%+v", called, summary.Runs)
	}
}

func TestComicBootstrapModeIsAvailableOnlyForTheComicStage(t *testing.T) {
	if !validStageMode(BookStageComic, StageModeBootstrap) {
		t.Fatal("comic bootstrap mode must be accepted for the comic stage")
	}
	if validStageMode(BookStageAssets, StageModeBootstrap) {
		t.Fatal("comic bootstrap mode must not be accepted for the asset stage")
	}
}

func TestLatestFailedBookStageRunIgnoresFailureResolvedByLaterAttempt(t *testing.T) {
	firstFailure := BookStageRun{ID: "assets-1", Stage: BookStageAssets, Status: ProductionFailed, Attempt: 1, UpdatedAt: time.Unix(10, 0)}
	resolvedRetry := BookStageRun{ID: "assets-2", Stage: BookStageAssets, Status: ProductionSucceeded, Attempt: 2, UpdatedAt: time.Unix(20, 0)}
	directorFailure := BookStageRun{ID: "director-1", Stage: BookStageDirector, Status: ProductionFailed, Attempt: 1, UpdatedAt: time.Unix(15, 0)}

	latest := LatestFailedBookStageRun([]BookStageRun{firstFailure, directorFailure, resolvedRetry})
	if latest == nil || latest.ID != directorFailure.ID {
		t.Fatalf("latest unresolved failure = %#v", latest)
	}
}

func TestLatestFailedBookStageRunIgnoresDuplicateDirectorRevisionConflict(t *testing.T) {
	duplicate := BookStageRun{
		ID:           "director-duplicate",
		Stage:        BookStageDirector,
		Status:       ProductionFailed,
		ErrorMessage: "revision conflict: 导演分镜已存在；请使用重新生成导演分镜",
		UpdatedAt:    time.Unix(30, 0),
	}
	if latest := LatestFailedBookStageRun([]BookStageRun{duplicate}); latest != nil {
		t.Fatalf("duplicate director revision conflict must not remain retryable: %#v", latest)
	}
}

func TestNormalizeStageExecutionErrorMarksTextProviderTransportFailureRetryable(t *testing.T) {
	err := normalizeStageExecutionError(BookStageDirector, errors.New("context deadline exceeded"))
	if !errors.Is(err, ErrUnavailable) || !strings.Contains(err.Error(), "文本模型请求失败") {
		t.Fatalf("normalized error = %v", err)
	}
}

func TestBookStageAssetMalformedModelOutputIsRetryable(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"assetImagesEnabled": rawSetting(t, true)}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	service := &BookStageService{
		Store:    store,
		Director: &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{"不是 JSON"}}},
	}

	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageAssets, StageModeMissing, "asset-malformed-output", "")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("malformed model output must be retryable, got %v", err)
	}
	if summary.LastFailed == nil || summary.LastFailed.Stage != BookStageAssets || !strings.Contains(summary.LastFailed.ErrorMessage, "导演模型没有返回合法 JSON") {
		t.Fatalf("failed asset run = %#v", summary.LastFailed)
	}
}

func TestBookStageMissingSkipsExistingComicAssets(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "漫画批次", Books: []CreateBookInput{{Title: "漫画书", SourceText: "林晚走进办公室。"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	comicConfig := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"presetId": "batch-comic-assets", "body": "提取漫画资产"},
	}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": comicConfig}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	store.bookAssets["comic-existing"] = memoryOwned[BookAsset]{Owner: "alice", Value: BookAsset{ID: "comic-existing", BatchID: batch.ID, BookID: book.ID, Kind: "character", Name: "林晚", Prompt: "黑发女性", Source: "comic"}}
	store.mu.Unlock()
	provider := &queuedDirectorProvider{}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}}

	if _, err := service.Run(ctx, "alice", batch.ID, book.ID, BookStageAssets, StageModeMissing, "assets-missing", ""); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 0 {
		t.Fatalf("existing comic assets triggered %d AI calls, want 0", len(provider.calls))
	}
}

func TestBookStageMissingReplacesLegacyVideoPaginationWithComicSourceSegments(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "漫画批次", Books: []CreateBookInput{{
		Title: "漫画书", SourceText: "林晚走进办公室。", Videos: []CreateVideoInput{{Label: "漫画页 01", VideoPrompt: "林晚进门。"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	comicConfig := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "完整中文漫画页"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": comicConfig}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}}

	if _, err := service.Run(ctx, "alice", batch.ID, book.ID, BookStageDirector, StageModeMissing, "director-missing", ""); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("legacy video pagination triggered %d segmentation calls, want 1", len(provider.calls))
	}
	segmentation, err := store.GetComicSourceSegmentation(ctx, "alice", batch.ID, book.ID)
	if err != nil || len(segmentation.Segments) != 1 || segmentation.Segments[0].SourceText != "林晚走进办公室。" {
		t.Fatalf("legacy pagination was not replaced by source segmentation: %#v, %v", segmentation, err)
	}
}

// 统一配置保存只写配置，不改内容；自动化重启时也只以 StageModeMissing 跑缺失
// 阶段。Run 在 Director 分支遇到已存在的 DirectorRevision 直接返回当前摘要
// （stage_service.go:283-284，不会再创建 stage run 或调用 provider），因此
// 已完成的分镜与提示词不会被重置为待执行。
func TestBookStageMissingKeepsCompletedDirectorStageAndPrompts(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批次", Books: []CreateBookInput{{Title: "书", SourceText: "她转身离开，没有回头。"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err := store.PersistDirectorRevision(ctx, "alice", book, DirectorSnapshot{}, "digest-1", "", DirectorResult{
		Storyboard: []DirectorVideo{{FinalPrompt: "固定提示词 A"}},
	}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	before := latest.Books[0]
	if before.DirectorRevision == nil || len(before.Videos) == 0 {
		t.Fatal("seed director revision failed")
	}
	beforePrompts := map[string]string{}
	for _, video := range before.Videos {
		beforePrompts[video.ID] = video.VideoPrompt
	}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store}}

	if _, err := service.Run(ctx, "alice", batch.ID, before.ID, BookStageDirector, StageModeMissing, "director-missing", ""); err != nil {
		t.Fatalf("已完成导演阶段没有跳过，err=%v", err)
	}
	store.mu.Lock()
	runCount := len(store.bookStageRuns)
	store.mu.Unlock()
	if runCount != 0 {
		t.Fatalf("已完成导演阶段仍然新建了 stage run：%d", runCount)
	}
	after, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	afterBook := after.Books[0]
	if afterBook.DirectorRevision == nil || afterBook.DirectorRevision.ID != before.DirectorRevision.ID {
		t.Fatal("统一配置保存后已完成导演阶段被重置")
	}
	if len(afterBook.Videos) != len(beforePrompts) {
		t.Fatalf("已生成分镜数量变化：got %d want %d", len(afterBook.Videos), len(beforePrompts))
	}
	for _, video := range afterBook.Videos {
		if video.VideoPrompt != beforePrompts[video.ID] {
			t.Fatalf("已生成提示词被改写：video=%s", video.ID)
		}
	}
}

func TestBookStageMissingSkipsCurrentComicSourceSegmentation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	source := strings.Repeat("这是完整剧情句子，角色继续推动事件。", 12)
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "漫画批次", Books: []CreateBookInput{{Title: "漫画书", SourceText: source}}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "完整中文漫画页"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": comicConfig}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	drafts, _ := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if _, err := store.ReplaceComicSourceSegments(ctx, "alice", batch.ID, batch.Books[0].ID, sourceDigest(source), drafts); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}}
	if _, err := service.Run(ctx, "alice", batch.ID, batch.Books[0].ID, BookStageDirector, StageModeMissing, "director-current", ""); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 0 {
		t.Fatalf("current segmentation triggered %d model calls", len(provider.calls))
	}
}

func TestBookStageMissingSkipsExistingComicPagePrompts(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "漫画批次", Books: []CreateBookInput{{
		Title: "漫画书", SourceText: "林晚走进办公室。", Videos: []CreateVideoInput{{Label: "漫画页 01", VideoPrompt: "林晚进门。"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	comicConfig := rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "完整中文漫画页"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"productionMode": rawSetting(t, "comic"), "aiPromptConfig": comicConfig}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	video := book.Videos[0]
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"comicPrompt": rawSetting(t, "导演说明"), "comicImagePrompt": rawSetting(t, "图片提示词")}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}}

	if _, err := service.Run(ctx, "alice", batch.ID, book.ID, BookStageComic, StageModeMissing, "comic-missing", ""); err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("missing comic source segmentation triggered %d AI calls, want 1", len(provider.calls))
	}
	segmentation, err := store.GetComicSourceSegmentation(ctx, "alice", batch.ID, book.ID)
	if err != nil || len(segmentation.Segments) != 1 || segmentation.Segments[0].SourceText != "林晚走进办公室。" {
		t.Fatalf("existing prompts did not get continuous source segmentation: %#v, %v", segmentation, err)
	}
}

func TestBookStageDirectorUsesStructuredH3KernelWhenH3VideoPresetIsSelected(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "h3_v12_complete_director_trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	store := NewMemoryStore()
	source := "五岁的我刚被认回豪门，爸妈就甩下一百万生活费。\n把我和陆晚晚扔在别墅里大眼瞪小眼。\n三个月后，爸妈提前回国，想给我们一个惊喜。"
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: source}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	normalizedSource := strings.Join(h3NonEmptyVideoSourceLines(source), "\n")
	hash := sourceDigest(normalizedSource)
	var fixture map[string]any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	fixture["video_source_revision"] = "video-source-" + hash[:16] + "-r1"
	fixture["video_source_hash"] = hash
	raw, err = json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	config := map[string]any{
		"originalDirector": map[string]any{"enabled": true, "scope": "all", "presetVersion": 1, "body": "完整 H3 导演结构"},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	seedH3CharacterAssets(t, store, batch.ID, book.ID)
	service := &BookStageService{Store: store, H3Director: true, Director: &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{string(raw)}}}}
	if _, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageDirector, StageModeForce, "h3-director-1", ""); err != nil {
		t.Fatalf("H3 director stage failed: %v", err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	revision := latest.Books[0].DirectorRevision
	if revision == nil || revision.Output.H3Director == nil || len(revision.Output.H3Director.DirectorCards) != 3 {
		t.Fatalf("director stage did not persist structured H3 output: %#v", revision)
	}
	if len(revision.Output.Storyboard) != 0 {
		t.Fatalf("H3 stage must not persist legacy storyboard as the director source: %#v", revision.Output.Storyboard)
	}
}

func TestBookStageDirectorUsesStandardDirectorForSelectedSDVideoPreset(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: "林晚推开客厅门，握紧玻璃杯。"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := map[string]any{
		"video": map[string]any{"enabled": true, "scope": "all", "presetId": "batch-video-sd", "presetVersion": 1, "body": "SD VIDEO RULE"},
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，缓慢推轨，林晚推开门。`
	provider := &queuedDirectorProvider{values: []string{sdResponse}}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}}
	if _, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageDirector, StageModeForce, "sd-director-1", ""); err != nil {
		t.Fatalf("SD director stage failed: %v", err)
	}
	if len(provider.calls) != 1 || !strings.Contains(provider.calls[0].SystemPrompt, "SD VIDEO RULE") {
		t.Fatalf("selected SD video preset was not used as director rule: %#v", provider.calls)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	revision := latest.Books[0].DirectorRevision
	if revision == nil || revision.Output.H3Director != nil || len(revision.Output.Storyboard) != 1 {
		t.Fatalf("SD preset must persist ordinary storyboard output, got %#v", revision)
	}
	if revision.Output.Storyboard[0].FinalPrompt == "" {
		t.Fatal("SD preset must persist FinalPrompt in storyboard output")
	}
}

func TestBookStageVisualRunsWithoutRecreatingDirectorRevision(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	config := map[string]any{"visual": map[string]any{"enabled": true, "scope": "all", "body": "ONLY VISUAL"}}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: SettingsPatch{"aiPromptConfig": rawSetting(t, config)}, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: book.Videos[0].ID}, SettingsUpdate{Patch: SettingsPatch{"videoPrompt": rawSetting(t, "人物站在窗边。")}, ExpectedRevision: book.Videos[0].Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{"画面提示词"}}}}
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageVisual, StageModeForce, "visual-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Runs) != 1 || summary.Runs[0].Stage != BookStageVisual || summary.Runs[0].Status != ProductionSucceeded {
		t.Fatalf("runs=%+v", summary.Runs)
	}
	latest, _ := store.GetBatch(context.Background(), "alice", batch.ID)
	if latest.Books[0].DirectorRevision != nil || latest.Books[0].Videos[0].VisualPrompt != "画面提示词" {
		t.Fatalf("visual stage must only update visual prompts: %+v", latest.Books[0])
	}
}

func TestBookStageVideoReportsImmediateProviderFailure(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: ProductionFailed}}
	service := &BookStageService{Store: store, Production: &ProductionService{
		Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter,
		Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15},
	}}
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageVideo, StageModeForce, "video-provider-failure", video.ID)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected retryable provider error, got %v", err)
	}
	if summary.LastFailed == nil || summary.LastFailed.Stage != BookStageVideo || summary.LastFailed.Status != ProductionFailed {
		t.Fatalf("last failed stage=%+v", summary.LastFailed)
	}
	if !strings.Contains(summary.LastFailed.ErrorMessage, "视频模型") {
		t.Fatalf("missing provider feedback: %+v", summary.LastFailed)
	}
}

func TestRetryLastFailedVideoWithoutExplicitIDTargetsLatestFailedVideo(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	provider := &queuedDirectorProvider{values: []string{twoDirectorVideosJSON(t)}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(context.Background(), "alice", batch.ID)
	book = batch.Books[0]
	adapter := &sequenceProductionAdapter{refs: []ProviderTaskRef{
		{State: ProductionSucceeded, ProviderTaskID: "provider-ok", MediaURL: "https://media.example/ok.mp4"},
		{State: ProductionFailed, ProviderTaskID: "provider-failed", ErrorMessage: "provider rejected"},
		{State: ProductionSucceeded, ProviderTaskID: "provider-retry", MediaURL: "https://media.example/retry.mp4"},
	}}
	production := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	stages := &BookStageService{Store: store, Production: production}
	if _, err := stages.Run(context.Background(), "alice", batch.ID, book.ID, BookStageVideo, StageModeMissing, "first-stage", ""); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("first stage err=%v, want failed provider task", err)
	}
	summary, err := stages.RetryLastFailed(context.Background(), "alice", batch.ID, book.ID, "retry-stage", "")
	if err != nil {
		t.Fatal(err)
	}
	if summary.LastFailed != nil {
		t.Fatalf("retry left unresolved failure: %+v", summary.LastFailed)
	}
	status, err := production.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	latest := status.Jobs[len(status.Jobs)-1]
	if len(latest.Tasks) != 1 || latest.Tasks[0].VideoID != book.Videos[1].ID || adapter.calls != 3 {
		t.Fatalf("retry must submit only failed VIDEO02: latest=%+v calls=%d", latest.Tasks, adapter.calls)
	}
}

func seedOpeningStageBook(t *testing.T, sdResponse string) (*MemoryStore, Batch, Book, *BookStageService) {
	t.Helper()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "batch", Books: []CreateBookInput{{Title: "book", SourceText: "林晚推开客厅门，握紧玻璃杯。", Videos: []CreateVideoInput{{Label: "v1"}, {Label: "v2"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	config := map[string]any{
		"video": map[string]any{"enabled": true, "scope": "all", "presetId": "batch-video-sd", "presetVersion": 1, "body": "SD VIDEO RULE"},
	}
	patch := SettingsPatch{
		"aiPromptConfig": rawSetting(t, config),
		"openingEnabled": rawSetting(t, true),
		"openingCount":   rawSetting(t, 4),
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: patch, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	variantResponse := "===VARIANT 1===\n时长：10秒\n画面故事脚本：雨夜推门而入的开场。\n\n===VARIANT 2===\n时长：10秒\n画面故事脚本：客厅灯光骤暗，她握紧玻璃杯。\n\n===VARIANT 3===\n时长：10秒\n画面故事脚本：门外脚步逼近，她回头看向玄关。"
	provider := &queuedDirectorProvider{values: []string{sdResponse, variantResponse}}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}, OpeningMeta: PresetSnapshot{ID: "batch-opening-meta", Body: "换开头元提示词"}}
	if _, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageDirector, StageModeForce, "director-1", ""); err != nil {
		t.Fatalf("director stage failed: %v", err)
	}
	batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	return store, batch, batch.Books[0], service
}

func openingRunSucceeded(summary BookStageSummary) bool {
	for _, run := range summary.Runs {
		if run.Stage == BookStageOpening {
			return run.Status == ProductionSucceeded
		}
	}
	return false
}

func TestBookStageOpeningSucceedsForTwoStoryboardSDBook(t *testing.T) {
	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，缓慢推轨，林晚推开门。
===VIDEO 02===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 近景，林晚握紧玻璃杯。`
	store, batch, book, service := seedOpeningStageBook(t, sdResponse)
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageOpening, StageModeForce, "opening-1", "")
	if err != nil {
		t.Fatalf("opening stage failed: %v", err)
	}
	if !openingRunSucceeded(summary) {
		t.Fatalf("opening run missing or not succeeded: %+v", summary.Runs)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := latest.Books[0].Videos[0].SettingsState.Patch["openingVariants"]
	if !ok {
		t.Fatalf("two-storyboard book must persist opening variants on VIDEO01: %+v", latest.Books[0].Videos[0].SettingsState.Patch)
	}
	var variants []OpeningVariant
	if err := json.Unmarshal(raw, &variants); err != nil {
		t.Fatal(err)
	}
	if len(variants) == 0 || variants[0].Status != "success" || !strings.Contains(variants[0].Prompt, "雨夜推门而入") {
		t.Fatalf("variants = %+v", variants)
	}
}

func TestBookStageOpeningFailsAndRecordsReasonWhenARequiredVariantIsMissing(t *testing.T) {
	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，缓慢推轨，林晚推开门。
===VIDEO 02===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 近景，林晚握紧玻璃杯。`
	_, batch, book, service := seedOpeningStageBook(t, sdResponse)
	service.Director.Provider = &queuedDirectorProvider{values: []string{
		"===VARIANT 1===\n时长：10秒\n画面故事脚本：雨夜推门而入的开场。",
		"这一轮没有返回任何变体标记，只有说明文字。",
	}}
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageOpening, StageModeForce, "opening-missing", "")
	if err == nil {
		t.Fatal("missing opening variants must fail the stage")
	}
	if summary.LastFailed == nil || summary.LastFailed.Stage != BookStageOpening || !strings.Contains(summary.LastFailed.ErrorMessage, "换开头2") {
		t.Fatalf("last failed stage = %#v", summary.LastFailed)
	}
}

func TestBookStageOpeningSkipsSingleStoryboardBook(t *testing.T) {
	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，缓慢推轨，林晚推开门。`
	store, batch, book, service := seedOpeningStageBook(t, sdResponse)
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageOpening, StageModeForce, "opening-1", "")
	if err != nil {
		t.Fatalf("opening stage failed: %v", err)
	}
	if !openingRunSucceeded(summary) {
		t.Fatalf("opening run missing or not succeeded: %+v", summary.Runs)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, video := range latest.Books[0].Videos {
		if _, ok := video.SettingsState.Patch["openingVariants"]; ok {
			t.Fatalf("single-storyboard book must not persist opening variants: %+v", video.SettingsState.Patch)
		}
	}
}

func TestBookStageUsesSelectedTextModelProvider(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "漫画批次", Books: []CreateBookInput{{Title: "漫画书", SourceText: "林晚走进办公室。"}}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"presetId": "batch-comic-assets", "body": "提取漫画资产"},
	}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"productionMode":     rawSetting(t, "comic"),
			"aiPromptConfig":     comicConfig,
			"textModelId":        rawSetting(t, "model-selected"),
			"assetImagesEnabled": rawSetting(t, true),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	defaultProvider := &queuedDirectorProvider{}
	selectedProvider := &queuedDirectorProvider{values: []string{`{"characters":[{"name":"林晚","prompt":"黑发女性"}],"scenes":[],"props":[]}`}}
	var resolvedOwner, resolvedModel string
	service := &BookStageService{
		Store:    store,
		Director: &DirectorService{Store: store, Provider: defaultProvider},
		TextProviderResolver: func(_ context.Context, owner, modelID string) (DirectorProvider, error) {
			resolvedOwner, resolvedModel = owner, modelID
			if modelID != "model-selected" {
				return nil, ErrConfigRequired
			}
			return selectedProvider, nil
		},
	}

	summary, err := service.Run(ctx, "alice", batch.ID, batch.Books[0].ID, BookStageAssets, StageModeMissing, "assets-selected-model", "")
	if err != nil {
		t.Fatal(err)
	}
	if resolvedOwner != "alice" || resolvedModel != "model-selected" {
		t.Fatalf("resolver got owner=%q model=%q, want alice/model-selected", resolvedOwner, resolvedModel)
	}
	if len(defaultProvider.calls) != 0 {
		t.Fatalf("account default provider was used %d times despite an explicit textModelId", len(defaultProvider.calls))
	}
	if len(selectedProvider.calls) != 1 {
		t.Fatalf("selected provider calls=%d, want 1", len(selectedProvider.calls))
	}
	if len(summary.Runs) != 1 || summary.Runs[0].Status != ProductionSucceeded {
		t.Fatalf("stage run was not recorded as succeeded: %+v", summary.Runs)
	}
}

func TestBookStageFailsExplicitlyWhenSelectedTextModelIsUnavailable(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "漫画批次", Books: []CreateBookInput{{Title: "漫画书", SourceText: "林晚走进办公室。"}}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig := rawSetting(t, map[string]any{"comic": map[string]any{
		"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"presetId": "batch-comic-assets", "body": "提取漫画资产"},
	}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"productionMode":     rawSetting(t, "comic"),
			"aiPromptConfig":     comicConfig,
			"textModelId":        rawSetting(t, "model-missing"),
			"assetImagesEnabled": rawSetting(t, true),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	defaultProvider := &queuedDirectorProvider{}
	service := &BookStageService{
		Store:                store,
		Director:             &DirectorService{Store: store, Provider: defaultProvider},
		TextProviderResolver: func(context.Context, string, string) (DirectorProvider, error) { return nil, nil },
	}

	summary, err := service.Run(ctx, "alice", batch.ID, batch.Books[0].ID, BookStageAssets, StageModeMissing, "assets-missing-model", "")
	if !errors.Is(err, ErrConfigRequired) {
		t.Fatalf("unavailable textModelId must fail as config-required, got %v", err)
	}
	if len(defaultProvider.calls) != 0 {
		t.Fatalf("account default provider silently replaced the missing selection: %d calls", len(defaultProvider.calls))
	}
	if summary.LastFailed == nil || summary.LastFailed.Stage != BookStageAssets || !strings.Contains(summary.LastFailed.ErrorMessage, "选择的文本模型不可用") {
		t.Fatalf("failed run does not explain the missing text model: %#v", summary.LastFailed)
	}
}

func TestBookStageOpeningPersistsDirectorRevisionAsInputRevision(t *testing.T) {
	sdResponse := `===VIDEO 01===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 中景，缓慢推轨，林晚推开门。
===VIDEO 02===
时长：10秒
段内执行约束：无台词。
[场景 1] 总时长：10.000秒
[镜头 1] 近景，林晚握紧玻璃杯。`
	store, batch, book, service := seedOpeningStageBook(t, sdResponse)
	summary, err := service.Run(context.Background(), "alice", batch.ID, book.ID, BookStageOpening, StageModeForce, "opening-1", "")
	if err != nil {
		t.Fatalf("opening stage failed: %v", err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	revision := latest.Books[0].DirectorRevision
	if revision == nil {
		t.Fatal("director revision missing")
	}
	var openingRun *BookStageRun
	for i := range summary.Runs {
		if summary.Runs[i].Stage == BookStageOpening {
			openingRun = &summary.Runs[i]
		}
	}
	if openingRun == nil || openingRun.Status != ProductionSucceeded {
		t.Fatalf("opening run missing or not succeeded: %+v", summary.Runs)
	}
	if openingRun.InputRevision != revision.ID {
		t.Fatalf("opening run input revision = %q, want director revision %q", openingRun.InputRevision, revision.ID)
	}
}
