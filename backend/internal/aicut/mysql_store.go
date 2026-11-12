package aicut

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type clipEditorData struct {
	CompoundID            string             `json:"compoundId,omitempty"`
	CompoundName          string             `json:"compoundName,omitempty"`
	CompoundPath          []CompoundRef      `json:"compoundPath,omitempty"`
	CatalogResourceID     string             `json:"catalogResourceId,omitempty"`
	TextResourceID        string             `json:"textResourceId,omitempty"`
	ResourceID            string             `json:"resourceId,omitempty"`
	TransitionTrackID     string             `json:"transitionTrackId,omitempty"`
	TransitionLeftClipID  string             `json:"transitionLeftClipId,omitempty"`
	TransitionRightClipID string             `json:"transitionRightClipId,omitempty"`
	Parameters            map[string]float64 `json:"parameters,omitempty"`
	CanvasTransform       *CanvasTransform   `json:"canvasTransform,omitempty"`
	Keyframes             []ClipKeyframe     `json:"keyframes,omitempty"`
	SpeedCurve            []SpeedCurvePoint  `json:"speedCurve,omitempty"`
	AudioProcessing       *AudioProcessing   `json:"audioProcessing,omitempty"`
	VisualProcessing      *VisualProcessing  `json:"visualProcessing,omitempty"`
	Reversed              bool               `json:"reversed,omitempty"`
	FreezeSourceMS        *int64             `json:"freezeSourceMs,omitempty"`
	BlendMode             string             `json:"blendMode,omitempty"`
}

func encodeClipEditorJSON(clip Clip) ([]byte, error) {
	return json.Marshal(clipEditorData{
		CompoundID:            clip.CompoundID,
		CompoundName:          clip.CompoundName,
		CompoundPath:          clip.CompoundPath,
		CatalogResourceID:     clip.CatalogResourceID,
		TextResourceID:        clip.TextResourceID,
		ResourceID:            clip.ResourceID,
		TransitionTrackID:     clip.TransitionTrackID,
		TransitionLeftClipID:  clip.TransitionLeftClipID,
		TransitionRightClipID: clip.TransitionRightClipID,
		Parameters:            clip.Parameters,
		CanvasTransform:       clip.CanvasTransform,
		Keyframes:             clip.Keyframes,
		SpeedCurve:            clip.SpeedCurve,
		AudioProcessing:       clip.AudioProcessing,
		VisualProcessing:      clip.VisualProcessing,
		Reversed:              clip.Reversed,
		FreezeSourceMS:        clip.FreezeSourceMS,
		BlendMode:             clip.BlendMode,
	})
}

