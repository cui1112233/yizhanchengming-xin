package agentworkspace

import (
	"context"
	"errors"
	"time"
)

var (
	ErrProjectNotFound  = errors.New("agent project not found")
	ErrProjectExists    = errors.New("agent project already exists")
	ErrRevisionConflict = errors.New("agent canvas revision conflict")
)

type Store interface {
	ListProjects(context.Context, string) ([]ProjectSummary, error)
	Project(context.Context, string, string) (Project, error)
	CreateProject(context.Context, string, Project) error
	RenameProject(context.Context, string, string, string, time.Time) (Project, error)
	ClearMessages(context.Context, string, string, time.Time) error
	DeleteProject(context.Context, string, string, time.Time) error
	AppendExchange(context.Context, string, string, Message, Message) (Project, error)
	Canvas(context.Context, string, string) (Canvas, error)
	SaveCanvas(context.Context, string, string, Canvas, int64, time.Time) (Canvas, error)
	Executions(context.Context, string, string) ([]Execution, error)
	ImportProject(context.Context, string, Project) error
	LegacyImport(context.Context, string, string) (LegacyImport, bool, error)
	RecordLegacyImport(context.Context, LegacyImport) error
	VideoJobStore
}
