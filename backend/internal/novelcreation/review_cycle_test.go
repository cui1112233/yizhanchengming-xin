package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func reviewWordBody(count int, marker string) string {
	head := marker
	if len([]rune(head)) >= count {
		return string([]rune(head)[:count])
	}
	return head + strings.Repeat("字", count-len([]rune(head)))
}

func TestReviewLengthInRangeUsesDocumentedRatios(t *testing.T) {
	for _, tc := range []struct {
		words  int
		target int
		want   bool
	}{
		{1099, 2000, false},
		{1100, 2000, true},
		{2000, 2000, true},
		{3200, 2000, true},
		{3201, 2000, false},
		{0, 0, true},
	} {
		if got := reviewLengthInRange(tc.words, tc.target); got != tc.want {
			t.Fatalf("words=%d target=%d got=%v want=%v", tc.words, tc.target, got, tc.want)
		}
	}
}

func TestBetterReviewSnapshotPrefersLengthAtEqualScore(t *testing.T) {
	best := reviewSnapshot{Score: 85, LengthInRange: false, Passed: false}
	candidate := reviewSnapshot{Score: 85, LengthInRange: true, Passed: false}
	if !betterReviewSnapshot(candidate, best) {
		t.Fatal("in-range candidate must win at equal score")
	}
	if betterReviewSnapshot(best, candidate) {
		t.Fatal("out-of-range candidate must not win")
	}
}

func TestBetterReviewSnapshotRejectsHighScoreOutOfRange(t *testing.T) {
	best := reviewSnapshot{Score: 82, LengthInRange: true, Passed: false}
	candidate := reviewSnapshot{Score: 95, LengthInRange: false, Passed: false}
	if betterReviewSnapshot(candidate, best) {
		t.Fatal("severely off-length draft must not replace an in-range best")
	}
	passing := reviewSnapshot{Score: 81, LengthInRange: false, Passed: true}
	if !betterReviewSnapshot(passing, best) {
		t.Fatal("passing draft must outrank a non-passing draft")
	}
}

func TestBetterReviewSnapshotRequiresNetImprovement(t *testing.T) {
	best := reviewSnapshot{Score: 80, LengthInRange: true}
	if betterReviewSnapshot(reviewSnapshot{Score: 80.2, LengthInRange: true}, best) {
		t.Fatal("improvement below epsilon must not win")
	}
	if !betterReviewSnapshot(reviewSnapshot{Score: 81, LengthInRange: true}, best) {
		t.Fatal("improvement above epsilon must win")
	}
}

func reviewLoopFixture(t *testing.T) (RunContext, *chapterPlanOutput) {
	t.Helper()
	input := fixturePipelineRunContext(NewMemoryRuntime())
	agent := NormalizeAgentSettings(AgentSettings{}, 2000)
	agent.EnableLengthNorm = false
	agent.EnablePostWrite = false
	agent.EnableAiTells = false
	agent.MaxRevisionRounds = 3
	// 阈值设高，避免 88 分候选被自动判定通过，从而能验证回退与择优。
	agent.PassScore = 95
	input.Agent = &agent
	input.ReviewLoopVersion = 1
	plan := chapterPlanOutput{ChapterNumber: 1, Title: "雨夜来信", TargetWords: 2000}
	return input, &plan
}

func reviewAudit(score float64, passed bool, issue string) string {
	issues := "[]"
	if issue != "" {
		issues = fmt.Sprintf(`[{"id":%q,"severity":"medium","description":"需要修订"}]`, issue)
	}
	return fmt.Sprintf(`{"score":%v,"issues":%s,"passed":%v}`, score, issues, passed)
}

func reviewRevision(content, issue string) string {
	return fmt.Sprintf(`{"strategy":"补足线索","content":%q,"resolvedIssueIDs":[%q]}`, content, issue)
}

func TestStructuredReviewLoopExitsOnNoNetImprovement(t *testing.T) {
	input, _ := reviewLoopFixture(t)
	initial := reviewWordBody(1200, "初始正文")
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, fmt.Sprintf(`{"content":%q}`, initial),
		reviewAudit(70, false, "i1"),
		reviewRevision(reviewWordBody(1200, "第一次修订"), "i1"),
		reviewAudit(70.2, false, "i1"),
		settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	report := reviewReportOf(t, bundle.RuntimeSnapshot)
	if report.ExitReason != "no_net_improvement" || report.Iterations != 1 {
		t.Fatalf("report=%+v", report)
	}
	if report.Revised {
		t.Fatalf("sub-epsilon revision must not be adopted: %+v", report)
	}
	if !strings.Contains(bundle.Chapter.Content, "初始正文") {
		t.Fatalf("initial content lost: %s", bundle.Chapter.Content[:20])
	}
}

