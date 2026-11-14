package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"qiantie/backend/internal/novelcreation"
)

const novelCreationJSONLimit = 4 << 20

type NovelCreationRunStarter func(context.Context, string, string)

var novelCreationModes = []map[string]string{
	{"id": "long_original", "name": "长篇原创", "description": "从设定、卷纲到章节持续创作"},
	{"id": "short_original", "name": "短篇原创", "description": "围绕单一核心冲突完成短篇"},
	{"id": "long_imitation", "name": "仿写长篇", "description": "提取参考作品风格后创作全新长篇"},
	{"id": "short_imitation", "name": "仿写短篇", "description": "按参考文本风格完成全新短篇"},
	{"id": "fanfiction", "name": "番外创作", "description": "基于既有人物与世界规则创作番外"},
	{"id": "continuation", "name": "续写创作", "description": "承接给定正文继续发展情节"},
}

func RegisterNovelCreationRoutes(mux *http.ServeMux, auth BridgeAuth, service *novelcreation.Service, starters ...NovelCreationRunStarter) {
	if mux == nil || service == nil {
		return
	}
	var starter NovelCreationRunStarter
	if len(starters) > 0 {
		starter = starters[0]
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "请先登录"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}
	decode := func(w http.ResponseWriter, req *http.Request, target any) bool {
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, novelCreationJSONLimit))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效，请检查后再试"})
			return false
		}
		return true
	}

	handle("GET /api/novel-creation/modes", func(w http.ResponseWriter, _ *http.Request, _ string) {
		modes := make([]map[string]any, 0, len(novelCreationModes))
		for _, mode := range novelCreationModes {
			modes = append(modes, map[string]any{
				"id": mode["id"], "name": mode["name"], "description": mode["description"],
				"inputs": novelcreation.ModeInputFields(mode["id"]),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"modes": modes})
	})
	handle("GET /api/novel-creation/projects", func(w http.ResponseWriter, req *http.Request, owner string) {
		projects, err := service.ListProjects(req.Context(), owner)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
	})
	handle("POST /api/novel-creation/projects", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input novelcreation.CreateProjectInput
		if !decode(w, req, &input) {
			return
		}
		project, err := service.CreateProject(req.Context(), owner, input)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"project": project})
	})
	handle("GET /api/novel-creation/projects/{projectID}/sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		sessions, err := service.ListSessions(req.Context(), owner, req.PathValue("projectID"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
	})
	handle("POST /api/novel-creation/projects/{projectID}/sessions", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Mode          string `json:"mode"`
			Title         string `json:"title"`
			ContextBudget int    `json:"contextBudget"`
		}
		if !decode(w, req, &input) {
			return
		}
		if !validNovelCreationMode(input.Mode) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择有效的创作模式"})
			return
		}
		session, err := service.CreateSession(req.Context(), owner, novelcreation.CreateSessionInput{
			ProjectID: req.PathValue("projectID"), Mode: input.Mode, Title: input.Title, ContextBudget: input.ContextBudget,
		})
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"session": session})
	})
	handle("POST /api/novel-creation/sessions/{sessionID}/messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ProjectID     string            `json:"projectId"`
			Content       string            `json:"content"`
			ModelID       string            `json:"modelId"`
			ThinkingLevel string            `json:"thinkingLevel"`
			Idempotency   string            `json:"idempotencyKey"`
			ModeInputs    map[string]string `json:"modeInputs"`
		}
		if !decode(w, req, &input) {
			return
		}
		run, err := service.SendMessage(req.Context(), owner, novelcreation.SendMessageInput{
			SessionID: req.PathValue("sessionID"), ProjectID: input.ProjectID, Content: input.Content,
			ModelID: input.ModelID, ThinkingLevel: input.ThinkingLevel, IdempotencyKey: input.Idempotency,
			ModeInputs: input.ModeInputs,
		})
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, run.ID)
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"run": run})
	})
	handle("GET /api/novel-creation/runs/{runID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		after, _ := strconv.ParseInt(req.URL.Query().Get("afterEventId"), 10, 64)
		snapshot, err := service.GetSnapshot(req.Context(), owner, req.PathValue("runID"), after)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
	})
	handle("POST /api/novel-creation/runs/{runID}/retry", func(w http.ResponseWriter, req *http.Request, owner string) {
		run, err := service.RetryRun(req.Context(), owner, req.PathValue("runID"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, run.ID)
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"run": run})
	})
	handle("POST /api/novel-creation/runs/{runID}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		run, err := service.CancelRun(req.Context(), owner, req.PathValue("runID"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run": run})
	})
	handle("POST /api/novel-creation/runs/{runID}/confirm", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Blueprint json.RawMessage `json:"blueprint"`
		}
		if !decode(w, req, &input) {
			return
		}
		run, err := service.ConfirmRun(req.Context(), owner, req.PathValue("runID"), input.Blueprint)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, run.ID)
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"run": run})
	})
	handle("GET /api/novel-creation/projects/{projectID}/settings", func(w http.ResponseWriter, req *http.Request, owner string) {
		settings, err := service.GetProjectSettings(req.Context(), owner, req.PathValue("projectID"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"settings": settings})
	})
	handle("PUT /api/novel-creation/projects/{projectID}/settings", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			StyleProfile string   `json:"styleProfile"`
			Rules        []string `json:"rules"`
		}
		if !decode(w, req, &input) {
			return
		}
		settings, err := service.UpdateProjectSettings(req.Context(), owner, req.PathValue("projectID"), novelcreation.ProjectSettings{
			StyleProfile: input.StyleProfile, Rules: input.Rules,
		})
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"settings": settings})
	})
	handle("GET /api/novel-creation/projects/{projectID}/export", func(w http.ResponseWriter, req *http.Request, owner string) {
		document, err := service.ExportProject(req.Context(), owner, req.PathValue("projectID"), req.URL.Query().Get("format"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		w.Header().Set("Content-Type", document.ContentType)
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(document.Filename))
		_, _ = w.Write([]byte(document.Content))
	})
}

func validNovelCreationMode(value string) bool {
	for _, mode := range novelCreationModes {
		if mode["id"] == strings.TrimSpace(value) {
			return true
		}
	}
	return false
}

func writeNovelCreationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "小说创作暂时无法处理，请稍后重试"
	var domainErr *novelcreation.DomainError
	if errors.As(err, &domainErr) {
		message = domainErr.Message
		switch domainErr.Code {
		case novelcreation.ErrInputRequired:
			status = http.StatusBadRequest
		case novelcreation.ErrProjectNotFound, novelcreation.ErrSessionNotFound:
			status = http.StatusNotFound
		case novelcreation.ErrRunInProgress, novelcreation.ErrProjectLocked, novelcreation.ErrInvalidTransition:
			status = http.StatusConflict
		case novelcreation.ErrModelUnavailable:
			status = http.StatusUnprocessableEntity
		case novelcreation.ErrGenerationTimeout:
			status = http.StatusGatewayTimeout
		}
	}
	writeJSON(w, status, map[string]string{"error": message})
}
