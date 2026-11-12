package aicut

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func (s *MySQLStore) ListAssetFolders(ctx context.Context, owner, project string) ([]AssetFolder, error) {
	if _, err := s.Project(ctx, owner, project); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at FROM aicut_asset_folders WHERE project_id = ? AND owner_username = ? ORDER BY created_at, id`, project, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AssetFolder, 0)
	for rows.Next() {
		f := AssetFolder{ProjectID: project}
		if err := rows.Scan(&f.ID, &f.Name, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
func (s *MySQLStore) CreateAssetFolder(ctx context.Context, owner string, folder AssetFolder) error {
	if folder.ID == "" || !validAssetFolderName(folder.Name) {
		return ErrInvalidAssetFolder
	}
	if _, err := s.Project(ctx, owner, folder.ProjectID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO aicut_asset_folders (id, project_id, owner_username, name, created_at) VALUES (?, ?, ?, ?, ?)`, folder.ID, folder.ProjectID, owner, strings.TrimSpace(folder.Name), folder.CreatedAt)
	return err
}
func (s *MySQLStore) RenameAssetFolder(ctx context.Context, owner, project, id, name string) error {
	if !validAssetFolderName(name) {
		return ErrInvalidAssetFolder
	}
	if _, err := s.Project(ctx, owner, project); err != nil {
		return err
	}
	// Lock/read first: unchanged names are valid even with MySQL changed-rows semantics.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var found string
	err = tx.QueryRowContext(ctx, `SELECT id FROM aicut_asset_folders WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, id, project, owner).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAssetFolderNotFound
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE aicut_asset_folders SET name = ? WHERE id = ? AND project_id = ? AND owner_username = ?`, strings.TrimSpace(name), id, project, owner); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *MySQLStore) DeleteAssetFolder(ctx context.Context, owner, project, id string) error {
	if _, err := s.Project(ctx, owner, project); err != nil {
		return err
	}
	// Memberships cascade away; the referenced assets are deliberately retained.
	result, err := s.db.ExecContext(ctx, `DELETE FROM aicut_asset_folders WHERE id = ? AND project_id = ? AND owner_username = ?`, id, project, owner)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return ErrAssetFolderNotFound
	}
	return err
}
func (s *MySQLStore) MoveAssetToFolder(ctx context.Context, owner, project, asset, folder string) error {
	if _, err := s.Project(ctx, owner, project); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var found string
	err = tx.QueryRowContext(ctx, `SELECT id FROM aicut_assets WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, asset, project, owner).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAssetNotFound
	}
	if err != nil {
		return err
	}
	if folder != "" {
		err = tx.QueryRowContext(ctx, `SELECT id FROM aicut_asset_folders WHERE id = ? AND project_id = ? AND owner_username = ? FOR UPDATE`, folder, project, owner).Scan(&found)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAssetFolderNotFound
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO aicut_asset_folder_members (asset_id, project_id, owner_username, folder_id) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE folder_id = VALUES(folder_id)`, asset, project, owner, folder)
	} else {
		_, err = tx.ExecContext(ctx, `DELETE FROM aicut_asset_folder_members WHERE asset_id = ? AND project_id = ? AND owner_username = ?`, asset, project, owner)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *MySQLStore) AssetFolderAssignments(ctx context.Context, owner, project string) (map[string]string, error) {
	if _, err := s.Project(ctx, owner, project); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT asset_id, folder_id FROM aicut_asset_folder_members WHERE project_id = ? AND owner_username = ?`, project, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var asset, folder string
		if err := rows.Scan(&asset, &folder); err != nil {
			return nil, err
		}
		out[asset] = folder
	}
	return out, rows.Err()
}
