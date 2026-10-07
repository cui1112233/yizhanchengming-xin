package agentstudio

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrNotFound            = errors.New("agent studio: not found")
	ErrForbidden           = errors.New("agent studio: forbidden")
	ErrInvalid             = errors.New("agent studio: invalid")
	ErrConflict            = errors.New("agent studio: conflict")
	ErrStorageUnavailable  = errors.New("agent studio: object storage is unavailable")
	ErrExecutorUnavailable = errors.New("executor_unavailable")
)

type Actor struct{ UserID, TeamID int64 }
type Project struct {
	ID, OwnerUserID, TeamID int64
	Title                   string
	BatchProjectID, BookID  int64
	CreatedAt, UpdatedAt    time.Time
}
type CreateProjectInput struct {
	Title                  string
	BatchProjectID, BookID int64
}
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

type Message struct {
	ID, ProjectID int64
	Role          MessageRole
	Content       string
	CreatedAt     time.Time
}
type ExecutionStatus string

const (
	ExecutionQueued      ExecutionStatus = "queued"
	ExecutionRunning     ExecutionStatus = "running"
	ExecutionCompleted   ExecutionStatus = "completed"
	ExecutionFailed      ExecutionStatus = "failed"
	ExecutionUnavailable ExecutionStatus = "executor_unavailable"
)

type Execution struct {
	ID, ProjectID                       int64
	Status                              ExecutionStatus
	RuntimeRef, ErrorCode, ErrorMessage string
	CreatedAt, UpdatedAt                time.Time
}
type ContinueInput struct {
	Content       string  `json:"content"`
	SkillIDs      []int64 `json:"skillIds"`
	AttachmentIDs []int64 `json:"attachmentIds"`
}
type ContinueResult struct {
	Project   Project
	User      Message
	Assistant *Message
	Execution Execution
}
type Canvas struct {
	ProjectID int64  `json:"projectId"`
	Revision  int    `json:"revision"`
	Document  []byte `json:"document"`
}
type CanvasVersion struct {
	ID, ProjectID int64
	Revision      int
	Document      []byte
	CreatedAt     time.Time
}
type Attachment struct {
	ID, ProjectID, OwnerUserID                       int64
	Bucket, ObjectKey, Filename, ContentType, SHA256 string
	ByteSize                                         int64
	CreatedAt                                        time.Time
}
type Skill struct {
	ID, OwnerUserID int64
	Name            string
	Version         int
	Body            string
	Enabled         bool
	CreatedAt       time.Time
}
type CreateSkillInput struct{ Name, Body string }

type Store interface {
	CreateProject(context.Context, Actor, CreateProjectInput) (Project, error)
	GetProject(context.Context, Actor, int64) (Project, error)
	CreateMessage(context.Context, Message) (Message, error)
	CreateExecution(context.Context, Execution) (Execution, error)
	UpdateExecution(context.Context, Execution) (Execution, error)
	ListProjects(context.Context, Actor) ([]Project, error)
	ListMessages(context.Context, Actor, int64) ([]Message, error)
	DeleteProject(context.Context, Actor, int64) error
	ListExecutions(context.Context, Actor, int64) ([]Execution, error)
	CreateSkill(context.Context, Actor, CreateSkillInput) (Skill, error)
	ListSkills(context.Context, Actor) ([]Skill, error)
	GetSkills(context.Context, Actor, []int64) ([]Skill, error)
	GetCanvas(context.Context, Actor, int64) (Canvas, error)
	SaveCanvas(context.Context, Actor, Canvas) (Canvas, error)
	ListCanvasVersions(context.Context, Actor, int64) ([]CanvasVersion, error)
	RestoreCanvas(context.Context, Actor, int64, int) (Canvas, error)
	CreateAttachment(context.Context, Actor, Attachment) (Attachment, error)
	ListAttachments(context.Context, Actor, int64) ([]Attachment, error)
	GetAttachment(context.Context, Actor, int64, int64) (Attachment, error)
	DeleteAttachment(context.Context, Actor, int64, int64) error
}

type ObjectStore interface {
	PutObjectFromFile(context.Context, string, string, string) error
	GetObject(context.Context, string, string) (io.ReadCloser, error)
	DeleteObject(context.Context, string, string) error
}

type Executor interface {
	Execute(context.Context, Execution, Project, Message, []int64, []int64) (string, string, error)
}
type UnavailableExecutor struct{}

func (UnavailableExecutor) Execute(context.Context, Execution, Project, Message, []int64, []int64) (string, string, error) {
	return "", "", ErrExecutorUnavailable
}
