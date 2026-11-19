package novelcreation

import (
	"context"
	"encoding/json"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

func TestReviewDimensionCatalogAndSelection(t *testing.T) {
	catalog := novelprompt.ReviewDimensionCatalog()
	if len(catalog) != 37 || catalog[31].Title != "读者期待管理" || catalog[36].Title != "正典事件一致性" {
		t.Fatal("original catalog missing")
	}
	ids := novelprompt.DefaultReviewDimensionIDs()
	if len(ids) != 29 || ids[28] != 33 {
		t.Fatal("original defaults missing")
	}
	f := newServiceFixture(t, &stubPipelineRunner{})
	ctx := context.Background()
	settings, e := f.service.GetProjectSettings(ctx, "alice", f.project.ID)
	if e != nil || settings.Agent.ReviewDimensionIDs == nil || len(*settings.Agent.ReviewDimensionIDs) != 29 {
		t.Fatal("effective defaults", e)
	}
	selected := []int{32, 1}
	settings.Agent.ReviewDimensionIDs = &selected
	saved, e := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, settings)
	if e != nil {
		t.Fatal(e)
	}
	run, e := f.service.SendMessage(ctx, "alice", f.sendInput("dimensions-frozen"))
	if e != nil {
		t.Fatal(e)
	}
	empty := []int{}
	saved.Agent.ReviewDimensionIDs = &empty
	if _, e = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, saved); e != nil {
		t.Fatal(e)
	}
	read, e := f.service.GetProjectSettings(ctx, "alice", f.project.ID)
	if e != nil || read.Agent.ReviewDimensionIDs == nil || len(*read.Agent.ReviewDimensionIDs) != 0 {
		t.Fatal("empty reset to defaults", e)
	}
	var snap struct {
		Settings *ProjectSettings `json:"project_settings"`
	}
	if e = json.Unmarshal(run.PromptSnapshot, &snap); e != nil || snap.Settings == nil || len(*snap.Settings.Agent.ReviewDimensionIDs) != 2 {
		t.Fatal("frozen selection", e, string(run.PromptSnapshot))
	}
	for _, bad := range [][]int{{0}, {38}, {1, 1}} {
		read.Agent.ReviewDimensionIDs = &bad
		if _, e = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, read); !IsCode(e, ErrInputRequired) {
			t.Fatal("invalid accepted", bad, e)
		}
	}
}

func TestReviewDimensionsReachAuditAndRevisionOnly(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	selected := []int{32}
	input.Agent = &AgentSettings{ReviewDimensionIDs: &selected}
	for _, def := range []struct {
		kind string
		want bool
	}{{"audit", true}, {"revision", true}, {"short", true}, {"draft", false}} {
		d := novelprompt.ChapterAudit
		switch def.kind {
		case "revision":
			d = novelprompt.ChapterRevision
		case "short":
			d = novelprompt.ShortReviewPrompt
		case "draft":
			d = novelprompt.ChapterDraft
		}
		p := agentRequirementsSystemPrompt(input, d)
		if strings.Contains(p, "读者期待管理") != def.want || strings.Contains(p, "正典事件一致性") {
			t.Fatal(def.kind, p)
		}
	}
	empty := []int{}
	input.Agent.ReviewDimensionIDs = &empty
	if strings.Contains(agentRequirementsSystemPrompt(input, novelprompt.ChapterAudit), "读者期待管理") {
		t.Fatal("empty injected")
	}
	input.Agent.ReviewDimensionIDs = nil
	if strings.Contains(agentRequirementsSystemPrompt(input, novelprompt.ChapterAudit), "内置审稿维度") {
		t.Fatal("legacy frozen run silently migrated")
	}
}

func TestReviewDimensionsActualProviderRequestAndLegacyOmission(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	ids := []int{1, 32}
	input.Agent = &AgentSettings{ReviewDimensionIDs: &ids, PassScore: 80}
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	if _, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); err != nil {
		t.Fatal(err)
	}
	if len(provider.requests) != 6 {
		t.Fatal("additional model calls", len(provider.requests))
	}
	for i, r := range provider.requests {
		if strings.Contains(r.SystemPrompt, "读者期待管理") != (i == 3) || strings.Contains(r.SystemPrompt, "正典事件一致性") {
			t.Fatal("selection not reflected in actual request", i)
		}
	}
	f := newServiceFixture(t, &stubPipelineRunner{})
	a := AgentSettings{Version: 1, ReviewDimensionIDs: &ids}
	if _, e := f.service.UpdateProjectSettings(context.Background(), "alice", f.project.ID, ProjectSettings{Agent: a}); e != nil {
		t.Fatal(e)
	}
	a.ReviewDimensionIDs = nil // Old clients omit the new field: preserve selection.
	saved, e := f.service.UpdateProjectSettings(context.Background(), "alice", f.project.ID, ProjectSettings{Agent: a})
	if e != nil || saved.Agent.ReviewDimensionIDs == nil || len(*saved.Agent.ReviewDimensionIDs) != 2 {
		t.Fatal("old client reset selection", e)
	}
}

func TestReviewDimensionsRealMySQLRoundTrip(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("dimension-sql")
	s := NewService(NewMySQLStore(db), NewMemoryRuntime(), nil)
	p, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "审稿维度持久化", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	a := AgentSettings{Version: 1}
	ids := []int{1, 32, 37}
	a.ReviewDimensionIDs = &ids
	if _, e = s.UpdateProjectSettings(ctx, owner, p.ID, ProjectSettings{Agent: a}); e != nil {
		t.Fatal(e)
	}
	read, e := NewMySQLStore(db).GetProjectSettings(ctx, owner, p.ID)
	if e != nil || read.Agent.ReviewDimensionIDs == nil || len(*read.Agent.ReviewDimensionIDs) != 3 {
		t.Fatal("fresh-store selection", e)
	}
	empty := []int{}
	a.ReviewDimensionIDs = &empty
	if _, e = s.UpdateProjectSettings(ctx, owner, p.ID, ProjectSettings{Agent: a}); e != nil {
		t.Fatal(e)
	}
	read, e = NewMySQLStore(db).GetProjectSettings(ctx, owner, p.ID)
	if e != nil || read.Agent.ReviewDimensionIDs == nil || len(*read.Agent.ReviewDimensionIDs) != 0 {
		t.Fatal("fresh-store empty list", e)
	}
}
