package prompt

import (
	"path/filepath"
	"testing"
)

func TestPersonalStorePersistsOwnedPromptsAndUsage(t *testing.T) {
	store := NewPersonalStore(filepath.Join(t.TempDir(), "script-constraint-prompts.json"))
	name := "镜头风格"
	item, err := store.CreateOrSaveDraft("alice", "prefix", &name, "电影感，冷色调")
	if err != nil || item.ID == "" {
		t.Fatalf("CreateOrSaveDraft() = %+v, %v", item, err)
	}
	if err := store.MarkUsed("alice", []string{item.ID}); err != nil {
		t.Fatal(err)
	}
	items, err := store.List("alice", "prefix")
	if err != nil || len(items) != 1 || items[0].LastUsedAt == nil {
		t.Fatalf("List() = %+v, %v", items, err)
	}
	body := "更新后的风格"
	updated, found, err := store.Update("alice", item.ID, nil, &body)
	if err != nil || !found || updated.Body != body {
		t.Fatalf("Update() = %+v, %v, %v", updated, found, err)
	}
	removed, err := store.Remove("alice", item.ID)
	if err != nil || !removed {
		t.Fatalf("Remove() = %v, %v", removed, err)
	}
}

func TestPersonalStoreDoesNotExposeOtherOwners(t *testing.T) {
	store := NewPersonalStore(filepath.Join(t.TempDir(), "script-constraint-prompts.json"))
	name := "私有"
	item, err := store.CreateOrSaveDraft("alice", "negative", &name, "不要水印")
	if err != nil {
		t.Fatal(err)
	}
	if removed, err := store.Remove("bob", item.ID); err != nil || removed {
		t.Fatalf("Remove by another owner = %v, %v", removed, err)
	}
}