func TestStructuredReviewLoopRollsBackToBestVersion(t *testing.T) {
	input, _ := reviewLoopFixture(t)
	initial := reviewWordBody(1200, "初始正文")
	first := reviewWordBody(1200, "高分修订")
	second := reviewWordBody(1200, "低分修订")
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, fmt.Sprintf(`{"content":%q}`, initial),
		reviewAudit(70, false, "i1"),
		reviewRevision(first, "i1"),
		reviewAudit(88, false, "i1"),
		reviewRevision(second, "i1"),
		reviewAudit(72, false, "i1"),
		settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	report := reviewReportOf(t, bundle.RuntimeSnapshot)
	if !report.RolledBack || !report.Revised {
		t.Fatalf("rollback flags missing: %+v", report)
	}
	if report.ExitReason != "no_net_improvement" || report.Iterations != 2 {
		t.Fatalf("exit=%s iterations=%d", report.ExitReason, report.Iterations)
	}
	if !strings.Contains(bundle.Chapter.Content, "高分修订") {
		t.Fatalf("best version not adopted: %s", bundle.Chapter.Content[:20])
	}
}

func TestStructuredReviewLoopStopsOnEmptyRevision(t *testing.T) {
	input, _ := reviewLoopFixture(t)
	initial := reviewWordBody(1200, "初始正文")
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, fmt.Sprintf(`{"content":%q}`, initial),
		reviewAudit(70, false, "i1"),
		reviewRevision(initial, "i1"),
		settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	report := reviewReportOf(t, bundle.RuntimeSnapshot)
	if report.ExitReason != "empty_revision" || report.Revised {
		t.Fatalf("report=%+v", report)
	}
}

func TestStructuredReviewLoopPassesWithoutRevision(t *testing.T) {
	input, _ := reviewLoopFixture(t)
	initial := reviewWordBody(1200, "初始正文")
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, fmt.Sprintf(`{"content":%q}`, initial),
		reviewAudit(96, true, ""),
		settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	report := reviewReportOf(t, bundle.RuntimeSnapshot)
	if report.ExitReason != "passed" || report.Iterations != 0 || bundle.FinalRunStatus != RunSucceeded {
		t.Fatalf("report=%+v status=%s", report, bundle.FinalRunStatus)
	}
}

func TestReviewLoopVersionGatesBehaviour(t *testing.T) {
	for _, version := range []int{0, 1} {
		input, _ := reviewLoopFixture(t)
		input.ReviewLoopVersion = version
		// 两条路径都只跑一轮，避免旧路径继续修订而消耗后续响应。
		input.Agent.MaxRevisionRounds = 1
		initial := reviewWordBody(1200, "初始正文")
		provider := &queuedPipelineProvider{responses: []string{
			foundationJSON, chapterPlanJSON, fmt.Sprintf(`{"content":%q}`, initial),
			reviewAudit(70, false, "i1"),
			reviewRevision(reviewWordBody(1200, "第一次修订"), "i1"),
			reviewAudit(70.2, false, "i1"),
			settlementJSON, memoryJSON,
		}}
		bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
		if err != nil {
			t.Fatal(err)
		}
		var snapshot map[string]json.RawMessage
		if err = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); err != nil {
			t.Fatal(err)
		}
		_, present := snapshot["reviewCycle"]
		if version == 1 && !present {
			t.Fatal("version 1 must record the review cycle report")
		}
		if version == 0 && present {
			t.Fatal("version 0 must keep the legacy loop without a report")
		}
		if version == 0 && strings.Contains(bundle.Chapter.Content, "初始正文") {
			t.Fatalf("legacy loop keeps highest score, not epsilon-gated best: %s", bundle.Chapter.Content[:20])
		}
	}
}

func reviewReportOf(t *testing.T, runtime json.RawMessage) reviewCycleReport {
	t.Helper()
	var snapshot map[string]json.RawMessage
	if err := json.Unmarshal(runtime, &snapshot); err != nil {
		t.Fatal(err)
	}
	var report reviewCycleReport
	if err := json.Unmarshal(snapshot["reviewCycle"], &report); err != nil || len(report.Snapshots) == 0 {
		t.Fatalf("reviewCycle missing: %v %s", err, snapshot["reviewCycle"])
	}
	return report
}
