package httpapi

import (
	"context"
	"fmt"
	"strings"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
)

type configuredAgentAudioGenerator struct {
	catalog novelfetchworkshop.Store
}

func (g *configuredAgentAudioGenerator) Generate(ctx context.Context, owner string, input agentworkspace.AudioGenerateInput) ([]agentworkspace.AudioOutput, error) {
	if g == nil || g.catalog == nil {
		return nil, fmt.Errorf("音频模型配置不可用")
	}
	settings, err := g.catalog.GetConfig(ctx, owner)
	if err != nil {
		return nil, err
	}
	modelID := strings.TrimSpace(input.ModelID)
	var selected map[string]any
	if modelID != builtInTTSModelID {
		for _, model := range modelCatalog(settings) {
			if model["enabled"] != true || !strings.EqualFold(modelString(model, "kind"), "audio") {
				continue
			}
			if modelString(model, "id") == modelID || modelString(model, "modelId") == modelID {
				selected = model
				break
			}
		}
	}
	if selected == nil && modelID != builtInTTSModelID {
		return nil, fmt.Errorf("所选音频模型不可用")
	}
	voice := audioInputString(input.Settings, "voice")
	if voice == "" {
		voice = audioInputString(input.Advanced, "voice")
	}
	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}
	speed := audioSettingNumber(input.Settings, "speed")
	if speed == 0 {
		speed = 1
	}
	pitch := input.Settings["pitch"]
	if pitch == nil {
		pitch = 0
	}
	style := strings.TrimSpace(string(input.Operation))
	if emotion := audioInputString(input.Advanced, "emotion"); emotion != "" {
		style = emotion
	}
	// Keep the operation visible to templates even when an emotion is selected.
	// Sound/music providers commonly route by this value rather than a TTS voice.
	if input.Operation == agentworkspace.AudioMusic || input.Operation == agentworkspace.AudioSoundEffect {
		style = strings.TrimSpace(string(input.Operation))
	}
	ttsInput := ttsRequest{Input: input.Prompt, Voice: voice, Speed: speed, Pitch: pitch, Style: style}
	var payload []byte
	var contentType string
	if modelID == builtInTTSModelID {
		payload, contentType, err = requestDefaultTTSBytes(ctx, ttsInput)
	} else {
		payload, contentType, err = requestConfiguredTTSBytes(ctx, ttsInput, selected)
	}
	if err != nil {
		return nil, err
	}
	extension := "mp3"
	if strings.Contains(strings.ToLower(contentType), "wav") {
		extension = "wav"
	} else if strings.Contains(strings.ToLower(contentType), "ogg") {
		extension = "ogg"
	} else if strings.Contains(strings.ToLower(contentType), "mp4") || strings.Contains(strings.ToLower(contentType), "m4a") {
		extension = "m4a"
	}
	return []agentworkspace.AudioOutput{{Name: string(input.Operation) + "." + extension, ContentType: contentType, Extension: extension, Bytes: payload}}, nil
}

func audioInputString(values map[string]any, key string) string {
	if values == nil || values[key] == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(values[key]))
}
