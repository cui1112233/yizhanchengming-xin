package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

func TestComicGenerationFailureStageSeparatesAssetImageFailures(t *testing.T) {
	stage, detail := comicGenerationFailureStage(errors.New("漫画资产图生成失败: temporary upstream failure"))
	if stage != "comic_asset_images" || detail != "漫画资产参考图生成失败" {
		t.Fatalf("asset image failure = %q / %q", stage, detail)
	}
	stage, detail = comicGenerationFailureStage(errors.New("当前漫画页生成失败"))
	if stage != "comic_images" || detail != "整本漫画出图失败" {
		t.Fatalf("page image failure = %q / %q", stage, detail)
	}
	missingAssetErr := errors.New("漫画资产图未全部写回（剩余 2 个资产）")
	stage, detail = comicGenerationFailureStage(missingAssetErr)
	if stage != "comic_asset_images" || detail != "漫画资产参考图生成失败" {
		t.Fatalf("missing asset image = %q / %q", stage, detail)
	}
	if !comicMissingAssetError(missingAssetErr) || comicMissingAssetError(errors.New("整本漫画出图失败")) {
		t.Fatal("only a real missing-asset error should trigger the backoff retry")
	}
}

func TestAcknowledgeComicRetryClearsQueuedMarker(t *testing.T) {
	entry := map[string]any{"status": "running", "retryRequested": true}

	acknowledgeComicRetry(entry)

	if _, ok := entry["retryRequested"]; ok {
		t.Fatalf("comic retry marker should be cleared after worker pickup: %#v", entry)
	}
}

func TestCompletedAutomationJobWithDueComicRetryCanBeRecovered(t *testing.T) {
	now := time.Date(2026, 11, 1, 10, 36, 0, 0, time.UTC)
	job := map[string]any{
		"state": "completed",
		"books": map[string]any{
			"book-retrying": map[string]any{
				"status":         "running",
				"retryRequested": true,
				"retryReason":    "content_policy",
				"retryAt":        "2026-11-01T10:35:22Z",
			},
			"book-ready": map[string]any{"status": "completed"},
		},
	}
	if !automationHasDueRetry(job, now) {
		t.Fatal("a completed job with a due book retry must be eligible for recovery")
	}

	entry := job["books"].(map[string]any)["book-retrying"].(map[string]any)
	entry["retryAt"] = "2026-11-01T10:40:00Z"
	if automationHasDueRetry(job, now) {
		t.Fatal("a retry that is not due must not reopen a completed job")
	}
}

func TestFailedAutomationJobWithDueComicRetryCanBeRecovered(t *testing.T) {
	now := time.Date(2026, 11, 1, 10, 36, 0, 0, time.UTC)
	job := map[string]any{
		"state": "failed",
		"books": map[string]any{
			"book-retrying": map[string]any{
				"status":         "running",
				"retryRequested": true,
				"retryReason":    "content_policy",
				"retryAt":        "2026-11-01T10:35:22Z",
			},
		},
	}
	if !automationJobRetryDispatchDue(job, now) {
		t.Fatal("a failed job with a due book retry must be resumed automatically")
	}
}

func TestComicAutomationRetriesTransientGenerationUpToThreeTimes(t *testing.T) {
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	coordinator := &BatchAutomationCoordinator{now: func() time.Time { return now }}
	entry := map[string]any{"status": "running", "stage": "comic_images"}
	persisted := 0
	attempts := 0
	collection, err := coordinator.generateComicWithRetry(
		context.Background(),
		entry,
		func() { persisted++ },
		func(context.Context) (batchfactoryv11.ComicCollection, error) {
			attempts++
			if attempts < 3 {
				return batchfactoryv11.ComicCollection{}, errors.New("temporary upstream failure")
			}
			return batchfactoryv11.ComicCollection{ID: "collection-1"}, nil
		},
		func(context.Context, time.Duration) error { return nil },
	)
	if err != nil || collection.ID != "collection-1" {
		t.Fatalf("collection=%#v err=%v, want success after retry", collection, err)
	}
	if attempts != 3 || persisted != 4 {
		t.Fatalf("attempts=%d persisted=%d, want 3 attempts and 4 retry state updates", attempts, persisted)
	}
	if entry["retryCount"] != 2 {
		t.Fatalf("retryCount=%#v, want 2", entry["retryCount"])
	}
	if _, ok := entry["retryAt"]; ok {
		t.Fatalf("retryAt should be cleared after success: %#v", entry)
	}
}

func TestComicAutomationClearsRetryScheduleBeforeRetryAttempt(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{now: time.Now}
	entry := map[string]any{"status": "running", "stage": "comic_images"}
	attempts := 0
	_, err := coordinator.generateComicWithRetry(
		context.Background(),
		entry,
		func() {},
		func(context.Context) (batchfactoryv11.ComicCollection, error) {
			attempts++
			if attempts == 1 {
				return batchfactoryv11.ComicCollection{}, errors.New("temporary upstream failure")
			}
			if _, exists := entry["retryAt"]; exists {
				t.Fatalf("retry attempt must not retain the old retryAt: %#v", entry)
			}
			return batchfactoryv11.ComicCollection{ID: "collection-2"}, nil
		},
		func(context.Context, time.Duration) error { return nil },
	)
	if err != nil || attempts != 2 {
		t.Fatalf("err=%v attempts=%d, want one retry and a successful second attempt", err, attempts)
	}
}

func TestComicAutomationDoesNotRetryConfigurationErrors(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{now: time.Now}
	entry := map[string]any{"status": "running", "stage": "comic_images"}
	attempts := 0
	_, err := coordinator.generateComicWithRetry(
		context.Background(),
		entry,
		func() {},
		func(context.Context) (batchfactoryv11.ComicCollection, error) {
			attempts++
			return batchfactoryv11.ComicCollection{}, fmt.Errorf("%w: 图片提示词缺失", batchfactoryv11.ErrConflict)
		},
		func(context.Context, time.Duration) error {
			t.Fatal("configuration error must not wait for retry")
			return nil
		},
	)
	if err == nil || attempts != 1 {
		t.Fatalf("err=%v attempts=%d, want one non-retryable attempt", err, attempts)
	}
	if !retryableComicAutomationError(fmt.Errorf("%w: 当前漫画页已有生成任务，请等待完成后再重新生成", batchfactoryv11.ErrConflict)) {
		t.Fatal("temporary page-generation conflict should be retried")
	}
}

func TestComicAutomationDoesNotRetryPermanentProviderHTTPFailures(t *testing.T) {
	for _, status := range []int{400, 401, 402, 403, 404, 422} {
		err := fmt.Errorf("temporary upstream failure: image provider returned HTTP %d: balance pre-deduction insufficient", status)
		if retryableComicAutomationError(err) {
			t.Fatalf("HTTP %d provider rejection must stop without retry", status)
		}
	}
	for _, status := range []int{408, 429, 500, 502, 503} {
		err := fmt.Errorf("temporary upstream failure: image provider returned HTTP %d", status)
		if !retryableComicAutomationError(err) {
			t.Fatalf("HTTP %d may be transient and should remain retryable", status)
		}
	}
}

func TestComicAutomationCallsProviderOnceForHTTP403(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{now: time.Now}
	entry := map[string]any{"status": "running", "stage": "comic_images"}
	attempts := 0
	_, err := coordinator.generateComicWithRetry(
		context.Background(),
		entry,
		func() {},
		func(context.Context) (batchfactoryv11.ComicCollection, error) {
			attempts++
			return batchfactoryv11.ComicCollection{}, errors.New("temporary upstream failure: image provider returned HTTP 403: balance pre-deduction insufficient")
		},
		func(context.Context, time.Duration) error {
			t.Fatal("HTTP 403 must not schedule a retry")
			return nil
		},
	)
	if err == nil || attempts != 1 {
		t.Fatalf("err=%v attempts=%d, want one provider call and a terminal error", err, attempts)
	}
	if _, ok := entry["retryAt"]; ok {
		t.Fatalf("terminal provider rejection must not retain a retry schedule: %#v", entry)
	}
}

func TestComicAutomationClearsRetryScheduleAfterFinalFailure(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{now: time.Now}
	entry := map[string]any{"status": "running", "stage": "comic_images"}
	attempts := 0
	_, err := coordinator.generateComicWithRetry(
		context.Background(),
		entry,
		func() {},
		func(context.Context) (batchfactoryv11.ComicCollection, error) {
			attempts++
			return batchfactoryv11.ComicCollection{}, fmt.Errorf("%w: 当前漫画页已有生成任务，请等待完成后再重新生成", batchfactoryv11.ErrConflict)
		},
		func(context.Context, time.Duration) error { return nil },
	)
	if err == nil || attempts != comicAutomaticRetryLimit {
		t.Fatalf("err=%v attempts=%d, want terminal failure after %d attempts", err, attempts, comicAutomaticRetryLimit)
	}
	if _, ok := entry["retryAt"]; ok {
		t.Fatalf("terminal comic failure must not retain a future retry timestamp: %#v", entry)
	}
}

func TestComicAutomationStopsRetryingAfterOverallContextDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	coordinator := &BatchAutomationCoordinator{now: time.Now}
	entry := map[string]any{"status": "running", "stage": "comic_images"}
	attempts := 0
	_, err := coordinator.generateComicWithRetry(
		ctx,
		entry,
		func() {},
		func(attemptCtx context.Context) (batchfactoryv11.ComicCollection, error) {
			attempts++
			cancel()
			<-attemptCtx.Done()
			return batchfactoryv11.ComicCollection{}, attemptCtx.Err()
		},
		func(context.Context, time.Duration) error { return nil },
	)
	if err == nil || attempts != 1 {
		t.Fatalf("err=%v attempts=%d, want one attempt after overall context cancellation", err, attempts)
	}
}

