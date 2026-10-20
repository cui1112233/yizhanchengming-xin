package agentworkspace

import (
	"context"
	"database/sql"
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
	appearanceJSON := `{"edgeColor":"coral","edgeAnimated":true,"edgeWidth":"default","backgroundDots":true,"alignmentGuides":true}`

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_projects (id, owner_username, title, title_mode, preview, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("project-1", "alice", "项目", "automatic", "", now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvases (project_id, owner_username, canvas_version, revision, viewport_json, appearance_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("project-1", "alice", 1, int64(0), nil, appearanceJSON, now, now).
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
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_projects (id, owner_username, title, title_mode, preview, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`)).
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT canvas_version, revision, viewport_json, appearance_json FROM agent_canvases WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"canvas_version", "revision", "viewport_json", "appearance_json"}).AddRow(1, int64(2), nil, `{"edgeColor":"cyan","edgeAnimated":false,"edgeWidth":"thick","backgroundDots":false,"alignmentGuides":true}`))
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
	if len(canvas.Nodes) != 1 || canvas.Nodes[0].GroupID != "group-1" || canvas.Nodes[0].Config["prompt"] != "雨夜追逐" || canvas.Nodes[0].Config["ratio"] != "9:16" || canvas.Appearance.EdgeColor != "cyan" || canvas.Appearance.EdgeAnimated || canvas.Appearance.EdgeWidth != "thick" || canvas.Appearance.BackgroundDots || !canvas.Appearance.AlignmentGuides {
		t.Fatalf("canvas=%+v", canvas)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreSaveCanvasWritesNodeGroupAndConfiguration(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 19, 1, 2, 3, 0, time.UTC)
	appearanceJSON := `{"edgeColor":"coral","edgeAnimated":true,"edgeWidth":"default","backgroundDots":true,"alignmentGuides":true}`

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(int64(0)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_canvases SET canvas_version = ?, revision = ?, viewport_json = ?, appearance_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`)).
		WithArgs(1, int64(1), nil, appearanceJSON, now, "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_canvas_edges WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ?`)).
		WithArgs("project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvas_nodes (id, project_id, owner_username, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("node-1", "project-1", "alice", "video", "节点", "内容", 12.0, 24.0, int64(1), "ready", nil, nil, `{"prompt":"雨夜追逐","ratio":"9:16"}`, sql.NullString{String: "group-1", Valid: true}, now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT IGNORE INTO agent_node_versions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_projects SET updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`)).
		WithArgs(now, "project-1", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	canvas := Canvas{Version: 1, Nodes: []Node{{ID: "node-1", Type: "video", Title: "节点", Summary: "内容", Position: Position{X: 12, Y: 24}, Version: 1, Status: "ready", Config: map[string]any{"prompt": "雨夜追逐", "ratio": "9:16"}, GroupID: "group-1"}}, Edges: []Edge{}}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, 0, now); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
