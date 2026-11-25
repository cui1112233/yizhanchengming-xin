package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func postWriteFixtureChapters() []Chapter {
	return []Chapter{
		{ID: "c1", Number: 1, Title: "春雨", Content: "雨水敲着窗，林澜没有开灯。"},
		{ID: "c2", Number: 2, Title: "旧信", Content: "信封上没有寄件人，只有一枚褪色的邮戳。"},
		{ID: "c3", Number: 3, Title: "门缝", Content: "门缝里塞进一张纸条，字迹和信封一致。"},
		{ID: "c4", Number: 4, Title: "雨停", Content: "天快亮时雨停了，楼下有人发动汽车。"},
	}
}

func TestBuildPostWriteContextFreezesRealChapterData(t *testing.T) {
	settings := ProjectSettings{Rules: []string{"本书采用第一人称叙述"}}
	project := Project{Genre: "悬疑", Protagonist: "林澜"}
	got := buildPostWriteContext(postWriteFixtureChapters(), 5, settings, project)

	if len(got.ExistingTitles) != 4 || got.ExistingTitles[3] != "雨停" {
		t.Fatalf("titles=%v", got.ExistingTitles)
	}
	if !got.FirstPerson || got.FirstPersonSource != "rules" {
		t.Fatalf("person=%v source=%s", got.FirstPerson, got.FirstPersonSource)
	}
	if got.Protagonist != "林澜" || got.Genre != "悬疑" {
		t.Fatalf("project=%+v", got)
	}
	if len(got.PreviousFullText) != 3 || len(got.SourceChapters) != 3 {
		t.Fatalf("previous=%d chapters=%v", len(got.PreviousFullText), got.SourceChapters)
	}
	if got.SourceChapters[0] != 2 || got.SourceChapters[2] != 4 {
		t.Fatalf("only the latest three chapters may participate: %v", got.SourceChapters)
	}
	if !strings.Contains(got.PreviousFullText[2], "雨停了") {
		t.Fatalf("latest full text missing: %v", got.PreviousFullText)
	}
	for _, word := range []string{"忍不住", "蛛丝马迹"} {
		found := false
		for _, value := range got.FatigueWords {
			if value == word {
				found = true
			}
		}
		if !found {
			t.Fatalf("fatigue words missing %s: %v", word, got.FatigueWords)
		}
	}
	status := got.postWriteStatus()
	for key, want := range map[string]string{
		"cross_chapter":   "checked_full_text",
		"duplicate_title": "checked_history_titles",
		"fatigue":         "checked_genre_fatigue_words",
		"person":          "checked_rules_declaration",
	} {
		if status[key] != want {
			t.Fatalf("%s=%s want=%s", key, status[key], want)
		}
	}
}

func TestBuildPostWriteContextExcludesTargetAndArchivedNumbers(t *testing.T) {
	got := buildPostWriteContext(postWriteFixtureChapters(), 3, ProjectSettings{}, Project{})
	if len(got.PreviousFullText) != 2 || got.SourceChapters[1] != 2 {
		t.Fatalf("target chapter leaked into history: %v", got.SourceChapters)
	}
	if got.FirstPersonSource != "absent" {
		t.Fatalf("undeclared person must stay absent: %s", got.FirstPersonSource)
	}
	status := got.postWriteStatus()
	if status["person"] != "not_checked_missing_person_declaration" {
		t.Fatalf("person status=%s", status["person"])
	}
}

func TestHistoricalRunWithoutPostWriteContextReportsUnchecked(t *testing.T) {
	p := postWriteContextOf(nil)
	status := p.postWriteStatus()
	for key, want := range map[string]string{
		"cross_chapter":   "not_checked_missing_full_text",
		"duplicate_title": "not_checked_missing_history_titles",
	} {
		if status[key] != want {
			t.Fatalf("%s=%s want=%s", key, status[key], want)
		}
	}
	if p.postWriteInput("春").Title != "春" {
		t.Fatal("title lost")
	}
}

func TestPipelineFeedsFullTextIntoCrossChapterCheck(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	a.EnableLengthNorm = false
	a.EnablePostWrite = true
	input.Agent = &a
	repeated := strings.Repeat("甲乙丙丁戊己庚辛", 3)
	input.PostWrite = PostWriteContext{PreviousFullText: []string{repeated}, SourceChapters: []int{1}}
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	var snapshot map[string]json.RawMessage
	if e = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); e != nil {
		t.Fatal(e)
	}
	var checks struct {
		PostWrite struct {
			Issues []LocalIssue `json:"issues"`
		} `json:"post_write"`
		Status map[string]string `json:"post_write_status"`
	}
	if e = json.Unmarshal(snapshot["localChecks"], &checks); e != nil {
		t.Fatal(e)
	}
	if checks.Status["cross_chapter"] != "checked_full_text" {
		t.Fatalf("status=%v", checks.Status)
	}
}
