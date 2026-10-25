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

//go:embed defaults/comic_styles.json
var comicStyleDefaults []byte

//go:embed defaults/comic_style_previews/*.jpg
var comicStylePreviews embed.FS

// EnsureComicPresets adds only missing built-in comic presets to the durable
// system catalog. Existing system-preset versions remain fully editable and
// are never replaced by a later application start.
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
	var defaults []map[string]any
	if err := json.Unmarshal(comicPresetDefaults, &defaults); err != nil {
		return fmt.Errorf("read built-in comic presets: %w", err)
	}
	var styles []map[string]any
	if err := json.Unmarshal(comicStyleDefaults, &styles); err != nil {
		return fmt.Errorf("read built-in comic styles: %w", err)
	}
	defaults = append(defaults, styles...)
	existing := make(map[string]bool, len(records))
	for _, record := range records {
		if id := stringValue(record["id"]); id != "" {
			existing[id] = true
		}
	}
	changed := false
	for _, record := range defaults {
		id := stringValue(record["id"])
		if id == "" || existing[id] {
			continue
		}
		record["createdAt"] = "2026-10-18T00:00:00Z"
		record["createdBy"] = "system"
		record["publishedAt"] = "2026-10-18T00:00:00Z"
		record["publishedBy"] = "system"
		records = append(records, record)
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

// OpenComicStylePreview returns only the display image paired with a built-in
// style preset. Selecting a style never adds this image to generation inputs.
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
