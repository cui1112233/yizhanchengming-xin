package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestNativeRoutedFailureRealMySQL(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("route-reject")
	localAudit := strings.Replace(failedAuditJSON, `"category":"plot"`, `"category":"plot","repairScope":"local"`, 1)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, localAudit, revisedJSON}}
	store := NewMySQLStore(db)
	s := NewService(store, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	project, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "错误修订方式不得保存", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: project.ID, Mode: "long_original", Title: "修订方式"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.UpdateProjectSettings(ctx, owner, project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, PauseAfterBlueprint: boolPointer(false), PassScore: 80, MaxRevisionRounds: 1}})
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "写第一章", ModelID: "fake-only", IdempotencyKey: "reject-wrong-route"})
	if e != nil {
		t.Fatal(e)
	}
	failed, e := s.Run(ctx, owner, run.ID)
	if !IsCode(e, ErrInvalidOutput) || failed.Status != RunFailed || !strings.Contains(failed.ErrorMessage, "局部修改") {
		t.Fatalf("wrong route not clearly failed: %+v %v", failed, e)
	}
	chapters, e := store.ListChapters(ctx, owner, project.ID)
	if e != nil || len(chapters) != 0 {
		t.Fatal("wrong-mode text persisted as chapter", e, len(chapters))
	}
	var count int
	if e = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM novel_revisions WHERE owner_username=? AND run_id=?", owner, run.ID).Scan(&count); e != nil || count != 0 {
		t.Fatal("wrong-mode revision persisted", e, count)
	}
	if _, e = s.Run(ctx, owner, run.ID); e != nil || len(provider.requests) != 5 {
		t.Fatal("failed run silently retried", e, len(provider.requests))
	}
	t.Log("local scope required PATCHES; full JSON reply rejected; durable run failed with clear reason, 0 chapters/revisions, repeat generated no new requests; 0 paid calls")
}

func TestNativeRoutedRevisionPipeline(t *testing.T) {
	whole := "=== FIXED_ISSUES ===\n修复主线。\n=== REVISED_CONTENT ===\n雨水敲着窗。林澜找到失踪者的指纹。"
	cases := []struct{ name, scope, reply, wantMode, errorText string }{
		{"local", "local", nativeFixturePatch, "patch-only", ""},
		{"structural", "structural", whole, "rewrite-only", ""},
		{"wrong whole", "local", whole, "patch-only", "局部修改"},
		{"wrong patches", "structural", nativeFixturePatch, "rewrite-only", "整章重写"},
		{"JSON cannot fake local", "local", revisedJSON, "patch-only", "局部修改"},
		{"unknown legacy compatible", "unknown", revisedJSON, "allow-full", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			input := fixturePipelineRunContext(NewMemoryRuntime())
			input.ReviewCycleVersion = 3
			var audit map[string]any
			_ = json.Unmarshal([]byte(failedAuditJSON), &audit)
			audit["issues"].([]any)[0].(map[string]any)["repair_scope"] = c.scope
			auditRaw, _ := json.Marshal(audit)
			provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, string(auditRaw), c.reply, passedAuditJSON, settlementJSON, memoryJSON}}
			bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
			if c.errorText != "" {
				if err == nil || !strings.Contains(err.Error(), c.errorText) || len(provider.requests) != 5 || bundle.Chapter.Content != "" {
					t.Fatalf("wrong route accepted: %v calls=%d", err, len(provider.requests))
				}
			} else {
				if err != nil || len(provider.requests) != 8 || len(bundle.Revisions) != 1 || !strings.Contains(bundle.Chapter.Content, "指纹") {
					t.Fatalf("route did not produce result: %v calls=%d", err, len(provider.requests))
				}
				var snapshot struct {
					Native nativePatchResult `json:"native_revision"`
				}
				if err = json.Unmarshal(bundle.Revisions[0].Snapshot, &snapshot); err != nil || snapshot.Native.Route == nil || snapshot.Native.Route.Mode != c.wantMode {
					t.Fatalf("route was not saved: %v %s", err, bundle.Revisions[0].Snapshot)
				}
			}
			var request map[string]any
			_ = json.Unmarshal([]byte(provider.requests[4].UserPrompt), &request)
			if request["revision_output_mode"] != c.wantMode {
				t.Fatal("route not supplied to own model", request)
			}
			system := provider.requests[4].SystemPrompt
			hasLocal := strings.Contains(system, "你必须只输出 PATCHES——不要整章改写")
			hasStructural := strings.Contains(system, "你必须输出 REVISED_CONTENT——禁止输出 PATCHES")
			if hasLocal != (c.wantMode == "patch-only") || hasStructural != (c.wantMode == "rewrite-only") {
				t.Fatal("missing or contradictory native instructions")
			}
		})
	}
}

