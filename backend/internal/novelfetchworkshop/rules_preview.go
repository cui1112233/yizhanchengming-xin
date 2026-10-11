package novelfetchworkshop

import (
	"regexp"
	"strconv"
	"strings"
)

var defaultChapterLine = regexp.MustCompile(`^\s*(?:0*\d+|第\s*(?:[0-9一二三四五六七八九十百千万零两]+)\s*(?:章|节|回|集|卷|部|篇|话|幕|季|场)|楔子|序章|序言|前言|引子|引言|正文|尾声|终章|后记|番外)\s*$`)

func ruleString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(toStringValue(value)), "<"), ">"))
}

func toStringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func normalizeRuleNewlines(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
}

func repeatedPunctuation(text string) string {
	for _, pair := range [][2]string{{"。。", "。"}, {"，，", "，"}, {"！！", "！"}, {"？？", "？"}, {"、、", "、"}, {"...", "..."}} {
		for strings.Contains(text, pair[0]) && pair[0] != pair[1] {
			text = strings.ReplaceAll(text, pair[0], pair[1])
		}
	}
	return text
}

func ruleBool(layout map[string]any, key string, fallback bool) bool {
	if value, ok := layout[key]; ok {
		if typed, ok := value.(bool); ok {
			return typed
		}
	}
	return fallback
}

func ruleInt(value any, fallback int) int {
	if number, ok := value.(float64); ok {
		return int(number)
	}
	if number, err := strconv.Atoi(ruleString(value)); err == nil {
		return number
	}
	return fallback
}

func previewChapterCleanup(text string, layout, chapter map[string]any) string {
	if !ruleBool(layout, "apply_chapter_cleanup", true) || chapter["chapter_cleanup_enabled"] == false {
		return text
	}
	removeInline := chapter["chapter_remove_inline"] == true
	patterns := []*regexp.Regexp{defaultChapterLine}
	if raw, ok := chapter["chapter_exact_rules"].([]any); ok {
		for _, value := range raw {
			if item, ok := value.(map[string]any); ok && item["enabled"] != false {
				if pattern, err := regexp.Compile(ruleString(item["find"])); err == nil {
					patterns = append(patterns, pattern)
				}
			}
		}
	}
	lines := strings.Split(normalizeRuleNewlines(text), "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		remove := false
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				remove = true
				break
			}
		}
		if !remove && removeInline && strings.HasPrefix(trimmed, "第") {
			if position := strings.IndexAny(trimmed, "：:、-—"); position > 0 {
				remove = defaultChapterLine.MatchString(strings.TrimSpace(trimmed[:position]))
			}
		}
		if remove {
			lines[index] = ""
		}
	}
	return strings.Join(lines, "\n")
}

func previewPairFill(text string, layout map[string]any, symbol map[string]any) string {
	if !ruleBool(layout, "apply_pair_fill", true) {
		return text
	}
	pairs := [][2]string{{"“", "”"}, {"「", "」"}, {"【", "】"}, {"‘", "’"}, {"『", "』"}, {"《", "》"}, {"（", "）"}, {"(", ")"}, {"[", "]"}}
	if raw, ok := symbol["pair_fill_rules"].([]any); ok {
		pairs = pairs[:0]
		for _, value := range raw {
			if item, ok := value.(map[string]any); ok && item["enabled"] != false {
				left, right := ruleString(item["find"]), ruleString(item["replace"])
				if left != "" && right != "" {
					pairs = append(pairs, [2]string{left, right})
				}
			}
		}
	}
	lines := strings.Split(normalizeRuleNewlines(text), "\n")
	for index, line := range lines {
		for _, pair := range pairs {
			left, right := pair[0], pair[1]
			if strings.Contains(line, left) && !strings.Contains(line, right) {
				line += right
			} else if strings.Contains(line, right) && !strings.Contains(line, left) {
				line = left + line
			}
		}
		lines[index] = line
	}
	return strings.Join(lines, "\n")
}

