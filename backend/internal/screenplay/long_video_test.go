package screenplay

import "testing"

func longVideoShots(secs ...float64) []StoryboardShot {
	shots := make([]StoryboardShot, 0, len(secs))
	for i, sec := range secs {
		shots = append(shots, StoryboardShot{ID: "shot-" + string(rune('a'+i)), Seq: i + 1, Duration: sec, VideoDesc: "镜头提示词"})
	}
	return shots
}

func TestPlanLongVideoSectionsMatchesLuoshuiExample(t *testing.T) {
	sections := PlanLongVideoSections(longVideoShots(15, 5, 9, 15), 300, 30)
	if len(sections) != 2 {
		t.Fatalf("sections=%d want 2", len(sections))
	}
	if sections[0].TotalSeconds != 29 || len(sections[0].ShotIDs) != 3 {
		t.Fatalf("first=%+v", sections[0])
	}
	if sections[1].TotalSeconds != 15 || len(sections[1].ShotIDs) != 1 {
		t.Fatalf("second=%+v", sections[1])
	}
	if sections[0].Label != "超长视频分镜*1" || sections[1].Label != "超长视频分镜*2" {
		t.Fatalf("labels=%q %q", sections[0].Label, sections[1].Label)
	}
}

func TestPlanLongVideoSectionsNeverExceedsTarget(t *testing.T) {
	shots := longVideoShots(15, 15, 15, 15, 15, 15, 15)
	sections := PlanLongVideoSections(shots, 100, 30)
	if total := LongVideoPlannedSeconds(sections); total > 100 {
		t.Fatalf("total=%v exceeded target", total)
	}
	if got := LongVideoPlannedSeconds(sections); got != 90 {
		t.Fatalf("total=%v want 90", got)
	}
}

func TestPlanLongVideoSectionsSingleShotOverCapIsClipped(t *testing.T) {
	sections := PlanLongVideoSections(longVideoShots(45, 10), 300, 30)
	if len(sections) != 2 || sections[0].TotalSeconds != 30 || len(sections[0].ShotIDs) != 1 {
		t.Fatalf("sections=%+v", sections)
	}
}

func TestPlanLongVideoSectionsSkipsExistingSections(t *testing.T) {
	shots := []StoryboardShot{
		{ID: "a", Seq: 1, Duration: 10, VideoDesc: "超长视频分镜*1\n已合并的提示词"},
		{ID: "b", Seq: 2, Duration: 10, VideoDesc: "新的镜头"},
	}
	sections := PlanLongVideoSections(shots, 300, 30)
	if len(sections) != 1 || len(sections[0].ShotIDs) != 1 || sections[0].ShotIDs[0] != "b" {
		t.Fatalf("sections=%+v", sections)
	}
}

func TestPlanLongVideoSectionsMergesNamesAndPrompt(t *testing.T) {
	shots := []StoryboardShot{
		{ID: "a", Seq: 1, Duration: 10, VideoDesc: "第一镜", Characters: "小雨、阿岚", Scene: "天台", Props: "伞"},
		{ID: "b", Seq: 2, Duration: 10, VideoDesc: "第二镜", Characters: "阿岚，老陈", Scene: "天台、走廊", Props: "伞；灯"},
	}
	sections := PlanLongVideoSections(shots, 300, 30)
	if len(sections) != 1 {
		t.Fatalf("sections=%+v", sections)
	}
	first := sections[0]
	if first.Prompt != "第一镜\n第二镜" {
		t.Fatalf("prompt=%q", first.Prompt)
	}
	if first.Characters != "小雨、阿岚、老陈" {
		t.Fatalf("characters=%q", first.Characters)
	}
	if first.Scene != "天台、走廊" {
		t.Fatalf("scene=%q", first.Scene)
	}
	if first.Props != "伞、灯" {
		t.Fatalf("props=%q", first.Props)
	}
}

func TestNormalizeLongVideoTargetOnlyAllowsFixedOptions(t *testing.T) {
	if NormalizeLongVideoTarget(200) != 200 || NormalizeLongVideoTarget(123) != 300 || NormalizeLongVideoTarget(0) != 300 {
		t.Fatal("target normalization is not locked to 100/200/250/300")
	}
	if NormalizeLongVideoCap(60) != 30 || NormalizeLongVideoCap(0) != 30 {
		t.Fatal("cap must never exceed the single-generation limit")
	}
}
