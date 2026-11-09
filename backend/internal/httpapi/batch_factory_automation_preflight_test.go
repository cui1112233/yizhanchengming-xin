package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type preflightTextProvider struct{}

func (preflightTextProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	return "{}", nil
}

func preflightIssueCodes(issues []automationPreflightIssue) map[string]string {
	codes := make(map[string]string, len(issues))
	for _, issue := range issues {
		codes[issue.Code] = issue.Message
	}
	return codes
}

func preflightCoordinator(t *testing.T, store batchfactoryv11.Store, stages *batchfactoryv11.BookStageService) *BatchAutomationCoordinator {
	t.Helper()
	coordinator := NewBatchAutomationCoordinator(store, nil, stages, nil, nil, nil, nil, nil, func() time.Time {
		return time.Unix(1700000000, 0)
	})
	coordinator.SetComicPublisher(func(context.Context, string, string, string, batchfactoryv11.ComicCollection, batchfactoryv11.SettingsPatch) (N8ComicPublishRecord, error) {
		return N8ComicPublishRecord{}, nil
	})
	return coordinator
}

func preflightComicBatch(t *testing.T, store batchfactoryv11.Store, bookCount int, videosPerBook int) batchfactoryv11.Batch {
	t.Helper()
	ctx := context.Background()
	books := make([]batchfactoryv11.CreateBookInput, 0, bookCount)
	for index := 0; index < bookCount; index++ {
		videos := make([]batchfactoryv11.CreateVideoInput, 0, videosPerBook)
		for page := 0; page < videosPerBook; page++ {
			videos = append(videos, batchfactoryv11.CreateVideoInput{Label: "第" + string(rune('一'+page)) + "页"})
		}
		books = append(books, batchfactoryv11.CreateBookInput{
			Title:      "漫画书" + string(rune('A'+index)),
			SourceText: "林晚走进办公室，陆沉抬头看向她。",
			Videos:     videos,
		})
	}
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: books})
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func savePreflightBatchSettings(t *testing.T, store batchfactoryv11.Store, batch batchfactoryv11.Batch, patch batchfactoryv11.SettingsPatch) batchfactoryv11.Batch {
	t.Helper()
	updated, err := store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: patch, ExpectedRevision: batch.Revision})
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision != reloaded.Revision {
		t.Fatalf("settings revision=%d, batch revision=%d", updated.Revision, reloaded.Revision)
	}
	return reloaded
}

func preflightComicSettingsPatch(t *testing.T, comicBody string, assetsBody string) batchfactoryv11.SettingsPatch {
	t.Helper()
	comic := map[string]any{"enabled": true}
	if comicBody != "" {
		comic["presetId"] = "batch-comic-page"
		comic["body"] = comicBody
	}
	if assetsBody != "" {
		comic["assets"] = map[string]any{"presetId": "batch-comic-assets", "body": assetsBody}
	}
	raw, err := json.Marshal(map[string]any{"comic": comic})
	if err != nil {
		t.Fatal(err)
	}
	publish, _ := json.Marshal(map[string]any{"organizationId": "1", "organizationName": "博量", "producerAdminName": "制作者"})
	mode, _ := json.Marshal("comic")
	return batchfactoryv11.SettingsPatch{
		"productionMode":       mode,
		"textModelId":          json.RawMessage(`"text-1"`),
		"imageModelId":         json.RawMessage(`"image-1"`),
		"assetImagesEnabled":   json.RawMessage("true"),
		"comicPublishSettings": publish,
		"aiPromptConfig":       raw,
	}
}

func preflightReadyStages(store batchfactoryv11.Store) *batchfactoryv11.BookStageService {
	return &batchfactoryv11.BookStageService{
		Store:      store,
		Director:   &batchfactoryv11.DirectorService{Store: store, Provider: preflightTextProvider{}},
		ImageReady: func(context.Context, string) (bool, string) { return true, "" },
		TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return preflightTextProvider{}, nil
		},
	}
}

func TestAutomationPreflightPassesComicBatchAndWritesNothing(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产"))
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))

	result, err := coordinator.Preflight(ctx, "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || !result.ReadOnly {
		t.Fatalf("result ok=%v readOnly=%v issues=%#v", result.OK, result.ReadOnly, result.BatchIssues)
	}
	if result.Summary.Total != 1 || result.Summary.Passed != 1 || result.Summary.Blocked != 0 {
		t.Fatalf("summary=%#v", result.Summary)
	}
	if len(result.RunnableBookIDs) != 1 || result.RunnableBookIDs[0] != batch.Books[0].ID {
		t.Fatalf("runnable=%v", result.RunnableBookIDs)
	}
	// 纯只读：批次修订号与设置必须保持原样，且不能创建 automation 状态文件。
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Revision != batch.Revision {
		t.Fatalf("revision=%d, want %d: preflight must not write settings", latest.Revision, batch.Revision)
	}
	if string(latest.SettingsState.Patch["comicPublishSettings"]) != string(batch.SettingsState.Patch["comicPublishSettings"]) {
		t.Fatal("preflight changed comic publish settings")
	}
	if coordinator.automationState != nil || coordinator.path != "" {
		t.Fatal("preflight must not create an automation job state")
	}
	if jobs, _ := store.ListProductionJobs(ctx, "alice", batch.ID); len(jobs) != 0 {
		t.Fatalf("preflight created production jobs: %#v", jobs)
	}
	if result.TaskState != "" || result.RunningJobID != "" || result.RunningSince != "" {
		t.Fatalf("preflight reported an active task without one: %#v", result)
	}
}

