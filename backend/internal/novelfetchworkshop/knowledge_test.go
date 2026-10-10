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
