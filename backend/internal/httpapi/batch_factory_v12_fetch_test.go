package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestV12FetchOriginalsUsesGoNovelFetcherAndKeepsRawText(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("bookid") != "book-1" || r.URL.Query().Get("platform") != "15" || r.URL.Query().Has("max_txt") {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"完整正文超过限制","bookinfo":{"category":"女频","genre":"现代"}}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now },
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals", map[string]any{
		"platform": 15, "bookIds": []string{"book-1"}, "maxTxt": 100,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"\"status\":\"ok\"", "完整正文超过限制", "\"gender\":\"女频\"", "\"rawData\""} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestV12FetchOriginalsAppliesNovelFetchRulesAndKeepsRawSource(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"第1章\n坏词。。","bookinfo":{"category":"女频","genre":"现代悬疑"}}`))
	}))
	defer upstream.Close()

	now := time.Unix(1700000000, 0)
	workshopStore := novelfetchworkshop.NewMemoryStore()
	if err := workshopStore.PutConfig(context.Background(), "alice", map[string]any{
		"appConfig": map[string]any{"layout": map[string]any{
			"apply_to_original": true, "apply_sensitive": true, "apply_chapter_cleanup": true,
		}},
		"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "合规词", "apply_to_original": true}},
		}},
		"sensitive_ai": map[string]any{"enabled": false},
	}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchStore: workshopStore,
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals", map[string]any{
		"platform": 15, "bookIds": []string{"book-1"}, "maxTxt": 100,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`"rawData":"第1章\n坏词。。"`,
		`"data":"合规词`,
		`"sourceContentVersion":"sensitive"`,
		`"sensitiveReplaced":1`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestV12FetchOriginalsClassifiesFetchedMetadata(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"一段现代悬疑正文","bookinfo":{}}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now },
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
		NovelFetchAIGenderJudge: func(context.Context, novelfetchworkshop.BookInfo) (string, error) {
			return `{"gender":"女频","confidence":0.9,"reason":"正文情感关系线突出"}`, nil
		},
		NovelFetchRewrite: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
		}),
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals", map[string]any{
		"platform": 15, "bookIds": []string{"book-1"}, "maxTxt": 100,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, want := range []string{`"gender":"女频"`, `"style":"现代悬疑"`, `"classifyStatus":"classified"`, `"originalRawStatus":"done"`, `"upstreamCode":200`, `"upstreamMessage":"成功"`, `"queueStage":"classification"`} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("response missing %q: %s", want, rec.Body.String())
		}
	}
}

func TestFinalizeFetchedBatchClassificationKeepsFetchedTextUsable(t *testing.T) {
	failed := finalizeFetchedBatchClassification(map[string]any{
		"originalStatus":    "done",
		"originalRawStatus": "done",
		"originalRawChars":  3565,
		"classifyStatus":    "processing",
	})
	if failed["classifyStatus"] != "classified" || failed["classifyNeedsReview"] != true {
		t.Fatalf("failed=%#v, want usable default classification", failed)
	}
	if failed["gender"] != "女频" || failed["style"] != "现代通用" {
		t.Fatalf("failed=%#v, want default labels", failed)
	}
	if failed["originalStatus"] != "done" || failed["originalRawChars"] != 3565 {
		t.Fatalf("failed=%#v, want fetched source state preserved", failed)
	}

	completed := finalizeFetchedBatchClassification(map[string]any{
		"gender": "女频",
		"style":  "现代甜文",
	})
	if completed["classifyStatus"] != "classified" || completed["classifyRetryable"] != false {
		t.Fatalf("completed=%#v, want classified state", completed)
	}
}

func TestFinalizeFetchedBatchClassificationAcceptsValidLabelsWithoutConfidence(t *testing.T) {
	metadata := finalizeFetchedBatchClassification(map[string]any{
		"gender": "女频", "genderSource": "ai", "genderConfidence": 0.9,
		"style": "现代悬疑", "styleSource": "ai", "styleConfidence": 0,
	})
	if metadata["classifyStatus"] != "classified" || metadata["classifyRetryable"] != false {
		t.Fatalf("metadata=%#v, want valid labels accepted", metadata)
	}
}

