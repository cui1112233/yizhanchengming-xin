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

func (s *MySQLStore) CreateVideoJob(ctx context.Context, owner string, input CreateVideoJobInput) (VideoJob, error) {
	input, err := normalizeCreateVideoJob(input)
	if err != nil {
		return VideoJob{}, err
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return VideoJob{}, ErrVideoJobInvalid
	}
	settingsJSON, err := json.Marshal(input.Settings)
	if err != nil {
		return VideoJob{}, ErrVideoJobInvalid
	}
	referencesJSON, err := json.Marshal(input.ReferenceAssetIDs)
	if err != nil {
		return VideoJob{}, ErrVideoJobInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VideoJob{}, err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, input.ProjectID, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrProjectNotFound
	} else if err != nil {
		return VideoJob{}, err
	}
	job := VideoJob{ID: input.ID, ProjectID: input.ProjectID, OwnerUsername: owner, NodeID: input.NodeID, NodeVersion: input.NodeVersion, State: VideoJobPreparing, ModelKey: input.ModelKey, Settings: input.Settings, ReferenceAssetIDs: input.ReferenceAssetIDs, IdempotencyKey: input.IdempotencyKey, CreatedAt: input.Now.UTC(), UpdatedAt: input.Now.UTC()}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_video_jobs (id, project_id, owner_username, node_id, node_version, state, model_key, settings_json, reference_asset_ids_json, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, job.ID, job.ProjectID, owner, job.NodeID, job.NodeVersion, job.State, job.ModelKey, string(settingsJSON), string(referencesJSON), job.IdempotencyKey, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return VideoJob{}, ErrVideoJobExists
		}
		return VideoJob{}, err
	}
	if err := tx.Commit(); err != nil {
		return VideoJob{}, err
	}
	return job, nil
}

func scanVideoJob(scanner interface{ Scan(...any) error }) (VideoJob, error) {
	var job VideoJob
	var settingsJSON, referencesJSON []byte
	var providerTaskID, resultAssetID, errorMessage sql.NullString
	err := scanner.Scan(&job.ID, &job.ProjectID, &job.OwnerUsername, &job.NodeID, &job.NodeVersion, &job.State, &job.ModelKey, &providerTaskID, &settingsJSON, &referencesJSON, &resultAssetID, &errorMessage, &job.IdempotencyKey, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return VideoJob{}, err
	}
	job.ProviderTaskID = providerTaskID.String
	job.ResultAssetID = resultAssetID.String
	job.ErrorMessage = errorMessage.String
	if json.Unmarshal(settingsJSON, &job.Settings) != nil || json.Unmarshal(referencesJSON, &job.ReferenceAssetIDs) != nil || !validVideoJobState(job.State) {
		return VideoJob{}, ErrVideoJobInvalid
	}
	if job.Settings == nil {
		job.Settings = map[string]any{}
	}
	if job.ReferenceAssetIDs == nil {
		job.ReferenceAssetIDs = []string{}
	}
	return job, nil
}

const videoJobColumns = `id, project_id, owner_username, node_id, node_version, state, model_key, provider_task_id, settings_json, reference_asset_ids_json, result_asset_id, error_message, idempotency_key, created_at, updated_at`

func (s *MySQLStore) VideoJob(ctx context.Context, owner, projectID, id string) (VideoJob, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ?`, strings.TrimSpace(id), strings.TrimSpace(projectID), strings.TrimSpace(owner))
	job, err := scanVideoJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	return job, err
}

func (s *MySQLStore) ListVideoJobs(ctx context.Context, owner, projectID string) ([]VideoJob, error) {
	if _, err := s.Project(ctx, owner, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE project_id = ? AND owner_username = ? ORDER BY updated_at DESC, id DESC`, strings.TrimSpace(projectID), strings.TrimSpace(owner))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]VideoJob, 0)
	for rows.Next() {
		job, scanErr := scanVideoJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, job)
	}
	return out, rows.Err()
}

