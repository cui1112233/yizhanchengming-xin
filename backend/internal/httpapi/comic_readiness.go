package httpapi

import (
	"context"
	"fmt"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

const (
	comicPromptStatusMissing   = "missing"
	comicPromptStatusComicOnly = "comicOnly"
	comicPromptStatusReady     = "ready"

	comicImageStatusNone       = "none"
	comicImageStatusCandidates = "candidates"
	comicImageStatusPrimary    = "primary"

	comicGenerationStateIdle       = "idle"
	comicGenerationStateGenerating = "generating"
	comicGenerationStateFailed     = "failed"
	comicGenerationStateSucceeded  = "succeeded"
)

// comicReadiness 是书级权威就绪判定。“有图”统一按“有主图”计算：只存在
// 候选图（未设置 is_primary）的页不会计入 imageReadyCount，而是通过
// primaryFallback 与 blockers 打标，避免读接口产生写副作用。
type comicReadiness struct {
	PagesTotal         int                  `json:"pagesTotal"`
	PromptReadyCount   int                  `json:"promptReadyCount"`
	ImageReadyCount    int                  `json:"imageReadyCount"`
	PrimaryReadyCount  int                  `json:"primaryReadyCount"`
	StaleCount         int                  `json:"staleCount"`
	CanGenerateImages  bool                 `json:"canGenerateImages"`
	CanBuildCollection bool                 `json:"canBuildCollection"`
	Skippable          bool                 `json:"skippable"`
	Blockers           []string             `json:"blockers"`
	Limits             comicReadinessLimits `json:"limits"`
}

type comicReadinessLimits struct {
	MaxPages int `json:"maxPages"`
}

func deriveComicPageStatus(page batchfactoryv11.ComicPage, hasPrimaryImage bool, hasAnyImage bool) (promptStatus, imageStatus, generationState string, primaryFallback bool) {
	comicPrompt := strings.TrimSpace(page.ComicPrompt)
	imagePrompt := strings.TrimSpace(page.ImagePrompt)
	switch {
	case comicPrompt == "" && imagePrompt == "":
		promptStatus = comicPromptStatusMissing
	case comicPrompt == "" || imagePrompt == "":
		promptStatus = comicPromptStatusComicOnly
	default:
		promptStatus = comicPromptStatusReady
	}

	switch {
	case hasPrimaryImage:
		imageStatus = comicImageStatusPrimary
	case hasAnyImage:
		// 历史“有图但无主图”的行只打标，不自动补主图。
		imageStatus = comicImageStatusCandidates
		primaryFallback = true
	default:
		imageStatus = comicImageStatusNone
	}

	status := strings.ToLower(strings.TrimSpace(page.Status))
	switch {
	case status == "generating" || status == "queued" || status == "running" || status == "pending":
		generationState = comicGenerationStateGenerating
	case strings.Contains(status, "fail") || status == "error":
		generationState = comicGenerationStateFailed
	case status == "succeeded" || status == "success" || status == "completed" || status == "done":
		generationState = comicGenerationStateSucceeded
	case hasAnyImage:
		// 没有状态但已经有图，按已出图处理，避免 UI 显示“等待生成”。
		generationState = comicGenerationStateSucceeded
	default:
		generationState = comicGenerationStateIdle
	}
	return promptStatus, imageStatus, generationState, primaryFallback
}

func comicPageImageFlags(page batchfactoryv11.ComicPage) (hasPrimary bool, hasAny bool) {
	for _, image := range page.Images {
		hasAny = true
		if image.IsPrimary {
			hasPrimary = true
		}
	}
	return hasPrimary, hasAny
}

// buildComicReadiness 由页面行（唯一事实源）计算书级就绪度。videos 仅用于
// 统计“缺页面行”的页数；传空时以页面行数量作为总页数。
func buildComicReadiness(videos []batchfactoryv11.Video, pages []batchfactoryv11.ComicPage, maxPages int) comicReadiness {
	readiness := comicReadiness{
		Blockers: []string{},
		Limits:   comicReadinessLimits{MaxPages: maxPages},
	}
	pagesByVideo := make(map[string]batchfactoryv11.ComicPage, len(pages))
	for _, page := range pages {
		pagesByVideo[page.VideoID] = page
	}
	if len(videos) > 0 {
		readiness.PagesTotal = len(videos)
	} else {
		readiness.PagesTotal = len(pages)
	}

	missingPageRows := 0
	candidateOnly := 0
	seen := make(map[string]bool, len(pages))
	consider := func(page batchfactoryv11.ComicPage) {
		hasPrimary, hasAny := comicPageImageFlags(page)
		promptStatus, imageStatus, _, primaryFallback := deriveComicPageStatus(page, hasPrimary, hasAny)
		if promptStatus == comicPromptStatusReady {
			readiness.PromptReadyCount++
		}
		if imageStatus == comicImageStatusPrimary {
			readiness.ImageReadyCount++
			readiness.PrimaryReadyCount++
		}
		if primaryFallback {
			candidateOnly++
		}
		if page.ConfigStale {
			readiness.StaleCount++
		}
	}
	if len(videos) > 0 {
		for _, video := range videos {
			seen[video.ID] = true
			page, ok := pagesByVideo[video.ID]
			if !ok {
				missingPageRows++
				continue
			}
			consider(page)
		}
	} else {
		for _, page := range pages {
			if seen[page.VideoID] {
				continue
			}
			seen[page.VideoID] = true
			consider(page)
		}
	}

	if readiness.PagesTotal == 0 {
		readiness.Blockers = append(readiness.Blockers, "没有可用的漫画页")
	}
	if missingPageRows > 0 {
		readiness.Blockers = append(readiness.Blockers, fmt.Sprintf("缺少漫画页记录：%d 页", missingPageRows))
	}
	if readiness.PromptReadyCount < readiness.PagesTotal {
		readiness.Blockers = append(readiness.Blockers, fmt.Sprintf("提示词未就绪：%d/%d 页", readiness.PromptReadyCount, readiness.PagesTotal))
	}
	if readiness.StaleCount > 0 {
		readiness.Blockers = append(readiness.Blockers, fmt.Sprintf("提示词待重新分析：%d 页", readiness.StaleCount))
	}
	if readiness.ImageReadyCount < readiness.PagesTotal {
		readiness.Blockers = append(readiness.Blockers, fmt.Sprintf("尚未出主图：%d 页", readiness.PagesTotal-readiness.ImageReadyCount))
	}
	if candidateOnly > 0 {
		readiness.Blockers = append(readiness.Blockers, fmt.Sprintf("历史图片未设主图：%d 页", candidateOnly))
	}

	readiness.CanGenerateImages = readiness.PagesTotal > 0 && readiness.PromptReadyCount == readiness.PagesTotal
	readiness.CanBuildCollection = readiness.CanGenerateImages && readiness.ImageReadyCount == readiness.PagesTotal && readiness.StaleCount == 0
	readiness.Skippable = readiness.PagesTotal > 0 && readiness.ImageReadyCount == readiness.PagesTotal
	return readiness
}

// comicReadinessByBook 给批次详情补书级权威就绪度。只在 store 支持页面行时
// 计算；任何一本书读取失败只记录为空，不影响批次详情本身返回。
func comicReadinessByBook(ctx context.Context, store batchfactoryv11.Store, presetCatalog *prompt.Catalog, owner string, batch batchfactoryv11.Batch) map[string]comicReadiness {
	comicStore, ok := store.(batchfactoryv11.ComicPageStore)
	if !ok {
		return nil
	}
	result := make(map[string]comicReadiness, len(batch.Books))
	for _, book := range batch.Books {
		pages, err := comicStore.ListComicPages(ctx, owner, batch.ID, book.ID)
		if err != nil {
			continue
		}
		for index := range pages {
			pages[index].ConfigStale = comicPageConfigStale(ctx, store, presetCatalog, owner, batch, book.ID, pages[index])
		}
		result[book.ID] = buildComicReadiness(book.Videos, pages, comicPageGenerationLimit)
	}
	return result
}
