package novelfetchworkshop

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/prompt/batchbook"
)

// RewriteRequest 是小说改文模型收到的最小请求。提示词放在后端，前端只能选择数量。
type RewriteRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

// RewriteProvider 由应用层接入统一文本模型，密钥不会进入任务记录。
type RewriteProvider interface {
	Complete(context.Context, RewriteRequest) (string, error)
}

type RewriteTextProcessor func(context.Context, string, string) (string, []map[string]any, []map[string]any, error)

const (
	rewriteProcessLines = 20
	rewriteAnchorLines  = 5
	rewriteTemperature  = 0.45
	rewriteMaxTokens    = 1200
)

func documentMetaText(document Document, keys ...string) string {
	for _, key := range keys {
		if value, ok := document.Meta[key]; ok {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func rewriteLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	for index := range lines {
		lines[index] = strings.TrimSpace(lines[index])
	}
	return lines
}

type rewriteRuntimePlan struct {
	ProcessLines               int
	AnchorLines                int
	Temperature                float64
	MaxTokens                  int
	Method                     string
	TemplateID                 string
	TemplateName               string
	StrategyName               string
	OpeningPhraseName          string
	HighImitationReferenceName string
	Instructions               []string
	SystemPrompt               string
	UserTemplate               string
	ResponseMode               string
}

func rewriteItemText(item map[string]any) string {
	for _, key := range []string{"prompt", "content", "instruction", "text", "body", "example"} {
		if text := strings.TrimSpace(fmt.Sprint(item[key])); text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func rewriteItemName(item map[string]any) string {
	for _, key := range []string{"name", "title", "label", "id"} {
		if text := strings.TrimSpace(fmt.Sprint(item[key])); text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func firstKnowledgeText(library map[string]any, fields ...string) string {
	for _, field := range fields {
		for _, item := range arrayMap(library[field]) {
			if item["enabled"] == false {
				continue
			}
			if text := rewriteItemText(item); text != "" {
				return text
			}
		}
	}
	return ""
}

func rewritePlan(settings map[string]any, versionID string, index int) rewriteRuntimePlan {
	app := mapOrEmpty(settings["appConfig"])
	rewrite := mapOrEmpty(settings["rewrite"])
	if configured := mapOrEmpty(app["rewrite"]); len(configured) > 0 {
		rewrite = configured
	}
	ai := mapOrEmpty(settings["ai"])
	if configured := mapOrEmpty(app["ai"]); len(configured) > 0 {
		ai = configured
	}
	plan := rewriteRuntimePlan{ProcessLines: ruleInt(rewrite["process_line_count"], rewriteProcessLines), AnchorLines: ruleInt(rewrite["anchor_line_count"], rewriteAnchorLines), Temperature: rewriteTemperature, MaxTokens: rewriteMaxTokens, TemplateID: rewriteString(rewrite["default_template_id"])}
	if value, ok := rewrite["temperature"].(float64); ok {
		plan.Temperature = value
	}
	if value, ok := ai["max_tokens"].(float64); ok && value > 0 {
		plan.MaxTokens = int(value)
	}
	if slots := mapOrEmpty(rewrite["ai_slot_methods"]); rewriteString(slots[versionID]) != "" && rewriteString(slots[versionID]) != "auto" {
		plan.Method = rewriteString(slots[versionID])
	}
	if plan.Method == "" {
		methods := anySlice(rewrite["method_sequence"])
		if len(methods) > 0 {
			plan.Method = strings.TrimSpace(fmt.Sprint(methods[(index-1)%len(methods)]))
		}
	}
	if plan.Method == "" {
		plan.Method = "instruction"
	}
	plan.StrategyName = map[string]string{"opening_instruction": "爆款开头", "high_imitation": "高仿改写", "instruction": "普通改文"}[plan.Method]
	if plan.StrategyName == "" {
		plan.StrategyName = plan.Method
	}
	if prompt := strings.TrimSpace(fmt.Sprint(rewrite["prompt"])); prompt != "" && prompt != "<nil>" {
		plan.Instructions = append(plan.Instructions, "账号改文要求："+prompt)
	}
	for _, knowledgeConfig := range []map[string]any{mapOrEmpty(settings["knowledge"]), mapOrEmpty(app["knowledge"]), mapOrEmpty(settings["knowledgeLibraries"])} {
		if usage := strings.TrimSpace(fmt.Sprint(knowledgeConfig["usage_prompt"])); usage != "" && usage != "<nil>" {
			plan.Instructions = append(plan.Instructions, "知识库使用要求："+usage)
			break
		}
	}
	templates := knowledgeLibrary(settings, "rewrite_templates")
	for _, item := range append(arrayMap(templates["profiles"]), arrayMap(templates["items"])...) {
		if item["enabled"] == false {
			continue
		}
		id := strings.TrimSpace(fmt.Sprint(item["id"]))
		if plan.TemplateID != "" && id != plan.TemplateID {
			continue
		}
		if text := rewriteItemText(item); text != "" || rewriteString(item["user_prompt_template"]) != "" || rewriteString(item["system_prompt"]) != "" {
			plan.Instructions = append(plan.Instructions, "改文模板："+text)
			plan.SystemPrompt = rewriteString(item["system_prompt"])
			plan.UserTemplate = rewriteString(item["user_prompt_template"])
			plan.ResponseMode = rewriteString(item["response_mode"])
			plan.TemplateName = rewriteItemName(item)
			if plan.TemplateID == "" {
				plan.TemplateID = id
			}
			break
		}
	}
	if temporary := firstKnowledgeText(templates, "temporary_instructions"); temporary != "" {
		plan.Instructions = append(plan.Instructions, "临时指令："+temporary)
	}
	if plan.Method == "opening_instruction" {
		openingLibrary := knowledgeLibrary(settings, "opening_phrases")
		if opening := firstKnowledgeText(openingLibrary, "items", "phrases"); opening != "" {
			plan.Instructions = append(plan.Instructions, "爆款开头参考（只参考结构，不得照搬）："+opening)
			for _, item := range append(arrayMap(openingLibrary["items"]), arrayMap(openingLibrary["phrases"])...) {
				if rewriteItemText(item) == opening {
					plan.OpeningPhraseName = rewriteItemName(item)
					break
				}
			}
		}
	}
	if plan.Method == "high_imitation" {
		highLibrary := knowledgeLibrary(settings, "high_imitation")
		if high := firstKnowledgeText(highLibrary, "items", "prompts", "references"); high != "" {
			plan.Instructions = append(plan.Instructions, "高仿参考（只参考结构，不得照搬）："+high)
			for _, field := range []string{"items", "prompts", "references"} {
				for _, item := range arrayMap(highLibrary[field]) {
					if rewriteItemText(item) == high {
						plan.HighImitationReferenceName = rewriteItemName(item)
						break
					}
				}
				if plan.HighImitationReferenceName != "" {
					break
				}
			}
		}
	}
	return plan
}

func buildRewritePrompt(document Document, original string, index int, plan rewriteRuntimePlan) (string, int, int) {
	lines := rewriteLines(original)
	processCount := len(lines)
	if plan.ProcessLines > 0 && processCount > plan.ProcessLines {
		processCount = plan.ProcessLines
	}
	if processCount == 0 {
		processCount = 1
	}
	anchorEnd := processCount + plan.AnchorLines
	if anchorEnd > len(lines) {
		anchorEnd = len(lines)
	}
	block := strings.Join(lines[:processCount], "\n")
	anchor := strings.Join(lines[processCount:anchorEnd], "\n")
	style := documentMetaText(document, "style", "genre", "classification")
	if style == "" {
		style = strings.TrimSpace(document.BookInfo.Genre)
	}
	gender := strings.TrimSpace(document.Gender)
	if gender == "" {
		gender = strings.TrimSpace(document.BookInfo.Category)
	}
	bookName := documentMetaText(document, "bookName", "title", "name")
	if bookName == "" {
		bookName = "未命名小说"
	}
	prompt := fmt.Sprintf("这是第 %d 个改写版本。\n小说名：%s\n男女频：%s\n风格：%s\n\n只改写下面的待处理片段，不要扩写成摘要或续写。保持原有事件、人物关系和信息量，句子简洁，长度尽量接近原文（最多比原文增加 35%%）。每一行对应原文叙事节奏。只输出改写后的正文，不要标题、解释、序号或 Markdown。\n\n待处理片段：\n%s", index, bookName, gender, style, block)
	if len(plan.Instructions) > 0 {
		prompt = strings.Join(plan.Instructions, "\n") + "\n\n" + prompt
	}
	if anchor != "" {
		prompt += "\n\n后续衔接参考（不要改写这部分）：\n" + anchor
	}
	if plan.UserTemplate != "" {
		prompt = fillRewriteTemplate(plan.UserTemplate, map[string]string{
			"content_id": document.BookID, "book_id": document.BookID, "book_name": bookName,
			"style": style, "gender": gender, "full_text": original,
			"target_lines_text": block, "anchor_lines_text": anchor,
			"line_count": strconv.Itoa(processCount), "extra_instruction": strings.Join(plan.Instructions, "\n"),
		})
	}
	return prompt, processCount, len([]rune(block))
}

func mergeRewriteBlock(original, changed string, processCount int) string {
	originalLines := rewriteLines(original)
	changedLines := rewriteLines(changed)
	if processCount > len(originalLines) {
		processCount = len(originalLines)
	}
	merged := append([]string{}, changedLines...)
	if processCount < len(originalLines) {
		merged = append(merged, originalLines[processCount:]...)
	}
	return strings.TrimSpace(strings.Join(merged, "\n"))
}

func rewriteWithinSize(result string, targetChars int) bool {
	if targetChars <= 0 {
		return strings.TrimSpace(result) != ""
	}
	maxChars := int(float64(targetChars)*1.35) + 80
	return len([]rune(strings.TrimSpace(result))) <= maxChars
}

// GenerateAIRewrites 生成并保存 ai1、ai2... 版本。任何一次模型失败都不会保存半截版本。
func GenerateAIRewrites(ctx context.Context, store Store, owner, bookID string, count int, provider RewriteProvider) (Document, error) {
	if provider == nil {
		return Document{}, errors.New("AI 改文模型未配置")
	}
	if count < 1 || count > 20 {
		return Document{}, errors.New("生成数量需为 1~20 的整数")
	}
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return Document{}, err
	}
	original := strings.TrimSpace(document.Original)
	if original == "" {
		original = strings.TrimSpace(document.ProcessedText)
	}
	if original == "" {
		return Document{}, errors.New("请先抓取原文，再生成 AI 改文")
	}
	versions := make([]string, 0, count)
	for index := 1; index <= count; index++ {
		versions = append(versions, fmt.Sprintf("ai%d", index))
	}
	return GenerateAIRewritesForVersions(ctx, store, owner, bookID, versions, provider)
}

// GenerateAIRewritesForVersions 按调用方明确选择的版本生成改文。
// 不能用“最大编号”推断数量，否则用户只选 ai3 时会额外生成 ai1、ai2，
// 任务列表会出现用户没有要求的版本，后续提交也可能误选它们。
func GenerateAIRewritesForVersions(ctx context.Context, store Store, owner, bookID string, versions []string, provider RewriteProvider) (Document, error) {
	return GenerateAIRewritesForVersionsProcessed(ctx, store, owner, bookID, versions, provider, nil)
}

// GenerateAIRewritesForVersionsProcessed applies the same persisted rule
// pipeline used for originals before each AI body becomes visible.
func GenerateAIRewritesForVersionsProcessed(ctx context.Context, store Store, owner, bookID string, versions []string, provider RewriteProvider, processor RewriteTextProcessor, methodOverrides ...map[string]string) (Document, error) {
	if provider == nil {
		return Document{}, errors.New("AI 改文模型未配置")
	}
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return Document{}, err
	}
	original := strings.TrimSpace(document.Original)
	if original == "" {
		original = strings.TrimSpace(document.ProcessedText)
	}
	if original == "" {
		return Document{}, errors.New("请先抓取原文，再生成 AI 改文")
	}
	document = normalizeDocument(document)
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return document, err
	}
	if len(methodOverrides) > 0 && len(methodOverrides[0]) > 0 {
		app := mapOrEmpty(settings["appConfig"])
		rewrite := mapOrEmpty(settings["rewrite"])
		if configured := mapOrEmpty(app["rewrite"]); len(configured) > 0 {
			rewrite = configured
		}
		slots := cloneMap(mapOrEmpty(rewrite["ai_slot_methods"]))
		for version, method := range methodOverrides[0] {
			switch method {
			case "", "auto", "high_imitation", "opening_instruction", "instruction":
			default:
				return document, fmt.Errorf("不支持的改文方式：%s", method)
			}
			if method == "" {
				method = "auto"
			}
			slots[version] = method
		}
		rewrite["ai_slot_methods"] = slots
		app["rewrite"] = rewrite
		settings["appConfig"] = app
	}
	seen := map[string]bool{}
	requested := make([]string, 0, len(versions))
	for _, versionID := range versions {
		versionID = strings.ToLower(strings.TrimSpace(versionID))
		if versionID == "" || versionID == "original" || seen[versionID] {
			continue
		}
		if !strings.HasPrefix(versionID, "ai") {
			return Document{}, fmt.Errorf("不支持的改文版本：%s", versionID)
		}
		index, parseErr := strconv.Atoi(strings.TrimPrefix(versionID, "ai"))
		if parseErr != nil || index < 1 || index > 20 {
			return Document{}, fmt.Errorf("不支持的改文版本：%s", versionID)
		}
		seen[versionID] = true
		requested = append(requested, versionID)
	}
	if len(requested) == 0 {
		return document, nil
	}
	for _, versionID := range requested {
		index, _ := strconv.Atoi(strings.TrimPrefix(versionID, "ai"))
		plan := rewritePlan(settings, versionID, index)
		if plan.TemplateID != "" && plan.TemplateName == "" {
			return document, fmt.Errorf("所选改文模板不可用，请重新选择模板")
		}
		userPrompt, processCount, targetChars := buildRewritePrompt(document, original, index, plan)
		request := RewriteRequest{
			SystemPrompt: batchbook.RewriteSystem,
			UserPrompt:   userPrompt,
			Temperature:  plan.Temperature,
			MaxTokens:    plan.MaxTokens,
		}
		if plan.SystemPrompt != "" {
			request.SystemPrompt = fillRewriteTemplate(plan.SystemPrompt, map[string]string{"content_id": document.BookID, "book_id": document.BookID})
		}
		result, generationErr := completeValidatedRewrite(ctx, provider, request, document.BookID, plan.ResponseMode, processCount, targetChars)
		if generationErr != nil {
			document.Meta["rewriteValidationHold"] = errors.Is(generationErr, ErrRewriteValidation)
			document.Meta["aiStatus"] = "failed"
			document.Meta["aiError"] = generationErr.Error()
			_ = store.PutDocument(ctx, owner, document)
			return document, generationErr
		}
		result = mergeRewriteBlock(original, result, processCount)
		if processor != nil {
			processed, trace, sensitiveItems, processErr := processor(ctx, versionID, result)
			if processErr != nil {
				document.Meta["aiStatus"] = "failed"
				document.Meta["aiError"] = processErr.Error()
				_ = store.PutDocument(ctx, owner, document)
				return document, processErr
			}
			result = processed
			traceByVersion, _ := document.Meta["rulesTraceByVersion"].(map[string]any)
			if traceByVersion == nil {
				traceByVersion = map[string]any{}
			}
			traceByVersion[versionID] = trace
			document.Meta["rulesTraceByVersion"] = traceByVersion
			sensitiveByVersion, _ := document.Meta["sensitiveByVersion"].(map[string]any)
			if sensitiveByVersion == nil {
				sensitiveByVersion = map[string]any{}
			}
			sensitiveByVersion[versionID] = sensitiveItems
			document.Meta["sensitiveByVersion"] = sensitiveByVersion
		}
		document.Versions[versionID] = result
		model := ""
		if identity, ok := provider.(interface{ ModelIdentity() string }); ok {
			model = strings.TrimSpace(identity.ModelIdentity())
			if model != "" {
				document.Meta["rewriteModel"] = model
			}
		}
		document.Meta["aiCurrentVersion"] = versionID
		evidence, _ := document.Meta["rewriteEvidenceByVersion"].(map[string]any)
		if evidence == nil {
			evidence = map[string]any{}
		}
		evidence[versionID] = map[string]any{"method": plan.Method, "strategy_name": plan.StrategyName, "template_id": plan.TemplateID, "template_name": plan.TemplateName, "opening_phrase_name": plan.OpeningPhraseName, "high_imitation_reference_name": plan.HighImitationReferenceName, "process_line_count": plan.ProcessLines, "anchor_line_count": plan.AnchorLines, "temperature": plan.Temperature, "max_tokens": plan.MaxTokens}
		if model != "" {
			evidence[versionID].(map[string]any)["model"] = model
		}
		document.Meta["rewriteEvidenceByVersion"] = evidence
		ref, bodyErr := store.PutBody(ctx, owner, BodyRecord{
			BookID:  document.BookID,
			BodyRef: BodyRef{VersionID: versionID, State: "ready"},
			Content: result,
		})
		if bodyErr != nil {
			document.Meta["aiStatus"] = "failed"
			document.Meta["aiError"] = bodyErr.Error()
			_ = store.PutDocument(ctx, owner, document)
			return document, bodyErr
		}
		document.BodyRefs[versionID] = ref
		// Persist each successful version before starting another model call.
		if err := store.PutDocument(ctx, owner, document); err != nil {
			return document, err
		}
	}
	document.Meta["aiStatus"] = "done"
	delete(document.Meta, "rewriteValidationHold")
	document.Meta["aiGeneratedCount"] = len(requested)
	document.Meta["aiError"] = ""
	document.Meta["aiUpdatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	if err := store.PutDocument(ctx, owner, document); err != nil {
		return Document{}, err
	}
	return document, nil
}
