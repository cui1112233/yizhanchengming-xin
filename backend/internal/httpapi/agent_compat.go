package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
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
	root.Handle("POST /api/agent/tasks", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		task := map[string]any{"id": "agent-go-" + strings.ReplaceAll(now, ":", ""), "title": "新聊天", "titleMode": "automatic", "createdAt": now, "updatedAt": now, "messages": []any{}, "canvas": map[string]any{"version": 1, "revision": 0, "nodes": []any{}, "edges": []any{}}, "executions": []any{}}
		tasks, _ := document["tasks"].([]any)
		document["tasks"] = append([]any{task}, tasks...)
		if err := saveAgentDocument(usersRoot, identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存 Agent 任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task})
	})))
	root.Handle("PATCH /api/agent/tasks/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			Title string `json:"title"`
		}
		if json.NewDecoder(req.Body).Decode(&body) != nil || strings.TrimSpace(body.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "标题不能为空"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		updated := false
		for _, raw := range documentTasks(document) {
			if task, ok := raw.(map[string]any); ok && task["id"] == req.PathValue("id") {
				task["title"] = strings.TrimSpace(body.Title)
				task["titleMode"] = "manual"
				task["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
				updated = true
				writeJSON(w, http.StatusOK, map[string]any{"task": task})
				break
			}
		}
		if !updated {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
			return
		}
		if err := saveAgentDocument(usersRoot, identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存 Agent 任务失败"})
		}
	})))
	root.Handle("DELETE /api/agent/tasks/{id}/messages", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		updateAgentTask(usersRoot, w, req, func(task map[string]any) { task["messages"] = []any{} })
	})))
	root.Handle("DELETE /api/agent/tasks/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		id := req.PathValue("id")
		filtered := make([]any, 0)
		deleted := false
		for _, raw := range documentTasks(document) {
			task, ok := raw.(map[string]any)
			if ok && task["id"] == id {
				deleted = true
				continue
			}
			filtered = append(filtered, raw)
		}
		if !deleted {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
			return
		}
		document["tasks"] = filtered
		if err := saveAgentDocument(usersRoot, identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存 Agent 任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})))
	root.Handle("GET /api/agent/tasks/{id}/canvas", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		if task := findAgentTask(document, req.PathValue("id")); task != nil {
			canvas, _ := task["canvas"].(map[string]any)
			if canvas == nil {
				canvas = map[string]any{"version": 1, "revision": 0, "nodes": []any{}, "edges": []any{}}
			}
			writeJSON(w, http.StatusOK, map[string]any{"canvas": canvas})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
	})))
	root.Handle("PATCH /api/agent/tasks/{id}/canvas", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			Canvas map[string]any `json:"canvas"`
		}
		if json.NewDecoder(req.Body).Decode(&body) != nil || body.Canvas == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "画布数据无效"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		task := findAgentTask(document, req.PathValue("id"))
		if task == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
			return
		}
		body.Canvas["revision"] = numberOr(body.Canvas["revision"], 0)
		task["canvas"] = body.Canvas
		task["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		if err := saveAgentDocument(usersRoot, identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存 Agent 画布失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": body.Canvas})
	})))
	root.Handle("GET /api/agent/tasks/{id}/executions", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		if task := findAgentTask(document, req.PathValue("id")); task != nil {
			executions, _ := task["executions"].([]any)
			writeJSON(w, http.StatusOK, map[string]any{"executions": executions})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
	})))
}

func RegisterAgentChatCompat(root *http.ServeMux, middleware BridgeAuth, usersRoot string, provider batchfactoryv11.DirectorProvider) {
	if root == nil || provider == nil || strings.TrimSpace(usersRoot) == "" {
		return
	}
	root.Handle("POST /api/agent/chat", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			TaskID  string         `json:"taskId"`
			Prompt  string         `json:"prompt"`
			Context map[string]any `json:"context"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 256<<10)).Decode(&body) != nil || strings.TrimSpace(body.Prompt) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入消息"})
			return
		}
		document, err := loadAgentDocument(usersRoot, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
			return
		}
		task := findAgentTask(document, body.TaskID)
		if task == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
			return
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		messages, _ := task["messages"].([]any)
		userMessage := map[string]any{"role": "user", "content": strings.TrimSpace(body.Prompt), "createdAt": now}
		contextText, _ := json.Marshal(body.Context)
		completion, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{
			SystemPrompt: "你是前贴平台的中文创作助手。请用中文直接回答，不能声称已经替用户执行外部操作。",
			UserPrompt:   strings.TrimSpace(body.Prompt) + "\n\n页面上下文：\n" + string(contextText),
			MaxTokens:    2048,
			Temperature:  0.7,
			DisableJSONResponse: true,
		})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "文本模型请求失败"})
			return
		}
		assistantMessage := map[string]any{"role": "assistant", "content": strings.TrimSpace(completion), "createdAt": time.Now().UTC().Format(time.RFC3339Nano)}
		task["messages"] = append(messages, userMessage, assistantMessage)
		task["updatedAt"] = assistantMessage["createdAt"]
		if err := saveAgentDocument(usersRoot, identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存 Agent 回复失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task, "user": userMessage, "assistant": assistantMessage})
	})))
}

func documentTasks(document map[string]any) []any {
	tasks, _ := document["tasks"].([]any)
	return tasks
}

func findAgentTask(document map[string]any, id string) map[string]any {
	for _, raw := range documentTasks(document) {
		if task, ok := raw.(map[string]any); ok && task["id"] == id {
			return task
		}
	}
	return nil
}

func numberOr(value any, fallback int) int {
	switch number := value.(type) {
	case float64:
		return int(number)
	case int:
		return number
	default:
		return fallback
	}
}

func loadAgentDocument(root, username string) (map[string]any, error) {
	username = strings.TrimSpace(username)
	if username == "" || strings.ContainsAny(username, `/\\`) || username == "." || username == ".." {
		return nil, os.ErrInvalid
	}
	path := filepath.Join(root, username, "agent-tasks.json")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{"version": 1, "tasks": []any{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil || document == nil {
		return nil, os.ErrInvalid
	}
	if _, ok := document["tasks"].([]any); !ok {
		document["tasks"] = []any{}
	}
	return document, nil
}

func saveAgentDocument(root, username string, document map[string]any) error {
	path := filepath.Join(root, strings.TrimSpace(username), "agent-tasks.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func updateAgentTask(root string, w http.ResponseWriter, req *http.Request, update func(map[string]any)) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	document, err := loadAgentDocument(root, identity.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取 Agent 任务失败"})
		return
	}
	for _, raw := range documentTasks(document) {
		if task, ok := raw.(map[string]any); ok && task["id"] == req.PathValue("id") {
			update(task)
			task["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
			if err := saveAgentDocument(root, identity.Username, document); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存 Agent 任务失败"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"task": task})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
}
