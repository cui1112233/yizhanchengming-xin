package workspace

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// HistoryItem contains only identifiers and safe summaries of persisted facts.
// There is no content, provider diagnostic, storage location or restore command.
type HistoryItem struct {
	ID           string     `json:"id"`
	Kind         string     `json:"kind"`
	Origin       string     `json:"origin"`
	SourceID     string     `json:"sourceId"`
	IntakeID     int64      `json:"intakeId"`
	ProjectID    int64      `json:"projectId"`
	BookID       int64      `json:"bookId"`
	BookRunID    int64      `json:"bookRunId"`
	Attempt      int64      `json:"attempt"`
	Revision     int64      `json:"revision"`
	Title        string     `json:"title"`
	ProjectName  string     `json:"projectName"`
	Status       string     `json:"status"`
	SourceStatus string     `json:"sourceStatus"`
	ErrorCode    string     `json:"errorCode"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	ArchivedAt   *time.Time `json:"archivedAt"`
	Href         string     `json:"href"`
	CanPreview   bool       `json:"canPreview"`
}
type HistoryQuery struct {
	UserID, TeamID            int64
	Elevated                  bool
	Page, Limit               int
	Q, Kind, Status, Archived string
}
type HistoryPage struct {
	Entries []HistoryItem `json:"entries"`
	Total   int           `json:"total"`
}

const historyKinds = "intake batch book_run script novel_panel tts shuihuo_image shuihuo_video video merge"
const historyStatuses = "completed pending queued scheduled running partial_failed failed retryable_failed cancelled skipped saved measured pending_executor unknown"

func enumContains(values, value string) bool {
	for _, v := range strings.Fields(values) {
		if v == value {
			return true
		}
	}
	return false
}

// ValidateHistoryQuery is shared by the HTTP adapter and the database reader.
func ValidateHistoryQuery(q HistoryQuery) error {
	if q.Page < 1 || q.Page > 2147483647 || q.Limit < 1 || q.Limit > 100 {
		return fmt.Errorf("invalid history pagination")
	}
	if utf8.RuneCountInString(q.Q) > 200 || (q.Kind != "" && !enumContains(historyKinds, q.Kind)) || (q.Status != "" && !enumContains(historyStatuses, q.Status)) || !enumContains("all active archived", q.Archived) {
		return fmt.Errorf("invalid history filter")
	}
	return nil
}
func normalizeHistoryStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "succeeded" {
		return "completed"
	}
	if enumContains(historyStatuses, status) {
		return status
	}
	return "unknown"
}

const historyScopedScope = `WITH visible_intakes AS (
 SELECT i.id,i.name,i.status,i.updated_at FROM intakes i WHERE EXISTS (
  SELECT 1 FROM auth_intake_ownership io WHERE io.intake_id=i.id
  AND (io.owner_user_id=? OR (? > 0 AND io.team_id IS NOT NULL AND io.team_id=?))
 )
), visible_projects AS (
 SELECT bp.id,bp.intake_id,bp.name,bp.updated_at,bp.archived_at FROM batch_projects bp JOIN visible_intakes vi ON vi.id=bp.intake_id
 WHERE EXISTS (SELECT 1 FROM auth_batch_project_ownership po WHERE po.batch_project_id=bp.id
 AND (po.owner_user_id=? OR (? > 0 AND po.team_id IS NOT NULL AND po.team_id=?)))
)`
const historyElevatedScope = `WITH visible_intakes AS (
 SELECT i.id,i.name,i.status,i.updated_at FROM intakes i
), visible_projects AS (
 SELECT bp.id,bp.intake_id,bp.name,bp.updated_at,bp.archived_at FROM batch_projects bp JOIN visible_intakes vi ON vi.id=bp.intake_id
)`

// All attempts/revisions are retained. A workspace is a fallback only when its
// current revision has no persisted history record. Linked video state is used
// only when both project and book match; only that valid video link deduplicates.
const historyProjection = `,
visible_books AS (
 SELECT b.id,b.title,vp.id AS project_id FROM books b JOIN visible_projects vp ON b.intake_id=vp.intake_id
), visible_book_runs AS (
 SELECT br.id,br.batch_project_id,br.book_id,br.attempt,br.status,br.updated_at FROM book_runs br JOIN visible_books b ON b.id=br.book_id AND b.project_id=br.batch_project_id
), visible_video AS (
 SELECT vpt.id,vpt.attempt,vpt.status,vpt.updated_at,vpj.batch_project_id,vpj.book_id FROM video_production_tasks vpt
 JOIN video_production_jobs vpj ON vpj.id=vpt.production_job_id
 JOIN visible_books b ON b.id=vpj.book_id AND b.project_id=vpj.batch_project_id
), visible_media AS (
 SELECT smt.id,smt.batch_project_id,smt.book_id,smt.media_kind,
 COALESCE(CASE WHEN linked_vpj.id IS NOT NULL THEN linked_vpt.status END,smt.status) AS status,
 GREATEST(smt.updated_at,COALESCE(CASE WHEN linked_vpj.id IS NOT NULL THEN linked_vpt.updated_at END,smt.updated_at)) AS updated_at,
 COALESCE(CASE WHEN linked_vpj.id IS NOT NULL THEN linked_vpt.attempt END,0) AS attempt
 FROM shuihuo_media_tasks smt
 JOIN visible_books b ON b.id=smt.book_id AND b.project_id=smt.batch_project_id
 LEFT JOIN video_production_tasks linked_vpt ON linked_vpt.id=smt.production_task_id AND smt.media_kind='video'
 LEFT JOIN video_production_jobs linked_vpj ON linked_vpj.id=linked_vpt.production_job_id AND linked_vpj.batch_project_id=smt.batch_project_id AND linked_vpj.book_id=smt.book_id
 WHERE smt.media_kind IN ('image','audio','video')
), history_facts AS (
 SELECT 'intake' AS kind,'intake' AS origin,CAST(vi.id AS CHAR) AS source_id,vi.id AS intake_id,
 0 AS project_id,0 AS book_id,0 AS book_run_id,0 AS attempt,0 AS revision,vi.name AS title,'' AS project_name,vi.status AS raw_status,vi.updated_at,vp.archived_at
 FROM visible_intakes vi LEFT JOIN visible_projects vp ON vp.intake_id=vi.id
 UNION ALL
 SELECT 'batch','project',CAST(vp.id AS CHAR),vp.intake_id,vp.id,0,0,0,0,vp.name,vp.name,
 COALESCE((SELECT r.status FROM runs r WHERE r.batch_project_id=vp.id ORDER BY r.updated_at DESC,r.id DESC LIMIT 1),vi.status),
 GREATEST(vp.updated_at,vi.updated_at,COALESCE((SELECT r.updated_at FROM runs r WHERE r.batch_project_id=vp.id ORDER BY r.updated_at DESC,r.id DESC LIMIT 1),vp.updated_at)),vp.archived_at
 FROM visible_projects vp JOIN visible_intakes vi ON vi.id=vp.intake_id
 UNION ALL
 SELECT 'book_run','book_run',CAST(br.id AS CHAR),vp.intake_id,vp.id,br.book_id,br.id,br.attempt,0,b.title,vp.name,br.status,br.updated_at,vp.archived_at
 FROM visible_book_runs br JOIN visible_projects vp ON vp.id=br.batch_project_id JOIN visible_books b ON b.id=br.book_id AND b.project_id=vp.id
 UNION ALL
 SELECT 'script','stage',CAST(sr.id AS CHAR),vp.intake_id,vp.id,sr.book_id,br.id,sr.attempt,0,b.title,vp.name,sr.status,sr.updated_at,vp.archived_at
 FROM stage_runs sr JOIN visible_book_runs br ON br.id=sr.book_run_id AND sr.book_id=br.book_id
 JOIN visible_projects vp ON vp.id=br.batch_project_id JOIN visible_books b ON b.id=sr.book_id AND b.project_id=vp.id WHERE sr.stage='SCRIPT'
 UNION ALL
 SELECT 'novel_panel','revision',nh.id,vp.intake_id,vp.id,0,0,0,nh.revision,vp.name,vp.name,'saved',nh.created_at,vp.archived_at
 FROM novel_panel_history nh JOIN visible_projects vp ON vp.id=nh.batch_project_id
 UNION ALL
 SELECT 'novel_panel','workspace',CAST(npw.batch_project_id AS CHAR),vp.intake_id,vp.id,0,0,0,npw.revision,vp.name,vp.name,'saved',npw.updated_at,vp.archived_at
 FROM novel_panel_workspaces npw JOIN visible_projects vp ON vp.id=npw.batch_project_id
 WHERE NOT EXISTS (SELECT 1 FROM novel_panel_history nh WHERE nh.batch_project_id=npw.batch_project_id AND nh.revision=npw.revision)
 UNION ALL
 SELECT 'tts','measurement',CAST(am.id AS CHAR),vp.intake_id,vp.id,am.book_id,0,0,0,b.title,vp.name,'measured',am.measured_at,vp.archived_at
 FROM audio_measurements am JOIN visible_projects vp ON vp.id=am.batch_project_id JOIN visible_books b ON b.id=am.book_id AND b.project_id=vp.id
 UNION ALL
 SELECT CASE WHEN vm.media_kind='audio' THEN 'tts' WHEN vm.media_kind='image' THEN 'shuihuo_image' ELSE 'shuihuo_video' END,
 'media',CAST(vm.id AS CHAR),vp.intake_id,vp.id,vm.book_id,0,vm.attempt,0,b.title,vp.name,vm.status,vm.updated_at,vp.archived_at
 FROM visible_media vm JOIN visible_projects vp ON vp.id=vm.batch_project_id JOIN visible_books b ON b.id=vm.book_id AND b.project_id=vp.id
 UNION ALL
 SELECT 'video','production_task',CAST(vv.id AS CHAR),vp.intake_id,vp.id,vv.book_id,0,vv.attempt,0,b.title,vp.name,vv.status,vv.updated_at,vp.archived_at
 FROM visible_video vv JOIN visible_projects vp ON vp.id=vv.batch_project_id JOIN visible_books b ON b.id=vv.book_id AND b.project_id=vp.id
 WHERE NOT EXISTS (SELECT 1 FROM shuihuo_media_tasks smt WHERE smt.production_task_id=vv.id AND smt.batch_project_id=vv.batch_project_id AND smt.book_id=vv.book_id AND smt.media_kind='video')
 UNION ALL
 SELECT 'merge','merge_attempt',CAST(vma.id AS CHAR),vp.intake_id,vp.id,vmj.book_id,0,vma.attempt,0,b.title,vp.name,vma.status,vma.updated_at,vp.archived_at
 FROM video_merge_attempts vma JOIN video_merge_jobs vmj ON vmj.id=vma.merge_job_id
 JOIN visible_projects vp ON vp.id=vmj.batch_project_id JOIN visible_books b ON b.id=vmj.book_id AND b.project_id=vp.id
), safe_history AS (
 SELECT kind,origin,source_id,intake_id,project_id,book_id,book_run_id,attempt,revision,
 COALESCE(NULLIF(TRIM(title),''),CONCAT('记录 #',source_id)) AS title,project_name,updated_at,archived_at,
 CASE WHEN LOWER(TRIM(raw_status)) IN ('succeeded','completed','pending','queued','scheduled','running','partial_failed','failed','retryable_failed','cancelled','skipped','saved','measured','pending_executor') THEN LOWER(TRIM(raw_status)) ELSE 'unknown' END AS source_status
 FROM history_facts
), normalized_history AS (
 SELECT safe_history.*,CASE WHEN source_status='succeeded' THEN 'completed' ELSE source_status END AS status FROM safe_history
), filtered_history AS (
 SELECT * FROM normalized_history WHERE 1=1`

func (s *MySQLStore) ListHistory(ctx context.Context, q HistoryQuery) (HistoryPage, error) {
	out := HistoryPage{Entries: make([]HistoryItem, 0)}
	if err := ValidateHistoryQuery(q); err != nil {
		return out, err
	}
	if s == nil || s.db == nil {
		return out, fmt.Errorf("workspace history store unavailable")
	}
	if !q.Elevated && q.UserID <= 0 {
		return out, fmt.Errorf("history owner required")
	}
	scope := historyScopedScope
	args := []any{q.UserID, q.TeamID, q.TeamID, q.UserID, q.TeamID, q.TeamID}
	if q.Elevated {
		scope = historyElevatedScope
		args = nil
	}
	cte := scope + historyProjection
	if q.Q != "" {
		cte += " AND (LOWER(title) LIKE ? OR LOWER(project_name) LIKE ? OR CAST(book_id AS CHAR) LIKE ? OR CAST(project_id AS CHAR) LIKE ? OR source_id LIKE ?)"
		like := "%" + strings.ToLower(q.Q) + "%"
		args = append(args, like, like, like, like, like)
	}
	if q.Kind != "" {
		cte += " AND kind=?"
		args = append(args, q.Kind)
	}
	if q.Status != "" {
		cte += " AND status=?"
		args = append(args, q.Status)
	}
	if q.Archived == "active" {
		cte += " AND archived_at IS NULL"
	} else if q.Archived == "archived" {
		cte += " AND archived_at IS NOT NULL"
	}
	cte += "\n)\n"
	if err := s.db.QueryRowContext(ctx, cte+"SELECT COUNT(*) FROM filtered_history", args...).Scan(&out.Total); err != nil {
		return out, fmt.Errorf("count history: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, cte+`SELECT kind,origin,source_id,intake_id,project_id,book_id,book_run_id,attempt,revision,title,project_name,status,source_status,updated_at,archived_at FROM filtered_history ORDER BY updated_at DESC,kind ASC,origin ASC,source_id DESC LIMIT ? OFFSET ?`, append(args, q.Limit, int64(q.Page-1)*int64(q.Limit))...)
	if err != nil {
		return out, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item HistoryItem
		var archived sql.NullTime
		if err := rows.Scan(&item.Kind, &item.Origin, &item.SourceID, &item.IntakeID, &item.ProjectID, &item.BookID, &item.BookRunID, &item.Attempt, &item.Revision, &item.Title, &item.ProjectName, &item.Status, &item.SourceStatus, &item.UpdatedAt, &archived); err != nil {
			return out, fmt.Errorf("scan history: %w", err)
		}
		item.Status = normalizeHistoryStatus(item.Status)
		item.SourceStatus = strings.ToLower(strings.TrimSpace(item.SourceStatus))
		if item.SourceStatus != "succeeded" && !enumContains(historyStatuses, item.SourceStatus) {
			item.SourceStatus = "unknown"
		}
		item.ID = item.Kind + ":" + item.Origin + ":" + item.SourceID
		if archived.Valid {
			item.ArchivedAt = &archived.Time
		}
		if item.Status == "failed" || item.Status == "retryable_failed" || item.Status == "partial_failed" {
			item.ErrorCode = "HISTORY_SOURCE_FAILED"
		}
		// Only Batch currently supports exact project navigation. Other sources remain
		// visible without implying content preview or historical restore support.
		if item.Kind == "batch" && item.ProjectID > 0 {
			item.Href = "/batch-factory?projectId=" + strconv.FormatInt(item.ProjectID, 10)
		}
		out.Entries = append(out.Entries, item)
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("iterate history: %w", err)
	}
	return out, nil
}
