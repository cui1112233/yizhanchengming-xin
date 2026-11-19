package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func translateTestShots() []StoryboardShot {
	return []StoryboardShot{
		{ID: "s1", Seq: 1, ImagePrompt: "夜晚的街道，张三 walking", VideoDesc: "镜头推进，街道尽头", Line: "你好"},
		{ID: "s2", Seq: 2, ImagePrompt: "", VideoDesc: "", Line: "只有台词的空镜"},
		{ID: "s3", Seq: 3, ImagePrompt: "卧室，清晨", VideoDesc: "", Line: ""},
		{ID: "s4", Seq: 4, ImagePrompt: "   ", VideoDesc: "学校操场", Line: "上课了"},
	}
}

func TestPlanBatchTranslateAllScope(t *testing.T) {
	items, err := PlanBatchTranslate(translateTestShots(), BatchTranslateInput{Field: TranslateFieldImage, Direction: "zh2en", Scope: TranslateScopeAll})
	if err != nil {
		t.Fatal(err)
	}
	// s2 画面提示词为空 → 回退台词「你好」？不对：s2 的台词是「只有台词的空镜」，回退后可翻译。
	if len(items) != 4 {
		t.Fatalf("全部范围应命中 4 条（s2 回退台词），got %d", len(items))
	}
	if !items[1].FromSubtitle {
		t.Fatalf("s2 应标记为字幕回退")
	}
	if items[1].Text != "只有台词的空镜" {
		t.Fatalf("s2 回退文本错误: %q", items[1].Text)
	}
}

func TestPlanBatchTranslateIncompleteScopeUsesFieldText(t *testing.T) {
	items, err := PlanBatchTranslate(translateTestShots(), BatchTranslateInput{Field: TranslateFieldImage, Direction: "zh2en", Scope: TranslateScopeIncomplete})
	if err != nil {
		t.Fatal(err)
	}
	// 洛水「未完成」= 目标字段提示词非空；s2 画面提示词为空、s4 是空白 → 只剩 s1、s3。
	if len(items) != 2 {
		t.Fatalf("未完成范围应命中 2 条，got %d", len(items))
	}
	if items[0].Seq != 1 || items[1].Seq != 3 {
		t.Fatalf("未完成范围应只含 seq 1 和 3: %+v", items)
	}
}

func TestPlanBatchTranslateVideoField(t *testing.T) {
	items, err := PlanBatchTranslate(translateTestShots(), BatchTranslateInput{Field: TranslateFieldVideo, Direction: "en2zh", Scope: TranslateScopeAll})
	if err != nil {
		t.Fatal(err)
	}
	// s3 视频描述为空且没有台词 → 跳过；s1/s2/s4 命中。
	if len(items) != 3 {
		t.Fatalf("视频字段全部范围应命中 3 条，got %d", len(items))
	}
	if items[1].Seq != 2 || !items[1].FromSubtitle {
		t.Fatalf("s2 应回退字幕: %+v", items[1])
	}
}

