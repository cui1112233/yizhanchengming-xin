package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	mu             sync.RWMutex
	nextID         int64
	nextEventID    int64
	projects       map[string]Project
	sessions       map[string]Session
	messages       map[string][]Message
	runs           map[string]Run
	runByKey       map[string]string
	runEvents      map[string][]RunEvent
	chapterBundles map[string]ChapterBundle
	settings       map[string]ProjectSettings
	foundations    map[string]Foundation
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects:       make(map[string]Project),
		sessions:       make(map[string]Session),
		messages:       make(map[string][]Message),
		runs:           make(map[string]Run),
		runByKey:       make(map[string]string),
		runEvents:      make(map[string][]RunEvent),
		chapterBundles: make(map[string]ChapterBundle),
		settings:       make(map[string]ProjectSettings),
		foundations:    make(map[string]Foundation),
	}
}

func memoryKey(parts ...string) string { return strings.Join(parts, "\x00") }

func cloneRawJSON(value json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), value...)
}

func cloneProject(project Project) Project {
	if project.DeletedAt != nil {
		deletedAt := *project.DeletedAt
		project.DeletedAt = &deletedAt
	}
	return project
}

func cloneMessage(message Message) Message {
	message.References = cloneRawJSON(message.References)
	message.TokenUsage = cloneRawJSON(message.TokenUsage)
	return message
}

func cloneRun(run Run) Run {
	run.PromptSnapshot = cloneRawJSON(run.PromptSnapshot)
	run.RuntimeSnapshot = cloneRawJSON(run.RuntimeSnapshot)
	if run.FinishedAt != nil {
		finishedAt := *run.FinishedAt
		run.FinishedAt = &finishedAt
	}
	return run
}

func cloneRunEvent(event RunEvent) RunEvent {
	event.Payload = cloneRawJSON(event.Payload)
	return event
}

func cloneChapterBundle(bundle ChapterBundle) ChapterBundle {
	bundle.ChapterVersion.Snapshot = cloneRawJSON(bundle.ChapterVersion.Snapshot)
	bundle.Foundation.Snapshot = cloneRawJSON(bundle.Foundation.Snapshot)
	bundle.RuntimeSnapshot = cloneRawJSON(bundle.RuntimeSnapshot)
	bundle.Memories = append([]Memory(nil), bundle.Memories...)
	for index := range bundle.Memories {
		bundle.Memories[index].Snapshot = cloneRawJSON(bundle.Memories[index].Snapshot)
	}
	bundle.Audits = append([]Audit(nil), bundle.Audits...)
	for index := range bundle.Audits {
		bundle.Audits[index].Snapshot = cloneRawJSON(bundle.Audits[index].Snapshot)
	}
	bundle.Revisions = append([]Revision(nil), bundle.Revisions...)
	for index := range bundle.Revisions {
		bundle.Revisions[index].Snapshot = cloneRawJSON(bundle.Revisions[index].Snapshot)
	}
	return bundle
}

func validateRawJSON(field string, value json.RawMessage) error {
	if len(value) == 0 {
		return nil
	}
	if !json.Valid(value) {
		return &DomainError{Code: ErrPersistenceFailed, Message: field + " must be valid JSON"}
	}
	return nil
}

