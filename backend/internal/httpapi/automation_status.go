package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// automationBookWorkerCeiling 是单个任务内同时处理的书籍数上限。任务可以
// 配置更小的 concurrency，但不能突破这个上限：再多的书同时跑只会把请求
// 堆在网关队列里，反而拖慢整批收口。
const automationBookWorkerCeiling = 4

// automationConcurrencyBudget 把多层并发限制展开给前端，说明任务实际被哪
// 一层卡住。只做展示，不参与调度；网关真实并发会在冷启动后自适应爬升，
// 因此这里标注的是冷启动上限与天花板，而不是瞬时值。
func automationConcurrencyBudget(job map[string]any) map[string]any {
	bookWorkers := automationNumberValue(job["concurrency"], 1)
	if bookWorkers < 1 {
		bookWorkers = 1
	}
	if bookWorkers > automationBookWorkerCeiling {
		bookWorkers = automationBookWorkerCeiling
	}
	images := automationNumberValue(job["imageGlobalConcurrency"], 8)
	if images < 1 {
		images = 1
	}
	coldStart := shuihuoAdaptiveInitialLimit
	effective := images
	limiting := "task_image_concurrency"
	if coldStart < effective {
		effective = coldStart
		limiting = "gateway_cold_start"
	}
	return map[string]any{
		"bookWorkers":               bookWorkers,
		"bookWorkerCeiling":         automationBookWorkerCeiling,
		"taskImageConcurrency":      images,
		"comicBookSlots":            comicBookGenerationConcurrency(),
		"comicBookSlotCeiling":      8,
		"gatewayColdStart":          coldStart,
		"gatewayCeiling":            shuihuoAdaptiveMaxLimit,
		"effectiveImageConcurrency": effective,
		"limitingLayer":             limiting,
		"note":                      "实际并发取各层最小值；网关会从冷启动逐步升到天花板",
	}
}

// automationConcurrencyBudgetLine 把并发预算压成一行中文，追加到书籍的
// 执行日志里。前端日志区已经会渲染这本书的 logs，所以不用改页面也能看到
// “为什么只跑 2 路”。
func automationConcurrencyBudgetLine(job map[string]any) string {
	budget := automationConcurrencyBudget(job)
	return fmt.Sprintf("并发预算：书 %v/%v 路 · 本任务出图 %v 路 · 同时出图书 %v 本 · 网关冷启动 %v 路（上限 %v 路）；实际并发取各层最小值",
		budget["bookWorkers"], budget["bookWorkerCeiling"], budget["taskImageConcurrency"],
		budget["comicBookSlots"], budget["gatewayColdStart"], budget["gatewayCeiling"])
}

type FileAutomationStatusReader struct {
	path       string
	stateStore AutomationStateStore
}

// Path exposes the shared persistence location to the native Go coordinator.
// The status reader and writer must use the same file so the UI observes one
// job regardless of which endpoint created it.
func (r *FileAutomationStatusReader) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

func NewFileAutomationStatusReader(path string) *FileAutomationStatusReader {
	return &FileAutomationStatusReader{path: strings.TrimSpace(path)}
}

func (r *FileAutomationStatusReader) SetStateStore(store AutomationStateStore) {
	if r != nil {
		r.stateStore = store
	}
}

