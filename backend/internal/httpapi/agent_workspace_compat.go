package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	promptcatalog "qiantie/backend/internal/prompt"
)

var agentSensitiveContextKey = regexp.MustCompile(`(?i)token|secret|password|authorization|credential|api.?key`)

func RegisterAgentWorkspaceCompat(root *http.ServeMux, middleware BridgeAuth, store agentworkspace.Store, legacyRoot string, provider batchfactoryv11.DirectorProvider, clock func() time.Time) {
	if root == nil || store == nil {
		return
	}
	if clock == nil {
		clock = time.Now
	}
	var importMu sync.Mutex
	ownerFor := func(w http.ResponseWriter, req *http.Request) (string, bool) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return "", false
		}
		owner := strings.TrimSpace(identity.Username)
		if strings.TrimSpace(legacyRoot) != "" {
			importMu.Lock()
			_, err := agentworkspace.ImportLegacyFile(req.Context(), store, owner, filepath.Join(legacyRoot, owner, "agent-tasks.json"), clock().UTC())
			importMu.Unlock()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "导入旧 Agent 项目失败"})
				return "", false
			}
		}
		return owner, true
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			owner, ok := ownerFor(w, req)
			if ok {
				fn(w, req, owner)
			}
		})))
	}

	handle("GET /api/agent/tasks", func(w http.ResponseWriter, req *http.Request, owner string) {
		projects, err := store.ListProjects(req.Context(), owner)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": projects})
	})
	handle("GET /api/agent/tasks/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		project, err := store.Project(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": project})
	})
	handle("POST /api/agent/tasks", func(w http.ResponseWriter, req *http.Request, owner string) {
		now := clock().UTC()
		project := agentworkspace.Project{ID: newAgentID("agent"), Title: "新项目", TitleMode: "automatic", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.EmptyCanvas()}
		if err := store.CreateProject(req.Context(), owner, project); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		created, err := store.Project(req.Context(), owner, project.ID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": created})
	})
	handle("PATCH /api/agent/tasks/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Title string `json:"title"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&body) != nil || strings.TrimSpace(body.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "标题不能为空"})
			return
		}
		project, err := store.RenameProject(req.Context(), owner, req.PathValue("id"), body.Title, clock().UTC())
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": project})
	})
	handle("DELETE /api/agent/tasks/{id}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.ClearMessages(req.Context(), owner, req.PathValue("id"), clock().UTC()); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	handle("DELETE /api/agent/tasks/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteProject(req.Context(), owner, req.PathValue("id"), clock().UTC()); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	handle("GET /api/agent/tasks/{id}/canvas", func(w http.ResponseWriter, req *http.Request, owner string) {
		canvas, err := store.Canvas(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": canvas})
	})
	handle("PATCH /api/agent/tasks/{id}/canvas", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Canvas   agentworkspace.Canvas `json:"canvas"`
			Revision int64                 `json:"revision"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 2<<20)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "画布数据无效"})
			return
		}
		canvas, err := store.SaveCanvas(req.Context(), owner, req.PathValue("id"), body.Canvas, body.Revision, clock().UTC())
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": canvas})
	})
	handle("GET /api/agent/tasks/{id}/executions", func(w http.ResponseWriter, req *http.Request, owner string) {
		executions, err := store.Executions(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"executions": executions})
	})

	if provider != nil {
		handle("POST /api/agent/chat", func(w http.ResponseWriter, req *http.Request, owner string) {
			var body struct {
				TaskID  string         `json:"taskId"`
				Prompt  string         `json:"prompt"`
				Context map[string]any `json:"context"`
			}
			if json.NewDecoder(http.MaxBytesReader(w, req.Body, 256<<10)).Decode(&body) != nil || strings.TrimSpace(body.Prompt) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入消息"})
				return
			}
			if _, err := store.Project(req.Context(), owner, body.TaskID); err != nil {
				writeAgentStoreError(w, err)
				return
			}
			cleanContext := sanitizeAgentContext(body.Context)
			contextJSON, _ := json.Marshal(cleanContext)
			completion, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{
				SystemPrompt: promptcatalog.AgentWorkspaceSystemPrompt,
				UserPrompt:   strings.TrimSpace(body.Prompt) + "\n\n页面上下文：\n" + string(contextJSON),
				MaxTokens:    2048, Temperature: 0.7, DisableJSONResponse: true,
			})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "文本模型请求失败"})
				return
			}
			now := clock().UTC()
			userMessage := agentworkspace.Message{Role: "user", Content: strings.TrimSpace(body.Prompt), Context: cleanContext, CreatedAt: now}
			assistantMessage := agentworkspace.Message{Role: "assistant", Content: strings.TrimSpace(completion), CreatedAt: now}
			project, err := store.AppendExchange(req.Context(), owner, body.TaskID, userMessage, assistantMessage)
			if err != nil {
				writeAgentStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"task": project, "user": userMessage, "assistant": assistantMessage})
		})
	}
}

func newAgentID(prefix string) string {
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return prefix + "-" + strings.ReplaceAll(time.Now().UTC().Format("20060102T150405.000000000"), ".", "")
	}
	return prefix + "-" + hex.EncodeToString(raw)
}

func sanitizeAgentContext(input map[string]any) map[string]any {
	clean, _ := sanitizeAgentValue(input, 0).(map[string]any)
	if clean == nil {
		return map[string]any{}
	}
	return clean
}

func sanitizeAgentValue(value any, depth int) any {
	if depth > 6 {
		return nil
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any)
		for key, child := range typed {
			if !agentSensitiveContextKey.MatchString(key) {
				out[key] = sanitizeAgentValue(child, depth+1)
			}
		}
		return out
	case []any:
		limit := len(typed)
		if limit > 50 {
			limit = 50
		}
		out := make([]any, 0, limit)
		for _, child := range typed[:limit] {
			out = append(out, sanitizeAgentValue(child, depth+1))
		}
		return out
	case string:
		if len(typed) > 4000 {
			return typed[:4000]
		}
		return typed
	case float64, bool, nil:
		return typed
	default:
		return nil
	}
}

func writeAgentStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrProjectNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
	case errors.Is(err, agentworkspace.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "画布已在其他位置更新，请重新加载"})
	case errors.Is(err, agentworkspace.ErrProjectExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "项目已存在"})
	default:
		log.Printf("agent workspace store error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Agent 数据保存失败"})
	}
}
