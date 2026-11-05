package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

const (
	// 列表行缩略图最长边 420px：列表行宽度约 220–320px，420px 留出 2x 屏的余量，
	// 同时比 480px 少约三成像素量。漫画页文字在 420px 下仍可辨认。
	comicThumbnailLongEdge = 420
	// 缩略图统一转 JPEG。仓库内没有 WebP 编码器（golang.org/x/image 只提供解码），
	// JPEG 是标准库唯一可用的有损编码，质量 78 在 420px 下约 40–80KB。
	comicThumbnailQuality = 78
	comicThumbnailDirName = "comic-thumbnails"
	// 防解压炸弹：单张原图解码后不超过 4000 万像素。
	comicThumbnailMaxPixels = 40_000_000
)

// comicThumbnailBudget 限制同时缩放的图片数。16 行首屏会同时请求缩略图，
// 原图是 3MB 级 PNG，不设上限会把 CPU 打满。
var comicThumbnailBudget = make(chan struct{}, 4)

// comicThumbnailCacheKey 把原图身份和原图本身的版本一起纳入缓存键：
// storage_ref 在重新生成时会换新，文件大小/修改时间再兜底一次原地覆盖。
func comicThumbnailCacheKey(record batchfactoryv11.ComicPageImage, info os.FileInfo) string {
	sum := sha256.Sum256([]byte(record.StorageRef + "|" + record.MediaType + "|" +
		strconv.FormatInt(info.Size(), 10) + "|" +
		strconv.FormatInt(info.ModTime().UnixNano(), 10) + "|" +
		strconv.Itoa(comicThumbnailLongEdge) + "|" + strconv.Itoa(comicThumbnailQuality)))
	return hex.EncodeToString(sum[:])
}

// ensureComicThumbnail 按需生成并落盘缓存；命中缓存时不重复缩放。
// 返回的路径只含十六进制字符，可直接用于 ServeContent。
func ensureComicThumbnail(files *localartifact.Store, record batchfactoryv11.ComicPageImage, original *os.File, key string) (string, error) {
	dir := filepath.Join(files.Root(), comicThumbnailDirName)
	path := filepath.Join(dir, key[:32]+".jpg")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	comicThumbnailBudget <- struct{}{}
	defer func() { <-comicThumbnailBudget }()
	// 等待并发额度期间可能已有同键请求生成完成。
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	if _, err := original.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	source, _, err := image.Decode(original)
	if err != nil {
		return "", fmt.Errorf("decode comic thumbnail: %w", err)
	}
	bounds := source.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return "", fmt.Errorf("decode comic thumbnail: empty image")
	}
	if bounds.Dx()*bounds.Dy() > comicThumbnailMaxPixels {
		return "", fmt.Errorf("decode comic thumbnail: image too large (%dx%d)", bounds.Dx(), bounds.Dy())
	}

	width, height := bounds.Dx(), bounds.Dy()
	longEdge := width
	if height > longEdge {
		longEdge = height
	}
	if longEdge > comicThumbnailLongEdge {
		width = width * comicThumbnailLongEdge / longEdge
		height = height * comicThumbnailLongEdge / longEdge
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	// 先铺白底再缩放：JPEG 没有 alpha，透明区域直接编码会变黑。
	target := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(target, target.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	xdraw.ApproxBiLinear.Scale(target, target.Bounds(), source, bounds, xdraw.Over, nil)

	var encoded bytes.Buffer
	encoded.Grow(width * height / 6)
	if err := jpeg.Encode(&encoded, target, &jpeg.Options{Quality: comicThumbnailQuality}); err != nil {
		return "", fmt.Errorf("encode comic thumbnail: %w", err)
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	temp, err := os.CreateTemp(dir, ".thumb-*.tmp")
	if err != nil {
		return "", err
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = temp.Close()
			_ = os.Remove(tempName)
		}
	}()
	if _, err := temp.Write(encoded.Bytes()); err != nil {
		return "", err
	}
	if err := temp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tempName, path); err != nil {
		return "", err
	}
	committed = true
	return path, nil
}

// findComicPageImage 按页/图 ID 定位一条图片记录。缩略图和原图共用同一套归属校验。
func findComicPageImage(ctx context.Context, comicStore batchfactoryv11.ComicPageStore, owner, batchID, bookID, pageID, imageID string) (batchfactoryv11.ComicPageImage, error) {
	pages, err := comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return batchfactoryv11.ComicPageImage{}, err
	}
	for _, page := range pages {
		if page.ID != pageID {
			continue
		}
		for _, candidate := range page.Images {
			if candidate.ID == imageID {
				return candidate, nil
			}
		}
	}
	return batchfactoryv11.ComicPageImage{}, batchfactoryv11.ErrNotFound
}

// handleComicPageThumbnail 返回列表行用的小图。原图始终保留在原 /content 路由，
// 只有详情/下载/预览才取原图。
func handleComicPageThumbnail(comicStore batchfactoryv11.ComicPageStore, files *localartifact.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "comic image storage is unavailable"})
			return
		}
		imageRecord, err := findComicPageImage(r.Context(), comicStore, owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("pageId"), r.PathValue("imageId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if imageRecord.StorageRef == "" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "image bytes are not stored locally"})
			return
		}
		original, err := files.Open(imageRecord.StorageRef)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "image file not found"})
			return
		}
		defer original.Close()
		info, err := original.Stat()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image file is not readable"})
			return
		}
		key := comicThumbnailCacheKey(imageRecord, info)
		path, err := ensureComicThumbnail(files, imageRecord, original, key)
		if err != nil {
			// 生成失败时返回 500，前端会自动回落到原图，不会停在占位文案。
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "thumbnail generation failed"})
			return
		}
		thumbnail, err := os.Open(path)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "thumbnail is not readable"})
			return
		}
		defer thumbnail.Close()
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// 键里已经包含原图版本，原图变化会换键，因此可以长缓存。
		w.Header().Set("Cache-Control", "private, max-age=604800, immutable")
		w.Header().Set("Etag", `"`+key[:32]+`"`)
		http.ServeContent(w, r, key[:32]+".jpg", time.Time{}, thumbnail)
	}
}

// comicThumbnailURL 保留给响应层复用，避免路径拼接逻辑散落。
func comicThumbnailURL(image batchfactoryv11.ComicPageImage) string {
	if image.ID == "" || image.BatchID == "" || image.BookID == "" || image.ComicPageID == "" {
		return ""
	}
	return "/api/batch-factory/v12/batches/" + image.BatchID + "/books/" + image.BookID +
		"/comic-pages/" + image.ComicPageID + "/images/" + image.ID + "/thumbnail"
}
