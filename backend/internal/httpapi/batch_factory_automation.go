package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/batchfactoryv11/external"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

// BatchAutomationCoordinator is the Go owner of the V12 start/stop controls.
// It intentionally persists the same public job shape that the old worker
// used, so the existing workbench does not need a second status protocol.
type BatchAutomationCoordinator struct {
	store                 batchfactoryv11.Store
	presets               AutomationPresetStore
	stages                *batchfactoryv11.BookStageService
	merge                 *batchfactoryv11.MergeService
	external              *external.Service
	ttsStore              novelfetchworkshop.Store
	fetcher               *novelfetchworkshop.UpstreamFetcher
	classification        batchFetchClassification
	audioProbe            batchfactoryv11.H3AudioDurationProbe
	localFiles            *localartifact.Store
	path                  string
	automationStore       AutomationStateStore
	automationState       *automationStateFile
	now                   func() time.Time
	providerSync          func(context.Context, string) error
	promptCatalog         *prompt.Catalog
	publish121            *BatchFactory121Publisher
	comicGenerate         func(context.Context, string, string, string) (batchfactoryv11.ComicCollection, error)
	comicPublish          func(context.Context, string, string, string, batchfactoryv11.ComicCollection, batchfactoryv11.SettingsPatch) (N8ComicPublishRecord, error)
	comicPublishPreflight func(context.Context, batchfactoryv11.SettingsPatch) error
	mu                    sync.Mutex
	controlMu             sync.Mutex
	startMu               sync.Mutex
	runCancels            map[string]context.CancelFunc
	dispatchNextCheck     time.Time
	dispatchNextCheckSet  bool
	dispatcherOnce        sync.Once
}

type comicGenerationProgressFunc func(stage, message string)

type comicGenerationProgressContextKey struct{}
type comicAssetConcurrencyContextKey struct{}

func withComicGenerationProgress(ctx context.Context, report comicGenerationProgressFunc) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, comicGenerationProgressContextKey{}, report)
}

func comicGenerationProgressFromContext(ctx context.Context) comicGenerationProgressFunc {
	if ctx == nil {
		return nil
	}
	report, _ := ctx.Value(comicGenerationProgressContextKey{}).(comicGenerationProgressFunc)
	return report
}

func withComicAssetConcurrency(ctx context.Context, concurrency int) context.Context {
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > 5 {
		concurrency = 5
	}
	return context.WithValue(ctx, comicAssetConcurrencyContextKey{}, concurrency)
}

func comicAssetConcurrencyFromContext(ctx context.Context) int {
	if ctx == nil {
		return 1
	}
	if concurrency, ok := ctx.Value(comicAssetConcurrencyContextKey{}).(int); ok && concurrency > 0 {
		if concurrency > 5 {
			return 5
		}
		return concurrency
	}
	return 1
}

func comicGenerationFailureStage(err error) (string, string) {
	if err != nil && (strings.Contains(err.Error(), "漫画资产图生成失败") || strings.Contains(err.Error(), "漫画资产图未全部写回")) {
		return "comic_asset_images", "漫画资产参考图生成失败"
	}
	return "comic_images", "整本漫画出图失败"
}

type automationStateFile struct {
	Version int                       `json:"version"`
	Jobs    map[string]map[string]any `json:"jobs"`
}

func NewBatchAutomationCoordinator(store batchfactoryv11.Store, presets AutomationPresetStore, stages *batchfactoryv11.BookStageService, merge *batchfactoryv11.MergeService, publisher *external.Service, ttsStore novelfetchworkshop.Store, audioProbe batchfactoryv11.H3AudioDurationProbe, reader *FileAutomationStatusReader, now func() time.Time, providerSync ...func(context.Context, string) error) *BatchAutomationCoordinator {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	path := ""
	if reader != nil {
		path = reader.Path()
	}
	var syncProvider func(context.Context, string) error
	if len(providerSync) > 0 {
		syncProvider = providerSync[0]
	}
	return &BatchAutomationCoordinator{store: store, presets: presets, stages: stages, merge: merge, external: publisher, ttsStore: ttsStore, audioProbe: audioProbe, path: path, now: now, providerSync: syncProvider}
}

func (c *BatchAutomationCoordinator) SetAutomationStateStore(store AutomationStateStore) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.automationStore = store
	c.automationState = nil
}

const (
	automationScheduleExecute = "execute"
	automationScheduleSubmit  = "submit"
)

func normalizeAutomationScheduleMode(value string) string {
	if strings.TrimSpace(strings.ToLower(value)) == automationScheduleSubmit {
		return automationScheduleSubmit
	}
	return automationScheduleExecute
}

func automationJobDue(job map[string]any, now time.Time) bool {
	scheduledAt := strings.TrimSpace(automationStringValue(job["scheduledAt"], ""))
	if scheduledAt == "" {
		return true
	}
	due, err := time.Parse(time.RFC3339Nano, scheduledAt)
	if err != nil {
		due, err = time.Parse(time.RFC3339, scheduledAt)
	}
	return err != nil || !now.Before(due)
}

func (c *BatchAutomationCoordinator) StartScheduledDispatcher() {
	if c == nil {
		return
	}
	c.dispatcherOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for range ticker.C {
				_ = c.DispatchDueJobs(context.Background())
			}
		}()
	})
}

// DispatchDueJobs promotes persisted scheduled jobs to running and starts
// them. It is also callable by tests and by the resident dispatcher.
func (c *BatchAutomationCoordinator) DispatchDueJobs(ctx context.Context) error {
	if c == nil || c.store == nil {
		return nil
	}
	now := c.now()
	type dueJob struct {
		owner, batchID, jobID string
		job                   map[string]any
	}
	due := make([]dueJob, 0)
	c.mu.Lock()
	state, err := c.loadLocked()
	if err != nil {
		c.mu.Unlock()
		return err
	}
	if c.dispatchNextCheckSet && now.Before(c.dispatchNextCheck) {
		c.mu.Unlock()
		return nil
	}
	c.dispatchNextCheckSet = true
	c.dispatchNextCheck = now.Add(30 * time.Second)
	changed := false
	for jobID, job := range state.Jobs {
		if automationJobRetryDispatchDue(job, now) {
			books, _ := job["books"].(map[string]any)
			retryDue := false
			for _, raw := range books {
				entry, _ := raw.(map[string]any)
				if entry == nil || !automationBoolValue(entry["retryRequested"]) || !automationRetryDue(entry, now) {
					continue
				}
				if automationStringValue(entry["retryReason"], "") == "content_policy" {
					stage := automationStringValue(entry["stage"], "")
					if stage == "" {
						stage = "comic_images"
					}
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "running", stage, "到达重试时间，正在重新生成被内容审核拒绝的内容", now
					delete(entry, "retryReason")
				} else {
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "running", "assets", "超时后自动重试中", now
				}
				delete(entry, "retryAt")
				delete(entry, "retryRequested")
				retryDue = true
			}
			if retryDue {
				// Older runtimes could mark the whole job complete while a book
				// still had a delayed retry queued. Reopen that persisted job only
				// when the retry is due; do not restart it immediately on startup.
				job["state"] = "running"
				delete(job, "completedAt")
				job["updatedAt"] = now
				ready, failed, running := 0, 0, 0
				for _, raw := range books {
					entry, _ := raw.(map[string]any)
					if entry == nil {
						continue
					}
					switch automationStringValue(entry["status"], "") {
					case "completed":
						ready++
					case "failed", "blocked", "stopped":
						failed++
					case "running":
						running++
					}
				}
				job["counts"] = automationCounts(len(books), running, failed, ready)
				due = append(due, dueJob{owner: automationStringValue(job["owner"], ""), batchID: automationStringValue(job["batchId"], ""), jobID: jobID, job: job})
				changed = true
				continue
			}
		}
		if automationStringValue(job["state"], "") != "scheduled" || !automationJobDue(job, now) {
			continue
		}
		job["state"], job["startedAt"], job["updatedAt"] = "running", now, now
		if books, ok := job["books"].(map[string]any); ok {
			for _, raw := range books {
				entry, _ := raw.(map[string]any)
				if entry == nil || automationStringValue(entry["status"], "") != "scheduled" || automationStringValue(entry["stage"], "") == "waiting_submit" {
					continue
				}
				entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "running", "scheduled", "已到设定时间，开始继续处理", now
			}
		}
		due = append(due, dueJob{owner: automationStringValue(job["owner"], ""), batchID: automationStringValue(job["batchId"], ""), jobID: jobID, job: job})
		changed = true
	}
	if changed {
		err = c.saveLocked(state)
	}
	if next := automationNextDispatchAt(state, now); !next.IsZero() {
		c.dispatchNextCheck = next
	} else {
		c.dispatchNextCheck = now.Add(30 * time.Second)
	}
	c.mu.Unlock()
	if err != nil {
		return err
	}
	for _, item := range due {
		batch, getErr := c.store.GetBatch(ctx, item.owner, item.batchID)
		if getErr != nil {
			item.job["state"], item.job["lastError"], item.job["updatedAt"] = "failed", getErr.Error(), c.now()
			_ = c.setJob(item.owner, item.batchID, item.jobID, item.job)
			continue
		}
		scoped := batch
		tracked := map[string]bool{}
		if books, ok := item.job["books"].(map[string]any); ok {
			for id := range books {
				tracked[id] = true
			}
		}
		scoped.Books = make([]batchfactoryv11.Book, 0, len(tracked))
		for _, book := range batch.Books {
			if tracked[book.ID] {
				scoped.Books = append(scoped.Books, book)
			}
		}
		if validationErr := c.validateAutomationGlobalPreflight(ctx, item.owner, batch, automationStringValue(item.job["runMode"], "full_submit")); validationErr != nil {
			item.job["state"], item.job["lastError"], item.job["updatedAt"] = "failed", validationErr.Error(), c.now()
			_ = c.setJob(item.owner, item.batchID, item.jobID, item.job)
			continue
		}
		c.launchRun(item.owner, item.batchID, item.jobID, item.job, scoped)
	}
	return nil
}

func automationNextDispatchAt(state automationStateFile, now time.Time) time.Time {
	var next time.Time
	for _, job := range state.Jobs {
		jobState := automationStringValue(job["state"], "")
		if jobState == "scheduled" {
			at := automationStoredJobTime(job["scheduledAt"])
			if at.IsZero() || !at.After(now) {
				return now
			}
			if next.IsZero() || at.Before(next) {
				next = at
			}
		}
		if jobState != "running" && jobState != "completed" && jobState != "failed" {
			continue
		}
		books, _ := job["books"].(map[string]any)
		for _, raw := range books {
			entry, _ := raw.(map[string]any)
			if entry == nil || !automationBoolValue(entry["retryRequested"]) {
				continue
			}
			at := automationStoredJobTime(entry["retryAt"])
			if at.IsZero() || !at.After(now) {
				return now
			}
			if next.IsZero() || at.Before(next) {
				next = at
			}
		}
	}
	return next
}

func automationRetryDue(entry map[string]any, now time.Time) bool {
	value := strings.TrimSpace(automationStringValue(entry["retryAt"], ""))
	if value == "" {
		return false
	}
	when, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		when, err = time.Parse(time.RFC3339, value)
	}
	return err == nil && !now.Before(when)
}

func automationHasDueRetry(job map[string]any, now time.Time) bool {
	books, _ := job["books"].(map[string]any)
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry != nil && automationBoolValue(entry["retryRequested"]) && automationRetryDue(entry, now) {
			return true
		}
	}
	return false
}

func automationJobHasRetryRequest(job map[string]any) bool {
	books, _ := job["books"].(map[string]any)
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry != nil && automationBoolValue(entry["retryRequested"]) {
			return true
		}
	}
	return false
}

func automationJobRetryDispatchDue(job map[string]any, now time.Time) bool {
	if !automationHasDueRetry(job, now) {
		return false
	}
	switch automationStringValue(job["state"], "") {
	case "running", "completed", "failed":
		return true
	default:
		return false
	}
}

func shouldScheduleAutomationTimeoutRetry(stage string) bool {
	switch strings.TrimSpace(stage) {
	case "comic_images", "comic_asset_images":
		// These stages already own a bounded whole-book retry budget. Reopening
		// the entire book after that budget expires creates a second retry loop.
		return false
	default:
		return true
	}
}

// automationBatchRetryRounds 是批次末尾统一回补失败书籍的最大轮次。每一轮都在
// 整批其余书结束后串行回补，避免失败书一直占着并发槽位。
const automationBatchRetryRounds = 2

// 视频供应商实测单批排队 + 生成会超过 10 分钟（本次线上最后一个分镜用了 11 分
// 09 秒），10 分钟死线会把正常视频误判成失败并触发整书回补。视频等待单独放宽
// 到 30 分钟，并且只要任务状态还在推进就续期（续期以 10 分钟空闲为界）。
const (
	videoWaitTimeout     = 30 * time.Minute
	videoWaitIdleTimeout = 10 * time.Minute
)

// errProductionInFlight 表示视频等待到达上限时，供应商侧任务仍在执行。
// 这种情形不能按失败处理：整书回补会强制重跑前面的阶段并重建视频记录，
// 把已经生成好、仍可播放的视频变成 orphaned。
var errProductionInFlight = errors.New("视频任务仍在生成")

// automationRetryResumeStage 把失败阶段映射成回补起点。回补只从失败阶段继续，
// 不能从 assets 从头强制重跑。
func automationRetryResumeStage(stage string) string {
	switch stage {
	case "assets", "director", "opening", "video":
		return stage
	case "merge", "submit", "publish", "upload_confirm":
		// 合成 / 提交阶段失败说明视频已经完成，回补只需回到 video 重新回读。
		return "video"
	default:
		return "assets"
	}
}

// deterministicAutomationFailure 识别同一请求重试也不会成功的上游错误：
// 401/402/403/400/404/422、余额不足/欠费文案、参数或配置错误。超时、429
// 和 5xx 仍是瞬时错误，必须继续保留回补能力。
func deterministicAutomationFailure(err error) bool {
	if err == nil || errors.Is(err, errProductionInFlight) || errors.Is(err, batchfactoryv11.ErrConflict) {
		return false
	}
	message := strings.ToLower(err.Error())
	for _, token := range []string{
		"余额不足", "insufficient balance", "balance pre-deduction insufficient",
		"payment required", "http 401", "http 402", "http 403",
	} {
		if strings.Contains(message, token) {
			return true
		}
	}
	// retryableComicAutomationError 已把 400/401/402/403/404/422、ErrInvalid、
	// ErrNotFound、ErrUnavailable 判为永久错误，把 408/429/5xx 判为可重试。
	return !retryableComicAutomationError(err)
}

// markDeterministicAutomationFailure 把确定性错误立即落为终态失败，并清掉
// 回补标记。返回 true 时调用方必须跳过 scheduleAutomationBatchEndRetry /
// enqueueBatchRetry，避免整书重跑再次提交付费任务。
func markDeterministicAutomationFailure(entry map[string]any, stage string, err error, message string, now time.Time) bool {
	if entry == nil || !deterministicAutomationFailure(err) {
		return false
	}
	entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", stage, err.Error(), message+"：确定性失败，跳过回补", now
	delete(entry, "batchRetryPending")
	delete(entry, "retryAt")
	delete(entry, "retryRequested")
	appendAutomationBookLog(entry, now, stage, "确定性失败，跳过回补")
	return true
}

// markAutomationBookQueued 把等待并发槽位的书切换为真实排队态。回补书只改
// status/message，保留失败阶段与 batchRetryCount，供拿到槽位后 resumeStage
// 从失败阶段续跑。
func markAutomationBookQueued(entry map[string]any, now time.Time) bool {
	if entry == nil {
		return false
	}
	if automationStringValue(entry["status"], "") == "running" {
		entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "queued", "queued", "排队中，等待并发槽位", now
		return true
	}
	if automationBoolValue(entry["batchRetryPending"]) {
		entry["status"], entry["message"], entry["updatedAt"] = "queued", "排队中，等待并发槽位后回补", now
		return true
	}
	return false
}

// markAutomationBookStarted 在真正拿到并发槽位后把书切回运行态。回补书走
// prepareAutomationBookRetry，保证从失败阶段续跑而不是从头重做。
func markAutomationBookStarted(entry map[string]any, now time.Time) bool {
	if entry == nil {
		return false
	}
	if automationBoolValue(entry["batchRetryPending"]) {
		return prepareAutomationBookRetry(entry, automationNumberValue(entry["batchRetryCount"], 0), now)
	}
	if automationStringValue(entry["status"], "") == "queued" {
		entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "running", "assets", "并发槽位已获得，开始执行", now
		return true
	}
	return false
}

// scheduleAutomationBatchEndRetry 把本书登记为“已跳过，等待批次末尾回补”。
// 返回 false 表示回补预算已用完，调用方应把本书置为终态失败。
func scheduleAutomationBatchEndRetry(entry map[string]any, stage string, err error, message string, round int, now time.Time) bool {
	if entry == nil || round >= automationBatchRetryRounds {
		return false
	}
	// 任务还在供应商侧执行时既不算失败，也不消耗回补预算。
	if errors.Is(err, errProductionInFlight) {
		return false
	}
	attempt := automationNumberValue(entry["batchRetryCount"], 0)
	if attempt >= automationBatchRetryRounds {
		return false
	}
	errorText := ""
	if err != nil {
		errorText = err.Error()
	}
	entry["status"], entry["stage"], entry["error"], entry["updatedAt"] = "failed", stage, errorText, now
	entry["batchRetryCount"] = attempt + 1
	entry["batchRetryPending"] = true
	entry["message"] = fmt.Sprintf("%s，已进入回补队列，并发槽位一空出就立即回补（%d/%d）", message, attempt+1, automationBatchRetryRounds)
	return true
}

// prepareAutomationBookRetry 把一本等待回补的书切回运行态。它既可以由
// 批末统一切换使用，也可以在某本书失败后立即重新入队时单本调用。
func prepareAutomationBookRetry(entry map[string]any, round int, now time.Time) bool {
	if entry == nil || !automationBoolValue(entry["batchRetryPending"]) {
		return false
	}
	// 回补从失败阶段续跑，而不是从 assets 强制重跑；从头重跑会重建视频记录，
	// 把已经生成好的视频轮换成 orphaned。
	stage := automationRetryResumeStage(automationStringValue(entry["stage"], ""))
	message := fmt.Sprintf("开始回补失败书籍（第 %d/%d 轮）", round, automationBatchRetryRounds)
	if automationStringValue(entry["stage"], "") == "comic_publish" && strings.TrimSpace(automationStringValue(entry["comicCollectionId"], "")) != "" {
		stage = "comic_publish"
		message = fmt.Sprintf("重新发布已生成漫画合集（第 %d/%d 轮）", round, automationBatchRetryRounds)
	}
	entry["resumeStage"] = stage
	entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", stage, message, "", now
	delete(entry, "batchRetryPending")
	appendAutomationBookLog(entry, now, stage, message)
	return true
}

