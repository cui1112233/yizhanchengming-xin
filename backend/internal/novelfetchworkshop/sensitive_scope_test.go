package novelfetchworkshop

import "testing"

func TestSensitiveRuleGroupsHonorEnabledAndScope(t *testing.T) {
	settings := map[string]any{"appConfig": map[string]any{"sensitive": map[string]any{"groups": []any{
		map[string]any{"name": "原文组", "enabled": true, "apply_to_original": true, "apply_to_ai": false, "rules": []any{map[string]any{"find": "原词", "replace": "原替"}}},
		map[string]any{"name": "AI组", "enabled": true, "apply_to_original": false, "apply_to_ai": true, "rules": []any{map[string]any{"find": "AI词", "replace": "AI替"}}},
		map[string]any{"name": "关闭组", "enabled": false, "apply_to_original": true, "apply_to_ai": true, "rules": []any{map[string]any{"find": "禁用词", "replace": "不应出现"}}},
	}}}}

	original, count, _ := ApplySensitiveRulesForScopeWithDetails("原词 AI词 禁用词", settings, "original")
	if original != "原替 AI词 禁用词" || count != 1 {
		t.Fatalf("original scope = %q count=%d", original, count)
	}
	ai, count, _ := ApplySensitiveRulesForScopeWithDetails("原词 AI词 禁用词", settings, "ai")
	if ai != "原词 AI替 禁用词" || count != 1 {
		t.Fatalf("ai scope = %q count=%d", ai, count)
	}
}

func TestSensitiveRulesDeduplicateMirroredConfiguration(t *testing.T) {
	settings := map[string]any{
		"sensitive": map[string]any{"groups": []any{map[string]any{
			"enabled": true, "apply_to_original": true,
			"rules": []any{map[string]any{"find": "执法", "replace": "执法人员", "enabled": true}},
		}}},
		"knowledge": map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{
			"enabled": true, "apply_to_original": true,
			"rules": []any{map[string]any{"find": "执法", "replace": "执法人员", "enabled": true}},
		}}}},
	}
	got, count, items := ApplySensitiveRulesForScopeWithDetails("我去了一趟执法。", settings, "original")
	if got != "我去了一趟执法人员。" || count != 1 || len(items) != 1 {
		t.Fatalf("mirrored rule applied more than once: text=%q count=%d items=%#v", got, count, items)
	}
	if items[0]["start"] != 6 || items[0]["end"] != 7 {
		t.Fatalf("hit position was not persisted: %#v", items[0])
	}
}

func TestSensitiveRuleDoesNotReprocessItsOwnReplacement(t *testing.T) {
	settings := map[string]any{"sensitive_rules": []any{map[string]any{
		"find": "A", "replace": "A说明", "enabled": true,
	}}}
	got, count, _ := ApplySensitiveRulesForScopeWithDetails("A", settings, "original")
	if got != "A说明" || count != 1 {
		t.Fatalf("self-containing replacement was reprocessed: text=%q count=%d", got, count)
	}
}

func TestSensitiveRulesPreferLongerPhraseBeforeShortWord(t *testing.T) {
	settings := map[string]any{"sensitive_rules": []any{
		map[string]any{"find": "犯", "replace": "违规者", "enabled": true},
		map[string]any{"find": "侵犯", "replace": "越界伤害", "enabled": true},
	}}
	got, _, _ := ApplySensitiveRulesForScopeWithDetails("侵犯行为", settings, "original")
	if got != "越界伤害行为" {
		t.Fatalf("short rule damaged longer phrase: %q", got)
	}
}
