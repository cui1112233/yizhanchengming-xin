package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelcreation"
	"sync/atomic"
	"testing"
	"time"
)

type studioHTTPProvider struct{ calls atomic.Int32 }

func (p *studioHTTPProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls.Add(1)
	return `{"message":"请确认创建这部小说。","action":{"type":"create_book","book":{"title":"雨夜归人","genre":"悬疑","world":"旧城","protagonist":"送信人","coreConflict":"寻找旧信真相","targetChapterWords":2000}}}`, nil
}
func TestNovelStudioRoutesPlanConfirmCancelAndOwnership(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	rt := novelcreation.NewMemoryRuntime()
	provider := &studioHTTPProvider{}
	pipeline := novelcreation.NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil })
	svc := novelcreation.NewService(st, rt, pipeline)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	create := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/studio-sessions", map[string]any{"idempotencyKey": "create-studio-1", "mode": "long_original"})
	if create.Code != 201 {
		t.Fatalf("create=%d %s", create.Code, create.Body.String())
	}
	var cr struct {
		Session novelcreation.StudioSession `json:"session"`
	}
	if e := json.Unmarshal(create.Body.Bytes(), &cr); e != nil {
		t.Fatal(e)
	}
	path := "/api/novel-creation/studio-sessions/" + cr.Session.ID
	replay := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/studio-sessions", map[string]any{"idempotencyKey": "create-studio-1", "mode": "long_original"})
	if replay.Code != 201 || !containsText(replay.Body.String(), cr.Session.ID) {
		t.Fatal("creation replay changed session")
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, path, nil)
	if foreign.Code != 404 {
		t.Fatalf("cross-owner=%d", foreign.Code)
	}
	bad := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/messages", map[string]any{"idempotencyKey": "plan1", "content": "内容", "modelId": "own-model", "owner": "bob"})
	if bad.Code != 400 || provider.calls.Load() != 0 {
		t.Fatal("forged field accepted/called model")
	}
	input := map[string]any{"idempotencyKey": "plan1", "content": "创作标题雨夜归人的悬疑长篇，主角是寻找旧信的送信人。", "modelId": "own-model", "thinkingLevel": "standard", "contextBudget": 12000}
	plan := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/messages", input)
	if plan.Code != 200 {
		t.Fatalf("plan=%d %s", plan.Code, plan.Body.String())
	}
	var snap novelcreation.StudioSnapshot
	if e := json.Unmarshal(plan.Body.Bytes(), &snap); e != nil {
		t.Fatal(e)
	}
	if snap.PendingAction == nil || snap.PendingAction.Book.Title != "雨夜归人" {
		t.Fatal("pending action absent")
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	if len(projects) != 0 {
		t.Fatal("planning made hidden project")
	}
	again := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/messages", input)
	if again.Code != 200 || provider.calls.Load() != 1 {
		t.Fatal("planning replay billed again")
	}
	confirm := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/actions/confirm", map[string]any{"actionId": snap.PendingAction.ID, "modelId": "own-model-2", "thinkingLevel": "deep"})
	if confirm.Code != 200 {
		t.Fatalf("confirm=%d %s", confirm.Code, confirm.Body.String())
	}
	var tr novelcreation.StudioTransition
	_ = json.Unmarshal(confirm.Body.Bytes(), &tr)
	if tr.Project.ID == "" || tr.ProductionSession.ID == "" || tr.Run.ID == "" || tr.Run.ThinkingLevel != "deep" || tr.Project.ModelID != "own-model-2" {
		t.Fatal("incomplete transition")
	}
	again = signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/actions/confirm", map[string]any{"actionId": snap.PendingAction.ID, "modelId": "own-model"})
	if again.Code != 200 || !containsText(again.Body.String(), tr.Run.ID) {
		t.Fatal("repeat confirm changed result")
	}
	projects, _ = st.ListProjects(context.Background(), "alice")
	if len(projects) != 1 || provider.calls.Load() != 1 {
		t.Fatal("confirm duplicated project or planning")
	}
	cancel := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/actions/cancel", map[string]any{"actionId": snap.PendingAction.ID})
	if cancel.Code != 409 {
		t.Fatalf("confirmed action cancel=%d", cancel.Code)
	}
}
