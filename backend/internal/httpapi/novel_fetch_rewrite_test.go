package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestNovelFetchWorkshopGenerateAIRouteUsesGoProvider(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "book-rewrite", Original: "原始正文"}); err != nil {
		t.Fatal(err)
	}
	provider := rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		if request.UserPrompt == "" {
			t.Fatal("rewrite prompt is empty")
		}
		return "Go 改写正文", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/book-rewrite/generate-ai", map[string]any{"count": 1}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("generate route status=%d body=%s", response.Code, response.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "book-rewrite")
	if err != nil {
		t.Fatal(err)
	}
	if document.Versions["ai1"] != "Go 改写正文" {
		t.Fatalf("Go rewrite version missing: %#v", document.Versions)
	}
}

func TestNovelFetchWorkshopAITestRouteUsesGoProvider(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	provider := rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		if request.UserPrompt != "请只回复：ok" {
			t.Fatalf("unexpected test prompt: %#v", request)
		}
		if request.MaxTokens < 32 {
			t.Fatalf("AI test budget is too small: %d", request.MaxTokens)
		}
		return "ok", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore(), NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/ai/test", map[string]any{}, now, secret)
	if response.Code != http.StatusOK || response.Body.String() == "" {
		t.Fatalf("ai test status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestApplySensitiveAIRepairRecordsExecutionAndReplacesHit(t *testing.T) {
	provider := rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		if request.UserPrompt == "" {
			t.Fatal("sensitive repair prompt is empty")
		}
		return "他说关系越界了", nil
	})
	settings := map[string]any{
		"sensitive_ai": map[string]any{
			"enabled": true,
			"model":   "[anti-2]gemini-3.8-flash",
			"retries": 1,
		},
	}
	items := []map[string]any{{"hit_index": 1, "keyword": "出轨", "before": "他说他出轨了", "after": "他说他关系越界了", "status": "done"}}
	text, items := applySensitiveAIRepair(context.Background(), provider, settings, "book-1", "他说他关系越界了", items)
	if text != "他说关系越界了" {
		t.Fatalf("repaired text = %q", text)
	}
	if items[0]["source"] != "ai" || items[0]["model"] != "[anti-2]gemini-3.8-flash" || items[0]["attempt"] != 1 || items[0]["status"] != "done" {
		t.Fatalf("repair audit = %#v", items[0])
	}
}

func TestApplySensitiveAIRepairReplacesRepeatedLinesInOrder(t *testing.T) {
	outputs := []string{"第一段已改", "第二段已改"}
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		output := outputs[0]
		outputs = outputs[1:]
		return output, nil
	})
	settings := map[string]any{"sensitive_ai": map[string]any{"enabled": true}}
	items := []map[string]any{
		{"hit_index": 1, "keyword": "敏感", "before": "相同敏感段", "after": "相同已替换段", "status": "done"},
		{"hit_index": 2, "keyword": "敏感", "before": "相同敏感段", "after": "相同已替换段", "status": "done"},
	}
	text, _ := applySensitiveAIRepair(context.Background(), provider, settings, "book-1", "相同已替换段\n相同已替换段", items)
	if text != "第一段已改\n第二段已改" {
		t.Fatalf("repeated hit order was not preserved: %q", text)
	}
}

func TestApplySensitiveAIRepairHonorsConfiguredConcurrencyAndPreservesOrder(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	provider := rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		current := active.Add(1)
		for observed := maximum.Load(); current > observed; observed = maximum.Load() {
			if maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		defer active.Add(-1)
		time.Sleep(25 * time.Millisecond)
		switch {
		case strings.Contains(request.UserPrompt, "敏感一"):
			return "改好一", nil
		case strings.Contains(request.UserPrompt, "敏感二"):
			return "", errors.New("单条修复失败")
		case strings.Contains(request.UserPrompt, "敏感三"):
			return "改好三", nil
		default:
			return "未知片段", nil
		}
	})
	settings := map[string]any{"sensitive_ai": map[string]any{"enabled": true, "group_hits": false, "concurrency": 2}}
	items := []map[string]any{
		{"hit_index": 1, "keyword": "敏感一", "before": "原文敏感一", "after": "替换一", "status": "done"},
		{"hit_index": 2, "keyword": "敏感二", "before": "原文敏感二", "after": "替换二", "status": "done"},
		{"hit_index": 3, "keyword": "敏感三", "before": "原文敏感三", "after": "替换三", "status": "done"},
	}
	text, got := applySensitiveAIRepair(context.Background(), provider, settings, "book-1", "替换一\n替换二\n替换三", items)
	if maximum.Load() != 2 {
		t.Fatalf("maximum concurrent calls = %d, want configured concurrency 2", maximum.Load())
	}
	if text != "改好一\n替换二\n改好三" {
		t.Fatalf("repaired text order = %q", text)
	}
	if got[0]["status"] != "done" || got[1]["status"] != "failed" || strings.TrimSpace(firstText(got[1]["error"])) == "" || got[2]["status"] != "done" {
		t.Fatalf("per-hit result order/error was not preserved: %#v", got)
	}
}

