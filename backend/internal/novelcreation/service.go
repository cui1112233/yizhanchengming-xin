package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type PipelineRunner interface {
	Run(context.Context, RunContext, EventSink) (ChapterBundle, error)
}

type Snapshot struct {
	CanExportShort bool       `json:"canExportShort"`
	ExecutionState string     `json:"executionState,omitempty"`
	Project        Project    `json:"project"`
	Session        Session    `json:"session"`
	Run            Run        `json:"run"`
	Messages       []Message  `json:"messages"`
	Events         []RunEvent `json:"events"`
	Chapters       []Chapter  `json:"chapters"`
}

type Service struct {
	archiveProject  func(context.Context, string, string) error
	restoreProject  func(context.Context, string, string) error
	documents       DocumentObjectStore
	documentParser  DocumentParser
	automationClock func() time.Time
	store           Store
	runtime         Runtime
	pipeline        PipelineRunner
	control         *workerControl
}

func NewService(store Store, runtime Runtime, pipeline PipelineRunner) *Service {
	return &Service{store: store, runtime: runtime, pipeline: pipeline, control: newWorkerControl()}
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

func (s *Service) createSession(ctx context.Context, owner string, input CreateSessionInput) (Session, error) {
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

func (s *Service) sendMessage(ctx context.Context, owner string, input SendMessageInput) (Run, error) {
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
	project, err := s.store.GetProject(ctx, owner, input.ProjectID)
	if err != nil {
		return Run{}, err
	}
	selected, err := s.prepareGenerationInput(ctx, owner, input.ProjectID, input.SessionID, AutomationInput{
		ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID,
	})
	if err != nil {
		return Run{}, err
	}
	chapters, err := s.store.ListChapters(ctx, owner, input.ProjectID)
	if err != nil {
		return Run{}, err
	}
	chapterNumber := 1
	for _, chapter := range chapters {
		if chapter.Number >= chapterNumber {
			chapterNumber = chapter.Number + 1
		}
	}
	imitation, err := loadLongImitation(ctx, s.store, owner, input.ProjectID)
	if err != nil {
		return Run{}, err
	}
	if imitation != nil && session.Mode != longImitationKind {
		return Run{}, inputError("此作品按仿写长篇配置生产，请使用仿写长篇会话继续")
	}
	if imitation != nil {
		if imitation != nil && chapterNumber > imitation.Input.TargetChapters {
			return Run{}, &DomainError{Code: ErrInvalidTransition, Message: "已达到本作品目标章数，已停止继续生成"}
		}
	}
	continuity, err := s.loadContinuity(ctx, owner, input.ProjectID, chapters)
	if err != nil {
		return Run{}, err
	}
	var confirmedFoundation json.RawMessage
	if foundation, found, findErr := s.store.GetFoundation(ctx, owner, input.ProjectID); findErr != nil {
		return Run{}, findErr
	} else if found {
		confirmedFoundation = foundationPlanFromSnapshot(foundation.Snapshot)
	}
	checkpoint := len(confirmedFoundation) == 0 && modeUsesCheckpoint(session.Mode)
	message, err := s.store.AppendMessage(ctx, owner, Message{
		SessionID: input.SessionID, ProjectID: input.ProjectID, Role: "user", Content: input.Content,
		Status: "accepted", ModelID: input.ModelID, ThinkingLevel: input.ThinkingLevel,
	})
	if err != nil {
		return Run{}, err
	}
	settings, err := s.store.GetProjectSettings(ctx, owner, input.ProjectID)
	if err != nil {
		return Run{}, err
	}
	settings.Agent = NormalizeAgentSettings(settings.Agent, project.TargetChapterWords)
	if imitation != nil {
		settings.Agent.TargetChapterWords = imitation.Input.TargetWordCount
	}
	snapshotFields := map[string]any{
		"project_settings":       settings,
		"agent_contract_version": 1,
		"source_type":            input.SourceType, "source_plan_id": input.SourcePlanID, "source_round": input.SourceRound, "source_run_id": input.SourceRunID,
		"version": 1, "mode": session.Mode, "user_message": input.Content,
		"mode_inputs": selected.ModeInputs, "chapter_number": chapterNumber, "checkpoint": checkpoint,
		"knowledge_ids": selected.KnowledgeIDs, "knowledge_materials": selected.Knowledge,
		"note_ids": selected.NoteIDs, "notes": selected.Notes, "style_profile_id": selected.StyleProfileID, "style_profile": selected.StyleProfile,
		"previous_truth": continuity.Truth, "previous_memory": continuity.Memory,
		"previous_chapters": continuity.Chapters,
		"message_id":        message.ID,
	}
	if imitation != nil {
		snapshotFields["long_imitation"] = imitation
	}
	if len(confirmedFoundation) > 0 {
		snapshotFields["confirmed_foundation"] = json.RawMessage(confirmedFoundation)
	}
	promptSnapshot, _ := json.Marshal(snapshotFields)
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

	return run, nil
}

func (s *Service) runPipeline(ctx context.Context, owner, runID string) (Run, error) {
	owner, runID = strings.TrimSpace(owner), strings.TrimSpace(runID)
	run, err := s.store.GetRun(ctx, owner, runID)
	if err != nil {
		return Run{}, err
	}
	if isTerminalRunStatus(run.Status) {
		return run, nil
	}
	var shortSnapshot struct {
		Story *ShortStory `json:"short_story"`
	}
	if json.Unmarshal(run.PromptSnapshot, &shortSnapshot) == nil && shortSnapshot.Story != nil {
		return s.runShortStory(ctx, owner, run)
	}
	if s.pipeline == nil || s.runtime == nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "小说创作运行服务尚未配置", nil)
	}
	project, err := s.store.GetProject(ctx, owner, run.ProjectID)
	if err != nil {
		return s.finishWithError(ctx, owner, run, ErrProjectNotFound, "作品不存在", err)
	}
	session, err := s.store.GetSession(ctx, owner, run.SessionID)
	if err != nil {
		return s.finishWithError(ctx, owner, run, ErrSessionNotFound, "创作会话不存在", err)
	}
	var prompt struct {
		LongImitation        *LongImitationCreation `json:"long_imitation"`
		AgentContractVersion int                    `json:"agent_contract_version"`
		Settings             *ProjectSettings       `json:"project_settings"`
		StyleProfile         *string                `json:"style_profile"`
		Notes                []ProjectNote          `json:"notes"`
		UserMessage          string                 `json:"user_message"`
		Mode                 string                 `json:"mode"`
		ModeInputs           map[string]string      `json:"mode_inputs"`
		ChapterNumber        int                    `json:"chapter_number"`
		Checkpoint           bool                   `json:"checkpoint"`
		ConfirmedFoundation  json.RawMessage        `json:"confirmed_foundation"`
		KnowledgeIDs         []string               `json:"knowledge_ids"`
		KnowledgeMaterials   []KnowledgeMaterial    `json:"knowledge_materials"`
		ExpectedTarget       *PreviousChapter       `json:"expected_target"`
		RetryOfRunID         string                 `json:"retry_of_run_id"`
		PreviousChapters     []PreviousChapter      `json:"previous_chapters"`
		PreviousTruth        json.RawMessage        `json:"previous_truth"`
		PreviousMemory       json.RawMessage        `json:"previous_memory"`
	}
	if err := json.Unmarshal(run.PromptSnapshot, &prompt); err != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "创作输入快照损坏", err)
	}
	if err := s.validateTargetVersion(ctx, owner, run.ProjectID, prompt.ExpectedTarget); err != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "前文版本已变更，请重新确认创作", err)
	}
	if err := s.validateContinuity(ctx, owner, run.ProjectID, prompt.PreviousChapters, prompt.ChapterNumber, prompt.RetryOfRunID != ""); err != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "前文资料读取失败，已停止本次创作", err)
	}
	settings, err := s.store.GetProjectSettings(ctx, owner, run.ProjectID)
	if err != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "读取作品写作风格失败", err)
	}
	if prompt.Settings != nil {
		settings = *prompt.Settings
	}
	agent := NormalizeAgentSettings(settings.Agent, project.TargetChapterWords)
	if prompt.LongImitation != nil {
		agent.TargetChapterWords = prompt.LongImitation.Input.TargetWordCount
	}
	var effectiveAgent *AgentSettings
	if prompt.AgentContractVersion > 0 {
		effectiveAgent = &agent
	} else {
		var legacy struct {
			Settings struct {
				Agent struct {
					Rounds int `json:"maxRevisionRounds"`
				} `json:"agent"`
			} `json:"project_settings"`
		}
		_ = json.Unmarshal(run.PromptSnapshot, &legacy)
		agent.MaxRevisionRounds = legacy.Settings.Agent.Rounds
		if agent.MaxRevisionRounds < 1 {
			agent.MaxRevisionRounds = 2
		}
		if agent.MaxRevisionRounds > 5 {
			agent.MaxRevisionRounds = 5
		}
	}
	project.TargetChapterWords = agent.TargetChapterWords
	if strings.TrimSpace(run.ThinkingLevel) == "" {
		run.ThinkingLevel = agent.ThinkingLevel
	}
	styleProfile := settings.StyleProfile
	if prompt.StyleProfile != nil {
		styleProfile = *prompt.StyleProfile
	}
	sink := serviceEventSink{service: s, owner: owner}
	var imitationAnalysis *LongImitationAnalysis
	var imitationRequirements *LongImitationRequirements
	if prompt.LongImitation != nil {
		var a LongImitationAnalysis
		found, e := readContent(ctx, s.store, owner, run.ProjectID, longImitationKind, "analysis", &a)
		if e != nil {
			return s.recoverDurableWriteFailure(ctx, owner, run, e)
		}
		if found {
			imitationAnalysis = &a
		}
		var r LongImitationRequirements
		found, e = readContent(ctx, s.store, owner, run.ProjectID, longImitationKind, "requirements", &r)
		if e != nil {
			return s.recoverDurableWriteFailure(ctx, owner, run, e)
		}
		if found {
			imitationRequirements = &r
		}
	}
	bundle, runErr := s.pipeline.Run(ctx, RunContext{
		LongImitation: prompt.LongImitation, ImitationAnalysis: imitationAnalysis, ImitationRequirements: imitationRequirements,
		OwnerUsername: owner, Run: run, Project: project, Session: session,
		UserMessage: prompt.UserMessage, ChapterNumber: prompt.ChapterNumber,
		PreviousTruth: prompt.PreviousTruth, PreviousMemory: prompt.PreviousMemory, PreviousChapters: prompt.PreviousChapters, Runtime: s.runtime,
		ModeInputs: prompt.ModeInputs, StyleProfile: styleProfile, Rules: settings.Rules, Notes: prompt.Notes,
		Checkpoint:          prompt.Checkpoint && len(prompt.ConfirmedFoundation) == 0,
		ConfirmedFoundation: prompt.ConfirmedFoundation,
		KnowledgeMaterials:  prompt.KnowledgeMaterials,
		MaxRevisionRounds:   agent.MaxRevisionRounds,
		Agent:               effectiveAgent,
	}, sink)
	if runErr != nil {
		if IsCode(runErr, ErrSecondaryConfirmation) {
			return s.finishModelCallConfirmation(ctx, owner, run)
		}
		if isDurableWriteFailure(runErr) || IsCode(runErr, ErrPersistenceFailed) {
			return s.recoverDurableWriteFailure(ctx, owner, run, runErr)
		}
		code, message := domainFailure(runErr)
		status := RunFailed
		if code == ErrResultUnknown {
			status = RunPaused
		}
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
	if effectiveAgent != nil {
		execution, _, loadErr := s.store.GetExecution(ctx, owner, run.ID)
		if loadErr != nil {
			return s.recoverDurableWriteFailure(ctx, owner, run, loadErr)
		}
		var runtimeFields map[string]any
		_ = json.Unmarshal(bundle.RuntimeSnapshot, &runtimeFields)
		if runtimeFields == nil {
			runtimeFields = map[string]any{}
		}
		runtimeFields["agentSettings"] = agent
		runtimeFields["modelCalls"] = callMetadata(execution)
		bundle.RuntimeSnapshot, _ = json.Marshal(runtimeFields)
	}
	if bundle.FinalRunStatus == RunAwaitingConfirmation {
		return s.finishCheckpoint(ctx, owner, run, bundle)
	}
	// A regenerated target is a new immutable chapter version associated with this run.
	if prompt.ExpectedTarget != nil {
		bundle.Chapter.Version = prompt.ExpectedTarget.Version + 1
		bundle.ChapterVersion.Version = bundle.Chapter.Version
	}
	if err := persistBundle(ctx, bundle); err != nil {
		return run, err
	}
	return s.savePending(ctx, owner, run, bundle)
}

