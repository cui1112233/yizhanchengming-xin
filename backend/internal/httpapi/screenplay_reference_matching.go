package httpapi

import (
	"strings"

	"qiantie/backend/internal/screenplay"
)

// The current storyboard format stores names as text. Resolve only whole names
// in the corresponding field; a partial or ambiguous name must not pick an image.
func screenplayReferenceTargets(shot screenplay.StoryboardShot, assets []screenplay.Asset) []string {
	requested := map[screenplay.AssetType]map[string]bool{}
	for kind, field := range map[screenplay.AssetType]string{
		screenplay.AssetCharacter: shot.Characters,
		screenplay.AssetScene:     shot.Scene,
		screenplay.AssetProp:      shot.Props,
	} {
		requested[kind] = map[string]bool{}
		// Keep the separators accepted by storyboard splitAssetNames. Spaces
		// inside a name (for example Alice Smith) are not list separators.
		for _, part := range strings.FieldsFunc(field, func(r rune) bool {
			return strings.ContainsRune(",，、;；/|\n@", r)
		}) {
			name := strings.TrimSpace(part)
			if name != "" && name != "无" && name != "无道具" && name != "none" && name != "-" {
				requested[kind][name] = true
			}
		}
	}
	type assetName struct {
		kind screenplay.AssetType
		name string
	}
	counts := map[assetName]int{}
	for _, asset := range assets {
		counts[assetName{asset.Type, strings.TrimSpace(asset.Name)}]++
	}
	targets := []string{}
	if shot.ID != "" {
		targets = append(targets, shot.ID)
	}
	for _, asset := range assets {
		name := strings.TrimSpace(asset.Name)
		if requested[asset.Type][name] && counts[assetName{asset.Type, name}] == 1 {
			targets = append(targets, asset.ID)
		}
	}
	return targets
}