func TestApplySensitiveAIRepairHonorsGroupModeAndHitLimit(t *testing.T) {
	var calls int
	provider := rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		calls++
		if !strings.Contains(request.UserPrompt, "1.") || !strings.Contains(request.UserPrompt, "2.") {
			t.Fatalf("group prompt did not include both hits: %q", request.UserPrompt)
		}
		return `["第一段已改","第二段已改"]`, nil
	})
	settings := map[string]any{"sensitive_ai": map[string]any{"enabled": true, "group_hits": true, "max_hits_per_task": 2}}
	items := []map[string]any{
		{"keyword": "敏感1", "before": "第一段敏感1", "after": "第一段已替换"},
		{"keyword": "敏感2", "before": "第二段敏感2", "after": "第二段已替换"},
		{"keyword": "敏感3", "before": "第三段敏感3", "after": "第三段已替换"},
	}
	text, got := applySensitiveAIRepair(context.Background(), provider, settings, "book-1", "第一段已替换\n第二段已替换\n第三段已替换", items)
	if calls != 1 {
		t.Fatalf("group mode calls=%d, want 1", calls)
	}
	if text != "第一段已改\n第二段已改\n第三段已替换" {
		t.Fatalf("group repair text=%q", text)
	}
	if got[2]["status"] != "skipped_limit" {
		t.Fatalf("third hit status=%#v, want skipped_limit", got[2]["status"])
	}
}

func TestApplySensitiveAIRepairRejectsUnsafeOrIncompleteGroupOutput(t *testing.T) {
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		return `["仍然包含敏感词"]`, nil
	})
	settings := map[string]any{"sensitive_ai": map[string]any{"enabled": true, "group_hits": true}}
	items := []map[string]any{
		{"keyword": "敏感词", "before": "原文敏感词", "after": "替换后"},
		{"keyword": "另一个词", "before": "原文另一个词", "after": "替换后2"},
	}
	text, got := applySensitiveAIRepair(context.Background(), provider, settings, "book-1", "替换后\n替换后2", items)
	if text != "替换后\n替换后2" {
		t.Fatalf("unsafe group changed text=%q", text)
	}
	if got[0]["status"] != "failed" || got[1]["status"] != "failed" {
		t.Fatalf("unsafe/incomplete statuses=%#v", got)
	}
}

func TestNovelFetchOriginalUsesEnabledSensitiveAIRepair(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"他说他出轨了","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345678901", MaxTxt: 100, Meta: map[string]any{"platformId": "15"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"sensitive_rules": []any{map[string]any{"find": "出轨", "replace": "他说他关系越界了", "enabled": true}},
		"sensitive_ai":    map[string]any{"enabled": true, "model": "[anti-2]gemini-3.8-flash"},
	}); err != nil {
		t.Fatal(err)
	}
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		return "他说关系越界了", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345678901/fetch", map[string]any{"sensitive_ai_enabled": true}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("fetch status=%d body=%s", response.Code, response.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.ProcessedText != "他说关系越界了" {
		t.Fatalf("AI repair did not replace text: %q", document.ProcessedText)
	}
	fixed, _ := document.Meta["sensitiveFixed"].(map[string]any)
	items, _ := fixed["items"].([]any)
	if len(items) != 1 {
		raw, _ := json.Marshal(document.Meta["sensitiveFixed"])
		t.Fatalf("sensitive audit missing: %s", raw)
	}
	item, _ := items[0].(map[string]any)
	if item["source"] != "ai" || item["model"] != "[anti-2]gemini-3.8-flash" {
		t.Fatalf("AI audit mismatch: %#v", item)
	}
}

func TestNovelFetchProcessStartRunsFetchAndSensitiveAIRepair(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"他说他出轨了","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{map[string]any{"find": "出轨", "replace": "他说他关系越界了", "enabled": true}}}}}}); err != nil {
		t.Fatal(err)
	}
	checkConfig, _ := store.GetConfig(context.Background(), "tester")
	if got, count := novelfetchworkshop.ApplySensitiveRules("他说他出轨了", checkConfig); count != 1 || got == "他说他出轨了" {
		t.Fatalf("test sensitive config not applied: %#v count=%d", checkConfig, count)
	}
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		return "他说关系越界了", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: novelfetchworkshop.NewMemoryQueue(), NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t测试书", "sensitive_ai_enabled": true}, now, secret)
	if response.Code != http.StatusAccepted {
		t.Fatalf("process start status=%d body=%s", response.Code, response.Body.String())
	}
	document := waitForNovelFetchDocument(t, store, "12345678901", func(document novelfetchworkshop.Document) bool {
		return document.ProcessedText == "他说关系越界了"
	})
	if document.ProcessedText != "他说关系越界了" {
		t.Fatalf("process start did not execute AI repair: %q response=%s", document.ProcessedText, response.Body.String())
	}
}

