package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type agentImageToolRequest struct {
	NodeID         string         `json:"nodeId"`
	NodeVersion    int64          `json:"nodeVersion"`
	Operation      string         `json:"operation"`
	SourceAssetID  string         `json:"sourceAssetId"`
	Settings       map[string]any `json:"settings"`
	IdempotencyKey string         `json:"idempotencyKey"`
}

func registerAgentImageToolRoutes(root *http.ServeMux, middleware BridgeAuth, services AgentNodeGenerationServices) {
	if root == nil || services.Projects == nil || services.Assets == nil || services.Objects == nil {
		return
	}
	if services.Now == nil {
		services.Now = time.Now
	}
	root.Handle("POST /api/agent/tasks/{id}/image-tools", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input agentImageToolRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "图片编辑参数无效"})
			return
		}
		input.NodeID = strings.TrimSpace(input.NodeID)
		input.SourceAssetID = strings.TrimSpace(input.SourceAssetID)
		input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
		operation, err := agentworkspace.ParseImageToolOperation(input.Operation)
		if err != nil || input.NodeID == "" || input.NodeVersion < 1 || input.SourceAssetID == "" || input.IdempotencyKey == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "图片编辑参数不完整"})
			return
		}
		settings := cloneAgentSettings(input.Settings)
		if operation == agentworkspace.ImageMirror && strings.TrimSpace(fmt.Sprint(settings["direction"])) == "" {
			settings["direction"] = "horizontal"
		}
		if err := agentworkspace.ValidateImageToolSettings(operation, settings); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "图片编辑参数无效"})
			return
		}
		owner, projectID := strings.TrimSpace(identity.Username), strings.TrimSpace(req.PathValue("id"))
		canvas, err := services.Projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		if !agentGenerationNodeMatches(canvas, input.NodeID, "image", input.NodeVersion) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "图片节点已变化，请重新打开后再编辑"})
			return
		}
		source, err := services.Assets.Asset(req.Context(), owner, projectID, input.SourceAssetID)
		if err != nil || source.Kind != "image" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择当前项目中的图片素材"})
			return
		}
		task, err := services.Projects.CreateGenerationTask(req.Context(), owner, agentworkspace.CreateGenerationTaskInput{
			ID: newAgentID("generation"), ProjectID: projectID, NodeID: input.NodeID, NodeVersion: input.NodeVersion,
			Kind: "image", Provider: "local_image_tools", ModelID: "local_image_tools", Mode: operation.GenerationMode(),
			Settings:          map[string]any{"operation": string(operation), "sourceAssetId": source.ID, "settings": settings, "source": map[string]any{"assetId": source.ID, "name": source.Name, "contentType": source.ContentType}},
			ReferenceAssetIDs: []string{source.ID}, IdempotencyKey: input.IdempotencyKey, Now: services.Now().UTC(),
		})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"task": task})
	})))
}

func runAgentImageTool(ctx context.Context, owner, projectID string, task agentworkspace.GenerationTask, services AgentNodeGenerationServices) ([]AgentImageProcessOutput, error) {
	if services.ImageProcess == nil || services.Assets == nil || services.Objects == nil {
		return nil, fmt.Errorf("图片编辑服务尚未配置")
	}
	operation, err := agentworkspace.ParseImageToolOperation(strings.TrimPrefix(strings.TrimSpace(task.Mode), "image_"))
	if err != nil {
		return nil, err
	}
	sourceID := strings.TrimSpace(fmt.Sprint(task.Settings["sourceAssetId"]))
	if sourceID == "" && len(task.ReferenceAssetIDs) == 1 {
		sourceID = strings.TrimSpace(task.ReferenceAssetIDs[0])
	}
	source, err := services.Assets.Asset(ctx, owner, projectID, sourceID)
	if err != nil || source.Kind != "image" {
		return nil, fmt.Errorf("源图片不存在或不属于当前项目")
	}
	body, err := services.Objects.Open(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("读取源图片失败")
	}
	defer body.Close()
	contents, err := io.ReadAll(io.LimitReader(body, agentImageResultMaxBytes+1))
	if err != nil || len(contents) == 0 || len(contents) > agentImageResultMaxBytes {
		return nil, fmt.Errorf("读取源图片失败")
	}
	settings, _ := task.Settings["settings"].(map[string]any)
	return services.ImageProcess.Process(ctx, AgentImageProcessInput{Operation: operation, Source: source, SourceBytes: contents, Settings: settings})
}

func persistAgentImageToolOutputs(ctx context.Context, owner, projectID string, outputs []AgentImageProcessOutput, services AgentNodeGenerationServices) ([]agentworkspace.GenerationResult, error) {
	if len(outputs) == 0 || services.Assets == nil || services.Objects == nil {
		return nil, fmt.Errorf("图片编辑没有返回结果")
	}
	results := make([]agentworkspace.GenerationResult, 0, len(outputs))
	for index, output := range outputs {
		if len(output.Bytes) == 0 || strings.TrimSpace(output.ContentType) != "image/png" || strings.TrimSpace(output.Extension) != "png" {
			return nil, fmt.Errorf("图片编辑返回了无效结果")
		}
		assetID := newAgentID("asset")
		objectKey, err := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, output.Extension)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSpace(output.Name)
		if name == "" {
			name = fmt.Sprintf("图片编辑结果-%d.png", index+1)
		}
		asset := agentworkspace.Asset{ID: assetID, ProjectID: projectID, Kind: "image", Name: path.Base(name), ContentType: output.ContentType, ObjectKey: objectKey, ByteSize: int64(len(output.Bytes)), CreatedAt: services.Now().UTC()}
		if err := services.Objects.Put(ctx, asset, bytes.NewReader(output.Bytes)); err != nil {
			return nil, err
		}
		if err := services.Assets.CreateAsset(ctx, owner, asset); err != nil {
			_ = services.Objects.Delete(ctx, asset)
			return nil, err
		}
		asset.URL = agentAssetContentURL(projectID, asset.ID)
		results = append(results, agentworkspace.GenerationResult{ID: asset.ID, AssetID: asset.ID, Kind: "image", Asset: map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": "image", "name": asset.Name, "url": asset.URL, "contentType": asset.ContentType, "byteSize": asset.ByteSize}, CreatedAt: asset.CreatedAt})
	}
	return results, nil
}

func imageToolTaskMode(mode string) bool {
	for _, operation := range []agentworkspace.ImageToolOperation{agentworkspace.ImageCrop, agentworkspace.ImageRotate, agentworkspace.ImageMirror, agentworkspace.ImageGridSplit} {
		if mode == operation.GenerationMode() {
			return true
		}
	}
	return false
}
