package novelcreation

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// MySQLStore is the production persistence implementation for novel creation.
// It deliberately does not share state with MemoryStore: a restarted process
// reads all user data back from MySQL.
type MySQLStore struct{ db sqlExecutor }

func NewMySQLStore(db *sql.DB) Store { return &MySQLStore{db: sqlDatabase{db}} }

func novelID(prefix string) string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(bytes)
}

func requiredOwner(owner string) (string, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return "", &DomainError{Code: ErrInputRequired, Message: "owner is required"}
	}
	return owner, nil
}

func persistenceError(err error) error {
	if err == nil {
		return nil
	}
	return &DomainError{Code: ErrPersistenceFailed, Message: "novel persistence failed", Cause: err}
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableRawJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return []byte(value)
}

func (s *MySQLStore) CreateProject(ctx context.Context, owner string, input CreateProjectInput) (Project, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Project{}, err
	}
	input, err = NormalizeCreateProjectInput(input)
	if err != nil {
		return Project{}, err
	}
	if input.Title == "" || input.Genre == "" || input.ModelID == "" {
		return Project{}, &DomainError{Code: ErrInputRequired, Message: "title, genre and modelId are required"}
	}
	now := time.Now().UTC()
	project := Project{ID: novelID("project"), Title: input.Title, Genre: input.Genre, World: input.World, Protagonist: input.Protagonist, CoreConflict: input.CoreConflict, ModelID: input.ModelID, TargetChapterWords: input.TargetChapterWords, Status: "draft", CreatedAt: now, UpdatedAt: now}
	_, err = s.db.ExecContext(ctx, `INSERT INTO novel_projects
(project_id, owner_username, title, genre, world, protagonist, core_conflict, model_id, model_policy_json, target_chapter_words, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, project.ID, owner, project.Title, project.Genre, nullableString(project.World), nullableString(project.Protagonist), nullableString(project.CoreConflict), project.ModelID, []byte(`{}`), project.TargetChapterWords, project.Status, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return Project{}, persistenceError(err)
	}
	return project, nil
}

const projectColumns = `project_id, title, genre, world, protagonist, core_conflict, model_id, target_chapter_words, status, created_at, updated_at, deleted_at`

type novelRowScanner interface{ Scan(...any) error }

func scanProject(scanner novelRowScanner) (Project, error) {
	var project Project
	var world, protagonist, coreConflict sql.NullString
	var deletedAt sql.NullTime
	err := scanner.Scan(&project.ID, &project.Title, &project.Genre, &world, &protagonist, &coreConflict, &project.ModelID, &project.TargetChapterWords, &project.Status, &project.CreatedAt, &project.UpdatedAt, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	if err != nil {
		return Project{}, persistenceError(err)
	}
	project.World, project.Protagonist, project.CoreConflict = world.String, protagonist.String, coreConflict.String
	if deletedAt.Valid {
		when := deletedAt.Time
		project.DeletedAt = &when
	}
	return project, nil
}

func (s *MySQLStore) GetProject(ctx context.Context, owner, projectID string) (Project, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Project{}, err
	}
	return scanProject(s.db.QueryRowContext(ctx, `SELECT `+projectColumns+`
FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL`, owner, strings.TrimSpace(projectID)))
}

func (s *MySQLStore) ListProjects(ctx context.Context, owner string) ([]Project, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+projectColumns+`
FROM novel_projects WHERE owner_username = ? AND deleted_at IS NULL ORDER BY updated_at DESC`, owner)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return projects, nil
}

func (s *MySQLStore) CreateSession(ctx context.Context, owner string, input CreateSessionInput) (Session, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Session{}, err
	}
	input.ProjectID, input.Mode, input.Title = strings.TrimSpace(input.ProjectID), strings.TrimSpace(input.Mode), strings.TrimSpace(input.Title)
	if input.ProjectID == "" || input.Mode == "" {
		return Session{}, &DomainError{Code: ErrInputRequired, Message: "projectId and mode are required"}
	}
	now := time.Now().UTC()
	session := Session{ID: novelID("session"), ProjectID: input.ProjectID, Mode: input.Mode, Title: input.Title, ContextBudget: input.ContextBudget, CreatedAt: now, UpdatedAt: now}
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_sessions
(session_id, owner_username, project_id, mode, title, context_budget, message_cursor, current_run_id, created_at, updated_at)
SELECT ?, ?, ?, ?, ?, ?, 0, NULL, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)`, session.ID, owner, session.ProjectID, session.Mode, session.Title, session.ContextBudget, session.CreatedAt, session.UpdatedAt, owner, session.ProjectID)
	if err != nil {
		return Session{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Session{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	return session, nil
}

const sessionColumns = `s.session_id, s.project_id, s.mode, s.title, s.context_budget, s.message_cursor, COALESCE(s.current_run_id, ''), s.created_at, s.updated_at`

func scanSession(scanner novelRowScanner) (Session, error) {
	var session Session
	err := scanner.Scan(&session.ID, &session.ProjectID, &session.Mode, &session.Title, &session.ContextBudget, &session.MessageCursor, &session.CurrentRunID, &session.CreatedAt, &session.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	if err != nil {
		return Session{}, persistenceError(err)
	}
	return session, nil
}

func (s *MySQLStore) GetSession(ctx context.Context, owner, sessionID string) (Session, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Session{}, err
	}
	return scanSession(s.db.QueryRowContext(ctx, `SELECT `+sessionColumns+`
FROM novel_sessions s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.owner_username = ? AND s.session_id = ? AND p.deleted_at IS NULL`, owner, strings.TrimSpace(sessionID)))
}

func (s *MySQLStore) ListSessions(ctx context.Context, owner, projectID string) ([]Session, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	projectID = strings.TrimSpace(projectID)
	query := `SELECT ` + sessionColumns + `
FROM novel_sessions s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.owner_username = ? AND p.deleted_at IS NULL`
	args := []any{owner}
	if projectID != "" {
		query += ` AND s.project_id = ?`
		args = append(args, projectID)
	}
	query += ` ORDER BY s.updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	sessions := make([]Session, 0)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return sessions, nil
}

func (s *MySQLStore) AppendMessage(ctx context.Context, owner string, message Message) (Message, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Message{}, err
	}
	if err := validateRawJSON("message references", message.References); err != nil {
		return Message{}, err
	}
	if err := validateRawJSON("message token usage", message.TokenUsage); err != nil {
		return Message{}, err
	}
	message.ID, message.SessionID, message.ProjectID = strings.TrimSpace(message.ID), strings.TrimSpace(message.SessionID), strings.TrimSpace(message.ProjectID)
	message.Role, message.Content = strings.TrimSpace(message.Role), strings.TrimSpace(message.Content)
	if message.SessionID == "" || message.ProjectID == "" || message.Role == "" || message.Content == "" {
		return Message{}, &DomainError{Code: ErrInputRequired, Message: "message sessionId, projectId, role and content are required"}
	}
	if message.ID == "" {
		message.ID = novelID("message")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, persistenceError(err)
	}
	defer tx.Rollback()
	var cursor int64
	err = tx.QueryRowContext(ctx, `SELECT s.message_cursor FROM novel_sessions s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.owner_username = ? AND s.session_id = ? AND s.project_id = ? AND p.deleted_at IS NULL FOR UPDATE`, owner, message.SessionID, message.ProjectID).Scan(&cursor)
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, &DomainError{Code: ErrSessionNotFound, Message: "session not found"}
	}
	if err != nil {
		return Message{}, persistenceError(err)
	}
	message.Sequence = cursor + 1
	message.CreatedAt, message.UpdatedAt = time.Now().UTC(), time.Now().UTC()
	_, err = tx.ExecContext(ctx, `INSERT INTO novel_messages
(message_id, owner_username, session_id, project_id, sequence_no, role, content, status, model_id, thinking_level, references_json, token_usage_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, message.ID, owner, message.SessionID, message.ProjectID, message.Sequence, message.Role, message.Content, message.Status, message.ModelID, message.ThinkingLevel, nullableRawJSON(message.References), nullableRawJSON(message.TokenUsage), message.CreatedAt, message.UpdatedAt)
	if err != nil {
		return Message{}, persistenceError(err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE novel_sessions SET message_cursor = ?, updated_at = ? WHERE owner_username = ? AND session_id = ? AND project_id = ?`, message.Sequence, message.UpdatedAt, owner, message.SessionID, message.ProjectID)
	if err != nil {
		return Message{}, persistenceError(err)
	}
	if err := tx.Commit(); err != nil {
		return Message{}, persistenceError(err)
	}
	return message, nil
}

func (s *MySQLStore) ListMessages(ctx context.Context, owner, sessionID string, after int64) ([]Message, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT m.message_id, m.session_id, m.project_id, m.role, m.content, m.status, m.model_id, m.thinking_level, m.references_json, m.token_usage_json, m.sequence_no, m.created_at, m.updated_at
FROM novel_messages m JOIN novel_projects p ON p.owner_username = m.owner_username AND p.project_id = m.project_id
WHERE m.owner_username = ? AND m.session_id = ? AND m.sequence_no > ? AND p.deleted_at IS NULL ORDER BY m.sequence_no`, owner, strings.TrimSpace(sessionID), after)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	messages := make([]Message, 0)
	for rows.Next() {
		var message Message
		var references, usage []byte
		if err := rows.Scan(&message.ID, &message.SessionID, &message.ProjectID, &message.Role, &message.Content, &message.Status, &message.ModelID, &message.ThinkingLevel, &references, &usage, &message.Sequence, &message.CreatedAt, &message.UpdatedAt); err != nil {
			return nil, persistenceError(err)
		}
		message.References, message.TokenUsage = cloneRawJSON(references), cloneRawJSON(usage)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return messages, nil
}

const runColumns = `run_id, project_id, session_id, idempotency_key, status, model_id, thinking_level, prompt_snapshot_json, runtime_snapshot_json, error_code, COALESCE(error_message, ''), created_at, updated_at, finished_at`

const qualifiedRunColumns = `r.run_id, r.project_id, r.session_id, r.idempotency_key, r.status, r.model_id, r.thinking_level, r.prompt_snapshot_json, r.runtime_snapshot_json, r.error_code, COALESCE(r.error_message, ''), r.created_at, r.updated_at, r.finished_at`

func scanRun(scanner novelRowScanner) (Run, error) {
	var run Run
	var prompt, runtime []byte
	var finishedAt sql.NullTime
	err := scanner.Scan(&run.ID, &run.ProjectID, &run.SessionID, &run.IdempotencyKey, &run.Status, &run.ModelID, &run.ThinkingLevel, &prompt, &runtime, &run.ErrorCode, &run.ErrorMessage, &run.CreatedAt, &run.UpdatedAt, &finishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	if err != nil {
		return Run{}, persistenceError(err)
	}
	run.PromptSnapshot, run.RuntimeSnapshot = cloneRawJSON(prompt), cloneRawJSON(runtime)
	if finishedAt.Valid {
		when := finishedAt.Time
		run.FinishedAt = &when
	}
	return run, nil
}

func (s *MySQLStore) FindRunByIdempotencyKey(ctx context.Context, owner, idempotencyKey string) (Run, bool, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Run{}, false, err
	}
	run, err := scanRun(s.db.QueryRowContext(ctx, `SELECT `+qualifiedRunColumns+`
FROM novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id
WHERE r.owner_username = ? AND r.idempotency_key = ? AND p.deleted_at IS NULL`, owner, strings.TrimSpace(idempotencyKey)))
	if IsCode(err, ErrProjectNotFound) {
		return Run{}, false, nil
	}
	return run, err == nil, err
}

func (s *MySQLStore) GetRun(ctx context.Context, owner, runID string) (Run, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Run{}, err
	}
	return s.runByID(ctx, owner, strings.TrimSpace(runID))
}

func (s *MySQLStore) CreateRun(ctx context.Context, owner string, run Run) (Run, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Run{}, err
	}
	if err := validateRawJSON("run prompt snapshot", run.PromptSnapshot); err != nil {
		return Run{}, err
	}
	if err := validateRawJSON("run runtime snapshot", run.RuntimeSnapshot); err != nil {
		return Run{}, err
	}
	run.ID, run.ProjectID, run.SessionID, run.IdempotencyKey = strings.TrimSpace(run.ID), strings.TrimSpace(run.ProjectID), strings.TrimSpace(run.SessionID), strings.TrimSpace(run.IdempotencyKey)
	if run.ProjectID == "" || run.SessionID == "" || run.IdempotencyKey == "" || run.Status == "" {
		return Run{}, &DomainError{Code: ErrInputRequired, Message: "run projectId, sessionId, idempotencyKey and status are required"}
	}
	if run.ID == "" {
		run.ID = novelID("run")
	}
	now := time.Now().UTC()
	run.CreatedAt, run.UpdatedAt = now, now
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_runs
(run_id, owner_username, project_id, session_id, idempotency_key, status, model_id, thinking_level, prompt_snapshot_json, runtime_snapshot_json, error_code, error_message, created_at, updated_at, finished_at)
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)
  AND EXISTS (SELECT 1 FROM novel_sessions WHERE owner_username = ? AND session_id = ? AND project_id = ?)`, run.ID, owner, run.ProjectID, run.SessionID, run.IdempotencyKey, run.Status, run.ModelID, run.ThinkingLevel, nullableRawJSON(run.PromptSnapshot), nullableRawJSON(run.RuntimeSnapshot), run.ErrorCode, nullableString(run.ErrorMessage), run.CreatedAt, run.UpdatedAt, run.FinishedAt, owner, run.ProjectID, owner, run.SessionID, run.ProjectID)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return Run{}, &DomainError{Code: ErrRunInProgress, Message: "idempotency key already has a run", Cause: err}
	}
	if err != nil {
		return Run{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "project or session not found"}
	}
	_, err = s.db.ExecContext(ctx, `UPDATE novel_sessions SET current_run_id = ?, updated_at = ? WHERE owner_username = ? AND session_id = ? AND project_id = ?`, run.ID, now, owner, run.SessionID, run.ProjectID)
	if err != nil {
		return Run{}, persistenceError(err)
	}
	return run, nil
}

func (s *MySQLStore) UpdateRun(ctx context.Context, owner string, update RunUpdate) (Run, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Run{}, err
	}
	if err := validateRawJSON("run runtime snapshot", update.RuntimeSnapshot); err != nil {
		return Run{}, err
	}
	update.ID = strings.TrimSpace(update.ID)
	if update.ID == "" {
		return Run{}, &DomainError{Code: ErrInputRequired, Message: "run id is required"}
	}
	sets, args := make([]string, 0, 5), make([]any, 0, 7)
	if update.Status != "" {
		sets, args = append(sets, "r.status = ?"), append(args, update.Status)
	}
	if update.RuntimeSnapshot != nil {
		sets, args = append(sets, "r.runtime_snapshot_json = ?"), append(args, nullableRawJSON(update.RuntimeSnapshot))
	}
	if update.ErrorCode != "" {
		sets, args = append(sets, "r.error_code = ?"), append(args, update.ErrorCode)
	}
	if update.ErrorMessage != "" {
		sets, args = append(sets, "r.error_message = ?"), append(args, update.ErrorMessage)
	}
	if update.FinishedAt != nil {
		sets, args = append(sets, "r.finished_at = ?"), append(args, update.FinishedAt.UTC())
	}
	sets, args = append(sets, "r.updated_at = ?"), append(args, time.Now().UTC(), owner, update.ID)
	result, err := s.db.ExecContext(ctx, `UPDATE novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id
SET `+strings.Join(sets, ", ")+` WHERE r.owner_username = ? AND r.run_id = ? AND p.deleted_at IS NULL`, args...)
	if err != nil {
		return Run{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Run{}, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	return s.runByID(ctx, owner, update.ID)
}

func (s *MySQLStore) runByID(ctx context.Context, owner, runID string) (Run, error) {
	return scanRun(s.db.QueryRowContext(ctx, `SELECT `+qualifiedRunColumns+`
FROM novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id
WHERE r.owner_username = ? AND r.run_id = ? AND p.deleted_at IS NULL`, owner, runID))
}

func (s *MySQLStore) AppendRunEvent(ctx context.Context, owner string, event RunEvent) (RunEvent, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return RunEvent{}, err
	}
	if err := validateRawJSON("run event payload", event.Payload); err != nil {
		return RunEvent{}, err
	}
	event.RunID, event.SessionID, event.ProjectID, event.Type = strings.TrimSpace(event.RunID), strings.TrimSpace(event.SessionID), strings.TrimSpace(event.ProjectID), strings.TrimSpace(event.Type)
	if event.RunID == "" || event.SessionID == "" || event.ProjectID == "" || event.Type == "" {
		return RunEvent{}, &DomainError{Code: ErrInputRequired, Message: "run event runId, sessionId, projectId and type are required"}
	}
	event.CreatedAt = time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_run_events
(owner_username, run_id, session_id, project_id, event_type, step_name, status, message, progress, payload_json, created_at)
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id WHERE r.owner_username = ? AND r.run_id = ? AND r.session_id = ? AND r.project_id = ? AND p.deleted_at IS NULL)`, owner, event.RunID, event.SessionID, event.ProjectID, event.Type, event.Step, event.Status, nullableString(event.Message), event.Progress, nullableRawJSON(event.Payload), event.CreatedAt, owner, event.RunID, event.SessionID, event.ProjectID)
	if err != nil {
		return RunEvent{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return RunEvent{}, &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
	}
	event.EventID, _ = result.LastInsertId()
	return event, nil
}

func (s *MySQLStore) ListRunEvents(ctx context.Context, owner, runID string, after int64) ([]RunEvent, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT e.event_id, e.run_id, e.session_id, e.project_id, e.event_type, e.step_name, e.status, COALESCE(e.message, ''), e.progress, e.payload_json, e.created_at
FROM novel_run_events e JOIN novel_projects p ON p.owner_username = e.owner_username AND p.project_id = e.project_id
WHERE e.owner_username = ? AND e.run_id = ? AND e.event_id > ? AND p.deleted_at IS NULL ORDER BY e.event_id`, owner, strings.TrimSpace(runID), after)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	events := make([]RunEvent, 0)
	for rows.Next() {
		var event RunEvent
		var payload []byte
		if err := rows.Scan(&event.EventID, &event.RunID, &event.SessionID, &event.ProjectID, &event.Type, &event.Step, &event.Status, &event.Message, &event.Progress, &payload, &event.CreatedAt); err != nil {
			return nil, persistenceError(err)
		}
		event.Payload = cloneRawJSON(payload)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return events, nil
}

func (s *MySQLStore) ListChapters(ctx context.Context, owner, projectID string) ([]Chapter, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT c.chapter_id, c.project_id, c.volume_id, c.number_no, c.title, c.content, c.word_count, c.audit_status, c.version_no, c.created_at, c.updated_at
FROM novel_chapters c
JOIN novel_projects p ON p.owner_username = c.owner_username AND p.project_id = c.project_id
WHERE c.owner_username = ? AND c.project_id = ? AND p.deleted_at IS NULL
ORDER BY c.number_no`, owner, strings.TrimSpace(projectID))
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	chapters := make([]Chapter, 0)
	for rows.Next() {
		var chapter Chapter
		var volumeID sql.NullString
		if err := rows.Scan(&chapter.ID, &chapter.ProjectID, &volumeID, &chapter.Number, &chapter.Title, &chapter.Content, &chapter.WordCount, &chapter.AuditStatus, &chapter.Version, &chapter.CreatedAt, &chapter.UpdatedAt); err != nil {
			return nil, persistenceError(err)
		}
		chapter.VolumeID = volumeID.String
		chapters = append(chapters, chapter)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return chapters, nil
}

func (s *MySQLStore) SaveChapterBundle(ctx context.Context, owner string, bundle ChapterBundle) error {
	owner, err := requiredOwner(owner)
	if err != nil {
		return err
	}
	if err := validateChapterBundleJSON(bundle); err != nil {
		return err
	}
	bundle.ProjectID, bundle.RunID = strings.TrimSpace(bundle.ProjectID), strings.TrimSpace(bundle.RunID)
	if bundle.ProjectID == "" || strings.TrimSpace(bundle.Chapter.ID) == "" || bundle.Chapter.Number <= 0 {
		return &DomainError{Code: ErrInputRequired, Message: "projectId, chapter id and chapter number are required"}
	}
	if err := validateChapterBundleBindings(bundle); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return persistenceError(err)
	}
	defer tx.Rollback()
	var one int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL FOR UPDATE`, owner, bundle.ProjectID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	if err != nil {
		return persistenceError(err)
	}
	if bundle.RunID != "" {
		err = tx.QueryRowContext(ctx, `SELECT 1 FROM novel_runs WHERE owner_username = ? AND run_id = ? AND project_id = ? FOR UPDATE`, owner, bundle.RunID, bundle.ProjectID).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return &DomainError{Code: ErrProjectNotFound, Message: "run not found"}
		}
		if err != nil {
			return persistenceError(err)
		}
	}
	now := time.Now().UTC()
	if bundle.Volume.ID != "" {
		if bundle.Volume.ProjectID == "" {
			bundle.Volume.ProjectID = bundle.ProjectID
		}
		if bundle.Volume.CreatedAt.IsZero() {
			bundle.Volume.CreatedAt = now
		}
		bundle.Volume.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `INSERT INTO novel_volumes (volume_id, owner_username, project_id, number_no, title, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE title = VALUES(title), updated_at = VALUES(updated_at)`, bundle.Volume.ID, owner, bundle.Volume.ProjectID, bundle.Volume.Number, bundle.Volume.Title, bundle.Volume.CreatedAt, bundle.Volume.UpdatedAt); err != nil {
			return persistenceError(err)
		}
	}
	chapter := bundle.Chapter
	chapter.ProjectID = bundle.ProjectID
	if chapter.CreatedAt.IsZero() {
		chapter.CreatedAt = now
	}
	chapter.UpdatedAt = now
	if chapter.Version == 0 {
		chapter.Version = 1
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO novel_chapters
(chapter_id, owner_username, project_id, volume_id, number_no, title, content, word_count, audit_status, version_no, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE volume_id = VALUES(volume_id), title = VALUES(title), content = VALUES(content), word_count = VALUES(word_count), audit_status = VALUES(audit_status), version_no = VALUES(version_no), updated_at = VALUES(updated_at)`, chapter.ID, owner, chapter.ProjectID, nullableString(chapter.VolumeID), chapter.Number, chapter.Title, chapter.Content, chapter.WordCount, chapter.AuditStatus, chapter.Version, chapter.CreatedAt, chapter.UpdatedAt); err != nil {
		return persistenceError(err)
	}
	if version := bundle.ChapterVersion; version.ID != "" || strings.TrimSpace(version.ChapterID) != "" || strings.TrimSpace(version.RunID) != "" {
		if version.ChapterID == "" {
			version.ChapterID = chapter.ID
		}
		if version.Version == 0 {
			version.Version = chapter.Version
		}
		if version.CreatedAt.IsZero() {
			version.CreatedAt = now
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO novel_chapter_versions (version_id, owner_username, project_id, chapter_id, run_id, version_no, content, snapshot_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, version.ID, owner, bundle.ProjectID, version.ChapterID, nullableString(version.RunID), version.Version, version.Content, nullableRawJSON(version.Snapshot), version.CreatedAt); err != nil {
			return persistenceError(err)
		}
	}
	if foundation := bundle.Foundation; foundation.ID != "" {
		if foundation.ProjectID == "" {
			foundation.ProjectID = bundle.ProjectID
		}
		if foundation.CreatedAt.IsZero() {
			foundation.CreatedAt = now
		}
		foundation.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `INSERT INTO novel_foundations (foundation_id, owner_username, project_id, snapshot_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE snapshot_json = VALUES(snapshot_json), updated_at = VALUES(updated_at)`, foundation.ID, owner, foundation.ProjectID, nullableRawJSON(foundation.Snapshot), foundation.CreatedAt, foundation.UpdatedAt); err != nil {
			return persistenceError(err)
		}
	}
	for _, memory := range bundle.Memories {
		if strings.TrimSpace(memory.ID) == "" {
			return &DomainError{Code: ErrInputRequired, Message: "memory id is required"}
		}
		if memory.ProjectID == "" {
			memory.ProjectID = bundle.ProjectID
		}
		if memory.ChapterID == "" {
			memory.ChapterID = chapter.ID
		}
		if memory.CreatedAt.IsZero() {
			memory.CreatedAt = now
		}
		memory.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `INSERT INTO novel_memories (memory_id, owner_username, project_id, chapter_id, snapshot_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE snapshot_json = VALUES(snapshot_json), updated_at = VALUES(updated_at)`, memory.ID, owner, memory.ProjectID, nullableString(memory.ChapterID), nullableRawJSON(memory.Snapshot), memory.CreatedAt, memory.UpdatedAt); err != nil {
			return persistenceError(err)
		}
	}
	for _, audit := range bundle.Audits {
		if strings.TrimSpace(audit.ID) == "" {
			return &DomainError{Code: ErrInputRequired, Message: "audit id is required"}
		}
		if audit.ProjectID == "" {
			audit.ProjectID = bundle.ProjectID
		}
		if audit.ChapterID == "" {
			audit.ChapterID = chapter.ID
		}
		if audit.CreatedAt.IsZero() {
			audit.CreatedAt = now
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO novel_audits (audit_id, owner_username, project_id, chapter_id, run_id, score, snapshot_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE score = VALUES(score), snapshot_json = VALUES(snapshot_json)`, audit.ID, owner, audit.ProjectID, audit.ChapterID, nullableString(audit.RunID), audit.Score, nullableRawJSON(audit.Snapshot), audit.CreatedAt); err != nil {
			return persistenceError(err)
		}
	}
	for _, revision := range bundle.Revisions {
		if strings.TrimSpace(revision.ID) == "" {
			return &DomainError{Code: ErrInputRequired, Message: "revision id is required"}
		}
		if revision.ProjectID == "" {
			revision.ProjectID = bundle.ProjectID
		}
		if revision.ChapterID == "" {
			revision.ChapterID = chapter.ID
		}
		if revision.CreatedAt.IsZero() {
			revision.CreatedAt = now
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO novel_revisions (revision_id, owner_username, project_id, chapter_id, run_id, strategy, content, snapshot_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE strategy = VALUES(strategy), content = VALUES(content), snapshot_json = VALUES(snapshot_json)`, revision.ID, owner, revision.ProjectID, revision.ChapterID, nullableString(revision.RunID), revision.Strategy, revision.Content, nullableRawJSON(revision.Snapshot), revision.CreatedAt); err != nil {
			return persistenceError(err)
		}
	}
	if bundle.RunID != "" && bundle.FinalRunStatus != "" {
		finishedAt := any(nil)
		if isTerminalRunStatus(bundle.FinalRunStatus) {
			finishedAt = now
		}
		if _, err := tx.ExecContext(ctx, `UPDATE novel_runs SET status = ?, runtime_snapshot_json = ?, finished_at = ?, updated_at = ? WHERE owner_username = ? AND run_id = ? AND project_id = ?`, bundle.FinalRunStatus, nullableRawJSON(bundle.RuntimeSnapshot), finishedAt, now, owner, bundle.RunID, bundle.ProjectID); err != nil {
			return persistenceError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return persistenceError(err)
	}
	return nil
}

func validateChapterBundleBindings(bundle ChapterBundle) error {
	invalid := func(message string) error {
		return &DomainError{Code: ErrInputRequired, Message: message}
	}
	projectID, chapterID, runID := strings.TrimSpace(bundle.ProjectID), strings.TrimSpace(bundle.Chapter.ID), strings.TrimSpace(bundle.RunID)
	matchesProject := func(kind, candidate string) error {
		if candidate != "" && strings.TrimSpace(candidate) != projectID {
			return invalid(kind + " projectId must match bundle projectId")
		}
		return nil
	}
	matchesChapter := func(kind, candidate string) error {
		if candidate != "" && strings.TrimSpace(candidate) != chapterID {
			return invalid(kind + " chapterId must match bundle chapter")
		}
		return nil
	}
	matchesRun := func(kind, candidate string) error {
		if candidate != "" && (runID == "" || strings.TrimSpace(candidate) != runID) {
			return invalid(kind + " runId must match bundle run")
		}
		return nil
	}
	if err := matchesProject("volume", bundle.Volume.ProjectID); err != nil {
		return err
	}
	if err := matchesProject("chapter", bundle.Chapter.ProjectID); err != nil {
		return err
	}
	if volumeID := strings.TrimSpace(bundle.Chapter.VolumeID); volumeID != "" {
		if strings.TrimSpace(bundle.Volume.ID) == "" || volumeID != strings.TrimSpace(bundle.Volume.ID) {
			return invalid("chapter volumeId must match bundle volume")
		}
	}
	if version := bundle.ChapterVersion; version.ID != "" || strings.TrimSpace(version.ChapterID) != "" || strings.TrimSpace(version.RunID) != "" {
		if strings.TrimSpace(version.ID) == "" {
			return invalid("chapter version id is required")
		}
		if err := matchesChapter("chapter version", version.ChapterID); err != nil {
			return err
		}
		if err := matchesRun("chapter version", version.RunID); err != nil {
			return err
		}
	}
	if err := matchesProject("foundation", bundle.Foundation.ProjectID); err != nil {
		return err
	}
	for _, memory := range bundle.Memories {
		if strings.TrimSpace(memory.ID) == "" {
			return invalid("memory id is required")
		}
		if err := matchesProject("memory", memory.ProjectID); err != nil {
			return err
		}
		if err := matchesChapter("memory", memory.ChapterID); err != nil {
			return err
		}
	}
	for _, audit := range bundle.Audits {
		if strings.TrimSpace(audit.ID) == "" {
			return invalid("audit id is required")
		}
		if err := matchesProject("audit", audit.ProjectID); err != nil {
			return err
		}
		if err := matchesChapter("audit", audit.ChapterID); err != nil {
			return err
		}
		if err := matchesRun("audit", audit.RunID); err != nil {
			return err
		}
	}
	for _, revision := range bundle.Revisions {
		if strings.TrimSpace(revision.ID) == "" {
			return invalid("revision id is required")
		}
		if err := matchesProject("revision", revision.ProjectID); err != nil {
			return err
		}
		if err := matchesChapter("revision", revision.ChapterID); err != nil {
			return err
		}
		if err := matchesRun("revision", revision.RunID); err != nil {
			return err
		}
	}
	return nil
}

var _ Store = (*MySQLStore)(nil)

func (s *MySQLStore) GetProjectSettings(ctx context.Context, owner, projectID string) (ProjectSettings, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return ProjectSettings{}, err
	}
	projectID = strings.TrimSpace(projectID)
	var settings ProjectSettings
	var style sql.NullString
	var rules, agent []byte
	err = s.db.QueryRowContext(ctx, `SELECT st.project_id, st.style_profile, st.rules_json, st.agent_json, st.updated_at
FROM novel_project_settings st
JOIN novel_projects p ON p.owner_username = st.owner_username AND p.project_id = st.project_id
WHERE st.owner_username = ? AND st.project_id = ? AND p.deleted_at IS NULL`, owner, projectID).
		Scan(&settings.ProjectID, &style, &rules, &agent, &settings.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if _, projectErr := s.GetProject(ctx, owner, projectID); projectErr != nil {
			return ProjectSettings{}, projectErr
		}
		return ProjectSettings{ProjectID: projectID, Rules: []string{}}, nil
	}
	if err != nil {
		return ProjectSettings{}, persistenceError(err)
	}
	settings.StyleProfile = style.String
	settings.Rules = []string{}
	if len(rules) > 0 {
		if unmarshalErr := json.Unmarshal(rules, &settings.Rules); unmarshalErr != nil {
			settings.Rules = []string{}
		}
	}
	if settings.Rules == nil {
		settings.Rules = []string{}
	}
	if len(agent) > 0 {
		if unmarshalErr := json.Unmarshal(agent, &settings.Agent); unmarshalErr != nil {
			settings.Agent = AgentSettings{}
		}
	}
	return settings, nil
}