func decodeClipEditorJSON(data []byte, clip *Clip) error {
	var editor clipEditorData
	if err := json.Unmarshal(data, &editor); err != nil {
		return err
	}
	clip.CatalogResourceID = editor.CatalogResourceID
	clip.CompoundID = editor.CompoundID
	clip.CompoundName = editor.CompoundName
	clip.CompoundPath = editor.CompoundPath
	clip.TextResourceID = editor.TextResourceID
	clip.ResourceID = editor.ResourceID
	clip.TransitionTrackID = editor.TransitionTrackID
	clip.TransitionLeftClipID = editor.TransitionLeftClipID
	clip.TransitionRightClipID = editor.TransitionRightClipID
	clip.Parameters = editor.Parameters
	clip.CanvasTransform = editor.CanvasTransform
	clip.Keyframes = editor.Keyframes
	clip.SpeedCurve = editor.SpeedCurve
	clip.AudioProcessing = editor.AudioProcessing
	clip.VisualProcessing = editor.VisualProcessing
	clip.Reversed = editor.Reversed
	clip.FreezeSourceMS = editor.FreezeSourceMS
	clip.BlendMode = editor.BlendMode
	return nil
}

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) ListProjects(ctx context.Context, owner string) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, aspect_ratio, background_mode, background_color, created_at, updated_at FROM aicut_projects WHERE owner_username = ? AND deleted_at IS NULL ORDER BY updated_at DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Title, &project.AspectRatio, &project.BackgroundMode, &project.BackgroundColor, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (s *MySQLStore) CreateProject(ctx context.Context, owner string, project Project) error {
	project, err := normalizeProject(project)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO aicut_projects (id, owner_username, title, aspect_ratio, background_mode, background_color, revision, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?)`, project.ID, owner, project.Title, project.AspectRatio, project.BackgroundMode, project.BackgroundColor, project.CreatedAt, project.UpdatedAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrProjectExists
	}
	if err != nil {
		return err
	}
	emptyJSON, _ := json.Marshal(Timeline{Tracks: []Track{}})
	if _, err = tx.ExecContext(ctx, `INSERT INTO aicut_timeline_history (project_id, owner_username, sequence_no, timeline_json, created_at) VALUES (?, ?, 0, ?, ?)`, project.ID, owner, string(emptyJSON), project.CreatedAt); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO aicut_timeline_history_state (project_id, owner_username, cursor_no, max_no, updated_at) VALUES (?, ?, 0, 0, ?)`, project.ID, owner, project.CreatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) Project(ctx context.Context, owner, id string) (Project, error) {
	var project Project
	err := s.db.QueryRowContext(ctx, `SELECT id, title, aspect_ratio, background_mode, background_color, created_at, updated_at FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, id, owner).Scan(&project.ID, &project.Title, &project.AspectRatio, &project.BackgroundMode, &project.BackgroundColor, &project.CreatedAt, &project.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	return project, err
}

func (s *MySQLStore) RenameProject(ctx context.Context, owner, id, title string, now time.Time) (Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Project{}, ErrInvalidProject
	}
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_projects SET title = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, title, now, id, owner)
	if err != nil {
		return Project{}, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return Project{}, ErrProjectNotFound
	}
	return s.Project(ctx, owner, id)
}

func (s *MySQLStore) UpdateProjectAppearance(ctx context.Context, owner, id, mode, color string, now time.Time) (Project, error) {
	candidate, err := normalizeProject(Project{ID: id, Title: "appearance", AspectRatio: "9:16", BackgroundMode: mode, BackgroundColor: color})
	if err != nil {
		return Project{}, err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_projects SET background_mode = ?, background_color = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, candidate.BackgroundMode, candidate.BackgroundColor, now, id, owner)
	if err != nil {
		return Project{}, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return Project{}, ErrProjectNotFound
	}
	return s.Project(ctx, owner, id)
}

func (s *MySQLStore) DeleteProject(ctx context.Context, owner, id string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_projects SET deleted_at = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now, now, id, owner)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (s *MySQLStore) Timeline(ctx context.Context, owner, id string) (Timeline, error) {
	var timeline Timeline
	var markers []byte
	if err := s.db.QueryRowContext(ctx, `SELECT revision, COALESCE(script_text, ''), COALESCE(markers_json, JSON_ARRAY()) FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, id, owner).Scan(&timeline.Revision, &timeline.Script, &markers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Timeline{}, ErrProjectNotFound
		}
		return Timeline{}, err
	}
	if len(markers) > 0 {
		if err := json.Unmarshal(markers, &timeline.Markers); err != nil {
			return Timeline{}, err
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, kind, name, sort_index, muted, locked, COALESCE(settings_json, JSON_OBJECT()) FROM aicut_tracks WHERE project_id = ? AND owner_username = ? ORDER BY sort_index, id`, id, owner)
	if err != nil {
		return Timeline{}, err
	}
	defer rows.Close()
	timeline.Tracks = make([]Track, 0)
	for rows.Next() {
		var track Track
		var settings []byte
		if err := rows.Scan(&track.ID, &track.Kind, &track.Name, &track.Order, &track.Muted, &track.Locked, &settings); err != nil {
			return Timeline{}, err
		}
		if len(settings) > 0 && string(settings) != "null" {
			if err := json.Unmarshal(settings, &track.Settings); err != nil {
				return Timeline{}, err
			}
		}
		track.Clips, err = s.trackClips(ctx, owner, id, track.ID)
		if err != nil {
			return Timeline{}, err
		}
		timeline.Tracks = append(timeline.Tracks, track)
	}
	return timeline, rows.Err()
}

func (s *MySQLStore) trackClips(ctx context.Context, owner, projectID, trackID string) ([]Clip, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, COALESCE(asset_id, ''), kind, name, start_ms, duration_ms, source_start_ms, source_duration_ms, COALESCE(text_content, ''), transition_type, transition_ms, review_status, COALESCE(review_reason, ''), link_group_id, playback_rate, volume, fade_in_ms, fade_out_ms, COALESCE(visual_effect_json, 'null'), COALESCE(editor_json, JSON_OBJECT()) FROM aicut_clips WHERE track_id = ? AND project_id = ? AND owner_username = ? ORDER BY start_ms, id`, trackID, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	clips := make([]Clip, 0)
	for rows.Next() {
		var clip Clip
		var visualEffect []byte
		var editorJSON []byte
		if err := rows.Scan(&clip.ID, &clip.AssetID, &clip.Kind, &clip.Name, &clip.StartMS, &clip.DurationMS, &clip.SourceStartMS, &clip.SourceDurationMS, &clip.Text, &clip.Transition, &clip.TransitionMS, &clip.ReviewStatus, &clip.ReviewReason, &clip.LinkGroupID, &clip.PlaybackRate, &clip.Volume, &clip.FadeInMS, &clip.FadeOutMS, &visualEffect, &editorJSON); err != nil {
			return nil, err
		}
		if len(editorJSON) > 0 {
			if err := decodeClipEditorJSON(editorJSON, &clip); err != nil {
				return nil, err
			}
		}
		if string(visualEffect) != "null" && len(visualEffect) > 0 {
			if err := json.Unmarshal(visualEffect, &clip.VisualEffect); err != nil {
				return nil, err
			}
		}
		clips = append(clips, clip)
	}
	return clips, rows.Err()
}

func (s *MySQLStore) SaveTimeline(ctx context.Context, owner, id string, timeline Timeline, expectedRevision int64, now time.Time) (Timeline, error) {
	if err := validateTimeline(timeline); err != nil {
		return Timeline{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Timeline{}, err
	}
	defer tx.Rollback()
	var currentRevision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, id, owner).Scan(&currentRevision); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Timeline{}, ErrProjectNotFound
		}
		return Timeline{}, err
	}
	if currentRevision != expectedRevision {
		return Timeline{}, ErrRevisionConflict
	}
	cursor, _, err := ensureTimelineHistoryTx(ctx, tx, owner, id, currentRevision, now)
	if err != nil {
		return Timeline{}, err
	}
	timeline.Revision = expectedRevision + 1
	if _, err := tx.ExecContext(ctx, `DELETE FROM aicut_timeline_history WHERE project_id = ? AND owner_username = ? AND sequence_no > ?`, id, owner, cursor); err != nil {
		return Timeline{}, err
	}
	next := cursor + 1
	encoded, _ := json.Marshal(timeline)
	if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_timeline_history (project_id, owner_username, sequence_no, timeline_json, created_at) VALUES (?, ?, ?, ?, ?)`, id, owner, next, string(encoded), now); err != nil {
		return Timeline{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE aicut_timeline_history_state SET cursor_no = ?, max_no = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`, next, next, now, id, owner); err != nil {
		return Timeline{}, err
	}
	if err := replaceTimelineTx(ctx, tx, owner, id, timeline, now); err != nil {
		return Timeline{}, err
	}
	if err := tx.Commit(); err != nil {
		return Timeline{}, err
	}
	return timeline, nil
}

