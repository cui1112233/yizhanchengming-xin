package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type createAgentGenerationRequest struct {
	NodeID            string         `json:"nodeId"`
	NodeVersion       int64          `json:"nodeVersion"`
	Kind              string         `json:"kind"`
	Provider          string         `json:"provider"`
	ModelID           string         `json:"modelId"`
	Mode              string         `json:"mode"`
	Prompt            string         `json:"prompt"`
	Parameters        map[string]any `json:"parameters"`
	Advanced          map[string]any `json:"advanced"`
	ReferenceAssetIDs []string       `json:"referenceAssetIds"`
	IdempotencyKey    string         `json:"idempotencyKey"`
}

func RegisterAgentGenerationTaskRoutes(root *http.ServeMux, middleware BridgeAuth, store agentworkspace.GenerationTaskStore, now func() time.Time) {
	if root == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, handler func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			handler(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("POST /api/agent/tasks/{id}/generations", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input createAgentGenerationRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 128<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "生成任务参数无效"})
			return
		}
		input.NodeID = strings.TrimSpace(input.NodeID)
		input.Kind = strings.TrimSpace(input.Kind)
		input.Provider = strings.TrimSpace(input.Provider)
		input.ModelID = strings.TrimSpace(input.ModelID)
		input.Mode = strings.TrimSpace(input.Mode)
		input.Prompt = strings.TrimSpace(input.Prompt)
		input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
		if input.NodeID == "" || input.NodeVersion < 1 || !agentGenerationKind(input.Kind) || input.Provider == "" || input.ModelID == "" || input.IdempotencyKey == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择正确的节点、生成通道和模型"})
			return
		}
		settings := cloneAgentSettings(input.Parameters)
		settings["parameters"] = cloneAgentSettings(input.Parameters)
		settings["advanced"] = cloneAgentSettings(input.Advanced)
		if input.Mode != "" {
			settings["mode"] = input.Mode
		}
		if input.Prompt != "" {
			settings["prompt"] = input.Prompt
		}
		validationSettings := cloneAgentSettings(input.Parameters)
		if input.Mode != "" {
			validationSettings["mode"] = input.Mode
		}
		if err := ValidateModelSettings(map[string]any{"kind": input.Kind}, input.Kind, validationSettings); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		projects, ok := store.(agentworkspace.Store)
		if !ok {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "项目画布存储未配置"})
			return
		}
		canvas, err := projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		if !agentGenerationNodeMatches(canvas, input.NodeID, input.Kind, input.NodeVersion) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "节点已变化，请重新打开后再生成"})
			return
		}
		task, err := store.CreateGenerationTask(req.Context(), owner, agentworkspace.CreateGenerationTaskInput{
			ID: newAgentID("generation"), ProjectID: projectID, NodeID: input.NodeID, NodeVersion: input.NodeVersion,
			Kind: input.Kind, Provider: input.Provider, ModelID: input.ModelID, Mode: input.Mode, Settings: settings,
			ReferenceAssetIDs: uniqueAgentGenerationIDs(input.ReferenceAssetIDs), IdempotencyKey: input.IdempotencyKey, Now: now().UTC(),
		})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"task": task})
	})

	handle("GET /api/agent/tasks/{id}/generations", func(w http.ResponseWriter, req *http.Request, owner string) {
		tasks, err := store.ListGenerationTasks(req.Context(), owner, strings.TrimSpace(req.PathValue("id")))
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
	})

	handle("GET /api/agent/tasks/{id}/generation-summary", func(w http.ResponseWriter, req *http.Request, owner string) {
		tasks, err := store.ListGenerationTasks(req.Context(), owner, strings.TrimSpace(req.PathValue("id")))
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"summary": SummarizeAgentGenerations(tasks)})
	})

	handle("GET /api/agent/tasks/{id}/generations/{generationID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		task, err := store.GenerationTask(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("generationID")))
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task})
	})

	handle("POST /api/agent/tasks/{id}/generations/{generationID}/confirm", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ResultID string `json:"resultId"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择要使用的生成结果"})
			return
		}
		if rejectGenericVideoAction(w, req, owner, store) {
			return
		}
		canvas, task, err := store.ConfirmGenerationTask(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("generationID")), strings.TrimSpace(input.ResultID), now().UTC())
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task, "canvas": agentCanvasForResponse(task.ProjectID, canvas)})
	})

	handle("POST /api/agent/tasks/{id}/generations/{generationID}/activate", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ResultID            string `json:"resultId"`
			ExpectedNodeVersion int64  `json:"expectedNodeVersion"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input); err != nil || strings.TrimSpace(input.ResultID) == "" || input.ExpectedNodeVersion < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择历史结果并提供当前节点版本"})
			return
		}
		if rejectGenericVideoAction(w, req, owner, store) {
			return
		}
		canvas, task, err := store.ActivateGenerationResult(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("generationID")), strings.TrimSpace(input.ResultID), input.ExpectedNodeVersion, now().UTC())
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task, "canvas": agentCanvasForResponse(task.ProjectID, canvas)})
	})

	handle("POST /api/agent/tasks/{id}/generations/{generationID}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		if rejectGenericVideoAction(w, req, owner, store) {
			return
		}
		task, err := store.CancelGenerationTask(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("generationID")), now().UTC())
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": task})
	})

	handle("POST /api/agent/tasks/{id}/generations/{generationID}/retry", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			IdempotencyKey string `json:"idempotencyKey"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input); err != nil || strings.TrimSpace(input.IdempotencyKey) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "重试请求参数无效"})
			return
		}
		if rejectGenericVideoAction(w, req, owner, store) {
			return
		}
		task, err := store.RetryGenerationTask(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("generationID")), agentworkspace.RetryGenerationTaskInput{
			ID: newAgentID("generation"), IdempotencyKey: strings.TrimSpace(input.IdempotencyKey), Now: now().UTC(),
		})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"task": task})
	})
}

func rejectGenericVideoAction(w http.ResponseWriter, req *http.Request, owner string, store agentworkspace.GenerationTaskStore) bool {
	task, err := store.GenerationTask(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("generationID")))
	if err != nil {
		writeAgentGenerationTaskError(w, err)
		return true
	}
	if task.Kind != "video" || strings.HasPrefix(task.Mode, "video_") {
		return false
	}
	writeJSON(w, http.StatusConflict, map[string]string{"error": "视频任务请使用视频任务接口确认、取消或重试"})
	return true
}

func agentGenerationKind(kind string) bool {
	return kind == "text" || kind == "image" || kind == "video" || kind == "audio"
}

func uniqueAgentGenerationIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func writeAgentGenerationTaskError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrProjectNotFound), errors.Is(err, agentworkspace.ErrGenerationNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "生成任务或项目不存在"})
	case errors.Is(err, agentworkspace.ErrGenerationNodeChanged), errors.Is(err, agentworkspace.ErrGenerationState), errors.Is(err, agentworkspace.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "生成任务状态或节点版本已变化，请刷新后重试"})
	case errors.Is(err, agentworkspace.ErrGenerationInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "生成任务参数无效"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成任务处理失败"})
	}
}
