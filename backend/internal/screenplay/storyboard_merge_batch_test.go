package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func mergeTestShots(durations ...float64) []StoryboardShot {
	shots := make([]StoryboardShot, 0, len(durations))
	for index, duration := range durations {
		shots = append(shots, StoryboardShot{
			ID:         "shot-" + string(rune('a'+index)),
			Duration:   duration,
			VideoDesc:  "镜头" + string(rune('1'+index)),
			Line:       "台词" + string(rune('1'+index)),
			Characters: []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛"}[index%8],
			Scene:      "场景" + string(rune('1'+index)),
			Props:      "",
		})
	}
	return shots
}

// 洛水更新日志原文举例：推荐时长 15s、5s、9s、15s，选 30s 时前三个（共 29s）合并，第四个单独保留。
func TestPlanStoryboardMergeMatchesLuoshuiExample(t *testing.T) {
	shots := mergeTestShots(15, 5, 9, 15)
	plan, err := PlanStoryboardMerge(shots, 30)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if plan.BeforeCount != 4 || plan.AfterCount != 2 || plan.RemovedCount != 2 {
		t.Fatalf("counts before=%d after=%d removed=%d", plan.BeforeCount, plan.AfterCount, plan.RemovedCount)
	}
	if plan.PendingGroups != 1 || plan.PendingShots != 3 {
		t.Fatalf("pending groups=%d shots=%d", plan.PendingGroups, plan.PendingShots)
	}
	group := plan.Groups[0]
	if group.KeepShotID != "shot-a" || len(group.ShotIDs) != 3 || group.Seconds != 29 {
		t.Fatalf("group = %+v", group)
	}
	if group.ShotIDs[0] != "shot-a" || group.ShotIDs[2] != "shot-c" {
		t.Fatalf("group order = %v", group.ShotIDs)
	}
}

func TestPlanStoryboardMergeRespectsMaxSeconds(t *testing.T) {
	plan, err := PlanStoryboardMerge(mergeTestShots(15, 5, 9, 15), 15)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	// 15 单独；5+9=14 合并；15 单独。
	if plan.PendingGroups != 1 || plan.AfterCount != 3 {
		t.Fatalf("groups=%d after=%d", plan.PendingGroups, plan.AfterCount)
	}
	if plan.Groups[0].KeepShotID != "shot-b" || plan.Groups[0].Seconds != 14 {
		t.Fatalf("group = %+v", plan.Groups[0])
	}
}

func TestPlanStoryboardMergeRejectsAllUntimed(t *testing.T) {
	_, err := PlanStoryboardMerge(mergeTestShots(0, 0, 0), 30)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
}

func TestPlanStoryboardMergeStopsAtUntimedShot(t *testing.T) {
	plan, err := PlanStoryboardMerge(mergeTestShots(10, 0, 10), 30)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	// 中间没有推荐时长的分镜是断点，不能跨过它把首尾合并。
	if plan.PendingGroups != 0 || plan.UntimedCount != 1 {
		t.Fatalf("groups=%d untimed=%d", plan.PendingGroups, plan.UntimedCount)
	}
}

func TestPlanStoryboardMergeNoGroupWhenNothingFits(t *testing.T) {
	plan, err := PlanStoryboardMerge(mergeTestShots(20, 20), 30)
	if err != nil {
		t.Fatalf("plan error: %v", err)
	}
	if plan.PendingGroups != 0 || plan.AfterCount != 2 {
		t.Fatalf("groups=%d after=%d", plan.PendingGroups, plan.AfterCount)
	}
}

func TestMergeStoryboardShotsByDurationMergesGroup(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	storyboard := Storyboard{ID: "sb-1", ProjectID: "comic-1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Shots: mergeTestShots(15, 5, 9, 15)}
	if _, err := store.SaveStoryboard(ctx, storyboard); err != nil {
		t.Fatalf("save storyboard: %v", err)
	}

	merged, plan, err := MergeStoryboardShotsByDuration(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, 30)
	if err != nil {
		t.Fatalf("merge error: %v", err)
	}
	if len(merged.Shots) != 2 {
		t.Fatalf("shots = %d, want 2", len(merged.Shots))
	}
	first := merged.Shots[0]
	if first.Duration != 29 {
		t.Fatalf("merged duration = %v, want 29", first.Duration)
	}
	if first.VideoDesc != "镜头1\n镜头2\n镜头3" {
		t.Fatalf("merged video desc = %q", first.VideoDesc)
	}
	if first.Line != "台词1\n台词2\n台词3" {
		t.Fatalf("merged line = %q", first.Line)
	}
	if first.Characters != "甲；乙；丙" {
		t.Fatalf("merged characters = %q", first.Characters)
	}
	if first.Scene != "场景1；场景2；场景3" {
		t.Fatalf("merged scene = %q", first.Scene)
	}
	if merged.Shots[0].Seq != 1 || merged.Shots[1].Seq != 2 || merged.ShotCount != 2 {
		t.Fatalf("seq/count wrong: %+v", merged.Shots)
	}
	var protocol map[string]any
	if err := json.Unmarshal(first.ProtocolFields, &protocol); err != nil {
		t.Fatalf("protocol fields: %v", err)
	}
	if protocol["merged_count"] != float64(3) {
		t.Fatalf("merged_count = %v", protocol["merged_count"])
	}
	if plan.RemovedCount != 2 {
		t.Fatalf("plan removed = %d", plan.RemovedCount)
	}

	reloaded, err := store.GetStoryboard(ctx, "alice", "comic-1", 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(reloaded.Shots) != 2 {
		t.Fatalf("reloaded shots = %d", len(reloaded.Shots))
	}
}

func TestMergeStoryboardShotsByDurationWithoutGroupsFails(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	if _, err := store.SaveStoryboard(ctx, Storyboard{ID: "sb-1", ProjectID: "comic-1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Shots: mergeTestShots(20, 20)}); err != nil {
		t.Fatalf("save storyboard: %v", err)
	}
	_, _, err := MergeStoryboardShotsByDuration(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, 30)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("err = %v, want ErrInvalidState", err)
	}
	stored, err := store.GetStoryboard(ctx, "alice", "comic-1", 1, EngineXiaoluoV7)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(stored.Shots) != 2 {
		t.Fatalf("shots changed on failure: %d", len(stored.Shots))
	}
}

func TestJoinStoryboardNamesDeduplicates(t *testing.T) {
	got := joinStoryboardNames([]string{"甲,乙", "乙、丙", "丙；甲"})
	if got != "甲；乙；丙" {
		t.Fatalf("names = %q", got)
	}
}
