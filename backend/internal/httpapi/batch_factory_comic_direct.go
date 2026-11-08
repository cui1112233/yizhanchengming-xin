package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/n8comic"
)

// 合集直传：把一个素材文件夹当作一个合集直接上传到 N8，不创建分镜、不创建
// 漫画页。请求按 1-10 张分片，最后一片带 finalize=true 时创建合集。
const (
	comicDirectRoute         = "POST /api/batch-factory/v12/comic-direct/import"
	comicDirectMaxChunk      = 10
	comicDirectRecordVersion = 1
	comicDirectDefaultRecord = "data/comic-direct-collections.json"
)

type n8ComicDirectClient interface {
	Configured() bool
	UploadImages(context.Context, n8comic.UploadImagesRequest) ([]int64, error)
	CreateCarousel(context.Context, n8comic.CreateCarouselRequest) (int64, error)
}

type comicDirectPageInput struct {
	Ordinal int    `json:"ordinal"`
	DataURL string `json:"dataUrl"`
}

type comicDirectInput struct {
	MaterialID        string                 `json:"materialId"`
	BookTitle         string                 `json:"bookTitle"`
	ProducerAdminName string                 `json:"producerAdminName"`
	TotalPages        int                    `json:"totalPages"`
	Finalize          bool                   `json:"finalize"`
	Pages             []comicDirectPageInput `json:"pages"`
	WorkIDs           []int64                `json:"workIds"`
	AudioCategoryID   int64                  `json:"audioCategoryId"`
	AudioID           int64                  `json:"audioId"`
}

type comicDirectEntry struct {
	MaterialID        string        `json:"materialId"`
	BookTitle         string        `json:"bookTitle"`
	OwnerUsername     string        `json:"ownerUsername,omitempty"`
	ProducerAdminName string        `json:"producerAdminName"`
	TotalPages        int           `json:"totalPages"`
	Name              string        `json:"name,omitempty"`
	ImageIDs          map[int]int64 `json:"imageIds"`
	CarouselID        int64         `json:"carouselId,omitempty"`
	ImageCount        int           `json:"imageCount,omitempty"`
	CreatedAt         string        `json:"createdAt"`
	UpdatedAt         string        `json:"updatedAt"`
}

type comicDirectRecordFile struct {
	Version int                          `json:"version"`
	Entries map[string]*comicDirectEntry `json:"entries"`
}

type comicDirectRecordStore struct {
	path string
	mu   sync.Mutex
}

func (s *comicDirectRecordStore) loadLocked() (*comicDirectRecordFile, error) {
	file := &comicDirectRecordFile{Version: comicDirectRecordVersion, Entries: map[string]*comicDirectEntry{}}
	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return file, nil
		}
		return nil, err
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return file, nil
	}
	if err := json.Unmarshal(content, file); err != nil {
		return nil, fmt.Errorf("合集直传记录文件损坏：%w", err)
	}
	if file.Entries == nil {
		file.Entries = map[string]*comicDirectEntry{}
	}
	file.Version = comicDirectRecordVersion
	return file, nil
}

func (s *comicDirectRecordStore) saveLocked(file *comicDirectRecordFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	temp := s.path + ".tmp"
	if err := os.WriteFile(temp, encoded, 0o644); err != nil {
		return err
	}
	return os.Rename(temp, s.path)
}

type comicDirectPageResult struct {
	Ordinal int   `json:"ordinal"`
	ImageID int64 `json:"imageId"`
}

type comicDirectResponse struct {
	Key              string                  `json:"key"`
	MaterialID       string                  `json:"materialId"`
	BookTitle        string                  `json:"bookTitle"`
	Expected         int                     `json:"expected"`
	UploadedNow      int                     `json:"uploadedNow"`
	TotalUploaded    int                     `json:"totalUploaded"`
	Finalized        bool                    `json:"finalized"`
	AlreadyPublished bool                    `json:"alreadyPublished,omitempty"`
	Pages            []comicDirectPageResult `json:"pages"`
	CarouselID       int64                   `json:"carouselId,omitempty"`
	Name             string                  `json:"name,omitempty"`
}

type comicDirectHandler struct {
	client  n8ComicDirectClient
	records *comicDirectRecordStore
	now     func() time.Time
}

func registerComicDirectRoutes(mux *http.ServeMux, publisher *BatchFactoryN8ComicPublisher, recordPath string, now func() time.Time) {
	var client n8ComicDirectClient
	if publisher != nil {
		if direct, ok := publisher.Client.(n8ComicDirectClient); ok {
			client = direct
		}
	}
	if strings.TrimSpace(recordPath) == "" {
		recordPath = comicDirectDefaultRecord
	}
	if absolute, err := filepath.Abs(recordPath); err == nil {
		recordPath = absolute
	}
	if now == nil {
		now = time.Now
	}
	handler := &comicDirectHandler{
		client:  client,
		records: &comicDirectRecordStore{path: recordPath},
		now:     now,
	}
	mux.HandleFunc(comicDirectRoute, handler.handle)
}

