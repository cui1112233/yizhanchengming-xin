package agentworkspace

import (
	"errors"
	"testing"
)

func TestNormalizeAudioJobRequestRequiresSourceOnlyForProcessing(t *testing.T) {
	generation, err := NormalizeAudioJobRequest(AudioJobRequest{Operation: AudioSpeech, Prompt: "旁白", ModelID: "tts-1"})
	if err != nil || generation.SourceAssetID != "" {
		t.Fatalf("generation=%+v err=%v", generation, err)
	}
	processed, err := NormalizeAudioJobRequest(AudioJobRequest{Operation: AudioTrim, SourceAssetID: "audio-source", Settings: map[string]any{"startMs": float64(1000), "endMs": float64(3000)}})
	if err != nil || processed.SourceAssetID != "audio-source" {
		t.Fatalf("processed=%+v err=%v", processed, err)
	}
	if _, err := NormalizeAudioJobRequest(AudioJobRequest{Operation: AudioTrim}); !errors.Is(err, ErrInvalidAudioJob) {
		t.Fatalf("missing source err=%v", err)
	}
}

func TestNormalizeAudioJobRequestSupportsAllPlannedOperations(t *testing.T) {
	for _, operation := range []AudioOperation{AudioSpeech, AudioMusic, AudioSoundEffect, AudioTrim, AudioSpeed, AudioSmartSplit, AudioCustomSplit} {
		input := AudioJobRequest{Operation: operation, Prompt: "声音", ModelID: "audio-model"}
		if operation.IsProcessing() {
			input.SourceAssetID = "audio-source"
		}
		if _, err := NormalizeAudioJobRequest(input); err != nil {
			t.Fatalf("operation %q should be supported: %v", operation, err)
		}
	}
}
