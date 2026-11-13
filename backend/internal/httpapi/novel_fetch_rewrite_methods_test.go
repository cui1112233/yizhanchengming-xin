package httpapi

import (
	"context"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
)

func TestWorkshopWorkerUsesSelectedRewriteMethod(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	_ = s.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "book-method", Gender: "女频", Original: "已有原文", ProcessedText: "已有原文", Meta: map[string]any{"originalStatus": "done", "style": "现代悬疑"}})
	_, _ = s.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{BookID: "book-method", BodyRef: novelfetchworkshop.BodyRef{VersionID: "processed_text"}, Content: "已有原文"})
	calls := 0
	w := &workshopProcessWorker{store: s, resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
		return rewriteProviderFunc(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			calls++
			return `{"id":"book-method","changed_lines":[{"line_no":1,"text":"改文结果"}]}`, nil
		}), nil
	}}
	if err := w.execute(ctx, novelfetchworkshop.ProcessJob{Owner: "tester", BookIDs: []string{"book-method"}, SelectedVersions: []string{"ai1"}, RewriteMethods: map[string]string{"ai1": "instruction"}}); err != nil {
		t.Fatal(err)
	}
	doc, _ := s.GetDocument(ctx, "tester", "book-method")
	evidence := doc.Meta["rewriteEvidenceByVersion"].(map[string]any)["ai1"].(map[string]any)
	if calls != 1 || evidence["method"] != "instruction" {
		t.Fatalf("selected method not consumed: calls=%d method=%v", calls, evidence["method"])
	}
}