func (h *comicDirectHandler) handle(w http.ResponseWriter, r *http.Request) {
	if h.client == nil || !h.client.Configured() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "N8 漫画接口未配置"})
		return
	}
	var input comicDirectInput
	if !decodeJSON(w, r, &input) {
		return
	}
	input.MaterialID = strings.TrimSpace(input.MaterialID)
	input.BookTitle = strings.TrimSpace(input.BookTitle)
	input.ProducerAdminName = strings.TrimSpace(input.ProducerAdminName)
	if input.MaterialID == "" || input.BookTitle == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "materialId 和 bookTitle 不能为空"})
		return
	}
	if input.ProducerAdminName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "producerAdminName 不能为空"})
		return
	}
	if input.TotalPages < 2 || input.TotalPages > 35 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "totalPages 必须在 2 到 35 之间"})
		return
	}
	if len(input.Pages) == 0 || len(input.Pages) > comicDirectMaxChunk {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("每片必须包含 1 到 %d 张图片", comicDirectMaxChunk)})
		return
	}
	seen := map[int]struct{}{}
	for _, page := range input.Pages {
		if page.Ordinal < 1 || page.Ordinal > input.TotalPages {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("页序号 %d 超出 1-%d 范围", page.Ordinal, input.TotalPages)})
			return
		}
		if _, exists := seen[page.Ordinal]; exists {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("页序号 %d 重复", page.Ordinal)})
			return
		}
		seen[page.Ordinal] = struct{}{}
	}

	owner := ""
	if identity, ok := BridgeIdentityFromContext(r.Context()); ok {
		owner = identity.Username
	}
	key := input.MaterialID + "|" + input.BookTitle

	h.records.mu.Lock()
	defer h.records.mu.Unlock()
	records, err := h.records.loadLocked()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	entry := records.Entries[key]
	if entry == nil {
		entry = &comicDirectEntry{
			MaterialID:        input.MaterialID,
			BookTitle:         input.BookTitle,
			OwnerUsername:     owner,
			ProducerAdminName: input.ProducerAdminName,
			TotalPages:        input.TotalPages,
			ImageIDs:          map[int]int64{},
			CreatedAt:         h.timestamp(),
		}
		records.Entries[key] = entry
	}
	if entry.TotalPages != input.TotalPages {
		writeJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("同一素材的页数不一致：记录为 %d，本次为 %d", entry.TotalPages, input.TotalPages)})
		return
	}
	entry.ProducerAdminName = input.ProducerAdminName
	entry.OwnerUsername = owner
	if entry.CarouselID > 0 {
		writeJSON(w, http.StatusOK, comicDirectResponse{
			Key: key, MaterialID: entry.MaterialID, BookTitle: entry.BookTitle,
			Expected: entry.TotalPages, UploadedNow: 0, TotalUploaded: len(entry.ImageIDs),
			Finalized: true, AlreadyPublished: true,
			Pages: comicDirectPages(entry), CarouselID: entry.CarouselID, Name: entry.Name,
		})
		return
	}

	ordered := append([]comicDirectPageInput(nil), input.Pages...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Ordinal < ordered[j].Ordinal })

	var images []n8comic.ImageInput
	var ordinals []int
	for _, page := range ordered {
		if entry.ImageIDs[page.Ordinal] > 0 {
			continue
		}
		_, payload, decodeErr := decodeImageDataURL(page.DataURL)
		if decodeErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("第 %d 页不是有效图片", page.Ordinal)})
			return
		}
		mediaType, ext, _, ok := comicDirectSniff(payload)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("第 %d 页不是有效图片", page.Ordinal)})
			return
		}
		images = append(images, n8comic.ImageInput{
			FileName:  fmt.Sprintf("%s-%03d%s", input.MaterialID, page.Ordinal, ext),
			MediaType: mediaType,
			Content:   payload,
		})
		ordinals = append(ordinals, page.Ordinal)
	}

	uploadedNow := 0
	if len(images) > 0 {
		imageIDs, uploadErr := h.client.UploadImages(r.Context(), n8comic.UploadImagesRequest{
			ProducerAdminName: input.ProducerAdminName,
			Images:            images,
		})
		for index, ordinal := range ordinals {
			if index < len(imageIDs) && imageIDs[index] > 0 {
				entry.ImageIDs[ordinal] = imageIDs[index]
				uploadedNow++
			}
		}
		entry.UpdatedAt = h.timestamp()
		if saveErr := h.records.saveLocked(records); saveErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存合集直传记录失败：" + saveErr.Error()})
			return
		}
		if uploadErr != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"error":         "N8 图片上传失败：" + uploadErr.Error(),
				"uploadedNow":   uploadedNow,
				"totalUploaded": len(entry.ImageIDs),
				"pages":         comicDirectPages(entry),
			})
			return
		}
	}

	if !input.Finalize {
		writeJSON(w, http.StatusOK, comicDirectResponse{
			Key: key, MaterialID: entry.MaterialID, BookTitle: entry.BookTitle,
			Expected: entry.TotalPages, UploadedNow: uploadedNow, TotalUploaded: len(entry.ImageIDs),
			Pages: comicDirectPages(entry),
		})
		return
	}

	missing := make([]string, 0)
	for ordinal := 1; ordinal <= entry.TotalPages; ordinal++ {
		if entry.ImageIDs[ordinal] <= 0 {
			missing = append(missing, fmt.Sprintf("%d", ordinal))
		}
	}
	if len(missing) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":   "图片尚未全部上传，缺少第 " + strings.Join(missing, "、") + " 页",
			"missing": missing,
		})
		return
	}
	if len(input.WorkIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "finalize 需要 workIds（N8 关联作品 ID）"})
		return
	}
	if input.AudioCategoryID <= 0 || input.AudioID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "finalize 需要大于 0 的 audioCategoryId 和 audioId"})
		return
	}
	name, err := n8ComicCollectionName(input.ProducerAdminName, input.BookTitle, h.now())
	if err != nil {
		writeStoreError(w, err)
		return
	}
	imageIDs := make([]int64, 0, entry.TotalPages)
	for ordinal := 1; ordinal <= entry.TotalPages; ordinal++ {
		imageIDs = append(imageIDs, entry.ImageIDs[ordinal])
	}
	carouselID, err := h.client.CreateCarousel(r.Context(), n8comic.CreateCarouselRequest{
		ProducerAdminName: input.ProducerAdminName,
		Name:              name,
		ImageIDs:          imageIDs,
		WorkIDs:           input.WorkIDs,
		AudioCategoryID:   input.AudioCategoryID,
		AudioID:           input.AudioID,
	})
	if err != nil {
		entry.UpdatedAt = h.timestamp()
		_ = h.records.saveLocked(records)
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "N8 创建合集失败：" + err.Error(), "name": name, "pages": comicDirectPages(entry)})
		return
	}
	entry.CarouselID = carouselID
	entry.Name = name
	entry.ImageCount = len(imageIDs)
	entry.UpdatedAt = h.timestamp()
	if err := h.records.saveLocked(records); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "合集已创建，但保存记录失败：" + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, comicDirectResponse{
		Key: key, MaterialID: entry.MaterialID, BookTitle: entry.BookTitle,
		Expected: entry.TotalPages, UploadedNow: uploadedNow, TotalUploaded: len(entry.ImageIDs),
		Finalized: true, Pages: comicDirectPages(entry), CarouselID: carouselID, Name: name,
	})
}

