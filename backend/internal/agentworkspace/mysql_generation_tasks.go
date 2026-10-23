package agentworkspace

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const generationTaskColumns = `id, project_id, owner_username, node_id, node_version, kind, provider, model_id, mode, state, settings_json, reference_asset_ids_json, selected_result_id, idempotency_key, retry_of_task_id, error_message, created_at, updated_at`

const generationResultColumns = `id, asset_id, kind, asset_json, created_at`

type generationQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func scanGenerationTask(scanner interface{ Scan(...any) error }) (GenerationTask, error) {
	var task GenerationTask
	var mode, selectedResultID, retryOfTaskID, errorMessage sql.NullString
	var settingsJSON, referencesJSON []byte
	if err := scanner.Scan(
		&task.ID, &task.ProjectID, &task.OwnerUsername, &task.NodeID, &task.NodeVersion, &task.Kind, &task.Provider, &task.ModelID,
		&mode, &task.State, &settingsJSON, &referencesJSON, &selectedResultID, &task.IdempotencyKey, &retryOfTaskID, &errorMessage,
		&task.CreatedAt, &task.UpdatedAt,
	); err != nil {
		return GenerationTask{}, err
	}
	task.Mode = mode.String
	task.SelectedResultID = selectedResultID.String
	task.RetryOfTaskID = retryOfTaskID.String
	task.ErrorMessage = errorMessage.String
	if !validGenerationState(task.State) || json.Unmarshal(settingsJSON, &task.Settings) != nil || json.Unmarshal(referencesJSON, &task.ReferenceAssetIDs) != nil {
		return GenerationTask{}, ErrGenerationInvalid
	}
	if task.Settings == nil {
		task.Settings = map[string]any{}
	}
	if task.ReferenceAssetIDs == nil {
		task.ReferenceAssetIDs = []string{}
	}
	task.Results = []GenerationResult{}
	return task, nil
}

