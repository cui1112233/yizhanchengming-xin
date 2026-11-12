package aicut

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
)

func (s *MySQLStore) CreateAgentRun(ctx context.Context, run AgentRun) (AgentRun, error) {
	normalized, err := normalizeAgentRun(run)
	if err != nil {
		return AgentRun{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO aicut_agent_runs (id, client_request_id, project_id, owner_username, instruction, mode, model_id, active_skill_id, expected_revision, state, turn_count, error_message, created_at, updated_at, timeline_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, normalized.ID, normalized.ClientRequestID, normalized.ProjectID, normalized.Owner, normalized.Instruction, normalized.Mode, normalized.ModelID, normalized.ActiveSkillID, normalized.ExpectedRevision, normalized.State, normalized.TurnCount, normalized.Error, normalized.CreatedAt, normalized.UpdatedAt, normalized.TimelineID)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		existing, e := s.agentRunByClientRequest(ctx, normalized.Owner, normalized.ProjectID, normalized.ClientRequestID)
		if e != nil {
			return AgentRun{}, e
		}
		if canonicalASRTimeline(existing.TimelineID) != canonicalASRTimeline(normalized.TimelineID) {
			return AgentRun{}, ErrInvalidAgentRun
		}
		return existing, nil
	}
	if err != nil {
		return AgentRun{}, err
	}
	return normalized, nil
}

type agentRunScanner interface{ Scan(...any) error }

func scanAgentRun(scanner agentRunScanner) (AgentRun, error) {
	var run AgentRun
	var cancelRequested sql.NullTime
	err := scanner.Scan(&run.ID, &run.ClientRequestID, &run.Owner, &run.ProjectID, &run.Instruction, &run.Mode, &run.ModelID, &run.ActiveSkillID, &run.ExpectedRevision, &run.State, &run.TurnCount, &run.Error, &cancelRequested, &run.CreatedAt, &run.UpdatedAt, &run.TimelineID)
	if err != nil {
		return AgentRun{}, err
	}
	if cancelRequested.Valid {
		value := cancelRequested.Time
		run.CancelRequestedAt = &value
	}
	return run, nil
}

const agentRunColumns = `id, client_request_id, owner_username, project_id, instruction, mode, model_id, active_skill_id, expected_revision, state, turn_count, COALESCE(error_message, ''), cancel_requested_at, created_at, updated_at, timeline_id`

func (s *MySQLStore) agentRunByClientRequest(ctx context.Context, owner, projectID, clientRequestID string) (AgentRun, error) {
	run, err := scanAgentRun(s.db.QueryRowContext(ctx, `SELECT `+agentRunColumns+` FROM aicut_agent_runs WHERE owner_username = ? AND project_id = ? AND client_request_id = ?`, owner, projectID, clientRequestID))
	if errors.Is(err, sql.ErrNoRows) {
		return AgentRun{}, ErrAgentRunNotFound
	}
	return run, err
}

func (s *MySQLStore) agentRunByID(ctx context.Context, id string) (AgentRun, error) {
	run, err := scanAgentRun(s.db.QueryRowContext(ctx, `SELECT `+agentRunColumns+` FROM aicut_agent_runs WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return AgentRun{}, ErrAgentRunNotFound
	}
	return run, err
}

func (s *MySQLStore) AgentRun(ctx context.Context, owner, projectID, id string) (AgentRun, error) {
	run, err := scanAgentRun(s.db.QueryRowContext(ctx, `SELECT `+agentRunColumns+` FROM aicut_agent_runs WHERE id = ? AND owner_username = ? AND project_id = ?`, id, owner, projectID))
	if errors.Is(err, sql.ErrNoRows) {
		return AgentRun{}, ErrAgentRunNotFound
	}
	return run, err
}

func (s *MySQLStore) ClaimAgentRun(ctx context.Context, id string, now time.Time) (AgentRun, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_agent_runs SET state = ?, updated_at = ? WHERE id = ? AND state = ?`, AgentRunning, now.UTC(), id, AgentQueued)
	if err != nil {
		return AgentRun{}, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		if _, err := s.agentRunByID(ctx, id); err != nil {
			return AgentRun{}, err
		}
		return AgentRun{}, ErrInvalidAgentTransition
	}
	return s.agentRunByID(ctx, id)
}

func (s *MySQLStore) UpdateAgentRunState(ctx context.Context, id string, state AgentRunState, message string, turnCount int, now time.Time) (AgentRun, error) {
	current, err := s.agentRunByID(ctx, id)
	if err != nil {
		return AgentRun{}, err
	}
	if !validAgentTransition(current.State, state) || turnCount < current.TurnCount || turnCount > 30 {
		return AgentRun{}, ErrInvalidAgentTransition
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE aicut_agent_runs SET state = ?, error_message = ?, turn_count = ?, updated_at = ? WHERE id = ? AND state = ?`, state, message, turnCount, now.UTC(), id, current.State); err != nil {
		return AgentRun{}, err
	}
	return s.agentRunByID(ctx, id)
}

func (s *MySQLStore) RequestAgentCancel(ctx context.Context, owner, projectID, id string, now time.Time) (AgentRun, error) {
	current, err := s.AgentRun(ctx, owner, projectID, id)
	if err != nil {
		return AgentRun{}, err
	}
	if current.State == AgentSucceeded || current.State == AgentFailed || current.State == AgentCanceled || current.State == AgentCancelRequested {
		return current, nil
	}
	when := now.UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE aicut_agent_runs SET state = ?, cancel_requested_at = ?, updated_at = ? WHERE id = ? AND owner_username = ? AND project_id = ? AND state IN (?, ?)`, AgentCancelRequested, when, when, id, owner, projectID, AgentQueued, AgentRunning)
	if err != nil {
		return AgentRun{}, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return s.AgentRun(ctx, owner, projectID, id)
	}
	current.State, current.CancelRequestedAt, current.UpdatedAt = AgentCancelRequested, &when, when
	return current, nil
}

func (s *MySQLStore) AppendAgentMessage(ctx context.Context, message AgentMessage) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO aicut_agent_messages (id, run_id, project_id, owner_username, role, content, tool_name, tool_call_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, message.ID, message.RunID, message.ProjectID, message.Owner, message.Role, message.Content, message.ToolName, message.ToolCallID, message.CreatedAt)
	return err
}

func (s *MySQLStore) AgentMessages(ctx context.Context, owner, projectID string, limit int) ([]AgentMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, run_id, project_id, owner_username, role, content, tool_name, tool_call_id, created_at FROM aicut_agent_messages WHERE owner_username = ? AND project_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`, owner, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AgentMessage, 0)
	for rows.Next() {
		var item AgentMessage
		if err := rows.Scan(&item.ID, &item.RunID, &item.ProjectID, &item.Owner, &item.Role, &item.Content, &item.ToolName, &item.ToolCallID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return items, rows.Err()
}

func (s *MySQLStore) AppendAgentEvent(ctx context.Context, event AgentEvent) (AgentEvent, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AgentEvent{}, err
	}
	defer tx.Rollback()
	var sequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(sequence_no), 0) FROM aicut_agent_events WHERE run_id = ? AND owner_username = ? AND project_id = ? FOR UPDATE`, event.RunID, event.Owner, event.ProjectID).Scan(&sequence); err != nil {
		return AgentEvent{}, err
	}
	event.Sequence = sequence + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO aicut_agent_events (run_id, project_id, owner_username, sequence_no, event_type, message, tool_name, tool_call_id, revision, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, event.RunID, event.ProjectID, event.Owner, event.Sequence, event.Type, event.Message, event.ToolName, event.ToolCallID, event.Revision, event.CreatedAt); err != nil {
		return AgentEvent{}, err
	}
	if err := tx.Commit(); err != nil {
		return AgentEvent{}, err
	}
	return event, nil
}

