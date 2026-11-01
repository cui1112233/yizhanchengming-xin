package httpapi

import (
	"strings"
	"testing"
)

// 婚宴那本书第 35 页的真实提示词片段，上游按“暴力内容”拒绝。
const shuihuoViolencePromptSample = "第二格镜：妍妍哭泣抢刀叉，沈远伸手抢夺失控，金属餐叉尖端狠狠擦过夏夏脸颊，鲜血飞溅而出。"

func TestSanitizeShuihuoImagePromptViolenceRemovesWeaponAndBlood(t *testing.T) {
	rewritten := sanitizeShuihuoImagePromptViolence(shuihuoViolencePromptSample)
	if rewritten == shuihuoViolencePromptSample {
		t.Fatal("violence prompt should be rewritten")
	}
	for _, banned := range []string{"刀叉", "餐叉", "叉尖", "鲜血", "血痕", "狠狠", "擦过"} {
		if strings.Contains(rewritten, banned) {
			t.Fatalf("rewritten prompt still contains %q: %s", banned, rewritten)
		}
	}
	if !strings.Contains(rewritten, "内容合规约束") {
		t.Fatalf("rewritten prompt missing compliance note: %s", rewritten)
	}
	// 台词、分格和剧情关键词必须保留，否则画面会和其它页对不上。
	if !strings.Contains(rewritten, "妍妍哭泣") || !strings.Contains(rewritten, "夏夏脸颊") {
		t.Fatalf("rewritten prompt dropped story detail: %s", rewritten)
	}
}

func TestSanitizeShuihuoImagePromptViolenceKeepsCleanPrompt(t *testing.T) {
	clean := "【四格纵滑叙事条漫】酒店宴会厅白天，夏夏冷漠挣脱，母亲端起水杯泼出冰水。"
	if rewritten := sanitizeShuihuoImagePromptViolence(clean); rewritten != clean {
		t.Fatalf("clean prompt should stay unchanged, got %s", rewritten)
	}
}
