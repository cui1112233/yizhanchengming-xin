package screenplay

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// 本文件实现「作品助手」的三项对话智能增强，对齐洛水 CreationAgentExecutor
// 的 maybe_compact / _ask_user / _coverage_for_job，但保持 V88 的安全边界：
//   1. 不自动执行任何工具或媒体生产（洛水 36 工具调度不照搬——会烧额度、越权）；
//   2. 只提供只读的链路能力清单 + 覆盖度自检 + 导航建议；
//   3. 主动追问只是把问题回给用户，不代替用户决策；
//   4. 会话压缩只折叠历史消息文本，不改任何作品数据。
//
// 这些函数都是纯函数，便于单测与回放。

// AssistantMessage 的扩展字段（不改动持久化结构，通过别名类型承载）。
// kind 取值："" 普通对话；"question" 助手追问；"compacted" 压缩摘要。
const (
	AssistantMessageKindQuestion  = "question"
	AssistantMessageKindCompacted = "compacted"
)

// AssistantCompactConfig 控制对话压缩阈值。
type AssistantCompactConfig struct {
	MaxMessages int // 超过该条数后压缩早期消息；<=0 用默认 40
	MaxChars    int // 早期消息累计字符超过该值也触发；<=0 用默认 24000
	KeepRecent  int // 始终保留的最近消息条数；<=0 用默认 12
}

func (c AssistantCompactConfig) normalized() AssistantCompactConfig {
	if c.MaxMessages <= 0 {
		c.MaxMessages = 40
	}
	if c.MaxChars <= 0 {
		c.MaxChars = 24000
	}
	if c.KeepRecent <= 0 {
		c.KeepRecent = 12
	}
	if c.KeepRecent >= c.MaxMessages {
		c.KeepRecent = c.MaxMessages - 1
	}
	if c.KeepRecent < 1 {
		c.KeepRecent = 1
	}
	return c
}

// CompactAssistantMessages 在消息过长时，把早期消息折叠成一条只读摘要占位，
// 保留最近 KeepRecent 条原文。摘要只统计角色与条数、不编造内容，避免模型把
// 占位当成事实。返回压缩后的消息与是否发生了压缩。
func CompactAssistantMessages(messages []AssistantMessage, cfg AssistantCompactConfig) ([]AssistantMessage, bool) {
	cfg = cfg.normalized()
	runes := 0
	for _, m := range messages {
		runes += utf8.RuneCountInString(m.Content)
	}
	if len(messages) <= cfg.MaxMessages && runes <= cfg.MaxChars {
		return messages, false
	}
	cut := len(messages) - cfg.KeepRecent
	if cut < 1 {
		cut = 1
	}
	// 已经折叠过的前缀不要重复计数/截断内容语义。
	compacted := 0
	for _, m := range messages[:cut] {
		if m.Kind == AssistantMessageKindCompacted {
			compacted += utf8.RuneCountInString(m.Content)
		}
	}
	folded := messages[:cut]
	kept := messages[cut:]
	summary := fmt.Sprintf("【历史对话已折叠】更早的 %d 条对话已省略（保留最近 %d 条原文）。如需追溯细节请让用户重新说明。", len(folded), len(kept))
	result := make([]AssistantMessage, 0, len(kept)+1)
	result = append(result, AssistantMessage{Role: "system", Content: summary, Kind: AssistantMessageKindCompacted})
	result = append(result, kept...)
	_ = compacted
	return result, true
}

// AssistantCapability 是链路中一个可导航的能力环节（只读）。
// Ready 表示作品中已有对应产物；Entry 是前端面板标识，用于导航建议。
type AssistantCapability struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Ready bool   `json:"ready"`
	Entry string `json:"entry"`
	Hint  string `json:"hint,omitempty"`
}

// AssistantCapabilitiesInput 是构建能力清单所需的作品状态快照。
type AssistantCapabilitiesInput struct {
	HasSource    bool
	HasSynopsis  bool
	HasChapters  bool
	HasScript    bool
	HasAssets    bool
	HasStoryboard bool
	HasImages    bool
	HasVideos    bool
	HasSubtitles bool
	HasFinalCut  bool
	HasMedical   bool // 占位：保留字段避免与未来能力冲突；当前恒 false
}