func previewParagraphStyle(text string, layout, layoutRules map[string]any) string {
	trimLines := ruleBool(layout, "trim_lines", true)
	dropEmpty := ruleBool(layout, "drop_empty_lines", true)
	indent := ruleInt(layoutRules["indent_spaces"], 0)
	blank := ruleInt(layoutRules["blank_lines_before"], 0)
	if indent < 0 {
		indent = 0
	}
	if blank < 0 {
		blank = 0
	}
	lines := strings.Split(normalizeRuleNewlines(text), "\n")
	out := make([]string, 0, len(lines))
	first := true
	for _, line := range lines {
		if trimLines {
			line = strings.TrimSpace(line)
		} else {
			line = strings.TrimRight(line, " \t")
		}
		if line == "" {
			if !dropEmpty {
				out = append(out, "")
			}
			continue
		}
		if !first {
			for index := 0; index < blank; index++ {
				out = append(out, "")
			}
		}
		out = append(out, strings.Repeat(" ", indent)+line)
		first = false
	}
	return strings.Trim(strings.Join(out, "\n"), "\n")
}

func ruleTraceStage(name, title, before, after string) map[string]any {
	return map[string]any{"stage": name, "title": title, "chars": len([]rune(after)), "lines": len(strings.Split(normalizeRuleNewlines(after), "\n")), "changed": before != after, "samples": ruleTraceSamples(before, after)}
}

func ruleTraceSamples(before, after string) []map[string]any {
	oldLines, newLines := strings.Split(normalizeRuleNewlines(before), "\n"), strings.Split(normalizeRuleNewlines(after), "\n")
	limit := len(oldLines)
	if len(newLines) > limit {
		limit = len(newLines)
	}
	result := make([]map[string]any, 0, 8)
	for index := 0; index < limit && len(result) < 8; index++ {
		oldLine, newLine := "", ""
		if index < len(oldLines) {
			oldLine = oldLines[index]
		}
		if index < len(newLines) {
			newLine = newLines[index]
		}
		if oldLine != newLine {
			result = append(result, map[string]any{"before": oldLine, "after": newLine})
		}
	}
	return result
}

// PreviewRules mirrors the workbench preview response without calling Node.
func PreviewRules(text, scope string, layout map[string]any, knowledge map[string]any) (string, []map[string]any) {
	if layout == nil {
		layout = map[string]any{}
	}
	if knowledge == nil {
		knowledge = map[string]any{}
	}
	layoutRules, _ := knowledge["layout_rules"].(map[string]any)
	symbolRules, _ := knowledge["symbol_rules"].(map[string]any)
	chapterRules, _ := knowledge["chapter_rules"].(map[string]any)
	if layoutRules == nil {
		layoutRules = map[string]any{}
	}
	if symbolRules == nil {
		symbolRules = map[string]any{}
	}
	if chapterRules == nil {
		chapterRules = map[string]any{}
	}
	result := normalizeRuleNewlines(text)
	stages := []map[string]any{map[string]any{"stage": "raw", "title": "抓取原文/raw", "chars": len([]rune(result)), "lines": len(strings.Split(result, "\n")), "changed": false, "samples": []map[string]any{}}}
	if (scope == "original" && !ruleBool(layout, "apply_to_original", true)) || (scope == "ai" && !ruleBool(layout, "apply_to_ai", true)) {
		return result, stages
	}
	before := result
	result = previewChapterCleanup(result, layout, chapterRules)
	stages = append(stages, ruleTraceStage("chapter_1", "章节清洗-首次", before, result))
	before = result
	result = repeatedPunctuation(result)
	stages = append(stages, ruleTraceStage("punctuation_1", "重复标点清理-首次", before, result))
	before = result
	result = previewPairFill(result, layout, symbolRules)
	stages = append(stages, ruleTraceStage("pair_fill", "成对补符号规则", before, result))
	before = result
	result = repeatedPunctuation(result)
	stages = append(stages, ruleTraceStage("punctuation_2", "重复标点清理-二次", before, result))
	before = result
	result = previewParagraphStyle(result, layout, layoutRules)
	stages = append(stages, ruleTraceStage("paragraph", "批量排版规则", before, result))
	return result, stages
}
