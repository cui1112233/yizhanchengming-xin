package batchfactoryv11

import (
	"context"
	"testing"
)

func TestNormalizeProductionModeRestoresLegacyEnabledComicModule(t *testing.T) {
	mode := normalizeProductionMode(SettingsPatch{"aiPromptConfig": rawSetting(t, map[string]any{
		"comic": map[string]any{"enabled": true, "body": "COMIC PAGE PRESET"},
	})})
	if mode != "comic" {
		t.Fatalf("mode=%q, want comic", mode)
	}
}

func TestNormalizeProductionModeExplicitVideoOverridesComicModule(t *testing.T) {
	mode := normalizeProductionMode(SettingsPatch{
		"productionMode": rawSetting(t, "video"),
		"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true, "body": "COMIC PAGE PRESET"}}),
	})
	if mode != "video" {
		t.Fatalf("mode=%q, want video", mode)
	}
}

func TestRunComicAssetExtractionReplacesOnlyComicAssets(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"productionMode": rawSetting(t, "comic"),
			"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
				"enabled": true, "body": "COMIC PAGE PRESET", "assets": map[string]any{"id": "vip-smart-character-derive", "body": "COMIC ASSET PRESET", "version": 1},
			}}),
		},
		ExpectedRevision: batch.SettingsState.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, CreateBookAssetInput{Kind: "prop", Name: "手动保留道具", Prompt: "人工设定"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PersistComicExtractedBookAssets(context.Background(), "alice", book, DirectorSnapshot{}, DirectorAssets{Characters: []NamedPrompt{{Name: "旧漫画角色", Prompt: "旧漫画提示词"}}}); err != nil {
		t.Fatal(err)
	}

	assets, err := (&DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{`{"characters":[{"name":"新漫画角色","prompt":"新漫画提示词"}],"scenes":[],"props":[]}`}}}).RunComicAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]string{}
	for _, asset := range assets {
		seen[asset.Name] = asset.Source
	}
	if seen["新漫画角色"] != "comic" {
		t.Fatalf("new comic asset source=%q, assets=%+v", seen["新漫画角色"], assets)
	}
	if _, found := seen["旧漫画角色"]; found {
		t.Fatalf("stale comic asset was not replaced: %+v", assets)
	}
	if seen["手动保留道具"] != "manual" {
		t.Fatalf("manual asset was changed: %+v", assets)
	}
}
