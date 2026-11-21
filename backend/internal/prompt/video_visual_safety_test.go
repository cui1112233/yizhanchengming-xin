package prompt

import (
	"strings"
	"testing"
)

func TestVideoVisualSafetyFallbackRemovesRejectedPhysicalConflictWithoutDroppingStructure(t *testing.T) {
	original := `视频提示词：[场景 1｜总时长 8 秒]
林倩当众撕毁沈律的身份证和户口本，破损证件留在桌面，重点看清撕毁动作。
[镜头 2｜时长 3.5 秒] 程意愕然望向证件。`

	safe := VideoVisualSafetyFallback(original)
	for _, rejected := range []string{"撕毁", "破损证件"} {
		if strings.Contains(safe, rejected) {
			t.Fatalf("fallback still contains %q: %s", rejected, safe)
		}
	}
	for _, preserved := range []string{"[场景 1｜总时长 8 秒]", "[镜头 2｜时长 3.5 秒]", "林倩", "沈律", "程意"} {
		if !strings.Contains(safe, preserved) {
			t.Fatalf("fallback dropped %q: %s", preserved, safe)
		}
	}
	if safe == original {
		t.Fatal("fallback must change a rejected prompt")
	}
}
