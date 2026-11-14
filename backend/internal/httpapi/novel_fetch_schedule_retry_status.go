package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

// A deferred fetch is retried by automation under a new process job. That job
// does not own the original schedule ID. Reconcile the original schedule from
// its own books instead of redispatching it or leaving its card waiting forever.
func reconcileNovelFetchScheduleRetry(ctx context.Context, store novelfetchworkshop.Store, owner string, item, settings map[string]any, now time.Time) (bool, error) {
	previous := firstText(item["status"])
	if previous != "waiting_retry" && item["retry_tracking"] != true {
		return false, nil
	}
	switch previous {
	case "completed", "failed", "stopped", "cancelled":
		return false, nil
	}
	ids := scheduleStringSlice(item["book_ids"])
	if len(ids) == 0 {
		return true, nil
	}
	states := map[string]int{}
	message := ""
	var nextRetry time.Time
	for _, bookID := range ids {
		document, err := store.GetDocument(ctx, owner, bookID)
		if err != nil {
			if !errors.Is(err, novelfetchworkshop.ErrNotFound) {
				return true, err
			}
			states["failed"]++
			message = "定时任务中的书籍记录不存在：" + bookID
			continue
		}
		meta := document.Meta
		status := strings.ToLower(firstText(meta["status"], meta["queueState"], meta["taskStatus"]))
		if status == "done" || status == "success" || status == "succeeded" {
			status = "completed"
		}
		submitted := strings.ToLower(firstText(meta["siteSubmitStatus"], meta["site_submit_status"]))
		if status == "completed" && item["auto_submit"] != false {
			switch submitted {
			case "completed", "done", "success", "succeeded", "submitted", "uploaded":
			case "waiting_login", "waiting_configuration":
				status = submitted
			default:
				status = "upload_pending"
			}
		}
		switch status {
		case "processing", "ai_processing":
			status = "running"
		case "done", "success", "succeeded":
			status = "completed"
		case "input_ready", "":
			status = "waiting_retry"
		}
		if status == "failed" && novelFetchAutomationEnabled(settings) && meta["manualOnly"] != true && meta["manual_only"] != true && meta["rewriteValidationHold"] != true {
			status = "waiting_retry"
		}
		states[status]++
		if status != "completed" && message == "" {
			message = firstText(meta["error"], meta["originalFetchLastError"], meta["automationSubmitError"])
		}
		if status == "waiting_retry" {
			raw := firstText(meta["originalFetchNextRetryAt"], meta["original_fetch_next_retry_at"], meta["automationNextRetryAt"], meta["automation_next_retry_at"])
			if due, err := time.Parse(time.RFC3339Nano, raw); err == nil && (nextRetry.IsZero() || due.Before(nextRetry)) {
				nextRetry = due
			}
		}
	}
	status := "waiting_retry"
	for _, candidate := range []string{"running", "queued", "waiting_retry", "waiting_login", "waiting_configuration", "waiting_classification", "waiting_sensitive_ai", "upload_pending", "failed", "stopped", "completed"} {
		if states[candidate] > 0 {
			status = candidate
			break
		}
	}
	if status == "completed" && states["completed"] != len(ids) {
		status = "waiting_retry"
	}
	if status == "completed" || status == "running" || status == "queued" {
		message = ""
	}
	next := ""
	if status == "waiting_retry" && !nextRetry.IsZero() {
		next = nextRetry.UTC().Format(time.RFC3339Nano)
	}
	tracking := status != "completed" && status != "failed" && status != "stopped"
	if status == previous && message == firstText(item["error"]) && next == firstText(item["next_retry_at"]) && item["retry_tracking"] == tracking {
		return true, nil
	}
	updated := cloneAnyMap(item)
	updated["_expected_status"] = previous
	updated["status"], updated["retry_tracking"] = status, tracking
	updated["next_retry_at"], updated["error"] = next, message
	updated["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
	return true, saveNovelFetchScheduleItem(ctx, store, owner, updated)
}