func TestComicAutomationTimeoutDoesNotScheduleAnotherWholeBookRun(t *testing.T) {
	if shouldScheduleAutomationTimeoutRetry("comic_images") {
		t.Fatal("comic image timeout must not schedule another whole-book retry")
	}
	if shouldScheduleAutomationTimeoutRetry("comic_asset_images") {
		t.Fatal("comic asset image timeout must not schedule another whole-book retry")
	}
	if !shouldScheduleAutomationTimeoutRetry("video") {
		t.Fatal("video timeout should keep its existing bounded retry path")
	}
}

type comicAutomationTextProvider struct {
	responses []string
	calls     int
}

type concurrentComicTextProvider struct {
	mu     sync.Mutex
	active int
	max    int
	calls  int
}

func (p *concurrentComicTextProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.mu.Lock()
	p.active++
	p.calls++
	if p.active > p.max {
		p.max = p.active
	}
	p.mu.Unlock()
	time.Sleep(20 * time.Millisecond)
	p.mu.Lock()
	p.active--
	p.mu.Unlock()
	return comicAutomationResponse(request), nil
}

func (p *comicAutomationTextProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	if len(p.responses) == 0 {
		return comicAutomationResponse(request), nil
	}
	response := p.responses[0]
	p.responses = p.responses[1:]
	return response, nil
}

func comicAutomationResponse(request batchfactoryv11.TextCompletionRequest) string {
	if strings.Contains(request.SystemPrompt, "漫画整书首次生成总控") {
		return `{"characters":[{"name":"林晚","prompt":"现代女性，黑色长发"},{"name":"陆沉","prompt":"现代男性，深色西装"}],"scenes":[{"name":"办公室","prompt":"现代办公室"}],"props":[],"storyboard":[{"end_boundary_id":"b1","characters":["林晚"],"props":[],"scene":"办公室","screen_prompt":"单页漫画导演说明","image_prompt":"林晚走进办公室，完整中文漫画页","roles":["林晚"],"role_stage":{"林晚":"林晚"},"scene_presets":["办公室"]}],"source_coverage":{"source_complete":true,"has_remaining_source":false}}`
	}
	if strings.Contains(request.SystemPrompt, "漫画分页编辑") {
		return `not-json-use-fallback`
	}
	// 当前漫画链路按“分页后逐页提取提示词”执行，单页合同要求模型
	// 返回单个页面对象；保留对整本 ordered_pages 合同的模拟兼容。
	if strings.Contains(request.UserPrompt, `"current_page"`) && !strings.Contains(request.UserPrompt, `"ordered_pages"`) {
		return `{"screen_prompt":"单页漫画导演说明","image_prompt":"林晚走进办公室，完整中文漫画页","roles":["林晚"],"role_stage":{"林晚":"林晚"},"scene_presets":["办公室"]}`
	}
	if strings.Contains(request.SystemPrompt, "comic assets") || strings.Contains(request.SystemPrompt, "漫画资产") {
		return `{"characters":[{"name":"林晚","prompt":"现代女性，黑色长发"},{"name":"陆沉","prompt":"现代男性，深色西装"}],"scenes":[{"name":"办公室","prompt":"现代办公室"}],"props":[]}`
	}
	var payload struct {
		Pages []struct {
			VideoID string `json:"video_id"`
		} `json:"ordered_pages"`
	}
	_ = json.Unmarshal([]byte(request.UserPrompt), &payload)
	pages := make([]map[string]any, 0, len(payload.Pages))
	for _, page := range payload.Pages {
		pages = append(pages, map[string]any{"video_id": page.VideoID, "screen_prompt": "单页漫画导演说明", "image_prompt": "林晚走进办公室，完整中文漫画页", "roles": []string{"林晚"}, "role_stage": map[string]string{"林晚": "林晚"}, "scene_presets": []string{"办公室"}})
	}
	raw, _ := json.Marshal(map[string]any{"pages": pages})
	return string(raw)
}

