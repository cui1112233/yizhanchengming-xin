package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

type singleFetchRewriteProvider func(context.Context, novelfetchworkshop.RewriteRequest) (string, error)

func (f singleFetchRewriteProvider) Complete(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
	return f(ctx, request)
}

func TestClassifyFetchedBatchBookUsesAIForUnspecifiedBookInfo(t *testing.T) {
	metadata := classifyFetchedBatchBook(context.Background(), "alice", novelfetchworkshop.BookInfo{}, "正文", map[string]any{}, batchFetchClassification{
		AIGenderJudge: func(context.Context, novelfetchworkshop.BookInfo) (string, error) { return "女频", nil },
		RewriteProvider: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代悬疑"}`, nil
		}),
	})
	if metadata["gender"] != "女频" || metadata["genderSource"] != "ai" {
		t.Fatalf("gender metadata = %#v", metadata)
	}
	if metadata["style"] != "现代悬疑" || metadata["styleSource"] != "ai" {
		t.Fatalf("style metadata = %#v", metadata)
	}
}
