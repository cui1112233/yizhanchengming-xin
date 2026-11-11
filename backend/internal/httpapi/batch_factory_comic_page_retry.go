package httpapi

import (
	"context"
	"fmt"

	"qiantie/backend/internal/batchfactoryv11"
)

// comicPageRetryFunc 由路由层注入，用于书级“重试失败步骤”在没有失败阶段时
// 回退到页级重试。返回本次重新提交出图的页数。
type comicPageRetryFunc func(ctx context.Context, owner, batchID, bookID string) (int, error)

type failedComicPage struct {
	videoID string
	input   batchfactoryv11.UpsertComicPageInput
}

// failedComicPages 收集需要重新出图的漫画页，只保留页级生成失败的记录，
// 避免把已成功或仍在生成中的页面重复提交给上游（重复提交会重复计费）。
func failedComicPages(pages []batchfactoryv11.ComicPage) []failedComicPage {
	failed := make([]failedComicPage, 0, len(pages))
	for _, page := range pages {
		if page.Status != "generation_failed" {
			continue
		}
		failed = append(failed, failedComicPage{
			videoID: page.VideoID,
			input: batchfactoryv11.UpsertComicPageInput{
				ComicPrompt:        page.ComicPrompt,
				ImagePrompt:        page.ImagePrompt,
				ReferenceImageURLs: page.ReferenceImageURLs,
				ModelID:            page.ModelID,
				AspectRatio:        page.AspectRatio,
			},
		})
	}
	return failed
}

// retryFailedComicPages 重新生成该书所有页级失败的漫画页。
// 与单页出图接口保持一致：使用后台上下文，浏览器断开不会丢弃上游已返回的图片。
func (g comicPageGenerator) retryFailedComicPages(ctx context.Context, owner, batchID, bookID string) (int, error) {
	comicStore, ok := g.store.(batchfactoryv11.ComicPageStore)
	if !ok || g.files == nil || g.catalogStore == nil {
		return 0, fmt.Errorf("%w: 漫画图片生成服务尚未配置", batchfactoryv11.ErrUnavailable)
	}
	pages, err := comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return 0, err
	}
	failed := failedComicPages(pages)
	if len(failed) == 0 {
		return 0, nil
	}
	generationCtx, cancel := comicPageGenerationContext(context.Background())
	defer cancel()
	generationCtx = withShuihuoImageGenerationCacheBypass(withShuihuoImageGenerationCache(generationCtx, g.imageCache))
	retried := 0
	for _, page := range failed {
		if _, _, genErr := g.generatePage(generationCtx, owner, batchID, bookID, page.videoID, page.input); genErr != nil {
			return retried, genErr
		}
		retried++
	}
	return retried, nil
}
