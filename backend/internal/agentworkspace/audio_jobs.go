package agentworkspace

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidAudioJob = errors.New("invalid agent audio job")

type AudioOperation string

const (
	AudioSpeech      AudioOperation = "speech"
	AudioMusic       AudioOperation = "music"
	AudioSoundEffect AudioOperation = "sound_effect"
	AudioTrim        AudioOperation = "trim"
	AudioSpeed       AudioOperation = "speed"
	AudioSmartSplit  AudioOperation = "smart_split"
	AudioCustomSplit AudioOperation = "custom_split"
)

func (operation AudioOperation) IsGeneration() bool {
	return operation == AudioSpeech || operation == AudioMusic || operation == AudioSoundEffect
}

func (operation AudioOperation) IsProcessing() bool {
	return operation == AudioTrim || operation == AudioSpeed || operation == AudioSmartSplit || operation == AudioCustomSplit
}

type AudioJobRequest struct {
	Operation     AudioOperation
	Prompt        string
	Provider      string
	ModelID       string
	SourceAssetID string
	Settings      map[string]any
	Advanced      map[string]any
}

type AudioGenerateInput struct {
	Operation AudioOperation
	Prompt    string
	Provider  string
	ModelID   string
	Settings  map[string]any
	Advanced  map[string]any
}

type AudioProcessInput struct {
	Operation         AudioOperation
	Source            Asset
	SourceBytes       []byte
	Settings          map[string]any
	Advanced          map[string]any
}

type AudioOutput struct {
	Name        string
	ContentType string
	Extension   string
	Bytes       []byte
}

type AudioGenerator interface {
	Generate(context.Context, string, AudioGenerateInput) ([]AudioOutput, error)
}

type AudioProcessor interface {
	Process(context.Context, AudioProcessInput) ([]AudioOutput, error)
}

func NormalizeAudioJobRequest(input AudioJobRequest) (AudioJobRequest, error) {
	input.Operation = AudioOperation(strings.TrimSpace(string(input.Operation)))
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.Provider = strings.TrimSpace(input.Provider)
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.SourceAssetID = strings.TrimSpace(input.SourceAssetID)
	if input.Settings == nil {
		input.Settings = map[string]any{}
	}
	if input.Advanced == nil {
		input.Advanced = map[string]any{}
	}
	if input.Operation.IsGeneration() {
		if input.Prompt == "" || input.ModelID == "" {
			return AudioJobRequest{}, ErrInvalidAudioJob
		}
		return input, nil
	}
	if input.Operation.IsProcessing() {
		if input.SourceAssetID == "" {
			return AudioJobRequest{}, ErrInvalidAudioJob
		}
		return input, nil
	}
	return AudioJobRequest{}, ErrInvalidAudioJob
}
