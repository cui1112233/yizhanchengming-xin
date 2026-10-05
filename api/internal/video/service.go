package video

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	store     Store
	prompts   FinalPromptSource
	providers ProviderFactory
	artifacts ArtifactStore
	masterKey []byte
}

func NewService(store Store, prompts FinalPromptSource, providers ProviderFactory, artifacts ArtifactStore, masterKey []byte) *Service {
	return &Service{store: store, prompts: prompts, providers: providers, artifacts: artifacts, masterKey: append([]byte(nil), masterKey...)}
}

func (s *Service) Start(ctx context.Context, req StartRequest) (StartResult, error) {
	if s == nil || s.store == nil || s.prompts == nil || s.providers == nil {
		return StartResult{}, providerError(ErrorProviderUnavailable, "video service is unavailable", nil)
	}
	providerForModel, ok := ProviderForModel(req.Model)
	if !ok || providerForModel != req.Provider {
		return StartResult{}, providerError(ErrorProviderUnavailable, "provider/model mapping mismatch", nil)
	}
	prompt, err := s.prompts.ResolveFinalPrompt(ctx, req.BatchProjectID, req.BookID)
	if err != nil {
		return StartResult{}, err
	}
	if prompt.StageRunID <= 0 || prompt.PromptVersion <= 0 || strings.TrimSpace(prompt.InputRevision) == "" || strings.TrimSpace(prompt.Text) == "" {
		return StartResult{}, fmt.Errorf("video: completed FINAL_PROMPT is required")
	}
	cfg, err := s.store.GetProviderConfig(ctx, req.Provider, req.Model)
	if err != nil {
		return StartResult{}, normalizeProviderConfigError(err)
	}
	if !cfg.Enabled || len(cfg.EncryptedSecret) == 0 || len(cfg.SecretNonce) == 0 {
		return StartResult{}, providerError(ErrorProviderUnconfigured, "provider is not configured", nil)
	}
	secret, err := DecryptSecret(s.masterKey, cfg.EncryptedSecret, cfg.SecretNonce)
	if err != nil {
		return StartResult{}, err
	}
	provider, err := s.providers.Build(cfg, secret)
	if err != nil {
		return StartResult{}, err
	}

	job := ProductionJob{
		BatchProjectID: req.BatchProjectID,
		BookID: req.BookID,
		Status: JobQueued,
		InputRevision: prompt.InputRevision,
		FinalPromptStageRunID: prompt.StageRunID,
		FinalPromptVersion: prompt.PromptVersion,
		Provider: req.Provider,
		Model: req.Model,
		IdempotencyKey: logicalIdempotencyKey(req, prompt),
	}
	job, created, err := s.store.CreateOrGetProductionJob(ctx, job)
	if err != nil {
		return StartResult{}, err
	}
	if !created {
		task, err := s.store.LatestTaskForJob(ctx, job.ID)
		if err != nil {
			return StartResult{Job: job}, err
		}
		return StartResult{Job: job, Task: task}, nil
	}

	task, err := s.store.CreateProductionTask(ctx, ProductionTask{
		ProductionJobID: job.ID,
		Attempt: 1,
		Provider: req.Provider,
		Model: req.Model,
		RequestID: req.RequestID,
		Status: TaskQueued,
	})
	if err != nil {
		return StartResult{Job: job}, err
	}

	submit, err := provider.Submit(ctx, SubmitRequest{Model: req.Model, Prompt: prompt.Text, RequestID: req.RequestID})
	if err != nil {
		code, message := safeProviderFailure(err)
		task.Status = TaskFailed
		task.ErrorCode = code
		task.ErrorMessage = message
		job.Status = JobFailed
		job.ErrorCode = code
		job.ErrorMessage = message
		_ = s.store.UpdateProductionTask(ctx, task)
		_ = s.store.UpdateProductionJob(ctx, job)
		return StartResult{Job: job, Task: task}, err
	}

	task.ProviderJobID = submit.ProviderJobID
	task.Status = submit.Status
	if task.Status == "" {
		task.Status = TaskQueued
	}
	job.Status = JobRunning
	if submit.Status == TaskSucceeded {
		if err := s.persistSucceededArtifact(ctx, &job, &task, submit.ArtifactURL); err != nil {
			return StartResult{Job: job, Task: task}, err
		}
	}
	if err := s.store.UpdateProductionTask(ctx, task); err != nil {
		return StartResult{Job: job, Task: task}, err
	}
	if err := s.store.UpdateProductionJob(ctx, job); err != nil {
		return StartResult{Job: job, Task: task}, err
	}
	return StartResult{Job: job, Task: task}, nil
}

