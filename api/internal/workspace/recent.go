package workspace

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// RecentItem is the safe, user-visible projection of an existing workspace
// fact. It intentionally excludes provider, storage, prompt and error payloads.
type RecentItem struct {
	Kind      string    `json:"kind"`
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
	Href      string    `json:"href"`
}

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

const scopedVisibleProjects = `WITH visible_projects AS (
SELECT bp.id AS project_id,bp.intake_id,bp.name AS project_name,bp.created_at AS project_created_at,bp.updated_at AS project_updated_at,i.status AS intake_status,i.created_at AS intake_created_at,i.updated_at AS intake_updated_at
FROM batch_projects bp
JOIN intakes i ON i.id=bp.intake_id
JOIN auth_batch_project_ownership o ON o.batch_project_id=bp.id
WHERE o.owner_user_id=? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id=?)
)`

const elevatedVisibleProjects = `WITH visible_projects AS (
SELECT bp.id AS project_id,bp.intake_id,bp.name AS project_name,bp.created_at AS project_created_at,bp.updated_at AS project_updated_at,i.status AS intake_status,i.created_at AS intake_created_at,i.updated_at AS intake_updated_at
FROM batch_projects bp
JOIN intakes i ON i.id=bp.intake_id
)`

const recentProjectionQuery = `,
ranked_runs AS (
 SELECT r.id,r.batch_project_id,r.status,r.updated_at,ROW_NUMBER() OVER (PARTITION BY r.batch_project_id ORDER BY r.updated_at DESC,r.id DESC) AS recent_rank
 FROM runs r JOIN visible_projects vp ON vp.project_id=r.batch_project_id
),
ranked_book_runs AS (
 SELECT br.id,br.batch_project_id,br.book_id,br.status,br.updated_at,ROW_NUMBER() OVER (PARTITION BY br.batch_project_id,br.book_id ORDER BY br.updated_at DESC,br.id DESC) AS recent_rank
 FROM book_runs br
 JOIN visible_projects vp ON vp.project_id=br.batch_project_id
 JOIN books b ON b.id=br.book_id AND b.intake_id=vp.intake_id
),
ranked_stages AS (
 SELECT sr.id,rbr.batch_project_id,sr.book_id,sr.status,sr.updated_at,ROW_NUMBER() OVER (PARTITION BY rbr.batch_project_id,sr.book_id ORDER BY sr.updated_at DESC,sr.id DESC) AS recent_rank
 FROM stage_runs sr
 JOIN ranked_book_runs rbr ON rbr.id=sr.book_run_id AND rbr.recent_rank=1
 WHERE sr.stage='SCRIPT'
),
ranked_audio AS (
 SELECT am.id,am.batch_project_id,am.book_id,am.measured_at AS updated_at,ROW_NUMBER() OVER (PARTITION BY am.batch_project_id,am.book_id ORDER BY am.measured_at DESC,am.id DESC) AS recent_rank
 FROM audio_measurements am
 JOIN visible_projects vp ON vp.project_id=am.batch_project_id
 JOIN books b ON b.id=am.book_id AND b.intake_id=vp.intake_id
),
ranked_video AS (
 SELECT vpt.id,vpj.batch_project_id,vpj.book_id,vpt.status,vpt.updated_at,ROW_NUMBER() OVER (PARTITION BY vpj.batch_project_id,vpj.book_id ORDER BY vpt.updated_at DESC,vpt.id DESC) AS recent_rank
 FROM video_production_tasks vpt
 JOIN video_production_jobs vpj ON vpj.id=vpt.production_job_id
 JOIN visible_projects vp ON vp.project_id=vpj.batch_project_id
 JOIN books b ON b.id=vpj.book_id AND b.intake_id=vp.intake_id
),
ranked_media AS (
 SELECT smt.id,smt.batch_project_id,smt.book_id,smt.media_kind,COALESCE(vpt.status,smt.status) AS status,GREATEST(smt.updated_at,COALESCE(vpt.updated_at,smt.updated_at)) AS updated_at,ROW_NUMBER() OVER (PARTITION BY smt.batch_project_id,smt.book_id,smt.media_kind ORDER BY GREATEST(smt.updated_at,COALESCE(vpt.updated_at,smt.updated_at)) DESC,smt.id DESC) AS recent_rank
 FROM shuihuo_media_tasks smt
 JOIN visible_projects vp ON vp.project_id=smt.batch_project_id
 JOIN books b ON b.id=smt.book_id AND b.intake_id=vp.intake_id
 LEFT JOIN video_production_tasks vpt ON vpt.id=smt.production_task_id
 WHERE smt.media_kind IN ('image','audio','video')
),
recent_items AS (
 SELECT 'batch' AS kind,vp.project_id AS source_id,COALESCE(NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)) AS title,COALESCE(rr.status,vp.intake_status,'pending') AS status,GREATEST(vp.project_updated_at,vp.intake_updated_at,COALESCE(rr.updated_at,vp.project_updated_at)) AS updated_at,vp.project_id,0 AS book_id,10 AS kind_rank
 FROM visible_projects vp LEFT JOIN ranked_runs rr ON rr.batch_project_id=vp.project_id AND rr.recent_rank=1
 UNION ALL
 SELECT 'intake',vp.intake_id,COALESCE(NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)),vp.intake_status,vp.intake_updated_at,vp.project_id,0,20
 FROM visible_projects vp
 UNION ALL
 SELECT 'script',rs.id,COALESCE(NULLIF(TRIM(b.title),''),NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)),rs.status,rs.updated_at,vp.project_id,b.id,30
 FROM ranked_stages rs JOIN visible_projects vp ON vp.project_id=rs.batch_project_id JOIN books b ON b.id=rs.book_id AND b.intake_id=vp.intake_id WHERE rs.recent_rank=1
 UNION ALL
 SELECT 'novel_panel',npw.batch_project_id,COALESCE(NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)),'saved',npw.updated_at,vp.project_id,0,40
 FROM novel_panel_workspaces npw JOIN visible_projects vp ON vp.project_id=npw.batch_project_id
 UNION ALL
 SELECT 'tts',ra.id,COALESCE(NULLIF(TRIM(b.title),''),NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)),'measured',ra.updated_at,vp.project_id,b.id,50
 FROM ranked_audio ra JOIN visible_projects vp ON vp.project_id=ra.batch_project_id JOIN books b ON b.id=ra.book_id AND b.intake_id=vp.intake_id WHERE ra.recent_rank=1
 UNION ALL
 SELECT CASE WHEN rm.media_kind='audio' THEN 'tts' WHEN rm.media_kind='image' THEN 'shuihuo_image' ELSE 'shuihuo_video' END,rm.id,COALESCE(NULLIF(TRIM(b.title),''),NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)),rm.status,rm.updated_at,vp.project_id,b.id,CASE WHEN rm.media_kind='audio' THEN 51 WHEN rm.media_kind='image' THEN 60 ELSE 70 END
 FROM ranked_media rm JOIN visible_projects vp ON vp.project_id=rm.batch_project_id JOIN books b ON b.id=rm.book_id AND b.intake_id=vp.intake_id WHERE rm.recent_rank=1
 UNION ALL
 SELECT 'video',rv.id,COALESCE(NULLIF(TRIM(b.title),''),NULLIF(TRIM(vp.project_name),''),CONCAT('项目 #',vp.project_id)),rv.status,rv.updated_at,vp.project_id,b.id,71
 FROM ranked_video rv JOIN visible_projects vp ON vp.project_id=rv.batch_project_id JOIN books b ON b.id=rv.book_id AND b.intake_id=vp.intake_id
 WHERE rv.recent_rank=1 AND NOT EXISTS (SELECT 1 FROM shuihuo_media_tasks smt WHERE smt.production_task_id=rv.id)
)
SELECT kind,source_id,title,status,updated_at,project_id
FROM recent_items
ORDER BY updated_at DESC,kind_rank ASC,project_id DESC,book_id DESC,source_id DESC
LIMIT ?`

