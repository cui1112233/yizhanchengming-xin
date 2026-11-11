package prompt

import (
	"strings"
	"testing"
)

func TestComicVisualSafetyRetryPlanMasksDialogueAndAdultVisuals(t *testing.T) {
	dialogue := "今晚来我房间，我们单独谈谈"
	input := `夜里，两人靠得很近，气氛暧昧。女主低声说：“` + dialogue + `。”`
	plan := ComicVisualSafetyRetryPlan("系统规则", input, 1)

	for _, rejected := range []string{dialogue, "靠得很近", "暧昧"} {
		if strings.Contains(plan.UserPrompt, rejected) {
			t.Fatalf("safe request still contains %q: %s", rejected, plan.UserPrompt)
		}
	}
	if !strings.Contains(plan.UserPrompt, "{对白1}") {
		t.Fatalf("safe request is missing dialogue placeholder: %s", plan.UserPrompt)
	}
	if !strings.Contains(plan.SystemPrompt, "禁止画面出现成人或暧昧内容") {
		t.Fatalf("safe system rule is missing adult-content protection: %s", plan.SystemPrompt)
	}
}

func TestComicVisualSafetyRetryPlanRestoresDialogueOnlyToScreen(t *testing.T) {
	dialogue := "今晚来我房间，我们单独谈谈。"
	plan := ComicVisualSafetyRetryPlan("系统规则", `女主说：“`+dialogue+`”`, 2)

	restored := plan.RestoreScreenPrompt("安全日常镜头，页面文字：{对白1}")
	if !strings.Contains(restored, dialogue) || strings.Contains(restored, "{对白1}") {
		t.Fatalf("dialogue was not restored: %s", restored)
	}
	if !strings.Contains(plan.SystemPrompt, "无身体接触") {
		t.Fatalf("strict retry rule is missing: %s", plan.SystemPrompt)
	}
}

func TestComicVisualSafetyRetryPlanAppendsDialogueWhenModelDropsPlaceholder(t *testing.T) {
	plan := ComicVisualSafetyRetryPlan("系统规则", `女主说：“保留这句原文。”`, 1)
	restored := plan.RestoreScreenPrompt("安全镜头")
	if !strings.Contains(restored, "保留这句原文。") {
		t.Fatalf("missing original dialogue fallback: %s", restored)
	}
}

func TestComicVisualSafetyRetryPlanStrictRoundDropsAllRawNarrative(t *testing.T) {
	dialogue := "这句对白最后再恢复"
	uniqueNarrative := "她把睡袍往下拉，露出肩膀上的红痕"
	input := `{"book_title":"测试书","current_page":"` + uniqueNarrative + `。女主说：“` + dialogue + `”","source_text":"` + uniqueNarrative + `","comic_assets":[{"asset_id":"role-1","kind":"character","name":"女主","prompt":"` + uniqueNarrative + `"}],"aspect_ratio":"9:16"}`
	plan := ComicVisualSafetyRetryPlan("系统规则", input, 2)

	if strings.Contains(plan.UserPrompt, uniqueNarrative) || strings.Contains(plan.UserPrompt, "睡袍") || strings.Contains(plan.UserPrompt, "红痕") {
		t.Fatalf("strict retry still contains raw narrative: %s", plan.UserPrompt)
	}
	if !strings.Contains(plan.UserPrompt, "role-1") || !strings.Contains(plan.UserPrompt, "女主") {
		t.Fatalf("strict retry lost safe asset identity: %s", plan.UserPrompt)
	}
	if !strings.Contains(plan.UserPrompt, "{对白1}") {
		t.Fatalf("strict retry lost dialogue placeholder: %s", plan.UserPrompt)
	}
}
