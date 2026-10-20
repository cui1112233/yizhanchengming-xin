package agentworkspace

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
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

func TestLocalAssetObjectStorePersistsAndDeletesAssetBytes(t *testing.T) {
	root := t.TempDir()
	store := NewLocalAssetObjectStore(root)
	asset := Asset{ObjectKey: "agent-assets/owner/project/asset-1.png"}
	want := []byte("local-agent-image")

	if err := store.Put(context.Background(), asset, bytes.NewReader(want)); err != nil {
		t.Fatal(err)
	}
	body, err := store.Open(context.Background(), asset)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(body)
	_ = body.Close()
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("got=%q err=%v", got, err)
	}
	if err := store.Delete(context.Background(), asset); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Open(context.Background(), asset); !errors.Is(err, ErrAssetNotFound) {
		t.Fatalf("open after delete err=%v", err)
	}
	if _, err := filepath.Abs(root); err != nil {
		t.Fatal(err)
	}
}

func TestLocalAssetObjectStoreRejectsEscapingObjectKeys(t *testing.T) {
	store := NewLocalAssetObjectStore(t.TempDir())
	asset := Asset{ObjectKey: "../outside.png"}
	if err := store.Put(context.Background(), asset, bytes.NewReader([]byte("bad"))); !errors.Is(err, ErrInvalidAsset) {
		t.Fatalf("put err=%v", err)
	}
}