func replaceTimelineTx(ctx context.Context, tx *sql.Tx, owner, id string, timeline Timeline, now time.Time) error {
	markers, err := json.Marshal(timeline.Markers)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE aicut_projects SET revision = ?, script_text = ?, markers_json = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, timeline.Revision, timeline.Script, string(markers), now, id, owner); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM aicut_tracks WHERE project_id = ? AND owner_username = ?`, id, owner); err != nil {
		return err
	}
	for _, track := range timeline.Tracks {
		settings, err := json.Marshal(track.Settings)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_tracks (id, project_id, owner_username, kind, name, sort_index, muted, locked, settings_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, track.ID, id, owner, track.Kind, track.Name, track.Order, track.Muted, track.Locked, string(settings), now, now); err != nil {
			return err
		}
		for _, clip := range track.Clips {
			var assetID any
			if clip.AssetID != "" {
				assetID = clip.AssetID
			}
			playbackRate := clip.PlaybackRate
			if playbackRate == 0 {
				playbackRate = 1
			}
			volume := clip.Volume
			if volume == 0 {
				volume = 1
			}
			var visualEffect any
			if clip.VisualEffect != nil {
				encoded, err := json.Marshal(clip.VisualEffect)
				if err != nil {
					return err
				}
				visualEffect = string(encoded)
			}
			editorJSON, err := encodeClipEditorJSON(clip)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_clips (id, track_id, project_id, owner_username, asset_id, kind, name, start_ms, duration_ms, source_start_ms, source_duration_ms, text_content, transition_type, transition_ms, review_status, review_reason, link_group_id, playback_rate, volume, fade_in_ms, fade_out_ms, visual_effect_json, editor_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, clip.ID, track.ID, id, owner, assetID, clip.Kind, clip.Name, clip.StartMS, clip.DurationMS, clip.SourceStartMS, clip.SourceDurationMS, clip.Text, clip.Transition, clip.TransitionMS, clip.ReviewStatus, clip.ReviewReason, clip.LinkGroupID, playbackRate, volume, clip.FadeInMS, clip.FadeOutMS, visualEffect, string(editorJSON), now, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MySQLStore) TimelineHistory(ctx context.Context, owner, id string) (TimelineHistoryState, error) {
	var cursor, max int64
	err := s.db.QueryRowContext(ctx, `SELECT cursor_no, max_no FROM aicut_timeline_history_state WHERE project_id = ? AND owner_username = ?`, id, owner).Scan(&cursor, &max)
	if errors.Is(err, sql.ErrNoRows) {
		var exists int
		if projectErr := s.db.QueryRowContext(ctx, `SELECT 1 FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, id, owner).Scan(&exists); errors.Is(projectErr, sql.ErrNoRows) {
			return TimelineHistoryState{}, ErrProjectNotFound
		} else if projectErr != nil {
			return TimelineHistoryState{}, projectErr
		}
		return TimelineHistoryState{}, nil
	}
	if err != nil {
		return TimelineHistoryState{}, err
	}
	return TimelineHistoryState{CanUndo: cursor > 0, CanRedo: cursor < max}, nil
}

