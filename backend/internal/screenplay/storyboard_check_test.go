package screenplay

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func checkShots() []StoryboardShot {
	return []StoryboardShot{
		{ID: "ck-1", Seq: 1, VideoDesc: "开场推进", ImagePrompt: "夜街", Characters: "林川", Scene: "街道", Duration: 5},
		{ID: "ck-2", Seq: 2, VideoDesc: "跟拍", Characters: "林川、阿紫", Duration: 4},
		{ID: "ck-3", Seq: 4, VideoDesc: "对峙", Characters: "阿紫", Scene: "屋顶", Duration: 6},
		{ID: "ck-4", Seq: 4, VideoDesc: "重复镜头"},
		{ID: "ck-5", Seq: 7, VideoDesc: "收尾"},
		{ID: "ck-6", Seq: -3, VideoDesc: "异常序号"},
	}
}

func TestCheckStoryboardShotsReportsIssues(t *testing.T) {
	report, err := CheckStoryboardShots(checkShots())
	if err != nil {
		t.Fatal(err)
	}
	// 有效序号：1,2,4,4,7 → max=7，缺失 3、5、6；4 重复；-3 异常。
	if report.Total != 6 {
		t.Fatalf("total=%d", report.Total)
	}
	if report.ExpectedTotal != 7 || report.MinSeq != 1 || report.MaxSeq != 7 {
		t.Fatalf("range: expected=%d min=%d max=%d", report.ExpectedTotal, report.MinSeq, report.MaxSeq)
	}
	if fmt.Sprint(report.Missing) != "[3 5 6]" {
		t.Fatalf("missing=%v", report.Missing)
	}
	if fmt.Sprint(report.Duplicates) != "[4]" {
		t.Fatalf("duplicates=%v", report.Duplicates)
	}
	if fmt.Sprint(report.Invalid) != "[-3]" {
		t.Fatalf("invalid=%v", report.Invalid)
	}
	if len(report.Extra) != 0 {
		t.Fatalf("V88 无原文对照，extra 应为空: %v", report.Extra)
	}
	if report.Passed {
		t.Fatal("有问题时应为未通过")
	}
}