func TestV12FetchOriginalsRejectsInvalidInput(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{}})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals", map[string]any{"platform": 15, "bookIds": []string{}, "maxTxt": 100})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12FetchOriginalsAcceptsQueryFallbackWhenPostBodyIsMissing(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"正文","bookinfo":{}}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now },
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals?platform=15&bookId=book-1&maxTxt=100", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestFetchBatchSourceWithRetryRetriesTransientUpstreamFailure(t *testing.T) {
	attempts := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"正文","bookinfo":{}}`))
	}))
	defer upstream.Close()
	result, err, attempts := fetchBatchSourceWithRetry(context.Background(), &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, "book-1", "15", 100)
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if attempts != 2 || result.ProcessedText != "正文" {
		t.Fatalf("attempts=%d result=%#v, want one retry and processed text", attempts, result)
	}
}

func TestV12SingleFetchMapsLegacyYGAliasAndCapturesFetchedSource(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "巨量素材书", BookID: "11010573332", Platform: "YG",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("bookid"); got != "11010573332" {
			t.Errorf("bookid=%q, want 11010573332", got)
		}
		if got := r.URL.Query().Get("platform"); got != "4" {
			t.Errorf("platform=%q, want 4", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"获取章节内容成功","data":"这是测试正文"}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/fetch-original"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"done"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Books[0].SourceText != "这是测试正文" || updated.Books[0].TxtText != "这是测试正文" {
		t.Fatalf("fetched source was not captured: %#v", updated.Books[0])
	}
}

func TestV12SingleFetchRetriesCaptureAfterConcurrentBookRevision(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "并发更新测试", BookID: "book-1", Platform: "15",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		latest, getErr := store.GetBatch(context.Background(), "alice", batch.ID)
		if getErr != nil {
			t.Fatal(getErr)
		}
		_, updateErr := store.UpdateBookMetadata(context.Background(), "alice", batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{
			Metadata: map[string]any{"backgroundUpdated": true}, ExpectedRevision: latest.Books[0].Revision,
		})
		if updateErr != nil {
			t.Fatal(updateErr)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"并发后仍应保存的正文"}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/fetch-original"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Books[0].SourceText != "并发后仍应保存的正文" || updated.Books[0].SourceMetadata["backgroundUpdated"] != true {
		t.Fatalf("concurrent metadata or fetched source was lost: %#v", updated.Books[0])
	}
}

func TestV12SingleFetchReusesExistingSourceUnlessForced(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "已有正文", BookID: "book-1", Platform: "15", SourceText: "已经保存的正文",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("upstream should not be called for an existing source")
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/fetch-original", map[string]any{"platform": "15", "maxTxt": 100})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"reused"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12SingleFetchRetriesProcessingFromSavedRawWithoutRefetching(t *testing.T) {
	upstreamCalls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"这是坏词正文","bookinfo":{"category":"女频","genre":"现代悬疑"}}`))
	}))
	defer upstream.Close()

	workshopStore := novelfetchworkshop.NewMemoryStore()
	settings := map[string]any{
		"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true, "apply_sensitive": true}},
		"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "坏词残留", "apply_to_original": true}},
		}},
		"sensitive_ai": map[string]any{"enabled": true, "retries": 0},
	}
	if err := workshopStore.PutConfig(context.Background(), "alice", settings); err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "batch", InitialBatchSettings: batchfactoryv11.SettingsPatch{"sensitiveAIOnFailure": json.RawMessage(`"fail"`)},
		Books: []batchfactoryv11.CreateBookInput{{Title: "测试小说", BookID: "book-1", Platform: "15"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, NovelFetchStore: workshopStore,
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/fetch-original"
	// 用户显式选择“失败即停”策略时仍保留旧语义：返回 422，但 raw 一定保留，可原地重试。
	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"platform": "15", "maxTxt": 100})
	if first.Code != http.StatusUnprocessableEntity || !strings.Contains(first.Body.String(), `"status":"processing_error"`) {
		t.Fatalf("first status=%d body=%s, want the opt-in fail strategy", first.Code, first.Body.String())
	}
	afterFailure, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if afterFailure.Books[0].SourceText != "" || afterFailure.Books[0].TxtText != "这是坏词正文" || afterFailure.Books[0].SourceMetadata["processingStatus"] != "failed" {
		t.Fatalf("failed processing did not retain raw text and retry state: %#v", afterFailure.Books[0])
	}
	// 切回默认“保留原文待确认”策略，并让规则结果安全：重试必须复用已存 raw，不再请求上游。
	if _, err := store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{
		Patch: batchfactoryv11.SettingsPatch{"sensitiveAIOnFailure": json.RawMessage(`"review"`)}, ExpectedRevision: batch.SettingsState.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	settings["sensitive_ai"] = map[string]any{"enabled": false}
	settings["knowledgeLibraries"] = map[string]any{"layout_rules": map[string]any{
		"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "合规词", "apply_to_original": true}},
	}}
	if err := workshopStore.PutConfig(context.Background(), "alice", settings); err != nil {
		t.Fatal(err)
	}
	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"platform": "15", "maxTxt": 100})
	if second.Code != http.StatusOK {
		t.Fatalf("retry status=%d body=%s", second.Code, second.Body.String())
	}
	if upstreamCalls != 1 {
		t.Fatalf("upstream calls=%d, want saved raw text to be reused", upstreamCalls)
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Books[0].SourceText != "这是合规词正文" || updated.Books[0].TxtText != "这是坏词正文" || updated.Books[0].SourceMetadata["processingStatus"] != "done" {
		t.Fatalf("retry did not finish from raw source: %#v", updated.Books[0])
	}
	if updated.Books[0].SourceMetadata["sensitiveReviewRequired"] == true {
		t.Fatalf("review flag must clear once rules resolve every hit: %#v", updated.Books[0].SourceMetadata)
	}
}

