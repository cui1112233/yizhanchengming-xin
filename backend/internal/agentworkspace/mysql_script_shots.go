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

const scriptShotColumns = `id, project_id, script_node_id, sequence, scene, characters, dialogue, narration, image_prompt, video_prompt, tts_text, revision, created_at, updated_at`

func scanScriptShot(scanner interface{ Scan(...any) error }) (ScriptShot, error) {
	var shot ScriptShot
	if err := scanner.Scan(&shot.ID, &shot.ProjectID, &shot.ScriptNodeID, &shot.Sequence, &shot.Scene, &shot.Characters, &shot.Dialogue, &shot.Narration, &shot.ImagePrompt, &shot.VideoPrompt, &shot.TTSText, &shot.Revision, &shot.CreatedAt, &shot.UpdatedAt); err != nil {
		return ScriptShot{}, err
	}
	return shot, nil
}

func (s *MySQLStore) ReplaceShots(ctx context.Context, owner, projectID, scriptNodeID string, shots []ScriptShot) error {
	owner, projectID, scriptNodeID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(scriptNodeID)
	if owner == "" || projectID == "" || scriptNodeID == "" {
		return ErrScriptShotInvalid
	}
	now := time.Now().UTC()
	normalized := make([]ScriptShot, 0, len(shots))
	ids := make(map[string]struct{}, len(shots))
	sequences := make(map[int]struct{}, len(shots))
	for _, shot := range shots {
		shot, err := normalizeScriptShot(owner, projectID, scriptNodeID, shot, now)
		if err != nil {
			return err
		}
		if _, exists := ids[shot.ID]; exists {
			return ErrScriptShotInvalid
		}
		if _, exists := sequences[shot.Sequence]; exists {
			return ErrScriptShotInvalid
		}
		ids[shot.ID] = struct{}{}
		sequences[shot.Sequence] = struct{}{}
		normalized = append(normalized, shot)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, projectID, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return ErrProjectNotFound
	} else if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_script_shots WHERE project_id = ? AND script_node_id = ? AND owner_username = ?`, projectID, scriptNodeID, owner); err != nil {
		return err
	}
	for _, shot := range normalized {
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_script_shots (id, project_id, script_node_id, owner_username, sequence, scene, characters, dialogue, narration, image_prompt, video_prompt, tts_text, revision, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, shot.ID, shot.ProjectID, shot.ScriptNodeID, owner, shot.Sequence, shot.Scene, shot.Characters, shot.Dialogue, shot.Narration, shot.ImagePrompt, shot.VideoPrompt, shot.TTSText, shot.Revision, shot.CreatedAt, shot.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *MySQLStore) ListShots(ctx context.Context, owner, projectID, scriptNodeID string) ([]ScriptShot, error) {
	owner, projectID, scriptNodeID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(scriptNodeID)
	if owner == "" || projectID == "" || scriptNodeID == "" {
		return nil, ErrScriptShotInvalid
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+scriptShotColumns+` FROM agent_script_shots WHERE owner_username = ? AND project_id = ? AND script_node_id = ? ORDER BY sequence, id`, owner, projectID, scriptNodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	shots := make([]ScriptShot, 0)
	for rows.Next() {
		shot, scanErr := scanScriptShot(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		shots = append(shots, shot)
	}
	return shots, rows.Err()
}

func (s *MySQLStore) UpdateShot(ctx context.Context, owner string, shot ScriptShot) (ScriptShot, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || shot.Revision < 1 {
		return ScriptShot{}, ErrScriptShotInvalid
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	normalized, err := normalizeScriptShot(owner, shot.ProjectID, shot.ScriptNodeID, shot, now)
	if err != nil {
		return ScriptShot{}, err
	}
	normalized.Revision = shot.Revision
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ScriptShot{}, err
	}
	defer tx.Rollback()
	var createdAt time.Time
	if err := tx.QueryRowContext(ctx, `SELECT created_at FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE`, normalized.ID, normalized.ProjectID, normalized.ScriptNodeID, owner).Scan(&createdAt); errors.Is(err, sql.ErrNoRows) {
		return ScriptShot{}, ErrScriptShotNotFound
	} else if err != nil {
		return ScriptShot{}, err
	}
	normalized.CreatedAt = createdAt.UTC().Truncate(time.Microsecond)
	result, err := tx.ExecContext(ctx, `UPDATE agent_script_shots SET sequence = ?, scene = ?, characters = ?, dialogue = ?, narration = ?, image_prompt = ?, video_prompt = ?, tts_text = ?, revision = revision + 1, updated_at = ? WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? AND revision = ?`, normalized.Sequence, normalized.Scene, normalized.Characters, normalized.Dialogue, normalized.Narration, normalized.ImagePrompt, normalized.VideoPrompt, normalized.TTSText, normalized.UpdatedAt, normalized.ID, normalized.ProjectID, normalized.ScriptNodeID, owner, normalized.Revision)
	if err != nil {
		return ScriptShot{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return ScriptShot{}, err
	}
	if rows == 0 {
		return ScriptShot{}, ErrRevisionConflict
	}
	normalized.Revision++
	if err := tx.Commit(); err != nil {
		return ScriptShot{}, err
	}
	return normalized, nil
}

func (s *MySQLStore) AttachGeneration(ctx context.Context, owner string, link ShotGenerationLink) error {
	link, err := normalizeShotGenerationLink(owner, link)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	owner = strings.TrimSpace(owner)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, link.ProjectID, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return ErrProjectNotFound
	} else if err != nil {
		return err
	}
	var shotID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE`, link.ShotID, link.ProjectID, link.ScriptNodeID, owner).Scan(&shotID); errors.Is(err, sql.ErrNoRows) {
		return ErrScriptShotNotFound
	} else if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_shot_generation_links (project_id, script_node_id, shot_id, owner_username, node_id, generation_id, attempt, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE script_node_id = VALUES(script_node_id), attempt = VALUES(attempt), active = VALUES(active), updated_at = VALUES(updated_at)`, link.ProjectID, link.ScriptNodeID, link.ShotID, owner, link.NodeID, link.GenerationID, link.Attempt, link.Active, now, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) ListGenerationLinks(ctx context.Context, owner, projectID, shotID string) ([]ShotGenerationLink, error) {
	owner, projectID, shotID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(shotID)
	if owner == "" || projectID == "" || shotID == "" {
		return nil, ErrScriptShotInvalid
	}
	rows, err := s.db.QueryContext(ctx, `SELECT project_id, script_node_id, shot_id, node_id, generation_id, attempt, active FROM agent_shot_generation_links WHERE owner_username = ? AND project_id = ? AND shot_id = ? ORDER BY attempt, node_id, generation_id`, owner, projectID, shotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := make([]ShotGenerationLink, 0)
	for rows.Next() {
		var link ShotGenerationLink
		if err := rows.Scan(&link.ProjectID, &link.ScriptNodeID, &link.ShotID, &link.NodeID, &link.GenerationID, &link.Attempt, &link.Active); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (s *MySQLStore) CreateShotGeneration(ctx context.Context, owner string, input CreateShotGenerationInput) (Canvas, GenerationTask, ShotGenerationLink, error) {
	input, link, err := normalizeCreateShotGeneration(owner, input)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	owner = strings.TrimSpace(owner)
	settingsJSON, err := json.Marshal(input.Task.Settings)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
	}
	referencesJSON, err := json.Marshal(input.Task.ReferenceAssetIDs)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
	}
	sourceJSON, err := nullableJSON(input.Node.SourceNodeIDs)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	assetJSON, err := nullableJSON(input.Node.AssetRef)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	configJSON, err := nullableJSON(input.Node.Config)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	nodeContentJSON, err := json.Marshal(input.Node)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, input.Task.ProjectID, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrProjectNotFound
	} else if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	var revision int64
	var snapshotJSON []byte
	if err := tx.QueryRowContext(ctx, `SELECT revision, canvas_json FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`, input.Task.ProjectID, owner).Scan(&revision, &snapshotJSON); errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrProjectNotFound
	} else if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if revision != input.ExpectedCanvasRevision {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrRevisionConflict
	}
	var canvas Canvas
	if len(snapshotJSON) == 0 || string(snapshotJSON) == "null" || json.Unmarshal(snapshotJSON, &canvas) != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrRevisionConflict
	}
	for _, node := range canvas.Nodes {
		if node.ID == input.Node.ID {
			return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
		}
	}
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE`, link.ShotID, link.ProjectID, link.ScriptNodeID, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrScriptShotNotFound
	} else if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}

	groupID := sql.NullString{String: input.Node.GroupID, Valid: input.Node.GroupID != ""}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_canvas_nodes (id, project_id, owner_username, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, input.Node.ID, input.Task.ProjectID, owner, input.Node.Type, input.Node.Title, input.Node.Summary, input.Node.Position.X, input.Node.Position.Y, input.Node.Version, input.Node.Status, sourceJSON, assetJSON, configJSON, groupID, input.Now, input.Now); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_node_versions (project_id, node_id, owner_username, version, content_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`, input.Task.ProjectID, input.Node.ID, owner, input.Node.Version, string(nodeContentJSON), input.Now); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_canvas_edges (id, project_id, owner_username, source_node_id, target_node_id, relation_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, input.Edge.ID, input.Task.ProjectID, owner, input.Edge.Source, input.Edge.Target, input.Edge.Relation, input.Now, input.Now); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	canvas.Nodes = append(canvas.Nodes, input.Node)
	canvas.Edges = append(canvas.Edges, input.Edge)
	canvas.Version = 1
	canvas.Revision = revision + 1
	canvasSnapshot, err := nullableJSON(canvas)
	if err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_canvases SET revision = ?, canvas_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND revision = ?`, canvas.Revision, canvasSnapshot, input.Now, input.Task.ProjectID, owner, revision); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	task := generationTaskFromInput(owner, input.Task)
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_generation_tasks (id, project_id, owner_username, node_id, node_version, kind, provider, model_id, mode, state, settings_json, reference_asset_ids_json, selected_result_id, idempotency_key, retry_of_task_id, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, task.ID, task.ProjectID, owner, task.NodeID, task.NodeVersion, task.Kind, task.Provider, task.ModelID, nullableString(task.Mode), task.State, string(settingsJSON), string(referencesJSON), nil, task.IdempotencyKey, nullableString(task.RetryOfTaskID), nil, task.CreatedAt, task.UpdatedAt); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return Canvas{}, GenerationTask{}, ShotGenerationLink{}, ErrGenerationInvalid
		}
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_shot_generation_links (project_id, script_node_id, shot_id, owner_username, node_id, generation_id, attempt, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, link.ProjectID, link.ScriptNodeID, link.ShotID, owner, link.NodeID, link.GenerationID, link.Attempt, link.Active, input.Now, input.Now); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, input.Now, input.Task.ProjectID, owner); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	if err := tx.Commit(); err != nil {
		return Canvas{}, GenerationTask{}, ShotGenerationLink{}, err
	}
	return canvas, task, link, nil
}
