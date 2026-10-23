package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
)

type AgentModelCapabilityResolver func(context.Context, string, string, string) (map[string]any, error)
type AgentShotGenerationDispatcher func(context.Context, string, string, agentworkspace.Node, agentworkspace.GenerationTask) error

type agentDispatchResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (r *agentDispatchResponse) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *agentDispatchResponse) WriteHeader(status int) { r.status = status }

func (r *agentDispatchResponse) Write(contents []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(contents)
}

type AgentScriptShotServices struct {
	Projects           agentworkspace.Store
	TextModels         AgentTextModelResolver
	TextComplete       AgentTextCompletionExecutor
	Models             AgentModelCapabilityResolver
	Dispatch           AgentShotGenerationDispatcher
	Now                func() time.Time
	ScriptSplitTimeout time.Duration
}

const defaultAgentScriptSplitTimeout = 90 * time.Second
const agentScriptSplitWorkerInterval = 2 * time.Second

type agentScriptSplitCoordinator struct {
	services AgentScriptShotServices
	mu       sync.Mutex
	inFlight map[string]struct{}
}

func newAgentScriptSplitCoordinator(services AgentScriptShotServices) *agentScriptSplitCoordinator {
	if services.Now == nil {
		services.Now = time.Now
	}
	if services.ScriptSplitTimeout <= 0 {
		services.ScriptSplitTimeout = defaultAgentScriptSplitTimeout
	}
	return &agentScriptSplitCoordinator{services: services, inFlight: map[string]struct{}{}}
}

// Run advances persisted script-split jobs independently of an open browser.
func (s *agentScriptSplitCoordinator) Run(ctx context.Context) {
	if s == nil || s.services.Projects == nil {
		return
	}
	_ = s.pollPendingOnce(ctx)
	ticker := time.NewTicker(agentScriptSplitWorkerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.pollPendingOnce(ctx)
		}
	}
}

func (s *agentScriptSplitCoordinator) pollPendingOnce(ctx context.Context) error {
	tasks, err := s.services.Projects.ListRunnableGenerationTasks(ctx, "script_split", 100)
	if err != nil {
		return err
	}
	var firstErr error
	for _, task := range tasks {
		if !s.claim(task.ID) {
			continue
		}
		err := s.execute(ctx, task)
		s.release(task.ID)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *agentScriptSplitCoordinator) claim(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.inFlight[id]; exists {
		return false
	}
	s.inFlight[id] = struct{}{}
	return true
}

func (s *agentScriptSplitCoordinator) release(id string) {
	s.mu.Lock()
	delete(s.inFlight, id)
	s.mu.Unlock()
}

func (s *agentScriptSplitCoordinator) execute(ctx context.Context, task agentworkspace.GenerationTask) error {
	services := s.services
	if task.State == agentworkspace.GenerationQueued {
		var err error
		task, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: task.UpdatedAt, Now: services.Now().UTC()})
		if err != nil {
			return err
		}
	}
	if task.State != agentworkspace.GenerationRunning {
		return nil
	}
	_, node, err := scriptNodeForRequest(ctx, services.Projects, task.OwnerUsername, task.ProjectID, task.NodeID, task.NodeVersion)
	if err != nil {
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, "剧本节点已变化，请重新拆分")
		return err
	}
	sourceText := strings.TrimSpace(stringValue(task.Settings["sourceText"]))
	if sourceText == "" {
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, "剧本正文为空")
		return errScriptSplitUpstream
	}
	if services.TextComplete == nil && services.TextModels == nil {
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, "文字模型执行器未配置")
		return errScriptSplitUpstream
	}
	complete := services.TextComplete
	if complete == nil {
		complete = newAgentTextCompletionExecutor(services.TextModels)
	}
	splitCtx, cancel := context.WithTimeout(ctx, services.ScriptSplitTimeout)
	defer cancel()
	contract, err := buildAgentDirectorContract(node.ID, node.Title, sourceText)
	if err != nil {
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, "剧本拆分规则初始化失败")
		return errScriptSplitUpstream
	}
	directorResult, err := completeAgentDirectorContract(complete, splitCtx, task.OwnerUsername, task.ModelID, contract)
	if err != nil {
		message := "剧本拆分失败"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(splitCtx.Err(), context.DeadlineExceeded) {
			message = "文字模型请求超时，请检查模型通道后重试"
		}
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, message)
		return errScriptSplitUpstream
	}
	shots, err := adaptDirectorResultToScriptShots(node.ID, directorResult)
	if err != nil {
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, "剧本拆分结果格式无效")
		return errScriptSplitUpstream
	}
	if err := services.Projects.ReplaceShots(ctx, task.OwnerUsername, task.ProjectID, node.ID, shots); err != nil {
		failScriptSplit(ctx, services, task.OwnerUsername, task.ProjectID, task, "镜头保存失败")
		return err
	}
	_, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationCompleted, ExpectedUpdatedAt: task.UpdatedAt, Now: services.Now().UTC()})
	if err != nil {
		return err
	}
	return nil
}

