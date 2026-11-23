package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestBatchFactorySensitiveStageTextPreflight(t *testing.T) {
	c := &BatchAutomationCoordinator{}
	batch := batchfactoryv11.Batch{SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"text"`)}}}
	book := batchfactoryv11.Book{TxtText: "原文", SourceText: "正文", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}
	if issues := c.automationGlobalPreflightIssues(context.Background(), "alice", batch, "video_no_submit"); len(issues) != 0 {
		t.Errorf("纯文字的本地排版不依赖媒体或模型执行器: %+v", issues)
	}
	if issues := c.automationBookPreflightIssues(context.Background(), "alice", batch, book, "video_no_submit", false); len(issues) != 0 {
		t.Errorf("本地规则不需要选择模型: %+v", issues)
	}
	batch.SettingsState.Patch["sensitiveMode"] = json.RawMessage(`"ai"`)
	if issues := c.automationBookPreflightIssues(context.Background(), "alice", batch, book, "video_no_submit", false); len(issues) != 1 || issues[0].Code != "text_model_missing" {
		t.Errorf("AI 需要文本模型: %+v", issues)
	}
	book.SourceMetadata = nil
	book.SettingsState.Patch = batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"off"`)}
	if issues := c.automationBookPreflightIssues(context.Background(), "alice", batch, book, "video_no_submit", false); len(issues) != 1 || !strings.Contains(issues[0].Message, "缺少完整原文") {
		t.Errorf("旧生产稿须先补全原文: %+v", issues)
	}
}

type sensitiveTestProvider func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error)

type failingRuleSettings struct{ novelfetchworkshop.Store }

func (failingRuleSettings) GetConfig(context.Context, string) (map[string]any, error) {
	return nil, errors.New("读取词库失败")
}

func TestBatchFactorySensitiveStagePersistsRulesFailure(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	b, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "失败时保留稿件", Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "前次生产稿", TxtText: "完整原文", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
	if err != nil {
		t.Fatal(err)
	}
	c := &BatchAutomationCoordinator{store: store, ttsStore: failingRuleSettings{}}
	if _, err := c.runBatchSensitiveStage(ctx, "alice", b.ID, b.Books[0].ID, false); err == nil {
		t.Fatal("expected rules error")
	}
	b, err = store.GetBatch(ctx, "alice", b.ID)
	if err != nil {
		t.Fatal(err)
	}
	book := b.Books[0]
	state := mapOrEmpty(book.SourceMetadata["batchSensitiveStage"])
	if state["status"] != "failed" || !strings.Contains(firstMetadataString(state, "error"), "读取词库失败") || book.SourceText != "前次生产稿" {
		t.Fatalf("缺少失败记录: %+v", book)
	}
}

func TestBatchFactorySensitiveStageProtectsManualEditBeforeFirstRun(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	b, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "用户编辑", InitialBatchSettings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"ai"`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "人工改过的段落", TxtText: "完整原文", SourceMetadata: map[string]any{"rawSourceStatus": "verified", "productionTextEdited": true}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := b.Books[0]
	_, err = store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: b.ID, BookID: book.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"off"`)}, ExpectedRevision: book.Revision})
	if err != nil {
		t.Fatal(err)
	}
	c := &BatchAutomationCoordinator{store: store}
	updated, err := c.runBatchSensitiveStage(ctx, "alice", b.ID, book.ID, false)
	if err != nil || updated.SourceText != "人工改过的段落" || updated.TxtText != "完整原文" {
		t.Fatalf("单书关闭后没有保留手改生产稿: book=%+v err=%v", updated, err)
	}
}

