package aicut

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestAgentLoadsPersistedUIContextBeforeFirstModelCall(t *testing.T) {
	model := &queuedAgentModel{}
	_, agents, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	_, err := agents.AppendAgentEvent(context.Background(), AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: "ui_context", Revision: run.ExpectedRevision, Message: `{"selectedClips":[{"trackId":"video-1","clipId":"clip-b"}],"playheadMs":1234}`})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Run(context.Background(), run.ID); err != nil {
		t.Fatal(err)
	}
	if len(model.requests) == 0 || !strings.Contains(model.requests[0].UserPrompt, `"clipId":"clip-b"`) || !strings.Contains(model.requests[0].UserPrompt, `"playheadMs":1234`) {
		t.Fatalf("initial UI data missing: %#v", model.requests)
	}
	messages, _ := agents.AgentMessages(context.Background(), run.Owner, run.ProjectID, 100)
	for _, m := range messages {
		if m.Role == AgentRoleUser && m.Content != run.Instruction {
			t.Fatalf("UI data leaked into user instruction: %#v", m)
		}
	}
}

func TestAgentRejectsStaleOrInvalidPersistedUIContext(t *testing.T) {
	for _, tc := range []struct {
		name, clip string
		stale      bool
	}{{"missing", "missing", false}, {"stale", "clip-b", true}} {
		t.Run(tc.name, func(t *testing.T) {
			model := &queuedAgentModel{}
			timelines, agents, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
			_, _ = agents.AppendAgentEvent(context.Background(), AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: "ui_context", Revision: run.ExpectedRevision, Message: `{"selectedClips":[{"trackId":"video-1","clipId":"` + tc.clip + `"}],"playheadMs":1234}`})
			if tc.stale {
				current, _ := timelines.Timeline(context.Background(), run.Owner, run.ProjectID)
				_, _ = timelines.SaveTimeline(context.Background(), run.Owner, run.ProjectID, current, current.Revision, run.CreatedAt)
			}
			finished, err := service.Run(context.Background(), run.ID)
			if err == nil || finished.State != AgentFailed || len(model.requests) != 0 {
				t.Fatalf("invalid UI context reached model: %#v %v requests=%d", finished, err, len(model.requests))
			}
		})
	}
}

func TestAgentSnapshotContainsOnlySurvivingInitialSelection(t *testing.T) {
	ctx := context.Background()
	_, agents, run, tools := newAgentToolFixture(t, false)
	_, _ = agents.AppendAgentEvent(ctx, AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: "ui_context", Revision: run.ExpectedRevision, Message: `{"selectedClips":[{"trackId":"video-1","clipId":"clip-b"}],"playheadMs":1234}`})
	snapshot, err := tools.Execute(ctx, run, ToolCall{ID: "before", Name: "get_project_snapshot", Arguments: []byte(`{}`)})
	if err != nil || !strings.Contains(string(snapshot.Data), `"initialUiContext":{"selectedClips":[{"trackId":"video-1","clipId":"clip-b"}],"playheadMs":1234}`) {
		t.Fatalf("missing initial selection: %s %v", snapshot.Data, err)
	}
	deleted, err := tools.Execute(ctx, run, ToolCall{ID: "delete", Name: "delete_clip", Arguments: []byte(`{"trackId":"video-1","clipId":"clip-b"}`)})
	if err != nil || !deleted.OK {
		t.Fatalf("delete: %#v %v", deleted, err)
	}
	run.ExpectedRevision = deleted.Revision
	snapshot, err = tools.Execute(ctx, run, ToolCall{ID: "after", Name: "get_project_snapshot", Arguments: []byte(`{}`)})
	if err != nil || !strings.Contains(string(snapshot.Data), `"initialUiContext":{"selectedClips":[],"playheadMs":1234}`) {
		t.Fatalf("deleted selection retained: %s %v", snapshot.Data, err)
	}
}

func TestAgentNamedUIContextUsesPersistedTimeline(t *testing.T) {
	ctx := context.Background()
	base := NewMemoryStore()
	now := time.Now()
	_ = base.CreateProject(ctx, "alice", Project{ID: "p", Title: "P"})
	collection := NewMemoryTimelineCollectionStore(base)
	named, _ := collection.Create(ctx, "alice", "p", "A")
	timeline, err := collection.Save(ctx, "alice", "p", named.ID, Timeline{Tracks: []Track{{ID: "t", Kind: TrackSubtitle, Clips: []Clip{{ID: "named-clip", Kind: ClipSubtitle, Text: "named", DurationMS: 1000}}}}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	agents := NewMemoryAgentStore()
	run := AgentRun{ID: "named-run", ClientRequestID: "named-request", Owner: "alice", ProjectID: "p", TimelineID: named.ID, Instruction: "剪这段", Mode: AgentModeAgent, ExpectedRevision: timeline.Revision, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	_, _ = agents.CreateAgentRun(ctx, run)
	_, _ = agents.AppendAgentEvent(ctx, AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: AgentEventUIContext, Revision: run.ExpectedRevision, Message: `{"selectedClips":[{"trackId":"t","clipId":"named-clip"}],"playheadMs":200}`})
	model := &queuedAgentModel{}
	tools := NewAgentToolExecutor(base, agents, time.Now, collection)
	service := NewAgentService(agents, tools, func(context.Context, string, string) (AgentModel, error) { return model, nil }, func(AgentMode) string { return "prompt" }, time.Now)
	if _, err := service.Run(ctx, run.ID); err != nil {
		t.Fatal(err)
	}
	if len(model.requests) != 1 || !strings.Contains(model.requests[0].UserPrompt, `"clipId":"named-clip"`) {
		t.Fatalf("named selection missing: %#v", model.requests)
	}
}
