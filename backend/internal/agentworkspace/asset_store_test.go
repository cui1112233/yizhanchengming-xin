package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryAssetStoreIsOwnerAndProjectScoped(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryAssetStore()
	asset := Asset{
		ID: "asset-1", ProjectID: "agent-1", Kind: "image", Name: "cover.png",
		ContentType: "image/png", ObjectKey: "agent-assets/abc/agent-1/asset-1.png",
		ByteSize: 12, CreatedAt: time.Date(2026, 10, 19, 2, 3, 4, 0, time.UTC),
	}
	if err := store.CreateAsset(ctx, "alice", asset); err != nil {
		t.Fatal(err)
	}
	items, err := store.ListAssets(ctx, "alice", "agent-1", "image")
	if err != nil || len(items) != 1 || items[0].ID != asset.ID {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if _, err := store.Asset(ctx, "bob", "agent-1", asset.ID); !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("cross-owner err=%v", err)
	}
	if _, err := store.Asset(ctx, "alice", "agent-2", asset.ID); !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("cross-project err=%v", err)
	}
}

func TestAgentAssetObjectKeyUsesSafeProjectAndAssetIDs(t *testing.T) {
	key, err := AgentAssetObjectKey("alice", "agent-1", "asset-1", ".PNG")
	if err != nil || key == "" || key[len(key)-4:] != ".png" {
		t.Fatalf("key=%q err=%v", key, err)
	}
	if _, err := AgentAssetObjectKey("alice", "../agent", "asset-1", "png"); !errors.Is(err, ErrInvalidAsset) {
		t.Fatalf("unsafe project err=%v", err)
	}
}