func TestComicAutomationStopsAfterWholeBookCollectionWithoutAudioOrVideo(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "漫画书", SourceText: "林晚走进办公室，陆沉抬头看向她。"}}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig, _ := json.Marshal(map[string]any{"comic": map[string]any{
		"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"id": "batch-comic-assets", "body": "提取漫画人物、场景和道具"},
		"imageDerivation": map[string]any{"body": "整理成可直接出图的提示词"},
	}})
	mode, _ := json.Marshal("comic")
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{
		Patch: batchfactoryv11.SettingsPatch{"productionMode": mode, "aiPromptConfig": comicConfig, "textModelId": json.RawMessage(`"text-1"`), "imageModelId": json.RawMessage(`"image-1"`)}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(ctx, "alice", batch.ID)
	book := batch.Books[0]
	provider := &comicAutomationTextProvider{}
	stages := &batchfactoryv11.BookStageService{
		Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider},
	}
	coordinator := NewBatchAutomationCoordinator(store, nil, stages, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
	generated := false
	coordinator.SetComicBookGenerator(func(_ context.Context, owner, batchID, bookID string) (batchfactoryv11.ComicCollection, error) {
		generated = owner == "alice" && batchID == batch.ID && bookID == book.ID
		latest, getErr := store.GetBatch(ctx, owner, batchID)
		if getErr != nil {
			return batchfactoryv11.ComicCollection{}, getErr
		}
		if len(latest.Books[0].Videos) != 1 || latest.Books[0].Videos[0].ComicImagePrompt == "" || latest.Books[0].Videos[0].ComicSourceSegment == nil {
			t.Fatalf("漫画分页提示词未在整本出图前保存: %+v", latest.Books[0].Videos)
		}
		return batchfactoryv11.ComicCollection{ID: "collection-1", Items: []batchfactoryv11.ComicCollectionItem{{VideoID: latest.Books[0].Videos[0].ID}}}, nil
	})
	entry := map[string]any{"bookId": book.ID, "status": "running", "stage": "assets"}
	job := map[string]any{"id": "job-comic", "owner": "alice", "batchId": batch.ID, "state": "running", "runMode": "video_no_submit", "books": map[string]any{book.ID: entry}}

	coordinator.run(ctx, "alice", batch.ID, "job-comic", job, batch)

	if !generated || provider.calls != 3 {
		t.Fatalf("generated=%v AI calls=%d entry=%#v, want separate comic asset, segmentation and prompt calls", generated, provider.calls, entry)
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(latest.Books[0].AssetRecords) != 3 || len(latest.Books[0].Videos) != 1 {
		t.Fatalf("unified comic result was not split into assets and pages: %+v", latest.Books[0])
	}
	if latest.Books[0].Videos[0].ComicPrompt != "单页漫画导演说明" || latest.Books[0].Videos[0].ComicImagePrompt == "" {
		t.Fatalf("unified comic page prompts were not persisted separately: %+v", latest.Books[0].Videos[0])
	}
	if entry["status"] != "completed" || entry["stage"] != "ready_for_upload" || entry["comicCollectionId"] != "collection-1" {
		t.Fatalf("comic automation entry=%#v", entry)
	}
	if jobs, _ := store.ListProductionJobs(ctx, "alice", batch.ID); len(jobs) != 0 {
		t.Fatalf("comic automation must not create video production: %+v", jobs)
	}
	if merges, _ := store.ListMergeJobs(ctx, "alice", batch.ID); len(merges) != 0 {
		t.Fatalf("comic automation must not create video merges: %+v", merges)
	}
}

func TestComicAutomationHonorsBookConcurrency(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{
		{Title: "第一本", SourceText: "林晚走进办公室。"},
		{Title: "第二本", SourceText: "林晚走进办公室。"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig, _ := json.Marshal(map[string]any{"comic": map[string]any{"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"id": "batch-comic-assets", "body": "提取漫画人物、场景和道具"}, "imageDerivation": map[string]any{"body": "整理成可直接出图的提示词"}}})
	mode, _ := json.Marshal("comic")
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{"productionMode": mode, "aiPromptConfig": comicConfig, "textModelId": json.RawMessage(`"text-1"`), "imageModelId": json.RawMessage(`"image-1"`)}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(ctx, "alice", batch.ID)
	provider := &concurrentComicTextProvider{}
	coordinator := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{
		Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider},
	}, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
	entries := map[string]any{}
	for _, book := range batch.Books {
		entries[book.ID] = map[string]any{"bookId": book.ID, "status": "running", "stage": "assets"}
	}
	job := map[string]any{"id": "job-concurrent", "owner": "alice", "batchId": batch.ID, "state": "running", "runMode": "storyboard_only", "concurrency": 2, "books": entries}

	coordinator.run(ctx, "alice", batch.ID, "job-concurrent", job, batch)

	provider.mu.Lock()
	maxActive := provider.max
	calls := provider.calls
	provider.mu.Unlock()
	if maxActive < 2 {
		t.Fatalf("maximum simultaneous AI calls=%d, want at least 2", maxActive)
	}
	if calls != 6 {
		t.Fatalf("AI calls=%d, want three independent comic preparation calls for each of 2 books", calls)
	}
	if job["state"] != "completed" {
		t.Fatalf("job state=%v, want completed", job["state"])
	}
}

func TestAutomationRunSkipsBlockedBook(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{
		{Title: "健康书", SourceText: "林晚走进办公室。"},
		{Title: "缺模型书", SourceText: "林晚走进办公室。"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig, _ := json.Marshal(map[string]any{"comic": map[string]any{"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"id": "batch-comic-assets", "body": "提取漫画人物、场景和道具"}, "imageDerivation": map[string]any{"body": "整理成可直接出图的提示词"}}})
	mode, _ := json.Marshal("comic")
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{"productionMode": mode, "aiPromptConfig": comicConfig, "textModelId": json.RawMessage(`"text-1"`), "imageModelId": json.RawMessage(`"image-1"`)}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(ctx, "alice", batch.ID)
	var healthy, blocked batchfactoryv11.Book
	for _, candidate := range batch.Books {
		if candidate.Title == "缺模型书" {
			blocked = candidate
		} else {
			healthy = candidate
		}
	}
	// 书级空值覆盖批级配置，模拟“只有这一本没配文本模型”。
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batch.ID, BookID: blocked.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{"textModelId": json.RawMessage(`""`)}, ExpectedRevision: blocked.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(ctx, "alice", batch.ID)
	provider := &comicAutomationTextProvider{}
	coordinator := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{
		Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider},
	}, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
	coordinator.SetComicBookGenerator(func(_ context.Context, _, _, _ string) (batchfactoryv11.ComicCollection, error) {
		latest, getErr := store.GetBatch(ctx, "alice", batch.ID)
		if getErr != nil {
			return batchfactoryv11.ComicCollection{}, getErr
		}
		for _, candidate := range latest.Books {
			if candidate.ID == healthy.ID {
				return batchfactoryv11.ComicCollection{ID: "collection-1", Items: []batchfactoryv11.ComicCollectionItem{{VideoID: candidate.Videos[0].ID}}}, nil
			}
		}
		return batchfactoryv11.ComicCollection{}, batchfactoryv11.ErrNotFound
	})
	entries := map[string]any{}
	for _, book := range batch.Books {
		entries[book.ID] = map[string]any{"bookId": book.ID, "status": "running", "stage": "assets"}
	}
	job := map[string]any{"id": "job-blocked", "owner": "alice", "batchId": batch.ID, "state": "running", "runMode": "video_no_submit", "books": entries}

	coordinator.run(ctx, "alice", batch.ID, "job-blocked", job, batch)

	blockedEntry := entries[blocked.ID].(map[string]any)
	if blockedEntry["status"] != "blocked" || blockedEntry["stage"] != "preflight" || !strings.Contains(fmt.Sprint(blockedEntry["message"]), "文本模型") {
		t.Fatalf("blocked entry=%#v, want blocked/preflight with reason", blockedEntry)
	}
	healthyEntry := entries[healthy.ID].(map[string]any)
	if healthyEntry["status"] != "completed" {
		t.Fatalf("healthy entry=%#v, want completed despite blocked neighbour", healthyEntry)
	}
}

func TestComicBootstrapReadyRequiresContentSegmentsAndCompletePrompts(t *testing.T) {
	effective := batchfactoryv11.SettingsPatch{"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true}}`)}
	book := batchfactoryv11.Book{
		AssetRecords: []batchfactoryv11.BookAsset{{ID: "asset-1", Kind: "character", Name: "林晚", Source: "comic", Prompt: "人物"}},
		Videos: []batchfactoryv11.Video{
			{ID: "page-1", ComicPrompt: "第一页", ComicImagePrompt: "第一页图片", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: "第一段"}},
			{ID: "page-2", ComicPrompt: "第二页", ComicImagePrompt: "第二页图片", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: "第二段"}},
		},
	}
	if !comicBootstrapReady(book, effective) {
		t.Fatal("content-driven pages with source and prompts must be reusable")
	}
	book.Videos[1].ComicSourceSegment = nil
	if comicBootstrapReady(book, effective) {
		t.Fatal("a page missing its source segment must resume from bootstrap")
	}
}

func TestComicPreparationActionRepairsOnlyMissingPagePrompts(t *testing.T) {
	effective := batchfactoryv11.SettingsPatch{"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true}}`)}
	book := batchfactoryv11.Book{
		AssetRecords: []batchfactoryv11.BookAsset{{ID: "asset-1", Kind: "character", Name: "林晚", Source: "comic", Prompt: "人物"}},
		Videos: []batchfactoryv11.Video{
			{ID: "page-1", ComicPrompt: "第一页", ComicImagePrompt: "第一页图片", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: "第一段"}},
			{ID: "page-2", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: "第二段"}},
		},
	}

	if got := selectComicPreparationAction(book, effective); got != comicPreparationMissingPrompts {
		t.Fatalf("comic preparation action = %q, want missing-page prompt recovery", got)
	}
	book.Videos[1].ComicSourceSegment = nil
	if got := selectComicPreparationAction(book, effective); got != comicPreparationBootstrap {
		t.Fatalf("comic preparation action without a source segment = %q, want whole-book bootstrap", got)
	}
}

func TestComicPreparationActionRecoversAmbiguousComicAssetIdentity(t *testing.T) {
	effective := batchfactoryv11.SettingsPatch{"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true}}`)}
	book := batchfactoryv11.Book{
		AssetRecords: []batchfactoryv11.BookAsset{{ID: "asset-1", Kind: "character", Name: "我", Source: "comic", Prompt: "人物"}},
		Videos:       []batchfactoryv11.Video{{ID: "page-1", ComicPrompt: "第一页", ComicImagePrompt: "第一页图片", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: "第一段"}}},
	}
	if got := selectComicPreparationAction(book, effective); got != comicPreparationBootstrap {
		t.Fatalf("ambiguous comic asset action = %q, want automatic bootstrap recovery", got)
	}
}

func TestComicPreparationActionDoesNotTreatVideoAssetsAsComicAssets(t *testing.T) {
	effective := batchfactoryv11.SettingsPatch{"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true}}`)}
	book := batchfactoryv11.Book{
		AssetRecords: []batchfactoryv11.BookAsset{{ID: "video-asset", Source: "director", Prompt: "视频人物"}},
		Videos: []batchfactoryv11.Video{
			{ID: "page-1", ComicPrompt: "第一页", ComicImagePrompt: "第一页图片", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: "第一段"}},
		},
	}
	if got := selectComicPreparationAction(book, effective); got != comicPreparationBootstrap {
		t.Fatalf("comic preparation action with only video assets = %q, want whole-book bootstrap", got)
	}
}

func TestRecoverInterruptedJobsMarksPersistedRunningJobsStopped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "automation.json")
	old := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {
			"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "running", "updatedAt": old,
			"counts": automationCounts(2, 2, 0, 0),
			"books": map[string]any{
				"book-1": map[string]any{"bookId": "book-1", "status": "running", "stage": "video"},
				"book-2": map[string]any{"bookId": "book-2", "status": "completed", "stage": "completed"},
			},
		},
	}}
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	coordinator := NewBatchAutomationCoordinator(nil, nil, nil, nil, nil, nil, nil, NewFileAutomationStatusReader(path), func() time.Time { return now })

	if err := coordinator.RecoverInterruptedJobs(); err != nil {
		t.Fatal(err)
	}

	// 状态存储已拆分：通过协调器读回（内部会完成旧文件迁移）。
	got, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	job := got.Jobs["job-1"]
	if job["state"] != "stopped" || !strings.Contains(automationStringValue(job["lastError"], ""), "服务重启") {
		t.Fatalf("recovered job = %#v, want stopped with restart interruption error", job)
	}
	counts, _ := job["counts"].(map[string]any)
	if running, ok := counts["running"].(float64); !ok || running != 0 {
		t.Fatalf("recovered counts = %#v, running must be zero", counts)
	}
	books, _ := job["books"].(map[string]any)
	book, _ := books["book-1"].(map[string]any)
	if book["status"] != "stopped" || !strings.Contains(automationStringValue(book["message"], ""), "中断") {
		t.Fatalf("recovered book = %#v, want stopped with interruption message", book)
	}
}

func TestResumeInterruptedAutomationJobOnlyReopensServiceRestart(t *testing.T) {
	now := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	restartJob := map[string]any{
		"state":     "stopped",
		"lastError": "服务重启，自动化已中断，请重新立即执行。",
		"books": map[string]any{
			"book-1": map[string]any{"status": "stopped", "stage": "stopped", "message": "服务重启，自动化已中断，请重新立即执行。"},
			"book-2": map[string]any{"status": "completed", "stage": "published"},
		},
	}
	if !resumeInterruptedAutomationJob(restartJob, now) {
		t.Fatal("service-restart job should be resumed")
	}
	if restartJob["state"] != "running" || restartJob["lastError"] != "" {
		t.Fatalf("reopened job = %#v", restartJob)
	}
	books, _ := restartJob["books"].(map[string]any)
	book, _ := books["book-1"].(map[string]any)
	// 书级恢复必须等逐书预检通过后才发生，这里只允许任务级重开。
	if book["status"] != "stopped" || book["stage"] != "stopped" {
		t.Fatalf("book must stay untouched before preflight, got %#v", book)
	}

	manualStop := map[string]any{"state": "stopped", "lastError": "", "books": map[string]any{}}
	if resumeInterruptedAutomationJob(manualStop, now) {
		t.Fatal("manual stop must not be resumed automatically")
	}
}

func TestResumeInterruptedAutomationJobKeepsCompletedComicCollectionForPublish(t *testing.T) {
	now := time.Date(2026, 11, 2, 5, 10, 0, 0, time.UTC)
	job := map[string]any{
		"state": "stopped", "runMode": "full_submit", "lastError": automationServiceRestartMessage,
		"books": map[string]any{
			"book-1": map[string]any{
				"status": "stopped", "stage": "comic_publish", "comicCollectionId": "collection-existing",
			},
		},
	}

	if !resumeInterruptedAutomationJob(job, now) {
		t.Fatal("service-restart job should be resumed")
	}
	book := job["books"].(map[string]any)["book-1"].(map[string]any)
	// 阶段不能被重置：恢复入口只重开任务，书级 stage 与已完成产物保持原样。
	if book["stage"] != "comic_publish" {
		t.Fatalf("publish-ready book stage must be preserved, got %#v", book)
	}
	if book["comicCollectionId"] != "collection-existing" {
		t.Fatalf("existing collection must be retained, got %#v", book)
	}
}

func TestApplyResumePreflightKeepsCompletedComicStage(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "漫画书"}}})
	if err != nil {
		t.Fatal(err)
	}
	settings := batchfactoryv11.SettingsPatch{
		"textModelId": json.RawMessage(`"text-1"`),
	}
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: settings, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
	entry := map[string]any{
		"bookId": batch.Books[0].ID, "status": "stopped", "stage": "comic_images",
		"promptReady": true, "comicImageIDs": []any{"page-1"},
	}
	job := map[string]any{"id": "job-1", "owner": "alice", "batchId": batch.ID, "runMode": "storyboard_only", "books": map[string]any{batch.Books[0].ID: entry}}

	coordinator.applyResumePreflight(ctx, "alice", batch, job, time.Unix(1700000000, 0))

	if entry["status"] != "running" || entry["stage"] != "comic_images" {
		t.Fatalf("resumed entry = %#v, want running at the preserved comic_images stage", entry)
	}
	if entry["promptReady"] != true {
		t.Fatalf("completed prompt state must be kept: %#v", entry)
	}
}

