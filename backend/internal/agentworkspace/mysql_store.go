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

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func nullableJSON(value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	raw, err := json.Marshal(value)
	if err != nil || string(raw) == "null" {
		return nil, err
	}
	return string(raw), nil
}

func (s *MySQLStore) ListProjects(ctx context.Context, owner string) ([]ProjectSummary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.id, p.title, p.title_mode, p.preview, p.created_at, p.updated_at,
 (SELECT COUNT(*) FROM agent_messages m WHERE m.project_id = p.id AND m.owner_username = p.owner_username)
FROM agent_projects p WHERE p.owner_username = ? AND p.deleted_at IS NULL ORDER BY p.updated_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ProjectSummary, 0)
	for rows.Next() {
		var item ProjectSummary
		if err := rows.Scan(&item.ID, &item.Title, &item.TitleMode, &item.Preview, &item.CreatedAt, &item.UpdatedAt, &item.MessageCount); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) Project(ctx context.Context, owner, id string) (Project, error) {
	var project Project
	err := s.db.QueryRowContext(ctx, `SELECT id, title, title_mode, preview, created_at, updated_at
FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, id, owner).
		Scan(&project.ID, &project.Title, &project.TitleMode, &project.Preview, &project.CreatedAt, &project.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}
	project.Messages, err = s.messages(ctx, owner, id)
	if err != nil {
		return Project{}, err
	}
	project.Canvas, err = s.Canvas(ctx, owner, id)
	if err != nil {
		return Project{}, err
	}
	project.Executions, err = s.Executions(ctx, owner, id)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *MySQLStore) messages(ctx context.Context, owner, id string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, role, content, context_json, created_at FROM agent_messages
WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`, id, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Message, 0)
	for rows.Next() {
		var item Message
		var contextJSON []byte
		if err := rows.Scan(&item.ID, &item.Role, &item.Content, &contextJSON, &item.CreatedAt); err != nil {
			return nil, err
		}
		if len(contextJSON) > 0 && json.Unmarshal(contextJSON, &item.Context) != nil {
			return nil, errors.New("invalid agent message context")
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) CreateProject(ctx context.Context, owner string, project Project) error {
	project = normalizeProject(project)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := insertProject(ctx, tx, owner, project); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrProjectExists
		}
		return err
	}
	return tx.Commit()
}

