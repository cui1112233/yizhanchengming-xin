package httpapi

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

const novelFetchAutomationConfigKey = "novel_fetch_automation"
const defaultNovelFetchOriginalFetchRetryInterval = time.Hour

func novelFetchAutomationEnabled(settings map[string]any) bool {
	return mapOrEmpty(settings[novelFetchAutomationConfigKey])["enabled"] == true
}

func novelFetchSensitiveAIEnabled(settings map[string]any) bool {
	if settings == nil {
		return false
	}
	if enabled, ok := settings["sensitive_ai_enabled"].(bool); ok {
		return enabled
	}
	if enabled, ok := settings["sensitiveAiEnabled"].(bool); ok {
		return enabled
	}
	ai := mapOrEmpty(settings["sensitive_ai"])
	if enabled, ok := ai["enabled"].(bool); ok {
		return enabled
	}
	return mapOrEmpty(mapOrEmpty(settings["appConfig"])["sensitive_ai"])["enabled"] == true
}

func saveNovelFetchAutomationEnabled(ctx context.Context, store novelfetchworkshop.Store, owner string, enabled bool) error {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return err
	}
	if settings == nil {
		settings = map[string]any{}
	}
	automation := mapOrEmpty(settings[novelFetchAutomationConfigKey])
	automation["enabled"] = enabled
	automation["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	settings[novelFetchAutomationConfigKey] = automation
	return store.PutConfig(ctx, owner, settings)
}

func novelFetchAutomationVersions(settings map[string]any) []string {
	automation := mapOrEmpty(settings[novelFetchAutomationConfigKey])
	versions := cleanStrings(anyStrings(automation["selected_versions"]))
	if len(versions) == 0 {
		web := mapOrEmpty(mapOrEmpty(settings["appConfig"])["web_submit"])
		versions = cleanStrings(anyStrings(web["submit_versions"]))
	}
	if len(versions) == 0 {
		versions = []string{"ai1"}
	}
	return versions
}

func novelFetchOriginalFetchRetryInterval(settings map[string]any) time.Duration {
	minutes := intValue(mapOrEmpty(settings["fetch"])["original_fetch_retry_interval_minutes"])
	if minutes <= 0 {
		minutes = intValue(mapOrEmpty(settings[novelFetchAutomationConfigKey])["original_fetch_retry_interval_minutes"])
	}
	if minutes <= 0 {
		return defaultNovelFetchOriginalFetchRetryInterval
	}
	if minutes > 24*60 {
		minutes = 24 * 60
	}
	return time.Duration(minutes) * time.Minute
}

// novelFetchAIManualRetryRequired 判断错误是否属于模型确定性失败。参考实现
// 对这类错误写 manual_ai_retry_required 并禁止自动重试：同一段超长文本反复
// 提交只会重复截断，既消耗额度也不可能自行恢复。
func novelFetchAIManualRetryRequired(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	for _, keyword := range []string{"文本模型输出被长度截断", "长度截断", "content policy", "内容审核", "模型拒绝", "refusal"} {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

// markNovelFetchAIManualRetry 把模型确定性失败落成人工状态：写 manualOnly、
// manualAiRetryRequired 和可读原因，让扫描器不再自动接管。用户点批量重试时
// 由 resetNovelFetchAutomationRetryState 清除这些标记。
func markNovelFetchAIManualRetry(ctx context.Context, store novelfetchworkshop.Store, owner, bookID string, failure error, now time.Time) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	reason := "文本模型无法完成本书，请更换输出上限更大的模型或减小单次文本量后重试：" + failure.Error()
	document.Meta["status"] = "failed"
	document.Meta["queueState"] = "failed"
	document.Meta["manualOnly"] = true
	document.Meta["manualAiRetryRequired"] = true
	document.Meta["error"] = reason
	delete(document.Meta, "automationNextRetryAt")
	appendNovelFetchDocumentLog(&document, "queue_failed", reason, now)
	return store.PutDocument(ctx, owner, document)
}

func markNovelFetchOriginalFetchRetry(ctx context.Context, store novelfetchworkshop.Store, owner, bookID string, fetchErr error, now time.Time) error {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return err
	}
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	delay := novelFetchOriginalFetchRetryInterval(settings)
	document.Meta["status"] = "waiting_retry"
	document.Meta["queueState"] = "waiting_retry"
	document.Meta["originalStatus"] = "waiting_fetch_retry"
	document.Meta["originalFetchRetryAttempts"] = intValue(document.Meta["originalFetchRetryAttempts"]) + 1
	document.Meta["originalFetchNextRetryAt"] = now.Add(delay).UTC().Format(time.RFC3339Nano)
	document.Meta["originalFetchLastError"] = firstText(fetchErr)
	document.Meta["error"] = firstText(fetchErr)
	return store.PutDocument(ctx, owner, document)
}

