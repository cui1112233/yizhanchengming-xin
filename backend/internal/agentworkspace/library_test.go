package agentworkspace

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStoreLibraryKeepsOwnerScopedFoldersAndPreventsDuplicateSourceAssets(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	folder, err := store.CreateLibraryFolder(ctx, "alice", CreateLibraryFolderInput{ID: "folder-1", Name: "角色参考", CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateLibraryFolder(ctx, "bob", CreateLibraryFolderInput{ID: "folder-1", Name: "其他", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	item, err := store.SaveLibraryItem(ctx, "alice", SaveLibraryItemInput{
		ID: "item-1", FolderID: folder.ID, SourceProjectID: "project-source", SourceAssetID: "asset-source",
		Kind: "image", Name: "女主参考.png", Tags: []string{"女主", "古装"}, CreatedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if item.SourceProjectID != "project-source" || item.SourceAssetID != "asset-source" || len(item.Tags) != 2 {
		t.Fatalf("item=%+v", item)
	}
	if _, err := store.SaveLibraryItem(ctx, "alice", SaveLibraryItemInput{
		ID: "item-2", FolderID: folder.ID, SourceProjectID: "project-source", SourceAssetID: "asset-source",
		Kind: "image", Name: "重复", CreatedAt: now,
	}); err != ErrLibraryItemExists {
		t.Fatalf("duplicate err=%v", err)
	}
	folders, err := store.ListLibraryFolders(ctx, "alice")
	if err != nil || len(folders) != 1 || folders[0].ID != folder.ID {
		t.Fatalf("folders=%+v err=%v", folders, err)
	}
	items, err := store.ListLibraryItems(ctx, "alice", folder.ID, LibraryItemFilter{Kind: "image", Query: "女主"})
	if err != nil || len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	bobItems, err := store.ListLibraryItems(ctx, "bob", "folder-1", LibraryItemFilter{})
	if err != nil || len(bobItems) != 0 {
		t.Fatalf("bob items=%+v err=%v", bobItems, err)
	}
}

func TestMemoryStoreLibraryMovesItemWithoutChangingSourceReference(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 24, 12, 5, 0, 0, time.UTC)
	for _, input := range []CreateLibraryFolderInput{
		{ID: "folder-a", Name: "待整理", CreatedAt: now},
		{ID: "folder-b", Name: "已整理", CreatedAt: now},
	} {
		if _, err := store.CreateLibraryFolder(ctx, "alice", input); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.SaveLibraryItem(ctx, "alice", SaveLibraryItemInput{
		ID: "item-1", FolderID: "folder-a", SourceProjectID: "project-source", SourceAssetID: "asset-source",
		Kind: "video", Name: "片段.mp4", CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	moved, err := store.MoveLibraryItem(ctx, "alice", "item-1", "folder-b", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if moved.FolderID != "folder-b" || moved.SourceProjectID != "project-source" || moved.SourceAssetID != "asset-source" {
		t.Fatalf("moved=%+v", moved)
	}
}
