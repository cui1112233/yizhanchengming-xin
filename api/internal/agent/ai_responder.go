package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const maxAgentHistoryMessages = 24
const maxAgentHistoryRunes = 48000

var agentSectionPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

const agentSystemPrompt = `你是一战晟铭的系统 Agent。你要用简洁中文帮助用户理解和操作系统。

当前可直接执行的工具只有“系统内部导航”和“记录任务”。不要声称已经生成图片、视频、剧本、小说或发布内容，除非工具系统明确提供了对应执行结果。
正式图片/视频都是 TOS 媒体资产，消息里只引用 media_asset_id，不要编造本地文件路径。

允许导航的系统页面：
- /agent Agent 工作区
- /novel-fetch 小说获取
- /script 剧本生成
- /shuihuo-production 水货生产
- /shuihuo-production/creative 创作漫剧
- /batch-factory 批量工厂
- /settings 设置
- /api-config API 配置

只返回一个 JSON 对象，不要 Markdown，不要额外文字：
{
  "content":"给用户看的回复",
  "navigate":{"path":"允许的内部路径","section":"可选区域"},
  "task":{"title":"可选任务标题","status":"in_progress|waiting|needs_decision|completed|failed","detail":"可选说明"}
}
不需要导航时省略 navigate；不需要创建任务时省略 task。不得输出外部 URL 作为 navigate.path。`

type AIResponderConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type AIResponder struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []chatCompletionMessage `json:"messages"`
	Temperature float64                 `json:"temperature"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type aiAgentDecision struct {
	Content  string         `json:"content"`
	Navigate *NavigateArgs  `json:"navigate,omitempty"`
	Task     *TaskProposal  `json:"task,omitempty"`
}

func NewAIResponder(cfg AIResponderConfig) (*AIResponder, error) {
	endpoint, err := agentChatCompletionsURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	apiKey := strings.TrimSpace(cfg.APIKey)
	model := strings.TrimSpace(cfg.Model)
	if apiKey == "" {
		return nil, errors.New("agent AI API key is required")
	}
	if model == "" {
		return nil, errors.New("agent AI model is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	return &AIResponder{endpoint: endpoint, apiKey: apiKey, model: model, client: client}, nil
}

func agentChatCompletionsURL(baseURL string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return "", errors.New("agent AI base URL is required")
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return "", errors.New("agent AI base URL must use http or https")
	}
	lower := strings.ToLower(trimmed)
	switch {
	case strings.HasSuffix(lower, "/chat/completions"):
		return trimmed, nil
	case strings.HasSuffix(lower, "/v1"):
		return trimmed + "/chat/completions", nil
	default:
		return trimmed + "/v1/chat/completions", nil
	}
}

func (r *AIResponder) Respond(ctx context.Context, input ResponseContext) (AgentResponse, error) {
	messages := []chatCompletionMessage{{Role: "system", Content: agentSystemPrompt}}
	for _, item := range recentAgentMessages(input.Messages) {
		role := item.Role
		if role != RoleUser && role != RoleAssistant {
			continue
		}
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		messages = append(messages, chatCompletionMessage{Role: role, Content: content})
	}
	current := strings.TrimSpace(input.Input)
	if len(input.MediaAssetIDs) > 0 {
		current += "\n\n当前显式引用的媒体资产：" + strings.Join(input.MediaAssetIDs, "、")
	}
	if current == "" {
		current = "请根据当前上下文继续。"
	}
	if len(messages) == 1 || messages[len(messages)-1].Role != RoleUser || messages[len(messages)-1].Content != current {
		messages = append(messages, chatCompletionMessage{Role: RoleUser, Content: current})
	}

	payload := chatCompletionRequest{Model: r.model, Messages: messages, Temperature: 0.25, MaxTokens: 1000}
	body, err := json.Marshal(payload)
	if err != nil {
		return AgentResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(body))
	if err != nil {
		return AgentResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	resp, err := r.client.Do(req)
	if err != nil {
		return AgentResponse{}, fmt.Errorf("agent AI request: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return AgentResponse{}, fmt.Errorf("read agent AI response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return AgentResponse{}, fmt.Errorf("agent AI request failed: HTTP %d", resp.StatusCode)
	}
	var decoded chatCompletionResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return AgentResponse{}, fmt.Errorf("decode agent AI response: %w", err)
	}
	if decoded.Error != nil && strings.TrimSpace(decoded.Error.Message) != "" {
		return AgentResponse{}, errors.New(decoded.Error.Message)
	}
	if len(decoded.Choices) == 0 {
		return AgentResponse{}, errors.New("agent AI response has no choices")
	}
	return parseAgentDecision(decoded.Choices[0].Message.Content)
}

func recentAgentMessages(messages []Message) []Message {
	if len(messages) > maxAgentHistoryMessages {
		messages = messages[len(messages)-maxAgentHistoryMessages:]
	}
	remaining := maxAgentHistoryRunes
	selected := make([]Message, 0, len(messages))
	for index := len(messages) - 1; index >= 0 && remaining > 0; index-- {
		item := messages[index]
		runes := []rune(item.Content)
		if len(runes) > remaining {
			runes = runes[len(runes)-remaining:]
			item.Content = string(runes)
		}
		remaining -= len(runes)
		selected = append(selected, item)
	}
	for left, right := 0, len(selected)-1; left < right; left, right = left+1, right-1 {
		selected[left], selected[right] = selected[right], selected[left]
	}
	return selected
}

func parseAgentDecision(content string) (AgentResponse, error) {
	cleaned := strings.TrimSpace(content)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```JSON")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start < 0 || end < start {
		return AgentResponse{}, errors.New("agent AI response does not contain JSON")
	}
	var decision aiAgentDecision
	if err := json.Unmarshal([]byte(cleaned[start:end+1]), &decision); err != nil {
		return AgentResponse{}, fmt.Errorf("decode agent decision: %w", err)
	}
	decision.Content = strings.TrimSpace(decision.Content)
	if decision.Content == "" {
		decision.Content = "我已经理解你的请求。"
	}
	result := AgentResponse{Content: decision.Content}
	if decision.Navigate != nil && IsAllowedNavigation(strings.TrimSpace(decision.Navigate.Path)) {
		section := strings.TrimSpace(decision.Navigate.Section)
		if section != "" && !agentSectionPattern.MatchString(section) {
			section = ""
		}
		destination := Destination{Name: destinationName(decision.Navigate.Path), Path: strings.TrimSpace(decision.Navigate.Path)}
		result.Tool = navigationTool(destination, section)
	}
	if decision.Task != nil {
		title := strings.TrimSpace(decision.Task.Title)
		status := strings.TrimSpace(decision.Task.Status)
		if title != "" && validTaskStatus(status) {
			result.Task = &TaskProposal{Title: cleanTitle(title), Status: status, Detail: strings.TrimSpace(decision.Task.Detail)}
		}
	}
	return result, nil
}

func destinationName(path string) string {
	for _, destination := range destinations {
		if destination.Path == path {
			return destination.Name
		}
	}
	return "系统页面"
}

type HybridResponder struct {
	Deterministic *Responder
	AI            *AIResponder
}

func NewHybridResponder(deterministic *Responder, ai *AIResponder) *HybridResponder {
	if deterministic == nil {
		deterministic = NewResponder()
	}
	return &HybridResponder{Deterministic: deterministic, AI: ai}
}

func (h *HybridResponder) Respond(ctx context.Context, input ResponseContext) (AgentResponse, error) {
	if h == nil {
		return AgentResponse{}, errors.New("agent responder unavailable")
	}
	if deterministicIntent(input.Input) || h.AI == nil {
		return h.Deterministic.Respond(ctx, input)
	}
	response, err := h.AI.Respond(ctx, input)
	if err == nil {
		return response, nil
	}
	return h.Deterministic.Respond(ctx, input)
}
