package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestFailedComicPagesIncludesOnlyExpiredGenerationLeases(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	pages := []batchfactoryv11.ComicPage{{VideoID: "old", Status: "generating", UpdatedAt: now.Add(-comicPageGenerationStaleAfter - time.Minute)}, {VideoID: "active", Status: "generating", UpdatedAt: now.Add(-time.Minute)}, {VideoID: "done", Status: "succeeded"}, {VideoID: "failed", Status: "generation_failed"}}
	got := failedComicPages(pages, now)
	if len(got) != 2 || got[0].videoID != "old" || got[1].videoID != "failed" {
		t.Fatalf("retry must protect active/succeeded pages: %+v", got)
	}
}

func TestFailedComicPagesKeepsOnlyGenerationFailed(t *testing.T) {
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "v-ready", Status: "image_ready", ModelID: "custom:ch_1"},
		{VideoID: "v-failed", Status: "generation_failed", ComicPrompt: "分格", ImagePrompt: "画面", ModelID: "custom:ch_1", AspectRatio: "9:16", ReferenceImageURLs: []string{"https://example.com/a.png"}},
		{VideoID: "v-generating", Status: "generating", ModelID: "custom:ch_1"},
		{VideoID: "v-failed-2", Status: "generation_failed"},
	}
	failed := failedComicPages(pages)
	if len(failed) != 2 {
		t.Fatalf("want 2 failed pages, got %d", len(failed))
	}
	if failed[0].videoID != "v-failed" || failed[1].videoID != "v-failed-2" {
		t.Fatalf("unexpected video order: %+v", failed)
	}
	first := failed[0].input
	if first.ComicPrompt != "分格" || first.ImagePrompt != "画面" || first.ModelID != "custom:ch_1" || first.AspectRatio != "9:16" {
		t.Fatalf("prompt/model fields not carried over: %+v", first)
	}
	if len(first.ReferenceImageURLs) != 1 || first.ReferenceImageURLs[0] != "https://example.com/a.png" {
		t.Fatalf("reference images not carried over: %+v", first.ReferenceImageURLs)
	}
}

func TestFailedComicPagesReturnsEmptyWhenAllSucceeded(t *testing.T) {
	failed := failedComicPages([]batchfactoryv11.ComicPage{
		{VideoID: "v1", Status: "image_ready"},
		{VideoID: "v2", Status: "generating"},
	})
	if len(failed) != 0 {
		t.Fatalf("want no failed pages, got %d", len(failed))
	}
}

func TestRetryFailedComicPagesRequiresGenerationServices(t *testing.T) {
	generator := comicPageGenerator{}
	if _, err := generator.retryFailedComicPages(context.Background(), "owner", "batch", "book"); !errors.Is(err, batchfactoryv11.ErrUnavailable) {
		t.Fatalf("want ErrUnavailable without comic services, got %v", err)
	}
}
