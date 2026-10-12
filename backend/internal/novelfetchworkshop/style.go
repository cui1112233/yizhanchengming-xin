package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"strings"
)

// StyleDecision is the result of the second-stage style classification.
// An empty Style means the text could not be classified safely.
type StyleDecision struct {
	Style  string `json:"style"`
	Source string `json:"styleSource"`
	Reason string `json:"styleReason"`
}

var knownNovelStyles = []string{"古风虐文", "古风甜文", "古风通用", "年代虐文", "年代甜文", "年代通用", "现代虐文", "现代甜文", "现代悬疑", "现代通用", "男频都市", "现代女主", "玄幻", "历史", "家庭奇葩", "家庭伤感", "职场打脸"}

// ResolveStyle uses explicit upstream metadata first and only asks the text
// model when the upstream value is missing or not one of the supported styles.
func ResolveStyle(ctx context.Context, info BookInfo, provider RewriteProvider) StyleDecision {
	if style := matchNovelStyle(info.Genre); style != "" {
		return StyleDecision{Style: style, Source: "bookinfo", Reason: "采用平台返回的明确风格"}
	}
	if style := matchNovelStyle(info.Category); style != "" {
		return StyleDecision{Style: style, Source: "bookinfo", Reason: "采用平台返回的明确风格"}
	}
	if provider == nil || strings.TrimSpace(info.Text) == "" {
		return StyleDecision{Reason: "平台未提供明确风格且 AI 风格判断未配置"}
	}
	content, err := provider.Complete(ctx, RewriteRequest{
		SystemPrompt: "你是小说分类器。只返回 JSON：{\"style\":\"风格\"}。style 必须从以下值中选择：古风虐文、古风甜文、古风通用、年代虐文、年代甜文、年代通用、现代虐文、现代甜文、现代悬疑、现代通用、男频都市、现代女主、玄幻、历史、家庭奇葩、家庭伤感、职场打脸。无法确定时返回空字符串。不要解释。",
		UserPrompt:   "请判断这段小说的投稿风格：\n" + info.Text,
		Temperature:  0,
		MaxTokens:    80,
	})
	if err != nil {
		return StyleDecision{Reason: "AI 风格判断失败"}
	}
	var payload struct {
		Style string `json:"style"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(content)), &payload) != nil {
		return StyleDecision{Reason: "AI 风格判断返回格式无效"}
	}
	if style := matchNovelStyle(payload.Style); style != "" {
		return StyleDecision{Style: style, Source: "ai", Reason: "平台风格不明确，采用 AI 判断结果"}
	}
	return StyleDecision{Reason: "AI 未返回有效风格"}
}

func matchNovelStyle(value string) string {
	value = strings.TrimSpace(value)
	for _, style := range knownNovelStyles {
		if value == style || strings.Contains(value, style) {
			return style
		}
	}
	return ""
}
