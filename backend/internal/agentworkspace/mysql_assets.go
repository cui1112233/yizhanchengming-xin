package agentworkspace

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type MySQLAssetStore struct{ db *sql.DB }

func NewMySQLAssetStore(db *sql.DB) *MySQLAssetStore { return &MySQLAssetStore{db: db} }

func (s *MySQLAssetStore) ListAssets(ctx context.Context, owner, projectID, kind string) ([]Asset, error) {
	query := `SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at
FROM agent_assets WHERE owner_username = ? AND project_id = ?`
	args := []any{owner, projectID}
	if kind != "" {
		query += ` AND kind = ?`
		args = append(args, kind)
	}
	query += ` ORDER BY created_at DESC, id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Asset, 0)
	for rows.Next() {
		var asset Asset
		if err := rows.Scan(&asset.ID, &asset.ProjectID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, asset)
	}
	return items, rows.Err()
}

func (s *MySQLAssetStore) CreateAsset(ctx context.Context, owner string, asset Asset) error {
	if err := ValidateAsset(asset); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO agent_assets (id, project_id, owner_username, kind, name, content_type, object_key, byte_size, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, asset.ID, asset.ProjectID, owner, asset.Kind, asset.Name, asset.ContentType, asset.ObjectKey, asset.ByteSize, asset.CreatedAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrAssetExists
	}
	return err
}

func (s *MySQLAssetStore) Asset(ctx context.Context, owner, projectID, assetID string) (Asset, error) {
	var asset Asset
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, kind, name, content_type, object_key, byte_size, created_at
FROM agent_assets WHERE owner_username = ? AND project_id = ? AND id = ?`, owner, projectID, assetID).
		Scan(&asset.ID, &asset.ProjectID, &asset.Kind, &asset.Name, &asset.ContentType, &asset.ObjectKey, &asset.ByteSize, &asset.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Asset{}, ErrAssetNotFound
	}
	return asset, err
}
