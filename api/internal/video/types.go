package video

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	ProviderPersonalAPI        = "personal_api"
	ProviderYFAISeedance       = "yfai_seedance"
	ProviderAutoDLComfyUI      = "autodl_comfyui"
	ProviderDoubaoLocalExecutor = "doubao_local_executor"

	ModelYD20Mini = "yd2.0-mini"
)

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
	JobCancelled JobStatus = "cancelled"
)

type ErrorCode string

const (
	ErrorProviderUnconfigured       ErrorCode = "provider_unconfigured"
	ErrorProviderUnavailable        ErrorCode = "provider_unavailable"
	ErrorProviderAuthFailed         ErrorCode = "provider_auth_failed"
	ErrorProviderRequestFailed      ErrorCode = "provider_request_failed"
	ErrorProviderInvalidResponse    ErrorCode = "provider_invalid_response"
	ErrorProviderCancelUnsupported  ErrorCode = "provider_cancel_unsupported"
	ErrorProviderConfigDecryptFailed ErrorCode = "provider_config_decrypt_failed"
)

type ProviderError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return fmt.Sprintf("video: %s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("video: %s", e.Code)
}

func (e *ProviderError) Unwrap() error { return e.Err }

func providerError(code ErrorCode, message string, err error) error {
	return &ProviderError{Code: code, Message: message, Err: err}
}

var ErrNotFound = errors.New("video: not found")

type ProviderConfig struct {
	ID              int64  `json:"id"`
	ProviderKey     string `json:"providerKey"`
	Model           string `json:"model"`
	CreateURL       string `json:"createUrl,omitempty"`
	TasksURL        string `json:"tasksUrl,omitempty"`
	ResultURL       string `json:"resultUrl,omitempty"`
	EncryptedSecret []byte `json:"-"`
	SecretNonce     []byte `json:"-"`
	Enabled         bool   `json:"enabled"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}

type ProviderConfigView struct {
	ID          int64  `json:"id"`
	ProviderKey string `json:"providerKey"`
	Model       string `json:"model"`
	CreateURL   string `json:"createUrl,omitempty"`
	TasksURL    string `json:"tasksUrl,omitempty"`
	ResultURL   string `json:"resultUrl,omitempty"`
	Enabled     bool   `json:"enabled"`
	Configured  bool   `json:"configured"`
}

func (c ProviderConfig) View() ProviderConfigView {
	return ProviderConfigView{
		ID: c.ID, ProviderKey: c.ProviderKey, Model: c.Model,
		CreateURL: c.CreateURL, TasksURL: c.TasksURL, ResultURL: c.ResultURL,
		Enabled: c.Enabled, Configured: len(c.EncryptedSecret) > 0 && len(c.SecretNonce) > 0,
	}
}

type SubmitRequest struct {
	Model              string
	Prompt             string
	RequestID          string
	DurationSeconds    int
	AspectRatio        string
	Resolution         string
	ReferenceImageURLs []string
}

type SubmitResult struct {
	ProviderJobID string
	Status        TaskStatus
	ArtifactURL   string
}

type PollResult struct {
	Status       TaskStatus
	ArtifactURL  string
	ErrorCode    ErrorCode
	ErrorMessage string
}

type CancelResult struct {
	Accepted bool
	Status   TaskStatus
}

type Provider interface {
	Submit(context.Context, SubmitRequest) (SubmitResult, error)
	Poll(context.Context, string) (PollResult, error)
	Cancel(context.Context, string) (CancelResult, error)
}

type ProviderFactory interface {
	Build(ProviderConfig, string) (Provider, error)
}

type FinalPrompt struct {
	StageRunID     int64
	PromptVersion  int
	InputRevision  string
	Text           string
}

type FinalPromptSource interface {
	ResolveFinalPrompt(context.Context, int64, int64) (FinalPrompt, error)
}

type Artifact struct {
	Bucket    string
	ObjectKey string
	URL       string
}

type ArtifactStore interface {
	Persist(context.Context, string, string) (Artifact, error)
}

type ProductionJob struct {
	ID                    int64
	BatchProjectID        int64
	BookID                 int64
	Status                 JobStatus
	InputRevision          string
	FinalPromptStageRunID  int64
	FinalPromptVersion     int
	Provider               string
	Model                  string
	IdempotencyKey         string
	ErrorCode              ErrorCode
	ErrorMessage           string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type ProductionTask struct {
	ID               int64
	ProductionJobID  int64
	Attempt          int
	Provider         string
	Model            string
	RequestID        string
	ProviderJobID    string
	Status           TaskStatus
	ErrorCode        ErrorCode
	ErrorMessage     string
	ArtifactSourceURL string
	OutputBucket     string
	OutputObjectKey  string
	OutputURL        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type StartRequest struct {
	BatchProjectID int64
	BookID         int64
	Provider       string
	Model          string
	RequestID      string
}

type StartResult struct {
	Job  ProductionJob
	Task ProductionTask
}

type Store interface {
	GetProviderConfig(context.Context, string, string) (ProviderConfig, error)
	CreateOrGetProductionJob(context.Context, ProductionJob) (ProductionJob, bool, error)
	CreateProductionTask(context.Context, ProductionTask) (ProductionTask, error)
	UpdateProductionTask(context.Context, ProductionTask) error
	UpdateProductionJob(context.Context, ProductionJob) error
	GetProductionTask(context.Context, int64) (ProductionTask, error)
	GetProductionJob(context.Context, int64) (ProductionJob, error)
	LatestTaskForJob(context.Context, int64) (ProductionTask, error)
	ListRecoverableTasks(context.Context, int) ([]ProductionTask, error)
}
