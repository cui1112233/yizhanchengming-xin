package generation

import (
	"context"
	"errors"
)

func (s *Service) BookSummary(ctx context.Context, projectID, bookID int64) (BookGenerationResult, error) {
	if projectID <= 0 || bookID <= 0 { return BookGenerationResult{}, ErrInvalid }
	if _, err := s.store.GetBookForProject(ctx, projectID, bookID); err != nil { return BookGenerationResult{}, err }
	run, err := s.store.LatestBookRun(ctx, projectID, bookID)
	if err != nil {
		if errors.Is(err, ErrNotFound) { return BookGenerationResult{Latest: map[Stage]StageRun{}}, nil }
		return BookGenerationResult{}, err
	}
	return s.result(ctx, run)
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
