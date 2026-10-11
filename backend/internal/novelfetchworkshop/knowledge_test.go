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
