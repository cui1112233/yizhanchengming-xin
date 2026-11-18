package luoshui

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestNativeAssetMessagesMatchRecoveredZIP(t *testing.T) {
	for kind, want := range map[string]string{
		"character": "3f48b4d2bdf3829d3df86b5413114f1e058d3101cc0a595ee9dc003d9052ec52",
		"scene":     "947e1a48ae0147231064c388be8670b64cbe99dff925a63c39cc29ded74596ba",
		"prop":      "75735262bbf476f63c16d8f2804c2e4b0b7241aa42347c22ec9c5e6f61098d0f",
	} {
		system, user, ok := NativeAssetMessages(kind, "UNCHANGED_SOURCE\n")
		if !ok || fmt.Sprintf("%x", sha256.Sum256([]byte(system))) != want {
			t.Fatalf("%s recovered system changed", kind)
		}
		if !strings.HasPrefix(user, "请分析以下小说文本") || !strings.HasSuffix(user, "\n\nUNCHANGED_SOURCE\n") {
			t.Fatalf("%s user envelope changed", kind)
		}
	}
	if _, _, ok := NativeAssetMessages("unknown", "text"); ok {
		t.Fatal("unknown type accepted")
	}
}
