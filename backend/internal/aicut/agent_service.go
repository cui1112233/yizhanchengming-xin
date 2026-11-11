package aicut

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AgentModelRequest struct {
	SystemPrompt string
	UserPrompt   string
}

type AgentModel interface {
	Complete(context.Context, AgentModelRequest) (string, error)
}

type AgentModelFunc func(context.Context, AgentModelRequest) (string, error)

func (fn AgentModelFunc) Complete(ctx context.Context, request AgentModelRequest) (string, error) {
	return fn(ctx, request)
}

type AgentModelResolver func(context.Context, string, string) (AgentModel, error)
type AgentPromptBuilder func(AgentMode) string

type AgentService struct {
	store        AgentStore
	tools        *AgentToolExecutor
	resolveModel AgentModelResolver
	buildPrompt  AgentPromptBuilder
	now          func() time.Time
}

func NewAgentService(store AgentStore, tools *AgentToolExecutor, resolver AgentModelResolver, promptBuilder AgentPromptBuilder, now func() time.Time) *AgentService {
	if now == nil {
		now = time.Now
	}
	return &AgentService{store: store, tools: tools, resolveModel: resolver, buildPrompt: promptBuilder, now: now}
}

type agentModelEnvelope struct {
	Reply     string     `json:"reply"`
	ToolCalls []ToolCall `json:"toolCalls"`
	Followup  any        `json:"followup"`
}

func (s *AgentService) appendEvent(ctx context.Context, run AgentRun, eventType AgentEventType, message, toolName, toolCallID string, revision int64) error {
	_, err := s.store.AppendAgentEvent(ctx, AgentEvent{RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Type: eventType, Message: message, ToolName: toolName, ToolCallID: toolCallID, Revision: revision, CreatedAt: s.now().UTC()})
	return err
}

func (s *AgentService) isCanceled(ctx context.Context, run AgentRun) (bool, AgentRun, error) {
	current, err := s.store.AgentRun(ctx, run.Owner, run.ProjectID, run.ID)
	if err != nil {
		return false, AgentRun{}, err
	}
	return current.State == AgentCancelRequested || current.State == AgentCanceled, current, nil
}

func (s *AgentService) finishCanceled(ctx context.Context, run AgentRun, turn int) (AgentRun, error) {
	finished, err := s.store.UpdateAgentRunState(ctx, run.ID, AgentCanceled, "已取消", turn, s.now().UTC())
	if err == nil {
		_ = s.appendEvent(ctx, run, AgentEventReply, "已取消", "", "", run.ExpectedRevision)
	}
	return finished, err
}

func (s *AgentService) fail(ctx context.Context, run AgentRun, turn int, message string, cause error) (AgentRun, error) {
	finished, updateErr := s.store.UpdateAgentRunState(ctx, run.ID, AgentFailed, message, turn, s.now().UTC())
	if updateErr != nil {
		return AgentRun{}, updateErr
	}
	_ = s.appendEvent(ctx, run, AgentEventReply, message, "", "", run.ExpectedRevision)
	if cause == nil {
		cause = errors.New(message)
	}
	return finished, cause
}

