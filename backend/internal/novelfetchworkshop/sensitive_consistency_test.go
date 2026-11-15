package novelfetchworkshop

import "testing"

func TestSensitiveCanonicalLibraryWinsOverLegacyMirrors(t *testing.T) {
	for _, groups := range []any{
		[]any{map[string]any{"enabled": true, "rules": []any{map[string]any{"find": "原词", "replace": "新词", "enabled": true}}}},
		[]any{},
		[]any{map[string]any{"enabled": false, "rules": []any{map[string]any{"find": "原词", "replace": "新词"}}}},
	} {
		settings := map[string]any{"knowledgeLibraries": map[string]any{"sensitive": map[string]any{"groups": groups}, "layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"find": "原词", "replace": "旧词"}}}}, "appConfig": map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{map[string]any{"find": "原词", "replace": "旧词"}}}}}}}
		want := "原词"
		if len(groups.([]any)) > 0 && groups.([]any)[0].(map[string]any)["enabled"] == true {
			want = "新词"
		}
		got, _, _ := ApplySensitiveRulesForScopeWithDetails("原词", settings, "original")
		if got != want {
			t.Errorf("got %q want %q for %#v", got, want, groups)
		}
		count := 0
		for _, g := range groups.([]any) {
			count += len(g.(map[string]any)["rules"].([]any))
		}
		if got := KnowledgeSummary(settings)["sensitive_rules"]; got != count {
			t.Errorf("summary=%d want %d", got, count)
		}
	}
}

func TestSensitiveExplicitDeletionDoesNotFallBackToReplacementAlias(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{map[string]any{"find": "代孕", "replace": "", "replacement": "旧替换", "replace_explicit": true}}}}}}
	got, _, _ := ApplySensitiveRulesForScopeWithDetails("甲代孕乙", settings, "ai")
	if got != "甲乙" {
		t.Fatalf("got %q", got)
	}
}

func TestSensitiveLegacyEditorPatchUpdatesCanonicalCatalog(t *testing.T) {
	old := map[string]any{"groups": []any{map[string]any{"rules": []any{map[string]any{"find": "原词", "replace": "旧词"}}}}}
	next := map[string]any{"groups": []any{map[string]any{"rules": []any{map[string]any{"find": "原词", "replace": "新词"}}}}}
	settings := map[string]any{"knowledgeLibraries": map[string]any{"sensitive": old}}
	SyncKnowledgeConfigViews(settings, map[string]any{"appConfig": map[string]any{"sensitive": next}})
	got, _, _ := ApplySensitiveRulesForScopeWithDetails("原词", settings, "original")
	if got != "新词" {
		t.Fatalf("saved legacy edit not applied: %q", got)
	}
}
