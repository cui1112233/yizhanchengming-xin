package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestComicPageImageContentServesPrivateCacheValidators(t *testing.T) {
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
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, book.Videos[0].ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := files.SaveImage("comic-content-cache", "image/png", bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatal(err)
	}
	image, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: artifact.StorageRef, MediaType: "image/png", Source: "upload"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + page.ID + "/images/" + image.ID + "/content"

	full := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if full.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", full.Code, full.Body.String())
	}
	if full.Body.Len() != len(pngBytes) {
		t.Fatalf("body bytes=%d want %d", full.Body.Len(), len(pngBytes))
	}
	etag := strings.TrimSpace(full.Header().Get("ETag"))
	if etag == "" {
		t.Fatal("content response is missing ETag")
	}
	if cacheControl := full.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "private") || !strings.Contains(cacheControl, "max-age=604800") {
		t.Fatalf("Cache-Control=%q", cacheControl)
	}
	if got := full.Header().Get("Accept-Ranges"); got != "bytes" {
		t.Fatalf("Accept-Ranges=%q", got)
	}
	if got := full.Header().Get("Content-Length"); got == "" {
		t.Fatal("content response is missing Content-Length")
	}

	// 同一个 URL 携带 If-None-Match 二次请求必须命中 304，且不重传图片字节。
	revalidate := httptest.NewRequest(http.MethodGet, path, nil)
	revalidate.Header.Set("If-None-Match", etag)
	SignBridgeRequest(revalidate, "alice", false, now, "secret")
	revalidated := httptest.NewRecorder()
	api.ServeHTTP(revalidated, revalidate)
	if revalidated.Code != http.StatusNotModified || revalidated.Body.Len() != 0 {
		t.Fatalf("revalidate status=%d bytes=%d", revalidated.Code, revalidated.Body.Len())
	}

	// Range 仍然返回 206，断点续传不受缓存头影响。
	ranged := httptest.NewRequest(http.MethodGet, path, nil)
	ranged.Header.Set("Range", "bytes=0-9")
	SignBridgeRequest(ranged, "alice", false, now, "secret")
	partial := httptest.NewRecorder()
	api.ServeHTTP(partial, ranged)
	if partial.Code != http.StatusPartialContent {
		t.Fatalf("range status=%d", partial.Code)
	}
	if got := partial.Header().Get("ETag"); got != etag {
		t.Fatalf("range ETag=%q want %q", got, etag)
	}
	if partial.Body.Len() != 10 {
		t.Fatalf("range bytes=%d want 10", partial.Body.Len())
	}
	if got := partial.Header().Get("Content-Range"); got == "" {
		t.Fatal("range response is missing Content-Range")
	}
}
