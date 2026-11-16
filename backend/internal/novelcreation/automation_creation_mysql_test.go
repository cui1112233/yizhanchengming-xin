package novelcreation

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
	"time"
)

type creationJSONArgument struct{ want *AutomationCreation }

func (a creationJSONArgument) Match(v driver.Value) bool {
	var raw []byte
	switch x := v.(type) {
	case []byte:
		raw = x
	case string:
		raw = []byte(x)
	default:
		return false
	}
	var in AutomationInput
	return json.Unmarshal(raw, &in) == nil && creationMatches(in.Creation, a.want) == nil
}
func creationSQLRows(kind, id string, raw any, add bool) *sqlmock.Rows {
	now := time.Now().UTC()
	if kind == "schedule" {
		r := sqlmock.NewRows([]string{"id", "owner", "project", "session", "interval", "enabled", "prompt", "model", "thinking", "next", "last", "run", "created", "updated", "input"})
		if add {
			r.AddRow(id, "alice", "project", "session", 30, false, "后来编辑", "model", "standard", now, nil, "", now, now, raw)
		}
		return r
	}
	r := sqlmock.NewRows([]string{"id", "owner", "project", "session", "status", "total", "completed", "run", "attempt", "prompt", "model", "thinking", "error", "created", "updated", "input"})
	if add {
		r.AddRow(id, "alice", "project", "session", BatchStopped, 5, 3, "run-current", 4, "后来编辑", "model", "standard", "", now, now, raw)
	}
	return r
}
func TestAutomationCreationIntentSQLAtomicCommitRollback(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		for _, fail := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
				st, m := newNovelCreationMockStore(t)
				ctx := context.Background()
				si := ScheduleInput{IdempotencyKey: "sql-creation", SessionID: "session", Prompt: "原始要求"}
				bi := BatchInput{IdempotencyKey: "sql-creation", SessionID: "session", Prompt: "原始要求"}
				var rawInput any = si
				if kind == "batch" {
					rawInput = bi
				}
				id, meta, _ := automationCreation("alice", "project", kind, "sql-creation", rawInput)
				columns := scheduleColumns
				table := "novel_schedules"
				if kind == "batch" {
					columns = batchColumns
					table = "novel_batches"
				}
				expectMissing := func() {
					m.ExpectQuery("SELECT "+regexp.QuoteMeta(columns)).WithArgs("alice", id).WillReturnRows(creationSQLRows(kind, id, nil, false))
				}
				expectProject := func() {
					m.ExpectQuery("SELECT "+regexp.QuoteMeta(projectColumns)).WithArgs("alice", "project").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "genre", "world", "protagonist", "conflict", "model", "target", "status", "created", "updated", "deleted"}).AddRow("project", "作品", "都市", nil, nil, nil, "model", 2000, "draft", time.Now(), time.Now(), nil))
				}
				expectSession := func() {
					m.ExpectQuery("SELECT "+regexp.QuoteMeta(sessionColumns)).WithArgs("alice", "session").WillReturnRows(sqlmock.NewRows([]string{"id", "project", "mode", "title", "budget", "cursor", "run", "created", "updated", "archived"}).AddRow("session", "project", "short_original", "会话", 120000, 0, "", time.Now(), time.Now(), nil))
				}
				expectMissing()
				m.ExpectBegin()
				m.ExpectQuery("SELECT 1 FROM novel_projects.*FOR UPDATE").WithArgs("alice", "project").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
				expectMissing()
				expectProject()
				expectSession()
				expectSession()
				expectProject()
				m.ExpectQuery("SELECT st.project_id.*FROM novel_project_settings").WithArgs("alice", "project").WillReturnRows(sqlmock.NewRows([]string{"project", "style", "rules", "agent", "updated"}).AddRow("project", nil, "[]", "{}", time.Now()))
				expectProject()
				m.ExpectQuery("SELECT m.material_id.*FROM novel_materials").WithArgs("alice", "project").WillReturnRows(sqlmock.NewRows([]string{"id", "project", "title", "content", "created", "updated"}))
				if kind == "batch" {
					m.ExpectQuery("SELECT "+regexp.QuoteMeta(batchColumns)).WithArgs("alice", "project").WillReturnRows(creationSQLRows(kind, id, nil, false))
				}
				args := []driver.Value{id, "alice", "project", "session"}
				if kind == "schedule" {
					args = append(args, 30, false, "原始要求", "model", "standard", sqlmock.AnyArg(), nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg(), creationJSONArgument{meta}, "alice", "project", "alice", "session", "project")
				} else {
					args = append(args, BatchRunning, 5, 0, nil, 1, "原始要求", "model", "standard", "", sqlmock.AnyArg(), sqlmock.AnyArg(), creationJSONArgument{meta}, "alice", "project", "alice", "session", "project")
				}
				e := m.ExpectExec("INSERT INTO " + table).WithArgs(args...)
				if fail {
					e.WillReturnError(errors.New("insert failed"))
					m.ExpectRollback()
				} else {
					e.WillReturnResult(sqlmock.NewResult(1, 1))
					m.ExpectCommit()
				}
				service := NewService(st, NewMemoryRuntime(), nil)
				var err error
				if kind == "schedule" {
					_, err = service.CreateSchedule(ctx, "alice", "project", si)
				} else {
					_, err = service.CreateBatch(ctx, "alice", "project", bi)
				}
				if (err != nil) != fail {
					t.Fatal(err)
				}
				if err = m.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
