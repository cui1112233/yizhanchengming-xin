package batchfactoryv11

import (
	"regexp"
	"strings"
)

// 这四个剥离函数逐个移植自前端 scriptFinalSegment.js（stripBaseSetupSection /
// stripStandaloneSetupSections / stripLegacySharedSetupSections / stripConstraintLines），
// 行为保持一致：剥掉模型输出里自带的共享设定段与约束行，最终只由程序按约束开关注入，
// 避免与程序生成版本重复，也防止模型绕过用户开关。

var (
	reStripBaseSetupStart  = regexp.MustCompile(`^(?:【基础设定】|基础设定[：:])`)
	reStripBaseSetupBody   = regexp.MustCompile(`^(?:#{1,6}\s*)?(?:镜头|分镜)\s*[第#]?\s*(?:\d+|[一二三四五六七八九十百千万两]+)`)
	reStripTimestamp       = regexp.MustCompile(`^\[?\d{1,2}:\d{2}\s*[-—~]`)
	reStripTimeRange       = regexp.MustCompile(`^[（(]?\d+\s*[-—~]\s*\d+\s*(?:s|秒)[)）]`)
	reStripBracket         = regexp.MustCompile(`^【`)
	reStripSharedBodyStart = regexp.MustCompile(`^(?:画面前缀|画质约束|画面约束提示词|视频提示词|画面限制|负面提示词)[：:]`)

	reStripStandaloneStart = regexp.MustCompile(`^(?:【|\[)?(?:人物与场景|人物卡|场景卡)(?:】|\])?(?:[：:].*)?$`)
	reStripContentStart    = regexp.MustCompile(`^(?:【|\[)?(?:时间轴|画面内容|声音设计|氛围与画质规范)(?:】|\])?|^(?:镜头画面[：:]|(?:#{1,6}\s*)?(?:镜头|分镜)\s*[第#]?\s*(?:\d+|[一二三四五六七八九十百千万两]+)|\[?\d{1,2}:\d{2}\s*[-—~])`)

	// 新 SD 模板把“场景环境：”作为单个场景内的正常字段输出，它不是旧版共享
	// 设定段的起点；只有已经进入“统一风格/统一人物”旧块时，才把它当作附属行。
	reStripLegacySetupStart = regexp.MustCompile(`^(?:统一风格|统一人物)[：:]`)
	// “[镜头 N｜…]”是新模板的镜头行，必须能被识别为旧块的结束边界，否则
	// 旧块跳过会一路吞掉后续镜头和下一个场景。
	reStripLegacyBodyStart = regexp.MustCompile(`^(?:镜头画面[：:]|【|\[\s*(?:镜头|分镜)|(?:#{1,6}\s*)?(?:镜头|分镜)\s*[第#]?\s*(?:\d+|[一二三四五六七八九十百千万两]+)|\[?\d{1,2}:\d{2}\s*[-—~])`)

	reStripPrefixLine      = regexp.MustCompile(`(?m)(^|\n)(?:【画面前缀】|画面前缀[：:])[^\n]*\n?`)
	reStripQualityLine     = regexp.MustCompile(`(?m)(^|\n)(?:【画质约束】|画质约束[：:])[^\n]*\n?`)
	reStripRestrictionLine = regexp.MustCompile(`(?m)(^|\n)(?:【画面限制】|画面限制[：:])[^\n]*\n?`)
	reStripNegativeLine    = regexp.MustCompile(`(?m)(^|\n)负面提示词[：:][^\n]*\n?`)
	reStripBlankRuns       = regexp.MustCompile(`\n{3,}`)
)

func collapseBlankRuns(text string) string {
	return reStripBlankRuns.ReplaceAllString(text, "\n\n")
}

// stripBaseSetupSection 剥掉模型输出中已有的【基础设定】段落（由程序生成版本替换）。
func stripBaseSetupSection(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	inSetup := false
	plainSetup := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if reStripBaseSetupStart.MatchString(trimmed) {
			inSetup = true
			plainSetup = !strings.HasPrefix(trimmed, "【")
			continue
		}
		if inSetup {
			boundary := reStripBracket.MatchString(trimmed) ||
				reStripBaseSetupBody.MatchString(trimmed) ||
				reStripTimestamp.MatchString(trimmed) ||
				reStripTimeRange.MatchString(trimmed) ||
				reStripSharedBodyStart.MatchString(trimmed) ||
				(!plainSetup && trimmed == "")
			if boundary {
				inSetup = false
				plainSetup = false
			} else {
				continue
			}
		}
		out = append(out, line)
	}
	return strings.TrimSpace(collapseBlankRuns(strings.Join(out, "\n")))
}

// stripStandaloneSetupSections 剥掉独立的共享设定区块（人物与场景/人物卡/场景卡），
// 防止它们绕过基础设定开关。
func stripStandaloneSetupSections(value string) string {
	lines := strings.Split(value, "\n")
	result := make([]string, 0, len(lines))
	skipping := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if reStripStandaloneStart.MatchString(trimmed) {
			skipping = true
			continue
		}
		if skipping {
			if trimmed == "" {
				continue
			}
			if !reStripContentStart.MatchString(trimmed) {
				continue
			}
			skipping = false
		}
		result = append(result, line)
	}
	return strings.TrimSpace(collapseBlankRuns(strings.Join(result, "\n")))
}

// stripLegacySharedSetupSections 剥掉旧版模型输出的“统一风格/统一人物/场景环境”，
// 这些现在由基础设定和画面前缀开关唯一负责。
func stripLegacySharedSetupSections(value string) string {
	lines := strings.Split(value, "\n")
	result := make([]string, 0, len(lines))
	skipping := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if reStripLegacySetupStart.MatchString(trimmed) {
			skipping = true
			continue
		}
		if skipping {
			if trimmed == "" {
				continue
			}
			if !reStripLegacyBodyStart.MatchString(trimmed) {
				continue
			}
			skipping = false
		}
		result = append(result, line)
	}
	return strings.TrimSpace(collapseBlankRuns(strings.Join(result, "\n")))
}

// stripConstraintLines 剥掉模型输出中已有的约束行（由程序生成版本替换）。
func stripConstraintLines(text string) string {
	out := reStripPrefixLine.ReplaceAllString(text, "$1")
	out = reStripQualityLine.ReplaceAllString(out, "$1")
	out = reStripRestrictionLine.ReplaceAllString(out, "$1")
	out = reStripNegativeLine.ReplaceAllString(out, "$1")
	return strings.TrimSpace(collapseBlankRuns(out))
}

// stripSharedPromptSections 按与前端剧本生成一致的顺序，对一段模型输出正文做完整剥离：
// 先剥旧版共享设定，再剥【基础设定】，再剥独立人物/场景区块，最后剥约束行。
func stripSharedPromptSections(text string) string {
	body := stripLegacySharedSetupSections(text)
	body = stripBaseSetupSection(body)
	body = stripStandaloneSetupSections(body)
	body = stripConstraintLines(body)
	return body
}
