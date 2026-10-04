package queue

import (
	"context"
	"encoding/json"
	"testing"
)

type agentToolPopperFake struct { payload []byte }
func (f agentToolPopperFake) Pop(context.Context, string) ([]byte, error) { return append([]byte(nil), f.payload...), nil }

func TestAgentToolQueuePushesOnlyToolCallID(t *testing.T) {
	pusher := &fakePusher{}
	queue := NewAgentToolQueueWithPusher("qiantie:agent:tools", pusher)
	if err := queue.Enqueue(context.Background(), "tool_123"); err != nil { t.Fatal(err) }
	if pusher.key != "qiantie:agent:tools" { t.Fatalf("key=%q", pusher.key) }
	var payload map[string]any
	if err := json.Unmarshal(pusher.payload, &payload); err != nil { t.Fatal(err) }
	if len(payload) != 1 || payload["tool_call_id"] != "tool_123" { t.Fatalf("payload=%v", payload) }
}

func TestAgentToolSourceReturnsToolCallID(t *testing.T) {
	source := NewAgentToolSourceWithPopper("qiantie:agent:tools", agentToolPopperFake{payload: []byte(`{"tool_call_id":"tool_123"}`)})
	id, err := source.Next(context.Background())
	if err != nil { t.Fatal(err) }
	if id != "tool_123" { t.Fatalf("id=%q", id) }
}

func TestAgentToolQueueRejectsEmptyReference(t *testing.T) {
	queue := NewAgentToolQueueWithPusher("qiantie:agent:tools", &fakePusher{})
	if err := queue.Enqueue(context.Background(), " "); err == nil { t.Fatal("expected empty tool call id error") }
}

func TestAgentToolSourceRejectsPayloadWithNoReference(t *testing.T) {
	source := NewAgentToolSourceWithPopper("qiantie:agent:tools", agentToolPopperFake{payload: []byte(`{"prompt":"must not be in redis"}`)})
	if _, err := source.Next(context.Background()); err == nil { t.Fatal("expected missing tool call id error") }
}
