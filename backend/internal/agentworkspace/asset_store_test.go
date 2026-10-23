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

type testAgentReferenceProvider func(context.Context, Asset, time.Duration) (string, error)

func (fn testAgentReferenceProvider) ReferenceURL(ctx context.Context, asset Asset, ttl time.Duration) (string, error) {
	return fn(ctx, asset, ttl)
}

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

func TestAgentAudioAssetIsValidProjectMaterial(t *testing.T) {
	asset := Asset{
		ID: "audio-1", ProjectID: "agent-1", Kind: "audio", Name: "voice.mp3",
		ContentType: "audio/mpeg", ObjectKey: "agent-assets/abc/agent-1/audio-1.mp3",
		ByteSize: 12, CreatedAt: time.Date(2026, 10, 21, 2, 3, 4, 0, time.UTC),
	}
	if !ValidAssetKind("audio") {
		t.Fatal("audio must be a supported Agent asset kind")
	}
	if err := ValidateAsset(asset); err != nil {
		t.Fatalf("audio asset should be valid: %v", err)
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

func TestMigratingAssetObjectStoreCopiesLegacyLocalAssetBeforeSigningReference(t *testing.T) {
	ctx := context.Background()
	asset := Asset{ObjectKey: "agent-assets/owner/project/asset-1.png"}
	legacy := NewLocalAssetObjectStore(t.TempDir())
	primary := NewLocalAssetObjectStore(t.TempDir())
	if err := legacy.Put(ctx, asset, bytes.NewReader([]byte("legacy-image"))); err != nil {
		t.Fatal(err)
	}
	store := NewMigratingAssetObjectStore(primary, testAgentReferenceProvider(func(_ context.Context, got Asset, ttl time.Duration) (string, error) {
		if got.ObjectKey != asset.ObjectKey || ttl != 15*time.Minute {
			t.Fatalf("asset=%+v ttl=%s", got, ttl)
		}
		return "https://tos.example/agent-assets/asset-1.png?signature=test", nil
	}), legacy)

	url, err := store.ReferenceURL(ctx, asset, 15*time.Minute)
	if err != nil || url == "" {
		t.Fatalf("url=%q err=%v", url, err)
	}
	body, err := primary.Open(ctx, asset)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(body)
	_ = body.Close()
	if err != nil || string(got) != "legacy-image" {
		t.Fatalf("copied bytes=%q err=%v", got, err)
	}
}
