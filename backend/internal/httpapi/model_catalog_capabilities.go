package httpapi

import (
	"fmt"
	"strconv"
	"strings"
)

type PublicDurationCapabilities struct {
	Min     int `json:"min,omitempty"`
	Max     int `json:"max,omitempty"`
	Default int `json:"default,omitempty"`
}

type PublicModelCapabilities struct {
	Ratios          []string                   `json:"ratios,omitempty"`
	Resolutions     []string                   `json:"resolutions,omitempty"`
	QualityLevels   []string                   `json:"qualityLevels,omitempty"`
	Duration        PublicDurationCapabilities `json:"duration,omitempty"`
	MaxOutputCount  int                        `json:"maxOutputCount,omitempty"`
	ReferenceKinds  []string                   `json:"referenceKinds,omitempty"`
	GenerationModes []string                   `json:"generationModes,omitempty"`
	SupportsAudio   bool                       `json:"supportsAudio"`
	AdvancedFlags   []string                   `json:"advancedFlags,omitempty"`
}

func publicModelCapabilities(model map[string]any) PublicModelCapabilities {
	kind := strings.TrimSpace(fmt.Sprint(model["kind"]))
	switch kind {
	case "image":
		return PublicModelCapabilities{
			Ratios: []string{"16:9", "9:16", "1:1"}, Resolutions: []string{"1K", "2K", "4K"},
			QualityLevels: []string{"low", "standard", "high"}, MaxOutputCount: 4,
			ReferenceKinds: []string{"image"}, GenerationModes: []string{"text_to_image", "image_to_image"},
			AdvancedFlags: []string{"negative_prompt", "seed", "autolink"},
		}
	case "video":
		return PublicModelCapabilities{
			Ratios: []string{"16:9", "9:16"}, Resolutions: []string{"480P", "720P", "1080P"},
			Duration: PublicDurationCapabilities{Min: 5, Max: 15, Default: 5}, MaxOutputCount: 1,
			ReferenceKinds: []string{"image"}, GenerationModes: []string{"text_to_video", "image_to_video", "first_frame"},
			AdvancedFlags: []string{"camera_motion", "autolink"},
		}
	case "audio":
		return PublicModelCapabilities{
			MaxOutputCount: 1, GenerationModes: []string{"speech", "music", "effect"},
			AdvancedFlags: []string{"voice", "speed", "pitch", "emotion", "format"},
		}
	default:
		return PublicModelCapabilities{
			MaxOutputCount: 1, ReferenceKinds: []string{"text", "image", "video", "audio"},
			GenerationModes: []string{"generate", "rewrite", "continue"},
			AdvancedFlags:   []string{"length", "creativity", "structure"},
		}
	}
}

func includesString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func settingString(settings map[string]any, key string) string {
	value, ok := settings[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func settingInt(settings map[string]any, key string) int {
	value := settings[key]
	switch typed := value.(type) {
	case int:
		return typed
	case float64:
		return int(typed)
	default:
		parsed, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
		return parsed
	}
}

func ValidateModelSettings(model map[string]any, kind string, settings map[string]any) error {
	modelKind := strings.TrimSpace(fmt.Sprint(model["kind"]))
	if modelKind != kind {
		return fmt.Errorf("kind: 模型类型为 %s，不能用于 %s 节点", modelKind, kind)
	}
	capabilities, ok := model["capabilities"].(PublicModelCapabilities)
	if !ok {
		capabilities = publicModelCapabilities(model)
	}
	if value := settingString(settings, "ratio"); value != "" && len(capabilities.Ratios) > 0 && !includesString(capabilities.Ratios, value) {
		return fmt.Errorf("ratio: 当前模型不支持 %s", value)
	}
	if value := strings.ToUpper(settingString(settings, "resolution")); value != "" && len(capabilities.Resolutions) > 0 && !includesString(capabilities.Resolutions, value) {
		return fmt.Errorf("resolution: 当前模型不支持 %s", value)
	}
	if value := settingString(settings, "quality"); value != "" && len(capabilities.QualityLevels) > 0 && !includesString(capabilities.QualityLevels, value) {
		return fmt.Errorf("quality: 当前模型不支持 %s", value)
	}
	if value := settingString(settings, "mode"); value != "" && len(capabilities.GenerationModes) > 0 && !includesString(capabilities.GenerationModes, value) {
		return fmt.Errorf("mode: 当前模型不支持 %s", value)
	}
	if count := settingInt(settings, "count"); count > 0 && capabilities.MaxOutputCount > 0 && count > capabilities.MaxOutputCount {
		return fmt.Errorf("count: 当前模型最多生成 %d 个结果", capabilities.MaxOutputCount)
	}
	if duration := settingInt(settings, "durationSeconds"); duration > 0 && capabilities.Duration.Max > 0 && (duration < capabilities.Duration.Min || duration > capabilities.Duration.Max) {
		return fmt.Errorf("durationSeconds: 当前模型只支持 %d-%d 秒", capabilities.Duration.Min, capabilities.Duration.Max)
	}
	return nil
}
