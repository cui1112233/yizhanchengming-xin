package project

import (
	"context"
	"encoding/json"
	"net/http"
)

type Repository interface {
	Create(context.Context, string, CreateInput) (Project, error)
	List(context.Context, string) ([]Project, error)
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
	if r.URL.Path != "/api/magic-cut/projects" {
		writeProjectError(w, http.StatusNotFound, "PROJECT_ROUTE_NOT_FOUND", "工程接口不存在")
		return
	}
	if h.repository == nil || h.owner == nil {
		writeProjectError(w, http.StatusServiceUnavailable, "PROJECT_STORE_UNAVAILABLE", "魔剪工程库尚未连接")
		return
	}
	owner, err := h.owner(r.Context(), r)
	if err != nil || owner == "" {
		writeProjectError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "请先从一战晟铭登录后再进入魔剪")
		return
	}

	switch r.Method {
	case http.MethodGet:
		projects, listErr := h.repository.List(r.Context(), owner)
		if listErr != nil {
			writeProjectError(w, http.StatusInternalServerError, "PROJECT_LIST_FAILED", "工程列表读取失败")
			return
		}
		writeProjectJSON(w, http.StatusOK, map[string]any{"projects": projects})
	case http.MethodPost:
		var input CreateInput
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&input); err != nil {
			writeProjectError(w, http.StatusBadRequest, "PROJECT_INPUT_INVALID", "工程参数不完整")
			return
		}
		created, createErr := h.repository.Create(r.Context(), owner, input)
		if createErr != nil {
			writeProjectError(w, http.StatusBadRequest, "PROJECT_CREATE_FAILED", createErr.Error())
			return
		}
		writeProjectJSON(w, http.StatusCreated, map[string]any{"project": created})
	default:
		writeProjectError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "请求方式不允许")
	}
}

func writeProjectError(w http.ResponseWriter, status int, code, message string) {
	writeProjectJSON(w, status, map[string]string{"code": code, "error": message})
}

func writeProjectJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