func ensureTimelineHistoryTx(ctx context.Context, tx *sql.Tx, owner, id string, revision int64, now time.Time) (int64, int64, error) {
	var cursor, max int64
	err := tx.QueryRowContext(ctx, `SELECT cursor_no, max_no FROM aicut_timeline_history_state WHERE project_id = ? AND owner_username = ? FOR UPDATE`, id, owner).Scan(&cursor, &max)
	if err == nil {
		return cursor, max, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, err
	}
	current, err := timelineFromTx(ctx, tx, owner, id)
	if err != nil {
		return 0, 0, err
	}
	if current.Revision != revision {
		return 0, 0, ErrRevisionConflict
	}
	encoded, _ := json.Marshal(current)
	if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_timeline_history (project_id, owner_username, sequence_no, timeline_json, created_at) VALUES (?, ?, 0, ?, ?)`, id, owner, string(encoded), now); err != nil {
		return 0, 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_timeline_history_state (project_id, owner_username, cursor_no, max_no, updated_at) VALUES (?, ?, 0, 0, ?)`, id, owner, now); err != nil {
		return 0, 0, err
	}
	return 0, 0, nil
}

func timelineFromTx(ctx context.Context, tx *sql.Tx, owner, id string) (Timeline, error) {
	var timeline Timeline
	var markers []byte
	if err := tx.QueryRowContext(ctx, `SELECT revision, COALESCE(script_text, ''), COALESCE(markers_json, JSON_ARRAY()) FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, id, owner).Scan(&timeline.Revision, &timeline.Script, &markers); err != nil {
		return Timeline{}, err
	}
	if len(markers) > 0 {
		if err := json.Unmarshal(markers, &timeline.Markers); err != nil {
			return Timeline{}, err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT id, kind, name, sort_index, muted, locked, COALESCE(settings_json, JSON_OBJECT()) FROM aicut_tracks WHERE project_id = ? AND owner_username = ? ORDER BY sort_index, id`, id, owner)
	if err != nil {
		return Timeline{}, err
	}
	defer rows.Close()
	tracks := []Track{}
	for rows.Next() {
		var track Track
		var settings []byte
		if err := rows.Scan(&track.ID, &track.Kind, &track.Name, &track.Order, &track.Muted, &track.Locked, &settings); err != nil {
			return Timeline{}, err
		}
		if len(settings) > 0 && string(settings) != "null" {
			if err := json.Unmarshal(settings, &track.Settings); err != nil {
				return Timeline{}, err
			}
		}
		tracks = append(tracks, track)
	}
	if err := rows.Err(); err != nil {
		return Timeline{}, err
	}
	if err := rows.Close(); err != nil {
		return Timeline{}, err
	}
	for index := range tracks {
		tracks[index].Clips, err = trackClipsFromTx(ctx, tx, owner, id, tracks[index].ID)
		if err != nil {
			return Timeline{}, err
		}
	}
	timeline.Tracks = tracks
	return timeline, nil
}

