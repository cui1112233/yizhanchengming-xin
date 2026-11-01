package httpapi

import (
	"context"
	"strings"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
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
	for _, banned := range []string{"武器", "刀具", "伤口", "血液", "昏迷", "殴打", "掐颈", "餐具尖端", "掠过夏夏脸颊"} {
		if strings.Contains(rewritten, banned) {
			t.Fatalf("compliance note reintroduced rejected content %q: %s", banned, rewritten)
		}
	}
	// 台词、分格和剧情关键词必须保留，否则画面会和其它页对不上。
	if !strings.Contains(rewritten, "妍妍哭泣") || !strings.Contains(rewritten, "夏夏侧身避开") {
		t.Fatalf("rewritten prompt dropped story detail: %s", rewritten)
	}
}

func TestSanitizeShuihuoImagePromptViolenceKeepsCleanPrompt(t *testing.T) {
	clean := "【四格纵滑叙事条漫】酒店宴会厅白天，夏夏冷漠挣脱，母亲端起水杯泼出冰水。"
	if rewritten := sanitizeShuihuoImagePromptViolence(clean); rewritten != clean {
		t.Fatalf("clean prompt should stay unchanged, got %s", rewritten)
	}
}

// 审核拒绝后的修复步骤必须真的调用文本模型，并且返回值仍要过一遍本地暴力
// 规则，防止模型重写时残留“鲜血/刀叉”这类词。
func TestRepairComicPageImagePromptUsesTextModelAndReSanitizes(t *testing.T) {
	called := false
	generator := comicPageGenerator{rewritePrompt: func(_ context.Context, owner string, request novelfetchworkshop.RewriteRequest) (string, error) {
		called = true
		if owner != "alice" {
			t.Fatalf("owner=%q", owner)
		}
		if !strings.Contains(request.UserPrompt, "抢刀叉") {
			t.Fatalf("repair request should carry the original prompt: %s", request.UserPrompt)
		}
		// 文本模型本身也可能留下违规词，这里模拟一个“改写不彻底”的返回。
		return "【合规改写】保持四格纵向叙事分格、Q版豆丁画风与全部台词不变，宴会厅场景、外留边、气泡样式保持原样。人物改用表情和后退躲避表现冲突：沈远伸手去接餐具，金属餐叉尖端掠过夏夏脸颊，鲜血仍在画面上，母亲端起水杯后退一步，夏夏侧身避开，画面以冷白顶光收尾。", nil
	}}
	repaired := generator.repairComicPageImagePrompt(context.Background(), "alice", shuihuoViolencePromptSample, "非常抱歉，该提示可能违反了关于暴力内容的防护限制。")
	if !called {
		t.Fatal("text model rewrite was not called")
	}
	for _, banned := range []string{"刀叉", "餐叉", "叉尖", "鲜血"} {
		if strings.Contains(repaired, banned) {
			t.Fatalf("repaired prompt still contains %q: %s", banned, repaired)
		}
	}
	if !strings.Contains(repaired, "合规改写") {
		t.Fatalf("repaired prompt should keep model output: %s", repaired)
	}
}
