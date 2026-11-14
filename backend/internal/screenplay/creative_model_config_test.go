package screenplay

import (
	"context"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

func TestCreativeModelConfigDefaultsPersistClearAndLegacy(t *testing.T) {
	p := Project{AspectRatio: "4:5"}
	c := EffectiveCreativeConfig(p)
	if c.ImageResolution != "2K" || c.AssetPromptType != "general" || c.VoiceDesignModel != "minimax" || !c.ShowOriginalText || c.ImageCoreInstruction != prompt.ScreenplayImageCoreDefault || c.AssetProcessPrompt != prompt.ScreenplayAssetProcessDefault || c.AspectRatio != "4:5" {
		t.Fatal(c)
	}
	c.ConfigVersion = 1
	c.ShowOriginalText = false
	c.ImageCoreInstruction = ""
	c.ImageCoreInstructionCleared = true
	c.AssetProcessPrompt = ""
	c.AssetProcessPromptCleared = true
	c.ImageResolution = "4K"
	c.AssetPromptType = "western"
	c.VoiceDesignModel = "qwen-tts"
	c.CharacterImageMode = "three_view"
	c.SceneImageMode = "none"
	raw, e := encodeCreativeConfig(&c)
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := decodeCreativeConfig(raw)
	if e != nil {
		t.Fatal(e)
	}
	effective := EffectiveCreativeConfig(Project{CreativeConfig: decoded, AspectRatio: "4:5"})
	if effective.ShowOriginalText || effective.ImageCoreInstruction != "" || effective.AssetProcessPrompt != "" || effective.ImageResolution != "4K" || effective.VoiceDesignModel != "qwen-tts" || effective.AssetPromptType != "western" || effective.CharacterImageMode != "three_view" || effective.SceneImageMode != "none" {
		t.Fatal(effective)
	}
	old := EffectiveCreativeConfig(Project{CreativeConfig: &CreativeConfig{ImageCoreInstruction: "", CharacterDesignPromptID: prompt.CreativeCharacter213}})
	if old.ImageCoreInstruction != "" || !old.ShowOriginalText || old.CharacterImageMode != "typedprompt:"+prompt.CreativeCharacter213 {
		t.Fatal(old)
	}
}

func TestCreativeAssetPromptTypeChangesActualExtractionInstruction(t *testing.T) {
	for _, kind := range []string{"general", "western"} {
		t.Run(kind, func(t *testing.T) {
			service, store, generator := newAssetService(t, PermissionBypass)
			p, _ := store.GetProject(context.Background(), "alice", "p1")
			_, e := store.SaveCreativeConfig(context.Background(), "alice", "p1", CreativeConfig{ConfigVersion: 1, AssetPromptType: kind}, p.Revision)
			if e != nil {
				t.Fatal(e)
			}
			_, e = service.Extract(context.Background(), AssetRequest{Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeMaster, Content: "海外原文内容", Confirmed: true, Model: "mock-model"})
			if e != nil {
				t.Fatal(e)
			}
			request := generator.requests[0]
			hasWestern := strings.Contains(request.SystemPrompt, prompt.ScreenplayWesternAssetExtraction)
			if hasWestern != (kind == "western") || request.Model != "mock-model" || !strings.Contains(request.UserPrompt, "海外原文内容") {
				t.Fatal(kind, request)
			}
		})
	}
}

func TestCreativeModelLegacyAssetProcessMigrationPreservesCoreClears(t *testing.T) {
	cases := []struct{ name, raw, wantProcess string }{
		{"legacy field absent", `{"config":{"image_core_instruction":""}}`, prompt.ScreenplayAssetProcessDefault},
		{"legacy clear flag", `{"config":{"image_core_instruction":"","asset_process_prompt":"","asset_process_prompt_cleared":true}}`, ""},
		{"legacy existing text", `{"config":{"image_core_instruction":"","asset_process_prompt":"用户原有处理规则"}}`, "用户原有处理规则"},
		{"current explicit empty", `{"config":{"config_version":1,"image_core_instruction":"","asset_process_prompt":""}}`, ""},
		{"current clear flag", `{"config":{"config_version":1,"image_core_instruction":"","asset_process_prompt":"","asset_process_prompt_cleared":true}}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, e := decodeCreativeConfig([]byte(tc.raw))
			if e != nil {
				t.Fatal(e)
			}
			effective := EffectiveCreativeConfig(Project{CreativeConfig: decoded})
			if effective.AssetProcessPrompt != tc.wantProcess || effective.ImageCoreInstruction != "" {
				t.Fatal(effective)
			}
			raw, e := encodeCreativeConfig(&effective)
			if e != nil {
				t.Fatal(e)
			}
			again, e := decodeCreativeConfig(raw)
			if e != nil {
				t.Fatal(e)
			}
			if got := EffectiveCreativeConfig(Project{CreativeConfig: again}); got.AssetProcessPrompt != tc.wantProcess || got.ImageCoreInstruction != "" {
				t.Fatal(got)
			}
		})
	}
}
