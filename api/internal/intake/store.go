package intake

import "context"

type Store interface {
	CreateIntake(ctx context.Context, name string) (Intake, error)
	GetIntake(ctx context.Context, id int64) (Intake, error)
	UpdateIntakeStatus(ctx context.Context, id int64, status Status) error
	UpsertBook(ctx context.Context, book Book) (Book, error)
	ListBooks(ctx context.Context, intakeID int64) ([]Book, error)
	CreateBatchProject(ctx context.Context, project BatchProject) (BatchProject, error)
	CreateRun(ctx context.Context, run Run) (Run, error)
}