// prepareAutomationBatchRetryRound 把上一轮被跳过、等待回补的书重新放回运行态，
// 返回本轮回补的书 ID 集合。返回空集合表示没有待回补的书，批次可以收尾。
func prepareAutomationBatchRetryRound(books map[string]any, round int, now time.Time) map[string]bool {
	retry := make(map[string]bool)
	for id, raw := range books {
		entry, _ := raw.(map[string]any)
		if prepareAutomationBookRetry(entry, round, now) {
			retry[id] = true
		}
	}
	return retry
}

func (c *BatchAutomationCoordinator) SetLocalFiles(files *localartifact.Store) {
	if c != nil {
		c.localFiles = files
	}
}

func (c *BatchAutomationCoordinator) SetSourceFetcher(fetcher *novelfetchworkshop.UpstreamFetcher) {
	if c != nil {
		c.fetcher = fetcher
	}
}

// SetSourceClassification keeps scheduled automation on the same metadata-first
// classification path as the manual fetch endpoint.
func (c *BatchAutomationCoordinator) SetSourceClassification(classification batchFetchClassification) {
	if c != nil {
		c.classification = classification
	}
}

func (c *BatchAutomationCoordinator) SetComicBookGenerator(generator func(context.Context, string, string, string) (batchfactoryv11.ComicCollection, error)) {
	if c != nil {
		c.comicGenerate = generator
	}
}

func (c *BatchAutomationCoordinator) SetComicPublisher(publisher func(context.Context, string, string, string, batchfactoryv11.ComicCollection, batchfactoryv11.SettingsPatch) (N8ComicPublishRecord, error)) {
	if c != nil {
		c.comicPublish = publisher
	}
}

func (c *BatchAutomationCoordinator) SetComicPublishPreflight(preflight func(context.Context, batchfactoryv11.SettingsPatch) error) {
	if c != nil {
		c.comicPublishPreflight = preflight
	}
}

func (c *BatchAutomationCoordinator) loadLocked() (automationStateFile, error) {
	if c.automationStore != nil && c.automationState != nil {
		return *c.automationState, nil
	}
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{}}
	if c.automationStore != nil {
		stored, found, err := c.automationStore.Load(context.Background())
		if err != nil {
			return state, err
		}
		if found {
			if stored.Jobs == nil {
				stored.Jobs = map[string]map[string]any{}
			}
			if normalizeAlreadyUploadedAutomationState(&stored) {
				if err := c.automationStore.Save(context.Background(), stored); err != nil {
					return state, err
				}
			}
			c.automationState = &stored
			return stored, nil
		}
	}
	if strings.TrimSpace(c.path) == "" {
		return state, nil
	}
	// 优先读增量布局（automation/index.json + automation/jobs/<jobId>.json）。
	if split, ok, err := c.loadSplitLocked(); err != nil {
		return state, err
	} else if ok {
		if c.automationStore != nil {
			if err := c.automationStore.Save(context.Background(), split); err != nil {
				return state, err
			}
		}
		if c.automationStore != nil {
			c.automationState = &split
		}
		return split, nil
	}
	// 增量布局不存在时尝试旧单文件并迁移；迁移失败不阻塞启动，沿用旧数据。
	raw, err := os.ReadFile(c.path)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, err
	}
	if state.Jobs == nil {
		state.Jobs = map[string]map[string]any{}
	}
	if normalizeAlreadyUploadedAutomationState(&state) {
		if c.automationStore == nil {
			if err := c.saveLocked(state); err != nil {
				return state, err
			}
		}
	}
	if err := c.migrateLegacyLocked(state); err != nil {
		// 迁移失败继续用旧布局，下次启动重试。
		return state, nil
	}
	return state, nil
}

// automationStorePaths 返回增量布局的索引文件与任务目录。
func (c *BatchAutomationCoordinator) automationStorePaths() (indexPath, jobsDir string) {
	dir := filepath.Join(filepath.Dir(c.path), "automation")
	return filepath.Join(dir, "index.json"), filepath.Join(dir, "jobs")
}

func writeAutomationJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (c *BatchAutomationCoordinator) loadSplitLocked() (automationStateFile, bool, error) {
	indexPath, jobsDir := c.automationStorePaths()
	raw, err := os.ReadFile(indexPath)
	if errors.Is(err, os.ErrNotExist) {
		return automationStateFile{}, false, nil
	}
	if err != nil {
		return automationStateFile{}, false, err
	}
	var index struct {
		Jobs []map[string]any `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return automationStateFile{}, false, err
	}
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{}}
	for _, entry := range index.Jobs {
		jobID := automationStringValue(entry["id"], "")
		if jobID == "" {
			continue
		}
		jobRaw, err := os.ReadFile(filepath.Join(jobsDir, jobID+".json"))
		if err != nil {
			continue // 单个任务文件缺失不拖垮整体加载
		}
		var job map[string]any
		if err := json.Unmarshal(jobRaw, &job); err != nil {
			continue // 单个任务文件损坏不拖垮整体加载
		}
		state.Jobs[jobID] = job
	}
	return state, true, nil
}

func (c *BatchAutomationCoordinator) saveLocked(state automationStateFile) error {
	if state.Jobs == nil {
		state.Jobs = map[string]map[string]any{}
	}
	if c.automationStore != nil {
		if err := c.automationStore.Save(context.Background(), state); err != nil {
			return err
		}
	}
	if c.automationStore != nil {
		c.automationState = &state
	}
	if strings.TrimSpace(c.path) == "" {
		return nil
	}
	indexPath, jobsDir := c.automationStorePaths()
	entries := make([]map[string]any, 0, len(state.Jobs))
	for jobID, job := range state.Jobs {
		if err := writeAutomationJSONAtomic(filepath.Join(jobsDir, jobID+".json"), job); err != nil {
			return err
		}
		entries = append(entries, map[string]any{
			"id": jobID, "owner": job["owner"], "batchId": job["batchId"],
			"state": job["state"], "updatedAt": job["updatedAt"],
		})
	}
	return writeAutomationJSONAtomic(indexPath, map[string]any{"version": 1, "jobs": entries})
}

// migrateLegacyLocked 把旧单文件拆分为增量布局，成功后旧文件改名 .migrated。
func (c *BatchAutomationCoordinator) migrateLegacyLocked(state automationStateFile) error {
	if err := c.saveLocked(state); err != nil {
		return err
	}
	return os.Rename(c.path, c.path+".migrated")
}

func (c *BatchAutomationCoordinator) setJob(owner, batchID, jobID string, job map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.automationStore != nil {
		state, err := c.loadLocked()
		if err != nil {
			return err
		}
		if job == nil {
			job = map[string]any{}
		}
		if automationStringValue(job["id"], "") == "" {
			job["id"] = jobID
		}
		if automationStringValue(job["owner"], "") == "" {
			job["owner"] = owner
		}
		if automationStringValue(job["batchId"], "") == "" {
			job["batchId"] = batchID
		}
		state.Jobs[jobID] = job
		if automationStringValue(job["state"], "") == "scheduled" || automationJobHasRetryRequest(job) {
			c.dispatchNextCheckSet = false
		}
		return c.saveLocked(state)
	}
	if strings.TrimSpace(c.path) == "" {
		return nil
	}
	indexPath, jobsDir := c.automationStorePaths()
	if err := writeAutomationJSONAtomic(filepath.Join(jobsDir, jobID+".json"), job); err != nil {
		return err
	}
	if automationStringValue(job["state"], "") == "scheduled" || automationJobHasRetryRequest(job) {
		c.dispatchNextCheckSet = false
	}
	if c.automationStore != nil && c.automationState != nil {
		if c.automationState.Jobs == nil {
			c.automationState.Jobs = map[string]map[string]any{}
		}
		c.automationState.Jobs[jobID] = job
	}
	// 索引只在任务存在性或 state 变化时重写，避免每次心跳全量写索引。
	entries := []map[string]any{}
	if raw, err := os.ReadFile(indexPath); err == nil {
		var index struct {
			Jobs []map[string]any `json:"jobs"`
		}
		if json.Unmarshal(raw, &index) == nil {
			entries = index.Jobs
		}
	}
	found := false
	stateChanged := false
	for i, entry := range entries {
		if automationStringValue(entry["id"], "") != jobID {
			continue
		}
		found = true
		if automationStringValue(entry["state"], "") != automationStringValue(job["state"], "") {
			stateChanged = true
		}
		entries[i] = map[string]any{"id": jobID, "owner": job["owner"], "batchId": job["batchId"], "state": job["state"], "updatedAt": job["updatedAt"]}
	}
	if !found {
		entries = append(entries, map[string]any{"id": jobID, "owner": job["owner"], "batchId": job["batchId"], "state": job["state"], "updatedAt": job["updatedAt"]})
	}
	if found && !stateChanged {
		return nil
	}
	return writeAutomationJSONAtomic(indexPath, map[string]any{"version": 1, "jobs": entries})
}

// RecoverInterruptedJobs converts persisted runs without a live goroutine into
// a terminal state so a restarted service cannot block a new immediate run.
func (c *BatchAutomationCoordinator) RecoverInterruptedJobs() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if c.now != nil {
		now = c.now()
	}
	changed := false
	for _, job := range state.Jobs {
		if automationStringValue(job["state"], "") != "running" {
			continue
		}
		job["state"], job["lastError"], job["updatedAt"] = "stopped", automationServiceRestartMessage, now
		job["stopAt"] = now
		books, _ := job["books"].(map[string]any)
		stopped, failed, ready := 0, 0, 0
		for _, raw := range books {
			entry, _ := raw.(map[string]any)
			if entry == nil {
				continue
			}
			switch automationStringValue(entry["status"], "") {
			case "running":
				entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "stopped", "stopped", automationServiceRestartMessage, now
				stopped++
			case "completed":
				ready++
			case "failed", "blocked", "stopped":
				failed++
			}
		}
		job["counts"] = automationCounts(len(books), 0, failed+stopped, ready)
		changed = true
	}
	if !changed {
		return nil
	}
	return c.saveLocked(state)
}

const automationServiceRestartMessage = "服务重启，自动化已中断，请重新立即执行。"

// resumeInterruptedAutomationJob reopens only a run that the server itself
// stopped during startup recovery. A user-stopped job has no matching
// lastError and remains stopped until the user explicitly retries it.
//
// 它只改任务级状态，不碰书级 status/stage：书是否恢复必须先过逐书预检
// （见 applyResumePreflight）。已完成阶段（comicCollectionId / comicImageIDs /
// promptReady 等）原样保留，只清中断错误。
func resumeInterruptedAutomationJob(job map[string]any, now time.Time) bool {
	if automationStringValue(job["state"], "") != "stopped" || automationStringValue(job["lastError"], "") != automationServiceRestartMessage {
		return false
	}
	job["state"], job["lastError"], job["updatedAt"] = "running", "", now
	delete(job, "stopAt")
	job["restartedAt"] = now
	return true
}

// applyResumePreflight 在恢复任务时执行逐书预检：不满足条件的书写 blocked
// 并附原因，满足的才重新置为 running。stage 保持原值，不重置为 assets。
func (c *BatchAutomationCoordinator) applyResumePreflight(ctx context.Context, owner string, batch batchfactoryv11.Batch, job map[string]any, now time.Time) {
	books, _ := job["books"].(map[string]any)
	bookByID := make(map[string]batchfactoryv11.Book, len(batch.Books))
	for _, book := range batch.Books {
		bookByID[book.ID] = book
	}
	runMode := automationStringValue(job["runMode"], "full_submit")
	running, failed, ready := 0, 0, 0
	for bookID, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		if automationStringValue(entry["status"], "") == "completed" {
			ready++
			continue
		}
		book, ok := bookByID[bookID]
		if !ok {
			failed++
			continue
		}
		if reason := c.automationBookPreflightReason(ctx, owner, batch, book, runMode); reason != "" {
			// 预检失败只改终态和原因，stage 保持原值，避免把已完成阶段清掉。
			entry["status"], entry["error"], entry["message"], entry["updatedAt"] = "blocked", "", reason, now
			entry["retryRequested"] = false
			appendAutomationBookLog(entry, now, "preflight", reason)
			failed++
			continue
		}
		// stage 保持恢复前的值：已经完成的漫画阶段不会被重新跑一遍。
		entry["status"], entry["message"], entry["error"], entry["updatedAt"] = "running", "服务已恢复，自动化继续执行", "", now
		entry["retryRequested"] = true
		running++
	}
	job["counts"] = automationCounts(len(books), running, failed, ready)
}

// automationResumeOnBootEnabled 控制服务启动时是否自动恢复被重启打断的自动化
// 任务。默认关闭：重启后不自动把书重新置为 running。
func automationResumeOnBootEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("QIANTIE_AUTOMATION_RESUME_ON_BOOT"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// resumeAutomationOnBoot 是启动期唯一入口：未显式开启时只保留
// RecoverInterruptedJobs 的标记行为，不自动 resume。
func resumeAutomationOnBoot(coordinator *BatchAutomationCoordinator) {
	if coordinator == nil {
		return
	}
	if !automationResumeOnBootEnabled() {
		log.Printf("skip resuming interrupted batch automation jobs: QIANTIE_AUTOMATION_RESUME_ON_BOOT is not enabled")
		return
	}
	if err := coordinator.ResumeInterruptedJobs(context.Background()); err != nil {
		log.Printf("resume interrupted batch automation jobs: %v", err)
		return
	}
	log.Printf("[DIAG] ResumeInterruptedJobs OK (no error)")
}

// ResumeInterruptedJobs restarts jobs that were interrupted only because the
// local Go service was restarted. Each run is launched against the books that
// were persisted in that job; comic stages are idempotent and reuse completed
// prompts and primary images.
func (c *BatchAutomationCoordinator) ResumeInterruptedJobs(ctx context.Context) error {
	if c == nil || c.store == nil {
		return nil
	}
	type candidate struct {
		owner, batchID, jobID string
		job                   map[string]any
	}
	resumable := make([]candidate, 0)
	c.mu.Lock()
	state, err := c.loadLocked()
	if err != nil {
		c.mu.Unlock()
		return err
	}
	changed := false
	now := c.now()
	for jobID, job := range state.Jobs {
		if !resumeInterruptedAutomationJob(job, now) {
			continue
		}
		resumable = append(resumable, candidate{owner: automationStringValue(job["owner"], ""), batchID: automationStringValue(job["batchId"], ""), jobID: jobID, job: job})
		changed = true
	}
	if changed {
		err = c.saveLocked(state)
	}
	c.mu.Unlock()
	if err != nil {
		return err
	}
	for _, item := range resumable {
		batch, getErr := c.store.GetBatch(ctx, item.owner, item.batchID)
		if getErr != nil {
			item.job["state"], item.job["lastError"], item.job["updatedAt"] = "failed", getErr.Error(), c.now()
			_ = c.setJob(item.owner, item.batchID, item.jobID, item.job)
			continue
		}
		// 自动恢复与正常派发共用同一套全局预检，避免重启后在没有执行器的
		// 情况下把整批书重新置为 running。
		if validationErr := c.validateAutomationGlobalPreflight(ctx, item.owner, batch, automationStringValue(item.job["runMode"], "full_submit")); validationErr != nil {
			item.job["state"], item.job["lastError"], item.job["updatedAt"] = "failed", validationErr.Error(), c.now()
			_ = c.setJob(item.owner, item.batchID, item.jobID, item.job)
			continue
		}
		c.applyResumePreflight(ctx, item.owner, batch, item.job, c.now())
		if saveErr := c.setJob(item.owner, item.batchID, item.jobID, item.job); saveErr != nil {
			log.Printf("save resumed batch automation job %s: %v", item.jobID, saveErr)
		}
		tracked := map[string]bool{}
		if books, ok := item.job["books"].(map[string]any); ok {
			for id := range books {
				tracked[id] = true
			}
		}
		scoped := batch
		scoped.Books = make([]batchfactoryv11.Book, 0, len(tracked))
		for _, book := range batch.Books {
			if tracked[book.ID] {
				scoped.Books = append(scoped.Books, book)
			}
		}
		c.launchRun(item.owner, item.batchID, item.jobID, item.job, scoped)
	}
	return nil
}

func automationCounts(total, running, failed, ready int) map[string]any {
	return map[string]any{"total": total, "ready": ready, "running": running, "queued": 0, "pending": 0, "failed": failed, "blocked": 0}
}

// automationCountsFromBooks 按持久化的书籍条目实时统计汇总数字。运行中的
// 任务不能用启动时的快照，否则一本失败后仍会显示 failed:0。
func automationCountsFromBooks(books map[string]any) map[string]any {
	counts := map[string]any{"total": len(books), "ready": 0, "running": 0, "queued": 0, "pending": 0, "failed": 0, "blocked": 0}
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "completed":
			counts["ready"] = counts["ready"].(int) + 1
		case "running":
			counts["running"] = counts["running"].(int) + 1
		case "queued":
			counts["queued"] = counts["queued"].(int) + 1
		case "scheduled", "pending_confirmation":
			counts["pending"] = counts["pending"].(int) + 1
		case "blocked":
			counts["blocked"] = counts["blocked"].(int) + 1
		case "failed", "stopped":
			counts["failed"] = counts["failed"].(int) + 1
		}
	}
	return counts
}

// selectedAutomationBooks 让后续“添加书籍”只创建新增书的任务。若重新启动
// 整个批次，会再次提交旧书，造成不同小说的生产记录混在一起。
func selectedAutomationBooks(batch batchfactoryv11.Batch, value any) ([]batchfactoryv11.Book, error) {
	requested := map[string]bool{}
	switch values := value.(type) {
	case nil:
	case []string:
		for _, id := range values {
			if id = strings.TrimSpace(id); id != "" {
				requested[id] = true
			}
		}
	case []any:
		for _, value := range values {
			if id := strings.TrimSpace(automationStringValue(value, "")); id != "" {
				requested[id] = true
			}
		}
	default:
		return nil, fmt.Errorf("%w: bookIds 格式无效", batchfactoryv11.ErrInvalid)
	}
	if len(requested) == 0 {
		return append([]batchfactoryv11.Book(nil), batch.Books...), nil
	}
	books := make([]batchfactoryv11.Book, 0, len(requested))
	for _, book := range batch.Books {
		if requested[book.ID] {
			books = append(books, book)
			delete(requested, book.ID)
		}
	}
	if len(requested) > 0 {
		return nil, fmt.Errorf("%w: 存在不属于当前批次的书籍", batchfactoryv11.ErrInvalid)
	}
	return books, nil
}

func (c *BatchAutomationCoordinator) Start(ctx context.Context, owner, batchID string, input map[string]any) (map[string]any, error) {
	if c == nil || c.store == nil || c.stages == nil {
		return nil, fmt.Errorf("%w: Go 自动化调度器未配置", batchfactoryv11.ErrUnavailable)
	}
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" {
		return nil, batchfactoryv11.ErrInvalid
	}
	// Serialize starts for one process. The persisted-state check below stops a
	// second request after the first one has created its job; this mutex closes
	// the small race between two simultaneous HTTP requests.
	c.startMu.Lock()
	defer c.startMu.Unlock()
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	targetBooks, err := selectedAutomationBooks(batch, input["bookIds"])
	if err != nil {
		return nil, err
	}
	if len(targetBooks) == 0 {
		return nil, fmt.Errorf("%w: 没有可自动生产的书籍", batchfactoryv11.ErrInvalid)
	}
	if active := c.activeAutomationJob(owner, batchID); active != nil {
		return nil, fmt.Errorf("%w: 当前批次已有自动生产任务（%s），请先暂停、停止或重试原任务", batchfactoryv11.ErrConflict, automationStringValue(active["id"], "未命名任务"))
	}
	presetID := automationStringValue(input["presetId"], "")
	if presetID == "" {
		return nil, fmt.Errorf("%w: 请先选择自动化预设", batchfactoryv11.ErrInvalid)
	}
	if c.presets == nil {
		return nil, fmt.Errorf("%w: 自动化预设服务未配置", batchfactoryv11.ErrUnavailable)
	}
	preset, err := c.presets.Get(owner, presetID)
	if err != nil || preset == nil {
		return nil, fmt.Errorf("自动化预设不存在或不属于当前账号")
	}
	batch, err = c.applyPreset(ctx, owner, batch, preset, automationStringValue(input["productionMode"], ""))
	if err != nil {
		return nil, err
	}
	// A private-customization run may carry one pasted task's media settings
	// (style, reference image and comic constraints) in addition to the saved
	// automation preset. Apply that small overlay after the preset so starting
	// automation cannot silently replace the choices made in the private form.
	if rawOverlay, ok := input["settingsPatch"].(map[string]any); ok && len(rawOverlay) > 0 {
		overlay := batchfactoryv11.SettingsPatch{}
		for key, value := range rawOverlay {
			if key == "aiPromptConfig" && c.promptCatalog != nil {
				rawValue, marshalErr := json.Marshal(value)
				if marshalErr != nil {
					return nil, marshalErr
				}
				hydrated, hydrateErr := hydrateBatchFactoryPromptConfig(c.promptCatalog, rawValue)
				if hydrateErr != nil {
					return nil, hydrateErr
				}
				overlay[key] = hydrated
				continue
			}
			raw, marshalErr := json.Marshal(value)
			if marshalErr != nil {
				return nil, marshalErr
			}
			overlay[key] = raw
		}
		if _, err := c.store.SaveSettings(ctx, owner, batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: overlay, ExpectedRevision: batch.Revision}); err != nil {
			return nil, err
		}
		batch, err = c.store.GetBatch(ctx, owner, batch.ID)
		if err != nil {
			return nil, err
		}
	}
	// Automatic production enters the same provider-sync boundary as the
	// explicit production endpoint. Without this, the UI can show a configured
	// video model while the Go worker's in-memory registry is still empty.
	if c.providerSync != nil {
		if err := c.providerSync(ctx, owner); err != nil {
			return nil, err
		}
	}
	jobID := fmt.Sprintf("go-auto-%d", c.now().UnixNano())
	runMode := automationStringValue(input["runMode"], "full_submit")
	autoPublish := automationBoolValue(input["autoPublish"])
	forceRegenerate := automationBoolValue(input["forceRegenerate"])
	scheduleMode := normalizeAutomationScheduleMode(automationStringValue(input["scheduleMode"], automationScheduleExecute))
	scheduledAt := strings.TrimSpace(automationStringValue(input["scheduledAt"], ""))
	pending := scheduledAt != "" && !automationJobDue(map[string]any{"scheduledAt": scheduledAt}, c.now())
	initialState := "running"
	if scheduleMode == automationScheduleExecute && pending {
		initialState = "scheduled"
	}
	// full_submit is the user-facing “生成并提交” mode. Keep the server
	// authoritative so older frontends that omitted autoPublish cannot silently
	// finish without the upload step.
	if runMode == "full_submit" {
		autoPublish = true
	}
	job := map[string]any{
		"id": jobID, "owner": owner, "batchId": batchID, "state": initialState,
		"runMode": runMode, "autoPublish": autoPublish, "forceRegenerate": forceRegenerate,
		"scheduleMode": scheduleMode, "scheduledAt": scheduledAt,
		"concurrency": automationNumberValue(input["concurrency"], 1), "preset": map[string]any{"id": preset.ID, "name": preset.Name, "version": preset.Version},
		"createdAt": c.now(), "updatedAt": c.now(),
		"counts": automationCounts(len(targetBooks), len(targetBooks), 0, 0), "books": map[string]any{},
	}
	if initialState == "running" {
		job["startedAt"] = c.now()
		if err := c.validateAutomationGlobalPreflight(ctx, owner, batch, runMode); err != nil {
			return nil, err
		}
	}
	books := job["books"].(map[string]any)
	for _, book := range targetBooks {
		// 还没有拿到并发槽位的书必须如实写 queued；前端只在 status=queued
		// 时显示“排队中”。真正开始执行时由队列循环迁移到 running/assets。
		status, stage, message := "queued", "queued", "排队中，等待并发槽位"
		if initialState == "scheduled" {
			status, stage, message = "scheduled", "scheduled", "等待北京时间到点后开始生产"
		}
		books[book.ID] = map[string]any{"bookId": book.ID, "title": book.Title, "status": status, "stage": stage, "message": message, "attempts": 0, "retryCount": 0, "updatedAt": c.now()}
	}
	if err := c.setJob(owner, batchID, jobID, job); err != nil {
		return nil, err
	}
	// The worker runs after the HTTP request has returned, so preserve the
	// authenticated owner in its context. Account-scoped text/video providers
	// must not fall back to an anonymous, unconfigured context in background
	// stages.
	// 任务只记录本次指定的书，阶段执行仍会从同一个已保存批次读取正文和配置。
	batch.Books = targetBooks
	if initialState == "running" {
		c.launchRun(owner, batchID, jobID, job, batch)
	}
	return publicAutomationJob(job), nil
}

func (c *BatchAutomationCoordinator) activeAutomationJob(owner, batchID string) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return nil
	}
	for _, job := range state.Jobs {
		if automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
			continue
		}
		switch automationStringValue(job["state"], "") {
		case "running", "paused", "scheduled", "needs_attention":
			return job
		}
	}
	return nil
}

// validateAutomationGlobalPreflight 只保留“全批不可执行”的条件：执行器整体
// 缺失或不可用。任意单书的配置缺陷不再整批拒绝，改由 automationBookPreflightReason
// 在 run() 里逐书跳过。
func (c *BatchAutomationCoordinator) validateAutomationGlobalPreflight(ctx context.Context, owner string, batch batchfactoryv11.Batch, runMode string) error {
	reasons := make([]string, 0)
	if c.stages == nil || c.stages.Director == nil || c.stages.Director.Provider == nil {
		reasons = append(reasons, "文本模型执行器未配置")
	}
	productionMode := comicSettingString(batch.SettingsState.Patch, "productionMode", "video")
	if productionMode == "comic" && runMode != "storyboard_only" {
		if c.stages == nil || c.stages.ImageReady == nil {
			reasons = append(reasons, "图片生成执行器未配置")
		} else if ready, reason := c.stages.ImageReady(ctx, owner); !ready {
			reasons = append(reasons, reasonOrDefault(reason, "图片生成执行器不可用"))
		}
	}
	if productionMode != "comic" && runMode != "storyboard_only" {
		if c.stages == nil || c.stages.VideoReady == nil {
			reasons = append(reasons, "视频生成执行器未配置")
		} else if ready, reason := c.stages.VideoReady(ctx, owner); !ready {
			reasons = append(reasons, reasonOrDefault(reason, "视频生成执行器不可用"))
		}
	}
	if runMode == "full_submit" {
		if productionMode == "comic" && c.comicPublish == nil {
			reasons = append(reasons, "漫画发布执行器未配置")
		}
		if productionMode != "comic" && c.publish121 == nil && c.external == nil {
			reasons = append(reasons, "视频发布执行器未配置")
		}
	}
	if len(reasons) == 0 {
		return nil
	}
	unique := make([]string, 0, len(reasons))
	seen := map[string]bool{}
	for _, reason := range reasons {
		if reason != "" && !seen[reason] {
			seen[reason] = true
			unique = append(unique, reason)
		}
	}
	return fmt.Errorf("%w: 自动生产前检查失败：%s", batchfactoryv11.ErrInvalid, strings.Join(unique, "；"))
}

// automationBookPreflightReason 逐书预检，返回空串表示该书可执行。定时任务与
// 立即任务都走 run()，因此同一套逐书预检对两种启动方式一致生效。
func (c *BatchAutomationCoordinator) automationBookPreflightReason(ctx context.Context, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, runMode string) string {
	if isDirectFirstSourceBlocked(book) {
		title := strings.TrimSpace(book.Title)
		if title == "" {
			title = "未命名作品"
		}
		return fmt.Sprintf("《%s》书城原文获取失败，等待人工处理", title)
	}
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if comicSettingString(effective, "textModelId", "") == "" {
		return "文本模型未配置"
	}
	if runMode == "storyboard_only" {
		return ""
	}
	productionMode := batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch)
	if productionMode == "comic" {
		if comicSettingString(effective, "imageModelId", "") == "" {
			return "图片模型未配置；" + setupHintImage
		}
	} else if comicSettingString(effective, "videoModelId", "") == "" {
		return "视频模型未配置；" + setupHintVideoModel
	}
	if runMode == "full_submit" {
		// 漫画发布与视频发布使用的是两套配置键：漫画走 comicPublishSettings
		// （组织/音频分类/制作者），视频走 publishSettings。之前统一只查
		// publishSettings，导致漫画批次哪怕预设里已经配好组织，也会被判
		// “发布配置未绑定网站或组织”，整批书在派发前全部 blocked。
		publishKey := "publishSettings"
		if productionMode == "comic" {
			publishKey = "comicPublishSettings"
		}
		var publish map[string]any
		if raw := effective[publishKey]; len(raw) > 0 {
			_ = json.Unmarshal(raw, &publish)
		}
		if publish == nil || !hasAutomationPublishProfile(publish) {
			if productionMode == "comic" {
				return "漫画发布设置未绑定网站或组织"
			}
			return "发布配置未绑定网站或组织"
		}
		// N8 目录查询属于发布阶段依赖（音频分类/音频文件列表），
		// 漫画出图阶段完全不碰它。这里之前在 preflight 里统一调
		// comicPublishPreflight，一旦 N8 后台改了路径/挂了就把整批
		// 书在派发最开头全部 blocked，连一张图都出不来。
		// 改为挪到 run() 里进入 publish 阶段时再查，让出图流水线
		// 先跑起来，发布依赖单独失败不拖累其它书。
	}
	return ""
}

func reasonOrDefault(reason, fallback string) string {
	if strings.TrimSpace(reason) != "" {
		return strings.TrimSpace(reason)
	}
	return fallback
}

func hasAutomationPublishProfile(publish map[string]any) bool {
	for _, key := range []string{"websiteProfileId", "versionProfile", "organization", "organizationId", "versionId", "accountId"} {
		if strings.TrimSpace(fmt.Sprint(publish[key])) != "" && fmt.Sprint(publish[key]) != "<nil>" {
			return true
		}
	}
	return false
}

// mergeAutomationPresetConfig 合并“自动化预设”与批次已保存的统一配置。
// 预设只是默认值：批次「统一配置」优先，单书覆盖则由运行时的
// ResolveSettings(batch, book) 再叠加一次，最终优先级为
// 单书显式值 > 批次统一配置 > 预设默认。
func mergeAutomationPresetConfig(current, config map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range config {
		merged[key] = value
	}
	for key, value := range current {
		merged[key] = value
	}
	for _, key := range []string{"aiPromptConfig", "publishSettings"} {
		left, _ := current[key].(map[string]any)
		right, _ := config[key].(map[string]any)
		if left != nil || right != nil {
			nested := map[string]any{}
			for k, v := range right {
				nested[k] = v
			}
			for k, v := range left {
				nested[k] = v
			}
			merged[key] = nested
		}
	}
	return merged
}

func (c *BatchAutomationCoordinator) applyPreset(ctx context.Context, owner string, batch batchfactoryv11.Batch, preset *AutomationPreset, requestedProductionMode ...string) (batchfactoryv11.Batch, error) {
	current := map[string]any{}
	for key, raw := range batch.SettingsState.Patch {
		var value any
		if json.Unmarshal(raw, &value) == nil {
			current[key] = value
		}
	}
	config := preset.Config
	if len(config) == 0 {
		return batchfactoryv11.Batch{}, fmt.Errorf("自动化预设没有可用配置")
	}
	merged := mergeAutomationPresetConfig(current, config)
	if len(requestedProductionMode) > 0 {
		mode := strings.TrimSpace(requestedProductionMode[0])
		if mode == "comic" || mode == "video" {
			merged["productionMode"] = mode
		}
	}
	// 预设里存的是提示词 ID（presetId/slot），正文 body 存在系统预设库里。
	// 之前只有「保存统一配置」的 settings 接口会冻结正文，直接点“立即执行”
	// 走预设合并时不会冻结，导致漫画资产提示词 body 为空，资产阶段直接报
	// “请在资产设置中选择漫画资产提示词”。这里在落库前统一补一次冻结。
	if c.promptCatalog != nil {
		if value, ok := merged["aiPromptConfig"]; ok && value != nil {
			raw, marshalErr := json.Marshal(value)
			if marshalErr != nil {
				return batchfactoryv11.Batch{}, marshalErr
			}
			hydrated, hydrateErr := hydrateBatchFactoryPromptConfig(c.promptCatalog, raw)
			if hydrateErr != nil {
				return batchfactoryv11.Batch{}, hydrateErr
			}
			merged["aiPromptConfig"] = hydrated
		}
	}
	patch := batchfactoryv11.SettingsPatch{}
	for key, value := range merged {
		raw, err := json.Marshal(value)
		if err != nil {
			return batchfactoryv11.Batch{}, err
		}
		patch[key] = raw
	}
	if _, err := c.store.SaveSettings(ctx, owner, batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: patch, ExpectedRevision: batch.Revision}); err != nil {
		return batchfactoryv11.Batch{}, err
	}
	return c.store.GetBatch(ctx, owner, batch.ID)
}

func (c *BatchAutomationCoordinator) launchRun(owner, batchID, jobID string, job map[string]any, batch batchfactoryv11.Batch) {
	baseCtx := context.WithValue(context.Background(), bridgeIdentityKey{}, BridgeIdentity{Username: owner})
	automationCtx, cancel := context.WithCancel(baseCtx)
	c.controlMu.Lock()
	if c.runCancels == nil {
		c.runCancels = map[string]context.CancelFunc{}
	}
	if previous := c.runCancels[jobID]; previous != nil {
		previous()
	}
	c.runCancels[jobID] = cancel
	c.controlMu.Unlock()
	go func() {
		// 单个自动化任务的 goroutine panic 不允许带走整个进程：落成 failed，
		// 保留 panic 文本，便于用户看到并重试。
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("batch automation job %s panicked: %v", jobID, recovered)
				job["state"], job["lastError"], job["updatedAt"] = "failed", fmt.Sprintf("panic: %v", recovered), c.now()
				_ = c.setJob(owner, batchID, jobID, job)
			}
		}()
		c.run(automationCtx, owner, batchID, jobID, job, batch)
	}()
}

func (c *BatchAutomationCoordinator) cancelRun(jobID string) {
	c.controlMu.Lock()
	cancel := c.runCancels[jobID]
	delete(c.runCancels, jobID)
	c.controlMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func automationStringValue(value any, fallback string) string {
	if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
		return text
	}
	return fallback
}
func automationBoolValue(value any) bool { v, _ := value.(bool); return v }
func automationNumberValue(value any, fallback int) int {
	if v, ok := value.(float64); ok && v > 0 {
		return int(v)
	}
	if v, ok := value.(int); ok && v > 0 {
		return v
	}
	return fallback
}

// automationVideoRequestID scopes the idempotency key to the current run attempt.
// The historical "-video-<book>" id is reused by a reopened (retry) run AND by
// the in-run batch backfill round, which made SubmitBookProductionWithOptions
// return the previous failed production job — together with its stored provider
// and error — without resubmitting.
func automationVideoRequestID(jobID, bookID string, retryAttempt, batchRetry int) string {
	if retryAttempt > 0 || batchRetry > 0 {
		return fmt.Sprintf("%s-video-%s-r%db%d", jobID, bookID, retryAttempt, batchRetry)
	}
	return jobID + "-video-" + bookID
}

func automationCountsWithPending(total, pending int) map[string]any {
	return map[string]any{"total": total, "ready": 0, "running": 0, "pending": pending, "failed": 0, "blocked": 0}
}

func (c *BatchAutomationCoordinator) automationSubmitDeferred(job map[string]any) bool {
	return normalizeAutomationScheduleMode(automationStringValue(job["scheduleMode"], automationScheduleExecute)) == automationScheduleSubmit &&
		automationStringValue(job["runMode"], "full_submit") == "full_submit" &&
		strings.TrimSpace(automationStringValue(job["scheduledAt"], "")) != "" && !automationJobDue(job, c.now())
}

func (c *BatchAutomationCoordinator) latestComicCollection(ctx context.Context, owner, batchID, bookID, collectionID string) (batchfactoryv11.ComicCollection, error) {
	comicStore, ok := c.store.(batchfactoryv11.ComicPageStore)
	if !ok {
		return batchfactoryv11.ComicCollection{}, fmt.Errorf("漫画合集存储未配置")
	}
	collections, err := comicStore.ListComicCollections(ctx, owner, batchID, bookID)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	for _, collection := range collections {
		if strings.TrimSpace(collectionID) != "" && collection.ID == collectionID {
			return collection, nil
		}
	}
	if len(collections) == 0 {
		return batchfactoryv11.ComicCollection{}, fmt.Errorf("漫画合集不存在")
	}
	return collections[len(collections)-1], nil
}

// errSensitiveReviewPending 表示正文里仍有未处理的敏感命中，必须人工确认后才能上传。
var errSensitiveReviewPending = errors.New("有敏感词未处理，已保留原文，需人工确认后上传")

func (c *BatchAutomationCoordinator) publishVideo(ctx context.Context, owner, batchID, bookID string) error {
	if count, pending := c.sensitiveReviewPending(ctx, owner, batchID, bookID); pending {
		return fmt.Errorf("%w（%d 处）", errSensitiveReviewPending, count)
	}
	if c.publish121 != nil {
		_, err := c.publish121.Submit(ctx, owner, batchID, bookID, map[string]any{})
		return err
	}
	if c.external == nil {
		return fmt.Errorf("视频管理系统上传服务未配置")
	}
	intent, err := c.external.CreateIntent(ctx, owner, string(external.Provider121), batchID, bookID, json.RawMessage(`{}`))
	if err != nil {
		return err
	}
	if _, err = c.external.ConfirmIntent(ctx, owner, intent.ID); err != nil {
		return err
	}
	_, err = c.external.Submit(ctx, owner, string(external.Provider121), intent.ID)
	return err
}

// sensitiveReviewPending 读取最新书状态，判断正文里是否还有未处理的敏感命中。
// 返回命中数量与是否需要人工确认；读取失败按“不阻断”处理，避免误伤正常批次。
func (c *BatchAutomationCoordinator) sensitiveReviewPending(ctx context.Context, owner, batchID, bookID string) (int, bool) {
	if c == nil || c.store == nil {
		return 0, false
	}
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return 0, false
	}
	for _, book := range batch.Books {
		if book.ID != bookID {
			continue
		}
		value := book.SourceMetadata["sensitiveReviewRequired"]
		if automationBoolValue(value) || strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), "true") {
			return automationNumberValue(book.SourceMetadata["sensitiveReviewCount"], 0), true
		}
		return 0, false
	}
	return 0, false
}

func comicBootstrapReady(book batchfactoryv11.Book, _ batchfactoryv11.SettingsPatch) bool {
	if !comicAssetsReady(book) || len(book.Videos) == 0 {
		return false
	}
	for _, page := range book.Videos {
		if page.ComicSourceSegment == nil || strings.TrimSpace(page.ComicSourceSegment.SourceText) == "" || strings.TrimSpace(page.ComicPrompt) == "" || strings.TrimSpace(page.ComicImagePrompt) == "" {
			return false
		}
	}
	return true
}

func comicAssetsReady(book batchfactoryv11.Book) bool {
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Source) == "comic" && strings.TrimSpace(asset.Prompt) != "" {
			return true
		}
	}
	return false
}

type comicPreparationAction string

const (
	comicPreparationBootstrap      comicPreparationAction = "bootstrap"
	comicPreparationMissingPrompts comicPreparationAction = "missing_prompts"
	comicPreparationReady          comicPreparationAction = "ready"
)

func selectComicPreparationAction(book batchfactoryv11.Book, effective batchfactoryv11.SettingsPatch) comicPreparationAction {
	if !comicAssetsReady(book) || batchfactoryv11.ComicAssetIdentityIssue(book) != "" || len(book.Videos) == 0 {
		return comicPreparationBootstrap
	}
	for _, page := range book.Videos {
		if page.ComicSourceSegment == nil || strings.TrimSpace(page.ComicSourceSegment.SourceText) == "" {
			return comicPreparationBootstrap
		}
	}
	if !comicBootstrapReady(book, effective) {
		return comicPreparationMissingPrompts
	}
	return comicPreparationReady
}

func automationBooksMissingSource(books []batchfactoryv11.Book) []batchfactoryv11.Book {
	missing := make([]batchfactoryv11.Book, 0)
	for _, book := range books {
		if len(productionLines(book)) == 0 {
			missing = append(missing, book)
		}
	}
	return missing
}

// acknowledgeComicRetry clears the retry marker once the comic worker has
// actually picked the book up. The marker is only needed by the video path to
// decide whether an existing video can be reused; leaving it on a comic entry
// makes the UI report "重试排队中" even after the retry has already finished
// and failed again.
func acknowledgeComicRetry(entry map[string]any) {
	if entry != nil {
		delete(entry, "retryRequested")
	}
}

const comicAutomaticRetryLimit = 3

// 缺页全部是在途/过期租约时的额外重排次数；这条路径不消耗内容审核预算。
const comicInFlightRetryLimit = 2

// 漫画整本生成不再使用固定的 60 分钟上限：一本 35 页漫画实测端到端需要
// 110~155 分钟，固定上限会把正常任务当作超时失败。这里改为“预算 + 进度
// 心跳”租约——初始预算按页数和并发量算，每完成一页就续期，只有真正停滞
// （连续 idle 时间没有任何进度事件）才取消。
const (
	comicGenerationPerPageBudget = 20 * time.Minute
	comicGenerationMinBudget     = 60 * time.Minute
	comicGenerationCushion       = 30 * time.Minute
	comicGenerationIdleExtend    = 20 * time.Minute
	comicGenerationHardCap       = 6 * time.Hour
	comicPublishStageTimeout     = 45 * time.Minute
)

// comicGenerationBudget 按“页数 / 有效并发”估算一本漫画的初始时间预算。
// 初始预算只需要覆盖第一页出结果之前的时间，之后由心跳续期接管。
func comicGenerationBudget(pageCount, concurrency int) time.Duration {
	if pageCount < 1 {
		pageCount = 1
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > 8 {
		concurrency = 8
	}
	rounds := (pageCount + concurrency - 1) / concurrency
	budget := time.Duration(rounds)*comicGenerationPerPageBudget + comicGenerationCushion
	if budget < comicGenerationMinBudget {
		budget = comicGenerationMinBudget
	}
	return budget
}

func comicGenerationPageCount(book batchfactoryv11.Book) int {
	pages := len(book.Videos)
	if pages > comicPageGenerationLimit {
		pages = comicPageGenerationLimit
	}
	if pages < 1 {
		pages = 1
	}
	return pages
}

// comicGenerationLease 是进度心跳租约。touch 每次把 deadline 推到“现在 +
// idle”，hardDeadline 防止心跳无限延长；两者都由后台 watchdog 检查。
type comicGenerationLease struct {
	mu           sync.Mutex
	deadline     time.Time
	hardDeadline time.Time
}

func (l *comicGenerationLease) touch(extend time.Duration) {
	if l == nil {
		return
	}
	l.mu.Lock()
	next := time.Now().Add(extend)
	if next.After(l.hardDeadline) {
		next = l.hardDeadline
	}
	if next.After(l.deadline) {
		l.deadline = next
	}
	l.mu.Unlock()
}

func (l *comicGenerationLease) expired(now time.Time) bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return now.After(l.deadline) || now.After(l.hardDeadline)
}

var comicGenerationLeaseTick = 15 * time.Second

func comicGenerationLeaseContext(parent context.Context, initial, hardCap time.Duration) (context.Context, *comicGenerationLease, context.CancelFunc) {
	start := time.Now()
	ctx, cancel := context.WithCancel(parent)
	lease := &comicGenerationLease{deadline: start.Add(initial), hardDeadline: start.Add(hardCap)}
	go func() {
		ticker := time.NewTicker(comicGenerationLeaseTick)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				if lease.expired(now) {
					cancel()
					return
				}
			}
		}
	}()
	return ctx, lease, cancel
}

func retryableComicAutomationError(err error) bool {
	if err == nil {
		return false
	}
	// A page can still have a provider task from the previous attempt. This is
	// a temporary coordination state, not a configuration error; wait and let
	// the idempotent book generator continue from the missing page.
	message := err.Error()
	if errors.Is(err, batchfactoryv11.ErrConflict) && (strings.Contains(message, "当前漫画页正在生成") || strings.Contains(message, "当前漫画页已有生成任务")) {
		return true
	}
	if status, ok := comicAutomationHTTPStatus(message); ok && status >= 400 && status < 500 && status != http.StatusRequestTimeout && status != http.StatusConflict && status != http.StatusTooManyRequests {
		return false
	}
	for _, permanent := range []error{
		batchfactoryv11.ErrConflict,
		batchfactoryv11.ErrInvalid,
		batchfactoryv11.ErrNotFound,
		batchfactoryv11.ErrUnavailable,
	} {
		if errors.Is(err, permanent) {
			return false
		}
	}
	return true
}

func comicAutomationHTTPStatus(message string) (int, bool) {
	_, suffix, found := strings.Cut(message, "HTTP ")
	if !found {
		return 0, false
	}
	fields := strings.Fields(suffix)
	if len(fields) == 0 {
		return 0, false
	}
	status, err := strconv.Atoi(strings.TrimRight(fields[0], ":,;"))
	return status, err == nil
}

func waitForComicAutomationRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// generateComicWithRetry retries only transient provider failures. Successful
// pages remain reusable inside the book generator, so a retry continues from
// the first missing page instead of regenerating the whole comic.
func (c *BatchAutomationCoordinator) generateComicWithRetry(
	ctx context.Context,
	entry map[string]any,
	persist func(),
	operation func(context.Context) (batchfactoryv11.ComicCollection, error),
	wait func(context.Context, time.Duration) error,
) (batchfactoryv11.ComicCollection, error) {
	if wait == nil {
		wait = waitForComicAutomationRetry
	}
	now := time.Now
	if c != nil && c.now != nil {
		now = c.now
	}
	for attempt := 1; attempt <= comicAutomaticRetryLimit; attempt++ {
		if err := ctx.Err(); err != nil {
			return batchfactoryv11.ComicCollection{}, err
		}
		collection, err := operation(ctx)
		if err == nil {
			delete(entry, "retryAt")
			return collection, nil
		}
		if !retryableComicAutomationError(err) || attempt == comicAutomaticRetryLimit {
			// retryAt only describes an active automatic retry. Leaving the
			// previous timestamp on a terminal failure makes the UI claim that
			// another retry is scheduled even though the worker has stopped.
			delete(entry, "retryAt")
			return batchfactoryv11.ComicCollection{}, err
		}
		delay := time.Duration(attempt*2) * time.Second
		retryAt := now().Add(delay)
		entry["retryCount"] = attempt
		entry["retryAt"] = retryAt
		entry["message"] = fmt.Sprintf("漫画图片生成失败，自动重试中（%d/%d）", attempt, comicAutomaticRetryLimit)
		entry["updatedAt"] = now()
		persist()
		if waitErr := wait(ctx, delay); waitErr != nil {
			return batchfactoryv11.ComicCollection{}, waitErr
		}
		// retryAt describes the sleep before a retry, not the retry attempt
		// itself. Clear it before the worker starts the next provider call so the
		// UI cannot show an old retry time while work is actively running.
		delete(entry, "retryAt")
		entry["message"] = "正在重新生成漫画图片"
		entry["updatedAt"] = now()
		persist()
	}
	return batchfactoryv11.ComicCollection{}, fmt.Errorf("漫画图片自动重试结束")
}

// comicContentPolicyRetryDelay 是整本等待被拒页重新生成的退避间隔。
const comicContentPolicyRetryDelay = 10 * time.Minute

// comicContentPolicyRetryLimit 是同一本书因图片审核拒绝而自动重试的次数上限。
const comicContentPolicyRetryLimit = 3

// comicMissingPageError 表示合集因为存在没有主图的漫画页而无法创建。自动化
// 路径下，这种情况只来自被内容审核拒绝后跳过的页。
func comicMissingPageError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "个漫画页缺少主图")
}

// comicMissingAssetError 表示漫画资产图阶段结束后仍有资产没有主图。审核拒绝
// 的资产图会被跳过而不是中断整本，所以这里同样按“等待退避重试”处理，避免
// 一次软拒绝就把整本书打成永久失败。
func comicMissingAssetError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "漫画资产图未全部写回")
}

// scheduleComicContentPolicyRetry 把整本标记为“等待缺页重试”，由调度器在
// 退避时间到达后重新执行。返回 false 表示重试预算已用完，调用方应转入失败。
func (c *BatchAutomationCoordinator) scheduleComicContentPolicyRetry(entry map[string]any, persist func(), stage string) bool {
	if entry == nil || persist == nil {
		return false
	}
	attempt := automationNumberValue(entry["contentPolicyRetryCount"], 0)
	if attempt >= comicContentPolicyRetryLimit {
		return false
	}
	attempt++
	if stage == "" {
		stage = "comic_images"
	}
	now := c.now()
	message := fmt.Sprintf("有漫画资产图或漫画页两次降级后仍被内容审核拒绝，%d 分钟后自动重试（%d/%d）", int(comicContentPolicyRetryDelay.Minutes()), attempt, comicContentPolicyRetryLimit)
	entry["contentPolicyRetryCount"] = attempt
	entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "running", stage, "", message, now
	entry["retryRequested"] = true
	entry["retryReason"] = "content_policy"
	entry["retryAt"] = now.Add(comicContentPolicyRetryDelay)
	appendAutomationBookLog(entry, now, stage, message)
	persist()
	return true
}

func automationOpeningEnabled(batch batchfactoryv11.Batch, book batchfactoryv11.Book) bool {
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	raw, ok := effective["openingEnabled"]
	return ok && string(raw) == "true"
}

func productionLines(book batchfactoryv11.Book) []string {
	limit := 0
	if raw, ok := book.SourceMetadata["contentRangeLines"]; ok {
		switch value := raw.(type) {
		case float64:
			limit = int(value)
		case int:
			limit = value
		case json.Number:
			if parsed, err := value.Int64(); err == nil {
				limit = int(parsed)
			}
		}
	}
	lines := make([]string, 0, limit)
	for _, line := range strings.Split(strings.ReplaceAll(book.SourceText, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if limit > 0 && len(lines) >= limit {
			break
		}
	}
	return lines
}

// hasReusableAudioPlan prevents a retry from needlessly calling TTS again when
// the selected source text and every measured line duration are already stored
// on the book. The plan remains part of the production chain; this only reuses
// the immutable result of the earlier timing step.
func hasReusableAudioPlan(book batchfactoryv11.Book) bool {
	lines := productionLines(book)
	if len(lines) == 0 {
		return false
	}
	patch := book.SettingsState.Patch
	var enabled bool
	if raw, ok := patch["audioPlanningEnabled"]; !ok || json.Unmarshal(raw, &enabled) != nil || !enabled {
		return false
	}
	var durations []float64
	if raw, ok := patch["audioLineDurations"]; !ok || json.Unmarshal(raw, &durations) != nil || len(durations) != len(lines) {
		return false
	}
	var total float64
	if raw, ok := patch["audioDurationSeconds"]; !ok || json.Unmarshal(raw, &total) != nil || total <= 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return false
	}
	measured := 0.0
	for _, duration := range durations {
		if duration <= 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
			return false
		}
		measured += duration
	}
	return math.Abs(measured-total) <= 0.05
}

func isDirectFirstSourceBlocked(book batchfactoryv11.Book) bool {
	strategy := strings.ToLower(strings.TrimSpace(stringValue(book.SourceMetadata["originalReadStrategy"])))
	if strategy == "" {
		strategy = strings.ToLower(strings.TrimSpace(stringValue(book.SourceMetadata["sourceReadStrategy"])))
	}
	stage := strings.ToLower(strings.TrimSpace(stringValue(book.SourceMetadata["originalReadStage"])))
	return strategy == "direct_first" && (stage == "failed" || stage == "source_blocked")
}

// ensureAutomationSource fills a book whose intake retained its ID but not its
// original text. Immediate runs and retries must recover this required input
// before starting audio planning rather than failing the whole production path.
func (c *BatchAutomationCoordinator) ensureAutomationSource(ctx context.Context, owner, batchID string, book batchfactoryv11.Book) (batchfactoryv11.Book, error) {
	// 调度队列里保存的是入队那一刻的书快照。重试轮次之间 persistBatchFetchFailure
	// 和其它阶段都会 revision+1，继续用快照的旧 revision 写会把自己制造的写入判成
	// 冲突。每次进入本阶段先回读最新快照，后续写入再由 _withRetry 兜底。
	book = c.refreshAutomationBook(ctx, owner, batchID, book)
	if strings.EqualFold(stringValue(book.SourceMetadata["processingStatus"]), "failed") && strings.TrimSpace(book.TxtText) != "" {
		maxTxt := int(targetNumber(book.SourceMetadata["maxTxt"]))
		fetched := batchFetchResultFromStoredRaw(book, maxTxt)
		fetched, metadata, processErr := prepareBatchFetchedSource(ctx, c.ttsStore, owner, book.BookID, fetched, c.classification, c.batchSensitiveConfigFor(ctx, owner, batchID, book))
		for key, value := range book.SourceMetadata {
			if _, exists := metadata[key]; !exists {
				metadata[key] = value
			}
		}
		productionText := fetched.ProcessedText
		if processErr != nil {
			productionText = ""
		}
		updated, err := c.captureAutomationSource(ctx, owner, batchID, book, batchfactoryv11.CaptureBookSourceInput{
			SourceText: productionText, RawSourceText: fetched.OriginalRaw, ReplaceSource: true,
			SourceTitle: book.Title, SourceMetadata: metadata,
		})
		if err != nil {
			return book, fmt.Errorf("保存重新处理的正文失败：%w", err)
		}
		if processErr != nil {
			return updated, fmt.Errorf("重新处理正文失败：%w", processErr)
		}
		return updated, nil
	}
	if isDirectFirstSourceBlocked(book) {
		return book, fmt.Errorf("书城原文获取失败，等待人工处理：%s", reasonOrDefault(strings.TrimSpace(stringValue(book.SourceMetadata["originalReadError"])), "请手动重试直取或确认后使用 OCR"))
	}
	if len(productionLines(book)) > 0 {
		// Classification is a metadata-enrichment step, not a production gate.
		// Any usable gender/style pair continues; an incomplete pair is
		// finalized into a default pair instead of stopping automation.
		gender := validBatchGender(firstMetadataString(book.SourceMetadata, "gender", "gender_name"))
		style := validBatchStyle(firstMetadataString(book.SourceMetadata, "style", "styleName", "style_name"))
		if gender != "" && style != "" {
			return book, nil
		}
		info := novelfetchworkshop.BookInfo{
			Category: firstMetadataString(book.SourceMetadata, "category"),
			Genre:    firstMetadataString(book.SourceMetadata, "genre"),
			Themes:   firstMetadataString(book.SourceMetadata, "themes", "theme"),
			Labels:   firstMetadataString(book.SourceMetadata, "labels", "tags"),
		}
		metadata := finalizeFetchedBatchClassification(classifyFetchedBatchBook(ctx, owner, info, book.SourceText, book.SourceMetadata, c.classification))
		updated, err := c.updateAutomationBookMetadata(ctx, owner, batchID, book, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata})
		if err != nil {
			return book, fmt.Errorf("保存分类状态失败：%w", err)
		}
		return updated, nil
	}
	if c == nil || c.fetcher == nil || c.store == nil {
		return book, fmt.Errorf("正文尚未获取，且自动获取服务未配置")
	}
	platform := strings.TrimSpace(book.Platform)
	if platform == "" {
		return book, fmt.Errorf("正文尚未获取，且缺少书城平台")
	}
	bookID := strings.TrimSpace(book.BookID)
	if bookID == "" {
		bookID = strings.TrimSpace(book.ID)
	}
	if bookID == "" {
		return book, fmt.Errorf("正文尚未获取，且缺少小说编号")
	}
	captureCharacters := 4000
	if value, ok := book.SourceMetadata["contentCaptureCharacters"]; ok {
		switch typed := value.(type) {
		case int:
			captureCharacters = typed
		case int64:
			captureCharacters = int(typed)
		case float64:
			captureCharacters = int(typed)
		case json.Number:
			if parsed, parseErr := typed.Int64(); parseErr == nil {
				captureCharacters = int(parsed)
			}
		}
	}
	if captureCharacters < 1000 {
		captureCharacters = 1000
	}
	if captureCharacters > 100000 {
		captureCharacters = 100000
	}
	fetched, err, attempts := fetchBatchSourceWithRetry(ctx, c.fetcher, bookID, platform, captureCharacters)
	if err != nil {
		persistBatchFetchFailure(ctx, c.store, owner, batchID, book, err, attempts)
		return book, fmt.Errorf("自动获取正文失败：%w", err)
	}
	if len(strings.TrimSpace(fetched.ProcessedText)) == 0 {
		return book, fmt.Errorf("自动获取正文失败：上游未返回可用正文")
	}
	fetched, metadata, processErr := prepareBatchFetchedSource(ctx, c.ttsStore, owner, bookID, fetched, c.classification, c.batchSensitiveConfigFor(ctx, owner, batchID, book))
	for key, value := range book.SourceMetadata {
		if _, exists := metadata[key]; !exists {
			metadata[key] = value
		}
	}
	productionText := fetched.ProcessedText
	if processErr != nil {
		productionText = ""
	}
	updated, err := c.captureAutomationSource(ctx, owner, batchID, book, batchfactoryv11.CaptureBookSourceInput{
		SourceText: productionText, RawSourceText: fetched.OriginalRaw, ReplaceSource: true, SourceTitle: book.Title, SourceMetadata: metadata,
	})
	if err != nil {
		return book, fmt.Errorf("保存自动获取正文失败：%w", err)
	}
	if processErr != nil {
		return updated, fmt.Errorf("自动处理获取正文失败：%w", processErr)
	}
	return updated, nil
}

// automationBookWriteAttempts bounds the optimistic-lock retries for book
// writes started by an already queued automation snapshot.
const automationBookWriteAttempts = 3

// refreshAutomationBook rereads the persisted book so optimistic-lock writes use
// the current revision instead of the one captured when the batch was queued.
func (c *BatchAutomationCoordinator) refreshAutomationBook(ctx context.Context, owner, batchID string, book batchfactoryv11.Book) batchfactoryv11.Book {
	if c == nil || c.store == nil || strings.TrimSpace(book.ID) == "" {
		return book
	}
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return book
	}
	for _, candidate := range batch.Books {
		if candidate.ID == book.ID {
			return candidate
		}
	}
	return book
}

// batchSensitiveConfigFor resolves the batch-level sensitive settings with the
// book-level override on top, matching the existing settings inheritance order.
func (c *BatchAutomationCoordinator) batchSensitiveConfigFor(ctx context.Context, owner, batchID string, book batchfactoryv11.Book) batchSensitiveConfig {
	if c == nil || c.store == nil {
		return defaultBatchSensitiveConfig()
	}
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return defaultBatchSensitiveConfig()
	}
	return resolveBatchSensitiveConfig(batch.SettingsState.Patch, book.SettingsState.Patch)
}

// captureAutomationSource saves fetched text and rereads the revision on a
// conflict instead of surfacing "revision conflict" as a failed fetch.
func (c *BatchAutomationCoordinator) captureAutomationSource(ctx context.Context, owner, batchID string, book batchfactoryv11.Book, input batchfactoryv11.CaptureBookSourceInput) (batchfactoryv11.Book, error) {
	if c == nil || c.store == nil {
		return book, fmt.Errorf("正文存储未配置")
	}
	latest := book
	var err error
	for attempt := 0; attempt < automationBookWriteAttempts; attempt++ {
		input.ExpectedRevision = latest.Revision
		var updated batchfactoryv11.Book
		updated, err = c.store.CaptureBookSource(ctx, owner, batchID, latest.ID, input)
		if err == nil {
			return updated, nil
		}
		if !errors.Is(err, batchfactoryv11.ErrConflict) {
			return latest, err
		}
		latest = c.refreshAutomationBook(ctx, owner, batchID, latest)
		if len(productionLines(latest)) > 0 {
			// 并发路径已经写入正文，直接使用最新结果，不再覆盖。
			return latest, nil
		}
	}
	return latest, err
}

// updateAutomationBookMetadata applies a metadata write with the same
// reread-and-retry policy as captureAutomationSource.
func (c *BatchAutomationCoordinator) updateAutomationBookMetadata(ctx context.Context, owner, batchID string, book batchfactoryv11.Book, input batchfactoryv11.UpdateBookMetadataInput) (batchfactoryv11.Book, error) {
	if c == nil || c.store == nil {
		return book, fmt.Errorf("正文存储未配置")
	}
	latest := book
	var err error
	for attempt := 0; attempt < automationBookWriteAttempts; attempt++ {
		input.ExpectedRevision = latest.Revision
		var updated batchfactoryv11.Book
		updated, err = c.store.UpdateBookMetadata(ctx, owner, batchID, latest.ID, input)
		if err == nil {
			return updated, nil
		}
		if !errors.Is(err, batchfactoryv11.ErrConflict) {
			return latest, err
		}
		latest = c.refreshAutomationBook(ctx, owner, batchID, latest)
	}
	return latest, err
}

// ensureAudioPlan is deliberately independent of the final audioSource switch.
// The switch only chooses which track is used during the final merge; every
// production run still creates TTS and measures every selected source line so
// the storyboard timeline is based on real durations.
func automationMergeAudioDuration(book batchfactoryv11.Book, _ batchfactoryv11.ProductionJob) (float64, error) {
	var duration float64
	raw, ok := book.SettingsState.Patch["audioDurationSeconds"]
	if !ok || json.Unmarshal(raw, &duration) != nil || duration <= 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return 0, fmt.Errorf("配音计划缺少有效的总时长")
	}
	return duration, nil
}

func (c *BatchAutomationCoordinator) ensureAudioPlan(ctx context.Context, owner, batchID string, book batchfactoryv11.Book) (string, float64, error) {
	if c == nil || c.ttsStore == nil || c.store == nil {
		return "", 0, fmt.Errorf("配音与逐行时长服务未配置")
	}
	if hasReusableAudioPlan(book) {
		var artifactID string
		_ = json.Unmarshal(book.SettingsState.Patch["ttsAudioArtifactId"], &artifactID)
		if strings.TrimSpace(artifactID) != "" && c.localFiles != nil {
			if file, err := c.localFiles.Open(strings.TrimSpace(artifactID) + ".mp3"); err == nil {
				_ = file.Close()
				duration, durationErr := automationMergeAudioDuration(book, batchfactoryv11.ProductionJob{})
				return strings.TrimSpace(artifactID), duration, durationErr
			}
		}
	}
	if c.localFiles == nil {
		return "", 0, fmt.Errorf("TTS 音频存储未配置")
	}
	lines := productionLines(book)
	if len(lines) == 0 {
		return "", 0, fmt.Errorf("正文没有可测量的生产行")
	}
	model, ok := configuredTTSModel(ctx, c.ttsStore, owner)
	probe := c.audioProbe
	if probe == nil {
		probe = batchfactoryv11.FFprobeH3AudioDuration{}
	}
	durations := make([]float64, 0, len(lines))
	estimated := false
	tempDir, err := os.MkdirTemp("", "qiantie-tts-plan-")
	if err != nil {
		return "", 0, err
	}
	defer os.RemoveAll(tempDir)
	audioInputs := make([]string, 0, len(lines))
	hash := sha256.New()
	for index, line := range lines {
		lineCtx, cancel := context.WithTimeout(ctx, 75*time.Second)
		ttsInput := ttsRequest{Input: line, Voice: "zh-CN-XiaoxiaoNeural", Speed: 1.8, Pitch: "10", Style: "general"}
		var audio []byte
		var err error
		if ok {
			audio, _, err = requestConfiguredTTSBytes(lineCtx, ttsInput, model)
		} else {
			// Match the /api/tts route: an enabled catalog model wins, otherwise
			// the configured default TTS service remains the valid fallback.
			audio, _, err = requestDefaultTTSBytes(lineCtx, ttsInput)
		}
		cancel()
		if err != nil {
			return "", 0, fmt.Errorf("TTS 生成失败：%w", err)
		}
		linePath := filepath.Join(tempDir, fmt.Sprintf("line-%04d.mp3", index))
		if writeErr := os.WriteFile(linePath, audio, 0o600); writeErr != nil {
			return "", 0, writeErr
		}
		audioInputs = append(audioInputs, linePath)
		measured, err := batchfactoryv11.MeasureAudioDuration(ctx, probe, audio)
		if err != nil {
			return "", 0, fmt.Errorf("TTS 时长测量失败：%w", err)
		}
		if measured.MS <= 0 {
			return "", 0, fmt.Errorf("TTS 时长测量失败：测量结果无效")
		}
		if measured.Estimated {
			estimated = true
			log.Printf("[TTS] 第 %d 行时长无法精确解析，已按估算值继续 source=%s durationMS=%d", index+1, measured.Source, measured.MS)
		}
		durations = append(durations, float64(measured.MS)/1000)
		_, _ = hash.Write(audio)
	}
	total := 0.0
	for _, duration := range durations {
		total += duration
	}
	combinedPath := filepath.Join(tempDir, "tts.mp3")
	if err := (&mergeworker.FFmpegRunner{}).ConcatAudioFiles(ctx, audioInputs, combinedPath); err != nil {
		return "", 0, fmt.Errorf("TTS 音频合成失败：%w", err)
	}
	digest := hash.Sum(nil)
	artifactID := fmt.Sprintf("tts-%x", digest[:20])
	combined, err := os.Open(combinedPath)
	if err != nil {
		return "", 0, err
	}
	_, saveErr := c.localFiles.SaveMedia(artifactID, "audio/mpeg", "mp3", combined)
	_ = combined.Close()
	if saveErr != nil && !errors.Is(saveErr, localartifact.ErrAlreadyExists) {
		return "", 0, fmt.Errorf("保存 TTS 音频失败：%w", saveErr)
	}
	latest, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return "", 0, err
	}
	for _, candidate := range latest.Books {
		if candidate.ID != book.ID {
			continue
		}
		lineRaw, _ := json.Marshal(durations)
		totalRaw, _ := json.Marshal(total)
		enabledRaw, _ := json.Marshal(true)
		estimatedRaw, _ := json.Marshal(estimated)
		_, err = c.store.SaveSettings(ctx, owner, batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batchID, BookID: book.ID}, batchfactoryv11.SettingsUpdate{
			Patch: batchfactoryv11.SettingsPatch{
				"audioPlanningEnabled":   enabledRaw,
				"audioDurationSeconds":   totalRaw,
				"audioDurationEstimated": estimatedRaw,
				"audioLineDurations":     lineRaw,
				"ttsFingerprint":         json.RawMessage(fmt.Sprintf("%q", func() string { digest := hash.Sum(nil); return fmt.Sprintf("tts-%x", digest[:12]) }())),
				"ttsAudioArtifactId":     json.RawMessage(fmt.Sprintf("%q", artifactID)),
			},
			ExpectedRevision: candidate.Revision,
		})
		return artifactID, total, err
	}
	return "", 0, batchfactoryv11.ErrNotFound
}

// GenerateBookAudioPlan runs only the TTS-and-duration part of a book's
// Batch Factory production. It intentionally does not enqueue assets,
// storyboards, videos, merges, or publishing work.
func (c *BatchAutomationCoordinator) GenerateBookAudioPlan(ctx context.Context, owner, batchID, bookID string) (map[string]any, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("%w: 配音与逐行时长服务未配置", batchfactoryv11.ErrUnavailable)
	}
	owner, batchID, bookID = strings.TrimSpace(owner), strings.TrimSpace(batchID), strings.TrimSpace(bookID)
	if owner == "" || batchID == "" || bookID == "" {
		return nil, batchfactoryv11.ErrInvalid
	}
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	for _, candidate := range batch.Books {
		if candidate.ID != bookID {
			continue
		}
		book, err := c.ensureAutomationSource(ctx, owner, batchID, candidate)
		if err != nil {
			return nil, err
		}
		artifactID, durationSeconds, err := c.ensureAudioPlan(ctx, owner, batchID, book)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"bookId":          book.ID,
			"audioArtifactId": artifactID,
			"durationSeconds": durationSeconds,
		}, nil
	}
	return nil, batchfactoryv11.ErrNotFound
}

func (c *BatchAutomationCoordinator) run(ctx context.Context, owner, batchID, jobID string, job map[string]any, batch batchfactoryv11.Batch) {
	// Each external stage must have a bounded lifetime.  The UI should receive a
	// failed state instead of leaving a book stuck at "running" when an upstream
	// model endpoint stops responding.
	const stageTimeout = 10 * time.Minute
	// 漫画提示词是逐页文本请求，35 页在 5 路并发下实测需要 10~11 分钟，已经正好
	// 压在 10 分钟的通用阶段超时上，导致每本书都要先超时再补跑一轮。这里单独给
	// 提示词阶段留出 20 分钟，避免把正常慢请求当成失败重试。
	const comicPromptStageTimeout = 20 * time.Minute
	// 提示词请求是纯文本调用，不占用图片网关，实测可以把并发从 5 提到 8 而不
	// 触发上游限流，直接缩短提示词阶段的排队轮次。
	const comicPromptConcurrency = 8
	// 批次末尾统一回补：单本书失败时先跳过，等本批其余书全部跑完，再按顺序重跑
	// 失败的书。生成链路是幂等的，回补会从缺失的页/阶段继续，不会重做已完成部分。
	books := job["books"].(map[string]any)
	// Video submissions are idempotent per request id. On a reopened (retry) run
	// the plain "-video-<book>" id would return the previous failed production
	// job instead of submitting again, so retries use a distinct id suffix.
	retryAttempt := automationNumberValue(job["retryAttempt"], 0)
	videoRequestID := func(bookID string, batchRetry int) string {
		return automationVideoRequestID(jobID, bookID, retryAttempt, batchRetry)
	}
	concurrency := automationNumberValue(job["concurrency"], 1)
	if concurrency > 4 {
		concurrency = 4
	}
	// 每本书各自最多 4 路出图，但整个任务还需要一个总并发上限：实测图片网关
	// 在并发响应体过多时会直接 RST（HTTP 200 后 wsarecv 被重置）。URL 模式下
	// 小响应很稳定，这里给到 8 路，资产图与漫画页各占一半，避免资产图阶段把
	// 漫画页饿住。可用 job 字段覆盖。
	imageGlobalConcurrency := automationNumberValue(job["imageGlobalConcurrency"], 8)
	if imageGlobalConcurrency < 1 {
		imageGlobalConcurrency = 8
	}
	ctx = withShuihuoGlobalImageConcurrency(ctx, int(imageGlobalConcurrency))
	var failed atomic.Int32
	var workers sync.WaitGroup
	var jobMu sync.Mutex
	semaphore := make(chan struct{}, concurrency)
	// queue 是本次要调度的书。失败的书一旦登记为待回补，就立刻重新入队；
	// 只要有一个并发槽位空出就立即回补，不再等整批其它书全部跑完。
	queue := append([]batchfactoryv11.Book(nil), batch.Books...)
	var queueMu sync.Mutex
	queueCond := sync.NewCond(&queueMu)
	inflight := 0
	enqueueBatchRetry := func(book batchfactoryv11.Book) {
		queueMu.Lock()
		queue = append(queue, book)
		queueCond.Signal()
		queueMu.Unlock()
	}
	qi := 0
	for {
		queueMu.Lock()
		for qi >= len(queue) && inflight > 0 {
			queueCond.Wait()
		}
		if qi >= len(queue) {
			queueMu.Unlock()
			break
		}
		book := queue[qi]
		qi++
		inflight++
		queueMu.Unlock()
		// Workers persist their book entry into the same maps, so the initial
		// read and clone must hold the same mutex to avoid a concurrent map
		// read/write against the previous round's workers.
		jobMu.Lock()
		initialEntry, _ := books[book.ID].(map[string]any)
		if initialEntry == nil {
			initialEntry = map[string]any{}
		}
		entry := cloneAutomationEntry(initialEntry)
		jobMu.Unlock()
		persist := func() {
			jobMu.Lock()
			clear(initialEntry)
			for key, value := range entry {
				initialEntry[key] = value
			}
			books[book.ID] = initialEntry
			job["counts"] = automationCountsFromBooks(books)
			job["updatedAt"] = c.now()
			_ = c.setJob(owner, batchID, jobID, job)
			jobMu.Unlock()
		}
		// 并发槽位还没拿到时如实写 queued，前端才会显示“排队中”。
		// 回补书保留原 stage/retryAt/batchRetryCount，拿到槽位后仍能续跑。
		if markAutomationBookQueued(entry, c.now()) {
			persist()
		}
		workers.Add(1)
		semaphore <- struct{}{}
		// 等待期间用户可能已暂停/停止；拿到槽位后以持久化状态为准，
		// 避免用阻塞前的旧快照覆盖控制态。
		jobMu.Lock()
		if latest, _ := books[book.ID].(map[string]any); latest != nil {
			if status := automationStringValue(latest["status"], ""); status == "paused" || status == "stopped" {
				entry = cloneAutomationEntry(latest)
			}
		}
		jobMu.Unlock()
		if markAutomationBookStarted(entry, c.now()) {
			persist()
		}
		go func(book batchfactoryv11.Book) {
			defer workers.Done()
			defer func() {
				<-semaphore
				queueMu.Lock()
				inflight--
				queueCond.Broadcast()
				queueMu.Unlock()
			}()
			// The one-item loop lets every existing stage exit its own book with
			// continue while the other book workers keep running.
			for range [1]struct{}{} {
				// 单本书的执行日志供前端展开查看；并发出图时多个 goroutine 会
				// 同时上报阶段与去敏、审核拒绝、降级重试事件，因此单独加锁，
				// 保证日志追加与持久化不并发写同一个 entry。
				var entryLogMu sync.Mutex
				var generationLease *comicGenerationLease
				recordComicEvent := func(stage, message string) {
					// 任何出图/资产进度事件都算心跳：只要还在推进，就不判定超时。
					if generationLease != nil {
						generationLease.touch(comicGenerationIdleExtend)
					}
					entryLogMu.Lock()
					entry["stage"], entry["message"], entry["updatedAt"] = stage, message, c.now()
					delete(entry, "retryAt")
					appendAutomationBookLog(entry, c.now(), stage, message)
					persist()
					entryLogMu.Unlock()
				}
				fail := func(stage string, err error, message string) {
					if ctx.Err() != nil {
						control, _, _ := c.currentJobControl(owner, batchID, jobID)
						if control == "paused" || control == "stopped" {
							entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = control, control, "", "自动化已"+map[string]string{"paused": "暂停", "stopped": "停止"}[control], c.now()
							persist()
							return
						}
					}
					// 余额不足/鉴权/参数校验这类确定性错误重试也只会再烧一次付费任务，
					// 必须立即终态失败，跳过整书回补与重新入队。
					if markDeterministicAutomationFailure(entry, stage, err, message, c.now()) {
						failed.Add(1)
						persist()
						return
					}
					// 还有回补预算时把本书重新入队；并发槽位一空出就立即回补，
					// 不再等本批其余书全部跑完。
					if scheduleAutomationBatchEndRetry(entry, stage, err, message, automationNumberValue(entry["batchRetryCount"], 0), c.now()) {
						appendAutomationBookLog(entry, c.now(), stage, message)
						persist()
						enqueueBatchRetry(book)
						return
					}
					entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", stage, err.Error(), message, c.now()
					appendAutomationBookLog(entry, c.now(), stage, message)
					if err != nil && shouldScheduleAutomationTimeoutRetry(stage) && (errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "超时") || strings.Contains(strings.ToLower(err.Error()), "deadline")) {
						retryCount := automationNumberValue(entry["retryCount"], 0)
						if retryCount < 3 {
							entry["retryCount"] = retryCount + 1
							entry["retryRequested"] = true
							entry["retryAt"] = c.now().Add(30 * time.Second)
							entry["message"] = fmt.Sprintf("%s，已安排自动重试（%d/3）", message, retryCount+1)
						}
					}
					failed.Add(1)
					persist()
				}
				entryStatus := automationStringValue(entry["status"], "")
				if entryStatus == "completed" {
					continue
				}
				// A retry is made runnable by resetRetryableAutomationBooks (or by
				// the timeout dispatcher). Keep other terminal books untouched so a
				// retry of one failed book cannot accidentally rerun its neighbours.
				if entryStatus == "failed" || entryStatus == "blocked" || entryStatus == "stopped" {
					continue
				}
				control, controlJob, controlErr := c.currentJobControl(owner, batchID, jobID)
				if controlErr == nil && controlJob != nil {
					if control == "paused" || control == "stopped" {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = control, control, "自动化已"+map[string]string{"paused": "暂停", "stopped": "停止"}[control], c.now()
						persist()
						return
					}
				}
				if ctx.Err() != nil {
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "stopped", "stopped", "测试任务已停止", c.now()
					persist()
					continue
				}
				if entryStatus == "scheduled" && automationStringValue(entry["stage"], "") == "waiting_submit" {
					if c.automationSubmitDeferred(job) {
						continue
					}
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "running", "submit", "到设定时间，正在提交成品", c.now()
					persist()
					submitSettings := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
					submitTimeout := stageTimeout
					if batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch) == "comic" {
						// 漫画发布要先上传 2–35 张图再建合集，N8 回调入库是异步的，
						// 10 分钟阶段超时会把正常发布误判成失败。
						submitTimeout = comicPublishStageTimeout
					}
					publishCtx, cancel := context.WithTimeout(ctx, submitTimeout)
					var submitErr error
					if batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch) == "comic" {
						collection, collectionErr := c.latestComicCollection(publishCtx, owner, batchID, book.ID, automationStringValue(entry["comicCollectionId"], ""))
						if collectionErr != nil {
							submitErr = collectionErr
						} else if c.comicPublish == nil {
							submitErr = fmt.Errorf("N8 漫画发布服务未配置")
						} else {
							_, submitErr = c.comicPublish(publishCtx, owner, batchID, book.ID, collection, submitSettings)
						}
					} else {
						submitErr = c.publishVideo(publishCtx, owner, batchID, book.ID)
					}
					cancel()
					if submitErr != nil {
						if errors.Is(submitErr, errSensitiveReviewPending) {
							// 未确认的敏感词绝不静默上传：停在待确认，不判整本失败。
							entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "pending_confirmation", "sensitive_review", "", submitErr.Error()+"，请先在卡片上确认", c.now()
							appendAutomationBookLog(entry, c.now(), "sensitive_review", submitErr.Error())
							persist()
							continue
						}
						if errors.Is(submitErr, errPublishUnconfirmed) {
							entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", "upload_confirm", submitErr.Error(), "已上传待确认：回读失败，重试时将优先回读", c.now()
							appendAutomationBookLog(entry, c.now(), "upload_confirm", "已上传待确认：回读失败，重试时将优先回读")
							failed.Add(1)
							persist()
							continue
						}
						fail("submit", submitErr, "定时提交失败")
						continue
					}
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "published", "已按定时提交时间完成上传", c.now()
					persist()
					continue
				}
				if reason := c.automationBookPreflightReason(ctx, owner, batch, book, automationStringValue(job["runMode"], "full_submit")); reason != "" {
					entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "blocked", "preflight", "", reason, c.now()
					appendAutomationBookLog(entry, c.now(), "preflight", reason)
					persist()
					continue
				}
				if len(productionLines(book)) == 0 {
					entry["stage"], entry["message"], entry["updatedAt"] = "fetch_original", "正在自动获取缺失正文", c.now()
					persist()
					fetchCtx, cancel := context.WithTimeout(ctx, stageTimeout)
					fetchedBook, fetchErr := c.ensureAutomationSource(fetchCtx, owner, batchID, book)
					cancel()
					if fetchErr != nil {
						fail("fetch_original", fetchErr, "自动获取正文失败")
						continue
					}
					book = fetchedBook
				}
				effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
				if batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch) == "comic" {
					if automationStringValue(job["runMode"], "full_submit") == "full_submit" && automationStringValue(entry["stage"], "") == "comic_publish" && strings.TrimSpace(automationStringValue(entry["comicCollectionId"], "")) != "" {
						entry["message"], entry["updatedAt"] = "正在继续发布已生成漫画合集", c.now()
						persist()
						publishCtx, cancel := context.WithTimeout(ctx, comicPublishStageTimeout)
						collection, collectionErr := c.latestComicCollection(publishCtx, owner, batchID, book.ID, automationStringValue(entry["comicCollectionId"], ""))
						var publishErr error
						var record N8ComicPublishRecord
						if collectionErr != nil {
							publishErr = collectionErr
						} else if c.comicPublish == nil {
							publishErr = fmt.Errorf("N8 漫画发布服务未配置")
						} else {
							record, publishErr = c.comicPublish(publishCtx, owner, batchID, book.ID, collection, effective)
						}
						cancel()
						if publishErr != nil || record.Status != "published" {
							// 只复用已上传 imageIDs、继续建合集；不能当成发布完成。
							if publishErr == nil {
								publishErr = errN8ComicCarouselIncomplete
							}
							entry["publishState"] = "upload_incomplete"
							if len(record.ImageIDs) > 0 {
								entry["comicImageIDs"] = record.ImageIDs
							}
							fail("comic_publish", publishErr, "漫画图片已上传，正在继续创建合集")
							continue
						}
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "published", "漫画合集已发布", c.now()
						delete(entry, "retryRequested")
						delete(entry, "retryAt")
						delete(entry, "retryReason")
						delete(entry, "publishState")
						entry["comicCarouselId"], entry["comicPublishName"] = record.CarouselID, record.Name
						entry["comicImageIDs"] = record.ImageIDs
						persist()
						continue
					}
					acknowledgeComicRetry(entry)
					switch selectComicPreparationAction(book, effective) {
					case comicPreparationBootstrap:
						identityIssue := batchfactoryv11.ComicAssetIdentityIssue(book)
						identityRecovery := identityIssue != ""
						if !comicAssetsReady(book) || identityRecovery {
							message := "正在提取漫画人物、场景和道具资产"
							assetMode := batchfactoryv11.StageModeMissing
							if identityRecovery {
								message = fmt.Sprintf("检测到漫画资产身份不完整（%s），正在自动重新提取人物身份", identityIssue)
								assetMode = batchfactoryv11.StageModeForce
							}
							entry["stage"], entry["message"], entry["updatedAt"] = "comic_assets", message, c.now()
							persist()
							stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
							_, assetErr := c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageAssets, assetMode, jobID+"-comic-assets-"+book.ID, "")
							cancel()
							if assetErr != nil {
								fail("comic_assets", assetErr, "漫画资产提取失败")
								continue
							}
							if identityRecovery {
								latestBatch, readErr := c.store.GetBatch(ctx, owner, batchID)
								if readErr != nil {
									fail("comic_assets", readErr, "自动修复漫画资产后读取结果失败")
									continue
								}
								refreshed := false
								for _, candidate := range latestBatch.Books {
									if candidate.ID == book.ID {
										book = candidate
										refreshed = true
										break
									}
								}
								if !refreshed {
									fail("comic_assets", batchfactoryv11.ErrNotFound, "自动修复漫画资产后找不到当前作品")
									continue
								}
								if remainingIssue := batchfactoryv11.ComicAssetIdentityIssue(book); remainingIssue != "" {
									fail("comic_assets", fmt.Errorf("漫画资产身份仍不完整：%s", remainingIssue), "自动修复漫画资产失败，已停止继续生图")
									continue
								}
							}
						}
						entry["stage"], entry["message"], entry["updatedAt"] = "comic_bootstrap", "正在分析原文、分页并生成每页漫画提示词", c.now()
						persist()
						stageCtx, cancel := context.WithTimeout(ctx, comicPromptStageTimeout)
						// 每页提示词都是一次独立的文本请求，串行会把一本书压在
						// 10 分钟超时线上，并把整批拖成“跑一早上”。
						stageCtx = batchfactoryv11.WithComicPromptConcurrency(stageCtx, comicPromptConcurrency)
						var bootstrapProgressMu sync.Mutex
						stageCtx = batchfactoryv11.WithComicPromptProgress(stageCtx, func(completed, total int, pageLabel string) {
							bootstrapProgressMu.Lock()
							defer bootstrapProgressMu.Unlock()
							if completed == 0 {
								return
							}
							entry["message"] = fmt.Sprintf("漫画页提示词 %d/%d 已完成，最近完成：%s", completed, total, pageLabel)
							entry["updatedAt"] = c.now()
							persist()
						})
						_, bootstrapErr := c.stages.RunComicBootstrap(stageCtx, owner, batchID, book.ID)
						cancel()
						if bootstrapErr != nil {
							fail("comic_bootstrap", bootstrapErr, "漫画原文分页与提示词提取失败")
							continue
						}
						if identityRecovery {
							entry["stage"], entry["message"], entry["updatedAt"] = "comic_prompt_recovery", "人物身份已修复，正在重新生成全部漫画页提示词", c.now()
							persist()
							stageCtx, cancel := context.WithTimeout(ctx, comicPromptStageTimeout)
							stageCtx = batchfactoryv11.WithComicPromptConcurrency(stageCtx, comicPromptConcurrency)
							_, promptErr := c.stages.Director.RunComicPromptExtraction(stageCtx, owner, batchID, book.ID)
							cancel()
							if promptErr != nil {
								fail("comic_prompt_recovery", promptErr, "漫画身份修复后的提示词重新生成失败")
								continue
							}
						}
					case comicPreparationMissingPrompts:
						entry["stage"], entry["message"], entry["updatedAt"] = "comic_prompt_recovery", "正在补齐缺失的漫画页提示词", c.now()
						persist()
						stageCtx, cancel := context.WithTimeout(ctx, comicPromptStageTimeout)
						stageCtx = batchfactoryv11.WithComicPromptConcurrency(stageCtx, comicPromptConcurrency)
						var promptProgressMu sync.Mutex
						stageCtx = batchfactoryv11.WithComicPromptProgress(stageCtx, func(completed, total int, pageLabel string) {
							promptProgressMu.Lock()
							defer promptProgressMu.Unlock()
							if completed == 0 {
								entry["message"] = fmt.Sprintf("正在并发生成漫画页提示词（共 %d 张）", total)
							} else {
								entry["message"] = fmt.Sprintf("漫画页提示词 %d/%d 已完成，最近完成：%s", completed, total, pageLabel)
							}
							entry["stage"], entry["updatedAt"] = "comic_prompt_recovery", c.now()
							persist()
						})
						_, recoveryErr := c.stages.RunMissingComicPromptExtraction(stageCtx, owner, batchID, book.ID)
						cancel()
						if recoveryErr != nil {
							fail("comic_prompt_recovery", recoveryErr, "漫画缺页提示词补齐失败")
							continue
						}
					case comicPreparationReady:
						entry["stage"], entry["message"], entry["updatedAt"] = "comic_resume", "已复用完整漫画资产和分页，继续补齐图片", c.now()
						persist()
					}
					if automationStringValue(job["runMode"], "full_submit") == "storyboard_only" {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "ready_for_images", "已完成漫画资产、分页与每页提示词", c.now()
						persist()
						continue
					}
					if c.comicGenerate == nil {
						fail("comic_images", fmt.Errorf("漫画图片生成服务未配置"), "整本漫画出图失败")
						continue
					}
					entry["stage"], entry["message"], entry["updatedAt"] = "comic_images", "正在逐页生成漫画并整理合集", c.now()
					persist()
					generationCtx, lease, generationCancel := comicGenerationLeaseContext(ctx, comicGenerationBudget(comicGenerationPageCount(book), 4), comicGenerationHardCap)
					generationLease = lease
					collection, generateErr := c.generateComicWithRetry(generationCtx, entry, persist, func(attemptCtx context.Context) (batchfactoryv11.ComicCollection, error) {
						// 每次 attempt 不再使用固定 60 分钟：attempt 继承整本的
						// 心跳租约，只有整本真正停滞时才被取消。
						generateCtx, cancel := context.WithCancel(attemptCtx)
						defer cancel()
						generateCtx = withComicGenerationProgress(generateCtx, recordComicEvent)
						generateCtx = withShuihuoImageEvent(generateCtx, recordComicEvent)
						generateCtx = withComicAssetConcurrency(generateCtx, 5)
						return c.comicGenerate(generateCtx, owner, batchID, book.ID)
					}, nil)
					generationCancel()
					if generateErr != nil {
						failureStage, failureDetail := comicGenerationFailureStage(generateErr)
						// 缺页如果只是“仍在生成”或“lease 已过期”，不是内容审核失败：
						// 不消耗内容审核重试预算，也不使用“两次降级后仍被拒绝”的文案。
						var shortage *comicPageShortageError
						if errors.As(generateErr, &shortage) && shortage.Total > 0 && shortage.Failed == 0 {
							attempts := automationNumberValue(entry["inflightRetryCount"], 0)
							if attempts < comicInFlightRetryLimit {
								inflightMessage := fmt.Sprintf("漫画页仍在生成中，等待后立即回补（%d/%d）", attempts+1, comicInFlightRetryLimit)
								entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", "comic_images", inflightMessage, generateErr.Error(), c.now()
								entry["inflightRetryCount"] = attempts + 1
								appendAutomationBookLog(entry, c.now(), "comic_images", inflightMessage)
								delete(entry, "retryAt")
								persist()
								enqueueBatchRetry(book)
								continue
							}
						}
						// 被内容审核拒绝的资产图会跳过，页会被跳过，合集/资产校验
						// 因此报缺图。这里不把整本打成永久失败，而是按退避时间自动
						// 重试缺失内容，其它书照常继续。
						if comicMissingPageError(generateErr) || comicMissingAssetError(generateErr) {
							if c.scheduleComicContentPolicyRetry(entry, persist, failureStage) {
								continue
							}
							fail(failureStage, generateErr, fmt.Sprintf("被内容审核拒绝的漫画资产图/漫画页自动重试 %d 次仍未通过，请调整对应提示词后手动重试", comicContentPolicyRetryLimit))
							continue
						}
						fail(failureStage, generateErr, failureDetail)
						continue
					}
					confirmCtx, cancelConfirm := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
					confirmedCollection, confirmErr := c.confirmComicCollection(confirmCtx, owner, batchID, book.ID, collection)
					cancelConfirm()
					if confirmErr != nil {
						fail("comic_images", confirmErr, "漫画图片已返回，但合集回填确认失败")
						continue
					}
					collection = confirmedCollection
					entry["comicCollectionId"], entry["comicPageCount"] = collection.ID, len(collection.Items)
					if automationStringValue(job["runMode"], "full_submit") == "full_submit" {
						if c.automationSubmitDeferred(job) {
							entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "scheduled", "waiting_submit", "漫画合集已提前完成，等待定时提交", c.now()
							persist()
							continue
						}
						if c.comicPublish == nil {
							fail("comic_publish", fmt.Errorf("N8 漫画发布服务未配置"), "漫画合集已完成，但自动发布不可用")
							continue
						}
						entry["stage"], entry["message"], entry["updatedAt"] = "comic_publish", "正在发布漫画合集", c.now()
						persist()
						publishCtx, cancel := context.WithTimeout(ctx, comicPublishStageTimeout)
						record, publishErr := c.comicPublish(publishCtx, owner, batchID, book.ID, collection, batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch))
						cancel()
						if publishErr != nil {
							if n8ImagesAlreadyUploadedError(publishErr.Error()) {
								entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "uploaded", "图片已上传完成，N8 已确认图片已占用；已停止重复上传", c.now()
								entry["publishError"], entry["error"], entry["retryRequested"] = publishErr.Error(), "", false
								delete(entry, "retryAt")
								persist()
								continue
							}
						}
						if publishErr != nil || record.Status != "published" {
							// 图片已上传但合集没建起来：这是可续建状态，不能映射成
							// completed。放回回补队列，下一轮只复用 imageIDs 建合集。
							if publishErr == nil {
								publishErr = errN8ComicCarouselIncomplete
							}
							entry["publishState"] = "upload_incomplete"
							if len(record.ImageIDs) > 0 {
								entry["comicImageIDs"] = record.ImageIDs
							}
							fail("comic_publish", publishErr, "漫画图片已上传，正在继续创建合集")
							continue
						}
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "published", "漫画合集已发布", c.now()
						delete(entry, "retryRequested")
						delete(entry, "retryAt")
						delete(entry, "retryReason")
						delete(entry, "publishState")
						entry["comicCarouselId"], entry["comicPublishName"] = record.CarouselID, record.Name
						entry["comicImageIDs"] = record.ImageIDs
						persist()
						continue
					}
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "ready_for_upload", "漫画合集已完成，等待上传", c.now()
					persist()
					continue
				}
				entry["stage"] = "audio_planning"
				entry["message"] = "正在逐行生成配音并测量时长"
				entry["updatedAt"] = c.now()
				persist()
				stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
				ttsArtifactID, audioDurationSeconds, stageErr := c.ensureAudioPlan(stageCtx, owner, batchID, book)
				cancel()
				if stageErr != nil {
					fail("audio_planning", stageErr, "配音与逐行时长失败")
					continue
				}
				production := batchfactoryv11.ProductionJob{}
				reusedProduction := false
				forceRegenerate := automationBoolValue(job["forceRegenerate"])
				// 视频阶段完成口径按 ordinal 统计，允许复用 orphaned 历史成片：
				// 只要每个分镜都已有 succeeded + mediaUrl，就不再提交新一轮视频任务。
				if !forceRegenerate && automationStringValue(job["runMode"], "full_submit") != "storyboard_only" {
					status, statusErr := c.stages.Production.GetBatchStatus(ctx, owner, batchID)
					if statusErr != nil {
						fail("video", statusErr, "读取已生成视频失败")
						continue
					}
					production, reusedProduction = automationSucceededProduction(status.Jobs, batch, book)
				}
				delete(entry, "retryRequested")
				if reusedProduction {
					entry["stage"], entry["message"], entry["updatedAt"] = "video", "视频已全部生成，跳过生成，继续合成", c.now()
					appendAutomationBookLog(entry, c.now(), "video", "视频已全部生成，跳过生成，继续合成")
					persist()
				} else {
					// 默认 Missing：已完成的阶段一律跳过，只补未完成的。
					// 只有用户在界面上显式勾选“重新生成”时才允许 force；
					// 重试/回补路径由各阶段的完成判定决定，不会再自动升级为 force。
					stageMode := batchfactoryv11.StageModeMissing
					if forceRegenerate {
						stageMode = batchfactoryv11.StageModeForce
					}
					if batchfactoryv11.AssetImagesEnabled(effective) {
						if automationAssetsComplete(book, effective) && !forceRegenerate {
							const skipMessage = "资产已生成，跳过资产提取"
							entry["stage"], entry["message"], entry["updatedAt"] = "assets", skipMessage, c.now()
							appendAutomationBookLog(entry, c.now(), "assets", skipMessage)
							persist()
						} else {
							entry["stage"] = "assets"
							entry["message"] = "正在提取资产"
							entry["updatedAt"] = c.now()
							stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
							_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageAssets, stageMode, jobID+"-assets-"+book.ID, "")
							cancel()
							if stageErr != nil {
								fail("assets", stageErr, "资产提取失败")
								continue
							}
						}
					} else {
						// 显式跳过，不能静默：日志里写清原因，后续分镜/视频/合成照常推进。
						const skipMessage = "因资产图片开关关闭，已跳过资产生成阶段"
						entry["stage"], entry["message"], entry["updatedAt"] = "assets", skipMessage, c.now()
						appendAutomationBookLog(entry, c.now(), "assets", skipMessage)
						persist()
					}
					if forceRegenerate || !automationDirectorComplete(book) {
						entry["stage"], entry["message"] = "director", "正在生成分镜提示词"
						entry["updatedAt"] = c.now()
						persist()
						directorMode := stageMode
						if !forceRegenerate && book.DirectorRevision != nil {
							// 正文/设置已变化：旧分镜不算完成，必须重建；Missing 会因为
							// 已存在 revision 直接跳过，所以这里只能显式 Force。
							directorMode = batchfactoryv11.StageModeForce
						}
						stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
						_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageDirector, directorMode, jobID+"-director-"+book.ID, "")
						cancel()
						if stageErr != nil {
							fail("director", stageErr, "分镜提示词生成失败")
							continue
						}
					} else {
						const skipMessage = "分镜提示词已完成且正文未变，跳过分镜生成"
						entry["stage"], entry["message"], entry["updatedAt"] = "director", skipMessage, c.now()
						appendAutomationBookLog(entry, c.now(), "director", skipMessage)
						persist()
					}
					if automationStringValue(job["runMode"], "full_submit") == "storyboard_only" {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "ready_for_video", "已完成资产与分镜提示词", c.now()
						persist()
						continue
					}
					if automationOpeningEnabled(batch, book) && (forceRegenerate || !batchfactoryv11.OpeningStageComplete(batch, book)) {
						entry["stage"], entry["message"], entry["updatedAt"] = "opening", "正在基于分镜提示词生成换开头", c.now()
						persist()
						stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
						_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageOpening, stageMode, jobID+"-opening-"+book.ID, "")
						cancel()
						if stageErr != nil {
							fail("opening", stageErr, "换开头提示词生成失败")
							continue
						}
					}
					entry["stage"], entry["message"], entry["updatedAt"] = "video", "正在提交视频生成", c.now()
					persist()
					stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
					_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageVideo, batchfactoryv11.StageModeMissing, videoRequestID(book.ID, automationNumberValue(entry["batchRetryCount"], 0)), "")
					cancel()
					if stageErr != nil {
						fail("video", stageErr, "视频提交失败")
						continue
					}
					stageCtx, cancel = context.WithTimeout(ctx, videoWaitTimeout)
					completedProduction, waitErr := c.waitProduction(stageCtx, owner, batchID, book.ID, videoRequestID(book.ID, automationNumberValue(entry["batchRetryCount"], 0)), videoWaitIdleTimeout)
					cancel()
					if waitErr != nil {
						if errors.Is(waitErr, errProductionInFlight) {
							// 供应商侧还在跑：不能整书回补（会重建视频记录、消耗回补预算）。
							entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "running", "video", "", "视频仍在生成，已超过等待上限，稍后可重新回读", c.now()
							appendAutomationBookLog(entry, c.now(), "video", "视频仍在生成，已超过等待上限，稍后可重新回读")
							persist()
							continue
						}
						fail("video", waitErr, "视频生成失败")
						continue
					}
					production = completedProduction
					confirmCtx, confirmCancel := context.WithTimeout(ctx, 30*time.Second)
					confirmErr := c.confirmProductionReadback(confirmCtx, owner, batchID, book.ID, production)
					confirmCancel()
					if confirmErr != nil {
						fail("video", confirmErr, "视频已返回，但成片回填确认失败")
						continue
					}
				}
				if automationStringValue(job["runMode"], "full_submit") == "video_no_submit" {
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "ready_for_upload", "视频已完成，等待上传", c.now()
					entry["productionTasks"] = len(production.Tasks)
					persist()
					continue
				}
				if c.merge == nil || !c.merge.Enabled {
					fail("merge", fmt.Errorf("合成服务未启用"), "最终合成失败")
					continue
				}
				// 本轮视频全部复用已完成成片时，已有成功合成就直接跳过；
				// 只要本轮提交过新视频，就必须重新合成，避免复用带旧视频的合成结果。
				if reusedProduction && c.automationMergeComplete(ctx, owner, batchID, batch, book) {
					const skipMessage = "合成成片已存在，跳过合成"
					entry["stage"], entry["message"], entry["updatedAt"] = "merge", skipMessage, c.now()
					appendAutomationBookLog(entry, c.now(), "merge", skipMessage)
					persist()
				} else {
					entry["stage"], entry["message"], entry["updatedAt"] = "merge", "正在合成最终视频", c.now()
					persist()
					mergeCtx, mergeCancel := context.WithTimeout(ctx, stageTimeout)
					audioSource := "tts"
					if raw, ok := book.SettingsState.Patch["audioSource"]; ok {
						_ = json.Unmarshal(raw, &audioSource)
					}
					mergeJobs, mergeErr := c.merge.SubmitBookMerge(mergeCtx, owner, batchID, book.ID, jobID+"-merge-"+book.ID, batchfactoryv11.MergeOptions{TimingMode: "audio", AudioSource: strings.ToLower(strings.TrimSpace(audioSource)), AudioArtifactID: ttsArtifactID, AudioDurationSeconds: audioDurationSeconds})
					if mergeErr == nil {
						_, mergeErr = c.waitMerge(mergeCtx, owner, batchID, book.ID, mergeJobs)
					}
					mergeCancel()
					if mergeErr != nil {
						fail("merge", mergeErr, "最终合成失败")
						continue
					}
				}
				if automationBoolValue(job["autoPublish"]) {
					if c.automationSubmitDeferred(job) {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "scheduled", "waiting_submit", "视频已提前生成并合成，等待定时提交", c.now()
						persist()
						continue
					}
					if automationBookUploaded(book) {
						const skipMessage = "已上传完成，跳过重复上传"
						entry["stage"], entry["message"], entry["updatedAt"] = "upload", skipMessage, c.now()
						appendAutomationBookLog(entry, c.now(), "upload", skipMessage)
						persist()
					} else {
						entry["stage"], entry["message"], entry["updatedAt"] = "upload", "正在上传视频管理系统", c.now()
						persist()
						publishCtx, publishCancel := context.WithTimeout(ctx, stageTimeout)
						publishErr := c.publishVideo(publishCtx, owner, batchID, book.ID)
						publishCancel()
						if publishErr != nil {
							if errors.Is(publishErr, errSensitiveReviewPending) {
								entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "pending_confirmation", "sensitive_review", "", publishErr.Error()+"，请先在卡片上确认", c.now()
								appendAutomationBookLog(entry, c.now(), "sensitive_review", publishErr.Error())
								persist()
								continue
							}
							if errors.Is(publishErr, errPublishUnconfirmed) {
								entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", "upload_confirm", publishErr.Error(), "已上传待确认：回读失败，重试时将优先回读", c.now()
								appendAutomationBookLog(entry, c.now(), "upload_confirm", "已上传待确认：回读失败，重试时将优先回读")
								failed.Add(1)
								persist()
								continue
							}
							fail("upload", publishErr, "上传失败")
							continue
						}
					}
				}
				entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "completed", "已完成视频生成、合成与上传", c.now()
				entry["productionTasks"] = len(production.Tasks)
				persist()
			}
		}(book)
	}
	workers.Wait()
	if control, _, controlErr := c.currentJobControl(owner, batchID, jobID); controlErr == nil && (control == "paused" || control == "stopped") {
		return
	}
	// Recount from persisted entries rather than only this goroutine's local
	// failures. A retry run intentionally skips unrelated terminal books; those
	// books must remain failed in the aggregate instead of being reported ready.
	failedCount := 0
	readyCount := 0
	autoRetryPending := false
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "failed", "blocked", "stopped":
			failedCount++
		case "completed":
			readyCount++
		}
		if automationBoolValue(entry["retryRequested"]) && strings.TrimSpace(automationStringValue(entry["retryAt"], "")) != "" {
			autoRetryPending = true
		}
	}
	if autoRetryPending {
		job["state"], job["updatedAt"] = "running", c.now()
		job["counts"] = automationCounts(len(batch.Books), 0, failedCount, readyCount)
		_ = c.setJob(owner, batchID, jobID, job)
		return
	}
	pendingCount := 0
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry != nil && (automationStringValue(entry["status"], "") == "scheduled" || automationStringValue(entry["status"], "") == "pending_confirmation") {
			pendingCount++
		}
	}
	if failedCount == 0 && pendingCount > 0 {
		job["state"], job["updatedAt"] = "scheduled", c.now()
		job["counts"] = automationCountsWithPending(len(batch.Books), pendingCount)
		_ = c.setJob(owner, batchID, jobID, job)
		return
	}
	job["state"] = "completed"
	job["completedAt"], job["updatedAt"] = c.now(), c.now()
	job["counts"] = automationCounts(len(batch.Books), 0, failedCount, readyCount)
	if failedCount > 0 {
		job["state"] = "failed"
	}
	_ = c.setJob(owner, batchID, jobID, job)
}

func cloneAutomationEntry(entry map[string]any) map[string]any {
	cloned := make(map[string]any, len(entry))
	for key, value := range entry {
		cloned[key] = value
	}
	return cloned
}

// automationBookLogLimit 限制单本书保留的执行日志条数，避免状态文件无限增长。
const automationBookLogLimit = 200

// appendAutomationBookLog 把一条阶段事件追加到单本书的执行日志里。
// 同一句重复上报只更新时间，不重复占位，避免进度刷屏。
func appendAutomationBookLog(entry map[string]any, at time.Time, stage, message string) {
	if entry == nil {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	logs := make([]any, 0, 8)
	switch typed := entry["logs"].(type) {
	case []any:
		logs = append(logs, typed...)
	case []map[string]any:
		for _, item := range typed {
			logs = append(logs, item)
		}
	}
	if len(logs) > 0 {
		if last, ok := logs[len(logs)-1].(map[string]any); ok && automationStringValue(last["message"], "") == message {
			last["at"] = at
			entry["logs"] = logs
			return
		}
	}
	logs = append(logs, map[string]any{"at": at, "stage": stage, "message": message})
	if len(logs) > automationBookLogLimit {
		logs = logs[len(logs)-automationBookLogLimit:]
	}
	entry["logs"] = logs
}

func (c *BatchAutomationCoordinator) currentJobControl(owner, batchID, jobID string) (string, map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return "", nil, err
	}
	job, ok := state.Jobs[jobID]
	if !ok || automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
		return "", nil, nil
	}
	return automationStringValue(job["state"], ""), job, nil
}

func (c *BatchAutomationCoordinator) waitProduction(ctx context.Context, owner, batchID, bookID, requestID string, idleTimeout time.Duration) (batchfactoryv11.ProductionJob, error) {
	if c.stages == nil || c.stages.Production == nil {
		return batchfactoryv11.ProductionJob{}, fmt.Errorf("视频生产服务未配置")
	}
	return waitProductionStatus(ctx, func(ctx context.Context) (batchfactoryv11.BatchStatus, error) {
		return c.stages.Production.GetBatchStatus(ctx, owner, batchID)
	}, bookID, requestID, idleTimeout)
}

// waitProductionStatus 是视频等待的核心循环。fetch 由调用方注入，便于测试。
func waitProductionStatus(ctx context.Context, fetch func(context.Context) (batchfactoryv11.BatchStatus, error), bookID, requestID string, idleTimeout time.Duration) (batchfactoryv11.ProductionJob, error) {
	if idleTimeout <= 0 {
		idleTimeout = videoWaitIdleTimeout
	}
	// progress 记录任务状态签名：只要任务在推进（新任务出现、状态变化、成片回填），
	// 就重置空闲计时，避免正常的长时间排队被当成失败。
	progress := ""
	job, found := batchfactoryv11.ProductionJob{}, false
	idle := time.NewTimer(idleTimeout)
	defer idle.Stop()
	for {
		status, err := fetch(ctx)
		if err != nil {
			if ctx.Err() == nil {
				return batchfactoryv11.ProductionJob{}, err
			}
			return finalizeProductionWait(fetch, bookID, requestID, job, found)
		}
		if current, ok := productionJobForRequest(status.Jobs, bookID, requestID); ok {
			job, found = current, true
			if productionJobFailed(current) {
				return current, fmt.Errorf("视频任务失败：%s", productionJobError(current))
			}
			if len(current.Tasks) > 0 && productionJobComplete(current) {
				return current, nil
			}
			if signature := productionProgressSignature(current); signature != progress {
				progress = signature
				if !idle.Stop() {
					select {
					case <-idle.C:
					default:
					}
				}
				idle.Reset(idleTimeout)
			}
		}
		select {
		case <-ctx.Done():
			return finalizeProductionWait(fetch, bookID, requestID, job, found)
		case <-idle.C:
			// 空闲超时但任务仍在跑，同样交给调用方按“仍在生成”处理。
			return finalizeProductionWait(fetch, bookID, requestID, job, found)
		case <-time.After(2 * time.Second):
		}
	}
}

// finalizeProductionWait 在等待失败时做一次终态回读：完成就按成功返回；
// 仍有在途任务时返回 errProductionInFlight，避免触发整书回补。
func finalizeProductionWait(fetch func(context.Context) (batchfactoryv11.BatchStatus, error), bookID, requestID string, last batchfactoryv11.ProductionJob, found bool) (batchfactoryv11.ProductionJob, error) {
	finalCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if status, err := fetch(finalCtx); err == nil {
		if current, ok := productionJobForRequest(status.Jobs, bookID, requestID); ok {
			if len(current.Tasks) > 0 && productionJobComplete(current) {
				return current, nil
			}
			if productionJobFailed(current) {
				return current, fmt.Errorf("视频任务失败：%s", productionJobError(current))
			}
			last, found = current, true
		}
	}
	if found && len(last.Tasks) > 0 {
		// 供应商侧仍有任务在跑：调用方不得整书回补。
		return last, errProductionInFlight
	}
	return batchfactoryv11.ProductionJob{}, fmt.Errorf("视频生成超时或已停止")
}

// productionProgressSignature 汇总任务的“状态 + 是否已有成片”，用于判断这一轮
// 是否还在推进。
func productionProgressSignature(job batchfactoryv11.ProductionJob) string {
	parts := make([]string, 0, len(job.Tasks))
	for _, task := range job.Tasks {
		media := "0"
		if strings.TrimSpace(task.MediaURL) != "" {
			media = "1"
		}
		parts = append(parts, fmt.Sprintf("%s:%s", task.Status, media))
	}
	return strings.Join(parts, ",")
}

func (c *BatchAutomationCoordinator) confirmProductionReadback(ctx context.Context, owner, batchID, bookID string, expected batchfactoryv11.ProductionJob) error {
	if c == nil || c.stages == nil || c.stages.Production == nil {
		return fmt.Errorf("视频状态回读服务未配置")
	}
	status, err := c.stages.Production.GetBatchStatus(ctx, owner, batchID)
	if err != nil {
		return err
	}
	job, ok := productionJobForRequest(status.Jobs, bookID, expected.RequestID)
	if !ok || !productionJobComplete(job) {
		return fmt.Errorf("视频管理状态尚未回填完整成片")
	}
	return nil
}

func (c *BatchAutomationCoordinator) confirmComicCollection(ctx context.Context, owner, batchID, bookID string, expected batchfactoryv11.ComicCollection) (batchfactoryv11.ComicCollection, error) {
	if strings.TrimSpace(expected.ID) == "" {
		return expected, fmt.Errorf("漫画生成服务未返回合集编号")
	}
	// Keep compatibility with older in-process generators used by integrations
	// that return only an ID and page references. The current persisted comic
	// generator always returns Status=ready; those results go through the full
	// collection read-back check below.
	if strings.TrimSpace(expected.Status) == "" {
		return expected, nil
	}
	confirmed, err := c.latestComicCollection(ctx, owner, batchID, bookID, expected.ID)
	if err != nil {
		return expected, err
	}
	if confirmed.ID != expected.ID || len(confirmed.Items) < len(expected.Items) {
		return expected, fmt.Errorf("漫画合集回填不完整：返回 %d 页，预期至少 %d 页", len(confirmed.Items), len(expected.Items))
	}
	for _, item := range confirmed.Items {
		if strings.TrimSpace(item.URL) == "" && strings.TrimSpace(item.StorageRef) == "" && strings.TrimSpace(item.ImageID) == "" {
			return expected, fmt.Errorf("漫画合集仍有页面没有主图")
		}
	}
	return confirmed, nil
}

func productionJobForRequest(jobs []batchfactoryv11.ProductionJob, bookID, requestID string) (batchfactoryv11.ProductionJob, bool) {
	for _, job := range jobs {
		if job.BookID == bookID && job.RequestID == requestID {
			return job, true
		}
	}
	return batchfactoryv11.ProductionJob{}, false
}

func productionJobComplete(job batchfactoryv11.ProductionJob) bool {
	for _, task := range job.Tasks {
		if task.Status != batchfactoryv11.ProductionSucceeded || strings.TrimSpace(task.MediaURL) == "" {
			return false
		}
	}
	return true
}

func reusableAutomationProduction(jobs []batchfactoryv11.ProductionJob, book batchfactoryv11.Book) (batchfactoryv11.ProductionJob, bool) {
	if book.DirectorRevision == nil || strings.TrimSpace(book.DirectorRevision.ID) == "" {
		return batchfactoryv11.ProductionJob{}, false
	}
	var selected batchfactoryv11.ProductionJob
	found := false
	for _, job := range jobs {
		if job.BookID != book.ID || job.DirectorRevisionID != book.DirectorRevision.ID || job.Status != batchfactoryv11.ProductionSucceeded || len(job.Tasks) == 0 || !productionJobComplete(job) {
			continue
		}
		if !found || job.UpdatedAt.After(selected.UpdatedAt) {
			selected, found = job, true
		}
	}
	return selected, found
}

func productionJobFailed(job batchfactoryv11.ProductionJob) bool {
	for _, task := range job.Tasks {
		if task.Status == batchfactoryv11.ProductionFailed || task.Status == batchfactoryv11.ProductionCancelled {
			return true
		}
	}
	return false
}

func productionJobError(job batchfactoryv11.ProductionJob) string {
	for _, task := range job.Tasks {
		if strings.TrimSpace(task.ErrorMessage) != "" {
			return task.ErrorMessage
		}
	}
	return "视频模型未返回成片"
}

// automationAssetsComplete 资产阶段完成口径：开关关闭视为完成；开关打开时
// 只要书里已有可用的资产条目就视为完成，不再重复调用模型提取。
func automationAssetsComplete(book batchfactoryv11.Book, effective batchfactoryv11.SettingsPatch) bool {
	if !batchfactoryv11.AssetImagesEnabled(effective) {
		return true
	}
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Prompt) != "" {
			return true
		}
	}
	return false
}

// automationDirectorComplete 分镜阶段完成口径：存在分镜，且其输入摘要与
// 当前正文一致。摘要变化说明正文被改过，必须重新生成，不能拿旧分镜硬跑。
func automationDirectorComplete(book batchfactoryv11.Book) bool {
	revision := book.DirectorRevision
	if revision == nil || strings.TrimSpace(revision.ID) == "" || len(book.Videos) == 0 {
		return false
	}
	return revision.SourceDigest == batchfactoryv11.SourceInputDigest(book.SourceText)
}

// automationSucceededProduction 汇总本书所有分镜的成功成片（含 orphaned 历史
// 槽位），并按 ordinal 判断视频阶段是否已经完成。返回的任务集合用于后续统计，
// 避免“只做了两轮、没有一条 job 全部成功”时又把整轮视频重新提交一遍。
//
// 完成口径：每个 active 分镜 ordinal 至少有 1 条 succeeded + 可访问 mediaUrl；
// 换开头开启时，VIDEO01 的每个成功变体也要有对应成片。
func automationSucceededProduction(jobs []batchfactoryv11.ProductionJob, batch batchfactoryv11.Batch, book batchfactoryv11.Book) (batchfactoryv11.ProductionJob, bool) {
	if len(book.Videos) == 0 {
		return batchfactoryv11.ProductionJob{}, false
	}
	ordinalByVideo := make(map[string]int, len(book.Videos)+len(book.VideoHistory))
	for _, video := range book.Videos {
		ordinalByVideo[video.ID] = video.Ordinal
	}
	for _, history := range book.VideoHistory {
		ordinalByVideo[history.ID] = history.Ordinal
	}
	done := map[[2]int]bool{}
	production := batchfactoryv11.ProductionJob{BookID: book.ID}
	for _, job := range jobs {
		if job.BookID != book.ID {
			continue
		}
		if production.ID == "" && len(job.Tasks) > 0 {
			production.ID, production.DirectorRevisionID = job.ID, job.DirectorRevisionID
		}
		for _, task := range job.Tasks {
			if task.Status != batchfactoryv11.ProductionSucceeded || strings.TrimSpace(task.MediaURL) == "" {
				continue
			}
			ordinal, ok := ordinalByVideo[task.VideoID]
			if !ok {
				continue
			}
			done[[2]int{ordinal, task.OpeningVariantIndex}] = true
			production.Tasks = append(production.Tasks, task)
		}
	}
	for _, video := range book.Videos {
		if !done[[2]int{video.Ordinal, 0}] {
			return production, false
		}
	}
	for _, index := range batchfactoryv11.RequiredOpeningVariantIndices(batch, book) {
		if index == 0 {
			continue
		}
		if !done[[2]int{book.Videos[0].Ordinal, index}] {
			return production, false
		}
	}
	return production, len(production.Tasks) > 0
}

// automationMergeComplete 合成阶段完成口径：本书当前需要的每个开头版本
// 都已有 succeeded 且有成片地址的合成任务。
func (c *BatchAutomationCoordinator) automationMergeComplete(ctx context.Context, owner, batchID string, batch batchfactoryv11.Batch, book batchfactoryv11.Book) bool {
	if c.merge == nil || !c.merge.Enabled {
		return false
	}
	jobs, err := c.merge.GetBatchStatus(ctx, owner, batchID)
	if err != nil {
		return false
	}
	done := map[int]bool{}
	for _, job := range jobs {
		if job.BookID != book.ID || job.Status != batchfactoryv11.MergeSucceeded || strings.TrimSpace(job.OutputURL) == "" {
			continue
		}
		done[job.OpeningVariantIndex] = true
	}
	for _, index := range batchfactoryv11.OpeningMergeIndices(batch, book) {
		if !done[index] {
			return false
		}
	}
	return true
}

// automationBookUploaded 判断上传阶段是否已有服务端回执。已上传的书不再进入
// 上传流程：目标站登录态失效时也不会被误判成上传失败。
func automationBookUploaded(book batchfactoryv11.Book) bool {
	status := strings.ToLower(strings.TrimSpace(automationStringValue(book.SourceMetadata["websiteSubmitStatus"], "")))
	if status == "" {
		status = strings.ToLower(strings.TrimSpace(automationStringValue(book.SourceMetadata["publishStatus"], "")))
	}
	return status == "uploaded"
}

func (c *BatchAutomationCoordinator) waitMerge(ctx context.Context, owner, batchID, bookID string, submitted []batchfactoryv11.MergeJob) ([]batchfactoryv11.MergeJob, error) {
	wanted := map[string]bool{}
	for _, job := range submitted {
		wanted[job.ID] = true
	}
	for {
		jobs, err := c.merge.GetBatchStatus(ctx, owner, batchID)
		if err != nil {
			return nil, err
		}
		found := 0
		for _, job := range jobs {
			if job.BookID != bookID || !wanted[job.ID] {
				continue
			}
			found++
			if job.Status == batchfactoryv11.MergeFailed {
				return jobs, fmt.Errorf("合成任务失败：%s", job.ErrorMessage)
			}
		}
		if found == len(wanted) && found > 0 {
			complete := true
			for _, job := range jobs {
				if job.BookID == bookID && wanted[job.ID] && job.Status != batchfactoryv11.MergeSucceeded {
					complete = false
				}
			}
			if complete {
				return jobs, nil
			}
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("最终合成超时或已停止")
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *BatchAutomationCoordinator) Cancel(owner, batchID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return err
	}
	stoppedIDs := make([]string, 0)
	for id, job := range state.Jobs {
		if automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
			continue
		}
		stateValue := automationStringValue(job["state"], "")
		if stateValue != "running" && stateValue != "paused" && stateValue != "scheduled" {
			continue
		}
		job["state"], job["updatedAt"], job["stopAt"] = "stopped", c.now(), c.now()
		job["lastError"] = "用户停止自动化"
		job["counts"] = automationCounts(0, 0, 0, 0)
		state.Jobs[id] = job
		stoppedIDs = append(stoppedIDs, id)
	}
	if err := c.saveLocked(state); err != nil {
		return err
	}
	for _, id := range stoppedIDs {
		c.cancelRun(id)
	}
	return nil
}

func (c *BatchAutomationCoordinator) Pause(owner, batchID string) (map[string]any, error) {
	return c.control(owner, batchID, "paused", func(job map[string]any) error {
		c.cancelRun(automationStringValue(job["id"], ""))
		return nil
	})
}

func (c *BatchAutomationCoordinator) Resume(ctx context.Context, owner, batchID string) (map[string]any, error) {
	return c.control(owner, batchID, "running", func(job map[string]any) error {
		batch, err := c.store.GetBatch(ctx, owner, batchID)
		if err != nil {
			return err
		}
		if c.providerSync != nil {
			if err := c.providerSync(ctx, owner); err != nil {
				return err
			}
		}
		c.launchRun(owner, batchID, automationStringValue(job["id"], ""), job, batch)
		return nil
	})
}

func (c *BatchAutomationCoordinator) Retry(ctx context.Context, owner, batchID string, bookIDs []string) (map[string]any, error) {
	set := map[string]bool{}
	for _, id := range bookIDs {
		if strings.TrimSpace(id) != "" {
			set[id] = true
		}
	}
	c.mu.Lock()
	state, err := c.loadLocked()
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}
	// Retry is also the recovery entry point after a run has reached a terminal
	// state (failed/completed/stopped). control() only matched active jobs, so
	// retrying a finished run returned 404 even though the workbench button was
	// available. Reopen the most recent terminal job instead.
	target, reopened := pickAutomationControlJob(state, owner, batchID)
	if target == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: 没有可控制的自动化任务", batchfactoryv11.ErrNotFound)
	}
	target["state"], target["updatedAt"] = "running", c.now()
	if reopened {
		// Terminal markers must not make the status reader expose the reopened
		// run as already finished.
		delete(target, "completedAt")
		delete(target, "stopAt")
		target["reopenedAt"] = c.now()
		// A reopened run must not reuse the previous run's stage request ids.
		// SubmitBookProductionWithOptions is idempotent per request id, so a
		// repeated "-video-<book>" id returned the OLD failed production job
		// (and its old provider/error) without ever resubmitting. The attempt
		// counter scopes the video request id to this retry.
		target["retryAttempt"] = automationNumberValue(target["retryAttempt"], 0) + 1
	}
	books, _ := target["books"].(map[string]any)
	tracked := make(map[string]bool, len(books))
	for id := range books {
		tracked[id] = true
	}
	resetRetryableAutomationBooks(books, set, c.now())
	if reopened {
		var running, failed, ready int
		for _, raw := range books {
			entry, _ := raw.(map[string]any)
			switch automationStringValue(entry["status"], "") {
			case "completed":
				ready++
			case "running":
				running++
			default:
				failed++
			}
		}
		target["counts"] = automationCounts(len(books), running, failed, ready)
	}
	if err := c.saveLocked(state); err != nil {
		c.mu.Unlock()
		return nil, err
	}
	c.mu.Unlock()
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	if c.providerSync != nil {
		if err := c.providerSync(ctx, owner); err != nil {
			return nil, err
		}
	}
	// run() indexes one job entry per batch book and would panic on books this
	// job never tracked (a batch run can cover a subset). Relaunch against the
	// tracked set only; completed entries are skipped inside run().
	scoped := batch
	scoped.Books = make([]batchfactoryv11.Book, 0, len(tracked))
	for _, book := range batch.Books {
		if tracked[book.ID] {
			scoped.Books = append(scoped.Books, book)
		}
	}
	c.launchRun(owner, batchID, automationStringValue(target["id"], ""), target, scoped)
	return publicAutomationJob(target), nil
}

// pickAutomationControlJob returns the active job (running/paused/
// needs_attention) for the batch. When no active job exists it falls back to
// the most recent terminal job (failed/completed/stopped) so retry can reopen
// it. The boolean is true when the returned job is a reopened terminal job.
func pickAutomationControlJob(state automationStateFile, owner, batchID string) (map[string]any, bool) {
	var terminal map[string]any
	var terminalAt time.Time
	for _, candidate := range state.Jobs {
		if automationStringValue(candidate["owner"], "") != owner || automationStringValue(candidate["batchId"], "") != batchID {
			continue
		}
		switch automationStringValue(candidate["state"], "") {
		case "running", "paused", "scheduled", "needs_attention":
			return candidate, false
		case "failed", "completed", "stopped":
		default:
			continue
		}
		at := automationJobTime(candidate)
		if terminal == nil || at.After(terminalAt) {
			terminal, terminalAt = candidate, at
		}
	}
	if terminal != nil {
		return terminal, true
	}
	return nil, false
}

// resetRetryableAutomationBooks moves failed/blocked/paused entries back to the
// running assets stage. When only is non-empty, just that subset is reset.
func resetRetryableAutomationBooks(books map[string]any, only map[string]bool, now time.Time) {
	for id, raw := range books {
		if len(only) > 0 && !only[id] {
			continue
		}
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		status := automationStringValue(entry["status"], "")
		if status == "failed" || status == "blocked" || status == "paused" || status == "stopped" {
			stage, message := "assets", "重新执行失败阶段"
			if automationStringValue(entry["stage"], "") == "comic_publish" && strings.TrimSpace(automationStringValue(entry["comicCollectionId"], "")) != "" {
				stage, message = "comic_publish", "重新发布已生成漫画合集"
			}
			entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", stage, message, "", now
			entry["retryRequested"] = true
			// 人工重试视为重新授权一轮自动重试预算；否则 3/3 用尽后的重试
			// 会被 scheduleComicContentPolicyRetry 直接判死，按钮等于失效。
			delete(entry, "contentPolicyRetryCount")
			delete(entry, "retryAt")
			delete(entry, "retryReason")
			// 人工重试同样重置“批次末尾回补”的预算与等待标记。
			delete(entry, "batchRetryCount")
			delete(entry, "batchRetryPending")
			// 人工重试是重新从 assets 开始，不能沿用上一轮回补的续跑点。
			delete(entry, "resumeStage")
		}
	}
}

func (c *BatchAutomationCoordinator) control(owner, batchID, nextState string, after func(map[string]any) error) (map[string]any, error) {
	c.mu.Lock()
	state, err := c.loadLocked()
	if err != nil {
		c.mu.Unlock()
		return nil, err
	}
	var target map[string]any
	for _, candidate := range state.Jobs {
		if automationStringValue(candidate["owner"], "") != owner || automationStringValue(candidate["batchId"], "") != batchID {
			continue
		}
		current := automationStringValue(candidate["state"], "")
		if current == "running" || current == "paused" || current == "needs_attention" {
			target = candidate
			break
		}
	}
	if target == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: 没有可控制的自动化任务", batchfactoryv11.ErrNotFound)
	}
	target["state"], target["updatedAt"] = nextState, c.now()
	if err := c.saveLocked(state); err != nil {
		c.mu.Unlock()
		return nil, err
	}
	c.mu.Unlock()
	if after != nil {
		if err := after(target); err != nil {
			return nil, err
		}
		if err := c.setJob(owner, batchID, automationStringValue(target["id"], ""), target); err != nil {
			return nil, err
		}
	}
	return publicAutomationJob(target), nil
}

func registerBatchAutomationRoutes(mux *http.ServeMux, auth BridgeAuth, coordinator *BatchAutomationCoordinator, version string) {
	if mux == nil || coordinator == nil {
		return
	}
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation/start", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		result, err := coordinator.Start(r.Context(), owner, r.PathValue("batchId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"automation": result})
	})))
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/books/{bookId}/audio-plan", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			RequestID string `json:"requestId"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		result, err := coordinator.GenerateBookAudioPlan(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if requestID := strings.TrimSpace(input.RequestID); requestID != "" {
			result["requestId"] = requestID
		}
		writeJSON(w, http.StatusOK, result)
	})))
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation/cancel", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if err := coordinator.Cancel(owner, r.PathValue("batchId")); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})))
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation/pause", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		result, err := coordinator.Pause(owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"automation": result})
	})))
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation/resume", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		result, err := coordinator.Resume(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"automation": result})
	})))
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation/retry", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			BookIDs []string `json:"bookIds"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		result, err := coordinator.Retry(r.Context(), owner, r.PathValue("batchId"), input.BookIDs)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"automation": result})
	})))
}
