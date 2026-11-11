package agentproxy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

const workspaceColumns = `id, name, description, permission_mode, created_at, updated_at`

func normalizeCreatePermissionMode(value string) (string, error) {
	mode := NormalizePermissionMode(value)
	if mode == "" {
		return PermissionPlan, nil
	}
	if mode == PermissionBypass {
		return "", ErrPermissionModeUnavailable
	}
	if !ValidPermissionMode(mode) {
		return "", ErrInvalidInput
	}
	return mode, nil
}

func (s *MySQLStore) ListWorkspaces(ctx context.Context, owner string) ([]Workspace, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT w.id, w.name, w.description, w.permission_mode, w.created_at, w.updated_at,
 (SELECT COUNT(*) FROM agent_proxy_sessions s WHERE s.owner_username = w.owner_username AND s.workspace_id = w.id AND s.deleted_at IS NULL)
FROM agent_proxy_workspaces w WHERE w.owner_username = ? AND w.deleted_at IS NULL ORDER BY w.updated_at DESC, w.id DESC`, strings.TrimSpace(owner))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Workspace, 0)
	for rows.Next() {
		var item Workspace
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.PermissionMode, &item.CreatedAt, &item.UpdatedAt, &item.SessionCount); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) Workspace(ctx context.Context, owner, id string) (Workspace, error) {
	var item Workspace
	err := s.db.QueryRowContext(ctx, `SELECT `+workspaceColumns+`,
 (SELECT COUNT(*) FROM agent_proxy_sessions s WHERE s.owner_username = w.owner_username AND s.workspace_id = w.id AND s.deleted_at IS NULL)
FROM agent_proxy_workspaces w WHERE w.id = ? AND w.owner_username = ? AND w.deleted_at IS NULL`, strings.TrimSpace(id), strings.TrimSpace(owner)).
		Scan(&item.ID, &item.Name, &item.Description, &item.PermissionMode, &item.CreatedAt, &item.UpdatedAt, &item.SessionCount)
	if errors.Is(err, sql.ErrNoRows) {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	return item, nil
}

func (s *MySQLStore) CreateWorkspace(ctx context.Context, owner string, input CreateWorkspaceInput, now time.Time) (Workspace, error) {
	owner = strings.TrimSpace(owner)
	id := strings.TrimSpace(input.ID)
	name := strings.TrimSpace(input.Name)
	if owner == "" || id == "" || name == "" || now.IsZero() {
		return Workspace{}, ErrInvalidInput
	}
	mode, err := normalizeCreatePermissionMode(input.PermissionMode)
	if err != nil {
		return Workspace{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO agent_proxy_workspaces (id, owner_username, name, description, permission_mode, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, owner, name, strings.TrimSpace(input.Description), mode, now.UTC(), now.UTC())
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return Workspace{}, ErrWorkspaceExists
	}
	if err != nil {
		return Workspace{}, err
	}
	return s.Workspace(ctx, owner, id)
}

func (s *MySQLStore) UpdateWorkspace(ctx context.Context, owner, id string, input UpdateWorkspaceInput, now time.Time) (Workspace, error) {
	current, err := s.Workspace(ctx, owner, id)
	if err != nil {
		return Workspace{}, err
	}
	name := current.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
		if name == "" {
			return Workspace{}, ErrInvalidInput
		}
	}
	description := current.Description
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}
	mode := current.PermissionMode
	if input.PermissionMode != nil {
		if mode, err = normalizeCreatePermissionMode(*input.PermissionMode); err != nil {
			return Workspace{}, err
		}
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE agent_proxy_workspaces SET name = ?, description = ?, permission_mode = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`,
		name, description, mode, now.UTC(), current.ID, strings.TrimSpace(owner)); err != nil {
		return Workspace{}, err
	}
	return s.Workspace(ctx, owner, id)
}

func (s *MySQLStore) DeleteWorkspace(ctx context.Context, owner, id string, now time.Time) error {
	owner = strings.TrimSpace(owner)
	id = strings.TrimSpace(id)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE agent_proxy_workspaces SET deleted_at = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`, now.UTC(), now.UTC(), id, owner)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return ErrWorkspaceNotFound
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_proxy_sessions SET deleted_at = ?, updated_at = ? WHERE workspace_id = ? AND owner_username = ? AND deleted_at IS NULL`, now.UTC(), now.UTC(), id, owner); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

