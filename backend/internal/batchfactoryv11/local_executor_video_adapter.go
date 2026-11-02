package batchfactoryv11

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/localexecutor"
)

type LocalVideoJobInput struct {
	SourceTaskID       string
	BatchID            string
	BookID             string
	VideoID            string
	Model              string
	Prompt             string
	Duration           int
	AspectRatio        string
	Resolution         string
	ReferenceImageURLs []string
}

type LocalVideoJob struct {
	ID           string
	State        string
	ArtifactID   string
	ErrorMessage string
}

type LocalVideoJobClient interface {
	CreateVideoJob(context.Context, string, LocalVideoJobInput) (LocalVideoJob, error)
	GetVideoJob(context.Context, string, string) (LocalVideoJob, error)
}

type LocalVideoJobCanceler interface {
	CancelVideoJob(context.Context, string, string) (LocalVideoJob, error)
}

type LocalVideoExecutorAvailability interface {
	HasOnlineVideoExecutor(context.Context, string) (bool, error)
}

type LocalExecutorVideoAdapter struct {
	Client        LocalVideoJobClient
	PublicBaseURL string
	// ArtifactSecret enables short-lived owner-bound URLs that remote merge
	// providers can fetch without a user cookie.
	ArtifactSecret   string
	ArtifactTokenTTL time.Duration
	Now              func() time.Time
}

func NewLocalExecutorVideoAdapter(client LocalVideoJobClient, publicBaseURL string) *LocalExecutorVideoAdapter {
	return &LocalExecutorVideoAdapter{Client: client, PublicBaseURL: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")}
}

func (a *LocalExecutorVideoAdapter) EnsureAvailable(ctx context.Context, owner string) error {
	if a == nil || a.Client == nil {
		return fmt.Errorf("%w: local executor is unavailable", ErrUnavailable)
	}
	if availability, ok := a.Client.(LocalVideoExecutorAvailability); ok {
		online, err := availability.HasOnlineVideoExecutor(ctx, owner)
		if err != nil {
			return err
		}
		if !online {
			return fmt.Errorf("%w: no online Doubao local executor is paired", ErrUnavailable)
		}
	}
	return nil
}

func (a *LocalExecutorVideoAdapter) Submit(ctx context.Context, owner string, input LocalVideoJobInput) (ProviderTaskRef, error) {
	if a == nil || a.Client == nil {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor is unavailable", ErrUnavailable)
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor task identity is required", ErrInvalid)
	}
	if strings.TrimSpace(input.SourceTaskID) == "" {
		var ok bool
		input.SourceTaskID, ok = derivedLocalSourceTaskID(input)
		if !ok {
			return ProviderTaskRef{}, fmt.Errorf("%w: local executor task identity is required", ErrInvalid)
		}
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return ProviderTaskRef{}, fmt.Errorf("%w: compiled prompt is required", ErrInvalid)
	}
	job, err := a.Client.CreateVideoJob(ctx, owner, input)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	if strings.TrimSpace(job.ID) == "" {
		return ProviderTaskRef{}, fmt.Errorf("local executor did not return a job id")
	}
	return ProviderTaskRef{ProviderTaskID: job.ID, State: mapLocalVideoState(job.State), RequestedDurationSeconds: float64(input.Duration)}, nil
}

// Cancel forwards only to the owned local-executor job. It never changes a V11
// task state before the executor accepts the cancellation request.
func (a *LocalExecutorVideoAdapter) Cancel(ctx context.Context, owner, id string) (ProviderTaskRef, error) {
	if a == nil || a.Client == nil {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor is unavailable", ErrUnavailable)
	}
	canceler, ok := a.Client.(LocalVideoJobCanceler)
	if !ok {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor cancellation is unavailable", ErrUnavailable)
	}
	owner, id = strings.TrimSpace(owner), strings.TrimSpace(id)
	if owner == "" || id == "" {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor task identity is required", ErrInvalid)
	}
	job, err := canceler.CancelVideoJob(ctx, owner, id)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	return ProviderTaskRef{ProviderTaskID: strings.TrimSpace(job.ID), State: mapLocalVideoState(job.State)}, nil
}

func derivedLocalSourceTaskID(input LocalVideoJobInput) (string, bool) {
	batchID := strings.TrimSpace(input.BatchID)
	bookID := strings.TrimSpace(input.BookID)
	videoID := strings.TrimSpace(input.VideoID)
	if batchID == "" || bookID == "" || videoID == "" {
		return "", false
	}
	return "bf11:" + batchID + ":" + bookID + ":" + videoID, true
}

func (a *LocalExecutorVideoAdapter) Poll(ctx context.Context, owner, jobID string) (ProviderTaskRef, error) {
	if a == nil || a.Client == nil {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor is unavailable", ErrUnavailable)
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return ProviderTaskRef{}, fmt.Errorf("%w: local executor job id is required", ErrInvalid)
	}
	job, err := a.Client.GetVideoJob(ctx, owner, jobID)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	ref := ProviderTaskRef{ProviderTaskID: job.ID, State: mapLocalVideoState(job.State)}
	if ref.State == ProductionSucceeded {
		artifactID := strings.TrimSpace(job.ArtifactID)
		if artifactID == "" {
			return ProviderTaskRef{}, fmt.Errorf("local executor succeeded without an artifact")
		}
		ref.MediaURL = a.artifactURL(owner, artifactID)
	}
	if ref.State == ProductionFailed && strings.TrimSpace(job.ErrorMessage) != "" {
		ref.ErrorMessage = boundedProviderError(job.ErrorMessage)
	}
	return ref, nil
}

func (a *LocalExecutorVideoAdapter) artifactURL(owner, id string) string {
	rawID := strings.TrimSpace(id)
	id = url.PathEscape(rawID)
	base := strings.TrimRight(strings.TrimSpace(a.PublicBaseURL), "/")
	if base == "" {
		return "/api/shuihuo-production/local-executor-artifacts/" + id
	}
	if strings.TrimSpace(a.ArtifactSecret) != "" {
		ttl := a.ArtifactTokenTTL
		if ttl <= 0 {
			ttl = 15 * time.Minute
		}
		now := time.Now
		if a.Now != nil {
			now = a.Now
		}
		if token, err := localexecutor.ArtifactPublicToken(a.ArtifactSecret, owner, rawID, now().Add(ttl)); err == nil {
			return base + "/api/local-executor/v1/artifacts/" + id + "?token=" + url.QueryEscape(token)
		}
	}
	return base + "/api/shuihuo-production/local-executor-artifacts/" + id
}

func mapLocalVideoState(state string) ProductionState {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "queued", "leased", "preparing", "submitting":
		return ProductionQueued
	case "accepted", "generating", "downloading", "uploading":
		return ProductionRunning
	case "succeeded", "success", "completed":
		return ProductionSucceeded
	case "failed", "cancelled", "canceled":
		if strings.EqualFold(strings.TrimSpace(state), "failed") {
			return ProductionFailed
		}
		return ProductionCancelled
	default:
		return ProductionRunning
	}
}

func boundedProviderError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 240 {
		return message[:240]
	}
	return message
}
