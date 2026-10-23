package batchfactoryv11

import "strings"

func normalizeCapturedSourceText(value string) (string, bool) {
	text := strings.TrimSpace(value)
	if text == "" || text == "[object Object]" {
		return "", false
	}
	return text, true
}
