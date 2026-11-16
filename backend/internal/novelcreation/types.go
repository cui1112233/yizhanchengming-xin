package novelcreation

import (
	"fmt"
	"time"
)

type RunStatus string

const (
	RunDraft                 RunStatus = "draft"
	RunClarifying            RunStatus = "clarifying"
	RunAwaitingConfirmation  RunStatus = "awaiting_confirmation"
	RunQueued                RunStatus = "queued"
	RunPlanning              RunStatus = "planning"
	RunWriting               RunStatus = "writing"
	RunNormalizing           RunStatus = "normalizing"
	RunAuditing              RunStatus = "auditing"
	RunRevising              RunStatus = "revising"
	RunSettling              RunStatus = "settling"
	RunIndexing              RunStatus = "indexing"
	RunSucceeded             RunStatus = "succeeded"
	RunSucceededWithWarnings RunStatus = "succeeded_with_warnings"
	RunFailed                RunStatus = "failed"
	RunCancelled             RunStatus = "cancelled"
)

const (
	ErrInputRequired       = "novel_input_required"
	ErrProjectNotFound     = "novel_project_not_found"
	ErrSessionNotFound     = "novel_session_not_found"
	ErrRunInProgress       = "novel_run_in_progress"
	ErrProjectLocked       = "novel_project_locked"
	ErrInvalidTransition   = "novel_invalid_transition"
	ErrModelUnavailable    = "novel_model_unavailable"
	ErrGenerationTimeout   = "novel_generation_timeout"
	ErrModelRateLimited    = "novel_model_rate_limited"
	ErrModelOverloaded     = "novel_model_overloaded"
	ErrModelAuthFailed     = "novel_model_auth_failed"
	ErrModelConnectionLost = "novel_model_connection_lost"
	ErrPersistenceFailed   = "novel_persistence_failed"
)

// Batch statuses for continuous chapter generation.
const (
	BatchRunning   = "running"
	BatchPaused    = "paused"
	BatchStopped   = "stopped"
	BatchCompleted = "completed"
	BatchFailed    = "failed"
)

type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code
	}
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type CreateProjectInput struct {
	Title              string
	Genre              string
	World              string
	Protagonist        string
	CoreConflict       string
	ModelID            string
	TargetChapterWords int
}

type SendMessageInput struct {
	SourceType     string `json:"sourceType,omitempty"`
	SourcePlanID   string `json:"sourcePlanId,omitempty"`
	SourceRound    string `json:"sourceRound,omitempty"`
	SourceRunID    string `json:"sourceRunId,omitempty"`
	SessionID      string
	ProjectID      string
	Content        string
	ModelID        string
	ThinkingLevel  string
	IdempotencyKey string
	// ModeInputs carries the mode-specific contract payload (reference text for
	// imitation, existing setting for fanfiction, previous text for
	// continuation). The service validates it against Session.Mode.
	ModeInputs map[string]string
	// KnowledgeIDs selects imported project materials for this run. The service
	// validates ownership and stores a bounded snapshot so retries stay stable.
	KnowledgeIDs   []string
	NoteIDs        []string
	StyleProfileID string
}

// KnowledgeMaterial is the bounded material snapshot injected into prompts.
type KnowledgeMaterial struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Material is a user-imported reference document attached to a project.
type Material struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateMaterialInput struct {
	Title   string
	Content string
}

// AgentSettings controls generation volume and revision rounds per project.
type AgentSettings struct {
	Version                 int               `json:"version"`
	ModelRoutes             map[string]string `json:"modelRoutes"`
	PassScore               float64           `json:"passScore"`
	BaseTemperature         float64           `json:"baseTemperature"`
	TemperatureStep         float64           `json:"temperatureStep"`
	EnableLengthNorm        bool              `json:"enableLengthNorm"`
	EnablePostWrite         bool              `json:"enablePostWrite"`
	EnableAiTells           bool              `json:"enableAiTells"`
	RequireSecondaryConfirm bool              `json:"requireSecondaryConfirm"`
	TargetChapterWords      int               `json:"targetChapterWords"`
	MaxRevisionRounds       int               `json:"maxRevisionRounds"`
	ThinkingLevel           string            `json:"thinkingLevel"`
}

