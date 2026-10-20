package agentworkspace

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	mu               sync.Mutex
	projects         map[string]Project
	imports          map[string]LegacyImport
	videoJobs        map[string]VideoJob
	videoIdempotency map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{projects: make(map[string]Project), imports: make(map[string]LegacyImport), videoJobs: make(map[string]VideoJob), videoIdempotency: make(map[string]string)}
}

func projectKey(owner, id string) string                   { return owner + "\x00" + id }
func importKey(owner, path string) string                  { return owner + "\x00" + path }
func videoJobKey(owner, id string) string                  { return owner + "\x00" + id }
func videoIdempotencyKey(owner, idempotency string) string { return owner + "\x00" + idempotency }

func cloneProject(project Project) Project {
	raw, _ := json.Marshal(project)
	var out Project
	_ = json.Unmarshal(raw, &out)
	if out.Messages == nil {
		out.Messages = []Message{}
	}
	if out.Canvas.Nodes == nil {
		out.Canvas.Nodes = []Node{}
	}
	if out.Canvas.Edges == nil {
		out.Canvas.Edges = []Edge{}
	}
	if out.Executions == nil {
		out.Executions = []Execution{}
	}
	return out
}

func (s *MemoryStore) ListProjects(_ context.Context, owner string) ([]ProjectSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ProjectSummary, 0)
	for key, project := range s.projects {
		if project.DeletedAt != nil || !strings.HasPrefix(key, owner+"\x00") {
			continue
		}
		out = append(out, ProjectSummary{ID: project.ID, Title: project.Title, TitleMode: project.TitleMode, Preview: project.Preview, MessageCount: len(project.Messages), CreatedAt: project.CreatedAt, UpdatedAt: project.UpdatedAt})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (s *MemoryStore) Project(_ context.Context, owner, id string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[projectKey(owner, id)]
	if !ok || project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	return cloneProject(project), nil
}

func normalizeProject(project Project) Project {
	project.ID = strings.TrimSpace(project.ID)
	project.Title = strings.TrimSpace(project.Title)
	if project.Title == "" {
		project.Title = "新项目"
	}
	if project.TitleMode == "" {
		project.TitleMode = "automatic"
	}
	if project.Messages == nil {
		project.Messages = []Message{}
	}
	if project.Canvas.Version == 0 {
		project.Canvas = EmptyCanvas()
	} else {
		project.Canvas.Appearance = normalizeCanvasAppearance(project.Canvas.Appearance, project.Canvas.appearanceConfigured)
		project.Canvas.appearanceConfigured = true
	}
	if project.Executions == nil {
		project.Executions = []Execution{}
	}
	return project
}

func (s *MemoryStore) CreateProject(_ context.Context, owner string, project Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, project.ID)
	if _, exists := s.projects[key]; exists {
		return ErrProjectExists
	}
	s.projects[key] = cloneProject(normalizeProject(project))
	return nil
}

func (s *MemoryStore) RenameProject(_ context.Context, owner, id, title string, now time.Time) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	project.Title = strings.TrimSpace(title)
	project.TitleMode = "manual"
	project.UpdatedAt = now
	s.projects[key] = project
	return cloneProject(project), nil
}

func (s *MemoryStore) ClearMessages(_ context.Context, owner, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	project.Messages = []Message{}
	project.Preview = ""
	project.UpdatedAt = now
	s.projects[key] = project
	return nil
}

func (s *MemoryStore) DeleteProject(_ context.Context, owner, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	project.DeletedAt = &now
	project.UpdatedAt = now
	s.projects[key] = project
	return nil
}

func (s *MemoryStore) AppendExchange(_ context.Context, owner, id string, user, assistant Message) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	project.Messages = append(project.Messages, user, assistant)
	project.Preview = assistant.Content
	if len(project.Preview) > 512 {
		project.Preview = project.Preview[:512]
	}
	project.UpdatedAt = assistant.CreatedAt
	s.projects[key] = project
	return cloneProject(project), nil
}

func (s *MemoryStore) Canvas(_ context.Context, owner, id string) (Canvas, error) {
	project, err := s.Project(context.Background(), owner, id)
	if err != nil {
		return Canvas{}, err
	}
	return project.Canvas, nil
}

func (s *MemoryStore) SaveCanvas(_ context.Context, owner, id string, canvas Canvas, expected int64, now time.Time) (Canvas, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Canvas{}, ErrProjectNotFound
	}
	if project.Canvas.Revision != expected {
		return Canvas{}, ErrRevisionConflict
	}
	canvas.Version = 1
	canvas.Revision = expected + 1
	canvas.Appearance = normalizeCanvasAppearance(canvas.Appearance, canvas.appearanceConfigured)
	canvas.appearanceConfigured = true
	if canvas.Nodes == nil {
		canvas.Nodes = []Node{}
	}
	if canvas.Edges == nil {
		canvas.Edges = []Edge{}
	}
	project.Canvas = canvas
	project.UpdatedAt = now
	s.projects[key] = project
	return cloneProject(project).Canvas, nil
}

func (s *MemoryStore) Executions(_ context.Context, owner, id string) ([]Execution, error) {
	project, err := s.Project(context.Background(), owner, id)
	if err != nil {
		return nil, err
	}
	return project.Executions, nil
}

func (s *MemoryStore) ImportProject(ctx context.Context, owner string, project Project) error {
	err := s.CreateProject(ctx, owner, project)
	if err == ErrProjectExists {
		return nil
	}
	return err
}

func (s *MemoryStore) LegacyImport(_ context.Context, owner, path string) (LegacyImport, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.imports[importKey(owner, path)]
	return record, ok, nil
}

func (s *MemoryStore) RecordLegacyImport(_ context.Context, record LegacyImport) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.imports[importKey(record.OwnerUsername, record.SourcePath)] = record
	return nil
}
