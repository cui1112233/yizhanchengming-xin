package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const projectRunLockTTL = 15 * time.Minute

type PipelineRunner interface {
	Run(context.Context, RunContext, EventSink) (ChapterBundle, error)
}

type Snapshot struct {
	Project  Project    `json:"project"`
	Session  Session    `json:"session"`
	Run      Run        `json:"run"`
	Messages []Message  `json:"messages"`
	Events   []RunEvent `json:"events"`
	Chapters []Chapter  `json:"chapters"`
}

type Service struct {
	store    Store
	runtime  Runtime
	pipeline PipelineRunner
}

func NewService(store Store, runtime Runtime, pipeline PipelineRunner) *Service {
	return &Service{store: store, runtime: runtime, pipeline: pipeline}
}

func (s *Service) CreateProject(ctx context.Context, owner string, input CreateProjectInput) (Project, error) {
	if s == nil || s.store == nil {
		return Project{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	input, err := NormalizeCreateProjectInput(input)
	if err != nil {
		return Project{}, err
	}
	if strings.TrimSpace(owner) == "" || input.Title == "" || input.Genre == "" || input.ModelID == "" {
		return Project{}, &DomainError{Code: ErrInputRequired, Message: "作品名称、题材和文本模型不能为空"}
	}
	return s.store.CreateProject(ctx, strings.TrimSpace(owner), input)
}

func (s *Service) CreateSession(ctx context.Context, owner string, input CreateSessionInput) (Session, error) {
	if s == nil || s.store == nil {
		return Session{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.Mode = strings.TrimSpace(input.Mode)
	input.Title = strings.TrimSpace(input.Title)
	if input.ProjectID == "" || input.Mode == "" {
		return Session{}, &DomainError{Code: ErrInputRequired, Message: "作品和创作模式不能为空"}
	}
	return s.store.CreateSession(ctx, strings.TrimSpace(owner), input)
}

func (s *Service) ListProjects(ctx context.Context, owner string) ([]Project, error) {
	return s.store.ListProjects(ctx, strings.TrimSpace(owner))
}

func (s *Service) ListSessions(ctx context.Context, owner, projectID string) ([]Session, error) {
	return s.store.ListSessions(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
}

func (s *Service) SendMessage(ctx context.Context, owner string, input SendMessageInput) (Run, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作运行服务尚未配置"}
	}
	input, err := NormalizeSendMessageInput(input)
	if err != nil {
		return Run{}, err
	}
	owner = strings.TrimSpace(owner)
	if existing, found, err := s.store.FindRunByIdempotencyKey(ctx, owner, input.IdempotencyKey); err != nil {
		return Run{}, err
	} else if found {
		return existing, nil
	}
	session, err := s.store.GetSession(ctx, owner, input.SessionID)
	if err != nil {
		return Run{}, err
	}
	if input.ProjectID == "" {
		input.ProjectID = session.ProjectID
	}
	if session.ProjectID != input.ProjectID {
		return Run{}, &DomainError{Code: ErrInputRequired, Message: "会话与作品不匹配"}
	}
	if _, err := s.store.GetProject(ctx, owner, input.ProjectID); err != nil {
		return Run{}, err
	}
	message, err := s.store.AppendMessage(ctx, owner, Message{
		SessionID: input.SessionID, ProjectID: input.ProjectID, Role: "user", Content: input.Content,
		Status: "accepted", ModelID: input.ModelID, ThinkingLevel: input.ThinkingLevel,
	})
	if err != nil {
		return Run{}, err
	}
	promptSnapshot, _ := json.Marshal(map[string]any{
		"version": 1, "user_message": input.Content, "chapter_number": 1,
		"previous_truth": map[string]any{}, "previous_memory": map[string]any{},
		"message_id": message.ID,
	})
	run, err := s.store.CreateRun(ctx, owner, Run{
		ProjectID: input.ProjectID, SessionID: input.SessionID, IdempotencyKey: input.IdempotencyKey,
		Status: RunQueued, ModelID: input.ModelID, ThinkingLevel: input.ThinkingLevel, PromptSnapshot: promptSnapshot,
	})
	if err != nil {
		if IsCode(err, ErrRunInProgress) {
			if existing, found, findErr := s.store.FindRunByIdempotencyKey(ctx, owner, input.IdempotencyKey); findErr == nil && found {
				return existing, nil
			}
		}
		return Run{}, err
	}
	if _, err := s.appendEvent(ctx, owner, run, "run.queued", RunQueued, "已加入小说创作队列", 0, nil); err != nil {
		return run, err
	}
	if err := s.runtime.Enqueue(ctx, run.ID); err != nil {
		failed, failErr := s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "小说创作任务入队失败", err)
		if failErr != nil {
			return failed, failErr
		}
		return failed, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作任务入队失败，请稍后重试", Cause: err}
	}
	return run, nil
}

func (s *Service) Run(ctx context.Context, owner, runID string) (Run, error) {
	owner, runID = strings.TrimSpace(owner), strings.TrimSpace(runID)
	run, err := s.store.GetRun(ctx, owner, runID)
	if err != nil {
		return Run{}, err
	}
	if isTerminalRunStatus(run.Status) {
		return run, nil
	}
	if s.pipeline == nil || s.runtime == nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "小说创作运行服务尚未配置", nil)
	}
	release, err := s.runtime.AcquireProject(ctx, run.ProjectID, run.ID, projectRunLockTTL)
	if err != nil {
		return run, err
	}
	defer release()
	project, err := s.store.GetProject(ctx, owner, run.ProjectID)
	if err != nil {
		return s.finishWithError(ctx, owner, run, ErrProjectNotFound, "作品不存在", err)
	}
	session, err := s.store.GetSession(ctx, owner, run.SessionID)
	if err != nil {
		return s.finishWithError(ctx, owner, run, ErrSessionNotFound, "创作会话不存在", err)
	}
	var prompt struct {
		UserMessage    string          `json:"user_message"`
		ChapterNumber  int             `json:"chapter_number"`
		PreviousTruth  json.RawMessage `json:"previous_truth"`
		PreviousMemory json.RawMessage `json:"previous_memory"`
	}
	if err := json.Unmarshal(run.PromptSnapshot, &prompt); err != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "创作输入快照损坏", err)
	}
	sink := serviceEventSink{service: s, owner: owner}
	bundle, runErr := s.pipeline.Run(ctx, RunContext{
		OwnerUsername: owner, Run: run, Project: project, Session: session,
		UserMessage: prompt.UserMessage, ChapterNumber: prompt.ChapterNumber,
		PreviousTruth: prompt.PreviousTruth, PreviousMemory: prompt.PreviousMemory, Runtime: s.runtime,
	}, sink)
	if runErr != nil {
		code, message := domainFailure(runErr)
		status := RunFailed
		if code == ErrRunCancelled || errors.Is(runErr, context.Canceled) {
			status = RunCancelled
		}
		return s.finish(ctx, owner, run, status, code, message, nil, runErr)
	}
	if cancelled, cancelErr := s.runtime.IsCancelled(ctx, run.ID); cancelErr != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "读取取消状态失败", cancelErr)
	} else if cancelled {
		cancelledErr := &DomainError{Code: ErrRunCancelled, Message: "小说创作已取消", Cause: context.Canceled}
		return s.finish(ctx, owner, run, RunCancelled, ErrRunCancelled, cancelledErr.Message, nil, cancelledErr)
	}
	if err := s.store.SaveChapterBundle(ctx, owner, bundle); err != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "保存章节失败", err)
	}
	status := bundle.FinalRunStatus
	if status != RunSucceededWithWarnings {
		status = RunSucceeded
	}
	message := "章节创作完成"
	if status == RunSucceededWithWarnings {
		message = "章节已生成，但仍有审核提醒"
	}
	return s.finish(ctx, owner, run, status, "", message, bundle.RuntimeSnapshot, nil)
}

