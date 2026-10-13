package batchfactoryv11

import "testing"

func TestParseSmartUnifiedAnalysisKeepsSDUnifiedStyle(t *testing.T) {
	analysis, err := parseSmartUnifiedAnalysis(`{"schema_version":"h3-style-system/v1","prompt":"兼容视觉基线","sd_unified_style":"古装内宅暖灰微冷，主体清晰，暗部保留木纹与衣料纹理。","fields":{"final_genre":"古装短剧","trailer_style":"电影感","story_era":"古代宅邸"},"preset":{"id":"script-constraint-prefix-smart-unified","name":"智能统一","version":2}}`)
	if err != nil {
		t.Fatalf("parse smart unified analysis: %v", err)
	}
	if analysis.SDUnifiedStyle == "" {
		t.Fatal("sd_unified_style was dropped from the frozen analysis")
	}
}
