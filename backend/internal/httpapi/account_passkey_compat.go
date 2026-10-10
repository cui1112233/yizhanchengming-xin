package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// Passkey registration still stays on the WebAuthn implementation. These
// read/delete handlers keep the security page on Go without exposing keys.
func RegisterAccountPasskeyCompat(root *http.ServeMux, middleware BridgeAuth, profiles *MemberProfileStore) {
	if root == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/account-recovery/passkeys", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		credentials := readPasskeyCredentials(profiles, identity.Username)
		writeJSON(w, http.StatusOK, map[string]any{"passkeys": credentials})
	})))
	root.Handle("DELETE /api/account-recovery/passkeys/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		path := adminSystemFile(profiles.SystemDir, "passkeys.json")
		state := map[string]any{"credentials": []any{}, "challenges": []any{}}
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &state)
		}
		rows, _ := state["credentials"].([]any)
		id := req.PathValue("id")
		removed := map[string]any(nil)
		next := make([]any, 0, len(rows))
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if ok && removed == nil && toString(row["username"]) == identity.Username && (toString(row["id"]) == id || toString(row["credentialId"]) == id) {
				removed = map[string]any{"id": row["id"], "username": row["username"], "name": row["name"]}
				continue
			}
			next = append(next, raw)
		}
		if removed == nil {
			managedAccountError(w, http.StatusNotFound, "Passkey 不存在")
			return
		}
		state["credentials"] = next
		raw, err := json.MarshalIndent(state, "", "  ")
		if err != nil || os.WriteFile(path, raw, 0o600) != nil {
			managedAccountError(w, http.StatusInternalServerError, "Passkey 删除失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"removed": removed})
	})))
}

func readPasskeyCredentials(profiles *MemberProfileStore, username string) []map[string]any {
	path := adminSystemFile(profiles.SystemDir, "passkeys.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	state := map[string]any{}
	if json.Unmarshal(data, &state) != nil {
		return []map[string]any{}
	}
	rows, _ := state["credentials"].([]any)
	result := make([]map[string]any, 0)
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(toString(row["username"])) != username {
			continue
		}
		result = append(result, map[string]any{"id": row["id"], "username": row["username"], "credentialId": row["credentialId"], "name": row["name"], "algorithm": row["algorithm"], "transports": row["transports"], "signCount": row["signCount"], "createdAt": row["createdAt"], "lastUsedAt": row["lastUsedAt"]})
	}
	return result
}
