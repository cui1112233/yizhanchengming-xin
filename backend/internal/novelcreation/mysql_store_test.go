package novelcreation

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func newNovelCreationMockStore(t *testing.T) (*MySQLStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewMySQLStore(db).(*MySQLStore), mock
}

func TestMySQLStoreProjectQueriesAreOwnerScopedAndIgnoreSoftDeleted(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	query := `SELECT project_id, title, genre, world, protagonist, core_conflict, model_id, target_chapter_words, status, created_at, updated_at, deleted_at
FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("alice", "project-1").WillReturnRows(sqlmock.NewRows([]string{
		"project_id", "title", "genre", "world", "protagonist", "core_conflict", "model_id", "target_chapter_words", "status", "created_at", "updated_at", "deleted_at",
	}))

	_, err := store.GetProject(context.Background(), "alice", "project-1")
	if !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("err=%v, want owner-scoped soft-deleted project to be invisible", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreScansNullableProjectTextFields(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	query := `SELECT project_id, title, genre, world, protagonist, core_conflict, model_id, target_chapter_words, status, created_at, updated_at, deleted_at
FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL`
	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("alice", "project-1").WillReturnRows(sqlmock.NewRows([]string{
		"project_id", "title", "genre", "world", "protagonist", "core_conflict", "model_id", "target_chapter_words", "status", "created_at", "updated_at", "deleted_at",
	}).AddRow("project-1", "空白字段", "悬疑", nil, nil, nil, "text-1", 2000, "draft", now, now, nil))

	project, err := store.GetProject(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if project.World != "" || project.Protagonist != "" || project.CoreConflict != "" {
		t.Fatalf("nullable fields=%+v, want empty strings", project)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCreateProjectWritesEmptyOptionalTextAsNull(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	mock.ExpectExec("INSERT INTO novel_projects").
		WithArgs(sqlmock.AnyArg(), "alice", "空白字段", "悬疑", nil, nil, nil, "text-1", sqlmock.AnyArg(), 2000, "draft", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if _, err := store.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "空白字段", Genre: "悬疑", ModelID: "text-1"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreIdempotentRunSurvivesReopen(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	ctx := context.Background()
	run := Run{ID: "run-1", ProjectID: "project-1", SessionID: "session-1", IdempotencyKey: "send-1", Status: RunQueued}
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO novel_runs
(run_id, owner_username, project_id, session_id, idempotency_key, status, model_id, thinking_level, prompt_snapshot_json, runtime_snapshot_json, error_code, error_message, created_at, updated_at, finished_at)
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)
  AND EXISTS (SELECT 1 FROM novel_sessions WHERE owner_username = ? AND session_id = ? AND project_id = ?)`)).
		WithArgs(run.ID, "alice", run.ProjectID, run.SessionID, run.IdempotencyKey, run.Status, "", "", nil, nil, "", nil, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "alice", run.ProjectID, "alice", run.SessionID, run.ProjectID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_sessions SET current_run_id = ?, updated_at = ? WHERE owner_username = ? AND session_id = ? AND project_id = ?`)).
		WithArgs(run.ID, sqlmock.AnyArg(), "alice", run.SessionID, run.ProjectID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	first, err := store.CreateRun(ctx, "alice", run)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != run.ID {
		t.Fatalf("run ID=%q, want %q", first.ID, run.ID)
	}

	query := `SELECT r.run_id, r.project_id, r.session_id, r.idempotency_key, r.status, r.model_id, r.thinking_level, r.prompt_snapshot_json, r.runtime_snapshot_json, r.error_code, COALESCE(r.error_message, ''), r.created_at, r.updated_at, r.finished_at
FROM novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id
WHERE r.owner_username = ? AND r.idempotency_key = ? AND p.deleted_at IS NULL`
	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("alice", "send-1").WillReturnRows(sqlmock.NewRows([]string{
		"run_id", "project_id", "session_id", "idempotency_key", "status", "model_id", "thinking_level", "prompt_snapshot_json", "runtime_snapshot_json", "error_code", "error_message", "created_at", "updated_at", "finished_at",
	}).AddRow("run-1", "project-1", "session-1", "send-1", RunQueued, "", "", nil, nil, "", "", now, now, nil))

	second, found, err := NewMySQLStore(store.db).FindRunByIdempotencyKey(ctx, "alice", "send-1")
	if err != nil || !found || second.ID != first.ID {
		t.Fatalf("recovered=%+v found=%v err=%v", second, found, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreDuplicateRunMapsToRunInProgress(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	run := Run{ID: "run-2", ProjectID: "project-1", SessionID: "session-1", IdempotencyKey: "send-1", Status: RunQueued}
	mock.ExpectExec("INSERT INTO novel_runs").WillReturnError(&mysql.MySQLError{Number: 1062, Message: "duplicate idempotency key"})

	_, err := store.CreateRun(context.Background(), "alice", run)
	if !IsCode(err, ErrRunInProgress) {
		t.Fatalf("err=%v, want %s", err, ErrRunInProgress)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCreateRunRejectsMissingOrSoftDeletedProject(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	run := Run{ID: "run-3", ProjectID: "deleted-project", SessionID: "session-1", IdempotencyKey: "send-3", Status: RunQueued}
	mock.ExpectExec("INSERT INTO novel_runs").WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := store.CreateRun(context.Background(), "alice", run)
	if !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("err=%v, want %s", err, ErrProjectNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreUpdateRunRejectsSoftDeletedProject(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id
SET r.status = ?, r.updated_at = ? WHERE r.owner_username = ? AND r.run_id = ? AND p.deleted_at IS NULL`)).
		WithArgs(RunWriting, sqlmock.AnyArg(), "alice", "run-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := store.UpdateRun(context.Background(), "alice", RunUpdate{ID: "run-1", Status: RunWriting})
	if !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("err=%v, want %s", err, ErrProjectNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreAppendMessageLocksCursorAndCommits(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	message := Message{ID: "message-1", SessionID: "session-1", ProjectID: "project-1", Role: "user", Content: "继续"}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT s.message_cursor FROM novel_sessions").WithArgs("alice", "session-1", "project-1").WillReturnRows(sqlmock.NewRows([]string{"message_cursor"}).AddRow(3))
	mock.ExpectExec("INSERT INTO novel_messages").WithArgs("message-1", "alice", "session-1", "project-1", int64(4), "user", "继续", "", "", "", nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE novel_sessions SET message_cursor = ?, updated_at = ? WHERE owner_username = ? AND session_id = ? AND project_id = ?`)).WithArgs(int64(4), sqlmock.AnyArg(), "alice", "session-1", "project-1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	saved, err := store.AppendMessage(context.Background(), "alice", message)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Sequence != 4 {
		t.Fatalf("sequence=%d, want 4", saved.Sequence)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreAppendMessageRollsBackOnWriteFailure(t *testing.T) {
	tests := []struct{ name string }{
		{name: "insert"},
		{name: "cursor update"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newNovelCreationMockStore(t)
			message := Message{ID: "message-1", SessionID: "session-1", ProjectID: "project-1", Role: "user", Content: "继续"}
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT s.message_cursor FROM novel_sessions").WithArgs("alice", "session-1", "project-1").WillReturnRows(sqlmock.NewRows([]string{"message_cursor"}).AddRow(0))
			insert := mock.ExpectExec("INSERT INTO novel_messages")
			if tt.name == "insert" {
				insert.WillReturnError(errors.New("insert failed"))
			} else {
				insert.WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE novel_sessions SET message_cursor").WillReturnError(errors.New("cursor update failed"))
			}
			mock.ExpectRollback()

			if _, err := store.AppendMessage(context.Background(), "alice", message); err == nil {
				t.Fatal("AppendMessage error=nil, want rollback-triggering error")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMySQLStoreListsRunEventsAfterCursor(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	query := `SELECT e.event_id, e.run_id, e.session_id, e.project_id, e.event_type, e.step_name, e.status, COALESCE(e.message, ''), e.progress, e.payload_json, e.created_at
FROM novel_run_events e JOIN novel_projects p ON p.owner_username = e.owner_username AND p.project_id = e.project_id
WHERE e.owner_username = ? AND e.run_id = ? AND e.event_id > ? AND p.deleted_at IS NULL ORDER BY e.event_id`
	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("alice", "run-1", int64(7)).WillReturnRows(sqlmock.NewRows([]string{
		"event_id", "run_id", "session_id", "project_id", "event_type", "step_name", "status", "message", "progress", "payload_json", "created_at",
	}).AddRow(8, "run-1", "session-1", "project-1", "status", "planning", RunPlanning, "规划中", 25, []byte(`{"stage":"plan"}`), now))

	events, err := store.ListRunEvents(context.Background(), "alice", "run-1", 7)
	if err != nil || len(events) != 1 || events[0].EventID != 8 || events[0].Status != RunPlanning {
		t.Fatalf("events=%+v err=%v", events, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreSaveChapterBundleRollsBackOnRevisionFailure(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	ctx := context.Background()
	bundle := ChapterBundle{
		ProjectID: "project-1", RunID: "run-1", FinalRunStatus: RunSucceeded,
		Chapter:         Chapter{ID: "chapter-1", Number: 1, Title: "雨夜", Content: "雨落在旧城。", Version: 1},
		ChapterVersion:  ChapterVersion{ID: "version-1", ChapterID: "chapter-1", RunID: "run-1", Version: 1, Content: "雨落在旧城。", Snapshot: []byte(`{}`)},
		Foundation:      Foundation{ID: "foundation-1", Snapshot: []byte(`{}`)},
		Memories:        []Memory{{ID: "memory-1", Snapshot: []byte(`{}`)}},
		Audits:          []Audit{{ID: "audit-1", Snapshot: []byte(`{}`)}},
		Revisions:       []Revision{{ID: "revision-1", Content: "修订版", Snapshot: []byte(`{}`)}},
		RuntimeSnapshot: []byte(`{}`),
	}
	revisionErr := errors.New("revision insert failed")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM novel_projects").WithArgs("alice", "project-1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM novel_runs").WithArgs("alice", "run-1", "project-1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO novel_chapters").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_chapter_versions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_foundations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_memories").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_audits").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_revisions").WillReturnError(revisionErr)
	mock.ExpectRollback()

	err := store.SaveChapterBundle(ctx, "alice", bundle)
	if !errors.Is(err, revisionErr) {
		t.Fatalf("err=%v, want revision failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreSaveChapterBundleRejectsCrossProjectChildBeforeTransaction(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ChapterBundle)
	}{
		{name: "memory project", mutate: func(bundle *ChapterBundle) { bundle.Memories[0].ProjectID = "other-project" }},
		{name: "chapter version chapter", mutate: func(bundle *ChapterBundle) {
			bundle.ChapterVersion.ID, bundle.ChapterVersion.ChapterID = "", "other-chapter"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newNovelCreationMockStore(t)
			bundle := validMySQLChapterBundle()
			tt.mutate(&bundle)

			err := store.SaveChapterBundle(context.Background(), "alice", bundle)
			if !IsCode(err, ErrInputRequired) {
				t.Fatalf("err=%v, want %s", err, ErrInputRequired)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("cross-project child must not begin, write, or rollback a transaction: %v", err)
			}
		})
	}
}

func TestMySQLStoreSaveChapterBundleCommitsAllRows(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	bundle := validMySQLChapterBundle()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM novel_projects").WithArgs("alice", "project-1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM novel_runs").WithArgs("alice", "run-1", "project-1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO novel_chapters").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_chapter_versions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_foundations").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_memories").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_audits").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_revisions").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE novel_runs SET status").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.SaveChapterBundle(context.Background(), "alice", bundle); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func validMySQLChapterBundle() ChapterBundle {
	return ChapterBundle{
		ProjectID: "project-1", RunID: "run-1", FinalRunStatus: RunSucceeded,
		Chapter:         Chapter{ID: "chapter-1", Number: 1, Title: "雨夜", Content: "雨落在旧城。", Version: 1},
		ChapterVersion:  ChapterVersion{ID: "version-1", ChapterID: "chapter-1", RunID: "run-1", Version: 1, Content: "雨落在旧城。", Snapshot: []byte(`{}`)},
		Foundation:      Foundation{ID: "foundation-1", Snapshot: []byte(`{}`)},
		Memories:        []Memory{{ID: "memory-1", Snapshot: []byte(`{}`)}},
		Audits:          []Audit{{ID: "audit-1", Snapshot: []byte(`{}`)}},
		Revisions:       []Revision{{ID: "revision-1", Content: "修订版", Snapshot: []byte(`{}`)}},
		RuntimeSnapshot: []byte(`{}`),
	}
}

var _ Store = (*MySQLStore)(nil)
