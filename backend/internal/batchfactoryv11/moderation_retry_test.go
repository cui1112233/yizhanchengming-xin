package batchfactoryv11

import (
	"encoding/json"
	"testing"
)

func TestModerationRetrySwitchDefaultsOff(t *testing.T) {
	t.Setenv(StripReferenceImagesOnModerationEnv, "")
	if stripReferenceImagesOnModerationEnabled() {
		t.Fatal("switch must default off")
	}
	t.Setenv(StripReferenceImagesOnModerationEnv, "1")
	if !stripReferenceImagesOnModerationEnabled() {
		t.Fatal("switch must accept 1")
	}
	t.Setenv(StripReferenceImagesOnModerationEnv, "true")
	if !stripReferenceImagesOnModerationEnabled() {
		t.Fatal("switch must accept true")
	}
	t.Setenv(StripReferenceImagesOnModerationEnv, "0")
	if stripReferenceImagesOnModerationEnabled() {
		t.Fatal("switch must reject 0")
	}
}

func TestModerationRetryRequiresModerationKeyword(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    bool
	}{
		{"chinese moderation", "内容不符合平台规范，请修改后重试", true},
		{"provider policy", "provider rejected: content_policy_violation", true},
		{"nsfw flag", "NSFW content detected", true},
		{"transient network", "connection reset by peer", false},
		{"quota", "insufficient balance", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		if got := isModerationRejectionMessage(tc.message); got != tc.want {
			t.Fatalf("%s: matches=%v want %v", tc.name, got, tc.want)
		}
	}
}

func TestStripReferenceImagesClearsEveryCarrier(t *testing.T) {
	prompt := FinalPrompt{
		CompiledPrompt:     "frozen prompt",
		ReferenceImageURLs: []string{"https://cdn.example.com/hero.png"},
		DowngradedAssetIDs: []string{"asset-1"},
		EffectiveSettings: EffectiveSettings{Values: SettingsPatch{
			"imageUrls":       json.RawMessage(`["https://cdn.example.com/hero.png"]`),
			"referenceImages": json.RawMessage(`[{"url":"https://cdn.example.com/hero.png"}]`),
			"durationSec":     json.RawMessage(`10`),
		}},
	}
	stripped := stripReferenceImagesFromPrompt(prompt)
	if len(stripped.ReferenceImageURLs) != 0 || len(stripped.DowngradedAssetIDs) != 0 {
		t.Fatalf("reference carriers survived: %#v", stripped)
	}
	if _, ok := stripped.EffectiveSettings.Values["imageUrls"]; ok {
		t.Fatal("imageUrls setting survived")
	}
	if _, ok := stripped.EffectiveSettings.Values["referenceImages"]; ok {
		t.Fatal("referenceImages setting survived")
	}
	if string(stripped.EffectiveSettings.Values["durationSec"]) != "10" {
		t.Fatal("unrelated settings must be preserved")
	}
	if stripped.CompiledPrompt != "frozen prompt" {
		t.Fatal("compiled prompt must not change")
	}
}