func (s *Service) retryRun(ctx context.Context, owner, runID string) (Run, error) {
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
	if _, err := writableSession(ctx, s.store, owner, source.ProjectID, source.SessionID); err != nil {
		return Run{}, err
	}
	if operationKind(source) != "" {
		var snap maintenanceSnapshot
		if e := json.Unmarshal(source.PromptSnapshot, &snap); e != nil {
			return Run{}, e
		}
		if e := s.validateMaintenanceSources(ctx, owner, source.ProjectID, snap); e != nil {
			return Run{}, e
		}
		retry, e := s.store.CreateRun(ctx, owner, Run{ProjectID: source.ProjectID, SessionID: source.SessionID, IdempotencyKey: key, Status: RunQueued, ModelID: source.ModelID, ThinkingLevel: source.ThinkingLevel, PromptSnapshot: cloneRawJSON(source.PromptSnapshot)})
		if e != nil {
			return retry, e
		}
		if e = s.trackStyleRetry(ctx, owner, source, retry, snap); e != nil {
			return retry, e
		}
		_, e = s.appendEvent(ctx, owner, retry, "run.queued", RunQueued, "分析重试已加入队列", 0, map[string]any{"retryOfRunId": source.ID})
		return retry, e
	}
	var snapshot map[string]any
	if err := json.Unmarshal(source.PromptSnapshot, &snapshot); err != nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "原任务输入快照损坏", Cause: err}
	}
	var original struct {
		ChapterNumber int `json:"chapter_number"`
	}
	if err := json.Unmarshal(source.PromptSnapshot, &original); err != nil {
		return Run{}, continuityError(err)
	}
	chapters, err := s.store.ListChapters(ctx, owner, source.ProjectID)
	if err != nil {
		return Run{}, continuityError(err)
	}
	snapshot["expected_target"] = PreviousChapter{Number: original.ChapterNumber, Version: 0}
	for _, chapter := range chapters {
		if chapter.Number == original.ChapterNumber {
			snapshot["expected_target"] = PreviousChapter{ChapterID: chapter.ID, Version: chapter.Version, Number: chapter.Number, Ending: chapter.Content}
		}
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

	return retry, nil
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
	execution, _, err := s.store.GetExecution(ctx, owner, run.ID)
	if err != nil {
		return Snapshot{}, err
	}
	canExport, err := s.canExportShort(ctx, owner, run, session)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{CanExportShort: canExport, ExecutionState: execution.State, Project: project, Session: session, Run: run, Messages: messages, Events: events, Chapters: chapters}, nil
}

