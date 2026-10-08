package app

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

func requestLogger(logger *slog.Logger, ctx context.Context) *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	if requestID := observability.RequestID(ctx); requestID != "" {
		return logger.With("request_id", requestID)
	}
	return logger
}

func logSafeFailure(logger *slog.Logger, ctx context.Context, subsystem, operation, code string, err error, attrs ...any) {
	args := []any{"subsystem", subsystem, "operation", operation, "error_code", code, "safe_error", observability.SafeError(err)}
	args = append(args, attrs...)
	requestLogger(logger, ctx).Error(operation+" failed", args...)
}

type observedGenerationService struct {
	next   httpapi.GenerationService
	logger *slog.Logger
}

func (s observedGenerationService) ProjectSummary(ctx context.Context, projectID int64) (generation.ProjectSummary, error) {
	return s.next.ProjectSummary(ctx, projectID)
}
func (s observedGenerationService) BookSummary(ctx context.Context, projectID, bookID int64) (generation.BookGenerationResult, error) {
	return s.next.BookSummary(ctx, projectID, bookID)
}
func (s observedGenerationService) RunBook(ctx context.Context, req generation.RunBookRequest) (generation.BookGenerationResult, error) {
	log := requestLogger(s.logger, ctx)
	log.Info("generation book started", "subsystem", "generation", "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "generation_request_id", req.RequestID)
	out, err := s.next.RunBook(ctx, req)
	if err != nil {
		logSafeFailure(s.logger, ctx, "generation", "run_book", "generation_failed", err, "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "generation_request_id", req.RequestID)
		return out, err
	}
	log.Info("generation book completed", "subsystem", "generation", "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "book_run_id", out.Run.ID, "status", out.Run.Status, "generation_request_id", req.RequestID)
	for _, stage := range out.Stages {
		log.Info("generation stage state", "subsystem", "generation", "book_run_id", stage.BookRunID, "book_id", stage.BookID, "stage_run_id", stage.ID, "stage", stage.Stage, "status", stage.Status, "generation_request_id", stage.RequestID)
	}
	return out, nil
}
func (s observedGenerationService) RunBatch(ctx context.Context, req generation.RunBatchRequest) (generation.BatchGenerationResult, error) {
	log := requestLogger(s.logger, ctx)
	log.Info("generation batch started", "subsystem", "generation", "batch_project_id", req.BatchProjectID, "generation_request_id", req.RequestID)
	out, err := s.next.RunBatch(ctx, req)
	if err != nil {
		logSafeFailure(s.logger, ctx, "generation", "run_batch", "generation_batch_failed", err, "batch_project_id", req.BatchProjectID, "generation_request_id", req.RequestID)
		return out, err
	}
	log.Info("generation batch completed", "subsystem", "generation", "batch_project_id", req.BatchProjectID, "completed", out.Completed, "failed", out.Failed, "generation_request_id", req.RequestID)
	return out, nil
}
func (s observedGenerationService) RetryStage(ctx context.Context, req generation.RetryStageRequest) (generation.BookGenerationResult, error) {
	log := requestLogger(s.logger, ctx)
	log.Info("generation stage retry started", "subsystem", "generation", "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "stage", req.Stage, "generation_request_id", req.RequestID)
	out, err := s.next.RetryStage(ctx, req)
	if err != nil {
		logSafeFailure(s.logger, ctx, "generation", "retry_stage", "generation_retry_failed", err, "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "stage", req.Stage, "generation_request_id", req.RequestID)
		return out, err
	}
	if stage, ok := out.Latest[req.Stage]; ok {
		log.Info("generation stage retry completed", "subsystem", "generation", "book_run_id", stage.BookRunID, "book_id", stage.BookID, "stage_run_id", stage.ID, "stage", stage.Stage, "status", stage.Status, "generation_request_id", req.RequestID)
	}
	return out, nil
}
func (s observedGenerationService) StageResult(ctx context.Context, projectID, bookID int64, stage generation.Stage) (generation.StageRun, error) {
	return s.next.StageResult(ctx, projectID, bookID, stage)
}
func (s observedGenerationService) ListPrompts(ctx context.Context) ([]generation.Prompt, error) {
	return s.next.ListPrompts(ctx)
}
func (s observedGenerationService) AudioMeasurement(ctx context.Context, projectID, bookID int64) (generation.AudioMeasurement, error) {
	return s.next.AudioMeasurement(ctx, projectID, bookID)
}
func (s observedGenerationService) MeasureAudio(ctx context.Context, req generation.AudioMeasurementRequest) (generation.AudioMeasurement, error) {
	log := requestLogger(s.logger, ctx)
	log.Info("audio measurement started", "subsystem", "audio", "batch_project_id", req.BatchProjectID, "book_id", req.BookID)
	out, err := s.next.MeasureAudio(ctx, req)
	if err != nil {
		code := "audio_measurement_failed"
		if errors.Is(err, generation.ErrAudioProbeUnavailable) {
			code = "audio_probe_unavailable"
		}
		logSafeFailure(s.logger, ctx, "audio", "measure", code, err, "batch_project_id", req.BatchProjectID, "book_id", req.BookID)
		return out, err
	}
	log.Info("audio measured duration", "subsystem", "audio", "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "duration_ms", out.DurationMS)
	return out, nil
}

type observedVideoService struct {
	next   httpapi.VideoService
	logger *slog.Logger
	states sync.Map
}

func (s *observedVideoService) Start(ctx context.Context, req video.StartRequest) (video.StartResult, error) {
	out, err := s.next.Start(ctx, req)
	if err != nil {
		logSafeFailure(s.logger, ctx, "video", "submit", "video_submit_failed", err, "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "provider", req.Provider, "model", req.Model)
		return out, err
	}
	log := requestLogger(s.logger, ctx)
	log.Info("video submitted", "subsystem", "video", "batch_project_id", req.BatchProjectID, "book_id", req.BookID, "production_job_id", out.Job.ID, "production_task_id", out.Task.ID, "provider", out.Task.Provider, "model", out.Task.Model, "provider_job_id", out.Task.ProviderJobID, "idempotency_key", out.Job.IdempotencyKey, "status", out.Task.Status)
	s.states.Store(out.Task.ID, out.Task.Status)
	return out, nil
}
func (s *observedVideoService) PollTask(ctx context.Context, taskID int64) (video.ProductionTask, error) {
	out, err := s.next.PollTask(ctx, taskID)
	if err != nil {
		logSafeFailure(s.logger, ctx, "video", "poll", "video_poll_failed", err, "production_task_id", taskID)
		return out, err
	}
	previous, loaded := s.states.LoadOrStore(taskID, out.Status)
	if loaded && previous != out.Status {
		requestLogger(s.logger, ctx).Info("video state transition", "subsystem", "video", "production_job_id", out.ProductionJobID, "production_task_id", out.ID, "provider", out.Provider, "model", out.Model, "provider_job_id", out.ProviderJobID, "from_status", previous, "to_status", out.Status)
		s.states.Store(taskID, out.Status)
	}
	return out, nil
}
func (s *observedVideoService) CancelTask(ctx context.Context, taskID int64) (video.ProductionTask, error) {
	out, err := s.next.CancelTask(ctx, taskID)
	if err != nil {
		logSafeFailure(s.logger, ctx, "video", "cancel", "video_cancel_failed", err, "production_task_id", taskID)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("video cancelled", "subsystem", "video", "production_job_id", out.ProductionJobID, "production_task_id", out.ID, "status", out.Status)
	s.states.Store(taskID, out.Status)
	return out, nil
}
func (s *observedVideoService) RetryTask(ctx context.Context, taskID int64, retryRequestID string) (video.StartResult, error) {
	out, err := s.next.RetryTask(ctx, taskID, retryRequestID)
	if err != nil {
		logSafeFailure(s.logger, ctx, "video", "retry", "video_retry_failed", err, "production_task_id", taskID)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("video retry submitted", "subsystem", "video", "production_job_id", out.Job.ID, "production_task_id", out.Task.ID, "provider", out.Task.Provider, "model", out.Task.Model, "provider_job_id", out.Task.ProviderJobID, "idempotency_key", out.Job.IdempotencyKey, "status", out.Task.Status)
	s.states.Store(out.Task.ID, out.Task.Status)
	return out, nil
}

type observedMergeService struct {
	next   httpapi.VideoMergeService
	logger *slog.Logger
}

func (s observedMergeService) StartFromProductionTasks(ctx context.Context, req video.MergeProductionStartRequest) (video.MergeResult, error) {
	requestLogger(s.logger, ctx).Info("merge queued", "subsystem", "merge", "batch_project_id", req.BatchProjectID, "book_id", req.BookID)
	out, err := s.next.StartFromProductionTasks(ctx, req)
	if err != nil {
		logSafeFailure(s.logger, ctx, "merge", "start", "merge_failed", err, "batch_project_id", req.BatchProjectID, "book_id", req.BookID)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("merge state", "subsystem", "merge", "batch_project_id", out.Job.BatchProjectID, "book_id", out.Job.BookID, "merge_job_id", out.Job.ID, "merge_attempt_id", out.Attempt.ID, "status", out.Job.Status)
	return out, nil
}
func (s observedMergeService) Get(ctx context.Context, jobID int64) (video.MergeJob, []video.MergeAttempt, error) {
	return s.next.Get(ctx, jobID)
}
func (s observedMergeService) RetryAttempt(ctx context.Context, attemptID int64) (video.MergeResult, error) {
	out, err := s.next.RetryAttempt(ctx, attemptID)
	if err != nil {
		logSafeFailure(s.logger, ctx, "merge", "retry", "merge_retry_failed", err, "merge_attempt_id", attemptID)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("merge retry state", "subsystem", "merge", "merge_job_id", out.Job.ID, "merge_attempt_id", out.Attempt.ID, "status", out.Job.Status)
	return out, nil
}

type observedPublishingService struct {
	next   httpapi.PublishingService
	logger *slog.Logger
}

func (s observedPublishingService) CreateAccount(ctx context.Context, user authn.User, input publishing.CreateAccountInput) (publishing.Account, error) {
	out, err := s.next.CreateAccount(ctx, user, input)
	if err != nil {
		logSafeFailure(s.logger, ctx, "publishing", "create_account", "publishing_account_failed", err, "user_id", user.ID, "platform", input.Platform)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("publishing account configured", "subsystem", "publishing", "user_id", user.ID, "publishing_account_id", out.ID, "platform", out.Platform)
	return out, nil
}
func (s observedPublishingService) ListAccounts(ctx context.Context, user authn.User) ([]publishing.Account, error) {
	return s.next.ListAccounts(ctx, user)
}
func (s observedPublishingService) ClaimBatchProject(ctx context.Context, user authn.User, projectID int64) error {
	return s.next.ClaimBatchProject(ctx, user, projectID)
}
func (s observedPublishingService) CreateIntent(ctx context.Context, user authn.User, input publishing.CreateIntentInput) (publishing.Intent, error) {
	out, err := s.next.CreateIntent(ctx, user, input)
	if err != nil {
		code := "publish_intent_failed"
		if errors.Is(err, publishing.ErrForbidden) {
			code = "publishing_permission_denied"
		}
		logSafeFailure(s.logger, ctx, "publishing", "create_intent", code, err, "user_id", user.ID, "batch_project_id", input.BatchProjectID, "book_id", input.BookID, "publishing_account_id", input.PublishingAccountID)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("publish intent created", "subsystem", "publishing", "user_id", user.ID, "publish_intent_id", out.ID, "batch_project_id", out.BatchProjectID, "book_id", out.BookID, "publishing_account_id", out.PublishingAccountID, "status", out.Status)
	return out, nil
}
func (s observedPublishingService) GetIntent(ctx context.Context, user authn.User, id int64) (publishing.Intent, error) {
	return s.next.GetIntent(ctx, user, id)
}
func (s observedPublishingService) ListAudits(ctx context.Context, user authn.User, id int64) ([]publishing.Audit, error) {
	return s.next.ListAudits(ctx, user, id)
}

type observedLocalExecutorService struct {
	next   httpapi.VideoLocalExecutorService
	logger *slog.Logger
}

func (s observedLocalExecutorService) Register(ctx context.Context, input video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error) {
	out, err := s.next.Register(ctx, input)
	if err != nil {
		logSafeFailure(s.logger, ctx, "local_executor", "register", "executor_register_failed", err, "provider", input.ProviderKey, "model", input.Model)
		return out, err
	}
	requestLogger(s.logger, ctx).Info("local executor registered", "subsystem", "local_executor", "executor_id", out.Executor.ID, "provider", out.Executor.ProviderKey, "model", out.Executor.Model, "online", out.Executor.Online)
	return out, nil
}
func (s observedLocalExecutorService) CreatePairingIntent(ctx context.Context, ownerUserID int64) (video.LocalExecutorPairingResult, error) {
	out, err := s.next.CreatePairingIntent(ctx, ownerUserID)
	if err != nil {
		logSafeFailure(s.logger, ctx, "local_executor", "create_pairing", "executor_pairing_create_failed", err, "owner_user_id", ownerUserID)
	}
	return out, err
}
func (s observedLocalExecutorService) RedeemPairingIntent(ctx context.Context, payload string, input video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error) {
	out, err := s.next.RedeemPairingIntent(ctx, payload, input)
	if err != nil {
		// Never log the opaque pairing payload or the issued device credential.
		logSafeFailure(s.logger, ctx, "local_executor", "redeem_pairing", "executor_pairing_redeem_failed", err, "provider", input.ProviderKey, "model", input.Model)
	}
	return out, err
}
func (s observedLocalExecutorService) ListForOwner(ctx context.Context, ownerUserID int64) ([]video.LocalExecutorIdentity, error) {
	return s.next.ListForOwner(ctx, ownerUserID)
}
func (s observedLocalExecutorService) UnbindForOwner(ctx context.Context, ownerUserID int64, executorID string) error {
	err := s.next.UnbindForOwner(ctx, ownerUserID, executorID)
	if err != nil {
		logSafeFailure(s.logger, ctx, "local_executor", "unbind", "executor_unbind_failed", err, "owner_user_id", ownerUserID)
	}
	return err
}
func (s observedLocalExecutorService) Identity(ctx context.Context, token string) (video.LocalExecutorIdentity, error) {
	return s.next.Identity(ctx, token)
}
func (s observedLocalExecutorService) Heartbeat(ctx context.Context, token string, input video.LocalExecutorHeartbeatInput) error {
	return s.next.Heartbeat(ctx, token, input)
}
func (s observedLocalExecutorService) List(ctx context.Context) ([]video.LocalExecutorIdentity, error) {
	return s.next.List(ctx)
}
func (s observedLocalExecutorService) CompleteTask(ctx context.Context, token, taskID string, input video.LocalExecutorCompleteInput) error {
	err := s.next.CompleteTask(ctx, token, taskID, input)
	if err != nil {
		logSafeFailure(s.logger, ctx, "local_executor", "complete_task", "executor_task_complete_failed", err, "production_task_id", taskID)
		return err
	}
	requestLogger(s.logger, ctx).Info("local executor task completed", "subsystem", "local_executor", "production_task_id", taskID)
	return nil
}
func (s observedLocalExecutorService) FailTask(ctx context.Context, token, taskID string, input video.LocalExecutorFailInput) error {
	err := s.next.FailTask(ctx, token, taskID, input)
	if err != nil {
		logSafeFailure(s.logger, ctx, "local_executor", "fail_task", "executor_task_fail_failed", err, "production_task_id", taskID)
		return err
	}
	requestLogger(s.logger, ctx).Error("local executor task failed", "subsystem", "local_executor", "production_task_id", taskID, "error_code", observability.SanitizeString(input.Code), "safe_error", observability.SanitizeString(input.Message))
	return nil
}
