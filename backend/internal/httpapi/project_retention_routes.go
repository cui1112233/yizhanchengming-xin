package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/projectretention"
)

// RegisterProjectRetentionRoutes 暴露全站统一的保留状态、归档库、恢复与彻底删除。
func RegisterProjectRetentionRoutes(mux *http.ServeMux, auth BridgeAuth, service *projectretention.Service) {
	if mux == nil || service == nil {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}
	writeRetentionError := func(w http.ResponseWriter, err error) {
		switch {
		case errors.Is(err, projectretention.ErrPurging):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error(), "code": "PROJECT_PURGING"})
		case errors.Is(err, projectretention.ErrNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "作品不存在或未归档"})
		case errors.Is(err, projectretention.ErrInvalid):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "只有已归档作品可以执行该操作"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	handle("GET /api/project-retention", func(w http.ResponseWriter, req *http.Request, owner string) {
		items, err := service.Status(req.Context(), owner)
		if err != nil {
			writeRetentionError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})
	handle("GET /api/project-retention/archived", func(w http.ResponseWriter, req *http.Request, owner string) {
		items, err := service.Archived(req.Context(), owner)
		if err != nil {
			writeRetentionError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})
	handle("POST /api/project-retention/archived/{module}/{id}/restore", func(w http.ResponseWriter, req *http.Request, owner string) {
		module := projectretention.Module(strings.TrimSpace(req.PathValue("module")))
		if err := service.Restore(req.Context(), owner, module, strings.TrimSpace(req.PathValue("id"))); err != nil {
			writeRetentionError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	handle("DELETE /api/project-retention/archived/{module}/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		module := projectretention.Module(strings.TrimSpace(req.PathValue("module")))
		if err := service.Purge(req.Context(), owner, module, strings.TrimSpace(req.PathValue("id"))); err != nil {
			writeRetentionError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
}
