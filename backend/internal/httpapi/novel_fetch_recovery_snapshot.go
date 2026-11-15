package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// Persist execution choices separately from mutable account defaults. Never
// persist provider credentials here. Older books use defaults until snapshotted.
func saveNovelFetchJobSnapshot(ctx context.Context, store novelfetchworkshop.Store, job novelfetchworkshop.ProcessJob) error {
	for _, id := range job.BookIDs {
		doc, err := store.GetDocument(ctx, job.Owner, id)
		if err != nil {
			return err
		}
		if doc.Meta == nil {
			doc.Meta = map[string]any{}
		}
		if novelFetchStopped(firstText(doc.Meta["status"], doc.Meta["queueState"])) {
			continue
		}
		if current := firstText(doc.Meta["currentProcessJobId"], doc.Meta["current_process_job_id"]); current != "" && current != job.ID {
			switch firstText(doc.Meta["status"], doc.Meta["queueState"]) {
			case "queued", "running", "processing", "ai_processing":
				continue
			}
		}
		doc.Meta["processSnapshot"] = map[string]any{"batch_id": job.BatchID, "schedule_id": job.ScheduleID, "platform_id": job.PlatformID, "max_txt": job.MaxTxt, "selected_versions": append([]string{}, job.SelectedVersions...), "rewrite_methods": job.RewriteMethods, "text_model_id": job.TextModelID, "sensitive_ai_enabled": job.SensitiveAIEnabled, "auto_submit": job.AutoSubmit}
		doc.Meta["nextSelectedVersions"] = append([]string{}, job.SelectedVersions...)
		if err = store.PutDocument(ctx, job.Owner, doc); err != nil {
			return err
		}
	}
	return nil
}

func novelFetchRecoveryJob(doc novelfetchworkshop.Document, settings map[string]any, owner string, automatic bool) novelfetchworkshop.ProcessJob {
	snap := mapOrEmpty(doc.Meta["processSnapshot"])
	if len(snap) == 0 {
		// Recover choices for pre-snapshot scheduled books from their own schedule,
		// never from another batch or a different account.
		for _, raw := range arrayOrEmpty(settings["novel_fetch_schedules"]) {
			item := mapOrEmpty(raw)
			matches := false
			for _, id := range scheduleStringSlice(item["book_ids"]) {
				if id == doc.BookID {
					matches = true
					break
				}
			}
			if matches {
				snap = cloneAnyMap(item)
				snap["schedule_id"] = item["id"]
				break
			}
		}
	}
	job := novelfetchworkshop.ProcessJob{Owner: owner, BatchID: firstText(doc.Meta["currentBatchId"], doc.Meta["current_batch_id"]), ScheduleID: firstText(doc.Meta["schedule_id"], doc.Meta["scheduleId"]), PlatformID: novelFetchResolvedPlatformID(settings, doc.Meta), MaxTxt: doc.MaxTxt, TextModelID: firstText(settings["text_model_id"], settings["textModelId"], "未指定"), SensitiveAIEnabled: novelFetchSensitiveAIEnabled(settings), AutoSubmit: automatic, SelectedVersions: novelFetchAutomationVersions(settings)}
	job.RewriteMethods = scheduleStringMap(mapOrEmpty(mapOrEmpty(settings["appConfig"])["rewrite"])["ai_slot_methods"])
	if len(snap) > 0 {
		job.BatchID = firstText(snap["batch_id"], job.BatchID)
		job.ScheduleID = firstText(snap["schedule_id"], job.ScheduleID)
		job.PlatformID = firstText(snap["platform_id"], job.PlatformID)
		job.MaxTxt = intValue(snap["max_txt"])
		job.TextModelID = firstText(snap["text_model_id"], job.TextModelID)
		// Empty selections are meaningful: original-only processing stays original-only.
		job.SelectedVersions = scheduleStringSlice(snap["selected_versions"])
		job.RewriteMethods = scheduleStringMap(snap["rewrite_methods"])
		job.SensitiveAIEnabled = snap["sensitive_ai_enabled"] == true
		if explicit, ok := snap["auto_submit"].(bool); ok {
			job.AutoSubmit = explicit
		}
	} else if !automatic {
		job.SelectedVersions = []string{}
		for i := 1; i <= workshopAIRequestedCount(doc, settings, 0); i++ {
			job.SelectedVersions = append(job.SelectedVersions, fmt.Sprintf("ai%d", i))
		}
	}
	// A later user edit is the next-run target; the old snapshot remains a
	// record of the already queued job, not an override of that edit.
	if versions, exists := doc.Meta["nextSelectedVersions"]; exists {
		job.SelectedVersions = scheduleStringSlice(versions)
	}
	if job.MaxTxt <= 0 {
		job.MaxTxt = 4000
	}
	return job
}

func novelFetchRecoveryGroup(job novelfetchworkshop.ProcessJob) string {
	job.ID = ""
	job.BookIDs = nil
	job.DedupeKey = ""
	job.DedupeBookIDs = nil
	job.Attempts = 0
	job.LastError = ""
	raw, _ := json.Marshal(job)
	return string(raw)
}

// A cancellation remains authoritative until an explicit user retry clears it.
// Fail closed on queue errors; do not consume model calls with uncertain intent.
func reconcileNovelFetchCancellation(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, doc novelfetchworkshop.Document) (bool, error) {
	status := strings.ToLower(firstText(doc.Meta["status"], doc.Meta["queueState"]))
	if status == "stopped" || status == "cancelled" || status == "canceled" {
		return true, nil
	}
	cancellation, ok := queue.(novelfetchworkshop.ProcessCancellationQueue)
	if !ok {
		return false, nil
	}
	cancelled, err := cancellation.IsBookCancelled(ctx, owner, doc.BookID)
	if err != nil {
		return false, err
	}
	if !cancelled {
		return false, nil
	}
	// Completed evidence remains intact, but all other stale queue states become stopped.
	if status != "completed" && status != "done" && status != "success" {
		if err = updateWorkshopTaskRuntimeState(ctx, store, owner, doc.BookID, "stopped", "任务已取消；手动重试后才会继续"); err != nil {
			return true, err
		}
	}
	return true, nil
}

func novelFetchSnapshotBeforeQueue(ctx context.Context, store novelfetchworkshop.Store, job novelfetchworkshop.ProcessJob) error {
	if job.MaxTxt <= 0 {
		job.MaxTxt = 4000
	}
	return saveNovelFetchJobSnapshot(ctx, store, job)
}
