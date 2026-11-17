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

func TestScreenplayZIP164SavedSelectionGeneratesListsAndExports(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	_, err := store.CreateProject(ctx, screenplay.Project{ID: "zip164-http", OwnerUsername: "alice", ModelID: "our-model164", PermissionMode: screenplay.PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	provider := screenplayProviderFunc(func(_ context.Context, in batchfactoryv11.TextCompletionRequest) (string, error) {
		calls++
		if !strings.Contains(in.SystemPrompt, "当前剧本164") || strings.Contains(in.SystemPrompt, "{{剧本内容}}") {
			t.Fatal("unrendered system prompt")
		}
		return `{"storyboard":[{"id":1,"source_elements":["ID-001"],"characters":"甲","scene":"房间","image_prompt":"四宫格图片词164","voice_actor":"","emotion":"平静","line":"","video_desc":"四帧视频词164"}]}`, nil
	})
	resolver := func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || model != "our-model164" {
			t.Fatal("own model not used", owner, model)
		}
		return provider, nil
	}
	now := time.Now()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterScreenplayWorkflowRoutes(mux, auth, store, resolver, func() time.Time { return now }, nil, nil)
	saved := signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/zip164-http/creative-config", map[string]any{"expected_revision": 0, "storyboard_prompt_id": "164"})
	if saved.Code != 200 {
		t.Fatal(saved.Code, saved.Body.String())
	}
	p, _ := store.GetProject(ctx, "alice", "zip164-http")
	if p.CreativeConfig.PromptSnapshots["storyboard"].Digest != "b3ea2ac397a09cee548464e5834d0c1dafcb2439649ae642fbe8547627dcf328" {
		t.Fatal("snapshot missing")
	}
	body := map[string]any{"idempotencyKey": "zip164-http-once", "content": "# 第1集\n当前剧本164"}
	for i := 0; i < 2; i++ {
		res := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/zip164-http/storyboards/generate", body)
		if res.Code != 200 {
			t.Fatal(res.Code, res.Body.String())
		}
	}
	if calls != 1 {
		t.Fatal("duplicate request", calls)
	}
	for _, suffix := range []string{"/workspace", "/storyboards", "/export"} {
		res := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects/zip164-http"+suffix, nil)
		if res.Code != 200 || !strings.Contains(res.Body.String(), "prompt_164") || !strings.Contains(res.Body.String(), "四宫格图片词164") || !strings.Contains(res.Body.String(), "四帧视频词164") {
			t.Fatal(suffix, res.Code, res.Body.String())
		}
	}
}
