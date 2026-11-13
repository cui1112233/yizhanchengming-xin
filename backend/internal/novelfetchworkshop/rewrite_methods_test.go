package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestRewriteJobMethodsOverrideConfigWithoutChangingAccount(t *testing.T) {
	for _, method := range []string{"high_imitation", "opening_instruction", "instruction", "auto"} {
		t.Run(method, func(t *testing.T) {
			ctx := context.Background()
			s := plainRewriteTestStore("alice")
			cfg, _ := s.GetConfig(ctx, "alice")
			app := mapOrEmpty(cfg["appConfig"])
			rewrite := mapOrEmpty(app["rewrite"])
			rewrite["ai_slot_methods"] = map[string]any{"ai1": "high_imitation"}
			rewrite["method_sequence"] = []any{"instruction"}
			app["rewrite"] = rewrite
			cfg["appConfig"] = app
			if err := s.PutConfig(ctx, "alice", cfg); err != nil {
				t.Fatal(err)
			}
			_ = s.PutDocument(ctx, "alice", Document{BookID: "book-method", Original: "首行原文"})
			calls := 0
			doc, err := GenerateAIRewritesForVersionsProcessed(ctx, s, "alice", "book-method", []string{"ai1"}, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) { calls++; return "首行改文", nil }), nil, map[string]string{"ai1": method})
			if err != nil {
				t.Fatal(err)
			}
			want := method
			if want == "auto" {
				want = "instruction"
			}
			actual := doc.Meta["rewriteEvidenceByVersion"].(map[string]any)["ai1"].(map[string]any)["method"]
			if actual != want || calls != 1 {
				t.Fatalf("method=%v want=%s calls=%d", actual, want, calls)
			}
			saved, _ := s.GetConfig(ctx, "alice")
			if mapOrEmpty(mapOrEmpty(mapOrEmpty(saved["appConfig"])["rewrite"])["ai_slot_methods"])["ai1"] != "high_imitation" {
				t.Fatal("job altered account defaults")
			}
		})
	}
}
