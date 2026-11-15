package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/prompt/batchbook"
	"strings"
	"time"
)

var knowledgeKinds = map[string]bool{
	"high_imitation":    true,
	"opening_phrases":   true,
	"rewrite_templates": true,
}

var openingStyles = []string{"现代通用", "现代女主", "现代虐文", "现代甜文", "现代悬疑", "古风通用", "古风虐文", "古风甜文", "年代通用", "年代虐文", "年代甜文", "男频都市", "玄幻", "历史", "家庭奇葩", "家庭伤感", "职场打脸"}

// SyncKnowledgeConfigViews makes knowledgeLibraries the canonical persisted
// view while keeping appConfig.knowledge compatible with the legacy page.
// Rules are edited through appConfig, whereas library entries are edited
// through knowledgeLibraries, so the changed view must win for each request.
func SyncKnowledgeConfigViews(settings, patch map[string]any) {
	if settings == nil {
		return
	}
	libraries, _ := settings["knowledgeLibraries"].(map[string]any)
	if libraries == nil {
		libraries = map[string]any{}
	}
	apply := func(source map[string]any) {
		for kind, value := range source {
			if kind == "usage_prompt" {
				libraries[kind] = value
				continue
			}
			if kind == "high_imitation" || kind == "opening_phrases" || kind == "rewrite_templates" || kind == "sensitive" || kind == "layout_rules" || kind == "symbol_rules" || kind == "chapter_rules" || kind == "opening_analyzer" {
				if library, ok := value.(map[string]any); ok {
					libraries[kind] = cloneMap(library)
				}
			}
		}
	}
	if root, ok := patch["knowledge"].(map[string]any); ok {
		apply(root)
	}
	if root, ok := patch["knowledgeLibraries"].(map[string]any); ok {
		apply(root)
	}
	if app, ok := patch["appConfig"].(map[string]any); ok {
		if sensitive, exists := app["sensitive"]; exists {
			apply(map[string]any{"sensitive": sensitive})
		}
		if root, ok := app["knowledge"].(map[string]any); ok {
			apply(root)
		}
	}
	settings["knowledgeLibraries"] = libraries
	app, _ := settings["appConfig"].(map[string]any)
	if app == nil {
		app = map[string]any{}
	}
	app["knowledge"] = cloneMap(libraries)
	settings["appConfig"] = app
}

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
	// Older configs kept the same data under appConfig.knowledge. Read that
	// view as a compatibility fallback while all writes use knowledgeLibraries.
	if len(libraries) == 0 {
		app, _ := settings["appConfig"].(map[string]any)
		libraries, _ = app["knowledge"].(map[string]any)
	}
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
	app, _ := settings["appConfig"].(map[string]any)
	if app == nil {
		app = map[string]any{}
	}
	appKnowledge, _ := app["knowledge"].(map[string]any)
	if appKnowledge == nil {
		appKnowledge = map[string]any{}
	}
	appKnowledge[kind] = cloneMap(library)
	app["knowledge"] = appKnowledge
	settings["appConfig"] = app
}

func ListKnowledge(settings map[string]any, kind string) (map[string]any, error) {
	if !knowledgeKinds[kind] && kind != "layout_rules" && kind != "symbol_rules" && kind != "chapter_rules" {
		return nil, fmt.Errorf("资料库类型无效：%s", kind)
	}
	// The public response is a view.  Keep the caller from changing the saved
	// configuration while also restoring the legacy sub-library identity that
	// the compact items list used to hide.
	library := cloneMap(knowledgeLibrary(settings, kind))
	if knowledgeKinds[kind] {
		library["items"] = knowledgeItemsWithCollections(library, kind)
	}
	return library, nil
}

func knowledgeCollectionNames(kind string) []string {
	switch kind {
	case "high_imitation":
		return []string{"prompts", "references"}
	case "rewrite_templates":
		return []string{"profiles", "temporary_instructions"}
	default:
		return nil
	}
}

