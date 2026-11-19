package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

func TestReviewDimensionsContextAndRevisionEffect(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	ids := []int{}
	for _, d := range novelprompt.ReviewDimensionCatalog() {
		ids = append(ids, d.ID)
	}
	input.Agent = &AgentSettings{ReviewDimensionIDs: &ids, PassScore: 80, MaxRevisionRounds: 1}
	input.StyleProfile = "句式简洁，用动作表达情绪"
	input.ModeInputs = map[string]string{"existingSetting": "角色不能读心，既定时间线从雨夜开始"}
	input.Rules = []string{"不能瞬移"}
	input.Notes = []ProjectNote{{ID: "memo-1", Title: "本章目标", Content: "必须落地指纹线索"}}
	input.KnowledgeMaterials = []KnowledgeMaterial{{ID: "canon-1", Title: "人物档案", Content: "林澜是一名记者"}}
	p := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	result, e := pipelineWithProvider(p).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	if len(result.Audits) != 2 || len(result.Revisions) != 1 || !strings.Contains(result.Chapter.Content, "失踪者的指纹") {
		t.Fatal("review did not change final content")
	}
	for _, i := range []int{3, 4, 5} {
		for _, d := range novelprompt.ReviewDimensionCatalog() {
			if !strings.Contains(p.requests[i].SystemPrompt, d.Title+"："+d.Content) {
				t.Fatalf("dimension %d missing at request %d", d.ID, i)
			}
		}
		var payload map[string]any
		if e = json.Unmarshal([]byte(p.requests[i].UserPrompt), &payload); e != nil {
			t.Fatal(e)
		}
		for _, key := range []string{"style_profile", "mode_inputs", "rules", "notes", "knowledge_materials", "foundation_plan", "chapter_plan", "previous_chapters", "previous_truth", "memory_context"} {
			if _, ok := payload[key]; !ok {
				t.Fatalf("%s missing at request %d", key, i)
			}
		}
	}
	if !strings.Contains(p.requests[4].UserPrompt, "issue-1") || !strings.Contains(p.requests[4].UserPrompt, "指纹线索没有落地") {
		t.Fatal("revision lost actionable issue")
	}
}

func TestReviewDimensionFindingReachesSavedChapter(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	p := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	f.service.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return p, nil })
	ids := []int{6, 32, 33}
	_, e := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, PauseAfterBlueprint: boolPointer(false), ReviewDimensionIDs: &ids, MaxRevisionRounds: 1, PassScore: 80}})
	if e != nil {
		t.Fatal(e)
	}
	r, e := f.service.SendMessage(ctx, "alice", f.sendInput("dimension-review-to-save"))
	if e != nil {
		t.Fatal(e)
	}
	r, e = f.service.Run(ctx, "alice", r.ID)
	if e != nil || r.Status != RunSucceeded {
		t.Fatal(r.Status, e)
	}
	chapters, e := f.store.ListChapters(ctx, "alice", f.project.ID)
	if e != nil || len(chapters) != 1 || !strings.Contains(chapters[0].Content, "失踪者的指纹") || chapters[0].AuditStatus != "passed" {
		t.Fatal("final revised chapter not saved", e, chapters)
	}
	if len(p.requests) != 8 {
		t.Fatal("unexpected provider count", len(p.requests))
	}
	for _, i := range []int{3, 4, 5} {
		if !strings.Contains(p.requests[i].SystemPrompt, "章节备忘偏离") || strings.Contains(p.requests[i].SystemPrompt, "正典事件一致性") {
			t.Fatal("selection changed across review loop", i)
		}
	}
}

func TestReviewDimensionFindingRealMySQL(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("dimension-effect")
	p := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	s := NewService(NewMySQLStore(db), NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return p, nil }))
	project, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "维度审稿修订验收", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: project.ID, Mode: "long_original", Title: "审稿修订"})
	if e != nil {
		t.Fatal(e)
	}
	ids := []int{6, 32, 33}
	_, e = s.UpdateProjectSettings(ctx, owner, project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, ReviewDimensionIDs: &ids, PauseAfterBlueprint: boolPointer(false), MaxRevisionRounds: 1, PassScore: 80}})
	if e != nil {
		t.Fatal(e)
	}
	r, e := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "写第一章，从一封匿名信开始。", ModelID: "fake-only", IdempotencyKey: "effect-run"})
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.Run(ctx, owner, r.ID)
	if e != nil || r.Status != RunSucceeded {
		t.Fatal(r.Status, e)
	}
	chapters, e := NewMySQLStore(db).ListChapters(ctx, owner, project.ID)
	if e != nil || len(chapters) != 1 || !strings.Contains(chapters[0].Content, "失踪者的指纹") {
		t.Fatal("revised output not durable", e)
	}
	for _, tc := range []struct {
		query string
		want  int
	}{{"SELECT COUNT(*) FROM novel_audits WHERE owner_username=? AND run_id=?", 2}, {"SELECT COUNT(*) FROM novel_revisions WHERE owner_username=? AND run_id=?", 1}} {
		var n int
		if e = db.QueryRowContext(ctx, tc.query, owner, r.ID).Scan(&n); e != nil || n != tc.want {
			t.Fatal("audit/revision evidence not durable", n, e)
		}
	}
	t.Log("fake-provider review -> actionable issue -> revision -> re-review -> one saved chapter; MySQL has2audits/1revision, zero paid calls")
}
