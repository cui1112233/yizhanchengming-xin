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
	"slices"
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
	giantMaterialResolver GiantMaterialResolver
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
	comicPublishRecord    func(context.Context, string, string, string) (N8ComicPublishRecord, bool)
	comicPublishPreflight func(context.Context, string, string, string, batchfactoryv11.SettingsPatch) error
	mu                    sync.Mutex
	controlMu             sync.Mutex
	batchRunMu            sync.Mutex
	batchRunLocks         map[string]*automationBatchRunLock
	runCancels            map[string]context.CancelFunc
	// runEpochs 记录每个 job 当前有效的执行轮次。重试会在同一 jobID 上启动
	// 新 run 并取消旧 run；旧 run 只能看到自己的 epoch，不能再写 job 终态、
	// counts 或把书标成失败/停止，避免两轮 run 互相覆盖。
	runEpochs            map[string]int64
	dispatchNextCheck    time.Time
	dispatchNextCheckSet bool
	dispatcherOnce       sync.Once
	reaperOnce           sync.Once
	sweepOnce            sync.Once
	// comicRunSlots 限制同时进入“资产+页图”出图的书籍数量。出图是最贵的上游
	// 环节，多本同时各开多页并发会把请求全堵在队列里，每本都缺几页、谁也无法
	// 提交。改成按本串行后，一本出完并发布，下一本才开始出图。
	comicRunSlots chan struct{}
}

// comicSlotStallTimeout 是出图槽位的停滞上限。持槽书本超过这个时间没有任何
// 真实进展（上游不回包、页图不落库），就主动中止并释放槽位，让队列里的下一本
// 先出图；本书按失败处理、可以直接重试，不会让整批一直卡在同一本上。
// 排队等待、响应头返回和正文传输都会刷新心跳，正常慢图不会被误判。
const comicSlotStallTimeout = 10 * time.Minute

// comicBookGenerationConcurrency 默认 3：同时在出图的书本数。设成 1 虽然
// 单本收口最快，但通道经常处于半空转（一本书在跑提示词时没人出图），总吞吐
// 明显下降；3 本可以填满通道且仍保持“一本一本地完成并发布”。可用
// QIANTIE_COMIC_BOOK_CONCURRENCY 调整，必须小于通道真实并发能力。
func comicBookGenerationConcurrency() int {
	const fallback = 3
	raw := strings.TrimSpace(os.Getenv("QIANTIE_COMIC_BOOK_CONCURRENCY"))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 8 {
		return fallback
	}
	return value
}

// automationBatchRunLock only serializes lifecycle transitions for one batch.
// A single global start mutex made an unrelated slow Start/provider sync block
// every completed job from writing its terminal state, leaving the UI in the
// impossible state "0 running books, job still running".
type automationBatchRunLock struct {
	mu   sync.Mutex
	refs int
}

func (c *BatchAutomationCoordinator) lockBatchRun(owner, batchID string) func() {
	key := strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(batchID)
	c.batchRunMu.Lock()
	if c.batchRunLocks == nil {
		c.batchRunLocks = map[string]*automationBatchRunLock{}
	}
	lock := c.batchRunLocks[key]
	if lock == nil {
		lock = &automationBatchRunLock{}
		c.batchRunLocks[key] = lock
	}
	lock.refs++
	c.batchRunMu.Unlock()

	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		c.batchRunMu.Lock()
		lock.refs--
		if lock.refs == 0 && c.batchRunLocks[key] == lock {
			delete(c.batchRunLocks, key)
		}
		c.batchRunMu.Unlock()
	}
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
	return &BatchAutomationCoordinator{store: store, presets: presets, stages: stages, merge: merge, external: publisher, ttsStore: ttsStore, audioProbe: audioProbe, path: path, now: now, providerSync: syncProvider, comicRunSlots: make(chan struct{}, comicBookGenerationConcurrency())}
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
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							log.Printf("batch automation dispatcher panic recovered: %v", recovered)
						}
					}()
					_ = c.DispatchDueJobs(context.Background())
				}()
			}
		}()
	})
	c.StartAutomationReaper()
	c.StartComicPageLeaseSweeper()
}

// comicPageLeaseReclaimAfter 是页级 generating 租约的回收阈值。
// 单页 detached 生成上限 15 分钟，这里留 5 分钟余量；刚更新过的页不会命中。
const comicPageLeaseReclaimAfter = 20 * time.Minute

// comicPageGenerationSweepInterval 是回收任务的扫描周期。
const comicPageGenerationSweepInterval = 5 * time.Minute

// StartComicPageLeaseSweeper 启动“卡死页”回收：启动时先扫一次，之后每 5 分钟
// 扫一次，把停在 generating 且超过 20 分钟没更新的页转成 generation_failed。
// 进程崩溃或请求被丢弃时不会再有人触碰这些页，必须由后台主动回收。
func (c *BatchAutomationCoordinator) StartComicPageLeaseSweeper() {
	if c == nil {
		return
	}
	c.sweepOnce.Do(func() {
		go func() {
			c.sweepStaleComicPagesOnce()
			ticker := time.NewTicker(comicPageGenerationSweepInterval)
			defer ticker.Stop()
			for range ticker.C {
				c.sweepStaleComicPagesOnce()
			}
		}()
	})
}

func (c *BatchAutomationCoordinator) sweepStaleComicPagesOnce() {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("comic page lease sweeper panic recovered: %v", recovered)
		}
	}()
	store, ok := c.store.(batchfactoryv11.StaleComicPageStore)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	before := time.Now().UTC().Add(-comicPageLeaseReclaimAfter)
	// 重启前留下的 generating 页不可能再有人写回：旧进程的请求已经全部作废，
	// 回收租约的写回操作本身也随进程结束而丢失。不立刻回收，每本书都要空等
	// 20 分钟才能补这几页，表现就是“一直缺几页、永远不发布”。
	if automationProcessStartedAt.After(before) {
		before = automationProcessStartedAt
	}
	pages, err := store.ListStaleGeneratingComicPages(ctx, before, 500)
	if err != nil {
		log.Printf("comic page lease sweep list failed: %v", err)
		return
	}
	for _, page := range pages {
		expired, err := store.ExpireStaleGeneratingComicPage(ctx, page, before)
		if err != nil {
			log.Printf("comic page lease sweep expire failed book=%s video=%s: %v", page.BookID, page.VideoID, err)
			continue
		}
		if expired {
			log.Printf("comic page lease reclaimed reason=超时未回收 owner=%s batch=%s book=%s video=%s stale_since=%s", page.Owner, page.BatchID, page.BookID, page.VideoID, page.UpdatedAt.Format(time.RFC3339))
		}
	}
}

// StartAutomationReaper 启动独立的僵尸回收循环。它不读 dispatchNextCheck，
// 因此不会被任何远期 scheduledAt/retryAt 饿死。
func (c *BatchAutomationCoordinator) StartAutomationReaper() {
	if c == nil {
		return
	}
	c.reaperOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							log.Printf("batch automation reaper panic recovered: %v", recovered)
						}
					}()
					if err := c.ReapZombieJobs(context.Background()); err != nil {
						log.Printf("batch automation reaper failed: %v", err)
					}
				}()
			}
		}()
	})
}

