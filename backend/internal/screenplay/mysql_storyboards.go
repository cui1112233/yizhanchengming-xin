package screenplay

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SaveStoryboard 幂等覆盖「项目 × 用户 × 剧集 × 引擎」这一条分镜记录：
// 行锁保证并发重复生成不会互相丢更新，旧分镜镜头整批替换。
func (s *MySQLStore) SaveStoryboard(ctx context.Context, storyboard Storyboard) (Storyboard, error) {
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
		if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_storyboard_shots
(id, storyboard_id, project_id, owner_username, seq, scene_id, duration_sec, characters_text, scene_name, props_text, video_desc, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			shot.ID, shot.StoryboardID, shot.ProjectID, shot.OwnerUsername, shot.Seq, shot.SceneID,
			shot.Duration, shot.Characters, shot.Scene, shot.Props, shot.VideoDesc, now, now); err != nil {
			return Storyboard{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Storyboard{}, err
	}
	return storyboard, nil
}

func (s *MySQLStore) GetStoryboard(ctx context.Context, owner, projectID string, episodeNo int, engine StoryboardEngine) (Storyboard, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, project_id, owner_username, episode_no, engine, status,
source_sha256, shot_count, revision, created_at, updated_at
FROM screenplay_storyboards WHERE project_id = ? AND owner_username = ? AND episode_no = ? AND engine = ?`,
		projectID, owner, episodeNo, string(engine))
	storyboard, err := scanStoryboard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Storyboard{}, fmt.Errorf("%w: 第 %d 集 %s 分镜", ErrNotFound, episodeNo, engine)
	}
	if err != nil {
		return Storyboard{}, err
	}
	shots, err := s.listStoryboardShots(ctx, owner, projectID, storyboard.ID)
	if err != nil {
		return Storyboard{}, err
	}
	storyboard.Shots = shots
	return storyboard, nil
}

func (s *MySQLStore) ListStoryboards(ctx context.Context, owner, projectID string, engine StoryboardEngine) ([]Storyboard, error) {
	query := `SELECT id, project_id, owner_username, episode_no, engine, status, source_sha256,
shot_count, revision, created_at, updated_at FROM screenplay_storyboards
WHERE project_id = ? AND owner_username = ?`
	args := []any{projectID, owner}
	if engine != "" {
		query += " AND engine = ?"
		args = append(args, string(engine))
	}
	query += " ORDER BY episode_no"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	storyboards := make([]Storyboard, 0)
	for rows.Next() {
		storyboard, err := scanStoryboard(rows)
		if err != nil {
			return nil, err
		}
		storyboards = append(storyboards, storyboard)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range storyboards {
		shots, err := s.listStoryboardShots(ctx, owner, projectID, storyboards[index].ID)
		if err != nil {
			return nil, err
		}
		storyboards[index].Shots = shots
	}
	return storyboards, nil
}

func (s *MySQLStore) listStoryboardShots(ctx context.Context, owner, projectID, storyboardID string) ([]StoryboardShot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, storyboard_id, project_id, owner_username, seq,
scene_id, duration_sec, characters_text, scene_name, props_text, video_desc
FROM screenplay_storyboard_shots WHERE storyboard_id = ? AND project_id = ? AND owner_username = ?
ORDER BY seq`, storyboardID, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	shots := make([]StoryboardShot, 0)
	for rows.Next() {
		var shot StoryboardShot
		if err := rows.Scan(&shot.ID, &shot.StoryboardID, &shot.ProjectID, &shot.OwnerUsername,
			&shot.Seq, &shot.SceneID, &shot.Duration, &shot.Characters, &shot.Scene, &shot.Props,
			&shot.VideoDesc); err != nil {
			return nil, err
		}
		shots = append(shots, shot)
	}
	return shots, rows.Err()
}

type storyboardScanner interface {
	Scan(dest ...any) error
}

func scanStoryboard(scanner storyboardScanner) (Storyboard, error) {
	var storyboard Storyboard
	err := scanner.Scan(&storyboard.ID, &storyboard.ProjectID, &storyboard.OwnerUsername,
		&storyboard.EpisodeNo, &storyboard.Engine, &storyboard.Status, &storyboard.SourceSHA256,
		&storyboard.ShotCount, &storyboard.Revision, &storyboard.CreatedAt, &storyboard.UpdatedAt)
	return storyboard, err
}
