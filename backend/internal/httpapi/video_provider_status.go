package httpapi

import (
	"context"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

type videoProviderConfigReader interface {
	GetConfig(context.Context, string) (map[string]any, error)
}

func registerVideoProviderStatusRoute(mux *http.ServeMux, auth BridgeAuth, reader videoProviderConfigReader, version string) {
	if mux == nil || reader == nil {
		return
	}
	mux.Handle("GET /api/batch-factory/"+version+"/video-provider/status", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		config, err := reader.GetConfig(r.Context(), owner)
		if err != nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		writeJSON(w, http.StatusOK, videoProviderStatus(config, r.URL.Query().Get("provider")))
	})))
}

func videoProviderStatus(config map[string]any, rawProvider string) map[string]any {
	provider := batchfactoryv11.NormalizeVideoProviderForHTTP(rawProvider)
	model := ""
	configured := false
	video, _ := config["video"].(map[string]any)
	catalogVersion, _ := config["modelCatalogVersion"].(float64)
	if catalog, ok := config["modelCatalog"].([]any); ok && catalogVersion >= 1 {
		catalogID := map[string]string{
			batchfactoryv11.VideoProviderPersonalAPI:  "yd2-mini-video",
			batchfactoryv11.VideoProviderAutoDLH3:     "minimax-h3-video",
			batchfactoryv11.VideoProviderYFAISeedance: "seedance-2-0-official",
		}[provider]
		for _, item := range catalog {
			entry, ok := item.(map[string]any)
			if !ok || strings.TrimSpace(toString(entry["id"])) != catalogID {
				continue
			}
			model = strings.TrimSpace(toString(entry["modelId"]))
			configured = strings.TrimSpace(toString(entry["credential"])) != ""
			break
		}
	} else {
		switch provider {
		case batchfactoryv11.VideoProviderPersonalAPI:
			model = batchfactoryv11.DefaultPersonalVideoModel
			configured = strings.TrimSpace(toString(video["ydApiKey"])) != "" || strings.TrimSpace(toString(video["apiKey"])) != ""
		case batchfactoryv11.VideoProviderAutoDLH3:
			model = batchfactoryv11.AutoDLH3Model
			configured = strings.TrimSpace(toString(video["h3ApiKey"])) != "" || strings.TrimSpace(toString(video["apiKey"])) != ""
		case batchfactoryv11.VideoProviderYFAISeedance:
			model = "seedance-2-0-official"
		case batchfactoryv11.VideoProviderDoubaoLocal:
			model = "doubao-seedance"
		}
	}
	if model == "" {
		switch provider {
		case batchfactoryv11.VideoProviderPersonalAPI:
			model = batchfactoryv11.DefaultPersonalVideoModel
		case batchfactoryv11.VideoProviderAutoDLH3:
			model = batchfactoryv11.AutoDLH3Model
		case batchfactoryv11.VideoProviderYFAISeedance:
			model = "seedance-2-0-official"
		case batchfactoryv11.VideoProviderDoubaoLocal:
			model = "doubao-seedance"
		}
	}
	return map[string]any{"provider": provider, "model": model, "configured": configured}
}
