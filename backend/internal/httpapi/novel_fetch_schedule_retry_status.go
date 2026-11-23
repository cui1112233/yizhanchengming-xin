package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

type novelFetchGenerationDecision struct {
	status          string
	skipReason      string
	dispatch        bool
	retryable       bool
	missingVersions []string
}

// Normalize each current signal before applying the schedule's hold and retry
// precedence. The general task status rule owns the persisted status aliases.
func novelFetchGenerationSignal(raw string) string {
	switch strings.ToLower(raw) {
	case "ai_processing":
		return novelfetchworkshop.TaskStatusProcessing
	case "pending_upload", "pending_verification", "pending_confirmation", novelfetchworkshop.TaskStatusPendingSubmit:
		return novelfetchworkshop.TaskStatusUploadPending
	}
	return novelfetchworkshop.DeriveTaskStatus(novelfetchworkshop.Document{Meta: map[string]any{"status": raw}})
}

func novelFetchScheduleGenerationDecision(ctx context.Context, store novelfetchworkshop.Store, owner string, doc novelfetchworkshop.Document, versions []string) novelFetchGenerationDecision {
	meta := doc.Meta
	queue := novelFetchGenerationSignal(firstText(meta["queueState"], meta["queue_state"]))
	status := novelFetchGenerationSignal(firstText(meta["status"], meta["taskStatus"], meta["task_status"]))
	if queue == novelfetchworkshop.TaskStatusStopped || status == novelfetchworkshop.TaskStatusStopped {
		return novelFetchGenerationDecision{status: "stopped"}
	}
	if strings.EqualFold(firstText(meta["originalStatus"], meta["original_status"]), "too_short") ||
		strings.EqualFold(firstText(meta["aiStatus"], meta["ai_status"]), "skipped_short_original") {
		return novelFetchGenerationDecision{status: "completed", skipReason: novelFetchShortOriginalSkipReason}
	}
	if meta["manualOnly"] == true || meta["manual_only"] == true || meta["manualOriginalFetchRequired"] == true ||
		meta["rewriteValidationHold"] == true || novelFetchRetryScheduleHasManualHold(meta) ||
		novelFetchGenerationSignal(firstText(meta["aiStatus"], meta["ai_status"])) == novelfetchworkshop.TaskStatusFailed {
		return novelFetchGenerationDecision{status: "failed"}
	}
	if status == novelfetchworkshop.TaskStatusInputReady {
		// Older documents can have just a queue marker or individual stage fields.
		if queue != novelfetchworkshop.TaskStatusInputReady {
			status = queue
		} else {
			status = novelfetchworkshop.DeriveTaskStatus(doc)
		}
	}
	if status == novelfetchworkshop.TaskStatusFailed || queue == novelfetchworkshop.TaskStatusFailed {
		return novelFetchGenerationDecision{status: "failed", retryable: true}
	}
	for _, state := range []string{queue, status} {
		switch state {
		case novelfetchworkshop.TaskStatusProcessing:
			return novelFetchGenerationDecision{status: "running"}
		case novelfetchworkshop.TaskStatusQueued, novelfetchworkshop.TaskStatusWaitingRetry,
			novelfetchworkshop.TaskStatusWaitingClassification, novelfetchworkshop.TaskStatusWaitingSensitiveAI,
			novelfetchworkshop.TaskStatusWaitingConfig, novelfetchworkshop.TaskStatusWaitingLogin:
			return novelFetchGenerationDecision{status: state}
		}
	}
	switch status {
	case novelfetchworkshop.TaskStatusCompleted, novelfetchworkshop.TaskStatusPendingSubmit, novelfetchworkshop.TaskStatusUploadPending:
		var missing []string
		for _, version := range novelFetchRequiredVersions(versions) {
			if !novelFetchVersionsGenerated(ctx, store, owner, doc, []string{version}) {
				missing = append(missing, version)
			}
		}
		if len(missing) > 0 {
			return novelFetchGenerationDecision{status: "waiting_retry", dispatch: true, missingVersions: missing}
		}
		return novelFetchGenerationDecision{status: "completed"}
	}
	return novelFetchGenerationDecision{status: status, dispatch: true}
}

