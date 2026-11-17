package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplaySceneGenerationUsesSingleAttemptAtResolverAndProvider(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	_, err := store.CreateProject(ctx, screenplay.Project{ID: "scene-one-attempt", OwnerUsername: "alice", ModelID: "own-text", PermissionMode: screenplay.PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	provider := screenplayProviderFunc(func(ctx context.Context, _ batchfactoryv11.TextCompletionRequest) (string, error) {
		if !batchfactoryv11.SingleAttempt(ctx) {
			t.Error("scene provider may retry beyond approved call count")
		}
		return `{"storyboard":[{"id":1,"scene_id":1,"duration_sec":12,"characters":"甲","scene":"房间","props":"无","video_desc":"甲打开窗。"}]}`, nil
	})
	resolver := func(ctx context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if !batchfactoryv11.SingleAttempt(ctx) {
			t.Error("scene resolver may retry its integration call")
		}
		return provider, nil
	}
	now := time.Now()
	mux := http.NewServeMux()
	RegisterScreenplayWorkflowRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now }, nil, nil)
	res := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/scene-one-attempt/storyboards/generate", map[string]any{"idempotencyKey": "scene-one-attempt", "engine": "xiaoluo_v7_15s", "content": "第1集\n甲打开窗。"})
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
}
