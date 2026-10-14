package batchfactoryv11

import "strings"

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
			assets = append(assets, BookAsset{BatchID: book.BatchID, BookID: book.ID, Kind: kind, Name: name, FormalName: strings.TrimSpace(value.FormalName), Facts: strings.TrimSpace(value.Facts), Appearance: strings.TrimSpace(value.Appearance), Relationships: strings.TrimSpace(value.Relationships), Prompt: prompt, Source: "director", ExtractionPresetID: preset.ID, ExtractionPresetVersion: int64(preset.Version)})
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
