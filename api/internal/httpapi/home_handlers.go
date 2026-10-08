package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workspace"
)

func (h handler) listWorkspaceRecent(w http.ResponseWriter, r *http.Request) {
	limit := 6
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 20 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "INVALID_LIMIT", "message": "limit 必须是 1 到 20 的整数"})
			return
		}
		limit = value
	}
	if h.deps.WorkspaceRecent == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "RECENT_UNAVAILABLE", "message": "最近创作暂不可用"})
		return
	}
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}

	var (
		items []workspace.RecentItem
		err   error
	)
	switch strings.ToLower(strings.TrimSpace(user.Role)) {
	case "admin", "owner":
		items, err = h.deps.WorkspaceRecent.ListRecentElevated(r.Context(), limit)
	default:
		items, err = h.deps.WorkspaceRecent.ListRecent(r.Context(), user.ID, user.TeamID, limit)
	}
	if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "RECENT_READ_FAILED", "读取最近创作失败", "workspace_recent", "list", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