func (s *Service) canExportShort(ctx context.Context, owner string, run Run, session Session) (bool, error) {
	if run.ID == "" || (session.Mode != "short_original" && session.Mode != "short_imitation") || operationKind(run) != "" || (run.Status != RunSucceeded && run.Status != RunSucceededWithWarnings) {
		return false, nil
	}
	saved, err := s.store.HasSavedRun(ctx, owner, run.ID)
	if err != nil {
		return false, persistenceError(err)
	}
	return saved, nil
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
	_ = s.service.runtime.Publish(ctx, event)
	return nil
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
	_ = s.runtime.Publish(ctx, event)
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
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, leaseLost()
	}
	now := time.Now().UTC()
	var finishedAt *time.Time
	if isTerminalRunStatus(status) {
		finishedAt = &now
	}
	eventType := "run.succeeded"
	switch status {
	case RunSucceededWithWarnings:
		eventType = "run.warning"
	case RunFailed:
		eventType = "run.failed"
	case RunCancelled:
		eventType = "run.cancelled"
	case RunPaused:
		eventType = "run.paused"
	}
	updated, event, err := g.finishRun(ctx, owner, RunUpdate{ID: run.ID, Status: status, RuntimeSnapshot: runtime, ErrorCode: code, ErrorMessage: message, FinishedAt: finishedAt}, nil, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: eventType, Status: status, Message: message, Progress: 100})
	if err != nil {
		return run, err
	}
	_ = s.runtime.Publish(ctx, event)
	if updated.Status == RunPaused && updated.ErrorCode == ErrResultUnknown {
		return updated, unknownResult()
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

// finishCheckpoint persists the generated blueprint and leaves the run in
// awaiting_confirmation so the user can confirm or edit it.
func (s *Service) finishCheckpoint(ctx context.Context, owner string, run Run, bundle ChapterBundle) (Run, error) {
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, leaseLost()
	}
	var input struct {
		Creation *LongImitationCreation `json:"long_imitation"`
	}
	_ = json.Unmarshal(run.PromptSnapshot, &input)
	updated, event, err := g.finishRun(ctx, owner, RunUpdate{ID: run.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: bundle.RuntimeSnapshot}, &bundle.Foundation, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.checkpoint", Status: RunAwaitingConfirmation, Message: "创作蓝图已生成，等待确认", Progress: 20}, func(tx Store) error {
		return applyImitationTitle(ctx, tx, owner, run, bundle.Foundation, input.Creation)
	})
	if err != nil {
		return run, err
	}
	_ = s.runtime.Publish(ctx, event)
	return updated, nil
}

