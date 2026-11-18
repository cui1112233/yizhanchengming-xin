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
	submissions      map[string]submissionReceipt
	contentRecords   map[string]ContentRecord
	automationRounds map[string]string
	executions       map[string]Execution
	slots            map[string]ProjectSlot
	mu               sync.RWMutex
	nextID           int64
	nextEventID      int64
	projects         map[string]Project
	sessions         map[string]Session
	messages         map[string][]Message
	runs             map[string]Run
	runByKey         map[string]string
	runEvents        map[string][]RunEvent
	chapterBundles   map[string]ChapterBundle
	chapterVersions  map[string]ChapterVersion
	settings         map[string]ProjectSettings
	foundations      map[string]Foundation
	materials        map[string]Material
	notes            map[string]ProjectNote
	styles           map[string]StyleProfile
	schedules        map[string]Schedule
	batches          map[string]Batch
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		submissions:      make(map[string]submissionReceipt),
		contentRecords:   make(map[string]ContentRecord),
		automationRounds: make(map[string]string),
		executions:       make(map[string]Execution),
		slots:            make(map[string]ProjectSlot),
		projects:         make(map[string]Project),
		sessions:         make(map[string]Session),
		messages:         make(map[string][]Message),
		runs:             make(map[string]Run),
		runByKey:         make(map[string]string),
		runEvents:        make(map[string][]RunEvent),
		chapterBundles:   make(map[string]ChapterBundle),
		chapterVersions:  make(map[string]ChapterVersion),
		settings:         make(map[string]ProjectSettings),
		foundations:      make(map[string]Foundation),
		materials:        make(map[string]Material),
		notes:            make(map[string]ProjectNote),
		styles:           make(map[string]StyleProfile),
		schedules:        make(map[string]Schedule),
		batches:          make(map[string]Batch),
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
	// Keep version rows immutable, matching the MySQL unique chapter/version key.
	if version := bundle.ChapterVersion; version.ID != "" {
		if version.Version == 0 {
			version.Version = bundle.Chapter.Version
			if version.Version == 0 {
				version.Version = 1
			}
		}
		key := memoryKey(owner, bundle.ProjectID, bundle.Chapter.ID, fmt.Sprint(version.Version))
		if _, exists := s.chapterVersions[key]; exists {
			return continuityError(fmt.Errorf("chapter version already exists"))
		}
		version.Snapshot = cloneRawJSON(version.Snapshot)
		s.chapterVersions[key] = version
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
	if bundle.RunID != "" && bundle.FinalRunStatus != "" {
		key := memoryKey(owner, bundle.RunID)
		run := s.runs[key]
		run.Status = bundle.FinalRunStatus
		run.RuntimeSnapshot = cloneRawJSON(bundle.RuntimeSnapshot)
		run.UpdatedAt = now
		if isTerminalRunStatus(run.Status) {
			run.FinishedAt = &now
		}
		s.runs[key] = cloneRun(run)
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
		return ProjectSettings{ProjectID: projectID, Rules: []string{}, Agent: NormalizeAgentSettings(AgentSettings{}, project.TargetChapterWords)}, nil
	}
	settings.Rules = append([]string(nil), settings.Rules...)
	settings.Agent = NormalizeAgentSettings(settings.Agent, project.TargetChapterWords)
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
		Rules: append([]string(nil), input.Rules...),
		Agent: NormalizeAgentSettings(input.Agent, project.TargetChapterWords), UpdatedAt: time.Now().UTC(),
	}
	s.settings[memoryKey(owner, projectID)] = settings
	settings.Agent = NormalizeAgentSettings(settings.Agent, project.TargetChapterWords)
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

func (s *MemoryStore) ListMaterials(_ context.Context, owner, projectID string) ([]Material, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	materials := make([]Material, 0)
	for key, material := range s.materials {
		if strings.HasPrefix(key, memoryKey(owner)+"\x00") && material.ProjectID == projectID {
			materials = append(materials, material)
		}
	}
	sort.Slice(materials, func(i, j int) bool { return materials[i].CreatedAt.After(materials[j].CreatedAt) })
	return materials, nil
}

