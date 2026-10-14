package batchfactoryv11

import (
	"encoding/json"
	"strings"
	"testing"
)

func rawSettingV13(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAudioMinimumVideoCountV13(t *testing.T) {
	if got := audioMinimumVideoCount(63, 10); got != 7 {
		t.Fatalf("audioMinimumVideoCount(63,10)=%d, want 7", got)
	}
	if got := audioMinimumVideoCount(48, 15); got != 4 {
		t.Fatalf("audioMinimumVideoCount(48,15)=%d, want 4", got)
	}
}

func TestAudioPlanningRequiredForNewFlowAndLegacyCompatibility(t *testing.T) {
	newFlow := SettingsPatch{
		"storyTimelineVersion": rawSettingV13(t, sdOriginalStoryVersion),
		"audioPlanningEnabled": rawSettingV13(t, false),
	}
	if !audioPlanningRequired(newFlow) {
		t.Fatal("new story-timeline flow must always require measured audio planning")
	}
	legacyExplicitOff := SettingsPatch{"audioPlanningEnabled": rawSettingV13(t, false)}
	if audioPlanningRequired(legacyExplicitOff) {
		t.Fatal("historical tasks with an explicit false must keep their legacy behavior")
	}
	measuredLegacy := SettingsPatch{"audioDurationSeconds": rawSettingV13(t, 12.4)}
	if !audioPlanningRequired(measuredLegacy) {
		t.Fatal("a legacy snapshot with measured audio must use the measured timeline")
	}
	if audioPlanningRequired(SettingsPatch{}) {
		t.Fatal("unmarked historical snapshots must keep their legacy behavior")
	}
}

func TestSnapshotRejectsFixedOpeningWithAudioPlanningV13(t *testing.T) {
	batch := Batch{SettingsState: SettingsState{Patch: SettingsPatch{
		"storyboardDurationLimit": rawSettingV13(t, 10),
		"fixedSingleVideo":        rawSettingV13(t, true),
		"audioPlanningEnabled":    rawSettingV13(t, true),
		"audioDurationSeconds":    rawSettingV13(t, 62.53),
	}}}
	_, err := snapshotForBook(batch, Book{})
	if err == nil || !strings.Contains(err.Error(), "固定开头") {
		t.Fatalf("expected fixed/audio conflict, got %v", err)
	}
}

func TestDirectorContractExplainsAudioPlanV13(t *testing.T) {
	snapshot := DirectorSnapshot{
		Effective: SettingsPatch{"audioPlanningEnabled": rawSettingV13(t, true)},
		Mode:      "original", MaxVideoDuration: 10,
		AudioDurationSeconds: 62.53, AudioTargetSeconds: 63,
		AspectRatio: "9:16",
	}
	contract, err := BuildDirectorContract(Book{ID: "b1", Title: "T", SourceText: "正文"}, HookRevision{}, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"至少规划 7 个 VIDEO", "不要机械平均", "minimum_video_count"} {
		if !strings.Contains(contract.SystemPrompt+contract.UserPrompt, want) {
			t.Fatalf("contract missing %q", want)
		}
	}
}

func TestNormalizeDirectorOutputRequiresEnoughVideosV13(t *testing.T) {
	storyboard := make([]map[string]any, 0, 6)
	for i := 0; i < 6; i++ {
		storyboard = append(storyboard, map[string]any{
			"id": i + 1, "duration_sec": 10, "characters": []any{}, "props": []any{}, "scene": "", "prefix_key": "", "video_desc": "动作",
			"shots": []map[string]any{{"start_sec": 0, "end_sec": 10, "shot_type": "中景", "camera": "固定", "description": "人物动作"}},
		})
	}
	raw, err := json.Marshal(map[string]any{"characters": []any{}, "scenes": []any{}, "props": []any{}, "storyboard": storyboard})
	if err != nil {
		t.Fatal(err)
	}
	_, err = NormalizeDirectorOutput(raw, DirectorSettings{MaxVideoDuration: 10, AudioTargetSeconds: 63, AspectRatio: "9:16"})
	if err == nil || !strings.Contains(err.Error(), "至少 7 个 VIDEO") {
		t.Fatalf("expected minimum video count error, got %v", err)
	}
}

func TestNormalizeDirectorOutputAcceptsSevenVideosFor63SecondsV13(t *testing.T) {
	storyboard := make([]map[string]any, 0, 7)
	for i := 0; i < 7; i++ {
		storyboard = append(storyboard, map[string]any{
			"id": i + 1, "duration_sec": 9, "characters": []any{}, "props": []any{}, "scene": "", "prefix_key": "", "video_desc": "动作",
			"shots": []map[string]any{{"start_sec": 0, "end_sec": 9, "shot_type": "中景", "camera": "固定", "description": "人物动作"}},
		})
	}
	raw, err := json.Marshal(map[string]any{"characters": []any{}, "scenes": []any{}, "props": []any{}, "storyboard": storyboard})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeDirectorOutput(raw, DirectorSettings{MaxVideoDuration: 10, AudioTargetSeconds: 63, AspectRatio: "9:16"}); err != nil {
		t.Fatalf("expected valid 63-second audio plan, got %v", err)
	}
}
