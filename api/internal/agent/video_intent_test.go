package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestResponderCreatesVideoToolFromReferencedAsset(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), ResponseContext{
		Input: "用这张图生成10秒视频，人物慢慢抬头，镜头推进",
		MediaAssetIDs: []string{"asset_img_1"},
	})
	if err != nil { t.Fatal(err) }
	if response.Tool == nil || response.Tool.Name != ToolVideoGenerate { t.Fatalf("response=%#v", response) }
	var args videoGenerateArguments
	if err := json.Unmarshal(response.Tool.Arguments, &args); err != nil { t.Fatal(err) }
	if args.Model != "yd2-mini-video" || args.Duration != 10 { t.Fatalf("args=%#v", args) }
	if len(args.ReferenceMediaAssetIDs) != 1 || args.ReferenceMediaAssetIDs[0] != "asset_img_1" { t.Fatalf("refs=%#v", args.ReferenceMediaAssetIDs) }
}

func TestResponderSelectsExplicitVideoModels(t *testing.T) {
	cases := []struct{ text, model string }{
		{"用H3生成视频", "minimax-h3-video"},
		{"用YFAI Seedance做视频", "seedance-2-0-official"},
		{"用豆包本地执行器生视频", "local-doubao-executor-video"},
	}
	for _, tc := range cases {
		response, err := NewResponder().Respond(context.Background(), ResponseContext{Input: tc.text})
		if err != nil { t.Fatal(err) }
		if response.Tool == nil { t.Fatalf("%q did not produce tool", tc.text) }
		var args videoGenerateArguments
		if err := json.Unmarshal(response.Tool.Arguments, &args); err != nil { t.Fatal(err) }
		if args.Model != tc.model { t.Fatalf("%q model=%q", tc.text, args.Model) }
	}
}
