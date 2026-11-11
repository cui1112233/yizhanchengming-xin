package agentchat

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type scriptedProvider struct {
	replies []string
	calls   int
}

func (p *scriptedProvider) Complete(ctx context.Context, system, user string, maxTokens int) (string, error) {
	if p.calls >= len(p.replies) {
		return "完成", nil
	}
	reply := p.replies[p.calls]
	p.calls++
	return reply, nil
}

type errorProvider struct{ err error }

func (p *errorProvider) Complete(context.Context, string, string, int) (string, error) {
	return "", p.err
}

func newTestStore(t *testing.T) (*FileStore, string) {
	t.Helper()
	base := t.TempDir()
	store := NewFileStore(base, func() time.Time { return time.UnixMilli(1794200000000) })
	if _, err := store.ListWorkspaces(context.Background(), "alice"); err != nil {
		t.Fatal(err)
	}
	return store, base
}

func TestFileStoreCreatesDefaultWorkspaceWithLuoshuiLayout(t *testing.T) {
	store, _ := newTestStore(t)
	workspaces, err := store.ListWorkspaces(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(workspaces) != 1 || !workspaces[0].IsDefault {
		t.Fatalf("expected one default workspace, got %+v", workspaces)
	}
	for _, rel := range []string{
		".luoshui/workspace.json",
		".luoshui/skills",
		".luoshui/context",
		".luoshui/sessions",
		".luoshui/runs",
		"input",
		"output",
	} {
		if _, err := os.Stat(filepath.Join(workspaces[0].RootPath, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}
}

func TestNormalizePermissionDefaultsToPlan(t *testing.T) {
	cases := map[string]string{
		"":         PermissionPlan,
		"   ":      PermissionPlan,
		"unknown":  PermissionPlan,
		"PLAN":     PermissionPlan,
		"auto":     PermissionAuto,
		"bypass":   PermissionBypass,
		" BYPASS ": PermissionBypass,
	}
	for input, want := range cases {
		if got := NormalizePermission(input); got != want {
			t.Fatalf("NormalizePermission(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSessionLifecyclePersistsMessages(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	workspaces, _ := store.ListWorkspaces(ctx, "alice")
	session, err := store.CreateSession(ctx, "alice", SessionInfo{Title: "新 Agent 会话", WorkspaceID: workspaces[0].ID, Model: "test-model"})
	if err != nil {
		t.Fatal(err)
	}
	if session.ID == 0 || session.PermissionMode != PermissionPlan {
		t.Fatalf("unexpected session %+v", session)
	}
	message := EmptyMessage("user")
	message.ID = 1
	message.Content = "你好"
	if err := store.SaveMessages(ctx, "alice", session.ID, []Message{message}); err != nil {
		t.Fatal(err)
	}
	messages, err := store.Messages(ctx, "alice", session.ID)
	if err != nil || len(messages) != 1 || messages[0].Content != "你好" {
		t.Fatalf("messages=%+v err=%v", messages, err)
	}
	updated, err := store.UpdateSession(ctx, "alice", session.ID, SessionPatch{Title: pstr("改名会话")})
	if err != nil || updated.Title != "改名会话" || updated.MessageCount != 1 {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if _, err := store.Session(ctx, "alice", session.ID); err != nil {
		t.Fatalf("renamed session must stay addressable by id: %v", err)
	}
}

func pstr(value string) *string { return &value }

func TestRunnerPlanModeOnlyPlansTools(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	workspaces, _ := store.ListWorkspaces(ctx, "alice")
	session, err := store.CreateSession(ctx, "alice", SessionInfo{Title: "计划", WorkspaceID: workspaces[0].ID, Model: "test-model", PermissionMode: PermissionPlan})
	if err != nil {
		t.Fatal(err)
	}
	provider := &scriptedProvider{replies: []string{
		`{"tool":"list_workspace_files","args":{}}`,
		"计划：先列出工作区文件，再读取需要修改的文档。",
	}}
	runner := NewRunner(store, func(ctx context.Context, owner, modelID string) (Provider, error) {
		return provider, nil
	}, func() time.Time { return time.UnixMilli(1794200000000) })
	runner.SetSkills(NewSkillCatalog([]Skill{{ID: "demo", Description: "示例技能", Body: "正文"}}))

	var events []string
	assistant, err := runner.Run(ctx, "alice", SendRequest{
		SessionID:   session.ID,
		WorkspaceID: workspaces[0].ID,
		Model:       "test-model",
		Permission:  PermissionPlan,
		Content:     "整理工作区",
	}, func(event Event) { events = append(events, event.Type) })
	if err != nil {
		t.Fatal(err)
	}
	if len(assistant.ToolBlocks) != 1 {
		t.Fatalf("expected one tool block, got %+v", assistant.ToolBlocks)
	}
	if assistant.ToolBlocks[0].Status != ToolPlanned {
		t.Fatalf("plan mode must not execute tools: %+v", assistant.ToolBlocks[0])
	}
	if assistant.Content == "" {
		t.Fatal("assistant reply must be persisted")
	}
	messages, err := store.Messages(ctx, "alice", session.ID)
	if err != nil || len(messages) != 2 {
		t.Fatalf("expected user + assistant messages, got %d (%v)", len(messages), err)
	}
	if len(events) == 0 {
		t.Fatal("expected streamed events")
	}
}

func TestRunnerBypassExecutesReadOnlyTool(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	workspaces, _ := store.ListWorkspaces(ctx, "alice")
	if err := os.WriteFile(filepath.Join(workspaces[0].RootPath, "note.md"), []byte("hello agent"), 0o600); err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession(ctx, "alice", SessionInfo{Title: "执行", WorkspaceID: workspaces[0].ID, Model: "test-model", PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	provider := &scriptedProvider{replies: []string{
		`{"tool":"read_workspace_file","args":{"path":"note.md"}}`,
		"文件内容是 hello agent。",
	}}
	runner := NewRunner(store, func(ctx context.Context, owner, modelID string) (Provider, error) {
		return provider, nil
	}, func() time.Time { return time.UnixMilli(1794200000000) })
	assistant, err := runner.Run(ctx, "alice", SendRequest{
		SessionID:   session.ID,
		WorkspaceID: workspaces[0].ID,
		Model:       "test-model",
		Permission:  PermissionBypass,
		Content:     "读 note.md",
	}, func(Event) {})
	if err != nil {
		t.Fatal(err)
	}
	if len(assistant.ToolBlocks) != 1 || assistant.ToolBlocks[0].Status != ToolDone {
		t.Fatalf("expected executed tool block, got %+v", assistant.ToolBlocks)
	}
	if assistant.ToolBlocks[0].ContentPreview == "" {
		t.Fatal("tool result preview must be kept")
	}
}

func TestRunnerProviderCreationFailureReturnsError(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	workspaces, _ := store.ListWorkspaces(ctx, "alice")
	session, err := store.CreateSession(ctx, "alice", SessionInfo{Title: "失败", WorkspaceID: workspaces[0].ID, Model: "test-model", PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(store, func(context.Context, string, string) (Provider, error) {
		return nil, errors.New("requested text model is not configured")
	}, func() time.Time { return time.UnixMilli(1794200000000) })
	var detail string
	message, runErr := runner.Run(ctx, "alice", SendRequest{
		SessionID: session.ID, WorkspaceID: workspaces[0].ID, Model: "test-model", Permission: PermissionBypass, Content: "你好",
	}, func(event Event) {
		if event.Type != EventError {
			return
		}
		if payload, ok := event.Data.(map[string]any); ok {
			detail, _ = payload["detail"].(string)
		}
	})
	if runErr == nil {
		t.Fatal("模型未配置不能当成功返回")
	}
	if message.Content == "请求失败" || !strings.Contains(message.Content, "未配置") {
		t.Fatalf("失败消息必须是中文结论，got %q", message.Content)
	}
	if !strings.Contains(detail, "not configured") {
		t.Fatalf("原始报错必须留在 detail 字段，got %q", detail)
	}
}

func TestRunnerProviderCallFailureKeepsHeadlineAndRaw(t *testing.T) {
	store, _ := newTestStore(t)
	ctx := context.Background()
	workspaces, _ := store.ListWorkspaces(ctx, "alice")
	session, err := store.CreateSession(ctx, "alice", SessionInfo{Title: "鉴权", WorkspaceID: workspaces[0].ID, Model: "test-model", PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	provider := &errorProvider{err: errors.New("文本模型鉴权失败，请检查模型密钥后重试。\n上游原文：HTTP 401 · model=test-model")}
	runner := NewRunner(store, func(context.Context, string, string) (Provider, error) {
		return provider, nil
	}, func() time.Time { return time.UnixMilli(1794200000000) })
	var headline, detail string
	assistant, runErr := runner.Run(ctx, "alice", SendRequest{
		SessionID: session.ID, WorkspaceID: workspaces[0].ID, Model: "test-model", Permission: PermissionBypass, Content: "你好",
	}, func(event Event) {
		if event.Type != EventError {
			return
		}
		if payload, ok := event.Data.(map[string]any); ok {
			headline, _ = payload["message"].(string)
			detail, _ = payload["detail"].(string)
		}
	})
	if runErr == nil {
		t.Fatal("模型调用失败必须返回错误")
	}
	if headline != "文本模型鉴权失败，请检查模型密钥后重试。" {
		t.Fatalf("主文案只保留中文结论，got %q", headline)
	}
	if !strings.Contains(detail, "HTTP 401") || strings.Contains(headline, "HTTP 401") {
		t.Fatalf("状态码应只出现在 detail，headline=%q detail=%q", headline, detail)
	}
	if assistant.Content != headline {
		t.Fatalf("失败消息不应伪装成正常回复: %q", assistant.Content)
	}
}
