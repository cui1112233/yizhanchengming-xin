package prompt

import (
	"strings"
	"testing"
)

func TestAICutAgentPromptSeparatesAgentAndQAModes(t *testing.T) {
	agent := BuildAICutAgentSystemPrompt("agent")
	ask := BuildAICutAgentSystemPrompt("ask")
	for _, required := range []string{"只使用提供的工具", "不得声称未执行的工具已经成功", "toolCalls", "reply"} {
		if !strings.Contains(agent, required) {
			t.Fatalf("agent prompt missing %q", required)
		}
	}
	if !strings.Contains(ask, "问答模式不会改时间轴") || !strings.Contains(ask, "禁止调用工具") {
		t.Fatalf("ask prompt does not prohibit tools: %s", ask)
	}
}

func TestAICutAgentPromptDeclaresEveryAllowedToolAndRevisionRule(t *testing.T) {
	prompt := BuildAICutAgentSystemPrompt("agent")
	for _, required := range []string{"get_project_snapshot", "place_asset", "move_clip", "delete_clip", "split_clip", "set_playback_rate", "replace_subtitle_text", "add_bgm", "undo_timeline", "redo_timeline", "先读取项目快照", "revision"} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("agent prompt missing tool or rule %q", required)
		}
	}
}