func (s *MySQLStore) UpdateProjectSettings(ctx context.Context, owner, projectID string, input ProjectSettings) (ProjectSettings, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return ProjectSettings{}, err
	}
	projectID = strings.TrimSpace(projectID)
	rules := input.Rules
	if rules == nil {
		rules = []string{}
	}
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return ProjectSettings{}, persistenceError(err)
	}
	agentJSON, err := json.Marshal(NormalizeAgentSettings(input.Agent, 0))
	if err != nil {
		return ProjectSettings{}, persistenceError(err)
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `INSERT INTO novel_project_settings
(owner_username, project_id, style_profile, rules_json, agent_json, created_at, updated_at)
SELECT ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)
ON DUPLICATE KEY UPDATE style_profile = VALUES(style_profile), rules_json = VALUES(rules_json), agent_json = VALUES(agent_json), updated_at = VALUES(updated_at)`,
		owner, projectID, nullableString(strings.TrimSpace(input.StyleProfile)), []byte(rulesJSON), []byte(agentJSON), now, now, owner, projectID)
	if err != nil {
		return ProjectSettings{}, persistenceError(err)
	}
	if _, projectErr := s.GetProject(ctx, owner, projectID); projectErr != nil {
		return ProjectSettings{}, projectErr
	}
	return s.GetProjectSettings(ctx, owner, projectID)
}

