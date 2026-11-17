package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplayRejectedModelFailsWithoutAutomaticRetry(t *testing.T) {
	cases := []struct {
		name     string
		upstream error
		want     string
	}{
		{"model missing503", &batchfactoryv11.DirectorHTTPError{StatusCode: 503, Body: `{"error":{"code":"model_not_found"}}`}, "failed"},
		{"wrapped401", fmt.Errorf("provider: %w", &batchfactoryv11.DirectorHTTPError{StatusCode: 401}), "failed"},
		{"gateway503", &batchfactoryv11.DirectorHTTPError{StatusCode: 503, Body: `{"error":{"message":"gateway unavailable"}}`}, "unknown"},
		{"unstructured body", &batchfactoryv11.DirectorHTTPError{StatusCode: 503, Body: `model_not_found`}, "unknown"},
		{"transport loss", errors.New("connection reset after send"), "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := screenplay.NewMemoryStore()
			_, err := store.CreateProject(ctx, screenplay.Project{ID: "reject-test", OwnerUsername: "alice", ModelID: "own", PermissionMode: screenplay.PermissionBypass})
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			gen := screenplayTextGenerator{owner: "alice", resolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
				return screenplayProviderFunc(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
					calls++
					return "", tc.upstream
				}), nil
			}}
			service := screenplay.StoryboardService{Store: store, Generator: gen}
			req := screenplay.StoryboardRequest{Owner: "alice", ProjectID: "reject-test", Engine: screenplay.EngineXiaoluoV7, Content: "第1集\n甲打开窗。", IdempotencyKey: "rejection"}
			_, _ = service.Generate(ctx, req)
			state, err := store.LoadStoryboardGenerations(ctx, "alice", "reject-test")
			if err != nil {
				t.Fatal(err)
			}
			attempt := state.Attempts[req.IdempotencyKey]
			if attempt == nil {
				t.Fatal("missing saved attempt")
			}
			if got := attempt.Episodes[1].Calls["storyboard"].Status; got != tc.want {
				t.Fatalf("status=%s want=%s", got, tc.want)
			}
			_, _ = service.Generate(ctx, req)
			if calls != 1 {
				t.Fatalf("replay billed %d calls", calls)
			}
		})
	}
}

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
