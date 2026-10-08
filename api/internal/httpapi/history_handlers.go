package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workspace"
)

func (h handler) listWorkspaceHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期", "request_id": requestIDFromRequest(r)})
		return
	}
	q := workspace.HistoryQuery{UserID: user.ID, TeamID: user.TeamID, Page: 1, Limit: 20, Q: strings.TrimSpace(r.URL.Query().Get("q")), Kind: strings.TrimSpace(r.URL.Query().Get("kind")), Status: strings.TrimSpace(r.URL.Query().Get("status")), Archived: "all"}
	switch strings.ToLower(strings.TrimSpace(user.Role)) {
	case "admin", "owner":
		q.Elevated = true
	}
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		q.Page, err = strconv.Atoi(value)
	}
	if err == nil {
		if value := r.URL.Query().Get("limit"); value != "" {
			q.Limit, err = strconv.Atoi(value)
		}
	}
	if value := r.URL.Query().Get("archived"); value != "" {
		q.Archived = value
	}
	if err != nil || workspace.ValidateHistoryQuery(q) != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "INVALID_HISTORY_QUERY", "message": "历史筛选或分页参数无效", "request_id": requestIDFromRequest(r)})
		return
	}
	if h.deps.WorkspaceHistory == nil {
		h.writeServiceError(w, r, http.StatusServiceUnavailable, "HISTORY_UNAVAILABLE", "历史记录暂不可用", "history", "list", nil)
		return
	}
	result, err := h.deps.WorkspaceHistory.ListHistory(r.Context(), q)
	if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "HISTORY_READ_FAILED", "读取历史记录失败", "history", "list", err)
		return
	}
	if result.Entries == nil {
		result.Entries = []workspace.HistoryItem{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": result.Entries, "page": q.Page, "limit": q.Limit, "total": result.Total})
}
