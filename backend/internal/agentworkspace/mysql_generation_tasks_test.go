package agentworkspace

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func generationTaskRow(now time.Time, state GenerationState) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "project_id", "owner_username", "node_id", "node_version", "kind", "provider", "model_id", "mode", "state",
		"settings_json", "reference_asset_ids_json", "selected_result_id", "idempotency_key", "retry_of_task_id", "error_message", "created_at", "updated_at",
	}).AddRow("generation-1", "project-1", "alice", "image-1", int64(1), "image", "personal_api", "image-model", "text_to_image", state,
		"{\"count\":2,\"ratio\":\"16:9\"}", "[\"asset-ref-1\"]", nil, "image-1-v1", nil, nil, now, now)
}

func TestMySQLGenerationTaskCreateIsProjectScopedAndIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 21, 13, 0, 0, 0, time.UTC)
	input := generationTestInput(now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+generationTaskColumns+" FROM agent_generation_tasks WHERE owner_username = ? AND project_id = ? AND idempotency_key = ?")).
		WithArgs("alice", "project-1", "image-1-v1").WillReturnError(sql.ErrNoRows)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE")).
		WithArgs("project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_generation_tasks (id, project_id, owner_username, node_id, node_version, kind, provider, model_id, mode, state, settings_json, reference_asset_ids_json, selected_result_id, idempotency_key, retry_of_task_id, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")).
		WithArgs("generation-1", "project-1", "alice", "image-1", int64(1), "image", "personal_api", "image-model", "text_to_image", GenerationQueued, "{\"count\":2,\"ratio\":\"16:9\"}", "[\"asset-ref-1\"]", nil, "image-1-v1", nil, nil, now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	created, err := store.CreateGenerationTask(context.Background(), "alice", input)
	if err != nil || created.State != GenerationQueued || created.ProjectID != "project-1" {
		t.Fatalf("created=%+v err=%v", created, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+generationTaskColumns+" FROM agent_generation_tasks WHERE owner_username = ? AND project_id = ? AND idempotency_key = ?")).
		WithArgs("alice", "project-1", "image-1-v1").WillReturnRows(generationTaskRow(now, GenerationQueued))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_id, kind, asset_json, created_at FROM agent_generation_results WHERE owner_username = ? AND project_id = ? AND task_id = ? ORDER BY created_at, id")).
		WithArgs("alice", "project-1", "generation-1").WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "kind", "asset_json", "created_at"}))

	duplicate, err := store.CreateGenerationTask(context.Background(), "alice", input)
	if err != nil || duplicate.ID != created.ID {
		t.Fatalf("duplicate=%+v err=%v", duplicate, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLGenerationTaskListsRunnableScriptSplitsAcrossProjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 23, 14, 0, 0, 0, time.UTC)
	columns := []string{"id", "project_id", "owner_username", "node_id", "node_version", "kind", "provider", "model_id", "mode", "state", "settings_json", "reference_asset_ids_json", "selected_result_id", "idempotency_key", "retry_of_task_id", "error_message", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+generationTaskColumns+" FROM agent_generation_tasks WHERE mode IN (?) AND state IN ('queued', 'running') ORDER BY updated_at ASC, id ASC LIMIT ?")).
		WithArgs("script_split", 2).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("split-old", "project-1", "alice", "script-1", int64(1), "text", "configured_model", "text-model", "script_split", GenerationQueued, `{"sourceText":"第一段"}`, `[]`, nil, "split-old", nil, nil, now, now).
			AddRow("split-running", "project-2", "bob", "script-2", int64(2), "text", "configured_model", "text-model", "script_split", GenerationRunning, `{"sourceText":"第二段"}`, `[]`, nil, "split-running", nil, nil, now, now.Add(time.Second)))

	tasks, err := store.ListRunnableGenerationTasks(context.Background(), "script_split", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].ID != "split-old" || tasks[0].OwnerUsername != "alice" || tasks[1].ID != "split-running" || tasks[1].State != GenerationRunning {
		t.Fatalf("tasks=%+v", tasks)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLGenerationTaskListsRunnableTasksForMultipleModes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 23, 14, 0, 0, 0, time.UTC)
	columns := []string{"id", "project_id", "owner_username", "node_id", "node_version", "kind", "provider", "model_id", "mode", "state", "settings_json", "reference_asset_ids_json", "selected_result_id", "idempotency_key", "retry_of_task_id", "error_message", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+generationTaskColumns+" FROM agent_generation_tasks WHERE mode IN (?,?) AND state IN ('queued', 'running') ORDER BY updated_at ASC, id ASC LIMIT ?")).
		WithArgs("text_to_image", "text_to_speech", 3).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("image-old", "project-1", "alice", "image-1", int64(1), "image", "image", "model", "text_to_image", GenerationQueued, `{}`, `[]`, nil, "image-old", nil, nil, now, now).
			AddRow("audio-running", "project-2", "bob", "audio-1", int64(1), "audio", "tts", "voice", "text_to_speech", GenerationRunning, `{}`, `[]`, nil, "audio-running", nil, nil, now, now.Add(time.Second)))

	tasks, err := store.ListRunnableGenerationTasksByModes(context.Background(), []string{"text_to_image", "text_to_speech"}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].ID != "image-old" || tasks[1].OwnerUsername != "bob" {
		t.Fatalf("tasks=%+v", tasks)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLGenerationConfirmRejectsChangedNodeBeforeWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 21, 14, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+generationTaskColumns+" FROM agent_generation_tasks WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("generation-1", "project-1", "alice").WillReturnRows(generationTaskRow(now, GenerationAwaitingConfirmation))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_id, kind, asset_json, created_at FROM agent_generation_results WHERE owner_username = ? AND project_id = ? AND task_id = ? ORDER BY created_at, id")).
		WithArgs("alice", "project-1", "generation-1").WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "kind", "asset_json", "created_at"}).
		AddRow("result-1", "asset-1", "image", "{\"id\":\"asset-1\",\"kind\":\"image\",\"name\":\"结果.png\"}", now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at FROM agent_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("asset-1", "project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "kind", "name", "content_type", "object_key", "byte_size", "created_at"}).
		AddRow("asset-1", "project-1", "image", "结果.png", "image/png", "agent/project-1/result.png", int64(1024), now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE")).
		WithArgs("project-1", "alice", "image-1").WillReturnRows(sqlmock.NewRows([]string{"id", "node_type", "title", "summary", "position_x", "position_y", "current_version", "status", "source_node_ids_json", "asset_ref_json", "config_json", "group_id"}).
		AddRow("image-1", "image", "图片", "", 0.0, 0.0, int64(2), "draft", nil, nil, nil, nil))
	mock.ExpectRollback()

	if _, _, err := store.ConfirmGenerationTask(context.Background(), "alice", "project-1", "generation-1", "result-1", now.Add(time.Second)); err != ErrGenerationNodeChanged {
		t.Fatalf("err=%v, want ErrGenerationNodeChanged", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLGenerationActivateRejectsStaleNodeVersionBeforeWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT "+generationTaskColumns+" FROM agent_generation_tasks WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("generation-1", "project-1", "alice").WillReturnRows(generationTaskRow(now, GenerationCompleted))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_id, kind, asset_json, created_at FROM agent_generation_results WHERE owner_username = ? AND project_id = ? AND task_id = ? ORDER BY created_at, id")).
		WithArgs("alice", "project-1", "generation-1").WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "kind", "asset_json", "created_at"}).
		AddRow("result-1", "asset-1", "image", "{\"id\":\"asset-1\",\"kind\":\"image\",\"name\":\"结果.png\"}", now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at FROM agent_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE")).
		WithArgs("asset-1", "project-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "kind", "name", "content_type", "object_key", "byte_size", "created_at"}).
		AddRow("asset-1", "project-1", "image", "结果.png", "image/png", "agent/project-1/result.png", int64(1024), now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, node_type, title, summary, position_x, position_y, current_version, status, source_node_ids_json, asset_ref_json, config_json, group_id FROM agent_canvas_nodes WHERE project_id = ? AND owner_username = ? AND id = ? FOR UPDATE")).
		WithArgs("project-1", "alice", "image-1").WillReturnRows(sqlmock.NewRows([]string{"id", "node_type", "title", "summary", "position_x", "position_y", "current_version", "status", "source_node_ids_json", "asset_ref_json", "config_json", "group_id"}).
		AddRow("image-1", "image", "图片", "", 0.0, 0.0, int64(3), "confirmed", nil, nil, nil, nil))
	mock.ExpectRollback()

	if _, _, err := store.ActivateGenerationResult(context.Background(), "alice", "project-1", "generation-1", "result-1", 2, now.Add(time.Second)); err != ErrGenerationNodeChanged {
		t.Fatalf("err=%v, want ErrGenerationNodeChanged", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