func TestResumeInterruptedJobsBlocksBooksThatFailPreflight(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "漫画书"}}})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "automation.json")
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {
			"id": "job-1", "owner": "alice", "batchId": batch.ID, "state": "stopped", "runMode": "storyboard_only",
			"lastError": automationServiceRestartMessage,
			"books":     map[string]any{batch.Books[0].ID: map[string]any{"bookId": batch.Books[0].ID, "status": "stopped", "stage": "stopped"}},
		},
	}}
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	stages := &batchfactoryv11.BookStageService{Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: &comicAutomationTextProvider{}}}
	coordinator := NewBatchAutomationCoordinator(store, nil, stages, nil, nil, nil, nil, NewFileAutomationStatusReader(path), func() time.Time { return time.Unix(1700000000, 0) })

	if err := coordinator.ResumeInterruptedJobs(ctx); err != nil {
		t.Fatal(err)
	}

	// 恢复后会异步 launchRun；等它收尾再断言，避免临时目录清理时的写入竞态。
	job := waitForTerminalAutomationJob(t, coordinator, "alice", batch.ID)
	entry, _ := job["books"].(map[string]any)[batch.Books[0].ID].(map[string]any)
	if entry["status"] != "blocked" {
		t.Fatalf("book missing text model must be blocked, got %#v", entry)
	}
	if reason := automationStringValue(entry["message"], ""); !strings.Contains(reason, "文本模型") {
		t.Fatalf("blocked reason = %q", reason)
	}
}

func TestAutomationResumeOnBootIsDisabledByDefault(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "批次", Books: []batchfactoryv11.CreateBookInput{{Title: "书"}}})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "automation.json")
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {
			"id": "job-1", "owner": "alice", "batchId": batch.ID, "state": "stopped", "runMode": "full_submit",
			"lastError": automationServiceRestartMessage,
			"books":     map[string]any{batch.Books[0].ID: map[string]any{"bookId": batch.Books[0].ID, "status": "stopped", "stage": "stopped"}},
		},
	}}
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	newCoordinator := func() *BatchAutomationCoordinator {
		return NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, NewFileAutomationStatusReader(path), func() time.Time { return time.Unix(1700000000, 0) })
	}

	t.Setenv("QIANTIE_AUTOMATION_RESUME_ON_BOOT", "")
	disabled := newCoordinator()
	resumeAutomationOnBoot(disabled)
	job := disabled.loadJobForTest(t, "alice", batch.ID)
	if job["state"] != "stopped" {
		t.Fatalf("boot resume must be off by default, got job state %#v", job["state"])
	}

	// 打开开关且缺少执行器时，resume 会走全局预检并把任务落成 failed，
	// 而不是把书重新置为 running。
	t.Setenv("QIANTIE_AUTOMATION_RESUME_ON_BOOT", "1")
	enabled := newCoordinator()
	resumeAutomationOnBoot(enabled)
	job = enabled.loadJobForTest(t, "alice", batch.ID)
	if job["state"] != "failed" || automationStringValue(job["lastError"], "") == "" {
		t.Fatalf("resume without executors must fail the job, got %#v", job)
	}
	entry, _ := job["books"].(map[string]any)[batch.Books[0].ID].(map[string]any)
	if entry["status"] != "stopped" {
		t.Fatalf("preflight-failed job must not reopen books, got %#v", entry)
	}
}

func (c *BatchAutomationCoordinator) loadJobForTest(t *testing.T, owner, batchID string) map[string]any {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	for _, job := range state.Jobs {
		if automationStringValue(job["owner"], "") == owner && automationStringValue(job["batchId"], "") == batchID {
			return job
		}
	}
	t.Fatalf("job not found for owner=%s batch=%s", owner, batchID)
	return nil
}

func waitForTerminalAutomationJob(t *testing.T, c *BatchAutomationCoordinator, owner, batchID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		job := c.loadJobForTest(t, owner, batchID)
		switch automationStringValue(job["state"], "") {
		case "completed", "failed":
			return job
		}
		if time.Now().After(deadline) {
			t.Fatalf("job did not reach a terminal state: %#v", job)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestNormalizeAutomationScheduleModeSeparatesExecutionAndSubmission(t *testing.T) {
	if got := normalizeAutomationScheduleMode("execute"); got != "execute" {
		t.Fatalf("execute mode=%q", got)
	}
	if got := normalizeAutomationScheduleMode("submit"); got != "submit" {
		t.Fatalf("submit mode=%q", got)
	}
	if got := normalizeAutomationScheduleMode("unknown"); got != "execute" {
		t.Fatalf("unknown mode=%q, want execute", got)
	}
}

func TestAutomationJobDueUsesPersistedScheduledTime(t *testing.T) {
	now := time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC)
	job := map[string]any{"scheduledAt": now.Add(2 * time.Minute).Format(time.RFC3339)}
	if automationJobDue(job, now) {
		t.Fatal("job became due before scheduledAt")
	}
	if !automationJobDue(job, now.Add(2*time.Minute)) {
		t.Fatal("job did not become due at scheduledAt")
	}
}

func TestAutomationGlobalPreflightRequiresServerSideExecutor(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{stages: &batchfactoryv11.BookStageService{}}
	batch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"video"`)}}}
	err := coordinator.validateAutomationGlobalPreflight(context.Background(), "alice", batch, "video_no_submit")
	if err == nil || !strings.Contains(err.Error(), "文本模型执行器未配置") {
		t.Fatalf("preflight error=%v, want missing server executor", err)
	}
}

func TestAutomationGlobalPreflightIgnoresBookConfig(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{}
	// 全局预检签名不再接收书列表：缺执行器时报全局原因，不含任何书级原因。
	err := coordinator.validateAutomationGlobalPreflight(context.Background(), "alice", batchfactoryv11.Batch{}, "storyboard_only")
	if err == nil || !strings.Contains(err.Error(), "文本模型执行器未配置") {
		t.Fatalf("err=%v, want global executor reason", err)
	}
	if strings.Contains(err.Error(), "《") || strings.Contains(err.Error(), "发布配置未绑定") {
		t.Fatalf("err=%v, must not contain per-book reasons", err)
	}
}

func TestAutomationBookPreflightReason(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{}
	batch := batchfactoryv11.Batch{}
	book := batchfactoryv11.Book{ID: "book-1", Title: "书一"}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batch, book, "execute_only"); !strings.Contains(reason, "文本模型") {
		t.Fatalf("reason=%q, want missing text model", reason)
	}
	book.SettingsState.Patch = batchfactoryv11.SettingsPatch{"textModelId": json.RawMessage(`"text-1"`)}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batch, book, "storyboard_only"); reason != "" {
		t.Fatalf("storyboard_only should pass, got %q", reason)
	}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batch, book, "execute_only"); !strings.Contains(reason, "视频模型") {
		t.Fatalf("reason=%q, want missing video model", reason)
	}
	book.SettingsState.Patch["videoModelId"] = json.RawMessage(`"video-1"`)
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batch, book, "execute_only"); reason != "" {
		t.Fatalf("execute_only should pass, got %q", reason)
	}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batch, book, "full_submit"); !strings.Contains(reason, "发布配置") {
		t.Fatalf("reason=%q, want missing publish profile", reason)
	}
}

func TestAutomationBookPreflightSkipsComicPublishPreflight(t *testing.T) {
	// comicPublishPreflight 现在挪到 run() 的 publish 阶段再查；
	// preflight 只检查出图必需的前置条件，N8 目录挂了不能卡出图。
	preflightCalled := false
	coordinator := &BatchAutomationCoordinator{
		comicPublishPreflight: func(context.Context, batchfactoryv11.SettingsPatch) error {
			preflightCalled = true
			return fmt.Errorf("漫画发布设置不完整：没有可用指定音频")
		},
	}
	batch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"productionMode":       json.RawMessage(`"comic"`),
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1","organizationName":"博量","producerAdminName":"制作者","workIds":[9001]}`),
	}}}
	book := batchfactoryv11.Book{ID: "book-1", SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"textModelId":     json.RawMessage(`"text-1"`),
		"imageModelId":    json.RawMessage(`"image-1"`),
		"publishSettings": json.RawMessage(`{"websiteProfileId":"p1"}`),
	}}}
	reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batch, book, "full_submit")
	if reason != "" {
		t.Fatalf("reason=%q, want empty: preflight should not call publish preflight", reason)
	}
	if preflightCalled {
		t.Fatal("comicPublishPreflight must not be called during preflight anymore")
	}

	// 漫画批次保存的是 comicPublishSettings（组织/音频分类/制作者），
	// 没有 publishSettings。之前统一只查 publishSettings，导致漫画批次
	// 在派发前被判“发布配置未绑定网站或组织”，整批一本书都跑不起来。
	comicBatch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"productionMode":       json.RawMessage(`"comic"`),
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1","organizationName":"博量","audioCategoryId":"12","producerAdminName":"制作者"}`),
	}}}
	comicBook := batchfactoryv11.Book{ID: "book-2", SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"textModelId":  json.RawMessage(`"text-1"`),
		"imageModelId": json.RawMessage(`"image-1"`),
	}}}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", comicBatch, comicBook, "full_submit"); reason != "" {
		t.Fatalf("reason=%q, want empty: comic mode should bind comicPublishSettings", reason)
	}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", comicBatch, comicBook, "full_submit"); strings.Contains(reason, "发布配置未绑定") {
		t.Fatalf("reason=%q, must not report missing publish profile", reason)
	}

	// 视频批次仍然只认 publishSettings：带了 comicPublishSettings 也不能放行。
	videoBatch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"productionMode":       json.RawMessage(`"video"`),
		"comicPublishSettings": json.RawMessage(`{"organizationId":"1"}`),
	}}}
	videoBook := batchfactoryv11.Book{ID: "book-3", SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"textModelId":  json.RawMessage(`"text-1"`),
		"videoModelId": json.RawMessage(`"video-1"`),
	}}}
	if reason := coordinator.automationBookPreflightReason(context.Background(), "alice", videoBatch, videoBook, "full_submit"); !strings.Contains(reason, "发布配置未绑定") {
		t.Fatalf("reason=%q, want missing publish profile for video mode", reason)
	}
}

