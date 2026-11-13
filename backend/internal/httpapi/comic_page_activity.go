package httpapi

import (
	"qiantie/backend/internal/batchfactoryv11"
	"time"
)

type comicPageActivity struct{ Generating, Failed, Interrupted int }

func comicPageActivityCounts(pages []batchfactoryv11.ComicPage, now time.Time) comicPageActivity {
	var result comicPageActivity
	for _, page := range pages {
		if comicPageGenerationTaskIsStale(page, now) {
			result.Interrupted++
			continue
		}
		primary, anyImage := comicPageImageFlags(page)
		_, _, state, _ := deriveComicPageStatus(page, primary, anyImage)
		switch state {
		case comicGenerationStateGenerating:
			result.Generating++
		case comicGenerationStateFailed:
			result.Failed++
		}
	}
	return result
}