// ReapZombieJobs 独立回收僵尸 running job。判定不依赖 dispatcher 调度门禁，
// 也不依赖是否存在排队书籍。
func (c *BatchAutomationCoordinator) ReapZombieJobs(ctx context.Context) error {
	if c == nil || c.store == nil {
		return nil
	}
	now := c.now()
	c.mu.Lock()
	state, err := c.loadLocked()
	if err != nil {
		c.mu.Unlock()
		return err
	}
	reclaimed := make([]string, 0)
	for jobID, job := range state.Jobs {
		if reclaimAutomationZombieRunningJob(job, jobID, now) {
			reclaimed = append(reclaimed, jobID)
		}
	}
	if len(reclaimed) == 0 {
		c.mu.Unlock()
		return nil
	}
	if err := c.saveLocked(state); err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()
	for _, jobID := range reclaimed {
		c.cancelRun(jobID)
	}
	return nil
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
	reclaimed := make([]string, 0)
	for jobID, job := range state.Jobs {
		if reclaimAutomationZombieRunningJob(job, jobID, now) {
			reclaimed = append(reclaimed, jobID)
			changed = true
			continue
		}
		if automationJobRetryDispatchDue(job, now) && !automationJobHasActiveBooks(job) {
			books, _ := job["books"].(map[string]any)
			retryDue := false
			for _, raw := range books {
				entry, _ := raw.(map[string]any)
				if entry == nil || !automationBoolValue(entry["retryRequested"]) || !automationRetryDue(entry, now) {
					continue
				}
				if reopenAutomationRetryBook(entry, now) {
					retryDue = true
				}
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
	for _, jobID := range reclaimed {
		c.cancelRun(jobID)
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
			c.blockAutomationBooksForJob(item.job, validationErr.Error(), c.now())
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

// automationRetryDue 必须同时接受内存中的 time.Time 与落库后的字符串。
// 只用字符串解析会让刚写入 30 秒续跑的 retryAt 在进程内永远判不出“已到点”，
// 表现为到点后既不重跑也不报错。
func automationRetryDue(entry map[string]any, now time.Time) bool {
	when := automationStoredJobTime(entry["retryAt"])
	return !when.IsZero() && !now.Before(when)
}

// automationBookRunVersionField 标记一本书当前由哪一轮 run 权威持有。
// 调度器重新打开书时递增，旧 run 的迟到写入因此自动失效。
const automationBookRunVersionField = "runVersion"

// reopenAutomationRetryBook 是书级自动重跑的唯一权威入口。
// 返回值表示是否真正重新打开。needs_manual 的书只清理重试标记，不重新激活。
func reopenAutomationRetryBook(entry map[string]any, now time.Time) bool {
	if entry == nil {
		return false
	}
	if automationBoolValue(entry["needsManual"]) {
		delete(entry, "retryRequested")
		delete(entry, "retryAt")
		entry["status"], entry["updatedAt"] = "failed", now
		entry["message"] = "需要人工处理，已停止自动重试"
		return false
	}
	stage := automationStringValue(entry["stage"], "")
	if automationStringValue(entry["retryReason"], "") == "content_policy" {
		if stage == "" {
			stage = "comic_images"
		}
		entry["message"] = "到达重试时间，正在重新生成被内容审核拒绝的内容"
		delete(entry, "retryReason")
	} else {
		if stage == "" {
			stage = "assets"
		}
		entry["message"] = "到达重试时间，继续未完成的页/节点"
	}
	entry["status"], entry["stage"], entry["updatedAt"] = "running", stage, now
	delete(entry, "retryAt")
	delete(entry, "retryRequested")
	entry[automationBookRunVersionField] = automationNumberValue(entry[automationBookRunVersionField], 0) + 1
	return true
}

// automationPendingBookRetry 判断 job 内是否还有已排期的书级续跑。
// retryAt 在内存中是 time.Time、落库后才是字符串，必须走统一解析入口，
// 否则 run 收尾会把「已有续跑计划」的 job 提前写成 failed。
func automationPendingBookRetry(books map[string]any) bool {
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		if automationBoolValue(entry["retryRequested"]) && !automationStoredJobTime(entry["retryAt"]).IsZero() {
			return true
		}
	}
	return false
}

// automationHasDueRetry 判断 job 内是否有已到点的书级续跑。
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

// automationJobHasActiveBooks 判断 job 内是否还有正在执行或排队等槽位的书。
// 有活跃书时不能启动书级续跑：launchRun 会取消当前 run，正在出图/视频的书
// 会被中断并重新提交，既拖慢其它书也会重复扣费。等这一轮全部收尾后 run()
// 会保持 job 为 running，再由调度器派发待续跑的书。
func automationJobHasActiveBooks(job map[string]any) bool {
	books, _ := job["books"].(map[string]any)
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "running", "queued":
			return true
		}
	}
	return false
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
// 整批其余书结束后串行回补，避免失败书一直占着并发槽位；3 轮仍不成功就停止
// 自动重试，把是否继续交给用户决定。
const automationBatchRetryRounds = 3

// automationZombieRunningIdleTimeout 是回收“僵尸 running”job 的空闲阈值。
// 回收只依赖「job 仍标记 running + 超过该时长没有任何状态写入」，不再要求
// 必须先没有任何 running 书：书级停滞由独立的停滞看门狗处理，僵尸回收
// 负责兜住整个执行轮次已经消失、但书还挂着 running 的情况。
// 比书级停滞阈值多留 2 分钟，让看门狗先落「跳过这本书」的可读状态。
const automationZombieRunningIdleTimeout = automationStallTimeout + 2*time.Minute

const automationZombieRunningMessage = "执行轮次已失效（长时间无进展）"

// automationZombieRunningReclaimable 判断 job 是否是执行轮次中断后残留的
// “running” 假状态：只要长时间没有任何状态写入即可回收，不再看书的在途状态。
func automationZombieRunningReclaimable(job map[string]any, now time.Time) bool {
	if job == nil || automationStringValue(job["state"], "") != "running" {
		return false
	}
	last := automationStoredJobTime(job["updatedAt"])
	return !last.IsZero() && now.Sub(last) >= automationZombieRunningIdleTimeout
}

// reclaimAutomationZombieRunningJob 把僵尸 running 落成 failed 终态，并按终态
// 语义清掉残留的重试标记，避免 dispatcher 再把一个已经没人执行的 job 重新派发。
func reclaimAutomationZombieRunningJob(job map[string]any, jobID string, now time.Time) bool {
	if !automationZombieRunningReclaimable(job, now) {
		return false
	}
	books, _ := job["books"].(map[string]any)
	// 执行轮次已经消失：还挂着 running/queued 的书不能再留在“进行中”，
	// 否则前端永远显示在跑，调度器也无法判断这本书需要重试。
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "running", "queued":
			entry["status"] = "failed"
			entry["stage"] = automationStringValue(entry["stage"], "unknown")
			entry["message"] = automationZombieRunningMessage
			entry["error"] = "执行轮次已失效：超过 10 分钟没有任何进展"
			entry["updatedAt"] = now
		}
	}
	job["state"] = "failed"
	job["lastError"] = automationZombieRunningMessage
	job["completedAt"], job["updatedAt"] = now, now
	job["counts"] = automationCountsFromBooks(books)
	job["needsManualPages"] = automationNeedsManualPages(books)
	clearAutomationRetryMarks(job)
	appendAutomationJobLog(job, now, automationZombieRunningMessage)
	log.Printf("automation: reclaim zombie running job %s owner=%s batch=%s", jobID,
		automationStringValue(job["owner"], ""), automationStringValue(job["batchId"], ""))
	return true
}

// clearAutomationRetryMarks 清掉任务与书上的残留重试标记。终态（stopped/failed）
// 的任务不应再被 retryAt 自动重新派发；只清标记，不改业务状态与错误信息。
func clearAutomationRetryMarks(job map[string]any) {
	if job == nil {
		return
	}
	delete(job, "retryRequested")
	delete(job, "retryAt")
	books, _ := job["books"].(map[string]any)
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		delete(entry, "retryRequested")
		delete(entry, "retryAt")
	}
}

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

// automationResumeShouldRunStage keeps continuation retries monotonic: a
// failure at video may continue into merge/upload, but must never call assets,
// director or opening again. An empty resume point is a normal first run.
func automationResumeShouldRunStage(resumeStage, stage string) bool {
	ranks := map[string]int{"assets": 1, "director": 2, "opening": 3, "video": 4}
	resumeRank := ranks[strings.TrimSpace(resumeStage)]
	stageRank := ranks[strings.TrimSpace(stage)]
	if resumeRank == 0 || stageRank == 0 {
		return true
	}
	return stageRank >= resumeRank
}

// automationResumeStageForBook prevents a retry that previously failed at
// video/merge/upload from skipping an opening prompt that is still missing.
// It only moves the resume point backwards to opening; completed assets and
// director work remain reused.
func automationResumeStageForBook(resumeStage string, batch batchfactoryv11.Batch, book batchfactoryv11.Book) string {
	if strings.TrimSpace(resumeStage) != "" && automationOpeningEnabled(batch, book) && !batchfactoryv11.OpeningStageComplete(batch, book) && !automationResumeShouldRunStage(resumeStage, "opening") {
		return "opening"
	}
	return resumeStage
}

