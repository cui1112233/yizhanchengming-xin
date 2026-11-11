package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
	promptcatalog "qiantie/backend/internal/prompt"
	promptagent "qiantie/backend/internal/prompt/agent"
	promptstoryboard "qiantie/backend/internal/prompt/storyboard"
)

type AgentTextModelResolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)
type AgentTextCompletionExecutor func(context.Context, string, string, batchfactoryv11.TextCompletionRequest) (string, error)
type AgentImageGenerationInput struct {
	ProjectID         string
	Provider          string
	ModelID           string
	Mode              string
	Prompt            string
	Size              string
	AspectRatio       string
	Quality           string
	Resolution        string
	Count             int
	Advanced          map[string]any
	ReferenceAssetIDs []string
}

type AgentImageGenerator func(context.Context, string, AgentImageGenerationInput) ([]string, int, error)

type AgentNodeGenerationServices struct {
	Projects      agentworkspace.Store
	Catalog       novelfetchworkshop.Store
	TextModels    AgentTextModelResolver
	TextComplete  AgentTextCompletionExecutor
	Assets        agentworkspace.AssetStore
	Objects       agentworkspace.AssetObjectStore
	ImageGenerate AgentImageGenerator
	ImageProcess  AgentImageProcessor
	// QueueImages makes the HTTP endpoint persist work only; the backend
	// coordinator performs generation after the response is returned.
	QueueImages bool
	Now         func() time.Time
}

func resolveAgentTextModel(ctx context.Context, store novelfetchworkshop.Store, userConfig *UserConfigStore, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
	if provider, matched, err := ResolveIntegrationTextProvider(ctx, userConfig, owner, modelID); matched {
		return provider, err
	}
	if store == nil {
		return nil, fmt.Errorf("text model catalog is unavailable")
	}
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return nil, err
	}
	modelID = strings.TrimSpace(modelID)
	for _, model := range modelCatalog(settings) {
		if !strings.EqualFold(modelString(model, "kind"), "text") || !modelEnabled(model) {
			continue
		}
		if modelString(model, "id") != modelID && modelString(model, "modelId") != modelID {
			continue
		}
		endpoint := strings.TrimRight(modelString(model, "baseUrl"), "/")
		if !strings.HasSuffix(endpoint, "/chat/completions") {
			endpoint += "/chat/completions"
		}
		provider := &batchfactoryv11.OpenAICompatibleProvider{Endpoint: endpoint, APIKey: shuihuoModelCredential(model), Model: modelString(model, "modelId")}
		if err := provider.Validate(); err != nil {
			return nil, err
		}
		return provider, nil
	}
	return nil, fmt.Errorf("requested text model is not configured")
}

type agentTextGenerationRequest struct {
	NodeID            string         `json:"nodeId"`
	NodeVersion       int64          `json:"nodeVersion"`
	Provider          string         `json:"provider"`
	ModelID           string         `json:"modelId"`
	Mode              string         `json:"mode"`
	Prompt            string         `json:"prompt"`
	Parameters        map[string]any `json:"parameters"`
	Advanced          map[string]any `json:"advanced"`
	ReferenceAssetIDs []string       `json:"referenceAssetIds"`
	IdempotencyKey    string         `json:"idempotencyKey"`
}

type agentImageGenerationRequest struct {
	NodeID            string         `json:"nodeId"`
	NodeVersion       int64          `json:"nodeVersion"`
	Provider          string         `json:"provider"`
	ModelID           string         `json:"modelId"`
	Mode              string         `json:"mode"`
	Prompt            string         `json:"prompt"`
	Parameters        map[string]any `json:"parameters"`
	Advanced          map[string]any `json:"advanced"`
	Size              string         `json:"size"`
	AspectRatio       string         `json:"aspectRatio"`
	Quality           string         `json:"quality"`
	Resolution        string         `json:"resolution"`
	Count             int            `json:"count"`
	ReferenceAssetIDs []string       `json:"referenceAssetIds"`
	IdempotencyKey    string         `json:"idempotencyKey"`
}