// ConfirmRun resumes a checkpointed run. An edited blueprint is validated and
// persisted before the continuation run is queued.
func (s *Service) confirmRun(ctx context.Context, owner, runID string, blueprint json.RawMessage) (Run, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作运行服务尚未配置"}
	}
	owner, runID = strings.TrimSpace(owner), strings.TrimSpace(runID)
	source, err := s.store.GetRun(ctx, owner, runID)
	if err != nil {
		return Run{}, err
	}
	if e, found, loadErr := s.store.GetExecution(ctx, owner, source.ID); loadErr != nil {
		return Run{}, loadErr
	} else if found && e.PendingConfirmation != nil {
		return Run{}, &DomainError{Code: ErrInputRequired, Message: "请提交当前模型调用的确认编号"}
	}
	key := "confirm:" + source.ID
	if existing, found, err := s.store.FindRunByIdempotencyKey(ctx, owner, key); err != nil {
		return Run{}, err
	} else if found {
		return existing, nil
	}
	if _, err := writableSession(ctx, s.store, owner, source.ProjectID, source.SessionID); err != nil {
		return Run{}, err
	}
	if source.Status != RunAwaitingConfirmation {
		return Run{}, &DomainError{Code: ErrInvalidTransition, Message: "当前任务不在蓝图确认状态"}
	}
	var snapshot map[string]any
	if err := json.Unmarshal(source.PromptSnapshot, &snapshot); err != nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "原任务输入快照损坏", Cause: err}
	}
	confirmed := blueprint
	if len(confirmed) > 0 {
		if err := validateFoundationDocument(confirmed); err != nil {
			return Run{}, err
		}
		if err := validateImitationConfirmation(source.PromptSnapshot, confirmed); err != nil {
			return Run{}, err
		}
		now := time.Now().UTC()
		foundationSnapshot, marshalErr := json.Marshal(map[string]any{"plan": json.RawMessage(confirmed)})
		if marshalErr != nil {
			return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "创作蓝图保存失败", Cause: marshalErr}
		}
		if err := s.store.SaveFoundation(ctx, owner, Foundation{
			ID: source.ProjectID + "-foundation", ProjectID: source.ProjectID,
			Snapshot: foundationSnapshot, CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			return Run{}, err
		}
	} else {
		confirmed = foundationPlanFromRuntime(source.RuntimeSnapshot)
		if len(confirmed) == 0 {
			return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "创作蓝图缺失，请重新生成"}
		}
	}
	snapshot["confirmed_foundation"] = json.RawMessage(confirmed)
	snapshot["checkpoint"] = false
	snapshot["confirmed_from_run_id"] = source.ID
	promptSnapshot, err := json.Marshal(snapshot)
	if err != nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "确认快照保存失败", Cause: err}
	}
	run, err := s.store.CreateRun(ctx, owner, Run{
		ProjectID: source.ProjectID, SessionID: source.SessionID, IdempotencyKey: key,
		Status: RunQueued, ModelID: source.ModelID, ThinkingLevel: source.ThinkingLevel, PromptSnapshot: promptSnapshot,
	})
	if err != nil {
		if IsCode(err, ErrRunInProgress) {
			if existing, found, findErr := s.store.FindRunByIdempotencyKey(ctx, owner, key); findErr == nil && found {
				return existing, nil
			}
		}
		return Run{}, err
	}
	if _, updateErr := s.store.UpdateRun(ctx, owner, RunUpdate{ID: source.ID, Status: RunSucceeded}); updateErr != nil {
		return run, updateErr
	}
	if _, err := s.appendEvent(ctx, owner, run, "run.queued", RunQueued, "蓝图已确认，开始写正文", 0, map[string]any{"confirmedFromRunId": source.ID}); err != nil {
		return run, err
	}

	return run, nil
}