func (r *FileAutomationStatusReader) Status(owner, batchID string) (map[string]any, error) {
	idle := map[string]any{"state": "idle", "counts": idleAutomationCounts()}
	if r == nil {
		return idle, nil
	}
	var state automationStateFile
	var splitLayout bool
	var err error
	if r.stateStore != nil {
		state, _, err = r.stateStore.Load(context.Background())
	} else if r.path != "" {
		state, splitLayout, err = readAutomationState(r.path)
	} else {
		return idle, nil
	}
	if err != nil {
		return nil, err
	}
	changedIDs := map[string]struct{}{}
	repaired := normalizeAlreadyUploadedAutomationStateInto(&state, changedIDs)
	if repairStaleRunningAutomationBooksInto(&state, time.Now(), changedIDs) {
		repaired = true
	}
	if reconcileTerminalAutomationJobsInto(&state, time.Now(), changedIDs) {
		repaired = true
	}
	if repaired {
		var writeErr error
		if r.stateStore != nil {
			if saver, ok := r.stateStore.(AutomationJobSaver); ok {
				if jobs := automationJobsSubset(state, changedIDs); len(jobs) > 0 {
					writeErr = saver.SaveJobs(context.Background(), jobs)
				}
			} else {
				writeErr = r.stateStore.Save(context.Background(), state)
			}
		} else if splitLayout {
			writeErr = writeSplitAutomationState(r.path, state)
		} else {
			writeErr = writeAutomationStateFile(r.path, state)
		}
		if writeErr != nil {
			return nil, fmt.Errorf("repair automation status: %w", writeErr)
		}
	}
	matched := make([]map[string]any, 0, len(state.Jobs))
	for _, job := range state.Jobs {
		if strings.TrimSpace(toString(job["owner"])) != strings.TrimSpace(owner) || strings.TrimSpace(toString(job["batchId"])) != strings.TrimSpace(batchID) {
			continue
		}
		matched = append(matched, job)
	}
	if len(matched) > 0 {
		// Jobs are persisted in a JSON object, whose iteration order is not
		// defined. Always expose the newest run; otherwise an old failed run
		// can randomly hide a later successful run in the UI.
		sort.SliceStable(matched, func(i, j int) bool {
			left := automationJobTime(matched[i])
			right := automationJobTime(matched[j])
			if left.Equal(right) {
				return toString(matched[i]["id"]) > toString(matched[j]["id"])
			}
			return left.After(right)
		})
		return publicAutomationJob(matched[0]), nil
	}
	return idle, nil
}

// automationStaleBookTimeout 是“任务已终止但书仍显示执行中”的判定阈值。
// 正常在途的书每分钟都会刷新 updatedAt 心跳，超过该阈值仍停留在 running
// 只可能是上次执行被中断留下的假状态。
const automationStaleBookTimeout = 5 * time.Minute

// automationProcessStartedAt 记录进程启动时间。任务心跳早于该时间，说明那
// 是上一次进程留下的 running 状态：重启后执行协程已不存在，不可能再继续，
// 必须立刻改判为 failed，否则重试会把它们当成“正在执行”跳过。
var automationProcessStartedAt = time.Now().UTC()

// repairStaleRunningAutomationBooks 把已终止任务（失败/完成/停止/暂停）里
// 遗留的 running 书改判为 failed，并回收重启前遗留的 running 任务。
// 重试逻辑会把 running 当成“正在执行”而跳过，界面上的“重试失败小说”也会
// 按 failed 计数显示 0，导致用户点了也没有任何效果。
func repairStaleRunningAutomationBooks(state *automationStateFile, now time.Time) bool {
	return repairStaleRunningAutomationBooksInto(state, now, nil)
}

// repairStaleRunningAutomationBooksInto 在修复的同时记录被改写的 jobID，
// 调用方可以只落库这些任务，而不是重写全部 job_json。
func repairStaleRunningAutomationBooksInto(state *automationStateFile, now time.Time, changedIDs map[string]struct{}) bool {
	if state == nil {
		return false
	}
	changed := false
	for jobID, job := range state.Jobs {
		if repairInterruptedRunningAutomationJob(job, now) {
			changed = true
			markAutomationJobChanged(changedIDs, jobID)
		}
		if repairStaleRunningAutomationJob(job, now) {
			changed = true
			markAutomationJobChanged(changedIDs, jobID)
		}
	}
	return changed
}

// repairInterruptedRunningAutomationJob 处理“进程重启”这一特定中断：任务
// 仍是 running，但最后一次心跳早于本进程启动时间，说明执行协程已经随旧
// 进程消失。直接改判失败，让用户/自动化可以立刻重试。
func repairInterruptedRunningAutomationJob(job map[string]any, now time.Time) bool {
	if job == nil || automationStringValue(job["state"], "") != "running" {
		return false
	}
	if updated, ok := automationEntryUpdatedAt(job); ok && !updated.Before(automationProcessStartedAt) {
		return false
	}
	books, _ := job["books"].(map[string]any)
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "running", "queued":
			entry["status"] = "failed"
			entry["message"] = "服务重启导致执行中断，可直接重试"
			entry["updatedAt"] = now
			delete(entry, "retryAt")
			entry["autoContinue"] = map[string]any{"reason": "服务重启导致执行中断，等待重试", "willContinue": false}
		}
	}
	if len(books) > 0 {
		job["counts"] = automationCountsFromBooks(books)
	}
	job["state"] = "failed"
	job["message"] = "服务重启导致执行中断，可直接重试"
	job["updatedAt"] = now
	delete(job, "retryAt")
	return true
}

