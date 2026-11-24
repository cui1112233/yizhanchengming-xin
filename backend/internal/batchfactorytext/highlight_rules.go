package batchfactorytext

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// A selected visual line locates its first visible character, then expands to
// the enclosing dialogue or complete sentence. Source text is never changed.
func sentenceHighlightAt(text string, anchor int, dialogue []HighlightRange) HighlightRange {
	for _, span := range dialogue {
		if anchor >= span.Start && anchor < span.End {
			return span
		}
	}
	rs := []rune(text)
	start, end := 0, len(rs)
	for _, boundary := range SentenceBreaks(text) {
		if boundary <= anchor {
			start = boundary
		} else {
			end = boundary
			break
		}
	}
	for start < end && unicode.IsSpace(rs[start]) {
		start++
	}
	for end > start && unicode.IsSpace(rs[end-1]) {
		end--
	}
	return HighlightRange{start, end}
}

func HighlightRuleEnabled(rules, rule string) bool {
	for _, v := range strings.Split(rules, ",") {
		if strings.TrimSpace(v) == rule {
			return true
		}
	}
	return false
}
func parseHighlightLines(value string) (map[int]bool, error) {
	out := map[int]bool{}
	if strings.TrimSpace(value) == "" {
		return out, nil
	}
	for _, v := range strings.Split(strings.ReplaceAll(value, "，", ","), ",") {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil || n < 1 || n > 1000 {
			return nil, fmt.Errorf("标红行号请填写 1 至 1000 的整数，多个行号用逗号隔开")
		}
		out[n] = true
	}
	return out, nil
}
func validateHighlightRules(c LayoutConfig) error {
	if c.HighlightRules != "" && c.HighlightRules != "off" {
		for _, v := range strings.Split(c.HighlightRules, ",") {
			if v != "dialogue" && v != "lines" && v != "keywords" {
				return fmt.Errorf("不支持的文字标红规则")
			}
		}
	}
	if HighlightRuleEnabled(c.HighlightRules, "lines") {
		if strings.TrimSpace(c.HighlightLines) == "" {
			return fmt.Errorf("请选择需要标红的行号，例如 1,3")
		}
		_, err := parseHighlightLines(c.HighlightLines)
		return err
	}
	return nil
}

// Rune coordinates preserve punctuation and page/source offsets, including nested quotes.
func DialogueHighlightRanges(text string) []HighlightRange {
	rs := []rune(NormalizeLineEndings(text))
	var stack []rune
	var out []HighlightRange
	start := 0
	pairs := map[rune]rune{'“': '”', '「': '」', '『': '』', '"': '"'}
	for i, r := range rs {
		if len(stack) > 0 && r == stack[len(stack)-1] {
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				out = append(out, HighlightRange{start, i + 1})
			}
			continue
		}
		if close, ok := pairs[r]; ok {
			if len(stack) == 0 {
				start = i
			}
			stack = append(stack, close)
		}
	}
	if len(stack) > 0 {
		out = append(out, HighlightRange{start, len(rs)})
	}
	return out
}

func DialoguePageRanges(source, pageText string, start, end int) ([]HighlightRange, bool) {
	rs := []rune(NormalizeLineEndings(source))
	if start < 0 || end < start || end > len(rs) {
		return nil, false
	}
	combined := string(rs[:start]) + pageText + string(rs[end:])
	var ranges []HighlightRange
	for _, v := range DialogueHighlightRanges(combined) {
		a, b := max(0, v.Start-start), min(len([]rune(pageText)), v.End-start)
		if a < b {
			ranges = append(ranges, HighlightRange{a, b})
		}
	}
	return ranges, true
}
