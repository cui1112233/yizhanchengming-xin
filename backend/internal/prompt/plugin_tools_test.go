package prompt

import "testing"

func TestPluginToolPromptOptionsMatchLuoshuiPanel(t *testing.T) {
	counts := map[string]int{
		"novel-to-drama":    8,
		"storyboard-master": 2,
		"viral-opening":     2,
		"character-edit":    1,
	}
	for tool, want := range counts {
		items, ok := PluginToolPromptOptions(tool)
		if !ok {
			t.Fatalf("%s: tool not found", tool)
		}
		if len(items) != want {
			t.Fatalf("%s: got %d prompts, want %d", tool, len(items), want)
		}
	}

	novel, _ := PluginToolPromptOptions("novel-to-drama")
	if novel[0].Name != "【Minimax情绪专用2】-不会修改原文" || novel[7].Name != "中长篇纯对话" {
		t.Fatalf("novel prompt order mismatch: %q ... %q", novel[0].Name, novel[7].Name)
	}
	storyboard, _ := PluginToolPromptOptions("storyboard-master")
	if storyboard[0].Name != "姬漫·分镜脚本-推文分镜1" {
		t.Fatalf("storyboard prompt order mismatch: %q", storyboard[0].Name)
	}
	viral, _ := PluginToolPromptOptions("viral-opening")
	if viral[0].Name != "短剧爆款开头" || viral[1].Name != "超强爆款开头" {
		t.Fatalf("viral-opening order mismatch: %q, %q", viral[0].Name, viral[1].Name)
	}
}

func TestPluginToolSystemPromptUsesLuoshuiCorpus(t *testing.T) {
	for _, tool := range []string{"novel-to-drama", "storyboard-master", "viral-opening", "character-edit"} {
		items, _ := PluginToolPromptOptions(tool)
		for _, item := range items {
			body, ok := PluginToolSystemPrompt(tool, item.ID)
			if !ok || len([]rune(body)) == 0 {
				t.Fatalf("%s/%s: corpus body not resolved", tool, item.ID)
			}
		}
	}

	cases := map[[2]string]int{
		{"novel-to-drama", "mid-length"}:             571,
		{"novel-to-drama", "default"}:                9663,
		{"storyboard-master", "jimman-storyboard-1"}: 1887,
		{"viral-opening", "short-drama"}:             6155,
		{"character-edit", "default"}:                4088,
	}
	for key, want := range cases {
		body, ok := PluginToolSystemPrompt(key[0], key[1])
		if !ok {
			t.Fatalf("%s/%s: not found", key[0], key[1])
		}
		if got := len([]rune(body)); got != want {
			t.Fatalf("%s/%s: body length %d, want %d", key[0], key[1], got, want)
		}
	}
}
