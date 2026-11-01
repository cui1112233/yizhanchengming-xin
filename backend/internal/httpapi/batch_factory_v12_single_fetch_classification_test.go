package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

type singleFetchRewriteProvider func(context.Context, novelfetchworkshop.RewriteRequest) (string, error)

func (f singleFetchRewriteProvider) Complete(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
	return f(ctx, request)
}

func TestBatchFactoryProcessingStageSummaryDoesNotCarrySensitiveDetails(t *testing.T) {
	summary := summarizeBatchProcessingStages([]map[string]any{{
		"stage": "sensitive", "title": "敏感词替换", "chars": 20, "changed": true,
		"samples": []any{"私密原文片段"}, "before": "私密原文片段", "keyword": "私密命中词",
	}})
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "私密原文片段") || strings.Contains(string(encoded), "私密命中词") || strings.Contains(string(encoded), "samples") {
		t.Fatalf("summary leaked sensitive details: %s", encoded)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil || len(decoded) != 1 || decoded[0]["changed"] != true || decoded[0]["chars"] != float64(20) {
		t.Fatalf("summary lost safe stage information: %s err=%v", encoded, err)
	}
}

func TestResolveSingleFetchPlatform(t *testing.T) {
	giantBook := batchfactoryv11.Book{
		Platform:       "七猫",
		SourceMetadata: map[string]any{"platformId": "3"},
	}
	cases := []struct {
		name  string
		input string
		book  batchfactoryv11.Book
		want  string
	}{
		{name: "explicit input wins", input: "15", book: giantBook, want: "15"},
		{name: "numeric book platform used", book: batchfactoryv11.Book{Platform: "3"}, want: "3"},
		{name: "giant material name falls back to metadata platformId", book: giantBook, want: "3"},
		{name: "platform_id key variant", book: batchfactoryv11.Book{Platform: "知乎", SourceMetadata: map[string]any{"platform_id": "15"}}, want: "15"},
		{name: "non-numeric metadata ignored", book: batchfactoryv11.Book{Platform: "番茄", SourceMetadata: map[string]any{"platformId": "fanqie"}}, want: "番茄"},
		{name: "empty stays empty", book: batchfactoryv11.Book{}, want: ""},
	}
	for _, tc := range cases {
		if got := resolveSingleFetchPlatform(tc.input, tc.book); got != tc.want {
			t.Errorf("%s: resolveSingleFetchPlatform(%q) = %q, want %q", tc.name, tc.input, got, tc.want)
		}
	}
}

func TestClassifyFetchedBatchBookUsesAIForUnspecifiedBookInfo(t *testing.T) {
	metadata := classifyFetchedBatchBook(context.Background(), "alice", novelfetchworkshop.BookInfo{}, "正文", map[string]any{}, batchFetchClassification{
		AIGenderJudge: func(context.Context, novelfetchworkshop.BookInfo) (string, error) {
			return `{"gender":"女频","confidence":0.9,"reason":"正文情感关系线突出"}`, nil
		},
		RewriteProvider: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
		}),
	})
	if metadata["gender"] != "女频" || metadata["genderSource"] != "ai" {
		t.Fatalf("gender metadata = %#v", metadata)
	}
	if metadata["genderConfidence"] != 0.9 {
		t.Fatalf("gender confidence = %#v", metadata["genderConfidence"])
	}
	if metadata["style"] != "现代悬疑" || metadata["styleSource"] != "ai" {
		t.Fatalf("style metadata = %#v", metadata)
	}
	if metadata["styleConfidence"] != 0.9 {
		t.Fatalf("style confidence = %#v", metadata["styleConfidence"])
	}
}
