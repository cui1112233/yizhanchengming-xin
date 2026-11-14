package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type StoryboardPrefixPreset struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
type StoryboardPreferences struct {
	Revision         int64                    `json:"revision"`
	FavoriteIDs      []string                 `json:"favorite_ids"`
	RecentIDs        []string                 `json:"recent_ids"`
	PrefixPresets    []StoryboardPrefixPreset `json:"prefix_presets"`
	SelectedPrefixID string                   `json:"selected_prefix_id"`
}
type StoryboardPreferencesStore interface {
	LoadStoryboardPreferences(context.Context, string) (StoryboardPreferences, error)
	MutateStoryboardPreferences(context.Context, string, func(*StoryboardPreferences) error) (StoryboardPreferences, error)
}

func initStoryboardPreferences(v *StoryboardPreferences) {
	kept := []StoryboardPrefixPreset{}
	for _, p := range v.PrefixPresets {
		if p.ID != "sb_preset_default" {
			kept = append(kept, p)
		}
	}
	v.PrefixPresets = kept
	if v.SelectedPrefixID == "sb_preset_default" {
		v.SelectedPrefixID = ""
	}
	if v.FavoriteIDs == nil {
		v.FavoriteIDs = []string{}
	}
	if v.RecentIDs == nil {
		v.RecentIDs = []string{}
	}
	if v.PrefixPresets == nil {
		v.PrefixPresets = []StoryboardPrefixPreset{}
	}
}
func cloneStoryboardPreferences(v StoryboardPreferences) StoryboardPreferences {
	b, _ := json.Marshal(v)
	var out StoryboardPreferences
	_ = json.Unmarshal(b, &out)
	initStoryboardPreferences(&out)
	return out
}
func (m *MemoryStore) LoadStoryboardPreferences(_ context.Context, owner string) (StoryboardPreferences, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneStoryboardPreferences(m.storyboardPreferences[owner]), nil
}
func (m *MemoryStore) MutateStoryboardPreferences(ctx context.Context, owner string, fn func(*StoryboardPreferences) error) (StoryboardPreferences, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ctx.Err() != nil {
		return StoryboardPreferences{}, ctx.Err()
	}
	v := cloneStoryboardPreferences(m.storyboardPreferences[owner])
	if e := fn(&v); e != nil {
		return v, e
	}
	v.Revision++
	if m.storyboardPreferences == nil {
		m.storyboardPreferences = map[string]StoryboardPreferences{}
	}
	m.storyboardPreferences[owner] = cloneStoryboardPreferences(v)
	return v, nil
}
func (s *MySQLStore) LoadStoryboardPreferences(ctx context.Context, owner string) (StoryboardPreferences, error) {
	var raw []byte
	var v StoryboardPreferences
	e := s.db.QueryRowContext(ctx, "SELECT state_json FROM screenplay_storyboard_preferences WHERE owner_username=?", owner).Scan(&raw)
	if errors.Is(e, sql.ErrNoRows) {
		initStoryboardPreferences(&v)
		return v, nil
	}
	if e != nil {
		return v, e
	}
	e = json.Unmarshal(raw, &v)
	initStoryboardPreferences(&v)
	return v, e
}
func (s *MySQLStore) MutateStoryboardPreferences(ctx context.Context, owner string, fn func(*StoryboardPreferences) error) (StoryboardPreferences, error) {
	var v StoryboardPreferences
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return v, e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(ctx, "INSERT IGNORE INTO screenplay_storyboard_preferences(owner_username,state_json) VALUES (?,'{}')", owner); e != nil {
		return v, e
	}
	var raw []byte
	if e = tx.QueryRowContext(ctx, "SELECT state_json FROM screenplay_storyboard_preferences WHERE owner_username=? FOR UPDATE", owner).Scan(&raw); e != nil {
		return v, e
	}
	if e = json.Unmarshal(raw, &v); e != nil {
		return v, e
	}
	initStoryboardPreferences(&v)
	if e = fn(&v); e != nil {
		return v, e
	}
	v.Revision++
	raw, e = json.Marshal(v)
	if e != nil {
		return v, e
	}
	if _, e = tx.ExecContext(ctx, "UPDATE screenplay_storyboard_preferences SET state_json=? WHERE owner_username=?", string(raw), owner); e != nil {
		return v, e
	}
	return v, tx.Commit()
}

type StoryboardPreferencesAction struct {
	ExpectedRevision *int64                   `json:"expected_revision"`
	Action           string                   `json:"action"`
	PromptID         string                   `json:"prompt_id"`
	PresetID         string                   `json:"preset_id"`
	Preset           StoryboardPrefixPreset   `json:"preset"`
	Presets          []StoryboardPrefixPreset `json:"presets"`
}

