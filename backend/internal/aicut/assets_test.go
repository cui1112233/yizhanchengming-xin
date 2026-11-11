package aicut

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStorePersistsProjectAssets(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 9, 16, 30, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-1", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	asset := Asset{ID: "asset-1", ProjectID: "cut-1", Kind: "video", Name: "source.mp4", ContentType: "video/mp4", ObjectKey: "aicut/alice/cut-1/asset-1.mp4", ByteSize: 1024, CreatedAt: now}
	if err := store.CreateAsset(ctx, "alice", asset); err != nil {
		t.Fatal(err)
	}
	items, err := store.ListAssets(ctx, "alice", "cut-1")
	if err != nil || len(items) != 1 || items[0].ID != asset.ID {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if items, err := store.ListAssets(ctx, "bob", "cut-1"); err == nil || items != nil {
		t.Fatalf("foreign items=%+v err=%v", items, err)
	}
}
