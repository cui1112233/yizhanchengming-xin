package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/agentworkspace"
)

func TestFFmpegAudioProcessorTrimAndSpeedUseSafeArguments(t *testing.T) {
	for _, test := range []struct {
		name      string
		operation agentworkspace.AudioOperation
		settings  map[string]any
		want      []string
	}{
		{name: "trim", operation: agentworkspace.AudioTrim, settings: map[string]any{"startMs": float64(1000), "endMs": float64(3000)}, want: []string{"-ss", "1.000", "-to", "3.000"}},
		{name: "speed", operation: agentworkspace.AudioSpeed, settings: map[string]any{"speed": float64(1.25)}, want: []string{"-filter:a", "atempo=1.25"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var args []string
			processor := &FFmpegAgentAudioProcessor{Exec: func(_ context.Context, _ string, values ...string) error {
				args = append([]string(nil), values...)
				return os.WriteFile(values[len(values)-1], []byte("processed"), 0o600)
			}}
			outputs, err := processor.Process(context.Background(), agentworkspace.AudioProcessInput{Operation: test.operation, Source: agentworkspace.Asset{Name: "source.mp3", ContentType: "audio/mpeg"}, SourceBytes: []byte("source"), Settings: test.settings})
			if err != nil || len(outputs) != 1 || string(outputs[0].Bytes) != "processed" {
				t.Fatalf("outputs=%+v err=%v", outputs, err)
			}
			joined := strings.Join(args, " ")
			for _, want := range test.want {
				if !strings.Contains(joined, want) {
					t.Fatalf("args=%q missing %q", joined, want)
				}
			}
		})
	}
}

func TestFFmpegAudioProcessorSmartAndCustomSplitReturnMultipleOutputs(t *testing.T) {
	smart := &FFmpegAgentAudioProcessor{Exec: func(_ context.Context, _ string, args ...string) error {
		pattern := args[len(args)-1]
		if err := os.WriteFile(strings.Replace(pattern, "%03d", "000", 1), []byte("one"), 0o600); err != nil { return err }
		return os.WriteFile(strings.Replace(pattern, "%03d", "001", 1), []byte("two"), 0o600)
	}}
	outputs, err := smart.Process(context.Background(), agentworkspace.AudioProcessInput{Operation: agentworkspace.AudioSmartSplit, Source: agentworkspace.Asset{Name: "source.wav", ContentType: "audio/wav"}, SourceBytes: []byte("source"), Settings: map[string]any{"segmentSeconds": float64(12)}})
	if err != nil || len(outputs) != 2 || string(outputs[1].Bytes) != "two" {
		t.Fatalf("smart outputs=%+v err=%v", outputs, err)
	}

	calls := 0
	custom := &FFmpegAgentAudioProcessor{Exec: func(_ context.Context, _ string, args ...string) error {
		calls++
		return os.WriteFile(filepath.Clean(args[len(args)-1]), []byte{byte('0' + calls)}, 0o600)
	}}
	outputs, err = custom.Process(context.Background(), agentworkspace.AudioProcessInput{Operation: agentworkspace.AudioCustomSplit, Source: agentworkspace.Asset{Name: "source.mp3", ContentType: "audio/mpeg"}, SourceBytes: []byte("source"), Settings: map[string]any{"ranges": []any{map[string]any{"startMs": float64(0), "endMs": float64(1000)}, map[string]any{"startMs": float64(1000), "endMs": float64(2500)}}}})
	if err != nil || len(outputs) != 2 || calls != 2 {
		t.Fatalf("custom outputs=%+v calls=%d err=%v", outputs, calls, err)
	}
}
