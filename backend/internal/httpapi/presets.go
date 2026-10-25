package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/prompt"
)

func registerPresetRoutes(root *http.ServeMux, auth BridgeAuth, catalog *prompt.Catalog) {
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/presets" {
			if r.Method != http.MethodGet {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			items, err := catalog.Published(r.URL.Query().Get("module"))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"catalog": items})
			return
		}
		const prefix = "/api/presets/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}
		presetPath := strings.TrimPrefix(r.URL.Path, prefix)
		if strings.HasPrefix(presetPath, "comic-style/") && strings.HasSuffix(presetPath, "/preview") {
			if r.Method != http.MethodGet {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			presetID := strings.TrimSuffix(strings.TrimPrefix(presetPath, "comic-style/"), "/preview")
			data, contentType, found := prompt.OpenComicStylePreview(presetID)
			if !found {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
			return
		}
		switch presetPath {
		case "constraint-text":
			if r.Method != http.MethodGet {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			parts := strings.Split(r.URL.Query().Get("ids"), ",")
			ids := make([]string, 0, len(parts))
			for _, part := range parts {
				if value := strings.TrimSpace(part); value != "" && len(ids) < 20 {
					ids = append(ids, value)
				}
			}
			texts, err := catalog.ConstraintText(ids)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"texts": texts})
		case "resolve":
			if r.Method != http.MethodPost {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			var input struct {
				Module    string   `json:"module"`
				PresetIDs []string `json:"presetIds"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid preset selection"})
				return
			}
			items, err := catalog.Resolve(strings.TrimSpace(input.Module), input.PresetIDs)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"presets": items})
		default:
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
		}
	}))
	root.Handle("/api/presets", handler)
	root.Handle("/api/presets/", handler)
}
