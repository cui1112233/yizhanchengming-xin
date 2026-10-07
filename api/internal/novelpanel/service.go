package novelpanel

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) GetWorkspace(ctx context.Context, projectID int64) (Workspace, error) {
	if s == nil || s.store == nil {
		return Workspace{}, fmt.Errorf("%w: store unavailable", ErrInvalid)
	}
	return s.store.GetWorkspace(ctx, projectID)
}

func (s *Service) ListHistory(ctx context.Context, projectID int64, limit int) ([]HistoryRecord, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("%w: store unavailable", ErrInvalid)
	}
	return s.store.ListHistory(ctx, projectID, limit)
}

func (s *Service) Save(ctx context.Context, request SaveRequest) (SaveResult, error) {
	if s == nil || s.store == nil {
		return SaveResult{}, fmt.Errorf("%w: store unavailable", ErrInvalid)
	}
	workspace, err := normalizeWorkspace(request.Workspace)
	if err != nil {
		return SaveResult{}, err
	}
	saved, history, err := s.store.SaveWorkspace(ctx, request.ExpectedRevision, workspace, request.Note)
	if err != nil {
		return SaveResult{}, err
	}
	return SaveResult{Workspace: saved, History: history}, nil
}

func (s *Service) Restore(ctx context.Context, request RestoreRequest) (SaveResult, error) {
	if s == nil || s.store == nil {
		return SaveResult{}, fmt.Errorf("%w: store unavailable", ErrInvalid)
	}
	if request.ProjectID <= 0 || strings.TrimSpace(request.HistoryID) == "" {
		return SaveResult{}, fmt.Errorf("%w: project id and history id are required", ErrInvalid)
	}
	history, err := s.store.GetHistory(ctx, request.ProjectID, strings.TrimSpace(request.HistoryID))
	if err != nil {
		return SaveResult{}, err
	}
	restored := history.Workspace
	restored.ProjectID = request.ProjectID
	restored.Revision = request.ExpectedRevision
	restored, err = normalizeWorkspace(restored)
	if err != nil {
		return SaveResult{}, err
	}
	saved, newHistory, err := s.store.SaveWorkspace(ctx, request.ExpectedRevision, restored, "恢复记录 "+history.ID)
	if err != nil {
		return SaveResult{}, err
	}
	return SaveResult{Workspace: saved, History: newHistory}, nil
}

func (s *Service) PrepareStoryboardRequest(workspace Workspace) (StoryboardRequest, error) {
	normalized, err := normalizeWorkspace(workspace)
	if err != nil {
		return StoryboardRequest{}, err
	}
	lines := SourceLines(normalized.ProcessedText)
	if len(lines) == 0 {
		return StoryboardRequest{}, fmt.Errorf("%w: original text is required before storyboard generation", ErrInvalid)
	}
	return StoryboardRequest{
		ProjectID:     normalized.ProjectID,
		Mode:          normalized.Mode,
		ContentType:   normalized.ContentType,
		UnifiedStyle:  normalized.UnifiedStyle,
		Density:       normalized.Density,
		SourceLines:   lines,
		Characters:    append([]CharacterCard(nil), normalized.Characters...),
		Relationships: append([]CharacterRelationship(nil), normalized.Relationships...),
		CaseLearning:  normalized.CaseLearning,
		Instructions:  normalized.Instructions,
		CurrentShots:  append([]Shot(nil), normalized.Shots...),
	}, nil
}

func normalizeWorkspace(workspace Workspace) (Workspace, error) {
	if workspace.ProjectID <= 0 {
		return Workspace{}, fmt.Errorf("%w: project id is required", ErrInvalid)
	}
	if workspace.Mode == "" {
		workspace.Mode = ModeNormal
	}
	if workspace.Mode != ModeNormal && workspace.Mode != ModePremiumIllustrated {
		return Workspace{}, fmt.Errorf("%w: unsupported mode %q", ErrInvalid, workspace.Mode)
	}
	if workspace.Density == "" {
		workspace.Density = DensityStandard
	}
	if workspace.Density != DensityCompact && workspace.Density != DensityStandard && workspace.Density != DensityDetailed {
		return Workspace{}, fmt.Errorf("%w: unsupported density %q", ErrInvalid, workspace.Density)
	}
	if utf8.RuneCountInString(workspace.OriginalText) > MaxNovelRunes {
		return Workspace{}, fmt.Errorf("%w: original text exceeds %d characters", ErrInvalid, MaxNovelRunes)
	}
	if utf8.RuneCountInString(workspace.ForcedRoster) > MaxForcedRosterRunes {
		return Workspace{}, fmt.Errorf("%w: forced roster exceeds %d characters", ErrInvalid, MaxForcedRosterRunes)
	}
	workspace.ContentType = strings.TrimSpace(workspace.ContentType)
	if workspace.ContentType == "" {
		workspace.ContentType = "小说推文"
	}
	workspace.UnifiedStyle = strings.TrimSpace(workspace.UnifiedStyle)
	if utf8.RuneCountInString(workspace.UnifiedStyle) > MaxStyleRunes {
		return Workspace{}, fmt.Errorf("%w: unified style is too long", ErrInvalid)
	}
	if workspace.Mode == ModePremiumIllustrated && workspace.UnifiedStyle == "" {
		return Workspace{}, fmt.Errorf("%w: premium illustrated mode requires a unified style", ErrInvalid)
	}
	workspace.ProcessedText = ProcessText(workspace.OriginalText, workspace.TextProcessing)

	forced := ParseForcedRoster(workspace.ForcedRoster)
	workspace.Characters = MergeForcedCharacters(forced, workspace.Characters)
	for index := range workspace.Characters {
		workspace.Characters[index].Appearance = strings.TrimSpace(workspace.Characters[index].Appearance)
		workspace.Characters[index].CardNote = strings.TrimSpace(workspace.Characters[index].CardNote)
		workspace.Characters[index].Gender = strings.TrimSpace(workspace.Characters[index].Gender)
		if workspace.Characters[index].Gender == "" {
			workspace.Characters[index].Gender = "待确认"
		}
	}
	if err := ValidateRelationships(workspace.Characters, workspace.Relationships); err != nil {
		return Workspace{}, err
	}
	if err := ValidateShots(workspace.ProcessedText, workspace.Shots); err != nil {
		return Workspace{}, err
	}
	workspace.CaseLearning.Material = strings.TrimSpace(workspace.CaseLearning.Material)
	if utf8.RuneCountInString(workspace.CaseLearning.Material) > MaxCaseLearningRunes {
		return Workspace{}, fmt.Errorf("%w: case learning material is too long", ErrInvalid)
	}
	workspace.Instructions.GenerationRules = strings.TrimSpace(workspace.Instructions.GenerationRules)
	workspace.Instructions.MustCoverDetails = strings.TrimSpace(workspace.Instructions.MustCoverDetails)
	workspace.Instructions.ShotRhythmRequirements = strings.TrimSpace(workspace.Instructions.ShotRhythmRequirements)
	workspace.Instructions.NegativeInstructions = strings.TrimSpace(workspace.Instructions.NegativeInstructions)
	if instructionRunes(workspace.Instructions) > MaxInstructionRunes {
		return Workspace{}, fmt.Errorf("%w: instruction settings are too long", ErrInvalid)
	}
	return workspace, nil
}

func instructionRunes(settings InstructionSettings) int {
	return utf8.RuneCountInString(settings.GenerationRules) +
		utf8.RuneCountInString(settings.MustCoverDetails) +
		utf8.RuneCountInString(settings.ShotRhythmRequirements) +
		utf8.RuneCountInString(settings.NegativeInstructions)
}
