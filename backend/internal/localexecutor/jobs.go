package localexecutor

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type JobStore interface {
	CreateJob(context.Context, JobRecord) error
	JobForOwner(context.Context, string, string) (JobRecord, error)
	CancelJob(context.Context, string, string, time.Time) (JobRecord, error)
	ClaimJob(context.Context, ExecutorRecord, SecretHash, time.Time, time.Time) (JobRecord, error)
	RenewJob(context.Context, string, string, SecretHash, int64, time.Time, time.Time) (JobRecord, error)
	SetJobState(context.Context, string, string, SecretHash, int64, JobState, time.Time) (JobRecord, error)
	AcceptJob(context.Context, string, string, SecretHash, int64, string, string, time.Time) (JobRecord, error)
	ReleaseJob(context.Context, string, string, SecretHash, int64, string, time.Time) (JobRecord, error)
	FailJob(context.Context, string, string, SecretHash, int64, string, string, time.Time) (JobRecord, error)
	CompleteJob(context.Context, string, string, SecretHash, int64, string, time.Time) (JobRecord, error)
}

func (s *Service) requireJobs() (JobStore, error) {
	if s.jobs == nil {
		return nil, ErrInvalidInput
	}
	return s.jobs, nil
}

func jobView(r JobRecord) JobView {
	return JobView{
		ID: r.ID, SourceTaskID: r.SourceTaskID, Platform: r.Platform, Payload: r.Payload,
		State: r.State, CancelRequested: r.CancelRequested, LeaseExecutorID: r.LeaseExecutorID,
		LeaseGeneration: r.LeaseGeneration, LeaseExpiresAt: r.LeaseExpiresAt, AcceptedAt: r.AcceptedAt,
		AcceptedAccountID: r.AcceptedAccountID, SubmissionID: r.SubmissionID, ArtifactID: r.ArtifactID,
		ErrorCode: r.ErrorCode, ErrorMessage: r.ErrorMessage, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (s *Service) CreateJob(ctx context.Context, owner string, input CreateJobInput) (JobView, error) {
	jobs, err := s.requireJobs()
	if err != nil {
		return JobView{}, err
	}
	owner = strings.TrimSpace(owner)
	source := strings.TrimSpace(input.SourceTaskID)
	platform := strings.ToLower(strings.TrimSpace(input.Platform))
	if owner == "" || source == "" {
		return JobView{}, ErrInvalidInput
	}
	if platform != PlatformDoubao {
		return JobView{}, ErrInvalidPlatform
	}
	payload, err := json.Marshal(input.Payload)
	if err != nil {
		return JobView{}, ErrInvalidInput
	}
	id, err := randomID("lej_", 12)
	if err != nil {
		return JobView{}, err
	}
	now := s.now().UTC()
	record := JobRecord{
		ID: id, OwnerUsername: owner, SourceTaskID: source, Platform: platform,
		Payload: payload, State: JobQueued, CreatedAt: now, UpdatedAt: now,
	}
	if err := jobs.CreateJob(ctx, record); err != nil {
		return JobView{}, err
	}
	s.notifyJobCreated()
	return jobView(record), nil
}

func (s *Service) GetJob(ctx context.Context, owner, id string) (JobView, error) {
	jobs, err := s.requireJobs()
	if err != nil {
		return JobView{}, err
	}
	record, err := jobs.JobForOwner(ctx, strings.TrimSpace(owner), strings.TrimSpace(id))
	if err != nil {
		return JobView{}, err
	}
	return jobView(record), nil
}

func (s *Service) CancelJob(ctx context.Context, owner, id string) (JobView, error) {
	jobs, err := s.requireJobs()
	if err != nil {
		return JobView{}, err
	}
	record, err := jobs.CancelJob(ctx, strings.TrimSpace(owner), strings.TrimSpace(id), s.now().UTC())
	if err != nil {
		return JobView{}, err
	}
	return jobView(record), nil
}

func (s *Service) ClaimJob(ctx context.Context, executorToken string) (ClaimResult, error) {
	jobs, err := s.requireJobs()
	if err != nil {
		return ClaimResult{}, err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return ClaimResult{}, err
	}
	leaseToken, err := randomToken()
	if err != nil {
		return ClaimResult{}, err
	}
	now := s.now().UTC()
	expires := now.Add(JobLeaseTTL)
	record, err := jobs.ClaimJob(ctx, executor, hashSecret(leaseToken), expires, now)
	if err != nil {
		return ClaimResult{}, err
	}
	if record.LeaseExpiresAt == nil {
		return ClaimResult{}, ErrJobConflict
	}
	return ClaimResult{
		Job: jobView(record), LeaseToken: leaseToken,
		LeaseGeneration: record.LeaseGeneration, LeaseExpiresAt: *record.LeaseExpiresAt,
	}, nil
}

func (s *Service) ClaimJobWait(ctx context.Context, executorToken string, wait time.Duration) (ClaimResult, error) {
	if wait <= 0 {
		return s.ClaimJob(ctx, executorToken)
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	for {
		signal := s.currentJobSignal()
		claim, err := s.ClaimJob(ctx, executorToken)
		if !errors.Is(err, ErrNoClaimableJob) {
			return claim, err
		}
		select {
		case <-ctx.Done():
			return ClaimResult{}, ctx.Err()
		case <-timer.C:
			return ClaimResult{}, ErrNoClaimableJob
		case <-signal:
		}
	}
}

func (s *Service) RenewJob(ctx context.Context, executorToken, jobID string, lease LeaseCredential) (LeaseView, error) {
	jobs, err := s.requireJobs()
	if err != nil {
		return LeaseView{}, err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return LeaseView{}, err
	}
	if strings.TrimSpace(lease.Token) == "" || lease.Generation < 1 {
		return LeaseView{}, ErrStaleLease
	}
	now := s.now().UTC()
	expires := now.Add(JobLeaseTTL)
	record, err := jobs.RenewJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, expires, now)
	if err != nil {
		return LeaseView{}, err
	}
	if record.LeaseExpiresAt == nil {
		return LeaseView{}, ErrStaleLease
	}
	return LeaseView{LeaseGeneration: record.LeaseGeneration, LeaseExpiresAt: *record.LeaseExpiresAt}, nil
}

func (s *Service) RecordProgress(ctx context.Context, executorToken, jobID string, lease LeaseCredential, next JobState) error {
	jobs, err := s.requireJobs()
	if err != nil {
		return err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return err
	}
	if !isProgressState(next) || strings.TrimSpace(lease.Token) == "" || lease.Generation < 1 {
		return ErrInvalidJobState
	}
	_, err = jobs.SetJobState(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, next, s.now().UTC())
	return err
}

func (s *Service) RecordAcceptance(ctx context.Context, executorToken, jobID string, lease LeaseCredential, input AcceptanceInput) error {
	jobs, err := s.requireJobs()
	if err != nil {
		return err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return err
	}
	accountID := strings.TrimSpace(input.AccountID)
	submissionID := strings.TrimSpace(input.SubmissionID)
	if accountID == "" || submissionID == "" || strings.TrimSpace(lease.Token) == "" || lease.Generation < 1 {
		return ErrInvalidInput
	}
	_, err = jobs.AcceptJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, accountID, submissionID, s.now().UTC())
	return err
}

func (s *Service) ReleaseJob(ctx context.Context, executorToken, jobID string, lease LeaseCredential, reason string) error {
	jobs, err := s.requireJobs()
	if err != nil {
		return err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return err
	}
	_, err = jobs.ReleaseJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, bounded(reason, 512), s.now().UTC())
	return err
}

