package video

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type MergeStatus string

const (
	MergeQueued    MergeStatus = "queued"
	MergeRunning   MergeStatus = "running"
	MergeSucceeded MergeStatus = "succeeded"
	MergeFailed    MergeStatus = "failed"
)

var (
	ErrMergeRetryNotAllowed = errors.New("video: merge retry not allowed")
	ErrMergeInputNotReady   = errors.New("video: merge input video is not ready")
	ErrMergeTimeout         = errors.New("video: merge execution timed out")
	ErrFFmpegUnavailable    = errors.New("video: ffmpeg unavailable")
)

type MergeInputAsset struct {
	ProductionTaskID int64  `json:"productionTaskId"`
	URL              string `json:"url"`
	Order            int    `json:"order"`
}

type MergeJob struct {
	ID             int64       `json:"id"`
	BatchProjectID int64       `json:"batchProjectId"`
	BookID         int64       `json:"bookId"`
	Status         MergeStatus `json:"status"`
	CurrentAttempt int         `json:"currentAttempt"`
	ErrorMessage   string      `json:"errorMessage,omitempty"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type MergeAttempt struct {
	ID              int64             `json:"id"`
	MergeJobID      int64             `json:"mergeJobId"`
	Attempt         int               `json:"attempt"`
	Status          MergeStatus       `json:"status"`
	Inputs          []MergeInputAsset `json:"inputs"`
	AspectRatio     string            `json:"aspectRatio"`
	Speed           float64           `json:"speed"`
	OutputBucket    string            `json:"outputBucket,omitempty"`
	OutputObjectKey string            `json:"outputObjectKey,omitempty"`
	OutputURL       string            `json:"outputUrl,omitempty"`
	ErrorMessage    string            `json:"errorMessage,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type MergeStartRequest struct {
	BatchProjectID int64             `json:"batchProjectId"`
	BookID         int64             `json:"bookId"`
	Inputs         []MergeInputAsset `json:"inputs"`
	AspectRatio    string            `json:"aspectRatio"`
	Speed          float64           `json:"speed"`
}

type MergeProductionStartRequest struct {
	BatchProjectID    int64   `json:"batchProjectId"`
	BookID            int64   `json:"bookId"`
	ProductionTaskIDs []int64 `json:"productionTaskIds"`
	AspectRatio       string  `json:"aspectRatio"`
	Speed             float64 `json:"speed"`
}

type MergeResult struct {
	Job     MergeJob     `json:"job"`
	Attempt MergeAttempt `json:"attempt"`
}

type MergeExecutionRequest struct {
	JobID       int64
	AttemptID   int64
	Inputs      []MergeInputAsset
	AspectRatio string
	Speed       float64
}

type MergeExecutor interface {
	Execute(context.Context, MergeExecutionRequest) (Artifact, error)
}

type MergeStore interface {
	CreateMergeJob(context.Context, MergeJob) (MergeJob, error)
	GetMergeJob(context.Context, int64) (MergeJob, error)
	UpdateMergeJob(context.Context, MergeJob) error
	CreateMergeAttempt(context.Context, MergeAttempt) (MergeAttempt, error)
	GetMergeAttempt(context.Context, int64) (MergeAttempt, error)
	UpdateMergeAttempt(context.Context, MergeAttempt) error
	ListMergeAttempts(context.Context, int64) ([]MergeAttempt, error)
}

type MergeInputResolver interface {
	ResolveSucceededMergeInputs(context.Context, int64, int64, []int64) ([]MergeInputAsset, error)
}

// MergeWorkCoordinator is only the domain consumption boundary for the future
// Task 9.4 shared queue/lease runtime. Task 14 intentionally provides no Redis,
// scheduler, or global queue implementation.
type MergeWorkCoordinator interface {
	ClaimMerge(context.Context) (MergeAttempt, error)
	RenewMerge(context.Context, int64, time.Time) error
	ReleaseMerge(context.Context, int64) error
	RequeueExpiredMerges(context.Context, time.Time, int) (int, error)
}

type MergeService struct {
	store    MergeStore
	executor MergeExecutor
	now      func() time.Time
}

func NewMergeService(store MergeStore, executor MergeExecutor) *MergeService {
	return &MergeService{store: store, executor: executor, now: time.Now}
}

