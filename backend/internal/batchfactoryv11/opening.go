package batchfactoryv11

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// OpeningVariant is one AI-generated alternative opening prompt for VIDEO01.
// Index 0 is reserved for the original director prompt and is never stored here.
type OpeningVariant struct {
	Index         int    `json:"index"`
	Label         string `json:"label"`
	Prompt        string `json:"prompt"`
	Status        string `json:"status"`
	DurationSec   int    `json:"durationSec,omitempty"`
	FailureReason string `json:"failureReason,omitempty"`
}

// Models occasionally preserve the requested section numbers while drifting
// from the exact decoration (for example "### VARIANT 1", "**变体 2：**",
// or "=== 换开头 3 ==="). Treat those as the same machine marker so a
// harmless Markdown-format change does not discard an otherwise valid slot.
// Horizontal whitespace is intentional: the match must never consume the
// following prompt line.
var openingVariantSectionPattern = regexp.MustCompile(`(?mi)^[\t ]*(?:#{1,6}[\t ]*)?(?:\*{1,2}|_{1,2})?(?:={2,}[\t ]*)?(?:VARIANT|变体|换开头)[\t ]*(\d+)[\t ]*(?:[:：])?[\t ]*(?:={2,})?(?:\*{1,2}|_{1,2})?[\t ]*`)

func openingVariantLabel(index int) string {
	return fmt.Sprintf("分镜一 | 换开头%d", index)
}

// parseOpeningVariants splits the meta-prompt reply into one OpeningVariant per
// ===VARIANT N=== section. The original VIDEO01 duration is authoritative:
// changing only the opening picture must not ask the model to infer timing
// again. A model-supplied duration is accepted as optional metadata and is
// rejected only when it is explicitly outside the configured limit.
// Every requested slot is returned so the UI can mark genuinely missing or
// empty variants without blocking on harmless output-format drift.
func parseOpeningVariants(raw string, maxVideoDuration, originalDuration, variantCount int) []OpeningVariant {
	out := make([]OpeningVariant, 0, variantCount)
	for index := 1; index <= variantCount; index++ {
		out = append(out, OpeningVariant{Index: index, Label: openingVariantLabel(index), Status: "failed", FailureReason: "模型未输出该变体分段"})
	}
	text := strings.TrimSpace(raw)
	if text == "" || variantCount <= 0 {
		return out
	}
	markers := openingVariantSectionPattern.FindAllStringSubmatchIndex(text, -1)
	for i, marker := range markers {
		number, _ := strconv.Atoi(text[marker[2]:marker[3]])
		if number < 1 || number > variantCount {
			continue
		}
		end := len(text)
		if i+1 < len(markers) {
			end = markers[i+1][0]
		}
		body := strings.TrimSpace(text[marker[1]:end])
		lines := strings.Split(body, "\n")
		promptLines := make([]string, 0, len(lines))
		explicitDuration := 0
		for _, line := range lines {
			durationMatch := sdDurationPattern.FindStringSubmatch(strings.TrimSpace(line))
			if durationMatch == nil || explicitDuration > 0 {
				promptLines = append(promptLines, line)
				continue
			}
			parsed, err := strconv.Atoi(durationMatch[1])
			if err != nil || parsed < 1 || parsed > maxVideoDuration {
				out[number-1].FailureReason = fmt.Sprintf("时长必须在 1-%d 秒之间", maxVideoDuration)
				explicitDuration = -1
				break
			}
			explicitDuration = parsed
		}
		if explicitDuration < 0 {
			continue
		}
		prompt := strings.TrimSpace(strings.Join(promptLines, "\n"))
		if prompt == "" {
			out[number-1].FailureReason = "变体正文为空"
			continue
		}
		if originalDuration < 1 || originalDuration > maxVideoDuration {
			out[number-1].FailureReason = "原分镜时长无效"
			continue
		}
		out[number-1] = OpeningVariant{Index: number, Label: openingVariantLabel(number), Prompt: prompt, Status: "success", DurationSec: originalDuration}
	}
	return out
}

func openingVariantFor(video Video, index int) (OpeningVariant, bool) {
	raw, ok := video.SettingsState.Patch["openingVariants"]
	if !ok {
		return OpeningVariant{}, false
	}
	var variants []OpeningVariant
	if json.Unmarshal(raw, &variants) != nil {
		return OpeningVariant{}, false
	}
	for _, variant := range variants {
		if variant.Index == index && variant.Status == "success" && strings.TrimSpace(variant.Prompt) != "" {
			return variant, true
		}
	}
	return OpeningVariant{}, false
}

func successfulOpeningVariants(video Video) []OpeningVariant {
	variants := openingVariants(video)
	out := []OpeningVariant{}
	for _, variant := range variants {
		if variant.Status == "success" && strings.TrimSpace(variant.Prompt) != "" {
			out = append(out, variant)
		}
	}
	return out
}

// OpeningStageComplete 判断「换开头」阶段是否已经做完：开关关闭、只有一个
// 分镜，或所需变体全部成功时视为完成。自动化每次触发都先查它，避免为了
// 补齐一个变体而把整轮开头重新生成。
func OpeningStageComplete(batch Batch, book Book) bool {
	if len(book.Videos) < 2 {
		return true
	}
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if !rawBool(effective, "openingEnabled", false) {
		return true
	}
	count := rawInt(effective, "openingCount", 4)
	if count < 1 || count > 8 {
		return false
	}
	return len(successfulOpeningVariants(book.Videos[0])) >= count-1
}

// RequiredOpeningVariantIndices 返回「换开头」设置要求的全部版本编号
// （0 为原开头）。自动化用它判断视频阶段是否真的补齐了全部开头条数，
// 而不是只看已经成功的变体。
func RequiredOpeningVariantIndices(batch Batch, book Book) []int {
	indices := []int{0}
	if len(book.Videos) < 2 {
		return indices
	}
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if !rawBool(effective, "openingEnabled", false) {
		return indices
	}
	count := rawInt(effective, "openingCount", 4)
	if count < 1 || count > 8 {
		return indices
	}
	for index := 1; index < count; index++ {
		indices = append(indices, index)
	}
	return indices
}

// OpeningMergeIndices 返回合成阶段应产出的开头版本编号：0 表示原开头，
// 其余是已成功生成的换开头变体。自动化据此判断历史合成是否已经覆盖
// 当前需要的全部版本。
func OpeningMergeIndices(batch Batch, book Book) []int {
	indices := []int{0}
	if len(book.Videos) < 2 {
		return indices
	}
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if !rawBool(effective, "openingEnabled", false) {
		return indices
	}
	for _, variant := range successfulOpeningVariants(book.Videos[0]) {
		indices = append(indices, variant.Index)
	}
	return indices
}

func openingVariants(video Video) []OpeningVariant {
	raw, ok := video.SettingsState.Patch["openingVariants"]
	if !ok {
		return nil
	}
	var variants []OpeningVariant
	if json.Unmarshal(raw, &variants) != nil {
		return nil
	}
	return variants
}