func clearNovelFetchOriginalFetchRetry(ctx context.Context, store novelfetchworkshop.Store, owner, bookID string) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		return nil
	}
	for _, key := range []string{"originalFetchRetryAttempts", "originalFetchNextRetryAt", "originalFetchLastError"} {
		delete(document.Meta, key)
	}
	return store.PutDocument(ctx, owner, document)
}

func novelFetchAutomationEligible(document novelfetchworkshop.Document, now time.Time) bool {
	meta := document.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	status := strings.ToLower(firstText(meta["status"], meta["queueState"], meta["taskStatus"]))
	// 自动重试预算耗尽后必须由人工介入，不能因为 status=failed 又被扫描器
	// 立刻捡回队列，否则“停止自动重试”实际上没有任何效果。
	if meta["manualOnly"] == true || meta["manual_only"] == true {
		return false
	}
	if meta["rewriteValidationHold"] == true {
		return false
	}
	submitStatus := strings.ToLower(firstText(meta["siteSubmitStatus"], meta["site_submit_status"]))
	if submitStatus == "waiting_configuration" || submitStatus == "waiting_login" {
		return false
	}
	if status == "queued" || status == "running" || status == "processing" {
		return false
	}
	if status == "waiting_classification" {
		return novelFetchClassificationResolved(document)
	}
	// Raw text is persisted before sensitive-word AI runs. If that second
	// stage fails, the task must be picked up again without fetching upstream
	// a second time.
	if strings.EqualFold(firstText(meta["originalStatus"], meta["original_status"]), "waiting_sensitive_ai") &&
		strings.TrimSpace(document.OriginalRaw) != "" {
		return true
	}
	if nextRaw := firstText(meta["originalFetchNextRetryAt"], meta["original_fetch_next_retry_at"], meta["automationNextRetryAt"], meta["automation_next_retry_at"]); nextRaw != "" {
		if next, err := time.Parse(time.RFC3339Nano, nextRaw); err == nil && next.After(now) {
			return false
		}
	}
	if submitStatus == "pending_upload" || submitStatus == "pending_confirmation" || submitStatus == "failed" || submitStatus == "error" {
		return true
	}
	// Older/manual batches can finish rewriting before automatic submission is
	// enabled. They have a completed process state but no submission state, so
	// recover them when automation is turned on instead of leaving them pending.
	if status == "completed" && submitStatus == "" &&
		strings.EqualFold(firstText(meta["originalStatus"], meta["original_status"]), "done") &&
		(strings.EqualFold(firstText(meta["aiStatus"], meta["ai_status"]), "ai_done") ||
			strings.EqualFold(firstText(meta["aiStatus"], meta["ai_status"]), "done") ||
			novelFetchDocumentHasGeneratedAI(document)) {
		return true
	}
	return status == "" || status == "input_ready" || status == "failed" || status == "waiting_retry"
}

func novelFetchAutomationTimeout(settings map[string]any) time.Duration {
	seconds := intValue(mapOrEmpty(settings["workflow"])["auto_queue_task_timeout_seconds"])
	if seconds <= 0 {
		seconds = 180
	}
	if seconds < 30 {
		seconds = 30
	}
	if seconds > 24*60*60 {
		seconds = 24 * 60 * 60
	}
	return time.Duration(seconds) * time.Second
}