func routeIssue(severity, category, description, camel, snake string) auditIssue {
	var v auditIssue
	b, _ := json.Marshal(map[string]any{"id": "route", "severity": severity, "category": category, "description": description, "repairScope": camel, "repair_scope": snake})
	_ = json.Unmarshal(b, &v)
	return v
}

func TestNativeRevisionRouting(t *testing.T) {
	cases := []struct {
		name   string
		issues []auditIssue
		want   string
	}{
		{"empty", nil, "allow-full"},
		{"only info", []auditIssue{routeIssue("info", "人设", "", "", "")}, "patch-only"},
		{"explicit structural beats local text", []auditIssue{routeIssue("high", "段落等长", "", "structural", "")}, "rewrite-only"},
		{"explicit local beats structural text", []auditIssue{routeIssue("high", "人设", "", "local", "")}, "patch-only"},
		{"snake field retained", []auditIssue{routeIssue("high", "未知", "", "", "local")}, "patch-only"},
		{"camel wins", []auditIssue{routeIssue("high", "未知", "", "local", "structural")}, "patch-only"},
		{"mixed explicit", []auditIssue{routeIssue("high", "未知", "", "local", ""), routeIssue("high", "未知", "", "structural", "")}, "rewrite-only"},
		{"partial scope falls back", []auditIssue{routeIssue("high", "段落等长", "", "local", ""), routeIssue("high", "陌生问题", "", "", "")}, "allow-full"},
		{"patterns local", []auditIssue{routeIssue("high", "HEDGE DENSITY", "", "", "")}, "patch-only"},
		{"patterns structural", []auditIssue{routeIssue("high", "普通分类", "存在主线偏离", "", "")}, "rewrite-only"},
		{"structural pattern wins", []auditIssue{routeIssue("high", "段落等长", "人设不一致", "", "")}, "rewrite-only"},
		{"unknown not guessed", []auditIssue{routeIssue("high", "未归类", "未知描述", "", "")}, "allow-full"},
		{"low is not info", []auditIssue{routeIssue("low", "时间线", "", "", "")}, "rewrite-only"},
		{"scope case exact", []auditIssue{routeIssue("high", "未知", "", "LOCAL", "")}, "allow-full"},
		{"info camel excluded", []auditIssue{routeIssue("info", "人设", "", "structural", "")}, "patch-only"},
		{"native snake precedence quirk", []auditIssue{routeIssue("info", "未知", "", "", "structural")}, "rewrite-only"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := resolveNativeRevisionRoute(c.issues)
			if got.Mode != c.want {
				t.Fatalf("got %+v want %s", got, c.want)
			}
		})
	}
}

func TestNativeVersion2DoesNotMigrateRoute(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ReviewCycleVersion = 2
	audit := strings.Replace(failedAuditJSON, `"category":"plot"`, `"category":"plot","repairScope":"local"`, 1)
	whole := "=== FIXED_ISSUES ===\n补充线索\n=== REVISED_CONTENT ===\n雨水敲着窗。林澜找到失踪者的指纹。"
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, audit, whole, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil || len(bundle.Revisions) != 1 || len(provider.requests) != 8 {
		t.Fatalf("old frozen protocol migrated: %v", e)
	}
	if strings.Contains(provider.requests[4].UserPrompt, "revision_output_mode") || strings.Contains(provider.requests[4].UserPrompt, "repairScope") || strings.Contains(provider.requests[4].SystemPrompt, "你必须只输出 PATCHES——不要整章改写") {
		t.Fatal("version2 request fingerprint changed")
	}
}