func TestNovelFetchProcessStartPersistsRawBeforeSensitiveAIFailure(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"他说他出轨了","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{map[string]any{"find": "出轨", "replace": "他说他关系越界了", "enabled": true}}}}},
	}); err != nil {
		t.Fatal(err)
	}
	provider := rewriteProviderFunc(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
		return "", errors.New("敏感词模型暂时不可用")
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: novelfetchworkshop.NewMemoryQueue(), NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t测试书", "sensitive_ai_enabled": true}, now, secret)
	if response.Code != http.StatusAccepted {
		t.Fatalf("process start status=%d body=%s", response.Code, response.Body.String())
	}
	document := waitForNovelFetchDocument(t, store, "12345678901", func(document novelfetchworkshop.Document) bool {
		return document.Meta["status"] == "waiting_retry" || document.Meta["status"] == "failed"
	})
	if document.OriginalRaw != "他说他出轨了" {
		t.Fatalf("raw original was lost after sensitive AI failure: raw=%q meta=%#v", document.OriginalRaw, document.Meta)
	}
	if document.Meta["originalStatus"] != "waiting_sensitive_ai" {
		t.Fatalf("original status=%#v, want waiting_sensitive_ai", document.Meta["originalStatus"])
	}
}

func TestNovelFetchProcessStartSkipsRewriteForShortOriginal(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"短文","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"workflow": map[string]any{"skip_short_original": true, "min_original_chars": 10},
	}); err != nil {
		t.Fatal(err)
	}
	provider := rewriteProviderFunc(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
		return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: novelfetchworkshop.NewMemoryQueue(), NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t测试书", "selected_versions": []string{"original", "ai1"}}, now, secret)
	if response.Code != http.StatusAccepted {
		t.Fatalf("process start status=%d body=%s", response.Code, response.Body.String())
	}
	document := waitForNovelFetchDocument(t, store, "12345678901", func(document novelfetchworkshop.Document) bool {
		return document.Meta["originalStatus"] == "too_short"
	})
	if document.Original != "短文" || document.ProcessedText != "短文" || document.Meta["aiStatus"] != "skipped_short_original" {
		t.Fatalf("short original was not preserved and skipped: %#v", document)
	}
}

func TestNovelFetchProcessStartGeneratesOnlySelectedAIVersion(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"原文第一行","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	rewriteCalls := 0
	classificationCalls := 0
	provider := rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		if strings.Contains(request.SystemPrompt, "小说分类器") {
			classificationCalls++
			return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
		}
		rewriteCalls++
		return "改写后的正文", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchQueue: novelfetchworkshop.NewMemoryQueue(), NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText": "12345678901\t测试书", "selected_versions": []string{"original", "ai3"},
	}, now, secret)
	if response.Code != http.StatusAccepted {
		t.Fatalf("process start status=%d body=%s", response.Code, response.Body.String())
	}
	document := waitForNovelFetchDocument(t, store, "12345678901", func(document novelfetchworkshop.Document) bool {
		return document.Versions["ai3"] == "改写后的正文"
	})
	if rewriteCalls != 1 || document.Versions["ai3"] != "改写后的正文" {
		t.Fatalf("selected AI version was not generated exactly once: rewrite=%d versions=%#v", rewriteCalls, document.Versions)
	}
	// The upstream category/genre already determines gender and style, so the
	// pipeline must not spend an extra classification AI call.
	if classificationCalls != 0 {
		t.Fatalf("platform classification was sufficient but AI was called %d times", classificationCalls)
	}
	if document.Gender != "男频" || firstText(document.Meta["style"]) != "男频都市" {
		t.Fatalf("platform classification was not applied: gender=%q style=%v", document.Gender, document.Meta["style"])
	}
	if _, exists := document.Versions["ai1"]; exists {
		t.Fatalf("unselected AI version was generated: %#v", document.Versions)
	}
}

type rewriteProviderFunc func(context.Context, novelfetchworkshop.RewriteRequest) (string, error)

func (f rewriteProviderFunc) Complete(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
	return f(ctx, request)
}

func waitForNovelFetchDocument(t *testing.T, store novelfetchworkshop.Store, bookID string, ready func(novelfetchworkshop.Document) bool) novelfetchworkshop.Document {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		document, err := store.GetDocument(context.Background(), "tester", bookID)
		if err == nil && ready(document) {
			return document
		}
		time.Sleep(10 * time.Millisecond)
	}
	document, _ := store.GetDocument(context.Background(), "tester", bookID)
	t.Fatalf("background novel fetch did not reach expected state: %#v", document)
	return novelfetchworkshop.Document{}
}