// recoverStaleNovelFetchTasks turns orphaned processing states into retryable
// work. A worker that disappears after marking a document processing must not
// leave that document excluded from automation forever.
func recoverStaleNovelFetchTasks(ctx context.Context, store novelfetchworkshop.Store, owner string, settings map[string]any, documents []novelfetchworkshop.Document, now time.Time, queues ...novelfetchworkshop.ProcessQueue) {
	cutoff := now.Add(-novelFetchAutomationTimeout(settings))
	staleBatches := map[string]bool{}
	var recoveryQueue novelfetchworkshop.ProcessStaleJobRecoveryQueue
	if len(queues) > 0 {
		recoveryQueue, _ = queues[0].(novelfetchworkshop.ProcessStaleJobRecoveryQueue)
	}
	for _, document := range documents {
		status := strings.ToLower(firstText(document.Meta["status"], document.Meta["queueState"], document.Meta["taskStatus"]))
		if status != "processing" && status != "running" && status != "ai_processing" && status != "queued" {
			continue
		}
		updated, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(document.UpdatedAt))
		if err != nil || updated.After(cutoff) {
			continue
		}
		if status == "queued" {
			// Age alone does not prove a queued job is orphaned. Never replace a
			// job which Redis still owns, or retry when Redis cannot be checked.
			jobID := firstText(document.Meta["currentProcessJobId"], document.Meta["current_process_job_id"])
			if len(queues) == 0 || jobID == "" {
				continue
			}
			checker, ok := queues[0].(interface {
				ProcessJobExists(context.Context, string) (bool, error)
			})
			if !ok {
				continue
			}
			exists, checkErr := checker.ProcessJobExists(ctx, jobID)
			if checkErr != nil || exists {
				continue
			}
		}
		if recoveryQueue != nil && status != "queued" {
			staleJobID := firstText(document.Meta["currentProcessJobId"], document.Meta["current_process_job_id"])
			if staleJobID == "" {
				batchID := firstText(document.Meta["currentBatchId"], document.Meta["current_batch_id"])
				if runs, runsErr := store.ListRuns(ctx, owner, batchID); runsErr == nil {
					for _, run := range runs {
						if run.Stage == "novel_fetch_process" && strings.EqualFold(strings.TrimSpace(run.Status), "running") {
							staleJobID = run.RunID
							break
						}
					}
				}
			}
			if staleJobID != "" {
				if _, recoverErr := recoveryQueue.RecoverStaleJob(ctx, staleJobID); recoverErr != nil {
					continue
				}
			}
		}
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		message := "任务超过执行超时时间，已转入等待重试"
		document.Meta["status"] = "waiting_retry"
		document.Meta["queueState"] = "waiting_retry"
		document.Meta["error"] = message
		document.Meta["automationNextRetryAt"] = now.UTC().Format(time.RFC3339Nano)
		appendNovelFetchDocumentLog(&document, "process_recovered", "超时任务已转入等待重试", now)
		appendNovelFetchDocumentLog(&document, "process_waiting_retry", message, now)
		if err := store.PutDocument(ctx, owner, document); err == nil {
			staleBatches[firstText(document.Meta["currentBatchId"], document.Meta["current_batch_id"])] = true
		}
	}
	if len(staleBatches) == 0 {
		return
	}
	runs, err := store.ListRuns(ctx, owner, "")
	if err != nil {
		return
	}
	for _, run := range runs {
		if !staleBatches[run.BookID] || strings.ToLower(strings.TrimSpace(run.Status)) != "running" {
			continue
		}
		updated, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(run.UpdatedAt))
		if parseErr != nil || updated.After(cutoff) {
			continue
		}
		run.Status = "waiting_retry"
		run.ErrorMessage = "任务超过执行超时时间，已转入等待重试"
		run.FinishedAt = now.UTC().Format(time.RFC3339Nano)
		_ = store.PutRun(ctx, owner, run)
	}
}

// Older workers could persist an AI body and leave aiStatus at "未开始" or
// "ai_processing". The stored body is the authoritative signal for automatic
// submission; an empty version must never be treated as ready.
func novelFetchDocumentHasGeneratedAI(document novelfetchworkshop.Document) bool {
	for index := 1; index <= 20; index++ {
		versionID := fmt.Sprintf("ai%d", index)
		if value, ok := document.Versions[versionID]; ok {
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" && text != "<nil>" {
				return true
			}
		}
		if ref, ok := document.BodyRefs[versionID]; ok && ref.CharCount > 0 && strings.EqualFold(strings.TrimSpace(ref.State), "ready") {
			return true
		}
	}
	return false
}