func TestActiveAutomationJobFindsOnlyLiveBatchRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	coordinator := &BatchAutomationCoordinator{path: path}
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"live": {"id": "live", "owner": "alice", "batchId": "batch-1", "state": "running"},
		"done": {"id": "done", "owner": "alice", "batchId": "batch-1", "state": "completed"},
	}}
	if err := coordinator.saveLocked(state); err != nil {
		t.Fatal(err)
	}
	if got := coordinator.activeAutomationJob("alice", "batch-1"); got == nil || got["id"] != "live" {
		t.Fatalf("active job=%#v, want live job", got)
	}
	if got := coordinator.activeAutomationJob("alice", "batch-2"); got != nil {
		t.Fatalf("active job for another batch=%#v, want nil", got)
	}
}

func TestStartPersistsScheduledExecutionWithoutLaunchingBeforeDueTime(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	now := time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC)
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "定时批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	presetStore, err := NewFileAutomationPresetStore(filepath.Join(t.TempDir(), "presets.json"), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	preset, err := presetStore.Create("alice", "漫画预设", map[string]any{"productionMode": "comic"})
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(store, presetStore, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, nil, func() time.Time { return now })
	job, err := coordinator.Start(ctx, "alice", batch.ID, map[string]any{
		"presetId": preset.ID, "runMode": "storyboard_only", "scheduleMode": "execute",
		"scheduledAt": now.Add(time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatal(err)
	}
	if job["state"] != "scheduled" || job["scheduleMode"] != "execute" || job["scheduledAt"] == "" {
		t.Fatalf("scheduled job=%#v", job)
	}
}

func TestProductionLinesHonorsConfiguredContentRange(t *testing.T) {
	book := batchfactoryv11.Book{SourceText: "第一行\n\n第二行\r\n第三行\n第四行", SourceMetadata: map[string]any{"contentRangeLines": 3}}
	got := productionLines(book)
	if len(got) != 3 || got[0] != "第一行" || got[2] != "第三行" {
		t.Fatalf("productionLines() = %#v, want first three non-empty lines", got)
	}
}

func TestExistingAudioPlanIsReusableForTheSameProductionLines(t *testing.T) {
	book := batchfactoryv11.Book{
		SourceText:     "第一行\n第二行\n第三行",
		SourceMetadata: map[string]any{"contentRangeLines": 3},
		SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
			"audioPlanningEnabled": json.RawMessage("true"),
			"audioLineDurations":   json.RawMessage(`[1.1,2.2,3.3]`),
			"audioDurationSeconds": json.RawMessage(`6.6`),
		}},
	}
	if !hasReusableAudioPlan(book) {
		t.Fatal("an existing valid line-duration plan should be reusable")
	}
}

func TestProductionJobForRequestDoesNotReuseAnOlderCompletedJob(t *testing.T) {
	jobs := []batchfactoryv11.ProductionJob{
		{BookID: "book-1", RequestID: "old-video", Status: batchfactoryv11.ProductionSucceeded, Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded, MediaURL: "old.mp4"}}},
		{BookID: "book-1", RequestID: "current-video", Status: batchfactoryv11.ProductionRunning, Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionRunning}}},
	}
	job, ok := productionJobForRequest(jobs, "book-1", "current-video")
	if !ok || job.RequestID != "current-video" {
		t.Fatalf("productionJobForRequest() = %#v, %v; want current job", job, ok)
	}
}

func TestProductionJobCompleteRequiresSuccessfulMedia(t *testing.T) {
	job := batchfactoryv11.ProductionJob{Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded}}}
	if productionJobComplete(job) {
		t.Fatal("productionJobComplete accepted a task without media URL")
	}
	job.Tasks[0].MediaURL = "/api/batch-factory/v12/media/video-1"
	if !productionJobComplete(job) {
		t.Fatal("productionJobComplete rejected a successful task with media URL")
	}
}

func TestSelectedAutomationBooks_WithBookIDs_ReturnsOnlyRequestedBooks(t *testing.T) {
	batch := batchfactoryv11.Batch{Books: []batchfactoryv11.Book{{ID: "book-a"}, {ID: "book-b"}}}

	books, err := selectedAutomationBooks(batch, []any{"book-b"})

	if err != nil {
		t.Fatalf("selectedAutomationBooks returned error: %v", err)
	}
	if len(books) != 1 || books[0].ID != "book-b" {
		t.Fatalf("selectedAutomationBooks = %#v, want only book-b", books)
	}
}

func TestSelectedAutomationBooks_WithUnknownBook_RejectsCrossBatchDispatch(t *testing.T) {
	batch := batchfactoryv11.Batch{Books: []batchfactoryv11.Book{{ID: "book-a"}}}

	_, err := selectedAutomationBooks(batch, []any{"not-in-this-batch"})

	if err == nil {
		t.Fatal("selectedAutomationBooks accepted a book outside the batch")
	}
}

func TestApplyPresetKeepsRequestedComicProductionMode(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本"}}})
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, nil)
	preset := &AutomationPreset{Config: map[string]any{"productionMode": "video", "imageModelId": "image-1"}}

	updated, err := coordinator.applyPreset(ctx, "alice", batch, preset, "comic")
	if err != nil {
		t.Fatal(err)
	}
	if got := comicSettingString(updated.SettingsState.Patch, "productionMode", ""); got != "comic" {
		t.Fatalf("productionMode=%q, want comic", got)
	}
}

