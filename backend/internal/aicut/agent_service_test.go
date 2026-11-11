package aicut

import (
	"context"
	"strings"
	"testing"
	"time"
)

type queuedAgentModel struct {
	responses []string
	requests  []AgentModelRequest
	onCall    func(int)
}

func (m *queuedAgentModel) Complete(_ context.Context, request AgentModelRequest) (string, error) {
	m.requests = append(m.requests, request)
	index := len(m.requests) - 1
	if m.onCall != nil {
		m.onCall(index)
	}
	if index >= len(m.responses) {
		return `{"reply":"完成","toolCalls":[]}`, nil
	}
	return m.responses[index], nil
}

func newAgentServiceFixture(t *testing.T, mode AgentMode, model *queuedAgentModel) (*MemoryStore, *MemoryAgentStore, AgentRun, *AgentService) {
	t.Helper()
	timelines, agents, run, tools := newAgentToolFixture(t, false)
	run.Mode = mode
	// newAgentToolFixture 已写入 run；测试模式需要同步到独立的新 store。
	agents = NewMemoryAgentStore()
	if _, err := agents.CreateAgentRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	tools = NewAgentToolExecutor(timelines, agents, func() time.Time { return time.Date(2026, 11, 10, 7, 0, 0, 0, time.UTC) })
	service := NewAgentService(agents, tools, func(context.Context, string, string) (AgentModel, error) { return model, nil }, func(mode AgentMode) string { return "mode=" + string(mode) }, func() time.Time { return time.Date(2026, 11, 10, 7, 0, 0, 0, time.UTC) })
	return timelines, agents, run, service
}

func TestAgentQAModeNeverExecutesToolCalls(t *testing.T) {
	model := &queuedAgentModel{responses: []string{`{"reply":"问答模式不会改时间轴。","toolCalls":[{"id":"call-1","name":"delete_clip","arguments":{"trackId":"video-1","clipId":"clip-b"}}]}`}}
	timelines, agents, run, service := newAgentServiceFixture(t, AgentModeAsk, model)
	finished, err := service.Run(context.Background(), run.ID)
	if err != nil || finished.State != AgentSucceeded {
		t.Fatalf("Run() = %#v, %v", finished, err)
	}
	timeline, _ := timelines.Timeline(context.Background(), run.Owner, run.ProjectID)
	if len(timeline.Tracks[0].Clips) != 2 {
		t.Fatalf("Q&A changed timeline: %#v", timeline)
	}
	events, _ := agents.AgentEvents(context.Background(), run.Owner, run.ProjectID, run.ID, 0)
	for _, event := range events {
		if event.Type == AgentEventToolStart {
			t.Fatalf("Q&A emitted tool start: %#v", events)
		}
	}
}

func TestAgentFeedsToolResultBackIntoNextModelTurn(t *testing.T) {
	model := &queuedAgentModel{responses: []string{
		`{"reply":"","toolCalls":[{"id":"call-1","name":"move_clip","arguments":{"trackId":"video-1","clipId":"clip-b","startMs":3000}}]}`,
		`{"reply":"已把第二段移到3秒。","toolCalls":[]}`,
	}}
	_, _, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	finished, err := service.Run(context.Background(), run.ID)
	if err != nil || finished.State != AgentSucceeded || len(model.requests) != 2 {
		t.Fatalf("Run() = %#v, %v, requests=%d", finished, err, len(model.requests))
	}
	if !strings.Contains(model.requests[1].UserPrompt, "片段已移动") || !strings.Contains(model.requests[1].UserPrompt, `"revision":2`) {
		t.Fatalf("second prompt missing tool result: %s", model.requests[1].UserPrompt)
	}
}

func TestAgentCancelStopsBeforeNextTool(t *testing.T) {
	model := &queuedAgentModel{responses: []string{`{"reply":"","toolCalls":[{"id":"call-1","name":"delete_clip","arguments":{"trackId":"video-1","clipId":"clip-b"}}]}`}}
	_, agents, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	model.onCall = func(int) {
		_, _ = agents.RequestAgentCancel(context.Background(), run.Owner, run.ProjectID, run.ID, time.Now())
	}
	finished, err := service.Run(context.Background(), run.ID)
	if err != nil || finished.State != AgentCanceled {
		t.Fatalf("Run(cancel) = %#v, %v", finished, err)
	}
	if _, ok, _ := agents.AgentToolResult(context.Background(), run.Owner, run.ProjectID, run.ID, "call-1"); ok {
		t.Fatal("tool executed after cancel")
	}
}

func TestAgentBadJSONRetriesOnceThenFails(t *testing.T) {
	model := &queuedAgentModel{responses: []string{"not-json", "still-not-json"}}
	_, _, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	finished, err := service.Run(context.Background(), run.ID)
	if err == nil || finished.State != AgentFailed || len(model.requests) != 2 {
		t.Fatalf("Run(bad json) = %#v, %v, requests=%d", finished, err, len(model.requests))
	}
}

func TestAgentStopsAfterThirtyTurns(t *testing.T) {
	responses := make([]string, 30)
	for index := range responses {
		responses[index] = `{"reply":"","toolCalls":[{"id":"read-` + string(rune('a'+index%26)) + `","name":"get_project_snapshot","arguments":{}}]}`
	}
	model := &queuedAgentModel{responses: responses}
	_, _, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	finished, err := service.Run(context.Background(), run.ID)
	if err == nil || finished.State != AgentFailed || finished.TurnCount != 30 {
		t.Fatalf("Run(30 turns) = %#v, %v", finished, err)
	}
}
