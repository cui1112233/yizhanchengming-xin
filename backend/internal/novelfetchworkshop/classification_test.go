package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestClassifyBookUsesOneSharedUpstreamFirstChain(t *testing.T) {
	called := 0
	result := ClassifyBook(context.Background(), "", BookInfo{Text: "正文"}, func(context.Context, BookInfo) (string, error) {
		called++
		return `{"gender":"女频","confidence":0.91,"reason":"情感关系线围绕女主展开"}`, nil
	}, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		called++
		return `{"style":"现代悬疑","confidence":0.9,"reason":"正文围绕案件线索推进"}`, nil
	}))
	if result.Gender.Gender != "女频" || result.Style.Style != "现代悬疑" {
		t.Fatalf("result=%#v", result)
	}
	if called != 2 {
		t.Fatalf("called=%d, want one gender and one style resolution", called)
	}
}
