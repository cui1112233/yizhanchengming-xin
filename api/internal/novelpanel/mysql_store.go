package novelpanel

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) GetWorkspace(ctx context.Context, projectID int64) (Workspace, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT workspace_json FROM novel_panel_workspaces WHERE batch_project_id=?`, projectID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return Workspace{ProjectID: projectID}, nil
	}
	if err != nil {
		return Workspace{}, err
	}
	var value Workspace
	if err := json.Unmarshal(raw, &value); err != nil {
		return Workspace{}, err
	}
	value.ProjectID = projectID
	return value, nil
}

func (s *MySQLStore) SaveWorkspace(ctx context.Context, expected int64, workspace Workspace, note string) (Workspace, HistoryRecord, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	defer tx.Rollback()
	var current int64
	err = tx.QueryRowContext(ctx, `SELECT revision FROM novel_panel_workspaces WHERE batch_project_id=? FOR UPDATE`, workspace.ProjectID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		current = 0
	} else if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	if current != expected {
		return Workspace{}, HistoryRecord{}, fmt.Errorf("%w: expected %d, current %d", ErrConflict, expected, current)
	}
	now := time.Now().UTC()
	workspace.Revision = current + 1
	workspace.UpdatedAt = now
	if workspace.CreatedAt.IsZero() {
		workspace.CreatedAt = now
	}
	raw, err := json.Marshal(workspace)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO novel_panel_workspaces(batch_project_id,revision,workspace_json,created_at,updated_at) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE revision=VALUES(revision),workspace_json=VALUES(workspace_json),updated_at=VALUES(updated_at)`, workspace.ProjectID, workspace.Revision, raw, workspace.CreatedAt, now)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	history := HistoryRecord{ID: fmt.Sprintf("hist_%d_%d_%d", workspace.ProjectID, workspace.Revision, now.UnixNano()), ProjectID: workspace.ProjectID, Revision: workspace.Revision, Note: strings.TrimSpace(note), Summary: SummarizeWorkspace(workspace), Workspace: workspace, CreatedAt: now}
	summary, _ := json.Marshal(history.Summary)
	snapshot, _ := json.Marshal(history.Workspace)
	_, err = tx.ExecContext(ctx, `INSERT INTO novel_panel_history(id,batch_project_id,revision,note,summary_json,workspace_json,created_at) VALUES(?,?,?,?,?,?,?)`, history.ID, history.ProjectID, history.Revision, history.Note, summary, snapshot, history.CreatedAt)
	if err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	if err = tx.Commit(); err != nil {
		return Workspace{}, HistoryRecord{}, err
	}
	return workspace, history, nil
}
func (s *MySQLStore) ListHistory(ctx context.Context, projectID int64, limit int) ([]HistoryRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,revision,note,summary_json,workspace_json,created_at FROM novel_panel_history WHERE batch_project_id=? ORDER BY created_at DESC LIMIT ?`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HistoryRecord
	for rows.Next() {
		var item HistoryRecord
		var summary, workspace []byte
		if err := rows.Scan(&item.ID, &item.Revision, &item.Note, &summary, &workspace, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.ProjectID = projectID
		if err := json.Unmarshal(summary, &item.Summary); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(workspace, &item.Workspace); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *MySQLStore) GetHistory(ctx context.Context, projectID int64, id string) (HistoryRecord, error) {
	var item HistoryRecord
	var summary, workspace []byte
	err := s.db.QueryRowContext(ctx, `SELECT revision,note,summary_json,workspace_json,created_at FROM novel_panel_history WHERE batch_project_id=? AND id=?`, projectID, id).Scan(&item.Revision, &item.Note, &summary, &workspace, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return item, ErrNotFound
	}
	if err != nil {
		return item, err
	}
	item.ID = id
	item.ProjectID = projectID
	if err = json.Unmarshal(summary, &item.Summary); err != nil {
		return item, err
	}
	if err = json.Unmarshal(workspace, &item.Workspace); err != nil {
		return item, err
	}
	return item, nil
}
