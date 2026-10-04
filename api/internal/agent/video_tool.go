package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/videogen"
)

type VideoGenerationService interface {
	Generate(context.Context, string, videogen.GenerateInput) (media.Asset, error)
}

type VideoGenerateTool struct {
	Service VideoGenerationService
}

type videoGenerateArguments struct {
	Model                  string   `json:"model"`
	FallbackProvider       string   `json:"fallback_provider,omitempty"`
	Prompt                 string   `json:"prompt"`
	Duration               int      `json:"duration"`
	AspectRatio            string   `json:"aspect_ratio,omitempty"`
	Resolution             string   `json:"resolution,omitempty"`
	ReferenceMediaAssetIDs []string `json:"reference_media_asset_ids,omitempty"`
}

func (t VideoGenerateTool) Execute(ctx context.Context, request ToolExecutionRequest) (ToolExecutionResult, error) {
	if t.Service == nil {
		return ToolExecutionResult{}, errors.New("video generation tool unavailable")
	}
	var args videoGenerateArguments
	if err := json.Unmarshal(request.Arguments, &args); err != nil {
		return ToolExecutionResult{}, fmt.Errorf("invalid video generation arguments: %w", err)
	}
	args.Model = strings.TrimSpace(args.Model)
	args.FallbackProvider = strings.TrimSpace(args.FallbackProvider)
	args.Prompt = strings.TrimSpace(args.Prompt)
	args.AspectRatio = strings.TrimSpace(args.AspectRatio)
	args.Resolution = strings.TrimSpace(args.Resolution)
	if args.Model == "" || args.Prompt == "" || args.Duration <= 0 {
		return ToolExecutionResult{}, errors.New("video generation requires model, prompt and positive duration")
	}
	if args.AspectRatio == "" { args.AspectRatio = "9:16" }
	if args.Resolution == "" { args.Resolution = "720p" }

	references := make([]string, 0, len(args.ReferenceMediaAssetIDs))
	seen := make(map[string]struct{}, len(args.ReferenceMediaAssetIDs))
	for _, raw := range args.ReferenceMediaAssetIDs {
		id := strings.TrimSpace(raw)
		if id == "" { continue }
		if !strings.HasPrefix(id, "asset_") {
			return ToolExecutionResult{}, fmt.Errorf("reference %q must be a media_asset_id", id)
		}
		if _, ok := seen[id]; ok { continue }
		seen[id] = struct{}{}
		references = append(references, id)
	}

	asset, err := t.Service.Generate(ctx, request.Owner, videogen.GenerateInput{
		ToolCallID: request.ToolCallID,
		Model: args.Model,
		FallbackProvider: args.FallbackProvider,
		Prompt: args.Prompt,
		Duration: args.Duration,
		AspectRatio: args.AspectRatio,
		Resolution: args.Resolution,
		ReferenceMediaAssetIDs: references,
	})
	if err != nil {
		return ToolExecutionResult{}, err
	}
	if asset.MediaType != media.TypeVideo || strings.TrimSpace(asset.ID) == "" {
		return ToolExecutionResult{}, errors.New("video generation did not return a valid video media asset")
	}
	result, _ := json.Marshal(map[string]any{
		"media_asset_id": asset.ID,
		"media_type": asset.MediaType,
		"model": args.Model,
	})
	return ToolExecutionResult{
		Content: "视频生成完成，已保存到 TOS。",
		MediaAssetIDs: []string{asset.ID},
		Result: result,
	}, nil
}
