package projectretention

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const novelActivitySQL = `SELECT p.project_id, COALESCE(NULLIF(p.title,''),'未命名作品集'), GREATEST(
 p.updated_at,
 COALESCE((SELECT MAX(s.updated_at) FROM novel_sessions s WHERE s.owner_username=p.owner_username AND s.project_id=p.project_id),p.updated_at),
 COALESCE((SELECT MAX(c.updated_at) FROM novel_chapters c WHERE c.owner_username=p.owner_username AND c.project_id=p.project_id),p.updated_at),
 COALESCE((SELECT MAX(r.updated_at) FROM novel_runs r WHERE r.owner_username=p.owner_username AND r.project_id=p.project_id),p.updated_at),
 COALESCE((SELECT MAX(d.updated_at) FROM novel_content_records d WHERE d.owner_username=p.owner_username AND d.project_id=p.project_id),p.updated_at),
 COALESCE((SELECT MAX(m.updated_at) FROM novel_materials m WHERE m.owner_username=p.owner_username AND m.project_id=p.project_id),p.updated_at),
 COALESCE((SELECT MAX(n.updated_at) FROM novel_project_notes n WHERE n.owner_username=p.owner_username AND n.project_id=p.project_id),p.updated_at),
 COALESCE((SELECT MAX(settings.updated_at) FROM novel_project_settings settings WHERE settings.owner_username=p.owner_username AND settings.project_id=p.project_id),p.updated_at)
) FROM novel_projects p WHERE p.owner_username=? AND p.deleted_at IS NULL`

