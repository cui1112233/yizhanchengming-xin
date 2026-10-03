package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
)

type fakeAgentAPI struct {
	threads []agent.Thread
	snapshot agent.ThreadSnapshot
	send agent.SendMessageResult
	createOwner string
	lastContent string
	updatedTask agent.Task
	updatedTool agent.ToolCall
}

func (f *fakeAgentAPI) ListThreads(context.Context, string, int) ([]agent.Thread, error) { return f.threads, nil }
func (f *fakeAgentAPI) CreateThread(_ context.Context, owner, title string) (agent.Thread, error) {
	f.createOwner = owner
	return agent.Thread{ID: "thread-1", Title: title, Status: "active"}, nil
}
func (f *fakeAgentAPI) GetThread(_ context.Context, owner, id string) (agent.ThreadSnapshot, error) {
	if id == "missing" || owner == "other" { return agent.ThreadSnapshot{}, agent.ErrNotFound }
	return f.snapshot, nil
}
func (f *fakeAgentAPI) SendMessage(_ context.Context, owner, threadID string, input agent.SendMessageInput) (agent.SendMessageResult, error) {
	if owner == "other" || threadID == "missing" { return agent.SendMessageResult{}, agent.ErrNotFound }
	f.lastContent = input.Content
	return f.send, nil
}
func (f *fakeAgentAPI) ListTasks(context.Context, string, string) ([]agent.Task, error) { return []agent.Task{}, nil }
func (f *fakeAgentAPI) UpdateTask(_ context.Context, owner, taskID string, input agent.UpdateTaskInput) (agent.Task, error) {
	if owner == "other" || taskID == "missing" { return agent.Task{}, agent.ErrNotFound }
	f.updatedTask = agent.Task{ID: taskID, Status: input.Status, ProgressCurrent: input.ProgressCurrent, ProgressTotal: input.ProgressTotal, Detail: input.Detail}
	return f.updatedTask, nil
}
func (f *fakeAgentAPI) UpdateToolCall(_ context.Context, owner, toolID string, input agent.UpdateToolCallInput) (agent.ToolCall, error) {
	if owner == "other" || toolID == "missing" { return agent.ToolCall{}, agent.ErrNotFound }
	f.updatedTool = agent.ToolCall{ID: toolID, Status: input.Status, Result: input.Result}
	return f.updatedTool, nil
}

func testOwner(_ *http.Request) (string, error) { return "owner", nil }

func TestAgentHandlerCreatesThreadWithServerOwner(t *testing.T) {
	api := &fakeAgentAPI{}
	handler := NewAgentHandler(api, testOwner)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/threads", strings.NewReader(`{"title":"新项目"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if api.createOwner != "owner" { t.Fatalf("owner=%q", api.createOwner) }
	var body map[string]agent.Thread
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if body["thread"].ID != "thread-1" { t.Fatalf("body=%s", rec.Body.String()) }
}

func TestAgentHandlerPostsMessageAndReturnsToolCall(t *testing.T) {
	args, _ := json.Marshal(agent.NavigateArgs{Path: "/novel-fetch"})
	api := &fakeAgentAPI{send: agent.SendMessageResult{
		UserMessage: agent.Message{ID: "u", Role: agent.RoleUser, Content: "打开小说获取"},
		AssistantMessage: agent.Message{ID: "a", Role: agent.RoleAssistant, Content: "可以"},
		ToolCall: &agent.ToolCall{ID: "tool", ToolName: agent.ToolNavigate, Status: agent.ToolProposed, Arguments: args},
	}}
	handler := NewAgentHandler(api, testOwner)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/threads/thread-1/messages", strings.NewReader(`{"content":"打开小说获取"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if api.lastContent != "打开小说获取" { t.Fatalf("content=%q", api.lastContent) }
	var body agent.SendMessageResult
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil { t.Fatal(err) }
	if body.ToolCall == nil || body.ToolCall.ToolName != agent.ToolNavigate { t.Fatalf("body=%s", rec.Body.String()) }
}

func TestAgentHandlerUpdatesToolLifecycle(t *testing.T) {
	api := &fakeAgentAPI{}
	handler := NewAgentHandler(api, testOwner)
	req := httptest.NewRequest(http.MethodPatch, "/api/agent/tool-calls/tool-1", strings.NewReader(`{"status":"completed","result":{"path":"/novel-fetch"}}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if api.updatedTool.Status != agent.ToolCompleted { t.Fatalf("tool=%#v", api.updatedTool) }
}

func TestAgentHandlerUpdatesTaskLifecycle(t *testing.T) {
	api := &fakeAgentAPI{}
	handler := NewAgentHandler(api, testOwner)
	req := httptest.NewRequest(http.MethodPatch, "/api/agent/tasks/task-1", strings.NewReader(`{"status":"completed","progress_current":3,"progress_total":3,"detail":"完成"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if api.updatedTask.Status != agent.TaskCompleted || api.updatedTask.ProgressCurrent != 3 { t.Fatalf("task=%#v", api.updatedTask) }
}

func TestAgentHandlerRejectsUnauthorizedRequest(t *testing.T) {
	api := &fakeAgentAPI{}
	handler := NewAgentHandler(api, func(*http.Request) (string, error) { return "", errors.New("no") })
	req := httptest.NewRequest(http.MethodGet, "/api/agent/threads", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized { t.Fatalf("status=%d", rec.Code) }
}

func TestAgentHandlerRejectsMalformedMessageJSON(t *testing.T) {
	api := &fakeAgentAPI{}
	handler := NewAgentHandler(api, testOwner)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/threads/thread-1/messages", strings.NewReader(`{"content":`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d", rec.Code) }
}

func TestAgentHandlerReturns404ForUnknownThread(t *testing.T) {
	api := &fakeAgentAPI{}
	handler := NewAgentHandler(api, testOwner)
	req := httptest.NewRequest(http.MethodGet, "/api/agent/threads/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
