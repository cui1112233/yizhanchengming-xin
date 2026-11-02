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
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "aiOriginalVideoPrompt"); ok {
		video.AIOriginalVideoPrompt = value
	}
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "visualPrompt"); ok {
		video.VisualPrompt = value
	}
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "comicPrompt"); ok {
		video.ComicPrompt = value
	}
	if value, ok := optionalPromptValue(video.SettingsState.Patch, "comicImagePrompt"); ok {
		video.ComicImagePrompt = value
	}
	if value, ok := optionalStringSliceValue(video.SettingsState.Patch, "comicRoles"); ok {
		video.ComicRoles = value
	}
	if value, ok := optionalStringMapValue(video.SettingsState.Patch, "comicRoleStage"); ok {
		video.ComicRoleStage = value
	}
	if value, ok := optionalStringSliceValue(video.SettingsState.Patch, "comicScenePresets"); ok {
		video.ComicScenePresets = value
	}
}

func optionalStringMapValue(patch SettingsPatch, key string) (map[string]string, bool) {
	raw, ok := patch[key]
	if !ok {
		return nil, false
	}
	var values map[string]string
	if json.Unmarshal(raw, &values) != nil {
		return nil, false
	}
	return normalizedStringMap(values), true
}

func optionalStringSliceValue(patch SettingsPatch, key string) ([]string, bool) {
	raw, ok := patch[key]
	if !ok {
		return nil, false
	}
	var values []string
	if json.Unmarshal(raw, &values) != nil {
		return nil, false
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized, true
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

// initializeAIOriginalVideoPrompt freezes the effective prompt that existed
// before a user first edits a legacy VIDEO. Later edits keep that stored
// original intact while changing only videoPrompt.
func initializeAIOriginalVideoPrompt(update SettingsUpdate, current SettingsPatch, preEditPrompt string) SettingsUpdate {
	if _, editingPrompt := update.Patch["videoPrompt"]; !editingPrompt {
		return update
	}
	if _, suppliedOriginal := update.Patch["aiOriginalVideoPrompt"]; suppliedOriginal {
		return update
	}
	if _, existingOriginal := current["aiOriginalVideoPrompt"]; existingOriginal {
		return update
	}
	if existingPrompt, overridden := optionalPromptValue(current, "videoPrompt"); overridden {
		preEditPrompt = existingPrompt
	}
	encoded, err := json.Marshal(preEditPrompt)
	if err != nil {
		return update
	}
	update.Patch = clonePatch(update.Patch)
	update.Patch["aiOriginalVideoPrompt"] = encoded
	return update
}