func repairStaleRunningAutomationJob(job map[string]any, now time.Time) bool {
	if job == nil || automationStringValue(job["state"], "") == "running" {
		return false
	}
	books, _ := job["books"].(map[string]any)
	if len(books) == 0 {
		return false
	}
	changed := false
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil || automationStringValue(entry["status"], "") != "running" {
			continue
		}
		if updated, ok := automationEntryUpdatedAt(entry); ok && now.Sub(updated) < automationStaleBookTimeout {
			continue
		}
		entry["status"] = "failed"
		entry["message"] = "上次执行已中断，可直接重试"
		entry["updatedAt"] = now
		delete(entry, "retryAt")
		entry["autoContinue"] = map[string]any{"reason": "上次执行已中断，等待重试", "willContinue": false}
		changed = true
	}
	if changed {
		// 只刷新计数，不能调用 normalizeAutomationJobCounts：它会按计数反推
		// 任务状态，把用户主动暂停的任务改写成 failed。
		job["counts"] = automationCountsFromBooks(books)
		job["updatedAt"] = now
	}
	return changed
}

func automationEntryUpdatedAt(entry map[string]any) (time.Time, bool) {
	value := strings.TrimSpace(toString(entry["updatedAt"]))
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

// reconcileTerminalAutomationJobs repairs an impossible aggregate state left
// by an interrupted/stalled coordinator: every tracked book has already
// reached a terminal state, no delayed retry exists, yet the parent job still
// says running. The parent must never block Start/Retry after all child work is
// finished.
func reconcileTerminalAutomationJobs(state *automationStateFile, now time.Time) bool {
	return reconcileTerminalAutomationJobsInto(state, now, nil)
}

func reconcileTerminalAutomationJobsInto(state *automationStateFile, now time.Time, changedIDs map[string]struct{}) bool {
	if state == nil {
		return false
	}
	changed := false
	for jobID, job := range state.Jobs {
		if reconcileTerminalAutomationJob(job, now) {
			changed = true
			markAutomationJobChanged(changedIDs, jobID)
		}
	}
	return changed
}

func reconcileTerminalAutomationJob(job map[string]any, now time.Time) bool {
	if job == nil || automationStringValue(job["state"], "") != "running" {
		return false
	}
	books, _ := job["books"].(map[string]any)
	if len(books) == 0 || automationPendingBookRetry(books) {
		return false
	}
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if !automationEntryTerminal(entry) {
			return false
		}
	}
	normalizeAutomationJobCounts(job, books)
	state := automationStringValue(job["state"], "")
	if state == "failed" || state == "completed" {
		job["completedAt"] = now
	}
	job["updatedAt"] = now
	return true
}

// readAutomationState supports both the legacy single JSON file and the
// current split layout used for live jobs. The writer migrated to split files
// so a growing batch cannot rewrite one large file on every progress update.
func readAutomationState(path string) (automationStateFile, bool, error) {
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{}}
	indexPath := filepath.Join(filepath.Dir(path), "automation", "index.json")
	indexRaw, indexErr := os.ReadFile(indexPath)
	if indexErr == nil {
		var index struct {
			Jobs []map[string]any `json:"jobs"`
		}
		if err := json.Unmarshal(indexRaw, &index); err != nil {
			return state, true, fmt.Errorf("decode automation index: %w", err)
		}
		jobsDir := filepath.Join(filepath.Dir(indexPath), "jobs")
		for _, entry := range index.Jobs {
			jobID := strings.TrimSpace(toString(entry["id"]))
			if jobID == "" {
				continue
			}
			jobRaw, err := os.ReadFile(filepath.Join(jobsDir, jobID+".json"))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return state, true, fmt.Errorf("read automation job: %w", err)
			}
			var job map[string]any
			if err := json.Unmarshal(jobRaw, &job); err != nil {
				return state, true, fmt.Errorf("decode automation job: %w", err)
			}
			state.Jobs[jobID] = job
		}
		return state, true, nil
	}
	if !errors.Is(indexErr, os.ErrNotExist) {
		return state, true, fmt.Errorf("read automation index: %w", indexErr)
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return state, false, nil
	}
	if err != nil {
		return state, false, fmt.Errorf("read automation status: %w", err)
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, false, fmt.Errorf("decode automation status: %w", err)
	}
	if state.Jobs == nil {
		state.Jobs = map[string]map[string]any{}
	}
	return state, false, nil
}

