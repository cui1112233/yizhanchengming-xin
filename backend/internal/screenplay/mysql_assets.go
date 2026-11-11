package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *MySQLStore) SaveAssetBatch(ctx context.Context, batch AssetBatch) error {
	now := time.Now().UTC()
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt = now
	}
	batch.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, `INSERT INTO screenplay_asset_batches
(id, project_id, owner_username, asset_type, degree, batch_index, chapter_start, chapter_end, status, item_count, error_message, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
ON DUPLICATE KEY UPDATE status = VALUES(status), item_count = VALUES(item_count),
error_message = VALUES(error_message), revision = revision + 1, updated_at = VALUES(updated_at)`,
		batch.ID, batch.ProjectID, batch.OwnerUsername, string(batch.Type), string(batch.Degree),
		batch.BatchIndex, batch.ChapterStart, batch.ChapterEnd, batch.Status, batch.ItemCount,
		nullString(batch.ErrorMessage), batch.CreatedAt, batch.UpdatedAt)
	return err
}

// MergeAssets 按 (项目, 类型, 名称) 合并同名资产；名称冲突时后写覆盖前写。
func (s *MySQLStore) MergeAssets(ctx context.Context, owner, projectID string, assets []Asset) ([]Asset, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	for index, asset := range assets {
		extra, marshalErr := marshalJSON(asset.Extra)
		if marshalErr != nil {
			return nil, marshalErr
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO screenplay_assets
(id, project_id, owner_username, asset_type, degree, name, identity_text, description, prompt, voice_text, episodes, tag, extra, sort_order, source_batch, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
ON DUPLICATE KEY UPDATE degree = VALUES(degree), identity_text = VALUES(identity_text),
description = VALUES(description), prompt = VALUES(prompt), voice_text = VALUES(voice_text),
episodes = VALUES(episodes), tag = VALUES(tag), extra = VALUES(extra),
sort_order = VALUES(sort_order), source_batch = VALUES(source_batch),
revision = revision + 1, updated_at = VALUES(updated_at)`,
			asset.ID, projectID, owner, string(asset.Type), string(asset.Degree), asset.Name,
			nullString(asset.Identity), nullString(asset.Description), nullString(asset.Prompt),
			nullString(asset.Voice), asset.Episodes, asset.Tag, nullableBytes(extra), index+1,
			asset.SourceBatch, now, now)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return assets, nil
}

func (s *MySQLStore) ListAssets(ctx context.Context, owner, projectID string, assetType AssetType) ([]Asset, error) {
	query := `SELECT id, project_id, owner_username, asset_type, degree, name,
COALESCE(identity_text, ''), COALESCE(description, ''), COALESCE(prompt, ''), COALESCE(voice_text, ''),
episodes, tag, extra, sort_order, source_batch, revision, created_at, updated_at
FROM screenplay_assets WHERE project_id = ? AND owner_username = ?`
	args := []any{projectID, owner}
	if assetType != "" {
		query += ` AND asset_type = ?`
		args = append(args, string(assetType))
	}
	query += ` ORDER BY asset_type, sort_order, name`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		var extra []byte
		if err := rows.Scan(&asset.ID, &asset.ProjectID, &asset.OwnerUsername, &asset.Type, &asset.Degree,
			&asset.Name, &asset.Identity, &asset.Description, &asset.Prompt, &asset.Voice,
			&asset.Episodes, &asset.Tag, &extra, &asset.SortOrder, &asset.SourceBatch,
			&asset.Revision, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
			return nil, err
		}
		if len(extra) > 0 {
			_ = json.Unmarshal(extra, &asset.Extra)
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func (s *MySQLStore) CreateAsset(ctx context.Context, asset Asset) (Asset, error) {
	asset.OwnerUsername = strings.TrimSpace(asset.OwnerUsername)
	asset.ProjectID = strings.TrimSpace(asset.ProjectID)
	asset.Name = strings.TrimSpace(asset.Name)
	if !validAssetType(asset.Type) || asset.Name == "" {
		return Asset{}, fmt.Errorf("%w: 资产类型和名称不能为空", ErrInvalidArgument)
	}
	if asset.Degree == "" {
		asset.Degree = DegreeStandard
	}
	if asset.ID == "" {
		asset.ID = newScreenplayID("asset")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Asset{}, err
	}
	defer tx.Rollback()
	var projectID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM screenplay_projects WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, asset.ProjectID, asset.OwnerUsername).Scan(&projectID); errors.Is(err, sql.ErrNoRows) {
		return Asset{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, asset.ProjectID)
	} else if err != nil {
		return Asset{}, err
	}
	var duplicate int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM screenplay_assets WHERE project_id = ? AND owner_username = ? AND asset_type = ? AND name = ?`, asset.ProjectID, asset.OwnerUsername, string(asset.Type), asset.Name).Scan(&duplicate); err != nil {
		return Asset{}, err
	}
	if duplicate > 0 {
		return Asset{}, fmt.Errorf("%w: 同类型资产名称已存在", ErrInvalidArgument)
	}
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 1 FROM screenplay_assets WHERE project_id = ? AND owner_username = ?`, asset.ProjectID, asset.OwnerUsername).Scan(&asset.SortOrder); err != nil {
		return Asset{}, err
	}
	extra, err := marshalJSON(asset.Extra)
	if err != nil {
		return Asset{}, err
	}
	now := time.Now().UTC()
	asset.Revision, asset.CreatedAt, asset.UpdatedAt = 1, now, now
	if _, err := tx.ExecContext(ctx, `INSERT INTO screenplay_assets
(id, project_id, owner_username, asset_type, degree, name, identity_text, description, prompt, voice_text, episodes, tag, extra, sort_order, source_batch, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`, asset.ID, asset.ProjectID, asset.OwnerUsername, string(asset.Type), string(asset.Degree), asset.Name, nullString(asset.Identity), nullString(asset.Description), nullString(asset.Prompt), nullString(asset.Voice), asset.Episodes, asset.Tag, nullableBytes(extra), asset.SortOrder, asset.SourceBatch, now, now); err != nil {
		return Asset{}, err
	}
	if err := tx.Commit(); err != nil {
		return Asset{}, err
	}
	return asset, nil
}

func (s *MySQLStore) UpdateAsset(ctx context.Context, owner, projectID, assetID string, changes Asset, expectedRevision int64) (Asset, error) {
	changes.Name = strings.TrimSpace(changes.Name)
	if !validAssetType(changes.Type) || changes.Name == "" {
		return Asset{}, fmt.Errorf("%w: 资产类型和名称不能为空", ErrInvalidArgument)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Asset{}, err
	}
	defer tx.Rollback()
	var current Asset
	err = tx.QueryRowContext(ctx, `SELECT degree, revision, created_at, sort_order, source_batch FROM screenplay_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, assetID, projectID, owner).Scan(&current.Degree, &current.Revision, &current.CreatedAt, &current.SortOrder, &current.SourceBatch)
	if errors.Is(err, sql.ErrNoRows) {
		return Asset{}, fmt.Errorf("%w: 资产 %s", ErrNotFound, assetID)
	}
	if err != nil {
		return Asset{}, err
	}
	if current.Revision != expectedRevision {
		return Asset{}, ErrRevisionConflict
	}
	var duplicate int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM screenplay_assets WHERE project_id = ? AND owner_username = ? AND asset_type = ? AND name = ? AND id <> ?`, projectID, owner, string(changes.Type), changes.Name, assetID).Scan(&duplicate); err != nil {
		return Asset{}, err
	}
	if duplicate > 0 {
		return Asset{}, fmt.Errorf("%w: 同类型资产名称已存在", ErrInvalidArgument)
	}
	if changes.Degree == "" {
		changes.Degree = current.Degree
	}
	extra, err := marshalJSON(changes.Extra)
	if err != nil {
		return Asset{}, err
	}
	now := time.Now().UTC()
	result, err := tx.ExecContext(ctx, `UPDATE screenplay_assets SET asset_type = ?, degree = ?, name = ?, identity_text = ?, description = ?, prompt = ?, voice_text = ?, episodes = ?, tag = ?, extra = ?, revision = revision + 1, updated_at = ? WHERE id = ? AND project_id = ? AND owner_username = ? AND revision = ?`, string(changes.Type), string(changes.Degree), changes.Name, nullString(changes.Identity), nullString(changes.Description), nullString(changes.Prompt), nullString(changes.Voice), changes.Episodes, changes.Tag, nullableBytes(extra), now, assetID, projectID, owner, expectedRevision)
	if err != nil {
		return Asset{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return Asset{}, ErrRevisionConflict
	}
	if err := tx.Commit(); err != nil {
		return Asset{}, err
	}
	changes.ID = assetID
	changes.ProjectID = projectID
	changes.OwnerUsername = owner
	changes.SortOrder = current.SortOrder
	changes.SourceBatch = current.SourceBatch
	changes.Revision = expectedRevision + 1
	changes.CreatedAt = current.CreatedAt
	changes.UpdatedAt = now
	return changes, nil
}

func (s *MySQLStore) DeleteAsset(ctx context.Context, owner, projectID, assetID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM screenplay_assets WHERE id = ? AND project_id = ? AND owner_username = ?`, assetID, projectID, owner)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return fmt.Errorf("%w: 资产 %s", ErrNotFound, assetID)
	}
	return nil
}
