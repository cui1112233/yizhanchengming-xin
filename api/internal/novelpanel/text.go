package novelpanel

import "strings"

func ProcessText(input string, settings TextProcessingSettings) string {
	value := strings.ReplaceAll(input, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	out := make([]string, 0, len(lines))
	blankPending := false
	for _, line := range lines {
		if settings.NormalizeFullWidthSpaces {
			line = strings.ReplaceAll(line, "\u3000", " ")
		}
		if settings.TrimLineWhitespace {
			line = strings.TrimSpace(line)
		}
		if settings.CollapseBlankLines && strings.TrimSpace(line) == "" {
			if len(out) == 0 || blankPending {
				continue
			}
			blankPending = true
			out = append(out, "")
			continue
		}
		blankPending = false
		out = append(out, line)
	}
	if settings.TrimLineWhitespace {
		for len(out) > 0 && strings.TrimSpace(out[0]) == "" {
			out = out[1:]
		}
		for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
			out = out[:len(out)-1]
		}
	}
	return strings.Join(out, "\n")
}

func SourceLines(text string) []SourceLine {
	value := strings.ReplaceAll(text, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	result := make([]SourceLine, 0, len(lines))
	for rawIndex, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, SourceLine{
			Index:   len(result) + 1,
			RawLine: rawIndex + 1,
			Text:    line,
		})
	}
	return result
}
