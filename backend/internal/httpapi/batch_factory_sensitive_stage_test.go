package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

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