const sessionColumns = `id, workspace_id, title, status, permission_mode, last_message_at, created_at, updated_at`

type sessionScanner interface {
	Scan(...any) error
}

func scanSession(scanner sessionScanner) (Session, error) {
	var item Session
	var lastMessageAt sql.NullTime
	if err := scanner.Scan(&item.ID, &item.WorkspaceID, &item.Title, &item.Status, &item.PermissionMode, &lastMessageAt, &item.CreatedAt, &item.UpdatedAt, &item.MessageCount); err != nil {
		return Session{}, err
	}
	if lastMessageAt.Valid {
		value := lastMessageAt.Time
		item.LastMessageAt = &value
	}
	return item, nil
}

func (s *MySQLStore) ListSessions(ctx context.Context, owner, workspaceID string) ([]Session, error) {
	owner = strings.TrimSpace(owner)
	rows, err := s.db.QueryContext(ctx, `SELECT s.id, s.workspace_id, s.title, s.status, s.permission_mode, s.last_message_at, s.created_at, s.updated_at,
 (SELECT COUNT(*) FROM agent_proxy_messages m WHERE m.owner_username = s.owner_username AND m.session_id = s.id)
FROM agent_proxy_sessions s WHERE s.owner_username = ? AND s.workspace_id = ? AND s.deleted_at IS NULL ORDER BY s.updated_at DESC, s.id DESC`, owner, strings.TrimSpace(workspaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Session, 0)
	for rows.Next() {
		item, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) Session(ctx context.Context, owner, id string) (Session, error) {
	row := s.db.QueryRowContext(ctx, `SELECT s.id, s.workspace_id, s.title, s.status, s.permission_mode, s.last_message_at, s.created_at, s.updated_at,
 (SELECT COUNT(*) FROM agent_proxy_messages m WHERE m.owner_username = s.owner_username AND m.session_id = s.id)
FROM agent_proxy_sessions s WHERE s.id = ? AND s.owner_username = ? AND s.deleted_at IS NULL`, strings.TrimSpace(id), strings.TrimSpace(owner))
	item, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return item, err
}

func (s *MySQLStore) CreateSession(ctx context.Context, owner, workspaceID string, input CreateSessionInput, now time.Time) (Session, error) {
	owner = strings.TrimSpace(owner)
	workspace, err := s.Workspace(ctx, owner, workspaceID)
	if err != nil {
		return Session{}, err
	}
	id := strings.TrimSpace(input.ID)
	title := strings.TrimSpace(input.Title)
	if id == "" || title == "" || now.IsZero() {
		return Session{}, ErrInvalidInput
	}
	mode := workspace.PermissionMode
	if strings.TrimSpace(input.PermissionMode) != "" {
		if mode, err = normalizeCreatePermissionMode(input.PermissionMode); err != nil {
			return Session{}, err
		}
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO agent_proxy_sessions (id, workspace_id, owner_username, title, status, permission_mode, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, workspace.ID, owner, title, SessionStatusActive, mode, now.UTC(), now.UTC())
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return Session{}, ErrSessionExists
	}
	if err != nil {
		return Session{}, err
	}
	return s.Session(ctx, owner, id)
}

func (s *MySQLStore) UpdateSession(ctx context.Context, owner, id string, input UpdateSessionInput, now time.Time) (Session, error) {
	current, err := s.Session(ctx, owner, id)
	if err != nil {
		return Session{}, err
	}
	title := current.Title
	if input.Title != nil {
		title = strings.TrimSpace(*input.Title)
		if title == "" {
			return Session{}, ErrInvalidInput
		}
	}
	status := current.Status
	if input.Status != nil {
		status = NormalizeSessionStatus(*input.Status)
		if !ValidSessionStatus(status) {
			return Session{}, ErrInvalidInput
		}
	}
	mode := current.PermissionMode
	if input.PermissionMode != nil {
		if mode, err = normalizeCreatePermissionMode(*input.PermissionMode); err != nil {
			return Session{}, err
		}
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE agent_proxy_sessions SET title = ?, status = ?, permission_mode = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`,
		title, status, mode, now.UTC(), current.ID, strings.TrimSpace(owner)); err != nil {
		return Session{}, err
	}
	return s.Session(ctx, owner, id)
}

func (s *MySQLStore) DeleteSession(ctx context.Context, owner, id string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE agent_proxy_sessions SET deleted_at = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND deleted_at IS NULL`,
		now.UTC(), now.UTC(), strings.TrimSpace(id), strings.TrimSpace(owner))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (s *MySQLStore) ListMessages(ctx context.Context, owner, sessionID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, role, content, metadata_json, created_at FROM agent_proxy_messages
WHERE owner_username = ? AND session_id = ? ORDER BY id DESC LIMIT ?`, strings.TrimSpace(owner), strings.TrimSpace(sessionID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Message, 0)
	for rows.Next() {
		var item Message
		var metadata []byte
		if err := rows.Scan(&item.ID, &item.Role, &item.Content, &metadata, &item.CreatedAt); err != nil {
			return nil, err
		}
		if len(metadata) > 0 && string(metadata) != "null" {
			if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
				return nil, err
			}
		}
		item.SessionID = strings.TrimSpace(sessionID)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (s *MySQLStore) AppendMessage(ctx context.Context, owner, sessionID string, input AppendMessageInput, now time.Time) (Message, error) {
	owner = strings.TrimSpace(owner)
	sessionID = strings.TrimSpace(sessionID)
	role := strings.ToLower(strings.TrimSpace(input.Role))
	content := strings.TrimSpace(input.Content)
	if owner == "" || sessionID == "" || content == "" || now.IsZero() || !ValidMessageRole(role) {
		return Message{}, ErrInvalidInput
	}
	var metadata any
	if input.Metadata != nil {
		raw, err := json.Marshal(input.Metadata)
		if err != nil {
			return Message{}, ErrInvalidInput
		}
		metadata = string(raw)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, err
	}
	var workspaceID string
	var storedSessionID string
	if err := tx.QueryRowContext(ctx, `SELECT id, workspace_id FROM agent_proxy_sessions WHERE id = ? AND owner_username = ? AND deleted_at IS NULL FOR UPDATE`, sessionID, owner).
		Scan(&storedSessionID, &workspaceID); errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return Message{}, ErrSessionNotFound
	} else if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO agent_proxy_messages (session_id, workspace_id, owner_username, role, content, metadata_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sessionID, workspaceID, owner, role, content, metadata, now.UTC())
	if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_proxy_sessions SET last_message_at = ?, updated_at = ? WHERE id = ? AND owner_username = ?`, now.UTC(), now.UTC(), sessionID, owner); err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	messageID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}
	return Message{ID: messageID, SessionID: sessionID, Role: role, Content: content, Metadata: input.Metadata, CreatedAt: now.UTC()}, nil
}

func (s *MySQLStore) ListSkills(ctx context.Context, owner, workspaceID string) ([]Skill, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, workspace_id, skill_key, display_name, description, source, status, created_at, updated_at
FROM agent_proxy_skills WHERE owner_username = ? AND workspace_id = ? ORDER BY updated_at DESC, id DESC`, strings.TrimSpace(owner), strings.TrimSpace(workspaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Skill, 0)
	for rows.Next() {
		var item Skill
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.SkillKey, &item.DisplayName, &item.Description, &item.Source, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *MySQLStore) ListScheduledJobs(ctx context.Context, owner, workspaceID string) ([]ScheduledJob, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, workspace_id, name, schedule, prompt, permission_mode, enabled, last_run_at, next_run_at, created_at, updated_at
FROM agent_proxy_scheduled_jobs WHERE owner_username = ? AND workspace_id = ? ORDER BY updated_at DESC, id DESC`, strings.TrimSpace(owner), strings.TrimSpace(workspaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ScheduledJob, 0)
	for rows.Next() {
		var item ScheduledJob
		var lastRunAt, nextRunAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.Name, &item.Schedule, &item.Prompt, &item.PermissionMode, &item.Enabled, &lastRunAt, &nextRunAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if lastRunAt.Valid {
			value := lastRunAt.Time
			item.LastRunAt = &value
		}
		if nextRunAt.Valid {
			value := nextRunAt.Time
			item.NextRunAt = &value
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
