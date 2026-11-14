package screenplay

import (
	"context"
	"errors"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

func TestStoryboardOwnerPreferencesCRUDImportRevisionAndIsolation(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	apply := func(action StoryboardPreferencesAction) StoryboardPreferences {
		t.Helper()
		v, e := store.MutateStoryboardPreferences(ctx, "alice", func(v *StoryboardPreferences) error { _, _, e := ApplyStoryboardPreferencesAction(v, action); return e })
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	v := apply(StoryboardPreferencesAction{ExpectedRevision: int64Ptr(0), Action: "prefix_save", Preset: StoryboardPrefixPreset{ID: "p", Name: "用户前缀", Content: "第一前缀"}})
	v = apply(StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "prefix_select", PresetID: "p"})
	if v.SelectedPrefix() != "第一前缀" {
		t.Fatal(v)
	}
	v = apply(StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "prefix_import", Presets: []StoryboardPrefixPreset{{ID: "p", Name: "冲突", Content: "不得覆盖"}, {ID: "q", Name: "第二", Content: "保留此文"}}})
	if len(v.PrefixPresets) != 2 || v.SelectedPrefix() != "第一前缀" {
		t.Fatal(v)
	}
	v = apply(StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "prefix_select", PresetID: "p"})
	if v.SelectedPrefixID != "" {
		t.Fatal(v)
	}
	v = apply(StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "recent", PromptID: prompt.CreativeStoryboard307})
	v = apply(StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "favorite_toggle", PromptID: prompt.CreativeStoryboard307})
	if len(v.RecentIDs) != 0 || len(v.FavoriteIDs) != 1 {
		t.Fatal(v)
	}
	other, _ := store.LoadStoryboardPreferences(ctx, "bob")
	if len(other.PrefixPresets) != 0 || len(other.FavoriteIDs) != 0 {
		t.Fatal("cross owner", other)
	}
	if _, e := store.MutateStoryboardPreferences(ctx, "alice", func(st *StoryboardPreferences) error {
		_, _, e := ApplyStoryboardPreferencesAction(st, StoryboardPreferencesAction{ExpectedRevision: int64Ptr(0), Action: "prefix_delete", PresetID: "q"})
		return e
	}); !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	v.RecentIDs = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	_, _, e := ApplyStoryboardPreferencesAction(&v, StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "recent", PromptID: "storyboard_xiaoluo_v7_15s"})
	if e != nil || len(v.RecentIDs) != 8 || v.RecentIDs[0] != "storyboard_xiaoluo_v7_15s" {
		t.Fatal(v, e)
	}
}
func int64Ptr(v int64) *int64 { return &v }

