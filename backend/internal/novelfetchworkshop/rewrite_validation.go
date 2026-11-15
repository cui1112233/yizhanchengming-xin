package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/prompt/batchbook"
	"sort"
	"strings"
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

func parseRewriteOutput(raw, mode, bookID string, lineCount int) (string, error) {
	raw = strings.TrimSpace(raw)
	if mode != "" && mode != "plain_text" && mode != "changed_lines_json" {
		return "", fmt.Errorf("不支持的改文响应模式：%s", mode)
	}
	if mode != "changed_lines_json" {
		return raw, ValidateRewriteProse(raw)
	}
	if strings.HasPrefix(raw, "```json\n") && strings.HasSuffix(raw, "```") {
		raw = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "```json\n"), "```"))
	}
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

func completeValidatedRewrite(ctx context.Context, provider RewriteProvider, request RewriteRequest, bookID, mode string, lineCount, targetChars int) (string, error) {
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
		if err == nil && !rewriteWithinSize(result, targetChars) {
			err = fmt.Errorf("AI 改文输出过长")
		}
		// Very short source snippets are legitimate. Apply the floor only to substantial blocks.
		if err == nil && targetChars >= 200 && len([]rune(result)) < targetChars/10 {
			err = fmt.Errorf("AI 改文输出异常过短")
		}
		if err == nil {
			return result, nil
		}
		lastErr = err
		request.UserPrompt = base + fmt.Sprintf(batchbook.Repair, err.Error())
	}
	return "", fmt.Errorf("%w，已停止（最多 3 次）：%v", ErrRewriteValidation, lastErr)
}
