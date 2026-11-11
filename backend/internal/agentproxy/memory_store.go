package agentproxy

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStore 是 Store 的内存实现，用于单元测试和没有 MySQL 的本地联调。
type MemoryStore struct {
	mu             sync.Mutex
	workspaces     map[string]Workspace
	workspaceOwner map[string]string
	sessions       map[string]Session
	sessionOwner   map[string]string
	messages       map[string][]Message
	nextID         int64
	skills         map[string][]Skill
	jobs           map[string][]ScheduledJob
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		workspaces:     map[string]Workspace{},
		workspaceOwner: map[string]string{},
		sessions:       map[string]Session{},
		sessionOwner:   map[string]string{},
		messages:       map[string][]Message{},
		skills:         map[string][]Skill{},
		jobs:           map[string][]ScheduledJob{},
	}
}

func (m *MemoryStore) ListWorkspaces(_ context.Context, owner string) ([]Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Workspace, 0)
	for id, item := range m.workspaces {
		if m.workspaceOwner[id] != owner {
			continue
		}
		item.SessionCount = m.sessionCount(owner, id)
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (m *MemoryStore) Workspace(_ context.Context, owner, id string) (Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.workspaces[strings.TrimSpace(id)]
	if !ok || m.workspaceOwner[item.ID] != owner {
		return Workspace{}, ErrWorkspaceNotFound
	}
	item.SessionCount = m.sessionCount(owner, item.ID)
	return item, nil
}

func (m *MemoryStore) CreateWorkspace(_ context.Context, owner string, input CreateWorkspaceInput, now time.Time) (Workspace, error) {
	mode, err := normalizeCreatePermissionMode(input.PermissionMode)
	if err != nil {
		return Workspace{}, err
	}
	id := strings.TrimSpace(input.ID)
	name := strings.TrimSpace(input.Name)
	if strings.TrimSpace(owner) == "" || id == "" || name == "" {
		return Workspace{}, ErrInvalidInput
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.workspaces[id]; exists {
		return Workspace{}, ErrWorkspaceExists
	}
	item := Workspace{ID: id, Name: name, Description: strings.TrimSpace(input.Description), PermissionMode: mode, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	m.workspaces[id] = item
	m.workspaceOwner[id] = owner
	return item, nil
}

func (m *MemoryStore) UpdateWorkspace(_ context.Context, owner, id string, input UpdateWorkspaceInput, now time.Time) (Workspace, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.workspaces[strings.TrimSpace(id)]
	if !ok || m.workspaceOwner[item.ID] != owner {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return Workspace{}, ErrInvalidInput
		}
		item.Name = name
	}
	if input.Description != nil {
		item.Description = strings.TrimSpace(*input.Description)
	}
	if input.PermissionMode != nil {
		mode, err := normalizeCreatePermissionMode(*input.PermissionMode)
		if err != nil {
			return Workspace{}, err
		}
		item.PermissionMode = mode
	}
	item.UpdatedAt = now.UTC()
	m.workspaces[item.ID] = item
	item.SessionCount = m.sessionCount(owner, item.ID)
	return item, nil
}

func (m *MemoryStore) DeleteWorkspace(_ context.Context, owner, id string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.workspaces[strings.TrimSpace(id)]
	if !ok || m.workspaceOwner[item.ID] != owner {
		return ErrWorkspaceNotFound
	}
	delete(m.workspaces, item.ID)
	delete(m.workspaceOwner, item.ID)
	for sessionID, session := range m.sessions {
		if session.WorkspaceID != item.ID || m.sessionOwner[sessionID] != owner {
			continue
		}
		session.UpdatedAt = now.UTC()
		delete(m.sessions, sessionID)
		delete(m.sessionOwner, sessionID)
	}
	return nil
}

func (m *MemoryStore) ListSessions(_ context.Context, owner, workspaceID string) ([]Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Session, 0)
	for id, item := range m.sessions {
		if m.sessionOwner[id] != owner || item.WorkspaceID != strings.TrimSpace(workspaceID) {
			continue
		}
		item.MessageCount = len(m.messages[item.ID])
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (m *MemoryStore) Session(_ context.Context, owner, id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.sessions[strings.TrimSpace(id)]
	if !ok || m.sessionOwner[item.ID] != owner {
		return Session{}, ErrSessionNotFound
	}
	item.MessageCount = len(m.messages[item.ID])
	return item, nil
}

func (m *MemoryStore) CreateSession(_ context.Context, owner, workspaceID string, input CreateSessionInput, now time.Time) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	workspace, ok := m.workspaces[strings.TrimSpace(workspaceID)]
	if !ok || m.workspaceOwner[workspace.ID] != owner {
		return Session{}, ErrWorkspaceNotFound
	}
	id := strings.TrimSpace(input.ID)
	title := strings.TrimSpace(input.Title)
	if id == "" || title == "" {
		return Session{}, ErrInvalidInput
	}
	if _, exists := m.sessions[id]; exists {
		return Session{}, ErrSessionExists
	}
	mode := workspace.PermissionMode
	if strings.TrimSpace(input.PermissionMode) != "" {
		var err error
		if mode, err = normalizeCreatePermissionMode(input.PermissionMode); err != nil {
			return Session{}, err
		}
	}
	item := Session{ID: id, WorkspaceID: workspace.ID, Title: title, Status: SessionStatusActive, PermissionMode: mode, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	m.sessions[id] = item
	m.sessionOwner[id] = owner
	return item, nil
}

func (m *MemoryStore) UpdateSession(_ context.Context, owner, id string, input UpdateSessionInput, now time.Time) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.sessions[strings.TrimSpace(id)]
	if !ok || m.sessionOwner[item.ID] != owner {
		return Session{}, ErrSessionNotFound
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return Session{}, ErrInvalidInput
		}
		item.Title = title
	}
	if input.Status != nil {
		status := NormalizeSessionStatus(*input.Status)
		if !ValidSessionStatus(status) {
			return Session{}, ErrInvalidInput
		}
		item.Status = status
	}
	if input.PermissionMode != nil {
		mode, err := normalizeCreatePermissionMode(*input.PermissionMode)
		if err != nil {
			return Session{}, err
		}
		item.PermissionMode = mode
	}
	item.UpdatedAt = now.UTC()
	item.MessageCount = len(m.messages[item.ID])
	m.sessions[item.ID] = item
	return item, nil
}

func (m *MemoryStore) DeleteSession(_ context.Context, owner, id string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.sessions[strings.TrimSpace(id)]
	if !ok || m.sessionOwner[item.ID] != owner {
		return ErrSessionNotFound
	}
	delete(m.sessions, item.ID)
	delete(m.sessionOwner, item.ID)
	return nil
}

func (m *MemoryStore) ListMessages(_ context.Context, owner, sessionID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[strings.TrimSpace(sessionID)]
	if !ok || m.sessionOwner[session.ID] != owner {
		return nil, ErrSessionNotFound
	}
	items := m.messages[session.ID]
	if len(items) > limit {
		items = items[len(items)-limit:]
	}
	return append([]Message(nil), items...), nil
}

func (m *MemoryStore) AppendMessage(_ context.Context, owner, sessionID string, input AppendMessageInput, now time.Time) (Message, error) {
	role := strings.ToLower(strings.TrimSpace(input.Role))
	content := strings.TrimSpace(input.Content)
	if !ValidMessageRole(role) || content == "" {
		return Message{}, ErrInvalidInput
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[strings.TrimSpace(sessionID)]
	if !ok || m.sessionOwner[session.ID] != owner {
		return Message{}, ErrSessionNotFound
	}
	m.nextID++
	item := Message{ID: m.nextID, SessionID: session.ID, Role: role, Content: content, Metadata: input.Metadata, CreatedAt: now.UTC()}
	m.messages[session.ID] = append(m.messages[session.ID], item)
	session.LastMessageAt = &item.CreatedAt
	session.UpdatedAt = item.CreatedAt
	m.sessions[session.ID] = session
	return item, nil
}

func (m *MemoryStore) ListSkills(_ context.Context, owner, workspaceID string) ([]Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]Skill(nil), m.skills[strings.TrimSpace(workspaceID)+"|"+owner]...), nil
}

func (m *MemoryStore) ListScheduledJobs(_ context.Context, owner, workspaceID string) ([]ScheduledJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ScheduledJob(nil), m.jobs[strings.TrimSpace(workspaceID)+"|"+owner]...), nil
}

func (m *MemoryStore) sessionCount(owner, workspaceID string) int {
	count := 0
	for id, item := range m.sessions {
		if m.sessionOwner[id] == owner && item.WorkspaceID == workspaceID {
			count++
		}
	}
	return count
}
