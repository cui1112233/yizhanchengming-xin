package agentworker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
)

type sourceFake struct { id string; err error }
func (f sourceFake) Next(context.Context) (string, error) { return f.id, f.err }

type storeFake struct {
	tool agent.ToolCall
	loadErr error
	updates []agent.UpdateToolCallInput
	messages []agent.AppendMessageInput
}
func (f *storeFake) LoadToolCall(context.Context, string) (agent.ToolCall, error) { return f.tool, f.loadErr }
func (f *storeFake) UpdateToolCall(_ context.Context, _ string, _ string, input agent.UpdateToolCallInput) (agent.ToolCall, error) {
	f.updates = append(f.updates, input)
	f.tool.Status = input.Status
	f.tool.Result = input.Result
	return f.tool, nil
}
func (f *storeFake) AppendMessage(_ context.Context, input agent.AppendMessageInput) (agent.Message, error) {
	f.messages = append(f.messages, input)
	return agent.Message{ID: "msg_result", ThreadID: input.ThreadID, Role: input.Role, Content: input.Content, MediaAssetIDs: input.MediaAssetIDs}, nil
}

type executorFake struct { result agent.ToolExecutionResult; err error; calls int; request agent.ToolExecutionRequest }
func (f *executorFake) Execute(_ context.Context, input agent.ToolExecutionRequest) (agent.ToolExecutionResult, error) {
	f.calls++
	f.request = input
	return f.result, f.err
}

func TestRunnerLoadsDurableToolCallAndPersistsMediaResult(t *testing.T) {
	store := &storeFake{tool: agent.ToolCall{
		ID: "tool_1", Owner: "owner", ThreadID: "thread_1", ToolName: agent.ToolImageGenerate,
		Status: agent.ToolProposed, Arguments: json.RawMessage(`{"prompt":"雨夜"}`),
	}}
	executor := &executorFake{result: agent.ToolExecutionResult{Content: "图片已生成", MediaAssetIDs: []string{"asset_1", "asset_2"}}}
	runner := Runner{Source: sourceFake{id: "tool_1"}, Store: store, Executor: executor}
	if err := runner.RunOnce(context.Background()); err != nil { t.Fatal(err) }
	if executor.calls != 1 || executor.request.ToolCallID != "tool_1" || executor.request.Owner != "owner" { t.Fatalf("execution=%#v", executor) }
	if len(store.updates) != 1 || store.updates[0].Status != agent.ToolCompleted { t.Fatalf("updates=%#v", store.updates) }
	if len(store.messages) != 1 || store.messages[0].Role != agent.RoleAssistant || len(store.messages[0].MediaAssetIDs) != 2 { t.Fatalf("messages=%#v", store.messages) }
	if !json.Valid(store.updates[0].Result) { t.Fatalf("result json=%q", store.updates[0].Result) }
}

func TestRunnerMarksToolFailedWithoutImmediateRequeue(t *testing.T) {
	store := &storeFake{tool: agent.ToolCall{ID: "tool_1", Owner: "owner", ThreadID: "thread_1", ToolName: agent.ToolVideoGenerate, Status: agent.ToolProposed, Arguments: json.RawMessage(`{}`)}}
	executor := &executorFake{err: errors.New("provider unavailable")}
	runner := Runner{Source: sourceFake{id: "tool_1"}, Store: store, Executor: executor}
	err := runner.RunOnce(context.Background())
	if err == nil { t.Fatal("expected execution error") }
	if len(store.updates) != 1 || store.updates[0].Status != agent.ToolFailed { t.Fatalf("updates=%#v", store.updates) }
	if len(store.messages) != 1 || store.messages[0].Content == "" { t.Fatalf("messages=%#v", store.messages) }
}

func TestRunnerIsIdempotentForCompletedToolCall(t *testing.T) {
	store := &storeFake{tool: agent.ToolCall{ID: "tool_1", Owner: "owner", ThreadID: "thread_1", ToolName: agent.ToolImageGenerate, Status: agent.ToolCompleted, Arguments: json.RawMessage(`{}`)}}
	executor := &executorFake{}
	runner := Runner{Source: sourceFake{id: "tool_1"}, Store: store, Executor: executor}
	if err := runner.RunOnce(context.Background()); err != nil { t.Fatal(err) }
	if executor.calls != 0 || len(store.updates) != 0 || len(store.messages) != 0 { t.Fatalf("unexpected side effects") }
}

func TestRunnerRejectsMissingDependencies(t *testing.T) {
	if err := (Runner{}).RunOnce(context.Background()); err == nil { t.Fatal("expected dependency error") }
}
