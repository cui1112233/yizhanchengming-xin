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
