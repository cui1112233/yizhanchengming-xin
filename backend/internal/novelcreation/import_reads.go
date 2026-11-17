package novelcreation

import (
	"context"
	"sort"
	"strings"
)

// Import choices expose metadata only. The selected IDs are validated again by ImportChapters.
type StyleProfileSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (s *Service) ListStyleProfiles(ctx context.Context, owner string) ([]StyleProfileSummary, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	return s.store.ListStyleProfiles(ctx, owner)
}

// Reuse the existing complete chapter contract, without creating a generation session.
func (s *Service) ListProjectChapters(ctx context.Context, owner, project string) ([]Chapter, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	project = strings.TrimSpace(project)
	if _, err = s.store.GetProject(ctx, owner, project); err != nil {
		return nil, err
	}
	return s.store.ListChapters(ctx, owner, project)
}

func (s *MemoryStore) ListStyleProfiles(_ context.Context, owner string) ([]StyleProfileSummary, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]StyleProfileSummary, 0)
	for key, style := range s.styles {
		if strings.HasPrefix(key, memoryKey(owner)+"\x00") {
			items = append(items, StyleProfileSummary{ID: style.ID, Title: style.Title})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func (s *MySQLStore) ListStyleProfiles(ctx context.Context, owner string) ([]StyleProfileSummary, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT style_profile_id, title FROM novel_style_profiles WHERE owner_username = ? ORDER BY style_profile_id`, owner)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	items := make([]StyleProfileSummary, 0)
	for rows.Next() {
		var item StyleProfileSummary
		if err = rows.Scan(&item.ID, &item.Title); err != nil {
			return nil, persistenceError(err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return items, nil
}
