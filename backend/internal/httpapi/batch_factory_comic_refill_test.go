package httpapi

import (
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestMissingComicPageWorkItems(t *testing.T) {
	videos := []batchfactoryv11.Video{{ID: "v1", Label: "第1页"}, {ID: "v2", Label: "第2页"}, {ID: "v3", Label: "第3页"}}
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "v1", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}},
		{VideoID: "v2", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: false}}},
	}

	work := missingComicPageWorkItems(videos, pages)
	if len(work) != 2 || work[0].page.ID != "v2" || work[1].page.ID != "v3" {
		t.Fatalf("unexpected missing work: %+v", work)
	}
	if work[0].index != 1 || work[1].index != 2 {
		t.Fatalf("page indexes not preserved: %+v", work)
	}
}
