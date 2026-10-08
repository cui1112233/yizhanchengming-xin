package intake

import "context"

type Store interface {
	CreateOwnedIntake(ctx context.Context, name string, actor ActorScope) (Intake, error)
	GetIntake(ctx context.Context, id int64) (Intake, error)
	ListIntakes(ctx context.Context) ([]Intake, error)
	ListVisibleIntakes(ctx context.Context, userID, teamID int64, elevated bool) ([]Intake, error)
	CanAccessIntake(ctx context.Context, intakeID, userID, teamID int64, elevated bool) (bool, error)
	UpdateIntakeStatus(ctx context.Context, id int64, status Status) error
	UpsertBook(ctx context.Context, book Book) (Book, error)
	ListBooks(ctx context.Context, intakeID int64) ([]Book, error)
	CreateBatchProject(ctx context.Context, project BatchProject) (BatchProject, error)
	CreateRun(ctx context.Context, run Run) (Run, error)
}
