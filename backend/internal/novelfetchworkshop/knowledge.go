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

// OptimizeKnowledge asks the unified text model to refine one saved knowledge
// item, then writes only non-empty returned fields back to the same item.
func OptimizeKnowledge(ctx context.Context, settings map[string]any, kind, id string, provider RewriteProvider) (map[string]any, error) {
	required := map[string][]string{
		"high_imitation":    {"title", "source", "style", "content"},
		"opening_phrases":   {"text", "style", "tags", "source"},
		"rewrite_templates": {"name", "style", "prompt"},
	}
	fields, ok := required[kind]
	if !ok {
		return nil, fmt.Errorf("资料库类型无效：%s", kind)
	}
	if provider == nil {
		return nil, errors.New("AI 改文模型未配置")
	}
	library, err := ListKnowledge(settings, kind)
	if err != nil {
		return nil, err
	}
	targetID := strings.TrimSpace(id)
	var current map[string]any
	for _, item := range normalizeKnowledgeItems(library["items"]) {
		if strings.TrimSpace(fmt.Sprint(item["id"])) == targetID {
			current = item
			break
		}
	}
	if current == nil {
		return nil, errors.New("知识条目不存在")
	}
	libraryNames := map[string]string{"high_imitation": "高仿文章库", "opening_phrases": "爆款开头词库", "rewrite_templates": "批量改文指令库"}
	payload := map[string]any{
		"library_name": libraryNames[kind], "library_type": kind, "current_item": current,
		"user_goal":       "让这个条目更适合批量自动改文，保留原意，输出更清楚。",
		"required_fields": fields,
		"rules":           []string{"只整理当前资料库条目，不要发散新功能。", "不要删除原有 id；没有把握的字段保持原值。"},
		"return_format":   map[string]any{"item": map[string]any{}},
	}
	requestBody, _ := json.Marshal(payload)
	content, err := provider.Complete(ctx, RewriteRequest{
		SystemPrompt: "你是批量小说改文系统的提示词整理助手。只根据用户给的目标优化当前资料库条目，不要发散新功能。必须返回 JSON 对象，字段 item 内只放可保存字段。",
		UserPrompt:   string(requestBody), Temperature: 0.35, MaxTokens: 3000,
	})
	if err != nil {
		return nil, err
	}
	content = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(content, "```"), "```json"))
	start, end := strings.Index(content, "{"), strings.LastIndex(content, "}")
	if start < 0 || end < start {
		return nil, errors.New("AI返回格式不正确")
	}
	var parsed struct {
		Item map[string]any `json:"item"`
	}
	if err := json.Unmarshal([]byte(content[start:end+1]), &parsed); err != nil || len(parsed.Item) == 0 {
		return nil, errors.New("AI返回格式不正确")
	}
	optimized := cloneMap(current)
	optimized["id"] = targetID
	for key, value := range parsed.Item {
		if value == nil {
			continue
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
			continue
		}
		optimized[key] = value
	}
	return SaveKnowledge(settings, kind, optimized)
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
