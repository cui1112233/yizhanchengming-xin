package novelcreation

import (
	"context"
	"strings"
	"testing"
)

func hookFixture(thread, status string, start, advanced, count int) settledForeshadowing {
	return settledForeshadowing{Thread: thread, Status: status, StartChapter: start, LastAdvancedChapter: advanced, AdvancingCount: count}
}

func TestAnalyzeHookHealthStaleDroughtAndBurst(t *testing.T) {
	hooks := []settledForeshadowing{
		hookFixture("雨夜来信", "open", 1, 1, 1),
		hookFixture("褪色邮戳", "open", 2, 2, 1),
		hookFixture("门缝纸条", "open", 3, 3, 1),
		hookFixture("旧车票", "resolved", 1, 2, 2),
	}
	report := analyzeHookHealth(hooks, 20)
	if report.ActiveCount != 3 || report.ResolvedCount != 1 {
		t.Fatalf("counts=%+v", report)
	}
	if report.StaleCount != 3 {
		t.Fatalf("stale=%+v", report.StaleHooks)
	}
	levels := map[string]string{}
	for _, alert := range report.Alerts {
		levels[alert.Level] = alert.Message
	}
	if !strings.Contains(levels["warning"], "沉寂") && !strings.Contains(levels["error"], "旱灾") {
		t.Fatalf("missing stale/drought alert: %+v", report.Alerts)
	}
	if !strings.Contains(levels["error"], "未推进任何伏笔") {
		t.Fatalf("missing drought error: %+v", report.Alerts)
	}
}

func TestAnalyzeHookHealthBurstWithoutResolve(t *testing.T) {
	hooks := []settledForeshadowing{
		hookFixture("甲", "open", 10, 10, 1),
		hookFixture("乙", "open", 11, 11, 1),
		hookFixture("丙", "open", 11, 11, 1),
	}
	report := analyzeHookHealth(hooks, 11)
	found := false
	for _, alert := range report.Alerts {
		if strings.Contains(alert.Message, "爆发警告") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing burst alert: %+v", report.Alerts)
	}
}

func TestAnalyzeHookHealthUnknownStartIsNotStale(t *testing.T) {
	report := analyzeHookHealth([]settledForeshadowing{{Thread: "历史伏笔", Status: "open"}}, 30)
	if report.StaleCount != 0 {
		t.Fatalf("unknown-start hook must not be reported stale: %+v", report.StaleHooks)
	}
	if report.UnknownStart != 1 {
		t.Fatalf("unknown start must be counted: %+v", report)
	}
}

func TestAnalyzeHookHealthDecayUsesHalfLife(t *testing.T) {
	fast := settledForeshadowing{Thread: "急线", Status: "open", StartChapter: 1, LastAdvancedChapter: 1, PayoffTiming: "immediate"}
	slow := settledForeshadowing{Thread: "慢线", Status: "open", StartChapter: 1, LastAdvancedChapter: 1, PayoffTiming: "slow-burn"}
	if hookDecay(fast, 9) <= hookDecay(slow, 9) {
		t.Fatalf("immediate payoff must decay faster: fast=%.3f slow=%.3f", hookDecay(fast, 9), hookDecay(slow, 9))
	}
}

func TestMergeForeshadowingBookkeepingTracksRealChapters(t *testing.T) {
	old := []settledForeshadowing{{Thread: "雨夜来信", Status: "open", StartChapter: 2, LastAdvancedChapter: 2, AdvancingCount: 1}}
	delta := []settledForeshadowing{
		{Thread: "雨夜来信", Status: "open"},
		{Thread: "门缝纸条", Status: "open"},
	}
	merged := mergeForeshadowingBookkeeping(old, delta, 5)
	byThread := map[string]settledForeshadowing{}
	for _, hook := range merged {
		byThread[hook.Thread] = hook
	}
	if got := byThread["雨夜来信"]; got.StartChapter != 2 || got.LastAdvancedChapter != 5 || got.AdvancingCount != 2 {
		t.Fatalf("existing hook bookkeeping=%+v", got)
	}
	// 新伏笔只记录埋下章号；lastAdvanced 留空，与洛水 Hook 默认值一致。
	if got := byThread["门缝纸条"]; got.StartChapter != 5 || got.LastAdvancedChapter != 0 || got.AdvancingCount != 0 {
		t.Fatalf("new hook bookkeeping=%+v", got)
	}
}

func TestMergeForeshadowingBookkeepingKeepsResolvedClosed(t *testing.T) {
	old := []settledForeshadowing{{Thread: "旧车票", Status: "resolved", StartChapter: 1, LastAdvancedChapter: 3, AdvancingCount: 2}}
	merged := mergeForeshadowingBookkeeping(old, []settledForeshadowing{{Thread: "旧车票", Status: "open"}}, 9)
	if merged[0].Status != "resolved" || merged[0].LastAdvancedChapter != 3 {
		t.Fatalf("resolved hook reopened: %+v", merged[0])
	}
}

func TestPlannerHookHealthInjectionIsVersionGated(t *testing.T) {
	truth := []byte(`{"facts":[],"characterChanges":[],"timeline":[],"foreshadowing":[{"thread":"雨夜来信","status":"open","startChapter":1,"lastAdvancedChapter":1}]}`)
	for _, tc := range []struct {
		version int
		want    bool
	}{{0, false}, {1, true}} {
		input := fixturePipelineRunContext(NewMemoryRuntime())
		a := NormalizeAgentSettings(AgentSettings{}, 2000)
		a.EnableLengthNorm = false
		input.Agent = &a
		input.HookHealthVersion = tc.version
		input.PreviousTruth = truth
		provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
		if _, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); e != nil {
			t.Fatal(e)
		}
		injected := false
		for _, request := range provider.requests {
			if strings.Contains(request.UserPrompt, "伏笔健康") {
				injected = true
			}
		}
		if injected != tc.want {
			t.Fatalf("version=%d injected=%v want=%v", tc.version, injected, tc.want)
		}
	}
}
