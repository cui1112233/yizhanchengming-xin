package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplay139FrozenContractRequiresExplicitReselectionBeforeGeneration(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	project, err := store.CreateProject(ctx, screenplay.Project{ID: "frozen-139", OwnerUsername: "alice", ModelID: "own-model", PermissionMode: screenplay.PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	selected, _ := prompt.ResolveCreativePrompt("139", "storyboard")
	frozen := selected
	frozen.Supported = false
	frozen.UnsupportedReason = "旧版未审核"
	frozen.ParserVersion = "storyboard_image_voice_v1"
	frozen.RequiredFields = append(append([]string{}, frozen.RequiredFields...), "image_prompt")
	other, _ := prompt.ResolveCreativePrompt("144", "script")
	other.Body += "\n其他冻结正文必须保留"
	config := screenplay.EffectiveCreativeConfig(project)
	config.ScriptPromptID = other.ID
	config.StoryboardPromptID = frozen.ID
	config.StoryboardAutoSyncAssets = false
	config.PromptSnapshots = map[string]screenplay.CreativePromptSnapshot{"storyboard": frozen, "script": other}
	project, err = store.SaveCreativeConfig(ctx, "alice", project.ID, config, project.Revision)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	now := time.Now()
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		return screenplayProviderFunc(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
			calls++
			return `{"storyboard":[{"id":1,"source_elements":["1-1-001"],"characters":"甲","scene":"家","voice_actor":"甲","emotion":"平静","line":"你好","video_desc":"甲在走廊说你好。"}]}`, nil
		}), nil
	}, func() time.Time { return now })
	generation := map[string]any{"idempotencyKey": "frozen-139-before-key", "engine": "prompt_139", "episodeNo": 1, "content": "# 第1集：测试\n1-1 日 内 家\n甲：你好", "confirmed": true}
	blocked := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+project.ID+"/storyboards/generate", generation)
	if blocked.Code != 400 || calls != 0 {
		t.Fatal("old frozen contract executed implicitly", blocked.Code, blocked.Body.String(), calls)
	}
	unrelated := signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/"+project.ID+"/creative-config", map[string]any{"expected_revision": project.Revision, "image_resolution": "2K"})
	if unrelated.Code != 200 {
		t.Fatal(unrelated.Code, unrelated.Body.String())
	}
	project, _ = store.GetProject(ctx, "alice", project.ID)
	if project.CreativeConfig.PromptSnapshots["storyboard"].Supported {
		t.Fatal("unrelated setting silently upgraded a paid protocol")
	}
	reselect := signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/"+project.ID+"/creative-config", map[string]any{"expected_revision": project.Revision, "storyboard_prompt_id": "139"})
	if reselect.Code != 200 {
		t.Fatal(reselect.Code, reselect.Body.String())
	}
	project, _ = store.GetProject(ctx, "alice", project.ID)
	current := project.CreativeConfig.PromptSnapshots["storyboard"]
	if !current.Supported || current.ParserVersion != "cm_storyboard_139_optional_image_v2" || current.Body != frozen.Body || current.Digest != frozen.Digest || project.CreativeConfig.PromptSnapshots["script"].Body != other.Body || calls != 0 {
		t.Fatalf("explicit refresh failed or changed frozen bodies: supported=%v parser=%s calls=%d", current.Supported, current.ParserVersion, calls)
	}
	generation["idempotencyKey"] = "frozen-139-after-key"
	generated := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+project.ID+"/storyboards/generate", generation)
	if generated.Code != 200 || calls != 1 {
		t.Fatal(generated.Code, generated.Body.String(), calls)
	}
	board, err := store.GetStoryboard(ctx, "alice", project.ID, 1, screenplay.StoryboardEngine("prompt_139"))
	if err != nil || len(board.Shots) != 1 || board.Shots[0].ImagePrompt != "" || board.Shots[0].VideoDesc == "" || board.Shots[0].Duration != 15 || !strings.Contains(string(board.Shots[0].ProtocolFields), "cm_storyboard_139_optional_image_v2") {
		t.Fatal(board, err)
	}
}
