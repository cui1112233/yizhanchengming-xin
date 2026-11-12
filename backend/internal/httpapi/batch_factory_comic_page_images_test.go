package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

const testComicPageImageDataURL = "data:image/png;base64,iVBORw0KGgoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

func newComicPageImagesTestBatch(t *testing.T, labels ...string) (*batchfactoryv11.MemoryStore, batchfactoryv11.Batch, *localartifact.Store, time.Time) {
	t.Helper()
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	videos := make([]batchfactoryv11.CreateVideoInput, 0, len(labels))
	for _, label := range labels {
		videos = append(videos, batchfactoryv11.CreateVideoInput{Label: label})
	}
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "测试书", SourceText: "漫画原文。", Videos: videos}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return store, batch, localartifact.NewStore(t.TempDir(), 1<<20), now
}

func TestComicPageImageUploadReplacesPrimaryByVideoID(t *testing.T) {
	store, batch, files, now := newComicPageImagesTestBatch(t, "第一页")
	book := batch.Books[0]
	videoID := book.Videos[0].ID
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, videoID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	imagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + videoID + "/images"

	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, imagesPath, map[string]any{"dataUrl": testComicPageImageDataURL})
	if first.Code != http.StatusCreated {
		t.Fatalf("upload status=%d body=%s", first.Code, first.Body.String())
	}
	firstPayload := decodeBody[struct {
		Page  batchfactoryv11.ComicPage      `json:"page"`
		Image batchfactoryv11.ComicPageImage `json:"image"`
	}](t, first)
	if firstPayload.Page.VideoID != videoID || !firstPayload.Image.IsPrimary {
		t.Fatalf("page=%+v image=%+v", firstPayload.Page, firstPayload.Image)
	}
	if !strings.HasSuffix(firstPayload.Image.URL, "/content") {
		t.Fatalf("image url=%q", firstPayload.Image.URL)
	}

	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, imagesPath, map[string]any{"dataUrl": testComicPageImageDataURL})
	if second.Code != http.StatusCreated {
		t.Fatalf("second upload status=%d body=%s", second.Code, second.Body.String())
	}
	secondImage := decodeBody[struct {
		Image batchfactoryv11.ComicPageImage `json:"image"`
	}](t, second).Image

	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/comic-pages", nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, listed).Pages
	if len(pages) != 1 || len(pages[0].Images) != 2 {
		t.Fatalf("pages=%+v", pages)
	}
	primaryCount := 0
	for _, image := range pages[0].Images {
		if image.IsPrimary {
			primaryCount++
			if image.ID != secondImage.ID {
				t.Fatalf("primary image=%s want %s", image.ID, secondImage.ID)
			}
		}
	}
	if primaryCount != 1 {
		t.Fatalf("primary count=%d", primaryCount)
	}
}

func TestComicPageImageUploadRejectsUnsupportedMediaType(t *testing.T) {
	store, batch, files, now := newComicPageImagesTestBatch(t, "第一页")
	book := batch.Books[0]
	videoID := book.Videos[0].ID
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, videoID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	imagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + videoID + "/images"
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, imagesPath, map[string]any{"dataUrl": "data:image/gif;base64,R0lGODlhAQABAAAAACw="})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestComicPageImagesZipKeepsPrimaryImagesAndManifestOnly(t *testing.T) {
	store, batch, files, now := newComicPageImagesTestBatch(t, "第一页", "第二页", "第三页")
	book := batch.Books[0]
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	// 只给第 1、3 页生成主图，第 2 页保持未出图；打包时应跳过缺图页。
	for _, videoIndex := range []int{0, 2} {
		video := book.Videos[videoIndex]
		page, upsertErr := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
		if upsertErr != nil {
			t.Fatal(upsertErr)
		}
		artifact, saveErr := files.SaveImage("comic-images-zip-"+page.ID, "image/png", bytes.NewReader(pngBytes))
		if saveErr != nil {
			t.Fatal(saveErr)
		}
		if _, createErr := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: artifact.StorageRef, MediaType: "image/png", Source: "upload"}); createErr != nil {
			t.Fatal(createErr)
		}
	}
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, book.Videos[1].ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"}); err != nil {
		t.Fatal(err)
	}

	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/comic-pages/images-zip", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/zip" {
		t.Fatalf("content-type=%q", got)
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(archive.File))
	for _, entry := range archive.File {
		names = append(names, entry.Name)
	}
	if len(names) != 3 || names[0] != "manifest.json" || names[1] != "漫画01.png" || names[2] != "漫画03.png" {
		t.Fatalf("entries=%v", names)
	}

	manifestFile, err := archive.File[0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer manifestFile.Close()
	var manifest comicPageImagesZipManifest
	if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.BookTitle != "测试书" || len(manifest.Pages) != 2 {
		t.Fatalf("manifest=%+v", manifest)
	}
	if manifest.Pages[0].PageIndex != 1 || manifest.Pages[0].VideoID != book.Videos[0].ID || manifest.Pages[0].FileName != "漫画01.png" {
		t.Fatalf("manifest page=%+v", manifest.Pages[0])
	}
	if manifest.Pages[1].PageIndex != 3 || manifest.Pages[1].VideoID != book.Videos[2].ID {
		t.Fatalf("manifest page=%+v", manifest.Pages[1])
	}
}

