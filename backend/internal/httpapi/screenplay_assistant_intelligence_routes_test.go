package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

// 助手主动追问与导航建议应落库，但不改作品内容、不产生 proposal（无 actions）。
func TestAuthoringAssistantQuestionAndSuggestions(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", SourceText: "原文", Synopsis: "概要"})
	repo := screenplay.NewMemoryAuthoringRepository()
	var gotSystem string
	mux := http.NewServeMux()
	RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, repo,
		func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return screenplayProviderFunc(func(_ context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
				gotSystem = r.SystemPrompt
				return `{"reply":"信息不足，先确认集数","summary":"","actions":[],"question":"你想先处理第几集？","suggestions":[{"label":"去资产页解析角色","entry":"assets","reason":"先建立角色一致性"},{"label":"跳到不存在的面板","entry":"hack","reason":"应被过滤"}]}`, nil
			}), nil
		})
	base := "/api/comic-screenplays/projects/p/assistant/chat"
	res := signedJSONRequest(t, mux, now, "alice", "POST", base, map[string]any{"expected_revision": 0, "message": "帮我做分镜"})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	if !strings.Contains(gotSystem, "当前作品链路状态") || !strings.Contains(gotSystem, "只读事实") {
		t.Fatalf("系统词未注入能力清单: %s", gotSystem)
	}
	st, _ := repo.Load(ctx, "alice", "p")
	if len(st.Messages) != 3 {
		t.Fatalf("消息数=%d want 3(user+reply+question)", len(st.Messages))
	}
	q := st.Messages[2]
	if q.Kind != screenplay.AssistantMessageKindQuestion || !strings.Contains(q.Content, "第几集") {
		t.Fatalf("追问未落库: %#v", q)
	}
	// 导航建议附加到普通助手回复（第 1 条 assistant），且 hack 入口被过滤
	if len(st.Messages[1].Suggestions) != 1 || st.Messages[1].Suggestions[0].Entry != "assets" {
		t.Fatalf("导航建议过滤失败: %#v", st.Messages[1].Suggestions)
	}
	if len(st.Proposals) != 0 {
		t.Fatalf("无 actions 不应产生 proposal: %#v", st.Proposals)
	}
}

// 会话超长时只压缩注入上下文与持久化消息，不改作品数据、不产生 proposal。
func TestAuthoringAssistantCompactsLongConversation(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", SourceText: "原文"})
	repo := screenplay.NewMemoryAuthoringRepository()
	svc := screenplay.AuthoringService{Store: s, Repository: repo}
	st, _ := svc.Load(ctx, "alice", "p")
	long := make([]screenplay.AssistantMessage, 0, 45)
	for i := 0; i < 45; i++ {
		long = append(long, screenplay.AssistantMessage{Role: "user", Content: "历史问题填充"})
		long = append(long, screenplay.AssistantMessage{Role: "assistant", Content: "历史回答填充"})
	}
	st.Messages = long
	svc.Save(ctx, "alice", "p", st, 0)
	var gotUser string
	mux := http.NewServeMux()
	RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, repo,
		func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return screenplayProviderFunc(func(_ context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
				gotUser = r.UserPrompt
				return `{"reply":"已了解","summary":"","actions":[]}`, nil
			}), nil
		})
	res := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/p/assistant/chat", map[string]any{"expected_revision": 1, "message": "继续"})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	if !strings.Contains(gotUser, "历史对话已折叠") {
		t.Fatalf("注入上下文未压缩: %.300s", gotUser)
	}
	saved, _ := repo.Load(ctx, "alice", "p")
	hasCompacted := false
	for _, m := range saved.Messages {
		if m.Kind == screenplay.AssistantMessageKindCompacted {
			hasCompacted = true
		}
	}
	if !hasCompacted {
		t.Fatal("持久化消息未压缩")
	}
}

// 回归：原有编辑确认流程不受增强影响。
func TestAuthoringAssistantEditFlowStillWorksWithIntelligence(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice", SourceText: "原文"})
	repo := screenplay.NewMemoryAuthoringRepository()
	svc := screenplay.AuthoringService{Store: s, Repository: repo}
	st, _ := svc.Load(ctx, "alice", "p")
	st.Chapters = []screenplay.Chapter{{ID: "c", Title: "旧", SourceText: "原文"}}
	svc.Save(ctx, "alice", "p", st, 0)
	mux := http.NewServeMux()
	RegisterScreenplayAuthoringRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, repo,
		func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return screenplayProviderFunc(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
				return `{"reply":"建议改名","summary":"改名","actions":[{"kind":"chapter","chapter":{"id":"c","title":"新","source_text":"原文","script_text":""}}]}`, nil
			}), nil
		})
	base := "/api/comic-screenplays/projects/p/assistant"
	res := signedJSONRequest(t, mux, now, "alice", "POST", base+"/chat", map[string]any{"expected_revision": 1, "message": "改名"})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	st, _ = repo.Load(ctx, "alice", "p")
	if len(st.Proposals) != 1 || st.Chapters[0].Title != "旧" {
		t.Fatalf("提案流程被破坏: %#v", st)
	}
	var body map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &body)
	res = signedJSONRequest(t, mux, now, "alice", "POST", base+"/apply", map[string]any{"expected_revision": 2, "proposal_id": st.Proposals[0].ID, "confirmed": true})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	st, _ = repo.Load(ctx, "alice", "p")
	if st.Chapters[0].Title != "新" {
		t.Fatal("确认后未应用")
	}
}
