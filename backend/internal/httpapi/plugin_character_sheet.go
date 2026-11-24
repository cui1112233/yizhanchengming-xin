package httpapi

import (
	"net/http"
	"strings"

	"qiantie/backend/internal/prompt"
)

// This is a user-requested prompt assembly, not a general catalog-body export.
func registerCharacterSheetPrompt(root *http.ServeMux, auth BridgeAuth, catalog *prompt.Catalog) {
	root.Handle("POST /api/plugins/character-turnaround/prompt", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			PresetID    string `json:"presetId"`
			Description string `json:"description"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		if catalog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "三视图模板目录未配置"})
			return
		}
		id, description := strings.TrimSpace(input.PresetID), strings.TrimSpace(input.Description)
		if id == "" || description == "" || len(description) > 100000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择模板并填写角色描述"})
			return
		}
		item, found, err := catalog.PublishedBody("shuihuo-production", id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "模板读取失败"})
			return
		}
		if !found || item.Slot != "shuihuo.asset.character-sheet" || strings.TrimSpace(item.Body) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "三视图模板未发布或归属不匹配"})
			return
		}
		// Keep composition in the server prompt module together with the catalog.
		result := prompt.BuildCharacterSheetPrompt(item.Body, description)
		writeJSON(w, http.StatusOK, map[string]any{"prompt": result, "preset": map[string]any{"id": item.ID, "revision": item.Revision}})
	})))
}
