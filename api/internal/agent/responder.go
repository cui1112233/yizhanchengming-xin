package agent

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var videoDurationPattern = regexp.MustCompile(`(?i)(\d{1,2})\s*(?:秒|s|sec|seconds?)`)

type ResponseContext struct {
	Owner         string
	ThreadID      string
	Input         string
	MediaAssetIDs []string
	Messages      []Message
	Tasks         []Task
}

type Responder struct{}

func NewResponder() *Responder { return &Responder{} }

func (r *Responder) Respond(_ context.Context, input ResponseContext) (AgentResponse, error) {
	text := strings.TrimSpace(input.Input)
	lower := strings.ToLower(text)
	if text == "" {
		return AgentResponse{Content: "你可以直接告诉我想做什么，例如“帮我打开小说获取”或“图片模型在哪里设置”。"}, nil
	}

	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
		return AgentResponse{Content: "为了安全，我不会把任意外部网址当成系统导航执行。你可以直接说要打开一战晟铭里的哪个功能，例如小说获取、剧本生成、创作漫剧或批量工厂。"}, nil
	}

	if tool, ok := videoToolForRequest(text, input.MediaAssetIDs); ok {
		return AgentResponse{Content: "可以，我会按当前引用素材和你的要求生成视频。生成完成后会直接在这里展示，并保存到 TOS。", Tool: tool}, nil
	}

	if title, ok := taskTitle(text); ok {
		return AgentResponse{
			Content: "已把这个事项记录到当前对话的任务区，你可以继续告诉我具体怎么处理。",
			Task: &TaskProposal{Title: title, Status: TaskInProgress},
		}, nil
	}

	if strings.Contains(lower, "图片模型") && (strings.Contains(text, "哪里") || strings.Contains(text, "设置") || strings.Contains(text, "打开")) {
		return AgentResponse{
			Content: "图片模型在「API 配置」里管理。我可以直接带你到图片模型区域。",
			Tool: navigationTool(Destination{Name: "API 配置", Path: "/api-config"}, "image-models"),
		}, nil
	}
	if strings.Contains(lower, "视频模型") && (strings.Contains(text, "哪里") || strings.Contains(text, "设置") || strings.Contains(text, "打开")) {
		return AgentResponse{
			Content: "视频模型在「API 配置」里管理。我可以直接带你到视频模型区域。",
			Tool: navigationTool(Destination{Name: "API 配置", Path: "/api-config"}, "video-models"),
		}, nil
	}
	if strings.Contains(lower, "文本模型") && (strings.Contains(text, "哪里") || strings.Contains(text, "设置") || strings.Contains(text, "打开")) {
		return AgentResponse{
			Content: "文本模型在「API 配置」里管理。我可以直接带你到文本模型区域。",
			Tool: navigationTool(Destination{Name: "API 配置", Path: "/api-config"}, "text-models"),
		}, nil
	}

	if destination, ok := destinationForText(text); ok && wantsNavigation(text) {
		return AgentResponse{
			Content: "可以，我已经准备打开「" + destination.Name + "」。",
			Tool: navigationTool(destination, ""),
		}, nil
	}

	if strings.Contains(text, "在哪里") || strings.Contains(text, "哪儿") || strings.Contains(text, "怎么进") {
		if destination, ok := destinationForText(text); ok {
			return AgentResponse{
				Content: "「" + destination.Name + "」就在系统工作区里，我可以直接帮你打开。",
				Tool: navigationTool(destination, ""),
			}, nil
		}
	}

	if strings.Contains(text, "能做什么") || strings.Contains(lower, "help") || text == "帮助" {
		return AgentResponse{Content: "当前 Agent 已经能帮你打开系统功能、回答常用位置问题、记录任务、携带媒体资产引用，并能把视频生成任务交给统一 Go Worker 执行。生成结果会统一保存到 TOS 并直接回到当前聊天。"}, nil
	}

	if strings.Contains(text, "图片") || strings.Contains(text, "修改") || strings.Contains(text, "生成") {
		return AgentResponse{Content: "我理解这是一个创作执行请求。视频生成已经接入统一工具链；图片生成和图片修改仍在接入同一个 Tool Registry，我不会假装已经生成。"}, nil
	}

	return AgentResponse{Content: "我可以直接帮你操作系统入口、解释设置位置、记录任务，并把执行结果放在同一个对话里。你可以更直接地说，例如“打开小说获取”“用这张图生成10秒视频”“创建任务：检查第 8 个视频失败原因”。"}, nil
}

func deterministicIntent(input string) bool {
	text := strings.TrimSpace(input)
	lower := strings.ToLower(text)
	if text == "" || strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
		return true
	}
	if _, ok := videoToolForRequest(text, nil); ok {
		return true
	}
	if _, ok := taskTitle(text); ok {
		return true
	}
	for _, kind := range []string{"图片模型", "视频模型", "文本模型"} {
		if strings.Contains(lower, kind) && (strings.Contains(text, "哪里") || strings.Contains(text, "设置") || strings.Contains(text, "打开")) {
			return true
		}
	}
	if destination, ok := destinationForText(text); ok && (wantsNavigation(text) || strings.Contains(text, "在哪里") || strings.Contains(text, "哪儿") || strings.Contains(text, "怎么进")) {
		_ = destination
		return true
	}
	return strings.Contains(text, "能做什么") || strings.Contains(lower, "help") || text == "帮助"
}

func videoToolForRequest(text string, mediaAssetIDs []string) (*ToolProposal, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if trimmed == "" { return nil, false }
	videoIntent := strings.Contains(trimmed, "生成视频") || strings.Contains(trimmed, "生视频") || strings.Contains(trimmed, "做视频") || strings.Contains(lower, "generate video") || strings.Contains(lower, "video generate")
	if !videoIntent { return nil, false }
	model := "yd2-mini-video"
	switch {
	case strings.Contains(lower, "h3") || strings.Contains(lower, "minimax"):
		model = "minimax-h3-video"
	case strings.Contains(lower, "yfai") || strings.Contains(lower, "seedance"):
		model = "seedance-2-0-official"
	case strings.Contains(trimmed, "豆包") || strings.Contains(trimmed, "本地执行器") || strings.Contains(lower, "local"):
		model = "local-doubao-executor-video"
	}
	duration := 5
	if match := videoDurationPattern.FindStringSubmatch(trimmed); len(match) == 2 {
		if parsed, err := strconv.Atoi(match[1]); err == nil && parsed > 0 { duration = parsed }
	}
	refs, _ := normalizeMediaAssetIDs(mediaAssetIDs)
	args := videoGenerateArguments{
		Model: model,
		Prompt: trimmed,
		Duration: duration,
		AspectRatio: "9:16",
		Resolution: "720p",
		ReferenceMediaAssetIDs: refs,
	}
	body, _ := json.Marshal(args)
	return &ToolProposal{Name: ToolVideoGenerate, Label: "生成视频", Arguments: body}, true
}

func wantsNavigation(text string) bool {
	for _, token := range []string{"打开", "进入", "跳到", "去", "带我到", "帮我开", "帮我打开", "在哪", "哪里"} {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func taskTitle(text string) (string, bool) {
	trimmed := strings.TrimSpace(text)
	prefixes := []string{"创建任务：", "创建任务:", "创建任务 ", "帮我记个任务：", "帮我记个任务:", "记个任务：", "记个任务:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if title == "" {
				return "", false
			}
			runes := []rune(title)
			if len(runes) > 120 {
				runes = runes[:120]
			}
			return string(runes), true
		}
	}
	return "", false
}