func (s *Service) FailJob(ctx context.Context, executorToken, jobID string, lease LeaseCredential, input FailureInput) error {
	jobs, err := s.requireJobs()
	if err != nil {
		return err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return err
	}
	_, err = jobs.FailJob(ctx, executor.ID, strings.TrimSpace(jobID), hashSecret(lease.Token), lease.Generation, bounded(input.Code, 96), bounded(input.Message, 512), s.now().UTC())
	return err
}

func (s *Service) CompleteJob(ctx context.Context, executorToken, jobID string, lease LeaseCredential, input ResultInput) error {
	jobs, err := s.requireJobs()
	if err != nil {
		return err
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return err
	}
	jobID = strings.TrimSpace(jobID)
	artifactID := strings.TrimSpace(input.ArtifactID)
	if artifactID == "" {
		return ErrInvalidInput
	}
	if err := s.validateResultArtifact(ctx, jobID, artifactID); err != nil {
		return err
	}
	_, err = jobs.CompleteJob(ctx, executor.ID, jobID, hashSecret(lease.Token), lease.Generation, artifactID, s.now().UTC())
	return err
}

func (s *Service) executorForToken(ctx context.Context, token string) (ExecutorRecord, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	record, err := s.store.ExecutorByTokenHash(ctx, hashSecret(token))
	if err != nil {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	return record, nil
}

// OwnerForToken returns the paired account owner for executor-only routes.
func (s *Service) OwnerForToken(ctx context.Context, token string) (string, error) {
	record, err := s.executorForToken(ctx, token)
	if err != nil {
		return "", err
	}
	return record.OwnerUsername, nil
}

func isProgressState(state JobState) bool {
	switch state {
	case JobPreparing, JobSubmitting, JobAcceptanceUnknown, JobGenerating, JobDownloading, JobUploading:
		return true
	default:
		return false
	}
}

func bounded(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

func validateTransition(from, to JobState, accepted bool) error {
	if from == to {
		return nil
	}
	if accepted {
		switch {
		case from == JobAccepted && to == JobGenerating:
			return nil
		case from == JobGenerating && to == JobDownloading:
			return nil
		case from == JobDownloading && to == JobUploading:
			return nil
		default:
			return ErrInvalidJobState
		}
	}
	switch {
	case from == JobLeased && to == JobPreparing:
		return nil
	case from == JobPreparing && to == JobSubmitting:
		return nil
	case from == JobSubmitting && to == JobAcceptanceUnknown:
		return nil
	default:
		return ErrInvalidJobState
	}
}

func checkLease(record JobRecord, executorID string, leaseHash SecretHash, generation int64, now time.Time) error {
	if record.CancelRequested || record.State == JobCancelled {
		return ErrJobCancelled
	}
	if record.LeaseExecutorID != executorID || record.LeaseGeneration != generation || record.LeaseTokenHash != leaseHash || record.LeaseExpiresAt == nil || !record.LeaseExpiresAt.After(now) {
		return ErrStaleLease
	}
	return nil
}

func isAccepted(record JobRecord) bool { return record.AcceptedAt != nil }
