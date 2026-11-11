// Package agentchat implements the conversational Agent workspace:
// workspace -> session -> message, a permission-gated tool loop, and the
// file conventions used by the original product (.luoshui/ inside the
// workspace root, file-based sessions, JSONL run audit).
package agentchat

import (
	"strings"
	"time"
)

// Permission modes mirror the original three-way switch. Omitted or unknown
// values fall back to plan, which matches the original product default;
// bypass is the highest level and must be requested explicitly.
const (
	PermissionPlan   = "plan"
	PermissionAuto   = "auto"
	PermissionBypass = "bypass"
)

// NormalizePermission maps omitted/unknown values to plan. Only an explicit
// bypass value is treated as bypass.
func NormalizePermission(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case PermissionPlan:
		return PermissionPlan
	case PermissionAuto:
		return PermissionAuto
	case PermissionBypass:
		return PermissionBypass
	default:
		return PermissionPlan
	}
}

// PermissionLabel returns the UI label used by the original permission picker.
func PermissionLabel(mode string) string {
	switch NormalizePermission(mode) {
	case PermissionAuto:
		return "自动审批"
	case PermissionBypass:
		return "完全运行"
	default:
		return "计划模式"
	}
}

// PermissionOptions is the static descriptor list consumed by the frontend
// dropdown. Descriptions are the original Chinese strings.
func PermissionOptions() []map[string]string {
	return []map[string]string{
		{"id": PermissionPlan, "label": "计划模式", "description": "仅规划不执行，查看工具使用计划"},
		{"id": PermissionAuto, "label": "自动审批", "description": "只读工具自动执行；写入与 Skill 仅列入计划"},
		{"id": PermissionBypass, "label": "完全运行", "description": "所有工具与 Skill 自动执行"},
	}
}

// Workspace is one entry of workspaces.json. Field names follow the original
// snake_case file format so existing data can be imported unchanged.
type Workspace struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	RootPath        string   `json:"root_path"`
	IsDefault       bool     `json:"is_default"`
	Pinned          bool     `json:"pinned"`
	Archived        bool     `json:"archived"`
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       int64    `json:"updated_at"`
	EnabledSkillIDs []string `json:"enabled_skill_ids"`
	Stage           string   `json:"stage"`
}

type workspaceIndex struct {
	Version    int         `json:"version"`
	Workspaces []Workspace `json:"workspaces"`
}

// workspaceMeta is <root>/.luoshui/workspace.json. The original file only has
// these four fields.
type workspaceMeta struct {
	Name            string   `json:"name"`
	Version         int      `json:"version"`
	EnabledSkillIDs []string `json:"enabled_skill_ids"`
	Stage           string   `json:"stage"`
}

// SessionInfo is <sessions>/<title>_<id>/info.json.
type SessionInfo struct {
	ID                  int64  `json:"id"`
	Title               string `json:"title"`
	Model               string `json:"model"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	Kind                string `json:"kind"`
	ChatMode            string `json:"chat_mode"`
	WorkspaceID         string `json:"workspace_id"`
	AgentPermissionMode string `json:"agent_permission_mode"`
	PermissionMode      string `json:"permission_mode"`
	MessageCount        int    `json:"message_count"`
}

// ToolBlock is one step in a message's tool timeline. Keys are camelCase
// because messages.json is written and read by the browser in the original.
type ToolBlock struct {
	ID             string         `json:"id"`
	ToolName       string         `json:"toolName"`
	Name           string         `json:"name"`
	Args           map[string]any `json:"args"`
	Status         string         `json:"status"`
	Message        string         `json:"message"`
	Preview        string         `json:"preview"`
	Summary        string         `json:"summary"`
	Error          string         `json:"error"`
	ContentPreview string         `json:"contentPreview"`
	Artifacts      []string       `json:"artifacts"`
	Expanded       bool           `json:"expanded"`
	StartedAt      int64          `json:"startedAt"`
	CompletedAt    int64          `json:"completedAt"`
}

// Tool block statuses.
const (
	ToolPlanned = "planned"
	ToolRunning = "running"
	ToolDone    = "done"
	ToolFailed  = "failed"
)

// Message is one entry of messages.json. Unknown-but-preserved fields use
// generic maps so imported conversations keep their shape.
type Message struct {
	ID              int64            `json:"id"`
	Role            string           `json:"role"`
	Content         string           `json:"content"`
	Timestamp       string           `json:"timestamp"`
	Model           string           `json:"model"`
	Attachments     []map[string]any `json:"attachments"`
	Todos           []map[string]any `json:"todos"`
	ToolBlocks      []ToolBlock      `json:"toolBlocks"`
	ThinkingContent string           `json:"thinkingContent"`
	StreamSegments  []map[string]any `json:"streamSegments"`
	GeneratedFiles  []string         `json:"generatedFiles"`
	MediaItems      []map[string]any `json:"mediaItems"`
	CanContinue     bool             `json:"canContinue,omitempty"`
}

// EmptyMessage keeps every array non-nil so the browser can iterate without
// null checks, matching the original payload shape.
func EmptyMessage(role string) Message {
	return Message{
		Role:        role,
		Attachments: []map[string]any{},
		Todos:       []map[string]any{},
		ToolBlocks:  []ToolBlock{},
		StreamSegments: []map[string]any{},
		GeneratedFiles: []string{},
		MediaItems:     []map[string]any{},
	}
}

// Event is one SSE payload pushed to the browser.
type Event struct {
	Type string
	Data any
}

// Event types follow the naming already used by the original stream hooks.
const (
	EventAgentStatus    = "agent_status"
	EventReasoningDelta = "reasoning_delta"
	EventToolStart      = "tool_start"
	EventToolProgress   = "tool_progress"
	EventToolDone       = "tool_done"
	EventTodos          = "todos"
	EventToken          = "token"
	EventDone           = "done"
	EventError          = "error"
)

// SendRequest is one turn submitted from the composer.
type SendRequest struct {
	SessionID   int64
	WorkspaceID string
	Model       string
	Permission  string
	Content     string
	Attachments []map[string]any
	EnabledSkillIDs []string
}

// nowStamp renders info.json timestamps. The original uses a space separator
// and second precision, not RFC3339.
func nowStamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// clockStamp renders the HH:MM timestamp stored on each message.
func clockStamp(t time.Time) string {
	return t.Format("15:04")
}