func (s *MySQLStore) GetFoundation(ctx context.Context, owner, projectID string) (Foundation, bool, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Foundation{}, false, err
	}
	var foundation Foundation
	var snapshot []byte
	err = s.db.QueryRowContext(ctx, `SELECT f.foundation_id, f.project_id, f.snapshot_json, f.created_at, f.updated_at
FROM novel_foundations f
JOIN novel_projects p ON p.owner_username = f.owner_username AND p.project_id = f.project_id
WHERE f.owner_username = ? AND f.project_id = ? AND p.deleted_at IS NULL
ORDER BY f.updated_at DESC LIMIT 1`, owner, strings.TrimSpace(projectID)).
		Scan(&foundation.ID, &foundation.ProjectID, &snapshot, &foundation.CreatedAt, &foundation.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Foundation{}, false, nil
	}
	if err != nil {
		return Foundation{}, false, persistenceError(err)
	}
	foundation.Snapshot = cloneRawJSON(snapshot)
	return foundation, true, nil
}

func (s *MySQLStore) SaveFoundation(ctx context.Context, owner string, foundation Foundation) error {
	owner, err := requiredOwner(owner)
	if err != nil {
		return err
	}
	if err := validateRawJSON("foundation snapshot", foundation.Snapshot); err != nil {
		return err
	}
	foundation.ID, foundation.ProjectID = strings.TrimSpace(foundation.ID), strings.TrimSpace(foundation.ProjectID)
	if foundation.ID == "" || foundation.ProjectID == "" {
		return &DomainError{Code: ErrInputRequired, Message: "foundation id and projectId are required"}
	}
	if _, projectErr := s.GetProject(ctx, owner, foundation.ProjectID); projectErr != nil {
		return projectErr
	}
	now := time.Now().UTC()
	if foundation.CreatedAt.IsZero() {
		foundation.CreatedAt = now
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO novel_foundations (foundation_id, owner_username, project_id, snapshot_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE snapshot_json = VALUES(snapshot_json), updated_at = VALUES(updated_at)`,
		foundation.ID, owner, foundation.ProjectID, nullableRawJSON(foundation.Snapshot), foundation.CreatedAt, now)
	if err != nil {
		return persistenceError(err)
	}
	return nil
}

func (s *MySQLStore) ListMaterials(ctx context.Context, owner, projectID string) ([]Material, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	if _, projectErr := s.GetProject(ctx, owner, projectID); projectErr != nil {
		return nil, projectErr
	}
	rows, err := s.db.QueryContext(ctx, `SELECT m.material_id, m.project_id, m.title, m.content, m.created_at, m.updated_at
FROM novel_materials m
JOIN novel_projects p ON p.owner_username = m.owner_username AND p.project_id = m.project_id
WHERE m.owner_username = ? AND m.project_id = ? AND p.deleted_at IS NULL
ORDER BY m.created_at DESC`, owner, strings.TrimSpace(projectID))
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	materials := make([]Material, 0)
	for rows.Next() {
		var material Material
		if err := rows.Scan(&material.ID, &material.ProjectID, &material.Title, &material.Content, &material.CreatedAt, &material.UpdatedAt); err != nil {
			return nil, persistenceError(err)
		}
		materials = append(materials, material)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return materials, nil
}

func (s *MySQLStore) CreateMaterial(ctx context.Context, owner, projectID string, input CreateMaterialInput) (Material, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Material{}, err
	}
	input, err = NormalizeMaterialInput(input)
	if err != nil {
		return Material{}, err
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return Material{}, &DomainError{Code: ErrInputRequired, Message: "projectId is required"}
	}
	now := time.Now().UTC()
	material := Material{ID: novelID("material"), ProjectID: projectID, Title: input.Title, Content: input.Content, CreatedAt: now, UpdatedAt: now}
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_materials
(material_id, owner_username, project_id, title, content, created_at, updated_at)
SELECT ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)`,
		material.ID, owner, projectID, material.Title, material.Content, material.CreatedAt, material.UpdatedAt, owner, projectID)
	if err != nil {
		return Material{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Material{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	return material, nil
}

func (s *MySQLStore) DeleteMaterial(ctx context.Context, owner, projectID, materialID string) error {
	owner, err := requiredOwner(owner)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE m FROM novel_materials m
JOIN novel_projects p ON p.owner_username = m.owner_username AND p.project_id = m.project_id
WHERE m.owner_username = ? AND m.project_id = ? AND m.material_id = ? AND p.deleted_at IS NULL`,
		owner, strings.TrimSpace(projectID), strings.TrimSpace(materialID))
	if err != nil {
		return persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return &DomainError{Code: ErrProjectNotFound, Message: "material not found"}
	}
	return nil
}

func (s *MySQLStore) ListNotes(ctx context.Context, owner, projectID string) ([]ProjectNote, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	if _, projectErr := s.GetProject(ctx, owner, projectID); projectErr != nil {
		return nil, projectErr
	}
	rows, err := s.db.QueryContext(ctx, `SELECT n.note_id, n.project_id, n.title, n.content, n.created_at, n.updated_at
FROM novel_project_notes n
JOIN novel_projects p ON p.owner_username = n.owner_username AND p.project_id = n.project_id
WHERE n.owner_username = ? AND n.project_id = ? AND p.deleted_at IS NULL
ORDER BY n.updated_at DESC`, owner, strings.TrimSpace(projectID))
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	notes := make([]ProjectNote, 0)
	for rows.Next() {
		var note ProjectNote
		if err := rows.Scan(&note.ID, &note.ProjectID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, persistenceError(err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return notes, nil
}

func (s *MySQLStore) CreateNote(ctx context.Context, owner, projectID string, input ProjectNoteInput) (ProjectNote, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return ProjectNote{}, err
	}
	input, err = NormalizeNoteInput(input)
	if err != nil {
		return ProjectNote{}, err
	}
	projectID = strings.TrimSpace(projectID)
	now := time.Now().UTC()
	note := ProjectNote{ID: novelID("note"), ProjectID: projectID, Title: input.Title, Content: input.Content, CreatedAt: now, UpdatedAt: now}
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_project_notes
(note_id, owner_username, project_id, title, content, created_at, updated_at)
SELECT ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)`,
		note.ID, owner, projectID, note.Title, note.Content, note.CreatedAt, note.UpdatedAt, owner, projectID)
	if err != nil {
		return ProjectNote{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return ProjectNote{}, &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	return note, nil
}

func (s *MySQLStore) UpdateNote(ctx context.Context, owner, projectID, noteID string, input ProjectNoteInput) (ProjectNote, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return ProjectNote{}, err
	}
	input, err = NormalizeNoteInput(input)
	if err != nil {
		return ProjectNote{}, err
	}
	projectID, noteID = strings.TrimSpace(projectID), strings.TrimSpace(noteID)
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `UPDATE novel_project_notes n
JOIN novel_projects p ON p.owner_username = n.owner_username AND p.project_id = n.project_id
SET n.title = ?, n.content = ?, n.updated_at = ?
WHERE n.owner_username = ? AND n.project_id = ? AND n.note_id = ? AND p.deleted_at IS NULL`,
		input.Title, input.Content, now, owner, projectID, noteID)
	if err != nil {
		return ProjectNote{}, persistenceError(err)
	}
	var note ProjectNote
	err = s.db.QueryRowContext(ctx, `SELECT n.note_id, n.project_id, n.title, n.content, n.created_at, n.updated_at
FROM novel_project_notes n
JOIN novel_projects p ON p.owner_username = n.owner_username AND p.project_id = n.project_id
WHERE n.owner_username = ? AND n.project_id = ? AND n.note_id = ? AND p.deleted_at IS NULL`,
		owner, projectID, noteID).
		Scan(&note.ID, &note.ProjectID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ProjectNote{}, &DomainError{Code: ErrProjectNotFound, Message: "note not found"}
	}
	if err != nil {
		return ProjectNote{}, persistenceError(err)
	}
	return note, nil
}

func (s *MySQLStore) DeleteNote(ctx context.Context, owner, projectID, noteID string) error {
	owner, err := requiredOwner(owner)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE n FROM novel_project_notes n
JOIN novel_projects p ON p.owner_username = n.owner_username AND p.project_id = n.project_id
WHERE n.owner_username = ? AND n.project_id = ? AND n.note_id = ? AND p.deleted_at IS NULL`,
		owner, strings.TrimSpace(projectID), strings.TrimSpace(noteID))
	if err != nil {
		return persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return &DomainError{Code: ErrProjectNotFound, Message: "note not found"}
	}
	return nil
}

func (s *MySQLStore) ListRuns(ctx context.Context, owner, projectID string) ([]Run, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+qualifiedRunColumns+`
FROM novel_runs r JOIN novel_projects p ON p.owner_username = r.owner_username AND p.project_id = r.project_id
WHERE r.owner_username = ? AND r.project_id = ? AND p.deleted_at IS NULL
ORDER BY r.created_at DESC`, owner, strings.TrimSpace(projectID))
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	runs := make([]Run, 0)
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return runs, nil
}

const scheduleColumns = `s.schedule_id, s.owner_username, s.project_id, s.session_id, s.interval_minutes, s.enabled, s.prompt, s.model_id, s.thinking_level, s.next_run_at, s.last_run_at, COALESCE(s.last_run_id, ''), s.created_at, s.updated_at, s.input_snapshot_json`

type novelScheduleScanner interface{ Scan(...any) error }

func scanSchedule(scanner novelScheduleScanner) (Schedule, error) {
	var schedule Schedule
	var inputJSON []byte
	var lastRunAt sql.NullTime
	err := scanner.Scan(&schedule.ID, &schedule.OwnerUsername, &schedule.ProjectID, &schedule.SessionID,
		&schedule.IntervalMinutes, &schedule.Enabled, &schedule.Prompt, &schedule.ModelID, &schedule.ThinkingLevel,
		&schedule.NextRunAt, &lastRunAt, &schedule.LastRunID, &schedule.CreatedAt, &schedule.UpdatedAt, &inputJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Schedule{}, &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	if err != nil {
		return Schedule{}, persistenceError(err)
	}
	if lastRunAt.Valid {
		when := lastRunAt.Time
		schedule.LastRunAt = &when
	}
	var input AutomationInput
	if len(inputJSON) > 0 {
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			return Schedule{}, persistenceError(err)
		}
	}
	schedule.ModeInputs = input.ModeInputs
	schedule.KnowledgeIDs = input.KnowledgeIDs
	schedule.NoteIDs = input.NoteIDs
	schedule.StyleProfileID = input.StyleProfileID
	schedule.ConfigurationError = input.ConfigurationError
	return schedule, nil
}

func (s *MySQLStore) ListSchedules(ctx context.Context, owner, projectID string) ([]Schedule, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+scheduleColumns+`
FROM novel_schedules s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.owner_username = ? AND s.project_id = ? AND p.deleted_at IS NULL
ORDER BY s.created_at DESC`, owner, strings.TrimSpace(projectID))
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	schedules := make([]Schedule, 0)
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return schedules, nil
}

func (s *MySQLStore) GetSchedule(ctx context.Context, owner, scheduleID string) (Schedule, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Schedule{}, err
	}
	return scanSchedule(s.db.QueryRowContext(ctx, `SELECT `+scheduleColumns+`
FROM novel_schedules s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.owner_username = ? AND s.schedule_id = ? AND p.deleted_at IS NULL`, owner, strings.TrimSpace(scheduleID)))
}

func (s *MySQLStore) CreateSchedule(ctx context.Context, owner string, schedule Schedule) (Schedule, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Schedule{}, err
	}
	schedule.ID, schedule.ProjectID, schedule.SessionID = strings.TrimSpace(schedule.ID), strings.TrimSpace(schedule.ProjectID), strings.TrimSpace(schedule.SessionID)
	if schedule.ID == "" {
		schedule.ID = novelID("schedule")
	}
	if schedule.ProjectID == "" || schedule.SessionID == "" {
		return Schedule{}, &DomainError{Code: ErrInputRequired, Message: "schedule projectId and sessionId are required"}
	}
	now := time.Now().UTC()
	if schedule.CreatedAt.IsZero() {
		schedule.CreatedAt = now
	}
	schedule.UpdatedAt = now
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_schedules
(schedule_id, owner_username, project_id, session_id, interval_minutes, enabled, prompt, model_id, thinking_level, next_run_at, last_run_at, last_run_id, created_at, updated_at, input_snapshot_json)
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)
  AND EXISTS (SELECT 1 FROM novel_sessions WHERE owner_username = ? AND session_id = ? AND project_id = ?)`,
		schedule.ID, owner, schedule.ProjectID, schedule.SessionID, schedule.IntervalMinutes, schedule.Enabled, schedule.Prompt,
		schedule.ModelID, schedule.ThinkingLevel, schedule.NextRunAt, schedule.LastRunAt, nullableString(schedule.LastRunID),
		schedule.CreatedAt, schedule.UpdatedAt, inputJSON(scheduleInput(schedule)), owner, schedule.ProjectID, owner, schedule.SessionID, schedule.ProjectID)
	if err != nil {
		return Schedule{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Schedule{}, &DomainError{Code: ErrProjectNotFound, Message: "project or session not found"}
	}
	schedule.OwnerUsername = owner
	return schedule, nil
}

func (s *MySQLStore) UpdateSchedule(ctx context.Context, owner string, schedule Schedule) (Schedule, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Schedule{}, err
	}
	schedule.ID, schedule.ProjectID, schedule.SessionID = strings.TrimSpace(schedule.ID), strings.TrimSpace(schedule.ProjectID), strings.TrimSpace(schedule.SessionID)
	if schedule.ID == "" || schedule.ProjectID == "" {
		return Schedule{}, &DomainError{Code: ErrInputRequired, Message: "schedule id and projectId are required"}
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `UPDATE novel_schedules s
JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
SET s.session_id = ?, s.interval_minutes = ?, s.enabled = ?, s.prompt = ?, s.model_id = ?, s.thinking_level = ?,
    s.next_run_at = ?, s.last_run_at = ?, s.last_run_id = ?, s.updated_at = ?, s.input_snapshot_json = ?
WHERE s.owner_username = ? AND s.schedule_id = ? AND s.project_id = ? AND p.deleted_at IS NULL`,
		schedule.SessionID, schedule.IntervalMinutes, schedule.Enabled, schedule.Prompt, schedule.ModelID, schedule.ThinkingLevel,
		schedule.NextRunAt, schedule.LastRunAt, nullableString(schedule.LastRunID), now, inputJSON(scheduleInput(schedule)), owner, schedule.ID, schedule.ProjectID)
	if err != nil {
		return Schedule{}, persistenceError(err)
	}
	return s.GetSchedule(ctx, owner, schedule.ID)
}

func (s *MySQLStore) DeleteSchedule(ctx context.Context, owner, projectID, scheduleID string) error {
	owner, err := requiredOwner(owner)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE s FROM novel_schedules s
JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.owner_username = ? AND s.project_id = ? AND s.schedule_id = ? AND p.deleted_at IS NULL`,
		owner, strings.TrimSpace(projectID), strings.TrimSpace(scheduleID))
	if err != nil {
		return persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return &DomainError{Code: ErrProjectNotFound, Message: "schedule not found"}
	}
	return nil
}

func (s *MySQLStore) ListDueSchedules(ctx context.Context, now time.Time, limit int) ([]Schedule, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+scheduleColumns+`
FROM novel_schedules s JOIN novel_projects p ON p.owner_username = s.owner_username AND p.project_id = s.project_id
WHERE s.enabled = 1 AND s.next_run_at <= ? AND p.deleted_at IS NULL
ORDER BY s.next_run_at LIMIT ?`, now.UTC(), limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	schedules := make([]Schedule, 0)
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return schedules, nil
}

const batchColumns = `b.batch_id, b.owner_username, b.project_id, b.session_id, b.status, b.total_chapters, b.completed_chapters, COALESCE(b.current_run_id, ''), b.attempt, b.prompt, b.model_id, b.thinking_level, b.last_error, b.created_at, b.updated_at, b.input_snapshot_json`

type novelBatchScanner interface{ Scan(...any) error }

func scanBatch(scanner novelBatchScanner) (Batch, error) {
	var batch Batch
	var inputJSON []byte
	err := scanner.Scan(&batch.ID, &batch.OwnerUsername, &batch.ProjectID, &batch.SessionID,
		&batch.Status, &batch.TotalChapters, &batch.CompletedChapters, &batch.CurrentRunID, &batch.Attempt,
		&batch.Prompt, &batch.ModelID, &batch.ThinkingLevel, &batch.LastError, &batch.CreatedAt, &batch.UpdatedAt, &inputJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Batch{}, &DomainError{Code: ErrProjectNotFound, Message: "batch not found"}
	}
	if err != nil {
		return Batch{}, persistenceError(err)
	}
	var input AutomationInput
	if len(inputJSON) > 0 {
		if err := json.Unmarshal(inputJSON, &input); err != nil {
			return Batch{}, persistenceError(err)
		}
	}
	batch.ModeInputs = input.ModeInputs
	batch.KnowledgeIDs = input.KnowledgeIDs
	batch.NoteIDs = input.NoteIDs
	batch.StyleProfileID = input.StyleProfileID
	return batch, nil
}

func (s *MySQLStore) ListBatches(ctx context.Context, owner, projectID string) ([]Batch, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+batchColumns+`
FROM novel_batches b JOIN novel_projects p ON p.owner_username = b.owner_username AND p.project_id = b.project_id
WHERE b.owner_username = ? AND b.project_id = ? AND p.deleted_at IS NULL
ORDER BY b.created_at DESC`, owner, strings.TrimSpace(projectID))
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	batches := make([]Batch, 0)
	for rows.Next() {
		batch, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return batches, nil
}

func (s *MySQLStore) GetBatch(ctx context.Context, owner, batchID string) (Batch, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Batch{}, err
	}
	return scanBatch(s.db.QueryRowContext(ctx, `SELECT `+batchColumns+`
FROM novel_batches b JOIN novel_projects p ON p.owner_username = b.owner_username AND p.project_id = b.project_id
WHERE b.owner_username = ? AND b.batch_id = ? AND p.deleted_at IS NULL`, owner, strings.TrimSpace(batchID)))
}

func (s *MySQLStore) CreateBatch(ctx context.Context, owner string, batch Batch) (Batch, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Batch{}, err
	}
	batch.ProjectID, batch.SessionID = strings.TrimSpace(batch.ProjectID), strings.TrimSpace(batch.SessionID)
	if batch.ID == "" {
		batch.ID = novelID("batch")
	}
	if batch.ProjectID == "" || batch.SessionID == "" {
		return Batch{}, &DomainError{Code: ErrInputRequired, Message: "batch projectId and sessionId are required"}
	}
	now := time.Now().UTC()
	if batch.CreatedAt.IsZero() {
		batch.CreatedAt = now
	}
	batch.UpdatedAt = now
	result, err := s.db.ExecContext(ctx, `INSERT INTO novel_batches
(batch_id, owner_username, project_id, session_id, status, total_chapters, completed_chapters, current_run_id, attempt, prompt, model_id, thinking_level, last_error, created_at, updated_at, input_snapshot_json)
SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
WHERE EXISTS (SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL)
  AND EXISTS (SELECT 1 FROM novel_sessions WHERE owner_username = ? AND session_id = ? AND project_id = ?)`,
		batch.ID, owner, batch.ProjectID, batch.SessionID, batch.Status, batch.TotalChapters, batch.CompletedChapters,
		nullableString(batch.CurrentRunID), batch.Attempt, batch.Prompt, batch.ModelID, batch.ThinkingLevel, batch.LastError,
		batch.CreatedAt, batch.UpdatedAt, inputJSON(batchInput(batch)), owner, batch.ProjectID, owner, batch.SessionID, batch.ProjectID)
	if err != nil {
		return Batch{}, persistenceError(err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Batch{}, &DomainError{Code: ErrProjectNotFound, Message: "project or session not found"}
	}
	batch.OwnerUsername = owner
	return batch, nil
}

func (s *MySQLStore) UpdateBatch(ctx context.Context, owner string, batch Batch) (Batch, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return Batch{}, err
	}
	batch.ID, batch.ProjectID = strings.TrimSpace(batch.ID), strings.TrimSpace(batch.ProjectID)
	if batch.ID == "" || batch.ProjectID == "" {
		return Batch{}, &DomainError{Code: ErrInputRequired, Message: "batch id and projectId are required"}
	}
	_, err = s.db.ExecContext(ctx, `UPDATE novel_batches b
JOIN novel_projects p ON p.owner_username = b.owner_username AND p.project_id = b.project_id
SET b.status = ?, b.completed_chapters = ?, b.current_run_id = ?, b.attempt = ?, b.prompt = ?, b.model_id = ?,
    b.thinking_level = ?, b.last_error = ?, b.updated_at = ?, b.input_snapshot_json = ?
WHERE b.owner_username = ? AND b.batch_id = ? AND b.project_id = ? AND p.deleted_at IS NULL`,
		batch.Status, batch.CompletedChapters, nullableString(batch.CurrentRunID), batch.Attempt, batch.Prompt,
		batch.ModelID, batch.ThinkingLevel, batch.LastError, time.Now().UTC(), inputJSON(batchInput(batch)), owner, batch.ID, batch.ProjectID)
	if err != nil {
		return Batch{}, persistenceError(err)
	}
	return s.GetBatch(ctx, owner, batch.ID)
}

func (s *MySQLStore) ListActiveBatches(ctx context.Context, limit int) ([]Batch, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+batchColumns+`
FROM novel_batches b JOIN novel_projects p ON p.owner_username = b.owner_username AND p.project_id = b.project_id
WHERE b.status = 'running' AND p.deleted_at IS NULL
ORDER BY b.updated_at LIMIT ?`, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	batches := make([]Batch, 0)
	for rows.Next() {
		batch, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return batches, nil
}

func (s *MySQLStore) GetStyleProfile(ctx context.Context, owner, id string) (StyleProfile, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return StyleProfile{}, err
	}
	var p StyleProfile
	err = s.db.QueryRowContext(ctx, `SELECT style_profile_id, title, content FROM novel_style_profiles WHERE owner_username = ? AND style_profile_id = ?`, owner, strings.TrimSpace(id)).Scan(&p.ID, &p.Title, &p.Content)
	if errors.Is(err, sql.ErrNoRows) {
		return p, &DomainError{Code: ErrInputRequired, Message: "所选文风不存在或无权使用，请重新选择"}
	}
	if err != nil {
		return p, persistenceError(err)
	}
	return p, nil
}
func (s *MySQLStore) SaveStyleProfile(ctx context.Context, owner string, p StyleProfile) (StyleProfile, error) {
	owner, err := requiredOwner(owner)
	if err != nil {
		return StyleProfile{}, err
	}
	p, err = normalizeStyleProfile(p)
	if err != nil {
		return p, err
	}
	if p.ID == "" {
		p.ID = novelID("style")
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO novel_style_profiles (owner_username, style_profile_id, title, content) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE title = VALUES(title), content = VALUES(content)`, owner, p.ID, p.Title, p.Content)
	if err != nil {
		return p, persistenceError(err)
	}
	return p, nil
}