func knowledgeItemCollection(library map[string]any, kind, id, requested string) string {
	for _, name := range knowledgeCollectionNames(kind) {
		if requested == name {
			return name
		}
		for _, item := range normalizeKnowledgeItems(library[name]) {
			if strings.TrimSpace(fmt.Sprint(item["id"])) == id {
				return name
			}
		}
	}
	if kind == "high_imitation" {
		if strings.Contains(strings.ToLower(id), "reference") {
			return "references"
		}
		return "prompts"
	}
	if kind == "rewrite_templates" {
		if strings.HasPrefix(strings.ToLower(id), "instruction_") {
			return "temporary_instructions"
		}
		return "profiles"
	}
	return ""
}

func knowledgeItemsWithCollections(library map[string]any, kind string) []map[string]any {
	items := normalizeKnowledgeItems(library["items"])
	collections := knowledgeCollectionNames(kind)
	if len(collections) == 0 {
		return items
	}
	collectionByID := map[string]string{}
	for _, name := range collections {
		for _, item := range normalizeKnowledgeItems(library[name]) {
			collectionByID[strings.TrimSpace(fmt.Sprint(item["id"]))] = name
		}
	}
	for _, item := range items {
		id := strings.TrimSpace(fmt.Sprint(item["id"]))
		if collection := knowledgeItemCollection(library, kind, id, strings.TrimSpace(fmt.Sprint(item["collection"]))); collection != "" {
			item["collection"] = collection
		} else if collectionByID[id] != "" {
			item["collection"] = collectionByID[id]
		}
	}
	return items
}

func replaceKnowledgeCollectionItem(library map[string]any, kind, id, collection string, item map[string]any) {
	for _, name := range knowledgeCollectionNames(kind) {
		current := normalizeKnowledgeItems(library[name])
		updated := make([]any, 0, len(current)+1)
		found := false
		for _, existing := range current {
			if strings.TrimSpace(fmt.Sprint(existing["id"])) != id {
				updated = append(updated, existing)
				continue
			}
			if name == collection {
				updated = append(updated, cloneMap(item))
				found = true
			}
		}
		if name == collection && !found {
			updated = append(updated, cloneMap(item))
		}
		library[name] = updated
	}
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
		existingIDs := map[string]bool{}
		for _, existing := range items {
			existingIDs[strings.TrimSpace(fmt.Sprint(existing["id"]))] = true
		}
		for index := len(items) + 1; ; index++ {
			id = fmt.Sprintf("%s_%03d", prefix, index)
			if !existingIDs[id] {
				break
			}
		}
	}
	normalized := cloneMap(item)
	normalized["id"] = id
	if collection := knowledgeItemCollection(library, kind, id, strings.TrimSpace(fmt.Sprint(item["collection"]))); collection != "" {
		normalized["collection"] = collection
	}
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
	if collection := strings.TrimSpace(fmt.Sprint(normalized["collection"])); collection != "" {
		replaceKnowledgeCollectionItem(library, kind, id, collection, normalized)
	}
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
	for _, name := range knowledgeCollectionNames(kind) {
		collection := normalizeKnowledgeItems(library[name])
		updated := make([]any, 0, len(collection))
		for _, item := range collection {
			if strings.TrimSpace(fmt.Sprint(item["id"])) != strings.TrimSpace(id) {
				updated = append(updated, item)
			}
		}
		library[name] = updated
	}
	library["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	putKnowledgeLibrary(settings, kind, library)
	return removed, nil
}

