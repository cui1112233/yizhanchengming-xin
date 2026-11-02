package agentworkspace

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func TestMySQLStoreCreatesProjectAndEmptyCanvasAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 18, 1, 2, 3, 0, time.UTC)
	appearanceJSON := `{"edgeColor":"neutral","edgeAnimated":false,"edgeWidth":"thin","backgroundDots":true,"alignmentGuides":true,"edgesVisible":true}`

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_projects (id, owner_username, title, title_mode, project_type, preview, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("project-1", "alice", "项目", "automatic", "video", "", now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvases (project_id, owner_username, canvas_version, revision, canvas_json, viewport_json, appearance_json, view_state_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("project-1", "alice", 1, int64(0), sqlmock.AnyArg(), nil, appearanceJSON, nil, now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = store.CreateProject(context.Background(), "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCreateProjectRejectsUnsafeCanvasBeforeWrites(t *testing.T) {
	now := time.Date(2026, 10, 23, 16, 30, 0, 0, time.UTC)
	for name, canvas := range map[string]Canvas{
		"forbidden node type": {Nodes: []Node{{ID: "asset-1", Type: "asset"}}},
		"legacy input type":   {Nodes: []Node{{ID: "input-1", Type: "input"}}},
		"private object key":  {Nodes: []Node{{ID: "image-1", Type: "image", AssetRef: map[string]any{"objectKey": "private/alice/image.png"}}}},
	} {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			store := NewMySQLStore(db)
			project := Project{ID: "unsafe-project", Title: "不安全项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}
			if err := store.CreateProject(context.Background(), "alice", project); !errors.Is(err, ErrCanvasInvalid) {
				t.Fatalf("err=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unexpected database write: %v", err)
			}
		})
	}
}

func TestMySQLStoreRejectsStaleCanvasRevisionBeforeWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(4)))
	mock.ExpectRollback()

	_, err = store.SaveCanvas(context.Background(), "alice", "project-1", EmptyCanvas(), 3, time.Now())
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreMapsDuplicateProjectForImportRecovery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 18, 1, 2, 3, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_projects (id, owner_username, title, title_mode, project_type, preview, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "duplicate"})
	mock.ExpectRollback()

	err = store.ImportProject(context.Background(), "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now})
	if !errors.Is(err, ErrProjectExists) {
		t.Fatalf("err=%v", err)
	}
}