func writeSplitAutomationState(path string, state automationStateFile) error {
	jobsDir := filepath.Join(filepath.Dir(path), "automation", "jobs")
	if err := os.MkdirAll(jobsDir, 0o700); err != nil {
		return err
	}
	for jobID, job := range state.Jobs {
		if err := writeAutomationJSONAtomic(filepath.Join(jobsDir, jobID+".json"), job); err != nil {
			return err
		}
	}
	return nil
}

// normalizeAlreadyUploadedAutomationState repairs old task records written by
// the pre-reconciliation publisher. N8 reports IMAGE_ALREADY_IN_CAROUSEL when
// the images are already accepted by a carousel, but the old code stored that
// response as a hard upload failure. Treating it as completed prevents a
// second upload and makes the live log reflect the actual remote result.
func normalizeAlreadyUploadedAutomationState(state *automationStateFile) bool {
	return normalizeAlreadyUploadedAutomationStateInto(state, nil)
}

func normalizeAlreadyUploadedAutomationStateInto(state *automationStateFile, changedIDs map[string]struct{}) bool {
	if state == nil {
		return false
	}
	changed := false
	for jobID, job := range state.Jobs {
		books, _ := job["books"].(map[string]any)
		if len(books) == 0 {
			continue
		}
		jobChanged := false
		jobTouched := false
		for _, raw := range books {
			entry, _ := raw.(map[string]any)
			if entry == nil {
				continue
			}
			if automationStringValue(entry["status"], "") == "completed" {
				if automationBoolValue(entry["retryRequested"]) || entry["retryAt"] != nil || entry["retryReason"] != nil {
					delete(entry, "retryRequested")
					delete(entry, "retryAt")
					delete(entry, "retryReason")
					changed = true
					jobTouched = true
				}
				continue
			}
			if automationStringValue(entry["status"], "") != "failed" {
				continue
			}
			errorText := automationStringValue(entry["error"], "")
			if !n8ImagesAlreadyUploadedError(errorText) {
				continue
			}
			entry["status"] = "completed"
			entry["stage"] = "uploaded"
			entry["message"] = "图片已上传完成，N8 已确认图片已占用；已停止重复上传"
			entry["publishError"] = errorText
			entry["error"] = ""
			entry["retryRequested"] = false
			delete(entry, "retryAt")
			changed = true
			jobChanged = true
			jobTouched = true
		}
		if jobChanged {
			normalizeAutomationJobCounts(job, books)
		}
		if jobTouched {
			markAutomationJobChanged(changedIDs, jobID)
		}
	}
	return changed
}

func n8ImagesAlreadyUploadedError(text string) bool {
	text = strings.ToUpper(strings.TrimSpace(text))
	return strings.Contains(text, "IMAGE_ALREADY_IN_CAROUSEL") ||
		strings.Contains(text, "ALREADY_IN_CAROUSEL") ||
		strings.Contains(text, "图片已被其他轮播图集占用")
}

// automationJobsSubset 抽取需要落库的任务副本，用于单任务增量保存。
func automationJobsSubset(state automationStateFile, changedIDs map[string]struct{}) map[string]map[string]any {
	jobs := make(map[string]map[string]any, len(changedIDs))
	for jobID := range changedIDs {
		if job, ok := state.Jobs[jobID]; ok {
			jobs[jobID] = job
		}
	}
	return jobs
}

// markAutomationJobChanged 在收集器非空时记录被改写的 jobID。
func markAutomationJobChanged(changedIDs map[string]struct{}, jobID string) {
	if changedIDs == nil || jobID == "" {
		return
	}
	changedIDs[jobID] = struct{}{}
}

func normalizeAutomationJobCounts(job map[string]any, books map[string]any) {
	counts := automationCountsFromBooks(books)
	job["counts"] = counts
	failed := automationNumberValue(counts["failed"], 0)
	running := automationNumberValue(counts["running"], 0) + automationNumberValue(counts["queued"], 0)
	pending := automationNumberValue(counts["pending"], 0)
	if failed > 0 {
		job["state"] = "failed"
	} else if running > 0 {
		job["state"] = "running"
	} else if pending > 0 {
		job["state"] = "scheduled"
	} else {
		job["state"] = "completed"
	}
}

