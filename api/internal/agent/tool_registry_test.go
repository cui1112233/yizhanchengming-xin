package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type toolHandlerFunc func(context.Context, ToolExecutionRequest) (ToolExecutionResult, error)
func (f toolHandlerFunc) Execute(ctx context.Context, input ToolExecutionRequest) (ToolExecutionResult, error) { return f(ctx, input) }

func TestToolRegistryExecutesRegisteredCreativeTool(t *testing.T) {
	registry := NewToolRegistry()
	called := false
	if err := registry.Register(ToolImageGenerate, toolHandlerFunc(func(_ context.Context, input ToolExecutionRequest) (ToolExecutionResult, error) {
		called = true
		if input.Owner != "owner" || input.ToolCallID != "tool_1" { t.Fatalf("input=%#v", input) }
		return ToolExecutionResult{Content: "生成完成", MediaAssetIDs: []string{"asset_1"}}, nil
	})); err != nil { t.Fatal(err) }

	result, err := registry.Execute(context.Background(), ToolExecutionRequest{
		Owner: "owner", ThreadID: "thread_1", ToolCallID: "tool_1", ToolName: ToolImageGenerate,
		Arguments: json.RawMessage(`{"prompt":"人物站在雨里"}`),
	})
	if err != nil { t.Fatal(err) }
	if !called || len(result.MediaAssetIDs) != 1 || result.MediaAssetIDs[0] != "asset_1" { t.Fatalf("result=%#v", result) }
}

func TestToolRegistryRejectsUnregisteredTool(t *testing.T) {
	registry := NewToolRegistry()
	if _, err := registry.Execute(context.Background(), ToolExecutionRequest{Owner: "owner", ToolCallID: "tool", ToolName: "shell.exec", Arguments: json.RawMessage(`{}`)}); !errors.Is(err, ErrToolUnavailable) {
		t.Fatalf("err=%v", err)
	}
}

func TestToolRegistryRejectsInvalidArgumentsBeforeHandler(t *testing.T) {
	registry := NewToolRegistry()
	calls := 0
	_ = registry.Register(ToolVideoGenerate, toolHandlerFunc(func(context.Context, ToolExecutionRequest) (ToolExecutionResult, error) {
		calls++
		return ToolExecutionResult{}, nil
	}))
	if _, err := registry.Execute(context.Background(), ToolExecutionRequest{Owner: "owner", ToolCallID: "tool", ToolName: ToolVideoGenerate, Arguments: json.RawMessage(`{`)}); err == nil {
		t.Fatal("expected invalid json error")
	}
	if calls != 0 { t.Fatalf("handler calls=%d", calls) }
}

func TestToolRegistryRejectsResultMediaURLsAndPaths(t *testing.T) {
	registry := NewToolRegistry()
	_ = registry.Register(ToolImageEdit, toolHandlerFunc(func(context.Context, ToolExecutionRequest) (ToolExecutionResult, error) {
		return ToolExecutionResult{MediaAssetIDs: []string{"https://example.com/a.png"}}, nil
	}))
	if _, err := registry.Execute(context.Background(), ToolExecutionRequest{Owner: "owner", ToolCallID: "tool", ToolName: ToolImageEdit, Arguments: json.RawMessage(`{}`)}); err == nil {
		t.Fatal("expected non-asset result to be rejected")
	}
}

func TestCreativeToolNamesAreStable(t *testing.T) {
	want := []string{ToolPromptRewrite, ToolImageGenerate, ToolImageEdit, ToolVideoGenerate}
	got := []string{"prompt.rewrite", "image.generate", "image.edit", "video.generate"}
	for index := range want {
		if want[index] != got[index] { t.Fatalf("tool[%d]=%q", index, want[index]) }
	}
}
