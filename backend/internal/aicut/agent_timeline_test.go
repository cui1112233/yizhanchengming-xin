package aicut

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestAgentNamedRunUsesPersistedTargetAndReplayCannotRetarget(t *testing.T) {
	ctx := context.Background()
	base := NewMemoryStore()
	now := time.Now()
	_ = base.CreateProject(ctx, "alice", Project{ID: "p", Title: "P", CreatedAt: now, UpdatedAt: now})
	collection := NewMemoryTimelineCollectionStore(base)
	a, _ := collection.Create(ctx, "alice", "p", "A")
	v, _ := collection.Save(ctx, "alice", "p", a.ID, Timeline{Tracks: []Track{{ID: "t", Kind: TrackSubtitle, Clips: []Clip{{ID: "c", Kind: ClipSubtitle, Text: "Before", DurationMS: 1000}}}}}, 0, now)
	agents := NewMemoryAgentStore()
	run := AgentRun{ID: "run", ClientRequestID: "request", Owner: "alice", ProjectID: "p", TimelineID: a.ID, Instruction: "Correct", Mode: AgentModeAgent, ExpectedRevision: v.Revision, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	if _, e := agents.CreateAgentRun(ctx, run); e != nil {
		t.Fatal(e)
	}
	model := &queuedAgentModel{responses: []string{`{"reply":"","toolCalls":[{"id":"call","name":"replace_subtitle_text","arguments":{"from":"Before","to":"After"}}]}`, `{"reply":"Done","toolCalls":[]}`}}
	tools := NewAgentToolExecutor(base, agents, func() time.Time { return now }, collection)
	service := NewAgentService(agents, tools, func(context.Context, string, string) (AgentModel, error) { return model, nil }, func(AgentMode) string { return "prompt" }, func() time.Time { return now })
	if _, e := service.Run(ctx, "run"); e != nil {
		t.Fatal(e)
	}
	got, _ := collection.Timeline(ctx, "alice", "p", a.ID)
	legacy, _ := base.Timeline(ctx, "alice", "p")
	if got.Tracks[0].Clips[0].Text != "After" || legacy.Revision != 0 {
		t.Fatalf("target=%+v legacy=%+v", got, legacy)
	}
	replay := run
	replay.ID = "other"
	replay.TimelineID = ""
	if _, e := agents.CreateAgentRun(ctx, replay); !errors.Is(e, ErrInvalidAgentRun) {
		t.Fatalf("retarget replay=%v", e)
	}
	call := ToolCall{ID: "call", Name: "replace_subtitle_text", Arguments: json.RawMessage(`{"from":"After","to":"Again"}`)}
	if _, e := tools.Execute(ctx, run, call); e != nil {
		t.Fatal(e)
	}
	got, _ = collection.Timeline(ctx, "alice", "p", a.ID)
	if got.Revision != 2 || got.Tracks[0].Clips[0].Text != "After" {
		t.Fatalf("replayed tool mutated=%+v", got)
	}
}
