package httpapi

import (
	"context"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelcreation"
)

// novelPlayText 把互动世界的单次补全请求转接到现有的模型解析器。
// 它不持有渠道凭据，只按账号与模型 ID 取已配置的 provider。
type novelPlayText struct {
	resolve novelcreation.ProviderResolver
}

func NewNovelPlayText(resolve novelcreation.ProviderResolver) novelcreation.PlayTextCompleter {
	return novelPlayText{resolve: resolve}
}

func (p novelPlayText) CompletePlayText(ctx context.Context, owner, modelID, systemPrompt, userPrompt string, maxTokens int, temperature float64) (string, error) {
	if p.resolve == nil {
		return "", batchfactoryv11.ErrUnavailable
	}
	provider, err := p.resolve(ctx, strings.TrimSpace(owner), strings.TrimSpace(modelID))
	if err != nil {
		return "", err
	}
	if provider == nil {
		return "", batchfactoryv11.ErrUnavailable
	}
	return provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    maxTokens,
		Temperature:  temperature,
	})
}
