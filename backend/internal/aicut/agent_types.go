package aicut

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrAgentRunNotFound       = errors.New("ai-cut agent run not found")
	ErrInvalidAgentRun        = errors.New("invalid ai-cut agent run")
	ErrInvalidAgentTransition = errors.New("invalid ai-cut agent state transition")
)

type AgentMode string

const (
	AgentModeAgent AgentMode = "agent"
	AgentModeAsk   AgentMode = "ask"
)

type AgentRunState string

const (
	AgentQueued          AgentRunState = "queued"
	AgentRunning         AgentRunState = "running"
	AgentCancelRequested AgentRunState = "cancel_requested"
	AgentSucceeded       AgentRunState = "succeeded"
	AgentFailed          AgentRunState = "failed"
	AgentCanceled        AgentRunState = "canceled"
)

type AgentRun struct {
	TimelineID        string        `json:"timelineId,omitempty"`
	ID                string        `json:"id"`
	ClientRequestID   string        `json:"clientRequestId"`
	Owner             string        `json:"-"`
	ProjectID         string        `json:"projectId"`
	Instruction       string        `json:"instruction"`
	Mode              AgentMode     `json:"mode"`
	ModelID           string        `json:"modelId,omitempty"`
	ActiveSkillID     string        `json:"activeSkillId,omitempty"`
	ExpectedRevision  int64         `json:"expectedRevision"`
	State             AgentRunState `json:"state"`
	TurnCount         int           `json:"turnCount"`
	Error             string        `json:"error,omitempty"`
	CancelRequestedAt *time.Time    `json:"cancelRequestedAt,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type AgentMessageRole string

const (
	AgentRoleUser      AgentMessageRole = "user"
	AgentRoleAssistant AgentMessageRole = "assistant"
	AgentRoleTool      AgentMessageRole = "tool"
)

type AgentMessage struct {
	ID         string           `json:"id"`
	RunID      string           `json:"runId"`
	Owner      string           `json:"-"`
	ProjectID  string           `json:"projectId"`
	Role       AgentMessageRole `json:"role"`
	Content    string           `json:"content"`
	ToolName   string           `json:"toolName,omitempty"`
	ToolCallID string           `json:"toolCallId,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
}

type AgentEventType string

const (
	AgentEventThinking   AgentEventType = "thinking"
	AgentEventToolStart  AgentEventType = "tool_start"
	AgentEventToolDone   AgentEventType = "tool_done"
	AgentEventToolFailed AgentEventType = "tool_failed"
	AgentEventReply      AgentEventType = "reply"
	AgentEventUsage      AgentEventType = "usage"
)

type AgentEvent struct {
	RunID      string         `json:"runId"`
	Owner      string         `json:"-"`
	ProjectID  string         `json:"projectId"`
	Sequence   int64          `json:"sequence"`
	Type       AgentEventType `json:"type"`
	Message    string         `json:"message,omitempty"`
	ToolName   string         `json:"toolName,omitempty"`
	ToolCallID string         `json:"toolCallId,omitempty"`
	Revision   int64          `json:"revision,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
}

type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
	CallID   string          `json:"callId"`
	Name     string          `json:"name"`
	OK       bool            `json:"ok"`
	Message  string          `json:"message"`
	Revision int64           `json:"revision,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

func normalizeAgentRun(run AgentRun) (AgentRun, error) {
	run.TimelineID = strings.TrimSpace(run.TimelineID)
	if run.TimelineID == DefaultTimelineID {
		run.TimelineID = ""
	}
	run.ID = strings.TrimSpace(run.ID)
	run.ClientRequestID = strings.TrimSpace(run.ClientRequestID)
	run.Owner = strings.TrimSpace(run.Owner)
	run.ProjectID = strings.TrimSpace(run.ProjectID)
	run.Instruction = strings.TrimSpace(run.Instruction)
	run.ModelID = strings.TrimSpace(run.ModelID)
	run.ActiveSkillID = strings.TrimSpace(run.ActiveSkillID)
	if run.ID == "" || run.ClientRequestID == "" || run.Owner == "" || run.ProjectID == "" || run.Instruction == "" || run.ExpectedRevision < 0 || len(run.TimelineID) > 64 {
		return AgentRun{}, ErrInvalidAgentRun
	}
	if run.Mode != AgentModeAgent && run.Mode != AgentModeAsk {
		return AgentRun{}, ErrInvalidAgentRun
	}
	if run.State == "" {
		run.State = AgentQueued
	}
	if !validAgentState(run.State) || run.TurnCount < 0 || run.TurnCount > 30 {
		return AgentRun{}, ErrInvalidAgentRun
	}
	return run, nil
}

func validAgentState(state AgentRunState) bool {
	switch state {
	case AgentQueued, AgentRunning, AgentCancelRequested, AgentSucceeded, AgentFailed, AgentCanceled:
		return true
	default:
		return false
	}
}

func validAgentTransition(from, to AgentRunState) bool {
	if from == to {
		return true
	}
	switch from {
	case AgentQueued:
		return to == AgentRunning || to == AgentCancelRequested || to == AgentFailed || to == AgentCanceled
	case AgentRunning:
		return to == AgentCancelRequested || to == AgentSucceeded || to == AgentFailed || to == AgentCanceled
	case AgentCancelRequested:
		return to == AgentCanceled || to == AgentFailed
	default:
		return false
	}
}
