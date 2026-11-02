package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StyleDecision is the result of the second-stage style classification.
// An empty Style means the text could not be classified safely.
type StyleDecision struct {
	Style      string  `json:"style"`
	Source     string  `json:"styleSource"`
	Reason     string  `json:"styleReason"`
	Confidence float64 `json:"confidence"`
}

type StyleAIResult struct {
	Style      string  `json:"style"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

var knownNovelStyles = []string{"古风虐文", "古风甜文", "古风通用", "年代虐文", "年代甜文", "年代通用", "现代虐文", "现代甜文", "现代悬疑", "现代通用", "男频都市", "现代女主", "玄幻", "历史", "家庭奇葩", "家庭伤感", "职场打脸"}

// ResolveStyle uses explicit upstream metadata first and only asks the text
// model when the upstream value is missing or not one of the supported styles.
func ResolveStyle(ctx context.Context, info BookInfo, provider RewriteProvider) StyleDecision {
	if style := matchNovelStyle(info.Genre); style != "" {
		return StyleDecision{Style: style, Source: "bookinfo", Reason: "采用平台返回的明确风格", Confidence: 1}
	}
	if style := matchNovelStyle(info.Category); style != "" {
		return StyleDecision{Style: style, Source: "bookinfo", Reason: "采用平台返回的明确风格", Confidence: 1}
	}
	if style := matchNovelStyleMetadata(info); style != "" {
		return StyleDecision{Style: style, Source: "bookinfo", Reason: "采用平台主题与标签中的明确风格", Confidence: 1}
	}
	if provider == nil || strings.TrimSpace(info.Text) == "" {
		return StyleDecision{Reason: "平台未提供明确风格且 AI 风格判断未配置"}
	}
	request := RewriteRequest{
		SystemPrompt: "你是小说分类器。只返回 JSON：{\"style\":\"风格\"}。style 必须从以下值中选择最接近的一项：古风虐文、古风甜文、古风通用、年代虐文、年代甜文、年代通用、现代虐文、现代甜文、现代悬疑、现代通用、男频都市、现代女主、玄幻、历史、家庭奇葩、家庭伤感、职场打脸。必须选择，不要留空，不要解释 JSON 以外的内容。",
		UserPrompt:   "请判断这段小说的投稿风格：\n" + info.Text,
		Temperature:  0,
		MaxTokens:    80,
	}
	var content string
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}
		content, err = provider.Complete(ctx, request)
		if err == nil || !retryableStyleError(err) || attempt == 2 {
			break
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 200 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			err = ctx.Err()
		case <-timer.C:
		}
	}
	if err != nil {
		return StyleDecision{Reason: "AI 风格判断失败：" + safeStyleError(err)}
	}
	if result := parseStyleAIResult(content); result.Style != "" {
		return StyleDecision{Style: result.Style, Source: "ai", Reason: result.Reason, Confidence: result.Confidence}
	}
	return StyleDecision{Reason: "AI 未返回有效风格"}
}

var styleHTTPStatus = regexp.MustCompile(`\bHTTP ([1-5][0-9]{2})\b`)

func styleErrorStatus(err error) int {
	match := styleHTTPStatus.FindStringSubmatch(err.Error())
	if len(match) != 2 {
		return 0
	}
	status, _ := strconv.Atoi(match[1])
	return status
}

func retryableStyleError(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if status := styleErrorStatus(err); status != 0 {
		return status == 408 || status == 429 || status >= 500
	}
	var networkErr net.Error
	return errors.As(err, &networkErr)
}

func safeStyleError(err error) string {
	if errors.Is(err, context.Canceled) {
		return "请求已取消"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "请求超时"
	}
	if status := styleErrorStatus(err); status != 0 {
		return fmt.Sprintf("模型服务返回 HTTP %d", status)
	}
	var networkErr net.Error
	if errors.As(err, &networkErr) {
		if networkErr.Timeout() {
			return "模型服务连接超时"
		}
		return "模型服务网络连接失败"
	}
	return "模型服务调用出错，请检查配置或稍后重试"
}

func parseStyleAIResult(content string) StyleAIResult {
	raw := strings.TrimSpace(content)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "```json"))
		raw = strings.TrimSpace(strings.TrimSuffix(raw, "```"))
	}
	var result StyleAIResult
	if json.Unmarshal([]byte(raw), &result) == nil {
		if style := matchNovelStyle(result.Style); style != "" {
			// confidence/reason are optional; a named style is enough to
			// continue production without asking the user to confirm.
			if result.Confidence <= 0 {
				result.Confidence = 0.8
			}
			if strings.TrimSpace(result.Reason) == "" {
				result.Reason = "采用模型返回的风格"
			}
			result.Style = style
			return result
		}
	}
	// Some models ignore the JSON instruction and answer with the style name.
	if style := matchNovelStyle(raw); style != "" {
		return StyleAIResult{Style: style, Confidence: 0.8, Reason: "采用模型返回的风格"}
	}
	return StyleAIResult{}
}