// BuildAssistantCapabilities 把作品状态快照转成固定顺序的能力清单。
// 顺序即链路顺序：原文 → 概要 → 章节 → 剧本 → 资产 → 分镜 → 图片 → 视频 → 字幕 → 成片。
func BuildAssistantCapabilities(in AssistantCapabilitiesInput) []AssistantCapability {
	return []AssistantCapability{
		{ID: "source", Label: "原文导入", Ready: in.HasSource, Entry: "source", Hint: "上传或粘贴小说正文，后续全部生成的源头"},
		{ID: "synopsis", Label: "故事概要", Ready: in.HasSynopsis, Entry: "script", Hint: "主舞台故事概要区可点击 AI 生成"},
		{ID: "chapters", Label: "章节规划", Ready: in.HasChapters, Entry: "script", Hint: "作品助手可先规划章节"},
		{ID: "script", Label: "剧本正文", Ready: in.HasScript, Entry: "script", Hint: "在剧本区生成或编辑分集剧本"},
		{ID: "assets", Label: "角色场景道具资产", Ready: in.HasAssets, Entry: "assets", Hint: "资产页解析，决定画面角色一致性"},
		{ID: "storyboard", Label: "分镜", Ready: in.HasStoryboard, Entry: "storyboard", Hint: "选集生成分镜，可续跑"},
		{ID: "images", Label: "分镜图片", Ready: in.HasImages, Entry: "media", Hint: "媒体面板图片页签生成"},
		{ID: "videos", Label: "镜头视频", Ready: in.HasVideos, Entry: "media", Hint: "媒体面板视频页签基于图片生成"},
		{ID: "subtitles", Label: "字幕", Ready: in.HasSubtitles, Entry: "media", Hint: "媒体面板识别字幕并编辑"},
		{ID: "final_cut", Label: "成片合成", Ready: in.HasFinalCut, Entry: "media", Hint: "媒体面板合成 MP4 成片"},
	}
}

// AssistantCoverage 是覆盖度自检结果：给出完成度与第一个待办能力。
type AssistantCoverage struct {
	Total     int    `json:"total"`
	Ready     int    `json:"ready"`
	Missing   int    `json:"missing"`
	NextID    string `json:"next_id,omitempty"`
	NextLabel string `json:"next_label,omitempty"`
	NextEntry string `json:"next_entry,omitempty"`
}

// BuildAssistantCoverage 从能力清单计算覆盖度，Next* 指向链路里第一个未就绪环节。
func BuildAssistantCoverage(caps []AssistantCapability) AssistantCoverage {
	cov := AssistantCoverage{Total: len(caps)}
	for _, cap := range caps {
		if cap.Ready {
			cov.Ready++
			continue
		}
		cov.Missing++
		if cov.NextID == "" {
			cov.NextID = cap.ID
			cov.NextLabel = cap.Label
			cov.NextEntry = cap.Entry
		}
	}
	return cov
}

// FormatCapabilitiesForPrompt 把能力清单与覆盖度格式化为注入模型的事实上下文。
// 只描述「当前作品里什么已就绪、下一步去哪个面板」，不含任何可执行工具签名。
func FormatCapabilitiesForPrompt(caps []AssistantCapability, cov AssistantCoverage) string {
	var b strings.Builder
	b.WriteString("【当前作品链路状态（只读事实，不要当成已完成的操作）】\n")
	for _, cap := range caps {
		state := "未就绪"
		if cap.Ready {
			state = "已就绪"
		}
		fmt.Fprintf(&b, "- %s：%s（入口：%s）", cap.Label, state, cap.Entry)
		if cap.Hint != "" {
			fmt.Fprintf(&b, "；%s", cap.Hint)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "覆盖度：%d/%d 环节已就绪，还缺 %d 个；建议下一步：%s（入口：%s）。\n",
		cov.Ready, cov.Total, cov.Missing, cov.NextLabel, cov.NextEntry)
	b.WriteString("你可以基于以上状态回答用户，或在信息不足时主动追问（question 字段），并给出导航建议（suggestions）。")
	b.WriteString("不要声称已经执行了任何生成或修改；任何内容改动仍需用户确认 proposal。")
	return b.String()
}
