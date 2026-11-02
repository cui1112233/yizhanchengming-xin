package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
)

type agentCanvasInputPreviewRequest struct {
	NodeID   string `json:"nodeId"`
	Provider string `json:"provider"`
	ModelID  string `json:"modelId"`
}

func registerAgentCanvasInputPreviewRoute(root *http.ServeMux, middleware BridgeAuth, projects agentworkspace.Store, catalog novelfetchworkshop.Store) {
	if root == nil || projects == nil {
		return
	}
	root.Handle("POST /api/agent/tasks/{id}/canvas-inputs/preview", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input agentCanvasInputPreviewRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || strings.TrimSpace(input.NodeID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择需要预览输入的节点"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		canvas, err := projects.Canvas(req.Context(), identity.Username, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		node, found := agentCanvasInputNode(canvas, input.NodeID)
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "画布节点不存在"})
			return
		}
		capability, err := resolveAgentCanvasInputCapability(req.Context(), catalog, identity.Username, node.Type, input.Provider, input.ModelID)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		resolved, err := agentworkspace.ResolveCanvasInputs(canvas, input.NodeID, capability)
		if err != nil {
			writeAgentCanvasInputError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"inputs": resolved})
	})))
}

func agentCanvasInputNode(canvas agentworkspace.Canvas, nodeID string) (agentworkspace.Node, bool) {
	for _, node := range canvas.Nodes {
		if node.ID == strings.TrimSpace(nodeID) {
			return node, true
		}
	}
	return agentworkspace.Node{}, false
}

func agentCanvasInputCapability(nodeType string) agentworkspace.CanvasInputCapability {
	switch strings.TrimSpace(nodeType) {
	case "image", "video":
		return agentworkspace.CanvasInputCapability{AcceptsTextPrompt: true, AcceptsImageReferences: true, MaxReferenceAssets: 10}
	case "audio":
		return agentworkspace.CanvasInputCapability{AcceptsTextPrompt: true}
	case "text":
		return agentworkspace.CanvasInputCapability{AcceptsTextPrompt: true}
	default:
		return agentworkspace.CanvasInputCapability{}
	}
}

func resolveAgentCanvasInputCapability(ctx context.Context, catalog novelfetchworkshop.Store, owner, nodeType, provider, modelID string) (agentworkspace.CanvasInputCapability, error) {
	nodeType, provider, modelID = strings.TrimSpace(nodeType), strings.TrimSpace(provider), strings.TrimSpace(modelID)
	if provider == "doubao_local_executor" && (nodeType == "image" || nodeType == "video") {
		return agentCanvasInputCapabilityFromPublic(publicModelCapabilities(map[string]any{"kind": nodeType})), nil
	}
	if catalog == nil || modelID == "" {
		return agentCanvasInputCapability(nodeType), nil
	}
	settings, err := catalog.GetConfig(ctx, owner)
	if err != nil {
		return agentworkspace.CanvasInputCapability{}, err
	}
	for _, model := range modelCatalog(settings) {
		if modelString(model, "id") != modelID && modelString(model, "modelId") != modelID {
			continue
		}
		if modelString(model, "kind") != nodeType || !modelEnabled(model) {
			return agentworkspace.CanvasInputCapability{}, fmt.Errorf("当前模型不可用")
		}
		return agentCanvasInputCapabilityFromPublic(publicModelCapabilities(model)), nil
	}
	return agentworkspace.CanvasInputCapability{}, fmt.Errorf("当前模型不可用")
}

func agentCanvasInputCapabilityFromPublic(capabilities PublicModelCapabilities) agentworkspace.CanvasInputCapability {
	return agentworkspace.CanvasInputCapability{
		AcceptsTextPrompt: capabilities.AcceptsTextPrompt, AcceptsImageReferences: capabilities.AcceptsImageReferences,
		AcceptsVideoReferences: capabilities.AcceptsVideoReferences, AcceptsAudioReferences: capabilities.AcceptsAudioReferences,
		MaxReferenceAssets: capabilities.MaxReferenceAssets,
	}
}

func writeAgentCanvasInputError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrCanvasInputTargetNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "画布节点不存在"})
	case errors.Is(err, agentworkspace.ErrCanvasInputUnsupported), errors.Is(err, agentworkspace.ErrCanvasInputUnconfirmed):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "画布输入解析失败"})
	}
}

// agentResolvedCanvasInputSnapshot records the request exactly as it will be
// submitted, together with the nodes and confirmed edge inputs that informed
// it. Settings are already persisted with every generation task, so this
// snapshot remains stable across queue retries.
func agentResolvedCanvasInputSnapshot(canvas agentworkspace.Canvas, nodeID, nodeType, submittedPrompt string, referenceAssetIDs []string) (map[string]any, error) {
	return agentResolvedCanvasInputSnapshotWithCapability(canvas, nodeID, submittedPrompt, referenceAssetIDs, agentCanvasInputCapability(nodeType))
}

func agentResolvedCanvasInputSnapshotWithCapability(canvas agentworkspace.Canvas, nodeID, submittedPrompt string, referenceAssetIDs []string, capability agentworkspace.CanvasInputCapability) (map[string]any, error) {
	inputs, err := agentworkspace.ResolveCanvasInputs(canvas, nodeID, capability)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"prompt":            strings.TrimSpace(submittedPrompt),
		"resolvedPrompt":    inputs.Prompt,
		"sourceNodeIds":     inputs.SourceNodeIDs,
		"images":            inputs.Images,
		"videos":            inputs.Videos,
		"audios":            inputs.Audios,
		"referenceAssetIds": uniqueAgentGenerationIDs(referenceAssetIDs),
		"warnings":          inputs.Warnings,
	}, nil
}