func trackClipsFromTx(ctx context.Context, tx *sql.Tx, owner, projectID, trackID string) ([]Clip, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, COALESCE(asset_id, ''), kind, name, start_ms, duration_ms, source_start_ms, source_duration_ms, COALESCE(text_content, ''), transition_type, transition_ms, review_status, COALESCE(review_reason, ''), link_group_id, playback_rate, volume, fade_in_ms, fade_out_ms, COALESCE(visual_effect_json, 'null'), COALESCE(editor_json, JSON_OBJECT()) FROM aicut_clips WHERE track_id = ? AND project_id = ? AND owner_username = ? ORDER BY start_ms, id`, trackID, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	clips := []Clip{}
	for rows.Next() {
		var clip Clip
		var visualEffect []byte
		var editorJSON []byte
		if err := rows.Scan(&clip.ID, &clip.AssetID, &clip.Kind, &clip.Name, &clip.StartMS, &clip.DurationMS, &clip.SourceStartMS, &clip.SourceDurationMS, &clip.Text, &clip.Transition, &clip.TransitionMS, &clip.ReviewStatus, &clip.ReviewReason, &clip.LinkGroupID, &clip.PlaybackRate, &clip.Volume, &clip.FadeInMS, &clip.FadeOutMS, &visualEffect, &editorJSON); err != nil {
			return nil, err
		}
		if len(editorJSON) > 0 {
			if err := decodeClipEditorJSON(editorJSON, &clip); err != nil {
				return nil, err
			}
		}
		if string(visualEffect) != "null" && len(visualEffect) > 0 {
			if err := json.Unmarshal(visualEffect, &clip.VisualEffect); err != nil {
				return nil, err
			}
		}
		clips = append(clips, clip)
	}
	return clips, rows.Err()
}

func (s *MySQLStore) moveTimelineHistory(ctx context.Context, owner, id string, expectedRevision int64, delta int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	defer tx.Rollback()
	var currentRevision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM aicut_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, id, owner).Scan(&currentRevision); errors.Is(err, sql.ErrNoRows) {
		return Timeline{}, TimelineHistoryState{}, ErrProjectNotFound
	} else if err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	if currentRevision != expectedRevision {
		return Timeline{}, TimelineHistoryState{}, ErrRevisionConflict
	}
	cursor, max, err := ensureTimelineHistoryTx(ctx, tx, owner, id, currentRevision, now)
	if err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	target := cursor + delta
	if target < 0 || target > max {
		return Timeline{}, TimelineHistoryState{CanUndo: cursor > 0, CanRedo: cursor < max}, ErrTimelineHistoryBoundary
	}
	var encoded []byte
	if err := tx.QueryRowContext(ctx, `SELECT timeline_json FROM aicut_timeline_history WHERE project_id = ? AND owner_username = ? AND sequence_no = ?`, id, owner, target).Scan(&encoded); err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	var timeline Timeline
	if err := json.Unmarshal(encoded, &timeline); err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	timeline.Revision = currentRevision + 1
	if err := validateTimeline(timeline); err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	if err := replaceTimelineTx(ctx, tx, owner, id, timeline, now); err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE aicut_timeline_history_state SET cursor_no = ?, updated_at = ? WHERE project_id = ? AND owner_username = ?`, target, now, id, owner); err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	if err := tx.Commit(); err != nil {
		return Timeline{}, TimelineHistoryState{}, err
	}
	return timeline, TimelineHistoryState{CanUndo: target > 0, CanRedo: target < max}, nil
}

