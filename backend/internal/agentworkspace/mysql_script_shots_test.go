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
)

func TestMySQLScriptShotStoreReplacesListsAndScopesShots(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE")).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM agent_script_shots WHERE project_id = ? AND script_node_id = ? AND owner_username = ?")).
		WithArgs("project-1", "script-1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_script_shots (id, project_id, script_node_id, owner_username, sequence, scene, characters, dialogue, narration, image_prompt, video_prompt, tts_text, revision, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")).
		WithArgs("shot-2", "project-1", "script-1", "alice", 2, "", "", "", "", "", "", "", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_script_shots (id, project_id, script_node_id, owner_username, sequence, scene, characters, dialogue, narration, image_prompt, video_prompt, tts_text, revision, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")).
		WithArgs("shot-1", "project-1", "script-1", "alice", 1, "", "", "", "", "", "", "", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.ReplaceShots(context.Background(), "alice", "project-1", "script-1", []ScriptShot{{ID: "shot-2", Sequence: 2}, {ID: "shot-1", Sequence: 1}}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, script_node_id, sequence, scene, characters, dialogue, narration, image_prompt, video_prompt, tts_text, revision, created_at, updated_at FROM agent_script_shots WHERE owner_username = ? AND project_id = ? AND script_node_id = ? ORDER BY sequence, id")).
		WithArgs("bob", "project-1", "script-1").WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "script_node_id", "sequence", "scene", "characters", "dialogue", "narration", "image_prompt", "video_prompt", "tts_text", "revision", "created_at", "updated_at"}))
	shots, err := store.ListShots(context.Background(), "bob", "project-1", "script-1")
	if err != nil || len(shots) != 0 {
		t.Fatalf("cross-owner shots=%+v err=%v, want empty", shots, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLScriptShotStoreUpdatesByRevisionAndRecordsLineage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 22, 4, 0, 0, 0, time.UTC)
	persistedCreatedAt := now.Add(-time.Hour)
	shot := ScriptShot{ID: "shot-1", ProjectID: "project-1", ScriptNodeID: "script-1", Sequence: 1, Scene: "雨夜小巷", Revision: 3, CreatedAt: now}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("shot-1", "project-1", "script-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(persistedCreatedAt))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE agent_script_shots SET sequence = ?, scene = ?, characters = ?, dialogue = ?, narration = ?, image_prompt = ?, video_prompt = ?, tts_text = ?, revision = revision + 1, updated_at = ? WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? AND revision = ?")).
		WithArgs(1, "雨夜小巷", "", "", "", "", "", "", sqlmock.AnyArg(), "shot-1", "project-1", "script-1", "alice", int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	updated, err := store.UpdateShot(context.Background(), "alice", shot)
	if err != nil || updated.Revision != 4 || !updated.CreatedAt.Equal(persistedCreatedAt) {
		t.Fatalf("updated=%+v err=%v, want revision 4 with persisted created_at", updated, err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE")).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("shot-1", "project-1", "script-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("shot-1"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_shot_generation_links (project_id, script_node_id, shot_id, owner_username, node_id, generation_id, attempt, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE script_node_id = VALUES(script_node_id), attempt = VALUES(attempt), active = VALUES(active), updated_at = VALUES(updated_at)")).
		WithArgs("project-1", "script-1", "shot-1", "alice", "image-1", "gen-1", 1, false, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := store.AttachGeneration(context.Background(), "alice", ShotGenerationLink{ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "shot-1", NodeID: "image-1", GenerationID: "gen-1", Attempt: 1}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, script_node_id, shot_id, node_id, generation_id, attempt, active FROM agent_shot_generation_links WHERE owner_username = ? AND project_id = ? AND shot_id = ? ORDER BY attempt, node_id, generation_id")).
		WithArgs("alice", "project-1", "shot-1").WillReturnRows(sqlmock.NewRows([]string{"project_id", "script_node_id", "shot_id", "node_id", "generation_id", "attempt", "active"}).AddRow("project-1", "script-1", "shot-1", "image-1", "gen-1", 1, false))
	links, err := store.ListGenerationLinks(context.Background(), "alice", "project-1", "shot-1")
	if err != nil || len(links) != 1 || links[0].GenerationID != "gen-1" {
		t.Fatalf("links=%+v err=%v", links, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLScriptShotStoreRejectsForeignProjectAndMissingShotLineage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE")).
		WithArgs("project-1", "bob").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	if err := store.ReplaceShots(context.Background(), "bob", "project-1", "script-1", []ScriptShot{{ID: "shot-1", Sequence: 1}}); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("foreign replace err=%v, want ErrProjectNotFound", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE")).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("missing", "project-1", "script-1", "alice").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	if err := store.AttachGeneration(context.Background(), "alice", ShotGenerationLink{ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "missing", NodeID: "image-1", GenerationID: "gen-1", Attempt: 1}); !errors.Is(err, ErrScriptShotNotFound) {
		t.Fatalf("missing lineage err=%v, want ErrScriptShotNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLScriptShotStoreReturnsRevisionConflictWhenUpdateAffectsNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	shot := ScriptShot{ID: "shot-1", ProjectID: "project-1", ScriptNodeID: "script-1", Sequence: 1, Revision: 3}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT created_at FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("shot-1", "project-1", "script-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Date(2026, 10, 22, 3, 0, 0, 0, time.UTC)))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE agent_script_shots SET sequence = ?, scene = ?, characters = ?, dialogue = ?, narration = ?, image_prompt = ?, video_prompt = ?, tts_text = ?, revision = revision + 1, updated_at = ? WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? AND revision = ?")).
		WithArgs(1, "", "", "", "", "", "", "", sqlmock.AnyArg(), "shot-1", "project-1", "script-1", "alice", int64(3)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	_, err = store.UpdateShot(context.Background(), "alice", shot)
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("update err=%v, want ErrRevisionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLShotGenerationRollsBackNodeTaskAndLinkTogether(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 22, 14, 30, 0, 0, time.UTC)
	canvas := EmptyCanvas()
	canvas.Nodes = []Node{{ID: "script-1", Type: "script", Version: 1}}
	snapshot, _ := json.Marshal(canvas)
	input := CreateShotGenerationInput{
		ExpectedCanvasRevision: 0,
		Node:                   Node{ID: "image-1", Type: "image", Title: "镜头 1 图片", Version: 1, Status: "queued", SourceNodeIDs: []string{"script-1"}},
		Edge:                   Edge{ID: "edge-1", Source: "script-1", Target: "image-1", Relation: "shot-generation"},
		Task:                   CreateGenerationTaskInput{ID: "generation-1", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "personal_api", ModelID: "image-model", Mode: "text_to_image", Settings: map[string]any{"prompt": "雨夜"}, IdempotencyKey: "shot:shot-1:text_to_image:1", Now: now},
		Link:                   ShotGenerationLink{ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "shot-1", NodeID: "image-1", GenerationID: "generation-1", Attempt: 1, Active: true},
		Now:                    now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT revision, canvas_json FROM agent_canvases WHERE project_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"revision", "canvas_json"}).AddRow(int64(0), snapshot))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM agent_script_shots WHERE id = ? AND project_id = ? AND script_node_id = ? AND owner_username = ? FOR UPDATE`)).
		WithArgs("shot-1", "project-1", "script-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvas_nodes (id, project_id, owner_username, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_node_versions (project_id, node_id, owner_username, version, content_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_canvas_edges (id, project_id, owner_username, source_node_id, target_node_id, relation_type, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE agent_canvases SET revision = ?, canvas_json = ?, updated_at = ? WHERE project_id = ? AND owner_username = ? AND revision = ?`)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_generation_tasks (id, project_id, owner_username, node_id, node_version, kind, provider, model_id, mode, state, settings_json, reference_asset_ids_json, selected_result_id, idempotency_key, retry_of_task_id, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_shot_generation_links (project_id, script_node_id, shot_id, owner_username, node_id, generation_id, attempt, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).WillReturnError(errors.New("lineage insert failed"))
	mock.ExpectRollback()

	if _, _, _, err := store.CreateShotGeneration(context.Background(), "alice", input); err == nil || err.Error() != "lineage insert failed" {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
