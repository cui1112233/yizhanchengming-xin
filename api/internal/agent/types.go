package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxMessageRunes = 12000

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

const (
	TaskInProgress    = "in_progress"
	TaskWaiting       = "waiting"
	TaskNeedsDecision = "needs_decision"
	TaskCompleted     = "completed"
	TaskFailed        = "failed"
)

const (
	ToolProposed  = "proposed"
	ToolCompleted = "completed"
	ToolFailed    = "failed"
)

var ErrNotFound = errors.New("agent resource not found")

type Thread struct {
	ID        string    `json:"id"`
	Owner     string    `json:"-"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID            string          `json:"id"`
	ThreadID      string          `json:"thread_id"`
	Owner         string          `json:"-"`
	Role          string          `json:"role"`
	Content       string          `json:"content"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	MediaAssetIDs []string        `json:"media_asset_ids,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type Task struct {
	ID              string    `json:"id"`
	ThreadID        string    `json:"thread_id"`
	Owner           string    `json:"-"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	ProgressCurrent uint      `json:"progress_current"`
	ProgressTotal   uint      `json:"progress_total"`
	Detail          string    `json:"detail,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ToolCall struct {
	ID        string          `json:"id"`
	ThreadID  string          `json:"thread_id"`
	MessageID string          `json:"message_id,omitempty"`
	Owner     string          `json:"-"`
	ToolName  string          `json:"tool_name"`
	Status    string          `json:"status"`
	Arguments json.RawMessage `json:"arguments"`
	Result    json.RawMessage `json:"result,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ThreadSnapshot struct {
	Thread    Thread     `json:"thread"`
	Messages  []Message  `json:"messages"`
	Tasks     []Task     `json:"tasks"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type AppendMessageInput struct {
	ThreadID      string
	Owner         string
	Role          string
	Content       string
	Metadata      json.RawMessage
	MediaAssetIDs []string
}

type CreateTaskInput struct {
	ThreadID        string
	Owner           string
	Title           string
	Status          string
	ProgressCurrent uint
	ProgressTotal   uint
	Detail          string
}

type CreateToolCallInput struct {
	ThreadID  string
	MessageID string
	Owner     string
	ToolName  string
	Status    string
	Arguments json.RawMessage
	Result    json.RawMessage
}

func validateMessageInput(input AppendMessageInput) error {
	if strings.TrimSpace(input.ThreadID) == "" || strings.TrimSpace(input.Owner) == "" {
		return errors.New("thread and owner are required")
	}
	if input.Role != RoleUser && input.Role != RoleAssistant && input.Role != RoleSystem {
		return errors.New("invalid message role")
	}
	content := strings.TrimSpace(input.Content)
	if content == "" && len(input.MediaAssetIDs) == 0 {
		return errors.New("message content or media is required")
	}
	if utf8.RuneCountInString(content) > MaxMessageRunes {
		return fmt.Errorf("message exceeds %d characters", MaxMessageRunes)
	}
	_, err := normalizeMediaAssetIDs(input.MediaAssetIDs)
	return err
}

func validTaskStatus(status string) bool {
	switch status {
	case TaskInProgress, TaskWaiting, TaskNeedsDecision, TaskCompleted, TaskFailed:
		return true
	default:
		return false
	}
}

func validToolStatus(status string) bool {
	switch status {
	case ToolProposed, ToolCompleted, ToolFailed:
		return true
	default:
		return false
	}
}
