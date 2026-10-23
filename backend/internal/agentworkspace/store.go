package agentworkspace

import (
	"context"
	"errors"
	"time"
)

var (
	ErrProjectNotFound      = errors.New("agent project not found")
	ErrProjectExists        = errors.New("agent project already exists")
	ErrRevisionConflict     = errors.New("agent canvas revision conflict")
	ErrExecutionExists      = errors.New("agent execution already exists")
	ErrExecutionNotFound    = errors.New("agent execution not found")
	ErrExecutionState       = errors.New("invalid agent execution state")
	ErrExecutionNodeChanged = errors.New("agent execution node changed")
	ErrCanvasInvalid        = errors.New("invalid agent canvas")
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
	BeginNodeExecution(context.Context, string, string, Execution) (Execution, error)
	SetNodeExecutionAwaitingConfirmation(context.Context, string, string, string, map[string]any, time.Time) (Execution, error)
	ApplyNodeExecution(context.Context, string, string, ApplyNodeExecutionInput) (Canvas, Execution, error)
	FailNodeExecution(context.Context, string, string, string, map[string]any, time.Time) (Execution, error)
	ImportProject(context.Context, string, Project) error
	LegacyImport(context.Context, string, string) (LegacyImport, bool, error)
	RecordLegacyImport(context.Context, LegacyImport) error
	VideoJobStore
	GenerationTaskStore
	ScriptShotStore
	BatchFactoryBindingStore
}