func TestBatchFactorySensitiveStageConcurrentStartAndMetadataEdit(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	b, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "并发", InitialBatchSettings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"ai"`), "textModelId": json.RawMessage(`"m"`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "生产稿", TxtText: "完整原文", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	bookID := b.Books[0].ID
	var c *BatchAutomationCoordinator
	c = &BatchAutomationCoordinator{store: store, stages: &batchfactoryv11.BookStageService{TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		return sensitiveTestProvider(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
			calls++
			if calls > 1 {
				t.Fatal("重复调用模型")
			}
			if _, err := c.runBatchSensitiveStage(ctx, "alice", b.ID, bookID, false); err == nil {
				t.Fatal("并发请求未识别正在执行的工作")
			}
			fresh, err := store.GetBatch(ctx, "alice", b.ID)
			if err != nil {
				t.Fatal(err)
			}
			br := fresh.Books[0]
			if _, err := store.UpdateBookMetadata(ctx, "alice", b.ID, bookID, batchfactoryv11.UpdateBookMetadataInput{Title: "新书名", Metadata: br.SourceMetadata, ExpectedRevision: br.Revision}); err != nil {
				t.Fatal(err)
			}
			return `{"changes":[{"before":"原文","after":"内容"}]}`, nil
		}), nil
	}}}
	updated, err := c.runBatchSensitiveStage(ctx, "alice", b.ID, bookID, false)
	if err != nil || calls != 1 || updated.Title != "新书名" || updated.SourceText != "完整内容" {
		t.Fatalf("同时改书名时应复用模型结果: %+v err=%v calls=%d", updated, err, calls)
	}
}

func (f sensitiveTestProvider) Complete(ctx context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
	return f(ctx, r)
}

func TestBatchFactorySensitiveStageDispatchAndResume(t *testing.T) {
	ctx := context.Background()
	for _, production := range []string{"comic", "video", "text"} {
		for _, mode := range []string{"ai", "off", "system"} {
			t.Run(production+"_"+mode, func(t *testing.T) {
				s := batchfactoryv11.NewMemoryStore()
				raw := " 坏词在这里。\n"
				b, err := s.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "测试", InitialBatchSettings: batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"` + production + `"`), "sensitiveMode": json.RawMessage(`"` + mode + `"`), "textModelId": json.RawMessage(`"selected"`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "书", TxtText: raw, SourceText: "旧稿", SourceMetadata: map[string]any{"rawSourceStatus": "verified", "rawSourceOrigin": "bookstore_raw"}}}})
				if err != nil {
					t.Fatal(err)
				}
				workshop := novelfetchworkshop.NewMemoryStore()
				if err := workshop.PutConfig(ctx, "alice", map[string]any{"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "规则词", "apply_to_original": true}}}}}); err != nil {
					t.Fatal(err)
				}
				calls := 0
				unavailable := true
				c := &BatchAutomationCoordinator{store: s, ttsStore: workshop, stages: &batchfactoryv11.BookStageService{TextProviderResolver: func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
					if owner != "alice" || model != "selected" {
						t.Fatalf("模型选择不正确: %s %s", owner, model)
					}
					return sensitiveTestProvider(func(_ context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
						calls++
						if !strings.Contains(r.UserPrompt, "坏词在这里") {
							t.Fatalf("模型没有收到原文: %v", r.UserPrompt)
						}
						if unavailable {
							return "", errors.New("模型临时不可用")
						}
						return `{"changes":[{"before":"坏词","after":"修复词"}]}`, nil
					}), nil
				}}}
				run := func(manual bool) (batchfactoryv11.Book, error) {
					return c.runBatchSensitiveStage(ctx, "alice", b.ID, b.Books[0].ID, manual)
				}
				if mode == "ai" {
					for i := 0; i < 2; i++ {
						if _, err := run(false); err == nil || !strings.Contains(err.Error(), "模型临时不可用") {
							t.Fatalf("失败未保留: %v", err)
						}
					}
					persisted, _ := s.GetBatch(ctx, "alice", b.ID)
					if calls != 1 || persisted.Books[0].SourceText != "旧稿" || persisted.Books[0].TxtText != raw {
						t.Fatalf("失败后原文、生产稿或调用数异常, calls=%d book=%+v", calls, persisted.Books[0])
					}
				}
				unavailable = false
				got, err := run(true)
				if err != nil {
					t.Fatal(err)
				}
				want := map[string]string{"ai": " 修复词在这里。\n", "system": " 规则词在这里。\n", "off": raw}[mode]
				if got.SourceText != want || got.TxtText != raw {
					t.Fatalf("got production=%q original=%q, want %q %q", got.SourceText, got.TxtText, want, raw)
				}
				if _, err := run(false); err != nil {
					t.Fatal(err)
				}
				if calls != map[string]int{"ai": 2}[mode] {
					t.Fatalf("模型重复执行: %d", calls)
				}
			})
		}
	}
}

