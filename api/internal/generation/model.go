package generation

import (
	"errors"
	"time"
)

var (
	ErrNotFound                 = errors.New("generation: not found")
	ErrConflict                 = errors.New("generation: conflict")
	ErrInvalid                  = errors.New("generation: invalid")
	ErrUnavailable              = errors.New("generation: unavailable")
	ErrAudioProbeUnavailable    = errors.New("audio_probe_unavailable")
	ErrAudioMeasurementRequired = errors.New("audio_measurement_required")
)

type Stage string

const (
	StageScript      Stage = "SCRIPT"
	StageHook        Stage = "HOOK"
	StageDirector    Stage = "DIRECTOR"
	StageFinalPrompt Stage = "FINAL_PROMPT"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusSkipped   Status = "skipped"
)

type DirectorMode string

const (
	DirectorNormal DirectorMode = "normal"
	DirectorH3     DirectorMode = "h3"
)

const (
	PromptScript         = "script.default"
	PromptScriptPlotMode = "script.plot_mode"
	PromptHook           = "hook.default"
	PromptDirector       = "director.default"
	PromptDirectorH3     = "director.h3"
	PromptFinal          = "final_prompt.default"
	PromptKeyScript      = PromptScript
)

type Prompt struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BookRun struct {
	ID             int64      `json:"id"`
	BatchProjectID int64      `json:"batchProjectId"`
	BookID         int64      `json:"bookId"`
	Status         Status     `json:"status"`
	RequestID      string     `json:"requestId,omitempty"`
	ErrorMessage   string     `json:"errorMessage,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	FinishedAt     *time.Time `json:"finishedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type StageRun struct {
	ID               int64      `json:"id"`
	BookRunID        int64      `json:"bookRunId"`
	BookID           int64      `json:"bookId"`
	Stage            Stage      `json:"stage"`
	Status           Status     `json:"status"`
	Attempt          int        `json:"attempt"`
	RequestID        string     `json:"requestId,omitempty"`
	PromptKey        string     `json:"promptKey,omitempty"`
	PromptVersion    int        `json:"promptVersion,omitempty"`
	InputSnapshot    string     `json:"inputSnapshot,omitempty"`
	OutputText       string     `json:"outputText,omitempty"`
	ErrorMessage     string     `json:"errorMessage,omitempty"`
	ValidationResult string     `json:"validationResult,omitempty"`
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	FinishedAt       *time.Time `json:"finishedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type AudioMeasurement struct {
	ID             int64     `json:"id"`
	BatchProjectID int64     `json:"batchProjectId"`
	BookID         int64     `json:"bookId"`
	AudioAsset     string    `json:"audioAsset"`
	AssetHash      string    `json:"-"`
	DurationMS     int64     `json:"durationMs"`
	MeasuredAt     time.Time `json:"measuredAt"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AudioMeasurementRequest struct {
	BatchProjectID int64  `json:"batchProjectId"`
	BookID         int64  `json:"bookId"`
	AudioAsset     string `json:"audioAsset"`
}

type TextRequest struct {
	BookID               int64        `json:"bookId"`
	Stage                Stage        `json:"stage"`
	SystemPrompt         string       `json:"systemPrompt"`
	UserPrompt           string       `json:"userPrompt"`
	DirectorMode         DirectorMode `json:"directorMode"`
	MatchAudio           bool         `json:"matchAudio"`
	AudioDurationSec     float64      `json:"audioDurationSec"`
	ShotDurationLimitSec int64        `json:"shotDurationLimitSec,omitempty"`
}

type RunBookRequest struct {
	BatchProjectID       int64        `json:"batchProjectId"`
	BookID               int64        `json:"bookId"`
	HookEnabled          bool         `json:"hookEnabled"`
	PlotMode             bool         `json:"plotMode"`
	DirectorMode         DirectorMode `json:"directorMode"`
	MatchAudio           bool         `json:"matchAudio,omitempty"`
	AudioDurationSec     float64      `json:"audioDurationSec,omitempty"`
	ShotDurationLimitSec int64        `json:"shotDurationLimitSec,omitempty"`
	RequestID            string       `json:"requestId"`
	ProcessingRules      string       `json:"processingRules,omitempty"`
	KnowledgeBase        string       `json:"knowledgeBase,omitempty"`
	ProjectConfig        string       `json:"projectConfig,omitempty"`
	UserConfig           string       `json:"userConfig,omitempty"`
	ModelConfig          string       `json:"modelConfig,omitempty"`
}

type RetryStageRequest struct {
	BatchProjectID int64  `json:"batchProjectId"`
	BookID         int64  `json:"bookId"`
	Stage          Stage  `json:"stage"`
	RequestID      string `json:"requestId"`
}

type RunBatchRequest struct {
	BatchProjectID       int64        `json:"batchProjectId"`
	HookEnabled          bool         `json:"hookEnabled"`
	PlotMode             bool         `json:"plotMode"`
	DirectorMode         DirectorMode `json:"directorMode"`
	MatchAudio           bool         `json:"matchAudio,omitempty"`
	AudioDurationSec     float64      `json:"audioDurationSec,omitempty"`
	ShotDurationLimitSec int64        `json:"shotDurationLimitSec,omitempty"`
	RequestID            string       `json:"requestId"`
}

type BookGenerationResult struct {
	Run    BookRun            `json:"run"`
	Stages []StageRun         `json:"stages"`
	Latest map[Stage]StageRun `json:"latest,omitempty"`
	Error  string             `json:"error,omitempty"`
}

type BatchBookResult struct {
	BookID int64    `json:"bookId"`
	Run    *BookRun `json:"run,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type BatchGenerationResult struct {
	BatchProjectID int64             `json:"batchProjectId"`
	Books          []BatchBookResult `json:"books"`
	Completed      int               `json:"completed"`
	Failed         int               `json:"failed"`
}

type ProjectSummary struct {
	BatchProjectID int64                   `json:"batchProjectId"`
	Books          []BookGenerationSummary `json:"books"`
	Pending        int                     `json:"pending"`
	Running        int                     `json:"running"`
	Completed      int                     `json:"completed"`
	Failed         int                     `json:"failed"`
}

type BookGenerationSummary struct {
	BookID int64              `json:"bookId"`
	Title  string             `json:"title,omitempty"`
	Run    *BookRun           `json:"run,omitempty"`
	Stages map[Stage]StageRun `json:"stages"`
}

type FinalPromptInput struct {
	SystemPreset    string
	Script          string
	Hook            string
	Director        string
	ProcessingRules string
	KnowledgeBase   string
	ProjectConfig   string
	UserConfig      string
	ModelConfig     string
}
