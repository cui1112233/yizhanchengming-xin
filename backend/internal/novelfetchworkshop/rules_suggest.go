package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/prompt/batchbook"
	"regexp"
	"strings"
)

func ruleSuggestionSchema(kind string) (map[string]any, error) {
	rule := map[string]any{"find": "实际命中内容", "replace": "替换内容", "enabled": true}
	switch kind {
	case "sensitive":
		return map[string]any{"groups": []any{map[string]any{"name": "AI建议", "enabled": true, "apply_to_original": true, "apply_to_ai": true, "rules": []any{rule}}}}, nil
	case "layout_rules":
		return map[string]any{"indent_spaces": 0, "blank_lines_before": 0, "drop_empty_lines": true, "trim_lines": true, "apply_to_original": true, "apply_to_ai": true, "apply_sensitive": true, "apply_chapter_cleanup": true, "apply_symbol_rules": true, "apply_pair_fill": true}, nil
	case "chapter_exact_rules", "chapter_inline_rules":
		return map[string]any{kind: []any{map[string]any{"find": "^第[0-9一二三四五六七八九十]+章", "replace": "", "enabled": true}}, "chapter_cleanup_enabled": true, "chapter_remove_standalone": true, "chapter_remove_inline": kind == "chapter_inline_rules"}, nil
	case "symbol_rules":
		return map[string]any{kind: []any{map[string]any{"find": "，", "replace": "", "enabled": true}}, "enable_symbol_trim": true, "symbol_trim_positions": "1,2", "symbol_trim_allow_adjacent": false}, nil
	case "pair_fill_rules":
		return map[string]any{kind: []any{map[string]any{"find": "“", "replace": "”", "enabled": true}}}, nil
	default:
		return nil, errors.New("不支持的规则类型")
	}
}

// SuggestRules returns only configuration fields understood by the editor and processing engine.
func SuggestRules(ctx context.Context, provider RewriteProvider, kind, goal, sample string) (map[string]any, error) {
	schema, err := ruleSuggestionSchema(kind)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, errors.New("AI 改文模型未配置")
	}
	encoded, _ := json.Marshal(schema)
	content, err := provider.Complete(ctx, RewriteRequest{SystemPrompt: batchbook.RulesSystem, UserPrompt: fmt.Sprintf(batchbook.RulesUser, kind, goal, string(encoded), sample), Temperature: 0.2, MaxTokens: 2000})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err = json.Unmarshal([]byte(strings.TrimSpace(content)), &result); err != nil {
		return nil, errors.New("AI 返回的规则不是有效 JSON，请重新生成")
	}
	if err = validateRuleSuggestions(kind, result, schema); err != nil {
		return nil, err
	}
	return result, nil
}

func validateRuleSuggestions(kind string, result, schema map[string]any) error {
	bad := func() error { return errors.New("AI 返回了不能执行的规则，请调整要求后重新生成") }
	if len(result) == 0 {
		return bad()
	}
	for key, value := range result {
		expected, ok := schema[key]
		if !ok {
			return bad()
		}
		switch expected.(type) {
		case bool:
			if _, ok := value.(bool); !ok {
				return bad()
			}
		case int:
			n, ok := value.(float64)
			if !ok || n < 0 || n > 20 || n != float64(int(n)) {
				return bad()
			}
		case string:
			if _, ok := value.(string); !ok {
				return bad()
			}
		}
	}
	validateRows := func(raw any) error {
		rows, ok := raw.([]any)
		if !ok || len(rows) == 0 {
			return bad()
		}
		for _, raw := range rows {
			item, ok := raw.(map[string]any)
			if !ok {
				return bad()
			}
			for key := range item {
				if key != "find" && key != "replace" && key != "enabled" {
					return bad()
				}
			}
			find, ok := item["find"].(string)
			if !ok || strings.TrimSpace(find) == "" {
				return bad()
			}
			replace, ok := item["replace"].(string)
			if !ok {
				return bad()
			}
			if enabled, exists := item["enabled"]; exists {
				if _, ok := enabled.(bool); !ok {
					return bad()
				}
			}
			if strings.HasPrefix(kind, "chapter_") {
				if replace != "" {
					return bad()
				}
				if _, err := regexp.Compile(find); err != nil {
					return bad()
				}
			}
			if kind == "symbol_rules" && len([]rune(find)) != 1 {
				return bad()
			}
			if kind == "pair_fill_rules" && strings.TrimSpace(replace) == "" {
				return bad()
			}
		}
		return nil
	}
	if kind == "sensitive" {
		groups, ok := result["groups"].([]any)
		if !ok || len(groups) == 0 {
			return bad()
		}
		for _, raw := range groups {
			group, ok := raw.(map[string]any)
			if !ok {
				return bad()
			}
			for key, value := range group {
				switch key {
				case "name":
					if _, ok := value.(string); !ok {
						return bad()
					}
				case "enabled", "apply_to_original", "apply_to_ai":
					if _, ok := value.(bool); !ok {
						return bad()
					}
				case "rules":
				default:
					return bad()
				}
			}
			if err := validateRows(group["rules"]); err != nil {
				return err
			}
		}
	} else if kind != "layout_rules" {
		if err := validateRows(result[kind]); err != nil {
			return err
		}
	}
	if kind == "symbol_rules" {
		if result["enable_symbol_trim"] != true || len(parseRulePositions(result["symbol_trim_positions"])) == 0 {
			return bad()
		}
	}
	if kind == "chapter_inline_rules" && result["chapter_remove_inline"] != true {
		return bad()
	}
	return nil
}