func TestSensitiveStageUsesSavedDraftForSegmentation(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	raw := "这是原始文本。"
	b, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画", Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "旧稿", TxtText: raw, SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
	if err != nil {
		t.Fatal(err)
	}
	bookID := b.Books[0].ID
	draft := "编辑的坏词。"
	rules := novelfetchworkshop.NewMemoryStore()
	if err = rules.PutConfig(ctx, "alice", map[string]any{"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "好词", "apply_to_original": true}}}}}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	director := &batchfactoryv11.DirectorService{Store: store, Provider: sensitiveTestProvider(func(_ context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
		calls++
		if !strings.Contains(r.UserPrompt, "编辑的好词。") || strings.Contains(r.UserPrompt, draft) {
			t.Fatalf("model received %q", r.UserPrompt)
		}
		return "{}", nil
	})}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, NovelFetchStore: rules, Director: director})
	rec := signedJSONRequest(t, api, now, "alice", "PUT", "/api/batch-factory/v12/drafts", batchfactoryv11.Draft{Key: "working-front:" + bookID, Kind: "working-front-content", Scope: b.ID, Content: draft})
	if rec.Code != 200 {
		t.Fatalf("save draft returned %d %s", rec.Code, rec.Body.String())
	}
	rec = signedJSONRequest(t, api, now, "alice", "POST", "/api/batch-factory/v12/batches/"+b.ID+"/books/"+bookID+"/comic-source-segments/segment", map[string]any{})
	if rec.Code != 200 || calls != 1 {
		t.Fatalf("segment calls=%d status=%d %s", calls, rec.Code, rec.Body.String())
	}
	latest, err := store.GetBatch(ctx, "alice", b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Books[0].SourceText != "编辑的好词。" || latest.Books[0].TxtText != raw {
		t.Fatalf("unexpected stored texts %+v", latest.Books[0])
	}
	saved, err := store.GetDraft(ctx, "alice", "working-front:"+bookID, "working-front-content", b.ID)
	if err != nil || saved.Content != draft {
		t.Fatalf("authored draft changed: %+v %v", saved, err)
	}
}

func TestSensitiveStageEmptyResultPreservesProductionAndCanRetry(t *testing.T) {
	for _, mode := range []string{"system", "ai"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			store := batchfactoryv11.NewMemoryStore()
			workshop := novelfetchworkshop.NewMemoryStore()
			b, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "书", InitialBatchSettings: batchfactoryv11.SettingsPatch{"sensitiveMode": comicTestRaw(t, mode), "textModelId": comicTestRaw(t, "选定模型")}, Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "前次生产稿", TxtText: "全文", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
			if err != nil {
				t.Fatal(err)
			}
			setRule := func(replace string) {
				t.Helper()
				if err := workshop.PutConfig(ctx, "alice", map[string]any{"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"enabled": true, "find": "全文", "replace": replace, "replace_explicit": true, "apply_to_original": true}}}}}); err != nil {
					t.Fatal(err)
				}
			}
			setRule("")
			reply := ""
			c := &BatchAutomationCoordinator{store: store, ttsStore: workshop, stages: &batchfactoryv11.BookStageService{TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
				return sensitiveTestProvider(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) { return reply, nil }), nil
			}}}
			if _, err := c.runBatchSensitiveStage(ctx, "alice", b.ID, b.Books[0].ID, false); err == nil {
				t.Fatal("空结果误报成功")
			}
			latest, err := store.GetBatch(ctx, "alice", b.ID)
			if err != nil {
				t.Fatal(err)
			}
			book := latest.Books[0]
			stage := mapOrEmpty(book.SourceMetadata["batchSensitiveStage"])
			if stage["status"] != "failed" || firstMetadataString(stage, "error") == "" || book.SourceText != "前次生产稿" || book.TxtText != "全文" {
				t.Fatalf("失败后状态异常 %+v", book)
			}
			setRule("规则替换")
			reply = `{"changes":[]}`
			updated, err := c.runBatchSensitiveStage(ctx, "alice", b.ID, book.ID, true)
			if err != nil || mapOrEmpty(updated.SourceMetadata["batchSensitiveStage"])["status"] != "done" {
				t.Fatalf("不能立即重试 %v %+v", err, updated)
			}
		})
	}
}

type frozenCollectionStore struct {
	*batchfactoryv11.MemoryStore
	collection batchfactoryv11.ComicCollection
}

func (s frozenCollectionStore) ListComicCollections(context.Context, string, string, string) ([]batchfactoryv11.ComicCollection, error) {
	return []batchfactoryv11.ComicCollection{s.collection}, nil
}

