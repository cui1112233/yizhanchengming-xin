package aicut

import (
	"context"
	"time"
)

func (s *MemoryStore) UpdateProjectRatio(_ context.Context, owner, id, ratio string, now time.Time) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, id)
	item, ok := s.projects[key]
	if !ok || item.project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	if ratio != "9:16" && ratio != "16:9" && ratio != "1:1" {
		return Project{}, ErrInvalidProject
	}
	item.project.AspectRatio = ratio
	item.project.UpdatedAt = now
	s.projects[key] = item
	return item.project, nil
}

func (s *MySQLStore) UpdateProjectRatio(ctx context.Context, owner, id, ratio string, now time.Time) (Project, error) {
	if ratio != "9:16" && ratio != "16:9" && ratio != "1:1" {
		return Project{}, ErrInvalidProject
	}
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_projects SET aspect_ratio = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, ratio, now, id, owner)
	if err != nil {
		return Project{}, err
	}
	// An unchanged ratio can report zero affected rows; the owner-scoped read distinguishes it from a missing project.
	if _, err = result.RowsAffected(); err != nil {
		return Project{}, err
	}
	return s.Project(ctx, owner, id)
}
