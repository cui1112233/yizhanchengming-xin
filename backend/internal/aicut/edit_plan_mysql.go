package aicut

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

type editPlanSQLReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

const editPlanSelect = `SELECT id, client_request_id, base_revision, model_id, state, request_json, summary_json, warnings_json, COALESCE(error_message, ''), applied_revision, created_at, updated_at FROM aicut_edit_plans`

func readEditPlanSQL(ctx context.Context, reader editPlanSQLReader, owner, projectID, predicate string, value string, forUpdate bool) (EditPlan, error) {
	query := editPlanSelect + ` WHERE owner_username = ? AND project_id = ? AND ` + predicate + ` = ?`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	var plan EditPlan
	var requestJSON, summaryJSON, warningsJSON []byte
	err := reader.QueryRowContext(ctx, query, owner, projectID, value).Scan(
		&plan.ID, &plan.ClientRequestID, &plan.BaseRevision, &plan.ModelID, &plan.State,
		&requestJSON, &summaryJSON, &warningsJSON, &plan.Error, &plan.AppliedRevision, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return EditPlan{}, ErrEditPlanNotFound
	}
	if err != nil {
		return EditPlan{}, err
	}
	plan.Owner, plan.ProjectID = owner, projectID
	if err := json.Unmarshal(requestJSON, &plan.Request); err != nil {
		return EditPlan{}, err
	}
	if err := json.Unmarshal(summaryJSON, &plan.Summary); err != nil {
		return EditPlan{}, err
	}
	if err := json.Unmarshal(warningsJSON, &plan.Warnings); err != nil {
		return EditPlan{}, err
	}
	rows, err := reader.QueryContext(ctx, `SELECT operation_json FROM aicut_edit_plan_operations WHERE plan_id = ? ORDER BY ordinal_no`, plan.ID)
	if err != nil {
		return EditPlan{}, err
	}
	defer rows.Close()
	plan.Operations = []EditOperation{}
	for rows.Next() {
		var encoded []byte
		var operation EditOperation
		if err := rows.Scan(&encoded); err != nil {
			return EditPlan{}, err
		}
		if err := json.Unmarshal(encoded, &operation); err != nil {
			return EditPlan{}, err
		}
		plan.Operations = append(plan.Operations, operation)
	}
	return plan, rows.Err()
}

