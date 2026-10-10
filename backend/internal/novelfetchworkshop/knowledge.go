package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var knowledgeKinds = map[string]bool{
	"high_imitation":    true,
	"opening_phrases":   true,
	"rewrite_templates": true,
}

var openingStyles = []string{"现代通用", "现代女主", "现代虐文", "现代甜文", "现代悬疑", "古风通用", "古风虐文", "古风甜文", "年代通用", "年代虐文", "年代甜文", "男频都市", "玄幻", "历史", "家庭奇葩", "家庭伤感", "职场打脸"}

func normalizeKnowledgeItems(raw any) []map[string]any {
	values, _ := raw.([]any)
	if typed, ok := raw.([]map[string]any); ok {
		return append([]map[string]any(nil), typed...)
	}
	items := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if item, ok := value.(map[string]any); ok {
			items = append(items, cloneMap(item))
		}
	}
	return items
}

// NormalizeKnowledgeItemsForHTTP exposes a safe copy for the HTTP layer.
func NormalizeKnowledgeItemsForHTTP(raw any) []map[string]any { return normalizeKnowledgeItems(raw) }

func knowledgeLibrary(settings map[string]any, kind string) map[string]any {
	libraries, _ := settings["knowledgeLibraries"].(map[string]any)
	if libraries == nil {
		libraries = map[string]any{}
	}
	library, _ := libraries[kind].(map[string]any)
	if library == nil {
		library = map[string]any{"source": "go", "updated_at": "", "items": []any{}}
	}
	return library
}

func putKnowledgeLibrary(settings map[string]any, kind string, library map[string]any) {
	libraries, _ := settings["knowledgeLibraries"].(map[string]any)
	if libraries == nil {
		libraries = map[string]any{}
	}
	libraries[kind] = library
	settings["knowledgeLibraries"] = libraries
}

func ListKnowledge(settings map[string]any, kind string) (map[string]any, error) {
	if !knowledgeKinds[kind] && kind != "layout_rules" && kind != "symbol_rules" && kind != "chapter_rules" {
		return nil, fmt.Errorf("资料库类型无效：%s", kind)
	}
	if !knowledgeKinds[kind] {
		return map[string]any{"source": "go", "updated_at": "", "items": []any{}}, nil
	}
	library := knowledgeLibrary(settings, kind)
	library["items"] = normalizeKnowledgeItems(library["items"])
	return library, nil
}

func SaveKnowledge(settings map[string]any, kind string, item map[string]any) (map[string]any, error) {
	if !knowledgeKinds[kind] {
		return nil, errors.New("该资料库只读或类型无效")
	}
	if item == nil {
		return nil, errors.New("知识条目格式无效")
	}
	library := knowledgeLibrary(settings, kind)
	items := normalizeKnowledgeItems(library["items"])
	id := strings.TrimSpace(fmt.Sprint(item["id"]))
	if id == "" || id == "<nil>" {
		prefix := map[string]string{"high_imitation": "hi", "opening_phrases": "opening", "rewrite_templates": "template"}[kind]
		id = fmt.Sprintf("%s_%03d", prefix, len(items)+1)
	}
	normalized := cloneMap(item)
	normalized["id"] = id
	if kind == "opening_phrases" {
		text := strings.TrimSpace(fmt.Sprint(firstValue(item, "text", "general_phrase", "content", "template")))
		if text == "" || text == "<nil>" {
			return nil, errors.New("开头词内容不能为空")
		}
		normalized["text"] = text
		normalized["style"] = normalizeOpeningStyle(fmt.Sprint(firstValue(item, "style", "category")))
		normalized["tags"] = normalizeTags(item["tags"])
		if strings.TrimSpace(fmt.Sprint(normalized["source"])) == "<nil>" || strings.TrimSpace(fmt.Sprint(normalized["source"])) == "" {
			normalized["source"] = "go"
		}
	}
	replaced := false
	for index, existing := range items {
		if strings.TrimSpace(fmt.Sprint(existing["id"])) == id {
			items[index] = normalized
			replaced = true
			break
		}
	}
	if !replaced {
		items = append(items, normalized)
	}
	library["items"] = items
	library["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	if kind == "opening_phrases" {
		styles := make([]any, 0)
		seen := map[string]bool{}
		for _, current := range items {
			style := strings.TrimSpace(fmt.Sprint(current["style"]))
			if style != "" && style != "<nil>" && !seen[style] {
				seen[style] = true
				styles = append(styles, style)
			}
		}
		library["styles"] = styles
	}
	putKnowledgeLibrary(settings, kind, library)
	return normalized, nil
}

