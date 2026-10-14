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

func TestValidateSDStoryScriptRowsKeepsOriginalIndexesAcrossBlankLines(t *testing.T) {
	book := Book{SourceText: "甲走进房间。\n\n乙抬头。"}
	valid := `{"rows":[{"source_index":1,"source_text":"甲走进房间。","story_script_line":"甲推门进入室内。","covered_source_span_keys":["L0001-P01"]},{"source_index":3,"source_text":"乙抬头。","story_script_line":"乙抬头看向门口。","covered_source_span_keys":["L0003-P01"]}]}`
	if _, err := validateSDStoryScriptRows(valid, book); err != nil {
		t.Fatalf("blank source line must not renumber the second row: %v", err)
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

func TestRunDirectorWithSDStoryVersionUsesSingleVideoPromptCall(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"storyTimelineVersion": rawSetting(t, sdOriginalStoryVersion),
			"audioDurationSeconds": rawSetting(t, 1),
			"aiPromptConfig": rawSetting(t, map[string]any{"video": map[string]any{
				"enabled": true, "scope": "all", "presetId": "batch-video-sd", "body": "SD VIDEO RULE",
			}}),
		}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	provider := &queuedDirectorProvider{values: []string{
		"===VIDEO 01===\n时长：1秒\n资产引用：人物=无；场景=无；道具=无\n[场景 1｜对应原文第 1 行｜总时长 1 秒]\n画面故事剧本：她低头离开人群。\n场景环境：空旷街道。\n[镜头 1｜出镜人物：无｜时长 1 秒] 她沉默离开。",
	}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatalf("SD unified director failed: %v", err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("SD unified director calls=%d, want one video prompt call", len(provider.calls))
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "单次视频提示词流程") || !strings.Contains(provider.calls[0].SystemPrompt, "画面故事剧本") {
		t.Fatalf("director call did not receive single-call story-script contract: %+v", provider.calls[0])
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