type scriptSplitRequest struct {
	NodeID      string `json:"nodeId"`
	NodeVersion int64  `json:"nodeVersion"`
	ModelID     string `json:"modelId"`
}

type batchShotNodesRequest struct {
	ScriptNodeID string   `json:"scriptNodeId"`
	ShotIDs      []string `json:"shotIds"`
	Kind         string   `json:"kind"`
	Provider     string   `json:"provider"`
	ModelID      string   `json:"modelId"`
}

type skippedShot struct {
	ShotID string `json:"shotId"`
	Reason string `json:"reason"`
}

func RegisterAgentScriptShotRoutes(root *http.ServeMux, middleware BridgeAuth, services AgentScriptShotServices) {
	if root == nil || services.Projects == nil {
		return
	}
	if services.Now == nil {
		services.Now = time.Now
	}
	if services.ScriptSplitTimeout <= 0 {
		services.ScriptSplitTimeout = defaultAgentScriptSplitTimeout
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

	handle("POST /api/agent/tasks/{id}/script-splits", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input scriptSplitRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "剧本拆分参数无效"})
			return
		}
		input.NodeID, input.ModelID = strings.TrimSpace(input.NodeID), strings.TrimSpace(input.ModelID)
		projectID := strings.TrimSpace(req.PathValue("id"))
		canvas, node, err := scriptNodeForRequest(req.Context(), services.Projects, owner, projectID, input.NodeID, input.NodeVersion)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		sourceText := strings.TrimSpace(stringValue(node.Config["sourceText"]))
		if sourceText == "" || input.ModelID == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "请先填写剧本正文并选择文字模型"})
			return
		}
		if _, err := resolveScriptShotModel(req.Context(), services, owner, input.ModelID, "text", ""); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "所选文字模型未配置或类型不匹配"})
			return
		}
		task, err := enqueueAgentScriptSplit(req.Context(), services, owner, projectID, node, input.ModelID, sourceText)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"canvas": agentCanvasForResponse(projectID, canvas), "generation": task})
	})

	handle("GET /api/agent/tasks/{id}/script-nodes/{nodeId}/shots", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, nodeID := strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("nodeId"))
		canvas, err := services.Projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		node, ok := agentCanvasNodeByID(canvas, nodeID)
		if !ok || node.Type != "script" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "剧本节点不存在"})
			return
		}
		shots, err := services.Projects.ListShots(req.Context(), owner, projectID, nodeID)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"shots": shots})
	})

	handle("PATCH /api/agent/tasks/{id}/shots/{shotId}", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, shotID := strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("shotId"))
		var patch map[string]any
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10))
		decoder.UseNumber()
		if decoder.Decode(&patch) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "镜头修改参数无效"})
			return
		}
		scriptNodeID := strings.TrimSpace(stringValue(patch["scriptNodeId"]))
		canvas, err := services.Projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		if node, ok := agentCanvasNodeByID(canvas, scriptNodeID); !ok || node.Type != "script" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "剧本节点不存在"})
			return
		}
		shots, err := services.Projects.ListShots(req.Context(), owner, projectID, scriptNodeID)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		shot, found := findScriptShot(shots, shotID)
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "镜头不存在"})
			return
		}
		revision, ok := requiredShotRevision(patch)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供有效的镜头 revision"})
			return
		}
		shot.Revision = revision
		applyShotPatch(&shot, patch)
		shot.ProjectID, shot.ScriptNodeID, shot.ID = projectID, scriptNodeID, shotID
		updated, err := services.Projects.UpdateShot(req.Context(), owner, shot)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"shot": updated})
	})

	handle("POST /api/agent/tasks/{id}/shots/batch-nodes", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input batchShotNodesRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "批量生成参数无效"})
			return
		}
		input.ScriptNodeID, input.Kind, input.Provider, input.ModelID = strings.TrimSpace(input.ScriptNodeID), strings.ToLower(strings.TrimSpace(input.Kind)), strings.ToLower(strings.TrimSpace(input.Provider)), strings.TrimSpace(input.ModelID)
		shotIDs, ok := nonEmptyUniqueShotIDs(input.ShotIDs)
		if !ok || input.ScriptNodeID == "" || input.ModelID == "" || (input.Kind != "image" && input.Kind != "video" && input.Kind != "audio") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择不重复的镜头、生成类型和模型"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		canvas, err := services.Projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		scriptNode, exists := agentCanvasNodeByID(canvas, input.ScriptNodeID)
		if !exists || scriptNode.Type != "script" || scriptNode.Version < 1 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "剧本节点不存在"})
			return
		}
		model, err := resolveScriptShotModel(req.Context(), services, owner, input.ModelID, input.Kind, input.Provider)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "所选模型未配置或类型不匹配"})
			return
		}
		shots, err := services.Projects.ListShots(req.Context(), owner, projectID, input.ScriptNodeID)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		response, err := createShotBatchNodes(req.Context(), services, owner, projectID, canvas, scriptNode, shots, shotIDs, input, model)
		if err != nil {
			writeScriptShotError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	})
}

