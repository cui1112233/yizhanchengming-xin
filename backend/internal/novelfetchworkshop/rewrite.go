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
		result, callErr := provider.Complete(ctx, RewriteRequest{
			SystemPrompt: "你是短剧小说改文编辑。保持人物身份、关键因果和时间顺序不变，增强冲突、悬念和可视化动作。只输出改写后的正文，不要标题、解释或 Markdown。",
			UserPrompt:   fmt.Sprintf("这是第 %d 个改写版本。\n\n原文：\n%s", index, original),
			Temperature:  0.75,
			MaxTokens:    12000,
		})
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
		document.Versions[fmt.Sprintf("ai%d", index)] = result
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
