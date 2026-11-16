package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/prompt/batchbook"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var ErrRewriteValidation = errors.New("改文校验未通过")

func rewriteString(value any) string {
	if value == nil {
		return ""
	}
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

// Replace only named placeholders, not the braces of the JSON example.
func fillRewriteTemplate(template string, values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		pairs = append(pairs, "{"+key+"}", values[key])
	}
	return strings.NewReplacer(pairs...).Replace(template)
}

// ValidateRewriteProse is also used immediately before an AI version is submitted.
func ValidateRewriteProse(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("AI 改文返回为空")
	}
	lower := strings.ToLower(text)
	for _, prefix := range []string{"抱歉，我无法", "对不起，我无法", "很抱歉，我不能", "我无法协助", "我不能帮助", "i'm sorry", "i cannot assist", "i can't assist", "as an ai", "internal server error", "bad gateway", "service unavailable", "error:", "<!doctype", "<html", "```", "{", "["} {
		if strings.HasPrefix(lower, prefix) {
			return fmt.Errorf("AI 改文包含拒答、错误消息或未解析格式，请重试或检查模型")
		}
	}
	return nil
}

// normalizeRewriteJSON tolerates the two formatting defects seen from
// OpenAI-compatible gateways: Markdown code fences around the object, and raw
// line breaks injected inside JSON string values. A newline is illegal inside
// a JSON string, so a model that otherwise followed the changed_lines contract
// would fail every retry merely because the transport wrapped long lines.
// Keep only the first complete top-level object as well, so trailing prose does
// not invalidate an otherwise usable rewrite.
func normalizeRewriteJSON(raw string) string {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "```") {
		if index := strings.IndexByte(text, '\n'); index >= 0 {
			text = text[index+1:]
		} else {
			text = strings.TrimPrefix(text, "```json")
			text = strings.TrimPrefix(text, "```JSON")
			text = strings.TrimPrefix(text, "```")
		}
		text = strings.TrimSpace(text)
	}
	if strings.HasSuffix(text, "```") {
		text = strings.TrimSpace(strings.TrimSuffix(text, "```"))
	}
	start := strings.IndexByte(text, '{')
	if start < 0 {
		return text
	}
	var builder strings.Builder
	builder.Grow(len(text) - start)
	inString := false
	escaped := false
	depth := 0
	for _, char := range text[start:] {
		if inString {
			switch {
			case char == '\n' || char == '\r' || char == '\t':
				continue
			case escaped:
				escaped = false
			case char == '\\':
				escaped = true
			case char == '"':
				inString = false
			}
			builder.WriteRune(char)
			continue
		}
		switch char {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
		}
		builder.WriteRune(char)
		if depth == 0 {
			return builder.String()
		}
	}
	return builder.String()
}

func parseRewriteOutput(raw, mode, bookID string, lineCount int) (string, error) {
	raw = strings.TrimSpace(raw)
	if mode != "" && mode != "plain_text" && mode != "changed_lines_json" {
		return "", fmt.Errorf("不支持的改文响应模式：%s", mode)
	}
	if mode != "changed_lines_json" {
		return raw, ValidateRewriteProse(raw)
	}
	raw = normalizeRewriteJSON(raw)
	var payload struct {
		ID    string `json:"id"`
		Lines []struct {
			Number int    `json:"line_no"`
			Text   string `json:"text"`
		} `json:"changed_lines"`
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return "", fmt.Errorf("改文 JSON 格式不正确")
	}
	// Unmarshal additionally rejects trailing documents/garbage.
	if !json.Valid([]byte(raw)) || payload.ID != bookID || len(payload.Lines) == 0 {
		return "", fmt.Errorf("改文书号不符或 changed_lines 为空")
	}
	seen := map[int]bool{}
	result := make([]string, 0, len(payload.Lines))
	previous := 0
	for _, line := range payload.Lines {
		if line.Number < 1 || line.Number > lineCount || seen[line.Number] || line.Number <= previous {
			return "", fmt.Errorf("改文行号重复、越界或顺序错误")
		}
		if err := ValidateRewriteProse(line.Text); err != nil {
			return "", err
		}
		seen[line.Number] = true
		previous = line.Number
		result = append(result, strings.TrimSpace(line.Text))
	}
	return strings.Join(result, "\n"), nil
}

// 模型偶尔会把后端提示词、JSON 协议字段或重试指令原样带进正文。
// 这些标记在正常小说正文中不会出现，命中即视为本次输出不可用。
var rewriteLeakMarkers = []string{
	"待处理片段：",
	"后续衔接参考",
	"只输出改写后的正文",
	"改文模板：",
	"临时指令：",
	"changed_lines",
	`"line_no"`,
	"接口连通性测试助手",
	"请只回复：ok",
}

var rewriteVersionLeakPattern = regexp.MustCompile(`这是第\s*\d+\s*个改写版本`)

