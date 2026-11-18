package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const requirementsJSON = `{"version":1,"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false,"reviewPrompts":[{"id":"r1","kind":"custom","title":"审稿标记","content":"审核必须检查伏笔甲","enabled":true},{"id":"r2","kind":"custom","title":"停用审稿","content":"停用内容乙","enabled":false}],"customWritePrompts":[{"id":"w1","category":"long_fiction","title":"长篇要求","content":"长篇要求丙","enabled":true},{"id":"w2","category":"short_fiction","title":"短篇要求","content":"短篇要求丁","enabled":true},{"id":"w3","category":"imitation","title":"仿写要求","content":"仿写要求戊","enabled":true},{"id":"w4","category":"long_fiction","title":"停用要求","content":"停用写作己","enabled":false}]}`

func TestAgentRequirementsSurviveJSONAndNormalization(t *testing.T) {
	var a AgentSettings
	if err := json.Unmarshal([]byte(requirementsJSON), &a); err != nil {
		t.Fatal(err)
	}
	a = NormalizeAgentSettings(a, 2000)
	raw, _ := json.Marshal(a)
	for _, text := range []string{"审核必须检查伏笔甲", "停用内容乙", "长篇要求丙", "短篇要求丁", "仿写要求戊"} {
		if !strings.Contains(string(raw), text) {
			t.Fatalf("requirement lost: %s", text)
		}
	}
}

func TestAgentRequirementsReachOnlyRelevantCalls(t *testing.T) {
	for _, tc := range []struct{ mode, want string }{{"long_original", "长篇要求丙"}, {"short_original", "短篇要求丁"}, {"long_imitation", "仿写要求戊"}, {"short_imitation", "仿写要求戊"}, {"fanfiction", "长篇要求丙"}, {"continuation", "长篇要求丙"}} {
		t.Run(tc.mode, func(t *testing.T) {
			input := fixturePipelineRunContext(NewMemoryRuntime())
			input.Session.Mode = tc.mode
			var a AgentSettings
			_ = json.Unmarshal([]byte(requirementsJSON), &a)
			input.Agent = &a
			provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
			if _, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); err != nil {
				t.Fatal(err)
			}
			if len(provider.requests) != 6 {
				t.Fatalf("extra paid call: %d", len(provider.requests))
			}
			for i, request := range provider.requests {
				s := request.SystemPrompt
				if strings.Contains(s, "停用内容乙") || strings.Contains(s, "停用写作己") {
					t.Fatal("disabled requirement injected")
				}
				if (i == 2) != strings.Contains(s, tc.want) {
					t.Fatalf("writing requirement missing/leaked at %d: %s", i, s)
				}
				if (i == 3) != strings.Contains(s, "审核必须检查伏笔甲") {
					t.Fatalf("review requirement missing/leaked at %d", i)
				}
				for _, other := range []string{"长篇要求丙", "短篇要求丁", "仿写要求戊"} {
					if other != tc.want && strings.Contains(s, other) {
						t.Fatal("category leaked")
					}
				}
			}
		})
	}
}

func TestAgentRequirementsValidationAndFrozenSave(t *testing.T) {
	ctx := context.Background()
	runner := &stubPipelineRunner{}
	f := newServiceFixture(t, runner)
	var a AgentSettings
	_ = json.Unmarshal([]byte(requirementsJSON), &a)
	if _, err := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a}); err != nil {
		t.Fatal(err)
	}
	run, err := f.service.SendMessage(ctx, "alice", f.sendInput("frozen-requirements"))
	if err != nil {
		t.Fatal(err)
	}
	a.CustomWritePrompts[0].Content = "新内容不能污染已排队任务"
	saved, err := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a})
	if err != nil {
		t.Fatal(err)
	}
	a.CustomWritePrompts[0].Content = "不能通过入参改数据库"
	saved.Agent.ReviewPrompts[0].Content = "不能通过返回值改数据库"
	read, err := f.service.GetProjectSettings(ctx, "alice", f.project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Agent.CustomWritePrompts[0].Content != "新内容不能污染已排队任务" || read.Agent.ReviewPrompts[0].Content != "审核必须检查伏笔甲" {
		t.Fatal("settings alias leaked")
	}
	if _, err = f.service.GetProjectSettings(ctx, "bob", f.project.ID); err == nil {
		t.Fatal("cross owner access")
	}
	if _, err = f.service.Run(ctx, "alice", run.ID); err != nil {
		t.Fatal(err)
	}
	if runner.lastCall().Agent.CustomWritePrompts[0].Content != "长篇要求丙" {
		t.Fatal("queued snapshot mutated")
	}
	for _, tc := range []struct {
		name string
		edit func(*AgentSettings)
	}{
		{"enabled blank", func(v *AgentSettings) { v.ReviewPrompts[0].Content = "  " }},
		{"duplicate id", func(v *AgentSettings) { v.ReviewPrompts[1].ID = v.ReviewPrompts[0].ID }},
		{"category", func(v *AgentSettings) { v.CustomWritePrompts[0].Category = "missing" }},
		{"title", func(v *AgentSettings) { v.CustomWritePrompts[0].Title = " " }},
		{"too long", func(v *AgentSettings) { v.CustomWritePrompts[0].Content = strings.Repeat("字", 8001) }},
		{"too many", func(v *AgentSettings) {
			for len(v.ReviewPrompts) < 51 {
				v.ReviewPrompts = append(v.ReviewPrompts, v.ReviewPrompts[0])
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var bad AgentSettings
			_ = json.Unmarshal([]byte(requirementsJSON), &bad)
			tc.edit(&bad)
			if _, e := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: bad}); !IsCode(e, ErrInputRequired) {
				t.Fatalf("invalid accepted: %v", e)
			}
		})
	}
	// Disabled blank drafts are valid; deleting every item persists empty lists.
	a.CustomWritePrompts = nil
	a.ReviewPrompts = []AgentPrompt{{ID: "draft", Kind: "custom", Title: "草稿", Enabled: false}}
	if _, err = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a}); err != nil {
		t.Fatal(err)
	}
	a.ReviewPrompts = nil
	if _, err = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a}); err != nil {
		t.Fatal(err)
	}
	read, _ = f.service.GetProjectSettings(ctx, "alice", f.project.ID)
	if len(read.Agent.ReviewPrompts) != 0 || len(read.Agent.CustomWritePrompts) != 0 {
		t.Fatal("deletion not durable")
	}
}

func TestAgentRequirementsRevisionAndRepairKeepContext(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	var a AgentSettings
	_ = json.Unmarshal([]byte(requirementsJSON), &a)
	a.MaxRevisionRounds = 1
	input.Agent = &a
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, "invalid json", draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	if _, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); err != nil {
		t.Fatal(err)
	}
	for _, i := range []int{2, 3, 5} {
		if !strings.Contains(provider.requests[i].SystemPrompt, "长篇要求丙") {
			t.Fatalf("lost requirements at call %d", i)
		}
	}
	if !strings.Contains(provider.requests[5].SystemPrompt, "审核必须检查伏笔甲") {
		t.Fatal("revision lost review requirements")
	}
}
