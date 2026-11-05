package httpapi

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func writeTestComicPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, canvas); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	if err := os.WriteFile(path, encoded.Bytes(), 0o644); err != nil {
		t.Fatalf("write png: %v", err)
	}
}

func TestEnsureComicThumbnailGeneratesAndCaches(t *testing.T) {
	dir := t.TempDir()
	files := localartifact.NewStore(dir, 1<<30)
	originalPath := filepath.Join(dir, "artifact-1.png")
	writeTestComicPNG(t, originalPath, 900, 1200)
	original, err := files.Open("artifact-1.png")
	if err != nil {
		t.Fatalf("open original: %v", err)
	}
	defer original.Close()
	info, err := original.Stat()
	if err != nil {
		t.Fatalf("stat original: %v", err)
	}
	record := batchfactoryv11.ComicPageImage{StorageRef: "artifact-1.png", MediaType: "image/png"}
	key := comicThumbnailCacheKey(record, info)
	path, err := ensureComicThumbnail(files, record, original, key)
	if err != nil {
		t.Fatalf("generate thumbnail: %v", err)
	}
	if filepath.Dir(path) != filepath.Join(dir, comicThumbnailDirName) {
		t.Fatalf("thumbnail should be cached under %s, got %s", comicThumbnailDirName, path)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open thumbnail: %v", err)
	}
	defer file.Close()
	decoded, _, err := image.DecodeConfig(file)
	if err != nil {
		t.Fatalf("decode thumbnail: %v", err)
	}
	if decoded.Width > comicThumbnailLongEdge || decoded.Height > comicThumbnailLongEdge {
		t.Fatalf("thumbnail %dx%d exceeds long edge %d", decoded.Width, decoded.Height, comicThumbnailLongEdge)
	}
	// 900x1200 按长边 420 缩放应得 315x420。
	if decoded.Width != 315 || decoded.Height != 420 {
		t.Fatalf("thumbnail size=%dx%d, want 315x420", decoded.Width, decoded.Height)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat cached thumbnail: %v", err)
	}
	cachedInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat cached thumbnail: %v", err)
	}
	if _, err := ensureComicThumbnail(files, record, original, key); err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	afterInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat cached thumbnail after second call: %v", err)
	}
	if !afterInfo.ModTime().Equal(cachedInfo.ModTime()) {
		t.Fatalf("cache should be reused, modtime changed %s -> %s", cachedInfo.ModTime(), afterInfo.ModTime())
	}
}

func TestComicThumbnailCacheKeyChangesWithOriginal(t *testing.T) {
	dir := t.TempDir()
	files := localartifact.NewStore(dir, 1<<30)
	originalPath := filepath.Join(dir, "artifact-2.png")
	writeTestComicPNG(t, originalPath, 300, 400)
	first, err := files.Open("artifact-2.png")
	if err != nil {
		t.Fatalf("open first: %v", err)
	}
	defer first.Close()
	firstInfo, err := first.Stat()
	if err != nil {
		t.Fatalf("stat first: %v", err)
	}
	record := batchfactoryv11.ComicPageImage{StorageRef: "artifact-2.png", MediaType: "image/png"}
	firstKey := comicThumbnailCacheKey(record, firstInfo)

	writeTestComicPNG(t, originalPath, 600, 800)
	second, err := files.Open("artifact-2.png")
	if err != nil {
		t.Fatalf("open second: %v", err)
	}
	defer second.Close()
	secondInfo, err := second.Stat()
	if err != nil {
		t.Fatalf("stat second: %v", err)
	}
	secondKey := comicThumbnailCacheKey(record, secondInfo)
	if firstKey == secondKey {
		t.Fatalf("cache key must change when original bytes change")
	}
}