// ProjectNote is a free-form memo bound to a project.
type ProjectNote struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProjectNoteInput struct {
	Title   string
	Content string
}

// Schedule is a recurring writing task.
type Schedule struct {
	AutomationState
	ConfigurationError string            `json:"configurationError,omitempty"`
	ID                 string            `json:"id"`
	OwnerUsername      string            `json:"-"`
	ProjectID          string            `json:"projectId"`
	SessionID          string            `json:"sessionId"`
	IntervalMinutes    int               `json:"intervalMinutes"`
	Enabled            bool              `json:"enabled"`
	Prompt             string            `json:"prompt"`
	ModelID            string            `json:"modelId"`
	ThinkingLevel      string            `json:"thinkingLevel"`
	NextRunAt          time.Time         `json:"nextRunAt"`
	LastRunAt          *time.Time        `json:"lastRunAt,omitempty"`
	LastRunID          string            `json:"lastRunId,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
	ModeInputs         map[string]string `json:"modeInputs"`
	KnowledgeIDs       []string          `json:"knowledgeIds"`
	NoteIDs            []string          `json:"noteIds"`
	StyleProfileID     string            `json:"styleProfileId"`
}

type ScheduleInput struct {
	DailyChapterCap        int
	MaxConsecutiveFailures int
	SessionID              string
	IntervalMinutes        int
	Enabled                bool
	Prompt                 string
	ModelID                string
	ThinkingLevel          string
	ModeInputs             map[string]string `json:"modeInputs"`
	KnowledgeIDs           []string          `json:"knowledgeIds"`
	NoteIDs                []string          `json:"noteIds"`
	StyleProfileID         string            `json:"styleProfileId"`
}

// Dashboard aggregates project progress for the data panel.
type Dashboard struct {
	ProjectID    string         `json:"projectId"`
	ChapterCount int            `json:"chapterCount"`
	TotalWords   int            `json:"totalWords"`
	RunCounts    map[string]int `json:"runCounts"`
	RunTotal     int            `json:"runTotal"`
	SuccessRate  float64        `json:"successRate"`
	LastRunAt    *time.Time     `json:"lastRunAt,omitempty"`
}

// Batch is a continuous chapter-generation task. Each chapter is a normal run,
// so auditing, revision, chapter storage and retry keep working unchanged.
type Batch struct {
	AutomationState
	ID                string            `json:"id"`
	OwnerUsername     string            `json:"-"`
	ProjectID         string            `json:"projectId"`
	SessionID         string            `json:"sessionId"`
	Status            string            `json:"status"`
	TotalChapters     int               `json:"totalChapters"`
	CompletedChapters int               `json:"completedChapters"`
	CurrentRunID      string            `json:"currentRunId,omitempty"`
	Attempt           int               `json:"-"`
	Prompt            string            `json:"prompt"`
	ModelID           string            `json:"modelId"`
	ThinkingLevel     string            `json:"thinkingLevel"`
	LastError         string            `json:"lastError,omitempty"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
	ModeInputs        map[string]string `json:"modeInputs"`
	KnowledgeIDs      []string          `json:"knowledgeIds"`
	NoteIDs           []string          `json:"noteIds"`
	StyleProfileID    string            `json:"styleProfileId"`
}

type BatchInput struct {
	MaxConsecutiveFailures int
	SessionID              string
	TotalChapters          int
	Prompt                 string
	ModelID                string
	ThinkingLevel          string
	ModeInputs             map[string]string `json:"modeInputs"`
	KnowledgeIDs           []string          `json:"knowledgeIds"`
	NoteIDs                []string          `json:"noteIds"`
	StyleProfileID         string            `json:"styleProfileId"`
}

// StyleProfile is an account-owned, reusable writing style. Import and analysis
// metadata can be added without changing the generation input contract.
type StyleProfile struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
