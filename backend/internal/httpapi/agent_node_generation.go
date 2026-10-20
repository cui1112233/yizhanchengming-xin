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
)

type AgentTextModelResolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)
type AgentImageGenerationInput struct {
	ModelID     string
	Prompt      string
	Size        string
	AspectRatio string
	Quality     string
	Resolution  string
	Count       int
}

type AgentImageGenerator func(context.Context, string, AgentImageGenerationInput) ([]string, int, error)

type AgentNodeGenerationServices struct {
	Projects      agentworkspace.Store
	TextModels    AgentTextModelResolver
	Assets        agentworkspace.AssetStore
	Objects       agentworkspace.AssetObjectStore
	ImageGenerate AgentImageGenerator
	Now           func() time.Time
}

func resolveAgentTextModel(ctx context.Context, store novelfetchworkshop.Store, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
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
	NodeID         string `json:"nodeId"`
	NodeVersion    int64  `json:"nodeVersion"`
	ModelID        string `json:"modelId"`
	Prompt         string `json:"prompt"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type agentImageGenerationRequest struct {
	NodeID            string   `json:"nodeId"`
	NodeVersion       int64    `json:"nodeVersion"`
	ModelID           string   `json:"modelId"`
	Prompt            string   `json:"prompt"`
	Size              string   `json:"size"`
	AspectRatio       string   `json:"aspectRatio"`
	Quality           string   `json:"quality"`
	Resolution        string   `json:"resolution"`
	Count             int      `json:"count"`
	ReferenceAssetIDs []string `json:"referenceAssetIds"`
	IdempotencyKey    string   `json:"idempotencyKey"`
}

func RegisterAgentNodeGenerationRoutes(root *http.ServeMux, middleware BridgeAuth, services AgentNodeGenerationServices) {
	if root == nil || services.Projects == nil {
		return
	}
	if services.Now == nil {
		services.Now = time.Now
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
		input.ModelID = strings.TrimSpace(input.ModelID)
		input.Prompt = strings.TrimSpace(input.Prompt)
		input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
		if input.NodeID == "" || input.NodeVersion < 1 || input.ModelID == "" || input.Prompt == "" || input.IdempotencyKey == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择文字模型并填写生成要求"})
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
		if services.TextModels == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "文字模型执行器未配置"})
			return
		}
		provider, err := services.TextModels(req.Context(), identity.Username, input.ModelID)
		if err != nil || provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "所选文字模型不可用，请检查 API 配置"})
			return
		}
		now := services.Now().UTC()
		execution, err := services.Projects.BeginNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.Execution{
			ID: newAgentID("execution"), ToolName: "text-generation", Status: "running",
			Detail:         map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "modelId": input.ModelID},
			IdempotencyKey: input.IdempotencyKey, CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			writeAgentNodeGenerationError(w, err)
			return
		}
		result, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{
			SystemPrompt: promptcatalog.AgentTextNodeSystemPrompt, UserPrompt: input.Prompt,
			Temperature: 0.7, MaxTokens: 4096, DisableJSONResponse: true,
		})
		result = strings.TrimSpace(result)
		if err != nil || result == "" {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{
				"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "modelId": input.ModelID, "error": "文字模型生成失败",
			}, services.Now().UTC())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "文字模型生成失败，请稍后重试"})
			return
		}
		updatedCanvas, completed, err := services.Projects.ApplyNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.ApplyNodeExecutionInput{
			ExecutionID: execution.ID, NodeID: input.NodeID, NodeVersion: input.NodeVersion, NodeType: "text", Summary: result,
			Detail: map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "modelId": input.ModelID}, Now: services.Now().UTC(),
		})
		if err != nil {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"error": "节点版本已变化", "result": result}, services.Now().UTC())
			writeAgentNodeGenerationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, updatedCanvas), "execution": completed, "result": result})
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
		input.ModelID = strings.TrimSpace(input.ModelID)
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
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "图片模型执行器未配置"})
			return
		}
		now := services.Now().UTC()
		execution, err := services.Projects.BeginNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.Execution{
			ID: newAgentID("execution"), ToolName: "image-generation", Status: "running",
			Detail:         map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "modelId": input.ModelID},
			IdempotencyKey: input.IdempotencyKey, CreatedAt: now, UpdatedAt: now,
		})
		if err != nil {
			writeAgentNodeGenerationError(w, err)
			return
		}
		generationInput, settingsErr := normalizeAgentImageGenerationInput(input)
		if settingsErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": settingsErr.Error()})
			return
		}
		resultURLs, status, err := services.ImageGenerate(req.Context(), identity.Username, generationInput)
		if err != nil {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"modelId": input.ModelID, "error": err.Error()}, services.Now().UTC())
			if status < 400 || status > 599 {
				status = http.StatusBadGateway
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		if len(resultURLs) == 0 {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"modelId": input.ModelID, "error": "图片模型未返回图片结果"}, services.Now().UTC())
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
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片结果保存失败：" + err.Error()})
			return
		}
		asset := assets[0]
		assetRef := map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": asset.Kind, "url": asset.URL, "contentType": asset.ContentType}
		updatedCanvas, completed, err := services.Projects.ApplyNodeExecution(req.Context(), identity.Username, projectID, agentworkspace.ApplyNodeExecutionInput{
			ExecutionID: execution.ID, NodeID: input.NodeID, NodeVersion: input.NodeVersion, NodeType: "image", AssetRef: assetRef,
			Detail: map[string]any{"nodeId": input.NodeID, "nodeVersion": input.NodeVersion, "modelId": input.ModelID, "assetId": asset.ID, "aspectRatio": generationInput.AspectRatio, "quality": generationInput.Quality, "resolution": generationInput.Resolution, "count": generationInput.Count}, Now: services.Now().UTC(),
		})
		if err != nil {
			_, _ = services.Projects.FailNodeExecution(req.Context(), identity.Username, projectID, execution.ID, map[string]any{"error": "节点版本已变化", "assetId": asset.ID}, services.Now().UTC())
			writeAgentNodeGenerationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"canvas": agentCanvasForResponse(projectID, updatedCanvas), "execution": completed, "resultAsset": asset, "resultAssets": assets})
	})))
}

func normalizeAgentImageGenerationInput(input agentImageGenerationRequest) (AgentImageGenerationInput, error) {
	ratio := input.AspectRatio
	if ratio == "" {
		ratio = "16:9"
	}
	quality := input.Quality
	if quality == "" {
		quality = "standard"
	}
	resolution := strings.ToUpper(input.Resolution)
	if resolution == "" {
		resolution = "2K"
	}
	count := input.Count
	if count == 0 {
		count = 1
	}
	if ratio != "16:9" && ratio != "9:16" && ratio != "1:1" {
		return AgentImageGenerationInput{}, fmt.Errorf("图片比例无效")
	}
	if quality != "standard" && quality != "high" {
		return AgentImageGenerationInput{}, fmt.Errorf("图片画质无效")
	}
	if resolution != "1K" && resolution != "2K" && resolution != "4K" {
		return AgentImageGenerationInput{}, fmt.Errorf("图片清晰度无效")
	}
	if count < 1 || count > 4 {
		return AgentImageGenerationInput{}, fmt.Errorf("图片数量必须为 1 到 4 张")
	}
	sizes := map[string]map[string]string{
		"1K": {"16:9": "1024x576", "9:16": "576x1024", "1:1": "1024x1024"},
		"2K": {"16:9": "2048x1152", "9:16": "1152x2048", "1:1": "2048x2048"},
		"4K": {"16:9": "4096x2304", "9:16": "2304x4096", "1:1": "4096x4096"},
	}
	return AgentImageGenerationInput{ModelID: input.ModelID, Prompt: input.Prompt, Size: sizes[resolution][ratio], AspectRatio: ratio, Quality: quality, Resolution: resolution, Count: count}, nil
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