func (s *MySQLStore) UndoTimeline(ctx context.Context, owner, id string, expectedRevision int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	return s.moveTimelineHistory(ctx, owner, id, expectedRevision, -1, now)
}

func (s *MySQLStore) RedoTimeline(ctx context.Context, owner, id string, expectedRevision int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	return s.moveTimelineHistory(ctx, owner, id, expectedRevision, 1, now)
}

func (s *MySQLStore) ListAssets(ctx context.Context, owner, projectID string) ([]Asset, error) {
	if _, err := s.Project(ctx, owner, projectID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, kind, name, content_type, object_key, byte_size, duration_ms, created_at FROM aicut_assets WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		asset.ProjectID = projectID
		if err := rows.Scan(&asset.ID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.DurationMS, &asset.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, asset)
	}
	return items, rows.Err()
}

func (s *MySQLStore) CreateAsset(ctx context.Context, owner string, asset Asset) error {
	if err := validateAsset(asset); err != nil {
		return err
	}
	_, err := s.Project(ctx, owner, asset.ProjectID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO aicut_assets (id, project_id, owner_username, kind, name, content_type, object_key, byte_size, duration_ms, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, asset.ID, asset.ProjectID, owner, asset.Kind, asset.Name, asset.ContentType, asset.ObjectKey, asset.ByteSize, asset.DurationMS, asset.CreatedAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrAssetExists
	}
	return err
}

func (s *MySQLStore) Asset(ctx context.Context, owner, projectID, id string) (Asset, error) {
	var asset Asset
	asset.ProjectID = projectID
	err := s.db.QueryRowContext(ctx, `SELECT id, kind, name, content_type, object_key, byte_size, duration_ms, created_at FROM aicut_assets WHERE id = ? AND project_id = ? AND owner_username = ?`, id, projectID, owner).Scan(&asset.ID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.DurationMS, &asset.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Asset{}, ErrAssetNotFound
	}
	return asset, err
}

const asrJobColumns = `id, owner_username, project_id, asset_id, model_id, language, expected_revision, state, progress, COALESCE(error_message, ''), imported_count, track_id, created_at, updated_at, timeline_id`

type rowScanner interface{ Scan(...any) error }

func scanASRJob(row rowScanner) (ASRJob, error) {
	var job ASRJob
	err := row.Scan(&job.ID, &job.Owner, &job.ProjectID, &job.AssetID, &job.ModelID, &job.Language, &job.ExpectedRevision, &job.State, &job.Progress, &job.Error, &job.Imported, &job.TrackID, &job.CreatedAt, &job.UpdatedAt, &job.TimelineID)
	if errors.Is(err, sql.ErrNoRows) {
		return ASRJob{}, ErrASRJobNotFound
	}
	return job, err
}

func (s *MySQLStore) CreateASRJob(ctx context.Context, owner string, job ASRJob) error {
	if err := validateASRJob(job); err != nil {
		return err
	}
	if _, err := s.Project(ctx, owner, job.ProjectID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO aicut_asr_jobs (id, project_id, owner_username, asset_id, model_id, language, expected_revision, state, progress, error_message, imported_count, track_id, created_at, updated_at, timeline_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, job.ID, job.ProjectID, owner, job.AssetID, job.ModelID, job.Language, job.ExpectedRevision, job.State, job.Progress, job.Error, job.Imported, job.TrackID, job.CreatedAt, job.UpdatedAt, job.TimelineID)
	return err
}

func (s *MySQLStore) ASRJob(ctx context.Context, owner, projectID, id string) (ASRJob, error) {
	return scanASRJob(s.db.QueryRowContext(ctx, `SELECT `+asrJobColumns+` FROM aicut_asr_jobs WHERE id = ? AND project_id = ? AND owner_username = ?`, id, projectID, owner))
}

func (s *MySQLStore) ListASRJobs(ctx context.Context, owner, projectID string) ([]ASRJob, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+asrJobColumns+` FROM aicut_asr_jobs WHERE project_id = ? AND owner_username = ? ORDER BY created_at DESC`, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ASRJob, 0)
	for rows.Next() {
		job, scanErr := scanASRJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, job)
	}
	return items, rows.Err()
}

