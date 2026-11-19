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
var openingVariantSectionPattern = regexp.MustCompile(`(?mi)^[\t ]*(?:#{1,6}[\t ]*)?(?:【|［|\[)?(?:\*{1,2}|_{1,2})?(?:={2,}[\t ]*)?(?:VARIANT|变体|换开头)[\t ]*(?:第)?[\t ]*(\d+|[一二三四五六七八九十]+)(?:[:：]|】|］|\]|={2,}|[\*_]|[\t ]+|$)[\t ]*(?:】|］|\])?(?:={2,})?(?:\*{1,2}|_{1,2})?[\t ]*`)

// openingVariantNumber accepts the Chinese numerals models occasionally use
// ("【变体二】") without dropping an otherwise valid section.
func openingVariantNumber(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if parsed, err := strconv.Atoi(raw); err == nil {
		return parsed, true
	}
	digits := map[rune]int{'一': 1, '二': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9, '十': 10}
	runes := []rune(raw)
	if len(runes) == 1 {
		if value, ok := digits[runes[0]]; ok {
			return value, true
		}
		return 0, false
	}
	if len(runes) == 2 && runes[0] == '十' {
		if value, ok := digits[runes[1]]; ok {
			return 10 + value, true
		}
	}
	return 0, false
}

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
		number, ok := openingVariantNumber(text[marker[2]:marker[3]])
		if !ok || number < 1 || number > variantCount {
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
	if len(markers) > 0 {
		return out
	}
	// Some models answer with a JSON object instead of the requested text markers.
	// Accept it before falling back to a whole-text copy so a valid variant is not
	// discarded just because the wire decoration changed.
	if parsed, ok := parseOpeningVariantsJSON(text, maxVideoDuration, originalDuration, variantCount); ok {
		return parsed
	}
	// A single requested variant has no real ambiguity: when the model omits the
	// marker, the whole reply is the replacement opening. This unblocks the common
	// openingCount=2 case without weakening multi-variant validation.
	if variantCount == 1 {
		prompt := openingVariantPromptOnly(text)
		if prompt != "" && originalDuration >= 1 && originalDuration <= maxVideoDuration {
			out[0] = OpeningVariant{Index: 1, Label: openingVariantLabel(1), Prompt: prompt, Status: "success", DurationSec: originalDuration}
		}
	}
	return out
}

type openingVariantJSON struct {
	Index       int     `json:"index"`
	Prompt      string  `json:"prompt"`
	Content     string  `json:"content"`
	Text        string  `json:"text"`
	DurationSec int     `json:"durationSec"`
	Duration    float64 `json:"duration"`
}

func parseOpeningVariantsJSON(text string, maxVideoDuration, originalDuration, variantCount int) ([]OpeningVariant, bool) {
	payload := openingJSONPayload(text)
	if payload == "" {
		return nil, false
	}
	var items []openingVariantJSON
	if err := json.Unmarshal([]byte(payload), &items); err != nil || len(items) == 0 {
		var wrapper struct {
			Variants []openingVariantJSON `json:"variants"`
			Openings []openingVariantJSON `json:"openings"`
		}
		if err := json.Unmarshal([]byte(payload), &wrapper); err != nil {
			return nil, false
		}
		items = wrapper.Variants
		if len(items) == 0 {
			items = wrapper.Openings
		}
	}
	if len(items) == 0 {
		return nil, false
	}
	out := make([]OpeningVariant, 0, variantCount)
	for index := 1; index <= variantCount; index++ {
		out = append(out, OpeningVariant{Index: index, Label: openingVariantLabel(index), Status: "failed", FailureReason: "模型未输出该变体分段"})
	}
	successes := 0
	for position, item := range items {
		number := item.Index
		if number < 1 || number > variantCount {
			number = position + 1
		}
		if number < 1 || number > variantCount {
			continue
		}
		prompt := strings.TrimSpace(item.Prompt)
		if prompt == "" {
			prompt = strings.TrimSpace(item.Content)
		}
		if prompt == "" {
			prompt = strings.TrimSpace(item.Text)
		}
		if prompt == "" {
			out[number-1].FailureReason = "变体正文为空"
			continue
		}
		duration := item.DurationSec
		if duration == 0 && item.Duration > 0 {
			duration = int(item.Duration)
		}
		if duration != 0 && (duration < 1 || duration > maxVideoDuration) {
			out[number-1].FailureReason = fmt.Sprintf("时长必须在 1-%d 秒之间", maxVideoDuration)
			continue
		}
		if originalDuration < 1 || originalDuration > maxVideoDuration {
			out[number-1].FailureReason = "原分镜时长无效"
			continue
		}
		out[number-1] = OpeningVariant{Index: number, Label: openingVariantLabel(number), Prompt: prompt, Status: "success", DurationSec: originalDuration}
		successes++
	}
	return out, successes > 0
}

// openingJSONPayload extracts the first balanced JSON array/object from a reply
// that may be wrapped in Markdown fences or surrounded by explanation text.
func openingJSONPayload(text string) string {
	start := strings.IndexAny(text, "{[")
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for index := start; index < len(text); index++ {
		ch := text[index]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				return text[start : index+1]
			}
		}
	}
	return ""
}

func openingVariantPromptOnly(text string) string {
	lines := strings.Split(text, "\n")
	promptLines := make([]string, 0, len(lines))
	for index, line := range lines {
		if index == 0 && sdDurationPattern.MatchString(strings.TrimSpace(line)) {
			continue
		}
		promptLines = append(promptLines, line)
	}
	return strings.TrimSpace(strings.Join(promptLines, "\n"))
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

// OpeningStageComplete 判断「换开头」阶段是否已经做完：开关关闭、只需要一个
// 开头版本、单分镜书，或所需变体全部成功时视为完成。自动化每次触发都先查它，
// 避免为了补齐一个变体而把整轮开头重新生成。
//
// VIDEO 槽位由导演阶段创建（Novel Fetch 追加的书在导演前没有视频），所以
// len(Videos)==0 只代表“分镜还没生成”，不能当作已完成；否则新书会在导演
// 完成后被直接跳过换开头，最终合成和上传都缺换开头版本。
func OpeningStageComplete(batch Batch, book Book) bool {
	effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if !rawBool(effective, "openingEnabled", false) {
		return true
	}
	count := rawInt(effective, "openingCount", 4)
	if count <= 1 {
		return true
	}
	if count < 1 || count > 8 {
		return false
	}
	if len(book.Videos) == 0 {
		return false
	}
	// 单分镜书整段视频就是开头，没有可替换的独立开头版本。
	if len(book.Videos) == 1 {
		return true
	}
	return len(successfulOpeningVariants(book.Videos[0])) >= count-1
}

// RequiredOpeningVariantIndices 返回视频阶段实际需要产出的开头版本编号
// （0 为原开头，其余是已成功生成的换开头变体）。缺失的变体是可重试的
// 单步缺口，不能让整本永远停在“等待生成视频”，因此这里不能按配置的
// openingCount 硬算 1..N，必须与 OpeningMergeIndices 使用同一口径：
// 有多少个成功变体就要求多少个，之后补生成的新变体会自然重新打开缺口。
func RequiredOpeningVariantIndices(batch Batch, book Book) []int {
	return OpeningMergeIndices(batch, book)
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
