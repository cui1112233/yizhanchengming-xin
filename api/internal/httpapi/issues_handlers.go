package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

type issueRow struct {
	ID, ProjectID, BookID                         int64
	Source, Status, ErrorCode, RequestID, Message string
	At                                            time.Time
}

// The task tables span migrations created with different MySQL default
// collations. Normalize every textual UNION column so an empty staging schema
// and an upgraded production schema both produce the same read-only projection.
const issueProjectionSQL = `SELECT br.id id,br.batch_project_id project_id,br.book_id book_id,_utf8mb4'book_run' COLLATE utf8mb4_unicode_ci source,CONVERT(br.status USING utf8mb4) COLLATE utf8mb4_unicode_ci status,CONVERT(COALESCE(br.error_code,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci error_code,CONVERT(COALESCE(br.request_id,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci request_id,CONVERT(COALESCE(br.error_message,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci error_message,COALESCE(br.finished_at,br.updated_at) at FROM book_runs br UNION ALL SELECT sr.id,sr2.batch_project_id,sr.book_id,_utf8mb4'stage_run' COLLATE utf8mb4_unicode_ci,CONVERT(sr.status USING utf8mb4) COLLATE utf8mb4_unicode_ci,_utf8mb4'' COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(sr.request_id,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(sr.error_message,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,COALESCE(sr.finished_at,sr.updated_at) FROM stage_runs sr JOIN book_runs sr2 ON sr2.id=sr.book_run_id UNION ALL SELECT vt.id,vj.batch_project_id,vj.book_id,_utf8mb4'video_task' COLLATE utf8mb4_unicode_ci,CONVERT(vt.status USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(vt.error_code,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(vt.request_id,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(vt.error_message,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,vt.updated_at FROM video_production_tasks vt JOIN video_production_jobs vj ON vj.id=vt.production_job_id UNION ALL SELECT mt.id,mt.batch_project_id,mt.book_id,_utf8mb4'media_task' COLLATE utf8mb4_unicode_ci,CONVERT(mt.status USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(mt.error_code,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(mt.request_id,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,CONVERT(COALESCE(mt.error_message,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci,mt.updated_at FROM shuihuo_media_tasks mt`

// listIssues is a read-only projection of durable execution facts. It never
// creates a second error-log store and scopes every row through project ownership.
func (h handler) listIssues(w http.ResponseWriter, r *http.Request) {
	if h.deps.Database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "ISSUES_UNAVAILABLE", "message": "问题记录暂不可用"})
		return
	}
	user, ok := authn.CurrentUser(r.Context())
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
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	elevated := strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
	where := ""
	args := []any{}
	if !elevated {
		where = " AND EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id=x.project_id AND (o.owner_user_id=? OR (o.team_id IS NOT NULL AND o.team_id=?)))"
		args = append(args, user.ID, user.TeamID)
	}
	filters := " WHERE 1=1"
	if source != "" {
		filters += " AND source=?"
		args = append(args, source)
	}
	if status != "" {
		filters += " AND status=?"
		args = append(args, status)
	}
	if search != "" {
		filters += " AND (message LIKE ? OR CAST(project_id AS CHAR) LIKE ? OR CAST(book_id AS CHAR) LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}
	predicate := filters + where + " AND (status IN ('failed','retryable_failed') OR error_message <> '')"
	countQuery := "SELECT COUNT(*) FROM (" + issueProjectionSQL + ") x" + predicate
	var total int
	if err := h.deps.Database.QueryRowContext(r.Context(), countQuery, args...).Scan(&total); err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "ISSUES_READ_FAILED", "读取问题记录失败", "issues", "count", err)
		return
	}
	query := "SELECT * FROM (" + issueProjectionSQL + ") x" + predicate + " ORDER BY at DESC,id DESC LIMIT ? OFFSET ?"
	queryArgs := append(append([]any{}, args...), limit, (page-1)*limit)
	rows, err := h.deps.Database.QueryContext(r.Context(), query, queryArgs...)
	if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "ISSUES_READ_FAILED", "读取问题记录失败", "issues", "list", err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var v issueRow
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.BookID, &v.Source, &v.Status, &v.ErrorCode, &v.RequestID, &v.Message, &v.At); err != nil {
			h.writeServiceError(w, r, 500, "ISSUES_READ_FAILED", "读取问题记录失败", "issues", "scan", err)
			return
		}
		out = append(out, map[string]any{"id": v.Source + "-" + strconv.FormatInt(v.ID, 10), "projectId": v.ProjectID, "bookId": v.BookID, "source": v.Source, "status": v.Status, "code": v.ErrorCode, "requestId": observability.SanitizeString(v.RequestID), "message": observability.SanitizeString(v.Message), "at": v.At})
	}
	if err := rows.Err(); err != nil {
		h.writeServiceError(w, r, 500, "ISSUES_READ_FAILED", "读取问题记录失败", "issues", "iterate", err)
		return
	}
	writeJSON(w, 200, map[string]any{"entries": out, "page": page, "limit": limit, "total": total})
}
