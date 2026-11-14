package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type StoryboardGenerationCall struct {
	Status      string `json:"status"`
	InputDigest string `json:"input_digest"`
	RawResponse string `json:"raw_response,omitempty"`
	Error       string `json:"error,omitempty"`
}
type StoryboardGenerationEpisode struct {
	EpisodeNo     int                                 `json:"episode_no"`
	Status        string                              `json:"status"`
	Error         string                              `json:"error,omitempty"`
	Calls         map[string]StoryboardGenerationCall `json:"calls"`
	Storyboard    *Storyboard                         `json:"storyboard,omitempty"`
	Warnings      []string                            `json:"warnings,omitempty"`
	AddedAssetIDs []string                            `json:"added_asset_ids,omitempty"`
	PendingAssets []Asset                             `json:"pending_assets,omitempty"`
}
type StoryboardGenerationAttempt struct {
	RequestKey        string                               `json:"request_key"`
	RetryOf           string                               `json:"retry_of,omitempty"`
	InputDigest       string                               `json:"input_digest"`
	Engine            StoryboardEngine                     `json:"engine"`
	Status            string                               `json:"status"`
	ChapterID         string                               `json:"chapter_id,omitempty"`
	PlanningChapterID string                               `json:"planning_chapter_id,omitempty"`
	UseOriginal       bool                                 `json:"use_original_content_directly"`
	UsePlanning       *bool                                `json:"use_planning"`
	PlanningApplied   bool                                 `json:"planning_applied"`
	EpisodeNumbers    []int                                `json:"episode_numbers"`
	Episodes          map[int]*StoryboardGenerationEpisode `json:"episodes"`
	Result            StoryboardResult                     `json:"result"`
	Error             string                               `json:"error,omitempty"`
	CancelRequested   bool                                 `json:"cancel_requested"`
	CreatedAt         time.Time                            `json:"created_at"`
	UpdatedAt         time.Time                            `json:"updated_at"`
	Roles             string                               `json:"roles"`
	Scenes            string                               `json:"scenes"`
	Props             string                               `json:"props"`
	ExternalHistory   map[int][]Storyboard                 `json:"external_history"`
}
type StoryboardGenerationState struct {
	Revision int64                                   `json:"revision"`
	Attempts map[string]*StoryboardGenerationAttempt `json:"attempts"`
}
type StoryboardGenerationStore interface {
	LoadStoryboardGenerations(context.Context, string, string) (StoryboardGenerationState, error)
	MutateStoryboardGenerations(context.Context, string, string, func(*StoryboardGenerationState) error) (StoryboardGenerationState, error)
	CommitStoryboardGeneration(context.Context, string, string, string, int, Storyboard, []string) (StoryboardGenerationAttempt, error)
}

