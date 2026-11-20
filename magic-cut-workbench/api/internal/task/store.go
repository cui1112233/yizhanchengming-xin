package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

const listTasksSQL = `SELECT id, project_id, owner_username, task_type, status, progress, input_json, output_json, error_code, error_message, created_at, updated_at FROM magic_cut_tasks WHERE owner_username = ? ORDER BY updated_at DESC LIMIT ?`
const saveTaskSQL = `INSERT INTO magic_cut_tasks (id, project_id, owner_username, task_type, status, progress, input_json, output_json, error_code, error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE status = VALUES(status), progress = VALUES(progress), output_json = VALUES(output_json), error_code = VALUES(error_code), error_message = VALUES(error_message), updated_at = VALUES(updated_at)`
const recoverableTasksSQL = `SELECT id, project_id, owner_username, task_type, status, progress, input_json, output_json, error_code, error_message, created_at, updated_at FROM magic_cut_tasks WHERE status IN (?, ?) ORDER BY updated_at ASC`
const cancelTaskSQL = `UPDATE magic_cut_tasks SET status = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND status IN (?, ?)`
const retryTaskSQL = `UPDATE magic_cut_tasks SET status = ?, progress = 0, error_code = NULL, error_message = NULL, updated_at = ? WHERE id = ? AND owner_username = ? AND status IN (?, ?)`

type StoreError struct {
	Code string
	Err  error
}

func (e *StoreError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}
func (e *StoreError) Unwrap() error { return e.Err }
func ErrorCode(err error) string {
	var typed *StoreError
	if errors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

func NewSQLStore(db *sql.DB, now func() time.Time) *SQLStore {
	if now == nil {
		now = time.Now
	}
	return &SQLStore{db: db, now: now}
}

func (s *SQLStore) List(ctx context.Context, owner string, limit int) ([]Task, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("task store is not configured")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, listTasksSQL, owner, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (s *SQLStore) Save(ctx context.Context, item Task) error {
	input, _ := json.Marshal(item.Input)
	output, _ := json.Marshal(item.Output)
	_, err := s.db.ExecContext(ctx, saveTaskSQL, item.ID, nullableString(item.ProjectID), item.Owner, item.Type, item.Status, item.Progress, input, output, nullableString(item.ErrorCode), nullableString(item.ErrorMessage), item.CreatedAt, item.UpdatedAt)
	return err
}

func (s *SQLStore) Recoverable(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, recoverableTasksSQL, StatusPending, StatusRunning)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (s *SQLStore) Cancel(ctx context.Context, owner, id string) error {
	result, err := s.db.ExecContext(ctx, cancelTaskSQL, StatusCanceled, s.now().UTC(), id, owner, StatusPending, StatusRunning)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return &StoreError{Code: "TASK_NOT_CANCELABLE", Err: errors.New("任务已结束，无法取消")}
	}
	return nil
}

func (s *SQLStore) Retry(ctx context.Context, owner, id string) error {
	result, err := s.db.ExecContext(ctx, retryTaskSQL, StatusPending, s.now().UTC(), id, owner, StatusFailed, StatusCanceled)
	if err != nil {
		return err
	}
	changed, _ := result.RowsAffected()
	if changed == 0 {
		return &StoreError{Code: "TASK_NOT_RETRYABLE", Err: errors.New("只有失败或已取消的任务可以重试")}
	}
	return nil
}

type rowScanner interface{ Scan(...any) error }
type rowsScanner interface {
	Next() bool
	Scan(...any) error
	Err() error
}

func scanTasks(rows rowsScanner) ([]Task, error) {
	items := make([]Task, 0)
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanTask(row rowScanner) (Task, error) {
	var item Task
	var projectID, errorCode, errorMessage sql.NullString
	var input, output []byte
	err := row.Scan(&item.ID, &projectID, &item.Owner, &item.Type, &item.Status, &item.Progress, &input, &output, &errorCode, &errorMessage, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Task{}, err
	}
	item.ProjectID, item.ErrorCode, item.ErrorMessage = projectID.String, errorCode.String, errorMessage.String
	if len(input) > 0 {
		_ = json.Unmarshal(input, &item.Input)
	}
	if len(output) > 0 {
		_ = json.Unmarshal(output, &item.Output)
	}
	return item, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

var _ Store = (*SQLStore)(nil)
