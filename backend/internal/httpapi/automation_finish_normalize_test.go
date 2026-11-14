package httpapi

import (
	"testing"
	"time"
)

func TestNormalizeQueuedBooksAtFinishMarksOnlyQueuedBooks(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	books := map[string]any{
		"book-a": map[string]any{"status": "queued"},
		"book-b": map[string]any{"status": "scheduled"},
		"book-c": map[string]any{"status": "completed"},
		"book-d": map[string]any{"status": "running"},
	}

	if got := normalizeQueuedBooksAtFinish(books, now); got != 1 {
		t.Fatalf("normalized=%d, want 1", got)
	}

	queued := books["book-a"].(map[string]any)
	if queued["status"] != "failed" {
		t.Fatalf("queued book status=%v, want failed", queued["status"])
	}
	if queued["message"] == "" || queued["error"] == "" {
		t.Fatalf("queued book needs operator-facing message, got message=%v error=%v", queued["message"], queued["error"])
	}
	if queued["updatedAt"] != now {
		t.Fatalf("queued book updatedAt=%v, want %v", queued["updatedAt"], now)
	}

	for _, id := range []string{"book-b", "book-c", "book-d"} {
		entry := books[id].(map[string]any)
		if entry["status"] != map[string]string{
			"book-b": "scheduled", "book-c": "completed", "book-d": "running",
		}[id] {
			t.Fatalf("%s status=%v, want untouched", id, entry["status"])
		}
	}
}

func TestNormalizeQueuedBooksAtFinishFeedsFailureCounts(t *testing.T) {
	books := map[string]any{
		"book-a": map[string]any{"status": "queued"},
		"book-b": map[string]any{"status": "completed"},
	}
	normalizeQueuedBooksAtFinish(books, time.Now())
	counts := automationCountsFromBooks(books)
	if automationNumberValue(counts["failed"], 0) != 1 {
		t.Fatalf("failed count=%v, want 1", counts["failed"])
	}
	if automationNumberValue(counts["queued"], 0) != 0 {
		t.Fatalf("queued count=%v, want 0", counts["queued"])
	}
}
