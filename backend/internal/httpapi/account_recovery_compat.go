package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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
				for _, value := range recoveryVerifiedEmails(state) {
					record, ok := value.(map[string]any)
					if !ok || fmt.Sprint(record["username"]) != identity.Username || email == "" || fmt.Sprint(record["email"]) != email {
						continue
					}
					verified = true
					verifiedAt = record["verifiedAt"]
					break
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"email": email, "verified": verified, "verifiedAt": verifiedAt, "deliveryConfigured": false})
	})))
	root.Handle("POST /api/account-recovery/email/verify", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Token string `json:"token"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || strings.TrimSpace(input.Token) == "" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "验证链接无效"})
			return
		}
		path := filepath.Join(systemDir, "account-recovery.json")
		state := map[string]any{}
		if raw, err := os.ReadFile(path); err == nil {
			if json.Unmarshal(raw, &state) != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "账号找回状态损坏"})
				return
			}
		}
		digest := sha256.Sum256([]byte(strings.TrimSpace(input.Token)))
		digestText := hex.EncodeToString(digest[:])
		now := time.Now().UTC()
		var matched map[string]any
		for _, value := range recoveryEmailTokens(state) {
			row, ok := value.(map[string]any)
			if !ok || fmt.Sprint(row["tokenDigest"]) != digestText {
				continue
			}
			if recoveryTokenUsed(row["usedAt"]) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "验证链接已使用"})
				return
			}
			expires, err := time.Parse(time.RFC3339, fmt.Sprint(row["expiresAt"]))
			if err != nil || !expires.After(now) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "验证链接已过期"})
				return
			}
			matched = row
			break
		}
		if matched == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "验证链接无效"})
			return
		}
		username := fmt.Sprint(matched["username"])
		email := fmt.Sprint(matched["email"])
		matched["usedAt"] = now.Format(time.RFC3339)
		verified := make([]any, 0)
		for _, value := range recoveryVerifiedEmails(state) {
			row, ok := value.(map[string]any)
			if !ok || fmt.Sprint(row["username"]) != username {
				verified = append(verified, value)
			}
		}
		verified = append(verified, map[string]any{"username": username, "email": email, "verifiedAt": now.Format(time.RFC3339)})
		state["verifiedEmails"] = verified
		if err := writeRecoveryState(path, state); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "邮箱验证状态保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"verified": true, "email": email, "verifiedAt": now.Format(time.RFC3339)})
	}))
}

func recoveryVerifiedEmails(state map[string]any) []any {
	values, _ := state["verifiedEmails"].([]any)
	return values
}

func recoveryEmailTokens(state map[string]any) []any {
	values, _ := state["emailTokens"].([]any)
	return values
}

func recoveryTokenUsed(value any) bool {
	if value == nil {
		return false
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	return text != "" && text != "<nil>"
}

func writeRecoveryState(path string, state map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}
