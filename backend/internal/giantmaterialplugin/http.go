package giantmaterialplugin

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/giantmaterialresolver"
)

type Resolver interface {
	Resolve(context.Context, string) (giantmaterialresolver.Material, error)
}

func NewQingyuHandler(pluginToken, upstreamEndpoint, upstreamToken string) http.Handler {
	return NewHandler(pluginToken, giantmaterialresolver.NewClient(nil, upstreamEndpoint, upstreamToken))
}

// NewHandler exposes the plugin's small, versioned HTTP contract. The plugin
// owns only upstream access; it has no access to the main application's users,
// batches, database, or Redis state.
func NewHandler(token string, resolver Resolver) http.Handler {
	expectedToken := strings.TrimSpace(token)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/v1/giant-material/resolve" {
			writeJSON(w, http.StatusNotFound, map[string]string{"code": "NOT_FOUND"})
			return
		}
		if resolver == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "PLUGIN_UNAVAILABLE"})
			return
		}
		if expectedToken == "" || req.Header.Get("Authorization") != "Bearer "+expectedToken {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"code": "UNAUTHORIZED"})
			return
		}
		var input struct {
			GiantMaterialID string `json:"giantMaterialId"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "INVALID_REQUEST"})
			return
		}
		material, err := resolver.Resolve(req.Context(), input.GiantMaterialID)
		if err != nil {
			writeResolverError(w, err)
			return
		}
		material.GiantMaterialID = strings.TrimSpace(input.GiantMaterialID)
		writeJSON(w, http.StatusOK, map[string]any{"material": material})
	})
}

func writeResolverError(w http.ResponseWriter, err error) {
	code := giantmaterialresolver.CodeOf(err)
	status := http.StatusBadGateway
	switch code {
	case "INVALID_GIANT_MATERIAL_ID":
		status = http.StatusBadRequest
	case "QINGYU_AUTH_NOT_CONFIGURED", "QINGYU_ENDPOINT_INVALID":
		status = http.StatusServiceUnavailable
	case "QINGYU_AUTH_FAILED":
		status = http.StatusUnauthorized
	case "QINGYU_TIMEOUT":
		status = http.StatusGatewayTimeout
	case "":
		code = "QINGYU_UPSTREAM_ERROR"
	}
	writeJSON(w, status, map[string]string{"code": code})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
