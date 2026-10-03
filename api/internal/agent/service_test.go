package agent

import (
	"context"
	"encoding/json"
	"testing"
)

type memoryStore struct {
	thread Thread
	messages []Message
	tasks []Task
	tools []ToolCall
}

func (m *memoryStore) CreateThread(_ context.Context, owner, title string) (Thread, error) {
	m.thread = Thread{ID: "thread-1", Owner: owner, Title: cleanTitle(title), Status: "active"}
	return m.thread, nil
}
func (m *memoryStore) ListThreads(_ context.Context, owner string, _ int) ([]Thread, error) {
	if m.thread.ID == "" || m.thread.Owner != owner { return []Thread{}, nil }
	return []Thread{m.thread}, nil
}
func (m *memoryStore) GetThread(_ context.Context, owner, id string) (ThreadSnapshot, error) {
	if m.thread.ID != id || m.thread.Owner != owner { return ThreadSnapshot{}, ErrNotFound }
	return ThreadSnapshot{Thread: m.thread, Messages: append([]Message(nil), m.messages...), Tasks: append([]Task(nil), m.tasks...), ToolCalls: append([]ToolCall(nil), m.tools...)}, nil
}
func (m *memoryStore) AppendMessage(_ context.Context, input AppendMessageInput) (Message, error) {
	item := Message{ID: "message-" + input.Role, ThreadID: input.ThreadID, Owner: input.Owner, Role: input.Role, Content: input.Content, MediaAssetIDs: append([]string(nil), input.MediaAssetIDs...)}
	m.messages = append(m.messages, item)
	return item, nil
}
func (m *memoryStore) CreateTask(_ context.Context, input CreateTaskInput) (Task, error) {
	item := Task{ID: "task-1", ThreadID: input.ThreadID, Owner: input.Owner, Title: input.Title, Status: input.Status, Detail: input.Detail}
	m.tasks = append(m.tasks, item)
	return item, nil
}
func (m *memoryStore) ListTasks(_ context.Context, owner, threadID string) ([]Task, error) {
	out := []Task{}
	for _, item := range m.tasks { if item.Owner == owner && (threadID == "" || item.ThreadID == threadID) { out = append(out, item) } }
	return out, nil
}
func (m *memoryStore) UpdateTask(_ context.Context, owner, taskID string, input UpdateTaskInput) (Task, error) {
	for index, item := range m.tasks {
		if item.ID != taskID || item.Owner != owner { continue }
		item.Status = input.Status
		item.ProgressCurrent = input.ProgressCurrent
		item.ProgressTotal = input.ProgressTotal
		item.Detail = input.Detail
		m.tasks[index] = item
		return item, nil
	}
	return Task{}, ErrNotFound
}
func (m *memoryStore) CreateToolCall(_ context.Context, input CreateToolCallInput) (ToolCall, error) {
	item := ToolCall{ID: "tool-1", ThreadID: input.ThreadID, MessageID: input.MessageID, Owner: input.Owner, ToolName: input.ToolName, Status: input.Status, Arguments: input.Arguments, Result: input.Result}
	m.tools = append(m.tools, item)
	return item, nil
}
func (m *memoryStore) UpdateToolCall(_ context.Context, owner, toolID string, input UpdateToolCallInput) (ToolCall, error) {
	for index, item := range m.tools {
		if item.ID != toolID || item.Owner != owner { continue }
		item.Status = input.Status
		item.Result = input.Result
		m.tools[index] = item
		return item, nil
	}
	return ToolCall{}, ErrNotFound
}
func (m *memoryStore) UpdateThreadTitle(_ context.Context, owner, id, title string) error {
	if m.thread.Owner != owner || m.thread.ID != id { return ErrNotFound }
	m.thread.Title = cleanTitle(title)
	return nil
}

type fixedResponder struct { response AgentResponse }
func (r fixedResponder) Respond(context.Context, string, string, string) (AgentResponse, error) { return r.response, nil }

