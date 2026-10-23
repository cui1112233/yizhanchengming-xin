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
		projectID := req.PathValue("id")
		canvas, err := store.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, canvas)})
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
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(req.PathValue("id"), canvas)})
	})
	handle("POST /api/agent/tasks/{id}/director/plans/{planId}/approve", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Revision int64 `json:"revision"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "导演方案确认数据无效"})
			return
		}
		projectID := req.PathValue("id")
		project, err := store.Project(req.Context(), owner, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		plan, ok := savedDirectorPlan(project.Messages, req.PathValue("planId"))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导演方案不存在或已失效"})
			return
		}
		canvas, err := agentworkspace.ApplyDirectorPlan(req.Context(), store, owner, projectID, body.Revision, plan, clock().UTC())
		if err != nil {
			writeDirectorPlanError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, canvas), "planId": plan.ID})
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
			project, err := store.Project(req.Context(), owner, body.TaskID)
			if err != nil {
				writeAgentStoreError(w, err)
				return
			}
			cleanContext := sanitizeAgentContext(body.Context)
			contextJSON, _ := json.Marshal(cleanContext)
			directorRequest := agentDirectorRequested(cleanContext)
			systemPrompt := promptcatalog.AgentWorkspaceSystemPrompt
			if directorRequest {
				systemPrompt = promptcatalog.AgentDirectorPlanSystemPrompt
			}
			completion, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{
				SystemPrompt: systemPrompt,
				UserPrompt:   strings.TrimSpace(body.Prompt) + "\n\n页面上下文：\n" + string(contextJSON),
				MaxTokens:    2048, Temperature: 0.7, DisableJSONResponse: !directorRequest,
			})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "文本模型请求失败"})
				return
			}
			now := clock().UTC()
			userMessage := agentworkspace.Message{Role: "user", Content: strings.TrimSpace(body.Prompt), Context: cleanContext, CreatedAt: now}
			assistantMessage := agentworkspace.Message{Role: "assistant", Content: strings.TrimSpace(completion), CreatedAt: now}
			var automaticallyAppliedCanvas *agentworkspace.Canvas
			if directorRequest {
				assistantMessage = directorAssistantMessage(project, strings.TrimSpace(completion), now)
				if agentDirectorAutomaticRequested(cleanContext) {
					if plan, ok := directorPlanFromMessage(assistantMessage); ok {
						canvas, applyErr := agentworkspace.ApplyDirectorPlan(req.Context(), store, owner, body.TaskID, project.Canvas.Revision, plan, now)
						if applyErr == nil {
							plan.Status = "applied"
							assistantMessage.Context["directorPlan"] = plan
							assistantMessage.Content = strings.TrimSpace(assistantMessage.Content) + "\n\n已按你的自动执行要求应用到画布。"
							automaticallyAppliedCanvas = &canvas
						} else if errors.Is(applyErr, agentworkspace.ErrDirectorCommandNeedsExecution) || errors.Is(applyErr, agentworkspace.ErrDirectorMissingConfiguration) {
							assistantMessage.Content = strings.TrimSpace(assistantMessage.Content) + "\n\n这一步需要真实生成或补齐模型配置，已暂停等待你处理。"
						}
					}
				}
			}
			updatedProject, err := store.AppendExchange(req.Context(), owner, body.TaskID, userMessage, assistantMessage)
			if err != nil {
				writeAgentStoreError(w, err)
				return
			}
			response := map[string]any{"task": updatedProject, "user": userMessage, "assistant": assistantMessage}
			if automaticallyAppliedCanvas != nil {
				response["canvas"] = agentCanvasForResponse(body.TaskID, *automaticallyAppliedCanvas)
			}
			writeJSON(w, http.StatusOK, response)
		})
	}
}

func agentDirectorRequested(context map[string]any) bool {
	director, ok := context["director"].(map[string]any)
	if !ok {
		return false
	}
	enabled, _ := director["enabled"].(bool)
	return enabled
}

func agentDirectorAutomaticRequested(context map[string]any) bool {
	director, ok := context["director"].(map[string]any)
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(stringValue(director["mode"])), "automatic")
}

func directorPlanFromMessage(message agentworkspace.Message) (agentworkspace.DirectorPlan, bool) {
	if message.Context == nil {
		return agentworkspace.DirectorPlan{}, false
	}
	rawPlan, exists := message.Context["directorPlan"]
	if !exists {
		return agentworkspace.DirectorPlan{}, false
	}
	rawJSON, err := json.Marshal(rawPlan)
	if err != nil {
		return agentworkspace.DirectorPlan{}, false
	}
	var plan agentworkspace.DirectorPlan
	if err := json.Unmarshal(rawJSON, &plan); err != nil {
		return agentworkspace.DirectorPlan{}, false
	}
	return plan, true
}

func directorAssistantMessage(project agentworkspace.Project, completion string, now time.Time) agentworkspace.Message {
	message := agentworkspace.Message{Role: "assistant", Content: completion, CreatedAt: now}
	var response struct {
		Reply string                       `json:"reply"`
		Plan  *agentworkspace.DirectorPlan `json:"plan"`
	}
	if json.Unmarshal([]byte(completion), &response) != nil {
		message.Content = "导演没有生成可确认的方案。请补充你想修改的内容后再试。"
		return message
	}
	if strings.TrimSpace(response.Reply) != "" {
		message.Content = strings.TrimSpace(response.Reply)
	}
	if response.Plan == nil {
		return message
	}
	if strings.TrimSpace(response.Plan.ID) == "" {
		response.Plan.ID = newAgentID("director-plan")
	}
	if err := agentworkspace.ValidateDirectorPlan(project, *response.Plan); err != nil {
		message.Content = strings.TrimSpace(message.Content) + "\n\n这份方案未通过画布校验，请补充或调整需求。"
		return message
	}
	message.Context = map[string]any{"directorPlan": response.Plan}
	return message
}

func savedDirectorPlan(messages []agentworkspace.Message, planID string) (agentworkspace.DirectorPlan, bool) {
	planID = strings.TrimSpace(planID)
	if planID == "" {
		return agentworkspace.DirectorPlan{}, false
	}
	for index := len(messages) - 1; index >= 0; index-- {
		message := messages[index]
		if message.Role != "assistant" || message.Context == nil {
			continue
		}
		rawPlan, exists := message.Context["directorPlan"]
		if !exists {
			continue
		}
		rawJSON, err := json.Marshal(rawPlan)
		if err != nil {
			continue
		}
		var plan agentworkspace.DirectorPlan
		if json.Unmarshal(rawJSON, &plan) == nil && plan.ID == planID {
			return plan, true
		}
	}
	return agentworkspace.DirectorPlan{}, false
}

func writeDirectorPlanError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrDirectorCommandNeedsExecution):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "该方案包含生成或应用结果，请先通过对应任务入口执行"})
	case errors.Is(err, agentworkspace.ErrDirectorMissingConfiguration):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "方案中的节点缺少可用模型或通道配置"})
	case errors.Is(err, agentworkspace.ErrDirectorCommandForbidden):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "导演不能删除项目素材"})
	case errors.Is(err, agentworkspace.ErrDirectorCommandInvalid), errors.Is(err, agentworkspace.ErrDirectorCommandUnknown):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "导演方案无效或已过期"})
	default:
		writeAgentStoreError(w, err)
	}
}

func agentCanvasForResponse(projectID string, canvas agentworkspace.Canvas) agentworkspace.Canvas {
	canvas.Nodes = append([]agentworkspace.Node(nil), canvas.Nodes...)
	publicFields := []string{"id", "assetId", "projectId", "kind", "mediaType", "name", "contentType", "byteSize", "width", "height", "duration"}
	for index := range canvas.Nodes {
		ref := canvas.Nodes[index].AssetRef
		if len(ref) == 0 {
			continue
		}
		publicRef := make(map[string]any, len(publicFields)+1)
		for _, field := range publicFields {
			if value, exists := ref[field]; exists {
				publicRef[field] = value
			}
		}
		assetID, _ := publicRef["id"].(string)
		kind, _ := publicRef["kind"].(string)
		if strings.TrimSpace(assetID) != "" && (kind == "image" || kind == "video" || kind == "audio") {
			publicRef["url"] = agentAssetContentURL(projectID, strings.TrimSpace(assetID))
		}
		canvas.Nodes[index].AssetRef = publicRef
	}
	return canvas
}

func agentCanvasNodeByID(canvas agentworkspace.Canvas, nodeID string) (agentworkspace.Node, bool) {
	nodeID = strings.TrimSpace(nodeID)
	for _, node := range canvas.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return agentworkspace.Node{}, false
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
