package httpapi

import (
	"net/http"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// RegisterShuihuoAccountAIConfigCompat accepts the short model shape sent by
// the retained 3001 compatibility gateway. The Go API Config catalog remains
// the source of truth; this adapter only translates the old top-level text and
// image fields into two managed models. Credentials are stored server-side and
// the response deliberately contains no configuration values.
func RegisterShuihuoAccountAIConfigCompat(root *http.ServeMux, middleware BridgeAuth, store novelfetchworkshop.Store) {
	if root == nil || store == nil {
		return
	}
	root.Handle("PUT /api/shuihuo-production/account-ai-config", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "无法读取模型配置"})
			return
		}
		catalog := modelCatalog(settings)
		updated := false
		for _, kind := range []string{"text", "image"} {
			payload := input
			if kind == "image" {
				if nested, nestedOK := input["image"].(map[string]any); nestedOK {
					payload = nested
				} else {
					continue
				}
			}
			model, modelOK := legacyShuihuoManagedModel(payload, kind)
			if !modelOK {
				continue
			}
			catalog = upsertLegacyShuihuoModel(catalog, model)
			if kind == "text" {
				settings["textModelId"] = model["id"]
			}
			updated = true
		}
		if !updated {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "至少需要一套完整的文本或图片模型配置"})
			return
		}
		settings["modelCatalog"] = catalog
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "模型配置保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})))
}

func legacyShuihuoManagedModel(payload map[string]any, kind string) (map[string]any, bool) {
	baseURL := strings.TrimSpace(firstText(payload["baseUrl"], payload["base_url"]))
	modelID := strings.TrimSpace(firstText(payload["model"], payload["modelId"], payload["model_id"]))
	credential := strings.TrimSpace(firstText(payload["apiKey"], payload["api_key"], payload["credential"]))
	if baseURL == "" || modelID == "" || credential == "" {
		return nil, false
	}
	model, err := normalizeManagedModel(map[string]any{
		"id":          "legacy-account-" + kind,
		"kind":        kind,
		"displayName": "兼容账号" + kind + "模型",
		"modelId":     modelID,
		"baseUrl":     baseURL,
		"credential":  credential,
		"enabled":     true,
	})
	if err != nil {
		return nil, false
	}
	return model, true
}

func upsertLegacyShuihuoModel(catalog []map[string]any, model map[string]any) []map[string]any {
	id, _ := model["id"].(string)
	for index, current := range catalog {
		if currentID, _ := current["id"].(string); currentID == id {
			catalog[index] = model
			return catalog
		}
	}
	return append(catalog, model)
}
