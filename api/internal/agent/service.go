package agent

import (
	"context"
	"errors"
	"strings"
)

type ResponseGenerator interface {
	Respond(context.Context, ResponseContext) (AgentResponse, error)
}

type Service struct {
	Store     Store
	Responder ResponseGenerator
}

type SendMessageInput struct {
	Content       string   `json:"content"`
	MediaAssetIDs []string `json:"media_asset_ids,omitempty"`
}

type SendMessageResult struct {
	UserMessage      Message   `json:"user_message"`
	AssistantMessage Message   `json:"assistant_message"`
	Task             *Task     `json:"task,omitempty"`
	ToolCall         *ToolCall `json:"tool_call,omitempty"`
}

func (s *Service) CreateThread(ctx context.Context, owner, title string) (Thread, error) {
	if s == nil || s.Store == nil {
		return Thread{}, errors.New("agent store unavailable")
	}
	return s.Store.CreateThread(ctx, owner, title)
}

func (s *Service) ListThreads(ctx context.Context, owner string, limit int) ([]Thread, error) {
	if s == nil || s.Store == nil {
		return nil, errors.New("agent store unavailable")
	}
	return s.Store.ListThreads(ctx, owner, limit)
}

func (s *Service) GetThread(ctx context.Context, owner, threadID string) (ThreadSnapshot, error) {
	if s == nil || s.Store == nil {
		return ThreadSnapshot{}, errors.New("agent store unavailable")
	}
	return s.Store.GetThread(ctx, owner, threadID)
}

func (s *Service) ListTasks(ctx context.Context, owner, threadID string) ([]Task, error) {
	if s == nil || s.Store == nil {
		return nil, errors.New("agent store unavailable")
	}
	return s.Store.ListTasks(ctx, owner, threadID)
}

func (s *Service) UpdateTask(ctx context.Context, owner, taskID string, input UpdateTaskInput) (Task, error) {
	if s == nil || s.Store == nil {
		return Task{}, errors.New("agent store unavailable")
	}
	return s.Store.UpdateTask(ctx, owner, taskID, input)
}

func (s *Service) UpdateToolCall(ctx context.Context, owner, toolID string, input UpdateToolCallInput) (ToolCall, error) {
	if s == nil || s.Store == nil {
		return ToolCall{}, errors.New("agent store unavailable")
	}
	return s.Store.UpdateToolCall(ctx, owner, toolID, input)
}

func (s *Service) SendMessage(ctx context.Context, owner, threadID string, input SendMessageInput) (SendMessageResult, error) {
	if s == nil || s.Store == nil || s.Responder == nil {
		return SendMessageResult{}, errors.New("agent service unavailable")
	}
	snapshot, err := s.Store.GetThread(ctx, owner, threadID)
	if err != nil {
		return SendMessageResult{}, err
	}
	userMessage, err := s.Store.AppendMessage(ctx, AppendMessageInput{
		ThreadID: threadID,
		Owner: owner,
		Role: RoleUser,
		Content: input.Content,
		MediaAssetIDs: input.MediaAssetIDs,
	})
	if err != nil {
		return SendMessageResult{}, err
	}
	if snapshot.Thread.Title == "" || snapshot.Thread.Title == "新对话" {
		_ = s.Store.UpdateThreadTitle(ctx, owner, threadID, titleFromMessage(input.Content))
	}

	response, err := s.Responder.Respond(ctx, ResponseContext{
		Owner: owner,
		ThreadID: threadID,
		Input: input.Content,
		MediaAssetIDs: append([]string(nil), input.MediaAssetIDs...),
		Messages: append([]Message(nil), snapshot.Messages...),
		Tasks: append([]Task(nil), snapshot.Tasks...),
	})
	if err != nil {
		return SendMessageResult{}, err
	}
	assistantMessage, err := s.Store.AppendMessage(ctx, AppendMessageInput{
		ThreadID: threadID,
		Owner: owner,
		Role: RoleAssistant,
		Content: response.Content,
	})
	if err != nil {
		return SendMessageResult{}, err
	}
	result := SendMessageResult{UserMessage: userMessage, AssistantMessage: assistantMessage}
	if response.Task != nil {
		task, err := s.Store.CreateTask(ctx, CreateTaskInput{
			ThreadID: threadID,
			Owner: owner,
			Title: response.Task.Title,
			Status: response.Task.Status,
			Detail: response.Task.Detail,
		})
		if err != nil {
			return SendMessageResult{}, err
		}
		result.Task = &task
	}
	if response.Tool != nil {
		toolCall, err := s.Store.CreateToolCall(ctx, CreateToolCallInput{
			ThreadID: threadID,
			MessageID: assistantMessage.ID,
			Owner: owner,
			ToolName: response.Tool.Name,
			Status: ToolProposed,
			Arguments: response.Tool.Arguments,
		})
		if err != nil {
			return SendMessageResult{}, err
		}
		result.ToolCall = &toolCall
	}
	return result, nil
}

func titleFromMessage(content string) string {
	content = strings.Join(strings.Fields(strings.TrimSpace(content)), " ")
	if content == "" {
		return "新对话"
	}
	runes := []rune(content)
	if len(runes) > 28 {
		runes = runes[:28]
		return string(runes) + "…"
	}
	return string(runes)
}