func TestPlanBatchTranslateCustomRange(t *testing.T) {
	items, err := PlanBatchTranslate(translateTestShots(), BatchTranslateInput{Field: TranslateFieldImage, Direction: "zh2en", Scope: TranslateScopeCustom, ShotStart: 2, ShotEnd: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Seq != 2 || items[1].Seq != 3 {
		t.Fatalf("自定义 2-3 应命中 s2（回退台词）与 s3: %+v", items)
	}
	if _, err := PlanBatchTranslate(translateTestShots(), BatchTranslateInput{Field: TranslateFieldImage, Direction: "zh2en", Scope: TranslateScopeCustom, ShotStart: 3, ShotEnd: 2}); err == nil {
		t.Fatalf("结束小于起始应报错")
	}
}

func TestPlanBatchTranslateValidation(t *testing.T) {
	cases := []BatchTranslateInput{
		{Field: "audio", Direction: "zh2en", Scope: TranslateScopeAll},
		{Field: TranslateFieldImage, Direction: "fr2de", Scope: TranslateScopeAll},
		{Field: TranslateFieldImage, Direction: "zh2en", Scope: "locked"},
	}
	for i, in := range cases {
		if _, err := PlanBatchTranslate(translateTestShots(), in); err == nil {
			t.Fatalf("case %d 应报错", i)
		}
	}
}

func TestBatchTranslateSystemPrompt(t *testing.T) {
	en := BatchTranslateSystemPrompt("zh2en")
	if !strings.Contains(en, "英文") {
		t.Fatalf("zh2en 系统提示词应指向英文: %q", en)
	}
	zh := BatchTranslateSystemPrompt("en2zh")
	if !strings.Contains(zh, "中文") {
		t.Fatalf("en2zh 系统提示词应指向中文: %q", zh)
	}
}

func seedTranslateStoryboard(t *testing.T, owner string) (*MemoryStore, Storyboard) {
	t.Helper()
	store := NewMemoryStore()
	row, err := store.SaveStoryboard(context.Background(), Storyboard{
		ProjectID: "comic-translate", OwnerUsername: owner, EpisodeNo: 1, Engine: EngineXiaoluoV7,
		Shots: []StoryboardShot{
			{ID: "tr-shot-1", Seq: 1, ImagePrompt: "夜晚的街道，张三走着", VideoDesc: "镜头推进，街道尽头", Line: "你好"},
			{ID: "tr-shot-2", Seq: 2, ImagePrompt: "卧室，清晨", VideoDesc: "", Line: ""},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return store, row
}

func TestTranslateStoryboardShotWritesBack(t *testing.T) {
	store, row := seedTranslateStoryboard(t, "alice")
	ctx := context.Background()
	saved, translated, err := TranslateStoryboardShot(ctx, store, "alice", row.ProjectID, 1, EngineXiaoluoV7, "tr-shot-1",
		BatchTranslateInput{Field: TranslateFieldVideo, Direction: "zh2en", Scope: TranslateScopeAll},
		func(ctx context.Context, text string) (string, error) { return "translated: " + text, nil })
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision <= row.Revision {
		t.Fatalf("保存应推进 revision: %d -> %d", row.Revision, saved.Revision)
	}
	if translated != "translated: 镜头推进，街道尽头" {
		t.Fatalf("translated = %q", translated)
	}
	reloaded, err := store.GetStoryboard(ctx, "alice", row.ProjectID, 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Shots[0].VideoDesc != "translated: 镜头推进，街道尽头" {
		t.Fatalf("视频描述应写回翻译结果: %q", reloaded.Shots[0].VideoDesc)
	}
	if reloaded.Shots[0].ImagePrompt != "夜晚的街道，张三走着" {
		t.Fatalf("未选字段不应被改动: %q", reloaded.Shots[0].ImagePrompt)
	}
}

func TestTranslateStoryboardShotEmptyTranslation(t *testing.T) {
	store, row := seedTranslateStoryboard(t, "bob")
	ctx := context.Background()
	_, _, err := TranslateStoryboardShot(ctx, store, "bob", row.ProjectID, 1, EngineXiaoluoV7, "tr-shot-1",
		BatchTranslateInput{Field: TranslateFieldImage, Direction: "zh2en", Scope: TranslateScopeAll},
		func(ctx context.Context, text string) (string, error) { return "  ", nil })
	if err == nil || !strings.Contains(err.Error(), "空翻译") {
		t.Fatalf("空翻译应报错: %v", err)
	}
	_, _, err = TranslateStoryboardShot(ctx, store, "bob", row.ProjectID, 1, EngineXiaoluoV7, "missing-shot",
		BatchTranslateInput{Field: TranslateFieldImage, Direction: "zh2en", Scope: TranslateScopeAll},
		func(ctx context.Context, text string) (string, error) { return "x", nil })
	if err == nil || !strings.Contains(err.Error(), "不存在") {
		t.Fatalf("不存在的分镜应报错: %v", err)
	}
	_, _, err = TranslateStoryboardShot(ctx, store, "bob", row.ProjectID, 1, EngineXiaoluoV7, "tr-shot-2",
		BatchTranslateInput{Field: TranslateFieldVideo, Direction: "zh2en", Scope: TranslateScopeAll},
		func(ctx context.Context, text string) (string, error) { return "x", nil })
	if err == nil || !strings.Contains(err.Error(), "没有可翻译的文本") {
		t.Fatalf("无字幕无提示词应报错: %v", err)
	}
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("错误应归入 ErrInvalidState")
	}
}
