package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

const (
	ToolPromptRewrite = "prompt.rewrite"
	ToolImageGenerate = "image.generate"
	ToolImageEdit     = "image.edit"
	ToolVideoGenerate = "video.generate"
)

var ErrToolUnavailable = errors.New("agent tool unavailable")

type ToolExecutionRequest struct {
	Owner      string          `json:"owner"`
	ThreadID   string          `json:"thread_id"`
	ToolCallID string          `json:"tool_call_id"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments"`
}

type ToolExecutionResult struct {
	Content       string          `json:"content,omitempty"`
	MediaAssetIDs []string        `json:"media_asset_ids,omitempty"`
	Result        json.RawMessage `json:"result,omitempty"`
}

type ToolHandler interface {
	Execute(context.Context, ToolExecutionRequest) (ToolExecutionResult, error)
}

type ToolRegistry struct {
	mu       sync.RWMutex
	handlers map[string]ToolHandler
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{handlers: make(map[string]ToolHandler)}
}

func (r *ToolRegistry) Register(name string, handler ToolHandler) error {
	if r == nil { return errors.New("tool registry unavailable") }
	name = strings.TrimSpace(name)
	if name == "" || handler == nil { return errors.New("tool name and handler are required") }
	if !isStableToolName(name) { return fmt.Errorf("unsupported agent tool %q", name) }
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[name]; exists { return fmt.Errorf("agent tool %q already registered", name) }
	r.handlers[name] = handler
	return nil
}

func (r *ToolRegistry) Execute(ctx context.Context, input ToolExecutionRequest) (ToolExecutionResult, error) {
	if r == nil { return ToolExecutionResult{}, errors.New("tool registry unavailable") }
	input.Owner = strings.TrimSpace(input.Owner)
	input.ThreadID = strings.TrimSpace(input.ThreadID)
	input.ToolCallID = strings.TrimSpace(input.ToolCallID)
	input.ToolName = strings.TrimSpace(input.ToolName)
	if input.Owner == "" || input.ToolCallID == "" || input.ToolName == "" {
		return ToolExecutionResult{}, errors.New("tool owner, call id and name are required")
	}
	if len(input.Arguments) == 0 { input.Arguments = json.RawMessage(`{}`) }
	if !json.Valid(input.Arguments) { return ToolExecutionResult{}, errors.New("tool arguments must be valid json") }
	r.mu.RLock()
	handler := r.handlers[input.ToolName]
	r.mu.RUnlock()
	if handler == nil { return ToolExecutionResult{}, fmt.Errorf("%w: %s", ErrToolUnavailable, input.ToolName) }
	result, err := handler.Execute(ctx, input)
	if err != nil { return ToolExecutionResult{}, err }
	ids, err := normalizeMediaAssetIDs(result.MediaAssetIDs)
	if err != nil { return ToolExecutionResult{}, fmt.Errorf("invalid tool media result: %w", err) }
	result.MediaAssetIDs = ids
	if len(result.Result) > 0 && !json.Valid(result.Result) { return ToolExecutionResult{}, errors.New("tool result must be valid json") }
	result.Content = strings.TrimSpace(result.Content)
	return result, nil
}

func isStableToolName(name string) bool {
	switch name {
	case ToolNavigate, ToolPromptRewrite, ToolImageGenerate, ToolImageEdit, ToolVideoGenerate:
		return true
	default:
		return false
	}
}