func enqueueAgentScriptSplit(ctx context.Context, services AgentScriptShotServices, owner, projectID string, node agentworkspace.Node, modelID, sourceText string) (agentworkspace.GenerationTask, error) {
	now := services.Now().UTC()
	return services.Projects.CreateGenerationTask(ctx, owner, agentworkspace.CreateGenerationTaskInput{
		ID: newAgentID("generation"), ProjectID: projectID, NodeID: node.ID, NodeVersion: node.Version,
		Kind: "text", Provider: "configured_model", ModelID: modelID, Mode: "script_split",
		Settings:       map[string]any{"prompt": sourceText, "sourceText": sourceText, "parameters": map[string]any{}, "advanced": map[string]any{}},
		IdempotencyKey: fmt.Sprintf("script-split:%s:v%d:%d", node.ID, node.Version, now.UnixNano()), Now: now,
	})
}

func executeAgentScriptSplit(ctx context.Context, services AgentScriptShotServices, owner, projectID string, canvas agentworkspace.Canvas, node agentworkspace.Node, modelID, sourceText string) (map[string]any, error) {
	now := services.Now().UTC()
	idempotencyKey := fmt.Sprintf("script-split:%s:v%d:%d", node.ID, node.Version, now.UnixNano())
	task, err := beginAgentGenerationTask(ctx, services.Projects, owner, projectID, "text", node.ID, node.Version, "configured_model", modelID, "script_split", sourceText, nil, nil, nil, idempotencyKey, now)
	if err != nil {
		return nil, err
	}
	if services.TextComplete == nil && services.TextModels == nil {
		failScriptSplit(ctx, services, owner, projectID, task, "文字模型执行器未配置")
		return nil, errScriptSplitUpstream
	}
	complete := services.TextComplete
	if complete == nil {
		complete = newAgentTextCompletionExecutor(services.TextModels)
	}
	splitCtx, cancel := context.WithTimeout(ctx, services.ScriptSplitTimeout)
	defer cancel()
	contract, err := buildAgentDirectorContract(node.ID, node.Title, sourceText)
	if err != nil {
		failScriptSplit(ctx, services, owner, projectID, task, "剧本拆分规则初始化失败")
		return nil, errScriptSplitUpstream
	}
	directorResult, err := completeAgentDirectorContract(complete, splitCtx, owner, modelID, contract)
	if err != nil {
		message := "剧本拆分失败"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(splitCtx.Err(), context.DeadlineExceeded) {
			message = "文字模型请求超时，请检查模型通道后重试"
		}
		failScriptSplit(ctx, services, owner, projectID, task, message)
		return nil, errScriptSplitUpstream
	}
	shots, err := adaptDirectorResultToScriptShots(node.ID, directorResult)
	if err != nil {
		failScriptSplit(ctx, services, owner, projectID, task, "剧本拆分结果格式无效")
		return nil, errScriptSplitUpstream
	}
	if err := services.Projects.ReplaceShots(ctx, owner, projectID, node.ID, shots); err != nil {
		failScriptSplit(ctx, services, owner, projectID, task, "镜头保存失败")
		return nil, err
	}
	task, err = services.Projects.TransitionGenerationTask(ctx, owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationCompleted, Now: services.Now().UTC()})
	if err != nil {
		return nil, err
	}
	return map[string]any{"canvas": agentCanvasForResponse(projectID, canvas), "generation": task, "split": agentworkspace.ScriptSplit{ProjectID: projectID, ScriptNodeID: node.ID, Shots: shots}}, nil
}

