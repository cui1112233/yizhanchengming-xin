package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/agentworkspace"
)

func TestFFmpegAgentVideoProcessorBuildsDeterministicCommands(t *testing.T) {
	workDir := t.TempDir()
	processor := &FFmpegAgentVideoProcessor{Exec: func(_ context.Context, _ string, args ...string) error {
		output := args[len(args)-1]
		if filepath.Ext(output) == ".jpg" {
			return os.WriteFile(output, []byte("frame"), 0o600)
		}
		return os.WriteFile(output, []byte("media"), 0o600)
	}}
	asset := agentworkspace.Asset{ID: "source-video", ProjectID: "project-1", Kind: "video", Name: "source.mp4", ContentType: "video/mp4"}
	for _, scenario := range []struct {
		name      string
		operation agentworkspace.VideoToolOperation
		settings  map[string]any
		kind      string
		count     int
	}{
		{name: "trim", operation: agentworkspace.VideoTrim, settings: map[string]any{"startMs": 0, "endMs": 1000}, kind: "video", count: 1},
		{name: "crop", operation: agentworkspace.VideoCrop, settings: map[string]any{"x": 0, "y": 0, "width": 320, "height": 180}, kind: "video", count: 1},
		{name: "frames", operation: agentworkspace.VideoExtractFrames, settings: map[string]any{"timestampsMs": []any{float64(0), float64(500)}}, kind: "image", count: 2},
		{name: "audio", operation: agentworkspace.VideoSeparateAudio, settings: map[string]any{}, kind: "audio", count: 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			outputs, err := processor.Process(context.Background(), AgentVideoProcessInput{Operation: scenario.operation, Source: asset, SourceBytes: []byte("source"), Settings: scenario.settings, WorkDir: workDir})
			if err != nil || len(outputs) != scenario.count {
				t.Fatalf("outputs=%d err=%v", len(outputs), err)
			}
			for _, output := range outputs {
				if output.Kind != scenario.kind || len(output.Bytes) == 0 {
					t.Fatalf("output=%+v", output)
				}
			}
		})
	}
}
