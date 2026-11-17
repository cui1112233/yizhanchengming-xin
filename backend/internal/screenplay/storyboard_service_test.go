package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

const storyboardJSON = `{"storyboard":[
{"id":1,"scene_id":1,"duration_sec":12,"characters":"顾苒苒,吴姐","scene":"蓝雨传媒28楼·经纪人办公室","props":"解约协议","video_desc":"画风: 3D国漫\n\n本片段场景设定在：办公室\n\n分镜1｜3s：顾苒苒坐着。"},
{"id":2,"scene_id":2,"duration_sec":13.5,"characters":"顾苒苒","scene":"顾苒苒公寓","props":"无","video_desc":"画风: 3D国漫\n\n本片段场景设定在：公寓\n\n分镜1｜4s：她瘫在沙发上。"}
]}`

func newStoryboardService(t *testing.T, mode PermissionMode) (*StoryboardService, *MemoryStore, *fakeGenerator) {
	t.Helper()
	store := NewMemoryStore()
	if _, err := store.CreateProject(context.Background(), Project{
		ID: "p1", OwnerUsername: "alice", Title: "测试", ArtStyle: "3D国漫", PermissionMode: mode,
	}); err != nil {
		t.Fatal(err)
	}
	generator := &fakeGenerator{responses: []string{storyboardJSON}}
	// Existing fixtures assert serial call order; dedicated concurrency tests use the default service.
	return &StoryboardService{Store: store, Generator: generator, SceneConcurrency: 1}, store, generator
}

func TestResolveStoryboardEngine(t *testing.T) {
	if engine, promptID, err := resolveStoryboardEngine(""); err != nil || engine != EngineXiaoyunqueV3 || promptID != "storyboard_xiaoyunque_v3" {
		t.Fatalf("默认引擎 = %s/%s/%v", engine, promptID, err)
	}
	if engine, promptID, err := resolveStoryboardEngine(EngineXiaoluoV7); err != nil || engine != EngineXiaoluoV7 || promptID != "storyboard_xiaoluo_v7_15s" {
		t.Fatalf("v7 引擎 = %s/%s/%v", engine, promptID, err)
	}
	if _, _, err := resolveStoryboardEngine("unknown"); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestGenerateStoryboardTwoEnginesStoredSeparately(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	if _, err := service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-1",
		Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3, Content: "# 第1集：A\n正文",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Generate(ctx, StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-2",
		Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Content: "# 第1集：A\n正文",
	}); err != nil {
		t.Fatal(err)
	}
	if generator.callCount() != 2 {
		t.Fatalf("调用次数 = %d, want 2", generator.callCount())
	}
	v3, err := store.ListStoryboards(ctx, "alice", "p1", EngineXiaoyunqueV3)
	if err != nil || len(v3) != 1 {
		t.Fatalf("v3 = %d/%v", len(v3), err)
	}
	v7, err := store.ListStoryboards(ctx, "alice", "p1", EngineXiaoluoV7)
	if err != nil || len(v7) != 1 {
		t.Fatalf("v7 = %d/%v", len(v7), err)
	}
	if v3[0].ID == v7[0].ID {
		t.Fatal("两套分镜必须独立落表，不能共享记录")
	}
	if v3[0].ShotCount != 2 || len(v7[0].Shots) != 2 {
		t.Fatalf("镜头数 v3=%d v7=%d", v3[0].ShotCount, len(v7[0].Shots))
	}
	if v3[0].Shots[0].Characters != "顾苒苒,吴姐" {
		t.Fatalf("characters = %q", v3[0].Shots[0].Characters)
	}
}

