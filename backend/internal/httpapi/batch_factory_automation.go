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
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/batchfactoryv11/external"
	"qiantie/backend/internal/novelfetchworkshop"
)

// BatchAutomationCoordinator is the Go owner of the V12 start/stop controls.
// It intentionally persists the same public job shape that the old worker
// used, so the existing workbench does not need a second status protocol.
type BatchAutomationCoordinator struct {
	store        batchfactoryv11.Store
	presets      AutomationPresetStore
	stages       *batchfactoryv11.BookStageService
	merge        *batchfactoryv11.MergeService
	external     *external.Service
	ttsStore     novelfetchworkshop.Store
	audioProbe   batchfactoryv11.H3AudioDurationProbe
	path         string
	now          func() time.Time
	providerSync func(context.Context, string) error
	publish121   *BatchFactory121Publisher
	mu           sync.Mutex
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

func automationCounts(total, running, failed, ready int) map[string]any {
	return map[string]any{"total": total, "ready": ready, "running": running, "pending": 0, "failed": failed, "blocked": 0}
}

func (c *BatchAutomationCoordinator) Start(ctx context.Context, owner, batchID string, input map[string]any) (map[string]any, error) {
	if c == nil || c.store == nil || c.stages == nil {
		return nil, fmt.Errorf("%w: Go 自动化调度器未配置", batchfactoryv11.ErrUnavailable)
	}
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" {
		return nil, batchfactoryv11.ErrInvalid
	}
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
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
	batch, err = c.applyPreset(ctx, owner, batch, preset)
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
	// full_submit is the user-facing “生成并提交” mode. Keep the server
	// authoritative so older frontends that omitted autoPublish cannot silently
	// finish without the upload step.
	if runMode == "full_submit" {
		autoPublish = true
	}
	job := map[string]any{
		"id": jobID, "owner": owner, "batchId": batchID, "state": "running",
		"runMode": runMode, "autoPublish": autoPublish,
		"concurrency": automationNumberValue(input["concurrency"], 1), "preset": map[string]any{"id": preset.ID, "name": preset.Name, "version": preset.Version},
		"createdAt": c.now(), "startedAt": c.now(), "updatedAt": c.now(),
		"counts": automationCounts(len(batch.Books), len(batch.Books), 0, 0), "books": map[string]any{},
	}
	books := job["books"].(map[string]any)
	for _, book := range batch.Books {
		books[book.ID] = map[string]any{"bookId": book.ID, "title": book.Title, "status": "running", "stage": "assets", "message": "Go 自动化已接管", "attempts": 0, "retryCount": 0, "updatedAt": c.now()}
	}
	if err := c.setJob(owner, batchID, jobID, job); err != nil {
		return nil, err
	}
	// The worker runs after the HTTP request has returned, so preserve the
	// authenticated owner in its context. Account-scoped text/video providers
	// must not fall back to an anonymous, unconfigured context in background
	// stages.
	c.launchRun(owner, batchID, jobID, job, batch)
	return publicAutomationJob(job), nil
}

func (c *BatchAutomationCoordinator) applyPreset(ctx context.Context, owner string, batch batchfactoryv11.Batch, preset *AutomationPreset) (batchfactoryv11.Batch, error) {
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
	automationCtx := context.WithValue(context.Background(), bridgeIdentityKey{}, BridgeIdentity{Username: owner})
	go c.run(automationCtx, owner, batchID, jobID, job, batch)
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

// ensureAudioPlan is deliberately independent of the final audioSource switch.
// The switch only chooses which track is used during the final merge; every
// production run still creates TTS and measures every selected source line so
// the storyboard timeline is based on real durations.
func (c *BatchAutomationCoordinator) ensureAudioPlan(ctx context.Context, owner, batchID string, book batchfactoryv11.Book) error {
	if c == nil || c.ttsStore == nil || c.store == nil {
		return fmt.Errorf("配音与逐行时长服务未配置")
	}
	if hasReusableAudioPlan(book) {
		return nil
	}
	lines := productionLines(book)
	if len(lines) == 0 {
		return fmt.Errorf("正文没有可测量的生产行")
	}
	model, ok := configuredTTSModel(ctx, c.ttsStore, owner)
	probe := c.audioProbe
	if probe == nil {
		probe = batchfactoryv11.FFprobeH3AudioDuration{}
	}
	durations := make([]float64, 0, len(lines))
	hash := sha256.New()
	for _, line := range lines {
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
			return fmt.Errorf("TTS 生成失败：%w", err)
		}
		durationMS, err := probe.DurationMS(ctx, audio)
		if err != nil {
			return fmt.Errorf("TTS 时长测量失败：%w", err)
		}
		durations = append(durations, float64(durationMS)/1000)
		_, _ = hash.Write(audio)
	}
	total := 0.0
	for _, duration := range durations {
		total += duration
	}
	latest, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return err
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
			},
			ExpectedRevision: candidate.Revision,
		})
		return err
	}
	return batchfactoryv11.ErrNotFound
}

