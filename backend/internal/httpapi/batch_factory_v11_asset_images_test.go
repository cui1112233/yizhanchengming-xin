package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/png"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestBookAssetPrimaryImageSwapsWithoutDeletingPreviousVersion(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "林晚", Prompt: "女医生"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + book.ID + "/assets/" + asset.ID + "/images"

	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"url": "https://images.example/first.png", "source": "provider"})
	if first.Code != http.StatusCreated {
		t.Fatalf("first status=%d body=%s", first.Code, first.Body.String())
	}
	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"url": "https://images.example/second.png", "source": "provider"})
	if second.Code != http.StatusCreated {
		t.Fatalf("second status=%d body=%s", second.Code, second.Body.String())
	}
	secondImage := decodeBody[struct {
		Image struct {
			ID string `json:"id"`
		} `json:"image"`
	}](t, second).Image

	setPrimary := signedJSONRequest(t, api, now, "alice", http.MethodPut, path+"/"+secondImage.ID+"/primary", map[string]any{})
	if setPrimary.Code != http.StatusOK {
		t.Fatalf("set primary status=%d body=%s", setPrimary.Code, setPrimary.Body.String())
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	images := decodeBody[struct {
		Images []struct {
			ID        string `json:"id"`
			URL       string `json:"url"`
			IsPrimary bool   `json:"isPrimary"`
		} `json:"images"`
	}](t, listed).Images
	if len(images) != 2 {
		t.Fatalf("images=%+v", images)
	}
	var primary string
	for _, image := range images {
		if image.IsPrimary {
			primary = image.ID
		}
	}
	if primary != secondImage.ID {
		t.Fatalf("primary=%q images=%+v", primary, images)
	}
}

func TestBookAssetImageUploadStoresBytesAndServesOwnedVersion(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "林晚", Prompt: "女医生"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1<<30)})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + book.ID + "/assets/" + asset.ID + "/images"
	canvas := image.NewRGBA(image.Rect(0, 0, 900, 1200))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		t.Fatal(err)
	}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())

	uploaded := signedJSONRequest(t, api, now, "alice", http.MethodPost, path+"/upload", map[string]any{"dataUrl": dataURL})
	if uploaded.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
	}
	uploadedImage := decodeBody[struct {
		Image struct {
			ID           string `json:"id"`
			URL          string `json:"url"`
			ThumbnailURL string `json:"thumbnailUrl"`
			Source       string `json:"source"`
		} `json:"image"`
	}](t, uploaded).Image
	if uploadedImage.ID == "" || uploadedImage.URL == "" || uploadedImage.ThumbnailURL == "" || uploadedImage.Source != "upload" {
		t.Fatalf("image=%+v", uploadedImage)
	}

	thumbnail := signedJSONRequest(t, api, now, "alice", http.MethodGet, uploadedImage.ThumbnailURL, nil)
	if thumbnail.Code != http.StatusOK || thumbnail.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("thumbnail status=%d type=%q body=%s", thumbnail.Code, thumbnail.Header().Get("Content-Type"), thumbnail.Body.String())
	}
	if got := thumbnail.Header().Get("Cache-Control"); got != "private, max-age=604800, immutable" {
		t.Fatalf("thumbnail Cache-Control=%q", got)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(thumbnail.Body.Bytes()))
	if err != nil {
		t.Fatalf("decode thumbnail: %v", err)
	}
	if config.Width != 315 || config.Height != 420 {
		t.Fatalf("thumbnail size=%dx%d, want 315x420", config.Width, config.Height)
	}

	content := signedJSONRequest(t, api, now, "alice", http.MethodGet, uploadedImage.URL, nil)
	if content.Code != http.StatusOK || content.Header().Get("Content-Type") != "image/png" || content.Body.Len() == 0 {
		t.Fatalf("content status=%d type=%q bytes=%d", content.Code, content.Header().Get("Content-Type"), content.Body.Len())
	}
	if got := content.Header().Get("Cache-Control"); got != "private, max-age=604800, immutable" {
		t.Fatalf("Cache-Control=%q", got)
	}
	if got := content.Header().Get("ETag"); got == "" {
		t.Fatal("asset image content is missing ETag")
	}
}

func TestBookAssetImageProviderEndpointRejectsPretendUpload(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "林晚", Prompt: "女医生"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + book.ID + "/assets/" + asset.ID + "/images"

	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"url": "https://images.example/not-a-upload.png", "source": "upload"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
