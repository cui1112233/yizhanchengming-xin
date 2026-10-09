package generation

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type StoryboardCard struct {
	ID       int64  `json:"id"`
	Position int    `json:"position"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Version  int    `json:"version"`
}
type StoryboardDocument struct {
	BatchProjectID   int64            `json:"batchProjectId"`
	BookID           int64            `json:"bookId"`
	Version          int              `json:"version"`
	SourceStageRunID int64            `json:"sourceStageRunId,omitempty"`
	Cards            []StoryboardCard `json:"cards"`
}
type SaveStoryboardCardRequest struct {
	Card            StoryboardCard `json:"card"`
	ExpectedVersion int            `json:"expectedVersion"`
}

type StoryboardStore interface {
	GetStoryboard(context.Context, int64, int64) (StoryboardDocument, error)
	CreateStoryboard(context.Context, StoryboardDocument) (StoryboardDocument, error)
	SaveStoryboardCard(context.Context, int64, int64, StoryboardCard, int) (StoryboardDocument, error)
	DeleteStoryboardCard(context.Context, int64, int64, int64, int) (StoryboardDocument, error)
	ReorderStoryboard(context.Context, int64, int64, []int64, int) (StoryboardDocument, error)
}

func (s *Service) storyboardStore() (StoryboardStore, error) {
	value, ok := s.store.(StoryboardStore)
	if !ok {
		return nil, ErrUnavailable
	}
	return value, nil
}
func splitStoryboard(text string) []StoryboardCard {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	re := regexp.MustCompile(`(?m)^###\s*(分镜[^\n]*)\s*$`)
	indexes := re.FindAllStringSubmatchIndex(text, -1)
	if len(indexes) == 0 {
		return []StoryboardCard{{Position: 1, Title: "分镜一", Content: text}}
	}
	cards := make([]StoryboardCard, 0, len(indexes))
	for i, m := range indexes {
		end := len(text)
		if i+1 < len(indexes) {
			end = indexes[i+1][0]
		}
		content := strings.TrimSpace(text[m[1]:end])
		content = strings.TrimSpace(strings.Trim(content, "-"))
		cards = append(cards, StoryboardCard{Position: i + 1, Title: strings.TrimSpace(text[m[2]:m[3]]), Content: content})
	}
	return cards
}
func (s *Service) Storyboard(ctx context.Context, projectID, bookID int64) (StoryboardDocument, error) {
	store, err := s.storyboardStore()
	if err != nil {
		return StoryboardDocument{}, err
	}
	value, err := store.GetStoryboard(ctx, projectID, bookID)
	if err == nil {
		return value, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return StoryboardDocument{}, err
	}
	run, err := s.store.LatestBookRun(ctx, projectID, bookID)
	if err != nil {
		return StoryboardDocument{}, err
	}
	director, err := s.store.LatestStageRun(ctx, run.ID, StageDirector)
	if err != nil {
		return StoryboardDocument{}, err
	}
	if director.Status != StatusCompleted {
		return StoryboardDocument{}, ErrConflict
	}
	return store.CreateStoryboard(ctx, StoryboardDocument{BatchProjectID: projectID, BookID: bookID, Version: 1, SourceStageRunID: director.ID, Cards: splitStoryboard(director.OutputText)})
}
func (s *Service) SaveStoryboardCard(ctx context.Context, projectID, bookID int64, req SaveStoryboardCardRequest) (StoryboardDocument, error) {
	if strings.TrimSpace(req.Card.Content) == "" || req.ExpectedVersion < 1 {
		return StoryboardDocument{}, ErrInvalid
	}
	store, err := s.storyboardStore()
	if err != nil {
		return StoryboardDocument{}, err
	}
	return store.SaveStoryboardCard(ctx, projectID, bookID, req.Card, req.ExpectedVersion)
}
func (s *Service) DeleteStoryboardCard(ctx context.Context, projectID, bookID, cardID int64, expected int) (StoryboardDocument, error) {
	if expected < 1 {
		return StoryboardDocument{}, ErrInvalid
	}
	store, err := s.storyboardStore()
	if err != nil {
		return StoryboardDocument{}, err
	}
	return store.DeleteStoryboardCard(ctx, projectID, bookID, cardID, expected)
}
func (s *Service) ReorderStoryboard(ctx context.Context, projectID, bookID int64, ids []int64, expected int) (StoryboardDocument, error) {
	if expected < 1 || len(ids) == 0 {
		return StoryboardDocument{}, ErrInvalid
	}
	store, err := s.storyboardStore()
	if err != nil {
		return StoryboardDocument{}, err
	}
	return store.ReorderStoryboard(ctx, projectID, bookID, ids, expected)
}
func (s *Service) RecompileStoryboard(ctx context.Context, projectID, bookID int64, requestID string) (BookGenerationResult, error) {
	run, err := s.store.LatestBookRun(ctx, projectID, bookID)
	if err != nil {
		return BookGenerationResult{}, err
	}
	if run.RunID != 0 {
		return BookGenerationResult{}, ErrConflict
	}
	doc, err := s.Storyboard(ctx, projectID, bookID)
	if err != nil {
		return BookGenerationResult{}, err
	}
	if len(doc.Cards) == 0 {
		return BookGenerationResult{}, ErrInvalid
	}
	sort.Slice(doc.Cards, func(i, j int) bool { return doc.Cards[i].Position < doc.Cards[j].Position })
	book, err := s.store.GetBookForProject(ctx, projectID, bookID)
	if err != nil {
		return BookGenerationResult{}, err
	}
	script, err := s.store.LatestStageRun(ctx, run.ID, StageScript)
	if err != nil {
		return BookGenerationResult{}, err
	}
	if script.Status != StatusCompleted {
		return BookGenerationResult{}, ErrConflict
	}
	hookText := ""
	if hook, e := s.store.LatestStageRun(ctx, run.ID, StageHook); e == nil && hook.Status == StatusCompleted {
		hookText = hook.OutputText
	}
	run.RequestID = strings.TrimSpace(requestID)
	run.Status = StatusRunning
	run.ErrorMessage = ""
	run.FinishedAt = nil
	if run, err = s.store.UpdateBookRun(ctx, run); err != nil {
		return BookGenerationResult{}, err
	}
	directorPrompt, err := s.resolver.Resolve(ctx, PromptDirector)
	if err != nil {
		return s.failRun(ctx, run, err)
	}
	parts := make([]string, 0, len(doc.Cards))
	for _, c := range doc.Cards {
		parts = append(parts, fmt.Sprintf("### %s\n%s", c.Title, c.Content))
	}
	directorText := strings.Join(parts, "\n\n---\n\n")
	snapshot := fmt.Sprintf(`{"manualStoryboardVersion":%d,"source":"script_canvas"}`, doc.Version)
	if _, err = s.completeLocalStage(ctx, run, book, StageDirector, directorPrompt, directorText, snapshot); err != nil {
		return s.failRun(ctx, run, err)
	}
	finalPrompt, err := s.resolver.Resolve(ctx, PromptFinal)
	if err != nil {
		return s.failRun(ctx, run, err)
	}
	input := FinalPromptInput{SystemPreset: finalPrompt.Content, Script: script.OutputText, Hook: hookText, Director: directorText}
	compiled := s.compiler.Compile(input)
	if _, err = s.completeLocalStage(ctx, run, book, StageFinalPrompt, finalPrompt, compiled, fmt.Sprintf(`{"manualStoryboardVersion":%d}`, doc.Version)); err != nil {
		return s.failRun(ctx, run, err)
	}
	now := s.now().UTC()
	run.Status = StatusCompleted
	run.FinishedAt = &now
	run, err = s.store.UpdateBookRun(ctx, run)
	if err != nil {
		return BookGenerationResult{}, err
	}
	return s.result(ctx, run)
}
