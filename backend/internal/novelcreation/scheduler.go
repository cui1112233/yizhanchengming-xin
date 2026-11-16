package novelcreation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DispatchedSchedule is one due schedule that produced a run.
type DispatchedSchedule struct {
	Schedule Schedule
	Run      Run
}

func (s *Service) ListSchedules(ctx context.Context, owner, projectID string) ([]Schedule, error) {
	if s == nil || s.store == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	plans, err := s.store.ListSchedules(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
	if err != nil {
		return nil, err
	}
	for i := range plans {
		plans[i].day(s.automationNow())
	}
	return plans, nil
}

func (s *Service) createSchedule(ctx context.Context, owner, projectID string, input ScheduleInput) (Schedule, error) {
	if s == nil || s.store == nil {
		return Schedule{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	if found, ok, err := s.findScheduleCreation(ctx, owner, projectID, input); err != nil || ok {
		return found, err
	}
	creationID, creation, err := automationCreation(owner, projectID, "schedule", input.IdempotencyKey, input)
	if err != nil {
		return Schedule{}, err
	}
	input, err = NormalizeScheduleInput(input)
	if err != nil {
		return Schedule{}, err
	}
	project, err := s.store.GetProject(ctx, owner, projectID)
	if err != nil {
		return Schedule{}, err
	}
	session, err := s.store.GetSession(ctx, owner, input.SessionID)
	if err != nil {
		return Schedule{}, err
	}
	if session.ProjectID != projectID {
		return Schedule{}, &DomainError{Code: ErrInputRequired, Message: "会话与作品不匹配"}
	}
	selected, err := s.prepareGenerationInput(ctx, owner, projectID, input.SessionID, AutomationInput{ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID})
	if err != nil {
		return Schedule{}, err
	}
	modelID := input.ModelID
	if modelID == "" {
		modelID = project.ModelID
	}
	now := s.automationNow()
	schedule := Schedule{
		Creation:        creation,
		AutomationState: AutomationState{Status: "not_started", DailyChapterCap: input.DailyChapterCap, MaxConsecutiveFailures: input.MaxConsecutiveFailures},
		ModeInputs:      selected.ModeInputs, KnowledgeIDs: selected.KnowledgeIDs, NoteIDs: selected.NoteIDs, StyleProfileID: selected.StyleProfileID,
		ID: novelID("schedule"), ProjectID: projectID, SessionID: input.SessionID,
		IntervalMinutes: input.IntervalMinutes, Enabled: input.Enabled, Prompt: input.Prompt,
		ModelID: modelID, ThinkingLevel: input.ThinkingLevel,
		NextRunAt: now.Add(time.Duration(input.IntervalMinutes) * time.Minute),
	}
	if creation != nil {
		schedule.ID = creationID
	}
	return s.store.CreateSchedule(ctx, owner, schedule)
}

func (s *Service) updateSchedule(ctx context.Context, owner, projectID, scheduleID string, input ScheduleInput) (Schedule, error) {
	if s == nil || s.store == nil {
		return Schedule{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID, scheduleID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(scheduleID)
	input, err := NormalizeScheduleInput(input)
	if err != nil {
		return Schedule{}, err
	}
	schedule, err := s.store.GetSchedule(ctx, owner, scheduleID)
	if err != nil {
		return Schedule{}, err
	}
	if schedule.ProjectID != projectID {
		return Schedule{}, &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	session, err := s.store.GetSession(ctx, owner, input.SessionID)
	if err != nil {
		return Schedule{}, err
	}
	if session.ProjectID != projectID {
		return Schedule{}, &DomainError{Code: ErrInputRequired, Message: "会话与作品不匹配"}
	}
	selected, err := s.prepareGenerationInput(ctx, owner, projectID, input.SessionID, AutomationInput{ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID})
	if err != nil {
		return Schedule{}, err
	}
	modelID := input.ModelID
	if modelID == "" {
		modelID = schedule.ModelID
	}
	now := s.automationNow()
	schedule.ConfigurationError = ""
	schedule.DailyChapterCap = input.DailyChapterCap
	schedule.MaxConsecutiveFailures = input.MaxConsecutiveFailures
	if input.Enabled {
		schedule.Status = "not_started"
		schedule.StoppedReason = ""
	} else {
		schedule.Status = "paused"
		schedule.StoppedReason = "用户暂停"
	}
	if schedule.SessionID == input.SessionID {
		preserved := map[string]string{}
		known := map[string]bool{}
		for _, field := range ModeInputFields(session.Mode) {
			known[field.Key] = true
		}
		for key, value := range schedule.ModeInputs {
			if !known[key] {
				preserved[key] = value
			}
		}
		for key, value := range selected.ModeInputs {
			preserved[key] = value
		}
		if len(preserved) == 0 {
			schedule.ModeInputs = nil
		} else {
			schedule.ModeInputs = preserved
		}
	} else {
		schedule.ModeInputs = selected.ModeInputs
	}
	schedule.KnowledgeIDs = selected.KnowledgeIDs
	schedule.NoteIDs = selected.NoteIDs
	schedule.StyleProfileID = selected.StyleProfileID
	schedule.SessionID = input.SessionID
	schedule.IntervalMinutes = input.IntervalMinutes
	schedule.Enabled = input.Enabled
	schedule.Prompt = input.Prompt
	schedule.ModelID = modelID
	schedule.ThinkingLevel = input.ThinkingLevel
	schedule.NextRunAt = now.Add(time.Duration(input.IntervalMinutes) * time.Minute)
	return s.store.UpdateSchedule(ctx, owner, schedule)
}

func (s *Service) UpdateSchedule(ctx context.Context, owner, project, id string, input ScheduleInput) (Schedule, error) {
	if s == nil || s.store == nil {
		return Schedule{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, project, id = strings.TrimSpace(owner), strings.TrimSpace(project), strings.TrimSpace(id)
	var out Schedule
	err := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		var e error
		out, e = s.automationChild(tx).updateSchedule(ctx, owner, project, id, input)
		return e
	})
	return out, err
}

func (s *Service) deleteSchedule(ctx context.Context, owner, projectID, scheduleID string) error {
	if s == nil || s.store == nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	return s.store.DeleteSchedule(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(scheduleID))
}

// RunSchedule dispatches one schedule immediately, regardless of its due time.
func (s *Service) RunSchedule(ctx context.Context, owner, projectID, scheduleID string) (Run, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return Run{}, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作运行服务尚未配置"}
	}
	owner = strings.TrimSpace(owner)
	schedule, err := s.store.GetSchedule(ctx, owner, strings.TrimSpace(scheduleID))
	if err != nil {
		return Run{}, err
	}
	if projectID = strings.TrimSpace(projectID); projectID != "" && schedule.ProjectID != projectID {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	return s.dispatchSchedule(ctx, owner, schedule, s.automationNow(), true)
}

// DispatchDueSchedules runs every due schedule once and advances its next run
// time. Individual failures are collected so one broken schedule cannot block
// the others.
func (s *Service) DispatchDueSchedules(ctx context.Context, now time.Time) ([]DispatchedSchedule, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作运行服务尚未配置"}
	}
	// Capture all pages before processing: settlement may update due dates and
	// enablement, so offset pagination over a mutating candidate set is unsafe.
	var due []Schedule
	var cursor ScheduleCursor
	for {
		page, err := s.store.ListDueSchedulePage(ctx, now, cursor, 20)
		if err != nil {
			return nil, err
		}
		due = append(due, page...)
		if len(page) < 20 {
			break
		}
		last := page[len(page)-1]
		cursor = ScheduleCursor{NextRunAt: last.NextRunAt, OwnerUsername: last.OwnerUsername, ID: last.ID}
	}
	dispatched := make([]DispatchedSchedule, 0, len(due))
	failures := make([]error, 0)
	for _, schedule := range due {
		run, dispatchErr := s.dispatchSchedule(ctx, schedule.OwnerUsername, schedule, now)
		if dispatchErr != nil {
			failures = append(failures, fmt.Errorf("schedule %s: %w", schedule.ID, dispatchErr))
			continue
		}
		if run.ID != "" {
			dispatched = append(dispatched, DispatchedSchedule{Schedule: schedule, Run: run})
		}
	}
	return dispatched, errors.Join(failures...)
}

func (s *Service) dispatchSchedule(ctx context.Context, owner string, stale Schedule, now time.Time, manual ...bool) (Run, error) {
	var out Run
	var observed error
	immediate := len(manual) > 0 && manual[0]
	err := s.store.Atomic(ctx, owner, stale.ProjectID, func(tx Store) error {
		child := s.automationChild(tx)
		p, e := tx.GetSchedule(ctx, owner, stale.ID)
		if e != nil {
			return e
		}
		p.defaults()
		p.day(now)
		save := func() error { _, e := tx.UpdateSchedule(ctx, owner, p); return e }
		if p.IntervalMinutes < 1 || p.IntervalMinutes > 1440 || p.DailyChapterCap < 1 || p.DailyChapterCap > 100 || p.MaxConsecutiveFailures < 1 || p.MaxConsecutiveFailures > 10 {
			p.Enabled = false
			p.Status = "paused"
			p.ConfigurationError = "配置需调整：定时间隔、每日上限或失败阈值超出范围"
			p.StoppedReason = p.ConfigurationError
			return save()
		}
		restoreLegacyScheduleRound(&p)
		if r := p.active(); r != nil {
			outcome, e := child.inspectRound(ctx, owner, r)
			if e != nil {
				return e
			}
			if outcome == "waiting" {
				p.WaitingReason = automationWaitReason(r.State)
				if p.Enabled && r.State == "awaiting_confirmation" {
					p.Status = "waiting"
				}
				if immediate {
					observed = &DomainError{Code: ErrRunInProgress, Message: "该作品正在创作，等待空闲"}
				}
				return save()
			}
			if outcome == "halt" && r.ErrorCode == ErrResultUnknown {
				p.Enabled = false
				p.Status = "paused"
				p.StoppedReason = "上游结果待确认，已暂停自动重试"
				p.WaitingReason = p.StoppedReason
				r.State = "paused"
				if immediate {
					observed = unknownResult()
				}
				return save()
			}
			if outcome == "success" {
				p.TotalSavedChapters++
				if r.FinishedAt == nil || r.FinishedAt.In(shanghai).Format("2006-01-02") == p.SuccessDay {
					p.DailySavedChapters++
				}
				p.ConsecutiveFailures = 0
				r.State = "saved"
			} else if outcome == "cancelled" {
				r.State = outcome
				p.Enabled = false
				p.Status = "stopped"
				p.StoppedReason = "用户停止"
			} else {
				r.State = outcome
				p.ConsecutiveFailures++
				p.LastFailure = r.ErrorMessage
				if outcome != "failure" || p.ConsecutiveFailures >= p.MaxConsecutiveFailures {
					p.Enabled = false
					if p.Status != "stopped" {
						p.Status = "paused"
					}
					p.StoppedReason = r.ErrorMessage
					if p.StoppedReason == "" {
						p.StoppedReason = "任务已暂停，请检查运行结果"
					}
				}
			}
			finished := now.UTC()
			r.FinishedAt = &finished
			if !r.Immediate {
				p.NextRunAt = finished.Add(time.Duration(p.IntervalMinutes) * time.Minute)
			}
			p.ActiveRound = ""
			p.WaitingReason = ""
			if p.Enabled {
				p.Status = "waiting"
			}
			if !immediate {
				return save()
			}
		}
		if !immediate && (!p.Enabled || p.NextRunAt.After(now)) {
			return nil
		}
		if immediate && (p.Status == "stopped" || p.ConfigurationError != "") {
			observed = &DomainError{Code: ErrInvalidTransition, Message: "请先调整配置并恢复定时任务"}
			return nil
		}
		if p.DailySavedChapters >= p.DailyChapterCap {
			local := now.In(shanghai)
			midnight := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, shanghai).UTC()
			p.EligibleAt = &midnight
			p.WaitingReason = "已达到每日章节上限"
			if p.Enabled {
				p.Status = "waiting"
			}
			if immediate {
				observed = &DomainError{Code: ErrInvalidTransition, Message: p.WaitingReason}
			}
			return save()
		}
		key := p.NextRunAt.UTC().Format(time.RFC3339Nano)
		planned := p.NextRunAt
		if immediate {
			key = novelID("manual")
			planned = now.UTC()
		}
		sourceType := "schedule"
		if immediate {
			sourceType = "schedule_immediate"
		}
		run, e := child.queueAutomationMessage(ctx, owner, SendMessageInput{SourceType: sourceType, SourcePlanID: p.ID, SourceRound: key, SessionID: p.SessionID, ProjectID: p.ProjectID, Content: p.Prompt, ModelID: p.ModelID, ThinkingLevel: p.ThinkingLevel, ModeInputs: p.ModeInputs, KnowledgeIDs: p.KnowledgeIDs, NoteIDs: p.NoteIDs, StyleProfileID: p.StyleProfileID, IdempotencyKey: "schedule:" + p.ID + ":" + key})
		if e != nil {
			if busy(e) {
				p.WaitingReason = "该作品正在创作，等待空闲"
				if p.Enabled {
					p.Status = "waiting"
				}
				if immediate {
					observed = e
				}
				return save()
			}
			if IsCode(e, ErrPersistenceFailed) || IsCode(e, ErrIdempotencyConflict) {
				return e
			}
			observed = e
			p.Enabled = false
			p.Status = "paused"
			p.ConfigurationError = "配置需调整：" + e.Error()
			p.StoppedReason = p.ConfigurationError
			return save()
		}
		if e = tx.InsertAutomationRound(ctx, owner, "schedule", p.ID, key, run.ID, planned); e != nil {
			return e
		}
		p.Rounds = append(p.Rounds, AutomationRound{Key: key, PlannedAt: planned, Immediate: immediate, RunID: run.ID, State: "queued"})
		p.ActiveRound = key
		p.LastRunID = run.ID
		started := now.UTC()
		p.LastRunAt = &started
		p.WaitingReason = ""
		if p.Enabled {
			p.Status = "running"
		}
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

// ControlSchedule preserves the active round while pausing. Stop also persists
// cancellation; resuming never discards an unaccounted saved result.
func (s *Service) ControlSchedule(ctx context.Context, owner, project, id, action string) (Schedule, error) {
	var out Schedule
	var cancelledID string
	err := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		p, e := tx.GetSchedule(ctx, owner, id)
		if e != nil {
			return e
		}
		if p.ProjectID != project {
			return &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
		}
		restoreLegacyScheduleRound(&p)
		switch action {
		case "pause":
			p.Enabled = false
			p.Status = "paused"
			p.StoppedReason = "用户暂停"
		case "stop":
			p.Enabled = false
			p.Status = "stopped"
			p.StoppedReason = "用户停止"
			if r := p.active(); r != nil {
				child := s.automationChild(tx)
				if _, e := child.inspectRound(ctx, owner, r); e != nil {
					return e
				}
				run, e := tx.GetRun(ctx, owner, r.RunID)
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
		case "resume":
			p.defaults()
			if p.IntervalMinutes < 1 || p.IntervalMinutes > 1440 || p.DailyChapterCap < 1 || p.DailyChapterCap > 100 || p.MaxConsecutiveFailures < 1 || p.MaxConsecutiveFailures > 10 || p.ConfigurationError != "" {
				return &DomainError{Code: ErrInputRequired, Message: "配置需调整"}
			}
			if r := p.active(); r != nil {
				if _, e := s.automationChild(tx).inspectRound(ctx, owner, r); e != nil {
					return e
				}
				if r.ErrorCode == ErrResultUnknown {
					return unknownResult()
				}
			}
			p.Enabled = true
			p.Status = "waiting"
			p.StoppedReason = ""
			p.ConsecutiveFailures = 0
		default:
			return &DomainError{Code: ErrInputRequired, Message: "无效操作"}
		}
		out, e = tx.UpdateSchedule(ctx, owner, p)
		return e
	})
	if err == nil {
		s.signalAutomationCancel(ctx, owner, cancelledID)
	}
	return out, err
}

// StartScheduler scans due schedules on a fixed interval until ctx is done.
func StartScheduler(ctx context.Context, service *Service, start func(context.Context, string, string), interval time.Duration) {
	if service == nil || start == nil {
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				dispatched, err := service.DispatchDueSchedules(context.Background(), service.automationNow())
				if err != nil {
					fmt.Printf("novel creation scheduler: %v\n", err)
				}
				for _, item := range dispatched {
					start(context.Background(), item.Schedule.OwnerUsername, item.Run.ID)
				}
			}
		}
	}()
}
