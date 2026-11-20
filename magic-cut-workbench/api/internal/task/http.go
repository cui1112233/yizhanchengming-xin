package task

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type CenterStore interface {
	List(context.Context, string, int) ([]Task, error)
	Cancel(context.Context, string, string) error
	Retry(context.Context, string, string) error
}
type OwnerResolver func(context.Context, *http.Request) (string, error)
type ActionDelegate interface {
	CancelTask(context.Context, string, string) (bool, error)
	RetryTask(context.Context, string, string) (bool, error)
}
type handler struct {
	store   CenterStore
	owner   OwnerResolver
	actions ActionDelegate
}

func NewHandler(store CenterStore, owner OwnerResolver) http.Handler {
	return &handler{store: store, owner: owner}
}
func NewHandlerWithActions(store CenterStore, owner OwnerResolver, actions ActionDelegate) http.Handler {
	return &handler{store: store, owner: owner, actions: actions}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.store == nil || h.owner == nil {
		writeTaskError(w, http.StatusServiceUnavailable, "TASK_STORE_UNAVAILABLE", "任务中心尚未连接")
		return
	}
	owner, err := h.owner(r.Context(), r)
	if err != nil || owner == "" {
		writeTaskError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "请先从一战晟铭登录后再进入魔剪")
		return
	}
	if r.URL.Path == "/api/magic-cut/tasks" && r.Method == http.MethodGet {
		items, listErr := h.store.List(r.Context(), owner, 100)
		if listErr != nil {
			writeTaskError(w, http.StatusInternalServerError, "TASK_LIST_FAILED", "任务列表读取失败")
			return
		}
		writeTaskJSON(w, http.StatusOK, map[string]any{"tasks": items})
		return
	}
	id, action, ok := taskActionPath(r.URL.Path)
	if !ok || r.Method != http.MethodPost {
		writeTaskError(w, http.StatusNotFound, "TASK_ROUTE_NOT_FOUND", "任务接口不存在")
		return
	}
	if h.actions != nil {
		var handled bool
		if action == "cancel" {
			handled, err = h.actions.CancelTask(r.Context(), owner, id)
		} else {
			handled, err = h.actions.RetryTask(r.Context(), owner, id)
		}
		if err != nil {
			writeTaskError(w, http.StatusConflict, "TASK_ACTION_FAILED", err.Error())
			return
		}
		if handled {
			writeTaskJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
	}
	if action == "cancel" {
		err = h.store.Cancel(r.Context(), owner, id)
	} else {
		err = h.store.Retry(r.Context(), owner, id)
	}
	if err != nil {
		code := ErrorCode(err)
		if code == "" {
			code = "TASK_ACTION_FAILED"
		}
		writeTaskError(w, http.StatusConflict, code, err.Error())
		return
	}
	writeTaskJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func taskActionPath(path string) (string, string, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/magic-cut/tasks/"), "/")
	if len(parts) != 2 || parts[0] == "" || strings.Contains(parts[0], "..") || (parts[1] != "cancel" && parts[1] != "retry") {
		return "", "", false
	}
	return parts[0], parts[1], true
}
func writeTaskError(w http.ResponseWriter, status int, code, message string) {
	writeTaskJSON(w, status, map[string]string{"code": code, "error": message})
}
func writeTaskJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
