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
