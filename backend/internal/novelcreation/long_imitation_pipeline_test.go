package novelcreation

// queuedPipelineProvider supplies deterministic fake responses only. These
// tests verify real pipeline execution and inputs, not paid-model output quality
// or a genuinely generated 50,000-character chapter.

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func imitationAnalysisJSON() string {
	dimensions := map[string]string{}
	for _, d := range ImitationStyleDimensions() {
		dimensions[d.Key] = "依据原文的真实分析示例"
	}
	raw, _ := json.Marshal(LongImitationAnalysis{Dimensions: dimensions, EmotionCurve: []ImitationEmotion{{"开端", "紧张", "冲突推动"}}, CoreValues: []string{"守护与选择"}, Directions: []ImitationDirection{{"方向一", "原创侦探成长"}, {"方向二", "星际寻亲"}}})
	return string(raw)
}
func TestLongImitationRealPipelineUsesAnalysisAndAutoNamesAtCheckpoint(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: []string{imitationAnalysisJSON(), foundationJSON}}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	in := imitationInput("chain")
	in.SelectedDirection = ""
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation || len(provider.requests) != 2 {
		t.Fatalf("analysis+foundation checkpoint: %+v %v calls=%d", run, e, len(provider.requests))
	}
	project, _ := st.GetProject(ctx, "alice", out.Project.ID)
	session, _ := st.GetSession(ctx, "alice", out.Session.ID)
	if project.Title != "雨幕记者" || session.Title != "雨幕记者" {
		t.Fatal("permanent fake title", project.Title, session.Title)
	}
	config, e := s.GetLongImitation(ctx, "alice", out.Project.ID)
	if e != nil || config.Analysis == nil || config.Analysis.SelectedDirection != "方向一：原创侦探成长" || config.Input.Title != "" {
		t.Fatalf("analysis config=%+v analysis=%+v err=%v", config, config.Analysis, e)
	}
	for i, req := range provider.requests {
		if !strings.Contains(req.UserPrompt, "50000") || !strings.Contains(req.UserPrompt, in.ReaderPreferences) || !strings.Contains(req.UserPrompt, in.SupplementaryInfo) {
			t.Fatal("full fields missing from actual provider request", i)
		}
	}
	if !strings.Contains(provider.requests[0].UserPrompt, in.ReferenceText) || strings.Contains(provider.requests[1].UserPrompt, in.ReferenceText) {
		t.Fatal("reference must occur only in its analysis chunk, not blueprint")
	}
	if !strings.Contains(provider.requests[1].UserPrompt, "原创侦探成长") {
		t.Fatal("foundation did not receive actual analysis")
	}
	provider.responses = append(provider.responses, strings.Replace(chapterPlanJSON, `"targetWords":2000`, `"targetWords":50000`, 1), draftJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON)
	_, e = st.UpdateProjectSettings(ctx, "alice", project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, TargetChapterWords: 50000, PassScore: 80}})
	if e != nil {
		t.Fatal(e)
	}
	next, e := s.ConfirmRun(ctx, "alice", run.ID, nil)
	if e != nil {
		t.Fatal(e)
	}
	finished, e := s.Run(ctx, "alice", next.ID)
	if e != nil || !isTerminalRunStatus(finished.Status) {
		t.Fatalf("writing continuation failed status=%s code=%s err=%v", finished.Status, finished.ErrorCode, e)
	}
	if len(provider.requests) != 8 {
		t.Fatal("analysis was billed again or unexpected stages", len(provider.requests))
	}
	if provider.requests[3].MaxTokens != 104000 {
		t.Fatal("50k output silently kept old token cap", provider.requests[3].MaxTokens)
	}
	if !strings.Contains(provider.requests[2].UserPrompt, `"target_chapter_words":50000`) {
		t.Fatal("agent silently clamped 50k to 8k")
	}
}
func TestLongImitationUnknownAnalysisNeverResends(t *testing.T) {
	provider := &queuedPipelineProvider{errors: []error{errors.New("connection lost")}}
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	out, e := s.StartLongImitation(context.Background(), "alice", imitationInput("unknown"))
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(context.Background(), "alice", out.Run.ID)
	if !IsCode(e, ErrResultUnknown) || run.Status != RunPaused {
		t.Fatalf("unknown not paused %+v %v", run, e)
	}
	if _, e = s.RetryRun(context.Background(), "alice", out.Run.ID); !IsCode(e, ErrResultUnknown) {
		t.Fatal("unknown retry accepted", e)
	}
	if len(provider.requests) != 1 {
		t.Fatal("unknown external result rebilled")
	}
}
func TestLongImitationOutlineAndChapterPlanEnforceFrozenTargets(t *testing.T) {
	c := LongImitationCreation{Input: imitationInput("check")}
	var f foundationPlanOutput
	_ = json.Unmarshal([]byte(foundationJSON), &f)
	if e := validateImitationFoundation(f, &c); e != nil {
		t.Fatal(e)
	}
	c.Input.TargetChapters = 19
	if e := validateImitationFoundation(f, &c); e == nil {
		t.Fatal("outline exceeded requested chapter count")
	}
	c.Input.TargetChapters = 20
	var plan chapterPlanOutput
	_ = json.Unmarshal([]byte(chapterPlanJSON), &plan)
	plan.TargetWords = 50000
	if e := validateImitationPlan(plan, 1, f.Volumes, &c); e != nil {
		t.Fatal(e)
	}
	plan.TargetWords = 8000
	if e := validateImitationPlan(plan, 1, f.Volumes, &c); e == nil {
		t.Fatal("planner silently changed target")
	}
}