func TestV12SingleFetchDegradesToMasksAndReviewFlagInsteadOfDiscardingText(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"这是坏词正文","bookinfo":{"category":"女频","genre":"现代悬疑"}}`))
	}))
	defer upstream.Close()
	workshopStore := novelfetchworkshop.NewMemoryStore()
	settings := map[string]any{
		"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true, "apply_sensitive": true}},
		"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "坏词残留", "apply_to_original": true}},
		}},
		"sensitive_ai": map[string]any{"enabled": true, "retries": 0},
	}
	if err := workshopStore.PutConfig(context.Background(), "alice", settings); err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "测试小说", BookID: "book-1", Platform: "15",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, NovelFetchStore: workshopStore,
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/fetch-original"
	// 默认策略（批次未配置时）：AI 不可用 + 规则结果仍不安全 → 正文打码保留、标待确认、
	// processingStatus=done 继续后续阶段，不再出现“原文待获取”。
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"platform": "15", "maxTxt": 100})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book := updated.Books[0]
	if strings.TrimSpace(book.SourceText) == "" || book.TxtText != "这是坏词正文" {
		t.Fatalf("degraded path discarded text or raw: %#v", book)
	}
	if strings.Contains(book.SourceText, "坏词") {
		t.Fatalf("unprocessed sensitive keyword was stored verbatim: %q", book.SourceText)
	}
	if book.SourceMetadata["processingStatus"] != "done" || book.SourceMetadata["originalRawStatus"] != "done" {
		t.Fatalf("degraded path did not continue: %#v", book.SourceMetadata)
	}
	if book.SourceMetadata["sensitiveReviewRequired"] != true || automationNumberValue(book.SourceMetadata["sensitiveReviewCount"], 0) < 1 {
		t.Fatalf("unprocessed hits were not flagged: %#v", book.SourceMetadata)
	}
	if book.SourceMetadata["sensitiveStatus"] != "review_required" {
		t.Fatalf("sensitiveStatus=%v, want review_required", book.SourceMetadata["sensitiveStatus"])
	}
}

func TestResolveBatchSensitiveConfigDefaultsAndBookOverride(t *testing.T) {
	// 旧批次没有任何新字段时必须保持原行为：AI 修复开、失败保留原文待确认。
	cfg := resolveBatchSensitiveConfig(nil, nil)
	if !cfg.Enabled || cfg.OnFailure != batchSensitiveOnFailureReview || len(cfg.Models) != 0 {
		t.Fatalf("legacy defaults changed: %#v", cfg)
	}
	batchPatch := batchfactoryv11.SettingsPatch{
		"sensitiveAIEnabled":   json.RawMessage(`false`),
		"sensitiveAIOnFailure": json.RawMessage(`"fail"`),
		"sensitiveAIModels":    json.RawMessage(`["primary-model","backup-model"]`),
	}
	cfg = resolveBatchSensitiveConfig(batchPatch, nil)
	if cfg.Enabled || cfg.OnFailure != batchSensitiveOnFailureFail || strings.Join(cfg.Models, ",") != "primary-model,backup-model" {
		t.Fatalf("batch level switch not applied: %#v", cfg)
	}
	// 单书覆盖优先于批次级，沿用现有 settings 继承顺序。
	cfg = resolveBatchSensitiveConfig(batchPatch, batchfactoryv11.SettingsPatch{"sensitiveAIEnabled": json.RawMessage(`true`)})
	if !cfg.Enabled {
		t.Fatalf("book level override did not win: %#v", cfg)
	}
}

func TestProcessBatchConfiguredTextFallsBackToBackupModel(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	settings := map[string]any{
		"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true, "apply_sensitive": true}},
		"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "敏感", "replace": "敏感残留", "apply_to_original": true}},
		}},
		"sensitive_ai": map[string]any{"enabled": true, "model": "global-model", "retries": 0},
	}
	if err := store.PutConfig(context.Background(), "alice", settings); err != nil {
		t.Fatal(err)
	}
	calls := 0
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		calls++
		if calls == 1 {
			return "", nil
		}
		return "他关系越界了", nil
	})
	cfg := batchSensitiveConfig{Enabled: true, Models: []string{"primary-model", "backup-model"}, OnFailure: batchSensitiveOnFailureReview}
	text, _, _, outcome, err := processBatchConfiguredText(context.Background(), store, "alice", "book-1", "他说他敏感了", "original", cfg, provider)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("AI model chain calls=%d, want primary then backup", calls)
	}
	if outcome.ReviewRequired || strings.Contains(text, "敏感") {
		t.Fatalf("backup model success was not accepted: text=%q outcome=%#v", text, outcome)
	}
}

func TestProcessBatchConfiguredTextUsesRulesWhenAISwitchOff(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	settings := map[string]any{
		"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true, "apply_sensitive": true}},
		"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "敏感", "replace": "合规", "apply_to_original": true}},
		}},
		"sensitive_ai": map[string]any{"enabled": true, "model": "global-model"},
	}
	if err := store.PutConfig(context.Background(), "alice", settings); err != nil {
		t.Fatal(err)
	}
	calls := 0
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		calls++
		return "不该被调用", nil
	})
	cfg := batchSensitiveConfig{Enabled: false, OnFailure: batchSensitiveOnFailureReview}
	text, _, _, outcome, err := processBatchConfiguredText(context.Background(), store, "alice", "book-1", "他说他敏感了", "original", cfg, provider)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("AI provider was called while the batch switch is off: %d", calls)
	}
	if outcome.ReviewRequired || strings.Contains(text, "敏感") {
		t.Fatalf("rule-only path did not finish safely: text=%q outcome=%#v", text, outcome)
	}
}

func TestPublishVideoBlocksUnconfirmedSensitiveReview(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "测试小说", BookID: "book-1", Platform: "15"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err := store.UpdateBookMetadata(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.UpdateBookMetadataInput{
		Metadata: map[string]any{"sensitiveReviewRequired": true, "sensitiveReviewCount": 2, "sensitiveStatus": "review_required"}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, nil, nil, nil, func() time.Time { return time.Unix(1700000000, 0) })
	err = coordinator.publishVideo(context.Background(), "alice", batch.ID, book.ID)
	if !errors.Is(err, errSensitiveReviewPending) {
		t.Fatalf("unconfirmed sensitive review must block upload, got %v", err)
	}
	if !strings.Contains(err.Error(), "2 处") {
		t.Fatalf("block message did not carry the unprocessed count: %v", err)
	}
}