func (s *Service) GetProjectSettings(ctx context.Context, owner, projectID string) (ProjectSettings, error) {
	if s == nil || s.store == nil {
		return ProjectSettings{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	settings, err := s.store.GetProjectSettings(ctx, owner, projectID)
	if err != nil {
		return ProjectSettings{}, err
	}
	project, err := s.store.GetProject(ctx, owner, projectID)
	if err != nil {
		return ProjectSettings{}, err
	}
	settings.Agent = NormalizeAgentSettings(settings.Agent, project.TargetChapterWords)
	creation, e := loadLongImitation(ctx, s.store, owner, projectID)
	if e != nil {
		return ProjectSettings{}, e
	}
	if creation != nil {
		settings.Agent.TargetChapterWords = creation.Input.TargetWordCount
	}
	return settings, nil
}

func (s *Service) updateProjectSettings(ctx context.Context, owner, projectID string, input ProjectSettings) (ProjectSettings, error) {
	if s == nil || s.store == nil {
		return ProjectSettings{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	project, err := s.store.GetProject(ctx, owner, projectID)
	if err != nil {
		return ProjectSettings{}, err
	}
	if err := validateAgentRequirements(input.Agent); err != nil {
		return ProjectSettings{}, err
	}
	agent := NormalizeAgentSettings(input.Agent, project.TargetChapterWords)
	creation, e := loadLongImitation(ctx, s.store, owner, projectID)
	if e != nil {
		return ProjectSettings{}, e
	}
	if creation != nil {
		agent.TargetChapterWords = creation.Input.TargetWordCount
	}
	return s.store.UpdateProjectSettings(ctx, owner, projectID, ProjectSettings{
		StyleProfile: strings.TrimSpace(input.StyleProfile),
		Rules:        normalizeRules(input.Rules),
		Agent:        agent,
	})
}

func selectKnowledgeMaterials(materials []Material, ids []string) ([]KnowledgeMaterial, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	byID := make(map[string]Material, len(materials))
	for _, material := range materials {
		byID[material.ID] = material
	}
	selected := make([]KnowledgeMaterial, 0, len(ids))
	for _, id := range ids {
		material, ok := byID[id]
		if !ok {
			return nil, &DomainError{Code: ErrInputRequired, Message: "所选素材不存在或不属于当前作品"}
		}
		selected = append(selected, KnowledgeMaterial{
			ID: material.ID, Title: material.Title,
			Content: truncateRunes(material.Content, maxKnowledgeMaterialRunes),
		})
	}
	return selected, nil
}

func (s *Service) ListMaterials(ctx context.Context, owner, projectID string) ([]Material, error) {
	if s == nil || s.store == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.ListMaterials(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
}

func (s *Service) createMaterial(ctx context.Context, owner, projectID string, input CreateMaterialInput) (Material, error) {
	if s == nil || s.store == nil {
		return Material{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.CreateMaterial(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), input)
}

func (s *Service) deleteMaterial(ctx context.Context, owner, projectID, materialID string) error {
	if s == nil || s.store == nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.DeleteMaterial(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(materialID))
}

func (s *Service) ListNotes(ctx context.Context, owner, projectID string) ([]ProjectNote, error) {
	if s == nil || s.store == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.ListNotes(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
}

func (s *Service) createNote(ctx context.Context, owner, projectID string, input ProjectNoteInput) (ProjectNote, error) {
	if s == nil || s.store == nil {
		return ProjectNote{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.CreateNote(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), input)
}

func (s *Service) updateNote(ctx context.Context, owner, projectID, noteID string, input ProjectNoteInput) (ProjectNote, error) {
	if s == nil || s.store == nil {
		return ProjectNote{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.UpdateNote(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(noteID), input)
}

func (s *Service) deleteNote(ctx context.Context, owner, projectID, noteID string) error {
	if s == nil || s.store == nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.DeleteNote(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(noteID))
}

// GetDashboard aggregates chapter volume and run outcomes for the data panel.
func (s *Service) GetDashboard(ctx context.Context, owner, projectID string) (Dashboard, error) {
	if s == nil || s.store == nil {
		return Dashboard{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	if _, err := s.store.GetProject(ctx, owner, projectID); err != nil {
		return Dashboard{}, err
	}
	chapters, err := s.store.ListChapters(ctx, owner, projectID)
	if err != nil {
		return Dashboard{}, err
	}
	runs, err := s.store.ListRuns(ctx, owner, projectID)
	if err != nil {
		return Dashboard{}, err
	}
	dashboard := Dashboard{ProjectID: projectID, ChapterCount: len(chapters), RunCounts: make(map[string]int)}
	for _, chapter := range chapters {
		dashboard.TotalWords += chapter.WordCount
	}
	terminal, successful := 0, 0
	for _, run := range runs {
		dashboard.RunTotal++
		dashboard.RunCounts[string(run.Status)]++
		switch run.Status {
		case RunSucceeded, RunSucceededWithWarnings:
			terminal++
			successful++
		case RunFailed, RunCancelled:
			terminal++
		}
		if dashboard.LastRunAt == nil || run.CreatedAt.After(*dashboard.LastRunAt) {
			created := run.CreatedAt
			dashboard.LastRunAt = &created
		}
	}
	if terminal > 0 {
		dashboard.SuccessRate = float64(successful) * 100 / float64(terminal)
	}
	return dashboard, nil
}

func normalizeRules(rules []string) []string {
	normalized := make([]string, 0, len(rules))
	for _, rule := range rules {
		if value := strings.TrimSpace(rule); value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

// ExportDocument is a rendered whole-book download.
type ExportDocument struct {
	Filename    string
	ContentType string
	Content     string
}

// ExportProject renders every stored chapter in reading order.
func (s *Service) ExportProject(ctx context.Context, owner, projectID, format string) (ExportDocument, error) {
	if s == nil || s.store == nil {
		return ExportDocument{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	project, err := s.store.GetProject(ctx, owner, projectID)
	if err != nil {
		return ExportDocument{}, err
	}
	chapters, err := s.store.ListChapters(ctx, owner, projectID)
	if err != nil {
		return ExportDocument{}, err
	}
	hasContent := false
	for _, c := range chapters {
		if strings.TrimSpace(c.Content) != "" {
			hasContent = true
		}
	}
	if !hasContent {
		return ExportDocument{}, inputError("暂无可导出正文")
	}
	normalized := strings.ToLower(strings.TrimSpace(format))
	switch normalized {
	case "", "markdown", "md":
		normalized = "markdown"
	case "txt", "text":
		normalized = "txt"
	default:
		return ExportDocument{}, &DomainError{Code: ErrInputRequired, Message: "导出格式仅支持 markdown 或 txt"}
	}
	var builder strings.Builder
	if normalized == "markdown" {
		builder.WriteString("# " + project.Title + "\n\n")
	}
	for _, chapter := range chapters {
		fmt.Fprintf(&builder, "第%d章 %s\n\n%s\n\n", chapter.Number, chapter.Title, strings.TrimSpace(chapter.Content))
	}
	content := builder.String()
	if strings.TrimSpace(content) == "" {
		content = "（暂无章节内容）\n"
	}
	document := ExportDocument{Filename: sanitizeExportFilename(project.Title) + ".md", ContentType: "text/markdown; charset=utf-8", Content: content}
	if normalized == "txt" {
		document.Filename = sanitizeExportFilename(project.Title) + ".txt"
		document.ContentType = "text/plain; charset=utf-8"
	}
	return document, nil
}

func sanitizeExportFilename(title string) string {
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_", "\n", "_", "\r", "_")
	name := strings.TrimSpace(replacer.Replace(strings.TrimSpace(title)))
	if name == "" {
		return "小说"
	}
	return name
}

func foundationPlanFromSnapshot(raw json.RawMessage) json.RawMessage {
	var wrapper struct {
		Plan json.RawMessage `json:"plan"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &wrapper) != nil || !json.Valid(wrapper.Plan) {
		return nil
	}
	return wrapper.Plan
}

func foundationPlanFromRuntime(raw json.RawMessage) json.RawMessage {
	var wrapper struct {
		FoundationPlan json.RawMessage `json:"foundationPlan"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &wrapper) != nil || !json.Valid(wrapper.FoundationPlan) {
		return nil
	}
	return wrapper.FoundationPlan
}

func validateFoundationDocument(raw json.RawMessage) error {
	var value foundationPlanOutput
	if err := decodeAndValidate(string(raw), &value, func() error { return validateFoundation(value) }); err != nil {
		return &DomainError{Code: ErrInputRequired, Message: "创作蓝图格式无效", Cause: err}
	}
	return nil
}
