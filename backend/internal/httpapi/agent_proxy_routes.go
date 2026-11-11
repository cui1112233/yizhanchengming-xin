package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/agentproxy"
)

// RegisterAgentProxyRoutes 挂载 Agent 代理最小骨架：工作区 → 会话 → 消息。
// 第一版只做持久化与权限档位校验，技能加载、工具循环、子代理委派尚未接入。
func RegisterAgentProxyRoutes(root *http.ServeMux, middleware BridgeAuth, store agentproxy.Store, clock func() time.Time) {
	if root == nil || store == nil {
		return
	}
	if clock == nil {
		clock = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("GET /api/agent/proxy/workspaces", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspaces, err := store.ListWorkspaces(req.Context(), owner)
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspaces": workspaces})
	})
	handle("POST /api/agent/proxy/workspaces", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Name           string `json:"name"`
			Description    string `json:"description"`
			PermissionMode string `json:"permissionMode"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body) != nil && req.ContentLength > 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "工作区数据无效"})
			return
		}
		mode, ok := requirePermissionMode(w, body.PermissionMode)
		if !ok {
			return
		}
		body.PermissionMode = mode
		name := strings.TrimSpace(body.Name)
		if name == "" {
			name = "新工作区"
		}
		workspace, err := store.CreateWorkspace(req.Context(), owner, agentproxy.CreateWorkspaceInput{
			ID: newAgentID("agentws"), Name: name, Description: body.Description, PermissionMode: body.PermissionMode,
		}, clock().UTC())
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspace": workspace})
	})
	handle("GET /api/agent/proxy/workspaces/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspace, err := store.Workspace(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspace": workspace})
	})
	handle("PATCH /api/agent/proxy/workspaces/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body agentproxy.UpdateWorkspaceInput
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "工作区数据无效"})
			return
		}
		if body.PermissionMode != nil {
			mode, ok := requirePermissionMode(w, *body.PermissionMode)
			if !ok {
				return
			}
			body.PermissionMode = &mode
		}
		workspace, err := store.UpdateWorkspace(req.Context(), owner, req.PathValue("id"), body, clock().UTC())
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspace": workspace})
	})
	handle("DELETE /api/agent/proxy/workspaces/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteWorkspace(req.Context(), owner, req.PathValue("id"), clock().UTC()); err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	handle("GET /api/agent/proxy/workspaces/{id}/sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspaceID := req.PathValue("id")
		if _, err := store.Workspace(req.Context(), owner, workspaceID); err != nil {
			writeAgentProxyError(w, err)
			return
		}
		sessions, err := store.ListSessions(req.Context(), owner, workspaceID)
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	})
	handle("POST /api/agent/proxy/workspaces/{id}/sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Title          string `json:"title"`
			PermissionMode string `json:"permissionMode"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body) != nil && req.ContentLength > 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "会话数据无效"})
			return
		}
		if strings.TrimSpace(body.PermissionMode) != "" {
			mode, ok := requirePermissionMode(w, body.PermissionMode)
			if !ok {
				return
			}
			body.PermissionMode = mode
		}
		title := strings.TrimSpace(body.Title)
		if title == "" {
			title = "新会话"
		}
		session, err := store.CreateSession(req.Context(), owner, req.PathValue("id"), agentproxy.CreateSessionInput{
			ID: newAgentID("agentsession"), Title: title, PermissionMode: body.PermissionMode,
		}, clock().UTC())
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session})
	})
	handle("GET /api/agent/proxy/workspaces/{id}/skills", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspaceID := req.PathValue("id")
		if _, err := store.Workspace(req.Context(), owner, workspaceID); err != nil {
			writeAgentProxyError(w, err)
			return
		}
		skills, err := store.ListSkills(req.Context(), owner, workspaceID)
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		// 技能包解析和加载在第一版没有实现，这里只返回表结构里的占位数据。
		writeJSON(w, http.StatusOK, map[string]any{"skills": skills, "skillsEnabled": false})
	})
	handle("GET /api/agent/proxy/workspaces/{id}/scheduled-jobs", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspaceID := req.PathValue("id")
		if _, err := store.Workspace(req.Context(), owner, workspaceID); err != nil {
			writeAgentProxyError(w, err)
			return
		}
		jobs, err := store.ListScheduledJobs(req.Context(), owner, workspaceID)
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		// 定时任务第一版只做列表展示，不调度、不执行。
		writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs, "schedulingEnabled": false})
	})

	handle("GET /api/agent/proxy/sessions/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		session, err := store.Session(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session})
	})
	handle("PATCH /api/agent/proxy/sessions/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body agentproxy.UpdateSessionInput
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "会话数据无效"})
			return
		}
		if body.PermissionMode != nil {
			mode, ok := requirePermissionMode(w, *body.PermissionMode)
			if !ok {
				return
			}
			body.PermissionMode = &mode
		}
		if body.Status != nil && !agentproxy.ValidSessionStatus(*body.Status) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "会话状态无效"})
			return
		}
		session, err := store.UpdateSession(req.Context(), owner, req.PathValue("id"), body, clock().UTC())
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session})
	})
	handle("DELETE /api/agent/proxy/sessions/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteSession(req.Context(), owner, req.PathValue("id"), clock().UTC()); err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	handle("GET /api/agent/proxy/sessions/{id}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		limit := 100
		if raw := strings.TrimSpace(req.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed <= 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit 参数无效"})
				return
			}
			limit = parsed
		}
		session, err := store.Session(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		messages, err := store.ListMessages(req.Context(), owner, session.ID, limit)
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"messages": messages, "permissionMode": session.PermissionMode})
	})
	handle("POST /api/agent/proxy/sessions/{id}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body agentproxy.AppendMessageInput
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 512<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "消息数据无效"})
			return
		}
		session, err := store.Session(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		// 接口层校验：plan 档位只读/只建议，不允许请求执行或写入工具结果。
		if session.PermissionMode == agentproxy.PermissionPlan {
			if body.Execute || strings.EqualFold(strings.TrimSpace(body.Role), agentproxy.MessageRoleTool) {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "当前会话是 plan 档位，只读不执行，请切换为 auto 后再执行"})
				return
			}
		}
		message, err := store.AppendMessage(req.Context(), owner, session.ID, body, clock().UTC())
		if err != nil {
			writeAgentProxyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"message": message})
	})
}

