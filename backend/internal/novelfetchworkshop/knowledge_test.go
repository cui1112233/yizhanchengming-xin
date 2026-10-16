package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestKnowledgeCRUDPersistsInConfig(t *testing.T) {
	settings := map[string]any{}
	item, err := SaveKnowledge(settings, "opening_phrases", map[string]any{"text": "她推开门，看见了自己的葬礼", "style": "现代", "tags": "悬疑,反转"})
	if err != nil {
		t.Fatal(err)
	}
	if item["id"] == "" || item["style"] != "现代通用" {
		t.Fatalf("unexpected normalized item: %#v", item)
	}
	library, err := ListKnowledge(settings, "opening_phrases")
	if err != nil || len(normalizeKnowledgeItems(library["items"])) != 1 {
		t.Fatalf("library not saved: %#v, err=%v", library, err)
	}
	removed, err := DeleteKnowledge(settings, "opening_phrases", item["id"].(string))
	if err != nil || !removed {
		t.Fatalf("delete failed: removed=%v err=%v", removed, err)
	}
}

func TestAnalyzeOpeningUsesProvider(t *testing.T) {
	items, err := AnalyzeOpening(context.Background(), rewriteProviderFunc(func(_ context.Context, request RewriteRequest) (string, error) {
		if request.UserPrompt == "" {
			t.Fatal("opening prompt is empty")
		}
		return "反转开头模板", nil
	}), "原始开头")
	if err != nil || len(items) != 1 || items[0]["text"] != "反转开头模板" {
		t.Fatalf("unexpected opening analysis: %#v err=%v", items, err)
	}
}

