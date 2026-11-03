package httpapi

import (
	"testing"
	"time"
)

func TestScheduleAutomationBatchEndRetrySkipsFirstThenExhausts(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	entry := map[string]any{}
	for attempt := 1; attempt <= automationBatchRetryRounds; attempt++ {
		if ok := scheduleAutomationBatchEndRetry(entry, "comic_images", errTestRefusedImage{}, "整本漫画出图失败", attempt-1, now); !ok {
			t.Fatalf("round %d should schedule batch-end retry", attempt)
		}
		if got := automationNumberValue(entry["batchRetryCount"], 0); got != attempt {
			t.Fatalf("round %d batchRetryCount = %d, want %d", attempt, got, attempt)
		}
		if !automationBoolValue(entry["batchRetryPending"]) {
			t.Fatalf("round %d lost batchRetryPending", attempt)
		}
		if automationStringValue(entry["stage"], "") != "comic_images" {
			t.Fatalf("round %d stage = %q, want comic_images", attempt, entry["stage"])
		}
	}
	if ok := scheduleAutomationBatchEndRetry(entry, "comic_images", errTestRefusedImage{}, "整本漫画出图失败", automationBatchRetryRounds, now); ok {
		t.Fatalf("exhausted retry budget must fall through to terminal failure")
	}
	if ok := scheduleAutomationBatchEndRetry(entry, "comic_images", errTestRefusedImage{}, "整本漫画出图失败", 0, now); ok {
		t.Fatalf("round beyond the batch limit must not schedule another retry")
	}
}

type errTestRefusedImage struct{}

func (errTestRefusedImage) Error() string { return "image provider returned HTTP 400: 无法生成图片" }

func TestPrepareAutomationBatchRetryRoundResetsOnlyPendingBooks(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	books := map[string]any{
		"book_a": map[string]any{"status": "failed", "stage": "comic_images", "batchRetryPending": true, "batchRetryCount": 1, "error": "整本漫画出图失败"},
		"book_b": map[string]any{"status": "failed", "stage": "comic_publish", "comicCollectionId": "collection-1", "batchRetryPending": true, "batchRetryCount": 1},
		"book_c": map[string]any{"status": "completed", "stage": "published"},
		"book_d": map[string]any{"status": "failed", "stage": "assets"},
	}
	retry := prepareAutomationBatchRetryRound(books, 2, now)
	if len(retry) != 2 || !retry["book_a"] || !retry["book_b"] {
		t.Fatalf("retry set = %v, want book_a and book_b", retry)
	}
	first := books["book_a"].(map[string]any)
	if automationStringValue(first["status"], "") != "running" || automationStringValue(first["error"], "") != "" {
		t.Fatalf("book_a not reset: %v", first)
	}
	if automationBoolValue(first["batchRetryPending"]) {
		t.Fatalf("book_a kept batchRetryPending after being scheduled")
	}
	second := books["book_b"].(map[string]any)
	if automationStringValue(second["stage"], "") != "comic_publish" {
		t.Fatalf("book_b stage = %v, want comic_publish", second["stage"])
	}
	if automationStringValue(books["book_c"].(map[string]any)["status"], "") != "completed" {
		t.Fatalf("completed book must not be reset")
	}
	if automationStringValue(books["book_d"].(map[string]any)["status"], "") != "failed" {
		t.Fatalf("non-pending failed book must stay untouched")
	}
	if next := prepareAutomationBatchRetryRound(books, 3, now); len(next) != 0 {
		t.Fatalf("books already handed to a retry round must not be queued twice, got %v", next)
	}
}
