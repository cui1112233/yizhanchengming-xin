package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSetupErrorCodeContract 守护四类缺配置错误码与中文文案的前后端一致：
// JSON 是唯一真相文件，Go 镜像必须逐字相同。
func TestSetupErrorCodeContract(t *testing.T) {
	path := filepath.Join("..", "..", "..", "frontend", "src", "shared", "apiErrorCodes.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error code contract: %v", err)
	}
	var contract struct {
		Codes []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"codes"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("parse error code contract: %v", err)
	}
	jsonMessages := map[string]string{}
	for _, item := range contract.Codes {
		jsonMessages[item.Code] = item.Message
	}
	goMessages := map[string]string{
		ErrorCodeImageModelNotConfigured:    MessageImageModelNotConfigured,
		ErrorCodeVideoProviderNotConfigured: MessageVideoProviderNotConfigured,
		ErrorCodeExecutorNotReady:           MessageExecutorNotReady,
		ErrorCodeGiantPluginNotConfigured:   MessageGiantPluginNotConfigured,
	}
	for code, want := range goMessages {
		got, ok := jsonMessages[code]
		if !ok {
			t.Errorf("code %q exists in Go but not in the JSON contract", code)
			continue
		}
		if got != want {
			t.Errorf("code %q message mismatch:\n go=%q\njson=%q", code, want, got)
		}
	}
}