func (h *comicDirectHandler) timestamp() string {
	return h.now().In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05")
}

func comicDirectPages(entry *comicDirectEntry) []comicDirectPageResult {
	pages := make([]comicDirectPageResult, 0, len(entry.ImageIDs))
	for ordinal := 1; ordinal <= entry.TotalPages; ordinal++ {
		if imageID := entry.ImageIDs[ordinal]; imageID > 0 {
			pages = append(pages, comicDirectPageResult{Ordinal: ordinal, ImageID: imageID})
		}
	}
	return pages
}

// comicDirectSniff 按真实文件头判断类型。素材供应商提供的 .png 文件名
// 经常实际是 JPEG，不能信任扩展名。
func comicDirectSniff(payload []byte) (string, string, string, bool) {
	switch {
	case len(payload) >= 8 && bytes.Equal(payload[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "image/png", ".png", "png", true
	case len(payload) >= 3 && payload[0] == 0xFF && payload[1] == 0xD8 && payload[2] == 0xFF:
		return "image/jpeg", ".jpg", "jpeg", true
	case len(payload) >= 6 && (string(payload[:6]) == "GIF87a" || string(payload[:6]) == "GIF89a"):
		return "image/gif", ".gif", "gif", true
	case len(payload) >= 2 && payload[0] == 'B' && payload[1] == 'M':
		return "image/bmp", ".bmp", "bmp", true
	}
	return "", "", "", false
}