func insertProject(ctx context.Context, tx *sql.Tx, owner string, project Project) error {
	project.Canvas.Version = 1
	project.Canvas.Appearance = normalizeCanvasAppearance(project.Canvas.Appearance, project.Canvas.appearanceConfigured)
	project.Canvas.appearanceConfigured = true
	canvasJSON, err := nullableJSON(project.Canvas)
	if err != nil {
		return err
	}
	viewportJSON, err := nullableJSON(project.Canvas.Viewport)
	if err != nil {
		return err
	}
	appearanceJSON, err := nullableJSON(project.Canvas.Appearance)
	if err != nil {
		return err
	}
	viewStateJSON, err := nullableJSON(project.Canvas.ViewState)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_projects (id, owner_username, title, title_mode, preview, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		project.ID, owner, project.Title, project.TitleMode, project.Preview, project.CreatedAt, project.UpdatedAt); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO agent_canvases (project_id, owner_username, canvas_version, revision, canvas_json, viewport_json, appearance_json, view_state_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		project.ID, owner, project.Canvas.Version, project.Canvas.Revision, canvasJSON, viewportJSON, appearanceJSON, viewStateJSON, project.CreatedAt, project.UpdatedAt); err != nil {
		return err
	}
	for _, message := range project.Messages {
		contextJSON, marshalErr := nullableJSON(message.Context)
		if marshalErr != nil {
			return marshalErr
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_messages (project_id, owner_username, role, content, context_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`, project.ID, owner, message.Role, message.Content, contextJSON, message.CreatedAt); err != nil {
			return err
		}
	}
	if len(project.Canvas.Nodes) > 0 || len(project.Canvas.Edges) > 0 {
		if err := insertCanvasItems(ctx, tx, owner, project.ID, project.Canvas, project.UpdatedAt); err != nil {
			return err
		}
	}
	for _, execution := range project.Executions {
		detailJSON, marshalErr := nullableJSON(execution.Detail)
		if marshalErr != nil {
			return marshalErr
		}
		var idempotency any
		if execution.IdempotencyKey != "" {
			idempotency = execution.IdempotencyKey
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_executions (id, project_id, owner_username, tool_name, status, detail_json, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, execution.ID, project.ID, owner, execution.ToolName, execution.Status, detailJSON, idempotency, execution.CreatedAt, execution.UpdatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) RenameProject(ctx context.Context, owner, id, title string, now time.Time) (Project, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE agent_projects SET title = ?, title_mode = 'manual', updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, strings.TrimSpace(title), now, id, owner)
	if err != nil {
		return Project{}, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return Project{}, ErrProjectNotFound
	}
	return s.Project(ctx, owner, id)
}

func (s *MySQLStore) ClearMessages(ctx context.Context, owner, id string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE agent_projects SET preview = '', updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now, id, owner)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrProjectNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_messages WHERE project_id = ? AND owner_username = ?`, id, owner); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) DeleteProject(ctx context.Context, owner, id string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE agent_projects SET deleted_at = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now, now, id, owner)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (s *MySQLStore) AppendExchange(ctx context.Context, owner, id string, user, assistant Message) (Project, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Project{}, err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, id, owner).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	} else if err != nil {
		return Project{}, err
	}
	for _, message := range []Message{user, assistant} {
		contextJSON, marshalErr := nullableJSON(message.Context)
		if marshalErr != nil {
			return Project{}, marshalErr
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_messages (project_id, owner_username, role, content, context_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`, id, owner, message.Role, message.Content, contextJSON, message.CreatedAt); err != nil {
			return Project{}, err
		}
	}
	preview := strings.TrimSpace(assistant.Content)
	if len(preview) > 512 {
		preview = preview[:512]
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET preview = ?, updated_at = ? WHERE id = ? AND owner_username = ?`, preview, assistant.CreatedAt, id, owner); err != nil {
		return Project{}, err
	}
	if err := tx.Commit(); err != nil {
		return Project{}, err
	}
	return s.Project(ctx, owner, id)
}

func (s *MySQLStore) Canvas(ctx context.Context, owner, id string) (Canvas, error) {
	var canvas Canvas
	var canvasJSON, viewportJSON, appearanceJSON, viewStateJSON []byte
	err := s.db.QueryRowContext(ctx, `SELECT canvas_version, revision, canvas_json, viewport_json, appearance_json, view_state_json FROM agent_canvases WHERE project_id = ? AND owner_username = ?`, id, owner).
		Scan(&canvas.Version, &canvas.Revision, &canvasJSON, &viewportJSON, &appearanceJSON, &viewStateJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, ErrProjectNotFound
	}
	if err != nil {
		return Canvas{}, err
	}
	if len(canvasJSON) > 0 && string(canvasJSON) != "null" {
		var snapshot Canvas
		if err := json.Unmarshal(canvasJSON, &snapshot); err != nil {
			return Canvas{}, errors.New("invalid agent canvas snapshot")
		}
		snapshot.Version = canvas.Version
		snapshot.Revision = canvas.Revision
		snapshot.Appearance = normalizeCanvasAppearance(snapshot.Appearance, snapshot.appearanceConfigured)
		snapshot.appearanceConfigured = true
		if snapshot.Nodes == nil {
			snapshot.Nodes = []Node{}
		}
		if snapshot.Edges == nil {
			snapshot.Edges = []Edge{}
		}
		return snapshot, nil
	}
	if len(viewportJSON) > 0 && json.Unmarshal(viewportJSON, &canvas.Viewport) != nil {
		return Canvas{}, errors.New("invalid agent canvas viewport")
	}
	if len(viewStateJSON) > 0 && json.Unmarshal(viewStateJSON, &canvas.ViewState) != nil {
		return Canvas{}, errors.New("invalid agent canvas view state")
	}
	canvas.appearanceConfigured = len(appearanceJSON) > 0 && string(appearanceJSON) != "null"
	if canvas.appearanceConfigured && json.Unmarshal(appearanceJSON, &canvas.Appearance) != nil {
		return Canvas{}, errors.New("invalid agent canvas appearance")
	}
	canvas.Appearance = normalizeCanvasAppearance(canvas.Appearance, canvas.appearanceConfigured)
	canvas.appearanceConfigured = true
	rows, err := s.db.QueryContext(ctx, `SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id
FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`, id, owner)
	if err != nil {
		return Canvas{}, err
	}
	canvas.Nodes = make([]Node, 0)
	for rows.Next() {
		var node Node
		var sourceJSON, assetJSON, configJSON []byte
		var groupID sql.NullString
		if err := rows.Scan(&node.ID, &node.Type, &node.Title, &node.Summary, &node.Position.X, &node.Position.Y, &node.Version, &node.Status, &sourceJSON, &assetJSON, &configJSON, &groupID); err != nil {
			rows.Close()
			return Canvas{}, err
		}
		if groupID.Valid {
			node.GroupID = groupID.String
		}
		if len(sourceJSON) > 0 {
			_ = json.Unmarshal(sourceJSON, &node.SourceNodeIDs)
		}
		if len(assetJSON) > 0 {
			_ = json.Unmarshal(assetJSON, &node.AssetRef)
		}
		if len(configJSON) > 0 {
			_ = json.Unmarshal(configJSON, &node.Config)
		}
		canvas.Nodes = append(canvas.Nodes, node)
	}
	if err := rows.Close(); err != nil {
		return Canvas{}, err
	}
	edgeRows, err := s.db.QueryContext(ctx, `SELECT id, source_node_id, target_node_id, relation_type FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`, id, owner)
	if err != nil {
		return Canvas{}, err
	}
	defer edgeRows.Close()
	canvas.Edges = make([]Edge, 0)
	for edgeRows.Next() {
		var edge Edge
		if err := edgeRows.Scan(&edge.ID, &edge.Source, &edge.Target, &edge.Relation); err != nil {
			return Canvas{}, err
		}
		canvas.Edges = append(canvas.Edges, edge)
	}
	return canvas, edgeRows.Err()
}

func (s *MySQLStore) SaveCanvas(ctx context.Context, owner, id string, canvas Canvas, expected int64, now time.Time) (Canvas, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Canvas{}, err
	}
	defer tx.Rollback()
	var revision int64
	err = tx.QueryRowContext(ctx, `SELECT revision FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`, id, owner).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, ErrProjectNotFound
	}
	if err != nil {
		return Canvas{}, err
	}
	if revision != expected {
		return Canvas{}, ErrRevisionConflict
	}
	canvas.Version = 1
	canvas.Revision = revision + 1
	canvas.Appearance = normalizeCanvasAppearance(canvas.Appearance, canvas.appearanceConfigured)
	canvas.appearanceConfigured = true
	canvasJSON, err := nullableJSON(canvas)
	if err != nil {
		return Canvas{}, err
	}
	viewportJSON, err := nullableJSON(canvas.Viewport)
	if err != nil {
		return Canvas{}, err
	}
	appearanceJSON, err := nullableJSON(canvas.Appearance)
	if err != nil {
		return Canvas{}, err
	}
	viewStateJSON, err := nullableJSON(canvas.ViewState)
	if err != nil {
		return Canvas{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_canvases SET canvas_version = ?, revision = ?, canvas_json = ?, viewport_json = ?, appearance_json = ?, view_state_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`, canvas.Version, canvas.Revision, canvasJSON, viewportJSON, appearanceJSON, viewStateJSON, now, id, owner); err != nil {
		return Canvas{}, err
	}
	if err := syncCanvasNodeIndex(ctx, tx, owner, id, canvas, now); err != nil {
		return Canvas{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now, id, owner); err != nil {
		return Canvas{}, err
	}
	if err := tx.Commit(); err != nil {
		return Canvas{}, err
	}
	return canvas, nil
}

// Nodes remain a small compatibility index for generation workers. Canvas JSON is the source of truth.
func syncCanvasNodeIndex(ctx context.Context, tx *sql.Tx, owner, id string, canvas Canvas, now time.Time) error {
	for _, node := range canvas.Nodes {
		sourceJSON, err := nullableJSON(node.SourceNodeIDs)
		if err != nil {
			return err
		}
		assetJSON, err := nullableJSON(node.AssetRef)
		if err != nil {
			return err
		}
		configJSON, err := nullableJSON(node.Config)
		if err != nil {
			return err
		}
		groupID := sql.NullString{String: node.GroupID, Valid: node.GroupID != ""}
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_canvas_nodes (id, project_id, owner_username, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE node_type = VALUES(node_type), title = VALUES(title), summary = VALUES(summary), position_x = VALUES(position_x), position_y = VALUES(position_y), current_version = VALUES(current_version), status = VALUES(status), source_node_ids_json = VALUES(source_node_ids_json), asset_ref_json = VALUES(asset_ref_json), config_json = VALUES(config_json), group_id = VALUES(group_id), updated_at = VALUES(updated_at)`, node.ID, id, owner, node.Type, node.Title, node.Summary, node.Position.X, node.Position.Y, node.Version, node.Status, sourceJSON, assetJSON, configJSON, groupID, now, now); err != nil {
			return err
		}
	}
	return nil
}