func (s *MergeService) StartFromProductionTasks(ctx context.Context, req MergeProductionStartRequest) (MergeResult, error) {
	if s == nil || s.store == nil {
		return MergeResult{}, providerError(ErrorProviderUnavailable, "merge store unavailable", nil)
	}
	resolver, ok := s.store.(MergeInputResolver)
	if !ok {
		return MergeResult{}, providerError(ErrorProviderUnavailable, "merge input resolver unavailable", nil)
	}
	if req.BatchProjectID <= 0 || req.BookID <= 0 || len(req.ProductionTaskIDs) == 0 {
		return MergeResult{}, fmt.Errorf("video: merge project, book and production tasks are required")
	}
	inputs, err := resolver.ResolveSucceededMergeInputs(ctx, req.BatchProjectID, req.BookID, req.ProductionTaskIDs)
	if err != nil {
		return MergeResult{}, err
	}
	return s.Start(ctx, MergeStartRequest{
		BatchProjectID: req.BatchProjectID,
		BookID: req.BookID,
		Inputs: inputs,
		AspectRatio: req.AspectRatio,
		Speed: req.Speed,
	})
}

func (s *MergeService) Start(ctx context.Context, req MergeStartRequest) (MergeResult, error) {
	if s == nil || s.store == nil {
		return MergeResult{}, providerError(ErrorProviderUnavailable, "merge store unavailable", nil)
	}
	inputs, aspect, speed, err := normalizeMergeRequest(req.Inputs, req.AspectRatio, req.Speed)
	if err != nil {
		return MergeResult{}, err
	}
	if req.BatchProjectID <= 0 || req.BookID <= 0 {
		return MergeResult{}, fmt.Errorf("video: merge batch project and book are required")
	}
	now := s.now().UTC()
	job, err := s.store.CreateMergeJob(ctx, MergeJob{
		BatchProjectID: req.BatchProjectID, BookID: req.BookID, Status: MergeQueued,
		CurrentAttempt: 1, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return MergeResult{}, err
	}
	attempt, err := s.store.CreateMergeAttempt(ctx, MergeAttempt{
		MergeJobID: job.ID, Attempt: 1, Status: MergeQueued, Inputs: inputs,
		AspectRatio: aspect, Speed: speed, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return MergeResult{Job: job}, err
	}
	return MergeResult{Job: job, Attempt: attempt}, nil
}

func (s *MergeService) Get(ctx context.Context, jobID int64) (MergeJob, []MergeAttempt, error) {
	if s == nil || s.store == nil || jobID <= 0 {
		return MergeJob{}, nil, ErrNotFound
	}
	job, err := s.store.GetMergeJob(ctx, jobID)
	if err != nil {
		return MergeJob{}, nil, err
	}
	attempts, err := s.store.ListMergeAttempts(ctx, jobID)
	if err != nil {
		return MergeJob{}, nil, err
	}
	return job, attempts, nil
}

func (s *MergeService) RetryAttempt(ctx context.Context, attemptID int64) (MergeResult, error) {
	if s == nil || s.store == nil {
		return MergeResult{}, providerError(ErrorProviderUnavailable, "merge store unavailable", nil)
	}
	previous, err := s.store.GetMergeAttempt(ctx, attemptID)
	if err != nil {
		return MergeResult{}, err
	}
	job, err := s.store.GetMergeJob(ctx, previous.MergeJobID)
	if err != nil {
		return MergeResult{}, err
	}
	attempts, err := s.store.ListMergeAttempts(ctx, job.ID)
	if err != nil {
		return MergeResult{}, err
	}
	if len(attempts) == 0 || attempts[len(attempts)-1].ID != previous.ID || previous.Status != MergeFailed {
		return MergeResult{Job: job, Attempt: previous}, ErrMergeRetryNotAllowed
	}
	now := s.now().UTC()
	next, err := s.store.CreateMergeAttempt(ctx, MergeAttempt{
		MergeJobID: job.ID, Attempt: previous.Attempt + 1, Status: MergeQueued,
		Inputs: copyMergeInputs(previous.Inputs), AspectRatio: previous.AspectRatio, Speed: previous.Speed,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		return MergeResult{Job: job, Attempt: previous}, err
	}
	job.Status = MergeQueued
	job.CurrentAttempt = next.Attempt
	job.ErrorMessage = ""
	job.UpdatedAt = now
	if err := s.store.UpdateMergeJob(ctx, job); err != nil {
		return MergeResult{Job: job, Attempt: next}, err
	}
	return MergeResult{Job: job, Attempt: next}, nil
}

func (s *MergeService) ExecuteAttempt(ctx context.Context, attemptID int64) (MergeAttempt, error) {
	if s == nil || s.store == nil || s.executor == nil {
		return MergeAttempt{}, providerError(ErrorProviderUnavailable, "merge executor unavailable", nil)
	}
	attempt, err := s.store.GetMergeAttempt(ctx, attemptID)
	if err != nil {
		return MergeAttempt{}, err
	}
	if attempt.Status != MergeQueued {
		return attempt, fmt.Errorf("video: merge attempt is not queued")
	}
	job, err := s.store.GetMergeJob(ctx, attempt.MergeJobID)
	if err != nil {
		return attempt, err
	}
	now := s.now().UTC()
	attempt.Status = MergeRunning
	attempt.UpdatedAt = now
	job.Status = MergeRunning
	job.ErrorMessage = ""
	job.UpdatedAt = now
	if err := s.store.UpdateMergeAttempt(ctx, attempt); err != nil {
		return attempt, err
	}
	if err := s.store.UpdateMergeJob(ctx, job); err != nil {
		return attempt, err
	}

	artifact, execErr := s.executor.Execute(ctx, MergeExecutionRequest{
		JobID: job.ID, AttemptID: attempt.ID, Inputs: copyMergeInputs(attempt.Inputs),
		AspectRatio: attempt.AspectRatio, Speed: attempt.Speed,
	})
	now = s.now().UTC()
	if execErr != nil {
		attempt.Status = MergeFailed
		attempt.ErrorMessage = truncateMergeError(execErr.Error())
		attempt.UpdatedAt = now
		job.Status = MergeFailed
		job.ErrorMessage = attempt.ErrorMessage
		job.UpdatedAt = now
		_ = s.store.UpdateMergeAttempt(ctx, attempt)
		_ = s.store.UpdateMergeJob(ctx, job)
		return attempt, execErr
	}
	attempt.Status = MergeSucceeded
	attempt.OutputBucket = artifact.Bucket
	attempt.OutputObjectKey = artifact.ObjectKey
	attempt.OutputURL = artifact.URL
	attempt.ErrorMessage = ""
	attempt.UpdatedAt = now
	job.Status = MergeSucceeded
	job.ErrorMessage = ""
	job.UpdatedAt = now
	if err := s.store.UpdateMergeAttempt(ctx, attempt); err != nil {
		return attempt, err
	}
	if err := s.store.UpdateMergeJob(ctx, job); err != nil {
		return attempt, err
	}
	return attempt, nil
}

func normalizeMergeRequest(inputs []MergeInputAsset, aspect string, speed float64) ([]MergeInputAsset, string, float64, error) {
	if len(inputs) == 0 {
		return nil, "", 0, fmt.Errorf("video: merge inputs are required")
	}
	if speed == 0 {
		speed = 1
	}
	if speed < 0.5 || speed > 4 {
		return nil, "", 0, fmt.Errorf("video: merge speed must be between 0.5 and 4")
	}
	aspect = strings.TrimSpace(aspect)
	if aspect == "" {
		aspect = "9:16"
	}
	if _, _, err := mergeCanvas(aspect); err != nil {
		return nil, "", 0, err
	}
	out := copyMergeInputs(inputs)
	seenOrder := map[int]struct{}{}
	for i := range out {
		if out[i].ProductionTaskID <= 0 || out[i].Order <= 0 {
			return nil, "", 0, fmt.Errorf("video: merge input production task/order is required")
		}
		if _, exists := seenOrder[out[i].Order]; exists {
			return nil, "", 0, fmt.Errorf("video: merge input order must be unique")
		}
		seenOrder[out[i].Order] = struct{}{}
		parsed, err := url.Parse(strings.TrimSpace(out[i].URL))
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
			return nil, "", 0, fmt.Errorf("video: merge input URL must use https")
		}
		out[i].URL = parsed.String()
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return out, aspect, speed, nil
}

func copyMergeInputs(inputs []MergeInputAsset) []MergeInputAsset {
	return append([]MergeInputAsset(nil), inputs...)
}

func truncateMergeError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 1024 {
		message = message[:1024]
	}
	return message
}