func TestAutomationCreationIntentSQLReplayOnlyReadsAndRoundTrips(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			st, m := newNovelCreationMockStore(t)
			ctx := context.Background()
			si := ScheduleInput{IdempotencyKey: "immutable", SessionID: "session", Prompt: "原请求"}
			bi := BatchInput{IdempotencyKey: "immutable", SessionID: "session", Prompt: "原请求"}
			var in any = si
			if kind == "batch" {
				in = bi
			}
			id, meta, _ := automationCreation("alice", "project", kind, "immutable", in)
			raw := inputJSON(AutomationInput{Creation: meta, State: &AutomationState{Status: "stopped", TotalSavedChapters: 7}})
			columns := scheduleColumns
			if kind == "batch" {
				columns = batchColumns
			}
			m.ExpectQuery("SELECT "+regexp.QuoteMeta(columns)).WithArgs("alice", id).WillReturnRows(creationSQLRows(kind, id, raw, true))
			service := NewService(st, NewMemoryRuntime(), nil)
			if kind == "schedule" {
				p, e := service.CreateSchedule(ctx, "alice", "project", si)
				if e != nil || p.ID != id || p.Prompt != "后来编辑" || p.TotalSavedChapters != 7 {
					t.Fatalf("%+v %v", p, e)
				}
				if creationMatches(scheduleStoredInput(p).Creation, meta) != nil {
					t.Fatal("schedule metadata dropped")
				}
			} else {
				p, e := service.CreateBatch(ctx, "alice", "project", bi)
				if e != nil || p.ID != id || p.CompletedChapters != 3 || p.Status != BatchStopped {
					t.Fatalf("%+v %v", p, e)
				}
				if creationMatches(batchStoredInput(p).Creation, meta) != nil {
					t.Fatal("batch metadata dropped")
				}
			}
			// No transaction, mutation, project touch, active-batch or mutable reference query is expected.
			if e := m.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestAutomationCreationIntentSQLUpdatesPreserveOriginalFingerprint(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			st, m := newNovelCreationMockStore(t)
			ctx := context.Background()
			meta := &AutomationCreation{IdempotencyKey: "original-key", Fingerprint: "original-immutable-fingerprint"}
			raw := inputJSON(AutomationInput{Creation: meta})
			columns := scheduleColumns
			if kind == "batch" {
				columns = batchColumns
			}
			expectRead := func() {
				m.ExpectQuery("SELECT "+regexp.QuoteMeta(columns)).WithArgs("alice", "plan").WillReturnRows(creationSQLRows(kind, "plan", raw, true))
			}
			expectRead()
			if kind == "schedule" {
				p, e := st.GetSchedule(ctx, "alice", "plan")
				if e != nil {
					t.Fatal(e)
				}
				p.Prompt = "修改后的要求"
				m.ExpectExec("UPDATE novel_schedules").WithArgs("session", 30, false, p.Prompt, "model", "standard", sqlmock.AnyArg(), nil, nil, sqlmock.AnyArg(), creationJSONArgument{meta}, "alice", "plan", "project").WillReturnResult(sqlmock.NewResult(0, 1))
				expectRead()
				p, e = st.UpdateSchedule(ctx, "alice", p)
				if e != nil || creationMatches(p.Creation, meta) != nil {
					t.Fatalf("%+v %v", p, e)
				}
			} else {
				p, e := st.GetBatch(ctx, "alice", "plan")
				if e != nil {
					t.Fatal(e)
				}
				p.Prompt = "修改后的要求"
				m.ExpectExec("UPDATE novel_batches").WithArgs(BatchStopped, 3, "run-current", 4, p.Prompt, "model", "standard", "", sqlmock.AnyArg(), creationJSONArgument{meta}, "alice", "plan", "project").WillReturnResult(sqlmock.NewResult(0, 1))
				expectRead()
				p, e = st.UpdateBatch(ctx, "alice", p)
				if e != nil || creationMatches(p.Creation, meta) != nil {
					t.Fatalf("%+v %v", p, e)
				}
			}
			if e := m.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}
