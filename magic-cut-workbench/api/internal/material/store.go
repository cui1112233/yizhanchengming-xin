package material

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

const insertMaterialSQL = `INSERT INTO magic_cut_materials (id, project_id, owner_username, original_name, media_type, storage_key, size_bytes, metadata_json, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
const insertTaskSQL = `INSERT INTO magic_cut_tasks (id, project_id, owner_username, task_type, status, progress, input_json, output_json, error_code, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
const listMaterialsSQL = `SELECT id, project_id, owner_username, original_name, media_type, storage_key, size_bytes, metadata_json, status, created_at, updated_at FROM magic_cut_materials WHERE owner_username = ? AND project_id = ? ORDER BY created_at DESC`

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) SaveImport(ctx context.Context, item Material, probeTask worktask.Task) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("material store is not configured")
	}
	metadata, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}
	input, err := json.Marshal(probeTask.Input)
	if err != nil {
		return err
	}
	output, err := json.Marshal(probeTask.Output)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, insertMaterialSQL, item.ID, item.ProjectID, item.Owner, item.OriginalName, item.MediaType, item.StorageKey, item.SizeBytes, metadata, item.Status, item.CreatedAt, item.UpdatedAt); err != nil {
		return fmt.Errorf("insert material: %w", err)
	}
	var errorCode, errorMessage any
	if probeTask.ErrorCode != "" {
		errorCode = probeTask.ErrorCode
	}
	if probeTask.ErrorMessage != "" {
		errorMessage = probeTask.ErrorMessage
	}
	if _, err = tx.ExecContext(ctx, insertTaskSQL, probeTask.ID, probeTask.ProjectID, probeTask.Owner, probeTask.Type, probeTask.Status, probeTask.Progress, input, output, errorCode, errorMessage, probeTask.CreatedAt, probeTask.UpdatedAt); err != nil {
		return fmt.Errorf("insert material task: %w", err)
	}
	return tx.Commit()
}

func (s *Store) List(ctx context.Context, owner, projectID string) ([]Material, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("material store is not configured")
	}
	rows, err := s.db.QueryContext(ctx, listMaterialsSQL, owner, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Material, 0)
	for rows.Next() {
		var item Material
		var metadata []byte
		if err = rows.Scan(&item.ID, &item.ProjectID, &item.Owner, &item.OriginalName, &item.MediaType, &item.StorageKey, &item.SizeBytes, &metadata, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &item.Metadata)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