// OptimizeKnowledge asks the unified text model to refine one saved knowledge
// item, then writes only non-empty returned fields back to the same item.
func OptimizeKnowledge(ctx context.Context, settings map[string]any, kind, id string, provider RewriteProvider, options ...map[string]any) (map[string]any, error) {
	required := map[string][]string{
		"high_imitation":    {"name", "system_prompt", "user_prompt_template", "analysis"},
		"opening_phrases":   {"name", "content", "analysis", "slot_hint", "style", "tags"},
		"rewrite_templates": {"name", "system_prompt", "user_prompt_template", "required_variables", "response_mode"},
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
	goal := "让这个条目更适合批量自动改文，保留原意，输出更清楚。"
	if len(options) > 0 {
		if draft, ok := options[0]["item"].(map[string]any); ok {
			current = cloneMap(current)
			for key, value := range draft {
				current[key] = value
			}
			current["id"] = targetID
		}
		if text, ok := options[0]["goal"].(string); ok && strings.TrimSpace(text) != "" {
			goal = strings.TrimSpace(text)
		}
	}
	if kind == "high_imitation" && current["collection"] == "references" {
		fields = []string{"name", "reference_text", "analysis", "style", "gender", "tags"}
	}
	if kind == "rewrite_templates" && current["collection"] == "temporary_instructions" {
		fields = []string{"name", "content"}
	}
	libraryNames := map[string]string{"high_imitation": "高仿文章库", "opening_phrases": "爆款开头词库", "rewrite_templates": "批量改文指令库"}
	payload := map[string]any{
		"library_name": libraryNames[kind], "library_type": kind, "current_item": current,
		"user_goal":       goal,
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
		if key == "id" || key == "collection" {
			continue
		}
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
		if kind == "high_imitation" || kind == "rewrite_templates" {
			result[kind] = len(normalizeKnowledgeItems(library["items"]))
			if result[kind] == 0 {
				result[kind] = len(arrayMap(library["prompts"])) + len(arrayMap(library["references"])) + len(arrayMap(library["profiles"])) + len(arrayMap(library["temporary_instructions"]))
			}
		} else {
			result[kind] = len(normalizeKnowledgeItems(library["items"]))
		}
	}
	opening := knowledgeLibrary(settings, "opening_phrases")
	high := knowledgeLibrary(settings, "high_imitation")
	rewrite := knowledgeLibrary(settings, "rewrite_templates")
	styles := map[string]bool{}
	for _, style := range normalizeKnowledgeItems(opening["items"]) {
		value := strings.TrimSpace(fmt.Sprint(style["style"]))
		if value != "" && value != "<nil>" {
			styles[value] = true
		}
	}
	storedStyles, _ := opening["styles"].([]any)
	for _, style := range storedStyles {
		value := strings.TrimSpace(fmt.Sprint(style))
		if value != "" && value != "<nil>" {
			styles[value] = true
		}
	}
	result["opening_styles"] = len(styles)
	for _, item := range knowledgeItemsWithCollections(rewrite, "rewrite_templates") {
		switch item["collection"] {
		case "profiles":
			result["rewrite_profiles"]++
		case "temporary_instructions":
			result["temporary_instructions"]++
		}
	}
	for _, item := range knowledgeItemsWithCollections(high, "high_imitation") {
		switch item["collection"] {
		case "prompts":
			result["high_imitation_prompts"]++
		case "references":
			result["high_imitation_references"]++
		}
	}
	layout := knowledgeLibrary(settings, "layout_rules")
	symbol := knowledgeLibrary(settings, "symbol_rules")
	chapter := knowledgeLibrary(settings, "chapter_rules")
	result["sensitive_rules"] = len(arrayMap(layout["sensitive_rules"])) + len(arrayMap(layout["sensitiveKeywords"]))
	if source, exists := sensitiveRuleSource(settings); exists {
		result["sensitive_rules"] = countSensitiveRuleEntries(source)
	}
	result["symbol_rules"] = len(arrayMap(symbol["symbol_rules"]))
	result["pair_fill_rules"] = len(arrayMap(symbol["pair_fill_rules"]))
	result["chapter_exact_rules"] = len(arrayMap(chapter["chapter_exact_rules"]))
	result["chapter_inline_rules"] = len(arrayMap(chapter["chapter_inline_rules"]))
	return result
}

func mapOrEmpty(value any) map[string]any {
	result, _ := value.(map[string]any)
	if result == nil {
		return map[string]any{}
	}
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
		SystemPrompt: batchbook.OpeningSystem,
		UserPrompt:   batchbook.OpeningUser + text,
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
		for _, item := range parsed.Items {
			normalized, normalizeErr := SaveKnowledge(map[string]any{}, "opening_phrases", item)
			if normalizeErr == nil {
				delete(normalized, "id")
				items = append(items, normalized)
			}
		}
		if len(items) > 0 {
			return items, nil
		}
	}
	return nil, errors.New("AI 未返回有效的开头拆解结果，请调整原文或要求后重试")
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
