package agentstudio

import (
	"context"
	"sync"
	"time"
)

type memoryStore struct {
	mu          sync.Mutex
	next        int64
	projects    map[int64]Project
	messages    map[int64][]Message
	executions  map[int64]Execution
	attachments map[int64]Attachment
	canvases    map[int64]Canvas
	versions    map[int64][]CanvasVersion
	skills      map[int64]Skill
}

func newMemoryStore() *memoryStore {
	return &memoryStore{projects: map[int64]Project{}, messages: map[int64][]Message{}, executions: map[int64]Execution{}, attachments: map[int64]Attachment{}, canvases: map[int64]Canvas{}, versions: map[int64][]CanvasVersion{}, skills: map[int64]Skill{}}
}
func (s *memoryStore) id() int64 { s.next++; return s.next }
func (s *memoryStore) CreateProject(_ context.Context, a Actor, in CreateProjectInput) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := time.Now().UTC()
	p := Project{ID: s.id(), OwnerUserID: a.UserID, TeamID: a.TeamID, Title: in.Title, BatchProjectID: in.BatchProjectID, BookID: in.BookID, CreatedAt: n, UpdatedAt: n}
	s.projects[p.ID] = p
	return p, nil
}
func (s *memoryStore) GetProject(_ context.Context, a Actor, id int64) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[id]
	if !ok {
		return Project{}, ErrNotFound
	}
	if p.OwnerUserID != a.UserID {
		return Project{}, ErrForbidden
	}
	return p, nil
}
func (s *memoryStore) CreateMessage(_ context.Context, m Message) (Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m.ID = s.id()
	s.messages[m.ProjectID] = append(s.messages[m.ProjectID], m)
	return m, nil
}
func (s *memoryStore) CreateExecution(_ context.Context, e Execution) (Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = s.id()
	s.executions[e.ID] = e
	return e, nil
}
func (s *memoryStore) UpdateExecution(_ context.Context, e Execution) (Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[e.ID] = e
	return e, nil
}
func (s *memoryStore) ListProjects(_ context.Context, a Actor) ([]Project, error) {
	out := []Project{}
	for _, p := range s.projects {
		if p.OwnerUserID == a.UserID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s *memoryStore) ListMessages(_ context.Context, a Actor, id int64) ([]Message, error) {
	if _, e := s.GetProject(context.Background(), a, id); e != nil {
		return nil, e
	}
	return s.messages[id], nil
}
func (s *memoryStore) DeleteProject(_ context.Context, a Actor, id int64) error {
	if _, e := s.GetProject(context.Background(), a, id); e != nil {
		return e
	}
	delete(s.projects, id)
	return nil
}
func (s *memoryStore) ListExecutions(_ context.Context, a Actor, id int64) ([]Execution, error) {
	if _, e := s.GetProject(context.Background(), a, id); e != nil {
		return nil, e
	}
	out := []Execution{}
	for _, v := range s.executions {
		if v.ProjectID == id {
			out = append(out, v)
		}
	}
	return out, nil
}
func (s *memoryStore) CreateSkill(_ context.Context, a Actor, in CreateSkillInput) (Skill, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := Skill{ID: s.id(), OwnerUserID: a.UserID, Name: in.Name, Body: in.Body, Version: 1, Enabled: true, CreatedAt: time.Now().UTC()}
	s.skills[v.ID] = v
	return v, nil
}
func (s *memoryStore) ListSkills(_ context.Context, a Actor) ([]Skill, error) {
	out := []Skill{}
	for _, v := range s.skills {
		if v.OwnerUserID == a.UserID {
			out = append(out, v)
		}
	}
	return out, nil
}
func (s *memoryStore) GetSkills(ctx context.Context, a Actor, ids []int64) ([]Skill, error) {
	all, _ := s.ListSkills(ctx, a)
	m := map[int64]Skill{}
	for _, v := range all {
		m[v.ID] = v
	}
	out := []Skill{}
	for _, id := range ids {
		v, ok := m[id]
		if !ok {
			return nil, ErrForbidden
		}
		out = append(out, v)
	}
	return out, nil
}
func (s *memoryStore) GetCanvas(_ context.Context, a Actor, id int64) (Canvas, error) {
	if _, e := s.GetProject(context.Background(), a, id); e != nil {
		return Canvas{}, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.canvases[id]; ok {
		c.Document = append([]byte(nil), c.Document...)
		return c, nil
	}
	return Canvas{ProjectID: id, Revision: 0, Document: []byte(`{"nodes":[],"edges":[]}`)}, nil
}
func (s *memoryStore) SaveCanvas(_ context.Context, a Actor, c Canvas) (Canvas, error) {
	if _, e := s.GetProject(context.Background(), a, c.ProjectID); e != nil {
		return Canvas{}, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.canvases[c.ProjectID]
	if exists && current.Revision != c.Revision {
		return Canvas{}, ErrConflict
	}
	if !exists && c.Revision != 0 {
		return Canvas{}, ErrConflict
	}
	c.Revision++
	c.Document = append([]byte(nil), c.Document...)
	s.canvases[c.ProjectID] = c
	v := CanvasVersion{ID: s.id(), ProjectID: c.ProjectID, Revision: c.Revision, Document: append([]byte(nil), c.Document...), CreatedAt: time.Now().UTC()}
	s.versions[c.ProjectID] = append(s.versions[c.ProjectID], v)
	return c, nil
}
func (s *memoryStore) ListCanvasVersions(_ context.Context, a Actor, id int64) ([]CanvasVersion, error) {
	if _, e := s.GetProject(context.Background(), a, id); e != nil {
		return nil, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	versions := append([]CanvasVersion(nil), s.versions[id]...)
	for i := range versions {
		versions[i].Document = append([]byte(nil), versions[i].Document...)
	}
	return versions, nil
}
func (s *memoryStore) RestoreCanvas(_ context.Context, a Actor, id int64, revision int) (Canvas, error) {
	if _, e := s.GetProject(context.Background(), a, id); e != nil {
		return Canvas{}, e
	}
	s.mu.Lock()
	var restored []byte
	for _, v := range s.versions[id] {
		if v.Revision == revision {
			restored = append([]byte(nil), v.Document...)
			break
		}
	}
	current, ok := s.canvases[id]
	s.mu.Unlock()
	if restored == nil {
		return Canvas{}, ErrNotFound
	}
	if !ok {
		return Canvas{}, ErrNotFound
	}
	return s.SaveCanvas(context.Background(), a, Canvas{ProjectID: id, Revision: current.Revision, Document: restored})
}
func (s *memoryStore) CreateAttachment(_ context.Context, a Actor, v Attachment) (Attachment, error) {
	if _, e := s.GetProject(context.Background(), a, v.ProjectID); e != nil {
		return Attachment{}, e
	}
	v.ID = s.id()
	s.attachments[v.ID] = v
	return v, nil
}
func (s *memoryStore) ListAttachments(_ context.Context, a Actor, p int64) ([]Attachment, error) {
	if _, e := s.GetProject(context.Background(), a, p); e != nil {
		return nil, e
	}
	out := []Attachment{}
	for _, v := range s.attachments {
		if v.ProjectID == p {
			out = append(out, v)
		}
	}
	return out, nil
}
func (s *memoryStore) GetAttachment(_ context.Context, a Actor, p, id int64) (Attachment, error) {
	if _, e := s.GetProject(context.Background(), a, p); e != nil {
		return Attachment{}, e
	}
	v, ok := s.attachments[id]
	if !ok || v.ProjectID != p {
		return Attachment{}, ErrNotFound
	}
	return v, nil
}
func (s *memoryStore) DeleteAttachment(_ context.Context, a Actor, p, id int64) error {
	if _, e := s.GetAttachment(context.Background(), a, p, id); e != nil {
		return e
	}
	delete(s.attachments, id)
	return nil
}