func TestApplyPresetHydratesComicPromptBodies(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "presets.json")
	data := `[
		{"id":"batch-comic-assets","module":"batch-factory","name":"漫画资产提示词","kind":"base","version":1,"revision":1,"status":"published","body":"漫画资产正文","protocolLock":{"format":"json","slot":"batch.comic-assets"}},
		{"id":"batch-comic-page","module":"batch-factory","name":"漫画提示词","kind":"base","version":1,"revision":1,"status":"published","body":"漫画页面正文","protocolLock":{"format":"json","slot":"batch.comic-page"}}
	]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本"}}})
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.promptCatalog = prompt.NewCatalog(path)
	// 预设里只有提示词 ID（浏览器保存的紧凑形态），正文由系统预设库冻结补齐。
	preset := &AutomationPreset{Config: map[string]any{
		"productionMode": "comic",
		"aiPromptConfig": map[string]any{"comic": map[string]any{
			"presetId": "batch-comic-page",
			"assets":   map[string]any{"presetId": "batch-comic-assets"},
		}},
	}}

	updated, err := coordinator.applyPreset(ctx, "alice", batch, preset)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(updated.SettingsState.Patch["aiPromptConfig"], &config); err != nil {
		t.Fatal(err)
	}
	comic, _ := config["comic"].(map[string]any)
	if body, _ := comic["body"].(string); body != "漫画页面正文" {
		t.Fatalf("comic page body=%q, want 漫画页面正文", body)
	}
	assets, _ := comic["assets"].(map[string]any)
	if body, _ := assets["body"].(string); body != "漫画资产正文" {
		t.Fatalf("comic assets body=%q, want 漫画资产正文", body)
	}
}

func TestPickAutomationControlJobReopensLatestTerminalJob(t *testing.T) {
	state := automationStateFile{Jobs: map[string]map[string]any{
		"completed-run": {"id": "completed-run", "owner": "bob", "batchId": "batch-1", "state": "completed", "updatedAt": "2026-10-16T10:00:00Z"},
		"failed-run":    {"id": "failed-run", "owner": "bob", "batchId": "batch-1", "state": "failed", "updatedAt": "2026-10-16T12:00:00Z"},
		"other-batch":   {"id": "other-batch", "owner": "bob", "batchId": "batch-2", "state": "failed", "updatedAt": "2026-10-16T13:00:00Z"},
		"other-owner":   {"id": "other-owner", "owner": "alice", "batchId": "batch-1", "state": "failed", "updatedAt": "2026-10-16T14:00:00Z"},
	}}

	job, reopened := pickAutomationControlJob(state, "bob", "batch-1")

	if !reopened || job == nil || job["id"] != "failed-run" {
		t.Fatalf("pickAutomationControlJob() = %#v, %v; want newest terminal job failed-run reopened", job, reopened)
	}
}

func TestPickAutomationControlJobPrefersActiveJob(t *testing.T) {
	state := automationStateFile{Jobs: map[string]map[string]any{
		"finished": {"id": "finished", "owner": "bob", "batchId": "batch-1", "state": "failed", "updatedAt": "2026-10-16T12:00:00Z"},
		"active":   {"id": "active", "owner": "bob", "batchId": "batch-1", "state": "paused", "updatedAt": "2026-10-16T11:00:00Z"},
	}}

	job, reopened := pickAutomationControlJob(state, "bob", "batch-1")

	if reopened || job == nil || job["id"] != "active" {
		t.Fatalf("pickAutomationControlJob() = %#v, %v; want active job without reopen", job, reopened)
	}
}

func TestPickAutomationControlJobReturnsNilWithoutAnyJob(t *testing.T) {
	state := automationStateFile{Jobs: map[string]map[string]any{
		"other-batch": {"id": "other-batch", "owner": "bob", "batchId": "batch-2", "state": "failed"},
	}}

	if job, reopened := pickAutomationControlJob(state, "bob", "batch-1"); job != nil || reopened {
		t.Fatalf("pickAutomationControlJob() = %#v, %v; want nil for batch without jobs", job, reopened)
	}
}

func TestResetRetryableAutomationBooksRestartsOnlyFailedEntries(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	books := map[string]any{
		"done": map[string]any{"status": "completed", "stage": "completed"},
		"bad":  map[string]any{"status": "failed", "stage": "video", "error": "video failed"},
		"hold": map[string]any{"status": "failed", "stage": "assets", "error": "assets failed"},
	}

	resetRetryableAutomationBooks(books, map[string]bool{"bad": true}, now)

	done, _ := books["done"].(map[string]any)
	if done["status"] != "completed" {
		t.Fatalf("completed entry was touched: %#v", done)
	}
	bad, _ := books["bad"].(map[string]any)
	if bad["status"] != "running" || bad["stage"] != "assets" || bad["error"] != "" || bad["updatedAt"] != now || bad["retryRequested"] != true {
		t.Fatalf("selected failed entry not reset: %#v", bad)
	}
	hold, _ := books["hold"].(map[string]any)
	if hold["status"] != "failed" {
		t.Fatalf("unselected failed entry should stay failed: %#v", hold)
	}
}

func TestResetRetryableAutomationBooksKeepsComicPublishStage(t *testing.T) {
	now := time.Date(2026, 11, 2, 5, 30, 0, 0, time.UTC)
	books := map[string]any{
		"publish": map[string]any{"status": "failed", "stage": "comic_publish", "comicCollectionId": "collection-existing", "error": "missing settings"},
	}

	resetRetryableAutomationBooks(books, nil, now)

	entry := books["publish"].(map[string]any)
	if entry["status"] != "running" || entry["stage"] != "comic_publish" || entry["comicCollectionId"] != "collection-existing" {
		t.Fatalf("publish retry must keep its collection and stage, got %#v", entry)
	}
}

func TestResetRetryableAutomationBooksWithoutFilterRestartsAllFailures(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	books := map[string]any{
		"failed":  map[string]any{"status": "failed"},
		"blocked": map[string]any{"status": "blocked"},
		"paused":  map[string]any{"status": "paused"},
		"stopped": map[string]any{"status": "stopped", "stage": "stopped", "error": "服务重启"},
		"running": map[string]any{"status": "running"},
	}

	resetRetryableAutomationBooks(books, nil, now)

	for id, want := range map[string]string{"failed": "running", "blocked": "running", "paused": "running", "stopped": "running", "running": "running"} {
		entry, _ := books[id].(map[string]any)
		if entry["status"] != want {
			t.Fatalf("%s status = %v, want %v", id, entry["status"], want)
		}
	}
	stopped, _ := books["stopped"].(map[string]any)
	if stopped["stage"] != "assets" || stopped["error"] != "" || stopped["updatedAt"] != now {
		t.Fatalf("stopped entry was not reset for retry: %#v", stopped)
	}
}

func TestAutomationRequiresOpeningStageWhenEnabled(t *testing.T) {
	batch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{"openingEnabled": json.RawMessage("true")}}}
	book := batchfactoryv11.Book{}
	if !automationOpeningEnabled(batch, book) {
		t.Fatal("opening-enabled batch must run the opening stage after director")
	}
}

func TestVideoAutomationAssetImagesSwitchGatesOnlyTheAssetsStage(t *testing.T) {
	cases := []struct {
		name          string
		setting       json.RawMessage
		precreateBook bool
		wantCalls     int
		wantSkipLog   bool
	}{
		{name: "default keeps the assets stage", setting: nil, wantCalls: 2},
		{name: "explicit false skips the assets stage", setting: json.RawMessage("false"), wantCalls: 1, wantSkipLog: true},
		{name: "disabled keeps existing assets", setting: json.RawMessage("false"), precreateBook: true, wantCalls: 1, wantSkipLog: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := batchfactoryv11.NewMemoryStore()
			batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "视频批次", Books: []batchfactoryv11.CreateBookInput{{Title: "视频书", SourceText: "林晚推门进入客厅。"}}})
			if err != nil {
				t.Fatal(err)
			}
			var existingAssetID string
			if tc.precreateBook {
				asset, err := store.CreateBookAsset(ctx, "alice", batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Source: "manual", Name: "手工角色", Prompt: "人物"})
				if err != nil {
					t.Fatal(err)
				}
				existingAssetID = asset.ID
			}
			textModel, _ := json.Marshal("text-1")
			patch := batchfactoryv11.SettingsPatch{"textModelId": textModel}
			if tc.setting != nil {
				patch["assetImagesEnabled"] = tc.setting
			}
			if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: patch, ExpectedRevision: batch.Revision}); err != nil {
				t.Fatal(err)
			}
			batch, err = store.GetBatch(ctx, "alice", batch.ID)
			if err != nil {
				t.Fatal(err)
			}
			// 预置可复用的配音计划，测试只关心资产开关对阶段编排的影响。
			book := batch.Books[0]
			audioPlan := batchfactoryv11.SettingsPatch{
				"audioPlanningEnabled": json.RawMessage("true"),
				"audioLineDurations":   json.RawMessage("[9.0]"),
				"audioDurationSeconds": json.RawMessage("9.0"),
				"ttsAudioArtifactId":   json.RawMessage(`"tts-test"`),
			}
			if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batch.ID, BookID: book.ID}, batchfactoryv11.SettingsUpdate{Patch: audioPlan, ExpectedRevision: book.Revision}); err != nil {
				t.Fatal(err)
			}
			batch, err = store.GetBatch(ctx, "alice", batch.ID)
			if err != nil {
				t.Fatal(err)
			}
			files := localartifact.NewStore(t.TempDir(), 20<<20)
			if _, err := files.SaveMedia("tts-test", "audio/mpeg", "mp3", bytes.NewReader([]byte("ID3"))); err != nil {
				t.Fatal(err)
			}
			provider := &directorHTTPProvider{output: directorHTTPJSON}
			coordinator := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{
				Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider},
			}, nil, nil, novelfetchworkshop.NewMemoryStore(), nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
			coordinator.SetLocalFiles(files)
			book = batch.Books[0]
			entry := map[string]any{"bookId": book.ID, "status": "running", "stage": "assets"}
			job := map[string]any{"id": "job-assets-toggle", "owner": "alice", "batchId": batch.ID, "state": "running", "runMode": "storyboard_only", "books": map[string]any{book.ID: entry}}

			coordinator.run(ctx, "alice", batch.ID, "job-assets-toggle", job, batch)

			if entry["status"] != "completed" || entry["stage"] != "ready_for_video" || automationStringValue(entry["error"], "") != "" {
				t.Fatalf("entry=%#v, want a completed storyboard stage without errors", entry)
			}
			if provider.calls != tc.wantCalls {
				t.Fatalf("provider calls=%d want %d", provider.calls, tc.wantCalls)
			}
			latest, err := store.GetBatch(ctx, "alice", batch.ID)
			if err != nil {
				t.Fatal(err)
			}
			if tc.setting == nil && len(latest.Books[0].AssetRecords) == 0 {
				t.Fatalf("default switch must still generate assets: %+v", latest.Books[0])
			}
			if tc.precreateBook {
				found := false
				for _, asset := range latest.Books[0].AssetRecords {
					if asset.ID == existingAssetID {
						found = true
					}
				}
				if !found {
					t.Fatalf("existing asset %s was dropped when the switch was off: %+v", existingAssetID, latest.Books[0].AssetRecords)
				}
				for _, asset := range latest.Books[0].AssetRecords {
					images, err := store.ListBookAssetImages(ctx, "alice", batch.ID, book.ID, asset.ID)
					if err != nil {
						t.Fatal(err)
					}
					if len(images) != 0 {
						t.Fatalf("asset %s got images while the switch was off: %+v", asset.ID, images)
					}
				}
			}
			if tc.wantSkipLog && !strings.Contains(fmt.Sprint(entry["logs"]), "因资产图片开关关闭，已跳过资产生成阶段") {
				t.Fatalf("logs=%v, want the explicit skip reason", entry["logs"])
			}
		})
	}
}

func TestComicAutomationWithAssetImagesDisabledStillRunsTheComicStages(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "漫画书", SourceText: "林晚走进办公室，陆沉抬头看向她。"}}})
	if err != nil {
		t.Fatal(err)
	}
	comicConfig, _ := json.Marshal(map[string]any{"comic": map[string]any{
		"enabled": true, "body": "完整中文漫画页", "assets": map[string]any{"id": "batch-comic-assets", "body": "提取漫画人物、场景和道具"},
		"imageDerivation": map[string]any{"body": "整理成可直接出图的提示词"},
	}})
	mode, _ := json.Marshal("comic")
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{
		Patch: batchfactoryv11.SettingsPatch{"productionMode": mode, "aiPromptConfig": comicConfig, "textModelId": json.RawMessage(`"text-1"`), "imageModelId": json.RawMessage(`"image-1"`), "assetImagesEnabled": json.RawMessage("false")}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(ctx, "alice", batch.ID)
	book := batch.Books[0]
	provider := &comicAutomationTextProvider{}
	coordinator := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{
		Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider},
	}, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
	generated := false
	coordinator.SetComicBookGenerator(func(_ context.Context, owner, batchID, bookID string) (batchfactoryv11.ComicCollection, error) {
		generated = owner == "alice" && batchID == batch.ID && bookID == book.ID
		latest, getErr := store.GetBatch(ctx, owner, batchID)
		if getErr != nil {
			return batchfactoryv11.ComicCollection{}, getErr
		}
		return batchfactoryv11.ComicCollection{ID: "collection-off", Items: []batchfactoryv11.ComicCollectionItem{{VideoID: latest.Books[0].Videos[0].ID}}}, nil
	})
	entry := map[string]any{"bookId": book.ID, "status": "running", "stage": "assets"}
	job := map[string]any{"id": "job-comic-assets-off", "owner": "alice", "batchId": batch.ID, "state": "running", "runMode": "video_no_submit", "books": map[string]any{book.ID: entry}}

	coordinator.run(ctx, "alice", batch.ID, "job-comic-assets-off", job, batch)

	// 漫画资产文本仍是分页提示词的必需前置，关闭开关只跳过资产图生成，不能把整条漫画链路打断。
	if !generated || provider.calls != 3 {
		t.Fatalf("generated=%v AI calls=%d entry=%#v, want comic preparation and generation to keep running", generated, provider.calls, entry)
	}
	if entry["status"] != "completed" || entry["stage"] != "ready_for_upload" || automationStringValue(entry["error"], "") != "" {
		t.Fatalf("entry=%#v, want a completed comic run without failures", entry)
	}
}

type conflictAfterCaptureStore struct{ batchfactoryv11.Store }

func (s conflictAfterCaptureStore) CaptureBookSource(ctx context.Context, owner, batchID, bookID string, input batchfactoryv11.CaptureBookSourceInput) (batchfactoryv11.Book, error) {
	if _, err := s.Store.CaptureBookSource(ctx, owner, batchID, bookID, input); err != nil {
		return batchfactoryv11.Book{}, err
	}
	return batchfactoryv11.Book{}, batchfactoryv11.ErrConflict
}

type conflictWithoutCaptureStore struct{ batchfactoryv11.Store }

func (s conflictWithoutCaptureStore) CaptureBookSource(context.Context, string, string, string, batchfactoryv11.CaptureBookSourceInput) (batchfactoryv11.Book, error) {
	return batchfactoryv11.Book{}, batchfactoryv11.ErrConflict
}

func TestEnsureAutomationSourceReturnsLatestBookWhenConcurrentCaptureAlreadyWroteBody(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", ID: "source-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	store := conflictAfterCaptureStore{Store: memoryStore}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"并发流程已写入的正文","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, book)

	if err != nil || len(productionLines(updated)) != 1 || updated.SourceText != "并发流程已写入的正文" {
		t.Fatalf("ensureAutomationSource() = %+v, %v; want concurrently persisted source", updated, err)
	}
}

func TestEnsureAutomationSourceReturnsOriginalConflictWhenLatestBookStillHasNoBody(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", ID: "source-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	store := conflictWithoutCaptureStore{Store: memoryStore}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"待保存正文","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	_, err = coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, book)

	if err == nil || !strings.Contains(err.Error(), "保存自动获取正文失败") || !errors.Is(err, batchfactoryv11.ErrConflict) {
		t.Fatalf("err=%v, want original capture conflict", err)
	}
}

func TestEnsureAutomationSourceRecoversFromStaleRevisionAfterFetchFailureBookkeeping(t *testing.T) {
	// 线上链路：第一轮抓取失败 -> persistBatchFetchFailure 写失败元数据(revision+1)
	// -> 回补轮次仍拿着入队时的旧快照 -> 第二轮抓取成功却用旧 revision 写入，报
	// "保存自动获取正文失败：revision conflict"。
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", ID: "source-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	staleBook := batch.Books[0]
	if _, err := memoryStore.UpdateBookMetadata(context.Background(), "alice", batch.ID, staleBook.ID, batchfactoryv11.UpdateBookMetadataInput{
		Metadata:         map[string]any{"fetchError": "upstream timeout", "queueStage": "original"},
		ExpectedRevision: staleBook.Revision,
	}); err != nil {
		t.Fatalf("simulate failure bookkeeping: %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"回补成功保存的正文","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, staleBook)

	if err != nil {
		t.Fatalf("stale queued snapshot must not surface revision conflict: %v", err)
	}
	if len(productionLines(updated)) != 1 || updated.SourceText != "回补成功保存的正文" {
		t.Fatalf("updated source = %+v; want fetched text persisted", updated)
	}
}

type bumpOnceBeforeCaptureStore struct {
	batchfactoryv11.Store
	batchID string
	once    sync.Once
}

func (s *bumpOnceBeforeCaptureStore) CaptureBookSource(ctx context.Context, owner, batchID, bookID string, input batchfactoryv11.CaptureBookSourceInput) (batchfactoryv11.Book, error) {
	s.once.Do(func() {
		batch, err := s.Store.GetBatch(ctx, owner, s.batchID)
		if err != nil {
			return
		}
		for _, book := range batch.Books {
			if book.ID != bookID {
				continue
			}
			// 模拟别的写路径在读取与写入之间递增 revision。
			_, _ = s.Store.UpdateBookMetadata(ctx, owner, s.batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: map[string]any{"concurrentWriter": true}, ExpectedRevision: book.Revision})
			return
		}
	})
	return s.Store.CaptureBookSource(ctx, owner, batchID, bookID, input)
}

func TestEnsureAutomationSourceRetriesCaptureAfterConcurrentRevisionBump(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", ID: "source-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	store := &bumpOnceBeforeCaptureStore{Store: memoryStore, batchID: batch.ID}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"并发后写入成功","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, batch.Books[0])

	if err != nil {
		t.Fatalf("conflict must be retried with a fresh revision, got %v", err)
	}
	if len(productionLines(updated)) != 1 {
		t.Fatalf("source lines = %d, want 1", len(productionLines(updated)))
	}
}

func TestEnsureAutomationSourceDirectFirstFailureDoesNotRefetch(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "直取失败书", ID: "source-1", Platform: "15", SourceMetadata: map[string]any{
			"originalReadStrategy": "direct_first",
			"originalReadStage":    "failed",
			"originalReadError":    "121 upstream business error 400: 参数不完整",
			"contentPending":       false,
		},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"code":200,"data":"不应读取","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	_, err = coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, batch.Books[0])

	if err == nil || !strings.Contains(err.Error(), "书城原文获取失败，等待人工处理") {
		t.Fatalf("err=%v, want direct-first source blocked", err)
	}
	if requests != 0 {
		t.Fatalf("upstream requests=%d, want 0 for blocked direct-first source", requests)
	}
}

func TestEnsureAutomationSourceLegacyEmptyBookStillFetches(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "历史空正文", ID: "source-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"code":200,"data":"历史正文","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, batch.Books[0])

	if err != nil || updated.SourceText != "历史正文" {
		t.Fatalf("updated=%+v err=%v, want legacy source recovery", updated, err)
	}
	if requests != 1 {
		t.Fatalf("upstream requests=%d, want 1 for legacy source recovery", requests)
	}
}

func TestAutomationStateStoreMigration(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "automation-state.json")
	legacyState := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {"id": "job-1", "owner": "alice", "batchId": "b1", "state": "completed", "updatedAt": "2026-01-01T00:00:00Z", "books": map[string]any{}},
		"job-2": {"id": "job-2", "owner": "alice", "batchId": "b2", "state": "stopped", "updatedAt": "2026-01-02T00:00:00Z", "books": map[string]any{}},
	}}
	raw, err := json.Marshal(legacyState)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	coordinator := NewBatchAutomationCoordinator(nil, nil, nil, nil, nil, nil, nil, NewFileAutomationStatusReader(legacy), nil)
	state, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Jobs) != 2 || state.Jobs["job-1"]["state"] != "completed" || state.Jobs["job-2"]["state"] != "stopped" {
		t.Fatalf("migrated state = %#v", state.Jobs)
	}
	// 旧文件已改名，新布局文件已生成且内容等价。
	if _, err := os.Stat(legacy); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy file should be renamed, stat err=%v", err)
	}
	if _, err := os.Stat(legacy + ".migrated"); err != nil {
		t.Fatalf("migrated backup missing: %v", err)
	}
	for _, jobID := range []string{"job-1", "job-2"} {
		payload, err := os.ReadFile(filepath.Join(dir, "automation", "jobs", jobID+".json"))
		if err != nil {
			t.Fatalf("job file missing: %v", err)
		}
		var job map[string]any
		if err := json.Unmarshal(payload, &job); err != nil {
			t.Fatal(err)
		}
		if job["id"] != jobID {
			t.Fatalf("job file content = %#v", job)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "automation", "index.json")); err != nil {
		t.Fatalf("index missing: %v", err)
	}

	// setJob 只写对应 job 文件：改 job-2 不影响 job-1 的内容。
	if err := coordinator.setJob("alice", "b2", "job-2", map[string]any{"id": "job-2", "owner": "alice", "batchId": "b2", "state": "failed", "updatedAt": "2026-01-03T00:00:00Z", "books": map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Jobs["job-2"]["state"] != "failed" || reloaded.Jobs["job-1"]["state"] != "completed" {
		t.Fatalf("after setJob state = %#v", reloaded.Jobs)
	}
}

func TestAutomationPreflightRejectsDirectFirstBlockedBook(t *testing.T) {
	coordinator := NewBatchAutomationCoordinator(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	book := batchfactoryv11.Book{Title: "直取失败书", SourceMetadata: map[string]any{
		"originalReadStrategy": "direct_first",
		"originalReadStage":    "failed",
		"originalReadError":    "121 upstream business error 400: 参数不完整",
	}}

	reason := coordinator.automationBookPreflightReason(context.Background(), "alice", batchfactoryv11.Batch{}, book, "storyboard_only")

	if !strings.Contains(reason, "《直取失败书》书城原文获取失败，等待人工处理") {
		t.Fatalf("reason=%q, want blocked direct-first preflight reason", reason)
	}
}

func TestEnsureAutomationSourceUsesBookContentCaptureCharacters(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", ID: "source-1", Platform: "15", SourceMetadata: map[string]any{"contentCaptureCharacters": 8000}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"` + strings.Repeat("字", 9000) + `","bookinfo":{}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, book)
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(updated.SourceText)); got != 8000 {
		t.Fatalf("captured characters=%d, want 8000", got)
	}
}

func TestEnsureAutomationSourcePersistsFetchedFacetsAndClassification(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", ID: "source-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"正文","bookinfo":{"category":"女频","genre":"现代言情","theme":[{"name":"婚恋"}],"labels":[{"name":"现代"}]}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})
	coordinator.SetSourceClassification(batchFetchClassification{
		RewriteProvider: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代女主","confidence":0.9,"reason":"女主关系线突出"}`, nil
		}),
	})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, batch.Books[0])
	if err != nil {
		t.Fatal(err)
	}
	if updated.SourceMetadata["themes"] != "婚恋" || updated.SourceMetadata["labels"] != "现代" {
		t.Fatalf("fetched facets = %#v, want themes and labels", updated.SourceMetadata)
	}
	if updated.SourceMetadata["gender"] != "女频" || strings.TrimSpace(fmt.Sprint(updated.SourceMetadata["style"])) == "" || updated.SourceMetadata["classifyStatus"] != "classified" {
		t.Fatalf("classification = %#v, want completed classification", updated.SourceMetadata)
	}
}

