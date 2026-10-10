package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestBookAssetsCanBeCreatedAndReadBack(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "batch",
		Books: []batchfactoryv11.CreateBookInput{{Title: "book"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + book.ID + "/assets"

	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"kind": "character", "name": "林晚", "prompt": "二十五岁女医生，白大褂",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}

	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	response := decodeBody[struct {
		Assets []struct {
			ID     string `json:"id"`
			Kind   string `json:"kind"`
			Name   string `json:"name"`
			Prompt string `json:"prompt"`
		} `json:"assets"`
	}](t, listed)
	if len(response.Assets) != 1 || response.Assets[0].ID == "" || response.Assets[0].Kind != "character" || response.Assets[0].Name != "林晚" || response.Assets[0].Prompt != "二十五岁女医生，白大褂" {
		t.Fatalf("assets=%+v", response.Assets)
	}
}

func TestV12BookAssetsUseGoAlias(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/assets"
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"kind": "scene", "name": "急诊室", "prompt": "深夜医院"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "急诊室") {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
}

func TestDirectorAssetsBecomeDurableBookAssets(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	_, err = store.PersistDirectorRevision(context.Background(), "alice", book, batchfactoryv11.DirectorSnapshot{Mode: "original"}, "digest", "", batchfactoryv11.DirectorResult{
		Characters: []batchfactoryv11.NamedPrompt{{Name: "林晚", Prompt: "二十五岁女医生"}},
		Scenes:     []batchfactoryv11.NamedPrompt{{Name: "急诊室", Prompt: "深夜医院急诊室"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	assets, err := store.ListBookAssets(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 2 || assets[0].Source != "director" || assets[0].Revision != 1 {
		t.Fatalf("durable assets=%+v", assets)
	}
}

func TestBookAssetUpdateRequiresOwningBookAndCurrentRevision(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "scene", Name: "急诊室", Prompt: "深夜医院"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + book.ID + "/assets/" + asset.ID

	foreign := signedJSONRequest(t, api, now, "bob", http.MethodPatch, path, map[string]any{"name": "急诊室", "prompt": "篡改", "expectedRevision": asset.Revision})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign status=%d body=%s", foreign.Code, foreign.Body.String())
	}
	updated := signedJSONRequest(t, api, now, "alice", http.MethodPatch, path, map[string]any{"name": "急诊室", "prompt": "深夜医院急诊室，冷白光", "expectedRevision": asset.Revision})
	if updated.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updated.Code, updated.Body.String())
	}
	value := decodeBody[struct {
		Asset batchfactoryv11.BookAsset `json:"asset"`
	}](t, updated).Asset
	if value.Prompt != "深夜医院急诊室，冷白光" || value.Source != "manual" || value.Revision != asset.Revision+1 {
		t.Fatalf("asset=%+v", value)
	}
}