func updateCanvasSnapshotNode(ctx context.Context, tx *sql.Tx, owner, projectID string, nodeID string, now time.Time, mutate func(*Node)) error {
	var revision int64
	var snapshotJSON []byte
	if err := tx.QueryRowContext(ctx, `SELECT revision, canvas_json FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`, projectID, owner).Scan(&revision, &snapshotJSON); errors.Is(err, sql.ErrNoRows) {
		return ErrProjectNotFound
	} else if err != nil {
		return err
	}
	if len(snapshotJSON) == 0 || string(snapshotJSON) == "null" {
		_, err := tx.ExecContext(ctx, `UPDATE agent_canvases SET revision = revision + 1, updated_at = ? WHERE project_id = ? AND owner_username = ?`, now, projectID, owner)
		return err
	}
	var canvas Canvas
	if err := json.Unmarshal(snapshotJSON, &canvas); err != nil {
		return errors.New("invalid agent canvas snapshot")
	}
	found := false
	for index := range canvas.Nodes {
		if canvas.Nodes[index].ID == nodeID {
			mutate(&canvas.Nodes[index])
			found = true
			break
		}
	}
	if !found {
		return ErrExecutionNodeChanged
	}
	canvas.Version = 1
	canvas.Revision = revision + 1
	canvas.Appearance = normalizeCanvasAppearance(canvas.Appearance, canvas.appearanceConfigured)
	canvas.appearanceConfigured = true
	updatedJSON, err := nullableJSON(canvas)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE agent_canvases SET revision = ?, canvas_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`, canvas.Revision, updatedJSON, now, projectID, owner)
	return err
}

func insertCanvasItems(ctx context.Context, tx *sql.Tx, owner, id string, canvas Canvas, now time.Time) error {
	for _, node := range canvas.Nodes {
		sourceJSON, err := nullableJSON(node.SourceNodeIDs)
		if err != nil {
			return err
		}
		assetJSON, err := nullableJSON(node.AssetRef)
		if err != nil {
			return err
		}
		configJSON, err := nullableJSON(node.Config)
		if err != nil {
			return err
		}
		groupID := sql.NullString{String: node.GroupID, Valid: node.GroupID != ""}
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_canvas_nodes (id, project_id, owner_username, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, node.ID, id, owner, node.Type, node.Title, node.Summary, node.Position.X, node.Position.Y, node.Version, node.Status, sourceJSON, assetJSON, configJSON, groupID, now, now); err != nil {
			return err
		}
		contentJSON, err := json.Marshal(node)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO agent_node_versions (project_id, node_id, owner_username, version, content_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`, id, node.ID, owner, node.Version, string(contentJSON), now); err != nil {
			return err
		}
	}
	for _, edge := range canvas.Edges {
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_canvas_edges (id, project_id, owner_username, source_node_id, target_node_id, relation_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, edge.ID, id, owner, edge.Source, edge.Target, edge.Relation, now, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *MySQLStore) Executions(ctx context.Context, owner, id string) ([]Execution, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, tool_name, status, detail_json, idempotency_key, created_at, updated_at FROM agent_executions
WHERE project_id = ? AND owner_username = ? ORDER BY updated_at DESC`, id, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Execution, 0)
	for rows.Next() {
		var item Execution
		var detailJSON []byte
		var idempotency sql.NullString
		if err := rows.Scan(&item.ID, &item.ToolName, &item.Status, &detailJSON, &idempotency, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.IdempotencyKey = idempotency.String
		if len(detailJSON) > 0 {
			_ = json.Unmarshal(detailJSON, &item.Detail)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) BeginNodeExecution(ctx context.Context, owner, projectID string, execution Execution) (Execution, error) {
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	execution.ID = strings.TrimSpace(execution.ID)
	execution.ToolName = strings.TrimSpace(execution.ToolName)
	execution.IdempotencyKey = strings.TrimSpace(execution.IdempotencyKey)
	if owner == "" || projectID == "" || execution.ID == "" || execution.ToolName == "" || execution.IdempotencyKey == "" {
		return Execution{}, ErrExecutionState
	}
	execution.Status = "running"
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Execution{}, err
	}
	defer tx.Rollback()
	var activeProjectID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, projectID, owner).Scan(&activeProjectID); errors.Is(err, sql.ErrNoRows) {
		return Execution{}, ErrProjectNotFound
	} else if err != nil {
		return Execution{}, err
	}
	detailJSON, err := nullableJSON(execution.Detail)
	if err != nil {
		return Execution{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO agent_executions (id, project_id, owner_username, tool_name, status, detail_json, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, execution.ID, projectID, owner, execution.ToolName, execution.Status, detailJSON, execution.IdempotencyKey, execution.CreatedAt, execution.UpdatedAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return Execution{}, ErrExecutionExists
		}
		return Execution{}, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, err
	}
	return execution, nil
}

func (s *MySQLStore) SetNodeExecutionAwaitingConfirmation(ctx context.Context, owner, projectID, executionID string, detail map[string]any, now time.Time) (Execution, error) {
	owner, projectID, executionID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(executionID)
	detailJSON, err := nullableJSON(detail)
	if err != nil {
		return Execution{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Execution{}, err
	}
	defer tx.Rollback()
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM agent_executions WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, executionID, projectID, owner).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return Execution{}, ErrExecutionNotFound
	} else if err != nil {
		return Execution{}, err
	}
	if status != "running" {
		return Execution{}, ErrExecutionState
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_executions SET status = ?, detail_json = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, "awaiting_confirmation", detailJSON, now.UTC(), executionID, projectID, owner); err != nil {
		return Execution{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now.UTC(), projectID, owner); err != nil {
		return Execution{}, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, err
	}
	return Execution{ID: executionID, Status: "awaiting_confirmation", Detail: cloneMap(detail), UpdatedAt: now.UTC()}, nil
}

func (s *MySQLStore) ApplyNodeExecution(ctx context.Context, owner, projectID string, input ApplyNodeExecutionInput) (Canvas, Execution, error) {
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Canvas{}, Execution{}, err
	}
	defer tx.Rollback()
	var executionStatus string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM agent_executions WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, input.ExecutionID, projectID, owner).Scan(&executionStatus); errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, Execution{}, ErrExecutionNotFound
	} else if err != nil {
		return Canvas{}, Execution{}, err
	}
	if executionStatus != "running" && !(input.NodeType == "image" && executionStatus == "awaiting_confirmation") {
		return Canvas{}, Execution{}, ErrExecutionState
	}
	var nodeType string
	var nodeVersion int64
	if err := tx.QueryRowContext(ctx, `SELECT node_type, current_version FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`, projectID, owner, input.NodeID).Scan(&nodeType, &nodeVersion); errors.Is(err, sql.ErrNoRows) {
		return Canvas{}, Execution{}, ErrExecutionNodeChanged
	} else if err != nil {
		return Canvas{}, Execution{}, err
	}
	if nodeType != input.NodeType || nodeVersion != input.NodeVersion {
		return Canvas{}, Execution{}, ErrExecutionNodeChanged
	}
	assetJSON, err := nullableJSON(input.AssetRef)
	if err != nil {
		return Canvas{}, Execution{}, err
	}
	switch input.NodeType {
	case "text":
		if strings.TrimSpace(input.Summary) == "" {
			return Canvas{}, Execution{}, ErrExecutionState
		}
		_, err = tx.ExecContext(ctx, `UPDATE agent_canvas_nodes SET summary = ?, current_version = current_version + 1, status = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, strings.TrimSpace(input.Summary), "confirmed", input.Now, projectID, owner, input.NodeID)
	case "image":
		if strings.TrimSpace(stringValue(input.AssetRef["id"])) == "" || stringValue(input.AssetRef["kind"]) != "image" {
			return Canvas{}, Execution{}, ErrExecutionState
		}
		_, err = tx.ExecContext(ctx, `UPDATE agent_canvas_nodes SET asset_ref_json = ?, current_version = current_version + 1, status = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`, assetJSON, "confirmed", input.Now, projectID, owner, input.NodeID)
	default:
		return Canvas{}, Execution{}, ErrExecutionState
	}
	if err != nil {
		return Canvas{}, Execution{}, err
	}
	if err := updateCanvasSnapshotNode(ctx, tx, owner, projectID, input.NodeID, input.Now, func(node *Node) {
		node.Status = "confirmed"
		node.Version++
		if input.NodeType == "text" {
			node.Summary = strings.TrimSpace(input.Summary)
		} else {
			node.AssetRef = cloneMap(input.AssetRef)
		}
	}); err != nil {
		return Canvas{}, Execution{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, input.Now, projectID, owner); err != nil {
		return Canvas{}, Execution{}, err
	}
	detailJSON, err := nullableJSON(input.Detail)
	if err != nil {
		return Canvas{}, Execution{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_executions SET status = ?, detail_json = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, "completed", detailJSON, input.Now, input.ExecutionID, projectID, owner); err != nil {
		return Canvas{}, Execution{}, err
	}
	if err := tx.Commit(); err != nil {
		return Canvas{}, Execution{}, err
	}
	canvas, err := s.Canvas(ctx, owner, projectID)
	if err != nil {
		return Canvas{}, Execution{}, err
	}
	return canvas, Execution{ID: input.ExecutionID, Status: "completed", Detail: cloneMap(input.Detail), UpdatedAt: input.Now}, nil
}

func (s *MySQLStore) FailNodeExecution(ctx context.Context, owner, projectID, executionID string, detail map[string]any, now time.Time) (Execution, error) {
	detailJSON, err := nullableJSON(detail)
	if err != nil {
		return Execution{}, err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE agent_executions SET status = ?, detail_json = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ? AND status = ?`, "failed", detailJSON, now, strings.TrimSpace(executionID), strings.TrimSpace(projectID), strings.TrimSpace(owner), "running")
	if err != nil {
		return Execution{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Execution{}, err
	}
	if affected == 0 {
		return Execution{}, ErrExecutionNotFound
	}
	return Execution{ID: strings.TrimSpace(executionID), Status: "failed", Detail: cloneMap(detail), UpdatedAt: now}, nil
}

func (s *MySQLStore) ImportProject(ctx context.Context, owner string, project Project) error {
	return s.CreateProject(ctx, owner, project)
}

func (s *MySQLStore) LegacyImport(ctx context.Context, owner, path string) (LegacyImport, bool, error) {
	var record LegacyImport
	err := s.db.QueryRowContext(ctx, `SELECT owner_username, source_path, source_sha256, imported_projects, imported_at FROM agent_legacy_imports WHERE owner_username = ? AND source_path = ?`, owner, path).
		Scan(&record.OwnerUsername, &record.SourcePath, &record.SourceSHA256, &record.ImportedProjects, &record.ImportedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return LegacyImport{}, false, nil
	}
	return record, err == nil, err
}

func (s *MySQLStore) RecordLegacyImport(ctx context.Context, record LegacyImport) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO agent_legacy_imports (owner_username, source_path, source_sha256, imported_projects, imported_at) VALUES (?, ?, ?, ?, ?)`, record.OwnerUsername, record.SourcePath, record.SourceSHA256, record.ImportedProjects, record.ImportedAt)
	return err
}