func (s *MySQLStore) AgentEvents(ctx context.Context, owner, projectID, runID string, after int64) ([]AgentEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT run_id, project_id, owner_username, sequence_no, event_type, COALESCE(message, ''), tool_name, tool_call_id, revision, created_at FROM aicut_agent_events WHERE run_id = ? AND owner_username = ? AND project_id = ? AND sequence_no > ? ORDER BY sequence_no`, runID, owner, projectID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AgentEvent, 0)
	for rows.Next() {
		var item AgentEvent
		if err := rows.Scan(&item.RunID, &item.ProjectID, &item.Owner, &item.Sequence, &item.Type, &item.Message, &item.ToolName, &item.ToolCallID, &item.Revision, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *MySQLStore) AgentToolResult(ctx context.Context, owner, projectID, runID, callID string) (ToolResult, bool, error) {
	var result ToolResult
	var data []byte
	result.CallID = callID
	err := s.db.QueryRowContext(ctx, `SELECT tool_name, ok, message, revision, data_json FROM aicut_agent_tool_results WHERE run_id = ? AND owner_username = ? AND project_id = ? AND tool_call_id = ?`, runID, owner, projectID, callID).Scan(&result.Name, &result.OK, &result.Message, &result.Revision, &data)
	if errors.Is(err, sql.ErrNoRows) {
		return ToolResult{}, false, nil
	}
	if err != nil {
		return ToolResult{}, false, err
	}
	if len(data) > 0 && string(data) != "null" {
		result.Data = append([]byte(nil), data...)
	}
	return result, true, nil
}

func (s *MySQLStore) SaveAgentToolResult(ctx context.Context, owner, projectID, runID string, result ToolResult) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO aicut_agent_tool_results (run_id, project_id, owner_username, tool_call_id, tool_name, ok, message, revision, data_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE tool_call_id = VALUES(tool_call_id)`, runID, projectID, owner, result.CallID, result.Name, result.OK, result.Message, result.Revision, []byte(result.Data), time.Now().UTC())
	return err
}

