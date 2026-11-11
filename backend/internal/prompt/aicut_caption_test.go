package prompt

import (
	"strings"
	"testing"
)

func TestAICutCaptionCorrectionContractKeepsIDsAndScriptAsData(t *testing.T) {
	contract, err := AICutCaptionCorrectionContract("林夏来到洛水之畔。", []AICutCaption{{ClipID: "sub-1", StartMS: 1200, EndMS: 3000, Text: "林夏来到落水之畔"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contract.SystemPrompt, `{"corrections"`) || !strings.Contains(contract.SystemPrompt, "不能修改时间") {
		t.Fatalf("system prompt misses output/timing contract: %s", contract.SystemPrompt)
	}
	for _, want := range []string{`"script":"林夏来到洛水之畔。"`, `"clipId":"sub-1"`, `"startMs":1200`, `"text":"林夏来到落水之畔"`} {
		if !strings.Contains(contract.UserPrompt, want) {
			t.Fatalf("user prompt missing %s: %s", want, contract.UserPrompt)
		}
	}
}

func TestAICutCaptionCorrectionContractRejectsEmptyInputs(t *testing.T) {
	if _, err := AICutCaptionCorrectionContract("", []AICutCaption{{ClipID: "sub-1", Text: "字幕"}}); err == nil {
		t.Fatal("empty script should fail")
	}
	if _, err := AICutCaptionCorrectionContract("剧本", nil); err == nil {
		t.Fatal("empty captions should fail")
	}
}
