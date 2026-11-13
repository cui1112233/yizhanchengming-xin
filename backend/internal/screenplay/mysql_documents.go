package screenplay

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func contentDigest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// SaveScreenplay 追加一版剧本。项目行锁保证 A1 / B 档不会并发覆盖。
func (s *MySQLStore) SaveScreenplay(ctx context.Context, doc Document, shots []Shot, characters []Character, scenes []Scene) (Document, error) {
	if doc.ProjectID == "" || doc.OwnerUsername == "" {
		return Document{}, fmt.Errorf("%w: 项目与 owner 不能为空", ErrInvalidArgument)
	}
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback()

	var projectRevision int64
	err = tx.QueryRowContext(ctx, `SELECT revision FROM screenplay_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`,
		doc.ProjectID, doc.OwnerUsername).Scan(&projectRevision)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, doc.ProjectID)
	}
	if err != nil {
		return Document{}, err
	}
	var maxVersion int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM screenplay_documents WHERE project_id = ? AND owner_username = ?`,
		doc.ProjectID, doc.OwnerUsername).Scan(&maxVersion); err != nil {
		return Document{}, err
	}
	doc.Version = maxVersion + 1
	doc.Status = "active"
	doc.ContentSHA256 = contentDigest(doc.Content)
	doc.Revision = 1
	doc.CreatedAt = now
	doc.UpdatedAt = now
	doc.EpisodeCount = len(uniqueEpisodes(shots))

	if _, err := tx.ExecContext(ctx, `UPDATE screenplay_documents SET status = 'superseded', updated_at = ? WHERE project_id = ? AND owner_username = ? AND status = 'active'`,
		now, doc.ProjectID, doc.OwnerUsername); err != nil {
		return Document{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_documents
(id, project_id, owner_username, source, version, status, content, raw_result, content_sha256, chapter_count, episode_count, segment_count, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.ProjectID, doc.OwnerUsername, doc.Source, doc.Version, doc.Status, doc.Content, doc.RawResult,
		doc.ContentSHA256, doc.ChapterCount, doc.EpisodeCount, doc.SegmentCount, doc.Revision, doc.CreatedAt, doc.UpdatedAt); err != nil {
		return Document{}, err
	}
	// 同一 document_id 重复写入时先清空，保证幂等。
	if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_shots WHERE document_id = ? AND project_id = ? AND owner_username = ?`, doc.ID, doc.ProjectID, doc.OwnerUsername); err != nil {
		return Document{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_characters WHERE document_id = ? AND project_id = ? AND owner_username = ?`, doc.ID, doc.ProjectID, doc.OwnerUsername); err != nil {
		return Document{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM screenplay_scenes WHERE document_id = ? AND project_id = ? AND owner_username = ?`, doc.ID, doc.ProjectID, doc.OwnerUsername); err != nil {
		return Document{}, err
	}
	for _, shot := range shots {
		charactersJSON, err := marshalJSON(shot.CharacterNames)
		if err != nil {
			return Document{}, err
		}
		propsJSON, err := marshalJSON(shot.PropNames)
		if err != nil {
			return Document{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_shots
(id, document_id, project_id, owner_username, episode_no, sort_order, heading, script_text, character_names, scene_name, prop_names, location_reference, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			shot.ID, doc.ID, doc.ProjectID, doc.OwnerUsername, shot.EpisodeNo, shot.SortOrder, shot.Heading,
			shot.ScriptText, nullableBytes(charactersJSON), shot.SceneName, nullableBytes(propsJSON), shot.LocationReference, now, now); err != nil {
			return Document{}, err
		}
	}
	for _, character := range characters {
		if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_characters
(id, document_id, project_id, owner_username, name, description, image_prompt, reference_image, sort_order, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			character.ID, doc.ID, doc.ProjectID, doc.OwnerUsername, character.Name, character.Description,
			character.ImagePrompt, character.ReferenceImage, character.SortOrder, now, now); err != nil {
			return Document{}, err
		}
	}
	for _, scene := range scenes {
		if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_scenes
(id, document_id, project_id, owner_username, name, description, image_prompt, reference_image, sort_order, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			scene.ID, doc.ID, doc.ProjectID, doc.OwnerUsername, scene.Name, scene.Description,
			scene.ImagePrompt, scene.ReferenceImage, scene.SortOrder, now, now); err != nil {
			return Document{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (s *MySQLStore) ActiveDocument(ctx context.Context, owner, projectID string) (Document, error) {
	var doc Document
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, owner_username, source, version, status, content, COALESCE(raw_result, ''), content_sha256, chapter_count, episode_count, segment_count, revision, created_at, updated_at
FROM screenplay_documents WHERE project_id = ? AND owner_username = ? AND status = 'active' ORDER BY version DESC LIMIT 1`,
		projectID, owner).Scan(&doc.ID, &doc.ProjectID, &doc.OwnerUsername, &doc.Source, &doc.Version, &doc.Status,
		&doc.Content, &doc.RawResult, &doc.ContentSHA256, &doc.ChapterCount, &doc.EpisodeCount, &doc.SegmentCount,
		&doc.Revision, &doc.CreatedAt, &doc.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, ErrNoActiveScript
	}
	return doc, err
}

func (s *MySQLStore) GetDocument(ctx context.Context, owner, projectID, documentID string) (Document, error) {
	var doc Document
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, owner_username, source, version, status, content, COALESCE(raw_result, ''), content_sha256, chapter_count, episode_count, segment_count, revision, created_at, updated_at
FROM screenplay_documents WHERE id = ? AND project_id = ? AND owner_username = ?`, documentID, projectID, owner).
		Scan(&doc.ID, &doc.ProjectID, &doc.OwnerUsername, &doc.Source, &doc.Version, &doc.Status,
			&doc.Content, &doc.RawResult, &doc.ContentSHA256, &doc.ChapterCount, &doc.EpisodeCount, &doc.SegmentCount,
			&doc.Revision, &doc.CreatedAt, &doc.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, fmt.Errorf("%w: 剧本版本 %s", ErrNotFound, documentID)
	}
	return doc, err
}

func (s *MySQLStore) ListDocuments(ctx context.Context, owner, projectID string) ([]Document, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, owner_username, source, version, status, '' AS content, '' AS raw_result, content_sha256, chapter_count, episode_count, segment_count, revision, created_at, updated_at
FROM screenplay_documents WHERE project_id = ? AND owner_username = ? ORDER BY version DESC`, projectID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	documents := make([]Document, 0)
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.ProjectID, &doc.OwnerUsername, &doc.Source, &doc.Version, &doc.Status,
			&doc.Content, &doc.RawResult, &doc.ContentSHA256, &doc.ChapterCount, &doc.EpisodeCount, &doc.SegmentCount,
			&doc.Revision, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, rows.Err()
}

func (s *MySQLStore) ListShots(ctx context.Context, owner, projectID, documentID string) ([]Shot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, document_id, project_id, owner_username, episode_no, sort_order, heading, script_text, character_names, scene_name, prop_names, location_reference
FROM screenplay_shots WHERE project_id = ? AND owner_username = ? AND document_id = ? ORDER BY episode_no, sort_order`,
		projectID, owner, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	shots := make([]Shot, 0)
	for rows.Next() {
		var shot Shot
		var characterRaw, propRaw []byte
		if err := rows.Scan(&shot.ID, &shot.DocumentID, &shot.ProjectID, &shot.OwnerUsername, &shot.EpisodeNo,
			&shot.SortOrder, &shot.Heading, &shot.ScriptText, &characterRaw, &shot.SceneName, &propRaw, &shot.LocationReference); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(characterRaw, &shot.CharacterNames)
		_ = json.Unmarshal(propRaw, &shot.PropNames)
		shots = append(shots, shot)
	}
	return shots, rows.Err()
}

func (s *MySQLStore) CreateShot(ctx context.Context, shot Shot) (Shot, error) {
	if shot.OwnerUsername == "" || shot.ProjectID == "" || shot.DocumentID == "" {
		return Shot{}, fmt.Errorf("%w: 项目、剧本版本与用户不能为空", ErrInvalidArgument)
	}
	if strings.TrimSpace(shot.Heading) == "" && strings.TrimSpace(shot.ScriptText) == "" {
		return Shot{}, fmt.Errorf("%w: 场次标题和内容不能同时为空", ErrInvalidArgument)
	}
	if shot.EpisodeNo <= 0 {
		shot.EpisodeNo = 1
	}
	if shot.ID == "" {
		shot.ID = newScreenplayID("shot")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Shot{}, err
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 1 FROM screenplay_shots WHERE document_id = ? AND project_id = ? AND owner_username = ? AND episode_no = ? FOR UPDATE`, shot.DocumentID, shot.ProjectID, shot.OwnerUsername, shot.EpisodeNo).Scan(&shot.SortOrder); err != nil {
		return Shot{}, err
	}
	charactersJSON, err := marshalJSON(shot.CharacterNames)
	if err != nil {
		return Shot{}, err
	}
	propsJSON, err := marshalJSON(shot.PropNames)
	if err != nil {
		return Shot{}, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_shots
(id, document_id, project_id, owner_username, episode_no, sort_order, heading, script_text, character_names, scene_name, prop_names, location_reference, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, shot.ID, shot.DocumentID, shot.ProjectID, shot.OwnerUsername, shot.EpisodeNo, shot.SortOrder, strings.TrimSpace(shot.Heading), strings.TrimSpace(shot.ScriptText), nullableBytes(charactersJSON), strings.TrimSpace(shot.SceneName), nullableBytes(propsJSON), strings.TrimSpace(shot.LocationReference), now, now); err != nil {
		return Shot{}, err
	}
	if err := tx.Commit(); err != nil {
		return Shot{}, err
	}
	return shot, nil
}

func (s *MySQLStore) UpdateShot(ctx context.Context, owner, projectID, shotID string, changes Shot) (Shot, error) {
	if strings.TrimSpace(changes.Heading) == "" && strings.TrimSpace(changes.ScriptText) == "" {
		return Shot{}, fmt.Errorf("%w: 场次标题和内容不能同时为空", ErrInvalidArgument)
	}
	charactersJSON, err := marshalJSON(changes.CharacterNames)
	if err != nil {
		return Shot{}, err
	}
	propsJSON, err := marshalJSON(changes.PropNames)
	if err != nil {
		return Shot{}, err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE screenplay_shots SET episode_no = ?, heading = ?, script_text = ?, character_names = ?, scene_name = ?, prop_names = ?, location_reference = ?, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, changes.EpisodeNo, strings.TrimSpace(changes.Heading), strings.TrimSpace(changes.ScriptText), nullableBytes(charactersJSON), strings.TrimSpace(changes.SceneName), nullableBytes(propsJSON), strings.TrimSpace(changes.LocationReference), time.Now().UTC(), shotID, projectID, owner)
	if err != nil {
		return Shot{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Shot{}, fmt.Errorf("%w: 镜头 %s", ErrNotFound, shotID)
	}
	var shot Shot
	var characterRaw, propRaw []byte
	err = s.db.QueryRowContext(ctx, `SELECT id, document_id, project_id, owner_username, episode_no, sort_order, heading, script_text, character_names, scene_name, prop_names, location_reference FROM screenplay_shots WHERE id = ? AND project_id = ? AND owner_username = ?`, shotID, projectID, owner).Scan(&shot.ID, &shot.DocumentID, &shot.ProjectID, &shot.OwnerUsername, &shot.EpisodeNo, &shot.SortOrder, &shot.Heading, &shot.ScriptText, &characterRaw, &shot.SceneName, &propRaw, &shot.LocationReference)
	_ = json.Unmarshal(characterRaw, &shot.CharacterNames)
	_ = json.Unmarshal(propRaw, &shot.PropNames)
	return shot, err
}

func (s *MySQLStore) MoveShot(ctx context.Context, owner, projectID, shotID string, direction int) ([]Shot, error) {
	if direction != -1 && direction != 1 {
		return nil, fmt.Errorf("%w: 移动方向无效", ErrInvalidArgument)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var documentID string
	var episodeNo, sortOrder int
	if err := tx.QueryRowContext(ctx, `SELECT document_id, episode_no, sort_order FROM screenplay_shots WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, shotID, projectID, owner).Scan(&documentID, &episodeNo, &sortOrder); errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: 镜头 %s", ErrNotFound, shotID)
	} else if err != nil {
		return nil, err
	}
	operator, order := "<", "DESC"
	if direction > 0 {
		operator, order = ">", "ASC"
	}
	query := fmt.Sprintf(`SELECT id, sort_order FROM screenplay_shots WHERE document_id = ? AND project_id = ? AND owner_username = ? AND episode_no = ? AND sort_order %s ? ORDER BY sort_order %s LIMIT 1 FOR UPDATE`, operator, order)
	var neighborID string
	var neighborOrder int
	if err := tx.QueryRowContext(ctx, query, documentID, projectID, owner, episodeNo, sortOrder).Scan(&neighborID, &neighborOrder); errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: 已到当前集边界", ErrInvalidState)
	} else if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE screenplay_shots SET sort_order = CASE id WHEN ? THEN ? WHEN ? THEN ? END, updated_at = ? WHERE id IN (?, ?) AND project_id = ? AND owner_username = ?`, shotID, neighborOrder, neighborID, sortOrder, time.Now().UTC(), shotID, neighborID, projectID, owner); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.ListShots(ctx, owner, projectID, documentID)
}

func (s *MySQLStore) DeleteShot(ctx context.Context, owner, projectID, shotID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM screenplay_shots WHERE id = ? AND project_id = ? AND owner_username = ?`, shotID, projectID, owner)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("%w: 镜头 %s", ErrNotFound, shotID)
	}
	return nil
}

func (s *MySQLStore) ListCharacters(ctx context.Context, owner, projectID, documentID string) ([]Character, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, document_id, project_id, owner_username, name, COALESCE(description, ''), COALESCE(image_prompt, ''), COALESCE(reference_image, ''), sort_order
FROM screenplay_characters WHERE project_id = ? AND owner_username = ? AND document_id = ? ORDER BY sort_order, name`,
		projectID, owner, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	characters := make([]Character, 0)
	for rows.Next() {
		var character Character
		if err := rows.Scan(&character.ID, &character.DocumentID, &character.ProjectID, &character.OwnerUsername,
			&character.Name, &character.Description, &character.ImagePrompt, &character.ReferenceImage, &character.SortOrder); err != nil {
			return nil, err
		}
		characters = append(characters, character)
	}
	return characters, rows.Err()
}

func (s *MySQLStore) ListScenes(ctx context.Context, owner, projectID, documentID string) ([]Scene, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, document_id, project_id, owner_username, name, COALESCE(description, ''), COALESCE(image_prompt, ''), COALESCE(reference_image, ''), sort_order
FROM screenplay_scenes WHERE project_id = ? AND owner_username = ? AND document_id = ? ORDER BY sort_order, name`,
		projectID, owner, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	scenes := make([]Scene, 0)
	for rows.Next() {
		var scene Scene
		if err := rows.Scan(&scene.ID, &scene.DocumentID, &scene.ProjectID, &scene.OwnerUsername,
			&scene.Name, &scene.Description, &scene.ImagePrompt, &scene.ReferenceImage, &scene.SortOrder); err != nil {
			return nil, err
		}
		scenes = append(scenes, scene)
	}
	return scenes, rows.Err()
}

func uniqueEpisodes(shots []Shot) []int {
	seen := map[int]bool{}
	out := make([]int, 0)
	for _, shot := range shots {
		if shot.EpisodeNo <= 0 || seen[shot.EpisodeNo] {
			continue
		}
		seen[shot.EpisodeNo] = true
		out = append(out, shot.EpisodeNo)
	}
	return out
}
