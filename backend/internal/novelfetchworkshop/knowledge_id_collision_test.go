package novelfetchworkshop

import "testing"

func TestSaveKnowledgeAfterDeleteDoesNotReplaceExistingItem(t *testing.T) {
	for _, kind := range []string{"high_imitation", "opening_phrases", "rewrite_templates"} {
		t.Run(kind, func(t *testing.T) {
			settings := map[string]any{}
			items := []map[string]any{}
			for _, text := range []string{"first", "second", "third"} {
				item, err := SaveKnowledge(settings, kind, map[string]any{"text": text, "name": text})
				if err != nil {
					t.Fatal(err)
				}
				items = append(items, item)
			}
			DeleteKnowledge(settings, kind, items[1]["id"].(string))
			created, err := SaveKnowledge(settings, kind, map[string]any{"text": "new", "name": "new"})
			if err != nil {
				t.Fatal(err)
			}
			if created["id"] == items[2]["id"] {
				t.Fatal("new item overwrote surviving item")
			}
			library, _ := ListKnowledge(settings, kind)
			if len(normalizeKnowledgeItems(library["items"])) != 3 {
				t.Fatalf("lost existing item: %#v", library)
			}
		})
	}
}
