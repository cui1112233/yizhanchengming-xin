package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/agentworkspace"
)

type agentCanvasInputPreviewRequest struct {
	NodeID string `json:"nodeId"`
}

func registerAgentCanvasInputPreviewRoute(root *http.ServeMux, middleware BridgeAuth, projects agentworkspace.Store) {
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
		resolved, err := agentworkspace.ResolveCanvasInputs(canvas, input.NodeID, agentCanvasInputCapability(node.Type))
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
