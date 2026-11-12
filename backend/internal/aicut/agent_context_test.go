package aicut

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAgentLoadsSameTimelineHistoryAndSkill(t *testing.T) {
	model := &queuedAgentModel{responses: []string{`{"reply":"明白","toolCalls":[]}`}}
	_, store, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	old := run
	old.ID = "older-run"
	old.ClientRequestID = "older-request"
	old.ActiveSkillID = "字幕选片"
	if _, err := store.CreateAgentRun(context.Background(), old); err != nil {
		t.Fatal(err)
	}
	store.AppendAgentMessage(context.Background(), AgentMessage{ID: "old-message", RunID: old.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleUser, Content: "保留介绍产品的部分", CreatedAt: time.Now().Add(-time.Minute)})
	other := old
	other.ID = "other-timeline"
	other.ClientRequestID = "other-request"
	other.TimelineID = "named-timeline"
	if _, err := store.CreateAgentRun(context.Background(), other); err != nil {
		t.Fatal(err)
	}
	store.AppendAgentMessage(context.Background(), AgentMessage{ID: "other-message", RunID: other.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleUser, Content: "这是另一个时间线的秘密指令"})
	for i := 0; i < 100; i++ {
		store.AppendAgentMessage(context.Background(), AgentMessage{ID: fmt.Sprintf("other-%d", i), RunID: other.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleUser, Content: "其他时间线消息"})
		store.AppendAgentMessage(context.Background(), AgentMessage{ID: fmt.Sprintf("tool-%d", i), RunID: old.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleTool, Content: "工具消息"})
	}
	service.ConfigureSkills(func(id string) (string, bool) {
		return "先读取字幕，再按完整句子选片", id == "字幕选片"
	})
	// Store the selected skill in the actual claimed run.
	store.mu.Lock()
	selected := store.runs[run.ID]
	selected.ActiveSkillID = "字幕选片"
	store.runs[run.ID] = selected
	store.mu.Unlock()
	if _, err := service.Run(context.Background(), run.ID); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(model.requests[0].SystemPrompt, "先读取字幕") || !strings.Contains(model.requests[0].UserPrompt, "保留介绍产品") {
		t.Fatalf("missing skill/history: %#v", model.requests[0])
	}
	if strings.Contains(model.requests[0].UserPrompt, "秘密指令") {
		t.Fatal("history crossed timelines")
	}
}

func TestAgentFollowupDoesNotExecuteMutations(t *testing.T) {
	model := &queuedAgentModel{responses: []string{`{"reply":"先确认范围","followup":{"question":"希望保留什么？","fields":[{"id":"topic","label":"主题","required":true,"options":["产品","访谈"]}]},"toolCalls":[{"id":"delete","name":"delete_clip","arguments":{"trackId":"video-1","clipId":"clip-b"}}]}`}}
	timelines, store, run, service := newAgentServiceFixture(t, AgentModeAgent, model)
	finished, err := service.Run(context.Background(), run.ID)
	if err != nil || finished.State != AgentSucceeded {
		t.Fatalf("%#v %v", finished, err)
	}
	timeline, _ := timelines.Timeline(context.Background(), run.Owner, run.ProjectID)
	if len(timeline.Tracks[0].Clips) != 2 {
		t.Fatal("followup executed deletion")
	}
	events, _ := store.AgentEvents(context.Background(), run.Owner, run.ProjectID, run.ID, 0)
	found := false
	for _, event := range events {
		if event.Type == "followup" {
			var value map[string]any
			if json.Unmarshal([]byte(event.Message), &value) != nil {
				t.Fatal("invalid followup event")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("missing structured followup")
	}
}