func TestAutomationPreflightPeekActiveJobWritesNothing(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产"))
	stateStore := &memoryAutomationStateStore{found: true, state: automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-active": {
			"id": "job-active", "owner": "alice", "batchId": batch.ID, "state": "running",
			"startedAt": "2026-11-05T09:58:00Z", "updatedAt": "2026-11-05T10:00:00Z",
		},
		// 别的账号 / 别的批次 / 已完成任务都不应被当成“进行中”。
		"job-other-owner": {"id": "job-other-owner", "owner": "bob", "batchId": batch.ID, "state": "running"},
		"job-other-batch": {"id": "job-other-batch", "owner": "alice", "batchId": "batch-x", "state": "running"},
		"job-completed":   {"id": "job-completed", "owner": "alice", "batchId": batch.ID, "state": "completed"},
	}}}
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))
	coordinator.SetAutomationStateStore(stateStore)
	before, err := json.Marshal(stateStore.state)
	if err != nil {
		t.Fatal(err)
	}
	loadsBefore := stateStore.loads

	result, err := coordinator.Preflight(ctx, "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	if result.TaskState != "running" || result.RunningJobID != "job-active" {
		t.Fatalf("taskState=%q runningJobId=%q, want the active job", result.TaskState, result.RunningJobID)
	}
	if result.RunningSince != "2026-11-05T09:58:00Z" || result.TaskPaused {
		t.Fatalf("runningSince=%q paused=%v, want startedAt and not paused", result.RunningSince, result.TaskPaused)
	}
	if result.OK {
		t.Fatal("an active job must block preflight ok")
	}
	if msg := preflightIssueCodes(result.BatchIssues)["automation_job_active"]; !strings.Contains(msg, "当前批次已有自动生产任务在进行（运行中，开始于 2026-11-05T09:58:00Z）") {
		t.Fatalf("batch issues=%#v, want the active job notice", result.BatchIssues)
	}
	// 纯只读：peek 只允许一次 Load，绝不能 Save，内存状态快照也不能被改。
	if stateStore.saves != 0 {
		t.Fatalf("peek wrote automation state: saves=%d", stateStore.saves)
	}
	if stateStore.loads != loadsBefore+1 {
		t.Fatalf("peek loads=%d, want exactly one read-only load", stateStore.loads-loadsBefore)
	}
	after, err := json.Marshal(stateStore.state)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("peek mutated job state: before=%s after=%s", before, after)
	}
	if coordinator.automationState != nil {
		t.Fatal("peek must not populate the coordinator's cached job state")
	}
}

func TestAutomationPreflightBlocksMissingComicPublishProfile(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	patch := preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产")
	delete(patch, "comicPublishSettings")
	batch = savePreflightBatchSettings(t, store, batch, patch)
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))

	result, err := coordinator.Preflight(context.Background(), "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Summary.Blocked != 1 || len(result.RunnableBookIDs) != 0 {
		t.Fatalf("summary=%#v ok=%v runnable=%v", result.Summary, result.OK, result.RunnableBookIDs)
	}
	blocking := preflightIssueCodes(result.Books[0].Blocking)
	if !strings.Contains(blocking["comic_publish_profile_missing"], "漫画发布设置未绑定网站或组织") {
		t.Fatalf("blocking=%#v, want comic publish profile reason", result.Books[0].Blocking)
	}
}

func TestAutomationPreflightBlocksInvalidComicAudioDirectory(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产"))
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))
	called := false
	coordinator.SetComicPublishPreflight(func(_ context.Context, owner, batchID, bookID string, _ batchfactoryv11.SettingsPatch) error {
		called = true
		if owner != "alice" || batchID != batch.ID || bookID != batch.Books[0].ID {
			t.Fatalf("unexpected scope owner=%q batch=%q book=%q", owner, batchID, bookID)
		}
		return fmt.Errorf("音频分类“古风-爽文（打脸）”不存在、未启用或名称不唯一")
	})

	result, err := coordinator.Preflight(context.Background(), "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	if !called || result.OK || result.Summary.Blocked != 1 {
		t.Fatalf("called=%v ok=%v summary=%#v", called, result.OK, result.Summary)
	}
	blocking := preflightIssueCodes(result.Books[0].Blocking)
	if !strings.Contains(blocking["comic_publish_directory_invalid"], "古风-爽文（打脸）") {
		t.Fatalf("blocking=%#v", result.Books[0].Blocking)
	}
}