// deterministicAutomationFailure 识别同一请求重试也不会成功的上游错误：
// 401/402/403/400/404/422、余额不足/欠费文案、参数或配置错误。超时、429
// 和 5xx 仍是瞬时错误，必须继续保留回补能力。
func deterministicAutomationFailure(err error) bool {
	if err == nil || errors.Is(err, errProductionInFlight) || errors.Is(err, batchfactoryv11.ErrConflict) {
		return false
	}
	var sourceInputErr *batchFetchUserError
	if errors.As(err, &sourceInputErr) {
		return true
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
	// 余额不足 / 鉴权失败这类确定性错误只能人工修复后重试，标记为需人工处理，
	// 防止残留重试标记把同一本书再跑一遍。
	entry["needsManual"] = true
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

func (c *BatchAutomationCoordinator) SetGiantMaterialResolver(resolver GiantMaterialResolver) {
	if c != nil {
		c.giantMaterialResolver = resolver
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

func (c *BatchAutomationCoordinator) SetComicPublishRecordReader(reader func(context.Context, string, string, string) (N8ComicPublishRecord, bool)) {
	if c != nil {
		c.comicPublishRecord = reader
	}
}

func (c *BatchAutomationCoordinator) SetComicPublishPreflight(preflight func(context.Context, string, string, string, batchfactoryv11.SettingsPatch) error) {
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
	return c.setJobLocked(owner, batchID, jobID, job)
}

func (c *BatchAutomationCoordinator) setJobLocked(owner, batchID, jobID string, job map[string]any) error {
	// 终态任务不再保留自动重试标记：否则残留的 retryRequested/retryAt 会让
	// DispatchDueJobs 把一个已经结束的 job 重新派发成 running。
	if state := automationStringValue(job["state"], ""); state == "stopped" || state == "failed" {
		clearAutomationRetryMarks(job)
	}
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

// persistAutomationBookEntry 是 worker 写书的唯一入口。它在一个 c.mu 临界区内
// 完成「读最新状态 → 校验 runVersion → 合并 → 落库」，避免旧 run 的迟到写入
// 把调度器刚写入的 running 覆盖回 failed（最后写入者胜）。
func (c *BatchAutomationCoordinator) persistAutomationBookEntry(owner, batchID, jobID, bookID string, target, entry map[string]any, entryVersion int, fallbackJob map[string]any) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return false, err
	}
	job := state.Jobs[jobID]
	if job == nil {
		// 没有持久化状态（内存 run / 测试）时，退回本轮 job 自身的 map。
		job = fallbackJob
	}
	if job == nil {
		return false, nil
	}
	books, _ := job["books"].(map[string]any)
	if books == nil {
		return false, nil
	}
	latest, _ := books[bookID].(map[string]any)
	if latest != nil && automationNumberValue(latest[automationBookRunVersionField], 0) > entryVersion {
		// 该书的权威状态已被调度器或新一轮 run 接管，旧 worker 不得写回。
		return false, nil
	}
	if latest == nil {
		latest = target
	}
	if latest == nil {
		return false, nil
	}
	// 原地合并，保持既有 map 引用（状态读取方与测试都持有旧引用）。
	clear(latest)
	for key, value := range entry {
		latest[key] = value
	}
	if target != nil {
		clear(target)
		for key, value := range entry {
			target[key] = value
		}
	}
	books[bookID] = latest
	job["counts"] = automationCountsFromBooks(books)
	// 批次级“需人工页清单”：跨书汇总，供列表/详情直接展示。
	job["needsManualPages"] = automationNeedsManualPages(books)
	job["updatedAt"] = c.now()
	if err := c.setJobLocked(owner, batchID, jobID, job); err != nil {
		return false, err
	}
	return true, nil
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
		clearAutomationRetryMarks(job)
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

// blockAutomationBooksForJob 在整批预检失败时把仍有运行标记的书改为 blocked。
// 没有可用执行器时它们不可能继续推进；若继续保留 running，线上会出现
// “任务已 failed、书还在跑”的假状态。
func (c *BatchAutomationCoordinator) blockAutomationBooksForJob(job map[string]any, reason string, now time.Time) {
	if job == nil || strings.TrimSpace(reason) == "" {
		return
	}
	books, _ := job["books"].(map[string]any)
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "completed", "failed", "blocked", "stopped":
			continue
		}
		entry["status"], entry["error"], entry["message"], entry["updatedAt"] = "blocked", "", reason, now
		entry["retryRequested"] = false
		appendAutomationBookLog(entry, now, "preflight", reason)
	}
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
			c.blockAutomationBooksForJob(item.job, validationErr.Error(), c.now())
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
	unlockBatchRun := c.lockBatchRun(owner, batchID)
	defer unlockBatchRun()
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
	if automationBoolValue(input["queueIfActive"]) {
		if queued, found, queueErr := c.queueBooksBehindActiveAutomation(owner, batchID, targetBooks, input); found || queueErr != nil {
			return queued, queueErr
		}
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
	// 同一批次只允许一个执行任务：先把旧的 running/scheduled/queued 停掉，
	// 否则两个任务会同时调度同一本书，重复出图、重复扣费。
	if err := c.supersedeActiveJobsForBatch(owner, batchID, jobID); err != nil {
		return nil, err
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

// supersedeActiveJobsForBatch 保证同一批次同一时刻只有一个执行任务在跑。
// 重复点“执行”或前端重试时，旧任务必须立刻停掉，否则两个任务会各自调度
// 同一本书、各自出图，造成重复扣费。
func (c *BatchAutomationCoordinator) supersedeActiveJobsForBatch(owner, batchID, keepJobID string) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return err
	}
	now := c.now()
	changed := false
	for jobID, job := range state.Jobs {
		if jobID == keepJobID || job == nil {
			continue
		}
		if automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
			continue
		}
		switch automationStringValue(job["state"], "") {
		case "running", "scheduled", "queued":
		default:
			continue
		}
		job["state"] = "stopped"
		job["message"] = "已被新的执行任务接管"
		job["updatedAt"] = now
		delete(job, "retryAt")
		if books, ok := job["books"].(map[string]any); ok {
			for _, raw := range books {
				entry, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				switch automationStringValue(entry["status"], "") {
				case "running", "queued", "scheduled":
					entry["status"] = "stopped"
					entry["message"] = "已被新的执行任务接管"
					entry["updatedAt"] = now
				}
			}
		}
		if cancel, ok := c.runCancels[jobID]; ok && cancel != nil {
			cancel()
			delete(c.runCancels, jobID)
		}
		c.runEpochs[jobID]++
		changed = true
	}
	if !changed {
		return nil
	}
	return c.saveLocked(state)
}

const automationFollowUpStage = "waiting_active_run"

// queueBooksBehindActiveAutomation 把用户明确选择的新书追加到同一任务的后续队列。
// 它只写 queued 条目，不取消、不重启当前 run，因此当前正在调用模型的书不会被
// 二次提交；本轮结束后 run() 再领取这些条目继续执行。
func (c *BatchAutomationCoordinator) queueBooksBehindActiveAutomation(owner, batchID string, targetBooks []batchfactoryv11.Book, input map[string]any) (map[string]any, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return nil, false, err
	}
	var active map[string]any
	for _, candidate := range state.Jobs {
		if automationStringValue(candidate["owner"], "") != owner || automationStringValue(candidate["batchId"], "") != batchID {
			continue
		}
		switch automationStringValue(candidate["state"], "") {
		case "running", "paused", "scheduled", "needs_attention":
			active = candidate
		}
		if active != nil {
			break
		}
	}
	if active == nil {
		return nil, false, nil
	}
	requestedPresetID := automationStringValue(input["presetId"], "")
	activePreset, _ := active["preset"].(map[string]any)
	activePresetID := automationStringValue(activePreset["id"], "")
	if requestedPresetID != "" && activePresetID != "" && requestedPresetID != activePresetID {
		return nil, true, fmt.Errorf("%w: 当前任务使用预设 %s；新增书必须沿用当前预设", batchfactoryv11.ErrConflict, automationStringValue(activePreset["name"], activePresetID))
	}
	books, _ := active["books"].(map[string]any)
	if books == nil {
		books = map[string]any{}
		active["books"] = books
	}
	added := 0
	for _, book := range targetBooks {
		if _, tracked := books[book.ID]; tracked {
			continue
		}
		books[book.ID] = map[string]any{
			"bookId": book.ID, "title": book.Title,
			"status": "queued", "stage": automationFollowUpStage,
			"message": "当前任务完成后自动继续", "attempts": 0, "retryCount": 0,
			"updatedAt": c.now(),
		}
		added++
	}
	if added == 0 {
		return nil, true, fmt.Errorf("%w: 选择的小说已在当前任务中，不会重复执行", batchfactoryv11.ErrConflict)
	}
	active["counts"] = automationCountsFromBooks(books)
	active["updatedAt"] = c.now()
	if err := c.saveLocked(state); err != nil {
		return nil, true, err
	}
	return publicAutomationJob(active), true, nil
}

// claimAutomationFollowUpBooks 在当前一轮全部到达终态后，原子领取运行期间追加的
// queued 条目。返回的 job 是持久化的最新快照，包含当前轮结果和后续书。
func (c *BatchAutomationCoordinator) claimAutomationFollowUpBooks(owner, batchID, jobID string) (map[string]any, []string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return nil, nil, err
	}
	job := state.Jobs[jobID]
	if job == nil || automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
		return nil, nil, nil
	}
	books, _ := job["books"].(map[string]any)
	queuedIDs := make([]string, 0)
	for id, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil || automationStringValue(entry["status"], "") != "queued" || automationStringValue(entry["stage"], "") != automationFollowUpStage {
			continue
		}
		entry["stage"], entry["message"], entry["updatedAt"] = "queued", "排队中，等待并发槽位", c.now()
		queuedIDs = append(queuedIDs, id)
	}
	if len(queuedIDs) == 0 {
		return nil, nil, nil
	}
	job["state"], job["updatedAt"] = "running", c.now()
	job["counts"] = automationCountsFromBooks(books)
	if err := c.saveLocked(state); err != nil {
		return nil, nil, err
	}
	return cloneAutomationMap(job), queuedIDs, nil
}

func (c *BatchAutomationCoordinator) activeAutomationJob(owner, batchID string) map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return nil
	}
	now := time.Now()
	if c.now != nil {
		now = c.now()
	}
	if reconcileTerminalAutomationJobs(&state, now) {
		if err := c.saveLocked(state); err != nil {
			return nil
		}
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
	issues := c.automationGlobalPreflightIssues(ctx, owner, batch, runMode)
	if len(issues) == 0 {
		return nil
	}
	unique := make([]string, 0, len(issues))
	seen := map[string]bool{}
	for _, issue := range issues {
		if issue.Message != "" && !seen[issue.Message] {
			seen[issue.Message] = true
			unique = append(unique, issue.Message)
		}
	}
	return fmt.Errorf("%w: 自动生产前检查失败：%s", batchfactoryv11.ErrInvalid, strings.Join(unique, "；"))
}

// automationBookPreflightReason 逐书预检，返回空串表示该书可执行。定时任务与
// 立即任务都走 run()，因此同一套逐书预检对两种启动方式一致生效。
func (c *BatchAutomationCoordinator) automationBookPreflightReason(ctx context.Context, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, runMode string) string {
	if issues := c.automationBookPreflightIssues(ctx, owner, batch, book, runMode, false); len(issues) > 0 {
		return issues[0].Message
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
	// Worker 只能修改自己的深拷贝，不能直接修改 c.automationState 里的共享 map。
	// 否则 Retry/状态接口在整理公开响应时会与 worker 并发读写同一张 map，
	// Go 会以 "concurrent map read and map write" 直接终止整个进程。
	runJob := cloneAutomationMap(job)
	baseCtx := context.WithValue(context.Background(), bridgeIdentityKey{}, BridgeIdentity{Username: owner})
	automationCtx, cancel := context.WithCancel(baseCtx)
	c.controlMu.Lock()
	if c.runCancels == nil {
		c.runCancels = map[string]context.CancelFunc{}
	}
	if previous := c.runCancels[jobID]; previous != nil {
		previous()
	}
	if c.runEpochs == nil {
		c.runEpochs = map[string]int64{}
	}
	c.runEpochs[jobID]++
	epoch := c.runEpochs[jobID]
	c.runCancels[jobID] = cancel
	c.controlMu.Unlock()
	go func(job map[string]any) {
		// 单个自动化任务的 goroutine panic 不允许带走整个进程：落成 failed，
		// 保留 panic 文本，便于用户看到并重试。
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("batch automation job %s panicked: %v", jobID, recovered)
				if !c.isCurrentRun(jobID, epoch) {
					return
				}
				job["state"], job["lastError"], job["updatedAt"] = "failed", fmt.Sprintf("panic: %v", recovered), c.now()
				_ = c.setJob(owner, batchID, jobID, job)
			}
		}()
		c.run(automationCtx, epoch, owner, batchID, jobID, job, batch)
	}(runJob)
}

