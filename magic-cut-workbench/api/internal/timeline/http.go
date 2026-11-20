package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Repository interface {
	Get(context.Context, string, string) (Timeline, error)
	Put(context.Context, string, string, Timeline, uint64) (Timeline, error)
}

type OwnerResolver func(context.Context, *http.Request) (string, error)
type handler struct {
	repository Repository
	owner      OwnerResolver
}

func NewHandler(repository Repository, owner OwnerResolver) http.Handler {
	return &handler{repository: repository, owner: owner}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectID, ok := timelineProjectID(r.URL.Path)
	if !ok {
		writeTimelineError(w, http.StatusNotFound, "TIMELINE_ROUTE_NOT_FOUND", "时间线接口不存在", nil)
		return
	}
	if h.repository == nil || h.owner == nil {
		writeTimelineError(w, http.StatusServiceUnavailable, "TIMELINE_STORE_UNAVAILABLE", "时间线服务尚未连接", nil)
		return
	}
	owner, err := h.owner(r.Context(), r)
	if err != nil || owner == "" {
		writeTimelineError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "请先从一战晟铭登录后再进入魔剪", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		value, getErr := h.repository.Get(r.Context(), owner, projectID)
		if getErr != nil {
			writeTimelineError(w, http.StatusInternalServerError, "TIMELINE_LOAD_FAILED", "时间线读取失败", nil)
			return
		}
		writeTimelineJSON(w, http.StatusOK, map[string]any{"timeline": value})
	case http.MethodPut:
		var input struct {
			ExpectedRevision uint64   `json:"expectedRevision"`
			Timeline         Timeline `json:"timeline"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&input); err != nil {
			writeTimelineError(w, http.StatusBadRequest, "TIMELINE_INPUT_INVALID", "时间线内容无效", nil)
			return
		}
		saved, putErr := h.repository.Put(r.Context(), owner, projectID, input.Timeline, input.ExpectedRevision)
		var conflict *RevisionConflictError
		if errors.As(putErr, &conflict) {
			writeTimelineError(w, http.StatusConflict, "TIMELINE_REVISION_CONFLICT", "工程已在其他页面修改，请确认后重试", map[string]any{"currentRevision": conflict.CurrentRevision})
			return
		}
		if putErr != nil {
			writeTimelineError(w, http.StatusBadRequest, "TIMELINE_SAVE_FAILED", putErr.Error(), nil)
			return
		}
		writeTimelineJSON(w, http.StatusOK, map[string]any{"timeline": saved})
	default:
		writeTimelineError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "请求方式不允许", nil)
	}
}

func timelineProjectID(path string) (string, bool) {
	const prefix = "/api/magic-cut/projects/"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, "/timeline") {
		return "", false
	}
	value := strings.TrimSuffix(strings.TrimPrefix(path, prefix), "/timeline")
	return value, value != "" && !strings.Contains(value, "/") && !strings.Contains(value, "..")
}

func writeTimelineError(w http.ResponseWriter, status int, code, message string, extra map[string]any) {
	payload := map[string]any{"code": code, "error": message}
	for key, value := range extra {
		payload[key] = value
	}
	writeTimelineJSON(w, status, payload)
}
func writeTimelineJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