func TestSensitiveStageHistoricalComicUploadUsesFrozenCollection(t *testing.T) {
	ctx := context.Background()
	store := frozenCollectionStore{batchfactoryv11.NewMemoryStore(), batchfactoryv11.ComicCollection{ID: "fixed"}}
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "旧合集", InitialBatchSettings: batchfactoryv11.SettingsPatch{"productionMode": comicTestRaw(t, "comic"), "sensitiveMode": comicTestRaw(t, "ai"), "comicPublishSettings": json.RawMessage(`{"producerAdminName":"新设置"}`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "旧书", SourceText: "历史片段"}, {Title: "仍需制作", TxtText: "新书正文。", SourceText: "新书正文。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	modelCalls := 0
	c := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{Store: store, TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		modelCalls++
		return nil, errors.New("unexpected model call")
	}}, nil, nil, nil, nil, nil, time.Now)
	uploadCalls := 0
	c.SetComicPublisher(func(_ context.Context, owner, batchID, bookID string, col batchfactoryv11.ComicCollection, settings batchfactoryv11.SettingsPatch) (N8ComicPublishRecord, error) {
		uploadCalls++
		if col.ID != "fixed" || string(settings["comicPublishSettings"]) != `{"producerAdminName":"原设置"}` {
			t.Fatalf("unexpected frozen input %s %v", col.ID, settings)
		}
		return N8ComicPublishRecord{Status: "published", CarouselID: 17}, nil
	})
	entry := map[string]any{"status": "running", "stage": "comic_publish", "comicCollectionId": "fixed", "frozenComicPublishSettings": `{"producerAdminName":"原设置"}`}
	pending := map[string]any{"status": "queued", "stage": "queued"}
	job := map[string]any{"id": "old", "owner": "alice", "batchId": batch.ID, "state": "running", "runMode": "full_submit", "books": map[string]any{book.ID: entry, batch.Books[1].ID: pending}}
	if err := c.validateResumedAutomation(ctx, "alice", batch, job); err != nil {
		t.Fatal(err)
	}
	c.applyResumePreflight(ctx, "alice", batch, job, time.Now())
	c.run(ctx, 0, "alice", batch.ID, "old", job, batch)
	if uploadCalls != 1 || modelCalls != 0 || entry["stage"] != "published" {
		t.Fatalf("uploads=%d model=%d entry=%v", uploadCalls, modelCalls, entry)
	}
	if pending["status"] != "blocked" || !strings.Contains(automationStringValue(pending["message"], ""), "模型") {
		t.Fatalf("second book missing readiness result: %v", pending)
	}
	if _, err := c.latestComicCollection(ctx, "alice", batch.ID, book.ID, "deleted"); err == nil {
		t.Fatal("missing collection fell back to a different artifact")
	}
}

func prepareComicPageSourceForGeneration(t *testing.T, store *batchfactoryv11.MemoryStore, catalog novelfetchworkshop.Store, batchID, bookID, videoID string) {
	t.Helper()
	ctx := context.Background()
	book, err := (&BatchAutomationCoordinator{store: store, ttsStore: catalog}).runBatchSensitiveStage(ctx, "alice", batchID, bookID, false)
	if err != nil {
		t.Fatal(err)
	}
	source := book.SourceText
	drafts, err := batchfactoryv11.FallbackComicSourceSegments(source, batchfactoryv11.EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.BindComicSourceSegments(ctx, "alice", batchID, bookID, comicSourceDigest(source), []string{videoID}, drafts); err != nil {
		t.Fatal(err)
	}
}

func TestProductionRejectsDraftChangedAfterPreflight(t *testing.T) {
	for _, action := range []string{"director", "segmentation"} {
		t.Run(action, func(t *testing.T) {
			ctx := context.Background()
			store := batchfactoryv11.NewMemoryStore()
			batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "竞争测试", Books: []batchfactoryv11.CreateBookInput{{Title: "书", TxtText: "完整原稿。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
			if err != nil {
				t.Fatal(err)
			}
			id := batch.Books[0].ID
			coordinator := &BatchAutomationCoordinator{store: store}
			prepared, err := coordinator.runBatchSensitiveStage(ctx, "alice", batch.ID, id, false)
			if err != nil {
				t.Fatal(err)
			}
			pinned := batchfactoryv11.WithPreparedProduction(ctx, prepared)
			if _, err = store.SaveDraft(ctx, "alice", batchfactoryv11.Draft{Kind: "working-front-content", Key: "working-front:" + id, Scope: batch.ID, Content: "此刻刚保存的未处理稿。"}); err != nil {
				t.Fatal(err)
			}
			calls := 0
			director := &batchfactoryv11.DirectorService{Store: store, Provider: sensitiveTestProvider(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
				calls++
				return "{}", nil
			})}
			verifyConflict := func(t *testing.T) {
				t.Helper()
				if action == "director" {
					_, err = director.RunDirector(pinned, "alice", batch.ID, id)
				} else {
					_, err = director.RunComicSourceSegmentation(pinned, "alice", batch.ID, id)
				}
				if !errors.Is(err, batchfactoryv11.ErrConflict) || calls != 0 {
					t.Fatalf("changed draft: err=%v providerCalls=%d", err, calls)
				}
			}
			verifyConflict(t)
			if _, err = coordinator.runBatchSensitiveStage(ctx, "alice", batch.ID, id, true); err != nil {
				t.Fatal(err)
			}
			verifyConflict(t)
		})
	}
}