func (c *BatchAutomationCoordinator) run(ctx context.Context, owner, batchID, jobID string, job map[string]any, batch batchfactoryv11.Batch) {
	// Each external stage must have a bounded lifetime.  The UI should receive a
	// failed state instead of leaving a book stuck at "running" when an upstream
	// model endpoint stops responding.
	const stageTimeout = 10 * time.Minute
	books := job["books"].(map[string]any)
	failed := 0
	for _, book := range batch.Books {
		entry := books[book.ID].(map[string]any)
		persist := func() { _ = c.setJob(owner, batchID, jobID, job) }
		fail := func(stage string, err error, message string) {
			entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", stage, err.Error(), message, c.now()
			failed++
			persist()
		}
		if entryStatus := automationStringValue(entry["status"], ""); entryStatus == "completed" {
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
		entry["stage"] = "audio_planning"
		entry["message"] = "正在逐行生成配音并测量时长"
		entry["updatedAt"] = c.now()
		persist()
		stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
		stageErr := c.ensureAudioPlan(stageCtx, owner, batchID, book)
		cancel()
		if stageErr != nil {
			fail("audio_planning", stageErr, "配音与逐行时长失败")
			continue
		}
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
		production, waitErr := c.waitProduction(stageCtx, owner, batchID, book.ID, jobID+"-video-"+book.ID)
		cancel()
		if waitErr != nil {
			fail("video", waitErr, "视频生成失败")
			continue
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
		mergeJobs, mergeErr := c.merge.SubmitBookMerge(mergeCtx, owner, batchID, book.ID, jobID+"-merge-"+book.ID, batchfactoryv11.MergeOptions{TimingMode: "audio", AudioDurationSeconds: productionAudioDuration(production)})
		if mergeErr == nil {
			_, mergeErr = c.waitMerge(mergeCtx, owner, batchID, book.ID, mergeJobs)
		}
		mergeCancel()
		if mergeErr != nil {
			fail("merge", mergeErr, "最终合成失败")
			continue
		}
		if automationBoolValue(job["autoPublish"]) {
			entry["stage"], entry["message"], entry["updatedAt"] = "upload", "正在上传视频管理系统", c.now()
			persist()
			if c.publish121 != nil {
				publishCtx, publishCancel := context.WithTimeout(ctx, stageTimeout)
				_, publishErr := c.publish121.Submit(publishCtx, owner, batchID, book.ID, map[string]any{})
				publishCancel()
				if publishErr != nil {
					fail("upload", publishErr, "上传失败")
					continue
				}
			} else if c.external == nil {
				fail("upload", fmt.Errorf("视频管理系统上传服务未配置"), "上传失败")
				continue
			} else {
				publishCtx, publishCancel := context.WithTimeout(ctx, stageTimeout)
				intent, publishErr := c.external.CreateIntent(publishCtx, owner, string(external.Provider121), batchID, book.ID, json.RawMessage(`{}`))
				if publishErr == nil {
					_, publishErr = c.external.ConfirmIntent(publishCtx, owner, intent.ID)
				}
				if publishErr == nil {
					_, publishErr = c.external.Submit(publishCtx, owner, string(external.Provider121), intent.ID)
				}
				publishCancel()
				if publishErr != nil {
					fail("upload", publishErr, "上传失败")
					continue
				}
			}
		}
		entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "completed", "已完成视频生成、合成与上传", c.now()
		entry["productionTasks"] = len(production.Tasks)
		persist()
	}
	job["state"] = "completed"
	job["completedAt"], job["updatedAt"] = c.now(), c.now()
	job["counts"] = automationCounts(len(batch.Books), 0, failed, len(batch.Books)-failed)
	if failed > 0 {
		job["state"] = "failed"
	}
	_ = c.setJob(owner, batchID, jobID, job)
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

func productionAudioDuration(job batchfactoryv11.ProductionJob) float64 {
	var total float64
	for _, task := range job.Tasks {
		if task.ActualDurationSeconds > 0 {
			total += task.ActualDurationSeconds
		} else {
			total += task.TargetDurationSeconds
		}
	}
	return total
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
	for id, job := range state.Jobs {
		if automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
			continue
		}
		stateValue := automationStringValue(job["state"], "")
		if stateValue != "running" && stateValue != "paused" {
			continue
		}
		job["state"], job["updatedAt"], job["stopAt"] = "stopped", c.now(), c.now()
		job["lastError"] = "用户停止自动化"
		job["counts"] = automationCounts(0, 0, 0, 0)
		state.Jobs[id] = job
	}
	return c.saveLocked(state)
}

func (c *BatchAutomationCoordinator) Pause(owner, batchID string) (map[string]any, error) {
	return c.control(owner, batchID, "paused", nil)
}

func (c *BatchAutomationCoordinator) Resume(ctx context.Context, owner, batchID string) (map[string]any, error) {
	return c.control(owner, batchID, "running", func(job map[string]any) error {
		batch, err := c.store.GetBatch(ctx, owner, batchID)
		if err != nil {
			return err
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
	return c.control(owner, batchID, "running", func(job map[string]any) error {
		books, _ := job["books"].(map[string]any)
		for id, raw := range books {
			if len(set) > 0 && !set[id] {
				continue
			}
			entry, _ := raw.(map[string]any)
			status := automationStringValue(entry["status"], "")
			if status == "failed" || status == "blocked" || status == "paused" {
				entry["status"], entry["stage"], entry["message"], entry["error"], entry["updatedAt"] = "running", "assets", "重新执行失败阶段", "", c.now()
			}
		}
		batch, err := c.store.GetBatch(ctx, owner, batchID)
		if err != nil {
			return err
		}
		c.launchRun(owner, batchID, automationStringValue(job["id"], ""), job, batch)
		return nil
	})
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
