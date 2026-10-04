package intake

import "context"

type Store interface {
	CreateIntake(ctx context.Context, name string) (Intake, error)
	GetIntake(ctx context.Context, id int64) (Intake, error)
	UpsertBook(ctx context.Context, book Book) (Book, error)
	CreateBatchProject(ctx context.Context, project BatchProject) (BatchProject, error)
	CreateRun(ctx context.Context, run Run) (Run, error)
}
