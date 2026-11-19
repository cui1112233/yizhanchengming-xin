package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestNativePostWriteErrorParticipatesInReviewCycle(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ReviewCycleVersion = 1
	input.Agent = &AgentSettings{Version: 1, EnablePostWrite: true, MaxRevisionRounds: 1, PassScore: 80}
	bad := `{"content":"这不是一场梦而是一场雨——核心动机，Chapter 2。"}`
	fix := `{"strategy":"移除分析术语及章节引用，重写动作","content":"雨水敲着窗。林澜拆开了那封没有寄件人的信。","resolvedIssueIDs":[]}`
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, bad, passedAuditJSON, fix, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	if len(provider.requests) != 8 || len(bundle.Audits) != 2 || len(bundle.Revisions) != 1 {
		t.Fatalf("local hard errors did not enter real revise loop: calls=%d audits=%d revisions=%d", len(provider.requests), len(bundle.Audits), len(bundle.Revisions))
	}
	if strings.Contains(bundle.Chapter.Content, "核心动机") || bundle.FinalRunStatus != RunSucceeded {
		t.Fatal("bad candidate persisted", bundle.Chapter.Content, bundle.FinalRunStatus)
	}
	if !strings.Contains(provider.requests[4].UserPrompt, "post_write") || !strings.Contains(provider.requests[4].UserPrompt, "critical") {
		t.Fatal("revision did not receive deterministic blocking issues")
	}
	var first chapterAuditOutput
	if e = json.Unmarshal(bundle.Audits[0].Snapshot, &first); e != nil {
		t.Fatal(e)
	}
	if first.Passed == nil || *first.Passed || len(first.Issues) < 4 || *first.Score != 92 {
		t.Fatal("native merge must block without inventing a score penalty", string(bundle.Audits[0].Snapshot))
	}
}

func TestLegacyFrozenRunKeepsReportingOnlyPostWrite(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.Agent = &AgentSettings{Version: 1, EnablePostWrite: true, MaxRevisionRounds: 1, PassScore: 80}
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, `{"content":"这不是一场梦而是一场雨——核心动机，Chapter 2。"}`, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	if len(provider.requests) != 6 || len(bundle.Revisions) != 0 || !strings.Contains(bundle.Chapter.Content, "核心动机") {
		t.Fatal("old journaled task silently migrated")
	}
}

func TestNativePostWriteSwitchAndZeroRevisionBudget(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		input := fixturePipelineRunContext(NewMemoryRuntime())
		input.ReviewCycleVersion = 1
		input.Agent = &AgentSettings{Version: 1, EnablePostWrite: enabled, MaxRevisionRounds: 0, PassScore: 80}
		provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, `{"content":"这不是一场梦而是一场雨——核心动机，Chapter 2。"}`, passedAuditJSON, settlementJSON, memoryJSON}}
		bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
		if e != nil {
			t.Fatal(e)
		}
		want := RunSucceeded
		if enabled {
			want = RunSucceededWithWarnings
		}
		if len(provider.requests) != 6 || bundle.FinalRunStatus != want {
			t.Fatal("switch or budget ignored", enabled, len(provider.requests), bundle.FinalRunStatus)
		}
	}
}

func TestNativePostWriteReviewRealMySQL(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("native-review")
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, `{"content":"这不是一场梦而是一场雨——核心动机，Chapter 2。"}`, passedAuditJSON, `{"strategy":"移除分析术语和章节引用","content":"雨水敲着窗。林澜拆开了那封没有寄件人的信。","resolvedIssueIDs":[]}`, passedAuditJSON, settlementJSON, memoryJSON}}
	s := NewService(NewMySQLStore(db), NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	p, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "原版后写审稿合并验证", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: p.ID, Mode: "long_original", Title: "审稿"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.UpdateProjectSettings(ctx, owner, p.ID, ProjectSettings{Agent: AgentSettings{Version: 1, EnablePostWrite: true, PauseAfterBlueprint: boolPointer(false), PassScore: 80, MaxRevisionRounds: 1}})
	if e != nil {
		t.Fatal(e)
	}
	r, e := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: p.ID, SessionID: session.ID, Content: "写第一章", ModelID: "fake-only", IdempotencyKey: "native-effect"})
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.Run(ctx, owner, r.ID)
	if e != nil || r.Status != RunSucceeded {
		t.Fatal(r.Status, e)
	}
	chapters, e := NewMySQLStore(db).ListChapters(ctx, owner, p.ID)
	if e != nil || len(chapters) != 1 || strings.Contains(chapters[0].Content, "核心动机") {
		t.Fatal("bad draft persisted", e)
	}
	var saved []byte
	if e = db.QueryRowContext(ctx, "SELECT snapshot_json FROM novel_audits WHERE owner_username=? AND run_id=? ORDER BY created_at,audit_id LIMIT 1", owner, r.ID).Scan(&saved); e != nil {
		t.Fatal(e)
	}
	var first chapterAuditOutput
	if e = json.Unmarshal(saved, &first); e != nil || first.Passed == nil || *first.Passed {
		t.Fatal("hard check did not block stored audit", e, string(saved))
	}
	if len(provider.requests) != 8 || !strings.Contains(string(saved), "post_write") {
		t.Fatal("missing real revision round", len(provider.requests))
	}
	t.Log("native-derived hard checks blocked LLM pass, entered revision/review, and saved corrected chapter with durable merged audit;0 paid calls")
}
