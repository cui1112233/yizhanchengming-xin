package novelfetchworkshop

import (
	"fmt"
	"strings"
	"time"
)

func questionMask(text string) string {
	return strings.Map(func(r rune) rune {
		if r > 127 {
			return '?'
		}
		return r
	}, text)
}

func onlyQuestions(text string) bool { return text != "" && strings.Trim(text, "?") == "" }

// Repair only a complete, ordered match to the bundled imported catalog. A
// changed/custom/reordered group is left untouched rather than guessed at.
func RepairImportedSensitive(settings map[string]any) (int, error) {
	library := knowledgeLibrary(settings, "sensitive")
	original := cloneMap(library)
	groups := arrayMap(library["groups"])
	seedGroups := arrayMap(readImportedSeed("sensitive.json")["groups"])
	repaired := 0
	for index, group := range groups {
		if index >= len(seedGroups) {
			continue
		}
		seed := seedGroups[index]
		name, seedName := fmt.Sprint(group["name"]), fmt.Sprint(seed["name"])
		if name != seedName && name != questionMask(seedName) {
			continue
		}
		rows, seedRows := arrayMap(group["rules"]), arrayMap(seed["rules"])
		if len(rows) != len(seedRows) {
			continue
		}
		matched, anchors := true, 0
		for i, row := range rows {
			actual, clean := fmt.Sprint(row["find"]), fmt.Sprint(seedRows[i]["find"])
			if actual != clean && actual != questionMask(clean) {
				matched = false
				break
			}
			if clean == questionMask(clean) && strings.Trim(clean, "?") != "" {
				anchors++
			}
		}
		if !matched || (len(rows) > 0 && anchors == 0) {
			continue
		}
		if name != seedName {
			group["name"] = seedName
		}
		for i, row := range rows {
			changed := false
			if clean := fmt.Sprint(seedRows[i]["find"]); fmt.Sprint(row["find"]) != clean {
				row["find"] = clean
				changed = true
			}
			if replacement, _ := row["replace"].(string); onlyQuestions(replacement) {
				row["replace"] = seedRows[i]["replace"]
				changed = true
			}
			if changed {
				repaired++
			}
		}
	}
	if repaired == 0 {
		return 0, nil
	}
	// Keep the original damaged catalog, without model keys or other settings.
	if settings["importedSensitiveRepairBackup"] == nil {
		settings["importedSensitiveRepairBackup"] = map[string]any{"at": time.Now().UTC().Format(time.RFC3339Nano), "sensitive": original}
	}
	putKnowledgeLibrary(settings, "sensitive", library)
	EnsureSensitiveRuleReplacements(settings)
	return repaired, nil
}
