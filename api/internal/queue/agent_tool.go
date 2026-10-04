package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const defaultAgentToolQueueKey = "qiantie:agent:tools"

type AgentToolQueue struct {
	key string
	pusher Pusher
}

type AgentToolSource struct {
	key string
	popper Popper
}

func NewAgentToolQueue(redisURL, key string) (*AgentToolQueue, error) {
	client, err := newNetworkRedis(redisURL)
	if err != nil { return nil, err }
	return NewAgentToolQueueWithPusher(key, client), nil
}

func NewAgentToolSource(redisURL, key string) (*AgentToolSource, error) {
	client, err := newNetworkRedis(redisURL)
	if err != nil { return nil, err }
	return NewAgentToolSourceWithPopper(key, client), nil
}

func NewAgentToolQueueWithPusher(key string, pusher Pusher) *AgentToolQueue {
	return &AgentToolQueue{key: normalizeAgentToolKey(key), pusher: pusher}
}

func NewAgentToolSourceWithPopper(key string, popper Popper) *AgentToolSource {
	return &AgentToolSource{key: normalizeAgentToolKey(key), popper: popper}
}

func normalizeAgentToolKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" { return defaultAgentToolQueueKey }
	return key
}

func (q *AgentToolQueue) Enqueue(ctx context.Context, toolCallID string) error {
	if q == nil || q.pusher == nil { return errors.New("agent tool queue pusher is required") }
	toolCallID = strings.TrimSpace(toolCallID)
	if toolCallID == "" { return errors.New("tool_call_id is required") }
	payload, err := json.Marshal(map[string]string{"tool_call_id": toolCallID})
	if err != nil { return err }
	return q.pusher.Push(ctx, q.key, payload)
}

func (s *AgentToolSource) Next(ctx context.Context) (string, error) {
	if s == nil || s.popper == nil { return "", errors.New("agent tool source popper is required") }
	payload, err := s.popper.Pop(ctx, s.key)
	if err != nil { return "", err }
	var message struct { ToolCallID string `json:"tool_call_id"` }
	if err := json.Unmarshal(payload, &message); err != nil { return "", fmt.Errorf("decode agent tool reference: %w", err) }
	message.ToolCallID = strings.TrimSpace(message.ToolCallID)
	if message.ToolCallID == "" { return "", errors.New("agent tool reference missing tool_call_id") }
	return message.ToolCallID, nil
}
