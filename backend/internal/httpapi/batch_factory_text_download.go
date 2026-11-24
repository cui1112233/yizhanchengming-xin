package httpapi

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"sort"
	"strings"
	"unicode"

	txt "qiantie/backend/internal/batchfactorytext"
)

// Downloads use the immutable collection manifest, never the current page selection.
func (a textAPI) downloadCollection(w http.ResponseWriter, r *http.Request, x textRequest) error {
	collection, err := a.repo.GetCollection(x, x.scope, r.PathValue("collectionId"))
	if err != nil {
		return err
	}
	if len(collection.Items) == 0 {
		return fmt.Errorf("%w：合集没有可下载的文字页", txt.ErrNotReady)
	}
	artifacts, err := a.repo.ListArtifacts(x, x.scope)
	if err != nil {
		return err
	}
	byID := map[string]txt.PageArtifact{}
	for _, art := range artifacts {
		byID[art.ID] = art
	}
	type downloadFile struct {
		name string
		file *os.File
	}
	var files []downloadFile
	defer func() {
		for _, entry := range files {
			_ = entry.file.Close()
		}
	}()
	title := strings.TrimSpace(strings.Map(func(c rune) rune {
		if unicode.IsControl(c) || strings.ContainsRune(`\/:*?"<>|`, c) {
			return '_'
		}
		return c
	}, x.book.Title))
	if title == "" {
		title = "文本合集"
	}
	items := append([]txt.CollectionItem(nil), collection.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].Ordinal < items[j].Ordinal })
	seen := map[int]bool{}
	for _, item := range items {
		art, ok := byID[item.ArtifactID]
		if !ok || art.Kind != "page" || art.PageID != item.PageID || art.StorageRef != item.StorageRef || art.SHA256 != item.SHA256 || art.MediaType != item.MediaType || art.ByteSize <= 0 || item.Ordinal < 1 || seen[item.Ordinal] {
			return fmt.Errorf("%w：合集第 %d 页记录不完整，请检查合集", txt.ErrNotReady, item.Ordinal)
		}
		seen[item.Ordinal] = true
		ext := map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/webp": "webp"}[item.MediaType]
		if ext == "" {
			return txt.ErrInvalid
		}
		f, err := a.files.OpenContext(x, item.StorageRef)
		if err != nil {
			return fmt.Errorf("%w：合集第 %d 页图片无法读取", txt.ErrNotReady, item.Ordinal)
		}
		files = append(files, downloadFile{fmt.Sprintf("%s-%03d.%s", title, item.Ordinal, ext), f})
		hash := sha256.New()
		size, err := io.Copy(hash, io.LimitReader(f, art.ByteSize+1))
		if err != nil || size != art.ByteSize || hex.EncodeToString(hash.Sum(nil)) != item.SHA256 {
			return fmt.Errorf("%w：合集第 %d 页图片不完整", txt.ErrNotReady, item.Ordinal)
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if err = x.Err(); err != nil {
			return err
		}
	}
	// All reads and checks finish before sending headers, so missing files cannot
	// masquerade as successful downloads. Keep handles open while streaming.
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": title + ".zip"}))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	writer := zip.NewWriter(w)
	for _, entry := range files {
		if x.Err() != nil {
			panic(http.ErrAbortHandler)
		}
		target, err := writer.CreateHeader(&zip.FileHeader{Name: entry.name, Method: zip.Store})
		if err == nil {
			_, err = io.Copy(target, entry.file)
		}
		if err != nil {
			panic(http.ErrAbortHandler)
		}
	}
	if err = writer.Close(); err != nil {
		panic(http.ErrAbortHandler)
	}
	return nil
}
