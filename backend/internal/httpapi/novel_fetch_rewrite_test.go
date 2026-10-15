package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	if items[0]["source"] != "AI鉴组处理" || items[0]["model"] != "[anti-2]gemini-3.8-flash" || items[0]["attempt"] != 1 || items[0]["status"] != "done" {
		t.Fatalf("repair audit = %#v", items[0])
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
	if item["source"] != "AI鉴组处理" || item["model"] != "[anti-2]gemini-3.8-flash" {
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
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t测试书", "sensitive_ai_enabled": true}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("process start status=%d body=%s", response.Code, response.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.ProcessedText != "他说关系越界了" {
		t.Fatalf("process start did not execute AI repair: %q response=%s", document.ProcessedText, response.Body.String())
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
	calls := 0
	provider := rewriteProviderFunc(func(_ context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
		calls++
		return "改写后的正文", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}, NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{
		"inputText": "12345678901\t测试书", "selected_versions": []string{"original", "ai3"},
	}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("process start status=%d body=%s", response.Code, response.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || document.Versions["ai3"] != "改写后的正文" {
		t.Fatalf("selected AI version was not generated exactly once: calls=%d versions=%#v", calls, document.Versions)
	}
	if _, exists := document.Versions["ai1"]; exists {
		t.Fatalf("unselected AI version was generated: %#v", document.Versions)
	}
}

type rewriteProviderFunc func(context.Context, novelfetchworkshop.RewriteRequest) (string, error)

func (f rewriteProviderFunc) Complete(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
	return f(ctx, request)
}
