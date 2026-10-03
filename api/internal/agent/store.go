package agent

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Store interface {
	CreateThread(context.Context, string, string) (Thread, error)
	ListThreads(context.Context, string, int) ([]Thread, error)
	GetThread(context.Context, string, string) (ThreadSnapshot, error)
	AppendMessage(context.Context, AppendMessageInput) (Message, error)
	CreateTask(context.Context, CreateTaskInput) (Task, error)
	ListTasks(context.Context, string, string) ([]Task, error)
	CreateToolCall(context.Context, CreateToolCallInput) (ToolCall, error)
	UpdateThreadTitle(context.Context, string, string, string) error
}

type agentTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type SQLStore struct {
	db    *sql.DB
	begin func(context.Context) (agentTx, error)
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return newSQLStoreWithBegin(db, func(ctx context.Context) (agentTx, error) {
		return db.BeginTx(ctx, nil)
	})
}

func newSQLStoreWithBegin(db *sql.DB, begin func(context.Context) (agentTx, error)) *SQLStore {
	return &SQLStore{db: db, begin: begin}
}

var mediaAssetIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func normalizeMediaAssetIDs(input []string) ([]string, error) {
	if len(input) > 24 {
		return nil, errors.New("too many media references")
	}
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, raw := range input {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if !mediaAssetIDPattern.MatchString(id) {
			return nil, fmt.Errorf("invalid media asset id %q", id)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func newID(prefix string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return prefix + "_" + hex.EncodeToString(buf)
}

func cleanTitle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "新对话"
	}
	runes := []rune(value)
	if len(runes) > 120 {
		runes = runes[:120]
	}
	return string(runes)
}

func (s *SQLStore) CreateThread(ctx context.Context, owner, title string) (Thread, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return Thread{}, errors.New("owner is required")
	}
	thread := Thread{
		ID: newID("thr"), Owner: owner, Title: cleanTitle(title), Status: "active",
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO agent_threads (id, owner, title, status) VALUES (?, ?, ?, ?)`, thread.ID, thread.Owner, thread.Title, thread.Status)
	if err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func (s *SQLStore) ListThreads(ctx context.Context, owner string, limit int) ([]Thread, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, owner, title, status, created_at, updated_at FROM agent_threads WHERE owner=? ORDER BY updated_at DESC, id DESC LIMIT ?`, owner, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	threads := make([]Thread, 0)
	for rows.Next() {
		var item Thread
		if err := rows.Scan(&item.ID, &item.Owner, &item.Title, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		threads = append(threads, item)
	}
	return threads, rows.Err()
}

func (s *SQLStore) GetThread(ctx context.Context, owner, threadID string) (ThreadSnapshot, error) {
	var snapshot ThreadSnapshot
	if err := s.db.QueryRowContext(ctx, `SELECT id, owner, title, status, created_at, updated_at FROM agent_threads WHERE id=? AND owner=? LIMIT 1`, threadID, owner).
		Scan(&snapshot.Thread.ID, &snapshot.Thread.Owner, &snapshot.Thread.Title, &snapshot.Thread.Status, &snapshot.Thread.CreatedAt, &snapshot.Thread.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ThreadSnapshot{}, ErrNotFound
		}
		return ThreadSnapshot{}, err
	}

	messages, err := s.loadMessages(ctx, owner, threadID)
	if err != nil {
		return ThreadSnapshot{}, err
	}
	tasks, err := s.ListTasks(ctx, owner, threadID)
	if err != nil {
		return ThreadSnapshot{}, err
	}
	tools, err := s.loadToolCalls(ctx, owner, threadID)
	if err != nil {
		return ThreadSnapshot{}, err
	}
	snapshot.Messages = messages
	snapshot.Tasks = tasks
	snapshot.ToolCalls = tools
	return snapshot, nil
}

func (s *SQLStore) loadMessages(ctx context.Context, owner, threadID string) ([]Message, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.id, m.thread_id, m.owner, m.role, m.content, m.metadata_json, m.created_at, mm.media_asset_id
FROM agent_messages m
LEFT JOIN agent_message_media mm ON mm.message_id=m.id
WHERE m.thread_id=? AND m.owner=?
ORDER BY m.created_at ASC, m.id ASC, mm.ordinal ASC`, threadID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]Message, 0)
	indexes := map[string]int{}
	for rows.Next() {
		var item Message
		var metadata []byte
		var mediaID sql.NullString
		if err := rows.Scan(&item.ID, &item.ThreadID, &item.Owner, &item.Role, &item.Content, &metadata, &item.CreatedAt, &mediaID); err != nil {
			return nil, err
		}
		index, exists := indexes[item.ID]
		if !exists {
			item.Metadata = append(json.RawMessage(nil), metadata...)
			messages = append(messages, item)
			index = len(messages) - 1
			indexes[item.ID] = index
		}
		if mediaID.Valid && mediaID.String != "" {
			messages[index].MediaAssetIDs = append(messages[index].MediaAssetIDs, mediaID.String)
		}
	}
	return messages, rows.Err()
}

func (s *SQLStore) AppendMessage(ctx context.Context, input AppendMessageInput) (Message, error) {
	if err := validateMessageInput(input); err != nil {
		return Message{}, err
	}
	mediaIDs, err := normalizeMediaAssetIDs(input.MediaAssetIDs)
	if err != nil {
		return Message{}, err
	}
	message := Message{
		ID: newID("msg"), ThreadID: strings.TrimSpace(input.ThreadID), Owner: strings.TrimSpace(input.Owner),
		Role: input.Role, Content: strings.TrimSpace(input.Content), Metadata: append(json.RawMessage(nil), input.Metadata...),
		MediaAssetIDs: mediaIDs, CreatedAt: time.Now().UTC(),
	}
	tx, err := s.begin(ctx)
	if err != nil {
		return Message{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	result, err := tx.ExecContext(ctx, `INSERT INTO agent_messages (id, thread_id, owner, role, content, metadata_json)
SELECT ?, id, ?, ?, ?, ? FROM agent_threads WHERE id=? AND owner=?`,
		message.ID, message.Owner, message.Role, message.Content, nullableJSON(message.Metadata), message.ThreadID, message.Owner)
	if err != nil {
		return Message{}, err
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		if err != nil { return Message{}, err }
		return Message{}, ErrNotFound
	}
	for ordinal, mediaID := range mediaIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO agent_message_media (message_id, media_asset_id, ordinal) VALUES (?, ?, ?)`, message.ID, mediaID, ordinal); err != nil {
			return Message{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE agent_threads SET updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND owner=?`, message.ThreadID, message.Owner); err != nil {
		return Message{}, err
	}
	if err := tx.Commit(); err != nil {
		return Message{}, err
	}
	committed = true
	return message, nil
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return []byte(value)
}

func (s *SQLStore) CreateTask(ctx context.Context, input CreateTaskInput) (Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Owner == "" || input.ThreadID == "" || input.Title == "" {
		return Task{}, errors.New("thread, owner and task title are required")
	}
	if input.Status == "" {
		input.Status = TaskInProgress
	}
	if !validTaskStatus(input.Status) {
		return Task{}, errors.New("invalid task status")
	}
	task := Task{ID: newID("task"), ThreadID: input.ThreadID, Owner: input.Owner, Title: input.Title, Status: input.Status, ProgressCurrent: input.ProgressCurrent, ProgressTotal: input.ProgressTotal, Detail: strings.TrimSpace(input.Detail), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	result, err := s.db.ExecContext(ctx, `INSERT INTO agent_tasks (id, thread_id, owner, title, status, progress_current, progress_total, detail)
SELECT ?, id, ?, ?, ?, ?, ?, ? FROM agent_threads WHERE id=? AND owner=?`,
		task.ID, task.Owner, task.Title, task.Status, task.ProgressCurrent, task.ProgressTotal, task.Detail, task.ThreadID, task.Owner)
	if err != nil {
		return Task{}, err
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		if err != nil { return Task{}, err }
		return Task{}, ErrNotFound
	}
	return task, nil
}

func (s *SQLStore) ListTasks(ctx context.Context, owner, threadID string) ([]Task, error) {
	query := `SELECT id, thread_id, owner, title, status, progress_current, progress_total, COALESCE(detail,''), created_at, updated_at FROM agent_tasks WHERE owner=?`
	args := []any{owner}
	if strings.TrimSpace(threadID) != "" {
		query += ` AND thread_id=?`
		args = append(args, threadID)
	}
	query += ` ORDER BY updated_at DESC, id DESC LIMIT 100`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Task, 0)
	for rows.Next() {
		var item Task
		if err := rows.Scan(&item.ID, &item.ThreadID, &item.Owner, &item.Title, &item.Status, &item.ProgressCurrent, &item.ProgressTotal, &item.Detail, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLStore) CreateToolCall(ctx context.Context, input CreateToolCallInput) (ToolCall, error) {
	input.Owner = strings.TrimSpace(input.Owner)
	input.ThreadID = strings.TrimSpace(input.ThreadID)
	input.ToolName = strings.TrimSpace(input.ToolName)
	if input.Owner == "" || input.ThreadID == "" || input.ToolName == "" {
		return ToolCall{}, errors.New("thread, owner and tool name are required")
	}
	if input.Status == "" {
		input.Status = ToolProposed
	}
	if !validToolStatus(input.Status) {
		return ToolCall{}, errors.New("invalid tool status")
	}
	if len(input.Arguments) == 0 || !json.Valid(input.Arguments) {
		return ToolCall{}, errors.New("tool arguments must be valid json")
	}
	if len(input.Result) > 0 && !json.Valid(input.Result) {
		return ToolCall{}, errors.New("tool result must be valid json")
	}
	item := ToolCall{ID: newID("tool"), ThreadID: input.ThreadID, MessageID: strings.TrimSpace(input.MessageID), Owner: input.Owner, ToolName: input.ToolName, Status: input.Status, Arguments: append(json.RawMessage(nil), input.Arguments...), Result: append(json.RawMessage(nil), input.Result...), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	var messageID any
	if item.MessageID != "" { messageID = item.MessageID }
	result, err := s.db.ExecContext(ctx, `INSERT INTO agent_tool_calls (id, thread_id, message_id, owner, tool_name, status, arguments_json, result_json)
SELECT ?, id, ?, ?, ?, ?, ?, ? FROM agent_threads WHERE id=? AND owner=?`,
		item.ID, messageID, item.Owner, item.ToolName, item.Status, []byte(item.Arguments), nullableJSON(item.Result), item.ThreadID, item.Owner)
	if err != nil {
		return ToolCall{}, err
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		if err != nil { return ToolCall{}, err }
		return ToolCall{}, ErrNotFound
	}
	return item, nil
}

func (s *SQLStore) loadToolCalls(ctx context.Context, owner, threadID string) ([]ToolCall, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, thread_id, COALESCE(message_id,''), owner, tool_name, status, arguments_json, result_json, created_at, updated_at
FROM agent_tool_calls WHERE owner=? AND thread_id=? ORDER BY created_at ASC, id ASC`, owner, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ToolCall, 0)
	for rows.Next() {
		var item ToolCall
		var arguments, result []byte
		if err := rows.Scan(&item.ID, &item.ThreadID, &item.MessageID, &item.Owner, &item.ToolName, &item.Status, &arguments, &result, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Arguments = append(json.RawMessage(nil), arguments...)
		item.Result = append(json.RawMessage(nil), result...)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *SQLStore) UpdateThreadTitle(ctx context.Context, owner, threadID, title string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE agent_threads SET title=?, updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND owner=?`, cleanTitle(title), threadID, owner)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
