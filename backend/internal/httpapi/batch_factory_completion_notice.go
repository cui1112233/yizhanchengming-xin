package httpapi

import (
	"encoding/json"
	"net/http"
	"sort"
)

func batchAutomationCompletionSummary(job map[string]any) map[string]any {
	books, _ := job["books"].(map[string]any)
	total, ready := len(books), 0
	for _, raw := range books {
		book, _ := raw.(map[string]any)
		if book != nil && book["status"] == "completed" {
			ready++
		}
	}
	stamp := job["completedAt"]
	if stamp == nil {
		stamp = job["updatedAt"]
	}
	encoded, _ := json.Marshal(stamp)
	return map[string]any{
		"id": job["id"], "batchId": job["batchId"], "state": job["state"], "updatedAt": job["updatedAt"],
		"total": total, "ready": ready, "success": job["state"] == "completed" && total > 0 && ready == total,
		"completionKey": automationStringValue(job["id"], "") + ":" + string(encoded),
	}
}

func (c *BatchAutomationCoordinator) CompletionNotices(owner string) ([]map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return nil, err
	}
	jobs := []map[string]any{}
	for _, job := range state.Jobs {
		if job["owner"] == owner {
			jobs = append(jobs, job)
		}
	}
	sort.SliceStable(jobs, func(i, j int) bool { return automationJobTime(jobs[i]).After(automationJobTime(jobs[j])) })
	result := []map[string]any{}
	terminalCount := 0
	for _, job := range jobs {
		status := automationStringValue(job["state"], "")
		if status == "completed" || status == "failed" || status == "stopped" {
			terminalCount++
			if terminalCount > 100 {
				continue
			}
		}
		result = append(result, batchAutomationCompletionSummary(job))
	}
	return result, nil
}

func registerBatchCompletionNoticesRoute(mux *http.ServeMux, auth BridgeAuth, c *BatchAutomationCoordinator, version string) {
	mux.Handle("GET /api/batch-factory/"+version+"/automation/notifications", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		jobs, err := c.CompletionNotices(owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
	})))
}
