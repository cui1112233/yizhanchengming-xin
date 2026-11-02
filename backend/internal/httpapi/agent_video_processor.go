package httpapi

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"qiantie/backend/internal/agentworkspace"
)

type AgentVideoProcessInput struct {
	Operation   agentworkspace.VideoToolOperation
	Source      agentworkspace.Asset
	SourceBytes []byte
	Settings    map[string]any
	WorkDir     string // Tests may provide a fixture directory; production uses an isolated temporary directory.
}

type AgentVideoProcessOutput struct {
	Kind        string
	Name        string
	ContentType string
	Extension   string
	Bytes       []byte
}

type AgentVideoProcessor interface {
	Process(context.Context, AgentVideoProcessInput) ([]AgentVideoProcessOutput, error)
}

type FFmpegAgentVideoProcessor struct {
	Binary string
	Exec   func(context.Context, string, ...string) error
}

func (p *FFmpegAgentVideoProcessor) Process(ctx context.Context, input AgentVideoProcessInput) ([]AgentVideoProcessOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if input.Source.Kind != "video" || len(input.SourceBytes) == 0 || agentworkspace.ValidateVideoToolSettings(input.Operation, input.Settings) != nil {
		return nil, agentworkspace.ErrInvalidVideoTool
	}
	workDir := strings.TrimSpace(input.WorkDir)
	removeWorkDir := false
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "agent-video-*")
		if err != nil {
			return nil, err
		}
		removeWorkDir = true
	}
	if removeWorkDir {
		defer os.RemoveAll(workDir)
	}
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(input.Source.Name)), ".")
	if extension == "" || len(extension) > 8 {
		extension = "mp4"
	}
	sourcePath := filepath.Join(workDir, "source."+extension)
	if err := os.WriteFile(sourcePath, input.SourceBytes, 0o600); err != nil {
		return nil, err
	}
	run := p.Exec
	if run == nil {
		run = runAgentFFmpeg
	}
	binary := strings.TrimSpace(p.Binary)
	if binary == "" {
		binary = "ffmpeg"
	}
	output := func(kind, name, contentType, extension, file string) (AgentVideoProcessOutput, error) {
		contents, err := os.ReadFile(file)
		if err != nil || len(contents) == 0 {
			return AgentVideoProcessOutput{}, fmt.Errorf("视频处理结果为空")
		}
		return AgentVideoProcessOutput{Kind: kind, Name: name, ContentType: contentType, Extension: extension, Bytes: contents}, nil
	}
	switch input.Operation {
	case agentworkspace.VideoTrim:
		resultPath := filepath.Join(workDir, "trimmed.mp4")
		if err := run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-ss", videoToolSeconds(input.Settings, "startMs"), "-to", videoToolSeconds(input.Settings, "endMs"), "-i", sourcePath, "-c:v", "libx264", "-c:a", "aac", resultPath); err != nil {
			return nil, err
		}
		item, err := output("video", "视频剪辑.mp4", "video/mp4", "mp4", resultPath)
		return []AgentVideoProcessOutput{item}, err
	case agentworkspace.VideoCrop:
		resultPath := filepath.Join(workDir, "cropped.mp4")
		filter := fmt.Sprintf("crop=%d:%d:%d:%d", videoToolInt(input.Settings, "width"), videoToolInt(input.Settings, "height"), videoToolInt(input.Settings, "x"), videoToolInt(input.Settings, "y"))
		if err := run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-i", sourcePath, "-vf", filter, "-c:v", "libx264", "-c:a", "aac", resultPath); err != nil {
			return nil, err
		}
		item, err := output("video", "视频裁剪.mp4", "video/mp4", "mp4", resultPath)
		return []AgentVideoProcessOutput{item}, err
	case agentworkspace.VideoExtractFrames:
		timestamps := input.Settings["timestampsMs"].([]any)
		outputs := make([]AgentVideoProcessOutput, 0, len(timestamps))
		for index, raw := range timestamps {
			resultPath := filepath.Join(workDir, fmt.Sprintf("frame-%03d.jpg", index+1))
			seconds := strconv.FormatFloat(raw.(float64)/1000, 'f', 3, 64)
			if err := run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-ss", seconds, "-i", sourcePath, "-frames:v", "1", "-q:v", "2", resultPath); err != nil {
				return nil, err
			}
			item, err := output("image", fmt.Sprintf("视频帧-%02d.jpg", index+1), "image/jpeg", "jpg", resultPath)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, item)
		}
		return outputs, nil
	case agentworkspace.VideoSeparateAudio:
		resultPath := filepath.Join(workDir, "audio.m4a")
		if err := run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-i", sourcePath, "-vn", "-c:a", "aac", resultPath); err != nil {
			return nil, err
		}
		item, err := output("audio", "视频原声.m4a", "audio/mp4", "m4a", resultPath)
		return []AgentVideoProcessOutput{item}, err
	default:
		return nil, agentworkspace.ErrInvalidVideoTool
	}
}

func videoToolInt(settings map[string]any, key string) int {
	switch value := settings[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	case int64:
		return int(value)
	default:
		return 0
	}
}

func videoToolSeconds(settings map[string]any, key string) string {
	return strconv.FormatFloat(float64(videoToolInt(settings, key))/1000, 'f', 3, 64)
}

func readAgentVideoSource(ctx context.Context, owner, projectID, assetID string, assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore) (agentworkspace.Asset, []byte, error) {
	asset, err := assets.Asset(ctx, owner, projectID, strings.TrimSpace(assetID))
	if err != nil || asset.Kind != "video" {
		return agentworkspace.Asset{}, nil, agentworkspace.ErrInvalidVideoTool
	}
	body, err := objects.Open(ctx, asset)
	if err != nil {
		return agentworkspace.Asset{}, nil, err
	}
	defer body.Close()
	contents, err := io.ReadAll(io.LimitReader(body, agentAssetUploadMaxBytes+1))
	if err != nil || len(contents) == 0 || len(contents) > agentAssetUploadMaxBytes {
		return agentworkspace.Asset{}, nil, fmt.Errorf("读取源视频失败")
	}
	return asset, contents, nil
}