func (s *MySQLStore) ClaimNextASRJob(ctx context.Context, now time.Time) (ASRJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ASRJob{}, err
	}
	defer tx.Rollback()
	job, err := scanASRJob(tx.QueryRowContext(ctx, `SELECT `+asrJobColumns+` FROM aicut_asr_jobs WHERE state = ? ORDER BY created_at, id LIMIT 1 FOR UPDATE SKIP LOCKED`, ASRQueued))
	if errors.Is(err, ErrASRJobNotFound) {
		return ASRJob{}, ErrASRQueueEmpty
	}
	if err != nil {
		return ASRJob{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE aicut_asr_jobs SET state = ?, progress = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ? AND state = ?`, ASRRunning, 5, now, job.ID, job.ProjectID, job.Owner, ASRQueued)
	if err != nil {
		return ASRJob{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ASRJob{}, ErrASRQueueEmpty
	}
	if err := tx.Commit(); err != nil {
		return ASRJob{}, err
	}
	return job.WithState(ASRRunning, 5, "", now), nil
}

func (s *MySQLStore) UpdateASRJob(ctx context.Context, job ASRJob) error {
	if err := validateASRJob(job); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_asr_jobs SET state = ?, progress = ?, error_message = ?, imported_count = ?, track_id = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, job.State, job.Progress, job.Error, job.Imported, job.TrackID, job.UpdatedAt, job.ID, job.ProjectID, job.Owner)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrASRJobNotFound
	}
	return nil
}

func (s *MySQLStore) RequestASRCancel(ctx context.Context, owner, projectID, id string, now time.Time) (ASRJob, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_asr_jobs SET state = CASE WHEN state = ? THEN ? WHEN state = ? THEN ? ELSE state END, progress = CASE WHEN state = ? THEN 100 ELSE progress END, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, ASRQueued, ASRCanceled, ASRRunning, ASRCancelRequested, ASRQueued, now, id, projectID, owner)
	if err != nil {
		return ASRJob{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ASRJob{}, ErrASRJobNotFound
	}
	return s.ASRJob(ctx, owner, projectID, id)
}

const exportJobColumns = `id, owner_username, project_id, format, expected_revision, state, progress, COALESCE(error_message, ''), output_path, options_json, project_snapshot_json, timeline_snapshot_json, created_at, updated_at`

func scanExportJob(row rowScanner) (ExportJob, error) {
	var job ExportJob
	var optionsJSON, projectJSON, timelineJSON []byte
	err := row.Scan(&job.ID, &job.Owner, &job.ProjectID, &job.Format, &job.ExpectedRevision, &job.State, &job.Progress, &job.Error, &job.OutputPath, &optionsJSON, &projectJSON, &timelineJSON, &job.CreatedAt, &job.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ExportJob{}, ErrExportJobNotFound
	}
	if err != nil {
		return ExportJob{}, err
	}
	if err := json.Unmarshal(optionsJSON, &job.Options); err != nil {
		return ExportJob{}, err
	}
	if err := json.Unmarshal(projectJSON, &job.ProjectSnapshot); err != nil {
		return ExportJob{}, err
	}
	if err := json.Unmarshal(timelineJSON, &job.TimelineSnapshot); err != nil {
		return ExportJob{}, err
	}
	return job, nil
}

func encodeExportJobSnapshots(job ExportJob) (string, string, string, error) {
	optionsJSON, err := json.Marshal(job.Options)
	if err != nil {
		return "", "", "", err
	}
	projectJSON, err := json.Marshal(job.ProjectSnapshot)
	if err != nil {
		return "", "", "", err
	}
	timelineJSON, err := json.Marshal(job.TimelineSnapshot)
	if err != nil {
		return "", "", "", err
	}
	// go-sql-driver/mysql binds []byte as binary. MySQL JSON columns reject that
	// value with Error 3144, so snapshots must be passed as UTF-8 text.
	return string(optionsJSON), string(projectJSON), string(timelineJSON), nil
}

func (s *MySQLStore) CreateExportJob(ctx context.Context, owner string, job ExportJob) error {
	job.Owner = strings.TrimSpace(owner)
	if err := validateExportJob(job); err != nil {
		return err
	}
	if _, err := s.Project(ctx, owner, job.ProjectID); err != nil {
		return err
	}
	optionsJSON, projectJSON, timelineJSON, err := encodeExportJobSnapshots(job)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO aicut_export_jobs (id, project_id, owner_username, format, expected_revision, state, progress, error_message, output_path, options_json, project_snapshot_json, timeline_snapshot_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, job.ID, job.ProjectID, owner, job.Format, job.ExpectedRevision, job.State, job.Progress, job.Error, job.OutputPath, optionsJSON, projectJSON, timelineJSON, job.CreatedAt, job.UpdatedAt)
	return err
}

func (s *MySQLStore) ExportJob(ctx context.Context, owner, projectID, id string) (ExportJob, error) {
	return scanExportJob(s.db.QueryRowContext(ctx, `SELECT `+exportJobColumns+` FROM aicut_export_jobs WHERE id = ? AND project_id = ? AND owner_username = ?`, id, projectID, owner))
}

func (s *MySQLStore) ListExportJobs(ctx context.Context, owner, projectID string) ([]ExportJob, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+exportJobColumns+` FROM aicut_export_jobs WHERE project_id = ? AND owner_username = ? ORDER BY created_at DESC, id DESC`, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ExportJob, 0)
	for rows.Next() {
		job, scanErr := scanExportJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, job)
	}
	return items, rows.Err()
}

func (s *MySQLStore) ClaimNextExportJob(ctx context.Context, now time.Time) (ExportJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ExportJob{}, err
	}
	defer tx.Rollback()
	job, err := scanExportJob(tx.QueryRowContext(ctx, `SELECT `+exportJobColumns+` FROM aicut_export_jobs WHERE state = ? ORDER BY created_at, id LIMIT 1 FOR UPDATE SKIP LOCKED`, ExportQueued))
	if errors.Is(err, ErrExportJobNotFound) {
		return ExportJob{}, ErrExportQueueEmpty
	}
	if err != nil {
		return ExportJob{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE aicut_export_jobs SET state = ?, progress = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ? AND state = ?`, ExportRunning, 5, now, job.ID, job.ProjectID, job.Owner, ExportQueued)
	if err != nil {
		return ExportJob{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ExportJob{}, ErrExportQueueEmpty
	}
	if err := tx.Commit(); err != nil {
		return ExportJob{}, err
	}
	return job.WithState(ExportRunning, 5, "", "", now), nil
}

func (s *MySQLStore) UpdateExportJob(ctx context.Context, job ExportJob) error {
	if err := validateExportJob(job); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_export_jobs SET state = ?, progress = ?, error_message = ?, output_path = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, job.State, job.Progress, job.Error, job.OutputPath, job.UpdatedAt, job.ID, job.ProjectID, job.Owner)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrExportJobNotFound
	}
	return nil
}

func (s *MySQLStore) RequestExportCancel(ctx context.Context, owner, projectID, id string, now time.Time) (ExportJob, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_export_jobs SET state = CASE WHEN state = ? THEN ? WHEN state = ? THEN ? ELSE state END, progress = CASE WHEN state = ? THEN 100 ELSE progress END, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, ExportQueued, ExportCanceled, ExportRunning, ExportCancelRequested, ExportQueued, now, id, projectID, owner)
	if err != nil {
		return ExportJob{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ExportJob{}, ErrExportJobNotFound
	}
	return s.ExportJob(ctx, owner, projectID, id)
}