func ValidateStoryboardPrefix(v *StoryboardPrefixPreset) error {
	v.Name = strings.TrimSpace(v.Name)
	v.Content = strings.TrimSpace(v.Content)
	if v.ID == "" {
		v.ID = newScreenplayID("sb_preset")
	}
	if len(v.ID) > 96 || strings.ContainsAny(v.ID, "/\\?#") || v.Name == "" || len([]rune(v.Name)) > 255 || v.Content == "" || len([]rune(v.Content)) > 5000 {
		return fmt.Errorf("%w: 前缀名称不能为空且最多255字，内容须1到5000字，ID须有效", ErrInvalidArgument)
	}
	return nil
}
func ApplyStoryboardPreferencesAction(v *StoryboardPreferences, in StoryboardPreferencesAction) (imported, skipped int, err error) {
	if in.ExpectedRevision == nil || *in.ExpectedRevision != v.Revision {
		return 0, 0, ErrRevisionConflict
	}
	initStoryboardPreferences(v)
	strip := func(ids []string, id string) []string {
		out := []string{}
		for _, x := range ids {
			if x != id {
				out = append(out, x)
			}
		}
		return out
	}
	switch in.Action {
	case "favorite_toggle", "recent":
		id := CreativePromptAlias(in.PromptID)
		if _, e := ResolveProjectPrompt(Project{}, id, "storyboard"); e != nil {
			return 0, 0, fmt.Errorf("%w: %v", ErrInvalidArgument, e)
		}
		if in.Action == "favorite_toggle" {
			found := false
			for _, old := range v.FavoriteIDs {
				if old == id {
					found = true
				}
			}
			v.FavoriteIDs = strip(v.FavoriteIDs, id)
			if !found {
				v.FavoriteIDs = append([]string{id}, v.FavoriteIDs...)
			}
		} else {
			v.RecentIDs = append([]string{id}, strip(v.RecentIDs, id)...)
		}
		favorites := map[string]bool{}
		for _, id := range v.FavoriteIDs {
			favorites[id] = true
		}
		seen := map[string]bool{}
		recent := []string{}
		for _, id := range v.RecentIDs {
			if !favorites[id] && !seen[id] && len(recent) < 8 {
				recent = append(recent, id)
				seen[id] = true
			}
		}
		v.RecentIDs = recent
	case "prefix_save":
		p := in.Preset
		if p.ID == "sb_preset_default" {
			return 0, 0, fmt.Errorf("%w: 此ID属于已移除的旧默认前缀", ErrInvalidArgument)
		}
		if e := ValidateStoryboardPrefix(&p); e != nil {
			return 0, 0, e
		}
		found := false
		for i := range v.PrefixPresets {
			if v.PrefixPresets[i].ID == p.ID {
				v.PrefixPresets[i] = p
				found = true
				break
			}
		}
		if !found {
			v.PrefixPresets = append(v.PrefixPresets, p)
		}
	case "prefix_delete":
		out := []StoryboardPrefixPreset{}
		found := false
		for _, p := range v.PrefixPresets {
			if p.ID == in.PresetID {
				found = true
			} else {
				out = append(out, p)
			}
		}
		if !found {
			return 0, 0, ErrNotFound
		}
		v.PrefixPresets = out
		if v.SelectedPrefixID == in.PresetID {
			v.SelectedPrefixID = ""
		}
	case "prefix_select":
		found := in.PresetID == ""
		for _, p := range v.PrefixPresets {
			if p.ID == in.PresetID {
				found = true
				break
			}
		}
		if !found {
			return 0, 0, ErrNotFound
		}
		if v.SelectedPrefixID == in.PresetID {
			v.SelectedPrefixID = ""
		} else {
			v.SelectedPrefixID = in.PresetID
		}
	case "prefix_import":
		if in.Presets == nil {
			return 0, 0, fmt.Errorf("%w: 导入须为presets数组", ErrInvalidArgument)
		}
		seen := map[string]bool{}
		for _, p := range v.PrefixPresets {
			seen[p.ID] = true
		}
		for _, p := range in.Presets {
			if p.ID == "sb_preset_default" {
				skipped++
				continue
			}
			if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Content) == "" {
				skipped++
				continue
			}
			if e := ValidateStoryboardPrefix(&p); e != nil {
				return 0, 0, e
			}
			if seen[p.ID] {
				skipped++
				continue
			}
			seen[p.ID] = true
			v.PrefixPresets = append(v.PrefixPresets, p)
			imported++
		}
	default:
		return 0, 0, fmt.Errorf("%w: 未知分镜偏好操作", ErrInvalidArgument)
	}
	if len(v.PrefixPresets) > 1000 || len(v.FavoriteIDs) > 1000 {
		return 0, 0, fmt.Errorf("%w: 收藏或前缀最多1000条", ErrInvalidArgument)
	}
	return imported, skipped, nil
}
func (v StoryboardPreferences) SelectedPrefix() string {
	for _, p := range v.PrefixPresets {
		if p.ID == v.SelectedPrefixID {
			return p.Content
		}
	}
	return ""
}
