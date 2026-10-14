package batchfactoryv11

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestH3AppearanceBatchOneRequestAndIdentityMapping(t *testing.T) {
	assets := DirectorAssets{Characters: []NamedPrompt{{Name: "林晚", Prompt: "原始甲"}, {Name: "陆沉", Prompt: "原始乙"}}}
	first, second := strings.Repeat("甲的完整外形", 25), strings.Repeat("乙的完整外形", 25)
	raw, _ := json.Marshal(map[string]any{"characters": []map[string]string{
		{"character_id": "C002", "name": "陆沉", "prompt": second},
		{"character_id": "C001", "name": "林晚", "prompt": first},
	}})
	provider := &queuedDirectorProvider{values: []string{string(raw)}}
	snapshot := DirectorSnapshot{Effective: SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{"assets": map[string]any{
		"enabled": true, "character": map[string]any{"presetId": "batch-character-h3", "presetKey": "h3-character-normal", "body": "后台统一人物外形要求"},
	}})}}
	got, err := (&DirectorService{Provider: provider}).compileH3AssetPrompts(context.Background(), Book{SourceText: "林晚见到陆沉。"}, snapshot, assets)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 1 {
		t.Fatalf("appearance calls=%d, want 1", len(provider.calls))
	}
	if got.Characters[0].Prompt != first || got.Characters[1].Prompt != second {
		t.Fatalf("identity mismatch: %+v", got)
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "后台统一人物外形要求") {
		t.Fatal("selected backend preset not used")
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "不得写镜头动作") {
		t.Fatal("H3 appearance prompt must enforce static asset boundaries")
	}
}

func TestH3AppearanceBatchRejectsIncompleteWithoutFallback(t *testing.T) {
	for _, response := range []string{
		`{"characters":[{"character_id":"C001","name":"甲","prompt":"外形"}]}`,
		`{"characters":[{"character_id":"C001","name":"甲","prompt":"外形"},{"character_id":"C001","name":"甲","prompt":"外形"}]}`,
		`{"characters":[{"character_id":"C001","name":"甲","prompt":"外形"},{"character_id":"C999","name":"乙","prompt":"外形"}]}`,
		`{"characters":[{"character_id":"C001","name":"乙","prompt":"外形"},{"character_id":"C002","name":"甲","prompt":"外形"}]}`,
		`{"characters":[{"character_id":"C001","name":"甲","prompt":"外形"},{"character_id":"C002","name":"乙","prompt":""}]}`,
		`{"characters":[`,
	} {
		t.Run(response, func(t *testing.T) {
			provider := &queuedDirectorProvider{values: []string{response}}
			assets := DirectorAssets{Characters: []NamedPrompt{{Name: "甲", Prompt: "原甲"}, {Name: "乙", Prompt: "原乙"}}}
			_, err := (&DirectorService{Provider: provider}).compileH3CharacterBatch(context.Background(), Book{SourceText: "甲乙"}, PresetSnapshot{Body: "后台规则"}, assets)
			if err == nil || len(provider.calls) != 1 {
				t.Fatalf("err=%v calls=%d", err, len(provider.calls))
			}
			if assets.Characters[0].Prompt != "原甲" || assets.Characters[1].Prompt != "原乙" {
				t.Fatal("partial result mutated assets")
			}
		})
	}
}

func TestH3AppearanceBatchEmptySkipsModel(t *testing.T) {
	p := &queuedDirectorProvider{}
	_, err := (&DirectorService{Provider: p}).compileH3CharacterBatch(context.Background(), Book{}, PresetSnapshot{}, DirectorAssets{})
	if err != nil || len(p.calls) != 0 {
		t.Fatalf("empty roster called model: %v", err)
	}
}
