package novelpanel

import (
	"context"
	"errors"
	"testing"
)

func premiumWorkspace() Workspace {
	characters := ParseForcedRoster("沈清月（青年）,顾川（青年）")
	characters[0].Appearance = "乌黑长发，红色礼服"
	characters[1].Appearance = "短发，深色西装"
	return Workspace{
		ProjectID:    9,
		Mode:         ModePremiumIllustrated,
		OriginalText: "结婚十周年纪念宴上，她宣布了消息。\n我平静地给出选择。",
		TextProcessing: TextProcessingSettings{
			TrimLineWhitespace: true,
		},
		ForcedRoster: "沈清月（青年）,顾川（青年）",
		Characters:   characters,
		Relationships: []CharacterRelationship{{
			ID: "rel-1", FromID: characters[0].ID, ToID: characters[1].ID, Type: "旧识", Description: "关系存在冲突",
		}},
		ContentType:  "小说推文",
		UnifiedStyle: "现代都市写实电影感，低饱和",
		Density:      DensityDetailed,
		CaseLearning: CaseLearningSettings{Enabled: true, Material: "案例：开场先给冲突，再补关系。"},
		Instructions: InstructionSettings{
			GenerationRules:        "不得丢失原文事件",
			MustCoverDetails:       "纪念宴、二选一",
			ShotRhythmRequirements: "前三秒建立冲突",
		},
		Shots: []Shot{
			{ID: "s1", SourceIndex: 1, SourceBasis: "结婚十周年纪念宴上，她宣布了消息。", Visual: "宴会厅中景，女方举杯宣布消息", DurationSec: 4},
			{ID: "s2", SourceIndex: 2, SourceBasis: "我平静地给出选择。", Visual: "男方近景，神情克制", DurationSec: 3},
		},
	}
}

func TestServiceSaveHistoryAndRestore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	service := NewService(store)
	initial := premiumWorkspace()

	first, err := service.Save(ctx, SaveRequest{Workspace: initial, ExpectedRevision: 0, Note: "初版"})
	if err != nil {
		t.Fatalf("save first version: %v", err)
	}
	if first.Workspace.Revision != 1 || first.History.Revision != 1 {
		t.Fatalf("unexpected revision after first save: %#v", first)
	}
	if first.Workspace.ProcessedText == "" || len(first.Workspace.Characters) != 2 || len(first.Workspace.Relationships) != 1 {
		t.Fatalf("normalized workspace incomplete: %#v", first.Workspace)
	}

	secondWorkspace := first.Workspace
	secondWorkspace.UnifiedStyle = "高对比都市夜景"
	second, err := service.Save(ctx, SaveRequest{Workspace: secondWorkspace, ExpectedRevision: 1, Note: "改风格"})
	if err != nil {
		t.Fatalf("save second version: %v", err)
	}
	if second.Workspace.Revision != 2 {
		t.Fatalf("expected revision 2, got %d", second.Workspace.Revision)
	}

	history, err := service.ListHistory(ctx, 9, 10)
	if err != nil || len(history) != 2 {
		t.Fatalf("history mismatch: len=%d err=%v", len(history), err)
	}
	restored, err := service.Restore(ctx, RestoreRequest{ProjectID: 9, HistoryID: first.History.ID, ExpectedRevision: 2})
	if err != nil {
		t.Fatalf("restore first version: %v", err)
	}
	if restored.Workspace.Revision != 3 || restored.Workspace.UnifiedStyle != "现代都市写实电影感，低饱和" {
		t.Fatalf("restore did not recover snapshot: %#v", restored.Workspace)
	}
	if restored.History.Note == "" {
		t.Fatal("restore must create a new auditable history record")
	}
}

func TestServiceRejectsRevisionConflictWithoutOverwrite(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore())
	first, err := service.Save(ctx, SaveRequest{Workspace: premiumWorkspace(), ExpectedRevision: 0})
	if err != nil {
		t.Fatal(err)
	}
	stale := first.Workspace
	stale.UnifiedStyle = "不应该覆盖"
	_, err = service.Save(ctx, SaveRequest{Workspace: stale, ExpectedRevision: 0})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	current, err := service.GetWorkspace(ctx, 9)
	if err != nil {
		t.Fatal(err)
	}
	if current.UnifiedStyle == "不应该覆盖" || current.Revision != 1 {
		t.Fatalf("stale save overwrote current workspace: %#v", current)
	}
}

func TestInvalidStoryboardNeverOverwritesSavedWorkspace(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewMemoryStore())
	first, err := service.Save(ctx, SaveRequest{Workspace: premiumWorkspace(), ExpectedRevision: 0})
	if err != nil {
		t.Fatal(err)
	}
	invalid := first.Workspace
	invalid.Shots = invalid.Shots[:1]
	if _, err := service.Save(ctx, SaveRequest{Workspace: invalid, ExpectedRevision: 1}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected validation error, got %v", err)
	}
	current, err := service.GetWorkspace(ctx, 9)
	if err != nil {
		t.Fatal(err)
	}
	if current.Revision != 1 || len(current.Shots) != 2 {
		t.Fatalf("invalid result changed saved data: %#v", current)
	}
}

func TestPrepareStoryboardRequestCarriesPanelOnlyConfigurationToSharedGeneratorBoundary(t *testing.T) {
	service := NewService(NewMemoryStore())
	workspace := premiumWorkspace()
	request, err := service.PrepareStoryboardRequest(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if request.Mode != ModePremiumIllustrated || request.Density != DensityDetailed || !request.CaseLearning.Enabled {
		t.Fatalf("panel configuration missing: %#v", request)
	}
	if len(request.SourceLines) != 2 || len(request.Characters) != 2 || len(request.Relationships) != 1 {
		t.Fatalf("storyboard request lost source/character context: %#v", request)
	}
	if request.Instructions.MustCoverDetails != "纪念宴、二选一" {
		t.Fatalf("instruction settings missing: %#v", request.Instructions)
	}
}

func TestPremiumModeRequiresUnifiedStyle(t *testing.T) {
	service := NewService(NewMemoryStore())
	workspace := premiumWorkspace()
	workspace.UnifiedStyle = ""
	_, err := service.Save(context.Background(), SaveRequest{Workspace: workspace})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected premium validation error, got %v", err)
	}
}
