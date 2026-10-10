package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/prompt"
)

func registerPersonalPromptRoutes(root *http.ServeMux, auth BridgeAuth, store *prompt.PersonalStore) {
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, _ := BridgeIdentityFromContext(r.Context())
		username := identity.Username
		if r.URL.Path == "/api/script-constraint-prompts" {
			if r.Method == http.MethodGet {
				items, err := store.List(username, r.URL.Query().Get("category"))
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "提示词参数无效"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"prompts": items})
				return
			}
			if r.Method == http.MethodPost {
				var input struct {
					Category string  `json:"category"`
					Name     *string `json:"name"`
					Body     string  `json:"body"`
				}
				if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&input); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "提示词参数无效"})
					return
				}
				item, err := store.CreateOrSaveDraft(username, input.Category, input.Name, input.Body)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "提示词参数无效"})
					return
				}
				writeJSON(w, http.StatusCreated, map[string]any{"prompt": item})
				return
			}
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		const prefix = "/api/script-constraint-prompts/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "提示词不存在"})
			return
		}
		tail := strings.TrimPrefix(r.URL.Path, prefix)
		if tail == "usage" && r.Method == http.MethodPost {
			var input struct {
				IDs []string `json:"ids"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&input); err != nil || input.IDs == nil || store.MarkUsed(username, input.IDs) != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "提示词参数无效"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}
		if tail == "" || strings.Contains(tail, "/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "提示词不存在"})
			return
		}
		if r.Method == http.MethodDelete {
			found, err := store.Remove(username, tail)
			if err != nil || !found {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "提示词不存在"})
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPut {
			var input struct {
				Name *string `json:"name"`
				Body *string `json:"body"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "提示词参数无效"})
				return
			}
			item, found, err := store.Update(username, tail, input.Name, input.Body)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "提示词参数无效"})
				return
			}
			if !found {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "提示词不存在"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"prompt": item})
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}))
	root.Handle("/api/script-constraint-prompts", handler)
	root.Handle("/api/script-constraint-prompts/", handler)
}
