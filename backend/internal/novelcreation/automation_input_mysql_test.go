package novelcreation

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type automationJSONArgument struct{ want AutomationInput }

func (a automationJSONArgument) Match(value driver.Value) bool {
	var raw []byte
	switch v := value.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return false
	}
	var got AutomationInput
	return json.Unmarshal(raw, &got) == nil && reflect.DeepEqual(got, a.want)
}

func scheduleSQLRows(input any) *sqlmock.Rows {
	now := time.Now().UTC()
	return sqlmock.NewRows([]string{"id", "owner", "project", "session", "interval", "enabled", "prompt", "model", "thinking", "next", "last", "run", "created", "updated", "input"}).AddRow("plan", "alice", "project", "session", 60, true, "创作", "model", "deep", now, nil, "", now, now, input)
}
func batchSQLRows(input any) *sqlmock.Rows {
	now := time.Now().UTC()
	return sqlmock.NewRows([]string{"id", "owner", "project", "session", "status", "total", "completed", "run", "attempt", "prompt", "model", "thinking", "error", "created", "updated", "input"}).AddRow("batch", "alice", "project", "session", BatchRunning, 2, 0, "", 1, "创作", "model", "deep", "", now, now, input)
}
func TestAutomationInputSQLRoundTripAndUpdates(t *testing.T) {
	ctx := context.Background()
	store, mock := newNovelCreationMockStore(t)
	input := AutomationInput{ModeInputs: map[string]string{"referenceText": "参考片段"}, KnowledgeIDs: []string{"material"}, NoteIDs: []string{"note"}, StyleProfileID: "style"}
	now := time.Now().UTC()
	plan := Schedule{ID: "plan", ProjectID: "project", SessionID: "session", IntervalMinutes: 60, Enabled: true, Prompt: "创作", ModelID: "model", ThinkingLevel: "deep", NextRunAt: now, ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID}
	mock.ExpectExec("INSERT INTO novel_schedules").WithArgs("plan", "alice", "project", "session", 60, true, "创作", "model", "deep", now, nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg(), automationJSONArgument{input}, "alice", "project", "alice", "session", "project").WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := store.CreateSchedule(ctx, "alice", plan); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(scheduleColumns)+".*WHERE s.owner_username = \\? AND s.schedule_id = \\?").WithArgs("alice", "plan").WillReturnRows(scheduleSQLRows(inputJSON(input)))
	got, err := store.GetSchedule(ctx, "alice", "plan")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(scheduleInput(got), input) {
		t.Fatalf("schedule=%+v", got)
	}
	input.ConfigurationError = "配置需调整"
	got.ConfigurationError = input.ConfigurationError
	got.Enabled = false
	mock.ExpectExec("UPDATE novel_schedules s.*s.input_snapshot_json = \\?").WithArgs("session", 60, false, "创作", "model", "deep", sqlmock.AnyArg(), nil, nil, sqlmock.AnyArg(), automationJSONArgument{input}, "alice", "plan", "project").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(scheduleColumns)).WithArgs("alice", "plan").WillReturnRows(scheduleSQLRows(inputJSON(input)))
	if _, err = store.UpdateSchedule(ctx, "alice", got); err != nil {
		t.Fatal(err)
	}
	input.ConfigurationError = ""
	batch := Batch{ID: "batch", ProjectID: "project", SessionID: "session", Status: BatchRunning, TotalChapters: 2, Attempt: 1, Prompt: "创作", ModelID: "model", ThinkingLevel: "deep", ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID}
	mock.ExpectExec("INSERT INTO novel_batches").WithArgs("batch", "alice", "project", "session", BatchRunning, 2, 0, nil, 1, "创作", "model", "deep", "", sqlmock.AnyArg(), sqlmock.AnyArg(), automationJSONArgument{input}, "alice", "project", "alice", "session", "project").WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err = store.CreateBatch(ctx, "alice", batch); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(batchColumns)+".*WHERE b.owner_username = \\? AND b.batch_id = \\?").WithArgs("alice", "batch").WillReturnRows(batchSQLRows(inputJSON(input)))
	gotBatch, err := store.GetBatch(ctx, "alice", "batch")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(batchInput(gotBatch), input) {
		t.Fatalf("batch=%+v", gotBatch)
	}
	mock.ExpectExec("UPDATE novel_batches b.*b.input_snapshot_json = \\?").WithArgs(BatchRunning, 0, nil, 1, "创作", "model", "deep", "", sqlmock.AnyArg(), automationJSONArgument{input}, "alice", "batch", "project").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(batchColumns)).WithArgs("alice", "batch").WillReturnRows(batchSQLRows(inputJSON(input)))
	if _, err = store.UpdateBatch(ctx, "alice", gotBatch); err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAutomationInputSQLLegacyNullAndInvalidSnapshot(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	ctx := context.Background()
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(scheduleColumns)).WithArgs("alice", "legacy").WillReturnRows(scheduleSQLRows(nil))
	if got, err := store.GetSchedule(ctx, "alice", "legacy"); err != nil || len(got.ModeInputs) != 0 {
		t.Fatalf("legacy=%+v %v", got, err)
	}
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(batchColumns)).WithArgs("alice", "legacy").WillReturnRows(batchSQLRows(nil))
	if got, err := store.GetBatch(ctx, "alice", "legacy"); err != nil || len(got.ModeInputs) != 0 {
		t.Fatalf("legacy=%+v %v", got, err)
	}
	mock.ExpectQuery("SELECT "+regexp.QuoteMeta(scheduleColumns)).WithArgs("alice", "bad").WillReturnRows(scheduleSQLRows(`{"modeInputs":4}`))
	if _, err := store.GetSchedule(ctx, "alice", "bad"); !IsCode(err, ErrPersistenceFailed) {
		t.Fatalf("corrupt contract accepted: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAutomationInputSQLStyleProfilesOwnerScoped(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	ctx := context.Background()
	mock.ExpectExec("INSERT INTO novel_style_profiles").WithArgs("alice", "style", "文风", "短句").WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := store.SaveStyleProfile(ctx, "alice", StyleProfile{ID: "style", Title: "文风", Content: "短句"}); err != nil {
		t.Fatal(err)
	}
	query := regexp.QuoteMeta("SELECT style_profile_id, title, content FROM novel_style_profiles WHERE owner_username = ? AND style_profile_id = ?")
	mock.ExpectQuery(query).WithArgs("alice", "style").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content"}).AddRow("style", "文风", "短句"))
	if got, err := store.GetStyleProfile(ctx, "alice", "style"); err != nil || got.Content != "短句" {
		t.Fatalf("style=%+v %v", got, err)
	}
	mock.ExpectQuery(query).WithArgs("bob", "style").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content"}))
	if _, err := store.GetStyleProfile(ctx, "bob", "style"); !IsCode(err, ErrInputRequired) || !strings.Contains(err.Error(), "重新选择") {
		t.Fatalf("foreign style=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAutomationInputSQLSessionQueriesUseOnlySessionSchemaColumns(t *testing.T) {
	// Independent of sessionColumns so a mistaken production projection cannot
	// silently become the mock's expected schema.
	const projection = `s.session_id, s.project_id, s.mode, s.title, s.context_budget, s.message_cursor, COALESCE(s.current_run_id, ''), s.created_at, s.updated_at`
	const join = ` FROM novel_sessions s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id `
	now := time.Now().UTC()
	rows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"session_id", "project_id", "mode", "title", "context_budget", "message_cursor", "current_run_id", "created_at", "updated_at"}).AddRow("session", "project", "continuation", "会话", 4096, 2, "run", now, now)
	}
	for _, kind := range []string{"get", "list-project", "list-all"} {
		t.Run(kind, func(t *testing.T) {
			store, mock := newNovelCreationMockStore(t)
			ctx := context.Background()
			var got Session
			if kind == "get" {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT "+projection+join+`WHERE s.owner_username = ? AND s.session_id = ? AND p.deleted_at IS NULL`)).WithArgs("alice", "session").WillReturnRows(rows())
				var err error
				got, err = store.GetSession(ctx, "alice", "session")
				if err != nil {
					t.Fatal(err)
				}
			} else {
				projectID := ""
				query := "SELECT " + projection + join + `WHERE s.owner_username = ? AND p.deleted_at IS NULL`
				if kind == "list-project" {
					projectID = "project"
					mock.ExpectQuery(regexp.QuoteMeta(query+` AND s.project_id = ? ORDER BY s.updated_at DESC`)).WithArgs("alice", "project").WillReturnRows(rows())
				} else {
					mock.ExpectQuery(regexp.QuoteMeta(query + ` ORDER BY s.updated_at DESC`)).WithArgs("alice").WillReturnRows(rows())
				}
				sessions, err := store.ListSessions(ctx, "alice", projectID)
				if err != nil {
					t.Fatal(err)
				}
				if len(sessions) != 1 {
					t.Fatalf("sessions=%v", sessions)
				}
				got = sessions[0]
			}
			if got.ID != "session" || got.ProjectID != "project" || got.CurrentRunID != "run" || got.MessageCursor != 2 || got.ContextBudget != 4096 || got.Mode != "continuation" || got.Title != "会话" || !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(now) {
				t.Fatalf("scan=%+v", got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
