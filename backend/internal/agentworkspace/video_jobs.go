package agentworkspace

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrVideoJobNotFound    = errors.New("agent video job not found")
	ErrVideoJobExists      = errors.New("agent video job already exists")
	ErrVideoJobInvalid     = errors.New("invalid agent video job")
	ErrVideoJobState       = errors.New("invalid agent video job state")
	ErrVideoJobNodeChanged = errors.New("agent video node changed")
)

type VideoJobState string

const (
	VideoJobPreparing            VideoJobState = "preparing"
	VideoJobQueued               VideoJobState = "queued"
	VideoJobRunning              VideoJobState = "running"
	VideoJobResultReady          VideoJobState = "result_ready"
	VideoJobAwaitingConfirmation VideoJobState = "awaiting_confirmation"
	VideoJobApplied              VideoJobState = "applied"
	VideoJobFailed               VideoJobState = "failed"
	VideoJobCanceled             VideoJobState = "canceled"
	VideoJobAbandoned            VideoJobState = "abandoned"
)

type VideoJob struct {
	ID                string         `json:"id"`
	ProjectID         string         `json:"projectId"`
	OwnerUsername     string         `json:"-"`
	NodeID            string         `json:"nodeId"`
	NodeVersion       int64          `json:"nodeVersion"`
	State             VideoJobState  `json:"state"`
	ModelKey          string         `json:"modelKey"`
	ProviderTaskID    string         `json:"providerTaskId,omitempty"`
	Settings          map[string]any `json:"settings"`
	ReferenceAssetIDs []string       `json:"referenceAssetIds"`
	ResultAssetID     string         `json:"resultAssetId,omitempty"`
	ErrorMessage      string         `json:"errorMessage,omitempty"`
	IdempotencyKey    string         `json:"idempotencyKey,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

type CreateVideoJobInput struct {
	ID                string
	ProjectID         string
	NodeID            string
	NodeVersion       int64
	ModelKey          string
	Settings          map[string]any
	ReferenceAssetIDs []string
	IdempotencyKey    string
	Now               time.Time
}

type RetryVideoJobInput struct {
	ID             string
	IdempotencyKey string
	Now            time.Time
}

type VideoJobStore interface {
	CreateVideoJob(context.Context, string, CreateVideoJobInput) (VideoJob, error)
	VideoJob(context.Context, string, string, string) (VideoJob, error)
	ListVideoJobs(context.Context, string, string) ([]VideoJob, error)
	UpdateVideoJob(context.Context, string, VideoJob, time.Time) (VideoJob, error)
	ConfirmVideoJob(context.Context, string, string, string, time.Time) (VideoJob, error)
	RetryVideoJob(context.Context, string, string, string, RetryVideoJobInput) (VideoJob, error)
}

func validVideoJobState(state VideoJobState) bool {
	switch state {
	case VideoJobPreparing, VideoJobQueued, VideoJobRunning, VideoJobResultReady,
		VideoJobAwaitingConfirmation, VideoJobApplied, VideoJobFailed, VideoJobCanceled, VideoJobAbandoned:
		return true
	default:
		return false
	}
}

func normalizeCreateVideoJob(input CreateVideoJobInput) (CreateVideoJobInput, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.NodeID = strings.TrimSpace(input.NodeID)
	input.ModelKey = strings.TrimSpace(input.ModelKey)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.ID == "" || input.ProjectID == "" || input.NodeID == "" || input.NodeVersion < 1 || input.ModelKey == "" || input.IdempotencyKey == "" {
		return CreateVideoJobInput{}, ErrVideoJobInvalid
	}
	if input.Settings == nil {
		input.Settings = map[string]any{}
	}
	input.ReferenceAssetIDs = uniqueNonEmpty(input.ReferenceAssetIDs)
	if input.Now.IsZero() {
		return CreateVideoJobInput{}, ErrVideoJobInvalid
	}
	return input, nil
}

func uniqueNonEmpty(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func isRetryableVideoJob(state VideoJobState) bool {
	return state == VideoJobFailed || state == VideoJobCanceled || state == VideoJobAbandoned
}
