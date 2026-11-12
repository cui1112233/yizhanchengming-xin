package prompt

import (
	"strings"
	"testing"

	"qiantie/backend/internal/aicut"
)

func TestAICutEditPlanContractContainsStrictWhitelistAndSafeSnapshot(t *testing.T) {
	timeline := aicut.Timeline{Revision: 3, Script: "第一场", Tracks: []aicut.Track{{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "clip-1", AssetID: "asset-1", Kind: aicut.ClipVideo, StartMS: 0, DurationMS: 2000, SourceDurationMS: 5000}}}}}
	request := aicut.EditPlanRequest{SelectedAssets: []aicut.EditPlanAsset{{ID: "asset-1", Name: "真实视频名称.mp4", Kind: "video", DurationMS: 4567}}, AssetIDs: []string{"asset-1"}, Goal: "剪成节奏紧凑的短片", TargetDurationMS: 1500, AspectRatio: "9:16", IncludeCaptions: true}
	contract, err := AICutEditPlanContract(timeline, request, []string{"fade", "flash-black"}, []string{"warm", "bright"}, []string{"title-basic"})
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"place_asset", "trim_clip", "remove_range", "move_clip", "add_caption", "add_title", "set_transition", "set_filter", "add_effect_layer", "add_bgm", "set_volume_envelope", "set_project_format"} {
		if !strings.Contains(contract.SystemPrompt, operation) {
			t.Fatalf("system prompt missing %s", operation)
		}
	}
	for _, required := range []string{"只输出 JSON", "不得输出命令", "不得输出文件路径", "不得输出 SQL", "不得输出 URL"} {
		if !strings.Contains(contract.SystemPrompt, required) {
			t.Fatalf("system prompt missing %q", required)
		}
	}
	for _, forbidden := range []string{"password", "mysql://", "secret", "access_key"} {
		if strings.Contains(strings.ToLower(contract.SystemPrompt+contract.UserPrompt), forbidden) {
			t.Fatalf("prompt leaked forbidden text %q", forbidden)
		}
	}
	for _, expected := range []string{"真实视频名称.mp4", "4567", "clip-1", "asset-1", "video-1", "第一场", "flash-black", "title-basic"} {
		if !strings.Contains(contract.UserPrompt, expected) {
			t.Fatalf("user prompt missing %q", expected)
		}
	}
}

func TestAICutEditPlanContractRejectsEmptyAssets(t *testing.T) {
	_, err := AICutEditPlanContract(aicut.Timeline{Revision: 1, Tracks: []aicut.Track{}}, aicut.EditPlanRequest{Goal: "自动剪辑", TargetDurationMS: 1000, AspectRatio: "9:16"}, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "素材") {
		t.Fatalf("err=%v", err)
	}
}
