package agent

import (
	"context"
	"strings"
)

type Responder struct{}

func NewResponder() *Responder { return &Responder{} }

func (r *Responder) Respond(_ context.Context, _ string, _ string, input string) (AgentResponse, error) {
	text := strings.TrimSpace(input)
	lower := strings.ToLower(text)
	if text == "" {
		return AgentResponse{Content: "你可以直接告诉我想做什么，例如“帮我打开小说获取”或“图片模型在哪里设置”。"}, nil
	}

	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
		return AgentResponse{Content: "为了安全，我不会把任意外部网址当成系统导航执行。你可以直接说要打开一战晟铭里的哪个功能，例如小说获取、剧本生成、创作漫剧或批量工厂。"}, nil
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
		return AgentResponse{Content: "当前 Agent 已经能帮你打开系统功能、回答常用位置问题、记录任务，并在聊天中携带媒体资产引用。你可以试着说：“帮我打开小说获取”“图片模型在哪里设置”“创建任务：检查今天的批量项目”。图片修改、直接生图、生视频等创作工具会继续接入同一个工具系统，不会另做第二套流程。"}, nil
	}

	if strings.Contains(text, "图片") || strings.Contains(text, "视频") || strings.Contains(text, "生成") || strings.Contains(text, "修改") {
		return AgentResponse{Content: "我已经理解这是一个创作执行请求，但当前这个 Agent V1 还没有把图片/视频生成 Provider 接到工具注册表里。我不会假装任务已经生成。你可以先把相关事项记录成任务，或者让我打开创作漫剧、批量工厂、API 配置等现有工作区。"}, nil
	}

	return AgentResponse{Content: "我可以直接帮你操作系统入口、解释设置位置、记录任务，并把后续执行结果放在同一个对话里。你可以更直接地说，例如“打开小说获取”“批量工厂在哪里”“创建任务：检查第 8 个视频失败原因”。"}, nil
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