func writeAutomationStateFile(path string, state automationStateFile) error {
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tmp := path + ".repair.tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func automationJobTime(job map[string]any) time.Time {
	for _, key := range []string{"updatedAt", "completedAt", "startedAt", "createdAt"} {
		value := strings.TrimSpace(toString(job[key]))
		if value == "" {
			continue
		}
		if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
			return parsed
		}
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func idleAutomationCounts() map[string]any {
	return map[string]any{"total": 0, "ready": 0, "running": 0, "pending": 0, "failed": 0, "blocked": 0}
}

// automationBookAutoContinue 告诉前端“系统会不会自己继续”，并说明缺什么。
// 状态接口不能再让 stopped/needs_manual 表现为“什么都没有”。
func automationBookAutoContinue(book map[string]any) map[string]any {
	status := strings.TrimSpace(toString(book["status"]))
	switch status {
	case "queued":
		return map[string]any{"willContinue": true, "reason": "已排队，等待并发槽位"}
	case "running":
		return map[string]any{"willContinue": true, "reason": "系统正在自动执行"}
	case "paused":
		return map[string]any{"willContinue": false, "reason": "已暂停，需手动继续"}
	case "stopped":
		return map[string]any{"willContinue": false, "reason": "已停止，需手动继续"}
	case "needs_manual":
		pages := automationNeedsManualPages(map[string]any{automationStringValue(book["bookId"], ""): book})
		reason := "需人工处理"
		if len(pages) > 0 {
			// 优先展示页级已经生成好的中文原因（含余额、审核等结论），
			// 不再把 AUTH_OR_QUOTA 这类内部码直接抛给用户。
			detail := strings.TrimSpace(toString(pages[0]["message"]))
			if index := strings.Index(detail, "\n"); index >= 0 {
				detail = strings.TrimSpace(detail[:index])
			}
			if detail == "" {
				detail = automationManualReasonText(toString(pages[0]["code"]))
			}
			reason = "需人工处理 " + toString(pages[0]["label"]) + "：" + detail
		}
		return map[string]any{"willContinue": false, "reason": reason}
	case "failed":
		if automationBoolValue(book["retryRequested"]) {
			return map[string]any{"willContinue": true, "reason": "已安排自动续跑"}
		}
		return map[string]any{"willContinue": false, "reason": "已失败，等待重试"}
	case "completed":
		return map[string]any{"willContinue": false, "reason": "已完成"}
	}
	return map[string]any{"willContinue": false, "reason": "状态未知，需人工确认"}
}

// publicAutomationBookLogs 只返回最近 limit 条日志，避免状态接口随任务时长膨胀。
func publicAutomationBookLogs(raw any, limit int) []any {
	logs := make([]any, 0, 8)
	switch typed := raw.(type) {
	case []any:
		logs = append(logs, typed...)
	case []map[string]any:
		for _, item := range typed {
			logs = append(logs, item)
		}
	}
	if limit > 0 && len(logs) > limit {
		logs = logs[len(logs)-limit:]
	}
	return logs
}

func publicAutomationJob(job map[string]any) map[string]any {
	result := map[string]any{
		"id": job["id"], "batchId": job["batchId"], "state": job["state"],
		"stopAt": job["stopAt"], "autoPublish": job["autoPublish"],
		"runMode": job["runMode"], "scheduleMode": job["scheduleMode"], "concurrency": job["concurrency"], "preset": job["preset"],
		"scheduledAt": job["scheduledAt"], "createdAt": job["createdAt"], "startedAt": job["startedAt"],
		"updatedAt": job["updatedAt"], "completedAt": job["completedAt"], "lastError": job["lastError"],
		"counts":            job["counts"],
		"priorityBookIds":   automationPriorityIDs(job),
		"concurrencyBudget": automationConcurrencyBudget(job),
	}
	if result["autoPublish"] == nil {
		result["autoPublish"] = false
	}
	if strings.TrimSpace(toString(result["runMode"])) == "" {
		result["runMode"] = "video_no_submit"
	}
	if rawBooks, ok := job["books"].(map[string]any); ok && len(rawBooks) > 0 && strings.TrimSpace(toString(job["state"])) != "stopped" {
		// 旧 job 的 counts 可能停留在启动快照；这里按 books 的实际状态重新派生，
		// 保证一本书失败/完成后状态接口立即反映。
		result["counts"] = automationCountsFromBooks(rawBooks)
	}
	if result["counts"] == nil {
		result["counts"] = idleAutomationCounts()
	}
	// 批次级“需人工页清单”：全批次跑完后，前端不必逐本点开就能看到
	// 哪一页需要人工、原因、已自动改写几次、最后拒绝原文。
	if rawBooks, ok := job["books"].(map[string]any); ok {
		result["needsManualPages"] = automationNeedsManualPages(rawBooks)
	} else {
		result["needsManualPages"] = []map[string]any{}
	}
	books := []any{}
	if raw, ok := job["books"].(map[string]any); ok {
		for _, value := range raw {
			book, ok := value.(map[string]any)
			if !ok {
				continue
			}
			status := strings.TrimSpace(toString(book["status"]))
			retryAt := book["retryAt"]
			errorValue := book["error"]
			retryRequested := book["retryRequested"]
			batchRetryPending := book["batchRetryPending"]
			batchRetryCount := book["batchRetryCount"]
			publishError := book["publishError"]
			// Old jobs may have reached published/completed after a successful
			// continuation but still carry the previous partial-run error. The
			// completed state is authoritative; expose no live failure/retry badge.
			if status == "completed" {
				errorValue = ""
				publishError = ""
				retryRequested = nil
				retryAt = nil
				batchRetryPending = nil
				batchRetryCount = nil
			}
			// retryAt is meaningful only while this book is still running.
			// Older failed jobs may retain the marker from their last attempt;
			// do not expose that stale timestamp as a future automatic retry.
			if status != "running" && !automationBoolValue(book["retryRequested"]) {
				retryAt = nil
			}
			// 历史状态里可能残留英文哨兵前缀；在返回给界面之前统一清掉。
			messageValue := batchfactoryv11.UserFacingErrorText(automationStringValue(book["message"], ""))
			if text, ok := errorValue.(string); ok && text != "" {
				errorValue = batchfactoryv11.UserFacingErrorText(text)
			}
			failureCategory, failureCategoryLabel := automationFailureCategory(status, messageValue, toString(errorValue), toString(publishError))
			books = append(books, map[string]any{
				"bookId": book["bookId"], "title": book["title"], "status": book["status"], "stage": book["stage"],
				"message": messageValue, "error": errorValue, "publishError": publishError, "updatedAt": book["updatedAt"],
				"attempts": book["attempts"], "retryCount": book["retryCount"], "retryAt": retryAt,
				"retryRequested": retryRequested, "logs": publicAutomationBookLogs(book["logs"], 50),
				"priorityRequested": book["priorityRequested"],
				"batchRetryPending": batchRetryPending, "batchRetryCount": batchRetryCount,
				"needsManualPages": automationNeedsManualPages(map[string]any{automationStringValue(book["bookId"], ""): book}),
				"autoContinue":     automationBookAutoContinue(book),
				"failureCategory":  failureCategory, "failureCategoryLabel": failureCategoryLabel,
			})
		}
	}
	result["books"] = books
	return result
}

// automationFailureCategory 把失败原因归成四类，前端据此给出对应操作按钮，
// 用户不需要从一长串错误文本里自己判断“重试有没有用、会不会再扣费”。
func automationFailureCategory(status, message, errText, publishError string) (string, string) {
	if strings.TrimSpace(status) != "failed" {
		return "", ""
	}
	text := strings.Join([]string{message, errText, publishError}, " ")
	lower := strings.ToUpper(text)
	switch {
	case strings.Contains(lower, "IMAGE_BILLED_NOT_RETRIEVED") || (strings.Contains(text, "已生成") && strings.Contains(text, "需人工处理")):
		return "billed_not_retrieved", "已计费未取回（重试会再扣一次图费）"
	case strings.Contains(lower, "TIMEOUT") || strings.Contains(lower, "DEADLINE") || strings.Contains(text, "超时") || strings.Contains(text, "没有任何进展"):
		return "upstream_timeout", "上游超时（可直接重试）"
	case strings.Contains(text, "没有同名作品") || strings.Contains(text, "漫画合集"):
		return "external_work_missing", "外部作品缺失（需先建作品或指定关联）"
	case strings.Contains(text, "未配置") || strings.Contains(text, "尚未完成识别") || strings.Contains(text, "没有可用音频") || strings.Contains(text, "缺少可用于音频"):
		return "config_missing", "配置缺失（需先补齐设置）"
	}
	return "other", "其他失败（可重试）"
}

func registerAutomationStatusRoute(mux *http.ServeMux, auth BridgeAuth, reader *FileAutomationStatusReader, version string) {
	if mux == nil || reader == nil {
		return
	}
	mux.Handle("GET /api/batch-factory/"+version+"/batches/{batchId}/automation", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		status, err := reader.Status(owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"automation": status})
	})))
}