func TestEnsureAutomationSourceUsesNovelFetchRulesAndKeepsRawText(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "book", ID: "source-1", Platform: "15",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	workshopStore := novelfetchworkshop.NewMemoryStore()
	if err := workshopStore.PutConfig(context.Background(), "alice", map[string]any{
		"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true, "apply_sensitive": true}},
		"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "合规词", "apply_to_original": true}},
		}},
		"sensitive_ai": map[string]any{"enabled": false},
	}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"data":"这是坏词正文","bookinfo":{"category":"女频","genre":"现代悬疑"}}`))
	}))
	defer server.Close()
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, workshopStore, nil, nil, nil)
	coordinator.SetSourceFetcher(&novelfetchworkshop.UpstreamFetcher{Client: server.Client(), Endpoint: server.URL})
	coordinator.SetSourceClassification(batchFetchClassification{
		RewriteProvider: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
		}),
	})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, batch.Books[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(updated.SourceText, "合规词") || strings.Contains(updated.SourceText, "坏词") {
		t.Fatalf("production source was not processed: %q", updated.SourceText)
	}
	if updated.TxtText != "这是坏词正文" || updated.SourceMetadata["sourceContentVersion"] != "sensitive" || updated.SourceMetadata["processingStatus"] != "done" {
		t.Fatalf("raw source or processing metadata was lost: %+v", updated)
	}
}

func TestEnsureAutomationSourceClassifiesExistingBodyBeforeProduction(t *testing.T) {
	memoryStore := batchfactoryv11.NewMemoryStore()
	batch, err := memoryStore.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "batch",
		Books: []batchfactoryv11.CreateBookInput{{
			Title:          "book",
			ID:             "source-1",
			SourceText:     "已经获取的正文",
			SourceMetadata: map[string]any{"category": "女频"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(memoryStore, nil, nil, nil, nil, nil, nil, nil, nil)
	coordinator.SetSourceClassification(batchFetchClassification{
		RewriteProvider: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代女主","confidence":0.9,"reason":"女主关系线突出"}`, nil
		}),
	})

	updated, err := coordinator.ensureAutomationSource(context.Background(), "alice", batch.ID, batch.Books[0])
	if err != nil {
		t.Fatal(err)
	}
	if updated.SourceText != "已经获取的正文" || updated.SourceMetadata["gender"] != "女频" || updated.SourceMetadata["style"] != "现代女主" || updated.SourceMetadata["classifyStatus"] != "classified" {
		t.Fatalf("existing body classification = %#v, want classified metadata without refetch", updated.SourceMetadata)
	}
}