var errScriptSplitUpstream = errors.New("script split upstream failed")

func failScriptSplit(ctx context.Context, services AgentScriptShotServices, owner, projectID string, task agentworkspace.GenerationTask, message string) {
	failAgentGenerationTask(ctx, services.Projects, owner, projectID, task, message, services.Now().UTC())
}

func decodeScriptShots(scriptNodeID, value string) ([]agentworkspace.ScriptShot, error) {
	scriptNodeID = strings.TrimSpace(scriptNodeID)
	if scriptNodeID == "" {
		return nil, agentworkspace.ErrScriptShotInvalid
	}
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "```json")
	value = strings.TrimPrefix(value, "```")
	value = strings.TrimSuffix(strings.TrimSpace(value), "```")
	var payload struct {
		Shots []agentworkspace.ScriptShot `json:"shots"`
	}
	if json.Unmarshal([]byte(value), &payload) != nil || len(payload.Shots) == 0 {
		return nil, agentworkspace.ErrScriptShotInvalid
	}
	seen := make(map[string]struct{}, len(payload.Shots))
	for index := range payload.Shots {
		shot := &payload.Shots[index]
		if shot.Sequence == 0 {
			shot.Sequence = index + 1
		}
		shot.ID = fmt.Sprintf("%s-shot-%d", scriptNodeID, index+1)
		if _, exists := seen[shot.ID]; exists || shot.Sequence != index+1 {
			return nil, agentworkspace.ErrScriptShotInvalid
		}
		seen[shot.ID] = struct{}{}
	}
	return payload.Shots, nil
}

func scriptNodeForRequest(ctx context.Context, store agentworkspace.Store, owner, projectID, nodeID string, version int64) (agentworkspace.Canvas, agentworkspace.Node, error) {
	canvas, err := store.Canvas(ctx, owner, projectID)
	if err != nil {
		return agentworkspace.Canvas{}, agentworkspace.Node{}, err
	}
	node, ok := agentCanvasNodeByID(canvas, nodeID)
	if !ok {
		return agentworkspace.Canvas{}, agentworkspace.Node{}, agentworkspace.ErrScriptShotNotFound
	}
	if node.Type != "script" || version < 1 || node.Version != version {
		return agentworkspace.Canvas{}, agentworkspace.Node{}, agentworkspace.ErrRevisionConflict
	}
	return canvas, node, nil
}

