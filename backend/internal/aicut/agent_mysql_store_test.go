package aicut

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLAgentStorePersistsRunMessageAndOrderedEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 11, 10, 5, 0, 0, 0, time.UTC)
	run := AgentRun{ID: "run-1", ClientRequestID: "request-1", Owner: "alice", ProjectID: "cut-1", Instruction: "粗剪", Mode: AgentModeAgent, ModelID: "model-1", ExpectedRevision: 3, State: AgentQueued, CreatedAt: now, UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO aicut_agent_runs")).
		WithArgs(run.ID, run.ClientRequestID, run.ProjectID, run.Owner, run.Instruction, run.Mode, run.ModelID, run.ActiveSkillID, run.ExpectedRevision, run.State, run.TurnCount, run.Error, run.CreatedAt, run.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	created, err := store.CreateAgentRun(context.Background(), run)
	if err != nil || created.ID != run.ID {
		t.Fatalf("CreateAgentRun() = %#v, %v", created, err)
	}

	message := AgentMessage{ID: "msg-1", RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleUser, Content: run.Instruction, CreatedAt: now}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO aicut_agent_messages")).
		WithArgs(message.ID, message.RunID, message.ProjectID, message.Owner, message.Role, message.Content, message.ToolName, message.ToolCallID, message.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.AppendAgentMessage(context.Background(), message); err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(MAX(sequence_no), 0) FROM aicut_agent_events WHERE run_id = ? AND owner_username = ? AND project_id = ? FOR UPDATE")).
		WithArgs(run.ID, run.Owner, run.ProjectID).
		WillReturnRows(sqlmock.NewRows([]string{"sequence_no"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO aicut_agent_events")).
		WithArgs(run.ID, run.ProjectID, run.Owner, int64(1), AgentEventThinking, "分析中", "", "", int64(0), now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	event, err := store.AppendAgentEvent(context.Background(), AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: AgentEventThinking, Message: "分析中", CreatedAt: now})
	if err != nil || event.Sequence != 1 {
		t.Fatalf("AppendAgentEvent() = %#v, %v", event, err)
	}
	toolResult := ToolResult{CallID: "call-1", Name: "move_clip", OK: true, Message: "片段已移动", Revision: 4}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO aicut_agent_tool_results")).
		WithArgs(run.ID, run.ProjectID, run.Owner, toolResult.CallID, toolResult.Name, toolResult.OK, toolResult.Message, toolResult.Revision, []byte(nil), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.SaveAgentToolResult(context.Background(), run.Owner, run.ProjectID, run.ID, toolResult); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tool_name, ok, message, revision, data_json FROM aicut_agent_tool_results WHERE run_id = ? AND owner_username = ? AND project_id = ? AND tool_call_id = ?")).
		WithArgs(run.ID, run.Owner, run.ProjectID, toolResult.CallID).
		WillReturnRows(sqlmock.NewRows([]string{"tool_name", "ok", "message", "revision", "data_json"}).AddRow(toolResult.Name, true, toolResult.Message, toolResult.Revision, nil))
	loaded, ok, err := store.AgentToolResult(context.Background(), run.Owner, run.ProjectID, run.ID, toolResult.CallID)
	if err != nil || !ok || loaded.Revision != toolResult.Revision {
		t.Fatalf("AgentToolResult() = %#v, %v, %v", loaded, ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