func validateChapterBundleJSON(bundle ChapterBundle) error {
	for _, value := range []struct {
		field string
		value json.RawMessage
	}{
		{field: "chapter version snapshot", value: bundle.ChapterVersion.Snapshot},
		{field: "foundation snapshot", value: bundle.Foundation.Snapshot},
		{field: "runtime snapshot", value: bundle.RuntimeSnapshot},
	} {
		if err := validateRawJSON(value.field, value.value); err != nil {
			return err
		}
	}
	for _, memory := range bundle.Memories {
		if err := validateRawJSON("memory snapshot", memory.Snapshot); err != nil {
			return err
		}
	}
	for _, audit := range bundle.Audits {
		if err := validateRawJSON("audit snapshot", audit.Snapshot); err != nil {
			return err
		}
	}
	for _, revision := range bundle.Revisions {
		if err := validateRawJSON("revision snapshot", revision.Snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStore) newID(prefix string) string {
	s.nextID++
	return fmt.Sprintf("%s-%d", prefix, s.nextID)
}

func (s *MemoryStore) CreateProject(_ context.Context, owner string, input CreateProjectInput) (Project, error) {
	input, err := NormalizeCreateProjectInput(input)
	if err != nil {
		return Project{}, err
	}
	if strings.TrimSpace(owner) == "" || input.Title == "" || input.Genre == "" || input.ModelID == "" {
		return Project{}, &DomainError{Code: ErrInputRequired, Message: "owner, title, genre and modelId are required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	project := Project{ID: s.newID("project"), Title: input.Title, Genre: input.Genre, World: input.World, Protagonist: input.Protagonist, CoreConflict: input.CoreConflict, ModelID: input.ModelID, TargetChapterWords: input.TargetChapterWords, Status: "draft", CreatedAt: now, UpdatedAt: now}
	s.projects[memoryKey(owner, project.ID)] = cloneProject(project)
	return cloneProject(project), nil
}

func (s *MemoryStore) GetProject(_ context.Context, owner, projectID string) (Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return Project{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	return cloneProject(project), nil
}

func (s *MemoryStore) ListProjects(_ context.Context, owner string) ([]Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	projects := make([]Project, 0)
	prefix := owner + "\x00"
	for key, project := range s.projects {
		if strings.HasPrefix(key, prefix) && project.DeletedAt == nil {
			projects = append(projects, cloneProject(project))
		}
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].UpdatedAt.After(projects[j].UpdatedAt) })
	return projects, nil
}

func (s *MemoryStore) CreateSession(_ context.Context, owner string, input CreateSessionInput) (Session, error) {
	input.ProjectID, input.Mode, input.Title = strings.TrimSpace(input.ProjectID), strings.TrimSpace(input.Mode), strings.TrimSpace(input.Title)
	if input.ProjectID == "" || input.Mode == "" {
		return Session{}, &DomainError{Code: ErrInputRequired, Message: "projectId and mode are required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[memoryKey(owner, input.ProjectID)]
	if !ok || project.DeletedAt != nil {
		return Session{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	now := time.Now().UTC()
	session := Session{ID: s.newID("session"), ProjectID: input.ProjectID, Mode: input.Mode, Title: input.Title, ContextBudget: input.ContextBudget, CreatedAt: now, UpdatedAt: now}
	s.sessions[memoryKey(owner, session.ID)] = session
	return session, nil
}

func (s *MemoryStore) GetSession(_ context.Context, owner, sessionID string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[memoryKey(owner, sessionID)]
	if !ok {
		return Session{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	return session, nil
}

func (s *MemoryStore) ListSessions(_ context.Context, owner, projectID string) ([]Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make([]Session, 0)
	prefix := owner + "\x00"
	for key, session := range s.sessions {
		if strings.HasPrefix(key, prefix) && (projectID == "" || session.ProjectID == projectID) {
			sessions = append(sessions, session)
		}
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt) })
	return sessions, nil
}

func (s *MemoryStore) AppendMessage(_ context.Context, owner string, message Message) (Message, error) {
	if err := validateRawJSON("message references", message.References); err != nil {
		return Message{}, err
	}
	if err := validateRawJSON("message token usage", message.TokenUsage); err != nil {
		return Message{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[memoryKey(owner, message.SessionID)]
	if !ok || session.ProjectID != message.ProjectID {
		return Message{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	message.ID, message.Role, message.Content = strings.TrimSpace(message.ID), strings.TrimSpace(message.Role), strings.TrimSpace(message.Content)
	if message.Role == "" || message.Content == "" {
		return Message{}, &DomainError{Code: ErrInputRequired, Message: "message role and content are required"}
	}
	key := memoryKey(owner, message.SessionID)
	message.Sequence = int64(len(s.messages[key]) + 1)
	if message.ID == "" {
		message.ID = s.newID("message")
	}
	now := time.Now().UTC()
	message.CreatedAt, message.UpdatedAt = now, now
	s.messages[key] = append(s.messages[key], cloneMessage(message))
	session.MessageCursor = message.Sequence
	session.UpdatedAt = now
	s.sessions[memoryKey(owner, session.ID)] = session
	return cloneMessage(message), nil
}

func (s *MemoryStore) ListMessages(_ context.Context, owner, sessionID string, after int64) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.sessions[memoryKey(owner, sessionID)]; !ok {
		return nil, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	messages := make([]Message, 0)
	for _, message := range s.messages[memoryKey(owner, sessionID)] {
		if message.Sequence > after {
			messages = append(messages, cloneMessage(message))
		}
	}
	return messages, nil
}

func (s *MemoryStore) FindRunByIdempotencyKey(_ context.Context, owner, idempotencyKey string) (Run, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.runByKey[memoryKey(owner, strings.TrimSpace(idempotencyKey))]
	if !ok {
		return Run{}, false, nil
	}
	return cloneRun(s.runs[memoryKey(owner, id)]), true, nil
}

func (s *MemoryStore) GetRun(_ context.Context, owner, runID string) (Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.runs[memoryKey(owner, strings.TrimSpace(runID))]
	if !ok {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	return cloneRun(run), nil
}

func (s *MemoryStore) CreateRun(_ context.Context, owner string, run Run) (Run, error) {
	if err := validateRawJSON("run prompt snapshot", run.PromptSnapshot); err != nil {
		return Run{}, err
	}
	if err := validateRawJSON("run runtime snapshot", run.RuntimeSnapshot); err != nil {
		return Run{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	run.ID, run.ProjectID, run.SessionID, run.IdempotencyKey = strings.TrimSpace(run.ID), strings.TrimSpace(run.ProjectID), strings.TrimSpace(run.SessionID), strings.TrimSpace(run.IdempotencyKey)
	if run.ProjectID == "" || run.SessionID == "" || run.IdempotencyKey == "" || run.Status == "" {
		return Run{}, &DomainError{Code: ErrInputRequired, Message: "run projectId, sessionId, idempotencyKey and status are required"}
	}
	project, ok := s.projects[memoryKey(owner, run.ProjectID)]
	if !ok || project.DeletedAt != nil {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	session, ok := s.sessions[memoryKey(owner, run.SessionID)]
	if !ok || session.ProjectID != run.ProjectID {
		return Run{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	key := memoryKey(owner, run.IdempotencyKey)
	if existingID, exists := s.runByKey[key]; exists {
		return cloneRun(s.runs[memoryKey(owner, existingID)]), &DomainError{Code: ErrRunInProgress, Message: "idempotency key already has a run"}
	}
	if run.ID == "" {
		run.ID = s.newID("run")
	}
	now := time.Now().UTC()
	run.CreatedAt, run.UpdatedAt = now, now
	s.runs[memoryKey(owner, run.ID)] = cloneRun(run)
	s.runByKey[key] = run.ID
	session.CurrentRunID, session.UpdatedAt = run.ID, now
	s.sessions[memoryKey(owner, session.ID)] = session
	return cloneRun(run), nil
}

func (s *MemoryStore) UpdateRun(_ context.Context, owner string, update RunUpdate) (Run, error) {
	if err := validateRawJSON("run runtime snapshot", update.RuntimeSnapshot); err != nil {
		return Run{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, strings.TrimSpace(update.ID))
	run, ok := s.runs[key]
	if !ok {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	if update.Status != "" {
		run.Status = update.Status
	}
	if update.RuntimeSnapshot != nil {
		run.RuntimeSnapshot = cloneRawJSON(update.RuntimeSnapshot)
	}
	if update.ErrorCode != "" {
		run.ErrorCode = update.ErrorCode
	}
	if update.ErrorMessage != "" {
		run.ErrorMessage = update.ErrorMessage
	}
	if update.FinishedAt != nil {
		finishedAt := update.FinishedAt.UTC()
		run.FinishedAt = &finishedAt
	}
	run.UpdatedAt = time.Now().UTC()
	s.runs[key] = cloneRun(run)
	return cloneRun(run), nil
}

func (s *MemoryStore) AppendRunEvent(_ context.Context, owner string, event RunEvent) (RunEvent, error) {
	if err := validateRawJSON("run event payload", event.Payload); err != nil {
		return RunEvent{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[memoryKey(owner, strings.TrimSpace(event.RunID))]
	if !ok || run.ProjectID != event.ProjectID || run.SessionID != event.SessionID {
		return RunEvent{}, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	s.nextEventID++
	event.EventID = s.nextEventID
	event.CreatedAt = time.Now().UTC()
	key := memoryKey(owner, event.RunID)
	s.runEvents[key] = append(s.runEvents[key], cloneRunEvent(event))
	return cloneRunEvent(event), nil
}

func (s *MemoryStore) ListRunEvents(_ context.Context, owner, runID string, after int64) ([]RunEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.runs[memoryKey(owner, runID)]; !ok {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	events := make([]RunEvent, 0)
	for _, event := range s.runEvents[memoryKey(owner, runID)] {
		if event.EventID > after {
			events = append(events, cloneRunEvent(event))
		}
	}
	return events, nil
}

func (s *MemoryStore) ListChapters(_ context.Context, owner, projectID string) ([]Chapter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	projectID = strings.TrimSpace(projectID)
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	chapters := make([]Chapter, 0)
	for key, bundle := range s.chapterBundles {
		if strings.HasPrefix(key, memoryKey(owner)+"\x00") && bundle.ProjectID == projectID {
			chapters = append(chapters, bundle.Chapter)
		}
	}
	sort.Slice(chapters, func(i, j int) bool { return chapters[i].Number < chapters[j].Number })
	return chapters, nil
}

func (s *MemoryStore) SaveChapterBundle(_ context.Context, owner string, bundle ChapterBundle) error {
	if err := validateChapterBundleJSON(bundle); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[memoryKey(owner, strings.TrimSpace(bundle.ProjectID))]
	if !ok || project.DeletedAt != nil {
		return &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	if bundle.RunID != "" {
		run, exists := s.runs[memoryKey(owner, bundle.RunID)]
		if !exists || run.ProjectID != bundle.ProjectID {
			return &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
		}
	}
	if strings.TrimSpace(bundle.Chapter.ID) == "" || bundle.Chapter.Number <= 0 {
		return &DomainError{Code: ErrInputRequired, Message: "chapter id and number are required"}
	}
	bundle.Chapter.ProjectID = bundle.ProjectID
	now := time.Now().UTC()
	if bundle.Chapter.CreatedAt.IsZero() {
		bundle.Chapter.CreatedAt = now
	}
	bundle.Chapter.UpdatedAt = now
	s.chapterBundles[memoryKey(owner, bundle.ProjectID, bundle.Chapter.ID)] = cloneChapterBundle(bundle)
	if bundle.Foundation.ID != "" {
		foundation := bundle.Foundation
		foundation.ProjectID = bundle.ProjectID
		if foundation.CreatedAt.IsZero() {
			foundation.CreatedAt = now
		}
		foundation.UpdatedAt = now
		foundation.Snapshot = cloneRawJSON(foundation.Snapshot)
		s.foundations[memoryKey(owner, bundle.ProjectID)] = foundation
	}
	return nil
}

func (s *MemoryStore) GetProjectSettings(_ context.Context, owner, projectID string) (ProjectSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return ProjectSettings{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	settings, exists := s.settings[memoryKey(owner, projectID)]
	if !exists {
		return ProjectSettings{ProjectID: projectID, Rules: []string{}}, nil
	}
	settings.Rules = append([]string(nil), settings.Rules...)
	return settings, nil
}

func (s *MemoryStore) UpdateProjectSettings(_ context.Context, owner, projectID string, input ProjectSettings) (ProjectSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return ProjectSettings{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	settings := ProjectSettings{
		ProjectID: projectID, StyleProfile: strings.TrimSpace(input.StyleProfile),
		Rules: append([]string(nil), input.Rules...), UpdatedAt: time.Now().UTC(),
	}
	s.settings[memoryKey(owner, projectID)] = settings
	return settings, nil
}

func (s *MemoryStore) GetFoundation(_ context.Context, owner, projectID string) (Foundation, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return Foundation{}, false, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	foundation, exists := s.foundations[memoryKey(owner, projectID)]
	if !exists {
		return Foundation{}, false, nil
	}
	foundation.Snapshot = cloneRawJSON(foundation.Snapshot)
	return foundation, true, nil
}

func (s *MemoryStore) SaveFoundation(_ context.Context, owner string, foundation Foundation) error {
	if err := validateRawJSON("foundation snapshot", foundation.Snapshot); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	foundation.ProjectID = strings.TrimSpace(foundation.ProjectID)
	if foundation.ProjectID == "" || strings.TrimSpace(foundation.ID) == "" {
		return &DomainError{Code: ErrInputRequired, Message: "foundation id and projectId are required"}
	}
	project, ok := s.projects[memoryKey(owner, foundation.ProjectID)]
	if !ok || project.DeletedAt != nil {
		return &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	now := time.Now().UTC()
	if foundation.CreatedAt.IsZero() {
		foundation.CreatedAt = now
	}
	foundation.UpdatedAt = now
	foundation.Snapshot = cloneRawJSON(foundation.Snapshot)
	s.foundations[memoryKey(owner, foundation.ProjectID)] = foundation
	return nil
}
