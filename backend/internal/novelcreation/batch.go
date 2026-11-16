package novelcreation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DispatchedBatch is one batch step that created a chapter run.
type DispatchedBatch struct {
	Batch Batch
	Run   Run
}

func (s *Service) ListBatches(ctx context.Context, owner, projectID string) ([]Batch, error) {
	if s == nil || s.store == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.ListBatches(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
}

func (s *Service) CreateBatch(ctx context.Context, owner, projectID string, input BatchInput) (Batch, error) {
	if s == nil || s.store == nil {
		return Batch{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	input, err := NormalizeBatchInput(input)
	if err != nil {
		return Batch{}, err
	}
	project, err := s.store.GetProject(ctx, owner, projectID)
	if err != nil {
		return Batch{}, err
	}
	session, err := s.store.GetSession(ctx, owner, input.SessionID)
	if err != nil {
		return Batch{}, err
	}
	if session.ProjectID != projectID {
		return Batch{}, &DomainError{Code: ErrInputRequired, Message: "会话与作品不匹配"}
	}
	selected, err := s.prepareGenerationInput(ctx, owner, projectID, input.SessionID, AutomationInput{ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID})
	if err != nil {
		return Batch{}, err
	}
	existing, err := s.store.ListBatches(ctx, owner, projectID)
	if err != nil {
		return Batch{}, err
	}
	for _, item := range existing {
		if item.Status == BatchRunning || item.Status == BatchPaused {
			return Batch{}, &DomainError{Code: ErrRunInProgress, Message: "该作品已有批量续写任务，请先停止现有任务"}
		}
	}
	modelID := input.ModelID
	if modelID == "" {
		modelID = project.ModelID
	}
	return s.store.CreateBatch(ctx, owner, Batch{
		ModeInputs: selected.ModeInputs, KnowledgeIDs: selected.KnowledgeIDs, NoteIDs: selected.NoteIDs, StyleProfileID: selected.StyleProfileID,
		ProjectID: projectID, SessionID: input.SessionID, Status: BatchRunning,
		TotalChapters: input.TotalChapters, Attempt: 1, Prompt: input.Prompt,
		ModelID: modelID, ThinkingLevel: input.ThinkingLevel,
	})
}

func (s *Service) PauseBatch(ctx context.Context, owner, projectID, batchID string) (Batch, error) {
	batch, err := s.batchInProject(ctx, owner, projectID, batchID)
	if err != nil {
		return Batch{}, err
	}
	if batch.Status == BatchPaused {
		return batch, nil
	}
	if batch.Status != BatchRunning {
		return Batch{}, &DomainError{Code: ErrInvalidTransition, Message: "当前批量任务状态不允许暂停"}
	}
	batch.Status = BatchPaused
	return s.store.UpdateBatch(ctx, strings.TrimSpace(owner), batch)
}

func (s *Service) ResumeBatch(ctx context.Context, owner, projectID, batchID string) (Batch, error) {
	batch, err := s.batchInProject(ctx, owner, projectID, batchID)
	if err != nil {
		return Batch{}, err
	}
	if batch.Status == BatchRunning {
		return batch, nil
	}
	if batch.Status != BatchPaused && batch.Status != BatchFailed {
		return Batch{}, &DomainError{Code: ErrInvalidTransition, Message: "当前批量任务状态不允许继续"}
	}
	if _, err := s.prepareGenerationInput(ctx, strings.TrimSpace(owner), batch.ProjectID, batch.SessionID, batchInput(batch)); err != nil {
		return Batch{}, err
	}
	if batch.Status == BatchFailed {
		batch.Attempt++
		batch.CurrentRunID = ""
	}
	batch.Status = BatchRunning
	batch.LastError = ""
	if batch.CurrentRunID != "" {
		run, runErr := s.store.GetRun(ctx, strings.TrimSpace(owner), batch.CurrentRunID)
		if runErr != nil || isTerminalRunStatus(run.Status) {
			batch.CurrentRunID = ""
			batch.Attempt++
		}
	}
	return s.store.UpdateBatch(ctx, strings.TrimSpace(owner), batch)
}

func (s *Service) StopBatch(ctx context.Context, owner, projectID, batchID string) (Batch, error) {
	batch, err := s.batchInProject(ctx, owner, projectID, batchID)
	if err != nil {
		return Batch{}, err
	}
	owner = strings.TrimSpace(owner)
	if batch.Status == BatchStopped || batch.Status == BatchCompleted {
		return batch, nil
	}
	if batch.CurrentRunID != "" {
		if run, runErr := s.store.GetRun(ctx, owner, batch.CurrentRunID); runErr == nil && !isTerminalRunStatus(run.Status) {
			if _, cancelErr := s.CancelRun(ctx, owner, run.ID); cancelErr != nil {
				return Batch{}, cancelErr
			}
		}
	}
	batch.Status = BatchStopped
	batch.CurrentRunID = ""
	batch.Attempt++
	return s.store.UpdateBatch(ctx, owner, batch)
}

func (s *Service) batchInProject(ctx context.Context, owner, projectID, batchID string) (Batch, error) {
	if s == nil || s.store == nil {
		return Batch{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	batch, err := s.store.GetBatch(ctx, strings.TrimSpace(owner), strings.TrimSpace(batchID))
	if err != nil {
		return Batch{}, err
	}
	if projectID = strings.TrimSpace(projectID); projectID != "" && batch.ProjectID != projectID {
		return Batch{}, &DomainError{Code: ErrProjectNotFound, Message: "batch not found"}
	}
	return batch, nil
}

// AdvanceBatches drives every running batch by one step: finish the current
// chapter, record its outcome, then queue the next chapter run.
func (s *Service) AdvanceBatches(ctx context.Context) ([]DispatchedBatch, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作运行服务尚未配置"}
	}
	active, err := s.store.ListActiveBatches(ctx, 20)
	if err != nil {
		return nil, err
	}
	dispatched := make([]DispatchedBatch, 0, len(active))
	var failures []error
	for _, batch := range active {
		run, advanceErr := s.advanceBatch(ctx, batch)
		if advanceErr != nil {
			failures = append(failures, fmt.Errorf("batch %s: %w", batch.ID, advanceErr))
		}
		if run.ID != "" {
			dispatched = append(dispatched, DispatchedBatch{Batch: batch, Run: run})
		}
	}
	return dispatched, errors.Join(failures...)
}

// advanceBatch returns the newly queued run, if any.
func (s *Service) advanceBatch(ctx context.Context, batch Batch) (Run, error) {
	owner := batch.OwnerUsername
	if batch.CompletedChapters >= batch.TotalChapters {
		batch.Status = BatchCompleted
		batch.LastError = ""
		if _, err := s.store.UpdateBatch(ctx, owner, batch); err != nil {
			return Run{}, err
		}
		return Run{}, nil
	}
	if batch.CurrentRunID != "" {
		run, err := s.store.GetRun(ctx, owner, batch.CurrentRunID)
		if err == nil && !isTerminalRunStatus(run.Status) {
			return Run{}, nil
		}
		if err == nil {
			switch run.Status {
			case RunSucceeded, RunSucceededWithWarnings:
				batch.CompletedChapters++
				batch.CurrentRunID = ""
				if batch.CompletedChapters >= batch.TotalChapters {
					batch.Status = BatchCompleted
					batch.LastError = ""
					if _, updateErr := s.store.UpdateBatch(ctx, owner, batch); updateErr != nil {
						return Run{}, updateErr
					}
					return Run{}, nil
				}
				if _, updateErr := s.store.UpdateBatch(ctx, owner, batch); updateErr != nil {
					return Run{}, updateErr
				}
			case RunFailed:
				batch.Status = BatchFailed
				batch.CurrentRunID = ""
				batch.LastError = run.ErrorMessage
				if batch.LastError == "" {
					batch.LastError = "章节生成失败，可继续批量续写重试"
				}
				if _, updateErr := s.store.UpdateBatch(ctx, owner, batch); updateErr != nil {
					return Run{}, updateErr
				}
				return Run{}, nil
			case RunCancelled:
				batch.Status = BatchStopped
				batch.CurrentRunID = ""
				if _, updateErr := s.store.UpdateBatch(ctx, owner, batch); updateErr != nil {
					return Run{}, updateErr
				}
				return Run{}, nil
			}
		}
	}
	run, err := s.dispatchBatchRun(ctx, owner, batch)
	if err != nil {
		if IsCode(err, ErrRunInProgress) || IsCode(err, ErrProjectLocked) {
			return Run{}, nil
		}
		batch.Status = BatchFailed
		batch.LastError = err.Error()
		if IsCode(err, ErrInputRequired) || IsCode(err, ErrSessionNotFound) || IsCode(err, ErrProjectNotFound) {
			batch.LastError = "配置需调整：" + err.Error()
		}
		if _, updateErr := s.store.UpdateBatch(ctx, owner, batch); updateErr != nil {
			return Run{}, errors.Join(err, updateErr)
		}
		return Run{}, err
	}
	batch.CurrentRunID = run.ID
	if _, updateErr := s.store.UpdateBatch(ctx, owner, batch); updateErr != nil {
		return run, updateErr
	}
	return run, nil
}

func (s *Service) dispatchBatchRun(ctx context.Context, owner string, batch Batch) (Run, error) {
	key := fmt.Sprintf("batch:%s:%d:%d", batch.ID, batch.CompletedChapters+1, batch.Attempt)
	return s.SendMessage(ctx, owner, SendMessageInput{
		SessionID: batch.SessionID, ProjectID: batch.ProjectID,
		Content: batch.Prompt, ModelID: batch.ModelID, ThinkingLevel: batch.ThinkingLevel,
		ModeInputs: batch.ModeInputs, KnowledgeIDs: batch.KnowledgeIDs, NoteIDs: batch.NoteIDs, StyleProfileID: batch.StyleProfileID,
		IdempotencyKey: key,
	})
}

// StartBatchWorker advances running batches and starts their chapter runs.
func StartBatchWorker(ctx context.Context, service *Service, start func(context.Context, string, string), interval time.Duration) {
	if service == nil || start == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dispatched, err := service.AdvanceBatches(context.Background())
				if err != nil {
					fmt.Printf("novel creation batch worker: %v\n", err)
				}
				for _, item := range dispatched {
					start(context.Background(), item.Batch.OwnerUsername, item.Run.ID)
				}
			}
		}
	}()
}
