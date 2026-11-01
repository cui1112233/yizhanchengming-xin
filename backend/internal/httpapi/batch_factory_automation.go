package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
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
)

// BatchAutomationCoordinator is the Go owner of the V12 start/stop controls.
// It intentionally persists the same public job shape that the old worker
// used, so the existing workbench does not need a second status protocol.
type BatchAutomationCoordinator struct {
	store          batchfactoryv11.Store
	presets        AutomationPresetStore
	stages         *batchfactoryv11.BookStageService
	merge          *batchfactoryv11.MergeService
	external       *external.Service
	ttsStore       novelfetchworkshop.Store
	fetcher        *novelfetchworkshop.UpstreamFetcher
	classification batchFetchClassification
	audioProbe     batchfactoryv11.H3AudioDurationProbe
	localFiles     *localartifact.Store
	path           string
	now            func() time.Time
	providerSync   func(context.Context, string) error
	publish121     *BatchFactory121Publisher
	comicGenerate  func(context.Context, string, string, string) (batchfactoryv11.ComicCollection, error)
	comicPublish   func(context.Context, string, string, string, batchfactoryv11.ComicCollection, batchfactoryv11.SettingsPatch) (N8ComicPublishRecord, error)
	mu             sync.Mutex
	controlMu      sync.Mutex
	startMu        sync.Mutex
	runCancels     map[string]context.CancelFunc
	dispatcherOnce sync.Once
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
	changed := false
	for jobID, job := range state.Jobs {
		if automationStringValue(job["state"], "") == "running" {
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
				job["updatedAt"] = now
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
		if validationErr := c.validateAutomationPreflight(ctx, item.owner, batch, scoped.Books, automationStringValue(item.job["runMode"], "full_submit")); validationErr != nil {
			item.job["state"], item.job["lastError"], item.job["updatedAt"] = "failed", validationErr.Error(), c.now()
			_ = c.setJob(item.owner, item.batchID, item.jobID, item.job)
			continue
		}
		c.launchRun(item.owner, item.batchID, item.jobID, item.job, scoped)
	}
	return nil
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

func (c *BatchAutomationCoordinator) loadLocked() (automationStateFile, error) {
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{}}
	if strings.TrimSpace(c.path) == "" {
		return state, nil
	}
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
		if err := c.saveLocked(state); err != nil {
			return state, err
		}
	}
	return state, nil
}

