package generation

import (
	"context"
	"errors"
)

func (s *Service) BookSummary(ctx context.Context, projectID, bookID int64) (BookGenerationResult, error) {
	if projectID <= 0 || bookID <= 0 { return BookGenerationResult{}, ErrInvalid }
	book, err := s.store.GetBookForProject(ctx, projectID, bookID)
	if err != nil { return BookGenerationResult{}, err }
	history, err := s.bookHistory(ctx, projectID, bookID)
	if err != nil { return BookGenerationResult{}, err }
	run, err := s.store.LatestBookRun(ctx, projectID, bookID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return BookGenerationResult{SourceText: book.OriginalText, Latest: map[Stage]StageRun{}, History: history}, nil
		}
		return BookGenerationResult{}, err
	}
	result, err := s.result(ctx, run)
	if err != nil { return BookGenerationResult{}, err }
	result.SourceText = book.OriginalText
	result.History = history
	return result, nil
}

func (s *Service) bookHistory(ctx context.Context, projectID, bookID int64) ([]GenerationHistoryEntry, error) {
	runs, err := s.store.ListBookRunsByProject(ctx, projectID)
	if err != nil { return nil, err }
	out := make([]GenerationHistoryEntry, 0, 20)
	for index := len(runs) - 1; index >= 0 && len(out) < 20; index-- {
		run := runs[index]
		if run.BookID != bookID { continue }
		stages, stageErr := s.store.ListStageRuns(ctx, run.ID)
		if stageErr != nil { return nil, stageErr }
		latest := map[Stage]StageRun{}
		for _, value := range stages {
			current, ok := latest[value.Stage]
			if !ok || value.Attempt > current.Attempt || (value.Attempt == current.Attempt && value.ID > current.ID) {
				latest[value.Stage] = value
			}
		}
		editableOutput, compiledPrompt := generationOutputs(latest)
		out = append(out, GenerationHistoryEntry{
			Run: run, Latest: latest,
			EditableOutput: editableOutput, CompiledPrompt: compiledPrompt,
		})
	}
	return out, nil
}

func (s *Service) StageResult(ctx context.Context, projectID, bookID int64, stage Stage) (StageRun, error) {
	run, err := s.store.LatestBookRun(ctx, projectID, bookID)
	if err != nil { return StageRun{}, err }
	return s.store.LatestStageRun(ctx, run.ID, stage)
}

func (s *Service) ListPrompts(ctx context.Context) ([]Prompt, error) {
	if s == nil || s.store == nil { return nil, ErrUnavailable }
	return s.store.ListPrompts(ctx)
}
