package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	var state struct {
		Jobs map[string]json.RawMessage `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("decode automation status: %w", err)
	}
	for _, encoded := range state.Jobs {
		var job map[string]any
		if json.Unmarshal(encoded, &job) != nil {
			continue
		}
		if strings.TrimSpace(toString(job["owner"])) != strings.TrimSpace(owner) || strings.TrimSpace(toString(job["batchId"])) != strings.TrimSpace(batchID) {
			continue
		}
		return publicAutomationJob(job), nil
	}
	return idle, nil
}

func idleAutomationCounts() map[string]any {
	return map[string]any{"total": 0, "ready": 0, "running": 0, "pending": 0, "failed": 0, "blocked": 0}
}

func publicAutomationJob(job map[string]any) map[string]any {
	result := map[string]any{
		"id": job["id"], "batchId": job["batchId"], "state": job["state"],
		"stopAt": job["stopAt"], "autoPublish": job["autoPublish"],
		"runMode": job["runMode"], "concurrency": job["concurrency"], "preset": job["preset"],
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
			books = append(books, map[string]any{
				"bookId": book["bookId"], "title": book["title"], "status": book["status"], "stage": book["stage"],
				"message": book["message"], "error": book["error"], "updatedAt": book["updatedAt"],
				"attempts": book["attempts"], "retryCount": book["retryCount"], "retryAt": book["retryAt"],
				"retryRequested": book["retryRequested"],
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
