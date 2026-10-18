package prompt

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed defaults/comic_presets.json
var comicPresetDefaults []byte

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
