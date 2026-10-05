package generation

import (
	"context"
	"strings"
	"testing"
)

func TestPromptResolverUsesEnabledLatestVersionAndRecordsVersion(t *testing.T) {
	store := newMemoryStore()
	store.prompts = []Prompt{
		{Key: PromptKeyScript, Version: 1, Content: "old", Enabled: true},
		{Key: PromptKeyScript, Version: 2, Content: "new", Enabled: true},
		{Key: PromptKeyScript, Version: 3, Content: "disabled", Enabled: false},
	}
	resolver := NewPromptResolver(store)
	prompt, err := resolver.Resolve(context.Background(), PromptKeyScript)
	if err != nil { t.Fatal(err) }
	if prompt.Version != 2 || prompt.Content != "new" {
		t.Fatalf("unexpected prompt: %#v", prompt)
	}
}

func TestStoryModeMetaPromptIsGenericStableAndVisual(t *testing.T) {
	body := StoryModeMetaPrompt
	for _, required := range []string{
		"首镜头", "空间", "动作", "特写", "中景", "全景", "推镜", "拉镜", "摇镜", "跟拍", "环境光", "阴影",
		"不得改变主剧情", "性暗示", "未成年人",
	} {
		if !strings.Contains(body, required) { t.Fatalf("missing %q", required) }
	}
	for _, banned := range []string{"七夕", "男友", "闺蜜", "煤矿"} {
		if strings.Contains(body, banned) { t.Fatalf("story mode prompt must be generic, found %q", banned) }
	}
}
