package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Store is the durable boundary for the novel-creation domain. Production
// implementations persist through MySQL; MemoryStore exists for unit tests.
type Store interface {
	CreateProject(context.Context, string, CreateProjectInput) (Project, error)
	GetProject(context.Context, string, string) (Project, error)
	ListProjects(context.Context, string) ([]Project, error)
	CreateSession(context.Context, string, CreateSessionInput) (Session, error)
	GetSession(context.Context, string, string) (Session, error)
	ListSessions(context.Context, string, string) ([]Session, error)
	AppendMessage(context.Context, string, Message) (Message, error)
	ListMessages(context.Context, string, string, int64) ([]Message, error)
	FindRunByIdempotencyKey(context.Context, string, string) (Run, bool, error)
	CreateRun(context.Context, string, Run) (Run, error)
	UpdateRun(context.Context, string, RunUpdate) (Run, error)
	AppendRunEvent(context.Context, string, RunEvent) (RunEvent, error)
	ListRunEvents(context.Context, string, string, int64) ([]RunEvent, error)
	SaveChapterBundle(context.Context, string, ChapterBundle) error
}

// TxStore marks a Store that participates in a caller-managed transaction.
// The concrete MySQL implementation is introduced by the following task.
type TxStore interface{ Store }

type Project struct {
	ID                 string
	Title              string
	Genre              string
	World              string
	Protagonist        string
	CoreConflict       string
	ModelID            string
	TargetChapterWords int
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

type CreateSessionInput struct {
	ProjectID     string
	Mode          string
	Title         string
	ContextBudget int
}

type Session struct {
	ID            string
	ProjectID     string
	Mode          string
	Title         string
	ContextBudget int
	MessageCursor int64
	CurrentRunID  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Message struct {
	ID            string
	SessionID     string
	ProjectID     string
	Role          string
	Content       string
	Status        string
	ModelID       string
	ThinkingLevel string
	References    json.RawMessage
	TokenUsage    json.RawMessage
	Sequence      int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Run struct {
	ID              string
	ProjectID       string
	SessionID       string
	IdempotencyKey  string
	Status          RunStatus
	ModelID         string
	ThinkingLevel   string
	PromptSnapshot  json.RawMessage
	RuntimeSnapshot json.RawMessage
	ErrorCode       string
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      *time.Time
}

type RunUpdate struct {
	ID              string
	Status          RunStatus
	RuntimeSnapshot json.RawMessage
	ErrorCode       string
	ErrorMessage    string
	FinishedAt      *time.Time
}

type RunEvent struct {
	EventID   int64
	RunID     string
	SessionID string
	ProjectID string
	Type      string
	Step      string
	Status    RunStatus
	Message   string
	Progress  int
	Payload   json.RawMessage
	CreatedAt time.Time
}

type Volume struct {
	ID        string
	ProjectID string
	Number    int
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Chapter struct {
	ID          string
	ProjectID   string
	VolumeID    string
	Number      int
	Title       string
	Content     string
	WordCount   int
	AuditStatus string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChapterVersion struct {
	ID        string
	ChapterID string
	RunID     string
	Version   int
	Content   string
	Snapshot  json.RawMessage
	CreatedAt time.Time
}

type Foundation struct {
	ID        string
	ProjectID string
	Snapshot  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Memory struct {
	ID        string
	ProjectID string
	ChapterID string
	Snapshot  json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Audit struct {
	ID        string
	ProjectID string
	ChapterID string
	RunID     string
	Score     float64
	Snapshot  json.RawMessage
	CreatedAt time.Time
}

type Revision struct {
	ID        string
	ProjectID string
	ChapterID string
	RunID     string
	Strategy  string
	Content   string
	Snapshot  json.RawMessage
	CreatedAt time.Time
}

type ChapterBundle struct {
	ProjectID       string
	RunID           string
	Volume          Volume
	Chapter         Chapter
	ChapterVersion  ChapterVersion
	Foundation      Foundation
	Memories        []Memory
	Audits          []Audit
	Revisions       []Revision
	FinalRunStatus  RunStatus
	RuntimeSnapshot json.RawMessage
}

func IsCode(err error, code string) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr) && domainErr.Code == code
}
