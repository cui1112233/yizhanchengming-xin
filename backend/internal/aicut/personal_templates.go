package aicut

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrInvalidPersonalTemplate = errors.New("invalid personal template")
var ErrPersonalTemplateNotFound = errors.New("personal template not found")

type PersonalTemplateSlot struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
}
type PersonalTemplateLayout struct {
	Version int                    `json:"version"`
	Slots   []PersonalTemplateSlot `json:"slots"`
	Tracks  []Track                `json:"tracks"`
}
type PersonalTemplate struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Layout    PersonalTemplateLayout `json:"layout"`
	CreatedAt time.Time              `json:"createdAt"`
}

// Both implementations keep ownership in their storage key; callers never supply an owner in the record.
type PersonalTemplateStore struct {
	mu    sync.Mutex
	db    *sql.DB
	items map[string]PersonalTemplate
}

func NewMemoryPersonalTemplateStore() *PersonalTemplateStore {
	return &PersonalTemplateStore{items: map[string]PersonalTemplate{}}
}
func NewMySQLPersonalTemplateStore(db *sql.DB) *PersonalTemplateStore {
	return &PersonalTemplateStore{db: db}
}
func DecodePersonalTemplateLayout(raw []byte) (PersonalTemplateLayout, error) {
	var l PersonalTemplateLayout
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&l); e != nil {
		return l, ErrInvalidPersonalTemplate
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return l, ErrInvalidPersonalTemplate
	}
	return l, ValidatePersonalTemplateLayout(l)
}
func ValidatePersonalTemplateLayout(l PersonalTemplateLayout) error {
	if l.Version != 1 || len(l.Tracks) == 0 || len(l.Tracks) > 100 || len(l.Slots) > 500 {
		return ErrInvalidPersonalTemplate
	}
	slots := map[string]string{}
	used := map[string]bool{}
	count := 0
	for i, s := range l.Slots {
		if s.ID != fmt.Sprintf("slot-%d", i+1) || (s.Kind != "video" && s.Kind != "image" && s.Kind != "audio") || s.Label != fmt.Sprintf("素材 %d", i+1) {
			return ErrInvalidPersonalTemplate
		}
		slots[s.ID] = s.Kind
	}
	for i, t := range l.Tracks {
		if t.ID != fmt.Sprintf("track-%d", i+1) || t.Order != i || t.Settings.FontAssetID != "" || t.Settings.CompoundID != "" || t.Settings.CompoundName != "" || len(t.Clips) == 0 || (t.Kind != TrackVideo && t.Kind != TrackAudio && t.Kind != TrackSubtitle) {
			return ErrInvalidPersonalTemplate
		}
		for _, c := range t.Clips {
			if len(c.SpeedCurve) > 0 {
				return ErrInvalidPersonalTemplate
			}
			count++
			if count > 500 || c.ID != fmt.Sprintf("clip-%d", count) || c.StartMS < 0 || c.StartMS+c.DurationMS > 86400000 || c.LinkGroupID != "" || c.CompoundID != "" || c.CompoundName != "" || len(c.CompoundPath) > 0 || c.Transition != "" || c.TransitionMS != 0 || c.TransitionTrackID != "" || c.TransitionLeftClipID != "" || c.TransitionRightClipID != "" {
				return ErrInvalidPersonalTemplate
			}
			if c.Kind == ClipSubtitle {
				if c.AssetID != "" {
					return ErrInvalidPersonalTemplate
				}
			} else if (c.Kind != ClipVideo && c.Kind != ClipImage && c.Kind != ClipAudio) || slots[c.AssetID] != string(c.Kind) {
				return ErrInvalidPersonalTemplate
			} else {
				used[c.AssetID] = true
			}
			if p := c.VisualProcessing; p != nil && (p.PersonSegmentation.Enabled || p.PersonSegmentation.MatteAssetID != "" || p.PersonSegmentation.Signature != "") {
				return ErrInvalidPersonalTemplate
			}
		}
	}
	if count == 0 || len(used) != len(slots) || validateTimeline(Timeline{Tracks: l.Tracks}) != nil {
		return ErrInvalidPersonalTemplate
	}
	return nil
}
func clonePersonalTemplate(t PersonalTemplate) PersonalTemplate {
	b, _ := json.Marshal(t)
	var v PersonalTemplate
	_ = json.Unmarshal(b, &v)
	return v
}
func (s *PersonalTemplateStore) Create(ctx context.Context, owner, title string, l PersonalTemplateLayout) (PersonalTemplate, error) {
	title = strings.TrimSpace(title)
	if strings.TrimSpace(owner) == "" || title == "" || len([]rune(title)) > 80 || ValidatePersonalTemplateLayout(l) != nil {
		return PersonalTemplate{}, ErrInvalidPersonalTemplate
	}
	v := PersonalTemplate{ID: collectionID(), Title: title, Layout: l, CreatedAt: time.Now().UTC()}
	raw, e := json.Marshal(l)
	if e != nil || len(raw) > 2<<20 {
		return PersonalTemplate{}, ErrInvalidPersonalTemplate
	}
	if s.db != nil {
		_, e = s.db.ExecContext(ctx, `INSERT INTO aicut_personal_templates (id,owner_username,title,layout_json,created_at) VALUES (?,?,?,?,?)`, v.ID, owner, v.Title, string(raw), v.CreatedAt)
		return v, e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[owner+"\x00"+v.ID] = clonePersonalTemplate(v)
	return clonePersonalTemplate(v), nil
}
func (s *PersonalTemplateStore) List(ctx context.Context, owner string) ([]PersonalTemplate, error) {
	out := []PersonalTemplate{}
	if s.db != nil {
		rows, e := s.db.QueryContext(ctx, `SELECT id,title,layout_json,created_at FROM aicut_personal_templates WHERE owner_username=? ORDER BY created_at DESC,id`, owner)
		if e != nil {
			return nil, e
		}
		defer rows.Close()
		for rows.Next() {
			var v PersonalTemplate
			var raw []byte
			if e = rows.Scan(&v.ID, &v.Title, &raw, &v.CreatedAt); e != nil {
				return nil, e
			}
			if v.Layout, e = DecodePersonalTemplateLayout(raw); e != nil {
				return nil, e
			}
			out = append(out, v)
		}
		return out, rows.Err()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.items {
		if strings.HasPrefix(k, owner+"\x00") {
			out = append(out, clonePersonalTemplate(v))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (s *PersonalTemplateStore) Delete(ctx context.Context, owner, id string) error {
	if s.db != nil {
		r, e := s.db.ExecContext(ctx, `DELETE FROM aicut_personal_templates WHERE owner_username=? AND id=?`, owner, id)
		if e != nil {
			return e
		}
		n, e := r.RowsAffected()
		if e != nil {
			return e
		}
		if n == 0 {
			return ErrPersonalTemplateNotFound
		}
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	k := owner + "\x00" + id
	if _, ok := s.items[k]; !ok {
		return ErrPersonalTemplateNotFound
	}
	delete(s.items, k)
	return nil
}
