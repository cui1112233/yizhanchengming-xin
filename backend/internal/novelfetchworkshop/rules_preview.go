package novelfetchworkshop

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
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
	removeStandalone := chapter["chapter_remove_standalone"] != false
	patterns := []*regexp.Regexp{defaultChapterLine}
	for _, key := range []string{"chapter_exact_rules", "chapter_builtin_exact_rules"} {
		if raw, ok := chapter[key].([]any); ok {
			for _, value := range raw {
				if item, ok := value.(map[string]any); ok && item["enabled"] != false {
					if pattern, err := regexp.Compile(ruleString(item["find"])); err == nil {
						patterns = append(patterns, pattern)
					}
				}
			}
		}
	}
	inlinePatterns := []*regexp.Regexp{}
	for _, key := range []string{"chapter_inline_rules", "chapter_builtin_inline_rules"} {
		for _, value := range anySlice(chapter[key]) {
			if item, ok := value.(map[string]any); ok && item["enabled"] != false {
				if pattern, err := regexp.Compile(ruleString(item["find"])); err == nil {
					inlinePatterns = append(inlinePatterns, pattern)
				}
			}
		}
	}
	lines := strings.Split(normalizeRuleNewlines(text), "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		remove := false
		if removeStandalone {
			for _, pattern := range patterns {
				if match := pattern.FindStringIndex(line); match != nil && match[0] == 0 {
					remove = true
					break
				}
			}
		}
		if !remove && removeInline {
			for _, pattern := range inlinePatterns {
				if match := pattern.FindStringIndex(line); match != nil && match[0] == 0 {
					remove = true
					break
				}
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

func textHasRuleContent(text string) bool {
	for _, char := range text {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '_' {
			return true
		}
	}
	return false
}

func symbolAttachesRight(line []rune, index int, symbol rune) bool {
	if strings.ContainsRune("“‘「『【《〈（([", symbol) {
		return true
	}
	if !strings.ContainsRune("”’」』", symbol) {
		return false
	}
	previous := rune(0)
	if index > 0 {
		previous = line[index-1]
	}
	return (index == 0 || strings.ContainsRune(" \t:：,，。！？?；;、", previous)) && textHasRuleContent(string(line[index+1:]))
}

func parseRulePositions(value any) []int {
	parts := strings.FieldsFunc(ruleString(value), func(char rune) bool { return char == ',' || char == '，' || unicode.IsSpace(char) })
	positions := make([]int, 0, len(parts))
	for _, part := range parts {
		if position, err := strconv.Atoi(part); err == nil && position > 0 {
			positions = append(positions, position)
		}
	}
	return positions
}

func previewSymbolTrim(text string, layout, symbol map[string]any) string {
	if !ruleBool(layout, "apply_symbol_rules", true) || !ruleBool(symbol, "enable_symbol_trim", false) {
		return text
	}
	positions := parseRulePositions(symbol["symbol_trim_positions"])
	if len(positions) == 0 {
		return text
	}
	triggers := map[rune]bool{}
	for _, raw := range anySlice(symbol["symbol_rules"]) {
		item, ok := raw.(map[string]any)
		if !ok || item["enabled"] == false {
			continue
		}
		chars := []rune(ruleString(item["find"]))
		if len(chars) == 1 {
			triggers[chars[0]] = true
		}
	}
	if len(triggers) == 0 {
		return text
	}
	allowAdjacent := ruleBool(symbol, "symbol_trim_allow_adjacent", false)
	out := make([]string, 0)
	for _, line := range strings.Split(normalizeRuleNewlines(text), "\n") {
		queue := []string{line}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			chars := []rune(current)
			matched := -1
			for _, position := range positions {
				indexes := []int{position - 1}
				if allowAdjacent {
					indexes = []int{position - 2, position - 1, position}
				}
				for _, index := range indexes {
					if index >= 0 && index < len(chars) && triggers[chars[index]] {
						matched = index
						break
					}
				}
				if matched >= 0 {
					break
				}
			}
			if matched < 0 {
				out = append(out, current)
				continue
			}
			leftEnd, rightStart := matched+1, matched+1
			if symbolAttachesRight(chars, matched, chars[matched]) {
				leftEnd, rightStart = matched, matched
			}
			left := strings.TrimSpace(string(chars[:leftEnd]))
			right := strings.TrimSpace(string(chars[rightStart:]))
			if textHasRuleContent(left) {
				out = append(out, left)
			}
			if textHasRuleContent(right) {
				queue = append([]string{right}, queue...)
			}
		}
	}
	return strings.Join(out, "\n")
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
	result = previewSymbolTrim(result, layout, symbolRules)
	stages = append(stages, ruleTraceStage("symbol_trim", "字位符号规则", before, result))
	before = result
	result = previewPairFill(result, layout, symbolRules)
	stages = append(stages, ruleTraceStage("pair_fill", "成对补符号规则", before, result))
	before = result
	result = repeatedPunctuation(result)
	stages = append(stages, ruleTraceStage("punctuation_2", "重复标点清理-二次", before, result))
	before = result
	result = previewChapterCleanup(result, layout, chapterRules)
	stages = append(stages, ruleTraceStage("chapter_2", "章节清洗-二次", before, result))
	before = result
	result = previewParagraphStyle(result, layout, layoutRules)
	stages = append(stages, ruleTraceStage("paragraph", "批量排版规则", before, result))
	return result, stages
}

// ProcessConfiguredText is the single rule entry point used by previews,
// fetched originals and generated AI versions.
func ProcessConfiguredText(text, scope string, settings map[string]any) (string, []map[string]any, []map[string]any) {
	if settings == nil {
		settings = map[string]any{}
	}
	app := mapOrEmpty(settings["appConfig"])
	layout := mapOrEmpty(settings["layout"])
	if configured := mapOrEmpty(app["layout"]); len(configured) > 0 {
		layout = configured
	}
	knowledge := mapOrEmpty(settings["knowledgeLibraries"])
	if len(knowledge) == 0 {
		knowledge = mapOrEmpty(app["knowledge"])
	}
	raw := normalizeRuleNewlines(text)
	trace := []map[string]any{{"stage": "raw", "title": "抓取原文/raw", "chars": len([]rune(raw)), "lines": len(strings.Split(raw, "\n")), "changed": false, "samples": []map[string]any{}}}
	if (scope == "original" && !ruleBool(layout, "apply_to_original", true)) || (scope == "ai" && !ruleBool(layout, "apply_to_ai", true)) {
		return raw, trace, []map[string]any{}
	}
	processed := raw
	hits := []map[string]any{}
	if ruleBool(layout, "apply_sensitive", true) {
		before := processed
		processed, _, hits = ApplySensitiveRulesForScopeWithDetails(processed, settings, scope)
		trace = append(trace, ruleTraceStage("sensitive", "敏感词替换", before, processed))
	}
	processed, stages := PreviewRules(processed, scope, layout, knowledge)
	if len(stages) > 1 {
		trace = append(trace, stages[1:]...)
	}
	return processed, trace, hits
}
