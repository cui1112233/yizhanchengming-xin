package screenplay

import (
	"strings"
	"unicode/utf8"
)

// Original helper compatibility: bounded overlap lookup, then Markdown-aware
// glue. This deliberately does not promise to remove all repeated paragraphs.
func appendStoryboardPlanChunk(previous, chunk string) string {
	a, b := strings.TrimSpace(previous), strings.TrimSpace(chunk)
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	r := []rune(a)
	for _, n := range []int{480, 240, 160, 80, 40} {
		if len(r) < n {
			continue
		}
		tail := string(r[len(r)-n:])
		idx := strings.Index(b, tail)
		if idx >= 0 && utf8.RuneCountInString(b[:idx]) < 120 {
			b = strings.TrimSpace(b[idx+len(tail):])
			break
		}
	}
	if b == "" {
		return a
	}
	glue := "\n"
	if strings.HasPrefix(b, "#") || strings.HasPrefix(b, "|") || strings.HasPrefix(b, "-") {
		glue = ""
	}
	return a + glue + b
}