func (s *Service) RetryRun(ctx context.Context, owner, runID string) (Run, error) {
	owner = strings.TrimSpace(owner)
	source, err := s.store.GetRun(ctx, owner, strings.TrimSpace(runID))
	if err != nil {
		return Run{}, err
	}
	if source.Status != RunFailed && source.Status != RunCancelled && source.Status != RunSucceededWithWarnings {
		return Run{}, &DomainError{Code: ErrInvalidTransition, Message: "当前任务状态不允许重试"}
	}
	key := "retry:" + source.ID
	if existing, found, err := s.store.FindRunByIdempotencyKey(ctx, owner, key); err != nil {
		return Run{}, err
	} else if found {
		return existing, nil
	}
	var snapshot map[string]any
	if err := json.Unmarshal(source.PromptSnapshot, &snapshot); err != nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "原任务输入快照损坏", Cause: err}
	}
	snapshot["retry_of_run_id"] = source.ID
	promptSnapshot, _ := json.Marshal(snapshot)
	retry, err := s.store.CreateRun(ctx, owner, Run{
		ProjectID: source.ProjectID, SessionID: source.SessionID, IdempotencyKey: key,
		Status: RunQueued, ModelID: source.ModelID, ThinkingLevel: source.ThinkingLevel, PromptSnapshot: promptSnapshot,
	})
	if err != nil {
		return Run{}, err
	}
	if _, err := s.appendEvent(ctx, owner, retry, "run.queued", RunQueued, "重试任务已加入队列", 0, map[string]any{"retryOfRunId": source.ID}); err != nil {
		return retry, err
	}
	if err := s.runtime.Enqueue(ctx, retry.ID); err != nil {
		failed, _ := s.finishWithError(ctx, owner, retry, ErrPersistenceFailed, "重试任务入队失败", err)
		return failed, &DomainError{Code: ErrPersistenceFailed, Message: "重试任务入队失败，请稍后重试", Cause: err}
	}
	return retry, nil
}

