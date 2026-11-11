package agentchat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Provider is the minimal text-completion contract the loop needs. The HTTP
// layer adapts the existing model catalog provider to it.
type Provider interface {
	Complete(ctx context.Context, system, user string, maxTokens int) (string, error)
}

// Runner executes one Agent turn: model call, tool loop, permission gating and
// event streaming.
type Runner struct {
	Store       Store
	ProviderFor func(ctx context.Context, owner, modelID string) (Provider, error)
	MaxRounds   int
	Now         func() time.Time

	locksMu sync.Mutex
	locks   map[int64]*sync.Mutex
	// SkillsFor resolves the Tier-1 catalog per request. Returning a function
	// (instead of a shared field) keeps concurrent owners race-free.
	SkillsFor func(owner string) *SkillCatalog
}

// ErrSessionBusy is returned when the same session is already running a turn.
var ErrSessionBusy = errors.New("当前会话正在执行中，请等待完成或先停止。")

// RunError 是一次 Agent 回合失败的可展示错误。Error 只返回中文结论，
// 上游原文放在 Detail 里，通过 error 事件的 detail 字段交给界面折叠展示，
// 不再把失败伪装成一条内容为「请求失败」的正常回复。
type RunError struct {
	Headline string
	Detail   string
	Cause    error
}

func (e *RunError) Error() string { return e.Headline }

func (e *RunError) Unwrap() error { return e.Cause }

// classifyProviderError 把模型通道错误归入「缺配置 / 鉴权 / 限流 / 上游」四类。
// provider 已经给出「中文结论 + \n上游原文：...」时直接拆开复用；其余按关键字
// 归类，主文案只留中文结论，原始报错进 Detail。
func classifyProviderError(err error) (headline, detail string) {
	if err == nil {
		return "", ""
	}
	text := strings.TrimSpace(err.Error())
	if conclusion, raw, found := strings.Cut(text, "\n上游原文："); found {
		if strings.TrimSpace(raw) == "" {
			raw = text
		}
		return strings.TrimSpace(conclusion), strings.TrimSpace(raw)
	}
	detail = text
	lower := strings.ToLower(text)
	switch {
	case errors.Is(err, context.Canceled):
		return "已停止", detail
	case errors.Is(err, context.DeadlineExceeded), strings.Contains(lower, "timeout"), strings.Contains(text, "超时"):
		return "文本模型请求超时，请稍后重试。", detail
	case strings.Contains(text, "模型设置"), strings.Contains(text, "未配置"), strings.Contains(text, "模型未就绪"):
		// provider.Validate 已经给出可执行的中文配置提示，直接作为结论。
		return text, detail
	case strings.Contains(lower, "not configured"), strings.Contains(lower, "catalog is unavailable"):
		return "文本模型未配置：请到「设置 → 模型」选择一个可用的文本模型后重试。", detail
	case strings.Contains(lower, "401"), strings.Contains(lower, "403"),
		strings.Contains(lower, "unauthorized"), strings.Contains(lower, "forbidden"),
		strings.Contains(lower, "invalid api key"), strings.Contains(lower, "invalid_api_key"),
		strings.Contains(text, "鉴权"), strings.Contains(text, "密钥"):
		return "文本模型鉴权失败，请检查模型密钥后重试。", detail
	case strings.Contains(lower, "429"), strings.Contains(lower, "rate limit"),
		strings.Contains(lower, "too many requests"), strings.Contains(text, "限流"):
		return "上游文本模型触发限流，请稍后重试。", detail
	default:
		return "文本模型通道暂时不可用（上游异常），请稍后重试。", detail
	}
}

func emitAgentFailure(emit func(Event), headline, detail string) {
	payload := map[string]any{"message": headline}
	if detail != "" {
		payload["detail"] = detail
	}
	emitEvent(emit, EventError, payload)
}

const roundLimitMessage = "已达到内部工具调用上限。已完成的工作进度已保存在工作区中。你可以发送「继续」让我接着上次的进度继续工作。"

// NewRunner builds a runner with sane defaults.
func NewRunner(store Store, providerFor func(ctx context.Context, owner, modelID string) (Provider, error), clock func() time.Time) *Runner {
	if clock == nil {
		clock = time.Now
	}
	return &Runner{
		Store:       store,
		ProviderFor: providerFor,
		MaxRounds:   8,
		Now:         clock,
		locks:       map[int64]*sync.Mutex{},
	}
}

