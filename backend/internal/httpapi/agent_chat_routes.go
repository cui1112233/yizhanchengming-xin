package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentchat"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// RegisterAgentChatRoutes wires the conversational Agent workspace:
// workspaces -> sessions -> messages, plus the permission-gated tool loop.
func RegisterAgentChatRoutes(
	root *http.ServeMux,
	middleware BridgeAuth,
	usersDir, systemDir string,
	catalogStore novelfetchworkshop.Store,
	userConfig *UserConfigStore,
	providerFor func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error),
	clock func() time.Time,
) {
	if root == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	if clock == nil {
		clock = time.Now
	}
	service := &agentChatService{
		usersDir:    usersDir,
		systemDir:   systemDir,
		providerFor: providerFor,
		clock:       clock,
		cancels:     map[int64]context.CancelFunc{},
	}
	store := agentchat.NewFileStore(usersDir, clock)
	service.store = store
	runner := agentchat.NewRunner(store, func(ctx context.Context, owner, modelID string) (agentchat.Provider, error) {
		if providerFor == nil {
			return nil, errors.New("text model catalog is unavailable")
		}
		provider, err := providerFor(ctx, owner, modelID)
		if err != nil {
			return nil, err
		}
		return agentChatProviderAdapter{inner: provider}, nil
	}, clock)
	runner.SkillsFor = func(owner string) *agentchat.SkillCatalog {
		return agentchat.NewSkillCatalog(agentchat.LoadSkillFiles(
			filepath.Join(systemDir, "agent-skills.json"),
			filepath.Join(usersDir, safeAgentUsername(owner), "agent-skills.json"),
		))
	}
	service.runner = runner

	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			owner, err := SafeUsername(identity.Username)
			if err != nil {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "invalid username"})
				return
			}
			fn(w, req, owner)
		})))
	}

	handle("GET /api/agent/permissions", func(w http.ResponseWriter, req *http.Request, owner string) {
		writeJSON(w, http.StatusOK, map[string]any{"options": agentchat.PermissionOptions()})
	})

	handle("GET /api/agent/models", func(w http.ResponseWriter, req *http.Request, owner string) {
		writeJSON(w, http.StatusOK, map[string]any{"models": agentChatTextModels(req.Context(), catalogStore, userConfig, owner)})
	})

	handle("GET /api/agent/workspaces", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspaces, err := service.store.ListWorkspaces(req.Context(), owner)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建工作区失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspaces": workspaces})
	})

	handle("POST /api/agent/workspaces", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Name     string `json:"name"`
			RootPath string `json:"rootPath"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10)).Decode(&body) != nil || strings.TrimSpace(body.Name) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先选择工作区"})
			return
		}
		rootPath := strings.TrimSpace(body.RootPath)
		workspace, err := service.store.CreateWorkspace(req.Context(), owner, agentchat.Workspace{
			Name: strings.TrimSpace(body.Name), RootPath: rootPath,
		})
		if err != nil {
			if errors.Is(err, agentchat.ErrWorkspaceRootOutsideUserDir) {
				// Never let a client point the workspace at an arbitrary directory
				// such as the data directory or a drive root.
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "工作区路径必须位于当前用户目录内"})
				return
			}
			writeJSON(w, http.StatusConflict, map[string]string{"error": "创建工作区失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspace": workspace})
	})

	handle("PATCH /api/agent/workspaces/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Name            *string   `json:"name"`
			Pinned          *bool     `json:"pinned"`
			Archived        *bool     `json:"archived"`
			EnabledSkillIDs *[]string `json:"enabledSkillIds"`
			Stage           *string   `json:"stage"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "更新工作区失败"})
			return
		}
		workspace, err := service.store.UpdateWorkspace(req.Context(), owner, req.PathValue("id"), agentchat.WorkspacePatch{
			Name: body.Name, Pinned: body.Pinned, Archived: body.Archived,
			EnabledSkillIDs: body.EnabledSkillIDs, Stage: body.Stage,
		})
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "工作区不存在"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"workspace": workspace})
	})

	handle("DELETE /api/agent/workspaces/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := service.store.DeleteWorkspace(req.Context(), owner, req.PathValue("id")); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "删除工作区失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	handle("GET /api/agent/workspaces/{id}/tree", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspace, err := service.store.Workspace(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "工作区不存在"})
			return
		}
		if !agentChatWorkspaceRootAllowed(service.store, owner, workspace.RootPath) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "工作区路径非法"})
			return
		}
		entries, err := agentChatTree(workspace.RootPath, req.URL.Query().Get("path"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "获取文件树失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
	})

	handle("GET /api/agent/workspaces/{id}/file", func(w http.ResponseWriter, req *http.Request, owner string) {
		workspace, err := service.store.Workspace(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "工作区不存在"})
			return
		}
		if !agentChatWorkspaceRootAllowed(service.store, owner, workspace.RootPath) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "工作区路径非法"})
			return
		}
		relative := req.URL.Query().Get("path")
		target, err := agentchat.ResolveWorkspacePathSecure(workspace.RootPath, relative)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "路径非法"})
			return
		}
		raw, err := os.ReadFile(target)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "文件不存在"})
			return
		}
		const limit = 200000
		content := string(raw)
		truncated := false
		if len([]rune(content)) > limit {
			content = string([]rune(content)[:limit])
			truncated = true
		}
		writeJSON(w, http.StatusOK, map[string]any{"path": relative, "content": content, "truncated": truncated})
	})

	handle("GET /api/agent/sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessions, err := service.store.ListSessions(req.Context(), owner, req.URL.Query().Get("workspaceId"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	})

	handle("POST /api/agent/sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		var body struct {
			Title       string `json:"title"`
			WorkspaceID string `json:"workspaceId"`
			Model       string `json:"model"`
			Permission  string `json:"permission"`
			ChatMode    string `json:"chatMode"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "创建会话失败"})
			return
		}
		workspaceID := strings.TrimSpace(body.WorkspaceID)
		if workspaceID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先选择工作区"})
			return
		}
		if _, err := service.store.Workspace(req.Context(), owner, workspaceID); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "工作区不存在"})
			return
		}
		model := strings.TrimSpace(body.Model)
		if model == "" {
			models := agentChatTextModels(req.Context(), catalogStore, userConfig, owner)
			if len(models) > 0 {
				model, _ = models[0]["id"].(string)
			}
		}
		session, err := service.store.CreateSession(req.Context(), owner, agentchat.SessionInfo{
			Title: body.Title, WorkspaceID: workspaceID, Model: model,
			PermissionMode: body.Permission, ChatMode: body.ChatMode,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "创建会话失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session})
	})

	handle("PATCH /api/agent/sessions/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessionID, ok := agentChatPathID(w, req)
		if !ok {
			return
		}
		var body struct {
			Title      *string `json:"title"`
			Model      *string `json:"model"`
			Permission *string `json:"permission"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "重命名失败"})
			return
		}
		session, err := service.store.UpdateSession(req.Context(), owner, sessionID, agentchat.SessionPatch{
			Title: body.Title, Model: body.Model, Permission: body.Permission,
		})
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session})
	})

	handle("DELETE /api/agent/sessions/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessionID, ok := agentChatPathID(w, req)
		if !ok {
			return
		}
		if err := service.store.DeleteSession(req.Context(), owner, sessionID); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		service.clearCancel(sessionID)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	handle("GET /api/agent/sessions/{id}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessionID, ok := agentChatPathID(w, req)
		if !ok {
			return
		}
		session, err := service.store.Session(req.Context(), owner, sessionID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		messages, err := service.store.Messages(req.Context(), owner, sessionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session, "messages": messages})
	})

	handle("DELETE /api/agent/sessions/{id}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessionID, ok := agentChatPathID(w, req)
		if !ok {
			return
		}
		if err := service.store.SaveMessages(req.Context(), owner, sessionID, []agentchat.Message{}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	handle("POST /api/agent/sessions/{id}/stop", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessionID, ok := agentChatPathID(w, req)
		if !ok {
			return
		}
		service.stop(sessionID)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	handle("POST /api/agent/sessions/{id}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessionID, ok := agentChatPathID(w, req)
		if !ok {
			return
		}
		var body struct {
			Content     string           `json:"content"`
			Model       string           `json:"model"`
			Permission  string           `json:"permission"`
			Attachments []map[string]any `json:"attachments"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 512<<10)).Decode(&body) != nil || strings.TrimSpace(body.Content) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "消息不能为空"})
			return
		}
		session, err := service.store.Session(req.Context(), owner, sessionID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "加载 Agent 会话失败:"})
			return
		}
		workspace, err := service.store.Workspace(req.Context(), owner, session.WorkspaceID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先选择工作区"})
			return
		}
		if !agentChatWorkspaceRootAllowed(service.store, owner, workspace.RootPath) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "工作区路径非法"})
			return
		}
		if len(body.Attachments) > 5 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "最多 5 个附件，请先移除部分附件"})
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "请求失败"})
			return
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		ctx, cancel := context.WithCancel(req.Context())
		service.setCancel(sessionID, cancel)
		defer func() {
			cancel()
			service.clearCancel(sessionID)
		}()
		emit := func(event agentchat.Event) { writeAgentChatEvent(w, flusher, event) }
		_, runErr := service.runner.Run(ctx, owner, agentchat.SendRequest{
			SessionID: sessionID, WorkspaceID: workspace.ID, Model: body.Model,
			Permission: body.Permission, Content: body.Content, Attachments: body.Attachments,
			EnabledSkillIDs: workspace.EnabledSkillIDs,
		}, emit)
		if runErr != nil {
			writeAgentChatEvent(w, flusher, agentchat.Event{Type: agentchat.EventError, Data: map[string]string{"message": runErr.Error()}})
		}
	})
}

type agentChatService struct {
	store       *agentchat.FileStore
	runner      *agentchat.Runner
	usersDir    string
	systemDir   string
	providerFor func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error)
	clock       func() time.Time

	mu      sync.Mutex
	cancels map[int64]context.CancelFunc
}

func (s *agentChatService) setCancel(sessionID int64, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancels == nil {
		s.cancels = map[int64]context.CancelFunc{}
	}
	s.cancels[sessionID] = cancel
}

func (s *agentChatService) clearCancel(sessionID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancels, sessionID)
}

func (s *agentChatService) stop(sessionID int64) {
	s.mu.Lock()
	cancel := s.cancels[sessionID]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

type agentChatProviderAdapter struct {
	inner batchfactoryv11.DirectorProvider
}

func (a agentChatProviderAdapter) Complete(ctx context.Context, system, user string, maxTokens int) (string, error) {
	if a.inner == nil {
		return "", errors.New("text model catalog is unavailable")
	}
	return a.inner.Complete(ctx, batchfactoryv11.TextCompletionRequest{
		SystemPrompt:        system,
		UserPrompt:          user,
		MaxTokens:           maxTokens,
		Temperature:         0.7,
		DisableJSONResponse: true,
	})
}

func agentChatPathID(w http.ResponseWriter, req *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(req.PathValue("id")), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "加载 Agent 会话失败:"})
		return 0, false
	}
	return id, true
}

func writeAgentChatEvent(w http.ResponseWriter, flusher http.Flusher, event agentchat.Event) {
	payload, err := json.Marshal(event.Data)
	if err != nil {
		payload = []byte(`{}`)
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload)
	flusher.Flush()
}

func agentChatWorkspaceRootAllowed(store *agentchat.FileStore, owner, root string) bool {
	inside, err := agentchat.WithinDir(store.UserAgentRoot(owner), root)
	return err == nil && inside
}

func agentChatTree(root, relative string) ([]map[string]any, error) {
	target, err := agentchat.ResolveWorkspacePathSecure(root, relative)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0, len(entries))
	base := strings.Trim(strings.ReplaceAll(relative, "\\", "/"), "/")
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && entry.Name() == "" {
			continue
		}
		path := name
		if base != "" {
			path = base + "/" + name
		}
		kind := "file"
		if entry.IsDir() {
			kind = "directory"
		}
		result = append(result, map[string]any{"name": name, "path": path, "type": kind})
	}
	return result, nil
}

// agentChatTextModels lists enabled text models for the composer picker.
// It also appends the "已启用 + 已填 Key" API 接口渠道，与 /api/models 的下拉一致。
func agentChatTextModels(ctx context.Context, store novelfetchworkshop.Store, userConfig *UserConfigStore, owner string) []map[string]any {
	models := []map[string]any{}
	if store != nil {
		if settings, err := store.GetConfig(ctx, owner); err == nil {
			for _, model := range modelCatalog(settings) {
				if !strings.EqualFold(modelString(model, "kind"), "text") || !modelEnabled(model) {
					continue
				}
				id := modelString(model, "id")
				if id == "" {
					id = modelString(model, "modelId")
				}
				if id == "" {
					continue
				}
				label := modelString(model, "name")
				if label == "" {
					label = modelString(model, "modelId")
				}
				models = append(models, map[string]any{"id": id, "name": label})
			}
		}
	}
	for _, candidate := range integrationChannelModels(ctx, "text", userConfig, owner, nil) {
		models = append(models, map[string]any{"id": candidate["id"], "name": candidate["displayName"]})
	}
	return models
}
