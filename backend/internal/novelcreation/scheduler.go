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
	return s.store.ListSchedules(ctx, strings.TrimSpace(owner), strings.TrimSpace(projectID))
}

func (s *Service) CreateSchedule(ctx context.Context, owner, projectID string, input ScheduleInput) (Schedule, error) {
	if s == nil || s.store == nil {
		return Schedule{}, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	owner, projectID = strings.TrimSpace(owner), strings.TrimSpace(projectID)
	input, err := NormalizeScheduleInput(input)
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
	now := time.Now().UTC()
	schedule := Schedule{
		ModeInputs: selected.ModeInputs, KnowledgeIDs: selected.KnowledgeIDs, NoteIDs: selected.NoteIDs, StyleProfileID: selected.StyleProfileID,
		ID: novelID("schedule"), ProjectID: projectID, SessionID: input.SessionID,
		IntervalMinutes: input.IntervalMinutes, Enabled: input.Enabled, Prompt: input.Prompt,
		ModelID: modelID, ThinkingLevel: input.ThinkingLevel,
		NextRunAt: now.Add(time.Duration(input.IntervalMinutes) * time.Minute),
	}
	return s.store.CreateSchedule(ctx, owner, schedule)
}

func (s *Service) UpdateSchedule(ctx context.Context, owner, projectID, scheduleID string, input ScheduleInput) (Schedule, error) {
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
	now := time.Now().UTC()
	schedule.ConfigurationError = ""
	schedule.ModeInputs = selected.ModeInputs
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

func (s *Service) DeleteSchedule(ctx context.Context, owner, projectID, scheduleID string) error {
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
	return s.dispatchSchedule(ctx, owner, schedule, time.Now().UTC())
}

// DispatchDueSchedules runs every due schedule once and advances its next run
// time. Individual failures are collected so one broken schedule cannot block
// the others.
func (s *Service) DispatchDueSchedules(ctx context.Context, now time.Time) ([]DispatchedSchedule, error) {
	if s == nil || s.store == nil || s.runtime == nil {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "小说创作运行服务尚未配置"}
	}
	due, err := s.store.ListDueSchedules(ctx, now, 20)
	if err != nil {
		return nil, err
	}
	dispatched := make([]DispatchedSchedule, 0, len(due))
	failures := make([]error, 0)
	for _, schedule := range due {
		run, dispatchErr := s.dispatchSchedule(ctx, schedule.OwnerUsername, schedule, now)
		if dispatchErr != nil {
			failures = append(failures, fmt.Errorf("schedule %s: %w", schedule.ID, dispatchErr))
			continue
		}
		dispatched = append(dispatched, DispatchedSchedule{Schedule: schedule, Run: run})
	}
	return dispatched, errors.Join(failures...)
}

func (s *Service) dispatchSchedule(ctx context.Context, owner string, schedule Schedule, now time.Time) (Run, error) {
	project, err := s.store.GetProject(ctx, owner, schedule.ProjectID)
	if err != nil {
		return Run{}, err
	}
	modelID := strings.TrimSpace(schedule.ModelID)
	if modelID == "" {
		modelID = project.ModelID
	}
	run, sendErr := s.SendMessage(ctx, owner, SendMessageInput{
		SessionID: schedule.SessionID, ProjectID: schedule.ProjectID,
		Content: schedule.Prompt, ModelID: modelID, ThinkingLevel: schedule.ThinkingLevel,
		ModeInputs: schedule.ModeInputs, KnowledgeIDs: schedule.KnowledgeIDs, NoteIDs: schedule.NoteIDs, StyleProfileID: schedule.StyleProfileID,
		IdempotencyKey: fmt.Sprintf("schedule:%s:%d", schedule.ID, now.UTC().Unix()),
	})
	if IsCode(sendErr, ErrInputRequired) || IsCode(sendErr, ErrSessionNotFound) || IsCode(sendErr, ErrProjectNotFound) {
		schedule.Enabled = false
		schedule.ConfigurationError = "配置需调整：" + sendErr.Error()
	}
	lastRunAt := now.UTC()
	schedule.LastRunAt = &lastRunAt
	if run.ID != "" {
		schedule.LastRunID = run.ID
	}
	schedule.NextRunAt = now.UTC().Add(time.Duration(schedule.IntervalMinutes) * time.Minute)
	if _, updateErr := s.store.UpdateSchedule(ctx, owner, schedule); updateErr != nil {
		return run, errors.Join(sendErr, updateErr)
	}
	return run, sendErr
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
				dispatched, err := service.DispatchDueSchedules(context.Background(), time.Now().UTC())
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
