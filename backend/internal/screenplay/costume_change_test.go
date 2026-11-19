package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func costumeTestShots() []StoryboardShot {
	return []StoryboardShot{
		{ID: "shot-1", Seq: 1, Scene: "卧室", Characters: "张三、小Y"},
		{ID: "shot-2", Seq: 2, Scene: "卧室", Characters: "张三"},
		{ID: "shot-3", Seq: 3, Scene: "学校", Characters: "张三、李四"},
		{ID: "shot-4", Seq: 4, Scene: "卧室", Characters: "李四"},
		{ID: "shot-5", Seq: 5, Scene: "卧室", Characters: "张三-睡衣"},
		{ID: "shot-6", Seq: 6, Scene: "客厅", Characters: "王五、张三"},
	}
}

// 洛水更新日志原文举例：筛选「卧室」把「张三」换成「张三-睡衣」，
// 只改该场景且已绑定张三的分镜；学校、未绑定张三、已经是张三-睡衣的分镜都不动。
func TestPlanCostumeChangeMatchesLuoshuiExample(t *testing.T) {
	plan, err := PlanCostumeChange(costumeTestShots(), CostumeChangeOptions{ShotStart: 1, ShotEnd: 6, Scene: "卧室", From: "张三", To: "张三-睡衣"})
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if plan.Matched != 2 {
		t.Fatalf("matched = %d, want 2", plan.Matched)
	}
	if plan.Targets[0].ShotID != "shot-1" || plan.Targets[1].ShotID != "shot-2" {
		t.Fatalf("targets = %+v", plan.Targets)
	}
	// 顺序必须保留，@图片N 才对得上。
	if plan.Targets[0].After != "张三-睡衣、小Y" {
		t.Fatalf("after = %q", plan.Targets[0].After)
	}
	if plan.Targets[0].Before != "张三、小Y" {
		t.Fatalf("before = %q", plan.Targets[0].Before)
	}
}

// 不限制场景时，范围内所有绑定张三的分镜都改，范围外的保持不动。
func TestPlanCostumeChangeWithoutSceneFilter(t *testing.T) {
	plan, err := PlanCostumeChange(costumeTestShots(), CostumeChangeOptions{ShotStart: 1, ShotEnd: 3, Scene: "", From: "张三", To: "张三-睡衣"})
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if plan.Matched != 3 {
		t.Fatalf("matched = %d, want 3", plan.Matched)
	}
	for _, target := range plan.Targets {
		if target.Seq > 3 {
			t.Fatalf("范围外分镜被改：%+v", target)
		}
	}
}

func TestPlanCostumeChangeSkipsShotsAlreadyTarget(t *testing.T) {
	shots := []StoryboardShot{
		{ID: "a", Seq: 1, Scene: "卧室", Characters: "张三-睡衣、张三"},
		{ID: "b", Seq: 2, Scene: "卧室", Characters: "张三、李四"},
	}
	plan, err := PlanCostumeChange(shots, CostumeChangeOptions{ShotStart: 1, ShotEnd: 2, From: "张三", To: "张三-睡衣"})
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	// 第一条已经是目标人物，整条不动；第二条换成目标人物后与已有李四共存。
	if plan.Matched != 1 || plan.Targets[0].ShotID != "b" {
		t.Fatalf("plan = %+v", plan)
	}
	if plan.Targets[0].After != "张三-睡衣、李四" {
		t.Fatalf("after = %q", plan.Targets[0].After)
	}
}

func TestReplaceCostumeNameKeepsOrderAndSeparators(t *testing.T) {
	cases := []struct {
		characters string
		from       string
		to         string
		want       string
		hit        bool
	}{
		{"小Y、张三、李四", "张三", "张三-睡衣", "小Y、张三-睡衣、李四", true},
		{"小Y，张三；李四", "张三", "张三-睡衣", "小Y、张三-睡衣、李四", true},
		{"张三/李四|王五", "李四", "李四-校服", "张三、李四-校服、王五", true},
		{"张三", "张三", "张三-睡衣", "张三-睡衣", true},
		{"李四、王五", "张三", "张三-睡衣", "李四、王五", false},
		{"张三-睡衣", "张三", "张三-睡衣", "张三-睡衣", false},
		{"", "张三", "张三-睡衣", "", false},
		{"张三", "张三", "张三", "张三", false},
	}
	for _, item := range cases {
		got, hit := ReplaceCostumeName(item.characters, item.from, item.to)
		if hit != item.hit || got != item.want {
			t.Fatalf("ReplaceCostumeName(%q,%q,%q) = (%q,%v), want (%q,%v)", item.characters, item.from, item.to, got, hit, item.want, item.hit)
		}
	}
}