func cloneStoryboardGenerations(v StoryboardGenerationState) StoryboardGenerationState {
	raw, _ := json.Marshal(v)
	var out StoryboardGenerationState
	_ = json.Unmarshal(raw, &out)
	if out.Attempts == nil {
		out.Attempts = map[string]*StoryboardGenerationAttempt{}
	}
	return out
}
func (m *MemoryStore) LoadStoryboardGenerations(_ context.Context, owner, project string) (StoryboardGenerationState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneStoryboardGenerations(m.storyboardGenerations[projectKey(owner, project)]), nil
}
func (m *MemoryStore) MutateStoryboardGenerations(ctx context.Context, owner, project string, fn func(*StoryboardGenerationState) error) (StoryboardGenerationState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := cloneStoryboardGenerations(m.storyboardGenerations[projectKey(owner, project)])
	if ctx.Err() != nil {
		return v, ctx.Err()
	}
	if e := fn(&v); e != nil {
		return v, e
	}
	v.Revision++
	if m.storyboardGenerations == nil {
		m.storyboardGenerations = map[string]StoryboardGenerationState{}
	}
	m.storyboardGenerations[projectKey(owner, project)] = cloneStoryboardGenerations(v)
	return v, nil
}
func (s *MySQLStore) LoadStoryboardGenerations(ctx context.Context, owner, project string) (StoryboardGenerationState, error) {
	var raw []byte
	var v StoryboardGenerationState
	e := s.db.QueryRowContext(ctx, "SELECT state_json FROM screenplay_storyboard_generations WHERE owner_username=? AND project_id=?", owner, project).Scan(&raw)
	if errors.Is(e, sql.ErrNoRows) {
		return cloneStoryboardGenerations(v), nil
	}
	if e != nil {
		return v, e
	}
	e = json.Unmarshal(raw, &v)
	return cloneStoryboardGenerations(v), e
}
func loadStoryboardGenerationsTx(ctx context.Context, tx *sql.Tx, owner, project string) (StoryboardGenerationState, error) {
	var v StoryboardGenerationState
	if _, e := tx.ExecContext(ctx, "INSERT IGNORE INTO screenplay_storyboard_generations(owner_username,project_id,state_json) VALUES (?,?,'{}')", owner, project); e != nil {
		return v, e
	}
	var raw []byte
	e := tx.QueryRowContext(ctx, "SELECT state_json FROM screenplay_storyboard_generations WHERE owner_username=? AND project_id=? FOR UPDATE", owner, project).Scan(&raw)
	if e != nil {
		return v, e
	}
	if e = json.Unmarshal(raw, &v); e != nil {
		return v, e
	}
	return cloneStoryboardGenerations(v), nil
}
func saveStoryboardGenerationsTx(ctx context.Context, tx *sql.Tx, owner, project string, v StoryboardGenerationState) error {
	raw, e := json.Marshal(v)
	if e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, "UPDATE screenplay_storyboard_generations SET state_json=? WHERE owner_username=? AND project_id=?", string(raw), owner, project)
	return e
}
func (s *MySQLStore) MutateStoryboardGenerations(ctx context.Context, owner, project string, fn func(*StoryboardGenerationState) error) (StoryboardGenerationState, error) {
	var v StoryboardGenerationState
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return v, e
	}
	defer tx.Rollback()
	v, e = loadStoryboardGenerationsTx(ctx, tx, owner, project)
	if e != nil {
		return v, e
	}
	if e = fn(&v); e != nil {
		return v, e
	}
	v.Revision++
	if e = saveStoryboardGenerationsTx(ctx, tx, owner, project, v); e != nil {
		return v, e
	}
	return v, tx.Commit()
}
func generationCommitReady(v StoryboardGenerationState, key string, episode int) (*StoryboardGenerationAttempt, *StoryboardGenerationEpisode, error) {
	a := v.Attempts[key]
	if a == nil || a.Status != "running" {
		return nil, nil, ErrInvalidState
	}
	ep := a.Episodes[episode]
	if ep == nil || ep.Status == "succeeded" {
		return nil, nil, ErrInvalidState
	}
	return a, ep, nil
}
func commitGenerationResult(a *StoryboardGenerationAttempt, ep *StoryboardGenerationEpisode, board Storyboard, warnings []string) {
	ep.Status = "succeeded"
	ep.Error = ""
	ep.Storyboard = &board
	ep.Warnings = warnings
	a.UpdatedAt = time.Now().UTC()
	a.Result = generationAttemptResult(*a)
}
func (m *MemoryStore) CommitStoryboardGeneration(ctx context.Context, owner, project, key string, episode int, board Storyboard, warnings []string) (StoryboardGenerationAttempt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := cloneStoryboardGenerations(m.storyboardGenerations[projectKey(owner, project)])
	a, ep, e := generationCommitReady(v, key, episode)
	if e != nil {
		return StoryboardGenerationAttempt{}, e
	}
	if board.OwnerUsername != owner || board.ProjectID != project || board.EpisodeNo != episode || board.Engine != a.Engine {
		return StoryboardGenerationAttempt{}, ErrInvalidArgument
	}
	if ctx.Err() != nil {
		return StoryboardGenerationAttempt{}, ctx.Err()
	}
	saved, e := m.saveStoryboardLocked(board)
	if e != nil {
		return StoryboardGenerationAttempt{}, e
	}
	commitGenerationResult(a, ep, saved, warnings)
	v.Revision++
	m.storyboardGenerations[projectKey(owner, project)] = cloneStoryboardGenerations(v)
	return *a, nil
}
func (s *MySQLStore) CommitStoryboardGeneration(ctx context.Context, owner, project, key string, episode int, board Storyboard, warnings []string) (StoryboardGenerationAttempt, error) {
	var zero StoryboardGenerationAttempt
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return zero, e
	}
	defer tx.Rollback()
	v, e := loadStoryboardGenerationsTx(ctx, tx, owner, project)
	if e != nil {
		return zero, e
	}
	a, ep, e := generationCommitReady(v, key, episode)
	if e != nil {
		return zero, e
	}
	if board.OwnerUsername != owner || board.ProjectID != project || board.EpisodeNo != episode || board.Engine != a.Engine {
		return zero, fmt.Errorf("%w: 分镜与生成目标不匹配", ErrInvalidArgument)
	}
	saved, e := saveStoryboardInTx(ctx, tx, board)
	if e != nil {
		return zero, e
	}
	commitGenerationResult(a, ep, saved, warnings)
	v.Revision++
	if e = saveStoryboardGenerationsTx(ctx, tx, owner, project, v); e != nil {
		return zero, e
	}
	if e = tx.Commit(); e != nil {
		return zero, e
	}
	return *a, nil
}
