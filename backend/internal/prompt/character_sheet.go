package prompt

import "strings"

func BuildCharacterSheetPrompt(body, description string) string {
	return strings.TrimSpace(body) + "\n\n" + strings.TrimSpace(description)
}
