package httpapi

import (
	"context"
	"net/http"
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
		return "ok", nil
	})
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore(), NovelFetchRewrite: provider})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/ai/test", map[string]any{}, now, secret)
	if response.Code != http.StatusOK || response.Body.String() == "" {
		t.Fatalf("ai test status=%d body=%s", response.Code, response.Body.String())
	}
}

type rewriteProviderFunc func(context.Context, novelfetchworkshop.RewriteRequest) (string, error)

func (f rewriteProviderFunc) Complete(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
	return f(ctx, request)
}
