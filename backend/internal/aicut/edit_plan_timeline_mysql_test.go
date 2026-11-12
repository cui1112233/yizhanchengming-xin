package aicut

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestMySQLNamedEditPlanApplyCommitsSnapshotHistoryAndStateTogether(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			db, m, e := sqlmock.New()
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
			plan := editPlanForApply(EditOperation{ID: "caption", Type: EditAddCaption, TrackID: "subtitle", StartMS: 0, DurationMS: 1000, Text: "Named", Reason: "Caption", Enabled: true})
			plan.ID, plan.Owner, plan.ProjectID, plan.TimelineID, plan.BaseRevision = "plan", "alice", "p", "named", 1
			plan.Request.TimelineID = "named"
			request, _ := json.Marshal(plan.Request)
			summary, _ := json.Marshal(plan.Summary)
			warnings, _ := json.Marshal(plan.Warnings)
			op, _ := json.Marshal(plan.Operations[0])
			m.ExpectBegin()
			m.ExpectQuery("SELECT id, client_request_id, base_revision").WithArgs("alice", "p", "plan").WillReturnRows(sqlmock.NewRows([]string{"id", "client_request_id", "base_revision", "model_id", "state", "request_json", "summary_json", "warnings_json", "error_message", "applied_revision", "created_at", "updated_at"}).AddRow("plan", plan.ClientRequestID, 1, plan.ModelID, EditPlanReady, string(request), string(summary), string(warnings), "", 0, now, now))
			m.ExpectQuery("SELECT operation_json").WithArgs("plan").WillReturnRows(sqlmock.NewRows([]string{"operation_json"}).AddRow(string(op)))
			m.ExpectQuery("SELECT id FROM aicut_projects").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p"))
			m.ExpectQuery("SELECT revision, cursor_no, timeline_json FROM aicut_timelines").WithArgs("p", "alice", "named").WillReturnRows(sqlmock.NewRows([]string{"revision", "cursor_no", "timeline_json"}).AddRow(1, 0, `{"revision":1,"script":"Named","tracks":[{"id":"subtitle","kind":"subtitle","clips":[]}]}`))
			m.ExpectExec("DELETE FROM aicut_named_timeline_history").WithArgs("p", "alice", "named", 0).WillReturnResult(sqlmock.NewResult(0, 0))
			m.ExpectExec("INSERT INTO aicut_named_timeline_history").WithArgs("p", "alice", "named", 1, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectExec("UPDATE aicut_timelines SET revision").WithArgs(int64(2), 1, sqlmock.AnyArg(), now, "p", "alice", "named").WillReturnResult(sqlmock.NewResult(0, 1))
			m.ExpectExec("UPDATE aicut_edit_plan_operations").WithArgs(sqlmock.AnyArg(), "plan", 0).WillReturnResult(sqlmock.NewResult(0, 1))
			exec := m.ExpectExec("UPDATE aicut_edit_plans SET state").WithArgs(EditPlanApplied, int64(2), now, sqlmock.AnyArg(), "plan")
			if fail {
				exec.WillReturnError(errors.New("state write failed"))
				m.ExpectRollback()
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectCommit()
			}
			stored, v, _, e := applyMySQLNamedPlan(context.Background(), db, "alice", "p", "plan", "named", nil, now)
			if fail {
				if e == nil {
					t.Fatal("failed plan state write committed")
				}
			} else if e != nil || stored.TimelineID != "named" || v.Revision != 2 || v.Tracks[0].Clips[0].Text != "Named" {
				t.Fatalf("stored=%+v snapshot=%+v %v", stored, v, e)
			}
			if e = m.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}
