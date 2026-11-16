package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelcreation"
	"testing"
	"time"
)

type novelConfirmationProvider struct{ calls int }

func (p *novelConfirmationProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	if p.calls == 1 {
		return `{"title":"故事","genre":"悬疑","world":"旧城","characters":[{"name":"林","role":"主角","goal":"追查","conflict":"失踪"}],"coreConflict":"追查真相","volumes":[{"number":1,"title":"第一卷","goal":"追查","chapterRange":[1,10]}]}`, nil
	}
	return `{"chapterNumber":1,"title":"信","goal":"追查","conflict":"失踪","turn":"找到信","hook":"门外脚步","targetWords":2000}`, nil
}

func TestNovelConfirmationHTTPBindsExactTokenAndOwner(t *testing.T) {
	ctx := context.Background()
	store := novelcreation.NewMemoryStore()
	p := &novelConfirmationProvider{}
	service := novelcreation.NewService(store, novelcreation.NewMemoryRuntime(), novelcreation.NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return p, nil }))
	project, err := service.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "故事", Genre: "悬疑", ModelID: "chosen"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := service.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: project.ID, Mode: "short_original"})
	if err != nil {
		t.Fatal(err)
	}
	var a novelcreation.AgentSettings
	_ = json.Unmarshal([]byte(`{"requireSecondaryConfirm":true,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	_, _ = service.UpdateProjectSettings(ctx, "alice", project.ID, novelcreation.ProjectSettings{Agent: a})
	run, err := service.SendMessage(ctx, "alice", novelcreation.SendMessageInput{ProjectID: project.ID, SessionID: session.ID, ModelID: "chosen", Content: "写一个故事", IdempotencyKey: "confirmation-http"})
	if err != nil {
		t.Fatal(err)
	}
	paused, err := service.Run(ctx, "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	var state struct {
		ID string `json:"confirmationId"`
	}
	_ = json.Unmarshal(paused.RuntimeSnapshot, &state)
	if state.ID == "" || p.calls != 1 {
		t.Fatalf("no durable confirmation: %s", paused.RuntimeSnapshot)
	}
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: service})
	path := "/api/novel-creation/runs/" + run.ID + "/confirm"
	noToken := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{})
	if noToken.Code < 400 {
		t.Fatal("HTTP accepted missing expected confirmation token")
	}
	other := signedJSONRequest(t, mux, now, "bob", http.MethodPost, path, map[string]any{"confirmationId": state.ID})
	if other.Code < 400 {
		t.Fatal("other account confirmed this model call")
	}
	accepted := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{"confirmationId": state.ID})
	if accepted.Code != http.StatusAccepted {
		t.Fatalf("confirm=%d %s", accepted.Code, accepted.Body.String())
	}
	paused, err = service.Run(ctx, "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	stale := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, map[string]any{"confirmationId": state.ID})
	if stale.Code != http.StatusAccepted {
		t.Fatalf("old confirmed token was not idempotent: %s", stale.Body.String())
	}
	current, _ := store.GetRun(ctx, "alice", run.ID)
	if current.Status != novelcreation.RunAwaitingConfirmation || p.calls != 2 {
		t.Fatalf("late HTTP retry authorized next request: status=%s calls=%d", current.Status, p.calls)
	}
}
