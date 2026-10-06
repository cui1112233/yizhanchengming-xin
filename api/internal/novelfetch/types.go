package novelfetch

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusPending         Status = "pending"
	StatusQueued          Status = "queued"
	StatusScheduled       Status = "scheduled"
	StatusRunning         Status = "running"
	StatusSucceeded       Status = "succeeded"
	StatusPartialFailed   Status = "partial_failed"
	StatusRetryableFailed Status = "retryable_failed"
	StatusFailed          Status = "failed"
	StatusBlocked         Status = "blocked"
)

var (
	ErrNotFound                = errors.New("novel fetch resource not found")
	ErrInvalid                 = errors.New("invalid novel fetch input")
	ErrNotDue                  = errors.New("novel fetch run is not due")
	ErrRuntimeUnavailable      = errors.New("novel fetch runtime unavailable")
	ErrModelUnavailable        = errors.New("novel fetch text model unavailable")
	ErrPublishUnavailable      = errors.New("novel fetch publish boundary unavailable")
	ErrHandoffUnavailable      = errors.New("novel fetch batch factory boundary unavailable")
	ErrMultipleRunsUnsupported = errors.New("multiple novel fetch runs per batch are not supported before shared persistence integration")
)

type Metadata struct {
	Category string `json:"category,omitempty"`
	Genre    string `json:"genre,omitempty"`
	Gender   string `json:"gender,omitempty"`
	Style    string `json:"style,omitempty"`
}

type Book struct {
	Key            string            `json:"key"`
	BatchID        string            `json:"batchId"`
	Source         string            `json:"source"`
	PlatformID     string            `json:"platformId"`
	ExternalBookID string            `json:"bookId"`
	Title          string            `json:"title"`
	OriginalRaw    string            `json:"-"`
	ProcessedText  string            `json:"processedText,omitempty"`
	OriginalChars  int               `json:"originalChars"`
	ProcessedChars int               `json:"processedChars"`
	Metadata       Metadata          `json:"metadata"`
	Versions       map[string]string `json:"versions,omitempty"`
	Status         Status            `json:"status"`
	CurrentStage   string            `json:"currentStage,omitempty"`
	Error          string            `json:"error,omitempty"`
}

type Batch struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BookKeys  []string  `json:"bookKeys"`
	CreatedAt time.Time `json:"createdAt"`
}

type Config struct {
	TextModelID           string            `json:"textModelId"`
	MaxText               int               `json:"maxText"`
	TargetVersions        []string          `json:"targetVersions"`
	RewriteProfiles       map[string]string `json:"rewriteProfiles,omitempty"`
	SensitiveReplacements map[string]string `json:"sensitiveReplacements,omitempty"`
	ChapterRemovePrefixes []string          `json:"chapterRemovePrefixes,omitempty"`
	TrimLines             bool              `json:"trimLines"`
	DropBlankLines        bool              `json:"dropBlankLines"`
}

type Run struct {
	ID             string    `json:"id"`
	BatchID        string    `json:"batchId"`
	Status         Status    `json:"status"`
	RunAt          time.Time `json:"runAt"`
	ConfigSnapshot Config    `json:"configSnapshot"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Record struct {
	RunID      string    `json:"runId"`
	BookKey    string    `json:"bookKey,omitempty"`
	Stage      string    `json:"stage"`
	Attempt    int       `json:"attempt"`
	Status     Status    `json:"status"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
}

type HistoryItem struct {
	Batch Batch `json:"batch"`
	Run   Run   `json:"run"`
}

type KnowledgeEntry struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Enabled bool   `json:"enabled"`
}

type BookInput struct {
	BookID string `json:"bookId"`
	Title  string `json:"title,omitempty"`
}

type GroupInput struct {
	Source     string      `json:"source"`
	PlatformID string      `json:"platformId"`
	Books      []BookInput `json:"books"`
}

type CreateBatchInput struct {
	Name   string       `json:"name"`
	Groups []GroupInput `json:"groups"`
}

type FetchRequest struct {
	Source     string
	PlatformID string
	BookID     string
}

type FetchResult struct {
	OriginalRaw string
	Title       string
	Category    string
	Genre       string
	Gender      string
	Style       string
}

type RewriteRequest struct {
	Book        Book
	Version     string
	Profile     string
	TextModelID string
	Knowledge  []KnowledgeEntry
}

type RunDispatch struct {
	RunID       string
	AvailableAt time.Time
}

type RetryDispatch struct {
	RunID       string
	BookKey     string
	AvailableAt time.Time
}

type HandoffRequest struct {
	Batch Batch
	Run   Run
	Books []Book
}

type HandoffResult struct {
	BatchProjectID int64 `json:"batchProjectId"`
}

type SubmitIntentRequest struct {
	BatchID             string
	RunID               string
	BookKey             string
	Version             string
	PublishingAccountID int64
}

type SubmitIntentResult struct {
	IntentID int64  `json:"intentId"`
	Status   string `json:"status"`
}

type Store interface {
	CreateBatch(context.Context, Batch) (Batch, error)
	GetBatch(context.Context, string) (Batch, error)
	ListBatches(context.Context) ([]Batch, error)
	PutBook(context.Context, Book) error
	UpdateBook(context.Context, Book) error
	GetBook(context.Context, string, string) (Book, error)
	ListBooks(context.Context, string) ([]Book, error)
	CreateRun(context.Context, Run) (Run, error)
	UpdateRun(context.Context, Run) error
	GetRun(context.Context, string) (Run, error)
	ListRuns(context.Context, string) ([]Run, error)
	AppendRecord(context.Context, Record) error
	ListRecords(context.Context, string) ([]Record, error)
	GetConfig(context.Context) (Config, error)
	SaveConfig(context.Context, Config) (Config, error)
	ListKnowledge(context.Context, string) ([]KnowledgeEntry, error)
	UpsertKnowledge(context.Context, KnowledgeEntry) (KnowledgeEntry, error)
	DeleteKnowledge(context.Context, string, string) error
}

type Fetcher interface {
	Fetch(context.Context, FetchRequest) (FetchResult, error)
}

type TextModel interface {
	Rewrite(context.Context, RewriteRequest) (string, error)
}

type Dispatcher interface {
	EnqueueRun(context.Context, RunDispatch) error
	EnqueueBookRetry(context.Context, RetryDispatch) error
}

type BatchFactoryBoundary interface {
	CreateFromNovelFetch(context.Context, HandoffRequest) (HandoffResult, error)
}

type PublishIntentBoundary interface {
	CreateIntent(context.Context, SubmitIntentRequest) (SubmitIntentResult, error)
}
