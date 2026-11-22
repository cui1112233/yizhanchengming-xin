package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

type novelFetchScheduleReadback func(context.Context, string, string, []string)

const novelFetchScheduleReadbackOperationPrefix = "schedule-readback-"

func newNovelFetchScheduleReadback(store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, usersDir string, target videomanagementconnector.Service) novelFetchScheduleReadback {
	if target == nil {
		return nil
	}
	return func(ctx context.Context, owner, bookID string, versions []string) {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		operationID := fmt.Sprintf("%s%d", novelFetchScheduleReadbackOperationPrefix, time.Now().UnixNano())
		_, _ = executeWebSubmit(ctx, owner, usersDir, store, webSubmitRequest{
			Queue: queue, TaskIDs: []string{bookID}, Versions: versions, OperationID: operationID, ReadbackOnly: true,
		}, nil, target)
	}
}

func novelFetchScheduleNeedsVersionDispatch(ctx context.Context, store novelfetchworkshop.Store, job novelfetchworkshop.ProcessJob) (bool, error) {
	for _, id := range job.BookIDs {
		doc, err := store.GetDocument(ctx, job.Owner, id)
		if err != nil {
			return false, err
		}
		if len(novelFetchScheduleReadbackVersions(doc, job.SelectedVersions)) > 0 {
			return true, nil
		}
		done := scheduleStringSlice(firstConfiguredValue(doc.Meta["siteSubmitDoneVersions"], doc.Meta["site_submit_done_versions"]))
		for _, version := range novelFetchRequiredVersions(job.SelectedVersions) {
			if seenString(done, version) {
				return true, nil
			}
		}
	}
	return false, nil
}

func novelFetchScheduleReadbackDue(doc novelfetchworkshop.Document, now time.Time) bool {
	meta := doc.Meta
	if meta["manualOnly"] == true || meta["manual_only"] == true || novelFetchRetryScheduleHasManualHold(meta) {
		return false
	}
	if due, err := time.Parse(time.RFC3339Nano, firstText(meta["automationNextRetryAt"], meta["automation_next_retry_at"])); err == nil && due.After(now) {
		return false
	}
	for i := len(doc.Logs) - 1; i >= 0; i-- {
		entry := mapOrEmpty(doc.Logs[i])
		if strings.HasPrefix(firstText(entry["operation_id"]), novelFetchScheduleReadbackOperationPrefix) {
			last, err := time.Parse(time.RFC3339Nano, firstText(entry["time"]))
			return err != nil || !last.Add(30*time.Second).After(now)
		}
	}
	return true
}

func novelFetchScheduleReadbackVersions(doc novelfetchworkshop.Document, versions []string) []string {
	var pending []string
	for _, version := range novelFetchRequiredVersions(versions) {
		if novelFetchSubmissionNeedsReadback(doc.Meta, version) {
			pending = append(pending, version)
		}
	}
	return pending
}

// A per-book lock must not abandon unrelated books. Pending IDs remain on the
// parent schedule until they are either independently queued or proven done.
// Child jobs do not own the parent's status: reconciliation checks every book.
func dispatchNovelFetchSchedulePending(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, item, settings map[string]any, now time.Time, readback novelFetchScheduleReadback) error {
	unique, ok := queue.(novelfetchworkshop.ProcessBookDedupQueue)
	if !ok {
		return fmt.Errorf("任务队列不支持防重复派发")
	}
	pending := []string{}
	busy := 0
	selected := scheduleStringSlice(item["selected_versions"])
	autoSubmit := item["auto_submit"] != false
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
		versions := selected
		if autoSubmit {
			if len(novelFetchScheduleReadbackVersions(doc, selected)) > 0 {
				pending = append(pending, id)
				continue // After readback, reassess the remaining versions.
			}
			done := scheduleStringSlice(firstConfiguredValue(doc.Meta["siteSubmitDoneVersions"], doc.Meta["site_submit_done_versions"]))
			versions = nil
			for _, version := range novelFetchRequiredVersions(selected) {
				if !seenString(done, version) {
					versions = append(versions, version)
				}
			}
			if len(versions) == 0 {
				continue
			}
		} else if novelFetchVersionsGenerated(ctx, store, owner, doc, selected) {
			continue
		}
		job := novelfetchworkshop.ProcessJob{
			ID:         firstText(item["job_id"]) + "-book-" + id,
			ScheduleID: firstText(item["id"]),
			Owner:      owner, BatchID: firstText(item["batch_id"]), BookIDs: []string{id},
			PlatformID: firstText(item["platform_id"], "2"), MaxTxt: intValue(item["max_txt"]),
			SelectedVersions: versions, AutoSubmit: autoSubmit,
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
			busy++
			_ = updateWorkshopRun(store, job, "duplicate_skipped", "该书正由其他任务处理，定时任务继续等待", started, true)
			continue
		}
		snapshot := job
		snapshot.SelectedVersions = selected
		if err := novelFetchSnapshotBeforeQueue(ctx, store, snapshot); err != nil {
			return err
		}
		if err := markNovelFetchTaskQueued(ctx, store, owner, id, job.ID, job.BatchID, now); err != nil {
			return err
		}
	}
	item["pending_book_ids"], item["retry_tracking"] = pending, true
	item["status"], item["updatedAt"] = "waiting_retry", now.UTC().Format(time.RFC3339Nano)
	item["error"] = ""
	if busy > 0 {
		item["error"] = fmt.Sprintf("%d 本书已有任务占用，将自动继续检查；其他书已独立派发", busy)
	}
	if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
		return err
	}
	_, err = reconcileNovelFetchScheduleRetry(ctx, store, owner, item, settings, now, readback)
	return err
}
