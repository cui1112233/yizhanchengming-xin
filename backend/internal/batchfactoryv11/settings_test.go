package batchfactoryv11

import (
	"encoding/json"
	"testing"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

func TestSparsePatchPreservesFalseEmptyZero(t *testing.T) {
	current := SettingsPatch{"enabled": raw(true), "prefix": raw("x"), "duration": raw(9)}
	got := ApplySparseUpdate(current, SettingsUpdate{Patch: SettingsPatch{"enabled": raw(false), "prefix": raw(""), "duration": raw(0)}})
	for key, want := range map[string]string{"enabled": "false", "prefix": "\"\"", "duration": "0"} {
		if string(got[key]) != want {
			t.Fatalf("%s=%s want %s", key, got[key], want)
		}
	}
}

func TestNormalizeProductionModeDefaultsToVideo(t *testing.T) {
	if got := normalizeProductionMode(SettingsPatch{}); got != "video" {
		t.Fatalf("got %q want video", got)
	}
	if got := normalizeProductionMode(SettingsPatch{"productionMode": raw("comic")}); got != "comic" {
		t.Fatalf("got %q want comic", got)
	}
	if got := normalizeProductionMode(SettingsPatch{"productionMode": raw("unknown")}); got != "video" {
		t.Fatalf("got %q want video", got)
	}
}

func TestRestoreInheritanceRemovesOnlySelectedCurrentScopeKeys(t *testing.T) {
	current := SettingsPatch{"quality": raw("book"), "enabled": raw(false), "duration": raw(0)}
	got := ApplySparseUpdate(current, SettingsUpdate{RestoreKeys: []string{"quality"}})
	if _, ok := got["quality"]; ok {
		t.Fatal("quality should inherit after restore")
	}
	if string(got["enabled"]) != "false" || string(got["duration"]) != "0" {
		t.Fatalf("got=%v", got)
	}
}

func TestResolveSettingsUsesPresenceNotTruthiness(t *testing.T) {
	effective := ResolveSettings(
		SettingsPatch{"enabled": raw(true), "prefix": raw("system"), "duration": raw(10)},
		SettingsPatch{"enabled": raw(false)},
		SettingsPatch{"prefix": raw("")},
		SettingsPatch{"duration": raw(0)},
	)
	if string(effective["enabled"]) != "false" || string(effective["prefix"]) != "\"\"" || string(effective["duration"]) != "0" {
		t.Fatalf("effective=%v", effective)
	}
}

func TestResolveSettingsDeepMergesAIPromptConfigModules(t *testing.T) {
	batch := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"assets": map[string]any{"enabled": true, "presetId": "batch-assets"},
		"visual": map[string]any{"enabled": false},
	})}
	book := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"constraints": map[string]any{"enabled": true, "presetId": "book-constraints"},
	})}

	effective := ResolveSettings(batch, book)
	var prompt map[string]json.RawMessage
	if err := json.Unmarshal(effective["aiPromptConfig"], &prompt); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"assets", "constraints", "visual"} {
		if _, ok := prompt[key]; !ok {
			t.Fatalf("missing %s in %#v", key, prompt)
		}
	}
}

func TestResolveSettingsKeepsLegacyFullBookPromptPayload(t *testing.T) {
	batch := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"assets": map[string]any{"presetId": "batch-assets"},
		"video":  map[string]any{"presetId": "batch-video"},
	})}
	book := SettingsPatch{"aiPromptConfig": raw(map[string]any{
		"assets": map[string]any{"presetId": "book-assets"},
		"video":  map[string]any{"presetId": "book-video"},
	})}

	effective := ResolveSettings(batch, book)
	var prompt map[string]map[string]string
	if err := json.Unmarshal(effective["aiPromptConfig"], &prompt); err != nil {
		t.Fatal(err)
	}
	if prompt["assets"]["presetId"] != "book-assets" || prompt["video"]["presetId"] != "book-video" {
		t.Fatalf("legacy book prompt payload changed: %#v", prompt)
	}
}

func TestNormalizeLegacyNestedPatchRestoresTheSavedUnifiedConfiguration(t *testing.T) {
	legacy := SettingsPatch{
		"patch": raw(map[string]any{
			"patch": map[string]any{
				"textModelId": "gpt-5.4",
				"publishSettings": map[string]any{
					"websiteProfileId": "121-330",
					"horizontalFlip":   true,
				},
			},
		}),
		"aiPromptConfig":    raw(map[string]any{"video": map[string]any{"presetId": "h3-director"}}),
		"expectedRevision": raw(3),
	}

	got := normalizeLegacyNestedPatch(legacy)
	if _, exists := got["patch"]; exists {
		t.Fatalf("legacy wrapper leaked into the effective settings: %#v", got)
	}
	if _, exists := got["expectedRevision"]; exists {
		t.Fatalf("request metadata leaked into the effective settings: %#v", got)
	}
	if string(got["textModelId"]) != "\"gpt-5.4\"" {
		t.Fatalf("text model not restored: %s", got["textModelId"])
	}
	var publish map[string]any
	if err := json.Unmarshal(got["publishSettings"], &publish); err != nil {
		t.Fatal(err)
	}
	if publish["websiteProfileId"] != "121-330" || publish["horizontalFlip"] != true {
		t.Fatalf("publish settings not restored: %#v", publish)
	}
	if string(got["aiPromptConfig"]) == "" {
		t.Fatalf("newer outer settings were not preserved: %#v", got)
	}
}
