package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"time"
)

type StudioCreateSessionInput struct {
	IdempotencyKey string `json:"idempotencyKey"`
	Mode           string `json:"mode"`
	Title          string `json:"title"`
}
type StudioSendInput struct {
	IdempotencyKey string `json:"idempotencyKey"`
	Content        string `json:"content"`
	ModelID        string `json:"modelId"`
	ThinkingLevel  string `json:"thinkingLevel"`
	ContextBudget  int    `json:"contextBudget"`
}
type StudioActionInput struct {
	ActionID      string `json:"actionId"`
	ModelID       string `json:"modelId"`
	ThinkingLevel string `json:"thinkingLevel,omitempty"`
}
type StudioSession struct {
	ID                  string    `json:"id"`
	Mode                string    `json:"mode"`
	Title               string    `json:"title"`
	ProjectID           string    `json:"projectId,omitempty"`
	ProductionSessionID string    `json:"productionSessionId,omitempty"`
	ProductionRunID     string    `json:"productionRunId,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
type StudioMessage struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Content string `json:"content"`
	Status  string `json:"status,omitempty"`
}
type StudioBook struct {
	Title              string `json:"title"`
	Genre              string `json:"genre"`
	World              string `json:"world"`
	Protagonist        string `json:"protagonist"`
	CoreConflict       string `json:"coreConflict"`
	TargetChapterWords int    `json:"targetChapterWords"`
}
type StudioAction struct {
	ID     string     `json:"id"`
	Type   string     `json:"type"`
	Status string     `json:"status"`
	Book   StudioBook `json:"book"`
}
type StudioTurn struct {
	ID             string `json:"id"`
	IdempotencyKey string `json:"idempotencyKey"`
	State          string `json:"state"`
	ErrorMessage   string `json:"errorMessage,omitempty"`
	ModelID        string `json:"modelId"`
}
type StudioSnapshot struct {
	Session       StudioSession   `json:"session"`
	Messages      []StudioMessage `json:"messages"`
	PendingAction *StudioAction   `json:"pendingAction,omitempty"`
	Turn          *StudioTurn     `json:"turn,omitempty"`
}
type StudioTransition struct {
	Session           StudioSession `json:"session"`
	Project           Project       `json:"project"`
	ProductionSession Session       `json:"productionSession"`
	Run               Run           `json:"run"`
}

// Internal records never expose request snapshots or raw provider replies.
type studioTurnRecord struct {
	StudioTurn
	Fingerprint       string
	Request           batchfactoryv11.TextCompletionRequest
	RawReply          string
	RawReplySHA256    string
	RawReplyBytes     int
	RawReplyTruncated bool
	SentAt            time.Time
}
type studioRecord struct {
	CreateKey         string
	CreateFingerprint string
	Session           StudioSession
	Messages          []StudioMessage
	Action            *StudioAction
	Turns             []studioTurnRecord
	Transition        *StudioTransition
}
type StudioStore interface {
	ReserveStudioCreation(context.Context, string, string, string, json.RawMessage) (json.RawMessage, error)
	LoadStudio(context.Context, string, string, bool) (json.RawMessage, error)
	SaveStudio(context.Context, string, string, json.RawMessage, bool) error
	StudioSessions(context.Context, string) ([]StudioSession, error)
}
type studioProviderResolver interface {
	ResolveStudioProvider(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)
}

func (p *Pipeline) ResolveStudioProvider(ctx context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
	if p == nil || p.resolveProvider == nil {
		return nil, inputError("小说聊天模型尚未配置")
	}
	return p.resolveProvider(ctx, owner, model)
}
