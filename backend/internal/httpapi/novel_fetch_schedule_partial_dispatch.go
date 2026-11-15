package httpapi

import (
	"context"
	"fmt"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

// A per-book lock must not abandon unrelated books. Pending IDs remain on the
// parent schedule until they are either independently queued or proven done.
// Child jobs do not own the parent's status: reconciliation checks every book.
func dispatchNovelFetchSchedulePending(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, item map[string]any, now time.Time) error {
	unique, ok := queue.(novelfetchworkshop.ProcessBookDedupQueue)
	if !ok {
		return fmt.Errorf("任务队列不支持防重复派发")
	}
	pending := []string{}
	runs, err := store.ListRuns(ctx, owner, "")
	if err != nil {
		return err
	}
	activeRuns := map[string]bool{}
	for _, run := range runs {
		if run.Status == "queued" || run.Status == "running" {
			activeRuns[run.RunID] = true
		}
	}
	for _, id := range scheduleStringSlice(item["pending_book_ids"]) {
		doc, err := store.GetDocument(ctx, owner, id)
		if err != nil {
			return err
		}
		if novelFetchStopped(firstText(doc.Meta["status"], doc.Meta["queueState"])) || doc.Meta["manualOnly"] == true || doc.Meta["manual_only"] == true || doc.Meta["rewriteValidationHold"] == true || novelFetchRetryScheduleHasManualHold(doc.Meta) {
			continue
		}
		if cancellation, ok := queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
			stopped, err := cancellation.IsBookCancelled(ctx, owner, id)
			if err != nil {
				return err
			}
			if stopped {
				continue
			}
		}
		versions := scheduleStringSlice(item["selected_versions"])
		ready := novelFetchVersionsGenerated(ctx, store, owner, doc, versions)
		if item["auto_submit"] != false {
			ready = novelFetchVersionsSubmitted(doc, versions)
		}
		if ready {
			continue
		}
		// An unconfirmed submission must be read back, never uploaded twice.
		if state := firstText(doc.Meta["siteSubmitStatus"], doc.Meta["site_submit_status"]); state == "pending_verification" || state == "pending_confirmation" {
			continue
		}
		job := novelfetchworkshop.ProcessJob{
			ID:         firstText(item["job_id"]) + "-book-" + id,
			ScheduleID: firstText(item["id"]),
			Owner:      owner, BatchID: firstText(item["batch_id"]), BookIDs: []string{id},
			PlatformID: firstText(item["platform_id"], "2"), MaxTxt: intValue(item["max_txt"]),
			SelectedVersions: versions, AutoSubmit: item["auto_submit"] != false,
			SensitiveAIEnabled: item["sensitive_ai_enabled"] == true,
			RewriteMethods:     scheduleStringMap(item["rewrite_methods"]), TextModelID: firstText(item["text_model_id"]),
		}
		if job.MaxTxt <= 0 {
			job.MaxTxt = 4000
		}
		started := now.UTC().Format(time.RFC3339Nano)
		if !activeRuns[job.ID] {
			if err := store.PutRun(ctx, owner, novelfetchworkshop.RunRecord{RunID: job.ID, BookID: job.BatchID, Stage: "novel_fetch_process", Status: "queued", StartedAt: started, TextModelID: job.TextModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "定时任务逐书派发"}); err != nil {
				return err
			}
		}
		queued, existing, err := unique.EnqueueUniqueBooks(ctx, job, 7*24*time.Hour)
		if err != nil {
			_ = updateWorkshopRun(store, job, "failed", "定时任务入队失败", started, true)
			return err
		}
		if !queued && existing != job.ID {
			if existing == "" {
				return fmt.Errorf("定时任务入队未确认，保留等待记录")
			}
			pending = append(pending, id)
			_ = updateWorkshopRun(store, job, "duplicate_skipped", "该书正由其他任务处理，定时任务继续等待", started, true)
			continue
		}
		if err := novelFetchSnapshotBeforeQueue(ctx, store, job); err != nil {
			return err
		}
		if err := markNovelFetchTaskQueued(ctx, store, owner, id, job.ID, job.BatchID, now); err != nil {
			return err
		}
	}
	item["pending_book_ids"], item["retry_tracking"] = pending, true
	item["status"], item["updatedAt"] = "waiting_retry", now.UTC().Format(time.RFC3339Nano)
	item["error"] = ""
	if len(pending) > 0 {
		item["error"] = fmt.Sprintf("%d 本书已有任务占用，将自动继续检查；其他书已独立派发", len(pending))
	}
	return saveNovelFetchScheduleItem(ctx, store, owner, item)
}