func (s *Service) CancelRun(ctx context.Context, owner, runID string) (Run, error) {
	owner = strings.TrimSpace(owner)
	run, err := s.store.GetRun(ctx, owner, strings.TrimSpace(runID))
	if err != nil {
		return Run{}, err
	}
	if isTerminalRunStatus(run.Status) {
		return run, nil
	}
	if err := s.runtime.Cancel(ctx, run.ID); err != nil {
		return run, err
	}
	return s.finish(ctx, owner, run, RunCancelled, ErrRunCancelled, "小说创作已取消", nil, nil)
}

func (s *Service) GetSnapshot(ctx context.Context, owner, runID string, afterEventID int64) (Snapshot, error) {
	owner = strings.TrimSpace(owner)
	run, err := s.store.GetRun(ctx, owner, strings.TrimSpace(runID))
	if err != nil {
		return Snapshot{}, err
	}
	project, err := s.store.GetProject(ctx, owner, run.ProjectID)
	if err != nil {
		return Snapshot{}, err
	}
	session, err := s.store.GetSession(ctx, owner, run.SessionID)
	if err != nil {
		return Snapshot{}, err
	}
	messages, err := s.store.ListMessages(ctx, owner, session.ID, 0)
	if err != nil {
		return Snapshot{}, err
	}
	events, err := s.store.ListRunEvents(ctx, owner, run.ID, afterEventID)
	if err != nil {
		return Snapshot{}, err
	}
	chapters, err := s.store.ListChapters(ctx, owner, run.ProjectID)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Project: project, Session: session, Run: run, Messages: messages, Events: events, Chapters: chapters}, nil
}

type serviceEventSink struct {
	service *Service
	owner   string
}

func (s serviceEventSink) Persist(ctx context.Context, event RunEvent) (RunEvent, error) {
	if event.Status != "" {
		current, err := s.service.store.GetRun(ctx, s.owner, event.RunID)
		if err != nil {
			return RunEvent{}, err
		}
		if current.Status != event.Status && CanTransition(current.Status, event.Status) {
			if _, err := s.service.store.UpdateRun(ctx, s.owner, RunUpdate{ID: event.RunID, Status: event.Status}); err != nil {
				return RunEvent{}, err
			}
		}
	}
	return s.service.store.AppendRunEvent(ctx, s.owner, event)
}

func (s serviceEventSink) Publish(ctx context.Context, event RunEvent) error {
	return s.service.runtime.Publish(ctx, event)
}

func (s *Service) appendEvent(ctx context.Context, owner string, run Run, eventType string, status RunStatus, message string, progress int, payload any) (RunEvent, error) {
	var raw json.RawMessage
	if payload != nil {
		raw, _ = json.Marshal(payload)
	}
	event, err := s.store.AppendRunEvent(ctx, owner, RunEvent{
		RunID: run.ID, SessionID: run.SessionID, ProjectID: run.ProjectID,
		Type: eventType, Step: string(status), Status: status, Message: message, Progress: progress, Payload: raw,
	})
	if err != nil {
		return RunEvent{}, err
	}
	if err := s.runtime.Publish(ctx, event); err != nil {
		return RunEvent{}, &DomainError{Code: ErrPersistenceFailed, Message: "发布任务状态失败", Cause: err}
	}
	return event, nil
}

func (s *Service) finishWithError(ctx context.Context, owner string, run Run, code, message string, cause error) (Run, error) {
	status := RunFailed
	if code == ErrRunCancelled {
		status = RunCancelled
	}
	domainErr := &DomainError{Code: code, Message: message, Cause: cause}
	return s.finish(ctx, owner, run, status, code, message, nil, domainErr)
}

func (s *Service) finish(ctx context.Context, owner string, run Run, status RunStatus, code, message string, runtime json.RawMessage, resultErr error) (Run, error) {
	finishedAt := time.Now().UTC()
	updated, err := s.store.UpdateRun(ctx, owner, RunUpdate{ID: run.ID, Status: status, RuntimeSnapshot: runtime, ErrorCode: code, ErrorMessage: message, FinishedAt: &finishedAt})
	if err != nil {
		return run, err
	}
	eventType := "run.succeeded"
	progress := 100
	if status == RunSucceededWithWarnings {
		eventType = "run.warning"
	} else if status == RunFailed {
		eventType = "run.failed"
	} else if status == RunCancelled {
		eventType = "run.cancelled"
	}
	if _, eventErr := s.appendEvent(ctx, owner, updated, eventType, status, message, progress, nil); eventErr != nil && resultErr == nil {
		resultErr = eventErr
	}
	return updated, resultErr
}

func domainFailure(err error) (string, string) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code, domainErr.Message
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrGenerationTimeout, "小说生成超时，请重试"
	}
	if errors.Is(err, context.Canceled) {
		return ErrRunCancelled, "小说创作已取消"
	}
	return ErrModelUnavailable, fmt.Sprintf("模型调用失败：%v", err)
}
