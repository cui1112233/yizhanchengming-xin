package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestOptimizeKnowledgeUsesUnsavedDraftAndGoalWithoutChangingIdentity(t *testing.T) {
	settings := map[string]any{}
	_, err := SaveKnowledge(settings, "opening_phrases", map[string]any{"id": "opening_draft", "text": "已保存开头"})
	if err != nil {
		t.Fatal(err)
	}
	optimized, err := OptimizeKnowledge(context.Background(), settings, "opening_phrases", "opening_draft", rewriteProviderFunc(func(_ context.Context, request RewriteRequest) (string, error) {
		var payload map[string]any
		if err := json.Unmarshal([]byte(request.UserPrompt), &payload); err != nil {
			t.Fatal(err)
		}
		current := payload["current_item"].(map[string]any)
		if current["text"] != "尚未保存的修改" || current["id"] != "opening_draft" || payload["user_goal"] != "保留变量" {
			t.Fatalf("draft or goal missing: %#v", payload)
		}
		return `{"item":{"id":"wrong-id","text":"完成优化"}}`, nil
	}), map[string]any{"item": map[string]any{"id": "wrong-draft", "text": "尚未保存的修改"}, "goal": "保留变量"})
	if err != nil || optimized["id"] != "opening_draft" || optimized["text"] != "完成优化" {
		t.Fatalf("result=%#v err=%v", optimized, err)
	}
}