func createShotBatchNodes(ctx context.Context, services AgentScriptShotServices, owner, projectID string, canvas agentworkspace.Canvas, scriptNode agentworkspace.Node, shots []agentworkspace.ScriptShot, shotIDs []string, input batchShotNodesRequest, model map[string]any) (map[string]any, error) {
	shotByID := make(map[string]agentworkspace.ScriptShot, len(shots))
	for _, shot := range shots {
		shotByID[shot.ID] = shot
	}
	nodes := make([]agentworkspace.Node, 0, len(shotIDs))
	generationLinks := make([]agentworkspace.ShotGenerationLink, 0, len(shotIDs))
	skipped := make([]skippedShot, 0)
	for _, shotID := range shotIDs {
		shot, ok := shotByID[shotID]
		if !ok {
			skipped = append(skipped, skippedShot{ShotID: shotID, Reason: "镜头不存在或不属于当前剧本"})
			continue
		}
		prompt := shotPrompt(shot, input.Kind)
		if prompt == "" {
			skipped = append(skipped, skippedShot{ShotID: shotID, Reason: "镜头缺少当前类型的生成内容"})
			continue
		}
		links, err := services.Projects.ListGenerationLinks(ctx, owner, projectID, shotID)
		if err != nil {
			return nil, err
		}
		attempt := nextShotOperationAttempt(canvas, links, input.Kind)
		operation := shotOperation(input.Kind)
		node := agentworkspace.Node{ID: newAgentID(input.Kind), Type: input.Kind, Title: shotNodeTitle(shot, input.Kind), Version: 1, Status: "queued", Summary: prompt, SourceNodeIDs: []string{scriptNode.ID}, Position: agentworkspace.Position{X: scriptNode.Position.X + 360, Y: scriptNode.Position.Y + float64(len(nodes)*180)}, Config: map[string]any{"modelId": input.ModelID, "prompt": prompt, "source": map[string]any{"scriptNodeId": scriptNode.ID, "shotId": shot.ID, "operation": shotOperation(input.Kind), "attempt": attempt}}}
		source, _ := node.Config["source"].(map[string]any)
		provider := strings.TrimSpace(modelString(model, "provider"))
		if provider == "" {
			provider = "configured_model"
		}
		node.Config["provider"] = provider
		idempotencyKey := fmt.Sprintf("shot:%s:%s:%d", shot.ID, operation, attempt)
		taskID := newAgentID("generation")
		taskIdempotencyKey := idempotencyKey
		if input.Kind == "video" {
			taskID = newAgentID("video-job")
			taskIdempotencyKey = "video:" + idempotencyKey
		}
		taskInput := agentworkspace.CreateGenerationTaskInput{ID: taskID, ProjectID: projectID, NodeID: node.ID, NodeVersion: node.Version, Kind: input.Kind, Provider: provider, ModelID: input.ModelID, Mode: operation, Settings: map[string]any{"prompt": prompt, "parameters": map[string]any{}, "advanced": map[string]any{}, "source": source, "submissionIdempotencyKey": idempotencyKey}, IdempotencyKey: taskIdempotencyKey, Now: services.Now().UTC()}
		link := agentworkspace.ShotGenerationLink{ProjectID: projectID, ScriptNodeID: scriptNode.ID, ShotID: shot.ID, NodeID: node.ID, GenerationID: taskID, Attempt: attempt, Active: true}
		createdCanvas, task, createdLink, err := services.Projects.CreateShotGeneration(ctx, owner, agentworkspace.CreateShotGenerationInput{ExpectedCanvasRevision: canvas.Revision, Node: node, Edge: agentworkspace.Edge{ID: newAgentID("edge"), Source: scriptNode.ID, Target: node.ID, Relation: "shot-generation"}, Task: taskInput, Link: link, Now: services.Now().UTC()})
		if err != nil {
			return nil, err
		}
		canvas = createdCanvas
		nodes = append(nodes, node)
		generationLinks = append(generationLinks, createdLink)
		if services.Dispatch == nil {
			return nil, errors.New("shot generation dispatcher unavailable")
		}
		if err := services.Dispatch(ctx, owner, projectID, node, task); err != nil {
			_, _ = services.Projects.TransitionGenerationTask(ctx, owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: err.Error(), Now: services.Now().UTC()})
		}
	}
	return map[string]any{"canvas": agentCanvasForResponse(projectID, canvas), "nodes": nodes, "generationLinks": generationLinks, "skipped": skipped}, nil
}