// A deferred fetch is retried by automation under a new process job. That job
// does not own the original schedule ID. Reconcile the original schedule from
// its own books instead of redispatching it or leaving its card waiting forever.
func reconcileNovelFetchScheduleRetry(ctx context.Context, store novelfetchworkshop.Store, owner string, item, settings map[string]any, now time.Time, readback ...novelFetchScheduleReadback) (bool, error) {
	previous := firstText(item["status"])
	awaitingVerification := previous == "pending_verification" || previous == "pending_confirmation"
	if previous != "waiting_retry" && previous != "upload_pending" && !awaitingVerification && item["retry_tracking"] != true {
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
	versions := scheduleStringSlice(item["selected_versions"])
	pendingBookIDs := scheduleStringSlice(item["pending_book_ids"])
	pendingChanged := false
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
		if novelFetchStopped(status) {
			status = "stopped"
		}
		if status != "stopped" && item["auto_submit"] != false {
			pending := novelFetchScheduleReadbackVersions(document, versions)
			if len(pending) > 0 {
				status = "pending_verification"
				if len(readback) > 0 && readback[0] != nil && novelFetchScheduleReadbackDue(document, now) {
					readback[0](ctx, owner, bookID, pending)
					document, err = store.GetDocument(ctx, owner, bookID)
					if err != nil {
						return true, err
					}
					meta = document.Meta
					if novelFetchStopped(firstText(meta["status"], meta["queueState"])) {
						status = "stopped"
					} else if len(novelFetchScheduleReadbackVersions(document, versions)) == 0 {
						status = "upload_pending"
					}
				}
			}
			done := novelFetchVersionsSubmitted(document, versions)
			if !done && status == "completed" {
				switch submitted {
				case "waiting_login", "waiting_configuration":
					status = submitted
				default:
					status = "upload_pending"
				}
			} else if done && status == "upload_pending" {
				status = "completed"
			}
			if status == "upload_pending" && (len(pending) > 0 || awaitingVerification) && !seenString(pendingBookIDs, bookID) {
				pendingBookIDs = append(pendingBookIDs, bookID)
				pendingChanged = true
			}
		}
		retryableFailure := false
		if item["auto_submit"] == false {
			decision := novelFetchScheduleGenerationDecision(ctx, store, owner, document, versions)
			status, retryableFailure = decision.status, decision.retryable
			if decision.dispatch {
				status = "waiting_retry"
				if len(decision.missingVersions) > 0 && !seenString(pendingBookIDs, bookID) {
					pendingBookIDs = append(pendingBookIDs, bookID)
					pendingChanged = true
				}
			}
		} else if status == "failed" {
			retryableFailure = meta["manualOnly"] != true && meta["manual_only"] != true && meta["rewriteValidationHold"] != true && !novelFetchRetryScheduleHasManualHold(meta)
		}
		switch status {
		case "processing", "ai_processing":
			status = "running"
		case "done", "success", "succeeded":
			status = "completed"
		case "input_ready", "":
			status = "waiting_retry"
		}
		if retryableFailure && novelFetchAutomationEnabled(settings) {
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
	for _, candidate := range []string{"running", "queued", "waiting_retry", "waiting_login", "waiting_configuration", "waiting_classification", "waiting_sensitive_ai", "pending_verification", "pending_confirmation", "upload_pending", "failed", "stopped", "completed"} {
		if states[candidate] > 0 {
			status = candidate
			break
		}
	}
	if item["auto_submit"] == false && status == "stopped" && states["completed"] > 0 && states["completed"]+states["stopped"] == len(ids) {
		status = "completed"
	} else if status == "completed" && states["completed"] != len(ids) {
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
	if status == "completed" && len(pendingBookIDs) > 0 {
		pendingBookIDs = nil
		pendingChanged = true
	}
	if !pendingChanged && status == previous && message == firstText(item["error"]) && next == firstText(item["next_retry_at"]) && item["retry_tracking"] == tracking {
		return true, nil
	}
	updated := cloneAnyMap(item)
	updated["_expected_status"] = previous
	updated["status"], updated["retry_tracking"] = status, tracking
	updated["next_retry_at"], updated["error"] = next, message
	updated["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
	if pendingChanged {
		updated["pending_book_ids"] = pendingBookIDs
		if item["auto_submit"] == false && len(pendingBookIDs) > 0 {
			updated["partial_dispatch"] = true
		}
	}
	return true, saveNovelFetchScheduleItem(ctx, store, owner, updated)
}
