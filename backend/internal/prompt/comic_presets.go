package prompt

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed defaults/comic_presets.json
var comicPresetDefaults []byte

//go:embed defaults/cm_asset_presets.json
var cmAssetPresetDefaults []byte

//go:embed defaults/comic_styles.json
var comicStyleDefaults []byte

//go:embed defaults/comic_reference_scenes.json
var comicReferenceSceneDefaults []byte

//go:embed defaults/comic_style_previews/*.jpg
var comicStylePreviews embed.FS

//go:embed defaults/comic_reference_scenes/*.jpg
var comicReferenceScenes embed.FS

// EnsureComicPresets adds missing built-in comic presets to the durable
// system catalog. It also performs narrowly scoped upgrades for recognized
// legacy CM records while preserving newer administrator-authored versions.
func EnsureComicPresets(path string) error {
	path = filepath.Clean(path)
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read system prompt catalog: %w", err)
	}
	var records []map[string]any
	if err := json.Unmarshal(raw, &records); err != nil {
		return fmt.Errorf("read system prompt catalog: %w", err)
	}
	filtered := records[:0]
	removedRetired := false
	for _, record := range records {
		if stringValue(record["id"]) == "batch-comic-bootstrap" && stringValue(record["createdBy"]) == "system" {
			removedRetired = true
			continue
		}
		filtered = append(filtered, record)
	}
	records = filtered
	var defaults []map[string]any
	if err := json.Unmarshal(comicPresetDefaults, &defaults); err != nil {
		return fmt.Errorf("read built-in comic presets: %w", err)
	}
	var styles []map[string]any
	if err := json.Unmarshal(comicStyleDefaults, &styles); err != nil {
		return fmt.Errorf("read built-in comic styles: %w", err)
	}
	var cmAssets []map[string]any
	if err := json.Unmarshal(cmAssetPresetDefaults, &cmAssets); err != nil {
		return fmt.Errorf("read built-in CM asset presets: %w", err)
	}
	var referenceScenes []map[string]any
	if err := json.Unmarshal(comicReferenceSceneDefaults, &referenceScenes); err != nil {
		return fmt.Errorf("read built-in comic reference scenes: %w", err)
	}
	defaults = append(defaults, styles...)
	defaults = append(defaults, cmAssets...)
	defaults = append(defaults, referenceScenes...)
	existingVersions := make(map[string]bool, len(records))
	latestVersion := make(map[string]int, len(records))
	publishedIndex := make(map[string]int, len(records))
	for index, record := range records {
		id := stringValue(record["id"])
		version := intValue(record["version"])
		if id == "" {
			continue
		}
		existingVersions[presetVersionKey(id, version)] = true
		if version > latestVersion[id] {
			latestVersion[id] = version
		}
		if stringValue(record["status"]) == "published" {
			publishedIndex[id] = index
		}
	}
	changed := removedRetired
	for _, record := range defaults {
		id := stringValue(record["id"])
		version := intValue(record["version"])
		if id == "" || version < 1 || existingVersions[presetVersionKey(id, version)] {
			continue
		}
		if latestVersion[id] > 0 {
			legacyMaxVersion, isCMUpgrade := cmLegacySystemMaxVersions[id]
			_, forceLegacyPageUpgrade := cmForcedLegacyPageUpgradeIDs[id]
			index, hasPublished := publishedIndex[id]
			if !isCMUpgrade || latestVersion[id] > legacyMaxVersion || !hasPublished || intValue(records[index]["version"]) > legacyMaxVersion || (!forceLegacyPageUpgrade && stringValue(records[index]["createdBy"]) != "system") {
				continue
			}
			records[index]["status"] = "archived"
			changed = true
		}
		record["createdAt"] = "2026-10-18T00:00:00Z"
		record["createdBy"] = "system"
		record["publishedAt"] = "2026-10-18T00:00:00Z"
		record["publishedBy"] = "system"
		records = append(records, record)
		existingVersions[presetVersionKey(id, version)] = true
		latestVersion[id] = version
		changed = true
	}
	if !changed {
		return nil
	}
	encoded, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("encode system prompt catalog: %w", err)
	}
	temporary := path + ".comic-presets.tmp"
	if err := os.WriteFile(temporary, encoded, 0o600); err != nil {
		return fmt.Errorf("write system prompt catalog: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("replace system prompt catalog: %w", err)
	}
	return nil
}

var cmLegacySystemMaxVersions = map[string]int{
	"batch-comic-assets":             2,
	"vip-smart-character-derive":     1,
	"batch-comic-page":               2,
	"batch-comic-page-fusion-bw":     2,
	"batch-comic-page-light":         2,
	"batch-comic-page-premium-strip": 2,
}

var cmForcedLegacyPageUpgradeIDs = map[string]struct{}{
	"batch-comic-page":               {},
	"batch-comic-page-fusion-bw":     {},
	"batch-comic-page-light":         {},
	"batch-comic-page-premium-strip": {},
}

func presetVersionKey(id string, version int) string {
	return fmt.Sprintf("%s\x00%d", id, version)
}

// OpenComicStylePreview returns the embedded image paired with a built-in
// style preset. It is used for both display and the optional global style reference.
func OpenComicStylePreview(presetID string) ([]byte, string, bool) {
	presetID = strings.TrimSpace(presetID)
	const prefix = "batch-comic-style-"
	if !strings.HasPrefix(presetID, prefix) || len(presetID) != len(prefix)+2 {
		return nil, "", false
	}
	name := strings.TrimPrefix(presetID, prefix) + ".jpg"
	data, err := comicStylePreviews.ReadFile("defaults/comic_style_previews/" + name)
	if err != nil {
		return nil, "", false
	}
	return data, "image/jpeg", true
}

// OpenComicReferenceScene returns the real scene image sent to the image model.
func OpenComicReferenceScene(presetID string) ([]byte, string, bool) {
	presetID = strings.TrimSpace(presetID)
	const prefix = "batch-comic-reference-scene-"
	if !strings.HasPrefix(presetID, prefix) {
		return nil, "", false
	}
	name := strings.TrimPrefix(presetID, prefix) + ".jpg"
	data, err := comicReferenceScenes.ReadFile("defaults/comic_reference_scenes/" + name)
	if err != nil {
		return nil, "", false
	}
	return data, "image/jpeg", true
}
