package batchfactoryv11

import (
	"encoding/json"
	"strings"
)

// applyVideoPromptOverrides exposes durable per-shot prompt edits in readback.
// The record stays the original director output while a video-scope setting is
// the user's current effective prompt.
func applyVideoPromptOverrides(video *Video) {
	if video == nil {
		return
	}
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "videoPrompt"); ok {
		video.VideoPrompt = value
	}
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "visualPrompt"); ok {
		video.VisualPrompt = value
	}
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "comicPrompt"); ok {
		video.ComicPrompt = value
	}
}

func optionalPromptValue(patch SettingsPatch, key string) (string, bool) {
	raw, ok := patch[key]
	if !ok {
		return "", false
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	return strings.TrimSpace(value), true
}
