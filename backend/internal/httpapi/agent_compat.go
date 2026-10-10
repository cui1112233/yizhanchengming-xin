package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RegisterAgentReadCompat serves the existing file-backed Agent task list and
// detail contract while the write/chat operations remain on their current
// bridge. It is intentionally read-only to avoid changing task data during
// the migration.
func RegisterAgentReadCompat(root *http.ServeMux, middleware BridgeAuth, usersRoot string) {
	if root == nil || strings.TrimSpace(usersRoot) == "" {
		return
	}
	readDocument := func(w http.ResponseWriter, req *http.Request) (map[string]any, bool) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return nil, false
		}
		username := strings.TrimSpace(identity.Username)
		if username == "" || strings.ContainsAny(username, `/\\`) || username == "." || username == ".." {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "用户无效"})
			return nil, false
		}
		path := filepath.Join(usersRoot, username, "agent-tasks.json")
		raw, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return map[string]any{"version": 1, "tasks": []any{}}, true
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return nil, false
		}
		var document map[string]any
		if json.Unmarshal(raw, &document) != nil || document == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Agent 任务数据无法读取"})
			return nil, false
		}
		return document, true
	}
	root.Handle("GET /api/agent/tasks", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		document, ok := readDocument(w, req)
		if !ok {
			return
		}
		tasks, _ := document["tasks"].([]any)
		summaries := make([]any, 0, len(tasks))
		for _, raw := range tasks {
			task, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			summary := map[string]any{}
			for _, key := range []string{"id", "title", "createdAt", "updatedAt", "messageCount", "preview"} {
				if value, exists := task[key]; exists {
					summary[key] = value
				}
			}
			if _, exists := summary["messageCount"]; !exists {
				if messages, ok := task["messages"].([]any); ok {
					summary["messageCount"] = len(messages)
				}
			}
			summaries = append(summaries, summary)
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": summaries})
	})))
	root.Handle("GET /api/agent/tasks/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		document, ok := readDocument(w, req)
		if !ok {
			return
		}
		id := req.PathValue("id")
		tasks, _ := document["tasks"].([]any)
		for _, raw := range tasks {
			task, ok := raw.(map[string]any)
			if ok && task["id"] == id {
				writeJSON(w, http.StatusOK, map[string]any{"task": task})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
	})))
}
