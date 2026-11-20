package export

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type Creator interface {
	Create(context.Context, string, CreateInput) (Export, worktask.Task, error)
}

type FileProvider interface {
	File(context.Context, string, string) (Export, error)
}

type OwnerResolver func(context.Context, *http.Request) (string, error)

type handler struct {
	creator Creator
	owner   OwnerResolver
}

func NewHandler(creator Creator, owner OwnerResolver) http.Handler {
	return &handler{creator: creator, owner: owner}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.creator == nil || h.owner == nil {
		writeError(w, http.StatusServiceUnavailable, "EXPORT_SERVICE_UNAVAILABLE", "导出服务尚未连接")
		return
	}
	owner, err := h.owner(r.Context(), r)
	if err != nil || owner == "" {
		writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "请先登录后再导出")
		return
	}
	if r.Method == http.MethodGet {
		id, ok := exportFileID(r.URL.Path)
		provider, supported := h.creator.(FileProvider)
		if !ok || !supported {
			writeError(w, http.StatusNotFound, "EXPORT_ROUTE_NOT_FOUND", "导出文件不存在")
			return
		}
		item, fileErr := provider.File(r.Context(), owner, id)
		if fileErr != nil || item.SHA256 == "" {
			writeError(w, http.StatusNotFound, "EXPORT_FILE_NOT_READY", "导出文件尚未生成")
			return
		}
		if _, statErr := os.Stat(item.StorageKey); statErr != nil {
			writeError(w, http.StatusNotFound, "EXPORT_FILE_MISSING", "导出文件已经丢失")
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(item.OutputName+".mp4"))
		http.ServeFile(w, r, item.StorageKey)
		return
	}
	if r.Method != http.MethodPost || r.URL.Path != "/api/magic-cut/exports" {
		writeError(w, http.StatusNotFound, "EXPORT_ROUTE_NOT_FOUND", "导出接口不存在")
		return
	}
	var input CreateInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "EXPORT_BODY_INVALID", "导出参数无法读取")
		return
	}
	created, queued, err := h.creator.Create(r.Context(), owner, input)
	if err != nil {
		status := http.StatusConflict
		if errors.Is(err, context.Canceled) {
			status = http.StatusRequestTimeout
		}
		code := ErrorCode(err)
		if code == "" {
			code = "EXPORT_CREATE_FAILED"
		}
		writeError(w, status, code, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"exportId": created.ID, "taskId": queued.ID, "status": queued.Status})
}

func exportFileID(path string) (string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/magic-cut/exports/"), "/")
	returnValue := ""
	if len(parts) == 2 && parts[0] != "" && parts[1] == "file" && !strings.Contains(parts[0], "..") {
		returnValue = parts[0]
	}
	return returnValue, returnValue != ""
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