func (s *MySQLStore) CreateEditPlan(ctx context.Context, plan EditPlan) (EditPlan, error) {
	normalized, err := NormalizeEditPlan(plan)
	if err != nil {
		return EditPlan{}, err
	}
	if existing, findErr := s.EditPlanByRequest(ctx, normalized.Owner, normalized.ProjectID, normalized.ClientRequestID); findErr == nil {
		return existing, nil
	} else if !errors.Is(findErr, ErrEditPlanNotFound) {
		return EditPlan{}, findErr
	}
	requestJSON, _ := json.Marshal(normalized.Request)
	summaryJSON, _ := json.Marshal(normalized.Summary)
	warningsJSON, _ := json.Marshal(normalized.Warnings)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EditPlan{}, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO aicut_edit_plans (id, client_request_id, project_id, owner_username, base_revision, model_id, state, request_json, summary_json, warnings_json, error_message, applied_revision, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, normalized.ID, normalized.ClientRequestID, normalized.ProjectID, normalized.Owner, normalized.BaseRevision, normalized.ModelID, normalized.State, requestJSON, summaryJSON, warningsJSON, normalized.Error, normalized.AppliedRevision, normalized.CreatedAt, normalized.UpdatedAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return s.EditPlanByRequest(ctx, normalized.Owner, normalized.ProjectID, normalized.ClientRequestID)
		}
		return EditPlan{}, err
	}
	for index, operation := range normalized.Operations {
		encoded, _ := json.Marshal(operation)
		if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_edit_plan_operations (plan_id, operation_id, ordinal_no, operation_json) VALUES (?, ?, ?, ?)`, normalized.ID, operation.ID, index, encoded); err != nil {
			return EditPlan{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return EditPlan{}, err
	}
	return normalized, nil
}

func (s *MySQLStore) EditPlan(ctx context.Context, owner, projectID, id string) (EditPlan, error) {
	return readEditPlanSQL(ctx, s.db, owner, projectID, "id", id, false)
}

func (s *MySQLStore) EditPlanByRequest(ctx context.Context, owner, projectID, requestID string) (EditPlan, error) {
	return readEditPlanSQL(ctx, s.db, owner, projectID, "client_request_id", requestID, false)
}

func (s *MySQLStore) UpdateEditPlan(ctx context.Context, plan EditPlan) (EditPlan, error) {
	normalized, err := NormalizeEditPlan(plan)
	if err != nil {
		return EditPlan{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EditPlan{}, err
	}
	defer tx.Rollback()
	current, err := readEditPlanSQL(ctx, tx, normalized.Owner, normalized.ProjectID, "id", normalized.ID, true)
	if err != nil {
		return EditPlan{}, err
	}
	if current.ClientRequestID != normalized.ClientRequestID || !validEditPlanStateTransition(current.State, normalized.State) {
		return EditPlan{}, ErrInvalidEditPlan
	}
	requestJSON, _ := json.Marshal(normalized.Request)
	summaryJSON, _ := json.Marshal(normalized.Summary)
	warningsJSON, _ := json.Marshal(normalized.Warnings)
	if _, err := tx.ExecContext(ctx, `UPDATE aicut_edit_plans SET state = ?, request_json = ?, summary_json = ?, warnings_json = ?, error_message = ?, applied_revision = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, normalized.State, requestJSON, summaryJSON, warningsJSON, normalized.Error, normalized.AppliedRevision, normalized.UpdatedAt, normalized.ID, normalized.ProjectID, normalized.Owner); err != nil {
		return EditPlan{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM aicut_edit_plan_operations WHERE plan_id = ?`, normalized.ID); err != nil {
		return EditPlan{}, err
	}
	for index, operation := range normalized.Operations {
		encoded, _ := json.Marshal(operation)
		if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_edit_plan_operations (plan_id, operation_id, ordinal_no, operation_json) VALUES (?, ?, ?, ?)`, normalized.ID, operation.ID, index, encoded); err != nil {
			return EditPlan{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return EditPlan{}, err
	}
	return normalized, nil
}

func (s *MySQLStore) UpdateEditPlanState(ctx context.Context, owner, projectID, id string, state EditPlanState, message string, now time.Time) (EditPlan, error) {
	plan, err := s.EditPlan(ctx, owner, projectID, id)
	if err != nil {
		return EditPlan{}, err
	}
	if !validEditPlanStateTransition(plan.State, state) {
		return EditPlan{}, ErrInvalidEditPlan
	}
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_edit_plans SET state = ?, error_message = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ? AND state = ?`, state, message, now.UTC(), id, projectID, owner, plan.State)
	if err != nil {
		return EditPlan{}, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return s.EditPlan(ctx, owner, projectID, id)
	}
	plan.State, plan.Error, plan.UpdatedAt = state, message, now.UTC()
	return plan, nil
}

func editPlanApplySummaryFromPlan(plan EditPlan) EditPlanApplySummary {
	summary := EditPlanApplySummary{EstimatedDurationMS: plan.Summary.EstimatedDurationMS}
	for _, operation := range plan.Operations {
		if operation.Enabled {
			summary.Applied++
		} else {
			summary.Skipped++
		}
	}
	return summary
}

func (s *MySQLStore) ApplyStoredEditPlan(ctx context.Context, owner, projectID, id string, enabledIDs []string, now time.Time) (EditPlan, Timeline, EditPlanApplySummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	defer tx.Rollback()
	plan, err := readEditPlanSQL(ctx, tx, owner, projectID, "id", id, true)
	if err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	var currentRevision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, projectID, owner).Scan(&currentRevision); errors.Is(err, sql.ErrNoRows) {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrProjectNotFound
	} else if err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	current, err := timelineFromTx(ctx, tx, owner, projectID)
	if err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	if plan.State == EditPlanApplied {
		return plan, current, editPlanApplySummaryFromPlan(plan), nil
	}
	if plan.State != EditPlanReady {
		return plan, current, EditPlanApplySummary{}, ErrInvalidEditPlan
	}
	if currentRevision != plan.BaseRevision || current.Revision != plan.BaseRevision {
		plan.State, plan.Error, plan.UpdatedAt = EditPlanStale, "草案已过期，请基于最新工程重新生成", now.UTC()
		if _, err := tx.ExecContext(ctx, `UPDATE aicut_edit_plans SET state = ?, error_message = ?, updated_at = ? WHERE id = ?`, plan.State, plan.Error, plan.UpdatedAt, plan.ID); err != nil {
			return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
		}
		if err := tx.Commit(); err != nil {
			return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
		}
		return plan, current, EditPlanApplySummary{}, ErrRevisionConflict
	}
	selected := enabledEditPlan(plan, enabledIDs)
	nextTimeline, summary, err := ApplyEditPlan(current, selected)
	if err != nil {
		return plan, current, EditPlanApplySummary{}, err
	}
	cursor, _, err := ensureTimelineHistoryTx(ctx, tx, owner, projectID, currentRevision, now.UTC())
	if err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	nextTimeline.Revision = currentRevision + 1
	if _, err := tx.ExecContext(ctx, `DELETE FROM aicut_timeline_history WHERE project_id = ? AND owner_username = ? AND sequence_no > ?`, projectID, owner, cursor); err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	sequence := cursor + 1
	encoded, _ := json.Marshal(nextTimeline)
	if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_timeline_history (project_id, owner_username, sequence_no, timeline_json, created_at) VALUES (?, ?, ?, ?, ?)`, projectID, owner, sequence, string(encoded), now.UTC()); err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE aicut_timeline_history_state SET cursor_no = ?, max_no = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`, sequence, sequence, now.UTC(), projectID, owner); err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	if err := replaceTimelineTx(ctx, tx, owner, projectID, nextTimeline, now.UTC()); err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	if summary.AspectRatio != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE aicut_projects SET aspect_ratio = ? WHERE id = ? AND owner_username = ?`, summary.AspectRatio, projectID, owner); err != nil {
			return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
		}
	}
	plan.State, plan.Error, plan.AppliedRevision, plan.UpdatedAt = EditPlanApplied, "", nextTimeline.Revision, now.UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE aicut_edit_plans SET state = ?, error_message = '', applied_revision = ?, updated_at = ? WHERE id = ?`, plan.State, plan.AppliedRevision, plan.UpdatedAt, plan.ID); err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	if err := tx.Commit(); err != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, err
	}
	return plan, nextTimeline, summary, nil
}