func TestPlanCostumeChangeValidatesInput(t *testing.T) {
	shots := costumeTestShots()
	if _, err := PlanCostumeChange(shots, CostumeChangeOptions{ShotStart: 1, ShotEnd: 6, From: "张三", To: "张三"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("同名应报错：%v", err)
	}
	if _, err := PlanCostumeChange(shots, CostumeChangeOptions{ShotStart: 1, ShotEnd: 6, From: "", To: "张三"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("空人物应报错：%v", err)
	}
	if _, err := PlanCostumeChange(shots, CostumeChangeOptions{ShotStart: 9, ShotEnd: 12, From: "张三", To: "张三-睡衣"}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("范围越界应报错：%v", err)
	}
	if _, err := PlanCostumeChange(nil, CostumeChangeOptions{ShotStart: 1, ShotEnd: 1, From: "张三", To: "张三-睡衣"}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("空分镜应报错：%v", err)
	}
}

// 结束分镜不填或超出本集时收敛到本集最后一条，与洛水「分镜范围默认 1-总数」一致。
func TestPlanCostumeChangeClampsShotRange(t *testing.T) {
	plan, err := PlanCostumeChange(costumeTestShots(), CostumeChangeOptions{ShotStart: 1, ShotEnd: 0, Scene: "卧室", From: "张三", To: "张三-睡衣"})
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if plan.ShotEnd != 6 || plan.Matched != 2 {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestCostumeChangeMessagesMatchLuoshui(t *testing.T) {
	empty := CostumeChangePlan{ShotStart: 1, ShotEnd: 30, Scene: "卧室", From: "张三", To: "张三-睡衣"}
	// 洛水原文限定场景时写作「中需绑定「场景」，且已绑定「原人物」」，这里照抄不改写。
	if got := CostumeChangeEmptyMessage(empty); got != "没有可换装的分镜：分镜 1-30 中需绑定「卧室」，且已绑定「张三」，并且还不是「张三-睡衣」" {
		t.Fatalf("empty message = %q", got)
	}
	unlimited := CostumeChangePlan{ShotStart: 1, ShotEnd: 30, From: "张三", To: "张三-睡衣"}
	if got := CostumeChangeEmptyMessage(unlimited); got != "没有可换装的分镜：分镜 1-30 中需绑定「张三」，并且还不是「张三-睡衣」" {
		t.Fatalf("empty message = %q", got)
	}
	done := CostumeChangePlan{ShotStart: 1, ShotEnd: 30, Scene: "", From: "张三", To: "张三-睡衣", Matched: 4}
	if got := CostumeChangeSuccessMessage(done); got != "已在分镜 1-30将「张三」换成「张三-睡衣」，共 4 个分镜" {
		t.Fatalf("success message = %q", got)
	}
	scoped := CostumeChangePlan{ShotStart: 1, ShotEnd: 30, Scene: "卧室", From: "张三", To: "张三-睡衣", Matched: 2}
	if got := CostumeChangeSuccessMessage(scoped); got != "已在分镜 1-30（卧室）将「张三」换成「张三-睡衣」，共 2 个分镜" {
		t.Fatalf("success message = %q", got)
	}
}

func TestBuildCostumeChangeOptionsView(t *testing.T) {
	assets := []Asset{
		{Type: AssetCharacter, Name: "张三"},
		{Type: AssetCharacter, Name: "张三-睡衣"},
		{Type: AssetScene, Name: "卧室"},
		{Type: AssetProp, Name: "台灯"},
	}
	view := BuildCostumeChangeOptionsView(costumeTestShots(), assets)
	if strings.Join(view.Characters, ",") != "张三,张三-睡衣,小Y,李四,王五" {
		t.Fatalf("characters = %v", view.Characters)
	}
	if strings.Join(view.Scenes, ",") != "卧室,学校,客厅" {
		t.Fatalf("scenes = %v", view.Scenes)
	}
	if view.ShotCount != 6 {
		t.Fatalf("shot count = %d", view.ShotCount)
	}
}

func TestApplyCostumeChangeSavesStoryboard(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if _, err := store.SaveStoryboard(ctx, Storyboard{ID: "board", ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Shots: costumeTestShots()}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	storyboard, plan, err := ApplyCostumeChange(ctx, store, "alice", "p1", 1, EngineXiaoluoV7, CostumeChangeOptions{ShotStart: 1, ShotEnd: 6, Scene: "卧室", From: "张三", To: "张三-睡衣"})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if plan.Matched != 2 {
		t.Fatalf("matched = %d", plan.Matched)
	}
	if storyboard.Shots[0].Characters != "张三-睡衣、小Y" || storyboard.Shots[1].Characters != "张三-睡衣" {
		t.Fatalf("shots = %+v", storyboard.Shots[:2])
	}
	// 学校、客厅、已是目标人物的分镜保持原样。
	if storyboard.Shots[2].Characters != "张三、李四" || storyboard.Shots[5].Characters != "王五、张三" {
		t.Fatalf("未命中分镜被改动：%+v", storyboard.Shots)
	}
	reloaded, err := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Shots[0].Characters != "张三-睡衣、小Y" {
		t.Fatalf("未写回：%q", reloaded.Shots[0].Characters)
	}
}

func TestApplyCostumeChangeWithoutMatchKeepsState(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if _, err := store.SaveStoryboard(ctx, Storyboard{ID: "board", ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Shots: costumeTestShots()}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, _, err := ApplyCostumeChange(ctx, store, "alice", "p1", 1, EngineXiaoluoV7, CostumeChangeOptions{ShotStart: 1, ShotEnd: 2, Scene: "学校", From: "张三", To: "张三-睡衣"})
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("应为无命中错误：%v", err)
	}
	if !strings.Contains(err.Error(), "没有可换装的分镜") {
		t.Fatalf("错误文案 = %v", err)
	}
	reloaded, _ := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	if reloaded.Shots[0].Characters != "张三、小Y" {
		t.Fatalf("无命中时不应写库：%q", reloaded.Shots[0].Characters)
	}
}
