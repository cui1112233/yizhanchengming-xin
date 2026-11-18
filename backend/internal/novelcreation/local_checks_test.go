package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPostWriteReportsWithoutRewriting(t *testing.T) {
	text := "# 第一章\n这不是一场梦而是一场雨——核心动机，Chapter 2。"
	r := ValidatePostWrite(text, PostWriteInput{})
	if r.Passed || r.Stats.ErrorCount != 4 {
		t.Fatalf("report=%+v", r)
	}
	want := []string{"pw_not_but", "pw_dash", "pw_analysis_terms", "pw_chapter_ref"}
	for i, id := range want {
		if r.Issues[i].RuleID != id {
			t.Fatalf("order=%+v", r.Issues)
		}
	}
	if r.Stats.TotalChars != len([]rune(strings.ReplaceAll(strings.SplitN(text, "\n", 2)[1], " ", ""))) {
		t.Fatal("wrong codepoint counter")
	}
}

func TestLengthNormalizationBoundary(t *testing.T) {
	for _, v := range []struct {
		count int
		want  bool
	}{{1099, true}, {1100, false}, {2000, false}, {2900, false}, {2901, true}} {
		if got := needsLengthNormalization(strings.Repeat("字", v.count), 2000); got != v.want {
			t.Fatalf("count=%d got=%v", v.count, got)
		}
	}
	if needsLengthNormalization("正文", 0) {
		t.Fatal("invalid target triggers generation")
	}
}

func TestLocalChecksDoNotCallModelsOrChangeBody(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	a.EnableLengthNorm = false
	input.Agent = &a
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	sink := &recordingEventSink{}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, sink)
	if e != nil {
		t.Fatal(e)
	}
	if len(provider.requests) != 6 || bundle.Chapter.Content != "雨水敲着窗。林澜拆开了那封没有寄件人的信。" {
		t.Fatalf("calls=%d body=%s", len(provider.requests), bundle.Chapter.Content)
	}
	checks := 0
	for _, event := range sink.persisted {
		if event.Type == "run.local_checks" {
			checks++
			if len(event.Payload) == 0 {
				t.Fatal("empty report")
			}
		}
	}
	if checks != 1 {
		t.Fatalf("durable reports=%d", checks)
	}
	var saved map[string]json.RawMessage
	if json.Unmarshal(bundle.RuntimeSnapshot, &saved) != nil || len(saved["localChecks"]) == 0 {
		t.Fatal("checks lost from durable run snapshot")
	}
}

func TestPostWriteEveryRecoveredRule(t *testing.T) {
	cases := []struct {
		id, body string
		in       PostWriteInput
	}{
		{"pw_not_but", "不是两个人而是三个人", PostWriteInput{}},
		{"pw_dash", "雨落——灯亮", PostWriteInput{}},
		{"pw_analysis_terms", "核心动机", PostWriteInput{}},
		{"pw_chapter_ref", "cHaPtEr ２", PostWriteInput{}},
		{"pw_dup_title", "正文", PostWriteInput{Title: "春，雨", ExistingTitles: []string{"春雨"}}},
		{"pw_surprise_density", strings.Repeat("字", 500) + "忽然忽然", PostWriteInput{}},
		{"pw_fatigue", "雨雨", PostWriteInput{FatigueWords: []string{"雨"}}},
		{"pw_meta_narration", "故事发展到了雨夜", PostWriteInput{}},
		{"pw_preaching", "显然这不言而喻", PostWriteInput{}},
		{"pw_collective_shock", "众人震惊。全场哗然。", PostWriteInput{}},
		{"pw_consecutive_le", strings.Repeat("他走了。", 6), PostWriteInput{}},
		{"pw_para_overflow", strings.Repeat("字", 301), PostWriteInput{}},
		{"pw_cross_repeat", strings.Repeat("甲乙丙丁戊己庚辛", 3), PostWriteInput{PreviousChapters: []string{"甲乙丙丁戊己庚辛"}}},
		{"pw_dialogue_pressure", strings.Repeat("叙述。\n", 16), PostWriteInput{}},
		{"pw_person_drift", strings.Repeat("他觉得林澜", 6) + strings.Repeat("字", 800), PostWriteInput{FirstPerson: true, Protagonist: "林澜"}},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			r := ValidatePostWrite(tc.body, tc.in)
			for _, i := range r.Issues {
				if i.RuleID == tc.id {
					if strings.Contains(i.Message, "%!") {
						t.Fatalf("broken native message formatting: %s", i.Message)
					}
					return
				}
			}
			t.Fatalf("missing %s: %+v", tc.id, r)
		})
	}
}

func TestPostWriteThresholdsAndUnicode(t *testing.T) {
	has := func(r PostWriteResult, id string) bool {
		for _, v := range r.Issues {
			if v.RuleID == id {
				return true
			}
		}
		return false
	}
	if has(ValidatePostWrite(strings.Repeat("字", 300), PostWriteInput{}), "pw_para_overflow") {
		t.Fatal("300 should not overflow")
	}
	if has(ValidatePostWrite(strings.Repeat("叙述\n", 15), PostWriteInput{}), "pw_dialogue_pressure") {
		t.Fatal("15 paragraphs should not trigger")
	}
	if has(ValidatePostWrite(strings.Repeat("字", 500)+"忽然", PostWriteInput{}), "pw_surprise_density") {
		t.Fatal("allowed marker count flagged")
	}
	r := ValidatePostWrite("# 标题\n \r\n😀\t　", PostWriteInput{})
	if r.Stats.TotalChars != 3 || r.Stats.ParagraphCount != 1 {
		t.Fatalf("codepoints/whitespace=%+v", r.Stats)
	}
	if has(ValidatePostWrite("显然显然显然", PostWriteInput{}), "pw_preaching") {
		t.Fatal("same word is not multiple word-list entries")
	}
	if has(ValidatePostWrite(strings.Repeat("他觉得林澜", 6)+strings.Repeat("字", 800), PostWriteInput{}), "pw_person_drift") {
		t.Fatal("person check lacks explicit first-person input")
	}
}
