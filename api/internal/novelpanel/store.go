package novelpanel

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Store interface {
	GetWorkspace(context.Context, int64) (Workspace, error)
	SaveWorkspace(context.Context, int64, Workspace, string) (Workspace, HistoryRecord, error)
	ListHistory(context.Context, int64, int) ([]HistoryRecord, error)
	GetHistory(context.Context, int64, string) (HistoryRecord, error)
}

type MemoryStore struct {
	mu         sync.Mutex
	workspaces map[int64]Workspace
	history    map[int64][]HistoryRecord
	sequence   uint64
	now        func() time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		workspaces: map[int64]Workspace{},
		history:    map[int64][]HistoryRecord{},
		now:        time.Now,
	}
}

func cloneJSON[T any](value T) (T, error) {
	var result T
	raw, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (s *MemoryStore) GetWorkspace(ctx context.Context, projectID int64) (Workspace, error) {
	if err := ctx.Err(); err != nil {
		return Workspace{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	workspace, ok := s.workspaces[projectID]
	if !ok {
		return Workspace{}, ErrNotFound
	}
	return cloneJSON(workspace)
}

func (s *MemoryStore) SaveWorkspace(ctx context.Context, expectedRevision int64, workspace Workspace, note string) (Workspace, HistoryRecord, error) {
	if err := ctx.Err(); err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	if workspace.ProjectID <= 0 {
		return Workspace{}, HistoryRecord{}, fmt.Errorf("%w: project id is required", ErrInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.workspaces[workspace.ProjectID]
	currentRevision := int64(0)
	if exists {
		currentRevision = current.Revision
	}
	if currentRevision != expectedRevision {
		return Workspace{}, HistoryRecord{}, fmt.Errorf("%w: expected %d, current %d", ErrConflict, expectedRevision, currentRevision)
	}
	now := s.now().UTC()
	workspace.Revision = currentRevision + 1
	if exists && !current.CreatedAt.IsZero() {
		workspace.CreatedAt = current.CreatedAt
	} else if workspace.CreatedAt.IsZero() {
		workspace.CreatedAt = now
	}
	workspace.UpdatedAt = now
	cloned, err := cloneJSON(workspace)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	s.workspaces[workspace.ProjectID] = cloned

	s.sequence++
	history := HistoryRecord{
		ID:        fmt.Sprintf("hist_%d_%d_%06d", workspace.ProjectID, workspace.Revision, s.sequence),
		ProjectID: workspace.ProjectID,
		Revision:  workspace.Revision,
		Note:      strings.TrimSpace(note),
		Summary:   SummarizeWorkspace(workspace),
		Workspace: cloned,
		CreatedAt: now,
	}
	historyClone, err := cloneJSON(history)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	s.history[workspace.ProjectID] = append(s.history[workspace.ProjectID], historyClone)
	return cloneJSONPair(cloned, historyClone)
}

func cloneJSONPair(workspace Workspace, history HistoryRecord) (Workspace, HistoryRecord, error) {
	workspaceClone, err := cloneJSON(workspace)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	historyClone, err := cloneJSON(history)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	return workspaceClone, historyClone, nil
}

func (s *MemoryStore) ListHistory(ctx context.Context, projectID int64, limit int) ([]HistoryRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items := append([]HistoryRecord(nil), s.history[projectID]...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return cloneJSON(items)
}

func (s *MemoryStore) GetHistory(ctx context.Context, projectID int64, historyID string) (HistoryRecord, error) {
	if err := ctx.Err(); err != nil {
		return HistoryRecord{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.history[projectID] {
		if item.ID == historyID {
			return cloneJSON(item)
		}
	}
	return HistoryRecord{}, ErrNotFound
}

func SummarizeWorkspace(workspace Workspace) HistorySummary {
	preview := strings.Join(strings.Fields(workspace.OriginalText), " ")
	previewRunes := []rune(preview)
	if len(previewRunes) > 80 {
		preview = string(previewRunes[:80])
	}
	duration := 0.0
	for _, shot := range workspace.Shots {
		duration += shot.DurationSec
	}
	return HistorySummary{
		SourcePreview:  preview,
		CharacterCount: len(workspace.Characters),
		RelationCount:  len(workspace.Relationships),
		ShotCount:      len(workspace.Shots),
		DurationSec:    duration,
		Mode:           workspace.Mode,
	}
}
