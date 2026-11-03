package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/giantmaterialresolver"
)

type GiantMaterialResolver interface {
	Resolve(context.Context, string) (giantmaterialresolver.Material, error)
}

func RegisterGiantMaterialResolverRoutes(root *http.ServeMux, auth BridgeAuth, resolver GiantMaterialResolver) {
	if root == nil || resolver == nil {
		return
	}
	root.Handle("POST /api/shuihuo-production/giant-material-resolve", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			GiantMaterialID string `json:"giantMaterialId"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		material, err := resolver.Resolve(req.Context(), input.GiantMaterialID)
		if err != nil {
			writeGiantMaterialResolverError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"ok": true, "stage": "resolved", "material": material})
	})))
}

func writeGiantMaterialResolverError(w http.ResponseWriter, err error) {
	code := giantmaterialresolver.CodeOf(err)
	if code == "" {
		code = "QINGYU_UPSTREAM_ERROR"
	}
	status := http.StatusBadGateway
	responseCode := code
	switch code {
	case "INVALID_GIANT_MATERIAL_ID":
		status = http.StatusBadRequest
	case "QINGYU_AUTH_NOT_CONFIGURED", "GIANT_PLUGIN_UNAVAILABLE":
		status = http.StatusServiceUnavailable
		// 两类旧码都表示“插件没配好”，统一收敛为稳定错误码；error 字段
		// 仍保留旧码以兼容现有调用方。
		responseCode = ErrorCodeGiantPluginNotConfigured
	case "QINGYU_AUTH_FAILED":
		status = http.StatusUnauthorized
	case "QINGYU_TIMEOUT", "GIANT_PLUGIN_TIMEOUT":
		status = http.StatusGatewayTimeout
	case "QINGYU_ENDPOINT_INVALID", "GIANT_PLUGIN_ENDPOINT_INVALID":
		status = http.StatusInternalServerError
	}
	writeGiantExecutorJSON(w, status, map[string]any{"error": code, "code": responseCode})
}
