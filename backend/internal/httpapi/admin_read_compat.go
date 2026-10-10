package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"qiantie/backend/internal/auth"
)

func adminSystemFile(systemDir, name string) string {
	paths := []string{filepath.Join(systemDir, name), filepath.Join(systemDir, "..", "..", "..", "data", "system", name)}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return paths[0]
}

func RegisterAdminReadCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/admin/error-logs", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || toString(self["role"]) != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
			return
		}
		limit := 100
		if value, err := strconv.Atoi(req.URL.Query().Get("limit")); err == nil && value > 0 {
			limit = value
		}
		if limit > 500 {
			limit = 500
		}
		entries := []map[string]any{}
		if raw, err := os.ReadFile(adminSystemFile(profiles.SystemDir, "error-logs.json")); err == nil {
			_ = json.Unmarshal(raw, &entries)
		}
		if len(entries) > limit {
			entries = entries[len(entries)-limit:]
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
	})))
}