func parseStyleResponse(content string) string {
	raw := strings.TrimSpace(content)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "```json"))
		raw = strings.TrimSpace(strings.TrimSuffix(raw, "```"))
	}
	var payload struct {
		Style string `json:"style"`
	}
	if json.Unmarshal([]byte(raw), &payload) == nil {
		if style := matchNovelStyle(payload.Style); style != "" {
			return style
		}
	}
	return matchNovelStyle(raw)
}

func matchNovelStyle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, style := range knownNovelStyles {
		if value == style || strings.Contains(value, style) {
			return style
		}
	}
	// Platform facets rarely use our internal names ("现代言情", "都市", "悬疑").
	// Map the common words deterministically so a simple platform value is enough.
	for _, rule := range novelStyleKeywordRules {
		for _, keyword := range rule.keywords {
			if strings.Contains(value, keyword) {
				return rule.style
			}
		}
	}
	return ""
}

var novelStyleKeywordRules = []struct {
	style    string
	keywords []string
}{
	{"现代女主", []string{"现代女主", "大女主", "女主", "言情", "婚恋", "情感", "现实情感"}},
	{"男频都市", []string{"男频", "都市", "兵王", "赘婿", "异能"}},
	{"玄幻", []string{"玄幻", "仙侠", "修真", "修仙", "奇幻"}},
	{"历史", []string{"历史", "朝堂", "争霸"}},
	{"现代悬疑", []string{"悬疑", "推理", "刑侦", "灵异", "惊悚"}},
	{"古风甜文", []string{"古言甜", "古风甜"}},
	{"古风虐文", []string{"古言虐", "古风虐", "宫斗", "宅斗"}},
	{"古风通用", []string{"古风", "古代", "古言", "宫廷", "武侠"}},
	{"年代甜文", []string{"年代甜"}},
	{"年代虐文", []string{"年代虐"}},
	{"年代通用", []string{"年代", "民国", "知青"}},
	{"家庭奇葩", []string{"家庭", "奇葩", "极品"}},
	{"家庭伤感", []string{"伤感", "虐心", "亲情"}},
	{"职场打脸", []string{"职场", "打脸", "逆袭", "爽文"}},
	{"现代甜文", []string{"甜宠", "甜文", "撒糖", "团宠"}},
	{"现代虐文", []string{"虐恋", "虐文", "追妻"}},
	{"现代通用", []string{"现代", "青春", "校园"}},
}

func defaultStyleForGender(gender string) string {
	if gender == "男频" {
		return "男频都市"
	}
	return "现代通用"
}

func genderForStyle(style string) string {
	switch style {
	case "男频都市", "玄幻", "历史":
		return "男频"
	default:
		return "女频"
	}
}

func matchNovelStyleMetadata(info BookInfo) string {
	metadata := strings.TrimSpace(info.Labels + "、" + info.Themes)
	if strings.Contains(metadata, "言情") || strings.Contains(metadata, "婚恋") {
		return "现代女主"
	}
	if strings.Contains(metadata, "现代") {
		return "现代通用"
	}
	return ""
}