func TestAutomationBooksMissingSourceReturnsOnlyBooksWithoutProductionLines(t *testing.T) {
	books := []batchfactoryv11.Book{
		{ID: "ready", SourceText: "第一行\n第二行"},
		{ID: "blank", SourceText: " \n\t "},
		{ID: "limited", SourceText: "\n\n", SourceMetadata: map[string]any{"contentRangeLines": 8}},
	}

	missing := automationBooksMissingSource(books)

	if len(missing) != 2 || missing[0].ID != "blank" || missing[1].ID != "limited" {
		t.Fatalf("automationBooksMissingSource() = %#v, want blank and limited", missing)
	}
}

func TestAutomationMergeAudioDurationUsesMeasuredTTSPlan(t *testing.T) {
	duration, _ := json.Marshal(25.25)
	book := batchfactoryv11.Book{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
		"audioDurationSeconds": duration,
	}}}
	production := batchfactoryv11.ProductionJob{Tasks: []batchfactoryv11.ProductionTask{
		{ActualDurationSeconds: 9},
		{ActualDurationSeconds: 10},
		{ActualDurationSeconds: 10},
		{ActualDurationSeconds: 10},
		{ActualDurationSeconds: 10},
	}}

	got, err := automationMergeAudioDuration(book, production)

	if err != nil || got != 25.25 {
		t.Fatalf("automationMergeAudioDuration() = %v, %v; want measured TTS duration 25.25", got, err)
	}
}

func TestGenerateBookAudioPlanTargetsOnlyTheRequestedBookAndNeverCreatesAMerge(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "batch",
		Books: []batchfactoryv11.CreateBookInput{
			{Title: "selected", SourceText: "第一行", SourceMetadata: map[string]any{"gender": "男频", "style": "现代", "classifyStatus": "classified"}},
			{Title: "other", SourceText: "第二行", SourceMetadata: map[string]any{"gender": "女频", "style": "现代", "classifyStatus": "classified"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	selected := batch.Books[0]
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err = coordinator.GenerateBookAudioPlan(ctx, "alice", batch.ID, selected.ID)

	if err == nil || !strings.Contains(err.Error(), "配音与逐行时长服务未配置") {
		t.Fatalf("GenerateBookAudioPlan() error = %v, want the audio-plan service error", err)
	}
	production, err := store.ListProductionJobs(ctx, "alice", batch.ID)
	if err != nil || len(production) != 0 {
		t.Fatalf("production jobs = %#v, %v; audio planning must not start video production", production, err)
	}
	merges, err := store.ListMergeJobs(ctx, "alice", batch.ID)
	if err != nil || len(merges) != 0 {
		t.Fatalf("merge jobs = %#v, %v; audio planning must not create a merge", merges, err)
	}
}

func TestReusableAutomationProductionRequiresCurrentDirectorAndCompletedMedia(t *testing.T) {
	now := time.Date(2026, 10, 17, 6, 30, 0, 0, time.UTC)
	book := batchfactoryv11.Book{ID: "book-1", DirectorRevision: &batchfactoryv11.DirectorRevision{ID: "director-current"}}
	jobs := []batchfactoryv11.ProductionJob{
		{ID: "old-director", BookID: "book-1", DirectorRevisionID: "director-old", Status: batchfactoryv11.ProductionSucceeded, UpdatedAt: now.Add(time.Minute), Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded, MediaURL: "https://media.example/old.mp4"}}},
		{ID: "incomplete", BookID: "book-1", DirectorRevisionID: "director-current", UpdatedAt: now.Add(2 * time.Minute), Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionRunning}}},
		{ID: "reusable", BookID: "book-1", DirectorRevisionID: "director-current", Status: batchfactoryv11.ProductionSucceeded, UpdatedAt: now, Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded, MediaURL: "https://media.example/current.mp4"}}},
	}

	got, ok := reusableAutomationProduction(jobs, book)

	if !ok || got.ID != "reusable" {
		t.Fatalf("reusableAutomationProduction() = %#v, %v; want reusable current-director job", got, ok)
	}
}

func TestMergeAutomationPresetConfigKeepsUnifiedConfigurationOverPresetDefaults(t *testing.T) {
	current := map[string]any{
		"videoProvider": "yfai",
		"videoModelId":  "seedance-2",
		"publishSettings": map[string]any{"websiteProfileId": "unified-profile"},
	}
	config := map[string]any{
		"videoProvider": "autodl_h3",
		"videoModelId":  "minimax-h3",
		"textModelId":   "gemini",
		"publishSettings": map[string]any{
			"websiteProfileId": "preset-profile",
			"organization":     "preset-org",
		},
	}

	merged := mergeAutomationPresetConfig(current, config)

	if merged["videoProvider"] != "yfai" || merged["videoModelId"] != "seedance-2" {
		t.Fatalf("批次统一配置应优先于预设默认，got provider=%v model=%v", merged["videoProvider"], merged["videoModelId"])
	}
	if merged["textModelId"] != "gemini" {
		t.Fatalf("批次未设置的键应由预设填充，got textModelId=%v", merged["textModelId"])
	}
	publish, _ := merged["publishSettings"].(map[string]any)
	if publish["websiteProfileId"] != "unified-profile" {
		t.Fatalf("批次统一 publishSettings 应优先，got %v", publish["websiteProfileId"])
	}
	if publish["organization"] != "preset-org" {
		t.Fatalf("批次未设置的嵌套字段应由预设填充，got %v", publish["organization"])
	}
}

func TestMergeAutomationPresetConfigUsesPresetWhenBatchSettingsAreEmpty(t *testing.T) {
	merged := mergeAutomationPresetConfig(nil, map[string]any{"videoProvider": "autodl_h3", "videoModelId": "minimax-h3"})
	if merged["videoProvider"] != "autodl_h3" || merged["videoModelId"] != "minimax-h3" {
		t.Fatalf("批次没有统一配置时应直接使用预设，got %#v", merged)
	}
}

// 线上事故：重试重新打开了同一条终端 job，视频阶段沿用 jobID+"-video-"+bookID
// 这一重复 requestID，SubmitBookProductionWithOptions 按幂等直接返回旧的失败任务
// （连同旧 provider 与旧错误），从未真正重交。重试必须使用不同的 requestID。
func TestAutomationRetryScopesTheVideoRequestID(t *testing.T) {
	const jobID, bookID = "job_0a9b", "book_411df5a5"
	first := automationVideoRequestID(jobID, bookID, 0, 0)
	if first != jobID+"-video-"+bookID {
		t.Fatalf("首次运行 requestID=%q", first)
	}
	// In-run batch backfill round (batchRetryCount=1) must not replay round 1.
	backfill := automationVideoRequestID(jobID, bookID, 0, 1)
	if backfill != jobID+"-video-"+bookID+"-r0b1" {
		t.Fatalf("回补轮 requestID=%q", backfill)
	}
	// Manual retry after a terminal run reopens the same job id.
	retried := automationVideoRequestID(jobID, bookID, 2, 0)
	if retried != jobID+"-video-"+bookID+"-r2b0" {
		t.Fatalf("重试 requestID=%q", retried)
	}
	for _, other := range []string{backfill, retried} {
		if other == first {
			t.Fatal("重试/回补 requestID 必须与首次运行不同，否则会命中旧失败任务")
		}
	}
}
