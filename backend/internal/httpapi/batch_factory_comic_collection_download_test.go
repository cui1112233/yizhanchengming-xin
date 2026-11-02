package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestComicCollectionImageDownloadReturnsOrderedBookNamedZip(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "测试书", SourceText: "漫画原文。", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}, {Label: "第二页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	collection := createDownloadableComicCollection(t, store, files, batch)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/comic-collections/" + collection.ID + "/download"
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/zip" {
		t.Fatalf("content-type=%q", got)
	}
	if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, "attachment") || !strings.Contains(got, url.QueryEscape("测试书.zip")) {
		t.Fatalf("content-disposition=%q", got)
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.File) != 2 || archive.File[0].Name != "测试书-001.png" || archive.File[1].Name != "测试书-002.png" {
		got := make([]string, 0, len(archive.File))
		for _, entry := range archive.File {
			got = append(got, entry.Name)
		}
		t.Fatalf("archive entries=%v", got)
	}
}

func TestComicCollectionImageDownloadRejectsMissingPrimaryImage(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "测试书", SourceText: "漫画原文。", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, batch.Books[0].ID, batch.Books[0].Videos[0].ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
	if err != nil {
		t.Fatal(err)
	}
	image, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, batch.Books[0].ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: "missing-image.png", MediaType: "image/png", Source: "upload"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetPrimaryComicPageImage(context.Background(), "alice", batch.ID, batch.Books[0].ID, page.ID, image.ID); err != nil {
		t.Fatal(err)
	}
	collection, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil || len(missing) != 0 {
		t.Fatalf("collection=%+v missing=%v err=%v", collection, missing, err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/comic-collections/" + collection.ID + "/download"
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), "第 1 页") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func createDownloadableComicCollection(t *testing.T, store *batchfactoryv11.MemoryStore, files *localartifact.Store, batch batchfactoryv11.Batch) batchfactoryv11.ComicCollection {
	t.Helper()
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	for _, video := range book.Videos {
		page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
		if err != nil {
			t.Fatal(err)
		}
		artifact, err := files.SaveImage("comic-collection-download-"+page.ID, "image/png", bytes.NewReader(pngBytes))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: artifact.StorageRef, MediaType: "image/png", Source: "upload"}); err != nil {
			t.Fatal(err)
		}
	}
	collection, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID)
	if err != nil || len(missing) != 0 {
		t.Fatalf("collection=%+v missing=%v err=%v", collection, missing, err)
	}
	return collection
}
