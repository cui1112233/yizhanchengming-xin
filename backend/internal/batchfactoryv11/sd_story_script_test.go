package batchfactoryv11

import (
	"context"
	"strings"
	"testing"
)

func TestValidateSDStoryScriptRowsRequiresOneRowPerSourceLine(t *testing.T) {
	book := Book{SourceText: "甲走进房间。\n乙抬头。"}
	valid := `{"rows":[{"source_index":1,"source_text":"甲走进房间。","story_script_line":"甲推门进入室内。","covered_source_span_keys":["L0001-P01"]},{"source_index":2,"source_text":"乙抬头。","story_script_line":"乙抬头看向门口。","covered_source_span_keys":["L0002-P01"]}]}`
	if _, err := validateSDStoryScriptRows(valid, book); err != nil {
		t.Fatalf("valid SD story script rejected: %v", err)
	}
	missing := `{"rows":[{"source_index":1,"source_text":"甲走进房间。","story_script_line":"甲推门进入室内。","covered_source_span_keys":["L0001-P01"]}]}`
	if _, err := validateSDStoryScriptRows(missing, book); err == nil {
		t.Fatal("missing source row was accepted")
	}
}

func TestBuildSDStoryScriptContractUsesOriginalStoryVersion(t *testing.T) {
	contract := buildSDStoryScriptContract(Book{SourceText: "甲走进房间。\n乙抬头。"}, DirectorSnapshot{Effective: SettingsPatch{}})
	if contract.SystemPrompt == "" || contract.UserPrompt == "" {
		t.Fatal("SD story script contract is empty")
	}
	if got := contract.UserPrompt; !containsAll(got, []string{"sd.original-story.v1", "line_contract", "source_text"}) {
		t.Fatalf("story script contract lost required fields: %s", got)
	}
}

func TestRunDirectorWithSDStoryVersionUsesStoryScriptHandoff(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"storyTimelineVersion": rawSetting(t, sdOriginalStoryVersion),
			"aiPromptConfig": rawSetting(t, map[string]any{"video": map[string]any{
				"enabled": true, "scope": "all", "presetId": "batch-video-sd", "body": "SD VIDEO RULE",
			}}),
		}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		`{"rows":[{"source_index":1,"source_text":"她被当众羞辱后沉默离开。","story_script_line":"她低头离开人群。","covered_source_span_keys":["L0001-P01"]}]}`,
		"===VIDEO 01===\n时长：1秒\n资产引用：人物=无；场景=无；道具=无\n她沉默离开。",
	}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatalf("SD staged director failed: %v", err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("SD staged director calls=%d, want story script plus director", len(provider.calls))
	}
	if !strings.Contains(provider.calls[1].UserPrompt, "story_script_rows") || !strings.Contains(provider.calls[1].SystemPrompt, "SD_STORY_SCRIPT") {
		t.Fatalf("director call did not receive frozen story-script handoff: %+v", provider.calls[1])
	}
}

func containsAll(value string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