func loadGenerationResults(ctx context.Context, queryer generationQueryer, owner, projectID, taskID string) ([]GenerationResult, error) {
	rows, err := queryer.QueryContext(ctx, `SELECT `+generationResultColumns+` FROM agent_generation_results WHERE owner_username = ? AND project_id = ? AND task_id = ? ORDER BY created_at, id`, owner, projectID, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]GenerationResult, 0)
	for rows.Next() {
		var result GenerationResult
		var assetJSON []byte
		if err := rows.Scan(&result.ID, &result.AssetID, &result.Kind, &assetJSON, &result.CreatedAt); err != nil {
			return nil, err
		}
		if len(assetJSON) > 0 && string(assetJSON) != "null" {
			if err := json.Unmarshal(assetJSON, &result.Asset); err != nil {
				return nil, ErrGenerationInvalid
			}
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (s *MySQLStore) generationTaskByIdempotency(ctx context.Context, owner, projectID, idempotencyKey string) (GenerationTask, error) {
	task, err := scanGenerationTask(s.db.QueryRowContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE owner_username = ? AND project_id = ? AND idempotency_key = ?`, owner, projectID, idempotencyKey))
	if err != nil {
		return GenerationTask{}, err
	}
	task.Results, err = loadGenerationResults(ctx, s.db, owner, projectID, task.ID)
	return task, err
}

func (s *MySQLStore) CreateGenerationTask(ctx context.Context, owner string, input CreateGenerationTaskInput) (GenerationTask, error) {
	owner = strings.TrimSpace(owner)
	input, err := normalizeCreateGenerationTask(input)
	if err != nil || owner == "" {
		return GenerationTask{}, ErrGenerationInvalid
	}
	if existing, lookupErr := s.generationTaskByIdempotency(ctx, owner, input.ProjectID, input.IdempotencyKey); lookupErr == nil {
		return existing, nil
	} else if !errors.Is(lookupErr, sql.ErrNoRows) {
		return GenerationTask{}, lookupErr
	}
	settingsJSON, err := json.Marshal(input.Settings)
	if err != nil {
		return GenerationTask{}, ErrGenerationInvalid
	}
	referencesJSON, err := json.Marshal(input.ReferenceAssetIDs)
	if err != nil {
		return GenerationTask{}, ErrGenerationInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GenerationTask{}, err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, input.ProjectID, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return GenerationTask{}, ErrProjectNotFound
	} else if err != nil {
		return GenerationTask{}, err
	}
	task := GenerationTask{
		ID: input.ID, ProjectID: input.ProjectID, OwnerUsername: owner, NodeID: input.NodeID, NodeVersion: input.NodeVersion,
		Kind: input.Kind, Provider: input.Provider, ModelID: input.ModelID, Mode: input.Mode, State: GenerationQueued,
		Settings: input.Settings, ReferenceAssetIDs: input.ReferenceAssetIDs, Results: []GenerationResult{}, IdempotencyKey: input.IdempotencyKey,
		RetryOfTaskID: input.RetryOfTaskID, CreatedAt: input.Now, UpdatedAt: input.Now,
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_generation_tasks (id, project_id, owner_username, node_id, node_version, kind, provider, model_id, mode, state, settings_json, reference_asset_ids_json, selected_result_id, idempotency_key, retry_of_task_id, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, task.ProjectID, owner, task.NodeID, task.NodeVersion, task.Kind, task.Provider, task.ModelID, nullableString(task.Mode), task.State,
		string(settingsJSON), string(referencesJSON), nil, task.IdempotencyKey, nullableString(task.RetryOfTaskID), nil, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			_ = tx.Rollback()
			if existing, lookupErr := s.generationTaskByIdempotency(ctx, owner, input.ProjectID, input.IdempotencyKey); lookupErr == nil {
				return existing, nil
			}
			return GenerationTask{}, ErrGenerationInvalid
		}
		return GenerationTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return GenerationTask{}, err
	}
	return task, nil
}

func (s *MySQLStore) GenerationTask(ctx context.Context, owner, projectID, id string) (GenerationTask, error) {
	owner, projectID, id = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(id)
	task, err := scanGenerationTask(s.db.QueryRowContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE id = ? AND project_id = ? AND owner_username = ?`, id, projectID, owner))
	if errors.Is(err, sql.ErrNoRows) {
		return GenerationTask{}, ErrGenerationNotFound
	}
	if err != nil {
		return GenerationTask{}, err
	}
	task.Results, err = loadGenerationResults(ctx, s.db, owner, projectID, id)
	return task, err
}

func (s *MySQLStore) ListGenerationTasks(ctx context.Context, owner, projectID string) ([]GenerationTask, error) {
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	if _, err := s.Project(ctx, owner, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE owner_username = ? AND project_id = ? ORDER BY updated_at DESC, id DESC`, owner, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]GenerationTask, 0)
	for rows.Next() {
		task, scanErr := scanGenerationTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range tasks {
		tasks[index].Results, err = loadGenerationResults(ctx, s.db, owner, projectID, tasks[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

func (s *MySQLStore) ListRunnableGenerationTasks(ctx context.Context, mode string, limit int) ([]GenerationTask, error) {
	return s.ListRunnableGenerationTasksByModes(ctx, []string{mode}, limit)
}

func (s *MySQLStore) ListRunnableGenerationTasksByModes(ctx context.Context, modes []string, limit int) ([]GenerationTask, error) {
	uniqueModes := make([]string, 0, len(modes))
	seen := make(map[string]struct{}, len(modes))
	for _, mode := range modes {
		mode = strings.TrimSpace(mode)
		if mode == "" {
			continue
		}
		if _, exists := seen[mode]; exists {
			continue
		}
		seen[mode] = struct{}{}
		uniqueModes = append(uniqueModes, mode)
	}
	if len(uniqueModes) == 0 {
		return nil, ErrGenerationInvalid
	}
	if limit <= 0 {
		limit = 100
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(uniqueModes)), ",")
	args := make([]any, 0, len(uniqueModes)+1)
	for _, mode := range uniqueModes {
		args = append(args, mode)
	}
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE mode IN (`+placeholders+`) AND state IN ('queued', 'running') ORDER BY updated_at ASC, id ASC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]GenerationTask, 0)
	for rows.Next() {
		task, scanErr := scanGenerationTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *MySQLStore) TransitionGenerationTask(ctx context.Context, owner, projectID, id string, transition GenerationTaskTransition) (GenerationTask, error) {
	owner, projectID, id = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(id)
	if owner == "" || projectID == "" || id == "" || !validGenerationState(transition.State) || transition.Now.IsZero() {
		return GenerationTask{}, ErrGenerationInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return GenerationTask{}, err
	}
	defer tx.Rollback()
	task, err := scanGenerationTask(tx.QueryRowContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, id, projectID, owner))
	if errors.Is(err, sql.ErrNoRows) {
		return GenerationTask{}, ErrGenerationNotFound
	}
	if err != nil {
		return GenerationTask{}, err
	}
	if !transition.ExpectedUpdatedAt.IsZero() && !normalizeGenerationTime(transition.ExpectedUpdatedAt).Equal(normalizeGenerationTime(task.UpdatedAt)) {
		return GenerationTask{}, ErrGenerationState
	}
	if !validGenerationTransition(task.State, transition.State) {
		return GenerationTask{}, ErrGenerationState
	}
	if transition.State == GenerationAwaitingConfirmation {
		results, normalizeErr := normalizeGenerationResults(transition.Results, task.Kind, transition.Now)
		if normalizeErr != nil || len(results) == 0 {
			return GenerationTask{}, ErrGenerationInvalid
		}
		selected := strings.TrimSpace(transition.SelectedResultID)
		if selected == "" {
			selected = results[0].ID
		}
		if !generationResultExists(results, selected) {
			return GenerationTask{}, ErrGenerationInvalid
		}
		for _, result := range results {
			assetJSON, marshalErr := json.Marshal(result.Asset)
			if marshalErr != nil {
				return GenerationTask{}, ErrGenerationInvalid
			}
			var assetValue any
			if result.Asset != nil {
				assetValue = string(assetJSON)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO agent_generation_results (id, task_id, project_id, owner_username, asset_id, kind, asset_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, result.ID, task.ID, task.ProjectID, owner, result.AssetID, result.Kind, assetValue, result.CreatedAt); err != nil {
				return GenerationTask{}, err
			}
		}
		task.Results = results
		task.SelectedResultID = selected
	}
	if transition.State == GenerationCompleted && strings.TrimSpace(transition.SelectedResultID) != "" {
		if len(task.Results) == 0 {
			task.Results, err = loadGenerationResults(ctx, tx, owner, projectID, id)
			if err != nil {
				return GenerationTask{}, err
			}
		}
		if !generationResultExists(task.Results, strings.TrimSpace(transition.SelectedResultID)) {
			return GenerationTask{}, ErrGenerationInvalid
		}
		task.SelectedResultID = strings.TrimSpace(transition.SelectedResultID)
	}
	task.State = transition.State
	task.ErrorMessage = strings.TrimSpace(transition.ErrorMessage)
	task.UpdatedAt = nextGenerationUpdatedAt(task.UpdatedAt, transition.Now)
	if _, err := tx.ExecContext(ctx, `UPDATE agent_generation_tasks SET state = ?, selected_result_id = ?, error_message = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, task.State, nullableString(task.SelectedResultID), nullableString(task.ErrorMessage), task.UpdatedAt, task.ID, task.ProjectID, owner); err != nil {
		return GenerationTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return GenerationTask{}, err
	}
	return task, nil
}

func (s *MySQLStore) ConfirmGenerationTask(ctx context.Context, owner, projectID, id, resultID string, now time.Time) (Canvas, GenerationTask, error) {
	owner, projectID, id = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(id)
	if owner == "" || projectID == "" || id == "" || now.IsZero() {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	now = normalizeGenerationTime(now)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	defer tx.Rollback()
	task, err := scanGenerationTask(tx.QueryRowContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, id, projectID, owner))
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ErrGenerationNotFound
	}
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if task.State != GenerationAwaitingConfirmation {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	task.Results, err = loadGenerationResults(ctx, tx, owner, projectID, id)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	result, ok := selectedGenerationResult(task, resultID)
	if !ok {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	var asset Asset
	err = tx.QueryRowContext(ctx, `SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at FROM agent_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, result.AssetID, projectID, owner).
		Scan(&asset.ID, &asset.ProjectID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if asset.Kind != task.Kind {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	var node Node
	var sourceJSON, oldAssetJSON, configJSON []byte
	var groupID sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, projectID, owner, task.NodeID).
		Scan(&node.ID, &node.Type, &node.Title, &node.Summary, &node.Position.X, &node.Position.Y, &node.Version, &node.Status, &sourceJSON, &oldAssetJSON, &configJSON, &groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ErrGenerationNodeChanged
	}
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if node.Type != task.Kind || node.Version != task.NodeVersion {
		return Canvas{}, GenerationTask{}, ErrGenerationNodeChanged
	}
	_ = json.Unmarshal(sourceJSON, &node.SourceNodeIDs)
	_ = json.Unmarshal(configJSON, &node.Config)
	if groupID.Valid {
		node.GroupID = groupID.String
	}
	node.AssetRef = map[string]any{"id": asset.ID, "projectId": asset.ProjectID, "kind": asset.Kind, "name": asset.Name, "contentType": asset.ContentType, "byteSize": asset.ByteSize}
	node.Status = "confirmed"
	node.Version++
	assetRefJSON, err := json.Marshal(node.AssetRef)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_canvas_nodes SET asset_ref_json = ?, status = ?, current_version = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, string(assetRefJSON), node.Status, node.Version, now, projectID, owner, node.ID); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	contentJSON, err := json.Marshal(node)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_node_versions (project_id, node_id, owner_username, version, content_json, source_execution_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, projectID, node.ID, owner, node.Version, string(contentJSON), task.ID, now); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if err := updateCanvasSnapshotNode(ctx, tx, owner, projectID, node.ID, now, func(snapshotNode *Node) {
		snapshotNode.AssetRef = cloneMap(node.AssetRef)
		snapshotNode.Status = node.Status
		snapshotNode.Version = node.Version
	}); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now, projectID, owner); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	task.State = GenerationCompleted
	task.SelectedResultID = result.ID
	task.UpdatedAt = nextGenerationUpdatedAt(task.UpdatedAt, now)
	if _, err := tx.ExecContext(ctx, `UPDATE agent_generation_tasks SET state = ?, selected_result_id = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, task.State, task.SelectedResultID, task.UpdatedAt, task.ID, task.ProjectID, owner); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	canvas, err := s.Canvas(ctx, owner, projectID)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	return canvas, task, nil
}

func (s *MySQLStore) ActivateGenerationResult(ctx context.Context, owner, projectID, id, resultID string, expectedNodeVersion int64, now time.Time) (Canvas, GenerationTask, error) {
	owner, projectID, id, resultID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(id), strings.TrimSpace(resultID)
	if owner == "" || projectID == "" || id == "" || resultID == "" || expectedNodeVersion < 1 || now.IsZero() {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	now = normalizeGenerationTime(now)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	defer tx.Rollback()
	task, err := scanGenerationTask(tx.QueryRowContext(ctx, `SELECT `+generationTaskColumns+` FROM agent_generation_tasks WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, id, projectID, owner))
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ErrGenerationNotFound
	}
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if task.State != GenerationCompleted {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	task.Results, err = loadGenerationResults(ctx, tx, owner, projectID, id)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	result, ok := selectedGenerationResult(task, resultID)
	if !ok || result.AssetID == "" || result.Kind != task.Kind {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	var asset Asset
	err = tx.QueryRowContext(ctx, `SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at FROM agent_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, result.AssetID, projectID, owner).
		Scan(&asset.ID, &asset.ProjectID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if asset.Kind != task.Kind {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	var node Node
	var sourceJSON, oldAssetJSON, configJSON []byte
	var groupID sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, projectID, owner, task.NodeID).
		Scan(&node.ID, &node.Type, &node.Title, &node.Summary, &node.Position.X, &node.Position.Y, &node.Version, &node.Status, &sourceJSON, &oldAssetJSON, &configJSON, &groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ErrGenerationNodeChanged
	}
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if node.Type != task.Kind || node.Version != expectedNodeVersion {
		return Canvas{}, GenerationTask{}, ErrGenerationNodeChanged
	}
	_ = json.Unmarshal(sourceJSON, &node.SourceNodeIDs)
	_ = json.Unmarshal(configJSON, &node.Config)
	if groupID.Valid {
		node.GroupID = groupID.String
	}
	node.AssetRef = map[string]any{"id": asset.ID, "projectId": asset.ProjectID, "kind": asset.Kind, "name": asset.Name, "contentType": asset.ContentType, "byteSize": asset.ByteSize}
	node.Status = "confirmed"
	node.Version++
	assetRefJSON, err := json.Marshal(node.AssetRef)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_canvas_nodes SET asset_ref_json = ?, status = ?, current_version = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, string(assetRefJSON), node.Status, node.Version, now, projectID, owner, node.ID); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	contentJSON, err := json.Marshal(node)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_node_versions (project_id, node_id, owner_username, version, content_json, source_execution_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, projectID, node.ID, owner, node.Version, string(contentJSON), task.ID, now); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if err := updateCanvasSnapshotNode(ctx, tx, owner, projectID, node.ID, now, func(snapshotNode *Node) {
		snapshotNode.AssetRef = cloneMap(node.AssetRef)
		snapshotNode.Status = node.Status
		snapshotNode.Version = node.Version
	}); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now, projectID, owner); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	task.SelectedResultID = result.ID
	task.UpdatedAt = nextGenerationUpdatedAt(task.UpdatedAt, now)
	if _, err := tx.ExecContext(ctx, `UPDATE agent_generation_tasks SET selected_result_id = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, task.SelectedResultID, task.UpdatedAt, task.ID, task.ProjectID, owner); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	canvas, err := s.Canvas(ctx, owner, projectID)
	if err != nil {
		return Canvas{}, GenerationTask{}, err
	}
	return canvas, task, nil
}

func (s *MySQLStore) CancelGenerationTask(ctx context.Context, owner, projectID, id string, now time.Time) (GenerationTask, error) {
	return s.TransitionGenerationTask(ctx, owner, projectID, id, GenerationTaskTransition{State: GenerationCancelled, Now: now})
}

func (s *MySQLStore) RetryGenerationTask(ctx context.Context, owner, projectID, id string, input RetryGenerationTaskInput) (GenerationTask, error) {
	owner, projectID, id = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(id)
	input.ID, input.IdempotencyKey = strings.TrimSpace(input.ID), strings.TrimSpace(input.IdempotencyKey)
	if input.ID == "" || input.IdempotencyKey == "" || input.Now.IsZero() {
		return GenerationTask{}, ErrGenerationInvalid
	}
	original, err := s.GenerationTask(ctx, owner, projectID, id)
	if err != nil {
		return GenerationTask{}, err
	}
	if original.State != GenerationFailed && original.State != GenerationCancelled {
		return GenerationTask{}, ErrGenerationState
	}
	canvas, err := s.Canvas(ctx, owner, projectID)
	if err != nil {
		return GenerationTask{}, err
	}
	matched := false
	for _, node := range canvas.Nodes {
		if node.ID == original.NodeID {
			matched = node.Type == original.Kind && node.Version == original.NodeVersion
			break
		}
	}
	if !matched {
		return GenerationTask{}, ErrGenerationNodeChanged
	}
	return s.CreateGenerationTask(ctx, owner, CreateGenerationTaskInput{
		ID: input.ID, ProjectID: projectID, NodeID: original.NodeID, NodeVersion: original.NodeVersion, Kind: original.Kind,
		Provider: original.Provider, ModelID: original.ModelID, Mode: original.Mode, Settings: original.Settings,
		ReferenceAssetIDs: original.ReferenceAssetIDs, IdempotencyKey: input.IdempotencyKey, RetryOfTaskID: original.ID, Now: input.Now,
	})
}
