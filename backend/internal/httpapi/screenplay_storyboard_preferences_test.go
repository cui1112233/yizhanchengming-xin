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

func TestScreenplayStoryboardPreferencesOwnerAndSharedSelectedSnapshot(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	_, _ = projects.CreateProject(ctx, screenplay.Project{ID: "shared-prefix", OwnerUsername: "alice"})
	authors := &screenplay.AuthoringService{Store: projects, Repository: screenplay.NewMemoryAuthoringRepository()}
	_, e := authors.Save(ctx, "alice", "shared-prefix", screenplay.AuthoringState{Members: []screenplay.MemberGrant{{Username: "bob", Read: true}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	_, _ = projects.MutateStoryboardPreferences(ctx, "alice", func(v *screenplay.StoryboardPreferences) error {
		v.PrefixPresets = []screenplay.StoryboardPrefixPreset{{ID: "shown", Name: "当前前缀", Content: "当前共享可见选择"}, {ID: "private", Name: "不相关私人规则", Content: "不可泄漏私有前缀"}}
		v.SelectedPrefixID = "shown"
		return nil
	})
	_, _ = projects.MutateStoryboardPreferences(ctx, "bob", func(v *screenplay.StoryboardPreferences) error {
		v.PrefixPresets = []screenplay.StoryboardPrefixPreset{{ID: "bob", Name: "Bob私有", Content: "Bob自己的前缀"}}
		return nil
	})
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayStoryboardPreferenceRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, authors)
	shared := signedJSONRequest(t, mux, now, "bob", "GET", "/api/screenplays/shared-prefix/storyboard-preferences", nil)
	if shared.Code != 200 || !strings.Contains(shared.Body.String(), "当前共享可见选择") || strings.Contains(shared.Body.String(), "不可泄漏私有前缀") || strings.Contains(shared.Body.String(), "Bob自己的前缀") || !strings.Contains(shared.Body.String(), `"manage":false`) {
		t.Fatal(shared.Code, shared.Body.String())
	}
	global := signedJSONRequest(t, mux, now, "bob", "GET", "/api/screenplays/storyboard-preferences", nil)
	if global.Code != 200 || !strings.Contains(global.Body.String(), "Bob自己的前缀") || strings.Contains(global.Body.String(), "当前共享可见选择") {
		t.Fatal(global.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "mallory", "GET", "/api/screenplays/shared-prefix/storyboard-preferences", nil)
	if foreign.Code == 200 {
		t.Fatal("foreign project leaked prefs")
	}
	response := signedJSONRequest(t, mux, now, "bob", "POST", "/api/screenplays/storyboard-preferences/actions", map[string]any{"expected_revision": 1, "action": "prefix_save", "preset": map[string]any{"id": "new", "name": "新增", "content": "只属于Bob"}})
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	owner, _ := projects.LoadStoryboardPreferences(ctx, "alice")
	if len(owner.PrefixPresets) != 2 {
		t.Fatal(owner)
	}
	response = signedJSONRequest(t, mux, now, "bob", "POST", "/api/screenplays/storyboard-preferences/actions", map[string]any{"expected_revision": 1, "action": "prefix_delete", "preset_id": "new"})
	if response.Code != 409 {
		t.Fatal(response.Code, response.Body.String())
	}
}

func TestScreenplayStoryboardAssistantDraftDoesNotCallOrPersist(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	_, _ = projects.CreateProject(ctx, screenplay.Project{ID: "draft-p", OwnerUsername: "alice"})
	authors := &screenplay.AuthoringService{Store: projects, Repository: screenplay.NewMemoryAuthoringRepository()}
	state, e := authors.Save(ctx, "alice", "draft-p", screenplay.AuthoringState{Chapters: []screenplay.Chapter{{ID: "c", Title: "目标章节", SourceText: "原文", ScriptText: "剧本", PlanningText: "已保存规划"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayStoryboardPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, authors, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		calls++
		return nil, nil
	})
	response := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/draft-p/storyboard-assistant-draft", map[string]any{"chapter_id": "c", "task": "generate", "plan_mode": "existing", "planning_chapter_id": "c", "plan_density": "sd25", "use_original_content_directly": true})
	if response.Code != 200 || calls != 0 || !strings.Contains(response.Body.String(), "复用已有规划") || !strings.Contains(response.Body.String(), "完整原文") || !strings.Contains(response.Body.String(), "30秒一组") || strings.Contains(response.Body.String(), "SystemPrompt") {
		t.Fatal(response.Code, response.Body.String(), calls)
	}
	after, _ := authors.Load(ctx, "alice", "draft-p")
	if after.Revision != state.Revision || after.Chapters[0].PlanningText != "已保存规划" {
		t.Fatal(after)
	}
}

func TestScreenplayStoryboardPlanAndGenerationHTTPConsumeSavedChoices(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	_, _ = projects.CreateProject(ctx, screenplay.Project{ID: "e2-p", OwnerUsername: "alice", ModelID: "chosen-text", PermissionMode: screenplay.PermissionBypass, CreativeConfig: &screenplay.CreativeConfig{StoryboardPromptID: "storyboard_xiaoluo_v7_15s"}})
	authors := &screenplay.AuthoringService{Store: projects, Repository: screenplay.NewMemoryAuthoringRepository()}
	state, e := authors.Save(ctx, "alice", "e2-p", screenplay.AuthoringState{Chapters: []screenplay.Chapter{{ID: "c", Title: "目标", SourceText: "完整原文输入", ScriptText: "完整剧本输入"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	provider := screenplayProviderFunc(func(_ context.Context, in batchfactoryv11.TextCompletionRequest) (string, error) {
		calls++
		if strings.HasPrefix(in.UserPrompt, "【规划体例】") {
			if !strings.Contains(in.UserPrompt, "【本集剧本】\n完整剧本输入") || !in.DisableJSONResponse {
				t.Fatal(in)
			}
			return "实际保存规划文本", nil
		}
		if calls == 2 {
			if !strings.Contains(in.UserPrompt, "完整剧本输入") || !strings.Contains(in.UserPrompt, "实际保存规划文本") {
				t.Fatal(in)
			}
		} else if calls == 3 && in.UserPrompt != "完整原文输入" {
			t.Fatal(in)
		}
		return `{"storyboard":[{"id":1,"scene_id":1,"duration_sec":12,"characters":"甲","scene":"房间","props":"无","video_desc":"甲走进房间。"}]}`, nil
	})
	resolver := func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" || model != "chosen-text" {
			t.Fatal(owner, model)
		}
		return provider, nil
	}
	mux := http.NewServeMux()
	now := time.Now()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterScreenplayWorkflowRoutes(mux, auth, projects, resolver, func() time.Time { return now }, nil, authors)
	RegisterScreenplayStoryboardPlanRoutes(mux, auth, authors, resolver)
	planBody := map[string]any{"chapter_id": "c", "expected_revision": state.Revision, "idempotencyKey": "http-e2-once", "auto_continue": false}
	planned := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/e2-p/storyboard-plan", planBody)
	if planned.Code != 200 || !strings.Contains(planned.Body.String(), `"planning_request_key":"http-e2-once"`) || !strings.Contains(planned.Body.String(), "实际保存规划文本") {
		t.Fatal(planned.Code, planned.Body.String())
	}
	replay := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/e2-p/storyboard-plan", planBody)
	if replay.Code != 200 || calls != 1 {
		t.Fatal(replay.Code, replay.Body.String(), calls)
	}
	generated := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/e2-p/storyboards/generate", map[string]any{"idempotencyKey": "http-generation-e72bdddb3b4aa8", "engine": screenplay.EngineXiaoluoV7, "episodeNo": 1, "chapterId": "c", "planningChapterId": "c", "usePlanning": true})
	if generated.Code != 200 {
		t.Fatal(generated.Code, generated.Body.String())
	}
	original := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/e2-p/storyboards/generate", map[string]any{"idempotencyKey": "http-generation-2075ca27e2a9b8", "engine": screenplay.EngineXiaoluoV7, "episodeNo": 1, "chapterId": "c", "usePlanning": false, "useOriginalContentDirectly": true})
	if original.Code != 200 || calls != 3 {
		t.Fatal(original.Code, original.Body.String(), calls)
	}
}