// rewriteComparableText 去掉空白和标点后比较，避免仅换行/标点差异被当成改写。
func rewriteComparableText(text string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(text) {
		if unicode.IsSpace(char) || unicode.IsPunct(char) {
			continue
		}
		builder.WriteRune(char)
	}
	return builder.String()
}

func rewriteBigramSet(text string) map[string]struct{} {
	runes := []rune(text)
	set := make(map[string]struct{}, len(runes))
	for index := 0; index+1 < len(runes); index++ {
		set[string(runes[index:index+2])] = struct{}{}
	}
	return set
}

// rewriteSimilarity 返回字符二元组的 Jaccard 相似度，用于识别模型只复读原文。
func rewriteSimilarity(left, right string) float64 {
	leftSet := rewriteBigramSet(left)
	rightSet := rewriteBigramSet(right)
	if len(leftSet) == 0 || len(rightSet) == 0 {
		return 0
	}
	intersection := 0
	for gram := range leftSet {
		if _, ok := rightSet[gram]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(leftSet)+len(rightSet)-intersection)
}

// validateRewriteQuality 在格式和长度校验之后检查内容质量：提示词泄漏、
// 段落级重复，以及输出与原片段几乎一致（等于没有改写）。
func validateRewriteQuality(result, source string) error {
	lower := strings.ToLower(result)
	for _, marker := range rewriteLeakMarkers {
		if strings.Contains(lower, marker) {
			return fmt.Errorf("AI 改文输出包含提示词或协议字段")
		}
	}
	if rewriteVersionLeakPattern.MatchString(result) {
		return fmt.Errorf("AI 改文输出包含提示词或协议字段")
	}
	counts := map[string]int{}
	for _, line := range strings.Split(result, "\n") {
		normalized := strings.TrimSpace(line)
		if len([]rune(normalized)) < 12 {
			continue
		}
		counts[normalized]++
		if counts[normalized] >= 3 {
			return fmt.Errorf("AI 改文输出出现重复段落")
		}
	}
	original := rewriteComparableText(source)
	rewritten := rewriteComparableText(result)
	if len([]rune(original)) >= 100 && len([]rune(rewritten)) > 0 {
		if original == rewritten {
			return fmt.Errorf("AI 改文输出与原文完全一致")
		}
		if rewriteSimilarity(original, rewritten) >= 0.96 {
			return fmt.Errorf("AI 改文输出与原文几乎一致，未实际改写")
		}
	}
	return nil
}

func completeValidatedRewrite(ctx context.Context, provider RewriteProvider, request RewriteRequest, bookID, mode string, lineCount, targetChars int, source string) (string, error) {
	if mode != "" && mode != "plain_text" && mode != "changed_lines_json" {
		return "", fmt.Errorf("不支持的改文响应模式：%s", mode)
	}
	base := request.UserPrompt
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		raw, err := provider.Complete(ctx, request)
		if err != nil {
			message := strings.ToLower(err.Error())
			truncated := strings.Contains(message, "输出被长度截断") || strings.Contains(message, "finish_reason=length") || strings.Contains(message, "max_output_tokens") && strings.Contains(message, "truncat")
			if truncated {
				lastErr = err
				// Compact budgets can be consumed by reasoning before any prose
				// is returned. Escalate only an explicit truncation, at most twice.
				// Exhaustion becomes a manual hold instead of another queue loop.
				if request.MaxTokens >= 24000 {
					break
				}
				next := request.MaxTokens * 2
				if next < 12000 {
					next = 12000
				}
				if next > 24000 {
					next = 24000
				}
				request.MaxTokens = next
				continue
			}
			return "", err
		}
		result, err := parseRewriteOutput(raw, mode, bookID, lineCount)
		// 与批量原文改文系统源码保持一致：不设 135% 长度硬上限。
		// 只保留空输出、污染、过短和几乎未改写这几类真正影响成品的检查。
		// Very short source snippets are legitimate. Apply the floor only to substantial blocks.
		if err == nil && targetChars >= 200 && len([]rune(result)) < targetChars/10 {
			err = fmt.Errorf("AI 改文输出异常过短")
		}
		if err == nil {
			err = validateRewriteQuality(result, source)
		}
		if err == nil {
			return result, nil
		}
		lastErr = err
		// 与源码一致：后续重试降低随机性并放大输出预算，
		// 避免同一个提示词原样重试三次仍然得到同样的问题输出。
		if request.Temperature > 0.1 {
			request.Temperature = 0.1
		}
		if request.MaxTokens < 24000 {
			next := request.MaxTokens * 2
			if next < 12000 {
				next = 12000
			}
			if next > 24000 {
				next = 24000
			}
			request.MaxTokens = next
		}
		request.UserPrompt = base + fmt.Sprintf(batchbook.Repair, err.Error())
	}
	return "", fmt.Errorf("%w，已停止（最多 3 次）：%v", ErrRewriteValidation, lastErr)
}
