package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestExplicitEmptySensitiveReplacementSurvivesDefaultsAndDeletesMatch(t *testing.T) {
	rule := map[string]any{"find": "代孕", "replace": "", "replace_explicit": true, "enabled": true}
	settings := map[string]any{"knowledgeLibraries": map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{
		"name": "明确替换", "enabled": true, "apply_to_original": true, "apply_to_ai": true, "rules": []any{rule},
	}}}}}
	if changed := EnsureSensitiveRuleReplacements(settings); changed != 0 || rule["replace"] != "" {
		t.Fatalf("explicit deletion overwritten: changed=%d rule=%#v", changed, rule)
	}
	store := NewMemoryStore()
	if err := store.PutConfig(context.Background(), "test-explicit", settings); err != nil {
		t.Fatal(err)
	}
	saved, err := store.GetConfig(context.Background(), "test-explicit")
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"original", "ai"} {
		output, hits, _ := ApplySensitiveRulesForScopeWithDetails("正文代孕结束", saved, scope)
		if output != "正文结束" || hits != 1 {
			t.Fatalf("scope=%s output=%q hits=%d", scope, output, hits)
		}
	}
}

func TestMissingSensitiveReplacementStillGetsDefault(t *testing.T) {
	rule := map[string]any{"find": "出轨", "replace": "", "enabled": true}
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{rule}}}}}
	if EnsureSensitiveRuleReplacements(settings) != 1 || rule["replace"] == "" {
		t.Fatalf("default repair disabled: %#v", rule)
	}
}
