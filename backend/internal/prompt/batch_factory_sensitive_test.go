package prompt

import "testing"

func TestBatchFactorySensitiveSystemPrompt(t *testing.T) {
	const want = "你是内容合规改写助手。只改写用户给出的命中片段，保持原意，去掉违规、擦边、色情、低俗表达。"
	if got := BatchFactorySensitiveSystemPrompt(); got != want {
		t.Fatalf("BatchFactorySensitiveSystemPrompt() = %q, want %q", got, want)
	}
}