func TestGenerateStoryboardRendersPlaceholdersAndEpisodeContext(t *testing.T) {
	service, store, _ := newStoryboardService(t, PermissionBypass)
	generator := &fakeGenerator{responses: []string{storyboardJSON, storyboardJSON, storyboardJSON}}
	service.Generator = generator
	if _, err := store.SaveScreenplay(context.Background(), Document{
		ID: "doc1", ProjectID: "p1", OwnerUsername: "alice", Source: "b", Content: sampleScreenplay,
	}, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	_, err := service.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-3",
		Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if generator.callCount() != 3 {
		t.Fatalf("调用次数 = %d, want 3（第一集两场，第二集一场）", generator.callCount())
	}
	first := generator.requests[0]
	if strings.Contains(first.SystemPrompt, "{CURRENT_SCENE_CONTENT}") {
		t.Fatal("占位符未被替换")
	}
	if !strings.Contains(first.SystemPrompt, "第1集") || !strings.Contains(first.SystemPrompt, "蓝雨传媒28楼") {
		t.Fatal("第一集内容未注入")
	}
	if strings.Contains(first.SystemPrompt, "小河村·祖宅门口") {
		t.Fatal("下一集不能当作相邻场景注入")
	}
	second := generator.requests[1]
	if !strings.Contains(second.SystemPrompt, "蓝雨传媒28楼") {
		t.Fatal("同集上一场内容应作为 PREVIOUS_SCENE_CONTENT 注入")
	}
	if !strings.Contains(second.SystemPrompt, "画风") || !strings.Contains(second.SystemPrompt, "3D国漫") {
		t.Fatal("style 占位符未注入")
	}
}

func TestGenerateStoryboardIdempotentOverwrite(t *testing.T) {
	service, store, _ := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	request := StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-4", Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Content: "# 第1集：A\n正文"}
	if _, err := service.Generate(ctx, request); err != nil {
		t.Fatal(err)
	}
	first, err := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	request.IdempotencyKey = "explicit-regenerate-new-intent"
	if _, err := service.Generate(ctx, request); err != nil {
		t.Fatal(err)
	}
	second, err := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("重复生成必须覆盖同一条记录: %s != %s", first.ID, second.ID)
	}
	if second.Revision != first.Revision+1 {
		t.Fatalf("revision = %d, want %d", second.Revision, first.Revision+1)
	}
	if len(second.Shots) != 2 {
		t.Fatalf("镜头数 = %d, want 2（应整批替换，不累积）", len(second.Shots))
	}
}

func TestGenerateStoryboardEmptyContentRejected(t *testing.T) {
	service, _, generator := newStoryboardService(t, PermissionBypass)
	_, err := service.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-5",
		Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3, Content: "   ",
	})
	if !errors.Is(err, ErrNoActiveScript) && !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("err = %v, want 空输入错误", err)
	}
	if generator.callCount() != 0 {
		t.Fatal("空输入不应调用模型")
	}
	_, err = service.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-6",
		Owner: "alice", ProjectID: "p1", Episodes: map[int]string{1: "  "},
	})
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("err = %v, want ErrEmptyContent", err)
	}
}

func TestGenerateStoryboardAutoPermissionReturnsPlanOnly(t *testing.T) {
	service, _, generator := newStoryboardService(t, PermissionAuto)
	result, err := service.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-7",
		Owner: "alice", ProjectID: "p1", Content: "# 第1集：A\n正文",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.RequiresConfirmation || len(result.PendingActions) != 1 {
		t.Fatalf("auto 权限应只返回计划: %+v", result)
	}
	if result.PendingActions[0].StepKey != "5_storyboard" {
		t.Fatalf("step = %q", result.PendingActions[0].StepKey)
	}
	if generator.callCount() != 0 {
		t.Fatal("auto 权限未确认时不应调用模型")
	}
}

func TestGenerateStoryboardCanceled(t *testing.T) {
	service, _, generator := newStoryboardService(t, PermissionBypass)
	generator.err = context.Canceled
	_, err := service.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-8",
		Owner: "alice", ProjectID: "p1", Content: "# 第1集：A\n正文",
	})
	if !errors.Is(err, ErrUserCancelled) {
		t.Fatalf("err = %v, want ErrUserCancelled", err)
	}
}

func TestGenerateStoryboardPartialFailureReturnsFailedEpisode(t *testing.T) {
	service, _, generator := newStoryboardService(t, PermissionBypass)
	generator.responses = []string{storyboardJSON, "这不是 JSON"}
	result, err := service.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "test-storyboardservicetestgo-9",
		Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3,
		Episodes: map[int]string{1: "# 第1集：A\n正文", 2: "# 第2集：B\n正文"},
	})
	if !errors.Is(err, ErrPartialFailure) {
		t.Fatalf("部分剧集失败必须返回 ErrPartialFailure，err = %v", err)
	}
	if !result.PartialFailure || len(result.FailedEpisodes) != 1 || result.FailedEpisodes[0] != 2 {
		t.Fatalf("失败集号不完整: %+v", result.FailedEpisodes)
	}
	if len(result.Storyboards) != 1 {
		t.Fatalf("storyboards = %d, want 1（成功集仍要保存）", len(result.Storyboards))
	}
}

