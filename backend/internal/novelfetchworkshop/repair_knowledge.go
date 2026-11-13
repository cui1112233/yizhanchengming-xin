package novelfetchworkshop

import (
	"fmt"
	"time"
)

type ImportedKnowledgeRepairReport struct {
	Repaired  int            `json:"repaired"`
	Libraries map[string]int `json:"libraries"`
}

// A lost character can be '?' or the Unicode replacement character. Require
// identical positions, length and surviving characters; never infer prose.
func degradedImportedText(actual, original string) bool {
	a, b := []rune(actual), []rune(original)
	if actual == original || len(a) != len(b) {
		return false
	}
	damaged := false
	for i, clean := range b {
		if a[i] == clean {
			continue
		}
		if clean <= 127 || (a[i] != '?' && a[i] != '\uFFFD') {
			return false
		}
		damaged = true
	}
	return damaged
}

func restoreImportedValue(actual, original any) (any, int) {
	switch current := actual.(type) {
	case string:
		if clean, ok := original.(string); ok && degradedImportedText(current, clean) {
			return clean, 1
		}
	case map[string]any:
		clean, ok := original.(map[string]any)
		if !ok {
			return actual, 0
		}
		if id, exists := clean["id"]; exists && fmt.Sprint(current["id"]) != fmt.Sprint(id) {
			return actual, 0
		}
		n := 0
		for key, value := range current {
			switch key {
			case "id", "source", "imported_at", "updated_at":
				continue
			}
			if cleanValue, exists := clean[key]; exists {
				fixed, count := restoreImportedValue(value, cleanValue)
				current[key] = fixed
				n += count
			}
		}
		return current, n
	case []any:
		clean, ok := original.([]any)
		if !ok || len(current) != len(clean) {
			return actual, 0
		}
		n := 0
		for i := range current {
			fixed, count := restoreImportedValue(current[i], clean[i])
			current[i] = fixed
			n += count
		}
		return current, n
	}
	return actual, 0
}

// Repair only known imported libraries. Existing custom fields, switches and
// unmatched text survive. Dry run works on copies and never changes settings.
func RepairImportedKnowledge(settings map[string]any, apply bool) ImportedKnowledgeRepairReport {
	seeds, _ := importedKnowledge()
	report := ImportedKnowledgeRepairReport{Libraries: map[string]int{}}
	backup := map[string]any{}
	for _, kind := range []string{"high_imitation", "rewrite_templates", "opening_phrases", "symbol_rules", "chapter_rules", "layout_rules"} {
		current := knowledgeLibrary(settings, kind)
		seed, _ := seeds[kind].(map[string]any)
		if seed == nil || seed["source"] == nil || current["source"] != seed["source"] {
			continue
		}
		candidate := cloneMap(current)
		_, count := restoreImportedValue(candidate, seed)
		// A single intentional '?' in a custom rule is not proof of a damaged
		// import. Require several matching damaged fields from the known catalog.
		if count < 3 {
			continue
		}
		report.Libraries[kind] = count
		report.Repaired += count
		if apply {
			backup[kind] = cloneMap(current)
			putKnowledgeLibrary(settings, kind, candidate)
		}
	}
	if apply && report.Repaired > 0 && settings["importedKnowledgeRepairBackup"] == nil {
		settings["importedKnowledgeRepairBackup"] = map[string]any{"at": time.Now().UTC().Format(time.RFC3339Nano), "libraries": backup}
	}
	return report
}
