package httpapi

import (
	"context"
	"errors"
	"qiantie/backend/internal/novelfetchworkshop"
)

// Scheduling may create documents and update the current batch in between
// writes. Persist only runtime fields on its own schedule, never the old full
// settings snapshot. A cancelled/deleted schedule must not be resurrected.
func saveNovelFetchScheduleItem(ctx context.Context, store novelfetchworkshop.Store, owner string, item map[string]any) error {
	for attempt := 0; attempt < 3; attempt++ {
		current, err := store.GetConfig(ctx, owner)
		if err != nil {
			return err
		}
		found := false
		for _, raw := range arrayOrEmpty(current["novel_fetch_schedules"]) {
			row := mapOrEmpty(raw)
			if firstText(row["id"]) != firstText(item["id"]) {
				continue
			}
			if firstText(row["status"]) == "cancelled" {
				return nil
			}
			if expected, ok := item["_expected_status"].(string); ok && firstText(row["status"]) != expected {
				return nil
			}
			if firstText(item["status"]) == "queued" {
				switch firstText(row["status"]) {
				case "running", "completed", "failed", "stopped":
					return nil
				}
			}
			for _, key := range []string{"status", "error", "updatedAt", "dispatch_id", "job_id", "batch_id", "book_ids", "pending_book_ids", "partial_dispatch", "dispatch_started_at", "skipped_existing_book_ids", "existing_job_id", "retry_tracking", "next_retry_at", "tasks_prepared", "created_ids", "requested_book_ids", "skipped_books", "parsed_count", "empty_id_count", "duplicate_count", "selected_versions"} {
				if value, ok := item[key]; ok {
					row[key] = value
				} else if key == "error" {
					delete(row, key)
				}
			}
			found = true
			break
		}
		if !found {
			return nil
		}
		if err = store.PutConfig(ctx, owner, current); !errors.Is(err, novelfetchworkshop.ErrConfigConflict) {
			return err
		}
	}
	return novelfetchworkshop.ErrConfigConflict
}