var _ AgentStore = (*MySQLStore)(nil)

func (s *MySQLStore) AgentTimelineMessages(ctx context.Context, o, p, id string, limit int) ([]AgentMessage, error) {
	return s.agentTimelineMessages(ctx, o, p, id, limit, false)
}
func (s *MySQLStore) AgentConversationMessages(ctx context.Context, o, p, id string, limit int) ([]AgentMessage, error) {
	return s.agentTimelineMessages(ctx, o, p, id, limit, true)
}
func (s *MySQLStore) agentTimelineMessages(ctx context.Context, o, p, id string, limit int, conversationOnly bool) ([]AgentMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if canonicalASRTimeline(id) == DefaultTimelineID {
		id = ""
	}
	roleClause := ""
	if conversationOnly {
		roleClause = " AND m.role IN ('user', 'assistant')"
	}
	rows, e := s.db.QueryContext(ctx, `SELECT m.id, m.run_id, m.project_id, m.owner_username, m.role, m.content, m.tool_name, m.tool_call_id, m.created_at FROM aicut_agent_messages m INNER JOIN aicut_agent_runs r ON r.id = m.run_id AND r.owner_username = m.owner_username AND r.project_id = m.project_id WHERE m.owner_username = ? AND m.project_id = ? AND r.timeline_id = ?`+roleClause+` ORDER BY m.created_at DESC, m.id DESC LIMIT ?`, o, p, id, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []AgentMessage{}
	for rows.Next() {
		var m AgentMessage
		if e = rows.Scan(&m.ID, &m.RunID, &m.ProjectID, &m.Owner, &m.Role, &m.Content, &m.ToolName, &m.ToolCallID, &m.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}