func TestOptimizeKnowledgePreservesIDAndUpdatesReturnedFields(t *testing.T) {
	settings := map[string]any{}
	item, err := SaveKnowledge(settings, "opening_phrases", map[string]any{"id": "opening_001", "text": "她回来了", "style": "现代通用", "tags": []any{"反转"}, "source": "旧"})
	if err != nil {
		t.Fatal(err)
	}
	optimized, err := OptimizeKnowledge(context.Background(), settings, "opening_phrases", item["id"].(string), rewriteProviderFunc(func(_ context.Context, request RewriteRequest) (string, error) {
		if request.SystemPrompt == "" || request.UserPrompt == "" {
			t.Fatal("optimization prompt is incomplete")
		}
		return "```json\n{\"item\":{\"text\":\"她带着秘密回来了\",\"style\":\"现代悬疑\"}}\n```", nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if optimized["id"] != "opening_001" || optimized["text"] != "她带着秘密回来了" || optimized["style"] != "现代悬疑" {
		t.Fatalf("unexpected optimized item: %#v", optimized)
	}
}

func TestDefaultWorkshopSettingsIncludeImportedKnowledgeAndRules(t *testing.T) {
	settings := DefaultWorkshopSettings()
	app, _ := settings["appConfig"].(map[string]any)
	knowledge, _ := app["knowledge"].(map[string]any)
	if sensitive, ok := knowledge["sensitive"].(map[string]any); !ok || len(arrayMap(sensitive["groups"])) == 0 {
		t.Fatalf("sensitive rule catalog missing from visible knowledge: %#v", knowledge["sensitive"])
	}
	if analyzer, ok := knowledge["opening_analyzer"].(map[string]any); !ok || analyzer["name"] != "爆款开头拆解" {
		t.Fatalf("opening analyzer seed missing: %#v", knowledge["opening_analyzer"])
	}
	got := KnowledgeSummary(settings)
	want := map[string]int{
		"opening_phrases": 174, "rewrite_templates": 14, "high_imitation": 1,
		"sensitive_rules": 862, "symbol_rules": 17, "pair_fill_rules": 10,
		"chapter_exact_rules": 9, "chapter_inline_rules": 9,
	}
	for key, expected := range want {
		if got[key] != expected {
			t.Fatalf("%s = %d, want %d (summary=%#v)", key, got[key], expected, got)
		}
	}
}

func TestEnsureWorkshopDefaultsDoesNotLetEmptyLegacyPlaceholdersHideCatalog(t *testing.T) {
	settings := EnsureWorkshopDefaults(map[string]any{
		"knowledgeLibraries": map[string]any{
			"opening_phrases": map[string]any{"items": []any{}},
		},
	})
	library, err := ListKnowledge(settings, "opening_phrases")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(normalizeKnowledgeItems(library["items"])); got != 174 {
		t.Fatalf("opening phrases after empty placeholder merge = %d, want 174", got)
	}
}

func TestEnsureWorkshopDefaultsDoesNotLetStructuredEmptyLibraryHideCatalog(t *testing.T) {
	settings := EnsureWorkshopDefaults(map[string]any{
		"knowledgeLibraries": map[string]any{
			"opening_phrases": map[string]any{
				"source":     "go",
				"updated_at": "",
				"items":      []any{},
			},
		},
	})
	library, err := ListKnowledge(settings, "opening_phrases")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(normalizeKnowledgeItems(library["items"])); got != 174 {
		t.Fatalf("opening phrases after structured empty placeholder merge = %d, want 174", got)
	}
}

func TestKnowledgeWritesSynchronizeLegacyAppConfigView(t *testing.T) {
	settings := EnsureWorkshopDefaults(map[string]any{})
	if _, err := SaveKnowledge(settings, "opening_phrases", map[string]any{"id": "opening_sync", "text": "她推门就看见了遗书", "style": "现代悬疑"}); err != nil {
		t.Fatal(err)
	}
	app := settings["appConfig"].(map[string]any)
	knowledge := app["knowledge"].(map[string]any)
	opening := knowledge["opening_phrases"].(map[string]any)
	if got := len(normalizeKnowledgeItems(opening["items"])); got != 175 {
		t.Fatalf("legacy appConfig view has %d opening phrases, want 175", got)
	}
}

func TestSyncKnowledgeConfigViewsPreservesRuleReplacements(t *testing.T) {
	settings := EnsureWorkshopDefaults(map[string]any{})
	SyncKnowledgeConfigViews(settings, map[string]any{
		"appConfig": map[string]any{"knowledge": map[string]any{
			"layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"find": "脏词", "replace": "文明词", "enabled": true}}},
		}},
	})
	libraries := settings["knowledgeLibraries"].(map[string]any)
	layout := libraries["layout_rules"].(map[string]any)
	rules := layout["sensitive_rules"].([]any)
	if len(rules) != 1 || rules[0].(map[string]any)["replace"] != "文明词" {
		t.Fatalf("rule replacement was not persisted: %#v", layout)
	}
}

func TestEnsureSensitiveRuleReplacementsFillsSafeWords(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{
		map[string]any{"find": "绿帽子", "replace": ""},
		map[string]any{"find": "卧槽", "replace": ""},
		map[string]any{"find": "已有词", "replace": "自定义"},
	}}}}}
	if got := EnsureSensitiveRuleReplacements(settings); got != 2 {
		t.Fatalf("changed=%d, want 2", got)
	}
	rules := settings["sensitive"].(map[string]any)["groups"].([]any)[0].(map[string]any)["rules"].([]any)
	if rules[0].(map[string]any)["replace"] != "颜色帽子" || rules[1].(map[string]any)["replace"] != "惊讶感叹" || rules[2].(map[string]any)["replace"] != "自定义" {
		t.Fatalf("unexpected safe replacements: %#v", rules)
	}
}

func TestSafeSensitiveReplacementUsesMeaningfulNeutralLanguage(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{
		map[string]any{"find": "第三者", "replace": ""},
		map[string]any{"find": "轮奸", "replace": ""},
		map[string]any{"find": "未知脏词", "replace": ""},
	}}}}}
	EnsureSensitiveRuleReplacements(settings)
	rules := settings["sensitive"].(map[string]any)["groups"].([]any)[0].(map[string]any)["rules"].([]any)
	want := []string{"关系介入者", "多人伤害", "未知脏词"}
	for i, expected := range want {
		if got := rules[i].(map[string]any)["replace"]; got != expected {
			t.Fatalf("rule %d replacement=%q, want %q", i, got, expected)
		}
	}
}

