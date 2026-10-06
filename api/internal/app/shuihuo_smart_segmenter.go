package app

import (
	"context"
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
)

func newShuihuoSmartSegmenter(provider generation.Provider) shuihuo.SmartSegmenter {
	return shuihuo.SmartSegmenterFunc(func(ctx context.Context, text string) ([]shuihuo.Candidate, error) {
		if provider == nil {
			return nil, shuihuo.ErrSmartUnavailable
		}
		output, err := provider.Complete(ctx, generation.TextRequest{
			Stage:        generation.StageScript,
			SystemPrompt: "你是短视频文本分段器。只返回 JSON 数组，每项包含 text，可选 speaker；不得添加解释。",
			UserPrompt:   text,
		})
		if errors.Is(err, generation.ErrUnavailable) {
			return nil, shuihuo.ErrSmartUnavailable
		}
		if err != nil {
			return nil, err
		}
		return shuihuo.ParseSmartCandidates(output)
	})
}