func DeleteKnowledge(settings map[string]any, kind, id string) (bool, error) {
	if !knowledgeKinds[kind] {
		return false, errors.New("该资料库只读或类型无效")
	}
	library := knowledgeLibrary(settings, kind)
	items := normalizeKnowledgeItems(library["items"])
	filtered := make([]map[string]any, 0, len(items))
	removed := false
	for _, item := range items {
		if strings.TrimSpace(fmt.Sprint(item["id"])) == strings.TrimSpace(id) {
			removed = true
			continue
		}
		filtered = append(filtered, item)
	}
	library["items"] = filtered
	library["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	putKnowledgeLibrary(settings, kind, library)
	return removed, nil
}

func KnowledgeSummary(settings map[string]any) map[string]int {
	result := map[string]int{}
	for kind := range knowledgeKinds {
		library := knowledgeLibrary(settings, kind)
		result[kind] = len(normalizeKnowledgeItems(library["items"]))
	}
	result["layout_rules"], result["symbol_rules"], result["chapter_rules"] = 0, 0, 0
	return result
}

func AnalyzeOpening(ctx context.Context, provider RewriteProvider, text string) ([]map[string]any, error) {
	if provider == nil {
		return nil, errors.New("AI 改文模型未配置")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("请先粘贴爆款开头原文")
	}
	content, err := provider.Complete(ctx, RewriteRequest{
		SystemPrompt: "你是中文短剧爆款开头拆解助手。只返回 JSON 对象，格式为 {\"items\":[{\"text\":\"开头句式\",\"style\":\"风格\",\"tags\":[\"标签\"],\"source\":\"来源\"}]}，不要解释。",
		UserPrompt:   "请拆解以下开头，保留句式骨架和剧情槽位：\n" + text,
		Temperature:  0.2,
		MaxTokens:    2000,
	})
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Items []map[string]any `json:"items"`
	}
	if json.Unmarshal([]byte(content), &parsed) == nil && len(parsed.Items) > 0 {
		items := make([]map[string]any, 0, len(parsed.Items))
		for index, item := range parsed.Items {
			normalized, normalizeErr := SaveKnowledge(map[string]any{}, "opening_phrases", item)
			if normalizeErr == nil {
				normalized["id"] = fmt.Sprintf("opening_%03d", index+1)
				items = append(items, normalized)
			}
		}
		if len(items) > 0 {
			return items, nil
		}
	}
	return []map[string]any{{"id": "opening_001", "text": strings.TrimSpace(content), "style": "现代通用", "tags": []any{}, "source": "go"}}, nil
}

func firstValue(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := item[key]; ok && value != nil {
			return value
		}
	}
	return ""
}

func normalizeTags(value any) []any {
	result := []any{}
	seen := map[string]bool{}
	values, ok := value.([]any)
	if !ok {
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
			values = []any{text}
		}
	}
	for _, raw := range values {
		for _, part := range strings.FieldsFunc(fmt.Sprint(raw), func(r rune) bool { return r == ',' || r == '，' || r == '、' }) {
			part = strings.TrimSpace(part)
			if part != "" && !seen[part] {
				seen[part] = true
				result = append(result, part)
			}
		}
	}
	return result
}

func normalizeOpeningStyle(value string) string {
	value = strings.TrimSpace(value)
	for _, style := range openingStyles {
		if value == style || (value != "" && strings.Contains(value, style)) {
			return style
		}
	}
	if strings.Contains(value, "都市") || strings.Contains(value, "男频") {
		return "男频都市"
	}
	return "现代通用"
}
