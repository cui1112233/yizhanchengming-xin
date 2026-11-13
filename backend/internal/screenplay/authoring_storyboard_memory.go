package screenplay

import (
	"context"
	"time"
)

func (m *MemoryStore) SaveStoryboardRevision(_ context.Context, storyboard Storyboard, expected int64) (Storyboard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(storyboard.OwnerUsername, storyboard.ProjectID)
	if m.storyboards == nil {
		m.storyboards = map[string][]Storyboard{}
	}
	stored := m.storyboards[key]
	now := time.Now().UTC()
	storyboard.ShotCount = len(storyboard.Shots)
	if storyboard.Status == "" {
		storyboard.Status = "done"
	}
	for index := range storyboard.Shots {
		shot := &storyboard.Shots[index]
		if shot.ID == "" {
			shot.ID = newScreenplayID("sb_shot")
		}
		if shot.Seq <= 0 {
			shot.Seq = index + 1
		}
		shot.ProjectID = storyboard.ProjectID
		shot.OwnerUsername = storyboard.OwnerUsername
	}
	for index := range stored {
		if stored[index].EpisodeNo != storyboard.EpisodeNo || stored[index].Engine != storyboard.Engine {
			continue
		}
		if stored[index].Revision != expected {
			return Storyboard{}, ErrRevisionConflict
		}
		storyboard.ID = stored[index].ID
		if storyboard.ID == "" {
			storyboard.ID = newScreenplayID("sb")
		}
		storyboard.Revision = stored[index].Revision + 1
		storyboard.CreatedAt = stored[index].CreatedAt
		storyboard.UpdatedAt = now
		for shotIndex := range storyboard.Shots {
			storyboard.Shots[shotIndex].StoryboardID = storyboard.ID
		}
		stored[index] = cloneStoryboard(storyboard)
		m.storyboards[key] = stored
		return cloneStoryboard(stored[index]), nil
	}
	if storyboard.ID == "" {
		storyboard.ID = newScreenplayID("sb")
	}
	for shotIndex := range storyboard.Shots {
		storyboard.Shots[shotIndex].StoryboardID = storyboard.ID
	}
	if expected != 0 {
		return Storyboard{}, ErrRevisionConflict
	}
	storyboard.Revision = 1
	storyboard.CreatedAt = now
	storyboard.UpdatedAt = now
	m.storyboards[key] = append(stored, cloneStoryboard(storyboard))
	return cloneStoryboard(storyboard), nil
}
