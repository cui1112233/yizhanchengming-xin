package aicut

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAgentRunStateAndOwnerIsolation(t *testing.T) {
	store := NewMemoryAgentStore()
	now := time.Date(2026, 11, 10, 4, 0, 0, 0, time.UTC)
	run := AgentRun{ID: "run-1", ClientRequestID: "request-1", Owner: "alice", ProjectID: "cut-1", Instruction: "删除第二段", Mode: AgentModeAgent, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	created, err := store.CreateAgentRun(context.Background(), run)
	if err != nil || created.ID != run.ID {
		t.Fatalf("CreateAgentRun() = %#v, %v", created, err)
	}
	if _, err := store.AgentRun(context.Background(), "bob", "cut-1", "run-1"); !errors.Is(err, ErrAgentRunNotFound) {
		t.Fatalf("cross-owner AgentRun error = %v, want ErrAgentRunNotFound", err)
	}
	claimed, err := store.ClaimAgentRun(context.Background(), "run-1", now.Add(time.Second))
	if err != nil || claimed.State != AgentRunning {
		t.Fatalf("ClaimAgentRun() = %#v, %v", claimed, err)
	}
	finished, err := store.UpdateAgentRunState(context.Background(), "run-1", AgentSucceeded, "", 2, now.Add(2*time.Second))
	if err != nil || finished.State != AgentSucceeded || finished.TurnCount != 2 {
		t.Fatalf("UpdateAgentRunState() = %#v, %v", finished, err)
	}
	if _, err := store.UpdateAgentRunState(context.Background(), "run-1", AgentRunning, "", 3, now.Add(3*time.Second)); !errors.Is(err, ErrInvalidAgentTransition) {
		t.Fatalf("terminal transition error = %v, want ErrInvalidAgentTransition", err)
	}
}

func TestAgentRunCreateIsIdempotentByClientRequestID(t *testing.T) {
	store := NewMemoryAgentStore()
	now := time.Date(2026, 11, 10, 4, 0, 0, 0, time.UTC)
	first := AgentRun{ID: "run-1", ClientRequestID: "same-request", Owner: "alice", ProjectID: "cut-1", Instruction: "粗剪", Mode: AgentModeAgent, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	second := first
	second.ID = "run-2"
	created, err := store.CreateAgentRun(context.Background(), first)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := store.CreateAgentRun(context.Background(), second)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.ID != created.ID {
		t.Fatalf("idempotent run ID = %q, want %q", replayed.ID, created.ID)
	}
}

func TestAgentMessagesEventsAndCancelAreProjectScopedAndOrdered(t *testing.T) {
	store := NewMemoryAgentStore()
	now := time.Date(2026, 11, 10, 4, 0, 0, 0, time.UTC)
	run := AgentRun{ID: "run-1", ClientRequestID: "request-1", Owner: "alice", ProjectID: "cut-1", Instruction: "粗剪", Mode: AgentModeAgent, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	if _, err := store.CreateAgentRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendAgentMessage(context.Background(), AgentMessage{ID: "msg-1", RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleUser, Content: run.Instruction, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	first, err := store.AppendAgentEvent(context.Background(), AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: AgentEventThinking, Message: "分析中", CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.AppendAgentEvent(context.Background(), AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: AgentEventToolStart, ToolName: "move_clip", CreatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if first.Sequence != 1 || second.Sequence != 2 {
		t.Fatalf("event sequences = %d,%d, want 1,2", first.Sequence, second.Sequence)
	}
	events, err := store.AgentEvents(context.Background(), run.Owner, run.ProjectID, run.ID, 1)
	if err != nil || len(events) != 1 || events[0].Sequence != 2 {
		t.Fatalf("AgentEvents(after=1) = %#v, %v", events, err)
	}
	canceled, err := store.RequestAgentCancel(context.Background(), run.Owner, run.ProjectID, run.ID, now.Add(time.Second))
	if err != nil || canceled.State != AgentCancelRequested {
		t.Fatalf("RequestAgentCancel() = %#v, %v", canceled, err)
	}
	again, err := store.RequestAgentCancel(context.Background(), run.Owner, run.ProjectID, run.ID, now.Add(2*time.Second))
	if err != nil || again.State != AgentCancelRequested {
		t.Fatalf("idempotent cancel = %#v, %v", again, err)
	}
}
