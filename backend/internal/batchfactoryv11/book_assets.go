package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func directorBookAssets(book Book, snapshot DirectorSnapshot, output DirectorResult) []BookAsset {
	config := aiReasoningPromptConfig(snapshot.Effective).Assets
	assets := make([]BookAsset, 0, len(output.Characters)+len(output.Scenes)+len(output.Props))
	seen := map[string]bool{}
	appendAssets := func(kind string, values []NamedPrompt, preset PresetSnapshot) {
		for _, value := range values {
			name, prompt := strings.TrimSpace(value.Name), strings.TrimSpace(value.Prompt)
			key := kind + "\x00" + name
			if name == "" || prompt == "" || seen[key] {
				continue
			}
			seen[key] = true
			assets = append(assets, BookAsset{BatchID: book.BatchID, BookID: book.ID, Kind: kind, Name: name, FormalName: strings.TrimSpace(value.FormalName), Facts: strings.TrimSpace(value.Facts), Appearance: strings.TrimSpace(value.Appearance), Relationships: strings.TrimSpace(value.Relationships), Prompt: mergeVisualAssetPrompt(value.Facts, value.Appearance, prompt, value.Relationships), Source: "director", ExtractionPresetID: preset.ID, ExtractionPresetVersion: int64(preset.Version)})
		}
	}
	characterPreset := config.Character
	if characterPreset.ID == "" {
		characterPreset = config.Extraction
	}
	scenePreset := config.Scene
	if scenePreset.ID == "" {
		scenePreset = config.Extraction
	}
	appendAssets("character", output.Characters, characterPreset)
	appendAssets("scene", output.Scenes, scenePreset)
	appendAssets("prop", output.Props, config.Extraction)
	return assets
}

// AssetPromptProofKind is reserved for internal asset writes. Each asset keeps one
// record binding its exact prompt input to the production version that created it.
const AssetPromptProofKind = "asset_prompt_version"

func assetPromptProof(asset BookAsset, version string) Draft {
	appearance := asset.Appearance
	if appearance == "" {
		appearance = asset.Prompt
	}
	fields, _ := json.Marshal([]string{version, asset.BatchID, asset.BookID, asset.ID, asset.Kind, asset.Name, asset.FormalName, asset.Facts, appearance, asset.Relationships, asset.Prompt})
	return Draft{Key: asset.ID, Scope: asset.BatchID, Kind: AssetPromptProofKind, Content: sourceDigest(string(fields))}
}
func CheckAssetPromptVersions(ctx context.Context, store Store, owner string, book Book, assetIDs []string) error {
	version := ProductionVersion(book)
	assets := map[string]BookAsset{}
	for _, asset := range book.AssetRecords {
		assets[asset.ID] = asset
	}
	for _, id := range assetIDs {
		asset, ok := assets[id]
		if !ok {
			return ErrNotFound
		}
		proof := assetPromptProof(asset, version)
		d, err := store.GetDraft(ctx, owner, proof.Key, proof.Kind, proof.Scope)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if version == "" || err != nil || d.Content != proof.Content {
			return NewConfigRequiredError("资产「" + asset.Name + "」提示词的正文版本已变化，请重新提取资产或编辑确认提示词")
		}
	}
	return nil
}