func RegisterAgentNodeGenerationRoutes(root *http.ServeMux, middleware BridgeAuth, services AgentNodeGenerationServices) {
	if root == nil || services.Projects == nil {
		return
	}
	registerAgentCanvasInputPreviewRoute(root, middleware, services.Projects, services.Catalog)
	if services.Now == nil {
		services.Now = time.Now
	}
	registerAgentImageToolRoutes(root, middleware, services)
	if services.TextComplete == nil {
		services.TextComplete = newAgentTextCompletionExecutor(services.TextModels)
	}
	root.Handle("POST /api/agent/tasks/{id}/text-generations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input agentTextGenerationRequest
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文字生成参数无效"})
			return
		}
		input.NodeID = strings.TrimSpace(input.NodeID)
		input.Provider = strings.TrimSpace(input.Provider)
		input.ModelID = strings.TrimSpace(input.ModelID)
		input.Mode = strings.TrimSpace(input.Mode)
		input.Prompt = strings.TrimSpace(input.Prompt)
		input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
		if input.NodeID == "" || input.NodeVersion < 1 || input.ModelID == "" || input.Prompt == "" || input.IdempotencyKey == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择文字模型并填写生成要求"})
			return
		}
		textSettings := cloneAgentSettings(input.Parameters)
		if input.Mode != "" {
			textSettings["mode"] = input.Mode
		}
		if err := ValidateModelSettings(map[string]any{"kind": "text"}, "text", textSettings); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		canvas, err := services.Projects.Canvas(req.Context(), identity.Username, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		if !agentGenerationNodeMatches(canvas, input.NodeID, "text", input.NodeVersion) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "文字节点已变化，请重新打开后再生成"})
			return
		}
		now := services.Now().UTC()
		generation, err := beginAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, "text", input.NodeID, input.NodeVersion, input.Provider, input.ModelID, input.Mode, input.Prompt, input.Parameters, input.Advanced, input.ReferenceAssetIDs, input.IdempotencyKey, now)
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		if services.TextComplete == nil {
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, "文字模型执行器未配置", services.Now().UTC())
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文字模型执行器未配置"})
			return
		}
		execution, err := services.Projects.BeginNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.Execution{
			ID: newAgentID("execution"), ToolName: "text-generation", Status: "running",
			Detail:         map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "provider": input.Provider, "modelId": input.ModelID, "mode": input.Mode, "parameters": input.Parameters, "advanced": input.Advanced, "generationTaskId": generation.ID},
			IdempotencyKey: input.IdempotencyKey, CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, err.Error(), services.Now().UTC())
			writeAgentNodeGenerationError(w, err)
			return
		}
		result, err := services.TextComplete(req.Context(), identity.Username, input.ModelID, batchfactoryv11.TextCompletionRequest{
			SystemPrompt: promptagent.TextNode.SystemPrompt, UserPrompt: input.Prompt,
			Temperature: 0.7, MaxTokens: 4096, DisableJSONResponse: true,
		})
		result = strings.TrimSpace(result)
		if err != nil || result == "" {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{
				"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "modelId": input.ModelID, "error": "文字模型生成失败",
			}, services.Now().UTC())
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, "文字模型生成失败", services.Now().UTC())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "文字模型生成失败，请稍后重试"})
			return
		}
		updatedCanvas, completed, err := services.Projects.ApplyNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.ApplyNodeExecutionInput{
			ExecutionID: execution.ID, NodeID: input.NodeID, NodeVersion: input.NodeVersion, NodeType: "text", Summary: result,
			Detail: map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "provider": input.Provider, "modelId": input.ModelID, "mode": input.Mode, "parameters": input.Parameters, "advanced": input.Advanced, "generationTaskId": generation.ID}, Now: services.Now().UTC(),
		})
		if err != nil {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"error": "节点版本已变化", "result": result}, services.Now().UTC())
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, "节点版本已变化", services.Now().UTC())
			writeAgentNodeGenerationError(w, err)
			return
		}
		generation, err = services.Projects.TransitionGenerationTask(req.Context(), identity.Username, projectID, generation.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationCompleted, Now: services.Now().UTC()})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, updatedCanvas), "execution": completed, "generation": generation, "result": result})
	})))
	root.Handle("POST /api/agent/tasks/{id}/image-generations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input agentImageGenerationRequest
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "图片生成参数无效"})
			return
		}
		input.NodeID = strings.TrimSpace(input.NodeID)
		input.Provider = strings.TrimSpace(input.Provider)
		input.ModelID = strings.TrimSpace(input.ModelID)
		input.Mode = strings.TrimSpace(input.Mode)
		input.Prompt = strings.TrimSpace(input.Prompt)
		input.Size = strings.TrimSpace(input.Size)
		input.AspectRatio = strings.TrimSpace(input.AspectRatio)
		input.Quality = strings.TrimSpace(input.Quality)
		input.Resolution = strings.TrimSpace(input.Resolution)
		input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
		if input.NodeID == "" || input.NodeVersion < 1 || input.ModelID == "" || input.Prompt == "" || input.IdempotencyKey == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择图片模型并填写生成要求"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		canvas, err := services.Projects.Canvas(req.Context(), identity.Username, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		if !agentGenerationNodeMatches(canvas, input.NodeID, "image", input.NodeVersion) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "图片节点已变化，请重新打开后再生成"})
			return
		}
		capability, err := resolveAgentCanvasInputCapability(req.Context(), services.Catalog, identity.Username, "image", input.Provider, input.ModelID)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		resolvedInputs, err := agentResolvedCanvasInputSnapshotWithCapability(canvas, input.NodeID, input.Prompt, input.ReferenceAssetIDs, capability)
		if err != nil {
			writeAgentCanvasInputError(w, err)
			return
		}
		for _, referenceID := range input.ReferenceAssetIDs {
			if services.Assets == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "项目素材存储尚未配置"})
				return
			}
			if _, err := services.Assets.Asset(req.Context(), identity.Username, projectID, strings.TrimSpace(referenceID)); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "引用素材不存在或不属于当前项目"})
				return
			}
		}
		if services.ImageGenerate == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "图片模型执行器未配置", "code": ErrorCodeImageModelNotConfigured})
			return
		}
		generationInput, settingsErr := normalizeAgentImageGenerationInput(input)
		if settingsErr != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": settingsErr.Error()})
			return
		}
		now := services.Now().UTC()
		mode := input.Mode
		if mode == "" {
			mode = "image_generation"
		}
		provider := input.Provider
		if provider == "" {
			provider = "configured_model"
		}
		settings := map[string]any{"prompt": input.Prompt, "parameters": cloneAgentSettings(input.Parameters), "advanced": cloneAgentSettings(input.Advanced), "size": generationInput.Size, "aspectRatio": generationInput.AspectRatio, "quality": generationInput.Quality, "resolution": generationInput.Resolution, "count": generationInput.Count, "resolvedInputs": resolvedInputs}
		generation, err := services.Projects.CreateGenerationTask(req.Context(), identity.Username, agentworkspace.CreateGenerationTaskInput{
			ID: newAgentID("generation"), ProjectID: projectID, NodeID: input.NodeID, NodeVersion: input.NodeVersion, Kind: "image", Provider: provider, ModelID: input.ModelID, Mode: mode,
			Settings: settings, ReferenceAssetIDs: uniqueAgentGenerationIDs(input.ReferenceAssetIDs), IdempotencyKey: input.IdempotencyKey, Now: now,
		})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		if services.QueueImages {
			writeJSON(w, http.StatusAccepted, map[string]any{"generation": generation})
			return
		}
		generation, err = services.Projects.TransitionGenerationTask(req.Context(), identity.Username, projectID, generation.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: generation.UpdatedAt, Now: now})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}

		generationInput.ProjectID = projectID
		generationInput.ReferenceAssetIDs = uniqueAgentGenerationIDs(input.ReferenceAssetIDs)
		execution, err := services.Projects.BeginNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.Execution{
			ID: newAgentID("execution"), ToolName: "image-generation", Status: "running",
			Detail:         map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "provider": input.Provider, "modelId": input.ModelID, "mode": input.Mode, "parameters": input.Parameters, "advanced": input.Advanced, "generationTaskId": generation.ID},
			IdempotencyKey: input.IdempotencyKey, CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, err.Error(), services.Now().UTC())
			writeAgentNodeGenerationError(w, err)
			return
		}
		resultURLs, status, err := services.ImageGenerate(req.Context(), identity.Username, generationInput)
		if err != nil {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"modelId": input.ModelID, "error": err.Error()}, services.Now().UTC())
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, err.Error(), services.Now().UTC())
			if status < 400 || status > 599 {
				status = http.StatusBadGateway
			}
			payload := map[string]string{"error": err.Error()}
			if code := errorCodeOf(err); code != "" {
				payload["code"] = code
			}
			writeJSON(w, status, payload)
			return
		}
		if len(resultURLs) == 0 {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"modelId": input.ModelID, "error": "图片模型未返回图片结果"}, services.Now().UTC())
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, "图片模型未返回图片结果", services.Now().UTC())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片模型未返回图片结果"})
			return
		}
		ingestor := newAgentMediaAssetIngestor(services.Assets, services.Objects, nil, services.Now)
		assets := make([]agentworkspace.Asset, 0, len(resultURLs))
		for _, resultURL := range resultURLs {
			asset, saveErr := ingestor.SaveImageResult(req.Context(), identity.Username, projectID, newAgentID("asset"), resultURL)
			if saveErr != nil {
				err = saveErr
				break
			}
			assets = append(assets, asset)
		}
		if err != nil || len(assets) == 0 {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"modelId": input.ModelID, "error": err.Error()}, services.Now().UTC())
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, "图片结果保存失败", services.Now().UTC())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片结果保存失败：" + err.Error()})
			return
		}
		assetIDs := make([]string, 0, len(assets))
		assetViews := make([]map[string]any, 0, len(assets))
		generationResults := make([]agentworkspace.GenerationResult, 0, len(assets))
		for _, asset := range assets {
			assetIDs = append(assetIDs, asset.ID)
			assetView := map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": asset.Kind, "name": asset.Name, "url": asset.URL, "contentType": asset.ContentType, "byteSize": asset.ByteSize}
			assetViews = append(assetViews, assetView)
			generationResults = append(generationResults, agentworkspace.GenerationResult{ID: asset.ID, AssetID: asset.ID, Kind: asset.Kind, Asset: assetView, CreatedAt: asset.CreatedAt})
		}
		detail := map[string]any{
			"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "provider": input.Provider, "modelId": input.ModelID,
			"mode": input.Mode, "parameters": input.Parameters, "advanced": input.Advanced, "resultAssetIds": assetIDs,
			"resultAssets": assetViews, "aspectRatio": generationInput.AspectRatio, "quality": generationInput.Quality,
			"resolution": generationInput.Resolution, "count": generationInput.Count, "generationTaskId": generation.ID,
		}
		pending, err := services.Projects.SetNodeExecutionAwaitingConfirmation(req.Context(), identity.Username, projectID, execution.ID, detail, services.Now().UTC())
		if err != nil {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"error": "图片结果进入待确认状态失败", "resultAssetIds": assetIDs}, services.Now().UTC())
			failAgentGenerationTask(req.Context(), services.Projects, identity.Username, projectID, generation, "图片结果进入待确认状态失败", services.Now().UTC())
			writeAgentNodeGenerationError(w, err)
			return
		}
		generation, err = services.Projects.TransitionGenerationTask(req.Context(), identity.Username, projectID, generation.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: generationResults, SelectedResultID: generationResults[0].ID, Now: services.Now().UTC()})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"execution": pending, "generation": generation, "resultAssets": assets})
	})))
	root.Handle("POST /api/agent/tasks/{id}/image-generations/{executionID}/confirm", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			AssetID string `json:"assetId"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "图片确认参数无效"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		executionID := strings.TrimSpace(req.PathValue("executionID"))
		input.AssetID = strings.TrimSpace(input.AssetID)
		if executionID == "" || input.AssetID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择要使用的图片"})
			return
		}
		executions, err := services.Projects.Executions(req.Context(), identity.Username, projectID)
		if err != nil {
			writeAgentNodeGenerationError(w, err)
			return
		}
		var execution *agentworkspace.Execution
		for index := range executions {
			if executions[index].ID == executionID {
				execution = &executions[index]
				break
			}
		}
		if execution == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "图片生成记录不存在"})
			return
		}
		if execution.ToolName != "image-generation" || execution.Status != "awaiting_confirmation" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前图片结果不能再确认"})
			return
		}
		if !agentExecutionIncludesAsset(execution.Detail, input.AssetID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "所选图片不属于本次生成结果"})
			return
		}
		if services.Assets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "项目素材存储尚未配置"})
			return
		}
		asset, err := services.Assets.Asset(req.Context(), identity.Username, projectID, input.AssetID)
		if err != nil || asset.Kind != "image" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "所选图片不存在或不属于当前项目"})
			return
		}
		asset.URL = agentAssetContentURL(projectID, asset.ID)
		assetRef := map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": asset.Kind, "name": asset.Name, "url": asset.URL, "contentType": asset.ContentType}
		detail := cloneAgentSettings(execution.Detail)
		detail["assetId"] = asset.ID
		detail["selectedAssetId"] = asset.ID
		updatedCanvas, completed, err := services.Projects.ApplyNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.ApplyNodeExecutionInput{
			ExecutionID: execution.ID, NodeID: agentExecutionString(execution.Detail["nodeId"]), NodeVersion: agentExecutionInt64(execution.Detail["nodeVersion"]),
			NodeType: "image", AssetRef: assetRef, Detail: detail, Now: services.Now().UTC(),
		})
		if err != nil {
			writeAgentNodeGenerationError(w, err)
			return
		}
		if generationID := agentExecutionString(execution.Detail["generationTaskId"]); generationID != "" {
			generation, generationErr := services.Projects.TransitionGenerationTask(req.Context(), identity.Username, projectID, generationID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationCompleted, SelectedResultID: asset.ID, Now: services.Now().UTC()})
			if generationErr != nil {
				writeAgentGenerationTaskError(w, generationErr)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, updatedCanvas), "execution": completed, "generation": generation, "resultAsset": asset})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, updatedCanvas), "execution": completed, "resultAsset": asset})
	})))
}

func newAgentTextCompletionExecutor(resolver AgentTextModelResolver) AgentTextCompletionExecutor {
	if resolver == nil {
		return nil
	}
	return func(ctx context.Context, owner, modelID string, request batchfactoryv11.TextCompletionRequest) (string, error) {
		provider, err := resolver(ctx, owner, modelID)
		if err != nil || provider == nil {
			return "", fmt.Errorf("requested text model is unavailable")
		}
		return provider.Complete(ctx, request)
	}
}

func beginAgentGenerationTask(ctx context.Context, store agentworkspace.Store, owner, projectID, kind, nodeID string, nodeVersion int64, provider, modelID, mode, prompt string, parameters, advanced map[string]any, references []string, idempotencyKey string, now time.Time) (agentworkspace.GenerationTask, error) {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		provider = "configured_model"
	}
	settings := map[string]any{"prompt": strings.TrimSpace(prompt), "parameters": cloneAgentSettings(parameters), "advanced": cloneAgentSettings(advanced)}
	definition := promptagent.TextNode
	if strings.TrimSpace(mode) == "script_split" {
		definition = promptstoryboard.ScriptSplit
	}
	settings["promptDefinition"] = agentPromptDefinitionSetting(definition)
	if strings.TrimSpace(mode) != "" {
		settings["mode"] = strings.TrimSpace(mode)
	}
	task, err := store.CreateGenerationTask(ctx, owner, agentworkspace.CreateGenerationTaskInput{
		ID: newAgentID("generation"), ProjectID: projectID, NodeID: nodeID, NodeVersion: nodeVersion, Kind: kind,
		Provider: provider, ModelID: modelID, Mode: mode, Settings: settings, ReferenceAssetIDs: uniqueAgentGenerationIDs(references),
		IdempotencyKey: idempotencyKey, Now: now,
	})
	if err != nil {
		return agentworkspace.GenerationTask{}, err
	}
	if task.State != agentworkspace.GenerationQueued {
		return task, nil
	}
	return store.TransitionGenerationTask(ctx, owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: task.UpdatedAt, Now: now})
}

func agentPromptDefinitionSetting(definition promptcatalog.Definition) map[string]any {
	return map[string]any{"id": definition.ID, "version": definition.Version}
}

func failAgentGenerationTask(ctx context.Context, store agentworkspace.Store, owner, projectID string, task agentworkspace.GenerationTask, message string, now time.Time) {
	if task.ID == "" || task.State == agentworkspace.GenerationCompleted || task.State == agentworkspace.GenerationFailed || task.State == agentworkspace.GenerationCancelled {
		return
	}
	_, _ = store.TransitionGenerationTask(ctx, owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: strings.TrimSpace(message), Now: now})
}

func agentExecutionIncludesAsset(detail map[string]any, assetID string) bool {
	values, ok := detail["resultAssetIds"].([]any)
	if !ok {
		if typed, typedOK := detail["resultAssetIds"].([]string); typedOK {
			for _, value := range typed {
				if strings.TrimSpace(value) == assetID {
					return true
				}
			}
		}
		return false
	}
	for _, value := range values {
		if agentExecutionString(value) == assetID {
			return true
		}
	}
	return false
}

func agentExecutionString(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func agentExecutionInt64(value any) int64 {
	switch number := value.(type) {
	case int64:
		return number
	case int:
		return int64(number)
	case float64:
		return int64(number)
	default:
		return 0
	}
}

func normalizeAgentImageGenerationInput(input agentImageGenerationRequest) (AgentImageGenerationInput, error) {
	parameters := cloneAgentSettings(input.Parameters)
	if _, ok := parameters["ratio"]; !ok && input.AspectRatio != "" {
		parameters["ratio"] = input.AspectRatio
	}
	if _, ok := parameters["quality"]; !ok && input.Quality != "" {
		parameters["quality"] = input.Quality
	}
	if _, ok := parameters["resolution"]; !ok && input.Resolution != "" {
		parameters["resolution"] = input.Resolution
	}
	if _, ok := parameters["count"]; !ok && input.Count != 0 {
		parameters["count"] = input.Count
	}
	modeForValidation := input.Mode
	if modeForValidation == "image_generation" {
		// Canvas uses image_generation as its default task label; model
		// capabilities use text_to_image for the equivalent generation mode.
		modeForValidation = "text_to_image"
	}
	if modeForValidation != "" {
		parameters["mode"] = modeForValidation
	}
	if err := ValidateModelSettings(map[string]any{"kind": "image"}, "image", parameters); err != nil {
		return AgentImageGenerationInput{}, err
	}
	ratio := settingString(parameters, "ratio")
	if ratio == "" {
		ratio = "16:9"
	}
	quality := settingString(parameters, "quality")
	if quality == "" {
		quality = "standard"
	}
	resolution := strings.ToUpper(settingString(parameters, "resolution"))
	if resolution == "" {
		resolution = "2K"
	}
	count := settingInt(parameters, "count")
	if count == 0 {
		count = 1
	}
	if !batchfactoryv11.IsSupportedImageAspectRatio(ratio) {
		return AgentImageGenerationInput{}, fmt.Errorf("图片比例无效")
	}
	if quality != "low" && quality != "standard" && quality != "high" {
		return AgentImageGenerationInput{}, fmt.Errorf("图片画质无效")
	}
	if resolution != "1K" && resolution != "2K" && resolution != "4K" {
		return AgentImageGenerationInput{}, fmt.Errorf("图片清晰度无效")
	}
	if count < 1 || count > 4 {
		return AgentImageGenerationInput{}, fmt.Errorf("图片数量必须为 1 到 4 张")
	}
	sizes := map[string]map[string]string{
		"1K": {"16:9": "1024x576", "9:16": "576x1024", "21:9": "1344x576", "1:1": "1024x1024", "3:4": "768x1024", "4:3": "1024x768", "2:3": "768x1152", "3:2": "1152x768"},
		"2K": {"16:9": "2048x1152", "9:16": "1152x2048", "21:9": "2688x1152", "1:1": "2048x2048", "3:4": "1536x2048", "4:3": "2048x1536", "2:3": "1536x2304", "3:2": "2304x1536"},
		"4K": {"16:9": "4096x2304", "9:16": "2304x4096", "21:9": "5376x2304", "1:1": "4096x4096", "3:4": "3072x4096", "4:3": "4096x3072", "2:3": "3072x4608", "3:2": "4608x3072"},
	}
	return AgentImageGenerationInput{Provider: input.Provider, ModelID: input.ModelID, Mode: input.Mode, Prompt: input.Prompt, Size: sizes[resolution][ratio], AspectRatio: ratio, Quality: quality, Resolution: resolution, Count: count, Advanced: cloneAgentSettings(input.Advanced)}, nil
}

func cloneAgentSettings(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func agentGenerationNodeMatches(canvas agentworkspace.Canvas, nodeID, nodeType string, version int64) bool {
	for _, node := range canvas.Nodes {
		if node.ID == nodeID {
			return node.Type == nodeType && node.Version == version
		}
	}
	return false
}

func writeAgentNodeGenerationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrExecutionExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "该生成请求已提交，请勿重复操作"})
	case errors.Is(err, agentworkspace.ErrExecutionNodeChanged), errors.Is(err, agentworkspace.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "节点已变化，请重新打开后再生成"})
	case errors.Is(err, agentworkspace.ErrProjectNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成结果保存失败"})
	}
}
