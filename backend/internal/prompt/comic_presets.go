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

// Administrator-authored presets that must exist in every deployment. They are
// seeded by id+version so local administrator edits can still take precedence.
//
//go:embed defaults/admin_presets.json
var adminPresetDefaults []byte

//go:embed defaults/admin_previews/*.jpg
var adminPresetPreviews embed.FS

// EnsureComicPresets adds missing built-in comic presets to the durable
// system catalog. It also performs narrowly scoped upgrades for recognized
// legacy CM records while preserving newer administrator-authored versions.
func EnsureComicPresets(path string, artifactDirs ...string) error {
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
	records, removedDuplicates := dedupePresetVersions(records)
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
	changed := removedRetired || removedDuplicates
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
	var adminSeeds []map[string]any
	if err := json.Unmarshal(adminPresetDefaults, &adminSeeds); err != nil {
		return fmt.Errorf("read built-in administrator presets: %w", err)
	}
	// Rebuild the version indexes after the built-in merge so administrator
	// seeds upgrade against the current catalog state.
	existingVersions = make(map[string]bool, len(records)+len(adminSeeds))
	latestVersion = make(map[string]int, len(records)+len(adminSeeds))
	for _, record := range records {
		id := stringValue(record["id"])
		version := intValue(record["version"])
		if id == "" {
			continue
		}
		existingVersions[presetVersionKey(id, version)] = true
		if version > latestVersion[id] {
			latestVersion[id] = version
		}
	}
	for _, seed := range adminSeeds {
		id := stringValue(seed["id"])
		version := intValue(seed["version"])
		if id == "" || version < 1 || existingVersions[presetVersionKey(id, version)] {
			continue
		}
		// An equal or higher local version belongs to the administrator and
		// must never be downgraded by the embedded seed.
		if latestVersion[id] >= version {
			continue
		}
		for _, record := range records {
			if stringValue(record["id"]) != id || intValue(record["version"]) >= version {
				continue
			}
			if stringValue(record["status"]) == "published" {
				record["status"] = "archived"
			}
		}
		seed["status"] = "published"
		if stringValue(seed["createdBy"]) == "" {
			seed["createdBy"] = "system"
		}
		if stringValue(seed["publishedBy"]) == "" {
			seed["publishedBy"] = "system"
		}
		if stringValue(seed["publishedAt"]) == "" {
			seed["publishedAt"] = "2026-11-04T00:00:00Z"
		}
		records = append(records, seed)
		existingVersions[presetVersionKey(id, version)] = true
		latestVersion[id] = version
		changed = true
	}
	ensureAdminPresetPreviews(adminSeeds, presetArtifactDir(path, artifactDirs))
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

// dedupePresetVersions collapses exact id+version duplicates that may have
// accumulated in older catalogs. It keeps the most useful record instead of
// dropping rows that only differ by status or revision.
func dedupePresetVersions(records []map[string]any) ([]map[string]any, bool) {
	if len(records) < 2 {
		return records, false
	}
	indices := make(map[string]int, len(records))
	result := records[:0]
	changed := false
	for _, record := range records {
		id := stringValue(record["id"])
		if id == "" {
			result = append(result, record)
			continue
		}
		key := presetVersionKey(id, intValue(record["version"]))
		index, found := indices[key]
		if !found {
			indices[key] = len(result)
			result = append(result, record)
			continue
		}
		changed = true
		if preferPresetRecord(record, result[index]) {
			result[index] = record
		}
	}
	return result, changed
}

func preferPresetRecord(candidate, current map[string]any) bool {
	candidatePublished := stringValue(candidate["status"]) == "published"
	currentPublished := stringValue(current["status"]) == "published"
	if candidatePublished != currentPublished {
		return candidatePublished
	}
	return intValue(candidate["revision"]) > intValue(current["revision"])
}

func presetArtifactDir(path string, override []string) string {
	for _, dir := range override {
		if value := strings.TrimSpace(dir); value != "" {
			return value
		}
	}
	return filepath.Join(filepath.Dir(filepath.Dir(path)), "local-executor-artifacts")
}

// ensureAdminPresetPreviews materializes the embedded style previews into the
// local artifact store used by /api/presets/comic-material/{id}/preview. It
// never overwrites an existing artifact, so administrator uploads stay intact.
func ensureAdminPresetPreviews(seeds []map[string]any, artifactDir string) {
	if strings.TrimSpace(artifactDir) == "" {
		return
	}
	for _, seed := range seeds {
		assetID := strings.TrimSpace(stringValue(seed["previewAssetId"]))
		if assetID == "" || filepath.Base(assetID) != assetID {
			continue
		}
		existing := false
		for _, extension := range []string{".png", ".jpg", ".jpeg", ".webp"} {
			if _, err := os.Stat(filepath.Join(artifactDir, assetID+extension)); err == nil {
				existing = true
				break
			}
		}
		if existing {
			continue
		}
		payload, err := adminPresetPreviews.ReadFile("defaults/admin_previews/" + assetID + ".jpg")
		if err != nil {
			continue
		}
		if err := os.MkdirAll(artifactDir, 0o755); err != nil {
			continue
		}
		temporary, err := os.CreateTemp(artifactDir, ".admin-preview-*.tmp")
		if err != nil {
			continue
		}
		_, writeErr := temporary.Write(payload)
		closeErr := temporary.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(temporary.Name())
			continue
		}
		if err := os.Rename(temporary.Name(), filepath.Join(artifactDir, assetID+".jpg")); err != nil {
			_ = os.Remove(temporary.Name())
		}
	}
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