// novelFetchAutomationConfigurationBlock identifies failures that cannot be
// corrected by waiting. Retrying these every few seconds hides the actionable
// problem and leaves the automatic queue permanently busy.
func novelFetchAutomationConfigurationBlock(document novelfetchworkshop.Document, submitErr error) string {
	messages := []string{firstText(submitErr)}
	for _, message := range mapOrEmpty(document.Meta["siteSubmitFailedErrors"]) {
		messages = append(messages, firstText(message))
	}
	for _, message := range messages {
		normalized := strings.ToLower(strings.TrimSpace(message))
		if normalized == "" {
			continue
		}
		if strings.Contains(normalized, "风格") || strings.Contains(normalized, "性别") ||
			strings.Contains(normalized, "正文复检未通过") ||
			strings.Contains(normalized, "未找到正文版本") || strings.Contains(normalized, "正文少于") ||
			strings.Contains(normalized, "未配置上传") || strings.Contains(normalized, "未登录") ||
			strings.Contains(normalized, "登录已失效") {
			return message
		}
	}
	return ""
}

func dispatchNovelFetchAutomationOwner(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, now time.Time) (*novelfetchworkshop.ProcessJob, error) {
	if queue == nil {
		return nil, nil
	}
	if reaper, ok := queue.(interface {
		ReapStaleProcessing(context.Context) (int, error)
	}); ok {
		if reaped, reapErr := reaper.ReapStaleProcessing(ctx); reapErr != nil {
			log.Printf("novel fetch automation: reap stale processing owner=%s: %v", owner, reapErr)
		} else if reaped > 0 {
			log.Printf("novel fetch automation: reaped %d stale processing jobs owner=%s", reaped, owner)
		}
	}
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return nil, err
	}
	automationEnabled := novelFetchAutomationEnabled(settings)
	documents, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return nil, err
	}
	recoverStaleNovelFetchTasks(ctx, store, owner, settings, documents, now, queue)
	if refreshed, refreshErr := store.ListDocuments(ctx, owner); refreshErr == nil {
		documents = refreshed
	}
	bookIDs := make([]string, 0, 50)
	batchID := ""
	platformID := "2"
	maxTxt := 4000
	for _, document := range documents {
		if len(bookIDs) >= 50 {
			continue
		}
		waitingFetchRetry := strings.EqualFold(firstText(document.Meta["originalStatus"], document.Meta["original_status"]), "waiting_fetch_retry")
		// A fetch failure has already created an explicit retry deadline. It must
		// survive a paused global automation switch and a service restart; only
		// ordinary pending work is gated by the switch.
		if !automationEnabled && !waitingFetchRetry {
			continue
		}
		eligible := novelFetchAutomationEligible(document, now)
		if !eligible {
			// Body text is stored separately from the lightweight document JSON.
			// Hydrate refs before deciding whether a stale legacy task is ready
			// for automatic submission.
			if refs, refsErr := store.ListBodyRefs(ctx, owner, document.BookID); refsErr == nil && len(refs) > 0 {
				if document.BodyRefs == nil {
					document.BodyRefs = map[string]novelfetchworkshop.BodyRef{}
				}
				for _, ref := range refs {
					document.BodyRefs[ref.VersionID] = ref
				}
				eligible = novelFetchAutomationEligible(document, now)
			}
		}
		if !eligible {
			continue
		}
		bookIDs = append(bookIDs, document.BookID)
		if batchID == "" {
			batchID = firstText(document.Meta["currentBatchId"], document.Meta["current_batch_id"])
			platformID = firstText(document.Meta["platformId"], document.Meta["platform_id"], "2")
			if document.MaxTxt > 0 {
				maxTxt = document.MaxTxt
			}
		}
	}
	if len(bookIDs) == 0 {
		return nil, nil
	}
	stamp := now.UnixNano()
	if batchID == "" {
		batchID = fmt.Sprintf("auto-batch-%d", stamp)
	}
	job := novelfetchworkshop.ProcessJob{
		ID: fmt.Sprintf("queue_%d_%x", now.UnixMilli(), stamp&0xffffffff), Owner: owner, BatchID: batchID,
		BookIDs: bookIDs, PlatformID: platformID, MaxTxt: maxTxt, SensitiveAIEnabled: novelFetchSensitiveAIEnabled(settings), AutoSubmit: true,
		SelectedVersions: novelFetchAutomationVersions(settings), TextModelID: firstText(settings["text_model_id"], settings["textModelId"], "未指定"),
	}
	var enqueueErr error
	if uniqueQueue, ok := queue.(novelfetchworkshop.ProcessBookDedupQueue); ok {
		var queued bool
		queued, _, enqueueErr = uniqueQueue.EnqueueUniqueBooks(ctx, job, 2*time.Hour)
		if enqueueErr == nil && !queued {
			return nil, nil
		}
	} else if uniqueQueue, ok := queue.(novelfetchworkshop.ProcessDedupQueue); ok {
		var queued bool
		queued, enqueueErr = uniqueQueue.EnqueueUnique(ctx, owner+":"+strings.Join(bookIDs, ","), job, 2*time.Hour)
		if enqueueErr == nil && !queued {
			return nil, nil
		}
	} else {
		enqueueErr = queue.Enqueue(ctx, job)
	}
	if enqueueErr != nil {
		log.Printf("novel fetch automation: enqueue failed owner=%s batch=%s books=%d: %v", owner, batchID, len(bookIDs), enqueueErr)
		markNovelFetchAutomationEnqueueRetry(ctx, store, owner, bookIDs, now, enqueueErr)
		return nil, enqueueErr
	}
	for _, bookID := range bookIDs {
		if err := markNovelFetchTaskQueued(ctx, store, owner, bookID, job.ID, batchID, now); err != nil {
			return nil, err
		}
	}
	if err := store.PutRun(ctx, owner, novelfetchworkshop.RunRecord{RunID: job.ID, BookID: batchID, Stage: "novel_fetch_process", Status: "queued", TextModelID: job.TextModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取全自动队列", StartedAt: now.UTC().Format(time.RFC3339Nano)}); err != nil {
		return nil, err
	}
	return &job, nil
}

