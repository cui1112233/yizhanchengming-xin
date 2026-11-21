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

func TestVideoVisualSafetyFallbackTurnsSensitiveNegativeWordsIntoPositiveSafetyRules(t *testing.T) {
	original := `基础设定：人物年龄未注明，现代都市男女。
视频提示词：[场景 1｜总时长 8 秒] 两人在室内正常交谈。
画面限制：禁止出现与国家、党徽、企业有关的logo，禁止漏胸、乳沟擦边镜头。`

	safe := VideoVisualSafetyFallback(original)
	for _, rejected := range []string{"年龄未注明", "国家", "党徽", "漏胸", "乳沟", "擦边"} {
		if strings.Contains(safe, rejected) {
			t.Fatalf("fallback still sends sensitive negative word %q to the video provider: %s", rejected, safe)
		}
	}
	for _, required := range []string{"人物均为成年人", "衣着完整得体", "[场景 1｜总时长 8 秒]", "两人在室内正常交谈"} {
		if !strings.Contains(safe, required) {
			t.Fatalf("fallback dropped positive safety rule or scene structure %q: %s", required, safe)
		}
	}
}

func TestEmbeddedAdminPresetsDoNotTeachSensitiveNegativeWords(t *testing.T) {
	text := string(adminPresetDefaults)
	for _, rejected := range []string{"禁止漏胸", "乳沟擦边", "国家、党徽"} {
		if strings.Contains(text, rejected) {
			t.Fatalf("embedded admin preset still contains provider-sensitive negative wording %q", rejected)
		}
	}
	for _, required := range []string{"人物均为成年人", "衣着完整得体", "画面不出现任何可识别的标识"} {
		if !strings.Contains(text, required) {
			t.Fatalf("embedded admin preset is missing positive safety rule %q", required)
		}
	}
}
