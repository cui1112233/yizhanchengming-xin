package httpapi

import (
	"encoding/json"
	"net/http"

	"qiantie/backend/internal/novelfetchworkshop"
)

// RegisterBatchRewriteConfigCompat exposes the read-only configuration shape
// expected by the legacy novel-fetch workbench. The actual settings are still
// stored in the Go workshop store, so the page no longer needs Node just to
// bootstrap its form.
func RegisterBatchRewriteConfigCompat(root *http.ServeMux, middleware BridgeAuth, store novelfetchworkshop.Store) {
	if root == nil || store == nil {
		return
	}
	read := func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取设置失败"})
			return
		}
		writeJSON(w, http.StatusOK, legacyBatchRewriteConfig(settings, req.URL.Path == "/api/batch-rewrite/bootstrap"))
	}
	handler := middleware.Middleware(http.HandlerFunc(read))
	root.Handle("GET /api/batch-rewrite/config", handler)
	root.Handle("GET /api/batch-rewrite/bootstrap", handler)
	root.Handle("POST /api/batch-rewrite/work-form", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			State map[string]any `json:"state"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 128<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取设置失败"})
			return
		}
		if body.State == nil {
			body.State = map[string]any{}
		}
		appConfig, ok := settings["appConfig"].(map[string]any)
		if !ok {
			appConfig = map[string]any{}
		}
		appConfig["work_form"] = body.State
		settings["appConfig"] = appConfig
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存设置失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "state": body.State})
	})))
}

func legacyBatchRewriteConfig(settings map[string]any, lightweight bool) map[string]any {
	appConfig := settings
	if configured, ok := settings["appConfig"].(map[string]any); ok {
		appConfig = configured
	}
	platforms := arrayOrEmpty(settings["platforms"])
	styles := arrayOrEmpty(settings["styles"])
	knowledge := arrayOrEmptyMap(appConfig["knowledge"])
	if lightweight {
		knowledge = map[string]any{}
	}
	workForm := mapOrEmpty(appConfig["work_form"])
	if len(workForm) == 0 {
		workForm = mapOrEmpty(settings["work_form"])
	}
	return map[string]any{
		"app_config":        appConfig,
		"platforms":         platforms,
		"styles":            styles,
		"sensitive":         mapOrEmpty(settings["sensitive"]),
		"web_submit":        mapOrEmpty(appConfig["web_submit"]),
		"knowledge":         knowledge,
		"knowledge_loaded":  !lightweight,
		"knowledge_summary": map[string]any{},
		"parse_modes": []any{
			map[string]any{"id": "smart", "name": "智能解析"},
			map[string]any{"id": "strict", "name": "严格解析"},
		},
		"column_presets": []any{},
		"work_form":      workForm,
	}
}

func arrayOrEmpty(value any) []any {
	if items, ok := value.([]any); ok {
		return items
	}
	return []any{}
}

func arrayOrEmptyMap(value any) map[string]any {
	if items, ok := value.(map[string]any); ok {
		return items
	}
	return map[string]any{}
}

func mapOrEmpty(value any) map[string]any {
	if items, ok := value.(map[string]any); ok {
		return items
	}
	return map[string]any{}
}
