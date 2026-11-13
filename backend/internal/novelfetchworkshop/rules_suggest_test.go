package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestSuggestedRulesExecuteThroughConfiguredPreview(t *testing.T) {
	for _, tc := range []struct{ kind, output, input, want string }{
		{"sensitive", `{"groups":[{"name":"建议","enabled":true,"rules":[{"find":"坏词","replace":"好词","enabled":true}]}]}`, "坏词", "好词"},
		{"chapter_inline_rules", `{"chapter_inline_rules":[{"find":"^目录前缀","replace":"","enabled":true}],"chapter_remove_inline":true}`, "目录前缀 正文\n保留正文", "保留正文"},
		{"pair_fill_rules", `{"pair_fill_rules":[{"find":"“","replace":"”","enabled":true}]}`, "“正文", "“正文”"},
		{"layout_rules", `{"indent_spaces":2}`, "保留故事", "  保留故事"},
		{"symbol_rules", `{"symbol_rules":[{"find":"，","replace":"","enabled":true}],"enable_symbol_trim":true,"symbol_trim_positions":"2"}`, "甲，乙", "甲，\n乙"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			result, err := SuggestRules(context.Background(), rewriteProviderFunc(func(_ context.Context, _ RewriteRequest) (string, error) { return tc.output, nil }), tc.kind, "处理样本", tc.input)
			if err != nil {
				t.Fatal(err)
			}
			key := tc.kind
			if key == "chapter_inline_rules" {
				key = "chapter_rules"
			}
			if key == "pair_fill_rules" {
				key = "symbol_rules"
			}
			settings := map[string]any{"knowledgeLibraries": map[string]any{key: result}}
			SyncKnowledgeConfigViews(settings, settings)
			processed, _, _ := ProcessConfiguredText(tc.input, "original", settings)
			if processed != tc.want {
				t.Fatalf("got %q want %q", processed, tc.want)
			}
		})
	}
}

func TestRuleSuggestionRejectsUnexecutableFieldsAndRegex(t *testing.T) {
	for _, raw := range []string{`{"summary":"建议","rules":[{"pattern":"x","replacement":"y"}]}`, `{"chapter_inline_rules":[{"find":"(?<=a)b","replace":"","enabled":true}],"chapter_remove_inline":true}`, `{"chapter_inline_rules":[],"chapter_remove_inline":true}`} {
		_, err := SuggestRules(context.Background(), rewriteProviderFunc(func(_ context.Context, _ RewriteRequest) (string, error) { return raw, nil }), "chapter_inline_rules", "", "样本")
		if err == nil {
			t.Fatalf("accepted invalid rules: %s", raw)
		}
	}
}

func TestChapterInlineRulesMatchOnlyAtLineStartAndRespectSwitch(t *testing.T) {
	chapter := map[string]any{"chapter_remove_inline": true, "chapter_remove_standalone": false, "chapter_inline_rules": []any{map[string]any{"find": "章节", "replace": "", "enabled": true}}}
	if got := previewChapterCleanup("正文章节\n章节正文\n1", nil, chapter); got != "正文章节\n\n1" {
		t.Fatalf("got %q", got)
	}
	chapter["chapter_remove_inline"] = false
	if got := previewChapterCleanup("章节正文", nil, chapter); got != "章节正文" {
		t.Fatalf("switch ignored: %q", got)
	}
}
