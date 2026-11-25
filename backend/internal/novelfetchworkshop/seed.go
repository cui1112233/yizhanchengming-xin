package novelfetchworkshop

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

// The imported workbench catalog is bundled with the Go service so a fresh
// user gets the same knowledge and rule libraries as the working V78 system.
//
//go:embed seed/*.json
var importedWorkshopSeed embed.FS

func readImportedSeed(name string) map[string]any {
	raw, err := importedWorkshopSeed.ReadFile("seed/" + name)
	if err != nil {
		panic(fmt.Sprintf("read workshop seed %s: %v", name, err))
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		panic(fmt.Sprintf("decode workshop seed %s: %v", name, err))
	}
	return value
}

func arrayMap(value any) []map[string]any {
	items, _ := value.([]any)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			result = append(result, object)
		}
	}
	return result
}

func appendObjects(values ...any) []any {
	result := make([]any, 0)
	for _, value := range values {
		for _, item := range arrayMap(value) {
			result = append(result, item)
		}
	}
	return result
}

func importedKnowledge() (map[string]any, map[string]any) {
	opening := readImportedSeed("opening_phrases.json")
	high := readImportedSeed("high_imitation.json")
	rewrite := readImportedSeed("rewrite_templates.json")
	layout := readImportedSeed("layout_rules.json")
	symbol := readImportedSeed("symbol_rules.json")
	chapter := readImportedSeed("chapter_rules.json")
	openingAnalyzer := readImportedSeed("opening_analyzer.json")
	sensitive := readImportedSeed("sensitive.json")

	// The old workbench keeps sensitive rules in a separate file, while the
	// summary and preview APIs read them through layout_rules. Keep both views
	// so old and new pages see the same catalog.
	groups, _ := sensitive["groups"].([]any)
	sensitiveRules := make([]any, 0)
	for _, rawGroup := range groups {
		group, _ := rawGroup.(map[string]any)
		for _, rule := range arrayMap(group["rules"]) {
			sensitiveRules = append(sensitiveRules, rule)
		}
	}
	layout["sensitive_rules"] = sensitiveRules

	// The modern API exposes one items list, while the legacy page keeps its
	// named collections. Preserve the original fields and add a normalized list.
	high["items"] = appendObjects(high["prompts"], high["references"])
	rewrite["items"] = appendObjects(rewrite["profiles"], rewrite["temporary_instructions"])
	knowledge := map[string]any{
		"opening_phrases":   opening,
		"opening_analyzer":  openingAnalyzer,
		"high_imitation":    high,
		"rewrite_templates": rewrite,
		// The rules page edits the canonical sensitive catalog directly. Keeping
		// it only as layout_rules.sensitive_rules made the page look empty even
		// though the same rules were present in the imported seed.
		"sensitive":     sensitive,
		"layout_rules":  layout,
		"symbol_rules":  symbol,
		"chapter_rules": chapter,
	}
	return knowledge, sensitive
}

// The bundled catalog is immutable. Parse it once, never cache user settings.
var workshopDefaults = sync.OnceValue(func() map[string]any {
	app := readImportedSeed("app.json")
	knowledge, sensitive := importedKnowledge()
	app["knowledge"] = knowledge
	app["sensitive"] = sensitive
	return map[string]any{
		"appConfig":          app,
		"knowledgeLibraries": knowledge,
	}
})

// DefaultWorkshopSettings returns a private copy of the imported V78 catalog.
// Saved user values still take precedence when defaults are merged.
func DefaultWorkshopSettings() map[string]any {
	return cloneWorkshopValue(workshopDefaults()).(map[string]any)
}

// Clone JSON-shaped trees directly. Fixed catalogs contain thousands of small
// objects; encoding and decoding every subtree wastes CPU on each scheduler read.
// The fallback retains the JSON normalization used for less common input types.
func cloneWorkshopValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		if value == nil {
			return nil
		}
		result := make(map[string]any, len(value))
		for key, item := range value {
			result[key] = cloneWorkshopValue(item)
		}
		return result
	case []any:
		if value == nil {
			return nil
		}
		result := make([]any, len(value))
		for index, item := range value {
			result[index] = cloneWorkshopValue(item)
		}
		return result
	case nil, bool, string, float64:
		return value
	default:
		raw, _ := json.Marshal(value)
		var result any
		_ = json.Unmarshal(raw, &result)
		return result
	}
}

func mergeWorkshopMaps(base, override map[string]any) map[string]any {
	result, _ := cloneWorkshopValue(base).(map[string]any)
	if result == nil {
		result = map[string]any{}
	}
	mergeWorkshopOverrides(result, override)
	return result
}

// result is already a private copy; don't clone every parent again on descent.
func mergeWorkshopOverrides(result, override map[string]any) {
	for key, value := range override {
		// Older accounts may already contain an empty placeholder for a
		// library. Treat that placeholder as missing so the bundled catalog is
		// actually visible after migration.
		if isEmptyWorkshopValue(value) {
			continue
		}
		if nested, ok := value.(map[string]any); ok {
			if previous, ok := result[key].(map[string]any); ok {
				mergeWorkshopOverrides(previous, nested)
			} else {
				result[key] = cloneWorkshopValue(nested)
			}
			continue
		}
		result[key] = value
	}
}

func isEmptyWorkshopValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	case string:
		return typed == ""
	default:
		return false
	}
}

// EnsureWorkshopDefaults adds the imported catalog only where a user has not
// supplied a value. Existing settings always win, including empty values the
// user intentionally saved.
func EnsureWorkshopDefaults(settings map[string]any) map[string]any {
	if settings == nil {
		settings = map[string]any{}
	}
	// Migrate the old appConfig.knowledge view before defaults are merged. This
	// prevents a user's saved library from being hidden by a fresh empty
	// knowledgeLibraries placeholder.
	if libraries, _ := settings["knowledgeLibraries"].(map[string]any); len(libraries) == 0 {
		if app, _ := settings["appConfig"].(map[string]any); app != nil {
			if knowledge, _ := app["knowledge"].(map[string]any); len(knowledge) > 0 {
				settings["knowledgeLibraries"] = cloneWorkshopValue(knowledge)
			}
		}
	}
	if app, _ := settings["appConfig"].(map[string]any); app != nil {
		if knowledge, _ := app["knowledge"].(map[string]any); knowledge != nil {
			if usage, exists := knowledge["usage_prompt"]; exists {
				libraries, _ := settings["knowledgeLibraries"].(map[string]any)
				if libraries == nil {
					libraries = map[string]any{}
				}
				if _, alreadySet := libraries["usage_prompt"]; !alreadySet {
					libraries["usage_prompt"] = usage
				}
				settings["knowledgeLibraries"] = libraries
			}
		}
	}
	promoteSavedSensitiveCatalog(settings)
	merged := mergeWorkshopMaps(workshopDefaults(), settings)
	// Keep the legacy view synchronized for older embedded pages. The Go
	// knowledgeLibraries map remains the canonical persisted source.
	if libraries, _ := merged["knowledgeLibraries"].(map[string]any); len(libraries) > 0 {
		app, _ := merged["appConfig"].(map[string]any)
		if app == nil {
			app = map[string]any{}
		}
		app["knowledge"] = cloneWorkshopValue(libraries)
		merged["appConfig"] = app
	}
	EnsureSensitiveRuleReplacements(merged)
	return merged
}
