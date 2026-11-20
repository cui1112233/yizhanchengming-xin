package novelcreation

import (
	"context"
	"encoding/json"
	"testing"
)

type agentDefaultsAPI interface {
	GetAgentDefaults(context.Context, string) (AgentSettings, error)
	UpdateAgentDefaults(context.Context, string, AgentSettings) (AgentSettings, error)
}

func TestAgentDefaultsAreOwnerScopedInheritedAndFrozen(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), nil)
	api, ok := any(s).(agentDefaultsAPI)
	if !ok {
		t.Fatal("无作品时没有可保存的账号Agent设置")
	}
	initial, e := api.GetAgentDefaults(ctx, "alice")
	if e != nil || !initial.EnableLengthNorm || !initial.EnablePostWrite || !initial.EnableAiTells {
		t.Fatal("未保存账号设置时丢失系统默认", e)
	}
	ids := []int{}
	a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 0, PassScore: 91, ModelRoutes: map[string]string{"chapter-audit": "review-model"}, ReviewDimensionIDs: &ids, CustomWritePrompts: []AgentPrompt{{ID: "custom", Title: "动机要求", Category: "short_fiction", Content: "保留人物动机", Enabled: true}}}, 2000)
	if _, e := api.UpdateAgentDefaults(ctx, "alice", a); e != nil {
		t.Fatal(e)
	}
	b, e := api.GetAgentDefaults(ctx, "bob")
	if e != nil || b.PassScore == 91 {
		t.Fatal("账号设置泄漏", e)
	}
	p, e := s.CreateProject(ctx, "alice", CreateProjectInput{Title: "新作品", Genre: "悬疑", ModelID: "fake"})
	if e != nil {
		t.Fatal(e)
	}
	got, e := s.GetProjectSettings(ctx, "alice", p.ID)
	if e != nil || got.Agent.PassScore != 91 || got.Agent.MaxRevisionRounds != 0 || got.Agent.ReviewDimensionIDs == nil || len(*got.Agent.ReviewDimensionIDs) != 0 {
		t.Fatal("新作品未继承完整设置", got.Agent, e)
	}
	out, e := s.StartShortStory(ctx, "alice", ShortStoryInput{IdempotencyKey: "defaults", Mode: "short_original", ModelID: "fake", Requirements: "写一个完整故事"})
	if e != nil {
		t.Fatal(e)
	}
	var snap struct {
		Settings ProjectSettings `json:"project_settings"`
	}
	if e = json.Unmarshal(out.Run.PromptSnapshot, &snap); e != nil || snap.Settings.Agent.PassScore != 91 || snap.Settings.Agent.ModelRoutes["chapter-audit"] != "review-model" {
		t.Fatal("任务没有冻结账号设置", e)
	}
	a.PassScore = 99
	if _, e = api.UpdateAgentDefaults(ctx, "alice", a); e != nil {
		t.Fatal(e)
	}
	got, e = s.GetProjectSettings(ctx, "alice", p.ID)
	if e != nil || got.Agent.PassScore != 91 {
		t.Fatal("修改默认改动了已有作品", e)
	}
	run, e := st.GetRun(ctx, "alice", out.Run.ID)
	if e != nil || string(run.PromptSnapshot) != string(out.Run.PromptSnapshot) {
		t.Fatal("历史快照被改动", e)
	}
	got.Agent.PassScore = 75
	if _, e = s.UpdateProjectSettings(ctx, "alice", p.ID, got); e != nil {
		t.Fatal(e)
	}
	a, e = api.GetAgentDefaults(ctx, "alice")
	if e != nil || a.PassScore != 99 {
		t.Fatal("作品覆盖污染账号默认", e)
	}
	if _, e = api.UpdateAgentDefaults(ctx, "", a); e == nil {
		t.Fatal("空账号可保存设置")
	}
}