func TestServiceSendMessagePersistsUserAssistantAndTool(t *testing.T) {
	store := &memoryStore{thread: Thread{ID: "thread-1", Owner: "owner", Title: "新对话", Status: "active"}}
	args, _ := json.Marshal(NavigateArgs{Path: "/novel-fetch"})
	service := &Service{Store: store, Responder: fixedResponder{response: AgentResponse{Content: "我来打开小说获取。", Tool: &ToolProposal{Name: ToolNavigate, Label: "打开小说获取", Arguments: args}}}}
	result, err := service.SendMessage(context.Background(), "owner", "thread-1", SendMessageInput{Content: "帮我打开小说获取", MediaAssetIDs: []string{"asset_1"}})
	if err != nil { t.Fatal(err) }
	if result.UserMessage.Role != RoleUser || result.AssistantMessage.Role != RoleAssistant { t.Fatalf("unexpected messages: %#v", result) }
	if len(store.messages) != 2 { t.Fatalf("messages=%d", len(store.messages)) }
	if len(store.tools) != 1 || store.tools[0].ToolName != ToolNavigate || store.tools[0].Status != ToolProposed { t.Fatalf("tools=%#v", store.tools) }
	if result.ToolCall == nil || result.ToolCall.ID == "" { t.Fatal("expected tool call in result") }
	if store.thread.Title == "新对话" { t.Fatal("first user message should title the thread") }
}

func TestServiceSendMessagePersistsTaskProposal(t *testing.T) {
	store := &memoryStore{thread: Thread{ID: "thread-1", Owner: "owner", Title: "测试", Status: "active"}}
	service := &Service{Store: store, Responder: fixedResponder{response: AgentResponse{Content: "已记录", Task: &TaskProposal{Title: "检查视频", Status: TaskInProgress}}}}
	result, err := service.SendMessage(context.Background(), "owner", "thread-1", SendMessageInput{Content: "创建任务：检查视频"})
	if err != nil { t.Fatal(err) }
	if result.Task == nil || result.Task.Title != "检查视频" { t.Fatalf("task=%#v", result.Task) }
	if len(store.tasks) != 1 { t.Fatalf("tasks=%d", len(store.tasks)) }
}

func TestServiceUpdatesToolLifecycle(t *testing.T) {
	store := &memoryStore{tools: []ToolCall{{ID: "tool-1", Owner: "owner", Status: ToolProposed}}}
	service := &Service{Store: store, Responder: fixedResponder{}}
	resultBody := json.RawMessage(`{"path":"/novel-fetch"}`)
	updated, err := service.UpdateToolCall(context.Background(), "owner", "tool-1", UpdateToolCallInput{Status: ToolCompleted, Result: resultBody})
	if err != nil { t.Fatal(err) }
	if updated.Status != ToolCompleted || string(updated.Result) != string(resultBody) { t.Fatalf("updated=%#v", updated) }
}

func TestServiceUpdatesTaskLifecycle(t *testing.T) {
	store := &memoryStore{tasks: []Task{{ID: "task-1", Owner: "owner", Status: TaskInProgress}}}
	service := &Service{Store: store, Responder: fixedResponder{}}
	updated, err := service.UpdateTask(context.Background(), "owner", "task-1", UpdateTaskInput{Status: TaskCompleted, ProgressCurrent: 3, ProgressTotal: 3, Detail: "完成"})
	if err != nil { t.Fatal(err) }
	if updated.Status != TaskCompleted || updated.ProgressCurrent != 3 { t.Fatalf("updated=%#v", updated) }
}

func TestServiceRejectsCrossOwnerThread(t *testing.T) {
	store := &memoryStore{thread: Thread{ID: "thread-1", Owner: "alice", Title: "A", Status: "active"}}
	service := &Service{Store: store, Responder: fixedResponder{response: AgentResponse{Content: "x"}}}
	if _, err := service.SendMessage(context.Background(), "bob", "thread-1", SendMessageInput{Content: "hello"}); err != ErrNotFound {
		t.Fatalf("err=%v", err)
	}
}
