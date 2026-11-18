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
	LongImitationStore
	StudioStore
	ChapterImportStore
	ContentStore
	ContentManagementStore
	HasSavedRun(context.Context, string, string) (bool, error)
	InsertAutomationRound(context.Context, string, string, string, string, string, time.Time) error
	Clock(context.Context) (time.Time, error)
	Atomic(context.Context, string, string, func(Store) error) error
	GetExecution(context.Context, string, string) (Execution, bool, error)
	PutExecution(context.Context, Execution) error
	GetSlot(context.Context, string, string) (ProjectSlot, error)
	PutSlot(context.Context, string, string, ProjectSlot) error
	ListDispatch(context.Context, int) ([]Execution, error)
	GetStyleProfile(context.Context, string, string) (StyleProfile, error)
	ListStyleProfiles(context.Context, string) ([]StyleProfileSummary, error)
	SaveStyleProfile(context.Context, string, StyleProfile) (StyleProfile, error)
	CreateProject(context.Context, string, CreateProjectInput) (Project, error)
	GetProject(context.Context, string, string) (Project, error)
	ListProjects(context.Context, string) ([]Project, error)
	CreateSession(context.Context, string, CreateSessionInput) (Session, error)
	GetSession(context.Context, string, string) (Session, error)
	ListSessions(context.Context, string, string) ([]Session, error)
	AppendMessage(context.Context, string, Message) (Message, error)
	ListMessages(context.Context, string, string, int64) ([]Message, error)
	FindRunByIdempotencyKey(context.Context, string, string) (Run, bool, error)
	GetRun(context.Context, string, string) (Run, error)
	CreateRun(context.Context, string, Run) (Run, error)
	UpdateRun(context.Context, string, RunUpdate) (Run, error)
	AppendRunEvent(context.Context, string, RunEvent) (RunEvent, error)
	ListRunEvents(context.Context, string, string, int64) ([]RunEvent, error)
	ListChapters(context.Context, string, string) ([]Chapter, error)
	GetChapterMemory(context.Context, string, string, string) (Memory, bool, error)
	SaveChapterBundle(context.Context, string, ChapterBundle) error
	// GetProjectSettings returns the project-level style profile and creation
	// rules. A project without a saved row returns zero values, not an error.
	GetProjectSettings(context.Context, string, string) (ProjectSettings, error)
	UpdateProjectSettings(context.Context, string, string, ProjectSettings) (ProjectSettings, error)
	// GetFoundation returns the newest confirmed or generated blueprint for a
	// project so later chapters reuse it instead of regenerating the outline.
	GetFoundation(context.Context, string, string) (Foundation, bool, error)
	SaveFoundation(context.Context, string, Foundation) error
	// Materials are imported reference documents selected per run.
	ListMaterials(context.Context, string, string) ([]Material, error)
	CreateMaterial(context.Context, string, string, CreateMaterialInput) (Material, error)
	DeleteMaterial(context.Context, string, string, string) error
	// Notes are free-form project memos.
	ListNotes(context.Context, string, string) ([]ProjectNote, error)
	CreateNote(context.Context, string, string, ProjectNoteInput) (ProjectNote, error)
	UpdateNote(context.Context, string, string, string, ProjectNoteInput) (ProjectNote, error)
	DeleteNote(context.Context, string, string, string) error
	// ListRuns returns project runs newest first for dashboard aggregation.
	ListRuns(context.Context, string, string) ([]Run, error)
	// Schedules are recurring writing tasks.
	ListSchedules(context.Context, string, string) ([]Schedule, error)
	GetSchedule(context.Context, string, string) (Schedule, error)
	CreateSchedule(context.Context, string, Schedule) (Schedule, error)
	UpdateSchedule(context.Context, string, Schedule) (Schedule, error)
	DeleteSchedule(context.Context, string, string, string) error
	ListDueSchedules(context.Context, time.Time, int) ([]Schedule, error)
	ListDueSchedulePage(context.Context, time.Time, ScheduleCursor, int) ([]Schedule, error)
	// Batches drive continuous multi-chapter generation.
	ListBatches(context.Context, string, string) ([]Batch, error)
	GetBatch(context.Context, string, string) (Batch, error)
	CreateBatch(context.Context, string, Batch) (Batch, error)
	UpdateBatch(context.Context, string, Batch) (Batch, error)
	ListActiveBatches(context.Context, int) ([]Batch, error)
}

// ProjectSettings holds the project-level writing style and hard rules that are
// injected into every generation call.
type ProjectSettings struct {
	ProjectID    string        `json:"projectId"`
	StyleProfile string        `json:"styleProfile"`
	Rules        []string      `json:"rules"`
	Agent        AgentSettings `json:"agent"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

// TxStore marks a Store that participates in a caller-managed transaction.
// The concrete MySQL implementation is introduced by the following task.
type TxStore interface{ Store }

type Project struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Genre              string     `json:"genre"`
	World              string     `json:"world"`
	Protagonist        string     `json:"protagonist"`
	CoreConflict       string     `json:"coreConflict"`
	ModelID            string     `json:"modelId"`
	TargetChapterWords int        `json:"targetChapterWords"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt,omitempty"`
}

type CreateSessionInput struct {
	ProjectID     string
	Mode          string
	Title         string
	ContextBudget int
}

type Session struct {
	ArchivedAt    *time.Time `json:"archivedAt,omitempty"`
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	Mode          string     `json:"mode"`
	Title         string     `json:"title"`
	ContextBudget int        `json:"contextBudget"`
	MessageCursor int64      `json:"messageCursor"`
	CurrentRunID  string     `json:"currentRunId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type Message struct {
	ID            string          `json:"id"`
	SessionID     string          `json:"sessionId"`
	ProjectID     string          `json:"projectId"`
	Role          string          `json:"role"`
	Content       string          `json:"content"`
	Status        string          `json:"status"`
	ModelID       string          `json:"modelId"`
	ThinkingLevel string          `json:"thinkingLevel"`
	References    json.RawMessage `json:"references,omitempty"`
	TokenUsage    json.RawMessage `json:"tokenUsage,omitempty"`
	Sequence      int64           `json:"sequence"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type Run struct {
	ID              string          `json:"id"`
	ProjectID       string          `json:"projectId"`
	SessionID       string          `json:"sessionId"`
	IdempotencyKey  string          `json:"idempotencyKey"`
	Status          RunStatus       `json:"status"`
	ModelID         string          `json:"modelId"`
	ThinkingLevel   string          `json:"thinkingLevel"`
	PromptSnapshot  json.RawMessage `json:"-"`
	RuntimeSnapshot json.RawMessage `json:"runtimeSnapshot,omitempty"`
	ErrorCode       string          `json:"errorCode,omitempty"`
	ErrorMessage    string          `json:"errorMessage,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	FinishedAt      *time.Time      `json:"finishedAt,omitempty"`
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
	EventID   int64           `json:"eventId"`
	RunID     string          `json:"runId"`
	SessionID string          `json:"sessionId"`
	ProjectID string          `json:"projectId"`
	Type      string          `json:"type"`
	Step      string          `json:"step"`
	Status    RunStatus       `json:"status"`
	Message   string          `json:"message"`
	Progress  int             `json:"progress"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
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
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	VolumeID    string    `json:"volumeId,omitempty"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	WordCount   int       `json:"wordCount"`
	AuditStatus string    `json:"auditStatus"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
	ChapterVersion int
	PreviousTruth  json.RawMessage
	ID             string
	ProjectID      string
	ChapterID      string
	Snapshot       json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
