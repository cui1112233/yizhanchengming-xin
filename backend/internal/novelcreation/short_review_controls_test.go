package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func controlledShortFixture(t *testing.T, responses []string, a AgentSettings) (*Pipeline, *queuedPipelineProvider, RunContext, ShortStory) {
	t.Helper()
	provider := &queuedPipelineProvider{responses: responses}
	in := fixturePipelineRunContext(NewMemoryRuntime())
	in.Session.Mode = "short_original"
	in.ShortReviewVersion = 1
	in.Agent = &a
	story := ShortStory{Input: ShortStoryInput{Mode: "short_original", TargetWordCount: 1000, TargetChapters: 1}, Outline: &ShortOutline{Title: "验收", Summary: "真相揭晓", Chapters: []ShortOutlineChapter{{Number: 1, Title: "起点", Summary: "完成冲突"}}}}
	return pipelineWithProvider(provider), provider, in, story
}
func shortControlText(s string) string {
	raw, _ := json.Marshal(shortText{Content: s})
	return string(raw)
}
func TestShortReviewControlsRespectZeroBudgetDisabledLengthAndPassScore(t *testing.T) {
	for _, kind := range []string{"zero_budget", "length_disabled", "low_score"} {
		t.Run(kind, func(t *testing.T) {
			body := strings.Repeat("完整正文。", 200)
			review := `{"score":95,"issues":[]}`
			if kind == "zero_budget" {
				review = `{"score":50,"issues":["主线逻辑不成立"]}`
			}
			if kind == "length_disabled" {
				body = "短正文。"
			}
			if kind == "low_score" {
				review = `{"score":1,"issues":[]}`
			}
			a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 0, PassScore: 80}, 2000)
			a.EnableLengthNorm = false
			p, provider, in, story := controlledShortFixture(t, []string{shortControlText(body), review, shortControlText(strings.Repeat("完整正文。", 200)), shortControlText("推荐")}, a)
			out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
			if e == nil || len(provider.requests) != 2 || out.Status == "completed" {
				t.Fatalf("%s: err=%v calls=%d status=%s", kind, e, len(provider.requests), out.Status)
			}
		})
	}
}
func TestShortRevisionIsReauditedAndBounded(t *testing.T) {
	for _, pass := range []bool{true, false} {
		t.Run(map[bool]string{true: "passes", false: "still_fails"}[pass], func(t *testing.T) {
			body := shortControlText(strings.Repeat("完整正文。", 200))
			after := `{"score":20,"issues":["仍不成立"]}`
			if pass {
				after = `{"score":90,"issues":[]}`
			}
			a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 1, PassScore: 80}, 2000)
			p, provider, in, story := controlledShortFixture(t, []string{body, `{"score":10,"issues":[]}`, body, after, shortControlText("推荐")}, a)
			out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
			if pass {
				if e != nil || out.Status != "completed" || len(provider.requests) != 5 || len(out.ChapterAudits) != 2 {
					t.Fatalf("%+v %v calls=%d", out, e, len(provider.requests))
				}
			} else if e == nil || out.Status == "completed" || len(provider.requests) != 4 {
				t.Fatal("failed re-audit escaped", e, len(provider.requests))
			}
			if len(provider.requests) >= 4 && !strings.Contains(provider.requests[3].UserPrompt, "pass_score") {
				t.Fatal("re-audit missing frozen pass score")
			}
		})
	}
}
func TestShortLengthRepairReauditedAndCancellationStops(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 1, PassScore: 80}, 2000)
	a.EnableLengthNorm = true
	body := shortControlText(strings.Repeat("完整正文。", 200))
	p, provider, in, story := controlledShortFixture(t, []string{shortControlText("短正文。"), `{"score":95,"issues":[]}`, body, `{"score":95,"issues":[]}`, shortControlText("推荐")}, a)
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if e != nil || out.Status != "completed" || len(provider.requests) != 5 {
		t.Fatal(e, len(provider.requests), out.Status)
	}
	runtime := NewMemoryRuntime()
	p, provider, in, story = controlledShortFixture(t, []string{body, `{"score":1,"issues":[]}`, body, `{"score":95,"issues":[]}`}, a)
	in.Runtime = runtime
	provider.afterCall = func(n int) {
		if n == 1 {
			_ = runtime.Cancel(context.Background(), in.Run.ID)
		}
	}
	out, e = p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if e == nil || out.Status == "completed" || len(provider.requests) != 2 {
		t.Fatal("cancel dispatched more calls", e, len(provider.requests))
	}
}
func TestShortLegacySnapshotKeepsPreviousBehavior(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 0, PassScore: 80}, 2000)
	body := shortControlText(strings.Repeat("完整正文。", 200))
	p, provider, in, story := controlledShortFixture(t, []string{body, `{"score":1,"issues":[]}`, shortControlText("推荐")}, a)
	in.ShortReviewVersion = 0
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if e != nil || out.Status != "completed" || len(provider.requests) != 3 {
		t.Fatal("legacy protocol changed", e, len(provider.requests))
	}
}

func TestShortOutlineUsesScoreBudgetAndReaudit(t *testing.T) {
	for _, rounds := range []int{0, 1} {
		t.Run(map[int]string{0: "no_revision", 1: "reaudit"}[rounds], func(t *testing.T) {
			a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: rounds, PassScore: 80}, 2000)
			outline := `{"title":"验收","summary":"真相揭晓","chapters":[{"number":1,"title":"起点","summary":"完成冲突"}]}`
			p, provider, in, story := controlledShortFixture(t, []string{outline, `{"score":1,"issues":[]}`, outline, `{"score":95,"issues":[]}`}, a)
			story.Outline = nil
			story.Input.PauseAfterOutline = true
			out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false)
			if rounds == 0 {
				if e == nil || len(provider.requests) != 2 || out.Status == "awaiting_confirmation" {
					t.Fatal("low score outline bypassed", e, len(provider.requests))
				}
			} else if e != nil || len(provider.requests) != 4 || out.Status != "awaiting_confirmation" || !strings.Contains(string(out.OutlineAudit), "95") {
				t.Fatal("outline not reaudited", e, len(provider.requests), out.Status)
			}
		})
	}
}

