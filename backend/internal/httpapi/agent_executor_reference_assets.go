package httpapi

import (
	"fmt"
	"io"
	"net/http"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/localexecutor"
)

// RegisterAgentExecutorReferenceAssetRoutes exposes project images only to a
// paired local executor that belongs to the same account.
func RegisterAgentExecutorReferenceAssetRoutes(root *http.ServeMux, executors *localexecutor.Service, assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore) {
	if root == nil || executors == nil || assets == nil || objects == nil {
		return
	}
	root.HandleFunc("GET /api/local-executor/v1/agent-reference-assets/{projectID}/{assetID}", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		owner, err := executors.OwnerForToken(req.Context(), token)
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		asset, err := assets.Asset(req.Context(), owner, req.PathValue("projectID"), req.PathValue("assetID"))
		if err != nil || asset.Kind != "image" {
			writeExecutorError(w, http.StatusNotFound, "reference image not found")
			return
		}
		body, err := objects.Open(req.Context(), asset)
		if err != nil {
			writeExecutorError(w, http.StatusBadGateway, "reference image is unavailable")
			return
		}
		defer body.Close()
		w.Header().Set("Content-Type", asset.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", asset.ByteSize))
		w.Header().Set("Cache-Control", "no-store")
		_, _ = io.Copy(w, body)
	})
}

func agentExecutorReferenceAssetURL(projectID, assetID string) string {
	return "/api/local-executor/v1/agent-reference-assets/" + projectID + "/" + assetID
}
