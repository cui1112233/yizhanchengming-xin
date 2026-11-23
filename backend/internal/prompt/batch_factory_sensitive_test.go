package prompt

import (
	"strings"
	"testing"
)

func TestBatchFactorySensitiveSystemPrompt(t *testing.T) {
	if got := BatchFactorySensitiveSystemPrompt(); strings.TrimSpace(got) == "" {
		t.Fatal("batch factory sensitive system prompt is empty")
	}
}
