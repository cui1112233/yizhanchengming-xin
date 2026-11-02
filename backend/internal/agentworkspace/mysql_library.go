package agentworkspace

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const libraryItemColumns = "id, folder_id, source_project_id, source_asset_id, kind, name, tags_json, created_at, updated_at"

func (s *MySQLStore) CreateLibraryFolder(ctx context.Context, owner string, input CreateLibraryFolderInput) (LibraryFolder, error) {
	input, err := normalizeLibraryFolderInput(input)
	if err != nil || strings.TrimSpace(owner) == "" {
		return LibraryFolder{}, ErrLibraryInvalid
	}
	folder := LibraryFolder{ID: input.ID, Name: input.Name, CreatedAt: input.CreatedAt.UTC(), UpdatedAt: input.CreatedAt.UTC()}
	_, err = s.db.ExecContext(ctx, `INSERT INTO agent_library_folders (id, owner_username, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, folder.ID, owner, folder.Name, folder.CreatedAt, folder.UpdatedAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return LibraryFolder{}, ErrLibraryFolderExists
	}
	return folder, err
}

func (s *MySQLStore) ListLibraryFolders(ctx context.Context, owner string) ([]LibraryFolder, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM agent_library_folders WHERE owner_username = ? ORDER BY updated_at DESC, id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]LibraryFolder, 0)
	for rows.Next() {
		var folder LibraryFolder
		if err := rows.Scan(&folder.ID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, folder)
	}
	return items, rows.Err()
}

func (s *MySQLStore) SaveLibraryItem(ctx context.Context, owner string, input SaveLibraryItemInput) (LibraryItem, error) {
	input, err := normalizeLibraryItemInput(input)
	if err != nil || strings.TrimSpace(owner) == "" {
		return LibraryItem{}, ErrLibraryInvalid
	}
	var exists int
	err = s.db.QueryRowContext(ctx, `SELECT 1 FROM agent_library_folders WHERE id = ? AND owner_username = ?`, input.FolderID, owner).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return LibraryItem{}, ErrLibraryFolderNotFound
	}
	if err != nil {
		return LibraryItem{}, err
	}
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		return LibraryItem{}, err
	}
	item := LibraryItem{ID: input.ID, FolderID: input.FolderID, SourceProjectID: input.SourceProjectID, SourceAssetID: input.SourceAssetID, Kind: input.Kind, Name: input.Name, Tags: append([]string(nil), input.Tags...), CreatedAt: input.CreatedAt.UTC(), UpdatedAt: input.CreatedAt.UTC()}
	_, err = s.db.ExecContext(ctx, `INSERT INTO agent_library_items (id, folder_id, owner_username, source_project_id, source_asset_id, kind, name, tags_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.ID, item.FolderID, owner, item.SourceProjectID, item.SourceAssetID, item.Kind, item.Name, string(tagsJSON), item.CreatedAt, item.UpdatedAt)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return LibraryItem{}, ErrLibraryItemExists
	}
	return item, err
}

func (s *MySQLStore) ListLibraryItems(ctx context.Context, owner, folderID string, filter LibraryItemFilter) ([]LibraryItem, error) {
	query := `SELECT ` + libraryItemColumns + ` FROM agent_library_items WHERE owner_username = ? AND folder_id = ?`
	args := []any{owner, strings.TrimSpace(folderID)}
	if kind := strings.TrimSpace(filter.Kind); kind != "" {
		query += ` AND kind = ?`
		args = append(args, kind)
	}
	if text := strings.ToLower(strings.TrimSpace(filter.Query)); text != "" {
		like := "%" + text + "%"
		query += ` AND (LOWER(name) LIKE ? OR LOWER(tags_json) LIKE ?)`
		args = append(args, like, like)
	}
	query += ` ORDER BY updated_at DESC, id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]LibraryItem, 0)
	for rows.Next() {
		item, err := scanLibraryItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) LibraryItem(ctx context.Context, owner, id string) (LibraryItem, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+libraryItemColumns+` FROM agent_library_items WHERE owner_username = ? AND id = ?`, owner, strings.TrimSpace(id))
	item, err := scanLibraryItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return LibraryItem{}, ErrLibraryItemNotFound
	}
	return item, err
}

func (s *MySQLStore) MoveLibraryItem(ctx context.Context, owner, itemID, folderID string, now time.Time) (LibraryItem, error) {
	if now.IsZero() {
		return LibraryItem{}, ErrLibraryInvalid
	}
	folderID = strings.TrimSpace(folderID)
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM agent_library_folders WHERE id = ? AND owner_username = ?`, folderID, owner).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return LibraryItem{}, ErrLibraryFolderNotFound
	}
	if err != nil {
		return LibraryItem{}, err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE agent_library_items SET folder_id = ?, updated_at = ? WHERE id = ? AND owner_username = ?`, folderID, now.UTC(), strings.TrimSpace(itemID), owner)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return LibraryItem{}, ErrLibraryItemExists
	}
	if err != nil {
		return LibraryItem{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return LibraryItem{}, err
	}
	if affected == 0 {
		return LibraryItem{}, ErrLibraryItemNotFound
	}
	return s.LibraryItem(ctx, owner, itemID)
}

type libraryItemScanner interface {
	Scan(...any) error
}

func scanLibraryItem(scanner libraryItemScanner) (LibraryItem, error) {
	var item LibraryItem
	var tagsJSON []byte
	if err := scanner.Scan(&item.ID, &item.FolderID, &item.SourceProjectID, &item.SourceAssetID, &item.Kind, &item.Name, &tagsJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return LibraryItem{}, err
	}
	if len(tagsJSON) > 0 && json.Unmarshal(tagsJSON, &item.Tags) != nil {
		return LibraryItem{}, ErrLibraryInvalid
	}
	if item.Tags == nil {
		item.Tags = []string{}
	}
	return item, nil
}
