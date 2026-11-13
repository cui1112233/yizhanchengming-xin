package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type MediaItem struct {
	Name       string    `json:"name,omitempty"`
	DurationMS int64     `json:"durationMs,omitempty"`
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Status     string    `json:"status"`
	URL        string    `json:"url,omitempty"`
	AssetID    string    `json:"assetId,omitempty"`
	ShotID     string    `json:"shotId,omitempty"`
	TargetID   string    `json:"targetId,omitempty"`
	Error      string    `json:"error,omitempty"`
	ProviderID string    `json:"providerId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}
type MediaState struct {
	Revision int64             `json:"revision"`
	Items    []MediaItem       `json:"items"`
	Selected map[string]string `json:"selected"`
	Voices   map[string]string `json:"voices"`
}
type MediaStore interface {
	LoadMedia(context.Context, string, string) (MediaState, error)
	MutateMedia(context.Context, string, string, func(*MediaState) error) (MediaState, error)
}

func initMedia(s *MediaState) {
	if s.Items == nil {
		s.Items = []MediaItem{}
	}
	if s.Selected == nil {
		s.Selected = map[string]string{}
	}
	if s.Voices == nil {
		s.Voices = map[string]string{}
	}
}

type MediaMemoryStore struct {
	mu     sync.Mutex
	states map[string]MediaState
}

func NewMediaMemoryStore() *MediaMemoryStore {
	return &MediaMemoryStore{states: map[string]MediaState{}}
}
func cloneMedia(s MediaState) MediaState {
	b, _ := json.Marshal(s)
	var out MediaState
	_ = json.Unmarshal(b, &out)
	initMedia(&out)
	return out
}
func (s *MediaMemoryStore) LoadMedia(_ context.Context, o, p string) (MediaState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneMedia(s.states[o+"\x00"+p]), nil
}
func (s *MediaMemoryStore) MutateMedia(_ context.Context, o, p string, fn func(*MediaState) error) (MediaState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := cloneMedia(s.states[o+"\x00"+p])
	if err := fn(&v); err != nil {
		return v, err
	}
	v.Revision++
	s.states[o+"\x00"+p] = cloneMedia(v)
	return v, nil
}
func (s *MySQLStore) LoadMedia(ctx context.Context, o, p string) (MediaState, error) {
	var raw []byte
	var out MediaState
	err := s.db.QueryRowContext(ctx, "SELECT state_json FROM screenplay_media_workbench WHERE owner_username=? AND project_id=?", o, p).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		initMedia(&out)
		return out, nil
	}
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(raw, &out)
	initMedia(&out)
	return out, err
}
func (s *MySQLStore) MutateMedia(ctx context.Context, o, p string, fn func(*MediaState) error) (MediaState, error) {
	var out MediaState
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT IGNORE INTO screenplay_media_workbench(owner_username,project_id,state_json) VALUES (?,?,'{}')", o, p)
	if err != nil {
		return out, err
	}
	var raw []byte
	err = tx.QueryRowContext(ctx, "SELECT state_json FROM screenplay_media_workbench WHERE owner_username=? AND project_id=? FOR UPDATE", o, p).Scan(&raw)
	if err != nil {
		return out, err
	}
	if err = json.Unmarshal(raw, &out); err != nil {
		return out, err
	}
	initMedia(&out)
	if err = fn(&out); err != nil {
		return out, err
	}
	out.Revision++
	raw, err = json.Marshal(out)
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, "UPDATE screenplay_media_workbench SET state_json=? WHERE owner_username=? AND project_id=?", raw, o, p)
	if err != nil {
		return out, err
	}
	return out, tx.Commit()
}