func (r *Runner) sessionLock(id int64) *sync.Mutex {
	r.locksMu.Lock()
	defer r.locksMu.Unlock()
	if r.locks == nil {
		r.locks = map[int64]*sync.Mutex{}
	}
	lock, ok := r.locks[id]
	if !ok {
		lock = &sync.Mutex{}
		r.locks[id] = lock
	}
	return lock
}

// Run executes one turn and emits events as they happen. The user message is
// persisted before the model call so a failed request never loses input.
func (r *Runner) Run(ctx context.Context, owner string, req SendRequest, emit func(Event)) (Message, error) {
	if strings.TrimSpace(req.Content) == "" {
		return Message{}, errors.New("消息不能为空")
	}
	if strings.TrimSpace(req.WorkspaceID) == "" {
		return Message{}, errors.New("请先选择工作区")
	}
	workspace, err := r.Store.Workspace(ctx, owner, req.WorkspaceID)
	if err != nil {
		return Message{}, errors.New("工作区不存在")
	}
	session, err := r.Store.Session(ctx, owner, req.SessionID)
	if err != nil {
		return Message{}, err
	}
	lock := r.sessionLock(req.SessionID)
	if !lock.TryLock() {
		return Message{}, ErrSessionBusy
	}
	defer lock.Unlock()

	mode := NormalizePermission(req.Permission)
	if req.Permission == "" {
		mode = NormalizePermission(session.PermissionMode)
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = session.Model
	}
	emitEvent(emit, EventAgentStatus, map[string]any{"status": "connecting", "text": "正在连接 Agent…"})

	history, err := r.Store.Messages(ctx, owner, req.SessionID)
	if err != nil {
		return Message{}, err
	}
	now := r.Now()
	userMessage := EmptyMessage("user")
	userMessage.ID = now.UnixMilli()
	userMessage.Content = strings.TrimSpace(req.Content)
	userMessage.Timestamp = clockStamp(now)
	userMessage.Model = model
	if len(req.Attachments) > 0 {
		userMessage.Attachments = req.Attachments
	}
	history = append(history, userMessage)
	if err := r.Store.SaveMessages(ctx, owner, req.SessionID, history); err != nil {
		return Message{}, err
	}

	provider, err := r.ProviderFor(ctx, owner, model)
	if err != nil {
		headline, detail := classifyProviderError(err)
		failure := EmptyMessage("assistant")
		failure.ID = r.Now().UnixMilli()
		failure.Content = headline
		failure.Timestamp = clockStamp(r.Now())
		failure.Model = model
		_ = r.appendMessages(owner, req.SessionID, failure)
		emitAgentFailure(emit, headline, detail)
		return failure, &RunError{Headline: headline, Detail: detail, Cause: err}
	}

	catalog := r.skillCatalog(owner, req.EnabledSkillIDs)
	systemPrompt := buildSystemPrompt(workspace, session, mode, catalog)
	transcript := buildTranscript(history)

	assistant := EmptyMessage("assistant")
	assistant.ID = now.UnixMilli() + 1
	assistant.Model = model
	rounds := r.MaxRounds
	if rounds <= 0 {
		rounds = 8
	}
	emitEvent(emit, EventAgentStatus, map[string]any{"status": "running", "text": "正在处理…"})

	var finalText string
	canContinue := false
	for round := 0; round < rounds; round++ {
		reply, callErr := provider.Complete(ctx, systemPrompt, transcript, 4096)
		if callErr != nil {
			if ctx.Err() != nil {
				assistant.Content = "已停止"
				break
			}
			headline, detail := classifyProviderError(callErr)
			assistant.Content = headline
			assistant.Timestamp = clockStamp(r.Now())
			assistant.StreamSegments = []map[string]any{{"type": "text"}}
			if len(assistant.ToolBlocks) > 0 {
				assistant.StreamSegments = append([]map[string]any{{"type": "tools"}}, assistant.StreamSegments...)
			}
			streamText(emit, assistant.Content)
			_ = r.appendMessages(owner, req.SessionID, assistant)
			r.writeAudit(workspace, session, model, mode, userMessage.Content, assistant)
			emitAgentFailure(emit, headline, detail)
			return assistant, &RunError{Headline: headline, Detail: detail, Cause: callErr}
		}
		toolName, args, isCall := parseToolCall(reply)
		if !isCall {
			finalText = strings.TrimSpace(reply)
			break
		}
		block := r.executeTool(ctx, workspace, catalog, mode, toolName, args, emit)
		assistant.ToolBlocks = append(assistant.ToolBlocks, block)
		transcript += fmt.Sprintf("\n\n工具调用：%s\n参数：%s\n工具结果（%s）：\n%s\n", toolName, jsonText(args), block.Status, toolObservation(block))
		if round == rounds-1 {
			canContinue = true
			finalText = roundLimitMessage
		}
	}
	if canContinue {
		assistant.Content = finalText
		assistant.CanContinue = true
	} else if finalText != "" {
		assistant.Content = finalText
	} else if assistant.Content == "" {
		assistant.Content = "执行失败"
	}
	assistant.Timestamp = clockStamp(r.Now())
	assistant.StreamSegments = []map[string]any{{"type": "text"}}
	if len(assistant.ToolBlocks) > 0 {
		assistant.StreamSegments = append([]map[string]any{{"type": "tools"}}, assistant.StreamSegments...)
	}

	streamText(emit, assistant.Content)
	if err := r.appendMessages(owner, req.SessionID, assistant); err != nil {
		return assistant, err
	}
	r.writeAudit(workspace, session, model, mode, userMessage.Content, assistant)
	emitEvent(emit, EventDone, map[string]any{"message": assistant, "canContinue": assistant.CanContinue})
	return assistant, nil
}