func (s *AgentService) Run(ctx context.Context, runID string) (AgentRun, error) {
	run, err := s.store.ClaimAgentRun(ctx, runID, s.now().UTC())
	if err != nil {
		return AgentRun{}, err
	}
	model, err := s.resolveModel(ctx, run.Owner, run.ModelID)
	if err != nil || model == nil {
		return s.fail(ctx, run, 0, "文本模型未配置，请先到 API 配置启用文本模型", err)
	}
	userMessage := AgentMessage{ID: run.ID + "-user", RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleUser, Content: run.Instruction, CreatedAt: s.now().UTC()}
	if err := s.store.AppendAgentMessage(ctx, userMessage); err != nil {
		return s.fail(ctx, run, 0, "保存用户指令失败", err)
	}
	conversation := []string{"用户指令：" + run.Instruction}
	for turn := 1; turn <= 30; turn++ {
		canceled, _, err := s.isCanceled(ctx, run)
		if err != nil {
			return AgentRun{}, err
		}
		if canceled {
			return s.finishCanceled(ctx, run, turn-1)
		}
		if err := s.appendEvent(ctx, run, AgentEventThinking, fmt.Sprintf("正在分析第 %d 步", turn), "", "", run.ExpectedRevision); err != nil {
			return s.fail(ctx, run, turn-1, "保存 Agent 事件失败", err)
		}
		request := AgentModelRequest{SystemPrompt: s.buildPrompt(run.Mode), UserPrompt: strings.Join(conversation, "\n")}
		raw, modelErr := model.Complete(ctx, request)
		if modelErr != nil {
			raw, modelErr = model.Complete(ctx, AgentModelRequest{SystemPrompt: request.SystemPrompt, UserPrompt: request.UserPrompt + "\n上一轮模型调用失败，请仅重试一次。"})
		}
		if modelErr != nil {
			return s.fail(ctx, run, turn, "模型调用失败："+modelErr.Error(), modelErr)
		}
		var envelope agentModelEnvelope
		if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
			repairRaw, repairErr := model.Complete(ctx, AgentModelRequest{SystemPrompt: request.SystemPrompt, UserPrompt: request.UserPrompt + "\n上一轮输出不是合法 JSON。现在只输出约定 JSON 对象。"})
			if repairErr != nil {
				return s.fail(ctx, run, turn, "模型输出无法解析", repairErr)
			}
			if err := json.Unmarshal([]byte(repairRaw), &envelope); err != nil {
				return s.fail(ctx, run, turn, "模型输出无法解析", err)
			}
		}
		if run.Mode == AgentModeAsk {
			reply := strings.TrimSpace(envelope.Reply)
			if reply == "" {
				reply = "问答模式不会改时间轴。请描述你想了解的剪辑问题。"
			}
			if !strings.Contains(reply, "问答模式不会改时间轴") {
				reply = "问答模式不会改时间轴。" + reply
			}
			return s.finishReply(ctx, run, turn, reply)
		}
		if len(envelope.ToolCalls) == 0 {
			reply := strings.TrimSpace(envelope.Reply)
			if reply == "" {
				return s.fail(ctx, run, turn, "模型没有返回可执行工具或说明", nil)
			}
			return s.finishReply(ctx, run, turn, reply)
		}
		for _, call := range envelope.ToolCalls {
			canceled, _, err := s.isCanceled(ctx, run)
			if err != nil {
				return AgentRun{}, err
			}
			if canceled {
				return s.finishCanceled(ctx, run, turn)
			}
			if err := s.appendEvent(ctx, run, AgentEventToolStart, "开始执行 "+call.Name, call.Name, call.ID, run.ExpectedRevision); err != nil {
				return s.fail(ctx, run, turn, "保存工具事件失败", err)
			}
			result, err := s.tools.Execute(ctx, run, call)
			if err != nil {
				return s.fail(ctx, run, turn, "工具执行失败："+err.Error(), err)
			}
			encoded, _ := json.Marshal(result)
			conversation = append(conversation, "工具结果："+string(encoded))
			message := AgentMessage{ID: fmt.Sprintf("%s-tool-%d-%s", run.ID, turn, call.ID), RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleTool, Content: string(encoded), ToolName: call.Name, ToolCallID: call.ID, CreatedAt: s.now().UTC()}
			if err := s.store.AppendAgentMessage(ctx, message); err != nil {
				return s.fail(ctx, run, turn, "保存工具结果失败", err)
			}
			eventType := AgentEventToolDone
			if !result.OK {
				eventType = AgentEventToolFailed
			}
			if err := s.appendEvent(ctx, run, eventType, result.Message, call.Name, call.ID, result.Revision); err != nil {
				return s.fail(ctx, run, turn, "保存工具结果事件失败", err)
			}
			if result.OK && result.Revision > 0 {
				run.ExpectedRevision = result.Revision
			}
		}
	}
	return s.fail(ctx, run, 30, "执行步骤过多，请缩小任务范围", nil)
}

func (s *AgentService) finishReply(ctx context.Context, run AgentRun, turn int, reply string) (AgentRun, error) {
	message := AgentMessage{ID: fmt.Sprintf("%s-assistant-%d", run.ID, turn), RunID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, Role: AgentRoleAssistant, Content: reply, CreatedAt: s.now().UTC()}
	if err := s.store.AppendAgentMessage(ctx, message); err != nil {
		return s.fail(ctx, run, turn, "保存 Agent 回复失败", err)
	}
	if err := s.appendEvent(ctx, run, AgentEventReply, reply, "", "", run.ExpectedRevision); err != nil {
		return s.fail(ctx, run, turn, "保存 Agent 回复事件失败", err)
	}
	return s.store.UpdateAgentRunState(ctx, run.ID, AgentSucceeded, "", turn, s.now().UTC())
}