func TestParseStoryboardJSONErrors(t *testing.T) {
	if _, err := parseStoryboardJSON("这不是 JSON"); !errors.Is(err, ErrAIInvalidJSON) {
		t.Fatalf("err = %v, want ErrAIInvalidJSON", err)
	}
	if _, err := parseStoryboardJSON(`{"storyboard": []}`); !errors.Is(err, ErrAIInvalidStoryboard) {
		t.Fatalf("err = %v, want ErrAIInvalidStoryboard", err)
	}
	if _, err := parseStoryboardJSON(`{"storyboard": [}`); !errors.Is(err, ErrAIJSONParseRaw) {
		t.Fatalf("err = %v, want ErrAIJSONParseRaw", err)
	}
	if _, err := parseStoryboardJSON(""); !errors.Is(err, ErrAIEmptyContent) {
		t.Fatalf("err = %v, want ErrAIEmptyContent", err)
	}
}

func TestParseStoryboardJSONTolerantFields(t *testing.T) {
	raw := `{"storyboard":[{"id":"3","scene_id":"2","duration_sec":"12s",
"characters":["顾苒苒","吴姐"],"scene":"办公室","props":"无","video_desc":"画风: x\n正文"}]}`
	shots, err := parseStoryboardJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(shots) != 1 {
		t.Fatalf("shots = %d", len(shots))
	}
	shot := shots[0]
	if shot.Seq != 3 || shot.SceneID != 2 || shot.Duration != 12 {
		t.Fatalf("字段解析错误: %+v", shot)
	}
	if shot.Characters != "顾苒苒，吴姐" {
		t.Fatalf("characters = %q", shot.Characters)
	}
}

func TestValidateStoryboardGroups(t *testing.T) {
	known := knownNames{characters: map[string]bool{"顾苒苒": true}, scenes: map[string]bool{"办公室": true}, props: map[string]bool{}}
	shots := []StoryboardShot{
		{Seq: 1, Duration: 16, Characters: "顾苒苒,吴姐", Scene: "办公室", Props: "解约协议", VideoDesc: "画风: x"},
	}
	warnings := validateStoryboardGroups(EngineXiaoluoV7, 2, shots, known)
	if len(warnings) != 2 {
		t.Fatalf("warnings = %v", warnings)
	}
	if !strings.Contains(strings.Join(warnings, "|"), "超过 15s 硬上限") {
		t.Fatal("应提示 v7 时长硬上限")
	}
	if !strings.Contains(strings.Join(warnings, "|"), "吴姐") {
		t.Fatal("应提示角色不在信息库")
	}

	v3Warnings := validateStoryboardGroups(EngineXiaoyunqueV3, 1, []StoryboardShot{
		{Seq: 1, Duration: 12, Characters: "无", Scene: "办公室", Props: "无", VideoDesc: "正文没有风格行"},
	}, known)
	if !strings.Contains(strings.Join(v3Warnings, "|"), "画风:") {
		t.Fatal("应提示 v3 首行缺少风格行")
	}
}

func TestSplitEpisodesKeepsOriginalText(t *testing.T) {
	episodes := splitEpisodes(sampleScreenplay)
	if len(episodes) != 2 || episodes[0].No != 1 || episodes[1].No != 2 {
		t.Fatalf("episodes = %+v", episodes)
	}
	if !strings.HasPrefix(episodes[0].Content, "# 第1集") {
		t.Fatal("切分必须保留集标题原文")
	}
	if strings.Contains(episodes[0].Content, "# 第2集") {
		t.Fatal("第一集不应包含第二集内容")
	}
}

func TestStoryboardJSONShapeStaysSixFields(t *testing.T) {
	shots := []StoryboardShot{{Seq: 1, SceneID: 1, Duration: 12, Characters: "A", Scene: "B", Props: "C", VideoDesc: "D"}}
	raw, err := json.Marshal(shots[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"id"`, `"scene_id"`, `"duration_sec"`, `"characters"`, `"scene"`, `"props"`, `"video_desc"`} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("缺少字段 %s: %s", key, raw)
		}
	}
}
