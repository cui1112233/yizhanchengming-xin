package video

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
)

type finalPromptGenerationReader struct {
	bookRun  generation.BookRun
	stageRun generation.StageRun
	bookErr  error
	stageErr error
}

func (r finalPromptGenerationReader) LatestBookRun(context.Context, int64, int64) (generation.BookRun, error) {
	return r.bookRun, r.bookErr
}

func (r finalPromptGenerationReader) LatestStageRun(context.Context, int64, generation.Stage) (generation.StageRun, error) {
	return r.stageRun, r.stageErr
}

func TestGenerationFinalPromptSourceTracksStageVersionAndInputRevision(t *testing.T) {
	input := `{"directorRevision":7,"projectConfig":"v3"}`
	sum := sha256.Sum256([]byte(input))
	reader := finalPromptGenerationReader{
		bookRun: generation.BookRun{ID: 501, BatchProjectID: 51, BookID: 31, Status: generation.StatusCompleted},
		stageRun: generation.StageRun{
			ID: 901, BookRunID: 501, BookID: 31, Stage: generation.StageFinalPrompt,
			Status: generation.StatusCompleted, PromptKey: generation.PromptFinal, PromptVersion: 7,
			InputSnapshot: input, OutputText: "compiled final video prompt",
		},
	}
	source := NewGenerationFinalPromptSource(reader)
	got, err := source.ResolveFinalPrompt(context.Background(), 51, 31)
	if err != nil {
		t.Fatal(err)
	}
	if got.StageRunID != 901 || got.PromptVersion != 7 || got.Text != "compiled final video prompt" {
		t.Fatalf("FinalPrompt = %+v", got)
	}
	if got.InputRevision != hex.EncodeToString(sum[:]) {
		t.Fatalf("InputRevision = %q", got.InputRevision)
	}
}

func TestGenerationFinalPromptSourceRejectsIncompleteFinalPrompt(t *testing.T) {
	reader := finalPromptGenerationReader{
		bookRun: generation.BookRun{ID: 501, BatchProjectID: 51, BookID: 31, Status: generation.StatusRunning},
		stageRun: generation.StageRun{ID: 901, BookRunID: 501, BookID: 31, Stage: generation.StageFinalPrompt, Status: generation.StatusRunning},
	}
	source := NewGenerationFinalPromptSource(reader)
	if _, err := source.ResolveFinalPrompt(context.Background(), 51, 31); !errors.Is(err, ErrFinalPromptNotReady) {
		t.Fatalf("err = %v, want ErrFinalPromptNotReady", err)
	}
}
