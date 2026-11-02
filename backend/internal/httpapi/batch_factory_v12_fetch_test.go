package httpapi

import (
	"context"
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
			"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "合规词", "apply_to_original": true}},
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
	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"platform": "15", "maxTxt": 100})
	if first.Code != http.StatusUnprocessableEntity || !strings.Contains(first.Body.String(), `"status":"processing_error"`) {
		t.Fatalf("first status=%d body=%s, want a visible processing failure", first.Code, first.Body.String())
	}
	afterFailure, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if afterFailure.Books[0].SourceText != "" || afterFailure.Books[0].TxtText != "这是坏词正文" || afterFailure.Books[0].SourceMetadata["processingStatus"] != "failed" {
		t.Fatalf("failed processing did not retain raw text and retry state: %#v", afterFailure.Books[0])
	}
	settings["sensitive_ai"] = map[string]any{"enabled": false}
	if err := workshopStore.PutConfig(context.Background(), "alice", settings); err != nil {
		t.Fatal(err)
	}
	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"platform": "15", "maxTxt": 100})
	if second.Code != http.StatusOK || !strings.Contains(second.Body.String(), `"status":"done"`) {
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
}
