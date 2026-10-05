package generation

import (
	"context"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type Store interface {
	GetBookForProject(context.Context, int64, int64) (intake.Book, error)
	ListBooksForProject(context.Context, int64) ([]intake.Book, error)
	ResolvePrompt(context.Context, string) (Prompt, error)
	ListPrompts(context.Context) ([]Prompt, error)
	CreateBookRun(context.Context, BookRun) (BookRun, error)
	UpdateBookRun(context.Context, BookRun) (BookRun, error)
	LatestBookRun(context.Context, int64, int64) (BookRun, error)
	ListBookRunsByProject(context.Context, int64) ([]BookRun, error)
	CreateStageRun(context.Context, StageRun) (StageRun, error)
	UpdateStageRun(context.Context, StageRun) (StageRun, error)
	ListStageRuns(context.Context, int64) ([]StageRun, error)
	LatestStageRun(context.Context, int64, Stage) (StageRun, error)
}

type Provider interface {
	Complete(context.Context, TextRequest) (string, error)
}
