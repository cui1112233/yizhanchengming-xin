package screenplay

import (
	"context"
	"sort"
	"time"
)

func cloneUserStyle(v UserStyle) UserStyle { v.Images = append([]string{}, v.Images...); return v }
func (m *MemoryStore) ListUserStyles(_ context.Context, owner string) ([]UserStyle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []UserStyle{}
	for _, v := range m.userStyles {
		if v.OwnerUsername == owner {
			out = append(out, cloneUserStyle(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *MemoryStore) GetUserStyle(_ context.Context, owner, id string) (UserStyle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.userStyles[projectKey(owner, id)]
	if !ok {
		return v, ErrNotFound
	}
	return cloneUserStyle(v), nil
}
func (m *MemoryStore) SaveUserStyle(_ context.Context, v UserStyle, expected *int64) (UserStyle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v.ID == "" || v.OwnerUsername == "" || v.Name == "" || v.Prompt == "" {
		return v, ErrInvalidArgument
	}
	key := projectKey(v.OwnerUsername, v.ID)
	old, ok := m.userStyles[key]
	if expected == nil && ok {
		return v, ErrRevisionConflict
	}
	if expected != nil && (!ok || old.Revision != *expected) {
		return v, ErrRevisionConflict
	}
	v.CreatedAt = old.CreatedAt
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	if expected != nil {
		v.Revision = old.Revision + 1
	}
	v.UpdatedAt = time.Now().UTC()
	if m.userStyles == nil {
		m.userStyles = map[string]UserStyle{}
	}
	m.userStyles[key] = cloneUserStyle(v)
	return cloneUserStyle(v), nil
}
func (m *MemoryStore) DeleteUserStyle(_ context.Context, owner, id string, expected int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, id)
	old, ok := m.userStyles[key]
	if !ok {
		return ErrNotFound
	}
	if old.Revision != expected {
		return ErrRevisionConflict
	}
	delete(m.userStyles, key)
	return nil
}
func (m *MemoryStore) SaveCreativeConfig(_ context.Context, owner, id string, c CreativeConfig, expected int64) (Project, error) {
	if err := ValidateCreativeVideoConfig(c); err != nil {
		return Project{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, id)
	p, ok := m.projects[key]
	if !ok {
		return Project{}, ErrNotFound
	}
	if p.Revision != expected {
		return Project{}, ErrRevisionConflict
	}
	applyCreativeConfig(&p, c)
	p.Revision++
	p.UpdatedAt = time.Now().UTC()
	m.projects[key] = p
	p.CreativeConfig = cloneCreativeConfig(p.CreativeConfig)
	return p, nil
}
