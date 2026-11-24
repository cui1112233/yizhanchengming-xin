package batchfactorytext

import "strings"

// Reading breaks are display boundaries in original rune coordinates. They
// never insert text, so highlights and editable source offsets stay stable.
func readingBreaks(text string, c LayoutConfig, width int) map[int]bool {
	runes := []rune(text)
	result := map[int]bool{}
	target := max(1, int(float64(width)/max(c.FontSize, 1))*2)
	start := 0
	var quotes straightQuotes
	for start < len(runes) {
		end := paragraphEnd(runes, start)
		group := start
		for cursor := start; cursor < end; {
			next := quotes.sentenceEnd(runes, cursor, end)
			sentence := strings.TrimSpace(string(runes[cursor:next]))
			dialogue := strings.HasPrefix(sentence, "“") || strings.HasPrefix(sentence, "「") || strings.HasPrefix(sentence, "\"")
			if cursor > group && (next-group > target || dialogue) {
				result[cursor] = true
				group = cursor
			}
			if dialogue && next < end {
				result[next] = true
				group = next
			}
			cursor = next
		}
		start = end
	}
	return result
}