func TestCheckStoryboardShotsPassesOnContinuous(t *testing.T) {
	report, err := CheckStoryboardShots([]StoryboardShot{
		{ID: "ok-1", Seq: 1},
		{ID: "ok-2", Seq: 2},
		{ID: "ok-3", Seq: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed || len(report.Missing) != 0 || len(report.Duplicates) != 0 || len(report.Invalid) != 0 {
		t.Fatalf("连续序号应通过: %+v", report)
	}
	if report.Total != 3 || report.ExpectedTotal != 3 {
		t.Fatalf("total=%d expected=%d", report.Total, report.ExpectedTotal)
	}
}

func TestCheckStoryboardShotsRejectsEmpty(t *testing.T) {
	if _, err := CheckStoryboardShots(nil); err == nil || !strings.Contains(err.Error(), "还没有分镜") {
		t.Fatalf("空分镜应报错: %v", err)
	}
}

func TestStoryboardCheckFillPromptCarriesContext(t *testing.T) {
	shots := checkShots()
	prev := &shots[0]
	next := &shots[2]
	prompt := StoryboardCheckFillPrompt(3, prev, next)
	for _, want := range []string{"分镜序号 3 缺失", "前一分镜（序号 1）", "后一分镜（序号 4）", "视频提示词：开场推进", "人物：林川", "JSON"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("提示词缺 %q: %s", want, prompt)
		}
	}
	// 没有邻居时也能生成（缺失在开头或结尾）。
	empty := StoryboardCheckFillPrompt(1, nil, nil)
	if !strings.Contains(empty, "分镜序号 1 缺失") {
		t.Fatalf("无邻居提示词异常: %s", empty)
	}
}

func TestStoryboardCheckFillSystemPrompt(t *testing.T) {
	prompt := StoryboardCheckFillSystemPrompt()
	if !strings.Contains(prompt, "补全") || !strings.Contains(prompt, "JSON") {
		t.Fatalf("系统提示词异常: %s", prompt)
	}
}

func TestParseStoryboardFillJSON(t *testing.T) {
	draft, err := parseStoryboardFillJSON(`{"characters":"林川","scene":"街道","props":"灯笼","video_desc":"镜头缓推","image_prompt":"夜街灯笼","line":"谁在那","duration_sec":6}`)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Characters != "林川" || draft.VideoDesc != "镜头缓推" || draft.Line != "谁在那" {
		t.Fatalf("字段解析错误: %+v", draft)
	}
	if value, err := draft.Duration.Float64(); err != nil || value != 6 {
		t.Fatalf("duration=%s err=%v", draft.Duration, err)
	}
	fenced, err := parseStoryboardFillJSON("```json\n{\"video_desc\":\"推进\"}\n```")
	if err != nil || fenced.VideoDesc != "推进" {
		t.Fatalf("代码块包裹应可解析: %+v err=%v", fenced, err)
	}
	if _, err := parseStoryboardFillJSON("模型只输出了文字"); err == nil {
		t.Fatal("无 JSON 应报错")
	}
	if _, err := parseStoryboardFillJSON("{broken"); err == nil {
		t.Fatal("坏 JSON 应报错")
	}
}

type checkStore struct{ *MemoryStore }

func TestFillMissingStoryboardShotInsertsBetween(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "check-project", OwnerUsername: "alice", Title: "分镜检查"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, Storyboard{
		ProjectID: "check-project", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7,
		Shots: []StoryboardShot{
			{ID: "f-1", Seq: 1, VideoDesc: "开场", Duration: 5},
			{ID: "f-2", Seq: 3, VideoDesc: "收尾"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	payload := `{"characters":"林川","scene":"街道","props":"灯笼","video_desc":"衔接镜头","image_prompt":"夜街","line":"跟上","duration_sec":4.5}`
	saved, shot, err := FillMissingStoryboardShot(ctx, store, "alice", "check-project", 1, EngineXiaoluoV7, 2, func(context.Context, string) (string, error) {
		return payload, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if shot.Seq != 2 || shot.VideoDesc != "衔接镜头" || shot.Duration != 4.5 {
		t.Fatalf("补全分镜错误: %+v", shot)
	}
	if shot.ID == "f-1" || shot.ID == "f-2" || shot.ID == "" {
		t.Fatalf("新分镜应生成新 ID: %q", shot.ID)
	}
	if len(saved.Shots) != 3 || saved.Shots[0].Seq != 1 || saved.Shots[1].Seq != 2 || saved.Shots[2].Seq != 3 {
		t.Fatalf("应按序号插入中间: %+v", saved.Shots)
	}
	stored, err := store.GetStoryboard(ctx, "alice", "check-project", 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Shots) != 3 || stored.Shots[1].VideoDesc != "衔接镜头" {
		t.Fatalf("未写库: %+v", stored.Shots)
	}
	report, err := CheckStoryboardShots(stored.Shots)
	if err != nil || !report.Passed {
		t.Fatalf("补全后应通过: %+v err=%v", report, err)
	}
}

func TestFillMissingStoryboardShotRejectsPresentSeq(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "check-project-2", OwnerUsername: "alice", Title: "分镜检查"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, Storyboard{
		ProjectID: "check-project-2", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7,
		Shots: []StoryboardShot{{ID: "p-1", Seq: 1}, {ID: "p-2", Seq: 2}},
	}); err != nil {
		t.Fatal(err)
	}
	called := false
	if _, _, err := FillMissingStoryboardShot(ctx, store, "alice", "check-project-2", 1, EngineXiaoluoV7, 2, func(context.Context, string) (string, error) {
		called = true
		return "{}", nil
	}); err == nil || !strings.Contains(err.Error(), "不是缺失分镜") {
		t.Fatalf("已存在序号应拒绝: %v", err)
	}
	if called {
		t.Fatal("不应调用模型")
	}
	if _, _, err := FillMissingStoryboardShot(ctx, store, "alice", "check-project-2", 1, EngineXiaoluoV7, 9, func(context.Context, string) (string, error) {
		return "{}", nil
	}); err == nil {
		t.Fatal("超出范围的序号也不是缺失分镜，应拒绝")
	}
}

func TestFillMissingStoryboardShotDefaultsDurationAndTrims(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "check-project-3", OwnerUsername: "alice", Title: "分镜检查"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, Storyboard{
		ProjectID: "check-project-3", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7,
		Shots: []StoryboardShot{{ID: "d-2", Seq: 2, Duration: 8}},
	}); err != nil {
		t.Fatal(err)
	}
	_, shot, err := FillMissingStoryboardShot(ctx, store, "alice", "check-project-3", 1, EngineXiaoluoV7, 1, func(context.Context, string) (string, error) {
		return `{"characters":" 林川 ","video_desc":"  开场  ","duration_sec":0}`, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// 无前一分镜（缺失在开头），duration_sec 无效 → 默认 5 秒。
	if shot.Duration != 5 {
		t.Fatalf("默认时长应 5 秒: %v", shot.Duration)
	}
	if shot.VideoDesc != "开场" || shot.Characters != "林川" {
		t.Fatalf("字段应去空白: %+v", shot)
	}
}
