package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/videogen"
)

type videoServiceFake struct {
	owner string
	input videogen.GenerateInput
	asset media.Asset
	err   error
}

func (f *videoServiceFake) Generate(_ context.Context, owner string, input videogen.GenerateInput) (media.Asset, error) {
	f.owner = owner
	f.input = input
	if f.asset.ID == "" {
		f.asset = media.Asset{ID: "asset_video_1", MediaType: media.TypeVideo}
	}
	return f.asset, f.err
}

func TestVideoGenerateToolExecutesUnifiedVideoService(t *testing.T) {
	service := &videoServiceFake{}
	handler := VideoGenerateTool{Service: service}
	arguments, _ := json.Marshal(map[string]any{
		"model": "yd2-mini-video",
		"prompt": "人物缓慢抬头，镜头向前推进",
		"duration": 10,
		"aspect_ratio": "9:16",
		"resolution": "720p",
		"reference_media_asset_ids": []string{"asset_img_1"},
	})
	result, err := handler.Execute(context.Background(), ToolExecutionRequest{
		Owner: "owner-a", ThreadID: "thread-a", ToolCallID: "tool-a", ToolName: ToolVideoGenerate, Arguments: arguments,
	})
	if err != nil { t.Fatal(err) }
	if service.owner != "owner-a" { t.Fatalf("owner=%q", service.owner) }
	if service.input.ToolCallID != "tool-a" || service.input.Model != "yd2-mini-video" || service.input.Duration != 10 {
		t.Fatalf("input=%#v", service.input)
	}
	if len(service.input.ReferenceMediaAssetIDs) != 1 || service.input.ReferenceMediaAssetIDs[0] != "asset_img_1" {
		t.Fatalf("references=%#v", service.input.ReferenceMediaAssetIDs)
	}
	if len(result.MediaAssetIDs) != 1 || result.MediaAssetIDs[0] != "asset_video_1" {
		t.Fatalf("result=%#v", result)
	}
	if result.Content == "" { t.Fatal("expected user-facing completion content") }
}

func TestVideoGenerateToolRejectsMissingRequiredArguments(t *testing.T) {
	handler := VideoGenerateTool{Service: &videoServiceFake{}}
	_, err := handler.Execute(context.Background(), ToolExecutionRequest{
		Owner: "owner-a", ToolCallID: "tool-a", ToolName: ToolVideoGenerate, Arguments: json.RawMessage(`{"model":"yd2-mini-video"}`),
	})
	if err == nil { t.Fatal("expected validation error") }
}

func TestVideoGenerateToolRejectsNonAssetReference(t *testing.T) {
	handler := VideoGenerateTool{Service: &videoServiceFake{}}
	_, err := handler.Execute(context.Background(), ToolExecutionRequest{
		Owner: "owner-a", ToolCallID: "tool-a", ToolName: ToolVideoGenerate,
		Arguments: json.RawMessage(`{"model":"yd2-mini-video","prompt":"x","duration":5,"reference_media_asset_ids":["https://example.com/a.png"]}`),
	})
	if err == nil { t.Fatal("expected media_asset_id validation error") }
}
