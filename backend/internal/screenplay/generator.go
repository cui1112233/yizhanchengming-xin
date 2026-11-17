package screenplay

import (
	"context"
	"errors"
	"fmt"
)

// 规格 §1 / §2 的失败提示原文。业务层只做包装（%w），不吞错、不改写。
var (
	ErrAIEmptyContent = errors.New("AI 未返回内容")
	ErrAIEmptyBody    = errors.New("AI未返回任何正文（本地流式长度为0；CM 后台仍可能已计费）")
	ErrAIInvalidJSON  = errors.New("AI 未返回有效JSON")
	// ErrAIInvalidStoryboard 是规格 §6/§7 的分镜专用失败串。
	ErrAIInvalidStoryboard = errors.New("AI 未返回有效分镜数据")
	ErrAIJSONParse         = errors.New("AI 返回的 JSON 解析失败：")
	ErrAIJSONParseRaw      = errors.New("JSON解析失败:")
	ErrAIRequest           = errors.New("API请求失败: HTTP")
	ErrAPIKeyMissing       = errors.New("API Key未配置，请在全局设置中配置")
	ErrUserCancelled       = errors.New("用户取消了推理")
	ErrArtifactInvalid     = errors.New("产物校验失败：")
	ErrBatchNotStarted     = errors.New("批次未启动，请先 batch_start")
	ErrEmptyContent        = errors.New("内容不能为空")
	ErrEmptyNovel          = errors.New("novel.md 缺失，批次卡在 0_ingest")
	ErrNoActiveScript      = errors.New("当前项目还没有激活的剧本")
	// ErrPartialFailure 表示批次/分集只有一部分成功。返回值里仍带着已成功
	// 落库的结果和失败定位信息（批次号、集号），上层据此阻断或逐条提示，
	// 不能只靠 warnings 当作整体成功。
	ErrPartialFailure = errors.New("部分内容生成失败")
)

// TextRequest 是对模型的最小调用契约，便于替换 provider 与测试。
type TextRequest struct {
	Messages        []TextMessage
	Model           string
	SystemPrompt    string
	UserPrompt      string
	MaxOutputTokens int
	Temperature     float64
	JSONResponse    bool
	SingleAttempt   bool `json:"single_attempt,omitempty"`
}

type TextMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TextResponse struct {
	Content string
	Model   string
}

// TextGenerator 只负责一次补全；重试、分段、落库都在业务层完成。
type TextGenerator interface {
	Generate(ctx context.Context, req TextRequest) (TextResponse, error)
}

// wrapAIError 把 provider 错误映射到规格错误串，同时保留原始错误链。
func wrapAIError(err error) error {
	if err == nil {
		return nil
	}
	for _, known := range []error{
		ErrAIEmptyContent, ErrAIEmptyBody, ErrAIInvalidJSON, ErrAIJSONParse,
		ErrAIRequest, ErrAPIKeyMissing, ErrUserCancelled,
	} {
		if errors.Is(err, known) {
			return err
		}
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %w", ErrUserCancelled, err)
	}
	return fmt.Errorf("%w: %w", ErrAIRequest, err)
}
