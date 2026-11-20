package screenplay

import (
	"strings"
	"testing"
)

func TestCompactAssistantMessages(t *testing.T) {
	mk := func(n int) []AssistantMessage {
		out := make([]AssistantMessage, 0, n)
		for i := 0; i < n; i++ {
			out = append(out, AssistantMessage{Role: "user", Content: "第" + string(rune('A'+i%26)) + "轮很长的内容重复填充"})
		}
		return out
	}
	// 未超阈值不压缩
	msgs := mk(10)
	got, changed := CompactAssistantMessages(msgs, AssistantCompactConfig{MaxMessages: 40, MaxChars: 100000, KeepRecent: 5})
	if changed || len(got) != 10 {
		t.Fatalf("不应压缩 changed=%v len=%d", changed, len(got))
	}
	// 条数超阈值：折叠早期、保留最近
	msgs = mk(30)
	got, changed = CompactAssistantMessages(msgs, AssistantCompactConfig{MaxMessages: 20, MaxChars: 100000, KeepRecent: 6})
	if !changed {
		t.Fatal("应压缩")
	}
	if len(got) != 7 {
		t.Fatalf("压缩后条数=%d want 7(1摘要+6最近)", len(got))
	}
	if got[0].Kind != AssistantMessageKindCompacted || got[0].Role != "system" {
		t.Fatalf("首条应为压缩摘要: %#v", got[0])
	}
	if !strings.Contains(got[0].Content, "24") {
		t.Fatalf("摘要应含被折叠条数: %s", got[0].Content)
	}
	// 最近原文完整保留
	if got[len(got)-1].Content != msgs[len(msgs)-1].Content {
		t.Fatal("最近消息未原样保留")
	}
	// 字符超阈值也触发
	msgs = []AssistantMessage{{Role: "user", Content: strings.Repeat("长", 5000)}}
	_, changed = CompactAssistantMessages(msgs, AssistantCompactConfig{MaxMessages: 40, MaxChars: 100, KeepRecent: 1})
	if !changed {
		t.Fatal("字符超阈值应压缩")
	}
}

func TestBuildAssistantCapabilitiesAndCoverage(t *testing.T) {
	caps := BuildAssistantCapabilities(AssistantCapabilitiesInput{
		HasSource: true, HasSynopsis: true, HasChapters: true,
		HasScript: true, HasAssets: true, HasStoryboard: true,
	})
	if len(caps) != 10 {
		t.Fatalf("能力项=%d want 10", len(caps))
	}
	cov := BuildAssistantCoverage(caps)
	if cov.Total != 10 || cov.Ready != 6 || cov.Missing != 4 {
		t.Fatalf("覆盖度错误: %#v", cov)
	}
	if cov.NextID != "images" || cov.NextEntry != "media" {
		t.Fatalf("下一步应指向 images/media: %#v", cov)
	}
	// 全就绪
	all := BuildAssistantCapabilities(AssistantCapabilitiesInput{
		HasSource: true, HasSynopsis: true, HasChapters: true, HasScript: true,
		HasAssets: true, HasStoryboard: true, HasImages: true, HasVideos: true,
		HasSubtitles: true, HasFinalCut: true,
	})
	cov = BuildAssistantCoverage(all)
	if cov.Missing != 0 || cov.NextID != "" {
		t.Fatalf("全就绪应无缺口: %#v", cov)
	}
}

func TestFormatCapabilitiesForPromptNoToolExecution(t *testing.T) {
	caps := BuildAssistantCapabilities(AssistantCapabilitiesInput{HasSource: true})
	text := FormatCapabilitiesForPrompt(caps, BuildAssistantCoverage(caps))
	if !strings.Contains(text, "只读事实") {
		t.Fatal("应声明只读")
	}
	if !strings.Contains(text, "仍需用户确认") {
		t.Fatal("应声明编辑需确认")
	}
	if !strings.Contains(text, "不要声称已经执行") {
		t.Fatal("应禁止声称执行")
	}
}