func (s *Service) PollTask(ctx context.Context, taskID int64) (ProductionTask, error) {
	task, err := s.store.GetProductionTask(ctx, taskID)
	if err != nil {
		return ProductionTask{}, err
	}
	job, err := s.store.GetProductionJob(ctx, task.ProductionJobID)
	if err != nil {
		return ProductionTask{}, err
	}
	cfg, err := s.store.GetProviderConfig(ctx, task.Provider, task.Model)
	if err != nil {
		return task, normalizeProviderConfigError(err)
	}
	secret, err := DecryptSecret(s.masterKey, cfg.EncryptedSecret, cfg.SecretNonce)
	if err != nil {
		return task, err
	}
	provider, err := s.providers.Build(cfg, secret)
	if err != nil {
		return task, err
	}
	poll, err := provider.Poll(ctx, task.ProviderJobID)
	if err != nil {
		return task, err
	}
	task.Status = poll.Status
	task.ErrorCode = poll.ErrorCode
	task.ErrorMessage = truncateError(poll.ErrorMessage)
	switch poll.Status {
	case TaskQueued, TaskRunning:
		job.Status = JobRunning
	case TaskFailed:
		job.Status = JobFailed
		job.ErrorCode = poll.ErrorCode
		job.ErrorMessage = truncateError(poll.ErrorMessage)
	case TaskSucceeded:
		if err := s.persistSucceededArtifact(ctx, &job, &task, poll.ArtifactURL); err != nil {
			return task, err
		}
	case TaskCancelled:
		job.Status = JobCancelled
	}
	if err := s.store.UpdateProductionTask(ctx, task); err != nil {
		return task, err
	}
	if err := s.store.UpdateProductionJob(ctx, job); err != nil {
		return task, err
	}
	return task, nil
}

func (s *Service) RecoverPending(ctx context.Context, limit int) ([]ProductionTask, error) {
	tasks, err := s.store.ListRecoverableTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]ProductionTask, 0, len(tasks))
	for _, task := range tasks {
		polled, pollErr := s.PollTask(ctx, task.ID)
		if pollErr != nil {
			continue
		}
		out = append(out, polled)
	}
	return out, nil
}

func (s *Service) persistSucceededArtifact(ctx context.Context, job *ProductionJob, task *ProductionTask, sourceURL string) error {
	if strings.TrimSpace(sourceURL) == "" {
		err := providerError(ErrorProviderInvalidResponse, "provider succeeded without artifact URL", nil)
		code, message := safeProviderFailure(err)
		task.Status, task.ErrorCode, task.ErrorMessage = TaskFailed, code, message
		job.Status, job.ErrorCode, job.ErrorMessage = JobFailed, code, message
		_ = s.store.UpdateProductionTask(ctx, *task)
		_ = s.store.UpdateProductionJob(ctx, *job)
		return err
	}
	if s.artifacts == nil {
		return providerError(ErrorProviderUnavailable, "artifact store is unavailable", nil)
	}
	artifact, err := s.artifacts.Persist(ctx, sourceURL, fmt.Sprintf("video/%d/%d.mp4", job.ID, task.ID))
	if err != nil {
		code, message := safeProviderFailure(providerError(ErrorProviderRequestFailed, "persist video artifact failed", err))
		task.Status, task.ErrorCode, task.ErrorMessage = TaskFailed, code, message
		job.Status, job.ErrorCode, job.ErrorMessage = JobFailed, code, message
		_ = s.store.UpdateProductionTask(ctx, *task)
		_ = s.store.UpdateProductionJob(ctx, *job)
		return err
	}
	task.ArtifactSourceURL = sourceURL
	task.OutputBucket = artifact.Bucket
	task.OutputObjectKey = artifact.ObjectKey
	task.OutputURL = artifact.URL
	task.Status = TaskSucceeded
	job.Status = JobSucceeded
	job.ErrorCode, job.ErrorMessage = "", ""
	return nil
}

func logicalIdempotencyKey(req StartRequest, prompt FinalPrompt) string {
	raw := fmt.Sprintf("%d|%d|VIDEO|%s|%d|%d|%s|%s", req.BatchProjectID, req.BookID, prompt.InputRevision, prompt.StageRunID, prompt.PromptVersion, req.Provider, req.Model)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func safeProviderFailure(err error) (ErrorCode, string) {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.Code, truncateError(providerErr.Message)
	}
	return ErrorProviderRequestFailed, "video provider request failed"
}

func normalizeProviderConfigError(err error) error {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return err
	}
	return providerError(ErrorProviderUnconfigured, "provider configuration not found", err)
}

func truncateError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 240 {
		return message[:240]
	}
	return message
}
