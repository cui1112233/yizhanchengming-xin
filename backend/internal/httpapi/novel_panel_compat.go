package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func RegisterNovelPanelDiagnosticsCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string) {
	if root == nil || middleware.AuthStore == nil || usersDir == "" {
		return
	}
	root.Handle("GET /api/novel-panel/diagnostics", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		limit := 50
		if raw := req.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if limit > 100 {
			limit = 100
		}
		entries := []map[string]any{}
		path := filepath.Join(usersDir, identity.Username, "novel-panel", "ai-diagnostics.json")
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &entries)
		}
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, http.StatusOK, map[string]any{"diagnostics": entries})
	})))
}