// novelFetchEnqueueMaxAttempts 是入队失败的自动重试上限。参考实现里重试
// 预算耗尽后会把任务标记为需要人工处理，而不是无限等待；这里保持同样的
// 语义，避免 Redis 长期不可用时任务永远停在“等待重试”。
const novelFetchEnqueueMaxAttempts = 20

// markNovelFetchAutomationEnqueueRetry 在入队失败时写入退避时间。Redis 持续
// 不可用时扫描器每分钟都会重试同一批书；没有退避既会放大数据库写入，也会
// 让“最后一次失败”永远停留在当前秒，无法判断是否已经恢复。
func markNovelFetchAutomationEnqueueRetry(ctx context.Context, store novelfetchworkshop.Store, owner string, bookIDs []string, now time.Time, enqueueErr error) {
	for _, bookID := range bookIDs {
		document, err := store.GetDocument(ctx, owner, bookID)
		if err != nil {
			continue
		}
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		attempts := intValue(document.Meta["automationEnqueueAttempts"]) + 1
		document.Meta["automationEnqueueAttempts"] = attempts
		if attempts >= novelFetchEnqueueMaxAttempts {
			document.Meta["status"] = "failed"
			document.Meta["queueState"] = "failed"
			document.Meta["manualOnly"] = true
			document.Meta["error"] = fmt.Sprintf("任务入队连续失败 %d 次，已停止自动重试：%s", attempts, firstText(enqueueErr))
			delete(document.Meta, "automationNextRetryAt")
			appendNovelFetchDocumentLog(&document, "queue_failed",
				fmt.Sprintf("入队连续失败 %d 次，已转人工处理", attempts), now)
			_ = store.PutDocument(ctx, owner, document)
			continue
		}
		delay := 30 * time.Second
		for step := 1; step < attempts && delay < 15*time.Minute; step++ {
			delay *= 2
		}
		if delay > 15*time.Minute {
			delay = 15 * time.Minute
		}
		document.Meta["status"] = "waiting_retry"
		document.Meta["queueState"] = "waiting_retry"
		document.Meta["error"] = firstText(enqueueErr)
		document.Meta["automationNextRetryAt"] = now.Add(delay).UTC().Format(time.RFC3339Nano)
		appendNovelFetchDocumentLog(&document, "process_waiting_retry", "入队失败："+firstText(enqueueErr), now)
		_ = store.PutDocument(ctx, owner, document)
	}
}

