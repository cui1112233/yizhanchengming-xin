package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

type toolRowScanner interface {
	Scan(...any) error
}

func (s *SQLStore) LoadToolCall(ctx context.Context, toolID string) (ToolCall, error) {
	if s == nil || s.db == nil { return ToolCall{}, errors.New("agent store unavailable") }
	toolID = strings.TrimSpace(toolID)
	if toolID == "" { return ToolCall{}, errors.New("tool call id is required") }
	row := s.db.QueryRowContext(ctx, `SELECT id, thread_id, COALESCE(message_id,''), owner, tool_name, status, arguments_json, result_json, created_at, updated_at
FROM agent_tool_calls WHERE id=? LIMIT 1`, toolID)
	return scanToolCall(row)
}

func scanToolCall(row toolRowScanner) (ToolCall, error) {
	var item ToolCall
	var arguments, result []byte
	if err := row.Scan(&item.ID, &item.ThreadID, &item.MessageID, &item.Owner, &item.ToolName, &item.Status, &arguments, &result, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) { return ToolCall{}, ErrNotFound }
		return ToolCall{}, err
	}
	item.Arguments = append(json.RawMessage(nil), arguments...)
	item.Result = append(json.RawMessage(nil), result...)
	return item, nil
}
