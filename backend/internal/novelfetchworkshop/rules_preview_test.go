package novelfetchworkshop

import (
	"strings"
	"testing"
)

func TestPreviewRulesMatchesWorkbenchCoreStages(t *testing.T) {
	text, trace := PreviewRules("第1章\n你好。。\n“测试", "ai", nil, nil)
	if strings.Contains(text, "第1章") || strings.Contains(text, "。。") || !strings.Contains(text, "“测试”") {
		t.Fatalf("unexpected preview text: %q", text)
	}
	if len(trace) < 5 {
		t.Fatalf("expected staged trace, got %#v", trace)
	}
	if trace[0]["stage"] != "raw" || trace[len(trace)-1]["stage"] != "paragraph" {
		t.Fatalf("unexpected trace boundaries: %#v", trace)
	}
}

func TestPreviewRulesHonorsScopeDisable(t *testing.T) {
	text, trace := PreviewRules("第1章\n你好。。", "original", map[string]any{"apply_to_original": false}, nil)
	if text != "第1章\n你好。。" || len(trace) != 1 {
		t.Fatalf("disabled scope should return normalized raw only: text=%q trace=%#v", text, trace)
	}
}

func TestApplySensitiveRulesUsesConfiguredKnowledgeAndKeepsRawCallerSeparate(t *testing.T) {
	settings := map[string]any{"knowledge": map[string]any{"layout_rules": map[string]any{
		"sensitive_rules": []any{
			map[string]any{"enabled": true, "find": "脏词", "replace": "文明词"},
			"@{enabled=True; find=坏词; replace=好词}",
		},
	}}}
	got, count := ApplySensitiveRules("脏词和坏词", settings)
	if got != "文明词和好词" || count != 2 {
		t.Fatalf("sensitive replacement = %q count=%d", got, count)
	}
}

func TestApplySensitiveRulesWithDetailsKeepsBeforeAndAfterItems(t *testing.T) {
	text := "这是坏词。\n坏词再次出现。"
	settings := map[string]any{"knowledge": map[string]any{"layout_rules": map[string]any{
		"sensitive_rules": []any{map[string]any{"find": "坏词", "replace": "合规词", "enabled": true}},
	}}}
	processed, count, items := ApplySensitiveRulesWithDetails(text, settings)
	if processed != "这是合规词。\n合规词再次出现。" || count != 2 || len(items) != 2 {
		t.Fatalf("processed=%q count=%d items=%#v", processed, count, items)
	}
	if items[0]["keyword"] != "坏词" || items[0]["before"] != "这是坏词。" || items[0]["after"] != "这是合规词。" {
		t.Fatalf("first item=%#v", items[0])
	}
}
