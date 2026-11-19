package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayExistingPlanningHTTPUsesOwnModelAndNoReplayCall(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	_, e := store.CreateProject(ctx, screenplay.Project{ID: "existing-p", OwnerUsername: "alice", ModelID: "OWN_MODEL", CreativeConfig: &screenplay.CreativeConfig{StoryboardPromptID: "storyboard_xiaoluo_v7_15s"}})
	if e != nil {
		t.Fatal(e)
	}
	authors := &screenplay.AuthoringService{Store: store, Repository: screenplay.NewMemoryAuthoringRepository()}
	st, e := authors.Save(ctx, "alice", "existing-p", screenplay.AuthoringState{Chapters: []screenplay.Chapter{{ID: "c", Title: "目标", ScriptText: "OWN_SCRIPT", PlanningText: "OLD_PLAN"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	provider := screenplayProviderFunc(func(_ context.Context, r batchfactoryv11.TextCompletionRequest) (string, error) {
		calls++
		if len(r.Messages) != 4 || r.Messages[2].Role != "assistant" || r.Messages[2].Content != "OLD_PLAN" || !strings.Contains(r.Messages[3].Content, "无缝续写") {
			t.Fatal("history lost before configured provider", r)
		}
		return "OWN_APPEND", nil
	})
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayStoryboardPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, authors, func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || model != "OWN_MODEL" {
			t.Fatal("foreign/default model used", owner, model)
		}
		return provider, nil
	})
	body := map[string]any{"is_continue": true, "chapter_id": "c", "expected_revision": st.Revision, "idempotencyKey": "one", "auto_continue": true, "max_rounds": 8}
	path := "/api/comic-screenplays/projects/existing-p/storyboard-plan"
	first := signedJSONRequest(t, mux, now, "alice", "POST", path, body)
	if first.Code != 200 || calls != 1 || !strings.Contains(first.Body.String(), `"planning_text":"OLD_PLAN\nOWN_APPEND"`) || !strings.Contains(first.Body.String(), `"planning_rounds":1`) {
		t.Fatal(first.Code, first.Body.String(), calls)
	}
	repeat := signedJSONRequest(t, mux, now, "alice", "POST", path, body)
	if repeat.Code != 200 || calls != 1 {
		t.Fatal("replay generated", repeat.Code, calls)
	}
	body["existing_plan"] = "CHANGED"
	changed := signedJSONRequest(t, mux, now, "alice", "POST", path, body)
	if changed.Code != 400 || calls != 1 {
		t.Fatal("changed seed generated", changed.Code, calls)
	}
	body["existing_plan"] = ""
	forbidden := signedJSONRequest(t, mux, now, "bob", "POST", path, body)
	if forbidden.Code != 403 || calls != 1 {
		t.Fatal("other user generated", forbidden.Code, calls)
	}
}
