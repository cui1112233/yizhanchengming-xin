package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	// 流式写出不预计算整包长度。
	if got := response.Header().Get("Content-Length"); got != "" {
		t.Fatalf("streamed zip must not set Content-Length, got %q", got)
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
	// 缺图在写任何 ZIP 字节前就返回可读 JSON，不会给客户端半截压缩包。
	if got := response.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type=%q", got)
	}
	if strings.HasPrefix(response.Body.String(), "PK") {
		t.Fatal("missing image must not return a partial zip")
	}
}

func TestComicCollectionResponseIncludesThumbnailURL(t *testing.T) {
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
	createDownloadableComicCollection(t, store, files, batch)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/comic-collections"
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Collections []batchfactoryv11.ComicCollection `json:"collections"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Collections) != 1 || len(payload.Collections[0].Items) != 2 {
		t.Fatalf("collections=%d items=%d", len(payload.Collections), len(payload.Collections[0].Items))
	}
	for _, item := range payload.Collections[0].Items {
		if !strings.HasSuffix(item.ThumbnailURL, "/thumbnail") {
			t.Fatalf("item %d thumbnailUrl=%q", item.Ordinal, item.ThumbnailURL)
		}
		if !strings.HasSuffix(item.URL, "/content") {
			t.Fatalf("item %d url=%q", item.Ordinal, item.URL)
		}
	}
}

// countingWriter 记录 ZIP 包的写入次数与首块大小：真正的流式写出会在压缩
// 过程中分多次 Write，首块远小于最终整包；整包缓冲则只会在最后一次性 Write。
type countingWriter struct {
	writes     int
	firstChunk int
	buffer     bytes.Buffer
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.firstChunk == 0 {
		w.firstChunk = len(p)
	}
	return w.buffer.Write(p)
}

func TestComicZipStreamWritesIncrementally(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<30)
	plan := comicZipPlan{filename: "测试书.zip", markdown: "# 提示词\n"}
	// 直接写满 512KB 不可压缩数据，越过 zip.Writer 内部缓冲，确保能观察到
	// 真正的分块写出，而不是等到 Close 才一次性吐出整包。
	payload := make([]byte, 512<<10)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 3; index++ {
		name := "comic-zip-stream-" + string(rune('a'+index)) + ".png"
		if err := os.WriteFile(filepath.Join(files.Root(), name), payload, 0o644); err != nil {
			t.Fatal(err)
		}
		plan.entries = append(plan.entries, comicZipEntry{name: "漫画页/" + name, storageRef: name})
	}

	sink := &countingWriter{}
	if err := writeComicZipStream(context.Background(), sink, files, plan); err != nil {
		t.Fatal(err)
	}
	t.Logf("streamed zip: writes=%d firstChunk=%dB total=%dB", sink.writes, sink.firstChunk, sink.buffer.Len())
	if sink.writes < 4 {
		t.Fatalf("zip should stream in multiple writes, got %d", sink.writes)
	}
	// 首块在剩余条目压缩完成前就已写出：首块只占整包的一小部分，而不是
	// 等 Close 后一次性吐出全部 1.5MB。
	if sink.firstChunk == 0 || sink.firstChunk >= sink.buffer.Len()/2 {
		t.Fatalf("first chunk=%d total=%d; zip was buffered instead of streamed", sink.firstChunk, sink.buffer.Len())
	}
	archive, err := zip.NewReader(bytes.NewReader(sink.buffer.Bytes()), int64(sink.buffer.Len()))
	if err != nil {
		t.Fatalf("streamed zip is not readable: %v", err)
	}
	if len(archive.File) != 4 {
		t.Fatalf("archive entries=%d want 4", len(archive.File))
	}
	if archive.File[0].Name != "最终完整出图提示词.md" {
		t.Fatalf("first entry=%q", archive.File[0].Name)
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