func TestStoryboardPlanOnlyDensityIdempotencyAndFinalModes(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, e := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c1", Title: "第一集", SourceText: "完整原文", ScriptText: "完整目标剧本"}, {ID: "c2", Title: "第二集", SourceText: "第二原文", ScriptText: "第二目标剧本"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.PlanDensity = "sd25"
	_, e = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	if e != nil {
		t.Fatal(e)
	}
	planGen := &fakeGenerator{responses: []string{"原文事件先后完整的30秒规划"}}
	request := StoryboardPlanRequest{ChapterID: "c1", Model: "selected-text", ExpectedRevision: &state.Revision, UseOriginal: true, IdempotencyKey: "plan-once"}
	result, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", request, planGen)
	if e != nil {
		t.Fatal(e)
	}
	if planGen.callCount() != 1 || !strings.Contains(planGen.requests[0].SystemPrompt, "30秒一组") || planGen.requests[0].UserPrompt != "完整原文" || result.Chapter.PlanningText == "" || result.Chapter.PlanDensity != "sd25" || result.Chapter.ScriptText != "完整目标剧本" {
		t.Fatal(result, planGen.requests)
	}
	if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", request, planGen); e != nil || planGen.callCount() != 1 {
		t.Fatal(e, planGen.callCount())
	}
	if boards, _ := store.ListStoryboards(ctx, "alice", "p1", EngineXiaoluoV7); len(boards) != 0 {
		t.Fatal("plan-only generated storyboard")
	}
	_, e = store.MutateStoryboardPreferences(ctx, "alice", func(v *StoryboardPreferences) error {
		v.PrefixPresets = []StoryboardPrefixPreset{{ID: "prefix", Name: "名称", Content: "全局前缀必须在原模板之前"}}
		v.SelectedPrefixID = "prefix"
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	service.Authoring = authors
	use := true
	_, e = service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-1", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, ChapterID: "c2", EpisodeNo: 2, PlanningChapterID: "c1", UsePlanning: &use, Model: "text-final"})
	if e != nil {
		t.Fatal(e)
	}
	sent := generator.requests[0]
	if !strings.HasPrefix(sent.SystemPrompt, "全局前缀必须在原模板之前\n\n") || !strings.HasPrefix(sent.UserPrompt, "第二目标剧本") || !strings.Contains(sent.UserPrompt, result.Chapter.PlanningText) {
		t.Fatal(sent)
	}
	use = false
	_, e = service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-2", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, ChapterID: "c2", EpisodeNo: 2, UsePlanning: &use, UseOriginalContentDirectly: true})
	if e != nil {
		t.Fatal(e)
	}
	sent = generator.requests[1]
	if sent.UserPrompt != "第二原文" || strings.Contains(sent.UserPrompt, result.Chapter.PlanningText) {
		t.Fatal(sent)
	}
}

func TestStoryboardContextUsesOnlyLastFiveAndAutoSyncOffDoesNotCall(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardUseContextHistory = true
	_, e := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	if e != nil {
		t.Fatal(e)
	}
	shots := []StoryboardShot{}
	for i := 1; i <= 7; i++ {
		shots = append(shots, StoryboardShot{ID: newScreenplayID("old"), Seq: i, Duration: 12, VideoDesc: "历史镜头" + string(rune('0'+i))})
	}
	_, e = store.SaveStoryboard(ctx, Storyboard{ID: "old-board", ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Status: "done", Shots: shots})
	if e != nil {
		t.Fatal(e)
	}
	_, e = service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-3", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, EpisodeNo: 2, Content: "第2集完整内容"})
	if e != nil {
		t.Fatal(e)
	}
	sent := generator.requests[0].UserPrompt
	if strings.Contains(sent, "历史镜头1") || strings.Contains(sent, "历史镜头2") || !strings.Contains(sent, "历史镜头3") || !strings.Contains(sent, "历史镜头7") || generator.callCount() != 1 {
		t.Fatal(sent, generator.callCount())
	}
	assets, _ := store.ListAssets(ctx, "alice", "p1", "")
	if len(assets) != 0 {
		t.Fatal("disabled auto sync changed assets", assets)
	}
}

func TestStoryboardAutoSyncAddsOnlyMissingAndKeepsExisting(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardAutoSyncAssets = true
	_, _ = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	_, _ = store.CreateAsset(ctx, Asset{ID: "old-hero", ProjectID: "p1", OwnerUsername: "alice", Type: AssetCharacter, Degree: DegreeMaster, Name: "顾苒苒", Description: "用户原有描述不可覆盖", Prompt: "用户原有提示词"})
	generator.responses = []string{storyboardJSON, `{"characters":[{"name":"吴姐","description":"角色补充","prompt":"经纪人形象"}],"scenes":[{"name":"蓝雨传媒28楼·经纪人办公室","description":"现代办公室","prompt":"整洁房间"},{"name":"顾苒苒公寓","description":"公寓","prompt":"公寓家具"}],"props":[{"name":"解约协议","description":"协议纸张","prompt":"白纸合同"}]}`}
	result, e := service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-4", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Content: "# 第1集\n完整内容", Model: "text-selected"})
	if e != nil {
		t.Fatal(e)
	}
	assets, _ := store.ListAssets(ctx, "alice", "p1", "")
	if result.AddedAssetCount != 4 || len(assets) != 5 || generator.callCount() != 2 || generator.requests[1].SystemPrompt != prompt.ScreenplayStoryboardMissingAssets {
		t.Fatal(result, len(assets), generator.callCount())
	}
	for _, a := range assets {
		if a.ID == "old-hero" && a.Description != "用户原有描述不可覆盖" {
			t.Fatal(a)
		}
	}
}

func TestStoryboardPlannerCancelPreservesOldPlanAndSameFailedKeyCannotRebill(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	_, _ = store.CreateProject(ctx, Project{ID: "plan-cancel", OwnerUsername: "alice", ModelID: "text-model"})
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, e := authors.Save(ctx, "alice", "plan-cancel", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "章节", SourceText: "原文", ScriptText: "剧本", PlanningText: "用户之前保存的规划", PlanDensity: "compact"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	in := StoryboardPlanRequest{ChapterID: "c", ExpectedRevision: &state.Revision, IdempotencyKey: "cancel-key"}
	_, e = authors.GenerateStoryboardPlan(cancelCtx, "alice", "plan-cancel", in, lateCanceledGenerator{cancel: cancel, content: "不能覆盖旧规划的晚返回"})
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	fresh, _ := authors.Load(ctx, "alice", "plan-cancel")
	if fresh.Chapters[0].PlanningText != "用户之前保存的规划" || fresh.Chapters[0].PlanningStatus != "cancelled" || fresh.Chapters[0].PlanningRequestKey != "cancel-key" {
		t.Fatal(fresh)
	}
	counted := &fakeGenerator{responses: []string{"不应调用"}}
	in.ExpectedRevision = &fresh.Revision
	if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "plan-cancel", in, counted); !errors.Is(e, ErrInvalidState) || counted.callCount() != 0 {
		t.Fatal(e, counted.callCount())
	}
	p, _ := store.GetProject(ctx, "alice", "plan-cancel")
	cfg := EffectiveCreativeConfig(p)
	cfg.PlanDensity = "sd25"
	_, _ = store.SaveCreativeConfig(ctx, "alice", "plan-cancel", cfg, p.Revision)
	unchanged, _ := authors.Load(ctx, "alice", "plan-cancel")
	if unchanged.Chapters[0].PlanningText != "用户之前保存的规划" || unchanged.Chapters[0].PlanDensity != "compact" || unchanged.Revision != fresh.Revision {
		t.Fatal("density save rewrote existing plan", unchanged)
	}
}

func TestStoryboardContextMultiEpisodeIsSequentialAndPreflightsMissingPlans(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardUseContextHistory = true
	_, _ = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	_, e := service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-5", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Episodes: map[int]string{1: "第一集完整输入", 2: "第二集完整输入"}})
	if e != nil {
		t.Fatal(e)
	}
	if len(generator.requests) != 2 || strings.Contains(generator.requests[0].UserPrompt, "最近5条已生成") || !strings.Contains(generator.requests[1].UserPrompt, "最近5条已生成") || !strings.Contains(generator.requests[1].UserPrompt, "顾苒苒坐着") {
		t.Fatal(generator.requests)
	}
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	_, _ = authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c1", Title: "1", SourceText: "服务端完整第一原文", ScriptText: "剧本1", PlanningText: "计划1"}, {ID: "c2", Title: "2", SourceText: "服务端完整第二原文", ScriptText: "剧本2"}}}, 0)
	service.Authoring = authors
	yes := true
	before := generator.callCount()
	_, e = service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-6", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Episodes: map[int]string{1: "1", 2: "2"}, UsePlanning: &yes})
	if !errors.Is(e, ErrInvalidArgument) || generator.callCount() != before {
		t.Fatal("missing later plan consumed first paid request", e, generator.callCount(), before)
	}
	no := false
	_, e = service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardpreparationtestgo-7", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Episodes: map[int]string{2: "客户端残片2", 1: "客户端残片1"}, UseOriginalContentDirectly: true, UsePlanning: &no})
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(generator.requests[before].UserPrompt, "服务端完整第一原文") || !strings.HasPrefix(generator.requests[before+1].UserPrompt, "服务端完整第二原文") || strings.Contains(generator.requests[before+1].UserPrompt, "客户端残片") {
		t.Fatal(generator.requests[before:])
	}
}
