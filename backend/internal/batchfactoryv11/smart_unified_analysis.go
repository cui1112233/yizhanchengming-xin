package batchfactoryv11

import (
	"encoding/json"
	"fmt"
	"strings"
)

const smartUnifiedAnalysisSchemaV1 = "h3-style-system/v1"

// SmartUnifiedAnalysis is the frozen outcome of the H3 style.system call.
// Prompt is the only text eligible for final VIDEO injection; Fields and
// Preset keep the result inspectable after the editable system preset changes.
type SmartUnifiedAnalysis struct {
	SchemaVersion string `json:"schema_version"`
	Prompt        string `json:"prompt"`
	// SDUnifiedStyle is the compact, provider-facing style paragraph required
	// by the SD preset. Prompt remains the legacy/H3-compatible field.
	SDUnifiedStyle string                     `json:"sd_unified_style,omitempty"`
	Fields         map[string]string          `json:"fields"`
	Preset         SmartUnifiedPresetSnapshot `json:"preset"`
}

type SmartUnifiedPresetSnapshot struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version int64  `json:"version"`
}

func parseSmartUnifiedAnalysis(value string) (*SmartUnifiedAnalysis, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, nil
	}
	if !strings.HasPrefix(raw, "{") {
		return &SmartUnifiedAnalysis{Prompt: raw}, nil
	}
	var analysis SmartUnifiedAnalysis
	if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
		return nil, fmt.Errorf("%w: invalid style.system analysis", ErrInvalid)
	}
	if analysis.SchemaVersion != smartUnifiedAnalysisSchemaV1 || strings.TrimSpace(analysis.Prompt) == "" {
		return nil, fmt.Errorf("%w: invalid style.system analysis", ErrInvalid)
	}
	for _, key := range []string{"final_genre", "trailer_style", "story_era"} {
		if strings.TrimSpace(analysis.Fields[key]) == "" {
			return nil, fmt.Errorf("%w: style.system field %s is required", ErrInvalid, key)
		}
	}
	if analysis.Preset.ID == "" || analysis.Preset.Version <= 0 {
		return nil, fmt.Errorf("%w: style.system preset snapshot is required", ErrInvalid)
	}
	analysis.Prompt = strings.TrimSpace(analysis.Prompt)
	return &analysis, nil
}
