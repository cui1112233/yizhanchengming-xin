package material

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type Importer interface {
	Import(context.Context, string, string, multipart.File, *multipart.FileHeader) (Material, worktask.Task, error)
}

type ListRepository interface {
	List(context.Context, string, string) ([]Material, error)
}
type FileRepository interface {
	Get(context.Context, string, string, string) (Material, error)
}
type OwnerResolver func(context.Context, *http.Request) (string, error)

type handler struct {
	importer   Importer
	repository ListRepository
	owner      OwnerResolver
}

func NewHandler(importer Importer, repository ListRepository, owner OwnerResolver) http.Handler {
	return &handler{importer: importer, repository: repository, owner: owner}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, materialID, fileRoute := materialFileIDs(r.URL.Path)
	if !fileRoute {
		var ok bool
		projectID, ok = materialProjectID(r.URL.Path)
		if !ok {
			writeError(w, http.StatusNotFound, "MATERIAL_ROUTE_NOT_FOUND", "素材接口不存在")
			return
		}
	}
	if h.owner == nil {
		writeError(w, http.StatusServiceUnavailable, "MATERIAL_STORE_UNAVAILABLE", "素材服务尚未连接")
		return
	}
	owner, err := h.owner(r.Context(), r)
	if err != nil || owner == "" {
		writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "请先从一战晟铭登录后再进入魔剪")
		return
	}
	if fileRoute {
		h.serveFile(w, r, owner, projectID, materialID)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if h.repository == nil {
			writeError(w, http.StatusServiceUnavailable, "MATERIAL_STORE_UNAVAILABLE", "素材服务尚未连接")
			return
		}
		items, listErr := h.repository.List(r.Context(), owner, projectID)
		if listErr != nil {
			writeError(w, http.StatusInternalServerError, "MATERIAL_LIST_FAILED", "素材列表读取失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"materials": items})
	case http.MethodPost:
		if h.importer == nil {
			writeError(w, http.StatusServiceUnavailable, "MATERIAL_STORE_UNAVAILABLE", "素材服务尚未连接")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 10<<30)
		if err = r.ParseMultipartForm(32 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "UPLOAD_INVALID", "素材上传内容无效")
			return
		}
		file, header, fileErr := r.FormFile("file")
		if fileErr != nil {
			writeError(w, http.StatusBadRequest, "FILE_REQUIRED", "请选择素材文件")
			return
		}
		defer file.Close()
		item, taskItem, importErr := h.importer.Import(r.Context(), owner, projectID, file, header)
		if importErr != nil {
			code := ErrorCode(importErr)
			status := http.StatusBadRequest
			if code == "FORMAT_NOT_SUPPORTED" {
				status = http.StatusUnsupportedMediaType
			}
			if code == "FILE_TOO_LARGE" {
				status = http.StatusRequestEntityTooLarge
			}
			if code == "FILE_PERMISSION_DENIED" || code == "DISK_SPACE_INSUFFICIENT" {
				status = http.StatusInsufficientStorage
			}
			if code == "" {
				code = "MATERIAL_IMPORT_FAILED"
			}
			writeError(w, status, code, importErr.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"material": item, "task": taskItem})
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "请求方式不允许")
	}
}

func (h *handler) serveFile(w http.ResponseWriter, r *http.Request, owner, projectID, materialID string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "请求方式不允许")
		return
	}
	repository, ok := h.repository.(FileRepository)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "MATERIAL_STORE_UNAVAILABLE", "素材服务尚未连接")
		return
	}
	item, err := repository.Get(r.Context(), owner, projectID, materialID)
	if err != nil || strings.TrimSpace(item.StorageKey) == "" {
		writeError(w, http.StatusNotFound, "MATERIAL_NOT_FOUND", "素材不存在")
		return
	}
	if item.Status != "ready" {
		writeError(w, http.StatusConflict, "MATERIAL_NOT_READY", "素材还在处理中")
		return
	}
	file, err := os.Open(item.StorageKey)
	if err != nil {
		writeError(w, http.StatusNotFound, "MATERIAL_FILE_NOT_FOUND", "素材文件不存在")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		writeError(w, http.StatusNotFound, "MATERIAL_FILE_NOT_FOUND", "素材文件不存在")
		return
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(item.OriginalName)))
	if contentType == "" {
		contentType = map[string]string{"video": "video/mp4", "audio": "audio/mpeg", "image": "image/jpeg"}[item.MediaType]
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", url.PathEscape(item.OriginalName)))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, item.OriginalName, info.ModTime(), file)
}

func materialProjectID(path string) (string, bool) {
	const prefix = "/api/magic-cut/projects/"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, "/materials") {
		return "", false
	}
	value := strings.TrimSuffix(strings.TrimPrefix(path, prefix), "/materials")
	if value == "" || strings.Contains(value, "/") || strings.Contains(value, "..") {
		return "", false
	}
	return value, true
}

func materialFileIDs(path string) (string, string, bool) {
	const prefix = "/api/magic-cut/projects/"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, "/file") {
		return "", "", false
	}
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(path, prefix), "/file"), "/")
	if len(parts) != 3 || parts[1] != "materials" || parts[0] == "" || parts[2] == "" || strings.Contains(parts[0], "..") || strings.Contains(parts[2], "..") {
		return "", "", false
	}
	return parts[0], parts[2], true
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "error": message})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil && !errors.Is(err, context.Canceled) {
		return
	}
}
