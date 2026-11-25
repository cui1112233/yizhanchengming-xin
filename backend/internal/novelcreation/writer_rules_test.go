package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriterRuleBlocksByGenreAndChapter(t *testing.T) {
	blocks := BuildWriterRuleBlocks(WriterRuleInput{Genre: "悬疑", ChapterNumber: 1})
	joined := strings.Join(blocks, "\n")
	for _, want := range []string{"输入治理契约", "题材写作要求", "悬疑", "写作铁律", "看点密度与断章", "去 AI 味铁律", "文笔执行", "开篇三章"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in writer blocks", want)
		}
	}
	later := strings.Join(BuildWriterRuleBlocks(WriterRuleInput{Genre: "悬疑", ChapterNumber: 12}), "\n")
	if strings.Contains(later, "开篇三章") {
		t.Fatal("opening-chapter rules leaked past chapter 3")
	}
	unknown := strings.Join(BuildWriterRuleBlocks(WriterRuleInput{Genre: "未知题材", ChapterNumber: 5}), "\n")
	if strings.Contains(unknown, "题材写作要求") {
		t.Fatal("unknown genre must not inject a body")
	}
}

func TestWriterRuleBlocksAddHookDebtOnlyWhenPresent(t *testing.T) {
	withHooks := strings.Join(BuildWriterRuleBlocks(WriterRuleInput{ChapterNumber: 5, HasHookDebt: true}), "\n")
	if !strings.Contains(withHooks, "伏笔兑现要求") {
		t.Fatal("hook debt block missing")
	}
	without := strings.Join(BuildWriterRuleBlocks(WriterRuleInput{ChapterNumber: 5}), "\n")
	if strings.Contains(without, "伏笔兑现要求") {
		t.Fatal("hook debt block injected without debt")
	}
}

func TestNormalizeDraftTitle(t *testing.T) {
	cases := map[string]string{
		"第3章 雨夜来信":             "雨夜来信",
		"第十二章：门缝":              "门缝",
		"Chapter 7 The Letter": "The Letter",
		"《春雨》":                 "春雨",
		"  雨停  ":               "雨停",
		"第三章":                  "第三章",
	}
	for input, want := range cases {
		if got := normalizeDraftTitle(input); got != want {
			t.Fatalf("normalize(%q)=%q want=%q", input, got, want)
		}
	}
}

func TestAdoptDraftTitleRejectsDuplicates(t *testing.T) {
	if got := adoptDraftTitle("春雨", "旧标题", []string{"春，雨"}); got != "" {
		t.Fatalf("duplicate title adopted: %q", got)
	}
	if got := adoptDraftTitle("新标题", "旧标题", []string{"春，雨"}); got != "新标题" {
		t.Fatalf("unique title rejected: %q", got)
	}
	if got := adoptDraftTitle("旧标题", "旧标题", nil); got != "" {
		t.Fatalf("identical title adopted: %q", got)
	}
	if got := adoptDraftTitle(strings.Repeat("字", 61), "旧标题", nil); got != "" {
		t.Fatalf("overlong title adopted: %q", got)
	}
}

func TestValidatePreWriteCheckRejectsIncompleteRows(t *testing.T) {
	if err := validatePreWriteCheck(nil); err != nil {
		t.Fatalf("empty rows must be allowed: %v", err)
	}
	if err := validatePreWriteCheck([]preWriteCheckRow{{Item: "当前任务"}}); err == nil {
		t.Fatal("missing record must fail")
	}
	if err := validatePreWriteCheck([]preWriteCheckRow{{Item: "甲", Record: "A"}, {Item: "甲", Record: "B"}}); err == nil {
		t.Fatal("duplicate item must fail")
	}
	rows := []preWriteCheckRow{{Item: " 当前任务 ", Record: " 写封信 ", Note: " x "}}
	if err := validatePreWriteCheck(rows); err != nil || rows[0].Item != "当前任务" || rows[0].Record != "写封信" || rows[0].Note != "x" {
		t.Fatalf("rows not normalized: %+v %v", rows, err)
	}
}

func TestWriterProtocolVersionGatesRuleBlocksAndSnapshot(t *testing.T) {
	for _, tc := range []struct {
		version int
		want    bool
	}{{0, false}, {1, true}} {
		input := fixturePipelineRunContext(NewMemoryRuntime())
		a := NormalizeAgentSettings(AgentSettings{}, 2000)
		a.EnableLengthNorm = false
		input.Agent = &a
		input.WriterRuleVersion = tc.version
		provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
		bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
		if e != nil {
			t.Fatal(e)
		}
		injected := false
		for _, request := range provider.requests {
			if strings.Contains(request.SystemPrompt, "写作铁律") {
				injected = true
			}
		}
		if injected != tc.want {
			t.Fatalf("version=%d rules injected=%v", tc.version, injected)
		}
		var snapshot map[string]json.RawMessage
		if e = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); e != nil {
			t.Fatal(e)
		}
		if _, ok := snapshot["preWriteCheck"]; ok {
			t.Fatalf("fixture draft has no preWriteCheck")
		}
	}
}

func TestPipelineKeepsPreWriteCheckOutOfContentAndAdoptsTitle(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	a.EnableLengthNorm = false
	input.Agent = &a
	input.WriterRuleVersion = 1
	draft := `{"preWriteCheck":[{"item":"当前任务","record":"拆开匿名信","note":"对齐章计划"}],"title":"第1章 雨夜来信","content":"雨水敲着窗。林澜拆开了那封没有寄件人的信。"}`
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draft, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(bundle.Chapter.Content, "preWriteCheck") || strings.Contains(bundle.Chapter.Content, "拆开匿名信") {
		t.Fatalf("pre-write check leaked into prose: %s", bundle.Chapter.Content)
	}
	if bundle.Chapter.Title != "雨夜来信" {
		t.Fatalf("draft title not adopted: %q", bundle.Chapter.Title)
	}
	var snapshot map[string]json.RawMessage
	if e = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); e != nil {
		t.Fatal(e)
	}
	var rows []preWriteCheckRow
	if e = json.Unmarshal(snapshot["preWriteCheck"], &rows); e != nil || len(rows) != 1 || rows[0].Item != "当前任务" {
		t.Fatalf("preWriteCheck snapshot=%s err=%v", snapshot["preWriteCheck"], e)
	}
}
