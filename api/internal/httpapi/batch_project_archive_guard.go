package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
)

// requireActiveBatchProject is placed after capability and ownership checks.
// It is a lifecycle guard only; it deliberately does not create a second
// authorization rule. Test-only unauthenticated handlers may omit the reader,
// while a configured authenticated application fails closed.
func (h handler) requireActiveBatchProject(pathKey string, next http.Handler) http.Handler {
	if h.deps.BatchProjectLifecycle == nil && h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.deps.BatchProjectLifecycle == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目状态校验暂不可用"})
			return
		}
		projectID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
		if err != nil || projectID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "BATCH_PROJECT_INVALID_REQUEST", "message": "批量项目 ID 无效"})
			return
		}
		if h.ensureBatchProjectActive(w, r, projectID) {
			next.ServeHTTP(w, r)
		}
	})
}

// requireActiveIntakeBatchProject closes the legacy Intake aliases that can
// mutate the one-to-one BatchProject (or its books/runs) after it is archived.
// An Intake without a BatchProject remains writable so the first project can
// still be created.
func (h handler) requireActiveIntakeBatchProject(pathKey string, next http.Handler) http.Handler {
	if h.deps.BatchProjectLifecycle == nil && h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.deps.BatchProjectLifecycle == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目状态校验暂不可用"})
			return
		}
		intakeID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
		if err != nil || intakeID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "INTAKE_INVALID_REQUEST", "message": "Intake ID 无效"})
			return
		}
		archived, err := h.deps.BatchProjectLifecycle.IsIntakeBatchProjectArchived(r.Context(), intakeID)
		switch {
		case err != nil:
			h.writeServiceError(w, r, http.StatusServiceUnavailable, "BATCH_PROJECT_POLICY_UNAVAILABLE", "项目状态校验暂不可用", "batch_project", "intake_archive_state", err)
		case archived:
			writeJSON(w, http.StatusConflict, map[string]any{"code": "BATCH_PROJECT_ARCHIVED", "message": "项目已归档，请先恢复后再修改"})
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func (h handler) ensureBatchProjectActive(w http.ResponseWriter, r *http.Request, projectID int64) bool {
	if h.deps.BatchProjectLifecycle == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目状态校验暂不可用"})
		return false
	}
	archived, err := h.deps.BatchProjectLifecycle.IsBatchProjectArchived(r.Context(), projectID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "BATCH_PROJECT_NOT_FOUND", "message": "批量项目不存在"})
	case err != nil:
		h.writeServiceError(w, r, http.StatusServiceUnavailable, "BATCH_PROJECT_POLICY_UNAVAILABLE", "项目状态校验暂不可用", "batch_project", "archive_state", err)
	case archived:
		writeJSON(w, http.StatusConflict, map[string]any{"code": "BATCH_PROJECT_ARCHIVED", "message": "项目已归档，请先恢复后再修改"})
	default:
		return true
	}
	return false
}