func resolveScriptShotModel(ctx context.Context, services AgentScriptShotServices, owner, modelID, kind, provider string) (map[string]any, error) {
	if provider == "doubao_local_executor" {
		switch kind {
		case "image":
			if _, ok := agentLocalDoubaoImageModel(AgentImageGenerationInput{Provider: provider, ModelID: modelID}); ok {
				return map[string]any{"id": modelID, "modelId": modelID, "kind": kind, "provider": provider, "enabled": true}, nil
			}
		case "video":
			if supportedAgentLocalVideoModel(modelID) {
				return map[string]any{"id": modelID, "modelId": modelID, "kind": kind, "provider": provider, "enabled": true}, nil
			}
		}
		return nil, errors.New("local Doubao model unavailable")
	}
	if services.Models == nil {
		return nil, errors.New("model catalog unavailable")
	}
	model, err := services.Models(ctx, owner, modelID, kind)
	if err != nil || model == nil || !strings.EqualFold(modelString(model, "kind"), kind) || !modelEnabled(model) {
		return nil, errors.New("model unavailable")
	}
	if err := ValidateModelSettings(model, kind, map[string]any{}); err != nil {
		return nil, err
	}
	return model, nil
}

func agentConfiguredModelResolver(store novelfetchworkshop.Store) AgentModelCapabilityResolver {
	return func(ctx context.Context, owner, modelID, kind string) (map[string]any, error) {
		if kind == "audio" && modelID == builtInTTSModelID {
			return map[string]any{"id": modelID, "modelId": modelID, "kind": "audio", "provider": "personal_api", "enabled": true}, nil
		}
		if store == nil {
			return nil, errors.New("model catalog unavailable")
		}
		settings, err := store.GetConfig(ctx, owner)
		if err != nil {
			return nil, err
		}
		for _, model := range modelCatalog(settings) {
			if !strings.EqualFold(modelString(model, "kind"), kind) || !modelEnabled(model) {
				continue
			}
			if modelString(model, "id") == modelID || modelString(model, "modelId") == modelID {
				return model, nil
			}
		}
		return nil, errors.New("model not configured")
	}
}

func newAgentShotGenerationDispatcher(root *http.ServeMux, auth BridgeAuth, now func() time.Time, videos *agentVideoCoordinator) AgentShotGenerationDispatcher {
	if root == nil {
		return nil
	}
	if now == nil {
		now = time.Now
	}
	return func(ctx context.Context, owner, projectID string, node agentworkspace.Node, task agentworkspace.GenerationTask) error {
		idempotencyKey := strings.TrimSpace(stringValue(task.Settings["submissionIdempotencyKey"]))
		prompt := strings.TrimSpace(stringValue(task.Settings["prompt"]))
		if idempotencyKey == "" || prompt == "" {
			return errors.New("生成任务缺少提交参数")
		}
		switch task.Kind {
		case "image":
			return dispatchAgentGenerationRoute(ctx, root, auth, now(), owner, projectID, "/image-generations", agentImageGenerationRequest{NodeID: node.ID, NodeVersion: node.Version, Provider: task.Provider, ModelID: task.ModelID, Mode: task.Mode, Prompt: prompt, IdempotencyKey: idempotencyKey})
		case "audio":
			return dispatchAgentGenerationRoute(ctx, root, auth, now(), owner, projectID, "/audio-generations", agentAudioRequest{NodeID: node.ID, NodeVersion: node.Version, Operation: task.Mode, Prompt: prompt, Provider: task.Provider, ModelID: task.ModelID, IdempotencyKey: idempotencyKey})
		case "video":
			if videos == nil {
				return errors.New("视频生成协调器未配置")
			}
			_, err := videos.submit(ctx, owner, projectID, agentVideoJobRequest{JobID: task.ID, NodeID: node.ID, NodeVersion: node.Version, Provider: task.Provider, ModelID: task.ModelID, ModelKey: task.ModelID, Mode: task.Mode, Prompt: prompt, IdempotencyKey: idempotencyKey})
			return err
		default:
			return errors.New("不支持的镜头生成类型")
		}
	}
}

