package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

// RegisterAccountRecoveryStatusCompat serves the harmless read-only portion of
// the recovery contract from the Go-owned profile files. Mail delivery,
// password reset, and passkey operations remain on their existing service.
func RegisterAccountRecoveryStatusCompat(root *http.ServeMux, middleware BridgeAuth, systemDir string) {
	if root == nil || middleware.AuthStore == nil || systemDir == "" {
		return
	}
	root.Handle("GET /api/account-recovery/email/status", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		email := ""
		if raw, err := os.ReadFile(filepath.Join(systemDir, "member-profile-details.json")); err == nil {
			var rows []map[string]any
			if json.Unmarshal(raw, &rows) == nil {
				for _, row := range rows {
					if row["username"] == identity.Username {
						if value, ok := row["email"].(string); ok {
							email = value
						}
						break
					}
				}
			}
		}
		verified, verifiedAt := false, any(nil)
		if raw, err := os.ReadFile(filepath.Join(systemDir, "account-recovery.json")); err == nil {
			var state map[string]any
			if json.Unmarshal(raw, &state) == nil {
				if record, ok := state[identity.Username].(map[string]any); ok {
					verified, _ = record["emailVerified"].(bool)
					verifiedAt = record["emailVerifiedAt"]
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"email": email, "verified": verified, "verifiedAt": verifiedAt, "deliveryConfigured": false})
	})))
}