// All novel writers first lock novel_projects. Retention follows that same
// order; state checks and archiving cannot race a queued claim/final commit.
func (s *Service) ArchiveNovel(ctx context.Context, owner, id string) error {
	return s.archiveNovel(ctx, owner, id, false)
}
func (s *Service) archiveNovel(ctx context.Context, owner, id string, automatic bool) error {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var title string
	if e = tx.QueryRowContext(ctx, `SELECT title FROM novel_projects WHERE owner_username=? AND project_id=? AND deleted_at IS NULL AND status<>'purging' FOR UPDATE`, owner, id).Scan(&title); e != nil {
		return e
	}
	now := s.now().UTC()
	if automatic {
		var due int
		e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_retention WHERE owner_username=? AND module='novel_creation' AND project_id=? AND archived_at IS NULL AND retention_expires_at<=?`, owner, id, now).Scan(&due)
		if e != nil {
			return e
		}
		if due == 0 {
			return nil
		}
		var currentID, currentTitle string
		var activity time.Time
		if e = tx.QueryRowContext(ctx, novelActivitySQL+` AND p.project_id=?`, owner, id).Scan(&currentID, &currentTitle, &activity); e != nil {
			return e
		}
		expires := s.expiresAt(activity, s.days(owner))
		if expires.After(now) {
			if _, e = tx.ExecContext(ctx, `UPDATE project_retention SET last_activity_at=?,retention_expires_at=?,updated_at=? WHERE owner_username=? AND module='novel_creation' AND project_id=? AND archived_at IS NULL`, activity, expires, now, owner, id); e != nil {
				return e
			}
			return tx.Commit()
		}

	}
	var active int
	e = tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM novel_runs WHERE owner_username=? AND project_id=? AND status NOT IN ('succeeded','succeeded_with_warnings','failed','cancelled','paused','awaiting_confirmation')) + (SELECT COUNT(*) FROM novel_content_records WHERE owner_username=? AND project_id=? AND ((kind='document' AND status IN ('queued','parsing')) OR (kind IN ('document','export') AND status='uploading' AND lease_until>UTC_TIMESTAMP(6))))`, owner, id, owner, id).Scan(&active)
	if e != nil {
		return e
	}
	if active > 0 {
		if automatic {
			return nil
		}
		return fmt.Errorf("请先停止正在运行或等待派发的任务，等待资料上传解析结束")
	}
	if automatic {
		_, e = tx.ExecContext(ctx, `UPDATE project_retention SET archived_at=?,updated_at=? WHERE owner_username=? AND module='novel_creation' AND project_id=? AND archived_at IS NULL`, now, now, owner, id)
	} else {
		_, e = tx.ExecContext(ctx, `INSERT INTO project_retention (owner_username,module,project_id,title,last_activity_at,retention_expires_at,archived_at,manual_archived,created_at,updated_at) VALUES (?,'novel_creation',?,?,?,?,?,1,?,?) ON DUPLICATE KEY UPDATE title=VALUES(title),archived_at=COALESCE(archived_at,VALUES(archived_at)),manual_archived=1,updated_at=VALUES(updated_at)`, owner, id, title, now, s.expiresAt(now, s.days(owner)), now, now, now)
	}
	if e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, `UPDATE novel_projects SET status='archived' WHERE owner_username=? AND project_id=?`, owner, id); e != nil {
		return e
	}
	if e = pauseNovelAutomation(ctx, tx, owner, id); e != nil {
		return e
	}
	return tx.Commit()
}
func pauseNovelAutomation(ctx context.Context, tx *sql.Tx, owner, id string) error {
	for _, q := range []string{`UPDATE novel_schedules SET enabled=0,input_snapshot_json=JSON_SET(COALESCE(input_snapshot_json,JSON_OBJECT()),'$.automationState.status','paused','$.automationState.stoppedReason','作品已归档，恢复后请手动继续') WHERE owner_username=? AND project_id=?`, `UPDATE novel_batches SET status='paused',input_snapshot_json=JSON_SET(COALESCE(input_snapshot_json,JSON_OBJECT()),'$.automationState.status','paused','$.automationState.stoppedReason','作品已归档，恢复后请手动继续') WHERE owner_username=? AND project_id=? AND status IN ('running','paused')`} {
		if _, e := tx.ExecContext(ctx, q, owner, id); e != nil {
			return e
		}
	}
	return nil
}
func (s *Service) archiveDueNovel(ctx context.Context, owner string) error {
	now := s.now().UTC()
	rows, e := s.DB.QueryContext(ctx, `SELECT project_id FROM project_retention WHERE owner_username=? AND module='novel_creation' AND archived_at IS NULL AND retention_expires_at<=?`, owner, now)
	if e != nil {
		return e
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return e
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, id := range ids {
		if e = s.archiveNovel(ctx, owner, id, true); e != nil && !errors.Is(e, sql.ErrNoRows) {
			return e
		}
	}
	// A retention-days increase follows the existing automatic restore policy;
	// explicit manual archives remain archived. Automation stays paused.
	rows, e = s.DB.QueryContext(ctx, `SELECT project_id FROM project_retention WHERE owner_username=? AND module='novel_creation' AND archived_at IS NOT NULL AND manual_archived=0 AND retention_expires_at>?`, owner, now)
	if e != nil {
		return e
	}
	ids = nil
	for rows.Next() {
		var id string
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			return e
		}
		ids = append(ids, id)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	for _, id := range ids {
		if e = s.restoreNovelAutomatic(ctx, owner, id); e != nil {
			return e
		}
	}
	return nil
}
func (s *Service) restoreNovel(ctx context.Context, owner, id string) error {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var status string
	if e = tx.QueryRowContext(ctx, `SELECT status FROM novel_projects WHERE owner_username=? AND project_id=? AND deleted_at IS NULL FOR UPDATE`, owner, id).Scan(&status); e != nil {
		return e
	}
	if status == "purging" {
		return ErrPurging
	}
	if status != "archived" {
		return ErrNotFound
	}
	now := s.now().UTC()
	r, e := tx.ExecContext(ctx, `UPDATE project_retention SET last_activity_at=?,retention_expires_at=?,archived_at=NULL,manual_archived=0,updated_at=? WHERE owner_username=? AND module='novel_creation' AND project_id=? AND archived_at IS NOT NULL`, now, now.AddDate(0, 0, s.days(owner)), now, owner, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return ErrNotFound
	}
	if _, e = tx.ExecContext(ctx, `UPDATE novel_projects SET status='draft',updated_at=? WHERE owner_username=? AND project_id=? AND deleted_at IS NULL`, now, owner, id); e != nil {
		return e
	}
	if e = pauseNovelAutomation(ctx, tx, owner, id); e != nil {
		return e
	}
	return tx.Commit()
}

// Changing retention settings may unarchive an automatic archive, but must not
// restart its retention clock. Only the explicit Restore action does that.
func (s *Service) restoreNovelAutomatic(ctx context.Context, owner, id string) error {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var one int
	if e = tx.QueryRowContext(ctx, `SELECT 1 FROM novel_projects WHERE owner_username=? AND project_id=? AND deleted_at IS NULL AND status='archived' FOR UPDATE`, owner, id).Scan(&one); e != nil {
		if errors.Is(e, sql.ErrNoRows) {
			return nil
		}
		return e
	}
	r, e := tx.ExecContext(ctx, `UPDATE project_retention SET archived_at=NULL,updated_at=? WHERE owner_username=? AND module='novel_creation' AND project_id=? AND archived_at IS NOT NULL AND manual_archived=0 AND retention_expires_at>?`, s.now().UTC(), owner, id, s.now().UTC())
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return nil
	}
	if _, e = tx.ExecContext(ctx, `UPDATE novel_projects SET status='draft' WHERE owner_username=? AND project_id=?`, owner, id); e != nil {
		return e
	}
	return tx.Commit()
}

// Persist a deletion intent before slow object I/O. Restore and all novel writers
// reject purging; partial object failures retain this intent and every source key.
func (s *Service) beginNovelPurge(ctx context.Context, owner, id string) error {
	tx, e := s.DB.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var one int
	if e = tx.QueryRowContext(ctx, `SELECT 1 FROM novel_projects WHERE owner_username=? AND project_id=? AND status IN ('archived','purging') FOR UPDATE`, owner, id).Scan(&one); e != nil {
		return e
	}
	var archived sql.NullTime
	if e = tx.QueryRowContext(ctx, `SELECT archived_at FROM project_retention WHERE owner_username=? AND module='novel_creation' AND project_id=? FOR UPDATE`, owner, id).Scan(&archived); e != nil {
		return e
	}
	if !archived.Valid {
		return ErrInvalid
	}
	if s.NovelObjects == nil {
		var objects int
		if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM novel_content_records WHERE owner_username=? AND project_id=? AND kind IN ('document','export')`, owner, id).Scan(&objects); e != nil {
			return e
		}
		if objects > 0 {
			return fmt.Errorf("小说原件清理服务未配置，归档已保留")
		}
	}
	if _, e = tx.ExecContext(ctx, `UPDATE novel_projects SET status='purging' WHERE owner_username=? AND project_id=?`, owner, id); e != nil {
		return e
	}
	return tx.Commit()
}