func (r *Runner) appendMessages(owner string, sessionID int64, message Message) error {
	messages, err := r.Store.Messages(context.Background(), owner, sessionID)
	if err != nil {
		return err
	}
	messages = append(messages, message)
	return r.Store.SaveMessages(context.Background(), owner, sessionID, messages)
}

func (r *Runner) skillCatalog(owner string, enabled []string) *SkillCatalog {
	all := NewSkillCatalog(nil)
	if r.SkillsFor != nil {
		if resolved := r.SkillsFor(owner); resolved != nil {
			all = resolved
		}
	}
	if len(enabled) == 0 {
		return all
	}
	allowed := map[string]bool{}
	for _, id := range enabled {
		allowed[strings.TrimSpace(id)] = true
	}
	filtered := make([]Skill, 0, len(all.List()))
	for _, skill := range all.List() {
		if allowed[skill.ID] {
			filtered = append(filtered, skill)
		}
	}
	return NewSkillCatalog(filtered)
}

// SetSkills pins a static catalog. The HTTP layer uses SkillsFor instead so
// every owner resolves its own skills.
func (r *Runner) SetSkills(catalog *SkillCatalog) {
	if catalog == nil {
		catalog = NewSkillCatalog(nil)
	}
	r.SkillsFor = func(string) *SkillCatalog { return catalog }
}

func (r *Runner) executeTool(ctx context.Context, workspace Workspace, catalog *SkillCatalog, mode, name string, args map[string]any, emit func(Event)) ToolBlock {
	now := r.Now().UnixMilli()
	tool, ok := FindTool(name)
	block := ToolBlock{
		ID:        fmt.Sprintf("call_%d", now),
		ToolName:  name,
		Args:      args,
		Status:    ToolPlanned,
		StartedAt: now,
	}
	if ok {
		block.Name = tool.Label
	} else {
		block.Name = name
	}
	allowed := mode == PermissionBypass || (mode == PermissionAuto && IsSafeTool(name))
	if !allowed {
		block.Message = "已列入计划"
		block.Preview = "仅列入计划，未执行"
		block.CompletedAt = now
		emitEvent(emit, EventToolStart, block)
		emitEvent(emit, EventToolDone, block)
		return block
	}
	if !ok {
		block.Status = ToolFailed
		block.Error = "未知工具"
		block.Message = "执行失败"
		block.CompletedAt = now
		emitEvent(emit, EventToolStart, block)
		emitEvent(emit, EventToolDone, block)
		return block
	}
	block.Status = ToolRunning
	block.Message = "正在执行：" + block.Name
	emitEvent(emit, EventToolStart, block)
	result, runErr := tool.Run(ctx, ToolEnv{WorkspaceRoot: workspace.RootPath, Skills: catalog}, args)
	block.CompletedAt = r.Now().UnixMilli()
	if runErr != nil {
		block.Status = ToolFailed
		block.Error = runErr.Error()
		block.Message = "执行失败"
	} else {
		block.Status = ToolDone
		block.Summary = result.Summary
		block.ContentPreview = truncateRunes(result.Content, 2000)
		block.Preview = truncateRunes(result.Content, 400)
		block.Message = "已完成：" + block.Name
		if result.Artifacts != nil {
			block.Artifacts = result.Artifacts
		}
	}
	emitEvent(emit, EventToolProgress, map[string]any{"id": block.ID, "message": block.Message})
	emitEvent(emit, EventToolDone, block)
	return block
}

