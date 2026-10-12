package novelfetchworkshop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
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

func buildRewritePrompt(document Document, original string, index int) (string, int, int) {
	lines := rewriteLines(original)
	processCount := len(lines)
	if processCount > rewriteProcessLines {
		processCount = rewriteProcessLines
	}
	if processCount == 0 {
		processCount = 1
	}
	anchorEnd := processCount + rewriteAnchorLines
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
	if anchor != "" {
		prompt += "\n\n后续衔接参考（不要改写这部分）：\n" + anchor
	}
	return prompt, processCount, len(strings.Join(lines[:processCount], "\n"))
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
	document = normalizeDocument(document)
	for index := 1; index <= count; index++ {
		userPrompt, processCount, targetChars := buildRewritePrompt(document, original, index)
		request := RewriteRequest{
			SystemPrompt: "你是短剧小说改文编辑。只处理用户标出的待处理片段，不得扩写、续写或解释；保持人物身份、关键因果、时间顺序和男女频风格不变。只输出改写后的正文。",
			UserPrompt:   userPrompt,
			Temperature:  rewriteTemperature,
			MaxTokens:    rewriteMaxTokens,
		}
		result, callErr := provider.Complete(ctx, request)
		result = strings.TrimSpace(result)
		if callErr != nil {
			document.Meta["aiStatus"] = "failed"
			document.Meta["aiError"] = callErr.Error()
			_ = store.PutDocument(ctx, owner, document)
			return document, callErr
		}
		if result == "" {
			err = errors.New("AI 改文返回为空")
			document.Meta["aiStatus"] = "failed"
			document.Meta["aiError"] = err.Error()
			_ = store.PutDocument(ctx, owner, document)
			return document, err
		}
		if !rewriteWithinSize(result, targetChars) {
			request.UserPrompt += "\n\n上一次输出过长。请删除解释和扩写，只保留与待处理片段等量的改写正文。"
			result, callErr = provider.Complete(ctx, request)
			result = strings.TrimSpace(result)
			if callErr != nil {
				document.Meta["aiStatus"] = "failed"
				document.Meta["aiError"] = callErr.Error()
				_ = store.PutDocument(ctx, owner, document)
				return document, callErr
			}
			if result == "" || !rewriteWithinSize(result, targetChars) {
				err = errors.New("AI 改文输出过长，请检查改文提示词或模型配置")
				document.Meta["aiStatus"] = "failed"
				document.Meta["aiError"] = err.Error()
				_ = store.PutDocument(ctx, owner, document)
				return document, err
			}
		}
		result = mergeRewriteBlock(original, result, processCount)
		document.Versions[fmt.Sprintf("ai%d", index)] = result
		versionID := fmt.Sprintf("ai%d", index)
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
	}
	document.Meta["aiStatus"] = "done"
	document.Meta["aiGeneratedCount"] = count
	document.Meta["aiError"] = ""
	document.Meta["aiUpdatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	if err := store.PutDocument(ctx, owner, document); err != nil {
		return Document{}, err
	}
	return document, nil
}
