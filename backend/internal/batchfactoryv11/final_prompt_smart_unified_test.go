package batchfactoryv11

import (
	"encoding/json"
	"testing"
)

func TestSmartUnifiedStylePrefersLatestPersistedAnalysis(t *testing.T) {
	encodedAnalysis, _ := json.Marshal(`{"schema_version":"h3-style-system/v1","prompt":"最新保存的全片统一风格。","fields":{"final_genre":"古装短剧","trailer_style":"电影感","story_era":"古代宅邸"},"preset":{"id":"script-constraint-prefix-smart-unified","version":1}}`)
	book := Book{
		SettingsState: SettingsState{Patch: SettingsPatch{
			"h3StyleAnalysis": encodedAnalysis,
		}},
		DirectorRevision: &DirectorRevision{Output: DirectorResult{SmartUnifiedStyle: "旧导演修订风格。"}},
	}
	config := AIReasoningPromptConfig{Constraints: AIReasoningConstraintModule{AIReasoningScope: AIReasoningScope{Enabled: true}, Selections: []PresetSnapshot{{ID: smartUnifiedPrefixPresetID, ConstraintCategory: "prefix"}}}}
	if got := smartUnifiedStyleForRevision(book, config); got != "最新保存的全片统一风格。" {
		t.Fatalf("smart unified style = %q, want latest persisted analysis", got)
	}
}
