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

func (s *Service) createBatch(ctx context.Context, owner, projectID string, input BatchInput) (Batch, error) {
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
		AutomationState: AutomationState{MaxConsecutiveFailures: input.MaxConsecutiveFailures},
		ModeInputs:      selected.ModeInputs, KnowledgeIDs: selected.KnowledgeIDs, NoteIDs: selected.NoteIDs, StyleProfileID: selected.StyleProfileID,
		ProjectID: projectID, SessionID: input.SessionID, Status: BatchRunning,
		TotalChapters: input.TotalChapters, Attempt: 1, Prompt: input.Prompt,
		ModelID: modelID, ThinkingLevel: input.ThinkingLevel,
	})
}

func (s *Service) CreateBatch(ctx context.Context, owner, project string, input BatchInput) (Batch, error) {
	if s == nil || s.store == nil {
		return Batch{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, project = strings.TrimSpace(owner), strings.TrimSpace(project)
	var out Batch
	err := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		var e error
		out, e = s.automationChild(tx).createBatch(ctx, owner, project, input)
		return e
	})
	return out, err
}
func (s *Service) PauseBatch(ctx context.Context, owner, project, id string) (Batch, error) {
	return s.controlBatch(ctx, owner, project, id, "pause")
}
func (s *Service) ResumeBatch(ctx context.Context, owner, project, id string) (Batch, error) {
	return s.controlBatch(ctx, owner, project, id, "resume")
}
func (s *Service) StopBatch(ctx context.Context, owner, project, id string) (Batch, error) {
	return s.controlBatch(ctx, owner, project, id, "stop")
}
func (s *Service) controlBatch(ctx context.Context, owner, project, id, action string) (Batch, error) {
	if s == nil || s.store == nil {
		return Batch{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, project, id = strings.TrimSpace(owner), strings.TrimSpace(project), strings.TrimSpace(id)
	var out Batch
	var cancelledID string
	err := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		child := s.automationChild(tx)
		p, e := child.batchInProject(ctx, owner, project, id)
		if e != nil {
			return e
		}
		switch action {
		case "pause":
			if p.Status != BatchRunning && p.Status != BatchPaused {
				return &DomainError{Code: ErrInvalidTransition, Message: "当前批量任务状态不允许暂停"}
			}
			p.Status = BatchPaused
			p.StoppedReason = "用户暂停"
		case "resume":
			if p.Status != BatchPaused && p.Status != BatchFailed && p.Status != BatchRunning {
				return &DomainError{Code: ErrInvalidTransition, Message: "当前批量任务状态不允许继续"}
			}
			p.defaults()
			if p.TotalChapters < 1 || p.TotalChapters > 50 || p.MaxConsecutiveFailures < 1 || p.MaxConsecutiveFailures > 10 {
				return &DomainError{Code: ErrInputRequired, Message: "配置需调整"}
			}
			if _, e = child.prepareGenerationInput(ctx, owner, p.ProjectID, p.SessionID, batchInput(p)); e != nil {
				return e
			}
			if r := p.active(); r != nil {
				if _, e := child.inspectRound(ctx, owner, r); e != nil {
					return e
				}
				if r.ErrorCode == ErrResultUnknown {
					return unknownResult()
				}
			}
			p.Status = BatchRunning
			p.StoppedReason = ""
			p.ConsecutiveFailures = 0
		case "stop":
			if p.Status == BatchCompleted {
				return nil
			}
			p.Status = BatchStopped
			p.StoppedReason = "用户停止"
			if r := p.active(); r != nil {
				if _, e := child.inspectRound(ctx, owner, r); e != nil {
					return e
				}
				p.CurrentRunID = r.RunID
			}
			if p.CurrentRunID != "" {
				run, e := tx.GetRun(ctx, owner, p.CurrentRunID)
				if e != nil {
					return e
				}
				if !isTerminalRunStatus(run.Status) {
					if e = s.cancelAutomationInTx(ctx, tx, owner, run.ID); e != nil {
						return e
					}
					cancelledID = run.ID
				}
			}
		}
		out, e = tx.UpdateBatch(ctx, owner, p)
		return e
	})
	if err == nil {
		s.signalAutomationCancel(ctx, owner, cancelledID)
	}
	return out, err
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
func (s *Service) advanceBatch(ctx context.Context, stale Batch) (Run, error) {
	var out Run
	var observed error
	owner := stale.OwnerUsername
	now := s.automationNow()
	err := s.store.Atomic(ctx, owner, stale.ProjectID, func(tx Store) error {
		child := s.automationChild(tx)
		p, e := tx.GetBatch(ctx, owner, stale.ID)
		if e != nil {
			return e
		}
		p.defaults()
		save := func() error { _, e := tx.UpdateBatch(ctx, owner, p); return e }
		if p.TotalChapters < 1 || p.TotalChapters > 50 || p.MaxConsecutiveFailures < 1 || p.MaxConsecutiveFailures > 10 {
			p.Status = BatchPaused
			p.StoppedReason = "配置需调整：章节数或失败阈值超出范围"
			p.LastError = p.StoppedReason
			return save()
		}
		if p.Status != BatchRunning && p.Status != BatchPaused && p.Status != BatchStopped {
			return nil
		}
		if p.CurrentRunID != "" && p.ActiveRound == "" {
			key := fmt.Sprintf("legacy:%d", p.Attempt)
			p.Rounds = append(p.Rounds, AutomationRound{Key: key, RunID: p.CurrentRunID, PlannedAt: now})
			p.ActiveRound = key
		}
		previous := ""
		if r := p.active(); r != nil {
			outcome, e := child.inspectRound(ctx, owner, r)
			if e != nil {
				return e
			}
			p.CurrentRunID = r.RunID
			if outcome == "waiting" {
				p.WaitingReason = automationWaitReason(r.State)
				return save()
			}
			if outcome == "halt" && r.ErrorCode == ErrResultUnknown {
				p.Status = BatchFailed
				p.StoppedReason = "上游结果待确认，已暂停自动重试"
				p.LastError = p.StoppedReason
				r.State = "paused"
				return save()
			}
			previous = r.RunID
			finished := now
			r.FinishedAt = &finished
			r.State = outcome
			if outcome == "success" {
				r.State = "saved"
				p.CompletedChapters++
				p.TotalSavedChapters++
				p.ConsecutiveFailures = 0
				p.LastError = ""
			} else if outcome == "cancelled" {
				p.Status = BatchStopped
				p.StoppedReason = "用户停止"
			} else {
				p.ConsecutiveFailures++
				p.LastError = r.ErrorMessage
				p.LastFailure = r.ErrorMessage
				if outcome != "failure" || p.ConsecutiveFailures >= p.MaxConsecutiveFailures {
					if p.Status != BatchStopped {
						p.Status = BatchFailed
					}
					if p.Status != BatchStopped {
						p.StoppedReason = r.ErrorMessage
					}
					if p.StoppedReason == "" {
						p.StoppedReason = "章节生成失败，请检查运行结果"
					}
				}
			}
			p.ActiveRound = ""
			p.CurrentRunID = ""
			p.Attempt++
			p.WaitingReason = ""
		}
		if p.CompletedChapters >= p.TotalChapters && p.Status != BatchStopped {
			p.Status = BatchCompleted
			p.StoppedReason = ""
			return save()
		}
		if p.Status != BatchRunning {
			return save()
		}
		if previous == "" && len(p.Rounds) > 0 {
			previous = p.Rounds[len(p.Rounds)-1].RunID
		}
		key := fmt.Sprintf("batch:%s:%d:%d", p.ID, p.CompletedChapters+1, p.Attempt)
		run, e := child.queueAutomationMessage(ctx, owner, SendMessageInput{SourceType: "batch", SourcePlanID: p.ID, SourceRound: key, SourceRunID: previous, SessionID: p.SessionID, ProjectID: p.ProjectID, Content: p.Prompt, ModelID: p.ModelID, ThinkingLevel: p.ThinkingLevel, ModeInputs: p.ModeInputs, KnowledgeIDs: p.KnowledgeIDs, NoteIDs: p.NoteIDs, StyleProfileID: p.StyleProfileID, IdempotencyKey: key})
		if e != nil {
			if busy(e) {
				p.WaitingReason = "该作品正在创作，等待空闲"
				return save()
			}
			if IsCode(e, ErrPersistenceFailed) || IsCode(e, ErrIdempotencyConflict) {
				return e
			}
			observed = e
			p.Status = BatchFailed
			p.LastError = e.Error()
			p.StoppedReason = e.Error()
			if IsCode(e, ErrInputRequired) || IsCode(e, ErrSessionNotFound) || IsCode(e, ErrProjectNotFound) {
				p.LastError = "配置需调整：" + e.Error()
				p.StoppedReason = p.LastError
			}
			return save()
		}
		if e = tx.InsertAutomationRound(ctx, owner, "batch", p.ID, key, run.ID, now); e != nil {
			return e
		}
		p.Rounds = append(p.Rounds, AutomationRound{Key: key, RunID: run.ID, PreviousRunID: previous, PlannedAt: now, State: "queued"})
		p.ActiveRound = key
		p.CurrentRunID = run.ID
		p.WaitingReason = ""
		out = run
		return save()
	})
	if err != nil {
		return Run{}, errors.Join(observed, err)
	}
	if out.ID != "" {
		s.Notify(ctx, owner, out.ID)
	}
	return out, observed
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