// requirePermissionMode 校验权限档位；bypass 在第一版只保留字段，不开放。
func requirePermissionMode(w http.ResponseWriter, raw string) (string, bool) {
	mode := agentproxy.NormalizePermissionMode(raw)
	switch {
	case mode == "":
		return agentproxy.PermissionPlan, true
	case mode == agentproxy.PermissionBypass:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bypass 档位尚未开放，第一版只支持 plan 和 auto"})
		return "", false
	case !agentproxy.ValidPermissionMode(mode):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "权限档位无效，只支持 plan 和 auto"})
		return "", false
	}
	return mode, true
}

func writeAgentProxyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentproxy.ErrWorkspaceNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "工作区不存在"})
	case errors.Is(err, agentproxy.ErrSessionNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "会话不存在"})
	case errors.Is(err, agentproxy.ErrWorkspaceExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "工作区已存在"})
	case errors.Is(err, agentproxy.ErrSessionExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "会话已存在"})
	case errors.Is(err, agentproxy.ErrPermissionModeUnavailable):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bypass 档位尚未开放，第一版只支持 plan 和 auto"})
	case errors.Is(err, agentproxy.ErrPermissionDenied):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "当前权限档位不允许该操作"})
	case errors.Is(err, agentproxy.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求数据无效"})
	default:
		log.Printf("agent proxy store error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Agent 代理数据保存失败"})
	}
}
