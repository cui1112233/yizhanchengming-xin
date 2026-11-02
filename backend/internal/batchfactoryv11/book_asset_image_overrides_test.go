package batchfactoryv11

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBookAssetImageOverrides(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{
		Title: "image overrides",
		Books: []CreateBookInput{{Title: "book"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	var overrideInput CreateBookAssetInput
	if err := json.Unmarshal([]byte(`{"kind":"character","name":"林晚","prompt":"女医生","imageModelId":" seedream-4.5 ","imageAspectRatio":"1:1"}`), &overrideInput); err != nil {
		t.Fatal(err)
	}
	overridden, err := store.CreateBookAsset(ctx, "alice", batch.ID, batch.Books[0].ID, overrideInput)
	if err != nil {
		t.Fatal(err)
	}
	assertBookAssetImageOverride(t, overridden, "seedream-4.5", "1:1")

	inherited, err := store.CreateBookAsset(ctx, "alice", batch.ID, batch.Books[0].ID, CreateBookAssetInput{
		Kind: "scene", Name: "医院走廊", Prompt: "深夜医院",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertBookAssetImageOverride(t, inherited, "", "")
}

func TestBookAssetImageOverridesRejectInvalidAspectRatio(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{
		Title: "image override validation",
		Books: []CreateBookInput{{Title: "book"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]

	var invalidCreate CreateBookAssetInput
	if err := json.Unmarshal([]byte(`{"kind":"character","name":"林晚","prompt":"女医生","imageAspectRatio":"4:3"}`), &invalidCreate); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, invalidCreate); err != ErrInvalid {
		t.Fatalf("CreateBookAsset() error = %v, want %v", err, ErrInvalid)
	}

	asset, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, CreateBookAssetInput{Kind: "character", Name: "顾遥", Prompt: "律师"})
	if err != nil {
		t.Fatal(err)
	}
	var invalidUpdate UpdateBookAssetInput
	if err := json.Unmarshal([]byte(`{"name":"顾遥","prompt":"律师","imageAspectRatio":"4:3","expectedRevision":1}`), &invalidUpdate); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpdateBookAsset(ctx, "alice", batch.ID, book.ID, asset.ID, invalidUpdate); err != ErrInvalid {
		t.Fatalf("UpdateBookAsset() error = %v, want %v", err, ErrInvalid)
	}
}

func assertBookAssetImageOverride(t *testing.T, asset BookAsset, wantModelID, wantAspectRatio string) {
	t.Helper()
	encoded, err := json.Marshal(asset)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	var gotModelID string
	if err := json.Unmarshal(payload["imageModelId"], &gotModelID); err != nil {
		t.Fatalf("imageModelId JSON = %s: %v; payload=%s", payload["imageModelId"], err, encoded)
	}
	if gotModelID != wantModelID {
		t.Fatalf("imageModelId = %q, want %q; payload=%s", gotModelID, wantModelID, encoded)
	}
	var gotAspectRatio string
	if err := json.Unmarshal(payload["imageAspectRatio"], &gotAspectRatio); err != nil {
		t.Fatalf("imageAspectRatio JSON = %s: %v; payload=%s", payload["imageAspectRatio"], err, encoded)
	}
	if gotAspectRatio != wantAspectRatio {
		t.Fatalf("imageAspectRatio = %q, want %q; payload=%s", gotAspectRatio, wantAspectRatio, encoded)
	}
}