func TestShortPiecesShareChapterRevisionBudget(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 1, PassScore: 80}, 2000)
	body := shortControlText(strings.Repeat("甲", 4000))
	p, provider, in, story := controlledShortFixture(t, []string{body, `{"score":1,"issues":[]}`, body, `{"score":95,"issues":[]}`, shortControlText("累积记忆"), body, `{"score":1,"issues":[]}`}, a)
	story.Input.TargetWordCount = 8000
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if e == nil || len(provider.requests) != 7 || out.Status == "completed" {
		t.Fatal("chunking multiplied revision budget", e, len(provider.requests))
	}
}

func TestShortControlledReviewRequiresScoreButAllowsExplicitZero(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 0, PassScore: 0}, 2000)
	body := shortControlText(strings.Repeat("完整正文。", 200))
	for _, valid := range []bool{false, true} {
		replies := []string{body, `{"issues":[]}`, `{"issues":[]}`}
		if valid {
			replies = []string{body, `{"score":0,"issues":[]}`, shortControlText("推荐")}
		}
		p, provider, in, story := controlledShortFixture(t, replies, a)
		out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
		if valid {
			if e != nil || out.Status != "completed" {
				t.Fatal("explicit zero rejected", e)
			}
		} else if e == nil || out.Status == "completed" || len(provider.requests) != 3 {
			t.Fatal("missing score became zero", e, len(provider.requests))
		}
	}
}

func TestShortReviewRepairCannotReuseMissingFields(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 0, PassScore: 0}, 2000)
	body := shortControlText(strings.Repeat("完整正文。", 200))
	p, provider, in, story := controlledShortFixture(t, []string{body, `{"score":0}`, `{"issues":[]}`}, a)
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if e == nil || out.Status == "completed" || len(provider.requests) != 3 {
		t.Fatal("repair reused stale score", e, len(provider.requests))
	}
}

func TestShortControlledRealMySQLReauditSaveAndRestart(t *testing.T) {
	db := studioOwnedDB(t)
	owner := novelID("short-control-mysql")
	body := shortControlText(strings.Repeat("完整正文。", 200))
	provider := &queuedPipelineProvider{responses: []string{`{"content":"写作参考分析"}`, `{"title":"验收","summary":"真相揭晓","chapters":[{"number":1,"title":"起点","summary":"完成冲突"}]}`, `{"score":95,"issues":[]}`, body, `{"score":1,"issues":[]}`, body, `{"score":95,"issues":[]}`, shortControlText("推荐")}}
	s := NewService(NewMySQLStore(db), NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput(novelID("create"))
	input.TargetWordCount = 1000
	input.PauseAfterOutline = false
	created, e := s.StartShortStory(context.Background(), owner, input)
	if e != nil {
		t.Fatal(e)
	}
	var frozen struct {
		ShortReviewVersion int `json:"short_review_version"`
	}
	if e = json.Unmarshal(created.Run.PromptSnapshot, &frozen); e != nil || frozen.ShortReviewVersion != 1 {
		t.Fatal("new task missing version", e)
	}
	run, e := s.Run(context.Background(), owner, created.Run.ID)
	if e != nil || run.Status != RunSucceeded || len(provider.requests) != 8 {
		t.Fatal(run.Status, e, len(provider.requests))
	}
	reopened := NewService(NewMySQLStore(db), NewMemoryRuntime(), pipelineWithProvider(provider))
	got, e := reopened.GetShortStory(context.Background(), owner, created.Project.ID)
	if e != nil || got.Status != "completed" || len(got.ChapterAudits) != 2 || len(got.Chapters) != 1 {
		t.Fatal("durable result missing", e, got.Status, len(got.ChapterAudits))
	}
	if _, e = reopened.Run(context.Background(), owner, created.Run.ID); e != nil || len(provider.requests) != 8 {
		t.Fatal("reopen repeated paid calls", e, len(provider.requests))
	}
	// A second real database task exhausts all three default revisions. Its
	// failed audit must not become a completed/exportable short story.
	provider.responses = append(provider.responses, `{"content":"参考分析"}`, `{"title":"未通过作品","summary":"真相揭晓","chapters":[{"number":1,"title":"起点","summary":"完成冲突"}]}`, `{"score":95,"issues":[]}`, body, `{"score":1,"issues":[]}`, body, `{"score":1,"issues":[]}`, body, `{"score":1,"issues":[]}`, body, `{"score":1,"issues":[]}`)
	input.IdempotencyKey = novelID("failed-create")
	bad, e := s.StartShortStory(context.Background(), owner, input)
	if e != nil {
		t.Fatal(e)
	}
	failed, e := s.Run(context.Background(), owner, bad.Run.ID)
	if e == nil || failed.Status != RunFailed || len(provider.requests) != 19 {
		t.Fatal("bad audit marked successful", failed.Status, e, len(provider.requests))
	}
	stored, e := reopened.GetShortStory(context.Background(), owner, bad.Project.ID)
	if e != nil || stored.Status != "failed" || len(stored.Artifacts) != 0 || !strings.Contains(failed.ErrorMessage, "审核未通过") {
		t.Fatal("failure not visible after restart", e, stored.Status, failed.ErrorMessage)
	}
}
