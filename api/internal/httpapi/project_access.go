package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func (h handler) batchProjectAllowed(ctx context.Context, user authn.User, projectID int64) (bool, error) {
	if h.deps.BatchProjectAccess == nil {
		return false, errVideoAccessUnavailable
	}
	role := strings.ToLower(strings.TrimSpace(user.Role))
	elevated := role == "admin" || role == "owner"
	return h.deps.BatchProjectAccess.CanAccessBatchProject(ctx, projectID, user.ID, user.TeamID, elevated)
}

func (h handler) requireBatchProjectAccess(pathKey string, next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		projectID, err := strconv.ParseInt(r.PathValue(pathKey), 10, 64)
		if err != nil || projectID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "AUTH_INVALID_REQUEST", "message": "批量项目 ID 无效"})
			return
		}
		user, ok := authn.CurrentUser(r.Context())
		if !ok || user.ID <= 0 {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
		allowed, err := h.batchProjectAllowed(r.Context(), user, projectID)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
			return
		}
		if !allowed {
			// Hide direct-object existence from an authenticated caller that has no
			// project ownership/team access.
			writeJSON(w, http.StatusNotFound, map[string]any{"code": "BATCH_PROJECT_NOT_FOUND", "message": "批量项目不存在"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