const scopedRecentQuery = scopedVisibleProjects + recentProjectionQuery
const elevatedRecentQuery = elevatedVisibleProjects + recentProjectionQuery

func (s *MySQLStore) ListRecent(ctx context.Context, userID, teamID int64, limit int) ([]RecentItem, error) {
	return s.listRecent(ctx, scopedRecentQuery, userID, teamID, teamID, limit)
}

func (s *MySQLStore) ListRecentElevated(ctx context.Context, limit int) ([]RecentItem, error) {
	return s.listRecent(ctx, elevatedRecentQuery, limit)
}

func (s *MySQLStore) listRecent(ctx context.Context, query string, args ...any) ([]RecentItem, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("workspace recent store unavailable")
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query workspace recent projection: %w", err)
	}
	defer rows.Close()

	items := make([]RecentItem, 0)
	for rows.Next() {
		var item RecentItem
		var sourceID, projectID int64
		if err := rows.Scan(&item.Kind, &sourceID, &item.Title, &item.Status, &item.UpdatedAt, &projectID); err != nil {
			return nil, fmt.Errorf("scan workspace recent projection: %w", err)
		}
		item.ID = item.Kind + ":" + strconv.FormatInt(sourceID, 10)
		item.Href = "/batch-factory?projectId=" + strconv.FormatInt(projectID, 10)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspace recent projection: %w", err)
	}
	return items, nil
}
