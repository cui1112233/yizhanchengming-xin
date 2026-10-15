package novelfetchworkshop

import (
	"context"
	"testing"
)

type styleProviderFunc func(context.Context, RewriteRequest) (string, error)

func (f styleProviderFunc) Complete(ctx context.Context, req RewriteRequest) (string, error) {
	return f(ctx, req)
}

func TestResolveStyleUsesAIOnlyWhenBookInfoIsNotExplicit(t *testing.T) {
	called := 0
	got := ResolveStyle(context.Background(), BookInfo{Genre: "现代悬疑", Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		called++
		return `{"style":"古风虐文"}`, nil
	}))
	if got.Style != "现代悬疑" || got.Source != "bookinfo" || called != 0 {
		t.Fatalf("got=%+v called=%d", got, called)
	}
}

func TestResolveStyleUsesPlatformLabelsBeforeAI(t *testing.T) {
	calls := 0
	got := ResolveStyle(context.Background(), BookInfo{Labels: "现代", Themes: "言情、婚恋", Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		calls++
		return `{"style":"玄幻"}`, nil
	}))
	if got.Style != "现代女主" || got.Source != "bookinfo" || calls != 0 {
		t.Fatalf("label style decision = %#v, calls=%d", got, calls)
	}
}

func TestResolveStyleFallsBackToAIAndKeepsUnknownPending(t *testing.T) {
	good := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return `{"style":"现代女主"}`, nil
	}))
	if good.Style != "现代女主" || good.Source != "ai" {
		t.Fatalf("good=%+v", good)
	}
	unknown := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return `{"style":"无法确定"}`, nil
	}))
	if unknown.Style != "" {
		t.Fatalf("unknown=%+v", unknown)
	}
}

func TestResolveStyleAcceptsCodeFenceAndPlainStyleResponses(t *testing.T) {
	for _, response := range []string{"```json\n{\"style\":\"现代女主\"}\n```", "现代女主"} {
		got := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
			return response, nil
		}))
		if got.Style != "现代女主" {
			t.Fatalf("response %q resolved to %q", response, got.Style)
		}
	}
}