func TestOptimizeKnowledgeFailureDoesNotSaveDraft(t *testing.T) {
	settings := map[string]any{}
	_, err := SaveKnowledge(settings, "opening_phrases", map[string]any{"id": "opening_failure", "text": "原始内容"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = OptimizeKnowledge(context.Background(), settings, "opening_phrases", "opening_failure", rewriteProviderFunc(func(_ context.Context, _ RewriteRequest) (string, error) {
		return "", fmt.Errorf("test provider failure")
	}), map[string]any{"item": map[string]any{"text": "未保存修改"}})
	if err == nil {
		t.Fatal("expected provider failure")
	}
	library, _ := ListKnowledge(settings, "opening_phrases")
	if normalizeKnowledgeItems(library["items"])[0]["text"] != "原始内容" {
		t.Fatal("failed optimization saved draft")
	}
}

func TestOptimizeKnowledgeRequestsFieldsConsumedByEachCollection(t *testing.T) {
	for _, tc := range []struct{ kind, collection, field string }{
		{"rewrite_templates", "profiles", "user_prompt_template"},
		{"rewrite_templates", "temporary_instructions", "content"},
		{"high_imitation", "references", "reference_text"},
		{"high_imitation", "prompts", "system_prompt"},
		{"opening_phrases", "", "content"},
	} {
		t.Run(tc.kind+tc.collection, func(t *testing.T) {
			settings := map[string]any{}
			_, err := SaveKnowledge(settings, tc.kind, map[string]any{"id": "entry", "collection": tc.collection, "name": "资料", "content": "内容"})
			if err != nil {
				t.Fatal(err)
			}
			_, err = OptimizeKnowledge(context.Background(), settings, tc.kind, "entry", rewriteProviderFunc(func(_ context.Context, request RewriteRequest) (string, error) {
				var payload map[string]any
				if err := json.Unmarshal([]byte(request.UserPrompt), &payload); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, field := range payload["required_fields"].([]any) {
					if field == tc.field {
						found = true
					}
				}
				if !found {
					t.Fatalf("runtime field %s missing: %s", tc.field, request.UserPrompt)
				}
				return `{"item":{"name":"已优化"}}`, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

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
		return `{"items":[{"text":"反转开头模板"}]}`, nil
	}), "原始开头")
	if err != nil || len(items) != 1 || items[0]["text"] != "反转开头模板" {
		t.Fatalf("unexpected opening analysis: %#v err=%v", items, err)
	}
}

func TestRepeatedOpeningAnalysisDoesNotOverwriteExistingLibrary(t *testing.T) {
	settings := map[string]any{}
	for _, text := range []string{"第一次开头", "第二次开头"} {
		items, err := AnalyzeOpening(context.Background(), rewriteProviderFunc(func(_ context.Context, _ RewriteRequest) (string, error) {
			return `{"items":[{"id":"opening_001","text":"` + text + `"}]}`, nil
		}), "样本")
		if err != nil {
			t.Fatal(err)
		}
		if _, exists := items[0]["id"]; exists {
			t.Fatal("analysis must not reuse a saved id")
		}
		if _, err := SaveKnowledge(settings, "opening_phrases", items[0]); err != nil {
			t.Fatal(err)
		}
	}
	library, _ := ListKnowledge(settings, "opening_phrases")
	if len(normalizeKnowledgeItems(library["items"])) != 2 {
		t.Fatal("new analysis overwrote existing opening")
	}
}

func TestOpeningAnalysisRejectsRefusalAndNonStructuredOutput(t *testing.T) {
	for _, content := range []string{"抱歉，无法处理", `{"items":[]}`, `{"items":[{}]}`} {
		_, err := AnalyzeOpening(context.Background(), rewriteProviderFunc(func(_ context.Context, _ RewriteRequest) (string, error) { return content, nil }), "样本")
		if err == nil {
			t.Fatalf("accepted invalid analysis %s", content)
		}
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
	settings := EnsureWorkshopDefaults(DefaultWorkshopSettings())
	app, _ := settings["appConfig"].(map[string]any)
	knowledge, _ := app["knowledge"].(map[string]any)
	if sensitive, ok := knowledge["sensitive"].(map[string]any); !ok || len(arrayMap(sensitive["groups"])) == 0 {
		t.Fatalf("sensitive rule catalog missing from visible knowledge: %#v", knowledge["sensitive"])
	}
	if sensitive, ok := knowledge["sensitive"].(map[string]any); ok {
		for _, group := range arrayMap(sensitive["groups"]) {
			for _, rule := range arrayMap(group["rules"]) {
				find := strings.TrimSpace(fmt.Sprint(rule["find"]))
				replace := strings.TrimSpace(fmt.Sprint(rule["replace"]))
				if find != "" && replace == "" {
					t.Fatalf("sensitive rule replacement missing for %q", rule["find"])
				}
				if find != "" && replace == find && find == "家暴" {
					t.Fatalf("sensitive rule replacement still equals original for %q", find)
				}
			}
		}
	}
	if analyzer, ok := knowledge["opening_analyzer"].(map[string]any); !ok || analyzer["name"] != "爆款开头拆解" {
		t.Fatalf("opening analyzer seed missing: %#v", knowledge["opening_analyzer"])
	}
	got := KnowledgeSummary(settings)
	want := map[string]int{
		"opening_phrases": 174, "opening_styles": 13,
		"rewrite_templates": 14, "rewrite_profiles": 8, "temporary_instructions": 6,
		"high_imitation": 1, "high_imitation_prompts": 1, "high_imitation_references": 0,
		"sensitive_rules": 862, "symbol_rules": 17, "pair_fill_rules": 10,
		"chapter_exact_rules": 9, "chapter_inline_rules": 9,
	}
	for key, expected := range want {
		if got[key] != expected {
			t.Fatalf("%s = %d, want %d (summary=%#v)", key, got[key], expected, got)
		}
	}
}

func TestKnowledgeSubcollectionsStaySeparatedWhenSaved(t *testing.T) {
	settings := EnsureWorkshopDefaults(DefaultWorkshopSettings())
	item, err := SaveKnowledge(settings, "rewrite_templates", map[string]any{
		"id": "instruction_test", "name": "测试临时指令", "collection": "temporary_instructions", "content": "只用于测试分类归属", "enabled": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if item["collection"] != "temporary_instructions" {
		t.Fatalf("saved collection = %v, want temporary_instructions", item["collection"])
	}
	library, err := ListKnowledge(settings, "rewrite_templates")
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, current := range normalizeKnowledgeItems(library["items"]) {
		if current["id"] == "instruction_test" {
			found = current["collection"] == "temporary_instructions"
		}
	}
	if !found {
		t.Fatalf("saved temporary instruction missing from its own collection: %#v", library["items"])
	}
	if got := KnowledgeSummary(settings)["temporary_instructions"]; got != 7 {
		t.Fatalf("temporary_instructions = %d, want 7", got)
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

func TestEnsureSensitiveRuleReplacementsRepairsPersistedOriginalReplacement(t *testing.T) {
	settings := map[string]any{"sensitive": map[string]any{"groups": []any{map[string]any{"rules": []any{
		map[string]any{"find": "家暴", "replace": "家暴"},
	}}}}}
	if got := EnsureSensitiveRuleReplacements(settings); got != 1 {
		t.Fatalf("changed=%d, want 1", got)
	}
	rules := settings["sensitive"].(map[string]any)["groups"].([]any)[0].(map[string]any)["rules"].([]any)
	if got := rules[0].(map[string]any)["replace"]; got != "家庭伤害" {
		t.Fatalf("replacement=%q, want 家庭伤害", got)
	}
}

func TestSafeSensitiveReplacementCoversCommonSensitiveCategories(t *testing.T) {
	for _, item := range []struct {
		find string
		want string
	}{
		{find: "家暴", want: "家庭伤害"},
		{find: "龌龊", want: "不雅表达"},
		{find: "荡妇", want: "失礼称呼"},
		{find: "乱伦", want: "不当关系"},
		{find: "迷药", want: "特殊药物"},
		{find: "人贩子", want: "非法拐卖者"},
		{find: "玩了多少小明星", want: "与多位公众人物交往"},
		{find: "知三当三", want: "明知介入关系"},
		{find: "龟儿子", want: "失礼称呼"},
		{find: "狗男女", want: "关系失当者"},
	} {
		if got := safeSensitiveReplacement(item.find); got != item.want {
			t.Fatalf("%s replacement=%q, want %q", item.find, got, item.want)
		}
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