func (r *Runner) writeAudit(workspace Workspace, session SessionInfo, model, mode, userMessage string, assistant Message) {
	calls := make([]map[string]string, 0, len(assistant.ToolBlocks))
	for _, block := range assistant.ToolBlocks {
		calls = append(calls, map[string]string{"name": block.ToolName, "status": block.Status})
	}
	line, err := json.Marshal(map[string]any{
		"run_id":          fmt.Sprintf("%d", assistant.ID),
		"ts":              r.Now().Format(time.RFC3339),
		"workspace_id":    workspace.ID,
		"session_id":      fmt.Sprintf("%d", session.ID),
		"model":           model,
		"permission_mode": mode,
		"user_message":    userMessage,
		"reply":           assistant.Content,
		"tool_calls":      calls,
		"artifacts":       assistant.GeneratedFiles,
	})
	if err != nil {
		return
	}
	if store, ok := r.Store.(*FileStore); ok {
		_ = store.AppendRunAudit(workspace.RootPath, fmt.Sprintf("%d", assistant.ID), line)
	}
}

func buildSystemPrompt(workspace Workspace, session SessionInfo, mode string, catalog *SkillCatalog) string {
	var builder strings.Builder
	builder.WriteString("你是 Agent 工作区的执行代理，直接完成用户委托的工作。\n")
	fmt.Fprintf(&builder, "当前工作区：%s\n当前权限模式：%s（%s）\n", workspace.Name, PermissionLabel(mode), mode)
	switch mode {
	case PermissionPlan:
		builder.WriteString("计划模式：只输出工具使用计划，不要声称已执行任何工具。\n")
	case PermissionAuto:
		builder.WriteString("自动审批：只读工具可直接调用；写入类与 Skill 执行只能列入计划。\n")
	default:
		builder.WriteString("完全运行：所有工具都可以直接调用。\n")
	}
	builder.WriteString("\n可用工具：\n")
	builder.WriteString(ToolPrompt())
	builder.WriteString("\n可用技能（仅提供 id 与描述；需要正文时调用 skill_view）：\n")
	if len(catalog.List()) == 0 {
		builder.WriteString("- （无）\n")
	} else {
		for _, skill := range catalog.List() {
			fmt.Fprintf(&builder, "- %s：%s\n", skill.ID, skill.Description)
		}
	}
	builder.WriteString("\n输出协议：\n")
	builder.WriteString("1. 需要调用工具时，只输出一个 JSON 对象：{\"tool\":\"工具名\",\"args\":{...}}，不要输出其它文字。\n")
	builder.WriteString("2. 任务完成或只需回答时，直接用中文输出最终答复，不要输出 JSON。\n")
	return builder.String()
}

func buildTranscript(messages []Message) string {
	var builder strings.Builder
	for _, message := range messages {
		role := "用户"
		if message.Role == "assistant" {
			role = "助手"
		}
		fmt.Fprintf(&builder, "%s：%s\n", role, strings.TrimSpace(message.Content))
	}
	builder.WriteString("\n请回答上面的最后一条用户消息。")
	return builder.String()
}

func parseToolCall(reply string) (string, map[string]any, bool) {
	text := strings.TrimSpace(reply)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return "", nil, false
	}
	var payload struct {
		Type string         `json:"type"`
		Tool string         `json:"tool"`
		Name string         `json:"name"`
		Args map[string]any `json:"args"`
	}
	if json.Unmarshal([]byte(text[start:end+1]), &payload) != nil {
		return "", nil, false
	}
	name := strings.TrimSpace(payload.Tool)
	if name == "" && payload.Type == "tool_call" {
		name = strings.TrimSpace(payload.Name)
	}
	if name == "" {
		return "", nil, false
	}
	if payload.Args == nil {
		payload.Args = map[string]any{}
	}
	return name, payload.Args, true
}

func toolObservation(block ToolBlock) string {
	if block.Status == ToolPlanned {
		return "该工具在当前权限模式下仅列入计划，未执行。"
	}
	if block.Status == ToolFailed {
		return "执行失败：" + block.Error
	}
	return block.ContentPreview
}

func streamText(emit func(Event), text string) {
	if emit == nil || strings.TrimSpace(text) == "" {
		return
	}
	runes := []rune(text)
	const chunk = 160
	for start := 0; start < len(runes); start += chunk {
		end := start + chunk
		if end > len(runes) {
			end = len(runes)
		}
		emitEvent(emit, EventToken, map[string]any{"text": string(runes[start:end])})
	}
}

func jsonText(value map[string]any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "…"
}

func emitEvent(emit func(Event), eventType string, data any) {
	if emit == nil {
		return
	}
	emit(Event{Type: eventType, Data: data})
}
