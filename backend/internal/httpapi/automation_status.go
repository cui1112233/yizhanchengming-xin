package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type FileAutomationStatusReader struct {
	path string
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

func (r *FileAutomationStatusReader) Status(owner, batchID string) (map[string]any, error) {
	idle := map[string]any{"state": "idle", "counts": idleAutomationCounts()}
	if r == nil || r.path == "" {
		return idle, nil
	}
	raw, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return idle, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read automation status: %w", err)
	}
	var state automationStateFile
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("decode automation status: %w", err)
	}
	if normalizeAlreadyUploadedAutomationState(&state) {
		if err := writeAutomationStateFile(r.path, state); err != nil {
			return nil, fmt.Errorf("repair automation status: %w", err)
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

// normalizeAlreadyUploadedAutomationState repairs old task records written by
// the pre-reconciliation publisher. N8 reports IMAGE_ALREADY_IN_CAROUSEL when
// the images are already accepted by a carousel, but the old code stored that
// response as a hard upload failure. Treating it as completed prevents a
// second upload and makes the live log reflect the actual remote result.
func normalizeAlreadyUploadedAutomationState(state *automationStateFile) bool {
	if state == nil {
		return false
	}
	changed := false
	for _, job := range state.Jobs {
		books, _ := job["books"].(map[string]any)
		if len(books) == 0 {
			continue
		}
		jobChanged := false
		for _, raw := range books {
			entry, _ := raw.(map[string]any)
			if entry == nil || automationStringValue(entry["status"], "") != "failed" {
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
		}
		if jobChanged {
			normalizeAutomationJobCounts(job, books)
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

func normalizeAutomationJobCounts(job map[string]any, books map[string]any) {
	failed, running, ready, pending := 0, 0, 0, 0
	for _, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		switch automationStringValue(entry["status"], "") {
		case "completed":
			ready++
		case "running":
			running++
		case "scheduled":
			pending++
		case "failed", "blocked", "stopped":
			failed++
		}
	}
	job["counts"] = map[string]any{"total": len(books), "ready": ready, "running": running, "pending": pending, "failed": failed, "blocked": 0}
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
		"counts": job["counts"],
	}
	if result["autoPublish"] == nil {
		result["autoPublish"] = false
	}
	if strings.TrimSpace(toString(result["runMode"])) == "" {
		result["runMode"] = "video_no_submit"
	}
	if result["counts"] == nil {
		result["counts"] = idleAutomationCounts()
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
			// retryAt is meaningful only while this book is still running.
			// Older failed jobs may retain the marker from their last attempt;
			// do not expose that stale timestamp as a future automatic retry.
			if status != "running" && !automationBoolValue(book["retryRequested"]) {
				retryAt = nil
			}
			books = append(books, map[string]any{
				"bookId": book["bookId"], "title": book["title"], "status": book["status"], "stage": book["stage"],
				"message": book["message"], "error": book["error"], "publishError": book["publishError"], "updatedAt": book["updatedAt"],
				"attempts": book["attempts"], "retryCount": book["retryCount"], "retryAt": retryAt,
				"retryRequested": book["retryRequested"], "logs": publicAutomationBookLogs(book["logs"], 50),
			})
		}
	}
	result["books"] = books
	return result
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
