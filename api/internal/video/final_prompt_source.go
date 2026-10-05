package video

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
)

var ErrFinalPromptNotReady = errors.New("video: final prompt not ready")

type GenerationReader interface {
	LatestBookRun(context.Context, int64, int64) (generation.BookRun, error)
	LatestStageRun(context.Context, int64, generation.Stage) (generation.StageRun, error)
}

type GenerationFinalPromptSource struct {
	reader GenerationReader
}

func NewGenerationFinalPromptSource(reader GenerationReader) *GenerationFinalPromptSource {
	return &GenerationFinalPromptSource{reader: reader}
}

func (s *GenerationFinalPromptSource) ResolveFinalPrompt(ctx context.Context, projectID, bookID int64) (FinalPrompt, error) {
	if s == nil || s.reader == nil {
		return FinalPrompt{}, fmt.Errorf("%w: generation reader unavailable", ErrFinalPromptNotReady)
	}
	bookRun, err := s.reader.LatestBookRun(ctx, projectID, bookID)
	if err != nil {
		return FinalPrompt{}, fmt.Errorf("%w: latest book run: %v", ErrFinalPromptNotReady, err)
	}
	if bookRun.ID <= 0 || bookRun.Status != generation.StatusCompleted {
		return FinalPrompt{}, ErrFinalPromptNotReady
	}
	stage, err := s.reader.LatestStageRun(ctx, bookRun.ID, generation.StageFinalPrompt)
	if err != nil {
		return FinalPrompt{}, fmt.Errorf("%w: final prompt stage: %v", ErrFinalPromptNotReady, err)
	}
	if stage.ID <= 0 || stage.Status != generation.StatusCompleted || stage.Stage != generation.StageFinalPrompt || stage.PromptVersion <= 0 || strings.TrimSpace(stage.OutputText) == "" {
		return FinalPrompt{}, ErrFinalPromptNotReady
	}
	sum := sha256.Sum256([]byte(stage.InputSnapshot))
	return FinalPrompt{
		StageRunID: stage.ID,
		PromptVersion: stage.PromptVersion,
		InputRevision: hex.EncodeToString(sum[:]),
		Text: stage.OutputText,
	}, nil
}