func TestComicPageImagesZipRejectsBookWithoutImages(t *testing.T) {
	store, batch, files, now := newComicPageImagesTestBatch(t, "第一页")
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/comic-pages/images-zip", nil)
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.HasPrefix(response.Body.String(), "PK") {
		t.Fatal("empty book must not return a zip")
	}
}

func TestComicPageDouyinPublishZipConvertsPagesToThreeFour(t *testing.T) {
	store, batch, files, now := newComicPageImagesTestBatch(t, "第一页", "第二页")
	book := batch.Books[0]
	source := image.NewRGBA(image.Rect(0, 0, 1024, 1536))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	for _, video := range book.Videos {
		page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
		if err != nil {
			t.Fatal(err)
		}
		artifact, err := files.SaveImage("douyin-"+page.ID, "image/png", bytes.NewReader(encoded.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: artifact.StorageRef, MediaType: "image/png", Source: "upload"}); err != nil {
			t.Fatal(err)
		}
	}

	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/comic-pages/douyin-zip", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/zip" {
		t.Fatalf("content-type=%q", got)
	}
	if disposition := response.Header().Get("Content-Disposition"); !strings.Contains(disposition, "douyin-publish.zip") {
		t.Fatalf("content-disposition=%q", disposition)
	}
	archive, err := zip.NewReader(bytes.NewReader(response.Body.Bytes()), int64(response.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	entries := make(map[string]*zip.File, len(archive.File))
	for _, entry := range archive.File {
		entries[entry.Name] = entry
	}
	for _, name := range []string{"图片/01.jpg", "图片/02.jpg", "发布文案.txt", "manifest.json"} {
		if entries[name] == nil {
			t.Fatalf("missing entry %q in %v", name, entryNames(archive.File))
		}
	}

	imageFile, err := entries["图片/01.jpg"].Open()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jpeg.Decode(imageFile)
	_ = imageFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	if bounds := decoded.Bounds(); bounds.Dx() != 1080 || bounds.Dy() != 1440 {
		t.Fatalf("douyin image size=%dx%d want 1080x1440", bounds.Dx(), bounds.Dy())
	}

	captionFile, err := entries["发布文案.txt"].Open()
	if err != nil {
		t.Fatal(err)
	}
	caption, err := io.ReadAll(captionFile)
	_ = captionFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(caption), "标题：测试书") || !strings.Contains(string(caption), "共 2 页") {
		t.Fatalf("caption=%q", string(caption))
	}

	manifestFile, err := entries["manifest.json"].Open()
	if err != nil {
		t.Fatal(err)
	}
	var manifest douyinPublishManifest
	decodeErr := json.NewDecoder(manifestFile).Decode(&manifest)
	_ = manifestFile.Close()
	if decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if manifest.PageCount != 2 || len(manifest.Images) != 2 {
		t.Fatalf("manifest=%+v", manifest)
	}
	if manifest.Images[0].FileName != "图片/01.jpg" || manifest.Images[0].SourceWidth != 1024 || manifest.Images[0].SourceHeight != 1536 {
		t.Fatalf("manifest image=%+v", manifest.Images[0])
	}
}

func TestComicPageDouyinPublishZipRejectsBookWithoutImages(t *testing.T) {
	store, batch, files, now := newComicPageImagesTestBatch(t, "第一页")
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/comic-pages/douyin-zip", nil)
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func entryNames(files []*zip.File) []string {
	names := make([]string, 0, len(files))
	for _, entry := range files {
		names = append(names, entry.Name)
	}
	return names
}