func TestAutomationPreflightReportsSingleBookOverride(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产"))
	book := batch.Books[0]
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batch.ID, BookID: book.ID}, batchfactoryv11.SettingsUpdate{
		Patch:            batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"organizationId":"","producerAdminName":""}`)},
		ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))

	result, err := coordinator.Preflight(ctx, "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary.Blocked != 1 {
		t.Fatalf("summary=%#v, want the book blocked by its own override", result.Summary)
	}
	entry := result.Books[0]
	if len(entry.Overrides) != 1 || entry.Overrides[0].Key != "comicPublishSettings" {
		t.Fatalf("overrides=%#v, want the single-book publish override listed", entry.Overrides)
	}
	if !strings.Contains(preflightIssueCodes(entry.Blocking)["comic_publish_profile_missing"], "漫画发布设置未绑定网站或组织") {
		t.Fatalf("blocking=%#v", entry.Blocking)
	}
}

func TestAutomationPreflightBlocksUnavailableTextModel(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产"))
	stages := preflightReadyStages(store)
	stages.TextProviderResolver = func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		return nil, nil
	}
	coordinator := preflightCoordinator(t, store, stages)

	result, err := coordinator.Preflight(context.Background(), "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	blocking := preflightIssueCodes(result.Books[0].Blocking)
	if !strings.Contains(blocking["text_model_unavailable"], "选择的文本模型不可用") {
		t.Fatalf("blocking=%#v, want unavailable text model reason", result.Books[0].Blocking)
	}
	if result.Books[0].Passed {
		t.Fatal("book with an unavailable text model must not pass preflight")
	}
}

func TestAutomationPreflightBlocksMissingComicPromptPresets(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "", ""))
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))

	result, err := coordinator.Preflight(context.Background(), "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	blocking := preflightIssueCodes(result.Books[0].Blocking)
	if !strings.Contains(blocking["comic_page_preset_missing"], "未选择漫画页提示词预设") {
		t.Fatalf("blocking=%#v, want missing comic page preset", result.Books[0].Blocking)
	}
	if !strings.Contains(blocking["comic_assets_preset_missing"], "请在资产设置中选择漫画资产提示词") {
		t.Fatalf("blocking=%#v, want missing comic assets preset", result.Books[0].Blocking)
	}
}

func TestAutomationPreflightBlocksUnfrozenComicPromptBody(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	patch := preflightComicSettingsPatch(t, "", "提取漫画资产")
	patch["aiPromptConfig"] = json.RawMessage(`{"comic":{"enabled":true,"presetId":"batch-comic-page","assets":{"presetId":"batch-comic-assets","body":"提取漫画资产"}}}`)
	batch = savePreflightBatchSettings(t, store, batch, patch)
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))

	result, err := coordinator.Preflight(context.Background(), "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := preflightIssueCodes(result.Books[0].Blocking)["comic_prompt_body_not_frozen"]; !ok {
		t.Fatalf("blocking=%#v, want unfrozen comic prompt body", result.Books[0].Blocking)
	}
}

func TestAutomationPreflightWarnsPageCountMismatch(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 3)
	patch := preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产")
	patch["aiPromptConfig"] = json.RawMessage(`{"comic":{"enabled":true,"body":"整本漫画页规则","targetPageCount":5,"assets":{"presetId":"batch-comic-assets","body":"提取漫画资产"}}}`)
	batch = savePreflightBatchSettings(t, store, batch, patch)
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))

	result, err := coordinator.Preflight(context.Background(), "alice", batch.ID, map[string]any{"runMode": "full_submit"})
	if err != nil {
		t.Fatal(err)
	}
	warning := preflightIssueCodes(result.Books[0].Warnings)
	if !strings.Contains(warning["comic_page_count_mismatch"], "执行后可能报「漫画页数量与分页不一致」") {
		t.Fatalf("warnings=%#v, want page count mismatch risk", result.Books[0].Warnings)
	}
	if !result.Books[0].Passed || result.Summary.WarningBooks != 1 {
		t.Fatalf("page count risk must warn without blocking: %#v", result.Summary)
	}
}

func TestV12AutomationPreflightRouteIsReadOnly(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 1, 0)
	batch = savePreflightBatchSettings(t, store, batch, preflightComicSettingsPatch(t, "整本漫画页规则", "提取漫画资产"))
	coordinator := preflightCoordinator(t, store, preflightReadyStages(store))
	api := NewRouter(RouterOptions{
		BridgeSecret:          "secret",
		Now:                   func() time.Time { return now },
		Slice:                 2,
		Store:                 store,
		AutomationCoordinator: coordinator,
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/automation/preflight", map[string]any{"runMode": "full_submit"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"readOnly":true`) || !strings.Contains(body, `"passed":1`) {
		t.Fatalf("body=%s", body)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Revision != batch.Revision {
		t.Fatalf("preflight route wrote settings: revision=%d want %d", latest.Revision, batch.Revision)
	}
}
