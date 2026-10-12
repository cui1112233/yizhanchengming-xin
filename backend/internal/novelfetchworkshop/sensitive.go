package novelfetchworkshop

import (
	"fmt"
	"strings"
)

// ApplySensitiveRules applies the saved sensitive-word replacements to the
// working text. It never touches Document.OriginalRaw; callers decide which
// processed field to save. Both the current map form and the legacy
// "@{enabled=True; find=x; replace=y}" form are accepted.
func ApplySensitiveRules(text string, settings map[string]any) (string, int) {
	result, changes, _ := ApplySensitiveRulesWithDetails(text, settings)
	return result, changes
}

// ApplySensitiveRulesWithDetails applies the same deterministic replacements
// as ApplySensitiveRules and also returns the before/after line for every hit.
// The detail records are used by the task detail page and are safe to persist
// without exposing the untouched raw document.
func ApplySensitiveRulesWithDetails(text string, settings map[string]any) (string, int, []map[string]any) {
	if text == "" || settings == nil {
		return text, 0, []map[string]any{}
	}
	rules := make([]any, 0)
	collectSensitiveRules(settings, &rules)
	result := text
	changes := 0
	items := make([]map[string]any, 0)
	for _, raw := range rules {
		find, replace, enabled := sensitiveRule(raw)
		if !enabled || find == "" || find == replace {
			continue
		}
		for {
			index := strings.Index(result, find)
			if index < 0 {
				break
			}
			before := sensitiveLine(result, index)
			result = result[:index] + replace + result[index+len(find):]
			after := sensitiveLine(result, index)
			changes++
			items = append(items, map[string]any{
				"hit_index": len(items) + 1,
				"keyword":   find,
				"before":    before,
				"after":     after,
				"status":    "done",
				"source":    "system_rule",
				"error":     "",
			})
		}
	}
	return result, changes, items
}

func sensitiveLine(text string, index int) string {
	start := strings.LastIndex(text[:index], "\n") + 1
	end := strings.IndexByte(text[index:], '\n')
	if end < 0 {
		end = len(text)
	} else {
		end += index
	}
	return text[start:end]
}

func collectSensitiveRules(value any, out *[]any) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"sensitive_rules", "sensitiveRules", "rules"} {
			if raw, ok := typed[key]; ok {
				if key == "rules" {
					// Only groups under a sensitive section are meaningful; plain
					// knowledge item rules are handled by their own features.
					if _, hasSensitive := typed["groups"]; !hasSensitive {
						continue
					}
				}
				collectSensitiveRules(raw, out)
			}
		}
		if groups, ok := typed["groups"]; ok {
			collectSensitiveRules(groups, out)
		}
		for _, key := range []string{"knowledge", "app_config", "appConfig", "layout_rules", "layout", "sensitive"} {
			if nested, ok := typed[key]; ok {
				collectSensitiveRules(nested, out)
			}
		}
	case []any:
		*out = append(*out, typed...)
	case []map[string]any:
		for _, item := range typed {
			*out = append(*out, item)
		}
	}
}

func sensitiveRule(raw any) (string, string, bool) {
	if item, ok := raw.(map[string]any); ok {
		find := strings.TrimSpace(stringValue(item["find"], item["pattern"], item["source"]))
		replace := stringValue(item["replace"], item["replacement"], item["target"])
		enabled := true
		if value, exists := item["enabled"]; exists {
			enabled = boolValue(value)
		}
		return find, replace, enabled
	}
	value := strings.TrimSpace(stringValue(raw))
	if !strings.HasPrefix(value, "@{") || !strings.HasSuffix(value, "}") {
		return "", "", false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(value, "@{"), "}")
	fields := map[string]string{}
	for _, part := range strings.Split(body, ";") {
		pieces := strings.SplitN(part, "=", 2)
		if len(pieces) == 2 {
			fields[strings.ToLower(strings.TrimSpace(pieces[0]))] = strings.TrimSpace(pieces[1])
		}
	}
	enabled := fields["enabled"] == "" || strings.EqualFold(fields["enabled"], "true") || fields["enabled"] == "1"
	return fields["find"], fields["replace"], enabled
}

func stringValue(values ...any) string {
	for _, value := range values {
		if value == nil {
			continue
		}
		if text, ok := value.(string); ok {
			if strings.TrimSpace(text) != "" {
				return text
			}
			continue
		}
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func boolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || strings.TrimSpace(typed) == "1"
	default:
		return false
	}
}
