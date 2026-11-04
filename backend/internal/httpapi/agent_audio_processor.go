package httpapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/ffmpegbin"
)

type FFmpegAgentAudioProcessor struct {
	Binary string
	Exec   func(context.Context, string, ...string) error
}

func (p *FFmpegAgentAudioProcessor) Process(ctx context.Context, input agentworkspace.AudioProcessInput) ([]agentworkspace.AudioOutput, error) {
	if len(input.SourceBytes) == 0 || !input.Operation.IsProcessing() {
		return nil, agentworkspace.ErrInvalidAudioJob
	}
	workDir, err := os.MkdirTemp("", "agent-audio-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(input.Source.Name), "."))
	if extension == "" || len(extension) > 8 {
		extension = "mp3"
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
	outputPaths := make([]string, 0)
	switch input.Operation {
	case agentworkspace.AudioTrim:
		start := audioSettingNumber(input.Settings, "startMs") / 1000
		end := audioSettingNumber(input.Settings, "endMs") / 1000
		if start < 0 || end <= start {
			return nil, agentworkspace.ErrInvalidAudioJob
		}
		output := filepath.Join(workDir, "trimmed.mp3")
		err = run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-ss", formatAudioSeconds(start), "-to", formatAudioSeconds(end), "-i", sourcePath, "-vn", "-c:a", "libmp3lame", "-b:a", "192k", output)
		outputPaths = append(outputPaths, output)
	case agentworkspace.AudioSpeed:
		speed := audioSettingNumber(input.Settings, "speed")
		if speed < 0.5 || speed > 2 {
			return nil, agentworkspace.ErrInvalidAudioJob
		}
		output := filepath.Join(workDir, "speed.mp3")
		err = run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-i", sourcePath, "-filter:a", "atempo="+strconv.FormatFloat(speed, 'f', -1, 64), "-vn", "-c:a", "libmp3lame", "-b:a", "192k", output)
		outputPaths = append(outputPaths, output)
	case agentworkspace.AudioSmartSplit:
		segmentSeconds := audioSettingNumber(input.Settings, "segmentSeconds")
		if segmentSeconds == 0 {
			segmentSeconds = 15
		}
		if segmentSeconds < 2 || segmentSeconds > 600 {
			return nil, agentworkspace.ErrInvalidAudioJob
		}
		pattern := filepath.Join(workDir, "segment-%03d.mp3")
		err = run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-i", sourcePath, "-vn", "-c:a", "libmp3lame", "-b:a", "192k", "-f", "segment", "-segment_time", strconv.FormatFloat(segmentSeconds, 'f', -1, 64), "-reset_timestamps", "1", pattern)
		if err == nil {
			outputPaths, err = filepath.Glob(filepath.Join(workDir, "segment-*.mp3"))
			sort.Strings(outputPaths)
		}
	case agentworkspace.AudioCustomSplit:
		ranges, ok := input.Settings["ranges"].([]any)
		if !ok || len(ranges) == 0 || len(ranges) > 100 {
			return nil, agentworkspace.ErrInvalidAudioJob
		}
		for index, raw := range ranges {
			value, ok := raw.(map[string]any)
			if !ok {
				return nil, agentworkspace.ErrInvalidAudioJob
			}
			start := audioSettingNumber(value, "startMs") / 1000
			end := audioSettingNumber(value, "endMs") / 1000
			if start < 0 || end <= start {
				return nil, agentworkspace.ErrInvalidAudioJob
			}
			output := filepath.Join(workDir, fmt.Sprintf("custom-%03d.mp3", index+1))
			if err = run(ctx, binary, "-hide_banner", "-loglevel", "error", "-y", "-ss", formatAudioSeconds(start), "-to", formatAudioSeconds(end), "-i", sourcePath, "-vn", "-c:a", "libmp3lame", "-b:a", "192k", output); err != nil {
				break
			}
			outputPaths = append(outputPaths, output)
		}
	default:
		return nil, agentworkspace.ErrInvalidAudioJob
	}
	if err != nil {
		return nil, err
	}
	if len(outputPaths) == 0 {
		return nil, fmt.Errorf("ffmpeg did not produce audio output")
	}
	outputs := make([]agentworkspace.AudioOutput, 0, len(outputPaths))
	for _, outputPath := range outputPaths {
		contents, readErr := os.ReadFile(outputPath)
		if readErr != nil || len(contents) == 0 {
			return nil, fmt.Errorf("audio output is empty")
		}
		outputs = append(outputs, agentworkspace.AudioOutput{Name: filepath.Base(outputPath), ContentType: "audio/mpeg", Extension: "mp3", Bytes: contents})
	}
	return outputs, nil
}

func audioSettingNumber(settings map[string]any, key string) float64 {
	value := settings[key]
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
		return parsed
	}
}

func formatAudioSeconds(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}

func runAgentFFmpeg(ctx context.Context, binary string, args ...string) error {
	return ffmpegbin.Run(ctx, binary, args...)
}
