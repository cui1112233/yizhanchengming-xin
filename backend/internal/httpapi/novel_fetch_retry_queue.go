package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

type workshopRetryQueueResult struct {
	Requested      int               `json:"requested"`
	Queued         int               `json:"queued"`
	Updated        int               `json:"updated"`
	Retried        int               `json:"retried"`
	Failed         int               `json:"failed"`
	Skipped        int               `json:"skipped"`
	AlreadyRunning int               `json:"already_running"`
	TaskIDs        []string          `json:"task_ids"`
	Jobs           []map[string]any  `json:"jobs"`
	Errors         map[string]string `json:"errors"`
	OK             bool              `json:"ok"`
}

// resetNovelFetchAutomationRetryState 在用户主动重试时清掉自动重试留下的
// 人工标记和退避字段。用户点击重试代表重新开始一轮自动处理；如果保留
// manualOnly，这一轮结束失败后扫描器仍然不会接管，任务会一直停住。
func resetNovelFetchAutomationRetryState(document *novelfetchworkshop.Document) bool {
	if document == nil || document.Meta == nil {
		return false
	}
	changed := false
	for _, key := range []string{"manualOnly", "manual_only", "automationEnqueueAttempts", "automationNextRetryAt", "manualOriginalFetchRequired", "originalFetchRetryAttempts", "originalFetchNextRetryAt", "originalFetchLastError", "manualAiRetryRequired", "manual_ai_retry_required", "manual_ai_retry_stage", "manual_ai_retry_error"} {
		if _, ok := document.Meta[key]; ok {
			delete(document.Meta, key)
			changed = true
		}
	}
	// 用户主动重试代表重新开始一轮处理，阶段重试表和退避时间必须一起清掉；
	// 否则扫描器仍会被旧的 next_retry_at 或 manual_only 挡住。
	if len(novelFetchRetrySchedule(document.Meta)) > 0 {
		clearNovelFetchTaskRetry(document.Meta, "")
		changed = true
	}
	return changed
}

// This function only persists work in the existing queue. Provider calls and
// body processing belong to the worker, whose context outlives the HTTP call.
func enqueueWorkshopRetries(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, ids []string, sensitiveAI bool) (workshopRetryQueueResult, error) {
	result := workshopRetryQueueResult{Requested: len(ids), TaskIDs: []string{}, Jobs: []map[string]any{}, Errors: map[string]string{}, OK: true}
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return result, err
	}
	uniqueQueue, queueReady := queue.(novelfetchworkshop.ProcessBookDedupQueue)
	autoSubmit := novelFetchAutomationEnabled(settings)
	for _, bookID := range ids {
		document, err := store.GetDocument(ctx, owner, bookID)
		if err != nil {
			result.Failed++
			result.Errors[bookID] = "读取任务失败"
			continue
		}
		if active := activeNovelFetchJobForBooks(ctx, store, owner, []string{bookID}); active != nil {
			result.Skipped++
			result.AlreadyRunning++
			result.Jobs = append(result.Jobs, active)
			continue
		}
		originalReady := strings.EqualFold(firstText(document.Meta["originalStatus"], document.Meta["original_status"]), "done")
		aiStatus := strings.ToLower(firstText(document.Meta["aiStatus"], document.Meta["ai_status"]))
		status := strings.ToLower(firstText(document.Meta["status"], document.Meta["queueState"]))
		aiDone := aiStatus == "done" || aiStatus == "ai_done" || aiStatus == "completed"
		uploadNeeded := autoSubmit && firstText(document.Meta["siteSubmitStatus"], document.Meta["site_submit_status"]) != "submitted"
		if originalReady && aiDone && status != "failed" && !uploadNeeded {
			result.Skipped++
			continue
		}
		if !queueReady {
			return result, fmt.Errorf("重试队列未配置或不支持防重复入队，未执行生成，请检查 Redis 队列")
		}
		if resetNovelFetchAutomationRetryState(&document) {
			if err := store.PutDocument(ctx, owner, document); err != nil {
				result.Failed++
				result.Errors[bookID] = "重置自动重试状态失败"
				continue
			}
		}
		if cancelled, ok := queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
			if err := cancelled.ClearCancelledBooks(ctx, owner, []string{bookID}); err != nil {
				result.Failed++
				result.Errors[bookID] = "清理取消状态失败"
				continue
			}
		}
		versions := []string{}
		for index := 1; index <= workshopAIRequestedCount(document, settings, 0); index++ {
			versions = append(versions, fmt.Sprintf("ai%d", index))
		}
		if autoSubmit {
			versions = cleanStrings(append(novelFetchAutomationVersions(settings), versions...))
		}
		batchID := firstText(document.Meta["currentBatchId"], document.Meta["current_batch_id"], "retry-batch")
		job := novelfetchworkshop.ProcessJob{ID: fmt.Sprintf("retry_%d", time.Now().UnixNano()), Owner: owner, BatchID: batchID, BookIDs: []string{bookID}, PlatformID: novelFetchResolvedPlatformID(settings, document.Meta), MaxTxt: document.MaxTxt, SensitiveAIEnabled: sensitiveAI || novelFetchSensitiveAIEnabled(settings), AutoSubmit: autoSubmit, SelectedVersions: versions, TextModelID: firstText(settings["text_model_id"], settings["textModelId"], "未指定")}
		if job.MaxTxt <= 0 {
			job.MaxTxt = 4000
		}
		job.RewriteMethods = map[string]string{}
		for version, method := range mapOrEmpty(mapOrEmpty(mapOrEmpty(settings["appConfig"])["rewrite"])["ai_slot_methods"]) {
			job.RewriteMethods[version] = firstText(method)
		}
		started := time.Now().UTC().Format(time.RFC3339Nano)
		if err := store.PutRun(ctx, owner, novelfetchworkshop.RunRecord{RunID: job.ID, BookID: batchID, Stage: "novel_fetch_process", Status: "queued", TextModelID: job.TextModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "批量测书后台重试", StartedAt: started}); err != nil {
			result.Failed++
			result.Errors[bookID] = "保存重试任务失败"
			continue
		}
		queued, existingID, err := uniqueQueue.EnqueueUniqueBooks(ctx, job, 24*time.Hour)
		if err != nil {
			_ = updateWorkshopRun(store, job, "failed", "重试入队失败", started, true)
			result.Failed++
			result.Errors[bookID] = "重试入队失败，请检查 Redis 队列"
			continue
		}
		if !queued {
			_ = updateWorkshopRun(store, job, "duplicate_skipped", "该书已有后台任务，本次没有重复入队", started, true)
			result.Skipped++
			result.AlreadyRunning++
			result.Jobs = append(result.Jobs, map[string]any{"id": existingID, "existing": true})
			continue
		}
		result.Queued++
		result.TaskIDs = append(result.TaskIDs, bookID)
		result.Jobs = append(result.Jobs, map[string]any{"id": job.ID, "book_id": bookID, "status": "queued"})
	}
	result.Updated, result.Retried = result.Queued, result.Queued
	result.OK = result.Failed == 0
	return result, nil
}
