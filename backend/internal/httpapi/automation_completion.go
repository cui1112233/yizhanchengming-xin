package httpapi

import (
	"sort"

	"qiantie/backend/internal/batchfactoryv11"
)

// comicPageContinuation 是「重试 = 幂等续跑」的完成清单视图。
// 进入任何页级节点前先用它算出工作集，已完成的目标绝不再次调用上游。
type comicPageContinuation struct {
	Pending []batchfactoryv11.Video
	Done    []batchfactoryv11.Video
}

// buildComicPageContinuation 按主图是否存在拆分工作集。判定复用生产代码
// 已有的 comicPageHasPrimary，避免出现第二套完成语义。
func buildComicPageContinuation(videos []batchfactoryv11.Video, pages []batchfactoryv11.ComicPage) comicPageContinuation {
	result := comicPageContinuation{
		Pending: make([]batchfactoryv11.Video, 0, len(videos)),
		Done:    make([]batchfactoryv11.Video, 0, len(videos)),
	}
	for _, video := range videos {
		if comicPageHasPrimary(pages, video.ID) {
			result.Done = append(result.Done, video)
			continue
		}
		result.Pending = append(result.Pending, video)
	}
	return result
}

// filterFailedComicPageWork 从续跑工作集中剔除本轮已经真实失败的目标。
// 同一轮内绝不对同一个失败页原地重试，失败只记账、等下一轮或升级重试。
func filterFailedComicPageWork(work []comicPageWorkItem, failed map[string]struct{}) []comicPageWorkItem {
	if len(failed) == 0 {
		return work
	}
	filtered := make([]comicPageWorkItem, 0, len(work))
	for _, item := range work {
		if _, exists := failed[item.page.ID]; exists {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

// sortedComicPageFailureIDs 让账本与日志顺序稳定，便于断言与排查。
func sortedComicPageFailureIDs(failures map[string]string) []string {
	ids := make([]string, 0, len(failures))
	for id := range failures {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
