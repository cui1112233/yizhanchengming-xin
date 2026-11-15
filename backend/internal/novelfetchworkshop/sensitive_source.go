package novelfetchworkshop

// A present catalog, even empty or disabled, owns sensitive rules. Legacy
// layout rules are used only when no sensitive catalog was saved.
func sensitiveRuleSource(settings map[string]any) (any, bool) {
	for _, path := range [][]string{
		{"knowledgeLibraries", "sensitive"}, {"appConfig", "knowledge", "sensitive"},
		{"appConfig", "sensitive"}, {"knowledge", "sensitive"}, {"sensitive"},
	} {
		var value any = settings
		found := true
		for _, key := range path {
			parent, ok := value.(map[string]any)
			if !ok {
				found = false
				break
			}
			value, ok = parent[key]
			if !ok {
				found = false
				break
			}
		}
		if found && value != nil {
			return value, true
		}
	}
	return settings, false
}

// Count saved entries, including disabled entries visible in the editor.
func countSensitiveRuleEntries(source any) int {
	if catalog, ok := source.(map[string]any); ok {
		if groups, exists := catalog["groups"]; exists {
			count := 0
			for _, value := range anySlice(groups) {
				if group, ok := value.(map[string]any); ok {
					count += len(anySlice(group["rules"]))
				}
			}
			return count
		}
		if rules, exists := catalog["rules"]; exists {
			return len(anySlice(rules))
		}
	}
	return len(anySlice(source))
}

// Promote explicitly saved legacy rules before imported defaults are merged.
// Otherwise the newly supplied default sensitive catalog hides the old rules.
func promoteSavedSensitiveCatalog(settings map[string]any) {
	libraries, _ := settings["knowledgeLibraries"].(map[string]any)
	if libraries == nil {
		libraries = map[string]any{}
	}
	if value, exists := libraries["sensitive"]; exists && value != nil {
		return
	}
	source, exists := sensitiveRuleSource(settings)
	if !exists {
		for _, path := range [][]string{
			{"knowledgeLibraries", "layout_rules", "sensitive_rules"}, {"knowledgeLibraries", "layout_rules", "sensitiveKeywords"},
			{"appConfig", "knowledge", "layout_rules", "sensitive_rules"}, {"knowledge", "layout_rules", "sensitive_rules"},
			{"appConfig", "layout", "sensitive_rules"}, {"layout_rules", "sensitive_rules"}, {"sensitive_rules"}, {"sensitiveRules"},
		} {
			var value any = settings
			found := true
			for _, key := range path {
				parent, ok := value.(map[string]any)
				if !ok {
					found = false
					break
				}
				value, ok = parent[key]
				if !ok {
					found = false
					break
				}
			}
			if found && value != nil {
				source = value
				exists = true
				break
			}
		}
	}
	if !exists {
		return
	}
	if catalog, ok := source.(map[string]any); ok {
		libraries["sensitive"] = cloneMap(catalog)
	} else {
		libraries["sensitive"] = map[string]any{"groups": []any{map[string]any{"name": "已有敏感词规则", "enabled": true, "rules": anySlice(source)}}}
	}
	settings["knowledgeLibraries"] = libraries
}
