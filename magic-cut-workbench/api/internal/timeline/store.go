package timeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const selectLatestRevisionForUpdateSQL = `SELECT revision FROM magic_cut_timeline_snapshots WHERE owner_username = ? AND project_id = ? ORDER BY revision DESC LIMIT 1 FOR UPDATE`
const selectLatestTimelineSQL = `SELECT timeline_json FROM magic_cut_timeline_snapshots WHERE owner_username = ? AND project_id = ? ORDER BY revision DESC LIMIT 1`
const insertTimelineSQL = `INSERT INTO magic_cut_timeline_snapshots (project_id, owner_username, revision, timeline_json, created_at) VALUES (?, ?, ?, ?, ?)`

type Store struct {
	db  *sql.DB
	now func() time.Time
}

func NewStore(db *sql.DB, now func() time.Time) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{db: db, now: now}
}

func Empty() Timeline {
	return Timeline{Revision: 0, DurationMs: 0, Tracks: []Track{{ID: "video-1", Kind: "video", Clips: []Clip{}}, {ID: "audio-1", Kind: "audio", Clips: []Clip{}}}}
}

func (s *Store) Get(ctx context.Context, owner, projectID string) (Timeline, error) {
	if s == nil || s.db == nil {
		return Timeline{}, errors.New("timeline store is not configured")
	}
	var payload []byte
	err := s.db.QueryRowContext(ctx, selectLatestTimelineSQL, owner, projectID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return Empty(), nil
	}
	if err != nil {
		return Timeline{}, err
	}
	var value Timeline
	if err := json.Unmarshal(payload, &value); err != nil {
		return Timeline{}, err
	}
	return value, nil
}

func (s *Store) Put(ctx context.Context, owner, projectID string, value Timeline, expectedRevision uint64) (Timeline, error) {
	if s == nil || s.db == nil {
		return Timeline{}, errors.New("timeline store is not configured")
	}
	if err := Validate(value); err != nil {
		return Timeline{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Timeline{}, err
	}
	defer tx.Rollback()
	var current uint64
	err = tx.QueryRowContext(ctx, selectLatestRevisionForUpdateSQL, owner, projectID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		current, err = 0, nil
	}
	if err != nil {
		return Timeline{}, err
	}
	if current != expectedRevision {
		return Timeline{}, &RevisionConflictError{CurrentRevision: current}
	}
	value.Revision = current + 1
	payload, err := json.Marshal(value)
	if err != nil {
		return Timeline{}, err
	}
	if _, err = tx.ExecContext(ctx, insertTimelineSQL, projectID, owner, value.Revision, payload, s.now().UTC()); err != nil {
		return Timeline{}, fmt.Errorf("insert timeline: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return Timeline{}, err
	}
	return value, nil
}
