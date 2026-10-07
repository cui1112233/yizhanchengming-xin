package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

// listWorkspaceHistory is a user-scoped read projection over existing project,
// book and generation facts; it deliberately owns no second history store.
func (h handler) listWorkspaceHistory(w http.ResponseWriter, r *http.Request) {
	if h.deps.Database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "HISTORY_UNAVAILABLE", "message": "历史记录暂不可用"})
		return
	}
	u, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	elevated := strings.EqualFold(u.Role, "admin") || strings.EqualFold(u.Role, "owner")
	where := " WHERE 1=1"
	args := []any{}
	if !elevated {
		where += " AND EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id=bp.id AND (o.owner_user_id=? OR (o.team_id IS NOT NULL AND o.team_id=?)))"
		args = append(args, u.ID, u.TeamID)
	}
	if q != "" {
		where += " AND (bp.name LIKE ? OR b.title LIKE ? OR CAST(b.id AS CHAR) LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if status != "" {
		where += " AND COALESCE(br.status,b.status)=?"
		args = append(args, status)
	}
	base := " FROM batch_projects bp JOIN books b ON b.intake_id=bp.intake_id LEFT JOIN book_runs br ON br.id=(SELECT r.id FROM book_runs r WHERE r.batch_project_id=bp.id AND r.book_id=b.id ORDER BY r.id DESC LIMIT 1)"
	var total int
	if err := h.deps.Database.QueryRowContext(r.Context(), "SELECT COUNT(*)"+base+where, args...).Scan(&total); err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "HISTORY_READ_FAILED", "读取项目历史失败", "history", "count", err)
		return
	}
	query := "SELECT bp.id,b.id,bp.name,b.title,COALESCE(br.status,b.status),COALESCE(br.updated_at,b.updated_at)" + base + where + " ORDER BY COALESCE(br.updated_at,b.updated_at) DESC,b.id DESC LIMIT ? OFFSET ?"
	rows, err := h.deps.Database.QueryContext(r.Context(), query, append(args, limit, (page-1)*limit)...)
	if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "HISTORY_READ_FAILED", "读取项目历史失败", "history", "list", err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var projectID, bookID int64
		var projectName, title, runStatus string
		var updated time.Time
		if err := rows.Scan(&projectID, &bookID, &projectName, &title, &runStatus, &updated); err != nil {
			h.writeServiceError(w, r, http.StatusInternalServerError, "HISTORY_READ_FAILED", "读取项目历史失败", "history", "scan", err)
			return
		}
		out = append(out, map[string]any{"id": strconv.FormatInt(projectID, 10) + "-" + strconv.FormatInt(bookID, 10), "projectId": projectID, "bookId": bookID, "projectName": projectName, "title": title, "status": runStatus, "updatedAt": updated})
	}
	if err := rows.Err(); err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "HISTORY_READ_FAILED", "读取项目历史失败", "history", "iterate", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": out, "page": page, "limit": limit, "total": total})
}