func (s *MemoryStore) CreateMaterial(_ context.Context, owner, projectID string, input CreateMaterialInput) (Material, error) {
	input, err := NormalizeMaterialInput(input)
	if err != nil {
		return Material{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return Material{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	now := time.Now().UTC()
	material := Material{ID: s.newID("material"), ProjectID: projectID, Title: input.Title, Content: input.Content, CreatedAt: now, UpdatedAt: now}
	s.materials[memoryKey(owner, material.ID)] = material
	return material, nil
}

func (s *MemoryStore) DeleteMaterial(_ context.Context, owner, projectID, materialID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, materialID)
	material, ok := s.materials[key]
	if !ok || material.ProjectID != projectID {
		return &DomainError{Code: ErrProjectNotFound, Message: "material not found"}
	}
	delete(s.materials, key)
	return nil
}

func (s *MemoryStore) ListNotes(_ context.Context, owner, projectID string) ([]ProjectNote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	notes := make([]ProjectNote, 0)
	for key, note := range s.notes {
		if strings.HasPrefix(key, memoryKey(owner)+"\x00") && note.ProjectID == projectID {
			notes = append(notes, note)
		}
	}
	sort.Slice(notes, func(i, j int) bool { return notes[i].UpdatedAt.After(notes[j].UpdatedAt) })
	return notes, nil
}

func (s *MemoryStore) CreateNote(_ context.Context, owner, projectID string, input ProjectNoteInput) (ProjectNote, error) {
	input, err := NormalizeNoteInput(input)
	if err != nil {
		return ProjectNote{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return ProjectNote{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	now := time.Now().UTC()
	note := ProjectNote{ID: s.newID("note"), ProjectID: projectID, Title: input.Title, Content: input.Content, CreatedAt: now, UpdatedAt: now}
	s.notes[memoryKey(owner, note.ID)] = note
	return note, nil
}

func (s *MemoryStore) UpdateNote(_ context.Context, owner, projectID, noteID string, input ProjectNoteInput) (ProjectNote, error) {
	input, err := NormalizeNoteInput(input)
	if err != nil {
		return ProjectNote{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, noteID)
	note, ok := s.notes[key]
	if !ok || note.ProjectID != projectID {
		return ProjectNote{}, &DomainError{Code: ErrProjectNotFound, Message: "note not found"}
	}
	note.Title, note.Content, note.UpdatedAt = input.Title, input.Content, time.Now().UTC()
	s.notes[key] = note
	return note, nil
}

func (s *MemoryStore) DeleteNote(_ context.Context, owner, projectID, noteID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, noteID)
	note, ok := s.notes[key]
	if !ok || note.ProjectID != projectID {
		return &DomainError{Code: ErrProjectNotFound, Message: "note not found"}
	}
	delete(s.notes, key)
	return nil
}

func (s *MemoryStore) ListRuns(_ context.Context, owner, projectID string) ([]Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	runs := make([]Run, 0)
	for key, run := range s.runs {
		if strings.HasPrefix(key, memoryKey(owner)+"\x00") && run.ProjectID == projectID {
			runs = append(runs, cloneRun(run))
		}
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].CreatedAt.After(runs[j].CreatedAt) })
	return runs, nil
}

func (s *MemoryStore) ListSchedules(_ context.Context, owner, projectID string) ([]Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	schedules := make([]Schedule, 0)
	for _, schedule := range s.schedules {
		if schedule.OwnerUsername == owner && schedule.ProjectID == projectID {
			schedules = append(schedules, cloneSchedule(schedule))
		}
	}
	sort.Slice(schedules, func(i, j int) bool { return schedules[i].CreatedAt.After(schedules[j].CreatedAt) })
	return schedules, nil
}

func (s *MemoryStore) GetSchedule(_ context.Context, owner, scheduleID string) (Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	schedule, ok := s.schedules[memoryKey(owner, scheduleID)]
	if !ok {
		return Schedule{}, &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	return cloneSchedule(schedule), nil
}

func (s *MemoryStore) CreateSchedule(_ context.Context, owner string, schedule Schedule) (Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[memoryKey(owner, schedule.ProjectID)]; !ok {
		return Schedule{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	session, ok := s.sessions[memoryKey(owner, schedule.SessionID)]
	if !ok || session.ProjectID != schedule.ProjectID {
		return Schedule{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	if schedule.ID == "" {
		schedule.ID = s.newID("schedule")
	}
	now := time.Now().UTC()
	schedule.OwnerUsername = owner
	if schedule.CreatedAt.IsZero() {
		schedule.CreatedAt = now
	}
	schedule.UpdatedAt = now
	s.schedules[memoryKey(owner, schedule.ID)] = cloneSchedule(schedule)
	return cloneSchedule(schedule), nil
}

func (s *MemoryStore) UpdateSchedule(_ context.Context, owner string, schedule Schedule) (Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, schedule.ID)
	existing, ok := s.schedules[key]
	if !ok || existing.ProjectID != schedule.ProjectID {
		return Schedule{}, &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	schedule.OwnerUsername = owner
	schedule.CreatedAt = existing.CreatedAt
	schedule.UpdatedAt = time.Now().UTC()
	s.schedules[key] = cloneSchedule(schedule)
	return cloneSchedule(schedule), nil
}

func (s *MemoryStore) DeleteSchedule(_ context.Context, owner, projectID, scheduleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, scheduleID)
	schedule, ok := s.schedules[key]
	if !ok || schedule.ProjectID != projectID {
		return &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	delete(s.schedules, key)
	return nil
}

func (s *MemoryStore) ListDueSchedules(ctx context.Context, now time.Time, limit int) ([]Schedule, error) {
	return s.ListDueSchedulePage(ctx, now, ScheduleCursor{}, limit)
}
func (s *MemoryStore) ListDueSchedulePage(_ context.Context, now time.Time, after ScheduleCursor, limit int) ([]Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 20
	}
	due := make([]Schedule, 0)
	for _, schedule := range s.schedules {
		if ((schedule.Enabled && !schedule.NextRunAt.After(now)) || schedule.ActiveRound != "") && scheduleAfter(schedule, after) {
			due = append(due, cloneSchedule(schedule))
		}
	}
	sort.Slice(due, func(i, j int) bool {
		return scheduleAfter(due[j], ScheduleCursor{NextRunAt: due[i].NextRunAt, OwnerUsername: due[i].OwnerUsername, ID: due[i].ID})
	})
	if len(due) > limit {
		due = due[:limit]
	}
	return due, nil
}

func (s *MemoryStore) ListBatches(_ context.Context, owner, projectID string) ([]Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	project, ok := s.projects[memoryKey(owner, projectID)]
	if !ok || project.DeletedAt != nil {
		return nil, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	batches := make([]Batch, 0)
	for _, batch := range s.batches {
		if batch.OwnerUsername == owner && batch.ProjectID == projectID {
			batches = append(batches, cloneBatch(batch))
		}
	}
	sort.Slice(batches, func(i, j int) bool { return batches[i].CreatedAt.After(batches[j].CreatedAt) })
	return batches, nil
}

func (s *MemoryStore) GetBatch(_ context.Context, owner, batchID string) (Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	batch, ok := s.batches[memoryKey(owner, batchID)]
	if !ok {
		return Batch{}, &DomainError{Code: ErrProjectNotFound, Message: "batch not found"}
	}
	return cloneBatch(batch), nil
}

func (s *MemoryStore) CreateBatch(_ context.Context, owner string, batch Batch) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[memoryKey(owner, batch.ProjectID)]; !ok {
		return Batch{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	session, ok := s.sessions[memoryKey(owner, batch.SessionID)]
	if !ok || session.ProjectID != batch.ProjectID {
		return Batch{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	if batch.ID == "" {
		batch.ID = s.newID("batch")
	}
	now := time.Now().UTC()
	batch.OwnerUsername = owner
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt = now
	}
	batch.UpdatedAt = now
	s.batches[memoryKey(owner, batch.ID)] = cloneBatch(batch)
	return cloneBatch(batch), nil
}

func (s *MemoryStore) UpdateBatch(_ context.Context, owner string, batch Batch) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := memoryKey(owner, batch.ID)
	existing, ok := s.batches[key]
	if !ok || existing.ProjectID != batch.ProjectID {
		return Batch{}, &DomainError{Code: ErrProjectNotFound, Message: "batch not found"}
	}
	batch.OwnerUsername = owner
	batch.CreatedAt = existing.CreatedAt
	batch.UpdatedAt = time.Now().UTC()
	s.batches[key] = cloneBatch(batch)
	return cloneBatch(batch), nil
}

func (s *MemoryStore) ListActiveBatches(_ context.Context, limit int) ([]Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 20
	}
	active := make([]Batch, 0)
	for _, batch := range s.batches {
		project, ok := s.projects[memoryKey(batch.OwnerUsername, batch.ProjectID)]
		if !ok || project.DeletedAt != nil || projectReadOnly(project.Status) {
			continue
		}
		if batch.Status == BatchRunning || batch.Status == BatchPaused || (batch.Status == BatchStopped && batch.ActiveRound != "") {
			active = append(active, cloneBatch(batch))
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].UpdatedAt.Before(active[j].UpdatedAt) })
	if len(active) > limit {
		active = active[:limit]
	}
	return active, nil
}
