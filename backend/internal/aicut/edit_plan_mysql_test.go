package aicut

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLCreateEditPlanPersistsPlanAndOrderedOperations(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	plan := editPlanForApply(EditOperation{ID: "op-caption", Type: EditAddCaption, TrackID: "subtitle-1", StartMS: 100, DurationMS: 800, Text: "自动字幕", Reason: "补充字幕", Enabled: true})
	plan.ID, plan.ClientRequestID, plan.Owner, plan.ProjectID = "plan-1", "request-1", "alice", "cut-1"
	plan.State, plan.CreatedAt, plan.UpdatedAt = EditPlanReady, now, now
	requestJSON, _ := json.Marshal(plan.Request)
	summaryJSON, _ := json.Marshal(plan.Summary)
	warningsJSON, _ := json.Marshal(plan.Warnings)
	operationJSON, _ := json.Marshal(plan.Operations[0])

	mock.ExpectQuery("SELECT id, client_request_id, base_revision").
		WithArgs("alice", "cut-1", "request-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO aicut_edit_plans")).
		WithArgs(plan.ID, plan.ClientRequestID, plan.ProjectID, plan.Owner, plan.BaseRevision, plan.ModelID, plan.State, string(requestJSON), string(summaryJSON), string(warningsJSON), "", int64(0), now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO aicut_edit_plan_operations")).
		WithArgs(plan.ID, "op-caption", 0, string(operationJSON)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	created, err := store.CreateEditPlan(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != plan.ID || created.Owner != "alice" || len(created.Operations) != 1 {
		t.Fatalf("created=%#v", created)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLUpdateEditPlanStateRejectsTerminalTransition(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	requestJSON := []byte(`{"assetIds":["asset-video"],"goal":"自动剪辑","targetDurationMs":1000,"aspectRatio":"9:16"}`)
	summaryJSON := []byte(`{"sourceDurationMs":1000,"estimatedDurationMs":1000}`)
	warningsJSON := []byte(`[]`)
	mock.ExpectQuery("SELECT id, client_request_id, base_revision").WithArgs("alice", "cut-1", "plan-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "client_request_id", "base_revision", "model_id", "state", "request_json", "summary_json", "warnings_json", "error_message", "applied_revision", "created_at", "updated_at"}).
			AddRow("plan-1", "request-1", 1, "model-1", EditPlanApplied, requestJSON, summaryJSON, warningsJSON, "", 2, now, now))
	mock.ExpectQuery("SELECT operation_json FROM aicut_edit_plan_operations").WithArgs("plan-1").
		WillReturnRows(sqlmock.NewRows([]string{"operation_json"}))

	if _, err := store.UpdateEditPlanState(context.Background(), "alice", "cut-1", "plan-1", EditPlanCanceled, "", now.Add(time.Second)); err != ErrInvalidEditPlan {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
