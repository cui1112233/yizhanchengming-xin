package prompt

import "testing"

func TestAICutFillerContractRequiresCaptionsAndKeepsPromptOnServer(t *testing.T) {
	if _, err := AICutFillerContract(nil); err == nil {
		t.Fatal("empty captions should fail")
	}
	contract, err := AICutFillerContract([]AICutCaption{{ClipID: "sub-1", Text: "嗯"}, {ClipID: "sub-2", Text: "嗯我知道了"}})
	if err != nil {
		t.Fatal(err)
	}
	if contract.SystemPrompt == "" || contract.UserPrompt == "" {
		t.Fatalf("contract=%+v", contract)
	}
}
