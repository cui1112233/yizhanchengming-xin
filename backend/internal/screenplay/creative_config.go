package screenplay

import (
	"qiantie/backend/internal/prompt"
	"strings"
)

type CreativePromptSnapshot = prompt.CreativePromptSnapshot
type CreativeStyleSnapshot struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Prompt   string   `json:"prompt"`
	Images   []string `json:"images"`
	Revision int64    `json:"revision"`
}
type CreativeConfig struct {
	PlanDensity                 string                            `json:"plan_density"`
	StoryboardAutoSyncAssets    bool                              `json:"storyboard_auto_sync_assets"`
	StoryboardUseContextHistory bool                              `json:"storyboard_use_context_history"`
	ConfigVersion               int                               `json:"config_version"`
	AspectRatio                 string                            `json:"aspect_ratio"`
	ImageResolution             string                            `json:"image_resolution"`
	AssetPromptType             string                            `json:"asset_prompt_type"`
	AssetProcessPrompt          string                            `json:"asset_process_prompt"`
	AssetProcessPromptCleared   bool                              `json:"asset_process_prompt_cleared"`
	ImageCoreInstructionCleared bool                              `json:"image_core_instruction_cleared"`
	VoiceDesignModel            string                            `json:"voice_design_model"`
	ShowOriginalText            bool                              `json:"show_original_text"`
	CharacterImageMode          string                            `json:"character_image_mode"`
	SceneImageMode              string                            `json:"scene_image_mode"`
	ScriptPromptID              string                            `json:"script_prompt_id"`
	StoryboardPromptID          string                            `json:"storyboard_prompt_id"`
	CharacterDesignPromptID     string                            `json:"character_design_prompt_id"`
	SceneDesignPromptID         string                            `json:"scene_design_prompt_id"`
	ImageCoreInstruction        string                            `json:"image_core_instruction"`
	ImageModelID                string                            `json:"image_model_id"`
	ImageEditModelID            string                            `json:"image_edit_model_id"`
	Style                       *CreativeStyleSnapshot            `json:"style,omitempty"`
	PromptSnapshots             map[string]CreativePromptSnapshot `json:"-"`
}

func selectedScriptPrompt(p Project) string {
	if p.CreativeConfig != nil && p.CreativeConfig.ScriptPromptID != "" {
		return p.CreativeConfig.ScriptPromptID
	}
	return p.PromptID
}
func ResolveProjectPrompt(p Project, id, purpose string) (CreativePromptSnapshot, error) {
	resolved, err := prompt.ResolveCreativePrompt(id, purpose)
	if err != nil {
		return resolved, err
	}
	if p.CreativeConfig != nil {
		if frozen, ok := p.CreativeConfig.PromptSnapshots[purpose]; ok && frozen.ID == resolved.ID && frozen.Type == purpose && frozen.Body != "" {
			// Older snapshots contain only body and identity. Reuse the verified
			// catalog contract only when the complete body still matches.
			if frozen.ID != "luoshui_storyboard_139" && frozen.OutputProtocol == "" && strings.ReplaceAll(frozen.Body, "\r\n", "\n") == strings.ReplaceAll(resolved.Body, "\r\n", "\n") {
				frozen.CreativePromptContract = resolved.CreativePromptContract
			}
			return frozen, nil
		}
	}
	return resolved, nil
}