func (c *BatchAutomationCoordinator) isCurrentRun(jobID string, epoch int64) bool {
	c.controlMu.Lock()
	defer c.controlMu.Unlock()
	return c.runEpochs[jobID] == epoch
}

func (c *BatchAutomationCoordinator) runSuperseded(jobID string, epoch int64) bool {
	c.controlMu.Lock()
	defer c.controlMu.Unlock()
	return c.runEpochs[jobID] > epoch
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

// Merge submissions are idempotent by request id too. Scope retries exactly
// like video submissions so a failed merge is truly submitted again instead
// of returning the previous failed record.
func automationMergeRequestID(jobID, bookID string, retryAttempt, batchRetry int) string {
	if retryAttempt > 0 || batchRetry > 0 {
		return fmt.Sprintf("%s-merge-%s-r%db%d", jobID, bookID, retryAttempt, batchRetry)
	}
	return jobID + "-merge-" + bookID
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

// publishedComicRecord 返回该书最新合集对应的已发布记录，用于避免重复发布。
func (c *BatchAutomationCoordinator) publishedComicRecord(ctx context.Context, owner, batchID, bookID string) (N8ComicPublishRecord, string, bool) {
	if c == nil || c.comicPublishRecord == nil {
		return N8ComicPublishRecord{}, "", false
	}
	collection, err := c.latestComicCollection(ctx, owner, batchID, bookID, "")
	if err != nil || strings.TrimSpace(collection.ID) == "" {
		return N8ComicPublishRecord{}, "", false
	}
	record, found := c.comicPublishRecord(ctx, owner, batchID, collection.ID)
	if !found || record.Status != "published" {
		return N8ComicPublishRecord{}, "", false
	}
	return record, collection.ID, true
}

func (c *BatchAutomationCoordinator) publishVideo(ctx context.Context, owner, batchID, bookID string) error {
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

func comicBootstrapReady(book batchfactoryv11.Book, effective batchfactoryv11.SettingsPatch) bool {
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
	// 漫画资产文字始终需要准备；资产图片开关只控制后续是否生成参考图。
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
	// 与 A 批停滞阈值对齐：连续 10 分钟没有任何进度心跳即判失效，
	// 不允许无限期挂 running。
	comicGenerationIdleExtend = automationStallTimeout
	comicGenerationHardCap    = 6 * time.Hour
	comicPublishStageTimeout  = 45 * time.Minute
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
	// 部分页成功不是整本可重试错误：调用方必须按页记账并安排带退避的续跑，
	// 不能在这里原地重复调用上游。
	var pagePartial *comicPagePartialError
	if errors.As(err, &pagePartial) {
		return false
	}
	// 已计费但响应体读取失败：默认不重发，避免同一张图重复扣费；由下一轮
	// 续跑在缺失页/缺失资产上重取。
	if comicBilledReadFailure(err) {
		return false
	}
	// 已计费但本地落盘被磁盘保护拒绝：重发只会再扣一次费，等磁盘恢复后
	// 由用户手动重试。
	if comicBilledSaveFailure(err) {
		return false
	}
	// A page can still have a provider task from the previous attempt. This is
	// a temporary coordination state, not a configuration error; wait and let
	// the idempotent book generator continue from the missing page.
	message := err.Error()
	if errors.Is(err, batchfactoryv11.ErrConflict) && (strings.Contains(message, "当前漫画页正在生成") || strings.Contains(message, "当前漫画页已有生成任务")) {
		return true
	}
	if status, ok := automationErrorHTTPStatus(err); ok && status >= 400 && status < 500 && status != http.StatusRequestTimeout && status != http.StatusConflict && status != http.StatusTooManyRequests {
		return false
	}
	// 文本模型偶发输出结构不符（例如 characters[0] 缺少提示词）是模型抖动，
	// 只是因为外层包了 ErrUnavailable 就被判成永久失败，会让整本停摆。错误
	// 文案本身已经声明“可重试”的，一律按可重试处理。
	if strings.Contains(message, "文本模型输出不符合当前格式要求") || strings.Contains(message, "可重试：") {
		return true
	}
	// 输出被长度上限截断（资产过多，或思考 token 占用输出预算）同样是通道
	// 抖动：稍后重跑有机会成功，不能按 ErrUnavailable 直接判成永久失败。
	if batchfactoryv11.IsModelOutputTruncated(err) {
		return true
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

// Once every comic page has been confirmed, an earlier partial-run error is
// historical only. Keeping it on the live entry makes the UI claim that pages
// are still missing while the collection is already being published.
func clearAutomationFailure(entry map[string]any) {
	if entry == nil {
		return
	}
	delete(entry, "error")
	delete(entry, "retryRequested")
	delete(entry, "retryAt")
	delete(entry, "retryReason")
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

// automationErrorHTTPStatus 优先读结构化状态码，只有拿不到类型信息的旧错误
// 才回退解析文案。分类“确定性失败/不再回补”必须走这里，不能只看中文文案。
func automationErrorHTTPStatus(err error) (int, bool) {
	if err == nil {
		return 0, false
	}
	var upstream shuihuoImageUpstreamError
	if errors.As(err, &upstream) && upstream.StatusCode > 0 {
		return upstream.StatusCode, true
	}
	var channel shuihuoImageChatChannelError
	if errors.As(err, &channel) && channel.StatusCode > 0 {
		return channel.StatusCode, true
	}
	return comicAutomationHTTPStatus(err.Error())
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

// comicPromptRecoveryIsPartial 判定补缺提示词是否算「部分成功」：只有至少
// 补上一页、且失败明细可逐页记账时，才允许跳过缺失页继续出图；一页都没
// 补上时仍按整本失败处理。
func comicPromptRecoveryIsPartial(err error, updated int) bool {
	if err == nil || updated <= 0 {
		return false
	}
	var promptPartial *batchfactoryv11.ComicPromptPartialError
	return errors.As(err, &promptPartial)
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

// automationAssetBackfillMaxAttempts 是资产参考图缺失时的最大回补轮次。
// 超过之后不再延迟发布：带着已有资产继续后续阶段，账本保留缺失明细。
const automationAssetBackfillMaxAttempts = 2

// scheduleComicAssetBackfill 在页图已完成、但资产参考图仍缺失时安排一次
// 回补轮。下一轮只会重新生成缺失的资产图，已完成的页不会重跑。
func (c *BatchAutomationCoordinator) scheduleComicAssetBackfill(entry map[string]any, persist func(), missing int) bool {
	if entry == nil || persist == nil || missing <= 0 {
		return false
	}
	attempt := automationNumberValue(entry["assetBackfillCount"], 0)
	if attempt >= automationAssetBackfillMaxAttempts {
		return false
	}
	attempt++
	entry["assetBackfillCount"] = attempt
	backoff := automationContinuationBackoff(attempt)
	message := fmt.Sprintf("漫画页已完成，仍有 %d 个资产参考图缺失，%d 秒后回补（%d/%d）",
		missing, int(backoff.Seconds()), attempt, automationAssetBackfillMaxAttempts)
	entry["status"], entry["stage"], entry["error"], entry["updatedAt"] = "failed", "comic_asset_images", "", c.now()
	entry["message"] = message
	entry["retryRequested"] = true
	entry["retryAt"] = c.now().Add(backoff)
	appendAutomationBookLog(entry, c.now(), "comic_asset_images", message)
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
			SourceText: productionText, RawSourceText: fetched.OriginalRaw, ReplaceSource: true, ReplaceTitle: true,
			SourceTitle: fetchedBookSourceTitle(fetched, book.Title), SourceMetadata: metadata,
		})
		if err != nil {
			return book, fmt.Errorf("保存重新处理的正文失败：%w", err)
		}
		if processErr != nil {
			return updated, fmt.Errorf("重新处理正文失败：%w", processErr)
		}
		return updated, nil
	}
	if isDirectFirstSourceBlocked(book) && len(productionLines(book)) == 0 && firstMetadataString(book.SourceMetadata, "giantMaterialId", "qingyuMaterialId") == "" {
		return book, fmt.Errorf("书城原文获取失败，等待人工处理：%s", reasonOrDefault(strings.TrimSpace(stringValue(book.SourceMetadata["originalReadError"])), "请手动重试直取或确认后使用 OCR"))
	}
	if len(productionLines(book)) > 0 {
		// A previous candidate can fail first and a sibling can later succeed.
		// Source text is the authority: once it exists, remove the obsolete intake
		// failure so later automation runs and the UI do not keep reporting it.
		if isDirectFirstSourceBlocked(book) {
			metadata := copyAnyMap(book.SourceMetadata)
			markBatchDirectReadSuccess(metadata)
			updated, err := c.updateAutomationBookMetadata(ctx, owner, batchID, book, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata})
			if err != nil {
				return book, fmt.Errorf("清理旧正文读取失败状态失败：%w", err)
			}
			book = updated
		}
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
	// 列表里保存的是给人看的书城名称（如“常读”“七猫”），121 正文接口
	// 只接受数字平台 ID。单本手动获取早已走统一解析，自动化也必须复用同一
	// 条链路，不能把展示名称直接传给上游。
	platform := resolveSingleFetchPlatform("", book)
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
	fetched, resolvedBookID, fallbackFrom, err, attempts := fetchBatchSourceWithGiantCandidate(ctx, c.fetcher, c.giantMaterialResolver, book, bookID, platform, captureCharacters)
	if err != nil {
		displayErr := batchFetchDisplayError(err, book)
		persistBatchFetchFailure(ctx, c.store, owner, batchID, book, displayErr, attempts)
		return book, fmt.Errorf("自动获取正文失败：%w", displayErr)
	}
	if len(strings.TrimSpace(fetched.ProcessedText)) == 0 {
		return book, fmt.Errorf("自动获取正文失败：上游未返回可用正文")
	}
	fetched, metadata, processErr := prepareBatchFetchedSource(ctx, c.ttsStore, owner, resolvedBookID, fetched, c.classification, c.batchSensitiveConfigFor(ctx, owner, batchID, book))
	if fallbackFrom != "" {
		metadata["platformBookId"] = resolvedBookID
		metadata["resolvedPlatformBookId"] = resolvedBookID
		metadata["platformCandidateFallbackFrom"] = fallbackFrom
	}
	for key, value := range book.SourceMetadata {
		if _, exists := metadata[key]; !exists {
			metadata[key] = value
		}
	}
	markBatchDirectReadSuccess(metadata)
	productionText := fetched.ProcessedText
	if processErr != nil {
		productionText = ""
	}
	updated, err := c.captureAutomationSource(ctx, owner, batchID, book, batchfactoryv11.CaptureBookSourceInput{
		SourceText: productionText, RawSourceText: fetched.OriginalRaw, ReplaceSource: true, ReplaceTitle: true, SourceBookID: resolvedBookID, SourceTitle: fetchedBookSourceTitle(fetched, book.Title), SourceMetadata: metadata,
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

func (c *BatchAutomationCoordinator) run(ctx context.Context, epoch int64, owner, batchID, jobID string, job map[string]any, batch batchfactoryv11.Batch) {
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
	mergeRequestID := func(bookID string, batchRetry int) string {
		return automationMergeRequestID(jobID, bookID, retryAttempt, batchRetry)
	}
	concurrency := automationNumberValue(job["concurrency"], 1)
	if concurrency > 4 {
		concurrency = 4
	}
	// 每本书默认 4 路出图（可配置，上限 8），整个任务还需要一个总并发上限：
	// 实测显示并发共享上游带宽会互相拖慢，12 路时慢图更容易撞总上限。恢复为
	// 8 路（2 本书 × 4 路）；可用 job 字段 imageGlobalConcurrency 覆盖。
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
		entryVersion := automationNumberValue(initialEntry[automationBookRunVersionField], 0)
		jobMu.Unlock()
		persist := func() {
			jobMu.Lock()
			defer jobMu.Unlock()
			// 本轮 run 一旦被同一 job 的新一轮接管，旧 worker 不得再写书状态与
			// job 计数，否则会覆盖新 run 的进度。
			if !c.isCurrentRun(jobID, epoch) {
				return
			}
			// 写书只走权威入口：runVersion 不匹配说明调度器已经把这本书
			// 重新打开，旧 worker 的 failed 终态必须丢弃。
			written, err := c.persistAutomationBookEntry(owner, batchID, jobID, book.ID, initialEntry, entry, entryVersion, job)
			if err != nil || !written {
				return
			}
			job["counts"] = automationCountsFromBooks(books)
			job["needsManualPages"] = automationNeedsManualPages(books)
			job["updatedAt"] = c.now()
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
			// 每本书一个可取消上下文：停滞看门狗只跳过这一本书，其它书的
			// goroutine 继续跑；取消原因用于区分「停滞」与上游错误。
			bookCtx, bookCancel := context.WithCancelCause(ctx)
			defer bookCancel(nil)
			ctx := bookCtx
			var entryLogMu sync.Mutex
			// 看门狗只写原子心跳、不直接写 entry，避免和书本 goroutine 的
			// 状态写入并发；真正落状态在下面的 defer 里由本书 goroutine 完成。
			var stallLast atomic.Value
			var stallStage atomic.Value
			var persistedHeartbeat atomic.Value
			stallLast.Store(c.now())
			stallStage.Store("")
			persistedHeartbeat.Store(c.now())
			defer func() {
				if !errors.Is(context.Cause(ctx), errAutomationBookStalled) {
					return
				}
				last, _ := stallLast.Load().(time.Time)
				stage, _ := stallStage.Load().(string)
				entryLogMu.Lock()
				markAutomationBookStalled(entry, stage, last, c.now())
				persist()
				entryLogMu.Unlock()
			}()
			// The one-item loop lets every existing stage exit its own book with
			// continue while the other book workers keep running.
			for range [1]struct{}{} {
				// 单本书的执行日志供前端展开查看；并发出图时多个 goroutine 会
				// 同时上报阶段与去敏、审核拒绝、降级重试事件，因此单独加锁，
				// 保证日志追加与持久化不并发写同一个 entry。
				var generationLease *comicGenerationLease
				// 停滞看门狗：书内超过 automationStallTimeout 没有任何进展时，
				// 取消这本书，把并发槽位让给其它书。
				stopWatchdog := make(chan struct{})
				defer close(stopWatchdog)
				go func() {
					ticker := time.NewTicker(automationStallWatchInterval)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-stopWatchdog:
							return
						case <-ticker.C:
							last, _ := stallLast.Load().(time.Time)
							if !automationProgressStalled(last, c.now()) {
								continue
							}
							stallLast.Store(last)
							bookCancel(errAutomationBookStalled)
							return
						}
					}
				}()
				recordComicEvent := func(stage, message string) {
					// 任何出图/资产进度事件都算心跳：只要还在推进，就不判定超时。
					if generationLease != nil {
						generationLease.touch(comicGenerationIdleExtend)
					}
					entryLogMu.Lock()
					now := c.now()
					entry["stage"], entry["message"], entry["updatedAt"] = stage, message, now
					entry["lastProgressAt"] = now.UTC().Format(time.RFC3339)
					stallLast.Store(now)
					stallStage.Store(stage)
					delete(entry, "retryAt")
					appendAutomationBookLog(entry, now, stage, message)
					persist()
					entryLogMu.Unlock()
				}
				fail := func(stage string, err error, message string) {
					if errors.Is(context.Cause(ctx), errAutomationBookStalled) {
						// 看门狗已经写好阶段/最后进展时间/跳过时长，并安排了续跑。
						// 这里不能再把 context canceled 写成新的失败。
						return
					}
					if ctx.Err() != nil {
						control, _, _ := c.currentJobControl(owner, batchID, jobID)
						if control == "paused" || control == "stopped" {
							entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = control, control, "", "自动化已"+map[string]string{"paused": "暂停", "stopped": "停止"}[control], c.now()
							persist()
							return
						}
						// ctx 被取消但控制态不是暂停/停止，说明本轮已被同一 job 的
						// 新一轮执行接管。被取代的书不能写成 failed/stopped，保持
						// 新一轮写入的状态即可。
						if c.runSuperseded(jobID, epoch) {
							markAutomationEntrySuperseded(entry, c.now())
							persist()
							return
						}
					}
					disposition := classifyAutomationFailure(err, stage)
					entry["errorCode"], entry["failureClass"] = disposition.Code, disposition.Class
					outcome := recordAutomationPageLedger(entry, err, stage, c.now())
					if disposition.needsManual() || outcome.NeedsManual {
						entry["status"], entry["stage"], entry["error"], entry["updatedAt"] = "failed", stage, err.Error(), c.now()
						// 已经落到页级账本时，用真正被转人工的原因码，而不是聚合
						// 错误的分类，避免“实际是已计费未取回却显示超时”的误导。
						reason := disposition.Code
						if len(outcome.ManualCodes) > 0 {
							reason = strings.Join(outcome.ManualCodes, "/")
						}
						entry["message"] = fmt.Sprintf("%s；%s", message, automationManualReasonText(reason))
						entry["needsManual"] = true
						delete(entry, "retryRequested")
						delete(entry, "retryAt")
						appendAutomationBookLog(entry, c.now(), stage, message)
						failed.Add(1)
						persist()
						return
					}
					// 页级失败但未达升级阈值：只跳过、不原地重试，安排一次带
					// 退避的续跑。失败累计到 3 次才进入整书重试队列。
					if outcome.HasFailures && !outcome.RetryQueued {
						backoff := automationContinuationBackoff(outcome.MaxAttempts)
						entry["status"], entry["stage"], entry["error"], entry["updatedAt"] = "failed", stage, err.Error(), c.now()
						entry["message"] = fmt.Sprintf("%s；已跳过本次，%d 秒后续跑未完成项（失败 %d/%d 次）",
							message, int(backoff.Seconds()), outcome.MaxAttempts, automationRetryEscalationThreshold)
						entry["retryRequested"] = true
						entry["retryAt"] = c.now().Add(backoff)
						appendAutomationBookLog(entry, c.now(), stage, message)
						failed.Add(1)
						persist()
						return
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
					scheduledRetry := false
					if err != nil && shouldScheduleAutomationTimeoutRetry(stage) && (errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "超时") || strings.Contains(strings.ToLower(err.Error()), "deadline")) {
						retryCount := automationNumberValue(entry["retryCount"], 0)
						if retryCount < automationBatchRetryRounds {
							entry["retryCount"] = retryCount + 1
							entry["retryRequested"] = true
							entry["retryAt"] = c.now().Add(30 * time.Second)
							entry["message"] = fmt.Sprintf("%s，已安排自动重试（%d/%d）", message, retryCount+1, automationBatchRetryRounds)
							scheduledRetry = true
						}
					}
					if !scheduledRetry {
						// 回补预算用尽：本书退出自动调度，其它书继续执行；是否
						// 再跑由用户点“重试失败小说”决定，避免无限循环烧钱。
						delete(entry, "retryRequested")
						delete(entry, "retryAt")
						entry["needsManual"] = true
						entry["message"] = fmt.Sprintf("%s；已自动跳过并回补 %d 轮仍未完成，已停止自动重试，请人工处理或手动重试", message, automationBatchRetryRounds)
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
					// 旧 run 被同一 job 的新一轮重试接管时，不能把书标成
					// "测试任务已停止"：这会被用户误读成主动停止，而且新一轮
					// 仍会重新调度这些书。保留原状态并标注接管原因即可。
					if c.runSuperseded(jobID, epoch) {
						markAutomationEntrySuperseded(entry, c.now())
						persist()
						continue
					}
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "stopped", "stopped", "测试任务已停止", c.now()
					persist()
					continue
				}
				// 幂等：已经发布成功的书，在重新执行/误点执行时直接跳过，
				// 避免在 N8 重复创建合集。强制重生成时才允许重新发布。
				if !automationBoolValue(job["forceRegenerate"]) &&
					automationStringValue(job["runMode"], "full_submit") == "full_submit" &&
					batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch) == "comic" {
					if record, collectionID, published := c.publishedComicRecord(ctx, owner, batchID, book.ID); published {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "published", "该书已发布过，跳过重复上传", c.now()
						entry["comicCollectionId"], entry["comicCarouselId"], entry["comicPublishName"] = collectionID, record.CarouselID, record.Name
						entry["comicImageIDs"] = record.ImageIDs
						appendAutomationBookLog(entry, c.now(), "publish", "检测到已发布记录，跳过重复发布")
						persist()
						continue
					}
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
					// 本轮补齐提示词后仍缺失的页数；storyboard_only 收尾必须保留它，
					// 不能把「部分页缺提示词」写成「每页提示词已完成」。
					comicPromptMissing := 0
					switch selectComicPreparationAction(book, effective) {
					case comicPreparationBootstrap:
						identityIssue := batchfactoryv11.ComicAssetIdentityIssue(book)
						identityRecovery := identityIssue != ""
						// 资产文字始终提取；是否生成资产参考图由后续出图阶段的开关决定。
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
						// 阶段成功即清零当前账期；后续新故障按新的第 1 次重新累计。
						if resetAutomationPageLedgerNode(entry, "comic_bootstrap", c.now()) {
							persist()
						}
						if identityRecovery {
							entry["stage"], entry["message"], entry["updatedAt"] = "comic_prompt_recovery", "人物身份已修复，正在重新生成全部漫画页提示词", c.now()
							persist()
							stageCtx, cancel := context.WithTimeout(ctx, comicPromptStageTimeout)
							stageCtx = batchfactoryv11.WithComicPromptConcurrency(stageCtx, comicPromptConcurrency)
							_, promptErr := c.stages.RunComicPromptExtraction(stageCtx, owner, batchID, book.ID)
							cancel()
							if promptErr != nil {
								fail("comic_prompt_recovery", promptErr, "漫画身份修复后的提示词重新生成失败")
								continue
							}
							if resetAutomationPageLedgerNode(entry, "comic_prompt_recovery", c.now()) {
								persist()
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
						partialResult, recoveryErr := c.stages.RunMissingComicPromptExtraction(stageCtx, owner, batchID, book.ID)
						cancel()
						if recoveryErr != nil {
							// 只有「一页都没补上」才允许整本失败；部分成功时失败页只记账，
							// 其余页继续进入页图阶段，页图阶段会自然跳过缺提示词的页。
							if !comicPromptRecoveryIsPartial(recoveryErr, len(partialResult.Updated)) {
								fail("comic_prompt_recovery", recoveryErr, "漫画缺页提示词补齐失败")
								continue
							}
							recordAutomationPageLedger(entry, recoveryErr, "comic_prompt_recovery", c.now())
							comicPromptMissing = len(partialResult.Failed)
							message := fmt.Sprintf("提示词 %d/%d 页就绪，%d 页缺失已跳过，继续生成其余页",
								len(partialResult.Updated), len(partialResult.Updated)+len(partialResult.Failed), len(partialResult.Failed))
							entry["stage"], entry["message"], entry["updatedAt"] = "comic_images", message, c.now()
							appendAutomationBookLog(entry, c.now(), "comic_prompt_recovery", message)
							persist()
						} else if resetAutomationPageLedgerNode(entry, "comic_prompt_recovery", c.now()) {
							persist()
						}
					case comicPreparationReady:
						entry["stage"], entry["message"], entry["updatedAt"] = "comic_resume", "已复用完整漫画资产和分页，继续补齐图片", c.now()
						persist()
					}
					if automationStringValue(job["runMode"], "full_submit") == "storyboard_only" {
						message := "已完成漫画资产、分页与每页提示词"
						if comicPromptMissing > 0 {
							// 缺页时必须保留真实缺口，不能覆盖成「每页提示词已完成」。
							message = fmt.Sprintf("部分完成：漫画资产与分页已就绪；还有 %d 页缺提示词，这些页的提示词未生成", comicPromptMissing)
							entry["partial"] = true
							entry["promptMissingPages"] = comicPromptMissing
							appendAutomationBookLog(entry, c.now(), "ready_for_images", message)
						}
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "ready_for_images", message, c.now()
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
						// 上游返回响应头、图片落库等真实进展只刷新看门狗心跳，
						// 不追加日志；慢图在重试预算内不会被误判为 stalled。
						generateCtx = withComicGenerationHeartbeat(generateCtx, func() {
							now := c.now()
							if generationLease != nil {
								generationLease.touch(comicGenerationIdleExtend)
							}
							stallLast.Store(now)
							lastPersisted, _ := persistedHeartbeat.Load().(time.Time)
							if now.Sub(lastPersisted) < time.Minute {
								return
							}
							entryLogMu.Lock()
							lastPersisted, _ = persistedHeartbeat.Load().(time.Time)
							if now.Sub(lastPersisted) >= time.Minute {
								entry["updatedAt"] = now
								entry["lastProgressAt"] = now.UTC().Format(time.RFC3339)
								persistedHeartbeat.Store(now)
								persist()
							}
							entryLogMu.Unlock()
						})
						generateCtx = withShuihuoImageEvent(generateCtx, recordComicEvent)
						generateCtx = withComicAssetConcurrency(generateCtx, 5)
						// 资产参考图缺失只记账不回滚：页图阶段继续用已有资产出图，
						// 成功后再安排一次回补轮。
						delete(entry, "assetShortfallCount")
						generateCtx = withComicAssetShortfall(generateCtx, func(shortfall comicAssetShortfall) {
							entryLogMu.Lock()
							recordComicAssetShortfall(entry, shortfall, c.now())
							persist()
							entryLogMu.Unlock()
						})
						// 自动改写画面/台词后写进页级账本，前端可以看到
						// “这一页被自动改写过”以及改了哪些字段。
						generateCtx = withComicPromptRepair(generateCtx, func(repair comicPromptRepair) {
							entryLogMu.Lock()
							recordComicPromptRepair(entry, repair, c.now())
							persist()
							entryLogMu.Unlock()
						})
						// 按本串行出图：拿到槽位才真正开始生成资产和页图，出完并发布后
						// 才释放给下一本。这样一本书几十页可以连续占满上游通道快速收口，
						// 不会再出现“每本都缺几页、一本也提交不了”。
						// 等待期间持续刷新整本心跳，避免排队被误判成停滞而取消。
						if c.comicRunSlots != nil {
							waitTicker := time.NewTicker(time.Minute)
							for waiting := true; waiting; {
								select {
								case c.comicRunSlots <- struct{}{}:
									waiting = false
								case <-waitTicker.C:
									if generationLease != nil {
										generationLease.touch(comicGenerationIdleExtend)
									}
								case <-generateCtx.Done():
									waitTicker.Stop()
									return batchfactoryv11.ComicCollection{}, generateCtx.Err()
								}
							}
							waitTicker.Stop()
							defer func() { <-c.comicRunSlots }()
						}
						// 出图槽位看门狗：占着槽位却长时间没有任何真实进展时，主动
						// 中止本次尝试并立刻释放槽位，让下一本先出图；本书按失败处理，
						// 可以直接重试，避免整批一直卡在同一本上。
						var slotStalled atomic.Bool
						slotWatchDone := make(chan struct{})
						go func() {
							ticker := time.NewTicker(30 * time.Second)
							defer ticker.Stop()
							for {
								select {
								case <-slotWatchDone:
									return
								case <-ticker.C:
									last, _ := stallLast.Load().(time.Time)
									if last.IsZero() || c.now().Sub(last) <= comicSlotStallTimeout {
										continue
									}
									slotStalled.Store(true)
									cancel()
									return
								}
							}
						}()
						defer close(slotWatchDone)
						collection, err := c.comicGenerate(generateCtx, owner, batchID, book.ID)
						if err != nil && slotStalled.Load() {
							return collection, fmt.Errorf("出图超过 %s 没有任何进展，已释放出图槽位，本书可直接重试：%w", comicSlotStallTimeout, err)
						}
						return collection, err
					}, nil)
					generationCancel()
					if generateErr != nil {
						failureStage, failureDetail := comicGenerationFailureStage(generateErr)
						// 部分页成功、部分页失败：只记账并安排续跑，不把整本判死。
						// 下一轮从缺失页继续，已经出好的页不会重跑。
						var pagePartial *comicPagePartialError
						if errors.As(generateErr, &pagePartial) {
							outcome := recordAutomationPageLedger(entry, generateErr, failureStage, c.now())
							if outcome.NeedsManual || outcome.MaxAttempts >= automationPageRetryMaxAttempts {
								detail := fmt.Sprintf(
									"已生成 %d/%d 页，剩余页连续 %d 次未补齐",
									pagePartial.Completed, pagePartial.Total, outcome.MaxAttempts)
								if outcome.NeedsManual {
									detail = fmt.Sprintf(
										"已生成 %d/%d 页，%d 页需要人工处理",
										pagePartial.Completed, pagePartial.Total, outcome.ManualPages)
								}
								fail(failureStage, generateErr, detail)
								continue
							}
							backoff := automationContinuationBackoff(outcome.MaxAttempts)
							message := fmt.Sprintf("已生成 %d/%d 页，%d 页失败已记账，%d 秒后续跑只补缺失页",
								pagePartial.Completed, pagePartial.Total, len(pagePartial.Failures), int(backoff.Seconds()))
							entry["status"], entry["stage"], entry["error"], entry["updatedAt"] = "failed", failureStage, generateErr.Error(), c.now()
							entry["message"] = message
							entry["retryRequested"] = true
							entry["retryAt"] = c.now().Add(backoff)
							appendAutomationBookLog(entry, c.now(), failureStage, message)
							persist()
							continue
						}
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
					// 整本出图成功说明所有页都已落地：清零当前账期，
					// 下一轮新故障从第 1 次重新累计。
					resetAutomationPageLedgerNode(entry, "comic_images", c.now())
					confirmCtx, cancelConfirm := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
					confirmedCollection, confirmErr := c.confirmComicCollection(confirmCtx, owner, batchID, book.ID, collection)
					cancelConfirm()
					if confirmErr != nil {
						fail("comic_images", confirmErr, "漫画图片已返回，但合集回填确认失败")
						continue
					}
					collection = confirmedCollection
					clearAutomationFailure(entry)
					entry["comicCollectionId"], entry["comicPageCount"] = collection.ID, len(collection.Items)
					// 资产参考图缺失不阻断整本：页图已全部完成，先安排回补轮
					// 把缺失的参考图补回来；回补预算用完就带着已有资产继续发布。
					shortfallCount := automationNumberValue(entry["assetShortfallCount"], 0)
					delete(entry, "assetShortfallCount")
					if shortfallCount > 0 {
						if c.scheduleComicAssetBackfill(entry, persist, shortfallCount) {
							continue
						}
						message := fmt.Sprintf("有 %d 个资产参考图多次回补仍未写回，用现有资产继续后续阶段", shortfallCount)
						entry["message"] = message
						appendAutomationBookLog(entry, c.now(), "comic_asset_images", message)
						persist()
					} else {
						delete(entry, "assetBackfillCount")
						resetAutomationPageLedgerNode(entry, "comic_asset_images", c.now())
					}
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
				resumeStage := automationResumeStageForBook(automationStringValue(entry["resumeStage"], ""), batch, book)
				// forceRegenerate authorizes the initial run only. A continuation
				// retry must preserve completed work even when the original run was
				// explicitly started as a full regeneration.
				forceRegenerate := automationBoolValue(job["forceRegenerate"]) && resumeStage == ""
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
					if !automationResumeShouldRunStage(resumeStage, "assets") {
						const skipMessage = "从失败阶段继续，跳过已完成的资产阶段"
						entry["stage"], entry["message"], entry["updatedAt"] = "assets", skipMessage, c.now()
						appendAutomationBookLog(entry, c.now(), "assets", skipMessage)
						persist()
					} else if batchfactoryv11.AssetImagesEnabled(effective) {
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
						if !bookHasStoredAssetDescription(book) {
							appendAutomationBookLog(entry, c.now(), "assets", "无已存资产描述，已跳过合并")
						}
						persist()
					}
					if !automationResumeShouldRunStage(resumeStage, "director") {
						const skipMessage = "从失败阶段继续，跳过已完成的分镜提示词阶段"
						entry["stage"], entry["message"], entry["updatedAt"] = "director", skipMessage, c.now()
						appendAutomationBookLog(entry, c.now(), "director", skipMessage)
						persist()
					} else if forceRegenerate || !automationDirectorComplete(book) {
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
					if automationResumeShouldRunStage(resumeStage, "opening") && automationOpeningEnabled(batch, book) && (forceRegenerate || !batchfactoryv11.OpeningStageComplete(batch, book)) {
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
					completedProduction, waitErr := c.waitProduction(stageCtx, owner, batchID, book.ID, videoRequestID(book.ID, automationNumberValue(entry["batchRetryCount"], 0)), videoWaitIdleTimeout, func() {
						// 供应商侧任务状态在推进也算书内进展，避免慢视频被停滞看门狗误杀。
						stallLast.Store(c.now())
					})
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
					mergeJobs, mergeErr := c.merge.SubmitBookMerge(mergeCtx, owner, batchID, book.ID, mergeRequestID(book.ID, automationNumberValue(entry["batchRetryCount"], 0)), batchfactoryv11.MergeOptions{TimingMode: "audio", AudioSource: strings.ToLower(strings.TrimSpace(audioSource)), AudioArtifactID: ttsArtifactID, AudioDurationSeconds: audioDurationSeconds})
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
	// 同一个 jobID 上可能已经启动了新一轮 run（重试/续跑）。旧 run 只有在自己的
	// epoch 仍然有效、并且本次调度的书都已到终态时才允许写 job 终态与 counts，
	// 否则会把新 run 的状态覆盖成 failed/completed（线上表现为 job 提前 failed
	// 而同一 job 仍有 3 本书在跑）。
	if !c.isCurrentRun(jobID, epoch) {
		return
	}
	for _, book := range batch.Books {
		entry, _ := books[book.ID].(map[string]any)
		if !automationEntryTerminal(entry) {
			return
		}
	}
	// Start 与本轮收尾共用同一批次锁，避免“刚检查完没有后续书，另一个请求又
	// 追加 queued，随后被本轮终态覆盖”的竞态。
	unlockBatchRun := c.lockBatchRun(owner, batchID)
	defer unlockBatchRun()
	if nextJob, queuedIDs, claimErr := c.claimAutomationFollowUpBooks(owner, batchID, jobID); claimErr != nil {
		log.Printf("claim follow-up books for automation %s: %v", jobID, claimErr)
		return
	} else if len(queuedIDs) > 0 {
		latestBatch, getErr := c.store.GetBatch(ctx, owner, batchID)
		if getErr != nil {
			log.Printf("load follow-up batch for automation %s: %v", jobID, getErr)
			return
		}
		queued := make(map[string]bool, len(queuedIDs))
		for _, id := range queuedIDs {
			queued[id] = true
		}
		latestBatch.Books = slices.DeleteFunc(latestBatch.Books, func(book batchfactoryv11.Book) bool { return !queued[book.ID] })
		c.launchRun(owner, batchID, jobID, nextJob, latestBatch)
		return
	}
	// Recount from persisted entries rather than only this goroutine's local
	// failures. A retry run intentionally skips unrelated terminal books; those
	// books must remain failed in the aggregate instead of being reported ready.
	failedCount := 0
	readyCount := 0
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
	}
	autoRetryPending := automationPendingBookRetry(books)
	// counts.total 必须按 job 跟踪的全部书计算：重试只调度未成功的书，
	// batch.Books 已收窄为本次要跑的子集，用 len(batch.Books) 会让总数变少。
	totalBooks := len(books)
	if autoRetryPending {
		job["state"], job["updatedAt"] = "running", c.now()
		job["counts"] = automationCounts(totalBooks, 0, failedCount, readyCount)
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
		job["counts"] = automationCountsWithPending(totalBooks, pendingCount)
		_ = c.setJob(owner, batchID, jobID, job)
		return
	}
	job["state"] = "completed"
	job["completedAt"], job["updatedAt"] = c.now(), c.now()
	job["counts"] = automationCounts(totalBooks, 0, failedCount, readyCount)
	if failedCount > 0 {
		// 书级失败不阻塞其它书；所有书都收尾后仍有失败，说明已跑满自动回补
		// 轮次（或遇到确定性错误），任务停止自动调度，把决定权交回用户。
		job["state"] = "failed"
		job["lastError"] = fmt.Sprintf("自动调度已停止：仍有 %d 本未完成；请人工处理或点“重试失败小说”继续", failedCount)
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

// automationEntryTerminal 判断单本书是否已经落到终态或明确的等待态。
// running/queued 说明本书仍在本轮（或新一轮）执行中，run() 不能在这种时候
// 写 job 终态，否则会出现“job 已 failed、但还有书在跑”的假状态。
func automationEntryTerminal(entry map[string]any) bool {
	if entry == nil {
		return true
	}
	switch automationStringValue(entry["status"], "") {
	case "completed", "failed", "blocked", "stopped", "paused", "scheduled", "pending_confirmation":
		return true
	}
	return false
}

// automationBookRetryable 判断重试是否应该重新调度这本书。已完成（含已发布）
// 与正在执行的书必须排除，否则重试会把它们再跑一遍；其余未成功状态
// （failed/blocked/stopped/paused/queued/等待中/从未执行）都属于可重试范围。
//
// currentSourceBookID 可选：传入批次记录里的当前 source_book_id。只有「获取原文」
// 阶段失败时才用它判断书号是否已被用户改过——书号填错导致的失败是确定性的，
// 同一书号反复重试不会成功，只有书号变化后才值得重新抓取。
func automationBookRetryable(entry map[string]any, currentSourceBookID ...string) bool {
	if entry == nil {
		return false
	}
	if automationStringValue(entry["stage"], "") == "published" {
		return false
	}
	status := automationStringValue(entry["status"], "")
	switch status {
	case "completed", "running":
		return false
	}
	if automationStringValue(entry["stage"], "") == "fetch_original" && status == "failed" {
		sourceBookID := ""
		if len(currentSourceBookID) > 0 {
			sourceBookID = strings.TrimSpace(currentSourceBookID[0])
		}
		if sourceBookID != "" {
			return automationFetchFailureBookChanged(entry, sourceBookID)
		}
	}
	return true
}

// automationFetchFailureBookChanged 判断一次「获取原文」失败是否因为用户已经
// 换过书号：错误文案里出现过书号（如“书号 794582 在七猫无法读取”），且没有
// 任何一个与批次记录里的当前 source_book_id 相同。文案里没有书号时无法判断，
// 沿用旧语义允许重试。
func automationFetchFailureBookChanged(entry map[string]any, currentSourceBookID string) bool {
	message := automationStringValue(entry["error"], "")
	bookIDs := automationDigitRuns(message)
	if len(bookIDs) == 0 {
		return true
	}
	for _, bookID := range bookIDs {
		if bookID == currentSourceBookID {
			return false
		}
	}
	return true
}

// automationDigitRuns 提取文本里长度不少于 3 的数字串，用来识别错误文案中的书号。
func automationDigitRuns(text string) []string {
	runs := make([]string, 0, 4)
	var current strings.Builder
	flush := func() {
		if current.Len() >= 3 {
			runs = append(runs, current.String())
		}
		current.Reset()
	}
	for _, char := range text {
		if char >= '0' && char <= '9' {
			current.WriteRune(char)
			continue
		}
		flush()
	}
	flush()
	return runs
}

// selectRetryableAutomationBooks 返回本次重试真正要调度的书。
// requested 为空表示“重试全部未成功的书”；非空时先与 job 实际跟踪的书取交集。
// sourceBookIDs 可选：job 里的书 ID 到批次记录当前 source_book_id 的映射。
func selectRetryableAutomationBooks(books map[string]any, requested map[string]bool, sourceBookIDs ...map[string]string) map[string]bool {
	currentSourceBookIDs := map[string]string(nil)
	if len(sourceBookIDs) > 0 {
		currentSourceBookIDs = sourceBookIDs[0]
	}
	selected := make(map[string]bool, len(books))
	for id, raw := range books {
		if len(requested) > 0 && !requested[id] {
			continue
		}
		entry, _ := raw.(map[string]any)
		if !automationBookRetryable(entry, currentSourceBookIDs[id]) {
			continue
		}
		selected[id] = true
	}
	return selected
}

// markAutomationEntrySuperseded 用于旧 run 被同一 job 的新一轮执行接管时：
// 保留书在新一轮中的状态，只标注接管原因，不能再写 stopped/“测试任务已停止”
// 把重试副作用显示成用户主动停止。
func markAutomationEntrySuperseded(entry map[string]any, now time.Time) {
	if entry == nil {
		return
	}
	entry["supersededByRun"] = true
	if automationStringValue(entry["message"], "") == "" {
		entry["message"] = "已被新一轮任务接管，等待重新执行"
	}
	entry["updatedAt"] = now
	appendAutomationBookLog(entry, now, automationStringValue(entry["stage"], ""), "旧执行轮次已被新一轮接管")
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

// appendAutomationJobLog 把一条任务级事件追加到 job 日志里（僵尸回收等没有
// 具体书归属的事件走这里）。
func appendAutomationJobLog(job map[string]any, at time.Time, message string) {
	if job == nil {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	logs := make([]any, 0, 8)
	switch typed := job["logs"].(type) {
	case []any:
		logs = append(logs, typed...)
	case []map[string]any:
		for _, item := range typed {
			logs = append(logs, item)
		}
	}
	logs = append(logs, map[string]any{"at": at, "stage": "automation", "message": message})
	if len(logs) > automationBookLogLimit {
		logs = logs[len(logs)-automationBookLogLimit:]
	}
	job["logs"] = logs
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

func (c *BatchAutomationCoordinator) waitProduction(ctx context.Context, owner, batchID, bookID, requestID string, idleTimeout time.Duration, onProgress ...func()) (batchfactoryv11.ProductionJob, error) {
	if c.stages == nil || c.stages.Production == nil {
		return batchfactoryv11.ProductionJob{}, fmt.Errorf("视频生产服务未配置")
	}
	return waitProductionStatus(ctx, func(ctx context.Context) (batchfactoryv11.BatchStatus, error) {
		return c.stages.Production.GetBatchStatus(ctx, owner, batchID)
	}, bookID, requestID, idleTimeout, onProgress...)
}

// waitProductionStatus 是视频等待的核心循环。fetch 由调用方注入，便于测试。
func waitProductionStatus(ctx context.Context, fetch func(context.Context) (batchfactoryv11.BatchStatus, error), bookID, requestID string, idleTimeout time.Duration, onProgress ...func()) (batchfactoryv11.ProductionJob, error) {
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
				for _, report := range onProgress {
					if report != nil {
						report()
					}
				}
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

// bookHasStoredAssetDescription reports whether the book already carries any
// stored asset description that the closed switch path can merge into prompts.
func bookHasStoredAssetDescription(book batchfactoryv11.Book) bool {
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Prompt) != "" {
			return true
		}
	}
	return false
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
	// 已有原版视频不代表换开头流程也完成。这里必须与上传阶段使用同一套
	// openingCount 口径；否则自动化会复用原版、跳过换开头，最后在上传时
	// 才因缺少 opening 1/2/... 失败。返回未完成后，Missing 模式只会补齐
	// 缺失的换开头提示词和对应视频，不会重做已有原版成片。
	if !batchfactoryv11.OpeningStageComplete(batch, book) {
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
		clearAutomationRetryMarks(job)
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
	if reconcileTerminalAutomationJob(target, c.now()) {
		reopened = true
	}
	// 主修：running 表示任务正在执行，重试绝不能在同一 jobID 上再起一轮 run。
	// 旧实现会先取消当前 run 的 ctx 再 launchRun，旧 run 的 worker 随后在
	// ctx.Err() 分支把书改写成 stopped（“测试任务已停止”），用户却以为自己
	// 只是点了重试。
	if automationStringValue(target["state"], "") == "running" {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: 任务正在执行，完成后再重试", batchfactoryv11.ErrConflict)
	}
	// 上次执行被中断后，书会停留在 running；选择重试书时会把它们当成
	// “正在执行”跳过，用户看到的就是“重试失败小说（0）”，点了也没有反应。
	if repairStaleRunningAutomationJob(target, c.now()) {
		reopened = true
	}
	books, _ := target["books"].(map[string]any)
	// 书号可能在抓取失败后被用户在「编辑内容」里改过。确定性抓取失败只有换号
	// 才可能成功，所以按批次记录的当前 source_book_id 判断是否值得重新抓取；
	// 读不到批次时沿用旧规则，不阻塞重试。
	currentSourceBookIDs := map[string]string{}
	if c.store != nil {
		if currentBatch, readErr := c.store.GetBatch(ctx, owner, batchID); readErr == nil {
			for _, book := range currentBatch.Books {
				currentSourceBookIDs[book.ID] = book.BookID
			}
		}
	}
	// 重试只调度未成功的书：已完成/已发布、正在执行、等待定时提交或等待人工
	// 确认的书都不进入新一轮 run。显式传了 bookIds 时先与 job 跟踪的书取交集。
	selected := selectRetryableAutomationBooks(books, set, currentSourceBookIDs)
	if len(selected) == 0 {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: 没有需要重试的书，已完成或正在执行的书不会重复执行", batchfactoryv11.ErrConflict)
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
	resetRetryableAutomationBooks(books, selected, c.now())
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
	// 必须在启动后台 worker 之前、且仍持有状态锁时生成公开响应。
	// worker 启动后会持续更新书/节点/页账本；这时再遍历 target 会触发
	// concurrent map read and map write，并直接杀死整个 Go 进程。
	response := publicAutomationJob(target)
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
	// job never tracked (a batch run can cover a subset). Relaunch only the
	// selected (not-yet-successful) books; completed entries stay out of this run.
	scoped := batch
	scoped.Books = make([]batchfactoryv11.Book, 0, len(selected))
	for _, book := range batch.Books {
		if selected[book.ID] {
			scoped.Books = append(scoped.Books, book)
		}
	}
	if len(scoped.Books) == 0 {
		return nil, fmt.Errorf("%w: 没有需要重试的书，已完成或正在执行的书不会重复执行", batchfactoryv11.ErrConflict)
	}
	c.launchRun(owner, batchID, automationStringValue(target["id"], ""), target, scoped)
	return response, nil
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
			failedStage := automationStringValue(entry["stage"], "")
			stage, message := automationRetryResumeStage(failedStage), "从失败阶段继续执行"
			if failedStage == "comic_publish" && strings.TrimSpace(automationStringValue(entry["comicCollectionId"], "")) != "" {
				stage, message = "comic_publish", "重新发布已生成漫画合集"
			}
			entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", stage, message, "", now
			entry["resumeStage"] = stage
			entry["retryRequested"] = true
			// 人工重试视为重新授权一轮自动重试预算；否则 3/3 用尽后的重试
			// 会被 scheduleComicContentPolicyRetry 直接判死，按钮等于失效。
			delete(entry, "needsManual")
			resetAutomationManualPageRetryBudget(entry)
			delete(entry, "contentPolicyRetryCount")
			delete(entry, "retryAt")
			delete(entry, "retryReason")
			// 人工重试同样重置“批次末尾回补”的预算与等待标记。
			delete(entry, "batchRetryCount")
			delete(entry, "batchRetryPending")
		}
	}
}

// resetAutomationManualPageRetryBudget 只在用户明确点击“重试”时调用。
// 已生成页仍由数据库里的主图判断并跳过；这里仅清除旧失败账本的人工门禁，
// 让缺失页获得一轮新的重试预算，同时保留历史错误和费用供审计。
func resetAutomationManualPageRetryBudget(entry map[string]any) {
	nodes, _ := entry["nodes"].(map[string]any)
	for _, nodeRaw := range nodes {
		node, _ := nodeRaw.(map[string]any)
		if node == nil {
			continue
		}
		ledger, _ := node["pageLedger"].(map[string]any)
		for _, pageRaw := range ledger {
			page, _ := pageRaw.(map[string]any)
			if page == nil {
				continue
			}
			page["attempts"] = 0
			page["needsManual"] = false
			page["retryQueued"] = false
			page["state"] = "retrying"
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
	// 与 Retry 一致：任务正在执行时不能再起第二个 run。旧实现会把 state 写回
	// running 并 launchRun，取消旧 run 的 ctx，造成两轮 run 并存。
	if nextState == "running" && automationStringValue(target["state"], "") == "running" {
		c.mu.Unlock()
		return nil, fmt.Errorf("%w: 任务正在执行，完成后再继续", batchfactoryv11.ErrConflict)
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
	registerBatchAutomationPreflightRoute(mux, auth, coordinator, version)
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
