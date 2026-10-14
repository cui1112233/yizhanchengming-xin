package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// BatchAutomationCoordinator is the Go owner of the V12 start/stop controls.
// It intentionally persists the same public job shape that the old worker
// used, so the existing workbench does not need a second status protocol.
type BatchAutomationCoordinator struct {
	store   batchfactoryv11.Store
	presets AutomationPresetStore
	stages  *batchfactoryv11.BookStageService
	path    string
	now     func() time.Time
	mu      sync.Mutex
}

type automationStateFile struct {
	Version int                       `json:"version"`
	Jobs    map[string]map[string]any `json:"jobs"`
}

func NewBatchAutomationCoordinator(store batchfactoryv11.Store, presets AutomationPresetStore, stages *batchfactoryv11.BookStageService, reader *FileAutomationStatusReader, now func() time.Time) *BatchAutomationCoordinator {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	path := ""
	if reader != nil {
		path = reader.Path()
	}
	return &BatchAutomationCoordinator{store: store, presets: presets, stages: stages, path: path, now: now}
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
	jobID := fmt.Sprintf("go-auto-%d", c.now().UnixNano())
	job := map[string]any{
		"id": jobID, "owner": owner, "batchId": batchID, "state": "running",
		"runMode": automationStringValue(input["runMode"], "full_submit"), "autoPublish": automationBoolValue(input["autoPublish"]),
		"concurrency": automationNumberValue(input["concurrency"], 1), "preset": input["preset"],
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
	go c.run(context.Background(), owner, batchID, jobID, job, batch)
	return publicAutomationJob(job), nil
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

func (c *BatchAutomationCoordinator) run(ctx context.Context, owner, batchID, jobID string, job map[string]any, batch batchfactoryv11.Batch) {
	// Each external stage must have a bounded lifetime.  The UI should receive a
	// failed state instead of leaving a book stuck at "running" when an upstream
	// model endpoint stops responding.
	const stageTimeout = 2 * time.Minute
	books := job["books"].(map[string]any)
	failed := 0
	for _, book := range batch.Books {
		entry := books[book.ID].(map[string]any)
		entry["stage"] = "assets"
		entry["message"] = "正在提取资产"
		entry["updatedAt"] = c.now()
		stageCtx, cancel := context.WithTimeout(ctx, stageTimeout)
		_, stageErr := c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageAssets, batchfactoryv11.StageModeForce, jobID+"-assets-"+book.ID, "")
		cancel()
		if stageErr != nil {
			entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", "assets", stageErr.Error(), "资产提取失败", c.now()
			failed++
			continue
		}
		entry["stage"], entry["message"] = "director", "正在生成分镜提示词"
		entry["updatedAt"] = c.now()
		stageCtx, cancel = context.WithTimeout(ctx, stageTimeout)
		_, stageErr = c.stages.Run(stageCtx, owner, batchID, book.ID, batchfactoryv11.BookStageDirector, batchfactoryv11.StageModeForce, jobID+"-director-"+book.ID, "")
		cancel()
		if stageErr != nil {
			entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", "director", stageErr.Error(), "分镜提示词生成失败", c.now()
			failed++
			continue
		}
		entry["status"], entry["stage"], entry["message"], entry["updatedAt"] = "completed", "director", "已完成资产与分镜提示词", c.now()
	}
	job["state"] = "completed"
	job["completedAt"], job["updatedAt"] = c.now(), c.now()
	job["counts"] = automationCounts(len(batch.Books), 0, failed, len(batch.Books)-failed)
	if failed > 0 {
		job["state"] = "failed"
	}
	_ = c.setJob(owner, batchID, jobID, job)
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
}