func (c *BatchAutomationCoordinator) saveLocked(state automationStateFile) error {
	if strings.TrimSpace(c.path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

func (c *BatchAutomationCoordinator) setJob(owner, batchID, jobID string, job map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return err
	}
	state.Jobs[jobID] = job
	return c.saveLocked(state)
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
func resumeInterruptedAutomationJob(job map[string]any, now time.Time) bool {
	if automationStringValue(job["state"], "") != "stopped" || automationStringValue(job["lastError"], "") != automationServiceRestartMessage {
		return false
	}
	job["state"], job["lastError"], job["updatedAt"] = "running", "", now
	delete(job, "stopAt")
	job["restartedAt"] = now
	books, _ := job["books"].(map[string]any)
	running, failed, ready := 0, 0, 0
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		if automationStringValue(entry["status"], "") != "completed" {
			entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", "assets", "服务已恢复，自动化继续执行", "", now
			entry["retryRequested"] = true
			running++
		} else {
			ready++
		}
	}
	job["counts"] = automationCounts(len(books), running, failed, ready)
	return true
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
	return map[string]any{"total": total, "ready": ready, "running": running, "pending": 0, "failed": failed, "blocked": 0}
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
		if err := c.validateAutomationPreflight(ctx, owner, batch, targetBooks, runMode); err != nil {
			return nil, err
		}
	}
	books := job["books"].(map[string]any)
	for _, book := range targetBooks {
		status, stage, message := "running", "assets", "Go 自动化已接管"
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

func (c *BatchAutomationCoordinator) validateAutomationPreflight(ctx context.Context, owner string, batch batchfactoryv11.Batch, books []batchfactoryv11.Book, runMode string) error {
	reasons := make([]string, 0)
	if c.stages == nil || c.stages.Director == nil || c.stages.Director.Provider == nil {
		reasons = append(reasons, "文本模型执行器未配置")
	}
	productionMode := comicSettingString(batch.SettingsState.Patch, "productionMode", "video")
	for _, book := range books {
		effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
		if comicSettingString(effective, "textModelId", "") == "" {
			reasons = append(reasons, "文本模型未配置")
		}
		if runMode == "storyboard_only" {
			continue
		}
		if productionMode == "comic" {
			if comicSettingString(effective, "imageModelId", "") == "" {
				reasons = append(reasons, "图片模型未配置")
			}
		} else if comicSettingString(effective, "videoModelId", "") == "" {
			reasons = append(reasons, "视频模型未配置")
		}
	}
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
		for _, book := range books {
			effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
			var publish map[string]any
			if raw := effective["publishSettings"]; len(raw) > 0 {
				_ = json.Unmarshal(raw, &publish)
			}
			if publish == nil || !hasAutomationPublishProfile(publish) {
				reasons = append(reasons, "发布配置未绑定网站或组织")
				break
			}
		}
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
	merged := map[string]any{}
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range config {
		merged[key] = value
	}
	if len(requestedProductionMode) > 0 {
		mode := strings.TrimSpace(requestedProductionMode[0])
		if mode == "comic" || mode == "video" {
			merged["productionMode"] = mode
		}
	}
	for _, key := range []string{"aiPromptConfig", "publishSettings"} {
		left, _ := current[key].(map[string]any)
		right, _ := config[key].(map[string]any)
		if left != nil || right != nil {
			nested := map[string]any{}
			for k, v := range left {
				nested[k] = v
			}
			for k, v := range right {
				nested[k] = v
			}
			merged[key] = nested
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
	go c.run(automationCtx, owner, batchID, jobID, job, batch)
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

// ensureAutomationSource fills a book whose intake retained its ID but not its
// original text. Immediate runs and retries must recover this required input
// before starting audio planning rather than failing the whole production path.
func (c *BatchAutomationCoordinator) ensureAutomationSource(ctx context.Context, owner, batchID string, book batchfactoryv11.Book) (batchfactoryv11.Book, error) {
	if strings.EqualFold(stringValue(book.SourceMetadata["processingStatus"]), "failed") && strings.TrimSpace(book.TxtText) != "" {
		maxTxt := int(targetNumber(book.SourceMetadata["maxTxt"]))
		fetched := batchFetchResultFromStoredRaw(book, maxTxt)
		fetched, metadata, processErr := prepareBatchFetchedSource(ctx, c.ttsStore, owner, book.BookID, fetched, c.classification)
		for key, value := range book.SourceMetadata {
			if _, exists := metadata[key]; !exists {
				metadata[key] = value
			}
		}
		productionText := fetched.ProcessedText
		if processErr != nil {
			productionText = ""
		}
		updated, err := c.store.CaptureBookSource(ctx, owner, batchID, book.ID, batchfactoryv11.CaptureBookSourceInput{
			SourceText: productionText, RawSourceText: fetched.OriginalRaw, ReplaceSource: true,
			SourceTitle: book.Title, SourceMetadata: metadata, ExpectedRevision: book.Revision,
		})
		if err != nil {
			return book, fmt.Errorf("保存重新处理的正文失败：%w", err)
		}
		if processErr != nil {
			return updated, fmt.Errorf("重新处理正文失败：%w", processErr)
		}
		return updated, nil
	}
	if len(productionLines(book)) > 0 {
		gender := firstMetadataString(book.SourceMetadata, "gender", "gender_name")
		style := firstMetadataString(book.SourceMetadata, "style", "styleName", "style_name")
		classifyStatus := strings.ToLower(strings.TrimSpace(stringValue(book.SourceMetadata["classifyStatus"])))
		if gender != "" && style != "" && (classifyStatus == "classified" || classifyStatus == "manual") {
			return book, nil
		}
		info := novelfetchworkshop.BookInfo{
			Category: firstMetadataString(book.SourceMetadata, "category"),
			Genre:    firstMetadataString(book.SourceMetadata, "genre"),
			Themes:   firstMetadataString(book.SourceMetadata, "themes", "theme"),
			Labels:   firstMetadataString(book.SourceMetadata, "labels", "tags"),
		}
		metadata := finalizeFetchedBatchClassification(classifyFetchedBatchBook(ctx, owner, info, book.SourceText, book.SourceMetadata, c.classification))
		updated, err := c.store.UpdateBookMetadata(ctx, owner, batchID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: book.Revision})
		if err != nil {
			return book, fmt.Errorf("保存分类状态失败：%w", err)
		}
		if metadata["classifyStatus"] != "classified" {
			return updated, fmt.Errorf("男女频和风格分类未完成：%s", stringValue(metadata["classifyError"]))
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
	fetched, metadata, processErr := prepareBatchFetchedSource(ctx, c.ttsStore, owner, bookID, fetched, c.classification)
	for key, value := range book.SourceMetadata {
		if _, exists := metadata[key]; !exists {
			metadata[key] = value
		}
	}
	productionText := fetched.ProcessedText
	if processErr != nil {
		productionText = ""
	}
	updated, err := c.store.CaptureBookSource(ctx, owner, batchID, book.ID, batchfactoryv11.CaptureBookSourceInput{
		SourceText: productionText, RawSourceText: fetched.OriginalRaw, ReplaceSource: true, SourceTitle: book.Title, SourceMetadata: metadata, ExpectedRevision: book.Revision,
	})
	if err != nil {
		if errors.Is(err, batchfactoryv11.ErrConflict) {
			latest, readErr := c.store.GetBatch(ctx, owner, batchID)
			if readErr == nil {
				for _, candidate := range latest.Books {
					if candidate.ID == book.ID && len(productionLines(candidate)) > 0 {
						return candidate, nil
					}
				}
			}
		}
		return book, fmt.Errorf("保存自动获取正文失败：%w", err)
	}
	if processErr != nil {
		return updated, fmt.Errorf("自动处理获取正文失败：%w", processErr)
	}
	return updated, nil
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
		durationMS, err := probe.DurationMS(ctx, audio)
		if err != nil {
			return "", 0, fmt.Errorf("TTS 时长测量失败：%w", err)
		}
		durations = append(durations, float64(durationMS)/1000)
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
		_, err = c.store.SaveSettings(ctx, owner, batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBook, BatchID: batchID, BookID: book.ID}, batchfactoryv11.SettingsUpdate{
			Patch: batchfactoryv11.SettingsPatch{
				"audioPlanningEnabled": enabledRaw,
				"audioDurationSeconds": totalRaw,
				"audioLineDurations":   lineRaw,
				"ttsFingerprint":       json.RawMessage(fmt.Sprintf("%q", func() string { digest := hash.Sum(nil); return fmt.Sprintf("tts-%x", digest[:12]) }())),
				"ttsAudioArtifactId":   json.RawMessage(fmt.Sprintf("%q", artifactID)),
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
	books := job["books"].(map[string]any)
	concurrency := automationNumberValue(job["concurrency"], 1)
	if concurrency > 4 {
		concurrency = 4
	}
	var failed atomic.Int32
	var workers sync.WaitGroup
	var jobMu sync.Mutex
	semaphore := make(chan struct{}, concurrency)
	for _, book := range batch.Books {
		book := book
		initialEntry, _ := books[book.ID].(map[string]any)
		if initialEntry == nil {
			initialEntry = map[string]any{}
		}
		entry := cloneAutomationEntry(initialEntry)
		workers.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer workers.Done()
			defer func() { <-semaphore }()
			// The one-item loop lets every existing stage exit its own book with
			// continue while the other book workers keep running.
			for range [1]struct{}{} {
				persist := func() {
					jobMu.Lock()
					clear(initialEntry)
					for key, value := range entry {
						initialEntry[key] = value
					}
					books[book.ID] = initialEntry
					_ = c.setJob(owner, batchID, jobID, job)
					jobMu.Unlock()
				}
				// 单本书的执行日志供前端展开查看；并发出图时多个 goroutine 会
				// 同时上报阶段与去敏、审核拒绝、降级重试事件，因此单独加锁，
				// 保证日志追加与持久化不并发写同一个 entry。
				var entryLogMu sync.Mutex
				recordComicEvent := func(stage, message string) {
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
					publishCtx, cancel := context.WithTimeout(ctx, stageTimeout)
					var submitErr error
					if comicSettingString(batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch), "productionMode", "video") == "comic" {
						collection, collectionErr := c.latestComicCollection(publishCtx, owner, batchID, book.ID, automationStringValue(entry["comicCollectionId"], ""))
						if collectionErr != nil {
							submitErr = collectionErr
						} else if c.comicPublish == nil {
							submitErr = fmt.Errorf("N8 漫画发布服务未配置")
						} else {
							_, submitErr = c.comicPublish(publishCtx, owner, batchID, book.ID, collection, batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch))
						}
					} else {
						submitErr = c.publishVideo(publishCtx, owner, batchID, book.ID)
					}
					cancel()
					if submitErr != nil {
						fail("submit", submitErr, "定时提交失败")
						continue
					}
					entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "published", "已按定时提交时间完成上传", c.now()
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
				if comicSettingString(effective, "productionMode", "video") == "comic" {
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
						stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
						_, bootstrapErr := c.stages.RunComicBootstrap(stageCtx, owner, batchID, book.ID)
						cancel()
						if bootstrapErr != nil {
							fail("comic_bootstrap", bootstrapErr, "漫画原文分页与提示词提取失败")
							continue
						}
						if identityRecovery {
							entry["stage"], entry["message"], entry["updatedAt"] = "comic_prompt_recovery", "人物身份已修复，正在重新生成全部漫画页提示词", c.now()
							persist()
							stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
							stageCtx = batchfactoryv11.WithComicPromptConcurrency(stageCtx, 5)
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
						stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
						stageCtx = batchfactoryv11.WithComicPromptConcurrency(stageCtx, 5)
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
					generationCtx, generationCancel := context.WithTimeout(ctx, comicBookDetachedGenerationTimeout)
					collection, generateErr := c.generateComicWithRetry(generationCtx, entry, persist, func(attemptCtx context.Context) (batchfactoryv11.ComicCollection, error) {
						generateCtx, cancel := context.WithTimeout(attemptCtx, 60*time.Minute)
						defer cancel()
						generateCtx = withComicGenerationProgress(generateCtx, recordComicEvent)
						generateCtx = withShuihuoImageEvent(generateCtx, recordComicEvent)
						generateCtx = withComicAssetConcurrency(generateCtx, 5)
						return c.comicGenerate(generateCtx, owner, batchID, book.ID)
					}, nil)
					generationCancel()
					if generateErr != nil {
						failureStage, failureDetail := comicGenerationFailureStage(generateErr)
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
						publishCtx, cancel := context.WithTimeout(ctx, stageTimeout)
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
							fail("comic_publish", publishErr, "漫画合集发布失败")
							continue
						}
						if record.Status == "uploaded" {
							entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "uploaded", "漫画图片已上传，已停止重复发布", c.now()
						} else {
							entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "published", "漫画合集已发布", c.now()
						}
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
				if !automationBoolValue(job["forceRegenerate"]) && automationStringValue(job["runMode"], "full_submit") != "storyboard_only" {
					status, statusErr := c.stages.Production.GetBatchStatus(ctx, owner, batchID)
					if statusErr != nil {
						fail("video", statusErr, "读取已生成视频失败")
						continue
					}
					production, reusedProduction = reusableAutomationProduction(status.Jobs, book)
				}
				delete(entry, "retryRequested")
				if reusedProduction {
					entry["stage"], entry["message"], entry["updatedAt"] = "video", "复用已完成视频，继续合成", c.now()
					persist()
				} else {
					entry["stage"] = "assets"
					entry["message"] = "正在提取资产"
					entry["updatedAt"] = c.now()
					stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
					_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageAssets, batchfactoryv11.StageModeForce, jobID+"-assets-"+book.ID, "")
					cancel()
					if stageErr != nil {
						fail("assets", stageErr, "资产提取失败")
						continue
					}
					entry["stage"], entry["message"] = "director", "正在生成分镜提示词"
					entry["updatedAt"] = c.now()
					persist()
					stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
					_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageDirector, batchfactoryv11.StageModeForce, jobID+"-director-"+book.ID, "")
					cancel()
					if stageErr != nil {
						fail("director", stageErr, "分镜提示词生成失败")
						continue
					}
					if automationStringValue(job["runMode"], "full_submit") == "storyboard_only" {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "ready_for_video", "已完成资产与分镜提示词", c.now()
						persist()
						continue
					}
					if automationOpeningEnabled(batch, book) {
						entry["stage"], entry["message"], entry["updatedAt"] = "opening", "正在基于分镜提示词生成换开头", c.now()
						persist()
						stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
						_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageOpening, batchfactoryv11.StageModeForce, jobID+"-opening-"+book.ID, "")
						cancel()
						if stageErr != nil {
							fail("opening", stageErr, "换开头提示词生成失败")
							continue
						}
					}
					entry["stage"], entry["message"], entry["updatedAt"] = "video", "正在提交视频生成", c.now()
					persist()
					stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
					_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageVideo, batchfactoryv11.StageModeMissing, jobID+"-video-"+book.ID, "")
					cancel()
					if stageErr != nil {
						fail("video", stageErr, "视频提交失败")
						continue
					}
					stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
					completedProduction, waitErr := c.waitProduction(stageCtx, owner, batchID, book.ID, jobID+"-video-"+book.ID)
					cancel()
					if waitErr != nil {
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
				if automationBoolValue(job["autoPublish"]) {
					if c.automationSubmitDeferred(job) {
						entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "scheduled", "waiting_submit", "视频已提前生成并合成，等待定时提交", c.now()
						persist()
						continue
					}
					entry["stage"], entry["message"], entry["updatedAt"] = "upload", "正在上传视频管理系统", c.now()
					persist()
					publishCtx, publishCancel := context.WithTimeout(ctx, stageTimeout)
					publishErr := c.publishVideo(publishCtx, owner, batchID, book.ID)
					publishCancel()
					if publishErr != nil {
						fail("upload", publishErr, "上传失败")
						continue
					}
				}
				entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "completed", "已完成视频生成、合成与上传", c.now()
				entry["productionTasks"] = len(production.Tasks)
				persist()
			}
		}()
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
		if entry != nil && automationStringValue(entry["status"], "") == "scheduled" {
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
const automationBookLogLimit = 100

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

func (c *BatchAutomationCoordinator) waitProduction(ctx context.Context, owner, batchID, bookID, requestID string) (batchfactoryv11.ProductionJob, error) {
	if c.stages == nil || c.stages.Production == nil {
		return batchfactoryv11.ProductionJob{}, fmt.Errorf("视频生产服务未配置")
	}
	for {
		status, err := c.stages.Production.GetBatchStatus(ctx, owner, batchID)
		if err != nil {
			return batchfactoryv11.ProductionJob{}, err
		}
		if job, ok := productionJobForRequest(status.Jobs, bookID, requestID); ok {
			if productionJobFailed(job) {
				return job, fmt.Errorf("视频任务失败：%s", productionJobError(job))
			}
			if len(job.Tasks) > 0 && productionJobComplete(job) {
				return job, nil
			}
		}
		select {
		case <-ctx.Done():
			return batchfactoryv11.ProductionJob{}, fmt.Errorf("视频生成超时或已停止")
		case <-time.After(2 * time.Second):
		}
	}
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
			entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", "assets", "重新执行失败阶段", "", now
			entry["retryRequested"] = true
			// 人工重试视为重新授权一轮自动重试预算；否则 3/3 用尽后的重试
			// 会被 scheduleComicContentPolicyRetry 直接判死，按钮等于失效。
			delete(entry, "contentPolicyRetryCount")
			delete(entry, "retryAt")
			delete(entry, "retryReason")
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
