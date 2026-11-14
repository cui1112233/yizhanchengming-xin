package screenplay

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s *MySQLStore) SaveStoryboardRevision(ctx context.Context, storyboard Storyboard, expected int64) (Storyboard, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Storyboard{}, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if storyboard.CreatedAt.IsZero() {
		storyboard.CreatedAt = now
	}
	storyboard.UpdatedAt = now
	storyboard.ShotCount = len(storyboard.Shots)
	if storyboard.Status == "" {
		storyboard.Status = "done"
	}

	var existingID string
	var revision int64
	err = tx.QueryRowContext(ctx, `SELECT id, revision FROM screenplay_storyboards
WHERE project_id = ? AND owner_username = ? AND episode_no = ? AND engine = ? FOR UPDATE`,
		storyboard.ProjectID, storyboard.OwnerUsername, storyboard.EpisodeNo, string(storyboard.Engine)).
		Scan(&existingID, &revision)
	switch {
	case err == nil:
		if revision != expected {
			return Storyboard{}, ErrRevisionConflict
		}
		storyboard.ID = existingID
		storyboard.Revision = revision + 1
		if _, err := tx.ExecContext(ctx, `UPDATE screenplay_storyboards
SET status = ?, source_sha256 = ?, shot_count = ?, revision = ?, updated_at = ?
WHERE id = ? AND project_id = ? AND owner_username = ?`,
			storyboard.Status, storyboard.SourceSHA256, storyboard.ShotCount, storyboard.Revision,
			storyboard.UpdatedAt, storyboard.ID, storyboard.ProjectID, storyboard.OwnerUsername); err != nil {
			return Storyboard{}, err
		}
	case errors.Is(err, sql.ErrNoRows):
		if expected != 0 {
			return Storyboard{}, ErrRevisionConflict
		}
		if storyboard.ID == "" {
			storyboard.ID = newScreenplayID("sb")
		}
		storyboard.Revision = 1
		if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_storyboards
(id, project_id, owner_username, episode_no, engine, status, source_sha256, shot_count, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			storyboard.ID, storyboard.ProjectID, storyboard.OwnerUsername, storyboard.EpisodeNo,
			string(storyboard.Engine), storyboard.Status, storyboard.SourceSHA256, storyboard.ShotCount,
			storyboard.Revision, storyboard.CreatedAt, storyboard.UpdatedAt); err != nil {
			return Storyboard{}, err
		}
	default:
		return Storyboard{}, err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_storyboard_shots
WHERE storyboard_id = ? AND project_id = ? AND owner_username = ?`,
		storyboard.ID, storyboard.ProjectID, storyboard.OwnerUsername); err != nil {
		return Storyboard{}, err
	}
	for index := range storyboard.Shots {
		shot := &storyboard.Shots[index]
		if shot.ID == "" {
			shot.ID = newScreenplayID("sb_shot")
		}
		if shot.Seq <= 0 {
			shot.Seq = index + 1
		}
		shot.StoryboardID = storyboard.ID
		shot.ProjectID = storyboard.ProjectID
		shot.OwnerUsername = storyboard.OwnerUsername
		protocol, err := marshalStoryboardProtocol(*shot)
		if err != nil {
			return Storyboard{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_storyboard_shots
(id, storyboard_id, project_id, owner_username, seq, scene_id, duration_sec, characters_text, scene_name, props_text, video_desc, protocol_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			shot.ID, shot.StoryboardID, shot.ProjectID, shot.OwnerUsername, shot.Seq, shot.SceneID,
			shot.Duration, shot.Characters, shot.Scene, shot.Props, shot.VideoDesc, protocol, now, now); err != nil {
			return Storyboard{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Storyboard{}, err
	}
	return storyboard, nil
}
