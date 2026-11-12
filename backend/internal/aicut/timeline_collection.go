package aicut

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

const DefaultTimelineID = "default"
const MaxProjectTimelines = 50

var (
	ErrTimelineNotFound          = errors.New("ai-cut timeline not found")
	ErrTimelineLimit             = errors.New("ai-cut timeline limit reached")
	ErrInvalidTimelineCollection = errors.New("invalid ai-cut timeline collection")
	ErrDefaultTimelineProtected  = errors.New("default timeline cannot be deleted")
	ErrTimelineJobUnsupported    = errors.New("job does not support named timelines")
)

type TimelineMetadata struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Order     int       `json:"order"`
	Default   bool      `json:"default"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type timelineCollectionRecord struct {
	TimelineMetadata
	Value   Timeline
	History []Timeline
	Cursor  int
}
type collectionRepository interface {
	access(context.Context, string, string, string, func(*[]timelineCollectionRecord) error) error
}
type TimelineCollectionStore struct {
	base Store
	repo collectionRepository
}
type memoryCollectionRepository struct {
	mu       sync.Mutex
	projects map[string][]timelineCollectionRecord
}

func NewMemoryTimelineCollectionStore(base Store) *TimelineCollectionStore {
	return &TimelineCollectionStore{base: base, repo: &memoryCollectionRepository{projects: map[string][]timelineCollectionRecord{}}}
}
func (r *memoryCollectionRepository) access(_ context.Context, o, p, target string, fn func(*[]timelineCollectionRecord) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := aicutProjectKey(o, p)
	items := append([]timelineCollectionRecord(nil), r.projects[key]...)
	for i := range items {
		items[i].Value = cloneTimeline(items[i].Value)
		items[i].History = append([]Timeline(nil), items[i].History...)
	}
	if len(items) == 0 {
		items = []timelineCollectionRecord{defaultCollectionRecord()}
	}
	if e := fn(&items); e != nil {
		return e
	}
	r.projects[key] = items
	return nil
}
func defaultCollectionRecord() timelineCollectionRecord {
	return timelineCollectionRecord{TimelineMetadata: TimelineMetadata{ID: DefaultTimelineID, Title: "默认时间线", Default: true}}
}
func collectionID() string {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return "tl_" + hex.EncodeToString(b[:])
}
func collectionTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || len([]rune(title)) > 100 {
		return "", ErrInvalidTimelineCollection
	}
	return title, nil
}
func (s *TimelineCollectionStore) access(ctx context.Context, o, p, target string, fn func(*[]timelineCollectionRecord) error) error {
	if _, e := s.base.Project(ctx, o, p); e != nil {
		return e
	}
	return s.repo.access(ctx, o, p, target, fn)
}
func collectionIndex(items []timelineCollectionRecord, id string) int {
	for i := range items {
		if items[i].ID == id {
			return i
		}
	}
	return -1
}
func (s *TimelineCollectionStore) List(ctx context.Context, o, p string) (out []TimelineMetadata, e error) {
	out = []TimelineMetadata{}
	e = s.access(ctx, o, p, "", func(items *[]timelineCollectionRecord) error {
		for _, r := range *items {
			out = append(out, r.TimelineMetadata)
		}
		return nil
	})
	return
}
func (s *TimelineCollectionStore) Create(ctx context.Context, o, p, title string) (TimelineMetadata, error) {
	return s.create(ctx, o, p, title, Timeline{Tracks: []Track{}})
}
func (s *TimelineCollectionStore) create(ctx context.Context, o, p, title string, value Timeline) (out TimelineMetadata, e error) {
	title, e = collectionTitle(title)
	if e != nil {
		return
	}
	value = cloneTimeline(value)
	value.Revision = 0
	now := time.Now().UTC()
	out = TimelineMetadata{ID: collectionID(), Title: title, CreatedAt: now, UpdatedAt: now}
	e = s.access(ctx, o, p, "", func(items *[]timelineCollectionRecord) error {
		if len(*items) >= MaxProjectTimelines {
			return ErrTimelineLimit
		}
		out.Order = len(*items)
		*items = append(*items, timelineCollectionRecord{TimelineMetadata: out, Value: value, History: []Timeline{cloneTimeline(value)}})
		return nil
	})
	return
}
func (s *TimelineCollectionStore) Duplicate(ctx context.Context, o, p, id, title string) (TimelineMetadata, error) {
	v, e := s.Timeline(ctx, o, p, id)
	if e != nil {
		return TimelineMetadata{}, e
	}
	return s.create(ctx, o, p, title, cloneCollectionTimeline(v))
}
func (s *TimelineCollectionStore) Rename(ctx context.Context, o, p, id, title string) (out TimelineMetadata, e error) {
	title, e = collectionTitle(title)
	if e != nil {
		return
	}
	e = s.access(ctx, o, p, "", func(items *[]timelineCollectionRecord) error {
		i := collectionIndex(*items, id)
		if i < 0 {
			return ErrTimelineNotFound
		}
		(*items)[i].Title = title
		(*items)[i].UpdatedAt = time.Now().UTC()
		out = (*items)[i].TimelineMetadata
		return nil
	})
	return
}
func (s *TimelineCollectionStore) Delete(ctx context.Context, o, p, id string) error {
	return s.access(ctx, o, p, "", func(items *[]timelineCollectionRecord) error {
		i := collectionIndex(*items, id)
		if i < 0 {
			return ErrTimelineNotFound
		}
		if id == DefaultTimelineID {
			return ErrDefaultTimelineProtected
		}
		*items = append((*items)[:i], (*items)[i+1:]...)
		for i := range *items {
			(*items)[i].Order = i
		}
		return nil
	})
}
func (s *TimelineCollectionStore) Reorder(ctx context.Context, o, p string, ids []string) error {
	return s.access(ctx, o, p, "", func(items *[]timelineCollectionRecord) error {
		if len(ids) != len(*items) {
			return ErrInvalidTimelineCollection
		}
		out := make([]timelineCollectionRecord, 0, len(ids))
		seen := map[string]bool{}
		for i, id := range ids {
			j := collectionIndex(*items, id)
			if j < 0 || seen[id] {
				return ErrInvalidTimelineCollection
			}
			seen[id] = true
			r := (*items)[j]
			r.Order = i
			out = append(out, r)
		}
		*items = out
		return nil
	})
}
func (s *TimelineCollectionStore) Timeline(ctx context.Context, o, p, id string) (out Timeline, e error) {
	if id == DefaultTimelineID {
		return s.base.Timeline(ctx, o, p)
	}
	e = s.access(ctx, o, p, id, func(items *[]timelineCollectionRecord) error {
		i := collectionIndex(*items, id)
		if i < 0 {
			return ErrTimelineNotFound
		}
		out = cloneTimeline((*items)[i].Value)
		return nil
	})
	return
}
func (s *TimelineCollectionStore) validateAssets(ctx context.Context, o, p string, v Timeline) error {
	refs := map[string]bool{}
	for _, m := range v.Markers {
		refs[m.AssetID] = true
	}
	for _, t := range v.Tracks {
		refs[t.Settings.FontAssetID] = true
		for _, c := range t.Clips {
			refs[c.AssetID] = true
			if c.VisualProcessing != nil {
				refs[c.VisualProcessing.PersonSegmentation.MatteAssetID] = true
			}
		}
	}
	for id := range refs {
		if id != "" {
			if _, e := s.base.Asset(ctx, o, p, id); e != nil {
				return e
			}
		}
	}
	return nil
}
func (s *TimelineCollectionStore) Save(ctx context.Context, o, p, id string, v Timeline, expected int64, now time.Time) (out Timeline, e error) {
	if id == DefaultTimelineID {
		return s.base.SaveTimeline(ctx, o, p, v, expected, now)
	}
	if e = validateTimeline(v); e != nil {
		return
	}
	if e = s.validateAssets(ctx, o, p, v); e != nil {
		return
	}
	e = s.access(ctx, o, p, id, func(items *[]timelineCollectionRecord) error {
		i := collectionIndex(*items, id)
		if i < 0 {
			return ErrTimelineNotFound
		}
		r := &(*items)[i]
		if expected < 0 || r.Value.Revision != expected {
			return ErrRevisionConflict
		}
		v = cloneTimeline(v)
		v.Revision = expected + 1
		r.Value = v
		r.History = append(r.History[:r.Cursor+1], cloneTimeline(v))
		r.Cursor++
		r.UpdatedAt = now
		out = cloneTimeline(v)
		return nil
	})
	return
}
func (s *TimelineCollectionStore) History(ctx context.Context, o, p, id string) (out TimelineHistoryState, e error) {
	if id == DefaultTimelineID {
		return s.base.TimelineHistory(ctx, o, p)
	}
	e = s.access(ctx, o, p, id, func(items *[]timelineCollectionRecord) error {
		i := collectionIndex(*items, id)
		if i < 0 {
			return ErrTimelineNotFound
		}
		r := (*items)[i]
		out = TimelineHistoryState{CanUndo: r.Cursor > 0, CanRedo: r.Cursor+1 < len(r.History)}
		return nil
	})
	return
}
func (s *TimelineCollectionStore) Undo(ctx context.Context, o, p, id string, expected int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	if id == DefaultTimelineID {
		return s.base.UndoTimeline(ctx, o, p, expected, now)
	}
	return s.move(ctx, o, p, id, expected, -1, now)
}
func (s *TimelineCollectionStore) Redo(ctx context.Context, o, p, id string, expected int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	if id == DefaultTimelineID {
		return s.base.RedoTimeline(ctx, o, p, expected, now)
	}
	return s.move(ctx, o, p, id, expected, 1, now)
}
func (s *TimelineCollectionStore) move(ctx context.Context, o, p, id string, expected int64, delta int, now time.Time) (out Timeline, h TimelineHistoryState, e error) {
	e = s.access(ctx, o, p, id, func(items *[]timelineCollectionRecord) error {
		i := collectionIndex(*items, id)
		if i < 0 {
			return ErrTimelineNotFound
		}
		r := &(*items)[i]
		if expected < 0 || r.Value.Revision != expected {
			return ErrRevisionConflict
		}
		next := r.Cursor + delta
		if next < 0 || next >= len(r.History) {
			return ErrTimelineHistoryBoundary
		}
		v := cloneTimeline(r.History[next])
		v.Revision = expected + 1
		r.Value = v
		r.Cursor = next
		r.UpdatedAt = now
		out = cloneTimeline(v)
		h = TimelineHistoryState{CanUndo: next > 0, CanRedo: next+1 < len(r.History)}
		return nil
	})
	return
}

// cloneCollectionTimeline preserves project asset/resource IDs while assigning
// fresh editable-object IDs and translating internal relationships together.
func cloneCollectionTimeline(value Timeline) Timeline {
	v := cloneTimeline(value)
	tracks, clips, groups, links := map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}
	fresh := func(m map[string]string, id string) string {
		if id == "" {
			return ""
		}
		if m[id] == "" {
			m[id] = collectionID()
		}
		return m[id]
	}
	for i := range v.Tracks {
		t := &v.Tracks[i]
		fresh(tracks, t.ID)
		fresh(groups, t.Settings.CompoundID)
		for _, c := range t.Clips {
			fresh(clips, c.ID)
			fresh(groups, c.CompoundID)
			fresh(links, c.LinkGroupID)
			for _, g := range c.CompoundPath {
				fresh(groups, g.ID)
			}
		}
	}
	for i := range v.Tracks {
		t := &v.Tracks[i]
		t.ID = tracks[t.ID]
		t.Settings.CompoundID = fresh(groups, t.Settings.CompoundID)
		for j := range t.Clips {
			c := &t.Clips[j]
			c.ID = clips[c.ID]
			c.CompoundID = fresh(groups, c.CompoundID)
			c.LinkGroupID = fresh(links, c.LinkGroupID)
			for k := range c.CompoundPath {
				c.CompoundPath[k].ID = fresh(groups, c.CompoundPath[k].ID)
			}
			if id := tracks[c.TransitionTrackID]; id != "" {
				c.TransitionTrackID = id
			}
			if id := clips[c.TransitionLeftClipID]; id != "" {
				c.TransitionLeftClipID = id
			}
			if id := clips[c.TransitionRightClipID]; id != "" {
				c.TransitionRightClipID = id
			}
		}
	}
	for i := range v.Markers {
		v.Markers[i].ID = collectionID()
	}
	return v
}
