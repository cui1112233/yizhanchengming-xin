package agentworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
)

type Source interface {
	Next(context.Context) (string, error)
}

type Store interface {
	LoadToolCall(context.Context, string) (agent.ToolCall, error)
	UpdateToolCall(context.Context, string, string, agent.UpdateToolCallInput) (agent.ToolCall, error)
	AppendMessage(context.Context, agent.AppendMessageInput) (agent.Message, error)
}

type Executor interface {
	Execute(context.Context, agent.ToolExecutionRequest) (agent.ToolExecutionResult, error)
}

type Runner struct {
	Source Source
	Store Store
	Executor Executor
}

type persistedResult struct {
	Content       string          `json:"content,omitempty"`
	MediaAssetIDs []string        `json:"media_asset_ids,omitempty"`
	Data          json.RawMessage `json:"data,omitempty"`
	Error         string          `json:"error,omitempty"`
}

func (r Runner) RunOnce(ctx context.Context) error {
	if r.Source == nil || r.Store == nil || r.Executor == nil {
		return errors.New("agent tool runner requires source, store and executor")
	}
	toolCallID, err := r.Source.Next(ctx)
	if err != nil { return err }
	toolCallID = strings.TrimSpace(toolCallID)
	if toolCallID == "" { return errors.New("agent tool source returned empty id") }
	call, err := r.Store.LoadToolCall(ctx, toolCallID)
	if err != nil { return fmt.Errorf("load agent tool call %s: %w", toolCallID, err) }
	if call.Status == agent.ToolCompleted || call.Status == agent.ToolFailed {
		return nil
	}
	if call.Status != agent.ToolProposed {
		return fmt.Errorf("agent tool call %s has unsupported status %q", call.ID, call.Status)
	}

	result, executeErr := r.Executor.Execute(ctx, agent.ToolExecutionRequest{
		Owner: call.Owner,
		ThreadID: call.ThreadID,
		ToolCallID: call.ID,
		ToolName: call.ToolName,
		Arguments: append(json.RawMessage(nil), call.Arguments...),
	})
	if executeErr != nil {
		return r.fail(ctx, call, executeErr)
	}
	body, err := json.Marshal(persistedResult{
		Content: strings.TrimSpace(result.Content),
		MediaAssetIDs: append([]string(nil), result.MediaAssetIDs...),
		Data: append(json.RawMessage(nil), result.Result...),
	})
	if err != nil { return r.fail(ctx, call, fmt.Errorf("encode tool result: %w", err)) }
	if _, err := r.Store.UpdateToolCall(ctx, call.Owner, call.ID, agent.UpdateToolCallInput{Status: agent.ToolCompleted, Result: body}); err != nil {
		return fmt.Errorf("mark agent tool completed: %w", err)
	}
	content := strings.TrimSpace(result.Content)
	if content == "" && len(result.MediaAssetIDs) > 0 { content = "执行完成。" }
	if content != "" || len(result.MediaAssetIDs) > 0 {
		if _, err := r.Store.AppendMessage(ctx, agent.AppendMessageInput{
			ThreadID: call.ThreadID,
			Owner: call.Owner,
			Role: agent.RoleAssistant,
			Content: content,
			MediaAssetIDs: append([]string(nil), result.MediaAssetIDs...),
		}); err != nil {
			return fmt.Errorf("append agent tool result message: %w", err)
		}
	}
	return nil
}

func (r Runner) fail(ctx context.Context, call agent.ToolCall, cause error) error {
	message := strings.TrimSpace(cause.Error())
	if len([]rune(message)) > 800 { message = string([]rune(message)[:800]) }
	body, _ := json.Marshal(persistedResult{Error: message})
	if _, err := r.Store.UpdateToolCall(ctx, call.Owner, call.ID, agent.UpdateToolCallInput{Status: agent.ToolFailed, Result: body}); err != nil {
		return fmt.Errorf("agent tool failed: %v; persist failure state: %w", cause, err)
	}
	_, appendErr := r.Store.AppendMessage(ctx, agent.AppendMessageInput{
		ThreadID: call.ThreadID,
		Owner: call.Owner,
		Role: agent.RoleAssistant,
		Content: "执行失败：" + message,
	})
	if appendErr != nil { return fmt.Errorf("agent tool failed: %v; append failure message: %w", cause, appendErr) }
	return fmt.Errorf("agent tool execution failed: %w", cause)
}