func markNovelFetchAutomationSubmitResult(ctx context.Context, store novelfetchworkshop.Store, owner string, bookIDs []string, submitErr error) {
	now := time.Now().UTC()
	for _, bookID := range bookIDs {
		document, err := store.GetDocument(ctx, owner, bookID)
		if err != nil {
			continue
		}
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		status := strings.ToLower(firstText(document.Meta["siteSubmitStatus"], document.Meta["site_submit_status"]))
		if submitErr == nil && status == "submitted" {
			delete(document.Meta, "automationSubmitAttempts")
			delete(document.Meta, "automationNextRetryAt")
			delete(document.Meta, "automationSubmitError")
		} else if workshopSubmitRequiresLogin(submitErr) {
			document.Meta["siteSubmitStatus"] = "waiting_login"
			document.Meta["status"], document.Meta["queueState"] = "waiting_login", "waiting_login"
			document.Meta["automationSubmitError"] = workshopLoginRequiredMessage
			delete(document.Meta, "automationSubmitAttempts")
			delete(document.Meta, "automationNextRetryAt")
		} else if blocked := novelFetchAutomationConfigurationBlock(document, submitErr); blocked != "" {
			document.Meta["siteSubmitStatus"] = "waiting_configuration"
			document.Meta["status"] = "waiting_configuration"
			document.Meta["queueState"] = "waiting_configuration"
			document.Meta["automationSubmitError"] = blocked
			delete(document.Meta, "automationSubmitAttempts")
			delete(document.Meta, "automationNextRetryAt")
		} else {
			attempts := intValue(document.Meta["automationSubmitAttempts"]) + 1
			delay := 30 * time.Second
			for step := 1; step < attempts && delay < 15*time.Minute; step++ {
				delay *= 2
			}
			if delay > 15*time.Minute {
				delay = 15 * time.Minute
			}
			if !strings.EqualFold(firstText(document.Meta["siteSubmitStatus"]), "pending_confirmation") {
				document.Meta["siteSubmitStatus"] = "pending_upload"
			}
			document.Meta["automationSubmitAttempts"] = attempts
			document.Meta["automationNextRetryAt"] = now.Add(delay).Format(time.RFC3339Nano)
			document.Meta["automationSubmitError"] = firstText(submitErr, document.Meta["siteSubmitFailedErrors"], "上传尚未确认")
		}
		_ = store.PutDocument(ctx, owner, document)
	}
}

func novelFetchAutomationStatus(ctx context.Context, store novelfetchworkshop.Store, registry *workshopProcessRegistry, owner string) (map[string]any, error) {
	var settings map[string]any
	var documents []novelfetchworkshop.Document
	var err error
	if reader, ok := store.(interface {
		AutomationStatusInputs(context.Context, string) (map[string]any, []novelfetchworkshop.Document, error)
	}); ok {
		settings, documents, err = reader.AutomationStatusInputs(ctx, owner)
	} else {
		settings, err = store.GetConfig(ctx, owner)
		if err == nil {
			documents, err = store.ListDocuments(ctx, owner)
		}
	}
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"running": 0, "queued": 0, "waiting_retry": 0, "waiting_configuration": 0, "waiting_login": 0, "failed": 0, "stopped": 0}
	for _, document := range documents {
		task := legacyNovelFetchTask(document)
		status := strings.ToLower(firstText(document.Meta["status"], document.Meta["queueState"], document.Meta["taskStatus"], task["ai_status"], document.Meta["aiStatus"]))
		switch {
		case status == "running" || status == "processing":
			counts["running"]++
		case status == "queued":
			counts["queued"]++
		case status == "waiting_retry":
			counts["waiting_retry"]++
		case status == "waiting_configuration":
			counts["waiting_configuration"]++
		case status == "waiting_login":
			counts["waiting_login"]++
		case status == "failed" || status == "error":
			counts["failed"]++
		case status == "stopped" || status == "cancelled":
			counts["stopped"]++
		}
	}
	latest := map[string]any{}
	activeJobID := ""
	if jobs := registry.latest(owner, 1); len(jobs) > 0 {
		latest = jobs[0]
	}
	if len(latest) == 0 {
		if runs, runErr := store.ListRuns(ctx, owner, ""); runErr == nil {
			for _, run := range runs {
				if run.Stage == "novel_fetch_process" {
					latest = persistedWorkshopRun(run)
					break
				}
			}
		}
	}
	status := strings.ToLower(firstText(latest["status"], latest["queue_state"]))
	if status == "queued" || status == "running" || status == "waiting_retry" {
		activeJobID = firstText(latest["id"])
	}
	return map[string]any{
		"ok": true, "enabled": novelFetchAutomationEnabled(settings), "auto_queue_enabled": novelFetchAutomationEnabled(settings),
		"active_job_id": activeJobID, "counts": counts, "processing_count": counts["running"],
		"pending_count": counts["queued"] + counts["waiting_retry"], "latest_job": novelFetchJobStatusSummary(latest),
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}