func TestSafeSensitiveReplacementCoversInstitutionsAndBenignWords(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{
		map[string]any{"find": "警察", "replace": ""},
		map[string]any{"find": "法院", "replace": ""},
		map[string]any{"find": "玉米", "replace": ""},
	}}}}}
	EnsureSensitiveRuleReplacements(settings)
	rules := settings["sensitive"].(map[string]any)["groups"].([]any)[0].(map[string]any)["rules"].([]any)
	want := []string{"执法人员", "司法机构", "玉米"}
	for i, expected := range want {
		if got := rules[i].(map[string]any)["replace"]; got != expected {
			t.Fatalf("rule %d replacement=%q, want %q", i, got, expected)
		}
	}
}

func TestSafeSensitiveReplacementVariesByWordMeaning(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{
		map[string]any{"find": "奸夫", "replace": ""},
		map[string]any{"find": "贱人", "replace": ""},
		map[string]any{"find": "未知脏词", "replace": ""},
	}}}}}
	EnsureSensitiveRuleReplacements(settings)
	rules := settings["sensitive"].(map[string]any)["groups"].([]any)[0].(map[string]any)["rules"].([]any)
	if got := rules[0].(map[string]any)["replace"]; got == "亲密行为" || got == "中性说法" {
		t.Fatalf("奸夫 collapsed to generic wording: %q", got)
	}
	if got := rules[1].(map[string]any)["replace"]; got == "不当称呼" || got == "中性说法" {
		t.Fatalf("贱人 collapsed to generic wording: %q", got)
	}
	if got := rules[2].(map[string]any)["replace"]; got != "未知脏词" {
		t.Fatalf("unclassified word=%q, want readable pass-through", got)
	}
}

func TestSafeSensitiveReplacementUsesSceneReadablePhrases(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{
		map[string]any{"find": "激情视频", "replace": ""},
		map[string]any{"find": "开房", "replace": ""},
	}}}}}
	EnsureSensitiveRuleReplacements(settings)
	rules := settings["sensitive"].(map[string]any)["groups"].([]any)[0].(map[string]any)["rules"].([]any)
	want := []string{"凌晨快乐视频", "外面睡觉"}
	for i, expected := range want {
		if got := rules[i].(map[string]any)["replace"]; got != expected {
			t.Fatalf("rule %d replacement=%q, want %q", i, got, expected)
		}
	}
}

func TestSafeSensitiveReplacementPreviewSetIsStable(t *testing.T) {
	preview := map[string]string{
		"捉奸": "发现关系越界", "爬上床": "深夜留宿", "避孕套": "防护用品", "强上": "强行靠近",
		"一夜情": "短暂关系", "禁脔": "专属对象", "婊子": "失礼之人", "骚货": "轻浮称呼",
		"私会": "私下见面", "搞我闺蜜": "介入朋友关系", "拉皮条": "私下牵线", "床上技术": "亲密相处方式",
		"鬼混": "混乱交往", "与多人有染": "同时与多人交往", "潜规则": "隐性约定", "伟哥": "活力药物",
		"把妹": "主动结识女生", "约炮": "临时约会", "裸聊": "私密聊天", "激情视频": "凌晨快乐视频", "开房": "外面睡觉",
	}
	for find, want := range preview {
		if got := safeSensitiveReplacement(find); got != want {
			t.Fatalf("%s replacement=%q, want %q", find, got, want)
		}
	}
}

func TestApplySensitiveRulesDoesNotLoopWhenReplacementContainsFind(t *testing.T) {
	settings := map[string]any{"sensitive_rules": []any{
		map[string]any{"find": "我", "replace": "我的"},
	}}
	got, changes, items := ApplySensitiveRulesWithDetails("我和我", settings)
	if got != "我的和我的" || changes != 2 || len(items) != 2 {
		t.Fatalf("got=%q changes=%d items=%d, want one-pass replacement", got, changes, len(items))
	}
}
