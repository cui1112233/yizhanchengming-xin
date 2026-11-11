package agentproxy

import (
	"context"
	"errors"
	"time"
)

var (
	ErrWorkspaceNotFound = errors.New("agent proxy workspace not found")
	ErrWorkspaceExists   = errors.New("agent proxy workspace already exists")
	ErrSessionNotFound   = errors.New("agent proxy session not found")
	ErrSessionExists     = errors.New("agent proxy session already exists")
	ErrInvalidInput      = errors.New("invalid agent proxy input")
	// ErrPermissionModeUnavailable 表示调用了第一版未开放的权限档位（bypass）。
	ErrPermissionModeUnavailable = errors.New("agent proxy permission mode unavailable")
	// ErrPermissionDenied 表示当前档位不允许这个动作，例如 plan 档位请求执行。
	ErrPermissionDenied = errors.New("agent proxy action denied by permission mode")
)

type Store interface {
	ListWorkspaces(context.Context, string) ([]Workspace, error)
	Workspace(context.Context, string, string) (Workspace, error)
	CreateWorkspace(context.Context, string, CreateWorkspaceInput, time.Time) (Workspace, error)
	UpdateWorkspace(context.Context, string, string, UpdateWorkspaceInput, time.Time) (Workspace, error)
	DeleteWorkspace(context.Context, string, string, time.Time) error

	ListSessions(context.Context, string, string) ([]Session, error)
	Session(context.Context, string, string) (Session, error)
	CreateSession(context.Context, string, string, CreateSessionInput, time.Time) (Session, error)
	UpdateSession(context.Context, string, string, UpdateSessionInput, time.Time) (Session, error)
	DeleteSession(context.Context, string, string, time.Time) error

	ListMessages(context.Context, string, string, int) ([]Message, error)
	AppendMessage(context.Context, string, string, AppendMessageInput, time.Time) (Message, error)

	ListSkills(context.Context, string, string) ([]Skill, error)
	ListScheduledJobs(context.Context, string, string) ([]ScheduledJob, error)
}
