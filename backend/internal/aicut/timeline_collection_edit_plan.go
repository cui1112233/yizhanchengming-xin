package aicut

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// ApplyStoredEditPlan commits the selected timeline, its history, and the plan
// state together. Named plans never enter the legacy default-timeline writer.
func (s *TimelineCollectionStore) ApplyStoredEditPlan(ctx context.Context, plans EditPlanStore, o, p, id, target string, enabled []string, now time.Time) (EditPlan, Timeline, EditPlanApplySummary, error) {
	plan, e := plans.EditPlan(ctx, o, p, id)
	if e != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
	}
	if canonicalASRTimeline(plan.TimelineID) != canonicalASRTimeline(target) {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrEditPlanNotFound
	}
	if canonicalASRTimeline(target) == DefaultTimelineID {
		return plans.ApplyStoredEditPlan(ctx, o, p, id, enabled, now)
	}
	if repo, ok := s.repo.(*memoryCollectionRepository); ok {
		if base, ok := s.base.(*MemoryStore); ok {
			return applyMemoryNamedPlan(base, repo, o, p, id, target, enabled, now)
		}
	}
	if repo, ok := s.repo.(*mysqlCollectionRepository); ok {
		return applyMySQLNamedPlan(ctx, repo.db, o, p, id, target, enabled, now)
	}
	return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrInvalidEditPlan
}
func timelineCollectionAssetIDs(v Timeline) map[string]bool {
	refs := map[string]bool{}
	for _, m := range v.Markers {
		refs[m.AssetID] = true
	}
	for _, t := range v.Tracks {
		refs[t.Settings.FontAssetID] = true
		for _, c := range t.Clips {
			refs[c.AssetID] = true
			if c.VisualProcessing != nil {
				refs[c.VisualProcessing.PersonSegmentation.MatteAssetID] = true
			}
		}
	}
	delete(refs, "")
	return refs
}
func applyMemoryNamedPlan(base *MemoryStore, repo *memoryCollectionRepository, o, p, id, target string, enabled []string, now time.Time) (EditPlan, Timeline, EditPlanApplySummary, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	base.mu.Lock()
	defer base.mu.Unlock()
	key := editPlanKey(o, p, id)
	plan, ok := base.editPlans[key]
	if !ok || canonicalASRTimeline(plan.TimelineID) != canonicalASRTimeline(target) {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrEditPlanNotFound
	}
	project, ok := base.projects[aicutProjectKey(o, p)]
	if !ok || project.project.DeletedAt != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrProjectNotFound
	}
	items := repo.projects[aicutProjectKey(o, p)]
	index := collectionIndex(items, target)
	if index < 0 {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrTimelineNotFound
	}
	record := cloneCollectionRecord(items[index])
	current := cloneTimeline(record.Value)
	if plan.State == EditPlanApplied {
		return cloneEditPlan(plan), current, base.editPlanSummaries[key], nil
	}
	if plan.State != EditPlanReady {
		return plan, current, EditPlanApplySummary{}, ErrInvalidEditPlan
	}
	if current.Revision != plan.BaseRevision {
		plan.State, plan.Error, plan.UpdatedAt = EditPlanStale, "草案已过期，请基于最新时间线重新生成", now.UTC()
		base.editPlans[key] = cloneEditPlan(plan)
		return plan, current, EditPlanApplySummary{}, ErrRevisionConflict
	}
	selected := enabledEditPlan(cloneEditPlan(plan), enabled)
	if e := refreshMemoryEditPlanAssets(base, &selected); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	next, summary, e := ApplyEditPlan(current, selected)
	if e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	for asset := range timelineCollectionAssetIDs(next) {
		if _, ok := base.assets[aicutAssetKey(o, p, asset)]; !ok {
			return plan, current, EditPlanApplySummary{}, ErrAssetNotFound
		}
	}
	next.Revision = current.Revision + 1
	record.Value = cloneTimeline(next)
	record.History = append(record.History[:record.Cursor+1], cloneTimeline(next))
	record.Cursor++
	record.UpdatedAt = now.UTC()
	items[index] = record
	repo.projects[aicutProjectKey(o, p)] = items
	if summary.AspectRatio != "" {
		project.project.AspectRatio = summary.AspectRatio
	}
	project.project.UpdatedAt = now.UTC()
	base.projects[aicutProjectKey(o, p)] = project
	// Persist exactly the selected operations so replay reports the same summary.
	plan.Operations = selected.Operations
	plan.Summary.EstimatedDurationMS = summary.EstimatedDurationMS
	plan.State, plan.Error, plan.AppliedRevision, plan.UpdatedAt = EditPlanApplied, "", next.Revision, now.UTC()
	base.editPlans[key] = cloneEditPlan(plan)
	base.editPlanSummaries[key] = summary
	return cloneEditPlan(plan), cloneTimeline(next), summary, nil
}
func applyMySQLNamedPlan(ctx context.Context, db *sql.DB, o, p, id, target string, enabled []string, now time.Time) (EditPlan, Timeline, EditPlanApplySummary, error) {
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
	}
	defer tx.Rollback()
	plan, e := readEditPlanSQL(ctx, tx, o, p, "id", id, true)
	if e != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
	}
	if canonicalASRTimeline(plan.TimelineID) != canonicalASRTimeline(target) {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrEditPlanNotFound
	}
	var parent string
	e = tx.QueryRowContext(ctx, `SELECT id FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, p, o).Scan(&parent)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrProjectNotFound
	}
	if e != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
	}
	var revision int64
	var cursor int
	var raw []byte
	e = tx.QueryRowContext(ctx, `SELECT revision, cursor_no, timeline_json FROM aicut_timelines WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, p, o, target).Scan(&revision, &cursor, &raw)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrTimelineNotFound
	}
	if e != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
	}
	var current Timeline
	if e = json.Unmarshal(raw, &current); e != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
	}
	current.Revision = revision
	if plan.State == EditPlanApplied {
		return plan, current, editPlanApplySummaryFromPlan(plan), nil
	}
	if plan.State != EditPlanReady {
		return plan, current, EditPlanApplySummary{}, ErrInvalidEditPlan
	}
	if revision != plan.BaseRevision {
		plan.State, plan.Error, plan.UpdatedAt = EditPlanStale, "草案已过期，请基于最新时间线重新生成", now.UTC()
		if _, e = tx.ExecContext(ctx, `UPDATE aicut_edit_plans SET state = ?, error_message = ?, updated_at = ? WHERE id = ?`, plan.State, plan.Error, plan.UpdatedAt, plan.ID); e != nil {
			return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
		}
		if e = tx.Commit(); e != nil {
			return EditPlan{}, Timeline{}, EditPlanApplySummary{}, e
		}
		return plan, current, EditPlanApplySummary{}, ErrRevisionConflict
	}
	selected := enabledEditPlan(cloneEditPlan(plan), enabled)
	if e := refreshSQLEditPlanAssets(ctx, tx, &selected); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	next, summary, e := ApplyEditPlan(current, selected)
	if e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	for asset := range timelineCollectionAssetIDs(next) {
		var found int
		e = tx.QueryRowContext(ctx, `SELECT 1 FROM aicut_assets WHERE id = ? AND project_id = ? AND owner_username = ?`, asset, p, o).Scan(&found)
		if errors.Is(e, sql.ErrNoRows) {
			e = ErrAssetNotFound
		}
		if e != nil {
			return plan, current, EditPlanApplySummary{}, e
		}
	}
	next.Revision = revision + 1
	raw, e = json.Marshal(next)
	if e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	if _, e = tx.ExecContext(ctx, `DELETE FROM aicut_named_timeline_history WHERE project_id = ? AND owner_username = ? AND timeline_id = ? AND sequence_no > ?`, p, o, target, cursor); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	if _, e = tx.ExecContext(ctx, `INSERT INTO aicut_named_timeline_history (project_id, owner_username, timeline_id, sequence_no, timeline_json) VALUES (?, ?, ?, ?, ?)`, p, o, target, cursor+1, string(raw)); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	if _, e = tx.ExecContext(ctx, `UPDATE aicut_timelines SET revision = ?, cursor_no = ?, timeline_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, next.Revision, cursor+1, string(raw), now.UTC(), p, o, target); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	if summary.AspectRatio != "" {
		if _, e = tx.ExecContext(ctx, `UPDATE aicut_projects SET aspect_ratio = ? WHERE id = ? AND owner_username = ?`, summary.AspectRatio, p, o); e != nil {
			return plan, current, EditPlanApplySummary{}, e
		}
	}
	for ordinal, op := range selected.Operations {
		encoded, _ := json.Marshal(op)
		if _, e = tx.ExecContext(ctx, `UPDATE aicut_edit_plan_operations SET operation_json = ? WHERE plan_id = ? AND ordinal_no = ?`, string(encoded), plan.ID, ordinal); e != nil {
			return plan, current, EditPlanApplySummary{}, e
		}
	}
	plan.Operations = selected.Operations
	plan.Summary.EstimatedDurationMS = summary.EstimatedDurationMS
	plan.State, plan.Error, plan.AppliedRevision, plan.UpdatedAt = EditPlanApplied, "", next.Revision, now.UTC()
	summaryJSON, _ := json.Marshal(plan.Summary)
	if _, e = tx.ExecContext(ctx, `UPDATE aicut_edit_plans SET state = ?, error_message = '', applied_revision = ?, updated_at = ?, summary_json = ? WHERE id = ?`, plan.State, plan.AppliedRevision, plan.UpdatedAt, string(summaryJSON), plan.ID); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	if e = tx.Commit(); e != nil {
		return plan, current, EditPlanApplySummary{}, e
	}
	return plan, next, summary, nil
}