func (s *MySQLStore) UpdateVideoJob(ctx context.Context, owner string, input VideoJob, now time.Time) (VideoJob, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.ProjectID) == "" || !validVideoJobState(input.State) || now.IsZero() || input.UpdatedAt.IsZero() {
		return VideoJob{}, ErrVideoJobInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VideoJob{}, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, input.ID, input.ProjectID, owner)
	current, err := scanVideoJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	if err != nil {
		return VideoJob{}, err
	}
	if !normalizeVideoJobTime(input.UpdatedAt).Equal(normalizeVideoJobTime(current.UpdatedAt)) || !validVideoJobTransition(current.State, input.State) {
		return VideoJob{}, ErrVideoJobState
	}
	current.State = input.State
	if strings.TrimSpace(input.ProviderTaskID) != "" {
		current.ProviderTaskID = strings.TrimSpace(input.ProviderTaskID)
	}
	if strings.TrimSpace(input.ResultAssetID) != "" {
		current.ResultAssetID = strings.TrimSpace(input.ResultAssetID)
	}
	current.ErrorMessage = strings.TrimSpace(input.ErrorMessage)
	if input.Settings != nil {
		current.Settings = input.Settings
	}
	current.UpdatedAt = nextVideoJobUpdatedAt(current.UpdatedAt, now)
	settingsJSON, marshalErr := json.Marshal(current.Settings)
	if marshalErr != nil {
		return VideoJob{}, ErrVideoJobInvalid
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_video_jobs SET state = ?, provider_task_id = ?, settings_json = ?, result_asset_id = ?, error_message = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, current.State, nullableString(current.ProviderTaskID), string(settingsJSON), nullableString(current.ResultAssetID), nullableString(current.ErrorMessage), current.UpdatedAt, current.ID, current.ProjectID, owner); err != nil {
		return VideoJob{}, err
	}
	if err := tx.Commit(); err != nil {
		return VideoJob{}, err
	}
	return current, nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func (s *MySQLStore) ConfirmVideoJob(ctx context.Context, owner, projectID, id string, now time.Time) (VideoJob, error) {
	now = normalizeVideoJobTime(now)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VideoJob{}, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, strings.TrimSpace(id), strings.TrimSpace(projectID), strings.TrimSpace(owner))
	job, err := scanVideoJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	if err != nil {
		return VideoJob{}, err
	}
	if job.State != VideoJobAwaitingConfirmation || job.ResultAssetID == "" {
		return VideoJob{}, ErrVideoJobState
	}
	var asset Asset
	err = tx.QueryRowContext(ctx, `SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at FROM agent_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, job.ResultAssetID, job.ProjectID, owner).Scan(&asset.ID, &asset.ProjectID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobState
	}
	if err != nil || asset.Kind != "video" {
		if err != nil {
			return VideoJob{}, err
		}
		return VideoJob{}, ErrVideoJobState
	}
	var node Node
	var sourceJSON, assetJSON, configJSON []byte
	var groupID sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, job.ProjectID, owner, job.NodeID).Scan(&node.ID, &node.Type, &node.Title, &node.Summary, &node.Position.X, &node.Position.Y, &node.Version, &node.Status, &sourceJSON, &assetJSON, &configJSON, &groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobNodeChanged
	}
	if err != nil {
		return VideoJob{}, err
	}
	if node.Type != "video" || node.Version != job.NodeVersion {
		return VideoJob{}, ErrVideoJobNodeChanged
	}
	_ = json.Unmarshal(sourceJSON, &node.SourceNodeIDs)
	_ = json.Unmarshal(configJSON, &node.Config)
	if groupID.Valid {
		node.GroupID = groupID.String
	}
	node.AssetRef = map[string]any{"id": asset.ID, "projectId": asset.ProjectID, "kind": asset.Kind, "name": asset.Name, "contentType": asset.ContentType, "byteSize": asset.ByteSize}
	node.Status = "confirmed"
	node.Version++
	assetRefJSON, marshalErr := json.Marshal(node.AssetRef)
	if marshalErr != nil {
		return VideoJob{}, marshalErr
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_canvas_nodes SET asset_ref_json = ?, status = ?, current_version = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, string(assetRefJSON), node.Status, node.Version, now.UTC(), job.ProjectID, owner, node.ID); err != nil {
		return VideoJob{}, err
	}
	contentJSON, marshalErr := json.Marshal(node)
	if marshalErr != nil {
		return VideoJob{}, marshalErr
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_node_versions (project_id, node_id, owner_username, version, content_json, source_execution_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, job.ProjectID, node.ID, owner, node.Version, string(contentJSON), job.ID, now.UTC()); err != nil {
		return VideoJob{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_canvases SET revision = revision + 1, updated_at = ? WHERE project_id = ? AND owner_username = ?`, now.UTC(), job.ProjectID, owner); err != nil {
		return VideoJob{}, err
	}
	job.State = VideoJobApplied
	job.UpdatedAt = now
	if _, err := tx.ExecContext(ctx, `UPDATE agent_video_jobs SET state = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, job.State, job.UpdatedAt, job.ID, job.ProjectID, owner); err != nil {
		return VideoJob{}, err
	}
	if err := tx.Commit(); err != nil {
		return VideoJob{}, err
	}
	return job, nil
}

func (s *MySQLStore) RetryVideoJob(ctx context.Context, owner, projectID, id string, input RetryVideoJobInput) (VideoJob, error) {
	owner, projectID, id = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(id)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VideoJob{}, err
	}
	defer tx.Rollback()
	current, err := scanVideoJob(tx.QueryRowContext(ctx, `SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, id, projectID, owner))
	if errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	if err != nil {
		return VideoJob{}, err
	}
	if !isRetryableVideoJob(current.State) {
		return VideoJob{}, ErrVideoJobState
	}
	var activeProjectID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, projectID, owner).Scan(&activeProjectID); errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrProjectNotFound
	} else if err != nil {
		return VideoJob{}, err
	}
	var nodeType string
	var nodeVersion int64
	if err := tx.QueryRowContext(ctx, `SELECT node_type, current_version FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, projectID, owner, current.NodeID).Scan(&nodeType, &nodeVersion); errors.Is(err, sql.ErrNoRows) {
		return VideoJob{}, ErrVideoJobNodeChanged
	} else if err != nil {
		return VideoJob{}, err
	}
	if nodeType != "video" || nodeVersion != current.NodeVersion {
		return VideoJob{}, ErrVideoJobNodeChanged
	}
	create, err := normalizeCreateVideoJob(CreateVideoJobInput{ID: input.ID, ProjectID: projectID, NodeID: current.NodeID, NodeVersion: current.NodeVersion, ModelKey: current.ModelKey, Settings: current.Settings, ReferenceAssetIDs: current.ReferenceAssetIDs, IdempotencyKey: input.IdempotencyKey, Now: input.Now})
	if err != nil || owner == "" {
		return VideoJob{}, ErrVideoJobInvalid
	}
	settingsJSON, err := json.Marshal(create.Settings)
	if err != nil {
		return VideoJob{}, ErrVideoJobInvalid
	}
	referencesJSON, err := json.Marshal(create.ReferenceAssetIDs)
	if err != nil {
		return VideoJob{}, ErrVideoJobInvalid
	}
	job := VideoJob{ID: create.ID, ProjectID: create.ProjectID, OwnerUsername: owner, NodeID: create.NodeID, NodeVersion: create.NodeVersion, State: VideoJobPreparing, ModelKey: create.ModelKey, Settings: create.Settings, ReferenceAssetIDs: create.ReferenceAssetIDs, IdempotencyKey: create.IdempotencyKey, CreatedAt: create.Now.UTC(), UpdatedAt: create.Now.UTC()}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_video_jobs (id, project_id, owner_username, node_id, node_version, state, model_key, settings_json, reference_asset_ids_json, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, job.ID, job.ProjectID, owner, job.NodeID, job.NodeVersion, job.State, job.ModelKey, string(settingsJSON), string(referencesJSON), job.IdempotencyKey, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return VideoJob{}, ErrVideoJobExists
		}
		return VideoJob{}, err
	}
	if err := tx.Commit(); err != nil {
		return VideoJob{}, err
	}
	return job, nil
}

var _ VideoJobStore = (*MySQLStore)(nil)
