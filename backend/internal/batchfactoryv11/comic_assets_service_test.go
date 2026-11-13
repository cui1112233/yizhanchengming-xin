package batchfactoryv11

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// seedComicAssetBook 准备一本已选漫画资产预设、可直接跑资产提取的书。
func seedComicAssetBook(t *testing.T) (*MemoryStore, Batch, Book) {
	t.Helper()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{
		Title:      "漫画资产截断",
		SourceText: "她叫林溪。周野推门进来。",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"productionMode": rawSetting(t, "comic"),
			"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
				"enabled": true, "assets": map[string]any{"presetId": "batch-comic-assets", "body": "提取漫画资产"},
			}}),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	return store, batch, batch.Books[0]
}

func TestRunComicAssetExtractionRetriesWithCompactPromptAfterTruncation(t *testing.T) {
	store, batch, book := seedComicAssetBook(t)
	provider := &truncatingDirectorProvider{values: []string{"truncated", `{"characters":[{"name":"林溪","prompt":"短发女主"}],"scenes":[],"props":[]}`}}

	assets, err := (&DirectorService{Store: store, Provider: provider}).RunComicAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatalf("comic asset extraction should recover from one truncated response: %v", err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("calls=%d, want exactly one compact retry", len(provider.calls))
	}
	if !strings.Contains(provider.calls[1].SystemPrompt, "更紧凑") || !provider.calls[1].DisableJSONResponse {
		t.Fatalf("retry must request compact output without JSON response mode: %#v", provider.calls[1])
	}
	if len(assets) == 0 || assets[0].Name != "林溪" {
		t.Fatalf("assets=%+v, want the recovered character", assets)
	}
}

func TestRunComicAssetExtractionReportsTruncationAfterSecondAttempt(t *testing.T) {
	store, batch, book := seedComicAssetBook(t)
	provider := &alwaysTruncatingDirectorProvider{}

	_, err := (&DirectorService{Store: store, Provider: provider}).RunComicAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err == nil || !strings.Contains(err.Error(), "长度截断") {
		t.Fatalf("err=%v, want an explicit truncation action for the user", err)
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err=%v, want ErrUnavailable so the stage can report a channel problem", err)
	}
	if provider.calls != 2 {
		t.Fatalf("calls=%d, want one original call plus one compact retry", provider.calls)
	}
}

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

func TestBuildComicAssetExtractionContractReportsMissingSelectionAsConfig(t *testing.T) {
	_, err := BuildComicAssetExtractionContract(Book{ID: "book-1", Title: "长篇漫画", SourceText: "完整小说正文"}, DirectorSnapshot{
		Effective: SettingsPatch{
			"productionMode": rawSetting(t, "comic"),
			"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{"enabled": true}}),
		},
	})
	if err == nil {
		t.Fatal("expected a missing comic asset prompt error")
	}
	// 缺配置必须是可以直接照做的动作文案，不能带 revision conflict 前缀。
	if err.Error() != "请在资产设置中选择漫画资产提示词" {
		t.Fatalf("err=%q, want plain action text", err.Error())
	}
	if !errors.Is(err, ErrConfigRequired) {
		t.Fatalf("err=%v, want ErrConfigRequired", err)
	}
	if errors.Is(err, ErrConflict) {
		t.Fatalf("missing configuration must not be reported as ErrConflict: %v", err)
	}
}

func TestBuildComicAssetExtractionContractAllowsCompleteCMOutput(t *testing.T) {
	contract, err := BuildComicAssetExtractionContract(Book{ID: "book-1", Title: "长篇漫画", SourceText: "完整小说正文"}, DirectorSnapshot{
		Effective: SettingsPatch{
			"productionMode": rawSetting(t, "comic"),
			"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
				"enabled": true,
				"assets":  map[string]any{"presetId": "batch-comic-assets", "body": "完整提取人物、场景和道具"},
			}}),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if contract.MaxTokens < 12000 {
		t.Fatalf("comic asset max tokens=%d, want at least 12000 for complete CM output", contract.MaxTokens)
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

func TestRunComicAssetExtractionPersistsResolvedGenderForAmbiguousFirstPerson(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{
		Title:          "女频书",
		SourceText:     "我推开房门。",
		SourceMetadata: map[string]any{"gender": "女频"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err = store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"productionMode": rawSetting(t, "comic"),
			"aiPromptConfig": rawSetting(t, map[string]any{"comic": map[string]any{
				"enabled": true, "assets": map[string]any{"presetId": "batch-comic-assets", "body": "提取漫画资产"},
			}}),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	assets, err := (&DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{`{"characters":[{"name":"我","facts":"身份：新来的住客。","appearance":"短发，深色外套。","prompt":"短发住客，深色外套。"}],"scenes":[],"props":[]}`}}}).RunComicAssetExtraction(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, asset := range assets {
		if asset.Source == "comic" && asset.Name == "我" {
			for _, value := range []string{asset.Facts, asset.Appearance, asset.Prompt} {
				if !containsAnyGenderMarker(value, "女") {
					t.Fatalf("persisted comic asset missed female fallback: %#v", asset)
				}
			}
			return
		}
	}
	t.Fatalf("first-person comic asset was not persisted: %#v", assets)
}

func TestUpdateComicBookAssetPreservesComicSource(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	assets, err := store.PersistComicExtractedBookAssets(context.Background(), "alice", book, DirectorSnapshot{}, DirectorAssets{
		Characters: []NamedPrompt{{Name: "阿月", Prompt: "青衣少女"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var comicAsset BookAsset
	for _, asset := range assets {
		if asset.Source == "comic" {
			comicAsset = asset
			break
		}
	}
	if comicAsset.ID == "" {
		t.Fatalf("comic asset not found: %+v", assets)
	}

	updated, err := store.UpdateBookAsset(context.Background(), "alice", batch.ID, book.ID, comicAsset.ID, UpdateBookAssetInput{
		Name: comicAsset.Name, Prompt: "青衣少女，银簪", ExpectedRevision: comicAsset.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Source != "comic" {
		t.Fatalf("updated source=%q, want comic", updated.Source)
	}
}

func TestLegacyManualSourceWithComicPresetIsStillAComicAsset(t *testing.T) {
	asset := normalizeLegacyComicBookAsset(BookAsset{Source: "manual", ExtractionPresetID: "vip-smart-character-derive"})
	if asset.Source != "comic" {
		t.Fatalf("normalized source=%q, want comic", asset.Source)
	}
	ordinary := normalizeLegacyComicBookAsset(BookAsset{Source: "manual", ExtractionPresetID: "batch-assets-h3"})
	if ordinary.Source != "manual" {
		t.Fatalf("ordinary source=%q, want manual", ordinary.Source)
	}
}
