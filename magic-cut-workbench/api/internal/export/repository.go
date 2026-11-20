package export

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"yizhanchengming/magic-cut-workbench/internal/material"
	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

const insertExportSQL = `INSERT INTO magic_cut_exports (id, project_id, task_id, owner_username, output_name, storage_key, media_metadata_json, sha256, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
const insertExportTaskSQL = `INSERT INTO magic_cut_tasks (id, project_id, owner_username, task_type, status, progress, input_json, output_json, error_code, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
const selectAttemptNumberSQL = `SELECT COALESCE(MAX(attempt_number), 0) FROM magic_cut_task_attempts WHERE task_id = ? FOR UPDATE`
const insertAttemptSQL = `INSERT INTO magic_cut_task_attempts (task_id, attempt_number, status, started_at, finished_at, error_code, error_message) VALUES (?, ?, ?, ?, NULL, NULL, NULL)`
const updateTaskRunningSQL = `UPDATE magic_cut_tasks SET status = ?, progress = 0, error_code = NULL, error_message = NULL, updated_at = ? WHERE id = ?`
const updateTaskProgressSQL = `UPDATE magic_cut_tasks SET progress = ?, updated_at = ? WHERE id = ? AND status = ?`
const updateExportSuccessSQL = `UPDATE magic_cut_exports SET storage_key = ?, media_metadata_json = ?, sha256 = ? WHERE id = ?`
const updateTaskSuccessSQL = `UPDATE magic_cut_tasks SET status = ?, progress = ?, output_json = ?, error_code = NULL, error_message = NULL, updated_at = ? WHERE id = ?`
const finishAttemptSQL = `UPDATE magic_cut_task_attempts SET status = ?, finished_at = ?, error_code = ?, error_message = ? WHERE task_id = ? AND status = ? ORDER BY attempt_number DESC LIMIT 1`
const updateTaskFailureSQL = `UPDATE magic_cut_tasks SET status = ?, error_code = ?, error_message = ?, updated_at = ? WHERE id = ?`
const selectOwnedTaskStatusSQL = `SELECT status FROM magic_cut_tasks WHERE id = ? AND owner_username = ? AND task_type = ? LIMIT 1`
const selectOwnedExportSQL = `SELECT id, project_id, task_id, owner_username, output_name, storage_key, media_metadata_json, sha256, created_at FROM magic_cut_exports WHERE id = ? AND owner_username = ? LIMIT 1`

type SQLRepository struct{ db *sql.DB }

func NewSQLRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) Create(ctx context.Context, item Export, queued worktask.Task) error {
	if r == nil || r.db == nil {
		return errors.New("export repository is not configured")
	}
	input, err := json.Marshal(queued.Input)
	if err != nil {
		return err
	}
	output, err := json.Marshal(queued.Output)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, insertExportTaskSQL, queued.ID, nullable(queued.ProjectID), queued.Owner, queued.Type, queued.Status, queued.Progress, input, output, nullable(queued.ErrorCode), nullable(queued.ErrorMessage), queued.CreatedAt, queued.UpdatedAt); err != nil {
		return fmt.Errorf("insert export task: %w", err)
	}
	if _, err = tx.ExecContext(ctx, insertExportSQL, item.ID, item.ProjectID, item.TaskID, item.Owner, item.OutputName, item.StorageKey, nil, nil, item.CreatedAt); err != nil {
		return fmt.Errorf("insert export: %w", err)
	}
	return tx.Commit()
}

func (r *SQLRepository) Start(ctx context.Context, taskID string, now time.Time) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var previous int
	if err = tx.QueryRowContext(ctx, selectAttemptNumberSQL, taskID).Scan(&previous); err != nil {
		return 0, err
	}
	attempt := previous + 1
	if _, err = tx.ExecContext(ctx, insertAttemptSQL, taskID, attempt, worktask.StatusRunning, now); err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, updateTaskRunningSQL, worktask.StatusRunning, now, taskID); err != nil {
		return 0, err
	}
	return attempt, tx.Commit()
}

func (r *SQLRepository) Progress(ctx context.Context, taskID string, progress float64, now time.Time) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 0.99 {
		progress = 0.99
	}
	_, err := r.db.ExecContext(ctx, updateTaskProgressSQL, progress, now, taskID, worktask.StatusRunning)
	return err
}

func (r *SQLRepository) Succeed(ctx context.Context, exportID, taskID string, result Result, metadata material.Metadata, now time.Time) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	outputJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, updateExportSuccessSQL, result.OutputPath, metadataJSON, result.SHA256, exportID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, updateTaskSuccessSQL, worktask.StatusSucceeded, float64(1), outputJSON, now, taskID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, finishAttemptSQL, worktask.StatusSucceeded, now, nil, nil, taskID, worktask.StatusRunning); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLRepository) Fail(ctx context.Context, taskID string, status worktask.Status, code, message string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, updateTaskFailureSQL, status, nullable(code), nullable(message), now, taskID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, finishAttemptSQL, status, now, nullable(code), nullable(message), taskID, worktask.StatusRunning); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLRepository) TaskStatus(ctx context.Context, owner, taskID string) (worktask.Status, error) {
	var status worktask.Status
	err := r.db.QueryRowContext(ctx, selectOwnedTaskStatusSQL, taskID, owner, "export_video").Scan(&status)
	return status, err
}

func (r *SQLRepository) Get(ctx context.Context, owner, exportID string) (Export, error) {
	var item Export
	var metadata []byte
	var storageKey, sha sql.NullString
	err := r.db.QueryRowContext(ctx, selectOwnedExportSQL, exportID, owner).Scan(&item.ID, &item.ProjectID, &item.TaskID, &item.Owner, &item.OutputName, &storageKey, &metadata, &sha, &item.CreatedAt)
	if err != nil {
		return Export{}, err
	}
	item.StorageKey, item.SHA256 = storageKey.String, sha.String
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &item.Metadata)
	}
	return item, nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

var _ Repository = (*SQLRepository)(nil)
