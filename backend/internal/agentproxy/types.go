package agentproxy

import (
	"strings"
	"time"
)

// 权限两档：plan 只读/只建议，auto 可直接执行。
// bypass 对应洛水三档中的最高档，第一版只保留取值位，接口层不开放。
const (
	PermissionPlan   = "plan"
	PermissionAuto   = "auto"
	PermissionBypass = "bypass"
)

// 会话状态取值，第一版只用到 active 和 archived。
const (
	SessionStatusActive   = "active"
	SessionStatusArchived = "archived"
)

const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
	MessageRoleTool      = "tool"
)

type Workspace struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	PermissionMode string    `json:"permissionMode"`
	SessionCount   int       `json:"sessionCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Session struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspaceId"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	PermissionMode string     `json:"permissionMode"`
	MessageCount   int        `json:"messageCount"`
	LastMessageAt  *time.Time `json:"lastMessageAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type Message struct {
	ID        int64          `json:"id"`
	SessionID string         `json:"sessionId"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

// Skill 是技能表的结构占位，第一版不解析技能包，列表恒为空。
type Skill struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	SkillKey    string    `json:"skillKey"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ScheduledJob 是定时任务的结构占位，第一版只读列表，不调度执行。
type ScheduledJob struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspaceId"`
	Name           string     `json:"name"`
	Schedule       string     `json:"schedule"`
	Prompt         string     `json:"prompt"`
	PermissionMode string     `json:"permissionMode"`
	Enabled        bool       `json:"enabled"`
	LastRunAt      *time.Time `json:"lastRunAt,omitempty"`
	NextRunAt      *time.Time `json:"nextRunAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type CreateWorkspaceInput struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PermissionMode string `json:"permissionMode"`
}

type UpdateWorkspaceInput struct {
	Name           *string `json:"name"`
	Description    *string `json:"description"`
	PermissionMode *string `json:"permissionMode"`
}

type CreateSessionInput struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	PermissionMode string `json:"permissionMode"`
}

type UpdateSessionInput struct {
	Title          *string `json:"title"`
	Status         *string `json:"status"`
	PermissionMode *string `json:"permissionMode"`
}

type AppendMessageInput struct {
	Role     string         `json:"role"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
	// Execute 表示这条消息请求直接执行工具，plan 档位下会被拒绝。
	Execute bool `json:"execute"`
}

// NormalizePermissionMode 统一大小写和空白，未知取值原样返回交给校验处理。
func NormalizePermissionMode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func NormalizeSessionStatus(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// ValidPermissionMode 只接受 plan/auto；bypass 暂不开放。
func ValidPermissionMode(value string) bool {
	mode := NormalizePermissionMode(value)
	return mode == PermissionPlan || mode == PermissionAuto
}

func ValidMessageRole(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleSystem, MessageRoleTool:
		return true
	}
	return false
}

func ValidSessionStatus(value string) bool {
	status := NormalizeSessionStatus(value)
	return status == SessionStatusActive || status == SessionStatusArchived
}