func TestNullableJSONUsesTextForMySQLJSONColumns(t *testing.T) {
	value, err := nullableJSON(map[string]any{"zoom": 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := value.(string); !ok {
		t.Fatalf("JSON driver value type=%T, want string", value)
	}
}

func TestMySQLStoreCanvasReadsNodeGroupAndConfiguration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT canvas_version, revision, canvas_json, viewport_json, appearance_json, view_state_json FROM agent_canvases WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"canvas_version", "revision", "canvas_json", "viewport_json", "appearance_json", "view_state_json"}).AddRow(1, int64(2), nil, nil, `{"edgeColor":"cyan","edgeAnimated":false,"edgeWidth":"thick","backgroundDots":false,"alignmentGuides":true,"edgesVisible":false}`, `{"activeView":"storyboard","drawerWidth":340}`))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id
FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`)).
		WithArgs("project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "node_type", "title", "summary", "position_x", "position_y", "current_version", "status", "source_node_ids_json", "asset_ref_json", "config_json", "group_id"}).
			AddRow("node-1", "video", "节点", "内容", 12.0, 24.0, int64(1), "ready", nil, nil, `{"prompt":"雨夜追逐","ratio":"9:16"}`, "group-1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, source_node_id, target_node_id, relation_type FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`)).
		WithArgs("project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_node_id", "target_node_id", "relation_type"}))

	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(canvas.Nodes) != 1 || canvas.Nodes[0].GroupID != "group-1" || canvas.Nodes[0].Config["prompt"] != "雨夜追逐" || canvas.Nodes[0].Config["ratio"] != "9:16" || canvas.Appearance.EdgeColor != "cyan" || canvas.Appearance.EdgeAnimated || canvas.Appearance.EdgeWidth != "thick" || canvas.Appearance.BackgroundDots || !canvas.Appearance.AlignmentGuides || canvas.ViewState["activeView"] != "storyboard" || canvas.ViewState["drawerWidth"] != float64(340) {
		t.Fatalf("canvas=%+v", canvas)
	}
	appearanceJSON, err := json.Marshal(canvas.Appearance)
	if err != nil || string(appearanceJSON) != `{"edgeColor":"cyan","edgeAnimated":false,"edgeWidth":"thick","backgroundDots":false,"alignmentGuides":true,"edgesVisible":false}` {
		t.Fatalf("appearance JSON=%s err=%v", appearanceJSON, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCanvasDefaultsLegacyEdgeVisibilityToTrue(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT canvas_version, revision, canvas_json, viewport_json, appearance_json, view_state_json FROM agent_canvases WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-legacy", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"canvas_version", "revision", "canvas_json", "viewport_json", "appearance_json", "view_state_json"}).AddRow(1, int64(2), nil, nil, `{"edgeColor":"cyan","edgeAnimated":false,"edgeWidth":"thick","backgroundDots":false,"alignmentGuides":true}`, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id
FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`)).
		WithArgs("project-legacy", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "node_type", "title", "summary", "position_x", "position_y", "current_version", "status", "source_node_ids_json", "asset_ref_json", "config_json", "group_id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, source_node_id, target_node_id, relation_type FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`)).
		WithArgs("project-legacy", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_node_id", "target_node_id", "relation_type"}))

	canvas, err := store.Canvas(context.Background(), "alice", "project-legacy")
	if err != nil {
		t.Fatal(err)
	}
	if !canvas.Appearance.EdgesVisible {
		t.Fatalf("legacy appearance hid edges: %+v", canvas.Appearance)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreSaveCanvasWritesSnapshotAndReconcilesIndexes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 19, 1, 2, 3, 0, time.UTC)
	appearanceJSON := `{"edgeColor":"neutral","edgeAnimated":false,"edgeWidth":"thin","backgroundDots":true,"alignmentGuides":true,"edgesVisible":false}`

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(0)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_canvases SET canvas_version = ?, revision = ?, canvas_json = ?, viewport_json = ?, appearance_json = ?, view_state_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`)).
		WithArgs(1, int64(1), sqlmock.AnyArg(), nil, appearanceJSON, `{"activeView":"storyboard","drawerWidth":340}`, now, "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id NOT IN (?)`)).
		WithArgs("project-1", "alice", "node-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvas_nodes (id, project_id, owner_username, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE node_type = VALUES(node_type), title = VALUES(title), summary = VALUES(summary), position_x = VALUES(position_x), position_y = VALUES(position_y), current_version = VALUES(current_version), status = VALUES(status), source_node_ids_json = VALUES(source_node_ids_json), asset_ref_json = VALUES(asset_ref_json), config_json = VALUES(config_json), group_id = VALUES(group_id), updated_at = VALUES(updated_at)`)).
		WithArgs("node-1", "project-1", "alice", "video", "节点", "内容", 12.0, 24.0, int64(1), "ready", nil, nil, `{"prompt":"雨夜追逐","ratio":"9:16"}`, sql.NullString{String: "group-1", Valid: true}, now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvas_edges (id, project_id, owner_username, source_node_id, target_node_id, relation_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("edge-1", "project-1", "alice", "node-1", "node-1", "reference", now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`)).
		WithArgs(now, "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	canvas := EmptyCanvas()
	canvas.Appearance.EdgesVisible = false
	canvas.ViewState = map[string]any{"activeView": "storyboard", "drawerWidth": 340}
	canvas.Nodes = []Node{{ID: "node-1", Type: "video", Title: "节点", Summary: "内容", Position: Position{X: 12, Y: 24}, Version: 1, Status: "ready", Config: map[string]any{"prompt": "雨夜追逐", "ratio": "9:16"}, GroupID: "group-1"}}
	canvas.Edges = []Edge{{ID: "edge-1", Source: "node-1", Target: "node-1", Relation: "reference"}}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, 0, now); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreSaveBatchFactoryProjectionRollsBackCanvasWhenOriginInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 23, 11, 30, 0, 0, time.UTC)
	appearanceJSON := `{"edgeColor":"neutral","edgeAnimated":false,"edgeWidth":"thin","backgroundDots":true,"alignmentGuides":true,"edgesVisible":true}`
	binding := BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: "project-1"}
	origin := BatchFactoryNodeOrigin{ProjectID: binding.ProjectID, NodeID: "node-1", BatchID: binding.BatchID, BookID: "book-1", ResourceType: "book_script", ResourceID: "book-1", VersionID: "v1"}
	wantErr := errors.New("origin insert failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(0)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_canvases SET canvas_version = ?, revision = ?, canvas_json = ?, viewport_json = ?, appearance_json = ?, view_state_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`)).
		WithArgs(1, int64(1), sqlmock.AnyArg(), nil, appearanceJSON, nil, now, "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM agent_batch_factory_bindings WHERE owner_username = ? AND batch_id = ? AND project_id = ?`)).
		WithArgs("alice", "batch-1", "project-1").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_batch_factory_node_origins WHERE owner_username = ? AND project_id = ? AND book_id = ?`)).
		WithArgs("alice", "project-1", "book-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_batch_factory_node_origins (project_id, node_id, owner_username, batch_id, book_id, resource_type, resource_id, version_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("project-1", "node-1", "alice", "batch-1", "book-1", "book_script", "book-1", "v1").WillReturnError(wantErr)
	mock.ExpectRollback()

	_, err = store.SaveBatchFactoryProjection(context.Background(), "alice", binding, "book-1", EmptyCanvas(), 0, []BatchFactoryNodeOrigin{origin}, now)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err=%v, want %v", err, wantErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreBeginsAndAppliesTextExecutionAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 20, 8, 20, 0, 0, time.UTC)
	execution := Execution{ID: "execution-1", ToolName: "text-generation", Status: "running", IdempotencyKey: "text-1-v1", CreatedAt: now, UpdatedAt: now}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("project-1"))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_executions (id, project_id, owner_username, tool_name, status, detail_json, idempotency_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("execution-1", "project-1", "alice", "text-generation", "running", nil, "text-1-v1", now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if _, err := store.BeginNodeExecution(context.Background(), "alice", "project-1", execution); err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM agent_executions WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("execution-1", "project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("running"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT node_type, current_version FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`)).
		WithArgs("project-1", "alice", "text-1").WillReturnRows(sqlmock.NewRows([]string{"node_type", "current_version"}).AddRow("text", int64(1)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_canvas_nodes SET summary = ?, current_version = current_version + 1, status = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND id = ?`)).
		WithArgs("生成后的正文", "confirmed", now.Add(time.Second), "project-1", "alice", "text-1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision, canvas_json FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision", "canvas_json"}).AddRow(int64(3), nil))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_canvases SET revision = revision + 1, updated_at = ? WHERE project_id = ? AND owner_username = ?`)).
		WithArgs(now.Add(time.Second), "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`)).
		WithArgs(now.Add(time.Second), "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_executions SET status = ?, detail_json = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`)).
		WithArgs("completed", `{"modelId":"text-primary"}`, now.Add(time.Second), "execution-1", "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT canvas_version, revision, canvas_json, viewport_json, appearance_json, view_state_json FROM agent_canvases WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"canvas_version", "revision", "canvas_json", "viewport_json", "appearance_json", "view_state_json"}).AddRow(1, int64(4), nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id
FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "node_type", "title", "summary", "position_x", "position_y", "current_version", "status", "source_node_ids_json", "asset_ref_json", "config_json", "group_id"}).
		AddRow("text-1", "text", "文字", "生成后的正文", 0.0, 0.0, int64(2), "confirmed", nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, source_node_id, target_node_id, relation_type FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "source_node_id", "target_node_id", "relation_type"}))

	canvas, completed, err := store.ApplyNodeExecution(context.Background(), "alice", "project-1", ApplyNodeExecutionInput{
		ExecutionID: "execution-1", NodeID: "text-1", NodeVersion: 1, NodeType: "text", Summary: "生成后的正文",
		Detail: map[string]any{"modelId": "text-primary"}, Now: now.Add(time.Second),
	})
	if err != nil || canvas.Revision != 4 || completed.Status != "completed" {
		t.Fatalf("canvas=%+v execution=%+v err=%v", canvas, completed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreMarksNodeExecutionAwaitingConfirmation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 21, 8, 30, 0, 0, time.UTC)
	detail := map[string]any{"nodeId": "image-1", "nodeVersion": int64(1), "resultAssetIds": []string{"asset-1", "asset-2"}}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM agent_executions WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("execution-1", "project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("running"))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_executions SET status = ?, detail_json = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`)).
		WithArgs("awaiting_confirmation", `{"nodeId":"image-1","nodeVersion":1,"resultAssetIds":["asset-1","asset-2"]}`, now, "execution-1", "project-1", "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`)).
		WithArgs(now, "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	pending, err := store.SetNodeExecutionAwaitingConfirmation(context.Background(), "alice", "project-1", "execution-1", detail, now)
	if err != nil || pending.Status != "awaiting_confirmation" || pending.Detail["nodeId"] != "image-1" {
		t.Fatalf("pending=%+v err=%v", pending, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreConfirmVideoJobReturnsNodeQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 20, 1, 20, 0, 0, time.UTC)
	nodeErr := errors.New("node query failed")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("job-1", "project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "node_id", "node_version", "state", "model_key", "provider_task_id", "settings_json", "reference_asset_ids_json", "result_asset_id", "error_message", "idempotency_key", "created_at", "updated_at"}).
			AddRow("job-1", "project-1", "alice", "video-1", int64(1), VideoJobAwaitingConfirmation, "minimax-h3-video", "provider-1", `{"prompt":"雨夜追逐"}`, `[]`, "result-1", nil, "request-1", now, now))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at FROM agent_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("result-1", "project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "kind", "name", "content_type", "object_key", "byte_size", "created_at"}).
			AddRow("result-1", "project-1", "video", "result.mp4", "video/mp4", "agent-assets/alice/project-1/result.mp4", int64(128), now))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE`)).
		WithArgs("project-1", "alice", "video-1").
		WillReturnError(nodeErr)
	mock.ExpectRollback()

	_, err = store.ConfirmVideoJob(context.Background(), "alice", "project-1", "job-1", now)
	if err != nodeErr {
		t.Fatalf("err=%v, want exact node query error", err)
	}
	if errors.Is(err, ErrVideoJobNodeChanged) {
		t.Fatalf("node query error was mapped to ErrVideoJobNodeChanged")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListsPollableVideoJobsAcrossProjectsInOldestOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	older := time.Date(2026, 10, 23, 8, 0, 0, 0, time.UTC)
	newer := older.Add(time.Minute)
	columns := []string{"id", "project_id", "owner_username", "node_id", "node_version", "state", "model_key", "provider_task_id", "settings_json", "reference_asset_ids_json", "result_asset_id", "error_message", "idempotency_key", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT ` + videoJobColumns + ` FROM agent_video_jobs WHERE state IN ('submitting', 'queued', 'running', 'result_ready', 'ingesting') ORDER BY updated_at ASC, id ASC LIMIT ?`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("job-old", "project-1", "alice", "video-1", int64(1), VideoJobSubmitting, "minimax-h3-video", nil, `{"prompt":"雨夜追逐"}`, `[]`, nil, nil, "request-old", older, older).
			AddRow("job-new", "project-2", "bob", "video-2", int64(1), VideoJobRunning, "minimax-h3-video", "provider-new", `{"prompt":"海边日出"}`, `[]`, nil, nil, "request-new", newer, newer))

	jobs, err := store.ListPollableVideoJobs(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].ID != "job-old" || jobs[0].OwnerUsername != "alice" || jobs[1].ID != "job-new" || jobs[1].ProjectID != "project-2" {
		t.Fatalf("jobs=%+v", jobs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLRetryVideoJobLocksAndRejectsNonRetryableStateBeforeInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 20, 4, 20, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("job-1", "project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "node_id", "node_version", "state", "model_key", "provider_task_id", "settings_json", "reference_asset_ids_json", "result_asset_id", "error_message", "idempotency_key", "created_at", "updated_at"}).
			AddRow("job-1", "project-1", "alice", "video-1", int64(1), VideoJobQueued, "minimax-h3-video", "provider-1", `{"prompt":"雨夜追逐"}`, `[]`, nil, nil, "request-1", now, now))
	mock.ExpectRollback()

	_, err = store.RetryVideoJob(context.Background(), "alice", "project-1", "job-1", RetryVideoJobInput{ID: "job-2", IdempotencyKey: "request-2", Now: now.Add(time.Second)})
	if !errors.Is(err, ErrVideoJobState) {
		t.Fatalf("err=%v, want ErrVideoJobState", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLRetryVideoJobRejectsDeletedProjectBeforeInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 20, 6, 20, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `+videoJobColumns+` FROM agent_video_jobs WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("job-1", "project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "node_id", "node_version", "state", "model_key", "provider_task_id", "settings_json", "reference_asset_ids_json", "result_asset_id", "error_message", "idempotency_key", "created_at", "updated_at"}).
			AddRow("job-1", "project-1", "alice", "video-1", int64(1), VideoJobFailed, "minimax-h3-video", "provider-1", `{"prompt":"雨夜追逐"}`, `[]`, nil, "failed", "request-1", now, now))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`)).
		WithArgs("project-1", "alice").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = store.RetryVideoJob(context.Background(), "alice", "project-1", "job-1", RetryVideoJobInput{ID: "job-2", IdempotencyKey: "request-2", Now: now.Add(time.Second)})
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("err=%v, want ErrProjectNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLAssetStoreMapsDuplicateKeyToAssetExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLAssetStore(db)
	asset := Asset{ID: "video-result-job-1", ProjectID: "project-1", Kind: "video", Name: "result.mp4", ContentType: "video/mp4", ObjectKey: "agent-assets/alice/project-1/video-result-job-1.mp4", ByteSize: 3, CreatedAt: time.Now()}
	mock.ExpectExec("INSERT INTO agent_assets").WillReturnError(&mysql.MySQLError{Number: 1062, Message: "duplicate"})

	if err := store.CreateAsset(context.Background(), "alice", asset); !errors.Is(err, ErrAssetExists) {
		t.Fatalf("err=%v, want ErrAssetExists", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