func dispatchAgentGenerationRoute(ctx context.Context, root *http.ServeMux, auth BridgeAuth, issuedAt time.Time, owner, projectID, suffix string, body any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/agent/tasks/"+projectID+suffix, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	SignBridgeRequest(req, owner, false, issuedAt.UTC(), auth.Secret)
	response := &agentDispatchResponse{}
	root.ServeHTTP(response, req)
	if response.status < 200 || response.status >= 300 {
		message := strings.TrimSpace(response.body.String())
		if message == "" {
			message = http.StatusText(response.status)
		}
		return fmt.Errorf("生成入口返回 %d: %s", response.status, message)
	}
	return nil
}

func nonEmptyUniqueShotIDs(values []string) ([]string, bool) {
	if len(values) == 0 {
		return nil, false
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false
		}
		if _, exists := seen[value]; exists {
			return nil, false
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out, true
}

func findScriptShot(shots []agentworkspace.ScriptShot, id string) (agentworkspace.ScriptShot, bool) {
	for _, shot := range shots {
		if shot.ID == id {
			return shot, true
		}
	}
	return agentworkspace.ScriptShot{}, false
}

func applyShotPatch(shot *agentworkspace.ScriptShot, patch map[string]any) {
	if value, ok := patch["sequence"]; ok {
		shot.Sequence = int(agentExecutionInt64(value))
		if number, ok := value.(json.Number); ok {
			parsed, _ := number.Int64()
			shot.Sequence = int(parsed)
		}
	}
	fields := map[string]*string{"scene": &shot.Scene, "characters": &shot.Characters, "dialogue": &shot.Dialogue, "narration": &shot.Narration, "imagePrompt": &shot.ImagePrompt, "videoPrompt": &shot.VideoPrompt, "ttsText": &shot.TTSText}
	for key, target := range fields {
		if value, exists := patch[key]; exists {
			*target = strings.TrimSpace(stringValue(value))
		}
	}
}

func requiredShotRevision(patch map[string]any) (int64, bool) {
	value, exists := patch["revision"]
	if !exists {
		return 0, false
	}
	if number, ok := value.(json.Number); ok {
		revision, err := number.Int64()
		return revision, err == nil && revision > 0
	}
	revision := agentExecutionInt64(value)
	return revision, revision > 0
}

func shotPrompt(shot agentworkspace.ScriptShot, kind string) string {
	switch kind {
	case "image":
		return strings.TrimSpace(shot.ImagePrompt)
	case "video":
		return strings.TrimSpace(shot.VideoPrompt)
	case "audio":
		return strings.TrimSpace(shot.TTSText)
	default:
		return ""
	}
}

func shotOperation(kind string) string {
	switch kind {
	case "image":
		return "text_to_image"
	case "video":
		return "text_to_video"
	case "audio":
		return "speech"
	default:
		return ""
	}
}

func nextShotOperationAttempt(canvas agentworkspace.Canvas, links []agentworkspace.ShotGenerationLink, kind string) int {
	nodeKinds := make(map[string]string, len(canvas.Nodes))
	for _, node := range canvas.Nodes {
		nodeKinds[node.ID] = node.Type
	}
	maxAttempt := 0
	for _, link := range links {
		if nodeKinds[link.NodeID] == kind && link.Attempt > maxAttempt {
			maxAttempt = link.Attempt
		}
	}
	return maxAttempt + 1
}

func shotNodeTitle(shot agentworkspace.ScriptShot, kind string) string {
	label := map[string]string{"image": "图片", "video": "视频", "audio": "配音"}[kind]
	return fmt.Sprintf("镜头 %d %s", shot.Sequence, label)
}

func writeScriptShotError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrProjectNotFound), errors.Is(err, agentworkspace.ErrScriptShotNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目、剧本节点或镜头不存在"})
	case errors.Is(err, agentworkspace.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "节点或镜头版本已变化，请刷新后重试"})
	case errors.Is(err, agentworkspace.ErrScriptShotInvalid), errors.Is(err, agentworkspace.ErrGenerationInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "镜头数据无效"})
	case errors.Is(err, errScriptSplitUpstream):
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "剧本拆分失败或结果格式无效"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "剧本镜头处理失败"})
	}
}
